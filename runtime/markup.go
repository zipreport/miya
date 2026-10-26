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

// markupPassthroughFilters return either their input or an argument.
var markupPassthroughFilters = map[string]bool{"default": true, "d": true}

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
	case markupMixingFilters[name] && markupContext(ctx):
		result, err = apply(value, args)
	case name == "map" && !safe && markupContext(ctx):
		result, err = e.mapMarkup(ctx, value, args, apply)
	case !safe && markupPassthroughFilters[name]:
		result, err = apply(value, args)
	case !safe && name == "replace" && markupContext(ctx) && isStringValue(value) && anySafeArg(args):
		result, err = apply(escapeFor(ctx, value), replaceArgs(ctx, args))
		if s, ok := result.(string); ok {
			result = Markup(s)
		}
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
	} else if list, ok := result.([]interface{}); ok && !markupContext(ctx) && name != "safe" && name != "marksafe" {
		// map('e') and friends: Markup items must not reach a non-HTML context raw.
		// Only lists are walked; maps are left alone.
		result = unmarkItems(list)
	}
	return result, nil
}

// isStringValue reports whether v is a string or a named string type.
func isStringValue(v interface{}) bool {
	_, ok := basicValue(v).(string)
	return ok
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
// the safe default of map(attribute=, default=).
func (e *DefaultEvaluator) mapMarkup(ctx Context, value interface{}, args []interface{}, apply func(interface{}, []interface{}) (interface{}, error)) (interface{}, error) {
	result, err := apply(value, keepDefault(args))
	if err != nil {
		return nil, err
	}
	list, ok := result.([]interface{})
	if !ok || len(args) == 0 {
		return result, nil
	}
	if _, kw := args[0].(KwArgs); kw {
		return result, nil
	}
	items, err := e.makeIterable(value)
	if err != nil || len(items) != len(list) {
		return result, nil
	}
	inner := ToString(args[0])
	innerArgs := args[1:]
	innerApply := func(v interface{}, a []interface{}) (interface{}, error) {
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
	out := list
	for i, it := range items {
		if _, undef := list[i].(*Undefined); undef {
			continue
		}
		_, safe := AsMarkup(it)
		if !safe {
			switch b := basicValue(it).(type) {
			case string, nil, *Undefined:
				_ = b
			default:
				continue
			}
		}
		r, err := e.applyFilterWithMarkup(inner, it, innerArgs, ctx, innerApply)
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
		if attr != nil {
			if _, undef := it.(*Undefined); undef || it == nil {
				escaped[i] = ""
				continue
			}
		}
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

// unmarkItems returns a copy of list with Markup elements turned into plain
// strings, recursing into nested lists. It returns list itself when nothing changes.
func unmarkItems(list []interface{}) []interface{} {
	var out []interface{}
	for i, it := range list {
		var n interface{}
		if sub, ok := it.([]interface{}); ok {
			r := unmarkItems(sub)
			if len(r) == 0 || len(sub) == 0 || &r[0] == &sub[0] {
				continue
			}
			n = r
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
