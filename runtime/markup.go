package runtime

import (
	"html"
	"reflect"
	"strings"
	"sync/atomic"
	"unsafe"
)

// Markup is a string marked safe for HTML output, the equivalent of Jinja2's
// markupsafe.Markup. It is only meaningful in the HTML family of escape
// contexts (HTML, XHTML, XML). Its Kind() is reflect.String, so reflection
// based value operations treat it as a string; type switches on string need
// a Markup case.
type Markup string

// String returns the markup content.
func (m Markup) String() string { return string(m) }

// KwArgs carries keyword arguments, as the last argument, to filters that
// declare **kwargs.
type KwArgs map[string]interface{}

// KeepMarkupFunc is a Go global that receives Markup arguments unchanged;
// for functions that only store values.
type KeepMarkupFunc func(args ...interface{}) (interface{}, error)

// AsMarkup reports whether value is marked safe (Markup, SafeValue or a
// non-nil *SafeValue) and returns its content as Markup.
func AsMarkup(value interface{}) (Markup, bool) {
	switch v := value.(type) {
	case Markup:
		return v, true
	case SafeValue:
		return markupOf(v.Value), true
	case *SafeValue:
		if v == nil {
			return "", false
		}
		return markupOf(v.Value), true
	}
	return "", false
}

func markupOf(v interface{}) Markup {
	if m, ok := v.(Markup); ok {
		return m
	}
	return Markup(ToString(v))
}

// Unmark returns the plain string content of a safe value, or value unchanged.
func Unmark(value interface{}) interface{} {
	if m, ok := AsMarkup(value); ok {
		return string(m)
	}
	if t, _, ok := asEscaped(value); ok {
		return t
	}
	return value
}

// escapedJS, escapedCSS, escapedURL and escapedJSON are template output
// (macro, caller, block set, filter block) already escaped for that escape
// context: the counterpart there of Markup, written as is when printed in the
// same context. Like Markup their Kind() is reflect.String. They stay inside
// the runtime: Unmark and Unbox turn them into plain strings for Go code.
type (
	escapedJS   string
	escapedCSS  string
	escapedURL  string
	escapedJSON string
)

// asEscaped returns the text and escape context of escaped output.
func asEscaped(v interface{}) (string, EscapeContext, bool) {
	switch t := v.(type) {
	case escapedJS:
		return string(t), EscapeContextJS, true
	case escapedCSS:
		return string(t), EscapeContextCSS, true
	case escapedURL:
		return string(t), EscapeContextURL, true
	case escapedJSON:
		return string(t), EscapeContextJSON, true
	}
	return "", "", false
}

// escapedOutputMade is set once any evaluator has tagged a value. Until then
// Unbox, which has no evaluator to ask, skips its scan for tagged values.
var escapedOutputMade atomic.Bool

// markEscaped tags s as already escaped for ec; s is returned as is for a
// context without a tag. The evaluator notes that it made escaped output:
// from then on the user's Go filters, tests and functions get the lists and
// dicts of this render scanned for it.
func (e *DefaultEvaluator) markEscaped(s string, ec EscapeContext) interface{} {
	switch ec {
	case EscapeContextJS, EscapeContextCSS, EscapeContextURL, EscapeContextJSON:
		if !e.escapedMade.Load() {
			e.escapedMade.Store(true)
		}
		if !escapedOutputMade.Load() {
			escapedOutputMade.Store(true)
		}
	}
	switch ec {
	case EscapeContextJS:
		return escapedJS(s)
	case EscapeContextCSS:
		return escapedCSS(s)
	case EscapeContextURL:
		return escapedURL(s)
	case EscapeContextJSON:
		return escapedJSON(s)
	}
	return s
}

// EscapeMarkup is markupsafe.escape: safe values pass through, anything else
// is converted to a string and HTML-escaped. It is for Go callers; the
// runtime escapes through escapeFor.
func EscapeMarkup(value interface{}) Markup {
	if m, ok := AsMarkup(value); ok {
		return m
	}
	return Markup(html.EscapeString(ToString(value)))
}

// resolveEscape walks the context chain through Unwrap. The first wrapper
// that states an autoescape setting decides whether escaping is on; the
// escape context and escaper come from the first ContextAwareContext.
func resolveEscape(ctx Context) (on bool, ec EscapeContext, esc *AutoEscaper) {
	decided := false
	for c := ctx; c != nil; {
		switch w := c.(type) {
		case *autoescapeContext:
			if !decided {
				on, decided = w.autoescape, true
			}
		case ContextAwareContext:
			if a := w.GetAutoEscaper(); a != nil {
				if !decided {
					on, decided = a.config.Enabled, true
				}
				return on, w.GetEscapeContext(), a
			}
		case AutoescapeContext:
			if !decided {
				on, decided = w.IsAutoescapeEnabled(), true
			}
		}
		u, ok := c.(interface{ Unwrap() Context })
		if !ok {
			break
		}
		c = u.Unwrap()
	}
	return on, EscapeContextHTML, nil
}

