package filters

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Pre-compiled regex patterns for string filters (performance optimization)
var (
	// Slugify patterns
	reSlugifySpaces    = regexp.MustCompile(`[\s\-_]+`)
	reSlugifyNonWord   = regexp.MustCompile(`[^\w\-]`)
	reRegexReplaceTest = regexp.MustCompile(`^/(.+)/([gimsuy]*)$`)
)

// titleCase converts a string to title case (capitalize first letter of each word).
// This replaces the deprecated strings.Title function.
func titleCase(s string) string {
	// Use a simple state machine to capitalize letters after spaces
	prev := ' '
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(prev) {
			prev = r
			return unicode.ToTitle(r)
		}
		prev = r
		return r
	}, s)
}

// capitalizeFirst capitalizes only the first letter of a string.
// Used for struct field/method name lookups.
func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError {
		return s
	}
	return string(unicode.ToUpper(r)) + s[size:]
}

// UpperFilter converts string to uppercase
func UpperFilter(value interface{}, args ...interface{}) (interface{}, error) {
	s := ToString(value)
	return strings.ToUpper(s), nil
}

// LowerFilter converts string to lowercase
func LowerFilter(value interface{}, args ...interface{}) (interface{}, error) {
	s := ToString(value)
	return strings.ToLower(s), nil
}

// CapitalizeFilter capitalizes the first character
func CapitalizeFilter(value interface{}, args ...interface{}) (interface{}, error) {
	s := ToString(value)
	if len(s) == 0 {
		return s, nil
	}

	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	for i := 1; i < len(runes); i++ {
		runes[i] = unicode.ToLower(runes[i])
	}

	return string(runes), nil
}

// TitleFilter converts string to title case
func TitleFilter(value interface{}, args ...interface{}) (interface{}, error) {
	s := ToString(value)
	return titleCase(s), nil
}

// TrimFilter removes leading and trailing whitespace
func TrimFilter(value interface{}, args ...interface{}) (interface{}, error) {
	s := ToString(value)
	if len(args) > 0 {
		// Trim specific characters
		chars := ToString(args[0])
		return strings.Trim(s, chars), nil
	}
	return strings.TrimSpace(s), nil
}

// LstripFilter removes leading whitespace
func LstripFilter(value interface{}, args ...interface{}) (interface{}, error) {
	s := ToString(value)
	if len(args) > 0 {
		// Trim specific characters from left
		chars := ToString(args[0])
		return strings.TrimLeft(s, chars), nil
	}
	return strings.TrimLeftFunc(s, unicode.IsSpace), nil
}

// RstripFilter removes trailing whitespace
func RstripFilter(value interface{}, args ...interface{}) (interface{}, error) {
	s := ToString(value)
	if len(args) > 0 {
		// Trim specific characters from right
		chars := ToString(args[0])
		return strings.TrimRight(s, chars), nil
	}
	return strings.TrimRightFunc(s, unicode.IsSpace), nil
}

// ReplaceFilter replaces occurrences of old with new
func ReplaceFilter(value interface{}, args ...interface{}) (interface{}, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("replace filter requires 2 arguments: old and new")
	}

	s := ToString(value)
	old := ToString(args[0])
	new := ToString(args[1])

	count := -1 // Replace all by default
	if len(args) > 2 {
		if c, err := ToInt(args[2]); err == nil {
			count = c
		}
	}

	if count == -1 {
		return strings.ReplaceAll(s, old, new), nil
	}
	return strings.Replace(s, old, new, count), nil
}

// TruncateFilter truncates string to specified length
func TruncateFilter(value interface{}, args ...interface{}) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("truncate filter requires length argument")
	}

	s := ToString(value)
	length, err := ToInt(args[0])
	if err != nil {
		return nil, fmt.Errorf("truncate length must be integer: %v", err)
	}

	killwords := false
	end := "..."

	if len(args) > 1 {
		killwords = ToBool(args[1])
	}
	if len(args) > 2 {
		end = ToString(args[2])
	}

	runes := []rune(s)
	if len(runes) <= length {
		return s, nil
	}

	if killwords {
		// Cut at exact length
		return string(runes[:length]) + end, nil
	}

	// Try to break at word boundary
	truncated := string(runes[:length])
	lastSpace := strings.LastIndex(truncated, " ")
	if lastSpace > 0 && lastSpace > length/2 {
		truncated = truncated[:lastSpace]
	}

	return truncated + end, nil
}

