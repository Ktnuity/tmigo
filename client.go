package tmigo

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Client represents a Twitch IRC client
type Client struct {
	*EventEmitter
	state  *clientState
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.RWMutex
}

// NewClient creates a new Twitch IRC client
func NewClient(opts *ClientOptions) *Client {
	if opts == nil {
		opts = &ClientOptions{}
	}

	// Set defaults
	if opts.Options == nil {
		opts.Options = &Options{}
	}
	if opts.Connection == nil {
		opts.Connection = &Connection{}
	}
	if opts.Identity == nil {
		opts.Identity = &Identity{}
	}
	if opts.Channels == nil {
		opts.Channels = []string{}
	}

	// Apply option defaults
	if opts.Options.GlobalDefaultChannel == "" {
		opts.Options.GlobalDefaultChannel = "#tmijs"
	}
	if opts.Options.JoinInterval == 0 {
		opts.Options.JoinInterval = 2000
	}
	if opts.Options.MessagesLogLevel == "" {
		opts.Options.MessagesLogLevel = "info"
	}

	// Apply connection defaults
	if opts.Connection.Server == "" {
		opts.Connection.Server = "irc-ws.chat.twitch.tv"
	}
	if opts.Connection.Port == 0 {
		opts.Connection.Port = 80
	}
	if opts.Connection.Secure {
		opts.Connection.Port = 443
	}
	if opts.Connection.Port == 443 {
		opts.Connection.Secure = true
	}
	if opts.Connection.ReconnectInterval == 0 {
		opts.Connection.ReconnectInterval = 1 * time.Second
	}
	if opts.Connection.ReconnectDecay == 0 {
		opts.Connection.ReconnectDecay = 1.5
	}
	if opts.Connection.MaxReconnectInterval == 0 {
		opts.Connection.MaxReconnectInterval = 30 * time.Second
	}
	if opts.Connection.MaxReconnectAttempts == 0 {
		opts.Connection.MaxReconnectAttempts = 999999 // Effectively infinite
	}
	if opts.Connection.Timeout == 0 {
		opts.Connection.Timeout = 9999 * time.Millisecond
	}
	opts.Connection.Reconnect = true // Default to true

	// Create logger
	logger := opts.Logger
	if logger == nil {
		logger = NewLogger()
	}
	if opts.Options.Debug {
		logger.SetLevel("info")
	} else {
		logger.SetLevel("error")
	}

	// Format channel names
	for i, ch := range opts.Channels {
		opts.Channels[i] = Channel(ch)
	}

	ctx, cancel := context.WithCancel(context.Background())

	state := &clientState{
		opts:                 opts,
		globalDefaultChannel: Channel(opts.Options.GlobalDefaultChannel),
		skipMembership:       opts.Options.SkipMembership,
		server:               opts.Connection.Server,
		port:                 opts.Connection.Port,
		secure:               opts.Connection.Secure,
		reconnect:            opts.Connection.Reconnect,
		reconnectDecay:       opts.Connection.ReconnectDecay,
		reconnectInterval:    opts.Connection.ReconnectInterval,
		maxReconnectInterval: opts.Connection.MaxReconnectInterval,
		maxReconnectAttempts: opts.Connection.MaxReconnectAttempts,
		reconnectTimer:       opts.Connection.ReconnectInterval,
		reconnecting:         false,
		reconnections:        0,
		username:             Username(opts.Identity.Username),
		channels:             []string{},
		emotes:               "",
		emotesets:            make(map[string]any),
		globalUserState:      GlobalUserState{},
		userState:            make(map[string]UserState),
		moderators:           make(map[string][]string),
		log:                  logger,
		currentLatency:       0,
		latency:              time.Now(),
		wasCloseCalled:       false,
	}

	// Generate justinfan username if none provided
	if state.username == "" {
		state.username = Justinfan()
	}

	client := &Client{
		EventEmitter: NewEventEmitter(),
		state:        state,
		ctx:          ctx,
		cancel:       cancel,
	}

	client.SetMaxListeners(0)

	return client
}

func (c *Client) SetLogLevel(logLevel string) error {
	return c.state.log.SetLevel(logLevel)
}

