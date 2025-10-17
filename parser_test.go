package tmigo

import (
	"reflect"
	"testing"
)

func TestParseMessage(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *IRCMessage
		wantNil bool
	}{
		{
			name:  "Simple PING",
			input: "PING :tmi.twitch.tv",
			want: &IRCMessage{
				Raw:     "PING :tmi.twitch.tv",
				Tags:    map[string]any{},
				Prefix:  "",
				Command: "PING",
				Params:  []string{"tmi.twitch.tv"},
			},
		},
		{
			name:  "Message with prefix",
			input: ":tmi.twitch.tv 001 username :Welcome!",
			want: &IRCMessage{
				Raw:     ":tmi.twitch.tv 001 username :Welcome!",
				Tags:    map[string]any{},
				Prefix:  "tmi.twitch.tv",
				Command: "001",
				Params:  []string{"username", "Welcome!"},
			},
		},
		{
			name:  "PRIVMSG",
			input: ":user!user@user.tmi.twitch.tv PRIVMSG #channel :Hello World",
			want: &IRCMessage{
				Raw:     ":user!user@user.tmi.twitch.tv PRIVMSG #channel :Hello World",
				Tags:    map[string]any{},
				Prefix:  "user!user@user.tmi.twitch.tv",
				Command: "PRIVMSG",
				Params:  []string{"#channel", "Hello World"},
			},
		},
		{
			name:  "Message with tags",
			input: "@badge-info=;badges=broadcaster/1;color=#FF0000 :user!user@user.tmi.twitch.tv PRIVMSG #channel :test",
			want: &IRCMessage{
				Raw:     "@badge-info=;badges=broadcaster/1;color=#FF0000 :user!user@user.tmi.twitch.tv PRIVMSG #channel :test",
				Tags:    map[string]any{"badge-info": true, "badges": "broadcaster/1", "color": "#FF0000"},
				Prefix:  "user!user@user.tmi.twitch.tv",
				Command: "PRIVMSG",
				Params:  []string{"#channel", "test"},
			},
		},
		{
			name:  "Tag with no value",
			input: "@flag :server COMMAND",
			want: &IRCMessage{
				Raw:     "@flag :server COMMAND",
				Tags:    map[string]any{"flag": true},
				Prefix:  "server",
				Command: "COMMAND",
				Params:  []string{},
			},
		},
		{
			name:  "Multiple params",
			input: ":server COMMAND param1 param2 param3",
			want: &IRCMessage{
				Raw:     ":server COMMAND param1 param2 param3",
				Tags:    map[string]any{},
				Prefix:  "server",
				Command: "COMMAND",
				Params:  []string{"param1", "param2", "param3"},
			},
		},
		{
			name:  "Command only",
			input: "COMMAND",
			want: &IRCMessage{
				Raw:     "COMMAND",
				Tags:    map[string]any{},
				Prefix:  "",
				Command: "COMMAND",
				Params:  []string{},
			},
		},
		{
			name:    "Empty string",
			input:   "",
			wantNil: true,
		},
		{
			name:    "Only tags no command",
			input:   "@tag=value",
			wantNil: true,
		},
		{
			name:    "Only prefix no command",
			input:   ":server.com",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseMessage(tt.input)
			if tt.wantNil {
				if got != nil {
					t.Errorf("ParseMessage(%q) = %+v, want nil", tt.input, got)
				}
				return
			}

			if got == nil {
				t.Fatalf("ParseMessage(%q) = nil, want non-nil", tt.input)
			}

			if got.Raw != tt.want.Raw {
				t.Errorf("Raw = %q, want %q", got.Raw, tt.want.Raw)
			}
			if got.Prefix != tt.want.Prefix {
				t.Errorf("Prefix = %q, want %q", got.Prefix, tt.want.Prefix)
			}
			if got.Command != tt.want.Command {
				t.Errorf("Command = %q, want %q", got.Command, tt.want.Command)
			}
			if !reflect.DeepEqual(got.Params, tt.want.Params) {
				t.Errorf("Params = %v, want %v", got.Params, tt.want.Params)
			}
			if !reflect.DeepEqual(got.Tags, tt.want.Tags) {
				t.Errorf("Tags = %v, want %v", got.Tags, tt.want.Tags)
			}
		})
	}
}

