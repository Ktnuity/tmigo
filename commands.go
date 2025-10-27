package tmigo

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// Say sends a message to a channel
func (c *Client) Say(channel, message string, tags ...map[string]string) error {
	channel = Channel(channel)

	// Check for commands
	if (strings.HasPrefix(message, ".") && !strings.HasPrefix(message, "..")) || strings.HasPrefix(message, "/") || strings.HasPrefix(message, "\\") {
		// Check for /me command
		if strings.HasPrefix(message, ".me ") || strings.HasPrefix(message, "/me ") {
			return c.Action(channel, message[4:], tags...)
		}
		return c.sendCommand(channel, message, tags...)
	}

	return c.sendMessage(channel, message, tags...)
}

// Action sends an action message (/me) to a channel
func (c *Client) Action(channel, message string, tags ...map[string]string) error {
	message = fmt.Sprintf("\x01ACTION %s\x01", message)
	return c.sendMessage(channel, message, tags...)
}

// Join joins a channel
func (c *Client) Join(channel string) error {
	channel = Channel(channel)

	return c.sendCommandWithResponse(
		"",
		fmt.Sprintf("JOIN %s", channel),
		"_promiseJoin",
		c.getPromiseDelay(),
		nil,
	)
}

// JoinMultiple joins 1 or more channels
func (c *Client) JoinMultiple(channels []string) error {
	if len(channels) == 0 {
		return nil
	}

	channels = ChannelAll(channels)

	return c.sendCommandWithResponse(
		"",
		fmt.Sprintf("JOIN %s", strings.Join(channels, ",")),
		"_promiseJoin",
		c.getPromiseDelay(),
		nil,
	)
}

// Part leaves a channel
func (c *Client) Part(channel string) error {
	channel = Channel(channel)

	return c.sendCommandWithResponse(
		"",
		fmt.Sprintf("PART %s", channel),
		"_promisePart",
		c.getPromiseDelay(),
		nil,
	)
}

// Ban bans a user from a channel
func (c *Client) Ban(channel, username, reason string) error {
	username = Username(username)
	if reason == "" {
		reason = ""
	}
	return c.sendCommandWithResponse(
		channel,
		fmt.Sprintf("/ban %s %s", username, reason),
		"_promiseBan",
		c.getPromiseDelay(),
		nil,
	)
}

// Timeout times out a user in a channel
func (c *Client) Timeout(channel, username string, seconds int, reason string) error {
	username = Username(username)
	if seconds == 0 {
		seconds = 300
	}
	if reason == "" {
		reason = ""
	}
	return c.sendCommandWithResponse(
		channel,
		fmt.Sprintf("/timeout %s %d %s", username, seconds, reason),
		"_promiseTimeout",
		c.getPromiseDelay(),
		nil,
	)
}

// Unban unbans a user from a channel
func (c *Client) Unban(channel, username string) error {
	username = Username(username)
	return c.sendCommandWithResponse(
		channel,
		fmt.Sprintf("/unban %s", username),
		"_promiseUnban",
		c.getPromiseDelay(),
		nil,
	)
}

// Clear clears chat in a channel
func (c *Client) Clear(channel string) error {
	return c.sendCommandWithResponse(
		channel,
		"/clear",
		"_promiseClear",
		c.getPromiseDelay(),
		nil,
	)
}

// Color changes the client's username color
func (c *Client) Color(newColor string) error {
	return c.sendCommandWithResponse(
		c.state.globalDefaultChannel,
		fmt.Sprintf("/color %s", newColor),
		"_promiseColor",
		c.getPromiseDelay(),
		nil,
	)
}

// Commercial runs a commercial on a channel
func (c *Client) Commercial(channel string, seconds int) error {
	if seconds == 0 {
		seconds = 30
	}
	return c.sendCommandWithResponse(
		channel,
		fmt.Sprintf("/commercial %d", seconds),
		"_promiseCommercial",
		c.getPromiseDelay(),
		nil,
	)
}

