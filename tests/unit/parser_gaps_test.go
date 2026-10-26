package miya_test

import (
	"testing"

	miya "github.com/zipreport/miya"
	"github.com/zipreport/miya/lexer"
	"github.com/zipreport/miya/loader"
	"github.com/zipreport/miya/parser"
)

// renderGap renders src with the variables every case shares. Expected
// values are Jinja2 3.1.6's output, except that miya prints booleans as
// true/false and whole floats without ".0".
func renderGap(t *testing.T, src string) (string, error) {
	t.Helper()
	sl := loader.NewStringLoader(loader.NewDirectTemplateParser())
	sl.AddTemplate("p", "[{{ x }}{{ g }}]")
	sl.AddTemplate("p2", "{% set leaked = 'L' %}")
	env := miya.NewEnvironment(miya.WithLoader(sl), miya.WithAutoEscape(false))
	env.AddGlobal("g", "G")
	tmpl, err := env.FromString(src)
	if err != nil {
		return "", err
	}
	return tmpl.Render(miya.NewContextFrom(map[string]interface{}{
		"x": 5, "i": -5, "a": 1, "y": []interface{}{1, 2},
		"L": []interface{}{[]interface{}{1, 2}, []interface{}{3, 4}},
		"d": map[string]interface{}{"if": 1, "in": 3, "not": 4, "true": 6, "none": 7, "block": 8, "is": 10, "é": 11},
	}))
}