func TestParseBadges(t *testing.T) {
	tests := []struct {
		name string
		tags map[string]any
		want map[string]any
	}{
		{
			name: "Single badge",
			tags: map[string]any{"badges": "broadcaster/1"},
			want: map[string]any{
				"badges":     map[string]any{"broadcaster": "1"},
				"badges-raw": "broadcaster/1",
			},
		},
		{
			name: "Multiple badges",
			tags: map[string]any{"badges": "broadcaster/1,moderator/1,subscriber/12"},
			want: map[string]any{
				"badges": map[string]any{
					"broadcaster": "1",
					"moderator":   "1",
					"subscriber":  "12",
				},
				"badges-raw": "broadcaster/1,moderator/1,subscriber/12",
			},
		},
		{
			name: "Empty badges",
			tags: map[string]any{"badges": true},
			want: map[string]any{
				"badges":     nil,
				"badges-raw": nil,
			},
		},
		{
			name: "No badges tag",
			tags: map[string]any{},
			want: map[string]any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseBadges(tt.tags)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseBadges() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestParseBadgeInfo(t *testing.T) {
	tests := []struct {
		name string
		tags map[string]any
		want map[string]any
	}{
		{
			name: "Subscriber badge info",
			tags: map[string]any{"badge-info": "subscriber/12"},
			want: map[string]any{
				"badge-info":     map[string]any{"subscriber": "12"},
				"badge-info-raw": "subscriber/12",
			},
		},
		{
			name: "No badge-info",
			tags: map[string]any{},
			want: map[string]any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseBadgeInfo(tt.tags)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseBadgeInfo() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestParseEmotes(t *testing.T) {
	tests := []struct {
		name string
		tags map[string]any
		want map[string]any
	}{
		{
			name: "Single emote with one occurrence",
			tags: map[string]any{"emotes": "25:0-4"},
			want: map[string]any{
				"emotes":     map[string]any{"25": [][]int{{0, 4}}},
				"emotes-raw": "25:0-4",
			},
		},
		{
			name: "Single emote with multiple occurrences",
			tags: map[string]any{"emotes": "25:0-4,6-10"},
			want: map[string]any{
				"emotes":     map[string]any{"25": [][]int{{0, 4}, {6, 10}}},
				"emotes-raw": "25:0-4,6-10",
			},
		},
		{
			name: "Multiple emotes",
			tags: map[string]any{"emotes": "25:0-4/1902:6-10"},
			want: map[string]any{
				"emotes": map[string]any{
					"25":   [][]int{{0, 4}},
					"1902": [][]int{{6, 10}},
				},
				"emotes-raw": "25:0-4/1902:6-10",
			},
		},
		{
			name: "No emotes",
			tags: map[string]any{"emotes": true},
			want: map[string]any{
				"emotes":     nil,
				"emotes-raw": nil,
			},
		},
		{
			name: "No emotes tag",
			tags: map[string]any{},
			want: map[string]any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseEmotes(tt.tags)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseEmotes() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestFormTags(t *testing.T) {
	tests := []struct {
		name string
		tags map[string]string
		want string
	}{
		{
			name: "Empty tags",
			tags: map[string]string{},
			want: "",
		},
		{
			name: "Single tag",
			tags: map[string]string{"key": "value"},
			want: "@key=value",
		},
		{
			name: "Tag with special characters",
			tags: map[string]string{"key": "hello world"},
			want: "@key=hello\\sworld",
		},
		{
			name: "Tag with semicolon",
			tags: map[string]string{"key": "val;ue"},
			want: "@key=val\\:ue",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormTags(tt.tags)

			// For multiple tags, we can't guarantee order, so just check it starts with @
			// and contains the expected elements
			if len(tt.tags) == 0 {
				if got != tt.want {
					t.Errorf("FormTags() = %q, want %q", got, tt.want)
				}
			} else if len(tt.tags) == 1 {
				if got != tt.want {
					t.Errorf("FormTags() = %q, want %q", got, tt.want)
				}
			} else {
				// Multiple tags - just verify format
				if got == "" || got[0] != '@' {
					t.Errorf("FormTags() = %q, should start with @", got)
				}
			}
		})
	}
}
