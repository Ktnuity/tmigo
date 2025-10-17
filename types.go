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
// TypeScript: { subscriber?: string; [other: string]: string | undefined; }
type BadgeInfo map[string]string

// Badges represents user badges from Twitch
// TypeScript: { admin?: string; bits?: string; ... [other: string]: string | undefined; }
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

// EmoteObj represents emote sets
type EmoteObj map[string][]EmoteSet

// EmoteSet represents a single emote in a set
type EmoteSet struct {
	Code string `json:"code"`
	ID   int    `json:"id"`
}

// SubMethod represents subscription tier
type SubMethod string

const (
	SubMethodPrime SubMethod = "Prime"
	SubMethod1000  SubMethod = "1000"
	SubMethod2000  SubMethod = "2000"
	SubMethod3000  SubMethod = "3000"
)

// SubMethods contains subscription method information
type SubMethods struct {
	Prime    bool      `json:"prime,omitempty"`
	Plan     SubMethod `json:"plan,omitempty"`
	PlanName string    `json:"planName,omitempty"`
}

// MsgID represents Twitch notice message IDs
type MsgID string

const (
	MsgIDAlreadyBanned          MsgID = "already_banned"
	MsgIDAlreadyEmoteOnlyOn     MsgID = "already_emote_only_on"
	MsgIDAlreadyEmoteOnlyOff    MsgID = "already_emote_only_off"
	MsgIDAlreadySubsOn          MsgID = "already_subs_on"
	MsgIDAlreadySubsOff         MsgID = "already_subs_off"
	MsgIDBadBanAdmin            MsgID = "bad_ban_admin"
	MsgIDBadBanAnon             MsgID = "bad_ban_anon"
	MsgIDBadBanBroadcaster      MsgID = "bad_ban_broadcaster"
	MsgIDBadBanGlobalMod        MsgID = "bad_ban_global_mod"
	MsgIDBadBanMod              MsgID = "bad_ban_mod"
	MsgIDBadBanSelf             MsgID = "bad_ban_self"
	MsgIDBadBanStaff            MsgID = "bad_ban_staff"
	MsgIDBadCommercialError     MsgID = "bad_commercial_error"
	MsgIDBadHostHosting         MsgID = "bad_host_hosting"
	MsgIDBadHostRateExceeded    MsgID = "bad_host_rate_exceeded"
	MsgIDBadModMod              MsgID = "bad_mod_mod"
	MsgIDBadModBanned           MsgID = "bad_mod_banned"
	MsgIDBadTimeoutAdmin        MsgID = "bad_timeout_admin"
	MsgIDBadTimeoutAnon         MsgID = "bad_timeout_anon"
	MsgIDBadTimeoutGlobalMod    MsgID = "bad_timeout_global_mod"
	MsgIDBadTimeoutMod          MsgID = "bad_timeout_mod"
	MsgIDBadTimeoutSelf         MsgID = "bad_timeout_self"
	MsgIDBadTimeoutStaff        MsgID = "bad_timeout_staff"
	MsgIDBadUnbanNoBan          MsgID = "bad_unban_no_ban"
	MsgIDBadUnmodMod            MsgID = "bad_unmod_mod"
	MsgIDBanSuccess             MsgID = "ban_success"
	MsgIDCmdsAvailable          MsgID = "cmds_available"
	MsgIDColorChanged           MsgID = "color_changed"
	MsgIDCommercialSuccess      MsgID = "commercial_success"
	MsgIDEmoteOnlyOn            MsgID = "emote_only_on"
	MsgIDEmoteOnlyOff           MsgID = "emote_only_off"
	MsgIDHostsRemaining         MsgID = "hosts_remaining"
	MsgIDHostTargetWentOffline  MsgID = "host_target_went_offline"
	MsgIDModSuccess             MsgID = "mod_success"
	MsgIDMsgBanned              MsgID = "msg_banned"
	MsgIDMsgCensoredBroadcaster MsgID = "msg_censored_broadcaster"
	MsgIDMsgChannelSuspended    MsgID = "msg_channel_suspended"
	MsgIDMsgDuplicate           MsgID = "msg_duplicate"
	MsgIDMsgEmoteonly           MsgID = "msg_emoteonly"
	MsgIDMsgRatelimit           MsgID = "msg_ratelimit"
	MsgIDMsgSubsonly            MsgID = "msg_subsonly"
	MsgIDMsgTimedout            MsgID = "msg_timedout"
	MsgIDMsgVerifiedEmail       MsgID = "msg_verified_email"
	MsgIDNoHelp                 MsgID = "no_help"
	MsgIDNoPermission           MsgID = "no_permission"
	MsgIDNotHosting             MsgID = "not_hosting"
	MsgIDTimeoutSuccess         MsgID = "timeout_success"
	MsgIDUnbanSuccess           MsgID = "unban_success"
	MsgIDUnmodSuccess           MsgID = "unmod_success"
	MsgIDUnrecognizedCmd        MsgID = "unrecognized_cmd"
	MsgIDUsageBan               MsgID = "usage_ban"
	MsgIDUsageClear             MsgID = "usage_clear"
	MsgIDUsageColor             MsgID = "usage_color"
	MsgIDUsageCommercial        MsgID = "usage_commercial"
	MsgIDUsageDisconnect        MsgID = "usage_disconnect"
	MsgIDUsageEmoteOnlyOn       MsgID = "usage_emote_only_on"
	MsgIDUsageEmoteOnlyOff      MsgID = "usage_emote_only_off"
	MsgIDUsageHelp              MsgID = "usage_help"
	MsgIDUsageHost              MsgID = "usage_host"
	MsgIDUsageMe                MsgID = "usage_me"
	MsgIDUsageMod               MsgID = "usage_mod"
	MsgIDUsageMods              MsgID = "usage_mods"
	MsgIDUsageR9kOn             MsgID = "usage_r9k_on"
	MsgIDUsageR9kOff            MsgID = "usage_r9k_off"
	MsgIDUsageSlowOn            MsgID = "usage_slow_on"
	MsgIDUsageSlowOff           MsgID = "usage_slow_off"
	MsgIDUsageSubsOn            MsgID = "usage_subs_on"
	MsgIDUsageSubsOff           MsgID = "usage_subs_off"
	MsgIDUsageTimeout           MsgID = "usage_timeout"
	MsgIDUsageUnban             MsgID = "usage_unban"
	MsgIDUsageUnhost            MsgID = "usage_unhost"
	MsgIDUsageUnmod             MsgID = "usage_unmod"
	MsgIDWhisperInvalidSelf     MsgID = "whisper_invalid_self"
	MsgIDWhisperLimitPerMin     MsgID = "whisper_limit_per_min"
	MsgIDWhisperLimitPerSec     MsgID = "whisper_limit_per_sec"
	MsgIDWhisperRestrictedRecip MsgID = "whisper_restricted_recipient"
)

