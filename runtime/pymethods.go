package runtime

import (
	"fmt"
	"html"
	"math"
	"reflect"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/zipreport/miya/parser"
)

// maxStringResult bounds strings built by the Python string methods and by
// string repetition (same bound as maxRepeatBytes in package filters, plan
// 030, which is unexported there).
const maxStringResult = 64 << 20

// maxListRepeat bounds the items of a repeated list.
const maxListRepeat = 1000000

// builtinMethod is a Python builtin method bound to a template value
// ('abc'.upper, l.append). It is a pointer to a struct, not a func, so
// makeIterable never calls it implicitly; it prints as <built-in method NAME>.
type builtinMethod struct {
	name string
	fn   func(args []interface{}, kwargs map[string]interface{}) (interface{}, error)
}

func (m *builtinMethod) String() string { return "<built-in method " + m.name + ">" }

// listBox holds a list stored by {% set %}, {% with %} or a namespace
// (Decision 5 in plan 048). Until the first append/extend/pop it only
// references orig, so reading it returns the original value unchanged;
// the first write copies the items into owned, so the application's data
// is never written.
type listBox struct {
	orig  interface{}   // the bound slice or array, until the first write
	items []interface{} // the box's own copy, after the first write
	owned bool
}

func isBoxable(v interface{}) bool {
	switch v.(type) {
	case nil, []byte:
		return false
	case []interface{}:
		return true
	}
	k := reflect.TypeOf(v).Kind()
	return k == reflect.Slice || k == reflect.Array
}

// BoxList wraps a slice or array (except []byte) in a new listBox; other
// values are returned unchanged. The root package's namespace() uses it.
func BoxList(v interface{}) interface{} {
	if isBoxable(v) {
		return &listBox{orig: v}
	}
	return v
}

// Unbox returns the value a template variable holds (for code that reads
// context variables directly): orig or items for a listBox, any other
// value unchanged, except that escaped template output, also inside lists
// and dicts, becomes a plain string.
func Unbox(v interface{}) interface{} {
	return plainEscaped(unbox(v))
}

// unbox is Unbox for the runtime itself: escaped output keeps its type.
func unbox(v interface{}) interface{} {
	if b, ok := v.(*listBox); ok {
		return b.view()
	}
	return v
}

func (b *listBox) view() interface{} {
	if b.owned {
		return b.items
	}
	return b.orig
}

// own copies the bound value into the box's own items on the first write.
func (b *listBox) own() {
	if b.owned {
		return
	}
	var items []interface{}
	if l, ok := b.orig.([]interface{}); ok {
		items = make([]interface{}, len(l), len(l)+1)
		copy(items, l)
	} else {
		rv := reflect.ValueOf(b.orig)
		items = make([]interface{}, rv.Len(), rv.Len()+1)
		for i := range items {
			items[i] = rv.Index(i).Interface()
		}
	}
	b.items, b.orig, b.owned = items, nil, true
}

func (b *listBox) list(e *DefaultEvaluator) ([]interface{}, error) {
	if b.owned {
		return b.items, nil
	}
	return e.makeIterable(b.orig)
}

const notStoredHint = "%s() can only change a list stored with {%% set %%}, {%% with %%} or namespace(); store it first, e.g. {%% set items = items|list %%}"

func isListMethodName(name string) bool {
	switch name {
	case "append", "extend", "pop", "index", "count":
		return true
	}
	return false
}

func isStringMethodName(name string) bool {
	switch name {
	case "upper", "lower", "strip", "lstrip", "rstrip", "split", "replace", "startswith",
		"endswith", "format", "title", "capitalize", "join", "count", "find":
		return true
	}
	return false
}