func TestParserGaps(t *testing.T) {
	tests := []struct{ name, src, want string }{
		// Tests without parentheses take one argument, as in Jinja2.
		{"test_arg_no_parens", "{{ 6 is divisibleby 3 }}", "true"},
		{"test_arg_then_and", "{{ 3 is gt 2 and true }}", "true"},
		{"test_not_no_parens", "{{ x is not divisibleby 2 }}", "true"},
		{"test_in_conditional", "{{ 1 if x is gt 2 else 0 }}", "1"},
		{"test_arg_attribute", "{{ x is divisibleby d.if }}", "true"},
		{"test_arg_keyword_name", "{% set block = 5 %}{{ 10 is divisibleby block }}", "true"},
		{"test_after_filter", "{{ y|first is eq 1 }}", "true"},
		{"test_on_test_result", "{{ 3 is gt 2 is true }}", "true"},
		// "if" and "recursive" after a test are never its argument. Jinja2
		// rejects the first form; miya accepts it today and keeps doing so.
		{"test_before_inline_if", "{{ x is odd if true else 0 }}", "true"},
		{"test_before_recursive", "{% for v in L if v is not none recursive %}{% if v is iterable %}{{ loop(v) }}{% else %}{{ v }}{% endif %}{% endfor %}", "1234"},
		{"is_in", "{{ 'a' is in 'abc' }}", "true"},
		{"is_not_in", "{{ x is not in y }}", "true"},
		{"is_true", "{{ true is true }}", "true"},
		{"is_true_int", "{{ 1 is true }}", "false"},
		{"is_false", "{{ false is false }}", "true"},
		{"is_filter", "{{ 'upper' is filter }}", "true"},
		{"is_filter_unknown", "{{ 'nope' is filter }}", "false"},
		{"is_test", "{{ 'odd' is test }}", "true"},
		{"is_test_unknown", "{{ 'nope' is test }}", "false"},
		// Keywords are names after "." and, where Jinja2 allows it, as variables.
		{"attr_if", "{{ d.if }}", "1"},
		{"attr_in", "{{ d.in }}", "3"},
		{"attr_not", "{{ d.not }}", "4"},
		{"attr_true", "{{ d.true }}", "6"},
		{"attr_none", "{{ d.none }}", "7"},
		{"attr_block", "{{ d.block }}", "8"},
		{"attr_is", "{{ d.is }}", "10"},
		{"var_block", "{% set block = 1 %}{{ block }}", "1"},
		{"var_filter", "{% set filter = 2 %}{{ filter }}", "2"},
		{"unicode_var", "{% set é = 1 %}{{ é }}", "1"},
		{"unicode_var_cjk", "{% set 名前 = 'n' %}{{ 名前 }}", "n"},
		{"unicode_attr", "{{ d.é }}", "11"},
		// An integer after "." is an index, and is never part of a float.
		{"attr_int_index", "{{ y.0 }}", "1"},
		{"attr_int_chain", "{{ L.0.1 }}", "2"},
		{"attr_int_chain_second", "{{ L.1.0 }}", "3"},
		// Comparisons chain: a < b < c is a < b and b < c.
		{"chain_true", "{{ 1 < 2 < 3 }}", "true"},
		{"chain_false", "{{ 1 < 3 < 2 }}", "false"},
		{"chain_eq", "{{ 1 == a == 1 }}", "true"},
		{"chain_mixed", "{{ 1 < x > 4 }}", "true"},
		// Unary minus binds tighter than ** and is applied before filters.
		{"neg_before_filter", "{{ -i|abs }}", "5"},
		{"neg_pow", "{{ -2**2 }}", "4"},
		{"pow_left_assoc", "{{ 2**3**2 }}", "64"},
		{"pow_neg_exponent", "{{ 2**-1 }}", "0.5"},
		// Jinja2 precedence changes these (CHANGELOG "Breaking"): "is" binds
		// to the nearest operand, and "~" binds tighter than "+".
		{"is_after_plus_parenthesized", "{{ (x + 1) is divisibleby(3) }}", "true"},
		{"is_binds_to_operand_eq", "{{ a == 1 is sameas(true) }}", "false"},
		{"tilde_before_plus", "{{ 'Page ' ~ x + 1 }}", "Page 51"}, // Jinja2: TypeError; miya adds the strings
		{"tilde_parenthesized", "{{ 'Page ' ~ (x + 1) }}", "Page 6"},
		{"plus_then_tilde", "{{ a + 2 ~ 3 }}", "24"},
		// Tuples.
		{"tuple_length", "{{ (1, 2)|length }}", "2"},
		{"tuple_empty", "{{ ()|length }}", "0"},
		{"tuple_single", "{{ (1,)|first }}", "1"},
		{"tuple_unpack_in_for", "{% for a, b in [(1, 2)] %}{{ a }}{{ b }}{% endfor %}", "12"},
		{"for_bare_tuple", "{% for v in 1, 2 %}{{ v }}{% endfor %}", "12"},
		{"set_tuple", "{% set a, b = 1, 2 %}{{ a }}{{ b }}", "12"},
		{"set_paren_targets", "{% set (a, b) = (5, 6) %}{{ a }}{{ b }}", "56"},
		{"set_bare_tuple", "{% set n = 1, 2 %}{{ n|length }}", "2"},
		// Numeric literals.
		{"int_underscore", "{{ 1_000 }}", "1000"},
		{"int_hex", "{{ 0x10 }}", "16"},
		{"int_octal", "{{ 0o17 }}", "15"},
		{"int_binary", "{{ 0b101 }}", "5"},
		{"int_prefix_needs_digit", "{{ 0 if 0or 0 else 2 }}", "2"},
		{"float_underscore", "{{ 1_0.5 }}", "10.5"},
		{"float_exponent", "{{ 1e3 }}", "1000"},
		{"float_exponent_upper", "{{ 2E3 }}", "2000"},
		{"float_neg_exponent", "{{ 1.5e-2 }}", "0.015"},
		// Adjacent string literals are one string.
		{"str_concat", "{{ 'a' 'b' }}", "ab"},
		{"str_concat_three", `{{ 'a' "b" 'c' }}`, "abc"},
		{"str_concat_tilde", "{{ 'a' ~ 'b' 'c' }}", "abc"},
		// An inline if without else is undefined when false.
		{"inline_if_no_else_false", "[{{ 'a' if false }}]", "[]"},
		{"inline_if_no_else_true", "[{{ 'a' if true }}]", "[a]"},
		// A comprehension's "if" filters; it is not an inline if on the iterable.
		{"comprehension_filter_cmp", "{{ [v for v in y if v > 1]|length }}", "1"},
		{"comprehension_filter_truthy", "{{ [v for v in y if v]|length }}", "2"},
		// include: Jinja2's [ignore missing] [with|without context] order.
		{"include_with_context", `{% include "p" with context %}`, "[5G]"},
		{"include_with_context_no_leak", `{% include "p2" with context %}[{{ leaked }}]`, "[]"},
		{"include_with_context_expr", `{% include "p" with context.x %}`, "[5G]"},
		{"include_without_context", `{% include "p" without context %}`, "[G]"},
		{"include_without_context_in_loop", `{% for x in [1] %}{% include "p" without context %}{% endfor %}`, "[G]"},
		{"include_in_macro_without_context", `{% macro m() %}{% include "p" without context %}{% endmacro %}{{ m() }}`, "[G]"},
		{"include_ignore_missing_without_context", `{% include "p" ignore missing without context %}`, "[G]"},
		{"include_missing_ignored_with_context", `{% include "nope" ignore missing with context %}`, ""},
		{"include_old_order", `{% include "nope" with context ignore missing %}`, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := renderGap(t, tt.src)
			if err != nil {
				t.Fatalf("%s: %v", tt.src, err)
			}
			if got != tt.want {
				t.Errorf("%s: got %q, want %q", tt.src, got, tt.want)
			}
		})
	}

	// Jinja2 rejects these; miya must too.
	for name, src := range map[string]string{
		"int_trailing_underscore": "{{ 1_ }}",
		"hex_no_digits":           "{{ 0x }}",
		"test_chained_on_test":    "{{ x is number is true }}",
		"include_without_alone":   `{% include "p" without %}`,
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := miya.NewEnvironment().FromString(src); err == nil {
				t.Errorf("%s: want a parse error, got none", src)
			}
		})
	}
}

