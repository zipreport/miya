package filters

import (
	"fmt"
	"html"
	"math"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/zipreport/miya/runtime"
)

// Pre-compiled regex patterns for HTML filters (performance optimization)
var (
	reURLs = regexp.MustCompile(`https?://[^\s<>"']+`)

	// Jinja2's urlize patterns; \w is Unicode-aware there and \d matches any decimal digit.
	reUrlizeHTTP   = regexp.MustCompile(`(?i)^(?:(?:https?://|www\.)(?:(?:[\p{L}\p{N}_%-]+\.)+)?(?:[a-z]{2,63}|xn--[\p{L}\p{N}_%]{2,59})|(?:[\p{L}\p{N}_%-]{2,63}\.)+(?:com|net|int|edu|gov|org|info|mil)|https?://(?:\p{Nd}{1,3}(?:\.\p{Nd}{1,3}){3}|\[(?:[\p{Nd}a-f]{0,4}:){2}(?:[\p{Nd}a-f]{0,4}:?){1,6}\]))(?::\p{Nd}{1,5})?(?:[/?#]\S*)?$`)
	reUrlizeEmail  = regexp.MustCompile(`^\S+@[\p{L}\p{N}_][\p{L}\p{N}_.-]*\.[\p{L}\p{N}_]+$`)
	reUrlizeScheme = regexp.MustCompile(`^[\p{L}\p{N}_.+-]{2,}:/{0,2}$`)
	reUrlizeSplit  = regexp.MustCompile("(" + pySpaceClass + "+)")
)

// SafeValue marks a value that must not be escaped. It is runtime.SafeValue,
// so the output path recognises values from either package.
type SafeValue = runtime.SafeValue

// EscapeFilter escapes HTML characters
func EscapeFilter(value interface{}, args ...interface{}) (interface{}, error) {
	// Don't escape if already marked as safe
	if m, ok := runtime.AsMarkup(value); ok {
		return m, nil
	}

	s := ToString(value)
	return runtime.Markup(html.EscapeString(s)), nil
}

// SafeFilter marks a value as safe (won't be escaped)
func SafeFilter(value interface{}, args ...interface{}) (interface{}, error) {
	if m, ok := runtime.AsMarkup(value); ok {
		return m, nil
	}
	return runtime.Markup(ToString(value)), nil
}

// urlQuote percent-encodes the UTF-8 bytes of s like Jinja2's url_quote,
// keeping A-Za-z0-9_.-~ (and / unless forQS, where a space becomes +).
func urlQuote(s string, forQS bool) string {
	const hex = "0123456789ABCDEF"
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c >= 'A' && c <= 'Z', c >= 'a' && c <= 'z', c >= '0' && c <= '9',
			c == '_', c == '.', c == '-', c == '~', c == '/' && !forQS:
			b.WriteByte(c)
		case c == ' ' && forQS:
			b.WriteByte('+')
		default:
			b.WriteByte('%')
			b.WriteByte(hex[c>>4])
			b.WriteByte(hex[c&15])
		}
	}
	return b.String()
}

// URLEncodeFilter percent-encodes a string, or turns a mapping or a list of
// pairs into a query string, like Jinja2.
func URLEncodeFilter(value interface{}, args ...interface{}) (interface{}, error) {
	if m, ok := runtime.AsMarkup(value); ok {
		return urlQuote(string(m), false), nil
	}
	rv := reflect.ValueOf(runtime.Indirect(value))
	var pairs [][2]interface{}
	switch rv.Kind() {
	case reflect.Map:
		type kv struct {
			k string
			v interface{}
		}
		items := make([]kv, 0, rv.Len())
		iter := rv.MapRange()
		for iter.Next() {
			items = append(items, kv{ToString(iter.Key().Interface()), iter.Value().Interface()})
		}
		sort.Slice(items, func(i, j int) bool { return items[i].k < items[j].k })
		for _, it := range items {
			pairs = append(pairs, [2]interface{}{it.k, it.v})
		}
	case reflect.Slice, reflect.Array:
		if rv.Type().Elem().Kind() == reflect.Uint8 {
			return urlQuote(ToString(value), false), nil
		}
		for i := 0; i < rv.Len(); i++ {
			el := rv.Index(i).Interface()
			ev := reflect.ValueOf(runtime.Indirect(el))
			if (ev.Kind() != reflect.Slice && ev.Kind() != reflect.Array) || ev.Len() != 2 {
				return nil, fmt.Errorf("urlencode: expected a string, a mapping or a list of pairs, got element %T", el)
			}
			pairs = append(pairs, [2]interface{}{ev.Index(0).Interface(), ev.Index(1).Interface()})
		}
	default:
		return urlQuote(ToString(value), false), nil
	}
	parts := make([]string, len(pairs))
	for i, p := range pairs {
		parts[i] = urlQuote(ToString(p[0]), true) + "=" + urlQuote(ToString(p[1]), true)
	}
	return strings.Join(parts, "&"), nil
}