// baseAutoescape is the escape setting of ctx ignoring every {% autoescape %} tag.
func baseAutoescape(ctx Context) bool {
	for c := ctx; c != nil; {
		switch w := c.(type) {
		case *autoescapeContext:
		case *goAutoescapeContext, *goEnvContext:
		case ContextAwareContext:
			if a := w.GetAutoEscaper(); a != nil {
				return a.config.Enabled
			}
		case AutoescapeContext:
			return w.IsAutoescapeEnabled()
		}
		u, ok := c.(interface{ Unwrap() Context })
		if !ok {
			break
		}
		c = u.Unwrap()
	}
	return false
}

// withBaseAutoescape wraps ctx so that the environment's escape setting
// applies, whatever {% autoescape %} tag surrounds it.
func withBaseAutoescape(ctx Context) Context {
	if on, _, _ := resolveEscape(ctx); on != baseAutoescape(ctx) {
		return &autoescapeContext{Context: ctx, autoescape: !on}
	}
	return ctx
}

// markupContext reports whether Markup is meaningful in ctx: escaping is on
// and the escape context is in the HTML family.
func markupContext(ctx Context) bool {
	on, ec, _ := resolveEscape(ctx)
	return on && htmlFamily(ec)
}

// htmlFamily reports whether ec escapes HTML, where Markup is safe as is.
func htmlFamily(ec EscapeContext) bool {
	switch ec {
	case EscapeContextHTML, EscapeContextXHTML, EscapeContextXML:
		return true
	}
	return false
}

// markIfSafe marks rendered template output as Markup when ctx escapes in an
// HTML-family context, and as escaped for the context in any other.
func (e *DefaultEvaluator) markIfSafe(s string, ctx Context) interface{} {
	on, ec, _ := resolveEscape(ctx)
	if !on {
		return s
	}
	if htmlFamily(ec) {
		return Markup(s)
	}
	return e.markEscaped(s, ec)
}

// escapeFor escapes v exactly as {{ }} would in ctx; safe values come back
// as their plain text.
func escapeFor(ctx Context, v interface{}) string {
	if m, ok := AsMarkup(v); ok {
		return string(m)
	}
	on, ec, esc := resolveEscape(ctx)
	if t, tec, ok := asEscaped(v); ok {
		// escaped for another context it is plain text here
		if tec == ec {
			return t
		}
		v = t
	}
	if !on {
		return ToString(v)
	}
	if esc != nil {
		if !esc.config.Enabled {
			// an {% autoescape true %} block switched escaping on
			cfg := *esc.config
			cfg.Enabled = true
			esc = &AutoEscaper{config: &cfg}
		}
		return esc.Escape(v, ec)
	}
	return ToString(escapeOutput(v))
}

// envContext finds the environment behind context wrappers.
func envContext(ctx Context) (EnvironmentContext, bool) {
	env, ok := filterContext(ctx).(EnvironmentContext)
	return env, ok
}

// markupPreservingFilters return Markup when given Markup in Jinja2 (markupsafe
// Markup methods, or slicing/padding of the input). The filter runs on the
// plain string with its arguments escaped, and a string result is Markup.
var markupPreservingFilters = map[string]bool{
	"upper": true, "lower": true, "capitalize": true, "trim": true, "strip": true,
	"replace": true, "center": true, "indent": true,
	"reverse": true, "last": true, "string": true,
	"title": true,
}

// markupAwareFilters receive the safe value unchanged: they implement the
// safe/unsafe distinction themselves, or must return their input as is.
var markupAwareFilters = map[string]bool{
	"safe": true, "marksafe": true, "escape": true, "e": true,
	"urlize": true, "urlizetruncate": true, "urlizetarget": true,
	"nl2br": true, "truncatehtml": true, "xmlattr": true,
}

// markupPlainWhenOff return a plain string when autoescape is off, as Jinja2's
// do_replace, do_title, do_wordwrap, do_join, do_urlize and do_xmlattr do;
// miya's HTML-generating extras follow urlize.
var markupPlainWhenOff = map[string]bool{
	"replace": true, "title": true, "wordwrap": true, "join": true, "xmlattr": true,
	"urlize": true, "urlizetruncate": true, "urlizetarget": true, "nl2br": true, "truncatehtml": true,
}

// escapedKeepingFilters keep escaped output escaped: they cannot split an
// escape sequence (trim and strip only without arguments, and not in CSS,
// where they would drop the space that ends an escape). title is not one:
// its word boundaries would fall inside escape sequences.
var escapedKeepingFilters = map[string]bool{
	"upper": true, "lower": true, "capitalize": true,
	"trim": true, "strip": true, "center": true, "indent": true, "string": true,
}

var stripNames = map[string]bool{"trim": true, "strip": true, "lstrip": true, "rstrip": true}