// Connect establishes a connection to the Twitch IRC server
func (c *Client) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Calculate reconnect timer
	c.state.reconnectTimer = time.Duration(float64(c.state.reconnectTimer) * c.state.reconnectDecay)
	c.state.reconnectTimer = min(c.state.reconnectTimer, c.state.maxReconnectInterval)

	return c.openConnection()
}

// openConnection opens a WebSocket connection to the server
func (c *Client) openConnection() error {
	protocol := "ws"
	if c.state.secure {
		protocol = "wss"
	}

	url := fmt.Sprintf("%s://%s:%d/", protocol, c.state.server, c.state.port)

	c.state.log.Info(fmt.Sprintf("Connecting to %s on port %d..", c.state.server, c.state.port))
	c.Emit("connecting", c.state.server, c.state.port)

	dialer := websocket.DefaultDialer
	ws, _, err := dialer.Dial(url, nil)
	if err != nil {
		c.state.log.Error(fmt.Sprintf("Connection error: %v", err))
		return err
	}

	c.state.ws = ws

	// Start handling messages
	go c.handleMessages()

	// Send authentication
	return c.authenticate()
}

// authenticate sends authentication to the server
func (c *Client) authenticate() error {
	c.state.log.Info("Sending authentication to server..")
	c.Emit("logon")

	// Request capabilities
	caps := "twitch.tv/tags twitch.tv/commands"
	if !c.state.skipMembership {
		caps += " twitch.tv/membership"
	}

	if err := c.state.ws.WriteMessage(websocket.TextMessage, fmt.Appendf(nil, "CAP REQ :%s", caps)); err != nil {
		return err
	}

	// Send password if provided
	password := c.state.opts.Identity.Password
	if password != "" {
		password = Password(password)
		if err := c.state.ws.WriteMessage(websocket.TextMessage, fmt.Appendf(nil, "PASS %s", password)); err != nil {
			return err
		}
	} else if IsJustinfan(c.state.username) {
		if err := c.state.ws.WriteMessage(websocket.TextMessage, []byte("PASS SCHMOOPIIE")); err != nil {
			return err
		}
	}

	// Send NICK
	if err := c.state.ws.WriteMessage(websocket.TextMessage, fmt.Appendf(nil, "NICK %s", c.state.username)); err != nil {
		return err
	}

	return nil
}

// handleMessages processes incoming WebSocket messages
func (c *Client) handleMessages() {
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			_, data, err := c.state.ws.ReadMessage()
			if err != nil {
				c.handleError(err)
				return
			}

			// Split by \r\n for multiple messages
			messages := strings.Split(strings.TrimSpace(string(data)), "\r\n")
			for _, msgStr := range messages {
				if msgStr == "" {
					continue
				}
				msg := ParseMessage(msgStr)
				if msg != nil {
					c.handleMessage(msg)
				}
			}
		}
	}
}

// handleError handles connection errors
func (c *Client) handleError(err error) {
	c.state.moderators = make(map[string][]string)
	c.state.userState = make(map[string]UserState)
	c.state.globalUserState = GlobalUserState{}

	if c.state.pingLoop != nil {
		c.state.pingLoop.Stop()
	}
	if c.state.pingTimeout != nil {
		c.state.pingTimeout.Stop()
	}

	c.state.reason = "Connection closed."
	if err != nil {
		c.state.reason = fmt.Sprintf("Unable to connect: %v", err)
	}

	c.Emit("disconnected", c.state.reason)

	// Reconnect logic
	if c.state.reconnect && c.state.reconnections < c.state.maxReconnectAttempts && !c.state.wasCloseCalled {
		c.state.reconnecting = true
		c.state.reconnections++

		c.state.log.Error(fmt.Sprintf("Reconnecting in %v..", c.state.reconnectTimer))
		c.Emit("reconnect")

		time.AfterFunc(c.state.reconnectTimer, func() {
			c.state.reconnecting = false
			c.Connect()
		})
	} else if c.state.reconnections >= c.state.maxReconnectAttempts {
		c.Emit("maxreconnect")
		c.state.log.Error("Maximum reconnection attempts reached.")
	}

	c.state.ws = nil
}