// CommonUserstate contains fields common to all userstate types
// TypeScript has [paramater: string]: any - use Extra field for additional properties
type CommonUserstate struct {
	Badges       map[string]string   `json:"badges,omitempty"`
	BadgeInfo    map[string]string   `json:"badge-info,omitempty"`
	Color        string              `json:"color,omitempty"`
	DisplayName  string              `json:"display-name,omitempty"`
	Emotes       map[string][]string `json:"emotes,omitempty"`
	ID           string              `json:"id,omitempty"`
	Mod          bool                `json:"mod,omitempty"`
	Turbo        bool                `json:"turbo,omitempty"`
	EmotesRaw    string              `json:"emotes-raw,omitempty"`
	BadgesRaw    string              `json:"badges-raw,omitempty"`
	BadgeInfoRaw string              `json:"badge-info-raw,omitempty"`
	RoomID       string              `json:"room-id,omitempty"`
	Subscriber   bool                `json:"subscriber,omitempty"`
	UserType     string              `json:"user-type,omitempty"` // "", "mod", "global_mod", "admin", or "staff"
	UserID       string              `json:"user-id,omitempty"`
	TMISentTs    string              `json:"tmi-sent-ts,omitempty"`
	Flags        string              `json:"flags,omitempty"`
	MessageType  string              `json:"message-type,omitempty"`
	// Extra holds any additional fields from [parameter: string]: any
	Extra        map[string]any      `json:"-"`
}