// keepsEscaped reports whether the filter or method name keeps output
// escaped for tec escaped.
func keepsEscaped(keeping map[string]bool, name string, tec EscapeContext, withArgs bool) bool {
	return keeping[name] && !(stripNames[name] && (withArgs || tec == EscapeContextCSS))
}

var caseFilters = map[string]bool{"upper": true, "lower": true, "capitalize": true}

// keepEscapes undoes a case change inside the escape sequences of JS or JSON
// text: the character after a backslash and the hex digits of \uXXXX and
// \xXX. ok is false when the change did not keep the number of characters.
func keepEscapes(orig, changed string) (string, bool) {
	o, c := []rune(orig), []rune(changed)
	if len(o) != len(c) {
		return "", false
	}
	for i := 0; i < len(o)-1; i++ {
		if o[i] != '\\' {
			continue
		}
		n := 1
		switch o[i+1] {
		case 'u':
			n = 5
		case 'x':
			n = 3
		}
		for j := i + 1; j <= i+n && j < len(o); j++ {
			c[j] = o[j]
		}
		i += n
	}
	return string(c), true
}

// markupPassthroughFilters return either their input or an argument, each
// as it is.
var markupPassthroughFilters = map[string]bool{"default": true, "d": true}

// markupEscapedArgFilters get their arguments escaped for a safe input even
// with autoescape off: markupsafe's Markup.center escapes its fill. Jinja2
// marks indent's string safe; miya escapes it as it does with autoescape on.
// Markup.strip does not escape its chars.
var markupEscapedArgFilters = map[string]bool{"center": true, "indent": true}

// markupMixingFilters receive their input and arguments unchanged, safe or not,
// and implement Jinja2's mixing of safe and plain text themselves.
var markupMixingFilters = map[string]bool{"format": true, "truncate": true, "wordwrap": true}

func isNumberKind(v interface{}) bool {
	switch reflect.ValueOf(v).Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64:
		return true
	}
	return false
}

// unmarkArgs converts safe arguments, also inside a KwArgs, to plain strings
// so Go filters that assert .(string) keep working.
func unmarkArgs(args []interface{}) []interface{} {
	out := make([]interface{}, len(args))
	for i, a := range args {
		if kw, ok := a.(KwArgs); ok {
			nk := make(KwArgs, len(kw))
			for k, v := range kw {
				nk[k] = Unmark(v)
			}
			out[i] = nk
			continue
		}
		out[i] = Unmark(a)
	}
	return out
}

// escapeArgs prepares the arguments of a filter applied to a safe value:
// safe arguments become plain text, numbers, bools and nil stay, everything
// else is escaped for ctx (markupsafe escapes the arguments of Markup methods).
func escapeArgs(ctx Context, args []interface{}) []interface{} {
	out := make([]interface{}, len(args))
	for i, a := range args {
		if kw, ok := a.(KwArgs); ok {
			nk := make(KwArgs, len(kw))
			for k, v := range kw {
				nk[k] = escapeArg(ctx, v)
			}
			out[i] = nk
			continue
		}
		out[i] = escapeArg(ctx, a)
	}
	return out
}

// markupAsText returns args with safe values, also inside a KwArgs, turned
// into plain strings and the rest unchanged.
func markupAsText(args []interface{}) []interface{} {
	out := make([]interface{}, len(args))
	for i, a := range args {
		if kw, ok := a.(KwArgs); ok {
			nk := make(KwArgs, len(kw))
			for k, v := range kw {
				if m, ok := AsMarkup(v); ok {
					v = string(m)
				}
				nk[k] = v
			}
			out[i] = nk
			continue
		}
		if m, ok := AsMarkup(a); ok {
			a = string(m)
		}
		out[i] = a
	}
	return out
}

func escapeArg(ctx Context, a interface{}) interface{} {
	if m, ok := AsMarkup(a); ok {
		return string(m)
	}
	if _, ok := a.(RoundedFloat); ok {
		return a
	}
	if a == nil {
		return a
	}
	if isNumberKind(a) || reflect.ValueOf(a).Kind() == reflect.Bool {
		// a number or bool stays one unless its string form needs escaping
		if s := escapeFor(ctx, a); s != ToString(a) {
			return s
		}
		return a
	}
	return escapeFor(ctx, a)
}

// unmarkCall returns copies of args and kwargs with safe values un-marked,
// for Go functions that expect plain strings; with nested true escaped
// output inside lists and dicts becomes a plain string too.
func unmarkCall(args []interface{}, kwargs map[string]interface{}, nested bool) ([]interface{}, map[string]interface{}) {
	if len(args) > 0 {
		na := make([]interface{}, len(args))
		for i, a := range args {
			na[i] = unmarkCallArg(a, nested)
		}
		args = na
	}
	if len(kwargs) > 0 {
		nk := make(map[string]interface{}, len(kwargs))
		for k, v := range kwargs {
			nk[k] = unmarkCallArg(v, nested)
		}
		kwargs = nk
	}
	return args, kwargs
}

