package runtime

import (
	"reflect"
	"testing"
	"time"
)

func markupTestCtx() Context {
	return &autoescapeContext{Context: &simpleContext{variables: map[string]interface{}{}}, autoescape: true}
}

func TestEscapeArgsKwArgs(t *testing.T) {
	out := escapeArgs(markupTestCtx(), []interface{}{"<x>", KwArgs{"n": "<s>", "m": Markup("<i>"), "k": 3}})
	if out[0] != "&lt;x&gt;" {
		t.Errorf("element 0 = %#v", out[0])
	}
	kw, ok := out[1].(KwArgs)
	if !ok {
		t.Fatalf("element 1 is %T, want KwArgs", out[1])
	}
	if kw["n"] != "&lt;s&gt;" {
		t.Errorf("n = %#v", kw["n"])
	}
	if s, ok := kw["m"].(string); !ok || s != "<i>" {
		t.Errorf("m = %#v, want plain string <i>", kw["m"])
	}
	if kw["k"] != 3 {
		t.Errorf("k = %#v", kw["k"])
	}
}

func TestWordwrapKeepsMarkupArgs(t *testing.T) {
	e := NewEvaluator()
	for _, tc := range []struct {
		name   string
		markup bool
	}{{"wordwrap", true}, {"upper", false}} {
		var got []interface{}
		apply := func(v interface{}, a []interface{}) (interface{}, error) {
			got = a
			return "", nil
		}
		if _, err := e.applyFilterWithMarkup(tc.name, "ab cd", []interface{}{5, true, Markup("<br>")}, markupTestCtx(), apply); err != nil {
			t.Fatal(err)
		}
		_, isMarkup := got[2].(Markup)
		if isMarkup != tc.markup {
			t.Errorf("%s: third argument %T, markup=%v want %v", tc.name, got[2], isMarkup, tc.markup)
		}
	}
}

// A context that cannot tell whether a filter exists gets the whole-list
// apply, then a re-apply per item.
func TestMapMarkupWithoutFilterLookup(t *testing.T) {
	ctx := &autoescapeContext{Context: &mockContext{data: map[string]interface{}{}}, autoescape: true}
	if _, ok := filterLookup(ctx); ok {
		t.Fatal("mockContext must not implement FilterLookup")
	}
	var calls [][]interface{}
	apply := func(v interface{}, a []interface{}) (interface{}, error) {
		calls = append(calls, v.([]interface{}))
		return []interface{}{"W", "W"}, nil
	}
	out, err := NewEvaluator().mapMarkup(ctx, []interface{}{1, "a"}, []interface{}{"f"}, apply)
	if err != nil {
		t.Fatal(err)
	}
	if len(calls) != 1 || len(calls[0]) != 2 {
		t.Errorf("apply calls = %v, want one call with the whole list", calls)
	}
	if got := out.([]interface{}); got[0] != 1 || got[1] != "a" {
		t.Errorf("out = %v, want the items re-applied through ApplyFilter", got)
	}

	// An Undefined result of the whole-list apply is kept as it is.
	undef := NewUndefined("x", UndefinedSilent, nil)
	apply = func(v interface{}, a []interface{}) (interface{}, error) {
		return []interface{}{undef, "W"}, nil
	}
	out, err = NewEvaluator().mapMarkup(ctx, []interface{}{1, "a"}, []interface{}{"f"}, apply)
	if err != nil {
		t.Fatal(err)
	}
	if got := out.([]interface{}); got[0] != undef {
		t.Errorf("out[0] = %#v, want the Undefined kept", got[0])
	}
}

type lookupCtx struct{ *mockContext }

func (lookupCtx) HasFilter(string) bool { return true }

// When the filter over the items that are not re-applied returns a list of
// another length, the whole list is filtered instead.
func TestMapMarkupSubListMismatch(t *testing.T) {
	ctx := lookupCtx{&mockContext{data: map[string]interface{}{}}}
	if _, ok := filterLookup(ctx); !ok {
		t.Fatal("lookupCtx must implement FilterLookup")
	}
	var calls [][]interface{}
	apply := func(v interface{}, a []interface{}) (interface{}, error) {
		list := v.([]interface{})
		calls = append(calls, list)
		if len(list) == 1 {
			return []interface{}{"X", "Y"}, nil
		}
		return []interface{}{"W1", "W2"}, nil
	}
	out, err := NewEvaluator().mapMarkup(ctx, []interface{}{1, []byte("x")}, []interface{}{"f"}, apply)
	if len(calls) != 2 || len(calls[1]) != 2 {
		t.Fatalf("apply calls = %v, want a sub-list call then a call with the whole 2-item list", calls)
	}
	if err != nil {
		t.Fatal(err)
	}
	if got := out.([]interface{}); len(got) != 2 || got[0] != "W1" || got[1] != "W2" {
		t.Errorf("out = %v, want [W1 W2]", got)
	}
}