// WordwrapFilter wraps words at specified width
func WordwrapFilter(value interface{}, args ...interface{}) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("wordwrap filter requires width argument")
	}

	s := ToString(value)
	width, err := ToInt(args[0])
	if err != nil {
		return nil, fmt.Errorf("wordwrap width must be integer: %v", err)
	}

	breakOnHyphens := true
	wrapString := "\n"

	if len(args) > 1 {
		breakOnHyphens = ToBool(args[1])
	}
	if len(args) > 2 {
		wrapString = ToString(args[2])
	}

	words := strings.Fields(s)
	if len(words) == 0 {
		return s, nil
	}

	var lines []string
	var currentLine strings.Builder

	for _, word := range words {
		if breakOnHyphens && strings.Contains(word, "-") {
			// Split on hyphens too
			parts := strings.Split(word, "-")
			for i, part := range parts {
				if i > 0 {
					part = "-" + part
				}
				if currentLine.Len()+len(part)+1 > width && currentLine.Len() > 0 {
					lines = append(lines, currentLine.String())
					currentLine.Reset()
				}
				if currentLine.Len() > 0 {
					currentLine.WriteString(" ")
				}
				currentLine.WriteString(part)
			}
		} else {
			if currentLine.Len()+len(word)+1 > width && currentLine.Len() > 0 {
				lines = append(lines, currentLine.String())
				currentLine.Reset()
			}
			if currentLine.Len() > 0 {
				currentLine.WriteString(" ")
			}
			currentLine.WriteString(word)
		}
	}

	if currentLine.Len() > 0 {
		lines = append(lines, currentLine.String())
	}

	return strings.Join(lines, wrapString), nil
}

// CenterFilter centers string in field of given width
func CenterFilter(value interface{}, args ...interface{}) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("center filter requires width argument")
	}

	s := ToString(value)
	width, err := ToInt(args[0])
	if err != nil {
		return nil, fmt.Errorf("center width must be integer: %v", err)
	}

	fillchar := " "
	if len(args) > 1 {
		fillchar = ToString(args[1])
		if len(fillchar) == 0 {
			fillchar = " "
		}
	}

	sLen := len([]rune(s))
	if sLen >= width {
		return s, nil
	}

	padding := width - sLen
	leftPad := padding / 2
	rightPad := padding - leftPad

	result := strings.Repeat(fillchar, leftPad) + s + strings.Repeat(fillchar, rightPad)
	return result, nil
}

// IndentFilter indents each line
func IndentFilter(value interface{}, args ...interface{}) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("indent filter requires width argument")
	}

	s := ToString(value)
	width, err := ToInt(args[0])
	if err != nil {
		return nil, fmt.Errorf("indent width must be integer: %v", err)
	}

	indentFirst := false
	indentString := " "

	if len(args) > 1 {
		indentFirst = ToBool(args[1])
	}
	if len(args) > 2 {
		indentString = ToString(args[2])
	}

	lines := strings.Split(s, "\n")
	prefix := strings.Repeat(indentString, width)

	for i, line := range lines {
		if i == 0 && !indentFirst {
			continue
		}
		if line != "" || i < len(lines)-1 {
			lines[i] = prefix + line
		}
	}

	return strings.Join(lines, "\n"), nil
}

// StringFilter converts value to string
func StringFilter(value interface{}, args ...interface{}) (interface{}, error) {
	return ToString(value), nil
}

// FormatFilter formats string using Printf-style formatting
// Note: This safely handles format strings by recovering from panics
// and validating that the format doesn't cause issues
func FormatFilter(value interface{}, args ...interface{}) (result interface{}, err error) {
	s := ToString(value)
	if len(args) == 0 {
		return s, nil
	}

	// Recover from any panics caused by malformed format strings
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("format error: %v", r)
			result = s
		}
	}()

	// Use the string as format and args as values
	return fmt.Sprintf(s, args...), nil
}

// Helper function to convert regex patterns for case-insensitive matching
func compileRegex(pattern string, ignoreCase bool) (*regexp.Regexp, error) {
	if ignoreCase {
		pattern = "(?i)" + pattern
	}
	return regexp.Compile(pattern)
}

// RegexReplaceFilter replaces text using regular expressions
func RegexReplaceFilter(value interface{}, args ...interface{}) (interface{}, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("regex_replace filter requires pattern and replacement arguments")
	}

	s := ToString(value)
	pattern := ToString(args[0])
	replacement := ToString(args[1])

	ignoreCase := false
	if len(args) > 2 {
		ignoreCase = ToBool(args[2])
	}

	regex, err := compileRegex(pattern, ignoreCase)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %v", err)
	}

	return regex.ReplaceAllString(s, replacement), nil
}

// RegexSearchFilter searches for pattern and returns first match
func RegexSearchFilter(value interface{}, args ...interface{}) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("regex_search filter requires pattern argument")
	}

	s := ToString(value)
	pattern := ToString(args[0])

	ignoreCase := false
	if len(args) > 1 {
		ignoreCase = ToBool(args[1])
	}

	regex, err := compileRegex(pattern, ignoreCase)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %v", err)
	}

	match := regex.FindString(s)
	if match == "" {
		return nil, nil
	}
	return match, nil
}

