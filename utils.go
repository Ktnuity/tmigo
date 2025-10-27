package tmigo

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	actionMessageRegex = regexp.MustCompile(`^\x01ACTION ([^\x01]+)\x01$`)
	justinfanRegex     = regexp.MustCompile(`^(justinfan)(\d+$)`)
	tokenRegex         = regexp.MustCompile(`^oauth:`)
)

var ircEscapedChars = map[rune]string{
	's':  " ",
	'n':  "",
	':':  ";",
	'r':  "",
	'\\': "\\",
}

var ircUnescapedChars = map[rune]string{
	' ':  "s",
	'\n': "n",
	';':  ":",
	'\r': "r",
	'\\': "\\",
}

// Justinfan returns a random justinfan username for anonymous connections
func Justinfan() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return fmt.Sprintf("justinfan%d", r.Intn(80000)+1000)
}

// IsJustinfan checks if a username is a justinfan username
func IsJustinfan(username string) bool {
	return justinfanRegex.MatchString(username)
}

// Channel returns a valid channel name (with # prefix)
func Channel(str string) string {
	channel := strings.ToLower(strings.TrimSpace(str))
	if channel == "" {
		return "#"
	}
	if strings.HasPrefix(channel, "#") {
		return channel
	}
	return "#" + channel
}

// ChannelAll returns valid channel names for all values
func ChannelAll(strs []string) []string {
	if len(strs) == 0 {
		return []string{}
	}

	res := make([]string, len(strs))

	for idx, str := range strs {
		res[idx] = Channel(str)
	}

	return res
}

// Username returns a valid username (without # prefix)
func Username(str string) string {
	username := strings.ToLower(strings.TrimSpace(str))
	if username == "" {
		return ""
	}
	if strings.HasPrefix(username, "#") {
		return username[1:]
	}
	return username
}

// Token returns a valid token (removes oauth: prefix if present)
func Token(str string) string {
	if str == "" {
		return ""
	}
	return tokenRegex.ReplaceAllString(str, "")
}

// Password returns a valid password with oauth: prefix
func Password(str string) string {
	token := Token(str)
	if token == "" {
		return ""
	}
	return "oauth:" + token
}

// IsActionMessage checks if a message is an action message (/me)
func IsActionMessage(msg string) (bool, string) {
	matches := actionMessageRegex.FindStringSubmatch(msg)
	if len(matches) > 1 {
		return true, matches[1]
	}
	return false, ""
}

// UnescapeIRC unescapes IRC message tag values
func UnescapeIRC(msg string) string {
	if msg == "" || !strings.Contains(msg, "\\") {
		return msg
	}

	result := strings.Builder{}
	escaped := false

	for _, ch := range msg {
		if escaped {
			if replacement, ok := ircEscapedChars[ch]; ok {
				result.WriteString(replacement)
			} else {
				result.WriteRune(ch)
			}
			escaped = false
		} else if ch == '\\' {
			escaped = true
		} else {
			result.WriteRune(ch)
		}
	}

	return result.String()
}

// EscapeIRC escapes values for IRC message tags
func EscapeIRC(msg string) string {
	if msg == "" {
		return msg
	}

	result := strings.Builder{}
	for _, ch := range msg {
		if replacement, ok := ircUnescapedChars[ch]; ok {
			result.WriteRune('\\')
			result.WriteString(replacement)
		} else {
			result.WriteRune(ch)
		}
	}

	return result.String()
}

// IsInteger checks if a string can be parsed as an integer
func IsInteger(input string) bool {
	_, err := strconv.Atoi(input)
	return err == nil
}

// ParseInt safely parses a string to int, returns 0 on error
func ParseInt(input string) int {
	val, err := strconv.Atoi(input)
	if err != nil {
		return 0
	}
	return val
}