// Disconnect closes the connection to the server
func (c *Client) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.state.ws == nil {
		return errors.New("not connected to server")
	}

	c.state.wasCloseCalled = true
	c.state.log.Info("Disconnecting from server..")

	err := c.state.ws.Close()
	c.cancel()

	c.Emit("disconnected", "Connection closed.")

	return err
}

// GetUsername returns the current username
func (c *Client) GetUsername() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state.username
}

// GetChannels returns the list of joined channels
func (c *Client) GetChannels() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	channels := make([]string, len(c.state.channels))
	copy(channels, c.state.channels)
	return channels
}

// IsMod checks if a username is a moderator in a channel
func (c *Client) IsMod(channel, username string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	ch := Channel(channel)
	mods, exists := c.state.moderators[ch]
	if !exists {
		return false
	}

	user := Username(username)
	return slices.Contains(mods, user)
}

// ReadyState returns the current connection state
func (c *Client) ReadyState() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.state.ws == nil {
		return "CLOSED"
	}

	// WebSocket states: connecting(0), open(1), closing(2), closed(3)
	// We'll just return OPEN or CLOSED for simplicity
	return "OPEN"
}

// isConnected checks if the WebSocket is connected
func (c *Client) isConnected() bool {
	return c.state.ws != nil
}

// Type-safe event handlers
// These methods provide compile-time type safety for event handlers
// instead of requiring manual type assertions from ...any arguments.

// OnMessage registers a type-safe handler for message events (both chat and action)
func (c *Client) OnMessage(handler func(channel string, userstate ChatUserstate, message string, self bool)) *Client {
	c.On("message", func(args ...any) {
		if len(args) >= 4 {
			channel, _ := args[0].(string)
			userstate, _ := args[1].(ChatUserstate)
			message, _ := args[2].(string)
			self, _ := args[3].(bool)
			handler(channel, userstate, message, self)
		}
	})
	return c
}

// OnChat registers a type-safe handler for regular chat messages
func (c *Client) OnChat(handler func(channel string, userstate ChatUserstate, message string, self bool)) *Client {
	c.On("chat", func(args ...any) {
		if len(args) >= 4 {
			channel, _ := args[0].(string)
			userstate, _ := args[1].(ChatUserstate)
			message, _ := args[2].(string)
			self, _ := args[3].(bool)
			handler(channel, userstate, message, self)
		}
	})
	return c
}

// OnAction registers a type-safe handler for action messages (/me)
func (c *Client) OnAction(handler func(channel string, userstate ChatUserstate, message string, self bool)) *Client {
	c.On("action", func(args ...any) {
		if len(args) >= 4 {
			channel, _ := args[0].(string)
			userstate, _ := args[1].(ChatUserstate)
			message, _ := args[2].(string)
			self, _ := args[3].(bool)
			handler(channel, userstate, message, self)
		}
	})
	return c
}

// OnWhisper registers a type-safe handler for whisper messages
func (c *Client) OnWhisper(handler func(from string, userstate ChatUserstate, message string, self bool)) *Client {
	c.On("whisper", func(args ...any) {
		if len(args) >= 4 {
			from, _ := args[0].(string)
			userstate, _ := args[1].(ChatUserstate)
			message, _ := args[2].(string)
			self, _ := args[3].(bool)
			handler(from, userstate, message, self)
		}
	})
	return c
}

// OnCheer registers a type-safe handler for cheer (bits) events
func (c *Client) OnCheer(handler func(channel string, userstate ChatUserstate, message string)) *Client {
	c.On("cheer", func(args ...any) {
		if len(args) >= 3 {
			channel, _ := args[0].(string)
			userstate, _ := args[1].(ChatUserstate)
			message, _ := args[2].(string)
			handler(channel, userstate, message)
		}
	})
	return c
}

// OnSubscription registers a type-safe handler for subscription events
func (c *Client) OnSubscription(handler func(channel string, username string, methods SubMethods, message string, userstate SubUserstate)) *Client {
	c.On("subscription", func(args ...any) {
		if len(args) >= 5 {
			channel, _ := args[0].(string)
			username, _ := args[1].(string)
			methods, _ := args[2].(SubMethods)
			message, _ := args[3].(string)
			userstate, _ := args[4].(SubUserstate)
			handler(channel, username, methods, message, userstate)
		}
	})
	return c
}

