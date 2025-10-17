package tmigo

import (
	"strings"
	"testing"
)

func TestJustinfan(t *testing.T) {
	username := Justinfan()

	if !strings.HasPrefix(username, "justinfan") {
		t.Errorf("Justinfan() should start with 'justinfan', got %s", username)
	}

	// Check that it generates different usernames
	username2 := Justinfan()
	if username == username2 {
		// It's possible (but unlikely) they're the same, try a few more times
		allSame := true
		for i := 0; i < 10; i++ {
			if Justinfan() != username {
				allSame = false
				break
			}
		}
		if allSame {
			t.Error("Justinfan() appears to generate the same username every time")
		}
	}
}

func TestIsJustinfan(t *testing.T) {
	tests := []struct {
		name     string
		username string
		want     bool
	}{
		{"Valid justinfan", "justinfan12345", true},
		{"Valid justinfan with different number", "justinfan99999", true},
		{"Not justinfan", "regular_user", false},
		{"Partial match", "justinfan", false},
		{"Justinfan with letters", "justinfanabc", false},
		{"Empty string", "", false},
		{"Uppercase", "JUSTINFAN12345", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsJustinfan(tt.username)
			if got != tt.want {
				t.Errorf("IsJustinfan(%q) = %v, want %v", tt.username, got, tt.want)
			}
		})
	}
}

func TestChannel(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"With hash", "#channel", "#channel"},
		{"Without hash", "channel", "#channel"},
		{"With spaces", "  channel  ", "#channel"},
		{"Uppercase", "CHANNEL", "#channel"},
		{"Mixed case with hash", "#MyChannel", "#mychannel"},
		{"Empty", "", "#"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Channel(tt.input)
			if got != tt.want {
				t.Errorf("Channel(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestUsername(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"Without hash", "username", "username"},
		{"With hash", "#username", "username"},
		{"With spaces", "  username  ", "username"},
		{"Uppercase", "USERNAME", "username"},
		{"Mixed case", "UserName", "username"},
		{"Empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Username(tt.input)
			if got != tt.want {
				t.Errorf("Username(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestToken(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"Without prefix", "abc123def", "abc123def"},
		{"With oauth prefix", "oauth:abc123def", "abc123def"},
		{"Empty", "", ""},
		{"Only prefix", "oauth:", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Token(tt.input)
			if got != tt.want {
				t.Errorf("Token(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestPassword(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"Without prefix", "abc123def", "oauth:abc123def"},
		{"With oauth prefix", "oauth:abc123def", "oauth:abc123def"},
		{"Empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Password(tt.input)
			if got != tt.want {
				t.Errorf("Password(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsActionMessage(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantIsAction bool
		wantMessage  string
	}{
		{"Action message", "\x01ACTION does something\x01", true, "does something"},
		{"Action with spaces", "\x01ACTION  multiple words  \x01", true, " multiple words  "},
		{"Regular message", "just a regular message", false, ""},
		{"Empty", "", false, ""},
		{"Only ACTION with space", "\x01ACTION  \x01", true, " "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIsAction, gotMessage := IsActionMessage(tt.input)
			if gotIsAction != tt.wantIsAction {
				t.Errorf("IsActionMessage(%q) isAction = %v, want %v", tt.input, gotIsAction, tt.wantIsAction)
			}
			if gotMessage != tt.wantMessage {
				t.Errorf("IsActionMessage(%q) message = %q, want %q", tt.input, gotMessage, tt.wantMessage)
			}
		})
	}
}

func TestUnescapeIRC(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"Space escape", `\s`, " "},
		{"Newline escape", `\n`, ""},
		{"Colon escape", `\:`, ";"},
		{"Carriage return escape", `\r`, ""},
		{"Backslash escape", `\\`, "\\"},
		{"Multiple escapes", `hello\sworld\:test`, "hello world;test"},
		{"No escapes", "hello world", "hello world"},
		{"Empty", "", ""},
		{"Unknown escape", `\x`, "x"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UnescapeIRC(tt.input)
			if got != tt.want {
				t.Errorf("UnescapeIRC(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestEscapeIRC(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"Space", " ", `\s`},
		{"Newline", "\n", `\n`},
		{"Semicolon", ";", `\:`},
		{"Carriage return", "\r", `\r`},
		{"Backslash", "\\", `\\`},
		{"Multiple characters", "hello world;test", `hello\sworld\:test`},
		{"No special chars", "hello", "hello"},
		{"Empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EscapeIRC(tt.input)
			if got != tt.want {
				t.Errorf("EscapeIRC(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestEscapeUnescapeRoundtrip(t *testing.T) {
	tests := []string{
		"hello world",
		"test;value",
		"back\\slash",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			escaped := EscapeIRC(input)
			unescaped := UnescapeIRC(escaped)
			if unescaped != input {
				t.Errorf("Roundtrip failed: %q -> %q -> %q", input, escaped, unescaped)
			}
		})
	}
}

func TestEscapeUnescapeNewlineCarriageReturn(t *testing.T) {
	// Note: IRC escaping converts \n and \r to empty strings when unescaping
	// This is intentional behavior for IRC message tag values
	tests := []struct {
		input    string
		expected string
	}{
		{"line\nbreak", "linebreak"},
		{"carriage\rreturn", "carriagereturn"},
		{"mixed; value\nwith\\stuff", "mixed; valuewith\\stuff"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			escaped := EscapeIRC(tt.input)
			unescaped := UnescapeIRC(escaped)
			if unescaped != tt.expected {
				t.Errorf("EscapeIRC(%q) -> UnescapeIRC = %q, want %q", tt.input, unescaped, tt.expected)
			}
		})
	}
}

func TestIsInteger(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"Valid positive", "123", true},
		{"Valid negative", "-123", true},
		{"Valid zero", "0", true},
		{"Invalid with letters", "123abc", false},
		{"Invalid with decimal", "123.45", false},
		{"Empty", "", false},
		{"Just minus", "-", false},
		{"Large number", "999999999", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsInteger(tt.input)
			if got != tt.want {
				t.Errorf("IsInteger(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseInt(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"Valid positive", "123", 123},
		{"Valid negative", "-123", -123},
		{"Valid zero", "0", 0},
		{"Invalid", "abc", 0},
		{"Empty", "", 0},
		{"Decimal", "123.45", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseInt(tt.input)
			if got != tt.want {
				t.Errorf("ParseInt(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}