// method returns the list method name bound to the box.
func (b *listBox) method(e *DefaultEvaluator, name string) *builtinMethod {
	fn := func(args []interface{}, kwargs map[string]interface{}) (interface{}, error) {
		if len(kwargs) > 0 {
			return nil, fmt.Errorf("%s() takes no keyword arguments", name)
		}
		switch name {
		case "append":
			if len(args) != 1 {
				return nil, fmt.Errorf("append() takes exactly one argument (%d given)", len(args))
			}
			b.own()
			if err := checkListGrowth(len(b.items), 1); err != nil {
				return nil, err
			}
			b.items = append(b.items, args[0])
			return nil, nil
		case "extend":
			if len(args) != 1 {
				return nil, fmt.Errorf("extend() takes exactly one argument (%d given)", len(args))
			}
			more, err := e.makeIterable(args[0])
			if err != nil {
				return nil, fmt.Errorf("extend() argument is not iterable: %w", err)
			}
			b.own()
			if err := checkListGrowth(len(b.items), len(more)); err != nil {
				return nil, err
			}
			b.items = append(b.items, more...)
			return nil, nil
		case "pop":
			if len(args) > 1 {
				return nil, fmt.Errorf("pop() takes at most 1 argument (%d given)", len(args))
			}
			idx := -1
			if len(args) == 1 {
				n, err := countArg("pop", args[0])
				if err != nil {
					return nil, err
				}
				idx = n
			}
			b.own()
			n := len(b.items)
			if n == 0 {
				return nil, fmt.Errorf("pop from empty list")
			}
			if idx < 0 {
				idx += n
			}
			if idx < 0 || idx >= n {
				return nil, fmt.Errorf("pop index out of range")
			}
			v := b.items[idx]
			if idx == n-1 {
				// the capacity cut makes the next append reallocate, so a
				// slice header that escaped earlier never sees a later write
				b.items = b.items[: n-1 : n-1]
			} else {
				b.items = append(append([]interface{}(nil), b.items[:idx]...), b.items[idx+1:]...)
			}
			return v, nil
		default:
			items, err := b.list(e)
			if err != nil {
				return nil, err
			}
			return listSearch(e, name, items, args)
		}
	}
	return &builtinMethod{name: name, fn: fn}
}

func checkListGrowth(have, add int) error {
	if have+add > maxConcatLength {
		return &SizeError{Msg: fmt.Sprintf("list too large: %d items (limit %d)", have+add, maxConcatLength)}
	}
	return nil
}

func listSearch(e *DefaultEvaluator, name string, items []interface{}, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("%s() takes exactly one argument (%d given)", name, len(args))
	}
	if name == "index" {
		for i, it := range items {
			if e.equal(it, args[0]) {
				return i, nil
			}
		}
		return nil, fmt.Errorf("%s is not in list", reprOf(args[0]))
	}
	n := 0
	for _, it := range items {
		if e.equal(it, args[0]) {
			n++
		}
	}
	return n, nil
}

// isPureChain reports whether node is a variable or an attribute chain rooted
// at one, so that evaluating it has no side effects.
func isPureChain(node parser.ExpressionNode) bool {
	for {
		switch n := node.(type) {
		case *parser.IdentifierNode:
			return true
		case *parser.AttributeNode:
			node = n.Object
		default:
			return false
		}
	}
}

// rawGetter is implemented by namespaces that can return a stored value
// without unboxing it.
type rawGetter interface {
	RawGet(name string) (interface{}, bool)
}

// nsRawGet returns the raw (possibly boxed) value of a namespace attribute.
func nsRawGet(ns NamespaceInterface, name string) (interface{}, bool) {
	if r, ok := ns.(rawGetter); ok {
		return r.RawGet(name)
	}
	return ns.Get(name)
}

// rawOperand returns the raw (possibly boxed) value of a variable or of a
// namespace attribute, without evaluating anything with side effects.
func (e *DefaultEvaluator) rawOperand(node parser.ExpressionNode, ctx Context) (interface{}, bool) {
	if !isPureChain(node) {
		return nil, false
	}
	switch n := node.(type) {
	case *parser.IdentifierNode:
		return ctx.GetVariable(n.Name)
	case *parser.AttributeNode:
		obj, err := e.EvalNode(n.Object, ctx)
		if err != nil {
			return nil, false
		}
		if ns, ok := obj.(NamespaceInterface); ok {
			return nsRawGet(ns, n.Attribute)
		}
	}
	return nil, false
}