// unmarkCallArg un-marks a Go function argument; with nested true escaped
// output in lists and dicts becomes a plain string too.
func unmarkCallArg(a interface{}, nested bool) interface{} {
	a = Unmark(a)
	if nested {
		return plainEscaped(a)
	}
	return a
}

// builtinLookup is implemented by an environment context that can tell
// miya's own filters and tests from the ones its user registered.
type builtinLookup interface {
	IsBuiltinFilter(name string) bool
	IsBuiltinTest(name string) bool
}

// filtersCallingByName are the built-in filters that call the filter or test
// named by their argument at that position.
var filtersCallingByName = map[string]struct {
	pos  int
	test bool
}{
	"map": {0, false}, "select": {0, true}, "reject": {0, true},
	"selectattr": {1, true}, "rejectattr": {1, true},
}

// builtinFilterCall reports whether filter name with args runs only miya's
// own code, which takes escaped output nested in lists and dicts as the
// strings they are. The evaluator's fallback filters are miya's own.
func builtinFilterCall(ctx Context, name string, args []interface{}) bool {
	fc := filterContext(ctx)
	b, ok := fc.(builtinLookup)
	if !ok {
		_, env := fc.(EnvironmentContext)
		return !env
	}
	if !b.IsBuiltinFilter(name) {
		return false
	}
	by, ok := filtersCallingByName[name]
	if !ok || by.pos >= len(args) {
		return true
	}
	inner, ok := Unmark(args[by.pos]).(string)
	if !ok {
		return true
	}
	if by.test {
		return b.IsBuiltinTest(inner)
	}
	return b.IsBuiltinFilter(inner)
}

// builtinTestCall reports whether test name is miya's own.
func builtinTestCall(ctx Context, name string) bool {
	fc := filterContext(ctx)
	b, ok := fc.(builtinLookup)
	if !ok {
		_, env := fc.(EnvironmentContext)
		return !env
	}
	return b.IsBuiltinTest(name)
}

// plainEscaped returns v with escaped output turned into plain strings, also
// inside nested []interface{} and map[string]interface{} values, for Go
// code. A container is copied only when something in it changes; one that
// contains itself is left as is where it recurs. A value found to hold no
// escaped output is returned without allocating.
func plainEscaped(v interface{}) interface{} {
	budget := plainScanBudget
	if !mayHoldEscaped(v, 0, &budget) {
		return v
	}
	r, _ := plainEscapedIn(v, nil)
	return r
}

// plainItems returns a list with its escaped output items turned into plain
// strings, copied only when one is; other values are returned unchanged.
func plainItems(v interface{}) interface{} {
	list, ok := v.([]interface{})
	if !ok {
		return v
	}
	var out []interface{}
	for i, it := range list {
		if t, _, ok := asEscaped(it); ok {
			if out == nil {
				out = append([]interface{}(nil), list...)
			}
			out[i] = t
		}
	}
	if out == nil {
		return v
	}
	return out
}

// plainScanDepth and plainScanBudget bound the scan of mayHoldEscaped, so a
// value that contains itself or shares containers stops it.
const (
	plainScanDepth  = 256
	plainScanBudget = 1 << 20
)

// mayHoldEscaped reports whether v holds escaped output, also nested; it
// also reports true when the scan reaches its depth or its budget of items.
func mayHoldEscaped(v interface{}, depth int, budget *int) bool {
	if _, _, ok := asEscaped(v); ok {
		return true
	}
	switch c := v.(type) {
	case []interface{}:
		if depth == plainScanDepth {
			return len(c) > 0
		}
		for _, it := range c {
			if *budget--; *budget < 0 || mayHoldEscaped(it, depth+1, budget) {
				return true
			}
		}
	case map[string]interface{}:
		if depth == plainScanDepth {
			return len(c) > 0
		}
		for _, it := range c {
			if *budget--; *budget < 0 || mayHoldEscaped(it, depth+1, budget) {
				return true
			}
		}
	}
	return false
}

// plainArgContainers applies plainEscaped to the list and dict arguments,
// also inside a KwArgs; escaped output given directly as an argument is left
// to the markup policy.
func plainArgContainers(args []interface{}) []interface{} {
	inside := func(v interface{}) (interface{}, bool) {
		if _, _, ok := asEscaped(v); ok {
			return v, false
		}
		budget := plainScanBudget
		if !mayHoldEscaped(v, 0, &budget) {
			return v, false
		}
		return plainEscapedIn(v, nil)
	}
	var out []interface{}
	for i, a := range args {
		n, changed := a, false
		if kw, ok := a.(KwArgs); ok {
			var nk KwArgs
			for k, v := range kw {
				if r, ch := inside(v); ch {
					if nk == nil {
						nk = make(KwArgs, len(kw))
						for k2, v2 := range kw {
							nk[k2] = v2
						}
					}
					nk[k] = r
				}
			}
			n, changed = nk, nk != nil
		} else {
			n, changed = inside(a)
		}
		if changed {
			if out == nil {
				out = append([]interface{}(nil), args...)
			}
			out[i] = n
		}
	}
	if out == nil {
		return args
	}
	return out
}

