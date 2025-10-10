// Package tmigo provides a Go implementation of the Twitch Messaging Interface (TMI)
// for interacting with Twitch IRC chat. It's a port of the tmi.js library.
//
// Basic usage:
//
//	client := tmigo.NewClient(&tmigo.ClientOptions{
//	    Identity: &tmigo.Identity{
//	        Username: "your_bot_name",
//	        Password: "oauth:your_oauth_token",
//	    },
//	    Channels: []string{"channel1", "channel2"},
//	})
//
//	client.On("message", func(args ...any) {
//	    channel := args[0].(string)
//	    message := args[2].(string)
//	    // Handle message
//	})
//
//	client.Connect()
package tmigo

import (
	"time"

	"github.com/gorilla/websocket"
)

// ClientOptions contains all configuration for the TMI client
type ClientOptions struct {
	Options    *Options
	Connection *Connection
	Identity   *Identity
	Channels   []string
	Logger     Logger
}

// Options contains general client options
type Options struct {
	Debug                bool
	GlobalDefaultChannel string
	SkipMembership       bool
	JoinInterval         int
	MessagesLogLevel     string
}

// Connection contains WebSocket connection options
type Connection struct {
	Server               string
	Port                 int
	Secure               bool
	Reconnect            bool
	ReconnectDecay       float64
	ReconnectInterval    time.Duration
	MaxReconnectInterval time.Duration
	MaxReconnectAttempts int
	Timeout              time.Duration
}

// Identity contains authentication credentials
type Identity struct {
	Username string
	Password string
}

// IRCMessage represents a parsed IRC message
type IRCMessage struct {
	Raw     string
	Tags    map[string]any
	Prefix  string
	Command string
	Params  []string
}

// BadgeInfo represents badge-info from Twitch
type BadgeInfo map[string]string

// Badges represents user badges
type Badges map[string]string

// Emotes represents emote data
type Emotes []EmotePosition

// EmotePosition represents an emote and its positions in a message
type EmotePosition struct {
	ID        string
	Positions []EmoteRange
}

// EmoteRange represents the start and end positions of an emote
type EmoteRange struct {
	Start int
	End   int
}

// GlobalUserState contains global user state information
type GlobalUserState struct {
	BadgeInfo    BadgeInfo
	Badges       Badges
	Color        string
	DisplayName  string
	EmoteSets    string
	UserID       string
	UserType     string
	BadgeInfoRaw string
	BadgesRaw    string
}

// UserState contains user state for a channel
type UserState struct {
	BadgeInfo    BadgeInfo
	Badges       Badges
	Color        string
	DisplayName  string
	EmoteSets    string
	Mod          bool
	Subscriber   bool
	UserType     string
	BadgeInfoRaw string
	BadgesRaw    string
	Username     string
}

// Client state
type clientState struct {
	// Connection
	ws             *websocket.Conn
	server         string
	port           int
	secure         bool
	reconnecting   bool
	reconnections  int
	reconnectTimer time.Duration
	wasCloseCalled bool
	reason         string
	currentLatency time.Duration
	latency        time.Time
	pingLoop       *time.Ticker
	pingTimeout    *time.Timer

	// Connection settings
	maxReconnectAttempts int
	maxReconnectInterval time.Duration
	reconnect            bool
	reconnectDecay       float64
	reconnectInterval    time.Duration

	// User state
	username        string
	emotes          string
	emotesets       map[string]any
	channels        []string
	globalUserState GlobalUserState
	userState       map[string]UserState
	lastJoined      string
	moderators      map[string][]string

	// Settings
	opts                 *ClientOptions
	globalDefaultChannel string
	skipMembership       bool

	// Logger
	log Logger
}

// EventHandler is a function that handles events
type EventHandler func(args ...any)

// OutgoingTags represents tags to send with outgoing messages
type OutgoingTags map[string]string