// GetExtra retrieves a value from the Extra map with a default fallback.
// If the key doesn't exist or the type assertion fails, orElse is returned.
func GetExtra[T any](cu *CommonUserstate, key string, orElse T) T {
	if cu.Extra == nil {
		return orElse
	}

	val, ok := cu.Extra[key]
	if !ok {
		return orElse
	}

	result, ok := val.(T)
	if !ok {
		return orElse
	}

	return result
}

// DeleteUserstate contains information about a deleted message
type DeleteUserstate struct {
	Login       string `json:"login,omitempty"`
	Message     string `json:"message,omitempty"`
	TargetMsgID string `json:"target-msg-id,omitempty"`
}

// UserNoticeState extends CommonUserstate for user notices
type UserNoticeState struct {
	CommonUserstate
	Login     string `json:"login,omitempty"`
	Message   string `json:"message,omitempty"`
	SystemMsg string `json:"system-msg,omitempty"`
}

// CommonSubUserstate extends UserNoticeState for subscription events
type CommonSubUserstate struct {
	UserNoticeState
	MsgParamSubPlan     SubMethod `json:"msg-param-sub-plan,omitempty"`
	MsgParamSubPlanName string    `json:"msg-param-sub-plan-name,omitempty"`
}

// CommonGiftSubUserstate extends CommonSubUserstate for gift subscriptions
type CommonGiftSubUserstate struct {
	CommonSubUserstate
	MsgParamRecipientDisplayName string `json:"msg-param-recipient-display-name,omitempty"`
	MsgParamRecipientID          string `json:"msg-param-recipient-id,omitempty"`
	MsgParamRecipientUserName    string `json:"msg-param-recipient-user-name,omitempty"`
	MsgParamMonths               string `json:"msg-param-months,omitempty"`
}

// ChatUserstate represents userstate for chat messages
type ChatUserstate struct {
	CommonUserstate
	Username string `json:"username,omitempty"`
	Bits     string `json:"bits,omitempty"`
}

// SubUserstate represents userstate for subscription events
type SubUserstate struct {
	CommonSubUserstate
	MsgParamCumulativeMonths  string `json:"msg-param-cumulative-months,omitempty"`
	MsgParamShouldShareStreak bool   `json:"msg-param-should-share-streak,omitempty"`
	MsgParamStreakMonths      string `json:"msg-param-streak-months,omitempty"`
}

// SubMysteryGiftUserstate represents userstate for mystery gift subs
type SubMysteryGiftUserstate struct {
	CommonSubUserstate
	MsgParamSenderCount string `json:"msg-param-sender-count,omitempty"`
	MsgParamOriginID    string `json:"msg-param-origin-id"`
}

// SubGiftUserstate represents userstate for gifted subs
type SubGiftUserstate struct {
	CommonGiftSubUserstate
	MsgParamSenderCount string `json:"msg-param-sender-count,omitempty"`
	MsgParamOriginID    string `json:"msg-param-origin-id"`
}

// AnonSubGiftUserstate represents userstate for anonymous gifted subs
type AnonSubGiftUserstate struct {
	CommonGiftSubUserstate
}

// AnonSubMysteryGiftUserstate represents userstate for anonymous mystery gift subs
type AnonSubMysteryGiftUserstate struct {
	CommonSubUserstate
}

// SubGiftUpgradeUserstate represents userstate for gift subscription upgrades
type SubGiftUpgradeUserstate struct {
	CommonSubUserstate
	MsgParamSenderName  string `json:"msg-param-sender-name,omitempty"`
	MsgParamSenderLogin string `json:"msg-param-sender-login,omitempty"`
}