// XMLAttrFilter formats attributes for XML/HTML
func XMLAttrFilter(value interface{}, args ...interface{}) (interface{}, error) {
	value = dictInput(value)
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
			if strings.ContainsAny(key, " \t\n\r\f\v/>=") {
				return nil, fmt.Errorf("invalid character in attribute name: %q", key)
			}
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
				if _, safe := runtime.AsMarkup(val); safe {
					result.WriteString(attrValue)
				} else {
					result.WriteString(html.EscapeString(attrValue))
				}
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
			if strings.ContainsAny(key, " \t\n\r\f\v/>=") {
				return nil, fmt.Errorf("invalid character in attribute name: %q", key)
			}
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

	return runtime.Markup(" " + result.String()), nil
}

// stripBetween removes every open...closing span, stopping when no closing
// follows the next open. It matches repeated find-and-cut from the start of
// the string (a removal can join text into a new open), in one linear pass:
// out holds the text that contains no open, and a new open can only begin in
// its last len(open)-1 bytes.
func stripBetween(s, open, closing string) string {
	out := make([]byte, 0, len(s))
	pos := 0
	for {
		rest := s[pos:]
		start, consumed := -1, 0 // start: offset in out; consumed: bytes of rest the open uses
		for o := max(0, len(out)-(len(open)-1)); o < len(out); o++ {
			n := len(out) - o
			if string(out[o:]) == open[:n] && strings.HasPrefix(rest, open[n:]) {
				start, consumed = o, len(open)-n
				break
			}
		}
		if start < 0 {
			i := strings.Index(rest, open)
			if i < 0 {
				return string(out) + rest
			}
			out = append(out, rest[:i]...)
			start, consumed, rest, pos = len(out), len(open), rest[i:], pos+i
		}
		j := strings.Index(rest[consumed:], closing)
		if j < 0 {
			return string(out) + rest
		}
		out = out[:start]
		pos += consumed + j + len(closing)
	}
}

// StripTagsFilter removes HTML/XML tags and comments, collapses whitespace and
// decodes entities, like markupsafe's Markup.striptags.
func StripTagsFilter(value interface{}, args ...interface{}) (interface{}, error) {
	s := stripBetween(ToString(value), "<!--", "-->")
	s = stripBetween(s, "<", ">")
	s = strings.Join(strings.FieldsFunc(s, isPySpace), " ")
	return html.UnescapeString(s), nil
}

// isSafe reports whether value is already marked safe.
func isSafe(value interface{}) bool {
	_, ok := runtime.AsMarkup(value)
	return ok
}

// linkify replaces URLs in s with link(url); the text between URLs is
// HTML-escaped unless it is already safe.
func linkify(s string, safe bool, link func(string) string) string {
	var b strings.Builder
	last := 0
	gap := func(g string) {
		if safe {
			b.WriteString(g)
		} else {
			b.WriteString(html.EscapeString(g))
		}
	}
	for _, m := range reURLs.FindAllStringIndex(s, -1) {
		gap(s[last:m[0]])
		b.WriteString(link(s[m[0]:m[1]]))
		last = m[1]
	}
	gap(s[last:])
	return b.String()
}

// urlizeTrim shortens x to limit runes plus "..." when it is longer, with
// Python's slice rules for a negative limit.
func urlizeTrim(x string, limit *int) string {
	if limit == nil {
		return x
	}
	r := []rune(x)
	if len(r) <= *limit {
		return x
	}
	end := *limit
	if end < 0 {
		end += len(r)
		if end < 0 {
			end = 0
		}
	}
	return string(r[:end]) + "..."
}

var (
	urlizeHeads = []string{"(", "<", "&lt;"}
	urlizeTails = []string{")", ">", ".", ",", "\n", "&gt;"}
	urlizePairs = [][2]string{{"(", ")"}, {"<", ">"}, {"&lt;", "&gt;"}}
)