// userFilterCtx is an environment context whose filter "kind" is the
// user's; every other filter and test is miya's own.
type userFilterCtx struct{ Context }

func (userFilterCtx) ApplyFilter(string, interface{}, ...interface{}) (interface{}, error) {
	return nil, nil
}
func (userFilterCtx) ApplyTest(string, interface{}, ...interface{}) (bool, error) { return false, nil }
func (userFilterCtx) IsBuiltinFilter(name string) bool                            { return name != "kind" }
func (userFilterCtx) IsBuiltinTest(string) bool                                   { return true }

// The scans that turn escaped output into plain strings for Go code run only
// for the user's filters, in a render whose evaluator made escaped output:
// built-in filters and HTML renders never pay for them, and an HTML render
// makes none. Unbox, which has no evaluator, scans once any evaluator made
// escaped output.
func TestEscapedScanOnlyAfterEscapedOutput(t *testing.T) {
	defer escapedOutputMade.Store(true)
	escapedOutputMade.Store(false)
	var seen interface{}
	kind := func(v interface{}, _ []interface{}) (interface{}, error) {
		seen = v.([]interface{})[1]
		return "", nil
	}
	e := NewEvaluator()
	ctx := &autoescapeContext{Context: userFilterCtx{&simpleContext{variables: map[string]interface{}{}}}, autoescape: true}
	if r := e.markIfSafe("<x>", ctx); r != Markup("<x>") {
		t.Fatalf("markIfSafe = %#v", r)
	}
	if _, ok := e.concatEscaped(ctx, "a", Markup("b")); ok || e.escapedMade.Load() || escapedOutputMade.Load() {
		t.Fatal("HTML output made escaped output")
	}
	list := []interface{}{"a", escapedJS("b")}
	scanned := func(name string, args ...interface{}) bool {
		t.Helper()
		seen = nil
		if _, err := e.applyFilterWithMarkup(name, list, args, ctx, kind); err != nil {
			t.Fatal(err)
		}
		_, plain := seen.(string)
		return plain
	}
	// a value made elsewhere is not looked for
	if scanned("kind") {
		t.Error("scanned before escaped output was made")
	}
	if got := Unbox(list).([]interface{})[1]; got != escapedJS("b") {
		t.Errorf("Unbox scanned before escaped output was made: got %T", got)
	}
	// another render made escaped output
	NewEvaluator().markEscaped("c", EscapeContextJS)
	if scanned("kind") {
		t.Error("scanned after another render made escaped output")
	}
	if got := Unbox(list).([]interface{})[1]; got != "b" {
		t.Errorf("Unbox did not scan after escaped output was made: got %T", got)
	}
	e.markEscaped("c", EscapeContextJS)
	for _, c := range []struct {
		name string
		args []interface{}
		want bool
	}{
		{"kind", nil, true},
		{"length", nil, false},
		{"map", []interface{}{"kind"}, true},
		{"map", []interface{}{"upper"}, false},
		{"select", []interface{}{"string"}, false},
	} {
		if got := scanned(c.name, c.args...); got != c.want {
			t.Errorf("%s%v: scanned = %v, want %v", c.name, c.args, got, c.want)
		}
	}
}

// plainEscaped walks a value that shares containers once per container, and
// stops where a container contains itself.
func TestPlainEscapedSharedAndSelfContaining(t *testing.T) {
	escapedOutputMade.Store(true)
	var l interface{} = []interface{}{escapedJS("a")}
	for i := 0; i < 40; i++ {
		l = []interface{}{l, l}
	}
	self := []interface{}{escapedJS("a"), nil}
	self[1] = self
	done := make(chan interface{}, 1)
	go func() { done <- plainEscaped(map[string]interface{}{"l": l, "self": self}) }()
	select {
	case r := <-done:
		leaf := r.(map[string]interface{})["l"]
		for i := 0; i < 40; i++ {
			leaf = leaf.([]interface{})[1]
		}
		if got := leaf.([]interface{})[0]; got != "a" {
			t.Errorf("shared leaf = %#v", got)
		}
		if got := r.(map[string]interface{})["self"].([]interface{})[0]; got != "a" {
			t.Errorf("self-containing item = %#v", got)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("timed out")
	}
	// the same shapes without escaped output come back as they are
	var p interface{} = []interface{}{"a"}
	for i := 0; i < 40; i++ {
		p = []interface{}{p, p}
	}
	pself := []interface{}{"a", nil}
	pself[1] = pself
	in := map[string]interface{}{"l": p, "self": pself}
	go func() { done <- plainEscaped(in) }()
	select {
	case r := <-done:
		if reflect.ValueOf(r).Pointer() != reflect.ValueOf(in).Pointer() {
			t.Error("a value without escaped output was copied")
		}
	case <-time.After(10 * time.Second):
		t.Fatal("timed out")
	}
}