// AnonSubGiftUpgradeUserstate represents userstate for anonymous gift upgrades
type AnonSubGiftUpgradeUserstate struct {
	CommonSubUserstate
}

// PrimeUpgradeUserstate represents userstate for Prime upgrades
type PrimeUpgradeUserstate struct {
	CommonSubUserstate
}

// RaidUserstate represents userstate for raid events
type RaidUserstate struct {
	UserNoticeState
	MsgParamDisplayName string `json:"msg-param-displayName,omitempty"`
	MsgParamLogin       string `json:"msg-param-login,omitempty"`
	MsgParamViewerCount string `json:"msg-param-viewerCount,omitempty"`
}

// RitualUserstate represents userstate for ritual events
type RitualUserstate struct {
	UserNoticeState
	MsgParamRitualName string `json:"msg-param-ritual-name,omitempty"` // "new_chatter"
}

// BanUserstate represents userstate for ban events
type BanUserstate struct {
	RoomID       string `json:"room-id,omitempty"`
	TargetUserID string `json:"target-user-id,omitempty"`
	TMISentTs    string `json:"tmi-sent-ts,omitempty"`
}

// TimeoutUserstate represents userstate for timeout events
type TimeoutUserstate struct {
	BanUserstate
	BanDuration string `json:"ban-duration,omitempty"`
}

// RoomState represents the state of a Twitch channel
type RoomState struct {
	BroadcasterLang string `json:"broadcaster-lang,omitempty"`
	EmoteOnly       bool   `json:"emote-only,omitempty"`
	FollowersOnly   string `json:"followers-only,omitempty"` // "-1" for off, or number of minutes
	R9K             bool   `json:"r9k,omitempty"`
	Rituals         bool   `json:"rituals,omitempty"`
	RoomID          string `json:"room-id,omitempty"`
	Slow            string `json:"slow,omitempty"` // "0" for off, or number of seconds
	SubsOnly        bool   `json:"subs-only,omitempty"`
	Channel         string `json:"channel,omitempty"`
}

// GlobalUserState contains global user state information
type GlobalUserState struct {
	BadgeInfo    map[string]string `json:"badge-info,omitempty"`
	Badges       map[string]string `json:"badges,omitempty"`
	Color        string            `json:"color,omitempty"`
	DisplayName  string            `json:"display-name,omitempty"`
	EmoteSets    string            `json:"emote-sets,omitempty"`
	UserID       string            `json:"user-id,omitempty"`
	UserType     string            `json:"user-type,omitempty"`
	BadgeInfoRaw string            `json:"badge-info-raw,omitempty"`
	BadgesRaw    string            `json:"badges-raw,omitempty"`
}

// UserState contains user state for a channel (kept for backward compatibility)
type UserState struct {
	BadgeInfo    map[string]string `json:"badge-info,omitempty"`
	Badges       map[string]string `json:"badges,omitempty"`
	Color        string            `json:"color,omitempty"`
	DisplayName  string            `json:"display-name,omitempty"`
	EmoteSets    string            `json:"emote-sets,omitempty"`
	Mod          bool              `json:"mod,omitempty"`
	Subscriber   bool              `json:"subscriber,omitempty"`
	UserType     string            `json:"user-type,omitempty"`
	BadgeInfoRaw string            `json:"badge-info-raw,omitempty"`
	BadgesRaw    string            `json:"badges-raw,omitempty"`
	Username     string            `json:"username,omitempty"`
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
type OutgoingTags struct {
	ClientNonce      string            `json:"client-nonce,omitempty"`
	ReplyParentMsgID string            `json:"reply-parent-msg-id,omitempty"`
	// Extra allows for additional custom tags
	Extra            map[string]string `json:"-"`
}

// GetExtraTag retrieves a value from the Extra map with a default fallback.
// If the key doesn't exist, orElse is returned.
func GetExtraTag(tags *OutgoingTags, key string, orElse string) string {
	if tags == nil || tags.Extra == nil {
		return orElse
	}

	val, ok := tags.Extra[key]
	if !ok {
		return orElse
	}

	return val
}