// stripRun removes a leading (or trailing) run of any of the tokens from s.
func stripRun(s string, tokens []string, leading bool) (rest, run string) {
	if leading {
		i := 0
		for {
			matched := false
			for _, t := range tokens {
				if strings.HasPrefix(s[i:], t) {
					i += len(t)
					matched = true
					break
				}
			}
			if !matched {
				return s[i:], s[:i]
			}
		}
	}
	j := len(s)
	for {
		matched := false
		for _, t := range tokens {
			if strings.HasSuffix(s[:j], t) {
				j -= len(t)
				matched = true
				break
			}
		}
		if !matched {
			return s[:j], s[j:]
		}
	}
}

// UrlizeFilter converts URLs in text to clickable links like Jinja2:
// urlize(trim_url_limit=None, nofollow=False, target=None, rel=None, extra_schemes=None).
func UrlizeFilter(value interface{}, args ...interface{}) (interface{}, error) {
	arg := func(i int) interface{} {
		if i < len(args) {
			return args[i]
		}
		return nil
	}

	var limit *int
	switch v := arg(0).(type) {
	case nil:
	case bool:
		n := 0
		if v {
			n = 1
		}
		limit = &n
	default:
		n, err := ToInt(v)
		if err != nil {
			return nil, fmt.Errorf("urlize: trim_url_limit must be an integer: %v", err)
		}
		limit = &n
	}

	relWords := map[string]bool{"noopener": true}
	if r := arg(3); r != nil {
		for _, w := range strings.FieldsFunc(ToString(r), isPySpace) {
			relWords[w] = true
		}
	}
	if ToBool(arg(1)) {
		relWords["nofollow"] = true
	}
	sortedRel := make([]string, 0, len(relWords))
	for w := range relWords {
		sortedRel = append(sortedRel, w)
	}
	sort.Strings(sortedRel)
	attrs := ` rel="` + html.EscapeString(strings.Join(sortedRel, " ")) + `"`
	if t := arg(2); t != nil {
		if ts := ToString(t); ts != "" {
			attrs += ` target="` + html.EscapeString(ts) + `"`
		}
	}

	var schemes []string
	if es := arg(4); es != nil {
		rv := reflect.ValueOf(runtime.Indirect(es))
		if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
			return nil, fmt.Errorf("urlize: extra_schemes must be a list")
		}
		for i := 0; i < rv.Len(); i++ {
			sc := ToString(rv.Index(i).Interface())
			if !reUrlizeScheme.MatchString(sc) {
				return nil, fmt.Errorf("urlize: invalid extra scheme %q", sc)
			}
			schemes = append(schemes, sc)
		}
	}

	text := ToString(value)
	if !isSafe(value) {
		text = html.EscapeString(text)
	}

	var out strings.Builder
	process := func(word string) string {
		middle, head := stripRun(word, urlizeHeads, true)
		middle, tail := stripRun(middle, urlizeTails, false)
		for _, p := range urlizePairs {
			open, closing := p[0], p[1]
			startCount := strings.Count(middle, open)
			if startCount <= strings.Count(middle, closing) {
				continue
			}
			cut := 0
			for n := minInt(startCount, strings.Count(tail, closing)); n > 0; n-- {
				i := strings.Index(tail[cut:], closing)
				if i < 0 {
					break
				}
				cut += i + len(closing)
			}
			middle += tail[:cut]
			tail = tail[cut:]
		}

		var link string
		switch {
		case reUrlizeHTTP.MatchString(middle):
			href := middle
			if !strings.HasPrefix(href, "http://") && !strings.HasPrefix(href, "https://") {
				href = "https://" + href
			}
			link = `<a href="` + href + `"` + attrs + `>` + urlizeTrim(middle, limit) + `</a>`
		case strings.HasPrefix(middle, "mailto:") && reUrlizeEmail.MatchString(middle[len("mailto:"):]):
			link = `<a href="` + middle + `">` + middle[len("mailto:"):] + `</a>`
		case strings.Contains(middle, "@") && !strings.HasPrefix(middle, "www.") && !strings.HasPrefix(middle, "@") &&
			!strings.Contains(middle, ":") && reUrlizeEmail.MatchString(middle):
			link = `<a href="mailto:` + middle + `">` + middle + `</a>`
		default:
			for _, sc := range schemes {
				if strings.HasPrefix(middle, sc) && middle != sc {
					link = `<a href="` + middle + `"` + attrs + `>` + middle + `</a>`
					break
				}
			}
		}
		if link == "" {
			return word
		}
		return head + link + tail
	}

	last := 0
	for _, m := range reUrlizeSplit.FindAllStringIndex(text, -1) {
		out.WriteString(process(text[last:m[0]]))
		out.WriteString(text[m[0]:m[1]])
		last = m[1]
	}
	out.WriteString(process(text[last:]))
	return runtime.Markup(out.String()), nil
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// UrlizeTargetFilter is like urlize but with a specific target
func UrlizeTargetFilter(value interface{}, args ...interface{}) (interface{}, error) {
	target := "_blank"
	if len(args) > 0 {
		target = ToString(args[0])
	}

	var rel interface{}
	if len(args) > 1 {
		rel = args[1]
	}
	return UrlizeFilter(value, nil, false, target, rel)
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
	if length < 0 {
		return nil, fmt.Errorf("truncatehtml length must not be negative")
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
		if isSafe(value) {
			return value, nil
		}
		return s, nil
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

	return runtime.Markup(html.EscapeString(truncated) + html.EscapeString(end)), nil
}

// FileSizeFormatFilter formats file sizes in human-readable format, following
// Jinja2's do_filesizeformat.
func FileSizeFormatFilter(value interface{}, args ...interface{}) (interface{}, error) {
	var size float64
	if str, ok := value.(string); ok {
		f, parsed := pyParseFloat(str)
		if !parsed {
			return nil, fmt.Errorf("filesizeformat requires numeric value: cannot convert %q to float", str)
		}
		size = f
	} else {
		f, err := ToFloat(value)
		if err != nil {
			return nil, fmt.Errorf("filesizeformat requires numeric value: %v", err)
		}
		size = f
	}

	binary := len(args) > 0 && ToBool(args[0])

	base := 1000.0
	prefixes := []string{"kB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"}
	if binary {
		base = 1024
		prefixes = []string{"KiB", "MiB", "GiB", "TiB", "PiB", "EiB", "ZiB", "YiB"}
	}

	if size == 1 {
		return "1 Byte", nil
	}
	if size < base {
		return fmt.Sprintf("%d Bytes", int64(size)), nil
	}
	var unit float64
	prefix := ""
	for i, p := range prefixes {
		unit = math.Pow(base, float64(i+2))
		prefix = p
		if size < unit {
			break
		}
	}
	return fmt.Sprintf("%.1f %s", base*size/unit, prefix), nil
}

// AutoEscapeFilter conditionally escapes based on context
//
// Deprecated: use EscapeFilter; the "autoescape" name cannot be used in templates.
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
	// ToString reads the content of a safe value
	s := ToString(value)
	// Return as Markup to prevent double-escaping
	return runtime.Markup(html.EscapeString(s)), nil
}

