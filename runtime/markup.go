package runtime

import (
	"html"
	"reflect"
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
	return value
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

// markupContext reports whether Markup is meaningful in ctx: escaping is on
// and the escape context is in the HTML family.
func markupContext(ctx Context) bool {
	on, ec, _ := resolveEscape(ctx)
	if !on {
		return false
	}
	switch ec {
	case EscapeContextHTML, EscapeContextXHTML, EscapeContextXML:
		return true
	}
	return false
}

// markIfSafe marks rendered template output as Markup when ctx escapes in an
// HTML-family context.
func markIfSafe(s string, ctx Context) interface{} {
	if markupContext(ctx) {
		return Markup(s)
	}
	return s
}

// escapeFor escapes v exactly as {{ }} would in ctx; safe values come back
// as their plain text.
func escapeFor(ctx Context, v interface{}) string {
	if m, ok := AsMarkup(v); ok {
		return string(m)
	}
	on, ec, esc := resolveEscape(ctx)
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
// wordwrap is a miya choice; Jinja2 returns a plain string there.
var markupPreservingFilters = map[string]bool{
	"upper": true, "lower": true, "capitalize": true, "trim": true, "strip": true,
	"replace": true, "center": true, "indent": true, "truncate": true,
	"reverse": true, "last": true, "string": true,
	"format": true, "title": true, "wordwrap": true,
}

// markupAwareFilters receive the safe value unchanged: they implement the
// safe/unsafe distinction themselves, or must return their input as is.
var markupAwareFilters = map[string]bool{
	"safe": true, "marksafe": true, "escape": true, "e": true,
	"urlize": true, "urlizetruncate": true, "urlizetarget": true,
	"nl2br": true, "truncatehtml": true, "xmlattr": true,
}

// markupPassthroughFilters return either their input or an argument.
var markupPassthroughFilters = map[string]bool{"default": true, "d": true}

// markupArgFilters get an unsafe input and their arguments unchanged, Markup
// included, because they treat a safe argument themselves.
var markupArgFilters = map[string]bool{"wordwrap": true}

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

func escapeArg(ctx Context, a interface{}) interface{} {
	if m, ok := AsMarkup(a); ok {
		return string(m)
	}
	if a == nil || isNumberKind(a) || reflect.ValueOf(a).Kind() == reflect.Bool {
		return a
	}
	return escapeFor(ctx, a)
}

// unmarkCall returns copies of args and kwargs with safe values un-marked,
// for Go functions that expect plain strings.
func unmarkCall(args []interface{}, kwargs map[string]interface{}) ([]interface{}, map[string]interface{}) {
	if len(args) > 0 {
		na := make([]interface{}, len(args))
		for i, a := range args {
			na[i] = Unmark(a)
		}
		args = na
	}
	if len(kwargs) > 0 {
		nk := make(map[string]interface{}, len(kwargs))
		for k, v := range kwargs {
			nk[k] = Unmark(v)
		}
		kwargs = nk
	}
	return args, kwargs
}

// applyFilterWithMarkup applies the markup policy around a filter call.
func (e *DefaultEvaluator) applyFilterWithMarkup(name string, value interface{}, args []interface{}, ctx Context,
	apply func(interface{}, []interface{}) (interface{}, error)) (interface{}, error) {
	m, safe := AsMarkup(value)

	var result interface{}
	var err error
	switch {
	case name == "join" && markupContext(ctx):
		result, err = e.joinMarkup(ctx, value, args, apply)
	case !safe && markupArgFilters[name]:
		result, err = apply(value, args)
	case !safe:
		result, err = apply(value, unmarkArgs(args))
	case markupAwareFilters[name]:
		result, err = apply(m, unmarkArgs(args))
	case markupPassthroughFilters[name], markupPreservingFilters[name]:
		result, err = apply(string(m), escapeArgs(ctx, args))
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
		if !markupContext(ctx) && name != "safe" && name != "marksafe" {
			return string(rm), nil
		}
	}
	return result, nil
}

// joinMarkup is Jinja2's do_join under autoescape: if any item or the
// separator is safe, escape the rest and return Markup.
func (e *DefaultEvaluator) joinMarkup(ctx Context, value interface{}, args []interface{}, apply func(interface{}, []interface{}) (interface{}, error)) (interface{}, error) {
	plain := func() (interface{}, error) { return apply(value, unmarkArgs(args)) }
	// join(sep, attribute) reads attributes of the items; leave it to the filter
	if len(args) > 1 {
		return plain()
	}
	if len(args) == 1 {
		if _, kw := args[0].(KwArgs); kw {
			return plain()
		}
	}
	items, err := e.makeIterable(value)
	if err != nil {
		return plain()
	}
	anySafe := false
	for _, it := range items {
		if _, ok := AsMarkup(it); ok {
			anySafe = true
			break
		}
	}
	if len(args) > 0 {
		if _, ok := AsMarkup(args[0]); ok {
			anySafe = true
		}
	}
	if !anySafe {
		return plain()
	}
	escaped := make([]interface{}, len(items))
	for i, it := range items {
		escaped[i] = escapeFor(ctx, it)
	}
	r, err := apply(escaped, escapeArgs(ctx, args))
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
