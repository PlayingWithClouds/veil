package stream

import "strings"

// attributeValue returns the value of name in an HLS attribute list, without
// quotes, or "" when absent.
func attributeValue(line, name string) string {
	_, attributes, _ := strings.Cut(line, ":")
	for attributes != "" {
		key, rest, found := strings.Cut(attributes, "=")
		if !found {
			return ""
		}
		var value string
		if strings.HasPrefix(rest, `"`) {
			quoted, remainder, _ := strings.Cut(rest[1:], `"`)
			value = quoted
			attributes = strings.TrimPrefix(remainder, ",")
		} else {
			unquoted, remainder, _ := strings.Cut(rest, ",")
			value = unquoted
			attributes = remainder
		}
		if strings.TrimSpace(key) == name {
			return value
		}
	}
	return ""
}
