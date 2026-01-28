package filters

import (
	"fmt"
	"html"
	"net/url"
	"regexp"
	"sort"
	"strings"
)

// Pre-compiled regex patterns for HTML filters (performance optimization)
var (
	reHTMLTags = regexp.MustCompile(`<[^>]*>`)
	reURLs     = regexp.MustCompile(`https?://[^\s<>"']+`)
)

// SafeValue represents a value that should not be escaped
type SafeValue struct {
	Value interface{}
}

// String returns the string representation of the safe value
func (s SafeValue) String() string {
	return ToString(s.Value)
}

// EscapeFilter escapes HTML characters
func EscapeFilter(value interface{}, args ...interface{}) (interface{}, error) {
	// Don't escape if already marked as safe
	if _, ok := value.(SafeValue); ok {
		return value, nil
	}

	s := ToString(value)
	return html.EscapeString(s), nil
}

// SafeFilter marks a value as safe (won't be escaped)
func SafeFilter(value interface{}, args ...interface{}) (interface{}, error) {
	return SafeValue{Value: value}, nil
}

// URLEncodeFilter URL-encodes a string
func URLEncodeFilter(value interface{}, args ...interface{}) (interface{}, error) {
	s := ToString(value)
	return url.QueryEscape(s), nil
}

// XMLAttrFilter formats attributes for XML/HTML
func XMLAttrFilter(value interface{}, args ...interface{}) (interface{}, error) {
	if value == nil {
		return "", nil
	}

	var result strings.Builder
	attrCount := 0

	switch v := value.(type) {
	case map[string]interface{}:
		// Sort keys for consistent output
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, key := range keys {
			val := v[key]
			if val == nil {
				continue
			}

			// Handle boolean attributes
			if b, ok := val.(bool); ok {
				if b {
					if attrCount > 0 {
						result.WriteByte(' ')
					}
					result.WriteString(html.EscapeString(key))
					attrCount++
				}
				continue
			}

			// Regular key="value" attributes
			attrValue := ToString(val)
			if attrValue != "" {
				if attrCount > 0 {
					result.WriteByte(' ')
				}
				result.WriteString(html.EscapeString(key))
				result.WriteString(`="`)
				result.WriteString(html.EscapeString(attrValue))
				result.WriteByte('"')
				attrCount++
			}
		}
	case map[string]string:
		// Sort keys for consistent output
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, key := range keys {
			val := v[key]
			if val != "" {
				if attrCount > 0 {
					result.WriteByte(' ')
				}
				result.WriteString(html.EscapeString(key))
				result.WriteString(`="`)
				result.WriteString(html.EscapeString(val))
				result.WriteByte('"')
				attrCount++
			}
		}
	default:
		return "", fmt.Errorf("xmlattr filter requires a mapping")
	}

	if attrCount == 0 {
		return "", nil
	}

	return " " + result.String(), nil
}

// StripTagsFilter removes HTML/XML tags
func StripTagsFilter(value interface{}, args ...interface{}) (interface{}, error) {
	s := ToString(value)

	// Simple HTML tag removal using pre-compiled regex
	result := reHTMLTags.ReplaceAllString(s, "")

	// Decode HTML entities
	result = html.UnescapeString(result)

	return result, nil
}

// UrlizeFilter converts URLs in text to clickable links
func UrlizeFilter(value interface{}, args ...interface{}) (interface{}, error) {
	s := ToString(value)

	trimURLs := true
	nofollow := false
	target := ""
	rel := ""

	if len(args) > 0 {
		trimURLs = ToBool(args[0])
	}
	if len(args) > 1 {
		nofollow = ToBool(args[1])
	}
	if len(args) > 2 {
		target = ToString(args[2])
	}
	if len(args) > 3 {
		rel = ToString(args[3])
	}

	// Use pre-compiled URL regex
	result := reURLs.ReplaceAllStringFunc(s, func(match string) string {
		displayURL := match

		// Trim long URLs
		if trimURLs && len(match) > 40 {
			displayURL = match[:37] + "..."
		}

		// Build link attributes
		var attrs []string
		attrs = append(attrs, fmt.Sprintf(`href="%s"`, html.EscapeString(match)))

		if target != "" {
			attrs = append(attrs, fmt.Sprintf(`target="%s"`, html.EscapeString(target)))
		}

		if nofollow {
			if rel != "" {
				rel = "nofollow " + rel
			} else {
				rel = "nofollow"
			}
		}

		if rel != "" {
			attrs = append(attrs, fmt.Sprintf(`rel="%s"`, html.EscapeString(rel)))
		}

		return fmt.Sprintf(`<a %s>%s</a>`, strings.Join(attrs, " "), html.EscapeString(displayURL))
	})

	return SafeValue{Value: result}, nil
}

// UrlizeTargetFilter is like urlize but with a specific target
func UrlizeTargetFilter(value interface{}, args ...interface{}) (interface{}, error) {
	target := "_blank"
	if len(args) > 0 {
		target = ToString(args[0])
	}

	// Call urlize with target parameter
	newArgs := []interface{}{true, false, target}
	if len(args) > 1 {
		newArgs = append(newArgs, args[1:]...)
	}

	return UrlizeFilter(value, newArgs...)
}

