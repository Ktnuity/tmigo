# tmigo

A Go port of [tmi.js](https://github.com/tmijs/tmi.js) - A Twitch Messaging Interface library for Go.

#### Twitch deprecations
- Twitch removed `/`-chat commands other than `/me` (action) through the IRC connection on February 18, 2023. [See the announcement](https://discuss.dev.twitch.tv/t/deprecation-of-chat-commands-through-irc/40486). This removed a lot of functionality which affetcs tmi.js and in turn tmigo. I have kept these functions in this port just to be 1:1 with the JS version. But be aware that just because they are there, they don't actually bypass twitch's *new* IRC restrictions.

## Installation

```bash
go get github.com/ktnuity/tmigo
```

## Usage

### Basic Example

```go
package main

import (
    "log"
    "github.com/ktnuity/tmigo"
)

func main() {
    // Create a new client
    client := tmigo.NewClient(&tmigo.ClientOptions{
        Options: &tmigo.Options{
            Debug: true,
        },
        Identity: &tmigo.Identity{
            Username: "your_bot_name",
            Password: "oauth:your_oauth_token",
        },
        Channels: []string{
            "channel1",
            "channel2",
        },
    })

    // Register event handlers
    client.On("message", func(args ...any) {
        channel := args[0].(string)
        tags := args[1].(map[string]any)
        message := args[2].(string)
        self := args[3].(bool)

        if self {
            return // Don't respond to own messages
        }

        username := tags["username"].(string)
        log.Printf("[%s] %s: %s", channel, username, message)

        // Simple command example
        if message == "!hello" {
            client.Say(channel, "@" + username + ", hello!")
        }
    })

    client.On("connected", func(args ...any) {
        log.Println("Connected to Twitch!")
    })

    // Connect
    if err := client.Connect(); err != nil {
        log.Fatal(err)
    }

    // Keep the program running
    select {}
}
```

## Available Methods

### Connection
- `Connect()` - Connect to Twitch IRC
- `Disconnect()` - Disconnect from Twitch IRC
- `GetUsername()` - Get the current username
- `GetChannels()` - Get list of joined channels
- `ReadyState()` - Get connection state

### Channel Management
- `Join(channel)` - Join a channel
- `Part(channel)` / `Leave(channel)` - Leave a channel

### Sending Messages
- `Say(channel, message)` - Send a message to a channel
- `Action(channel, message)` - Send an action message (/me)
- `Reply(channel, message, replyParentMsgID)` - Reply to a message
- `Whisper(username, message)` - Send a whisper
- `Announce(channel, message)` - Send an announcement

### Moderation
- `Ban(channel, username, reason)` - Ban a user
- `Unban(channel, username)` - Unban a user
- `Timeout(channel, username, seconds, reason)` - Timeout a user
- `Clear(channel)` - Clear chat
- `DeleteMessage(channel, messageUUID)` - Delete a specific message

### Channel Modes
- `Slow(channel, seconds)` / `SlowOff(channel)` - Slow mode
- `FollowersOnly(channel, minutes)` / `FollowersOnlyOff(channel)` - Followers-only mode
- `Subscribers(channel)` / `SubscribersOff(channel)` - Subscribers-only mode
- `EmoteOnly(channel)` / `EmoteOnlyOff(channel)` - Emote-only mode
- `R9KBeta(channel)` / `R9KBetaOff(channel)` - R9K mode

### User Management
- `Mod(channel, username)` / `Unmod(channel, username)` - Mod/unmod a user
- `VIP(channel, username)` / `Unvip(channel, username)` - VIP/unvip a user
- `Mods(channel)` - Get list of moderators
- `VIPs(channel)` - Get list of VIPs
- `IsMod(channel, username)` - Check if user is a mod

### Other
- `Host(channel, target)` / `Unhost(channel)` - Host/unhost
- `Commercial(channel, seconds)` - Run a commercial
- `Color(newColor)` - Change username color
- `Ping()` - Ping the server
- `Raw(command)` - Send a raw IRC command

## Events

### Connection Events
- `connecting` - Fired when connecting to server
- `connected` - Fired when successfully connected
- `disconnected` - Fired when disconnected
- `reconnect` - Fired when attempting to reconnect
- `logon` - Fired when sending authentication

### Message Events
- `message` - All messages (chat, action, whisper)
- `chat` - Regular chat messages
- `action` - Action messages (/me)
- `whisper` - Whisper messages

### Channel Events
- `join` - User joined a channel
- `part` - User left a channel
- `names` - List of users in channel
- `roomstate` - Room state changed

### Subscription Events
- `subscription` / `sub` - New subscription
- `resub` / `subanniversary` - Resubscription
- `subgift` - Gifted subscription
- `submysterygift` - Mystery gift subscription

### Other Events
- `cheer` - Bits cheered
- `raided` - Channel raided
- `hosted` - Channel hosted
- `hosting` - Now hosting another channel
- `ban` - User banned
- `timeout` - User timed out
- `clearchat` - Chat cleared
- `messagedeleted` - Message deleted
- `emotesets` - Emote sets changed
- `notice` - Notice from Twitch
- `raw_message` - Raw IRC message

## Configuration Options

### Options
```go
Options: &tmigo.Options{
    Debug: bool,                    // Enable debug logging
    GlobalDefaultChannel: string,   // Default channel for global commands
    SkipMembership: bool,           // Skip JOIN/PART events
    JoinInterval: int,              // Delay between joins (ms)
    MessagesLogLevel: string,       // Log level for messages
}
```

### Connection
```go
Connection: &tmigo.Connection{
    Server: string,                 // IRC server
    Port: int,                      // IRC port
    Secure: bool,                   // Use secure connection (WSS)
    Reconnect: bool,                // Auto-reconnect
    ReconnectInterval: time.Duration,
    ReconnectDecay: float64,
    MaxReconnectInterval: time.Duration,
    MaxReconnectAttempts: int,
    Timeout: time.Duration,
}
```

### Identity
```go
Identity: &tmigo.Identity{
    Username: string,  // Bot username
    Password: string,  // OAuth token (oauth:xxx or just xxx)
}
```

## Getting an OAuth Token

Get your OAuth token from: https://twitchapps.com/tmi/

## Differences from tmi.js

While this library aims to maintain API compatibility with tmi.js, there are some differences due to Go's nature:

1. **Events**: Instead of Promise-based methods, most operations are fire-and-forget. Use event listeners for responses.
2. **Callbacks**: Event handlers use `func(args ...any)` instead of typed callbacks.
3. **Concurrency**: The library is designed to be thread-safe with proper mutex usage.
4. **Error Handling**: Methods return errors in idiomatic Go style.

## Example

See [example/main.go](example/main.go) for a complete working example.

## License

MIT

## Credits

This is a Go port of [tmi.js](https://github.com/tmijs/tmi.js) by Schmoopiie.

This project was ported using **Claude Sonnet 4.5** (claude-sonnet-4-5-20250929).