// OnResub registers a type-safe handler for resubscription events
func (c *Client) OnResub(handler func(channel string, username string, months int, message string, userstate SubUserstate, methods SubMethods)) *Client {
	c.On("resub", func(args ...any) {
		if len(args) >= 6 {
			channel, _ := args[0].(string)
			username, _ := args[1].(string)
			months, _ := args[2].(int)
			message, _ := args[3].(string)
			userstate, _ := args[4].(SubUserstate)
			methods, _ := args[5].(SubMethods)
			handler(channel, username, months, message, userstate, methods)
		}
	})
	return c
}

// OnSubGift registers a type-safe handler for gifted subscription events
func (c *Client) OnSubGift(handler func(channel string, username string, streakMonths int, recipient string, methods SubMethods, userstate SubGiftUserstate)) *Client {
	c.On("subgift", func(args ...any) {
		if len(args) >= 6 {
			channel, _ := args[0].(string)
			username, _ := args[1].(string)
			streakMonths, _ := args[2].(int)
			recipient, _ := args[3].(string)
			methods, _ := args[4].(SubMethods)
			userstate, _ := args[5].(SubGiftUserstate)
			handler(channel, username, streakMonths, recipient, methods, userstate)
		}
	})
	return c
}

// OnSubMysteryGift registers a type-safe handler for mystery gift subscription events
func (c *Client) OnSubMysteryGift(handler func(channel string, username string, numbOfSubs int, methods SubMethods, userstate SubMysteryGiftUserstate)) *Client {
	c.On("submysterygift", func(args ...any) {
		if len(args) >= 5 {
			channel, _ := args[0].(string)
			username, _ := args[1].(string)
			numbOfSubs, _ := args[2].(int)
			methods, _ := args[3].(SubMethods)
			userstate, _ := args[4].(SubMysteryGiftUserstate)
			handler(channel, username, numbOfSubs, methods, userstate)
		}
	})
	return c
}

// OnAnonSubGift registers a type-safe handler for anonymous gifted subscriptions
func (c *Client) OnAnonSubGift(handler func(channel string, streakMonths int, recipient string, methods SubMethods, userstate AnonSubGiftUserstate)) *Client {
	c.On("anonsubgift", func(args ...any) {
		if len(args) >= 5 {
			channel, _ := args[0].(string)
			streakMonths, _ := args[1].(int)
			recipient, _ := args[2].(string)
			methods, _ := args[3].(SubMethods)
			userstate, _ := args[4].(AnonSubGiftUserstate)
			handler(channel, streakMonths, recipient, methods, userstate)
		}
	})
	return c
}

// OnAnonSubMysteryGift registers a type-safe handler for anonymous mystery gift subscriptions
func (c *Client) OnAnonSubMysteryGift(handler func(channel string, numbOfSubs int, methods SubMethods, userstate AnonSubMysteryGiftUserstate)) *Client {
	c.On("anonsubmysterygift", func(args ...any) {
		if len(args) >= 4 {
			channel, _ := args[0].(string)
			numbOfSubs, _ := args[1].(int)
			methods, _ := args[2].(SubMethods)
			userstate, _ := args[3].(AnonSubMysteryGiftUserstate)
			handler(channel, numbOfSubs, methods, userstate)
		}
	})
	return c
}

// OnGiftPaidUpgrade registers a type-safe handler for gift subscription upgrades
func (c *Client) OnGiftPaidUpgrade(handler func(channel string, username string, sender string, userstate SubGiftUpgradeUserstate)) *Client {
	c.On("giftpaidupgrade", func(args ...any) {
		if len(args) >= 4 {
			channel, _ := args[0].(string)
			username, _ := args[1].(string)
			sender, _ := args[2].(string)
			userstate, _ := args[3].(SubGiftUpgradeUserstate)
			handler(channel, username, sender, userstate)
		}
	})
	return c
}

// OnAnonGiftPaidUpgrade registers a type-safe handler for anonymous gift subscription upgrades
func (c *Client) OnAnonGiftPaidUpgrade(handler func(channel string, username string, userstate AnonSubGiftUpgradeUserstate)) *Client {
	c.On("anongiftpaidupgrade", func(args ...any) {
		if len(args) >= 3 {
			channel, _ := args[0].(string)
			username, _ := args[1].(string)
			userstate, _ := args[2].(AnonSubGiftUpgradeUserstate)
			handler(channel, username, userstate)
		}
	})
	return c
}

