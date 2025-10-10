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

func main() {
	// Create a new client with options
	client := tmigo.NewClient(&tmigo.ClientOptions{
		Options: &tmigo.Options{
			Debug: true, // Enable debug logging
		},
		Identity: &tmigo.Identity{
			Username: "your_bot_name",
			Password: "oauth:your_bot_token", // See README for how to obtain OAuth token
		},
		Channels: []string{
			"your_channel",
		},
	})

	// Register event handlers
	setupEventHandlers(client)

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

func setupEventHandlers(client *tmigo.Client) {
	// Connected event
	client.On("connected", func(args ...any) {
		server := args[0].(string)
		port := args[1].(int)
		log.Printf("Connected to %s:%d", server, port)
	})

	// Disconnected event
	client.On("disconnected", func(args ...any) {
		reason := args[0].(string)
		log.Printf("Disconnected: %s", reason)
	})

	// Join event
	client.On("join", func(args ...any) {
		channel := args[0].(string)
		username := args[1].(string)
		self := args[2].(bool)

		if self {
			log.Printf("Joined channel: %s", channel)
		} else {
			log.Printf("User %s joined %s", username, channel)
		}
	})

	// Message event
	// Note: tags is map[string]any, but for typed access you can convert to ChatUserstate
	client.On("message", func(args ...any) {
		channel := args[0].(string)
		tags := args[1].(map[string]any)
		message := args[2].(string)
		self := args[3].(bool)

		// Don't respond to own messages
		if self {
			return
		}

		username := ""
		if val, ok := tags["username"].(string); ok {
			username = val
		}

		// You can also check for specific fields from ChatUserstate:
		// - tags["display-name"] for display name
		// - tags["color"] for username color
		// - tags["badges"] for user badges (map[string]string)
		// - tags["subscriber"] for subscriber status (bool)
		// - tags["mod"] for moderator status (bool)
		// - tags["bits"] for bit amount in cheers (string)

		log.Printf("[%s] %s: %s", channel, username, message)

		// Simple command handler
		handleCommands(client, channel, username, message)
	})

	// Chat event (regular messages only)
	client.On("chat", func(args ...any) {
		channel := args[0].(string)
		tags := args[1].(map[string]any)
		message := args[2].(string)
		self := args[3].(bool)

		if self {
			return
		}

		username := ""
		if val, ok := tags["username"].(string); ok {
			username = val
		}

		log.Printf("[CHAT] [%s] %s: %s", channel, username, message)
	})

	// Action event (/me messages)
	client.On("action", func(args ...any) {
		channel := args[0].(string)
		tags := args[1].(map[string]any)
		message := args[2].(string)

		username := ""
		if val, ok := tags["username"].(string); ok {
			username = val
		}

		log.Printf("[ACTION] [%s] * %s %s", channel, username, message)
	})

	// Subscription events
	client.On("subscription", func(args ...any) {
		channel := args[0].(string)
		username := args[1].(string)
		log.Printf("[SUB] %s subscribed to %s!", username, channel)

		// Thank the subscriber
		client.Say(channel, fmt.Sprintf("Thank you for subscribing, @%s!", username))
	})

	client.On("resub", func(args ...any) {
		channel := args[0].(string)
		username := args[1].(string)
		months := args[2].(int)
		log.Printf("[RESUB] %s resubscribed to %s for %d months!", username, channel, months)

		client.Say(channel, fmt.Sprintf("Thank you for %d months, @%s!", months, username))
	})

	// Raid event
	client.On("raided", func(args ...any) {
		channel := args[0].(string)
		username := args[1].(string)
		viewers := args[2].(int)
		log.Printf("[RAID] %s raided %s with %d viewers!", username, channel, viewers)

		client.Say(channel, fmt.Sprintf("Welcome raiders from @%s!", username))
	})

	// Cheer event (bits)
	client.On("cheer", func(args ...any) {
		channel := args[0].(string)
		tags := args[1].(map[string]any)
		message := args[2].(string)

		username := ""
		if val, ok := tags["username"].(string); ok {
			username = val
		}

		bits := ""
		if val, ok := tags["bits"].(string); ok {
			bits = val
		}

		log.Printf("[CHEER] %s cheered %s bits in %s: %s", username, bits, channel, message)
	})

	// Host events
	client.On("hosting", func(args ...any) {
		channel := args[0].(string)
		target := args[1].(string)
		viewers := args[2].(int)
		log.Printf("[HOST] %s is now hosting %s for %d viewers", channel, target, viewers)
	})

	// Ban/timeout events
	client.On("ban", func(args ...any) {
		channel := args[0].(string)
		username := args[1].(string)
		log.Printf("[BAN] %s was banned from %s", username, channel)
	})

	client.On("timeout", func(args ...any) {
		channel := args[0].(string)
		username := args[1].(string)
		duration := args[3].(int)
		log.Printf("[TIMEOUT] %s was timed out in %s for %d seconds", username, channel, duration)
	})
}

func handleCommands(client *tmigo.Client, channel, username, message string) {
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
		client.Say(channel, fmt.Sprintf("@%s, hello!", username))

	case "!ping":
		client.Say(channel, "Pong!")

	case "!commands":
		client.Say(channel, "Available commands: !hello, !ping, !uptime, !commands")

	case "!uptime":
		client.Say(channel, "I don't track uptime yet, but I'm here!")

	case "!dice":
		// Example of using action
		client.Action(channel, "rolls a dice... 🎲")

	default:
		// Unknown command, do nothing
	}
}
