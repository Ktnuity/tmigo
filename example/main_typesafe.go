package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/ktnuity/tmigo"
)

// This example demonstrates the NEW type-safe API
// Compare this to main.go to see the difference!

func mainTypeSafe() {
	// Create a new client with options
	client := tmigo.NewClient(&tmigo.ClientOptions{
		Options: &tmigo.Options{
			Debug: true, // Enable debug logging
		},
		Identity: &tmigo.Identity{
			Username: "your_bot_name",
			Password: "oauth:your_bot_token",
		},
		Channels: []string{
			"your_channel",
		},
	})

	// Register event handlers using type-safe methods
	setupTypeSafeHandlers(client)

	// Connect to Twitch
	log.Println("Connecting to Twitch...")
	if err := client.Connect(); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down...")
	client.Disconnect()
}

func setupTypeSafeHandlers(client *tmigo.Client) {
	// Connected event - type-safe!
	client.OnConnected(func(server string, port int) {
		log.Printf("Connected to %s:%d", server, port)
	})

	// Disconnected event - type-safe!
	client.OnDisconnected(func(reason string) {
		log.Printf("Disconnected: %s", reason)
	})

	// Join event - type-safe!
	client.OnJoin(func(channel string, username string, self bool) {
		if self {
			log.Printf("Joined channel: %s", channel)
		} else {
			log.Printf("User %s joined %s", username, channel)
		}
	})

	// Message event - type-safe with ChatUserstate!
	// NO MORE type assertions needed!
	client.OnMessage(func(channel string, userstate tmigo.ChatUserstate, message string, self bool) {
		// Don't respond to own messages
		if self {
			return
		}

		// Direct access to typed fields - no more type assertions!
		log.Printf("[%s] %s: %s", channel, userstate.Username, message)

		// Access other fields directly with type safety:
		if userstate.Mod {
			log.Printf("  -> %s is a moderator", userstate.Username)
		}
		if userstate.Subscriber {
			log.Printf("  -> %s is a subscriber", userstate.Username)
		}
		if userstate.Bits != "" {
			log.Printf("  -> %s cheered %s bits", userstate.Username, userstate.Bits)
		}
		if userstate.Color != "" {
			log.Printf("  -> Username color: %s", userstate.Color)
		}

		// Simple command handler
		handleTypeSafeCommands(client, channel, userstate, message)
	})

	// Chat event (regular messages only) - type-safe!
	client.OnChat(func(channel string, userstate tmigo.ChatUserstate, message string, self bool) {
		if !self {
			log.Printf("[CHAT] [%s] %s: %s", channel, userstate.Username, message)
		}
	})

	// Action event (/me messages) - type-safe!
	client.OnAction(func(channel string, userstate tmigo.ChatUserstate, message string, self bool) {
		log.Printf("[ACTION] [%s] * %s %s", channel, userstate.Username, message)
	})

	// Subscription events - type-safe with SubUserstate!
	client.OnSubscription(func(channel string, username string, methods tmigo.SubMethods, message string, userstate tmigo.SubUserstate) {
		log.Printf("[SUB] %s subscribed to %s with plan: %s!", username, channel, methods.Plan)

		// Access typed userstate fields
		if userstate.DisplayName != "" {
			log.Printf("  -> Display name: %s", userstate.DisplayName)
		}

		client.Say(channel, fmt.Sprintf("Thank you for subscribing, @%s!", username))
	})

	// Resub event - type-safe!
	client.OnResub(func(channel string, username string, months int, message string, userstate tmigo.SubUserstate, methods tmigo.SubMethods) {
		log.Printf("[RESUB] %s resubscribed to %s for %d months!", username, channel, months)

		// Access cumulative months and streak
		if userstate.MsgParamCumulativeMonths != "" {
			log.Printf("  -> Total months: %s", userstate.MsgParamCumulativeMonths)
		}
		if userstate.MsgParamStreakMonths != "" {
			log.Printf("  -> Streak: %s months", userstate.MsgParamStreakMonths)
		}

		client.Say(channel, fmt.Sprintf("Thank you for %d months, @%s!", months, username))
	})

	// Gift sub event - type-safe with SubGiftUserstate!
	client.OnSubGift(func(channel string, username string, streakMonths int, recipient string, methods tmigo.SubMethods, userstate tmigo.SubGiftUserstate) {
		log.Printf("[SUBGIFT] %s gifted a sub to %s in %s!", username, recipient, channel)
		client.Say(channel, fmt.Sprintf("Thank you @%s for gifting a sub to @%s!", username, recipient))
	})

	// Raid event - type-safe!
	client.OnRaided(func(channel string, username string, viewers int) {
		log.Printf("[RAID] %s raided %s with %d viewers!", username, channel, viewers)
		client.Say(channel, fmt.Sprintf("Welcome raiders from @%s! 🎉", username))
	})

	// Cheer event (bits) - type-safe with ChatUserstate!
	client.OnCheer(func(channel string, userstate tmigo.ChatUserstate, message string) {
		log.Printf("[CHEER] %s cheered %s bits in %s: %s", userstate.Username, userstate.Bits, channel, message)
		client.Say(channel, fmt.Sprintf("Thank you for the %s bits, @%s!", userstate.Bits, userstate.Username))
	})

	// Host events - type-safe!
	client.OnHosting(func(channel string, target string, viewers int) {
		log.Printf("[HOST] %s is now hosting %s for %d viewers", channel, target, viewers)
	})

	client.OnHosted(func(channel string, username string, viewers int, autohost bool) {
		hostType := "hosting"
		if autohost {
			hostType = "auto-hosting"
		}
		log.Printf("[HOSTED] %s is %s with %d viewers!", username, hostType, viewers)
		client.Say(channel, fmt.Sprintf("Thank you @%s for the host!", username))
	})

	// Ban/timeout events - type-safe with BanUserstate/TimeoutUserstate!
	client.OnBan(func(channel string, username string, reason string, userstate tmigo.BanUserstate) {
		log.Printf("[BAN] %s was banned from %s", username, channel)
		// Access typed ban metadata
		if userstate.RoomID != "" {
			log.Printf("  -> Room ID: %s", userstate.RoomID)
		}
	})

	client.OnTimeout(func(channel string, username string, reason string, duration int, userstate tmigo.TimeoutUserstate) {
		log.Printf("[TIMEOUT] %s was timed out in %s for %d seconds", username, channel, duration)
		// Access timeout-specific metadata
		if userstate.BanDuration != "" {
			log.Printf("  -> Duration from tags: %s", userstate.BanDuration)
		}
	})

	// Room state events - type-safe with RoomState!
	client.OnRoomstate(func(channel string, state tmigo.RoomState) {
		log.Printf("[ROOMSTATE] %s room state changed", channel)
		if state.EmoteOnly {
			log.Printf("  -> Emote-only mode is ON")
		}
		if state.SubsOnly {
			log.Printf("  -> Subscribers-only mode is ON")
		}
		if state.Slow != "" && state.Slow != "0" {
			log.Printf("  -> Slow mode: %s seconds", state.Slow)
		}
	})

	// Slowmode event - type-safe!
	client.OnSlowmode(func(channel string, enabled bool, length int) {
		if enabled {
			log.Printf("[SLOWMODE] Slow mode enabled in %s: %d seconds", channel, length)
		} else {
			log.Printf("[SLOWMODE] Slow mode disabled in %s", channel)
		}
	})
}

