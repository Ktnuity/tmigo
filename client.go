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
