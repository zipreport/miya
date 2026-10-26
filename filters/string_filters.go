package filters

import (
	"fmt"
	"math"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/zipreport/miya/runtime"
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
	if len(args) > 0 && args[0] != nil {
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
	return truncate(5, value, args...)
}

// TruncateFilterWithLeeway returns a truncate filter whose default leeway is
// the given value (Jinja2's truncate.leeway policy).
func TruncateFilterWithLeeway(leeway int) FilterFunc {
	return func(value interface{}, args ...interface{}) (interface{}, error) {
		return truncate(leeway, value, args...)
	}
}

// truncate follows Jinja2's do_truncate(s, length=255, killwords=False,
// end="...", leeway=5).
func truncate(defaultLeeway int, value interface{}, args ...interface{}) (interface{}, error) {
	s := ToString(value)

	length := 255
	if len(args) > 0 && args[0] != nil {
		l, err := ToInt(args[0])
		if err != nil {
			return nil, fmt.Errorf("truncate length must be integer: %v", err)
		}
		length = l
	}
	killwords := len(args) > 1 && ToBool(args[1])
	end := "..."
	if len(args) > 2 {
		end = ToString(args[2])
	}
	leeway := defaultLeeway
	if len(args) > 3 && args[3] != nil {
		l, err := ToInt(args[3])
		if err != nil {
			return nil, fmt.Errorf("truncate leeway must be integer: %v", err)
		}
		leeway = l
	}

	endRunes := []rune(end)
	if length < len(endRunes) {
		return nil, fmt.Errorf("truncate: expected length >= %d, got %d", len(endRunes), length)
	}
	if leeway < 0 {
		return nil, fmt.Errorf("truncate: expected leeway >= 0, got %d", leeway)
	}

	runes := []rune(s)
	if len(runes) <= length+leeway {
		return s, nil
	}

	prefix := string(runes[:length-len(endRunes)])
	if killwords {
		return prefix + end, nil
	}
	if i := strings.LastIndex(prefix, " "); i >= 0 {
		prefix = prefix[:i]
	}
	return prefix + end, nil
}

// WordwrapFilter wraps text like Jinja2's wordwrap, which uses Python's
// textwrap: (width=79, break_long_words=true, wrapstring="\n", break_on_hyphens=true).
func WordwrapFilter(value interface{}, args ...interface{}) (interface{}, error) {
	arg := func(i int) interface{} {
		if i < len(args) {
			return args[i]
		}
		return nil
	}
	width := 79
	if arg(0) != nil {
		w, err := ToInt(arg(0))
		if err != nil {
			return nil, fmt.Errorf("wordwrap width must be integer: %v", err)
		}
		width = w
	}
	breakLongWords := arg(1) == nil || ToBool(arg(1))
	breakOnHyphens := arg(3) == nil || ToBool(arg(3))

	wrapMarkup, safeWrap := runtime.AsMarkup(arg(2))
	wrapString := "\n"
	if safeWrap {
		wrapString = string(wrapMarkup)
	} else if arg(2) != nil {
		wrapString = ToString(arg(2))
	}

	input := pySplitlines(ToString(value))
	if len(input) == 0 {
		return "", nil
	}
	if width <= 0 {
		return nil, fmt.Errorf("wordwrap: invalid width %d (must be > 0)", width)
	}
	var out []string
	for _, line := range input {
		wrapped := textwrapLines(line, width, breakLongWords, breakOnHyphens)
		if safeWrap {
			for i, l := range wrapped {
				wrapped[i] = string(runtime.EscapeMarkup(l))
			}
		}
		out = append(out, strings.Join(wrapped, wrapString))
	}
	result := strings.Join(out, wrapString)
	if safeWrap {
		return runtime.Markup(result), nil
	}
	return result, nil
}

// maxRepeatBytes bounds the padding a template can ask a filter to build.
const maxRepeatBytes = 64 << 20

func repeatBounded(s string, count int) (string, error) {
	if count <= 0 {
		return "", nil
	}
	if len(s) > 0 && count > maxRepeatBytes/len(s) {
		return "", fmt.Errorf("filter output too large")
	}
	return strings.Repeat(s, count), nil
}

// CenterFilter centers string in field of given width
func CenterFilter(value interface{}, args ...interface{}) (interface{}, error) {
	s := ToString(value)
	width := 80
	if len(args) > 0 && args[0] != nil {
		w, err := ToInt(args[0])
		if err != nil {
			return nil, fmt.Errorf("center width must be integer: %v", err)
		}
		width = w
	}

	fillchar := " "
	if len(args) > 1 {
		fillchar = ToString(args[1])
		if len(fillchar) == 0 {
			fillchar = " "
		}
	}

	marg := width - len([]rune(s))
	if marg <= 0 {
		return s, nil
	}
	// Python's str.center puts the odd space on the left when width is odd.
	leftPad := marg/2 + (marg & width & 1)

	pad, err := repeatBounded(fillchar, marg)
	if err != nil {
		return nil, err
	}
	split := leftPad * len(fillchar)
	return pad[:split] + s + pad[split:], nil
}

// IndentFilter indents each line, following Jinja2's do_indent(s, width=4,
// first=False, blank=False).
func IndentFilter(value interface{}, args ...interface{}) (interface{}, error) {
	s := ToString(value)

	var indention string
	if len(args) > 0 && args[0] != nil {
		if str, ok := args[0].(string); ok {
			indention = str
		} else {
			width, err := ToInt(args[0])
			if err != nil {
				return nil, fmt.Errorf("indent width must be integer: %v", err)
			}
			indention, err = repeatBounded(" ", width)
			if err != nil {
				return nil, err
			}
		}
	} else {
		indention = "    "
	}
	first := len(args) > 1 && ToBool(args[1])
	blank := len(args) > 2 && ToBool(args[2])

	lines := strings.Split(strings.NewReplacer("\r\n", "\n", "\r", "\n").Replace(s+"\n"), "\n")
	lines = lines[:len(lines)-1]
	if len(indention) > 0 && len(lines) > maxRepeatBytes/len(indention) {
		return nil, fmt.Errorf("indent output too large")
	}

	var rv string
	if blank {
		rv = strings.Join(lines, "\n"+indention)
	} else {
		rv = lines[0]
		if len(lines) > 1 {
			rest := make([]string, len(lines)-1)
			for i, line := range lines[1:] {
				if line != "" {
					line = indention + line
				}
				rest[i] = line
			}
			rv += "\n" + strings.Join(rest, "\n")
		}
	}
	if first {
		rv = indention + rv
	}
	return rv, nil
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
	s, args = pyIntVerbs(s, args)
	return fmt.Sprintf(s, args...), nil
}

// pyIntVerbs makes %d and %i accept float operands as Python's % operator
// does (truncating like int()), and rewrites %i, which Go lacks, to %d.
// Other verbs and their arguments are left unchanged. A format using Go's
// explicit argument indexes (%[1]d) is returned untouched.
func pyIntVerbs(format string, args []interface{}) (string, []interface{}) {
	if strings.Contains(format, "%[") {
		return format, args
	}
	out := append([]interface{}(nil), args...)
	var b strings.Builder
	argi := 0
	for i := 0; i < len(format); i++ {
		if format[i] != '%' {
			b.WriteByte(format[i])
			continue
		}
		j := i + 1
		for j < len(format) && strings.IndexByte("+-# 0123456789.*", format[j]) >= 0 {
			if format[j] == '*' {
				argi++
			}
			j++
		}
		if j >= len(format) {
			b.WriteString(format[i:])
			break
		}
		verb := format[j]
		b.WriteString(format[i:j])
		i = j
		if verb == '%' {
			b.WriteByte('%')
			continue
		}
		if verb == 'i' {
			verb = 'd'
		}
		if verb == 'd' && argi < len(out) {
			var f float64
			isFloat := true
			switch v := out[argi].(type) {
			case float64:
				f = v
			case float32:
				f = float64(v)
			default:
				isFloat = false
			}
			if isFloat && !math.IsNaN(f) && math.Abs(f) < 9e18 {
				out[argi] = int64(f)
			}
		}
		b.WriteByte(verb)
		argi++
	}
	return b.String(), out
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
	pad, err := repeatBounded(fillchar, padding)
	if err != nil {
		return nil, err
	}
	return pad + s, nil
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
	pad, err := repeatBounded(fillchar, padding)
	if err != nil {
		return nil, err
	}
	return s + pad, nil
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
