package filters

import (
	"fmt"
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

// reTitleSep splits on Jinja2's title separators: '-', Python whitespace and
// the opening brackets ( { [ <.
var reTitleSep = regexp.MustCompile(`[-\p{Z}\t-\r\x{1c}-\x{1f}\x{85}({\[<]+`)

// reWord matches a Unicode-aware \w+ run, as counted by Jinja2's wordcount.
var reWord = regexp.MustCompile(`[\p{L}\p{N}_]+`)

// titleCase converts a string to title case like Jinja2: the first rune of each
// piece between separators is upper-cased and the rest lower-cased.
func titleCase(s string) string {
	var b strings.Builder
	last := 0
	word := func(w string) {
		if w == "" {
			return
		}
		r, size := utf8.DecodeRuneInString(w)
		b.WriteString(strings.ToUpper(string(r)))
		b.WriteString(strings.ToLower(w[size:]))
	}
	for _, m := range reTitleSep.FindAllStringIndex(s, -1) {
		word(s[last:m[0]])
		b.WriteString(s[m[0]:m[1]])
		last = m[1]
	}
	word(s[last:])
	return b.String()
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
	_, valSafe := runtime.AsMarkup(value)

	length := 255
	if len(args) > 0 && args[0] != nil {
		l, err := ToInt(runtime.Unmark(args[0]))
		if err != nil {
			return nil, fmt.Errorf("truncate length must be integer: %v", err)
		}
		length = l
	}
	killwords := len(args) > 1 && ToBool(args[1])
	end := "..."
	endSafe := false
	if len(args) > 2 {
		end = ToString(args[2])
		_, endSafe = runtime.AsMarkup(args[2])
	}
	leeway := defaultLeeway
	if len(args) > 3 && args[3] != nil {
		l, err := ToInt(runtime.Unmark(args[3]))
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
		if valSafe {
			return value, nil
		}
		return s, nil
	}

	prefix := string(runes[:length-len(endRunes)])
	if !killwords {
		if i := strings.LastIndex(prefix, " "); i >= 0 {
			prefix = prefix[:i]
		}
	}
	if valSafe || endSafe {
		part := func(text string, safe bool) string {
			if safe {
				return text
			}
			return string(runtime.EscapeMarkup(text))
		}
		return runtime.Markup(part(prefix, valSafe) + part(end, endSafe)), nil
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

	_, inSafe := runtime.AsMarkup(value)
	// A safe input with a plain, non-blank wrapstring is escaped as a whole
	// by the caller, so the result must be plain (Jinja2's result).
	plainResult := inSafe && !safeWrap && strings.TrimSpace(wrapString) != ""

	input := pySplitlines(ToString(value))
	if len(input) == 0 {
		if inSafe && !plainResult {
			return runtime.Markup(""), nil
		}
		return "", nil
	}
	if width <= 0 {
		return nil, fmt.Errorf("wordwrap: invalid width %d (must be > 0)", width)
	}
	var out []string
	for _, line := range input {
		wrapped := textwrapLines(line, width, breakLongWords, breakOnHyphens)
		if safeWrap && !inSafe {
			for i, l := range wrapped {
				wrapped[i] = string(runtime.EscapeMarkup(l))
			}
		}
		out = append(out, strings.Join(wrapped, wrapString))
	}
	result := strings.Join(out, wrapString)
	if (safeWrap || inSafe) && !plainResult {
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
		return "", &runtime.SizeError{Msg: "filter output too large"}
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
		return nil, &runtime.SizeError{Msg: "indent output too large"}
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

// FormatFilter formats the string with Python's % operator: positional
// arguments, or keyword arguments for %(name)s conversions.
func FormatFilter(value interface{}, args ...interface{}) (interface{}, error) {
	var kwargs map[string]interface{}
	if n := len(args); n > 0 {
		if kw, ok := args[n-1].(KwArgs); ok {
			kwargs = map[string]interface{}(kw)
			args = args[:n-1]
		}
	}
	if len(kwargs) > 0 && len(args) > 0 {
		return nil, fmt.Errorf("format: can't handle positional and keyword arguments at the same time")
	}
	fm, safe := runtime.AsMarkup(value)
	format := ToString(value)
	if safe {
		format = string(fm)
	} else {
		args = unmarkAll(args)
		if len(kwargs) > 0 {
			plain := make(map[string]interface{}, len(kwargs))
			for k, v := range kwargs {
				plain[k] = runtime.Unmark(v)
			}
			kwargs = plain
		}
	}
	out, err := pyFormat(format, args, kwargs, safe)
	if err != nil {
		return nil, err
	}
	if safe {
		return runtime.Markup(out), nil
	}
	return out, nil
}

func unmarkAll(args []interface{}) []interface{} {
	out := make([]interface{}, len(args))
	for i, a := range args {
		out[i] = runtime.Unmark(a)
	}
	return out
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
	return len(reWord.FindAllString(ToString(value), -1)), nil
}

// Helper function to extract regex flags from string
func extractRegexFlags(s string) (pattern string, ignoreCase bool) {
	if strings.HasPrefix(s, "(?i)") {
		return s[4:], true
	}
	return s, false
}
