package runtime

import "testing"

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
}