// OnPrimePaidUpgrade registers a type-safe handler for Prime subscription upgrades
func (c *Client) OnPrimePaidUpgrade(handler func(channel string, username string, methods SubMethods, userstate PrimeUpgradeUserstate)) *Client {
	c.On("primepaidupgrade", func(args ...any) {
		if len(args) >= 4 {
			channel, _ := args[0].(string)
			username, _ := args[1].(string)
			methods, _ := args[2].(SubMethods)
			userstate, _ := args[3].(PrimeUpgradeUserstate)
			handler(channel, username, methods, userstate)
		}
	})
	return c
}

// OnRaided registers a type-safe handler for raid events
func (c *Client) OnRaided(handler func(channel string, username string, viewers int)) *Client {
	c.On("raided", func(args ...any) {
		if len(args) >= 3 {
			channel, _ := args[0].(string)
			username, _ := args[1].(string)
			viewers, _ := args[2].(int)
			handler(channel, username, viewers)
		}
	})
	return c
}

// OnRedeem registers a type-safe handler for channel point redemption events
func (c *Client) OnRedeem(handler func(channel string, username string, rewardType string, tags ChatUserstate, message string)) *Client {
	c.On("redeem", func(args ...any) {
		if len(args) >= 5 {
			channel, _ := args[0].(string)
			username, _ := args[1].(string)
			rewardType, _ := args[2].(string)
			tags, _ := args[3].(ChatUserstate)
			message, _ := args[4].(string)
			handler(channel, username, rewardType, tags, message)
		}
	})
	return c
}

// OnBan registers a type-safe handler for ban events
func (c *Client) OnBan(handler func(channel string, username string, reason string, userstate BanUserstate)) *Client {
	c.On("ban", func(args ...any) {
		if len(args) >= 4 {
			channel, _ := args[0].(string)
			username, _ := args[1].(string)
			reason, _ := args[2].(string)
			userstate, _ := args[3].(BanUserstate)
			handler(channel, username, reason, userstate)
		}
	})
	return c
}

// OnTimeout registers a type-safe handler for timeout events
func (c *Client) OnTimeout(handler func(channel string, username string, reason string, duration int, userstate TimeoutUserstate)) *Client {
	c.On("timeout", func(args ...any) {
		if len(args) >= 5 {
			channel, _ := args[0].(string)
			username, _ := args[1].(string)
			reason, _ := args[2].(string)
			duration, _ := args[3].(int)
			userstate, _ := args[4].(TimeoutUserstate)
			handler(channel, username, reason, duration, userstate)
		}
	})
	return c
}

// OnMessageDeleted registers a type-safe handler for deleted message events
func (c *Client) OnMessageDeleted(handler func(channel string, username string, deletedMessage string, userstate DeleteUserstate)) *Client {
	c.On("messagedeleted", func(args ...any) {
		if len(args) >= 4 {
			channel, _ := args[0].(string)
			username, _ := args[1].(string)
			deletedMessage, _ := args[2].(string)
			userstate, _ := args[3].(DeleteUserstate)
			handler(channel, username, deletedMessage, userstate)
		}
	})
	return c
}

// OnJoin registers a type-safe handler for join events
func (c *Client) OnJoin(handler func(channel string, username string, self bool)) *Client {
	c.On("join", func(args ...any) {
		if len(args) >= 3 {
			channel, _ := args[0].(string)
			username, _ := args[1].(string)
			self, _ := args[2].(bool)
			handler(channel, username, self)
		}
	})
	return c
}

// OnPart registers a type-safe handler for part (leave) events
func (c *Client) OnPart(handler func(channel string, username string, self bool)) *Client {
	c.On("part", func(args ...any) {
		if len(args) >= 3 {
			channel, _ := args[0].(string)
			username, _ := args[1].(string)
			self, _ := args[2].(bool)
			handler(channel, username, self)
		}
	})
	return c
}