// plainWalk holds the containers on the current path of a plainEscaped
// walk, and the result for each container already walked: a value that
// shares containers is walked once per container.
type plainWalk struct {
	path map[reprKey]bool
	done map[reprKey]plainResult
}

type plainResult struct {
	v       interface{}
	changed bool
}

func plainEscapedIn(v interface{}, w *plainWalk) (interface{}, bool) {
	if t, _, ok := asEscaped(v); ok {
		return t, true
	}
	var key reprKey
	switch c := v.(type) {
	case []interface{}:
		if len(c) == 0 {
			return v, false
		}
		key = reprKey{uintptr(unsafe.Pointer(&c[0])), len(c)}
	case map[string]interface{}:
		if len(c) == 0 {
			return v, false
		}
		key = reprKey{ptr: reflect.ValueOf(c).Pointer()}
	default:
		return v, false
	}
	if w == nil {
		w = &plainWalk{path: map[reprKey]bool{}, done: map[reprKey]plainResult{}}
	}
	if r, ok := w.done[key]; ok {
		return r.v, r.changed
	}
	if w.path[key] {
		return v, false
	}
	w.path[key] = true
	r := plainResult{v: v}
	switch c := v.(type) {
	case []interface{}:
		var out []interface{}
		for i, it := range c {
			if n, changed := plainEscapedIn(it, w); changed {
				if out == nil {
					out = append([]interface{}(nil), c...)
				}
				out[i] = n
			}
		}
		if out != nil {
			r = plainResult{out, true}
		}
	case map[string]interface{}:
		var out map[string]interface{}
		for k, it := range c {
			if n, changed := plainEscapedIn(it, w); changed {
				if out == nil {
					out = make(map[string]interface{}, len(c))
					for k2, v2 := range c {
						out[k2] = v2
					}
				}
				out[k] = n
			}
		}
		if out != nil {
			r = plainResult{out, true}
		}
	}
	delete(w.path, key)
	w.done[key] = r
	return r.v, r.changed
}

// applyFilterWithMarkup applies the markup policy around a filter call.
func (e *DefaultEvaluator) applyFilterWithMarkup(name string, value interface{}, args []interface{}, ctx Context,
	apply func(interface{}, []interface{}) (interface{}, error)) (interface{}, error) {
	// Markup stays Markup in the HTML family whether or not autoescape is on,
	// as markupsafe's does; only a few filters check autoescape themselves.
	on, ec, _ := resolveEscape(ctx)
	html := htmlFamily(ec)
	keep := html && (on || !markupPlainWhenOff[name])

	if t, tec, ok := asEscaped(value); ok {
		switch {
		case name == "safe" || name == "marksafe":
			return value, nil
		case markupPassthroughFilters[name]:
			r, err := e.applyFilterWithMarkup(name, t, args, ctx, apply)
			if s, ok := r.(string); err == nil && ok && s == t {
				// the input came back
				return value, nil
			}
			return r, err
		case tec == ec && keepsEscaped(escapedKeepingFilters, name, tec, len(args) > 0):
			// the arguments are escaped for the text's context even where
			// autoescape is off, as markupsafe does for Markup
			argCtx := ctx
			if !on {
				argCtx = &autoescapeContext{Context: ctx, autoescape: true}
			}
			// a safe argument is HTML-safe, not safe for tec
			r, err := apply(t, escapeArgs(argCtx, markupAsText(args)))
			if err != nil {
				return nil, err
			}
			if s, ok := r.(string); ok {
				if caseFilters[name] && (tec == EscapeContextJS || tec == EscapeContextJSON) {
					if s, ok = keepEscapes(t, s); !ok {
						return r, nil
					}
				}
				return e.markEscaped(s, tec), nil
			}
			return r, nil
		}
		// any other filter gets the text, and its result is escaped again
		value = t
	}
	// The user's Go filters never see escaped output, even nested
	if e.escapedMade.Load() && !builtinFilterCall(ctx, name, args) {
		value = plainEscaped(value)
		args = plainArgContainers(args)
	} else if name == "map" && !html {
		// map gives each item to a filter, which takes escaped output as
		// its text as it does when applied to it directly
		value = plainItems(value)
	}
	m, safe := AsMarkup(value)

	var result interface{}
	var err error
	switch {
	case name == "join" && markupContext(ctx):
		result, err = e.joinMarkup(ctx, value, args, apply)
	case markupMixingFilters[name] && keep:
		result, err = apply(value, args)
	case name == "map" && !safe && html:
		result, err = e.mapMarkup(ctx, value, args, apply)
	case markupPassthroughFilters[name] && !safe:
		result, err = apply(value, args)
	case markupPassthroughFilters[name]:
		result, err = apply(string(m), args)
		if s, ok := result.(string); ok && s == string(m) {
			// the input came back
			result = m
		}
	case !safe && name == "replace" && markupContext(ctx) && isReplaceInput(value) && anySafeArg(args):
		result, err = apply(escapeFor(ctx, value), replaceArgs(ctx, args))
		if s, ok := result.(string); ok {
			result = Markup(s)
		}
	case !safe:
		result, err = apply(value, unmarkArgs(args))
	case markupAwareFilters[name]:
		result, err = apply(m, unmarkArgs(args))
	case markupPreservingFilters[name] && keep:
		// with autoescape off the arguments stay as they are, except where
		// a markupsafe method escapes them
		argCtx := ctx
		if !on && markupEscapedArgFilters[name] {
			argCtx = &autoescapeContext{Context: ctx, autoescape: true}
		}
		result, err = apply(string(m), escapeArgs(argCtx, args))
		if s, ok := result.(string); ok {
			result = Markup(s)
		}
	default:
		result, err = apply(string(m), unmarkArgs(args))
	}
	if err != nil {
		return nil, err
	}
	// Normalise SafeValue results (third-party filters) to Markup so later
	// value operations see a string.
	if rm, ok := AsMarkup(result); ok {
		result = rm
		// Outside the HTML family only safe and marksafe stay safe: the
		// context escaper must run for everything else.
		if !keep && name != "safe" && name != "marksafe" {
			return string(rm), nil
		}
	} else if list, ok := result.([]interface{}); ok && !keep && name != "safe" && name != "marksafe" {
		// map('e') and friends: Markup items must not reach a non-HTML context raw.
		// Only lists are walked; maps are left alone.
		result = unmarkItems(list, nil)
	}
	return result, nil
}

