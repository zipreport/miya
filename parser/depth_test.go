package parser

import (
	"strings"
	"testing"

	"github.com/zipreport/miya/lexer"
)

func TestParseDepthLimit(t *testing.T) {
	rep := strings.Repeat
	expr := func(s string) string { return "{{ " + s + " }}" }
	cases := []struct {
		name    string
		src     string
		wantErr bool
	}{
		{"paren_at_limit", expr(rep("(", 499) + "1" + rep(")", 499)), false},
		{"paren_over_limit", expr(rep("(", 501) + "1" + rep(")", 501)), true},
		{"paren_huge", expr(rep("(", 200000) + "1" + rep(")", 200000)), true},
		{"list_huge", expr(rep("[", 200000) + "1" + rep("]", 200000)), true},
		{"if_at_limit", rep("{% if true %}", 499) + "x" + rep("{% endif %}", 499), false},
		{"if_huge", rep("{% if true %}", 1000000) + "x" + rep("{% endif %}", 1000000), true},
		{"not_chain_at_limit", expr(rep("not ", 499) + "1"), false},
		{"not_chain_huge", expr(rep("not ", 200000) + "1"), true},
		{"else_chain_at_limit", expr(rep("1 if x else ", 499) + "1"), false},
		{"else_chain_huge", expr(rep("1 if x else ", 200000) + "1"), true},
		{"pow_chain_at_limit", expr(rep("2 ** ", 499) + "2"), false},
		// binary operator, filter and attribute chains nest left-deep, one level per term
		{"pow_chain_huge", expr(rep("2 ** ", 200000) + "2"), true},
		{"plus_chain_huge", expr(rep("1 + ", 200000) + "1"), true},
		{"plus_chain_at_limit", expr(rep("1 + ", 400) + "1"), false},
		{"concat_chain_huge", expr(rep("'a' ~ ", 200000) + "'a'"), true},
		{"filter_chain_huge", expr("x" + rep("|upper", 200000)), true},
		{"attr_chain_huge", expr("a" + rep(".b", 200000)), true},
		{"or_chain_huge", expr(rep("x or ", 200000) + "x"), true},
		{"neg_chain", expr(rep("-", 200000) + "1"), true},
		{"neg_at_limit", expr(rep("-", 400) + "1"), false},
		{"many_siblings", expr(rep("1 + ", 400) + "1"), false},
		{"many_sibling_exprs", rep("{{ 1 + 1 }}", 600), false},
		{"many_sibling_tags", rep("{% if true %}x{% endif %}", 600), false},
		{"many_sibling_conditionals", rep("{{ 1 if x else 2 }}", 600) + rep("{{ not x }}", 600) + rep("{{ 2 ** 3 }}", 600), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tokens, err := lexer.NewLexer(tc.src, nil).Tokenize()
			if err != nil {
				t.Fatalf("tokenize: %v", err)
			}
			_, err = NewParser(tokens).Parse()
			if !tc.wantErr {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), "nesting too deep") {
				t.Fatalf("want error containing %q, got %v", "nesting too deep", err)
			}
		})
	}
}