// OnHosted registers a type-safe handler for hosted events
func (c *Client) OnHosted(handler func(channel string, username string, viewers int, autohost bool)) *Client {
	c.On("hosted", func(args ...any) {
		if len(args) >= 4 {
			channel, _ := args[0].(string)
			username, _ := args[1].(string)
			viewers, _ := args[2].(int)
			autohost, _ := args[3].(bool)
			handler(channel, username, viewers, autohost)
		}
	})
	return c
}

// OnHosting registers a type-safe handler for hosting events
func (c *Client) OnHosting(handler func(channel string, target string, viewers int)) *Client {
	c.On("hosting", func(args ...any) {
		if len(args) >= 3 {
			channel, _ := args[0].(string)
			target, _ := args[1].(string)
			viewers, _ := args[2].(int)
			handler(channel, target, viewers)
		}
	})
	return c
}

// OnUnhost registers a type-safe handler for unhost events
func (c *Client) OnUnhost(handler func(channel string, viewers int)) *Client {
	c.On("unhost", func(args ...any) {
		if len(args) >= 2 {
			channel, _ := args[0].(string)
			viewers, _ := args[1].(int)
			handler(channel, viewers)
		}
	})
	return c
}

// OnMod registers a type-safe handler for mod events
func (c *Client) OnMod(handler func(channel string, username string)) *Client {
	c.On("mod", func(args ...any) {
		if len(args) >= 2 {
			channel, _ := args[0].(string)
			username, _ := args[1].(string)
			handler(channel, username)
		}
	})
	return c
}

// OnUnmod registers a type-safe handler for unmod events
func (c *Client) OnUnmod(handler func(channel string, username string)) *Client {
	c.On("unmod", func(args ...any) {
		if len(args) >= 2 {
			channel, _ := args[0].(string)
			username, _ := args[1].(string)
			handler(channel, username)
		}
	})
	return c
}

// OnMods registers a type-safe handler for mods list events
func (c *Client) OnMods(handler func(channel string, mods []string)) *Client {
	c.On("mods", func(args ...any) {
		if len(args) >= 2 {
			channel, _ := args[0].(string)
			mods, _ := args[1].([]string)
			handler(channel, mods)
		}
	})
	return c
}

// OnVips registers a type-safe handler for VIPs list events
func (c *Client) OnVips(handler func(channel string, vips []string)) *Client {
	c.On("vips", func(args ...any) {
		if len(args) >= 2 {
			channel, _ := args[0].(string)
			vips, _ := args[1].([]string)
			handler(channel, vips)
		}
	})
	return c
}

// OnNotice registers a type-safe handler for notice events
func (c *Client) OnNotice(handler func(channel string, msgid MsgID, message string)) *Client {
	c.On("notice", func(args ...any) {
		if len(args) >= 3 {
			channel, _ := args[0].(string)
			msgid, _ := args[1].(MsgID)
			message, _ := args[2].(string)
			handler(channel, msgid, message)
		}
	})
	return c
}

// OnRoomstate registers a type-safe handler for roomstate events
func (c *Client) OnRoomstate(handler func(channel string, state RoomState)) *Client {
	c.On("roomstate", func(args ...any) {
		if len(args) >= 2 {
			channel, _ := args[0].(string)
			state, _ := args[1].(RoomState)
			handler(channel, state)
		}
	})
	return c
}

// OnClearchat registers a type-safe handler for clearchat events
func (c *Client) OnClearchat(handler func(channel string)) *Client {
	c.On("clearchat", func(args ...any) {
		if len(args) >= 1 {
			channel, _ := args[0].(string)
			handler(channel)
		}
	})
	return c
}

// OnEmoteonly registers a type-safe handler for emote-only mode events
func (c *Client) OnEmoteonly(handler func(channel string, enabled bool)) *Client {
	c.On("emoteonly", func(args ...any) {
		if len(args) >= 2 {
			channel, _ := args[0].(string)
			enabled, _ := args[1].(bool)
			handler(channel, enabled)
		}
	})
	return c
}

// OnFollowersonly registers a type-safe handler for followers-only mode events
func (c *Client) OnFollowersonly(handler func(channel string, enabled bool, length int)) *Client {
	c.On("followersonly", func(args ...any) {
		if len(args) >= 3 {
			channel, _ := args[0].(string)
			enabled, _ := args[1].(bool)
			length, _ := args[2].(int)
			handler(channel, enabled, length)
		}
	})
	return c
}