// isStringValue reports whether v is a string or a named string type.
func isStringValue(v interface{}) bool {
	_, ok := basicValue(v).(string)
	return ok
}

// isReplaceInput reports whether v is a string, number, RoundedFloat or bool.
func isReplaceInput(v interface{}) bool {
	if _, ok := v.(RoundedFloat); ok {
		return true
	}
	b := basicValue(v)
	if _, ok := b.(bool); ok {
		return true
	}
	return isStringValue(b) || isNumberKind(b)
}

// FilterLookup is implemented by contexts that can tell whether a filter is
// registered.
type FilterLookup interface {
	HasFilter(name string) bool
}

// filterLookup finds a FilterLookup behind context wrappers.
func filterLookup(ctx Context) (FilterLookup, bool) {
	fl, ok := filterContext(ctx).(FilterLookup)
	return fl, ok
}

// mapItemReappliable reports whether no attribute of the filter's name can
// win over the filter for this item, so the markup policy may be re-applied.
// A slice or array takes a name whose first segment is all digits as an index.
func mapItemReappliable(it interface{}, name string) bool {
	switch it.(type) {
	case GroupTuple, []byte:
		return false
	case RoundedFloat:
		return true
	case []interface{}:
		return !digitSegment(name)
	}
	switch b := basicValue(it).(type) {
	case string, nil, *Undefined, bool:
		return true
	default:
		if isNumberKind(b) {
			return true
		}
		switch reflect.ValueOf(b).Kind() {
		case reflect.Slice, reflect.Array:
			return !digitSegment(name)
		}
	}
	return false
}