// DeleteMessage deletes a specific message
func (c *Client) DeleteMessage(channel, messageUUID string) error {
	return c.sendCommandWithResponse(
		channel,
		fmt.Sprintf("/delete %s", messageUUID),
		"_promiseDeletemessage",
		c.getPromiseDelay(),
		nil,
	)
}

// EmoteOnly enables emote-only mode in a channel
func (c *Client) EmoteOnly(channel string) error {
	return c.sendCommandWithResponse(
		channel,
		"/emoteonly",
		"_promiseEmoteonly",
		c.getPromiseDelay(),
		nil,
	)
}

// EmoteOnlyOff disables emote-only mode in a channel
func (c *Client) EmoteOnlyOff(channel string) error {
	return c.sendCommandWithResponse(
		channel,
		"/emoteonlyoff",
		"_promiseEmoteonlyoff",
		c.getPromiseDelay(),
		nil,
	)
}

// FollowersOnly enables followers-only mode in a channel
func (c *Client) FollowersOnly(channel string, minutes int) error {
	if minutes == 0 {
		minutes = 30
	}
	return c.sendCommandWithResponse(
		channel,
		fmt.Sprintf("/followers %d", minutes),
		"_promiseFollowers",
		c.getPromiseDelay(),
		nil,
	)
}

// FollowersOnlyOff disables followers-only mode in a channel
func (c *Client) FollowersOnlyOff(channel string) error {
	return c.sendCommandWithResponse(
		channel,
		"/followersoff",
		"_promiseFollowersoff",
		c.getPromiseDelay(),
		nil,
	)
}

// Host hosts another channel
func (c *Client) Host(channel, target string) error {
	target = Username(target)
	return c.sendCommandWithResponse(
		channel,
		fmt.Sprintf("/host %s", target),
		"_promiseHost",
		2*time.Second,
		nil,
	)
}

// Unhost stops hosting
func (c *Client) Unhost(channel string) error {
	return c.sendCommandWithResponse(
		channel,
		"/unhost",
		"_promiseUnhost",
		2*time.Second,
		nil,
	)
}

// Mod gives mod status to a user
func (c *Client) Mod(channel, username string) error {
	username = Username(username)
	return c.sendCommandWithResponse(
		channel,
		fmt.Sprintf("/mod %s", username),
		"_promiseMod",
		c.getPromiseDelay(),
		nil,
	)
}

// Unmod removes mod status from a user
func (c *Client) Unmod(channel, username string) error {
	username = Username(username)
	return c.sendCommandWithResponse(
		channel,
		fmt.Sprintf("/unmod %s", username),
		"_promiseUnmod",
		c.getPromiseDelay(),
		nil,
	)
}

// Mods gets the list of moderators in a channel
func (c *Client) Mods(channel string) error {
	channel = Channel(channel)
	return c.sendCommandWithResponse(
		channel,
		"/mods",
		"_promiseMods",
		c.getPromiseDelay(),
		nil,
	)
}

// VIP gives VIP status to a user
func (c *Client) VIP(channel, username string) error {
	username = Username(username)
	return c.sendCommandWithResponse(
		channel,
		fmt.Sprintf("/vip %s", username),
		"_promiseVip",
		c.getPromiseDelay(),
		nil,
	)
}

// Unvip removes VIP status from a user
func (c *Client) Unvip(channel, username string) error {
	username = Username(username)
	return c.sendCommandWithResponse(
		channel,
		fmt.Sprintf("/unvip %s", username),
		"_promiseUnvip",
		c.getPromiseDelay(),
		nil,
	)
}

// VIPs gets the list of VIPs in a channel
func (c *Client) VIPs(channel string) error {
	return c.sendCommandWithResponse(
		channel,
		"/vips",
		"_promiseVips",
		c.getPromiseDelay(),
		nil,
	)
}