// TestParserGapsPrecedence checks the tree where rendering cannot show it:
// Jinja2 raises TypeError for 1 + 2 ~ 3, and x + 1 is divisibleby(3) needs
// bool arithmetic (plan 047) to render.
func TestParserGapsPrecedence(t *testing.T) {
	for src, want := range map[string]string{
		"{{ 1 + 2 ~ 3 }}":               "BinOp(Literal(1) + BinOp(Literal(2) ~ Literal(3)))",
		"{{ 1 ~ 2 + 3 }}":               "BinOp(BinOp(Literal(1) ~ Literal(2)) + Literal(3))",
		"{{ -2 ** 2 }}":                 "BinOp(UnaryOp(- Literal(2)) ** Literal(2))",
		"{{ x + 1 is divisibleby(3) }}": "BinOp(Id(x) + Test(Literal(1) is divisibleby(Literal(3))))",
		"{{ a == 1 is sameas(true) }}":  "BinOp(Id(a) == Test(Literal(1) is sameas(Literal(true))))",
		"{{ 1 < b }}":                   "BinOp(Literal(1) < Id(b))",
		"{{ 1 < b < 3 }}":               "Compare(Literal(1) < Id(b) < Literal(3))",
		"{{ 'a' if b }}":                "Conditional(Id(b) ? Literal(a))",
	} {
		tokens, err := lexer.NewLexer(src, nil).Tokenize()
		if err != nil {
			t.Fatal(err)
		}
		tmpl, err := parser.NewParser(tokens).Parse()
		if err != nil {
			t.Fatal(err)
		}
		got := tmpl.Children[0].(*parser.VariableNode).Expression.String()
		if got != want {
			t.Errorf("%s: got %s, want %s", src, got, want)
		}
	}
}

// TestParserGapsChainEvaluatesOnce: as in Jinja2, each operand of a chain is
// evaluated once, and evaluation stops at the first false comparison.
func TestParserGapsChainEvaluatesOnce(t *testing.T) {
	for _, tt := range []struct {
		src, want string
		calls     int
	}{
		{"{{ 1 < tick() < 3 }}", "true", 1},
		{"{{ 3 < 2 < tick() }}", "false", 0},
	} {
		calls := 0
		env := miya.NewEnvironment()
		env.AddGlobal("tick", func() int { calls++; return 2 })
		tmpl, err := env.FromString(tt.src)
		if err != nil {
			t.Fatal(err)
		}
		got, err := tmpl.Render(miya.NewContext())
		if err != nil {
			t.Fatal(err)
		}
		if got != tt.want || calls != tt.calls {
			t.Errorf("%s: got %q with %d calls, want %q with %d", tt.src, got, calls, tt.want, tt.calls)
		}
	}
}

// TestParserGapsChainRelease: releasing a template that holds a chain must
// return each node to the pool once. A node returned twice is handed to two
// later parses, which then share it ("{{ p ~ '-' ~ q }}" printed "Q-Q").
// Run without -race: sync.Pool keeps nothing between Gets under the race
// detector, so the test cannot fail there.
func TestParserGapsChainRelease(t *testing.T) {
	first, err := miya.NewEnvironment().FromString("{{ 1 < b < 3 }}")
	if err != nil {
		t.Fatal(err)
	}
	first.Release()
	for i := 0; i < 20; i++ {
		tmpl, err := miya.NewEnvironment().FromString("{{ p ~ '-' ~ q }}")
		if err != nil {
			t.Fatal(err)
		}
		got, err := tmpl.Render(miya.NewContextFrom(map[string]interface{}{"p": "P", "q": "Q"}))
		if err != nil || got != "P-Q" {
			t.Fatalf("got %q, %v; want \"P-Q\"", got, err)
		}
	}
}

// TestParserGapsUserTestWins: a test the user registers as "filter" or
// "test" replaces the built-in meaning of "is filter" / "is test".
func TestParserGapsUserTestWins(t *testing.T) {
	env := miya.NewEnvironment()
	if err := env.AddTest("filter", func(value interface{}, args ...interface{}) (bool, error) { return true, nil }); err != nil {
		t.Fatal(err)
	}
	tmpl, err := env.FromString("{{ 'nope' is filter }}")
	if err != nil {
		t.Fatal(err)
	}
	if got, err := tmpl.Render(miya.NewContext()); err != nil || got != "true" {
		t.Errorf("got %q, %v; want \"true\"", got, err)
	}
}