// digitSegment reports whether the first '.'-separated segment of name is all digits.
func digitSegment(name string) bool {
	seg, _, _ := strings.Cut(name, ".")
	if seg == "" {
		return false
	}
	for _, c := range seg {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// anySafeArg reports whether any argument, including the values of a trailing
// KwArgs, is safe.
func anySafeArg(args []interface{}) bool {
	for _, a := range args {
		if kw, ok := a.(KwArgs); ok {
			for _, v := range kw {
				if _, ok := AsMarkup(v); ok {
					return true
				}
			}
		} else if _, ok := AsMarkup(a); ok {
			return true
		}
	}
	return false
}

// replaceArgs prepares replace's arguments as markupsafe's Markup.replace gets
// them: old is searched as given, new is escaped.
func replaceArgs(ctx Context, args []interface{}) []interface{} {
	out := make([]interface{}, len(args))
	for i, a := range args {
		if kw, ok := a.(KwArgs); ok {
			nk := make(KwArgs, len(kw))
			for k, v := range kw {
				if k == "new" {
					nk[k] = escapeArg(ctx, v)
				} else {
					nk[k] = Unmark(v)
				}
			}
			out[i] = nk
		} else if i == 1 {
			out[i] = escapeArg(ctx, a)
		} else {
			out[i] = Unmark(a)
		}
	}
	return out
}

// keepDefault un-marks every argument except the value of a KwArgs "default".
func keepDefault(args []interface{}) []interface{} {
	out := unmarkArgs(args)
	for i, a := range args {
		if kw, ok := a.(KwArgs); ok {
			if d, ok := kw["default"]; ok {
				out[i].(KwArgs)["default"] = d
			}
		}
	}
	return out
}

// mapMarkup applies the markup policy to each item of map(filter) and keeps
// the safe default of map(attribute=, default=). The filter runs once per item:
// items whose markup policy is re-applied are filtered one by one, the others
// together through apply.
func (e *DefaultEvaluator) mapMarkup(ctx Context, value interface{}, args []interface{}, apply func(interface{}, []interface{}) (interface{}, error)) (interface{}, error) {
	whole := func() (interface{}, error) { return apply(value, keepDefault(args)) }
	if len(args) == 0 {
		return whole()
	}
	if _, kw := args[0].(KwArgs); kw {
		return whole()
	}
	inner := ToString(args[0])
	lookup, haveLookup := filterLookup(ctx)
	if !haveLookup {
		return e.mapMarkupWhole(ctx, value, args, inner, apply)
	}
	if !lookup.HasFilter(inner) {
		return whole()
	}
	items, err := e.makeIterable(value)
	if err != nil {
		return whole()
	}
	again := make([]bool, len(items))
	var rest []interface{}
	anyAgain := false
	for i, it := range items {
		_, safe := AsMarkup(it)
		if safe || mapItemReappliable(it, inner) {
			again[i] = true
			anyAgain = true
		} else {
			rest = append(rest, it)
		}
	}
	if !anyAgain {
		return whole()
	}
	out := make([]interface{}, len(items))
	if len(rest) > 0 {
		res, err := apply(rest, keepDefault(args))
		if err != nil {
			return nil, err
		}
		list, ok := res.([]interface{})
		if !ok || len(list) != len(rest) {
			return whole()
		}
		j := 0
		for i := range items {
			if !again[i] {
				out[i] = list[j]
				j++
			}
		}
	}
	innerApply := e.mapInnerApply(ctx, inner)
	for i, it := range items {
		if !again[i] {
			continue
		}
		r, err := e.applyFilterWithMarkup(inner, it, args[1:], ctx, innerApply)
		if err != nil {
			return nil, err
		}
		out[i] = r
	}
	return out, nil
}

// mapInnerApply applies the filter named inner to a single item of map(filter).
func (e *DefaultEvaluator) mapInnerApply(ctx Context, inner string) func(interface{}, []interface{}) (interface{}, error) {
	return func(v interface{}, a []interface{}) (interface{}, error) {
		env, ok := envContext(ctx)
		if !ok {
			return e.applyFilter(inner, v, a)
		}
		if n := len(a); n > 0 {
			if kw, ok := a[n-1].(KwArgs); ok {
				if applier, ok := env.(KwargsFilterApplier); ok {
					return applier.ApplyFilterKwargs(inner, v, a[:n-1], map[string]interface{}(kw))
				}
			}
		}
		return env.ApplyFilter(inner, v, a...)
	}
}

// mapMarkupWhole is mapMarkup for contexts that cannot tell whether a filter
// is registered: the filter runs over the whole list, then again for each
// item the markup policy applies to.
func (e *DefaultEvaluator) mapMarkupWhole(ctx Context, value interface{}, args []interface{}, inner string, apply func(interface{}, []interface{}) (interface{}, error)) (interface{}, error) {
	result, err := apply(value, keepDefault(args))
	if err != nil {
		return nil, err
	}
	list, ok := result.([]interface{})
	if !ok {
		return result, nil
	}
	items, err := e.makeIterable(value)
	if err != nil || len(items) != len(list) {
		return result, nil
	}
	innerApply := e.mapInnerApply(ctx, inner)
	out := list
	for i, it := range items {
		if _, undef := list[i].(*Undefined); undef {
			continue
		}
		_, safe := AsMarkup(it)
		if !safe && !mapItemReappliable(it, inner) {
			continue
		}
		r, err := e.applyFilterWithMarkup(inner, it, args[1:], ctx, innerApply)
		if err != nil {
			return nil, err
		}
		if &out[0] == &list[0] {
			out = append([]interface{}(nil), list...)
		}
		out[i] = r
	}
	return out, nil
}

// joinMarkup is Jinja2's do_join under autoescape: if any item or the
// separator is safe, escape the rest and return Markup.
func (e *DefaultEvaluator) joinMarkup(ctx Context, value interface{}, args []interface{}, apply func(interface{}, []interface{}) (interface{}, error)) (interface{}, error) {
	plain := func() (interface{}, error) { return apply(value, unmarkArgs(args)) }
	var sep interface{} = ""
	var attr interface{}
	var trailingKw bool
	if n := len(args); n > 0 {
		pos := args
		if kw, ok := args[n-1].(KwArgs); ok {
			trailingKw = true
			pos = args[:n-1]
			for k := range kw {
				if k != "d" && k != "attribute" {
					return plain()
				}
			}
			if d, ok := kw["d"]; ok {
				if len(pos) > 0 {
					return plain()
				}
				sep = d
			}
			attr = kw["attribute"]
			if _, ok := kw["attribute"]; ok && len(pos) > 1 {
				return plain()
			}
		}
		if len(pos) > 2 || (len(pos) == 2 && attr != nil) {
			return plain()
		}
		if len(pos) > 0 {
			sep = pos[0]
		}
		if len(pos) == 2 {
			attr = pos[1]
		}
	}
	items, err := e.makeIterable(value)
	if err != nil {
		return plain()
	}
	if attr != nil {
		applier, ok := filterContext(ctx).(KwargsFilterApplier)
		if !ok {
			return plain()
		}
		mapped, err := applier.ApplyFilterKwargs("map", value, nil, map[string]interface{}{"attribute": Unmark(attr)})
		if err != nil {
			return plain()
		}
		list, ok := mapped.([]interface{})
		if !ok {
			return plain()
		}
		items = list
	}
	anySafe := false
	for _, it := range items {
		if _, ok := AsMarkup(it); ok {
			anySafe = true
			break
		}
	}
	if _, ok := AsMarkup(sep); ok {
		anySafe = true
	}
	if !anySafe {
		return plain()
	}
	escaped := make([]interface{}, len(items))
	for i, it := range items {
		escaped[i] = escapeFor(ctx, it)
	}
	shaped := []interface{}{}
	if len(args) > 0 {
		shaped = append(shaped, escapeArg(ctx, sep))
	}
	if trailingKw {
		shaped = append(shaped, KwArgs{})
	}
	r, err := apply(escaped, shaped)
	if err != nil {
		return nil, err
	}
	if s, ok := r.(string); ok {
		return Markup(s), nil
	}
	return r, nil
}

// concatMarkup implements ~ and + when an operand is safe: the other operand
// is escaped and the result is Markup.
func concatMarkup(ctx Context, left, right interface{}) (Markup, bool) {
	_, ls := AsMarkup(left)
	_, rs := AsMarkup(right)
	if !ls && !rs {
		return "", false
	}
	return Markup(escapeFor(ctx, left) + escapeFor(ctx, right)), true
}

// concatEscaped implements ~ and + when an operand is output escaped for the
// active escape context: the other operand, safe or not, is escaped for the
// context and the result stays escaped.
func (e *DefaultEvaluator) concatEscaped(ctx Context, left, right interface{}) (interface{}, bool) {
	on, ec, _ := resolveEscape(ctx)
	_, lec, ls := asEscaped(left)
	_, rec, rs := asEscaped(right)
	if !on || !(ls && lec == ec || rs && rec == ec) {
		return nil, false
	}
	// a safe value is HTML-safe, not safe for ec
	if !ls || lec != ec {
		left = Unmark(left)
	}
	if !rs || rec != ec {
		right = Unmark(right)
	}
	return e.markEscaped(escapeFor(ctx, left)+escapeFor(ctx, right), ec), true
}

// unmarkItems returns a copy of list with Markup elements and escaped output
// turned into plain strings, recursing into nested lists. It returns list
// itself when nothing changes.
func unmarkItems(list []interface{}, path map[reprKey]bool) []interface{} {
	return unmarkList(list, path, nil)
}

// unmarkList is unmarkItems with done holding the result for each nested list
// already un-marked, so a list shared many times is walked once.
func unmarkList(list []interface{}, path map[reprKey]bool, done map[reprKey][]interface{}) []interface{} {
	var out []interface{}
	for i, it := range list {
		var n interface{}
		if sub, ok := it.([]interface{}); ok {
			if len(sub) == 0 {
				continue
			}
			// path holds the lists being un-marked, the top-level one
			// included: one that contains itself is left as is where it
			// recurs
			if path == nil {
				path = map[reprKey]bool{{uintptr(unsafe.Pointer(&list[0])), len(list)}: true}
			}
			if done == nil {
				done = map[reprKey][]interface{}{}
			}
			key := reprKey{uintptr(unsafe.Pointer(&sub[0])), len(sub)}
			if path[key] {
				continue
			}
			r, seen := done[key]
			if !seen {
				path[key] = true
				r = unmarkList(sub, path, done)
				delete(path, key)
				done[key] = r
			}
			if len(r) == 0 || &r[0] == &sub[0] {
				continue
			}
			n = r
		} else if t, _, ok := asEscaped(it); ok {
			n = t
		} else if m, ok := AsMarkup(it); ok {
			n = string(m)
		} else {
			continue
		}
		if out == nil {
			out = append([]interface{}(nil), list...)
		}
		out[i] = n
	}
	if out == nil {
		return list
	}
	return out
}