// R9KBeta enables R9K mode in a channel
func (c *Client) R9KBeta(channel string) error {
	return c.sendCommandWithResponse(
		channel,
		"/r9kbeta",
		"_promiseR9kbeta",
		c.getPromiseDelay(),
		nil,
	)
}

// R9KBetaOff disables R9K mode in a channel
func (c *Client) R9KBetaOff(channel string) error {
	return c.sendCommandWithResponse(
		channel,
		"/r9kbetaoff",
		"_promiseR9kbetaoff",
		c.getPromiseDelay(),
		nil,
	)
}

// Slow enables slow mode in a channel
func (c *Client) Slow(channel string, seconds int) error {
	if seconds == 0 {
		seconds = 300
	}
	return c.sendCommandWithResponse(
		channel,
		fmt.Sprintf("/slow %d", seconds),
		"_promiseSlow",
		c.getPromiseDelay(),
		nil,
	)
}

// SlowOff disables slow mode in a channel
func (c *Client) SlowOff(channel string) error {
	return c.sendCommandWithResponse(
		channel,
		"/slowoff",
		"_promiseSlowoff",
		c.getPromiseDelay(),
		nil,
	)
}

// Subscribers enables subscribers-only mode in a channel
func (c *Client) Subscribers(channel string) error {
	return c.sendCommandWithResponse(
		channel,
		"/subscribers",
		"_promiseSubscribers",
		c.getPromiseDelay(),
		nil,
	)
}

// SubscribersOff disables subscribers-only mode in a channel
func (c *Client) SubscribersOff(channel string) error {
	return c.sendCommandWithResponse(
		channel,
		"/subscribersoff",
		"_promiseSubscribersoff",
		c.getPromiseDelay(),
		nil,
	)
}

// Whisper sends a whisper to a user
func (c *Client) Whisper(username, message string) error {
	username = Username(username)

	if username == c.GetUsername() {
		return errors.New("cannot send a whisper to the same account")
	}

	return c.sendCommandWithResponse(
		c.state.globalDefaultChannel,
		fmt.Sprintf("/w %s %s", username, message),
		"_promiseWhisper",
		c.getPromiseDelay(),
		nil,
	)
}

// Ping sends a ping to the server
func (c *Client) Ping() error {
	c.state.latency = time.Now()
	c.state.pingTimeout = time.AfterFunc(c.state.opts.Connection.Timeout, func() {
		if c.state.ws != nil {
			c.state.wasCloseCalled = false
			c.state.log.Error("Ping timeout.")
			c.state.ws.Close()
			if c.state.pingLoop != nil {
				c.state.pingLoop.Stop()
			}
			if c.state.pingTimeout != nil {
				c.state.pingTimeout.Stop()
			}
		}
	})

	return c.sendCommandRaw("PING", nil)
}

// Raw sends a raw IRC command
func (c *Client) Raw(command string, tags ...map[string]string) error {
	return c.sendCommandRaw(command, tags...)
}

// Announce announces a message in a channel
func (c *Client) Announce(channel, message string) error {
	return c.sendMessage(channel, fmt.Sprintf("/announce %s", message))
}

// Reply sends a message as a reply to another message
func (c *Client) Reply(channel, message, replyParentMsgID string, tags ...map[string]string) error {
	var tagMap map[string]string
	if len(tags) > 0 {
		tagMap = tags[0]
	} else {
		tagMap = make(map[string]string)
	}

	if replyParentMsgID == "" {
		return errors.New("replyParentMsgId is required")
	}

	tagMap["reply-parent-msg-id"] = replyParentMsgID
	return c.Say(channel, message, tagMap)
}

