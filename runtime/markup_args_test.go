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