// callArg evaluates a macro argument or default. A stored list keeps its box,
// so the macro can change the caller's list; anything else is evaluated once.
func (e *DefaultEvaluator) callArg(expr parser.ExpressionNode, ctx Context) (interface{}, error) {
	if raw, ok := e.rawOperand(expr, ctx); ok {
		if b, ok := raw.(*listBox); ok {
			return b, nil
		}
	}
	return e.EvalNode(expr, ctx)
}

// bindArg prepares a macro parameter: a list the macro may change.
func bindArg(v interface{}) interface{} {
	if b, ok := v.(*listBox); ok {
		return b
	}
	return BoxList(v)
}

// bindValue prepares the value stored by {% set %}, {% with %} or a
// namespace attribute assignment: a list becomes a listBox, and binding a
// variable that already holds a box shares that box.
func (e *DefaultEvaluator) bindValue(expr parser.ExpressionNode, value interface{}, ctx Context) interface{} {
	if expr != nil {
		if raw, ok := e.rawOperand(expr, ctx); ok {
			if b, ok := raw.(*listBox); ok {
				return b
			}
		}
	}
	return BoxList(value)
}

// builtinMethodFor returns the builtin method name of a string or list value.
// Dicts are handled in getAttribute's map case, after their keys. ctx may be
// nil when only the method's existence matters.
func (e *DefaultEvaluator) builtinMethodFor(obj interface{}, name string, ctx Context) (*builtinMethod, bool) {
	isStr, isList := isStringMethodName(name), isListMethodName(name)
	if obj == nil || (!isStr && !isList) {
		return nil, false
	}
	if isStr {
		if t, tec, ok := asEscaped(obj); ok {
			return e.escapedMethod(t, tec, name, ctx), true
		}
		if m, ok := AsMarkup(obj); ok {
			// Outside the HTML family a safe receiver is plain text: the
			// context escaper must run on the method's result, even when
			// autoescape is off where the method is called.
			_, ec, _ := resolveEscape(ctx)
			return e.stringMethod(string(m), htmlFamily(ec), name), true
		}
		if s, ok := basicValue(obj).(string); ok {
			return e.stringMethod(s, false, name), true
		}
	}
	if isList {
		v := Indirect(obj)
		if v == nil {
			return nil, false
		}
		if _, ok := v.([]byte); ok {
			return nil, false
		}
		if !isBoxable(v) {
			return nil, false
		}
		if name == "index" || name == "count" {
			return &builtinMethod{name: name, fn: func(args []interface{}, kwargs map[string]interface{}) (interface{}, error) {
				if len(kwargs) > 0 {
					return nil, fmt.Errorf("%s() takes no keyword arguments", name)
				}
				items, err := e.makeIterable(v)
				if err != nil {
					return nil, err
				}
				return listSearch(e, name, items, args)
			}}, true
		}
		return &builtinMethod{name: name, fn: func(args []interface{}, kwargs map[string]interface{}) (interface{}, error) {
			return nil, fmt.Errorf(notStoredHint, name)
		}}, true
	}
	return nil, false
}

// escapedMethod is a string method of escaped output t: it works on the
// text, and the case and whitespace-strip methods called without arguments
// in t's escape context keep the result escaped, as the filters do.
func (e *DefaultEvaluator) escapedMethod(t string, tec EscapeContext, name string, ctx Context) *builtinMethod {
	m := e.stringMethod(t, false, name)
	if _, ec, _ := resolveEscape(ctx); tec != ec || !keepsEscaped(escapedKeepingMethods, name, tec, false) {
		return m
	}
	return &builtinMethod{name: name, fn: func(args []interface{}, kwargs map[string]interface{}) (interface{}, error) {
		r, err := m.fn(args, kwargs)
		s, isStr := r.(string)
		if err != nil || !isStr || len(args) > 0 {
			return r, err
		}
		if caseFilters[name] && (tec == EscapeContextJS || tec == EscapeContextJSON) {
			if s, isStr = keepEscapes(t, s); !isStr {
				return r, nil
			}
		}
		return markEscaped(s, tec), nil
	}}
}