// sendMessage sends a message to a channel
func (c *Client) sendMessage(channel, message string, tags ...map[string]string) error {
	if !c.isConnected() {
		return errors.New("not connected to server")
	}

	if IsJustinfan(c.GetUsername()) {
		return errors.New("cannot send anonymous messages")
	}

	channel = Channel(channel)

	// Split long messages
	if len(message) > 500 {
		firstPart := message[:500]
		lastSpace := strings.LastIndex(firstPart, " ")
		if lastSpace == -1 {
			lastSpace = 500
		}

		err := c.sendMessage(channel, message[:lastSpace], tags...)
		if err != nil {
			return err
		}

		time.Sleep(350 * time.Millisecond)
		return c.sendMessage(channel, message[lastSpace:], tags...)
	}

	tagStr := ""
	if len(tags) > 0 && tags[0] != nil {
		tagStr = FormTags(tags[0])
		if tagStr != "" {
			tagStr += " "
		}
	}

	return c.state.ws.WriteMessage(1, fmt.Appendf(nil, "%sPRIVMSG %s :%s", tagStr, channel, message))
}

// sendCommand sends a command to a channel
func (c *Client) sendCommand(channel, command string, tags ...map[string]string) error {
	if !c.isConnected() {
		return errors.New("not connected to server")
	}

	channel = Channel(channel)

	tagStr := ""
	if len(tags) > 0 && tags[0] != nil {
		tagStr = FormTags(tags[0])
		if tagStr != "" {
			tagStr += " "
		}
	}

	if channel != "" {
		c.state.log.Info(fmt.Sprintf("[%s] Executing command: %s", channel, command))
		return c.state.ws.WriteMessage(1, fmt.Appendf(nil, "%sPRIVMSG %s :%s", tagStr, channel, command))
	} else {
		c.state.log.Info(fmt.Sprintf("Executing command: %s", command))
		return c.state.ws.WriteMessage(1, fmt.Appendf(nil, "%s%s", tagStr, command))
	}

}

// sendCommandRaw sends a raw command
func (c *Client) sendCommandRaw(command string, tags ...map[string]string) error {
	if !c.isConnected() {
		return errors.New("not connected to server")
	}

	tagStr := ""
	if len(tags) > 0 && tags[0] != nil {
		tagStr = FormTags(tags[0])
		if tagStr != "" {
			tagStr += " "
		}
	}

	c.state.log.Info(fmt.Sprintf("Executing command: %s", command))
	return c.state.ws.WriteMessage(1, fmt.Appendf(nil, "%s%s", tagStr, command))
}

// sendCommandWithResponse sends a command and waits for a response event
func (c *Client) sendCommandWithResponse(channel, command, responseEvent string, timeout time.Duration, tags ...map[string]string) error {
	if !c.isConnected() {
		return errors.New("not connected to server")
	}

	// Send the command
	if channel != "" {
		err := c.sendCommand(channel, command, tags...)
		if err != nil {
			return err
		}
	} else {
		err := c.sendCommandRaw(command, tags...)
		if err != nil {
			return err
		}
	}

	// In a real implementation, we'd wait for the response event
	// For now, we'll just return nil
	// The event will be emitted when the server responds
	return nil
}

// getPromiseDelay returns the promise delay based on latency
func (c *Client) getPromiseDelay() time.Duration {
	minDelay := 600 * time.Millisecond
	latencyDelay := c.state.currentLatency + 100*time.Millisecond
	return max(latencyDelay, minDelay)
}

// Aliases
func (c *Client) FollowersMode(channel string, minutes int) error {
	return c.FollowersOnly(channel, minutes)
}

func (c *Client) FollowersModeOff(channel string) error {
	return c.FollowersOnlyOff(channel)
}

func (c *Client) Leave(channel string) error {
	return c.Part(channel)
}

func (c *Client) SlowMode(channel string, seconds int) error {
	return c.Slow(channel, seconds)
}

func (c *Client) SlowModeOff(channel string) error {
	return c.SlowOff(channel)
}

func (c *Client) R9KMode(channel string) error {
	return c.R9KBeta(channel)
}

func (c *Client) R9KModeOff(channel string) error {
	return c.R9KBetaOff(channel)
}

func (c *Client) UniqueChat(channel string) error {
	return c.R9KBeta(channel)
}

func (c *Client) UniqueChatOff(channel string) error {
	return c.R9KBetaOff(channel)
}