// OnSlowmode registers a type-safe handler for slow mode events
func (c *Client) OnSlowmode(handler func(channel string, enabled bool, length int)) *Client {
	c.On("slowmode", func(args ...any) {
		if len(args) >= 3 {
			channel, _ := args[0].(string)
			enabled, _ := args[1].(bool)
			length, _ := args[2].(int)
			handler(channel, enabled, length)
		}
	})
	return c
}

// OnSubscribers registers a type-safe handler for subscribers-only mode events
func (c *Client) OnSubscribers(handler func(channel string, enabled bool)) *Client {
	c.On("subscribers", func(args ...any) {
		if len(args) >= 2 {
			channel, _ := args[0].(string)
			enabled, _ := args[1].(bool)
			handler(channel, enabled)
		}
	})
	return c
}

// OnR9kbeta registers a type-safe handler for R9K mode events
func (c *Client) OnR9kbeta(handler func(channel string, enabled bool)) *Client {
	c.On("r9kbeta", func(args ...any) {
		if len(args) >= 2 {
			channel, _ := args[0].(string)
			enabled, _ := args[1].(bool)
			handler(channel, enabled)
		}
	})
	return c
}

// OnConnected registers a type-safe handler for connected events
func (c *Client) OnConnected(handler func(address string, port int)) *Client {
	c.On("connected", func(args ...any) {
		if len(args) >= 2 {
			address, _ := args[0].(string)
			port, _ := args[1].(int)
			handler(address, port)
		}
	})
	return c
}

// OnConnecting registers a type-safe handler for connecting events
func (c *Client) OnConnecting(handler func(address string, port int)) *Client {
	c.On("connecting", func(args ...any) {
		if len(args) >= 2 {
			address, _ := args[0].(string)
			port, _ := args[1].(int)
			handler(address, port)
		}
	})
	return c
}

// OnDisconnected registers a type-safe handler for disconnected events
func (c *Client) OnDisconnected(handler func(reason string)) *Client {
	c.On("disconnected", func(args ...any) {
		if len(args) >= 1 {
			reason, _ := args[0].(string)
			handler(reason)
		}
	})
	return c
}

// OnLogon registers a type-safe handler for logon events
func (c *Client) OnLogon(handler func()) *Client {
	c.On("logon", func(args ...any) {
		handler()
	})
	return c
}

// OnReconnect registers a type-safe handler for reconnect events
func (c *Client) OnReconnect(handler func()) *Client {
	c.On("reconnect", func(args ...any) {
		handler()
	})
	return c
}

// OnPing registers a type-safe handler for ping events
func (c *Client) OnPing(handler func()) *Client {
	c.On("ping", func(args ...any) {
		handler()
	})
	return c
}

// OnPong registers a type-safe handler for pong events
func (c *Client) OnPong(handler func(latency float64)) *Client {
	c.On("pong", func(args ...any) {
		if len(args) >= 1 {
			latency, _ := args[0].(float64)
			handler(latency)
		}
	})
	return c
}

// OnEmotesets registers a type-safe handler for emoteset events
func (c *Client) OnEmotesets(handler func(sets string, obj map[string]any)) *Client {
	c.On("emotesets", func(args ...any) {
		if len(args) >= 2 {
			sets, _ := args[0].(string)
			obj, _ := args[1].(map[string]any)
			handler(sets, obj)
		}
	})
	return c
}

// OnRawMessage registers a type-safe handler for raw IRC message events
func (c *Client) OnRawMessage(handler func(message *IRCMessage)) *Client {
	c.On("raw_message", func(args ...any) {
		if len(args) >= 1 {
			message, _ := args[0].(*IRCMessage)
			handler(message)
		}
	})
	return c
}

// OnAnnouncement registers a type-safe handler for announcement events
func (c *Client) OnAnnouncement(handler func(channel string, userstate ChatUserstate, message string, self bool, color string)) *Client {
	c.On("announcement", func(args ...any) {
		if len(args) >= 5 {
			channel, _ := args[0].(string)
			userstate, _ := args[1].(ChatUserstate)
			message, _ := args[2].(string)
			self, _ := args[3].(bool)
			color, _ := args[4].(string)
			handler(channel, userstate, message, self, color)
		}
	})
	return c
}
