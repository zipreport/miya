package runtime

import (
	"fmt"
	"testing"

	"github.com/zipreport/miya/parser"
)

func TestEscapeOutput(t *testing.T) {
	tests := []struct {
		name string
		in   interface{}
		want string
	}{
		{"markup", Markup("<a>"), "<a>"},
		{"safe_value", SafeValue{Value: "<a>"}, "<a>"},
		{"safe_value_ptr", &SafeValue{Value: "<a>"}, "<a>"},
		{"safe_value_nil_ptr", (*SafeValue)(nil), ""},
		{"string", "<a>", "&lt;a&gt;"},
		{"int", 42, "42"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToString(escapeOutput(tt.in)); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestStrictUndefinedFromCall(t *testing.T) {
	e := NewEvaluator()
	e.SetUndefinedBehavior(UndefinedStrict)
	ctx := &simpleContext{variables: map[string]interface{}{
		"f": func() interface{} { return NewUndefined("x", UndefinedStrict, nil) },
	}}
	node := &parser.VariableNode{Expression: &parser.CallNode{Function: &parser.IdentifierNode{Name: "f"}}}
	got, err := e.EvalVariableNode(node, ctx)
	if err == nil {
		t.Fatalf("want an error for a strict undefined returned by a call, got %v", got)
	}
}

func TestRoundedFloatFormat(t *testing.T) {
	r := RoundedFloat{Value: 2.5, Places: 2}
	for _, tt := range []struct{ format, want string }{
		{"%v", "2.50"}, {"%s", "2.50"}, {"%d", "2"}, {"%.1f", "2.5"},
	} {
		if got := fmt.Sprintf(tt.format, r); got != tt.want {
			t.Errorf("Sprintf(%q) = %q, want %q", tt.format, got, tt.want)
		}
	}
}