// TruncateHTMLFilter truncates HTML content safely
func TruncateHTMLFilter(value interface{}, args ...interface{}) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("truncatehtml filter requires length argument")
	}

	s := ToString(value)
	length, err := ToInt(args[0])
	if err != nil {
		return nil, fmt.Errorf("truncatehtml length must be integer: %v", err)
	}

	killwords := false
	end := "..."

	if len(args) > 1 {
		killwords = ToBool(args[1])
	}
	if len(args) > 2 {
		end = ToString(args[2])
	}

	// Strip tags for length calculation
	strippedResult, _ := StripTagsFilter(s)
	stripped := ToString(strippedResult)

	if len([]rune(stripped)) <= length {
		return SafeValue{Value: s}, nil
	}

	// This is a simplified version - a full implementation would
	// need to parse HTML properly and maintain tag balance
	runes := []rune(stripped)
	truncated := string(runes[:length])

	if !killwords {
		lastSpace := strings.LastIndex(truncated, " ")
		if lastSpace > 0 && lastSpace > length/2 {
			truncated = truncated[:lastSpace]
		}
	}

	return SafeValue{Value: truncated + end}, nil
}

// FileSizeFormatFilter formats file sizes in human-readable format
func FileSizeFormatFilter(value interface{}, args ...interface{}) (interface{}, error) {
	size, err := ToFloat(value)
	if err != nil {
		return nil, fmt.Errorf("filesizeformat requires numeric value: %v", err)
	}

	binary := false
	if len(args) > 0 {
		binary = ToBool(args[0])
	}

	var units []string
	var base float64

	if binary {
		units = []string{"Bytes", "KiB", "MiB", "GiB", "TiB", "PiB"}
		base = 1024
	} else {
		units = []string{"Bytes", "KB", "MB", "GB", "TB", "PB"}
		base = 1000
	}

	if size < base {
		return fmt.Sprintf("%.0f %s", size, units[0]), nil
	}

	for i := 1; i < len(units); i++ {
		if size < base*base {
			return fmt.Sprintf("%.1f %s", size/base, units[i]), nil
		}
		size /= base
	}

	return fmt.Sprintf("%.1f %s", size, units[len(units)-1]), nil
}

// AutoEscapeFilter conditionally escapes based on context
func AutoEscapeFilter(value interface{}, args ...interface{}) (interface{}, error) {
	// This would be called by the template engine based on auto-escape settings
	return EscapeFilter(value, args...)
}

// MarkSafeFilter marks content as safe (alias for safe)
func MarkSafeFilter(value interface{}, args ...interface{}) (interface{}, error) {
	return SafeFilter(value, args...)
}

// ForceEscapeFilter forces HTML escaping even for safe values
func ForceEscapeFilter(value interface{}, args ...interface{}) (interface{}, error) {
	// Force escape even if marked as safe
	var s string
	if safeVal, ok := value.(SafeValue); ok {
		s = ToString(safeVal.Value)
	} else {
		s = ToString(value)
	}
	// Return as SafeValue to prevent double-escaping
	return SafeValue{Value: html.EscapeString(s)}, nil
}

// NL2BRFilter converts newlines to HTML <br> tags
func NL2BRFilter(value interface{}, args ...interface{}) (interface{}, error) {
	s := ToString(value)

	// Replace different types of line breaks
	result := strings.ReplaceAll(s, "\r\n", "<br>")   // Windows
	result = strings.ReplaceAll(result, "\r", "<br>") // Mac
	result = strings.ReplaceAll(result, "\n", "<br>") // Unix

	return SafeValue{Value: result}, nil
}

// UrlizeTruncateFilter combines urlize with truncation
func UrlizeTruncateFilter(value interface{}, args ...interface{}) (interface{}, error) {
	length := 40
	killwords := false
	target := ""
	rel := ""

	if len(args) > 0 {
		var err error
		length, err = ToInt(args[0])
		if err != nil {
			return nil, fmt.Errorf("urlizetruncate length must be integer: %v", err)
		}
	}
	if len(args) > 1 {
		killwords = ToBool(args[1])
	}
	if len(args) > 2 {
		target = ToString(args[2])
	}
	if len(args) > 3 {
		rel = ToString(args[3])
	}

	s := ToString(value)

	// Use pre-compiled URL regex (same as urlize)
	result := reURLs.ReplaceAllStringFunc(s, func(match string) string {
		displayURL := match

		// Truncate the display URL
		if len([]rune(match)) > length {
			runes := []rune(match)
			if killwords {
				// Ensure the truncation includes the "..." in the length calculation
				truncateAt := length - 3
				if truncateAt < 0 {
					truncateAt = 0
				}
				displayURL = string(runes[:truncateAt]) + "..."
			} else {
				// Find last slash or dot before the limit
				truncateAt := length - 3
				if truncateAt < 0 {
					truncateAt = 0
				}
				truncated := string(runes[:truncateAt])
				lastSlash := strings.LastIndex(truncated, "/")
				lastDot := strings.LastIndex(truncated, ".")

				breakPoint := -1
				if lastSlash > lastDot {
					breakPoint = lastSlash
				} else if lastDot > 0 {
					breakPoint = lastDot
				}

				if breakPoint > truncateAt/2 {
					displayURL = truncated[:breakPoint] + "..."
				} else {
					displayURL = truncated + "..."
				}
			}
		}

		// Build link attributes
		var attrs []string
		attrs = append(attrs, fmt.Sprintf(`href="%s"`, html.EscapeString(match)))

		if target != "" {
			attrs = append(attrs, fmt.Sprintf(`target="%s"`, html.EscapeString(target)))
		}

		if rel != "" {
			attrs = append(attrs, fmt.Sprintf(`rel="%s"`, html.EscapeString(rel)))
		}

		return fmt.Sprintf(`<a %s>%s</a>`, strings.Join(attrs, " "), html.EscapeString(displayURL))
	})

	return SafeValue{Value: result}, nil
}