// RegexFindallFilter finds all matches of pattern
func RegexFindallFilter(value interface{}, args ...interface{}) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("regex_findall filter requires pattern argument")
	}

	s := ToString(value)
	pattern := ToString(args[0])

	ignoreCase := false
	if len(args) > 1 {
		ignoreCase = ToBool(args[1])
	}

	regex, err := compileRegex(pattern, ignoreCase)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %v", err)
	}

	matches := regex.FindAllString(s, -1)
	if matches == nil {
		return []interface{}{}, nil
	}

	// Convert to []interface{} for consistency
	result := make([]interface{}, len(matches))
	for i, match := range matches {
		result[i] = match
	}

	return result, nil
}

// SplitFilter splits string on delimiter
func SplitFilter(value interface{}, args ...interface{}) (interface{}, error) {
	s := ToString(value)
	delimiter := " " // Default delimiter

	if len(args) > 0 {
		delimiter = ToString(args[0])
	}

	maxSplit := -1
	if len(args) > 1 {
		if ms, err := ToInt(args[1]); err == nil {
			maxSplit = ms
		}
	}

	var parts []string
	if maxSplit == -1 {
		if delimiter == " " {
			// Special case: split on any whitespace
			parts = strings.Fields(s)
		} else {
			parts = strings.Split(s, delimiter)
		}
	} else {
		parts = strings.SplitN(s, delimiter, maxSplit+1)
	}

	// Convert to []interface{} for consistency
	result := make([]interface{}, len(parts))
	for i, part := range parts {
		result[i] = part
	}

	return result, nil
}

// StartswithFilter checks if string starts with prefix
func StartswithFilter(value interface{}, args ...interface{}) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("startswith filter requires prefix argument")
	}

	s := ToString(value)
	prefix := ToString(args[0])

	return strings.HasPrefix(s, prefix), nil
}

// EndswithFilter checks if string ends with suffix
func EndswithFilter(value interface{}, args ...interface{}) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("endswith filter requires suffix argument")
	}

	s := ToString(value)
	suffix := ToString(args[0])

	return strings.HasSuffix(s, suffix), nil
}

// ContainsFilter checks if string contains substring
func ContainsFilter(value interface{}, args ...interface{}) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("contains filter requires substring argument")
	}

	s := ToString(value)
	substring := ToString(args[0])

	return strings.Contains(s, substring), nil
}

// SlugifyFilter creates URL-friendly slug from string
func SlugifyFilter(value interface{}, args ...interface{}) (interface{}, error) {
	s := ToString(value)

	// Convert to lowercase
	s = strings.ToLower(s)

	// Replace spaces and common punctuation with hyphens (using pre-compiled regex)
	s = reSlugifySpaces.ReplaceAllString(s, "-")
	s = reSlugifyNonWord.ReplaceAllString(s, "")

	// Remove leading/trailing hyphens
	s = strings.Trim(s, "-")

	return s, nil
}

// PadLeftFilter pads string on the left
func PadLeftFilter(value interface{}, args ...interface{}) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("pad_left filter requires width argument")
	}

	s := ToString(value)
	width, err := ToInt(args[0])
	if err != nil {
		return nil, fmt.Errorf("pad_left width must be integer: %v", err)
	}

	fillchar := " "
	if len(args) > 1 {
		fillchar = ToString(args[1])
		if len(fillchar) == 0 {
			fillchar = " "
		}
	}

	sLen := len([]rune(s))
	if sLen >= width {
		return s, nil
	}

	padding := width - sLen
	return strings.Repeat(fillchar, padding) + s, nil
}

// PadRightFilter pads string on the right
func PadRightFilter(value interface{}, args ...interface{}) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("pad_right filter requires width argument")
	}

	s := ToString(value)
	width, err := ToInt(args[0])
	if err != nil {
		return nil, fmt.Errorf("pad_right width must be integer: %v", err)
	}

	fillchar := " "
	if len(args) > 1 {
		fillchar = ToString(args[1])
		if len(fillchar) == 0 {
			fillchar = " "
		}
	}

	sLen := len([]rune(s))
	if sLen >= width {
		return s, nil
	}

	padding := width - sLen
	return s + strings.Repeat(fillchar, padding), nil
}

// WordcountFilter counts words in a string
func WordcountFilter(value interface{}, args ...interface{}) (interface{}, error) {
	s := ToString(value)
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, nil
	}

	// Split on whitespace and count non-empty parts
	words := strings.Fields(s)
	return len(words), nil
}

// Helper function to extract regex flags from string
func extractRegexFlags(s string) (pattern string, ignoreCase bool) {
	if strings.HasPrefix(s, "(?i)") {
		return s[4:], true
	}
	return s, false
}