// NL2BRFilter converts newlines to HTML <br> tags
func NL2BRFilter(value interface{}, args ...interface{}) (interface{}, error) {
	s := ToString(value)
	if !isSafe(value) {
		s = html.EscapeString(s)
	}

	// Replace different types of line breaks
	result := strings.ReplaceAll(s, "\r\n", "<br>")   // Windows
	result = strings.ReplaceAll(result, "\r", "<br>") // Mac
	result = strings.ReplaceAll(result, "\n", "<br>") // Unix

	return runtime.Markup(result), nil
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
	safe := isSafe(value)
	// Safe input is already escaped, as in urlize.
	escape := html.EscapeString
	if safe {
		escape = func(s string) string { return s }
	}

	// Use pre-compiled URL regex (same as urlize)
	link := func(match string) string {
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
		attrs = append(attrs, fmt.Sprintf(`href="%s"`, escape(match)))

		if target != "" {
			attrs = append(attrs, fmt.Sprintf(`target="%s"`, html.EscapeString(target)))
		}

		if rel != "" {
			attrs = append(attrs, fmt.Sprintf(`rel="%s"`, html.EscapeString(rel)))
		}

		return fmt.Sprintf(`<a %s>%s</a>`, strings.Join(attrs, " "), escape(displayURL))
	}

	return runtime.Markup(linkify(s, safe, link)), nil
}