func handleTypeSafeCommands(client *tmigo.Client, channel string, userstate tmigo.ChatUserstate, message string) {
	// Simple command parsing
	if !strings.HasPrefix(message, "!") {
		return
	}

	parts := strings.Fields(message)
	if len(parts) == 0 {
		return
	}

	command := strings.ToLower(parts[0])

	switch command {
	case "!hello":
		// Use DisplayName if available, otherwise Username
		displayName := userstate.DisplayName
		if displayName == "" {
			displayName = userstate.Username
		}
		client.Say(channel, fmt.Sprintf("@%s, hello!", displayName))

	case "!ping":
		client.Say(channel, "Pong!")

	case "!commands":
		client.Say(channel, "Available commands: !hello, !ping, !whoami, !commands")

	case "!whoami":
		info := fmt.Sprintf("You are: %s", userstate.Username)
		if userstate.DisplayName != "" && userstate.DisplayName != userstate.Username {
			info += fmt.Sprintf(" (%s)", userstate.DisplayName)
		}

		var badges []string
		if userstate.Mod {
			badges = append(badges, "Moderator")
		}
		if userstate.Subscriber {
			badges = append(badges, "Subscriber")
		}
		if userstate.Turbo {
			badges = append(badges, "Turbo")
		}

		if len(badges) > 0 {
			info += fmt.Sprintf(" | Badges: %s", strings.Join(badges, ", "))
		}

		client.Say(channel, info)

	case "!dice":
		// Example of using action
		client.Action(channel, "rolls a dice... 🎲")

	default:
		// Unknown command, do nothing
	}
}