var escapedKeepingMethods = map[string]bool{
	"upper": true, "lower": true, "capitalize": true,
	"strip": true, "lstrip": true, "rstrip": true,
}

// mapGetMethod is d.get(key, default=None).
func mapGetMethod(m map[string]interface{}) *builtinMethod {
	return &builtinMethod{name: "get", fn: func(args []interface{}, kwargs map[string]interface{}) (interface{}, error) {
		if len(kwargs) > 0 {
			return nil, fmt.Errorf("get() takes no keyword arguments")
		}
		if len(args) < 1 || len(args) > 2 {
			return nil, fmt.Errorf("get() takes 1 or 2 arguments (%d given)", len(args))
		}
		if v, ok := m[DictKey(args[0])]; ok {
			return v, nil
		}
		if len(args) == 2 {
			return args[1], nil
		}
		return nil, nil
	}}
}

func stringMap(m map[string]string) map[string]interface{} {
	out := make(map[string]interface{}, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

// ---- argument helpers

func strArg(name string, v interface{}) (string, error) {
	if m, ok := AsMarkup(v); ok {
		return string(m), nil
	}
	if s, ok := basicValue(v).(string); ok {
		return s, nil
	}
	return "", fmt.Errorf("%s() argument must be str, not %T", name, v)
}

// textArg is strArg, HTML-escaping plain strings for a safe receiver.
func textArg(name string, v interface{}, esc bool) (string, error) {
	if m, ok := AsMarkup(v); ok {
		return string(m), nil
	}
	s, err := strArg(name, v)
	if err != nil {
		return "", err
	}
	if esc {
		s = html.EscapeString(s)
	}
	return s, nil
}

func countArg(name string, v interface{}) (int, error) {
	i, u, f, kind, ok := numericValue(v)
	if ok {
		switch kind {
		case 'i':
			return clampInt(i), nil
		case 'u':
			if u > math.MaxInt64 {
				return math.MaxInt, nil
			}
			return clampInt(int64(u)), nil
		case 'f':
			if f == math.Trunc(f) && !math.IsInf(f, 0) {
				if f > math.MaxInt64 {
					return math.MaxInt, nil
				}
				if f < math.MinInt64 {
					return math.MinInt, nil
				}
				return clampInt(int64(f)), nil
			}
		}
	}
	return 0, fmt.Errorf("%s() argument must be an integer, not %T", name, v)
}

func clampInt(i int64) int {
	if i > math.MaxInt {
		return math.MaxInt
	}
	if i < math.MinInt {
		return math.MinInt
	}
	return int(i)
}

func noKwargs(name string, kwargs map[string]interface{}) error {
	if len(kwargs) > 0 {
		return fmt.Errorf("%s() takes no keyword arguments", name)
	}
	return nil
}

func argCount(name string, args []interface{}, min, max int) error {
	if len(args) < min || len(args) > max {
		if min == max {
			return fmt.Errorf("%s() takes exactly %d argument(s) (%d given)", name, min, len(args))
		}
		return fmt.Errorf("%s() takes %d to %d arguments (%d given)", name, min, max, len(args))
	}
	return nil
}

// ---- string methods

// stringMethod binds a string method to s. esc is set for a Markup receiver:
// string results become Markup, and format, replace and join escape the
// arguments that are not themselves Markup.
func (e *DefaultEvaluator) stringMethod(s string, esc bool, name string) *builtinMethod {
	wrap := func(r string) interface{} {
		if esc {
			return Markup(r)
		}
		return r
	}
	fn := func(args []interface{}, kwargs map[string]interface{}) (interface{}, error) {
		if name != "format" {
			if err := noKwargs(name, kwargs); err != nil {
				return nil, err
			}
		}
		switch name {
		case "upper":
			if err := argCount(name, args, 0, 0); err != nil {
				return nil, err
			}
			return wrap(strings.ToUpper(s)), nil
		case "lower":
			if err := argCount(name, args, 0, 0); err != nil {
				return nil, err
			}
			return wrap(strings.ToLower(s)), nil
		case "title":
			if err := argCount(name, args, 0, 0); err != nil {
				return nil, err
			}
			return wrap(pyTitle(s)), nil
		case "capitalize":
			if err := argCount(name, args, 0, 0); err != nil {
				return nil, err
			}
			r, size := utf8.DecodeRuneInString(s)
			if size == 0 {
				return wrap(""), nil
			}
			return wrap(string(unicode.ToTitle(r)) + strings.ToLower(s[size:])), nil
		case "strip", "lstrip", "rstrip":
			if err := argCount(name, args, 0, 1); err != nil {
				return nil, err
			}
			if len(args) == 0 || args[0] == nil {
				switch name {
				case "strip":
					return wrap(strings.TrimFunc(s, unicode.IsSpace)), nil
				case "lstrip":
					return wrap(strings.TrimLeftFunc(s, unicode.IsSpace)), nil
				}
				return wrap(strings.TrimRightFunc(s, unicode.IsSpace)), nil
			}
			chars, err := strArg(name, args[0])
			if err != nil {
				return nil, err
			}
			switch name {
			case "strip":
				return wrap(strings.Trim(s, chars)), nil
			case "lstrip":
				return wrap(strings.TrimLeft(s, chars)), nil
			}
			return wrap(strings.TrimRight(s, chars)), nil
		case "split":
			return e.pySplit(s, esc, args)
		case "replace":
			return pyReplace(s, esc, args)
		case "startswith", "endswith":
			if err := argCount(name, args, 1, 1); err != nil {
				return nil, err
			}
			arg, err := strArg(name, args[0])
			if err != nil {
				return nil, err
			}
			if name == "startswith" {
				return strings.HasPrefix(s, arg), nil
			}
			return strings.HasSuffix(s, arg), nil
		case "count", "find":
			if err := argCount(name, args, 1, 1); err != nil {
				return nil, err
			}
			arg, err := strArg(name, args[0])
			if err != nil {
				return nil, err
			}
			if name == "count" {
				return strings.Count(s, arg), nil
			}
			i := strings.Index(s, arg)
			if i < 0 {
				return -1, nil
			}
			return utf8.RuneCountInString(s[:i]), nil
		case "join":
			return e.pyJoin(s, esc, args)
		case "format":
			out, err := e.pyFormat(s, esc, args, kwargs)
			if err != nil {
				return nil, err
			}
			return wrap(out), nil
		}
		return nil, fmt.Errorf("%s() is not supported", name)
	}
	return &builtinMethod{name: name, fn: fn}
}

func pyTitle(s string) string {
	var b strings.Builder
	prevCased := false
	for _, r := range s {
		if unicode.IsUpper(r) || unicode.IsLower(r) || unicode.IsTitle(r) {
			if prevCased {
				b.WriteRune(unicode.ToLower(r))
			} else {
				b.WriteRune(unicode.ToTitle(r))
			}
			prevCased = true
		} else {
			b.WriteRune(r)
			prevCased = false
		}
	}
	return b.String()
}

func (e *DefaultEvaluator) pySplit(s string, esc bool, args []interface{}) (interface{}, error) {
	if err := argCount("split", args, 0, 2); err != nil {
		return nil, err
	}
	maxsplit := -1
	if len(args) == 2 {
		n, err := countArg("split", args[1])
		if err != nil {
			return nil, err
		}
		maxsplit = n
	}
	var parts []string
	if len(args) == 0 || args[0] == nil {
		parts = splitWhitespace(s, maxsplit)
	} else {
		sep, err := strArg("split", args[0])
		if err != nil {
			return nil, err
		}
		if sep == "" {
			return nil, fmt.Errorf("split(): empty separator")
		}
		switch {
		case maxsplit == 0:
			parts = []string{s}
		case maxsplit < 0:
			parts = strings.Split(s, sep)
		default:
			parts = strings.SplitN(s, sep, maxsplit+1)
		}
	}
	out := make([]interface{}, len(parts))
	for i, p := range parts {
		if esc {
			out[i] = Markup(p)
		} else {
			out[i] = p
		}
	}
	return out, nil
}

func splitWhitespace(s string, maxsplit int) []string {
	if maxsplit < 0 {
		return strings.Fields(s)
	}
	out := []string{}
	i := 0
	for len(out) < maxsplit {
		for i < len(s) {
			r, size := utf8.DecodeRuneInString(s[i:])
			if !unicode.IsSpace(r) {
				break
			}
			i += size
		}
		if i >= len(s) {
			return out
		}
		start := i
		for i < len(s) {
			r, size := utf8.DecodeRuneInString(s[i:])
			if unicode.IsSpace(r) {
				break
			}
			i += size
		}
		out = append(out, s[start:i])
	}
	for i < len(s) {
		r, size := utf8.DecodeRuneInString(s[i:])
		if !unicode.IsSpace(r) {
			break
		}
		i += size
	}
	if i < len(s) {
		out = append(out, s[i:])
	}
	return out
}

func pyReplace(s string, esc bool, args []interface{}) (interface{}, error) {
	if err := argCount("replace", args, 2, 3); err != nil {
		return nil, err
	}
	old, err := textArg("replace", args[0], false)
	if err != nil {
		return nil, err
	}
	repl, err := textArg("replace", args[1], esc)
	if err != nil {
		return nil, err
	}
	count := -1
	if len(args) == 3 {
		if count, err = countArg("replace", args[2]); err != nil {
			return nil, err
		}
	}
	n := int64(strings.Count(s, old))
	if old == "" {
		n = int64(utf8.RuneCountInString(s)) + 1
	}
	if count >= 0 && int64(count) < n {
		n = int64(count)
	}
	if size := int64(len(s)) + n*int64(len(repl)-len(old)); size > maxStringResult {
		return nil, fmt.Errorf("replace(): %w", ErrOutputTooLarge)
	}
	r := strings.Replace(s, old, repl, count)
	if esc {
		return Markup(r), nil
	}
	return r, nil
}

func (e *DefaultEvaluator) pyJoin(sep string, esc bool, args []interface{}) (interface{}, error) {
	if err := argCount("join", args, 1, 1); err != nil {
		return nil, err
	}
	items, err := e.makeIterable(args[0])
	if err != nil {
		return nil, fmt.Errorf("join() argument is not iterable: %w", err)
	}
	parts := make([]string, len(items))
	total := int64(0)
	for i, it := range items {
		var p string
		if _, isMarkup := AsMarkup(it); isMarkup || !esc {
			p = ToString(Unmark(it))
		} else {
			p = html.EscapeString(ToString(it))
		}
		parts[i] = p
		total += int64(len(p))
		if i > 0 {
			total += int64(len(sep))
		}
		if total > maxStringResult {
			return nil, fmt.Errorf("join(): %w", ErrOutputTooLarge)
		}
	}
	r := strings.Join(parts, sep)
	if esc {
		return Markup(r), nil
	}
	return r, nil
}

// pyFormat implements str.format (Decision 8 in plan 048).
func (e *DefaultEvaluator) pyFormat(s string, esc bool, args []interface{}, kwargs map[string]interface{}) (string, error) {
	var out strings.Builder
	auto, mode := 0, 0 // mode 1: automatic numbering, 2: manual
	for i := 0; i < len(s); {
		c := s[i]
		switch c {
		case '{':
			if i+1 < len(s) && s[i+1] == '{' {
				out.WriteByte('{')
				i += 2
				continue
			}
			j, inBracket := i+1, false
			for ; j < len(s); j++ {
				if s[j] == '[' {
					inBracket = true
				} else if s[j] == ']' {
					inBracket = false
				} else if s[j] == '}' && !inBracket {
					break
				}
			}
			if j >= len(s) {
				return "", fmt.Errorf("format(): expected '}' before end of string")
			}
			text, err := e.formatField(s[i+1:j], esc, args, kwargs, &auto, &mode)
			if err != nil {
				return "", err
			}
			out.WriteString(text)
			i = j + 1
		case '}':
			if i+1 < len(s) && s[i+1] == '}' {
				out.WriteByte('}')
				i += 2
				continue
			}
			return "", fmt.Errorf("format(): single '}' encountered in format string")
		default:
			out.WriteByte(c)
			i++
		}
		if out.Len() > maxStringResult {
			return "", fmt.Errorf("format(): %w", ErrOutputTooLarge)
		}
	}
	return out.String(), nil
}

func (e *DefaultEvaluator) formatField(field string, esc bool, args []interface{}, kwargs map[string]interface{}, auto, mode *int) (string, error) {
	nameEnd, inBracket := len(field), false
	for k := 0; k < len(field); k++ {
		if field[k] == '[' {
			inBracket = true
		} else if field[k] == ']' {
			inBracket = false
		} else if !inBracket && (field[k] == '!' || field[k] == ':') {
			nameEnd = k
			break
		}
	}
	name, rest := field[:nameEnd], field[nameEnd:]
	conv := byte(0)
	if strings.HasPrefix(rest, "!") {
		if len(rest) < 2 {
			return "", fmt.Errorf("format(): end of string while looking for conversion specifier")
		}
		conv = rest[1]
		rest = rest[2:]
		if conv != 'r' && conv != 's' {
			return "", fmt.Errorf("format(): unknown conversion specifier %c", conv)
		}
	}
	if strings.HasPrefix(rest, ":") && len(rest) > 1 {
		return "", fmt.Errorf("format(): format spec %q is not supported; use the format filter", rest[1:])
	}
	if rest != "" && rest != ":" {
		return "", fmt.Errorf("format(): expected ':' after conversion specifier")
	}

	first, parts := name, ""
	if k := strings.IndexAny(name, ".["); k >= 0 {
		first, parts = name[:k], name[k:]
	}
	var value interface{}
	switch {
	case first == "":
		if *mode == 2 {
			return "", fmt.Errorf("format(): cannot switch from manual field specification to automatic field numbering")
		}
		*mode = 1
		idx := *auto
		*auto++
		if idx >= len(args) {
			return "", fmt.Errorf("IndexError: Replacement index %d out of range for positional args tuple", idx)
		}
		value = args[idx]
	case isDigits(first):
		if *mode == 1 {
			return "", fmt.Errorf("format(): cannot switch from automatic field numbering to manual field specification")
		}
		*mode = 2
		idx := 0
		for _, d := range first {
			idx = idx*10 + int(d-'0')
			if idx > math.MaxInt32 {
				break
			}
		}
		if idx >= len(args) {
			return "", fmt.Errorf("IndexError: Replacement index %d out of range for positional args tuple", idx)
		}
		value = args[idx]
	default:
		v, ok := kwargs[first]
		if !ok {
			return "", fmt.Errorf("KeyError: '%s'", first)
		}
		value = v
	}

	for parts != "" {
		if parts[0] == '.' {
			end := strings.IndexAny(parts[1:], ".[")
			var attr string
			if end < 0 {
				attr, parts = parts[1:], ""
			} else {
				attr, parts = parts[1:1+end], parts[1+end:]
			}
			if attr == "" {
				return "", fmt.Errorf("format(): empty attribute in format string")
			}
			if !e.attributeExists(value, attr) {
				return "", fmt.Errorf("'%s' object has no attribute '%s'", e.getObjectName(value), attr)
			}
			value = e.getAttribute(value, attr)
			continue
		}
		end := strings.IndexByte(parts, ']')
		if end < 0 {
			return "", fmt.Errorf("format(): missing ']' in format string")
		}
		key := parts[1:end]
		parts = parts[end+1:]
		if parts != "" && parts[0] != '.' && parts[0] != '[' {
			return "", fmt.Errorf("format(): only '.' or '[' may follow ']' in format field specifier")
		}
		if m, ok := StringKeyMap(value); ok {
			v, found := m[key]
			if !found {
				return "", fmt.Errorf("KeyError: '%s'", key)
			}
			value = v
			continue
		}
		var k interface{} = key
		if isDigits(key) {
			n := 0
			for _, d := range key {
				n = n*10 + int(d-'0')
				if n > math.MaxInt32 {
					break
				}
			}
			k = n
		}
		v, err := e.getItem(value, k)
		if err != nil {
			return "", err
		}
		if u, ok := v.(*Undefined); ok && u != nil {
			return "", fmt.Errorf("IndexError: index %s out of range", key)
		}
		value = v
	}

	var text string
	if conv == 'r' {
		text = reprOf(value)
	} else {
		if m, ok := AsMarkup(value); ok && esc {
			return string(m), nil
		}
		text = ToString(value)
	}
	if esc {
		text = html.EscapeString(text)
	}
	return text, nil
}

func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// ---- repetition

// repeatValue implements seq * count for a string or a list.
func repeatValue(seq, count interface{}) (interface{}, error) {
	n, err := repeatCount(count)
	if err != nil {
		return nil, err
	}
	if s, ok := basicValue(seq).(string); ok {
		if n <= 0 || s == "" {
			return "", nil
		}
		if n > maxStringResult/len(s) {
			return nil, fmt.Errorf("string repetition %w", ErrOutputTooLarge)
		}
		return strings.Repeat(s, n), nil
	}
	rv := reflect.ValueOf(Indirect(seq))
	if n <= 0 || rv.Len() == 0 {
		return []interface{}{}, nil
	}
	if n > maxListRepeat/rv.Len() {
		return nil, fmt.Errorf("list repetition %w", ErrOutputTooLarge)
	}
	out := make([]interface{}, 0, n*rv.Len())
	for k := 0; k < n; k++ {
		for i := 0; i < rv.Len(); i++ {
			out = append(out, rv.Index(i).Interface())
		}
	}
	return out, nil
}

func repeatCount(count interface{}) (int, error) {
	i, u, f, kind, ok := numericValue(count)
	if ok {
		switch kind {
		case 'i':
			return clampInt(i), nil
		case 'u':
			if u > math.MaxInt64 {
				return math.MaxInt, nil
			}
			return clampInt(int64(u)), nil
		case 'f':
			if f == math.Trunc(f) && !math.IsInf(f, 0) {
				if f > math.MaxInt64 {
					return math.MaxInt, nil
				}
				if f < math.MinInt64 {
					return math.MinInt, nil
				}
				return clampInt(int64(f)), nil
			}
		}
	}
	return 0, fmt.Errorf("can't multiply sequence by non-integer of type %T", count)
}

func isRepeatableList(v interface{}) bool {
	v = Indirect(v)
	if v == nil {
		return false
	}
	if _, ok := v.([]byte); ok {
		return false
	}
	return isBoxable(v)
}

// isSequenceCount reports whether a and b are a string or list and a number,
// in either order, and returns them as (sequence, count).
func sequenceOperands(a, b interface{}) (seq, count interface{}, ok bool) {
	isSeq := func(v interface{}) bool {
		_, isStr := basicValue(v).(string)
		return isStr || isRepeatableList(v)
	}
	isNum := func(v interface{}) bool { _, _, _, _, n := numericValue(v); return n }
	switch {
	case isSeq(a) && isNum(b):
		return a, b, true
	case isSeq(b) && isNum(a):
		return b, a, true
	}
	return nil, nil, false
}
