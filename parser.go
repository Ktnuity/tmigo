package tmigo

import (
	"strconv"
	"strings"
)

// ParseMessage parses an IRC message into an IRCMessage struct
func ParseMessage(data string) *IRCMessage {
	if data == "" {
		return nil
	}

	message := &IRCMessage{
		Raw:     data,
		Tags:    make(map[string]any),
		Prefix:  "",
		Command: "",
		Params:  []string{},
	}

	position := 0
	nextspace := 0

	// Parse IRCv3.2 message tags
	if len(data) > 0 && data[0] == '@' {
		nextspace = strings.Index(data, " ")
		if nextspace == -1 {
			return nil
		}

		rawTags := strings.Split(data[1:nextspace], ";")
		for _, tag := range rawTags {
			idx := strings.Index(tag, "=")
			if idx == -1 {
				message.Tags[tag] = true
			} else {
				key := tag[:idx]
				value := tag[idx+1:]
				if value == "" {
					message.Tags[key] = true
				} else {
					message.Tags[key] = value
				}
			}
		}

		position = nextspace + 1
	}

	// Skip trailing whitespace
	for position < len(data) && data[position] == ' ' {
		position++
	}

	// Extract prefix if present
	if position < len(data) && data[position] == ':' {
		nextspace = strings.Index(data[position:], " ")
		if nextspace == -1 {
			return nil
		}
		nextspace += position

		message.Prefix = data[position+1 : nextspace]
		position = nextspace + 1

		// Skip trailing whitespace
		for position < len(data) && data[position] == ' ' {
			position++
		}
	}

	// Extract command
	nextspace = strings.Index(data[position:], " ")
	if nextspace == -1 {
		if len(data) > position {
			message.Command = data[position:]
			return message
		}
		return nil
	}
	nextspace += position

	message.Command = data[position:nextspace]
	position = nextspace + 1

	// Skip trailing whitespace
	for position < len(data) && data[position] == ' ' {
		position++
	}

	// Extract parameters
	for position < len(data) {
		// Trailing parameter
		if data[position] == ':' {
			message.Params = append(message.Params, data[position+1:])
			break
		}

		nextspace = strings.Index(data[position:], " ")
		if nextspace == -1 {
			message.Params = append(message.Params, data[position:])
			break
		}
		nextspace += position

		message.Params = append(message.Params, data[position:nextspace])
		position = nextspace + 1

		// Skip trailing whitespace
		for position < len(data) && data[position] == ' ' {
			position++
		}
	}

	return message
}

// ParseBadges parses the badges tag
func ParseBadges(tags map[string]any) map[string]any {
	return parseComplexTag(tags, "badges", ",", "/", "")
}

// ParseBadgeInfo parses the badge-info tag
func ParseBadgeInfo(tags map[string]any) map[string]any {
	return parseComplexTag(tags, "badge-info", ",", "/", "")
}

// ParseEmotes parses the emotes tag
func ParseEmotes(tags map[string]any) map[string]any {
	return parseComplexTag(tags, "emotes", "/", ":", ",")
}

func parseComplexTag(tags map[string]any, tagKey string, splA, splB, splC string) map[string]any {
	raw, exists := tags[tagKey]
	if !exists {
		return tags
	}

	rawStr, isString := raw.(string)
	if isString {
		tags[tagKey+"-raw"] = rawStr
	} else {
		tags[tagKey+"-raw"] = nil
	}

	// If raw is true (boolean), set tag to null
	if rawBool, isBool := raw.(bool); isBool && rawBool {
		tags[tagKey] = nil
		return tags
	}

	result := make(map[string]any)
	tags[tagKey] = result

	if !isString {
		return tags
	}

	for part := range strings.SplitSeq(rawStr, splA) {
		subParts := strings.Split(part, splB)
		if len(subParts) < 1 {
			continue
		}

		key := subParts[0]
		var value any

		if len(subParts) > 1 {
			val := subParts[1]
			if splC != "" && val != "" {
				// For emotes, we need to parse the positions
				positions := strings.Split(val, splC)
				posArray := make([][]int, 0, len(positions))
				for _, pos := range positions {
					rangeParts := strings.Split(pos, "-")
					if len(rangeParts) == 2 {
						start, err1 := strconv.Atoi(rangeParts[0])
						end, err2 := strconv.Atoi(rangeParts[1])
						if err1 == nil && err2 == nil {
							posArray = append(posArray, []int{start, end})
						}
					}
				}
				value = posArray
			} else {
				value = val
			}
		}

		if value == nil || value == "" {
			result[key] = nil
		} else {
			result[key] = value
		}
	}

	return tags
}

// FormTags formats tags for outgoing messages
func FormTags(tags map[string]string) string {
	if len(tags) == 0 {
		return ""
	}

	parts := make([]string, 0, len(tags))
	for k, v := range tags {
		parts = append(parts, EscapeIRC(k)+"="+EscapeIRC(v))
	}

	return "@" + strings.Join(parts, ";")
}
