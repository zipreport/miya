package miya_test

import (
	"testing"

	miya "github.com/zipreport/miya"
	"github.com/zipreport/miya/loader"
	"github.com/zipreport/miya/runtime"
)

// Issue #1: super() inside an expression renders the parent block's content.
func TestIssue1SuperInExpression(t *testing.T) {
	const base = "[{% block a %}a{% endblock %}]"
	const cyc = "{% set c = cycler(1, 2, 3) %}[{% block a %}{{ c.next() }}{% endblock %}]"
	const strictBase = "[{% block a %}{{ missing }}{% endblock %}]"
	two := func(b, child string) map[string]string {
		return map[string]string{"base": b, "child": "{% extends 'base' %}" + child}
	}
	// want values are Jinja2 3.1.6's output
	tests := []struct {
		name       string
		autoescape bool
		strict     bool
		templates  map[string]string
		want       string
	}{
		{"filter", false, false, two(base, "{% block a %}{{ super()|upper }}{% endblock %}"), "[A]"},
		{"macro_arg", false, false, two(base, "{% macro w(x) %}({{ x }}){% endmacro %}{% block a %}{{ w(super()) }}{% endblock %}"), "[(a)]"},
		{"macro_kwarg", false, false, two(base, "{% macro w(x) %}({{ x }}){% endmacro %}{% block a %}{{ w(x=super()) }}{% endblock %}"), "[(a)]"},
		{"concat", false, false, two(base, "{% block a %}{{ super() ~ '!' }}{% endblock %}"), "[a!]"},
		{"set", false, false, two(base, "{% block a %}{% set p = super() %}<{{ p }}>{% endblock %}"), "[<a>]"},
		{"if", false, false, two(base, "{% block a %}{% if super() %}yes{% endif %}{% endblock %}"), "[yes]"},
		{"escaped_not_twice", true, false, two(`[{% block a %}{{ "<b>" }}{% endblock %}]`, "{% block a %}{{ super()|upper }}{% endblock %}"), "[&LT;B&GT;]"},
		// further shapes of the same defect
		{"elif", false, false, two(base, "{% block a %}{% if false %}no{% elif super() == 'a' %}elif{% endif %}{% endblock %}"), "[elif]"},
		{"in_for_body", false, false, two(base, "{% block a %}{% for i in [1, 2] %}{{ super()|upper }}{% endfor %}{% endblock %}"), "[AA]"},
		{"autoescape_false_around", true, false, two(`[{% block a %}{{ "<b>" }}{% endblock %}]`, "{% block a %}{% autoescape false %}{{ super()|upper }}{% endautoescape %}{% endblock %}"), "[&LT;B&GT;]"},
		{"two_blocks", false, false, two("[{% block a %}a{% endblock %}{% block b %}b{% endblock %}]", "{% block a %}{{ super()|upper }}{% endblock %}{% block b %}{{ super() ~ '!' }}{% endblock %}"), "[Ab!]"},
		{"three_levels", false, false, map[string]string{
			"g":     "[{% block a %}g{% endblock %}]",
			"m":     "{% extends 'g' %}{% block a %}{{ super()|upper }}m{% endblock %}",
			"child": "{% extends 'm' %}{% block a %}{{ super() ~ '!' }}{% endblock %}",
		}, "[Gm!]"},
		// the middle block is a single statement
		{"three_levels_single_statement", false, false, map[string]string{
			"g":     "[{% block a %}g{% endblock %}]",
			"m":     "{% extends 'g' %}{% block a %}{{ super()|upper }}{% endblock %}",
			"child": "{% extends 'm' %}{% block a %}{{ super() ~ '!' }}{% endblock %}",
		}, "[G!]"},
		// only the middle template calls super(); the child overrides nothing
		{"middle_only", false, false, map[string]string{
			"g":     "[{% block a %}g{% endblock %}]",
			"m":     "{% extends 'g' %}{% block a %}{{ super()|upper }}{% endblock %}",
			"child": "{% extends 'm' %}{% block b %}{% endblock %}",
		}, "[G]"},
		{"three_levels_twice", false, false, map[string]string{
			"g":     "[{% block a %}g{% endblock %}]",
			"m":     "{% extends 'g' %}{% block a %}{{ super()|upper }}{% endblock %}",
			"child": "{% extends 'm' %}{% block a %}{{ super() ~ super() }}{% endblock %}",
		}, "[GG]"},
		// the parent block renders where and as often as super() is called
		// (the cycler counts renders; strict undefined fails a parent render)
		{"once_expr", false, false, two(cyc, "{% block a %}{{ super() ~ '!' }}{% endblock %}"), "[1!]"},
		{"twice_in_one_statement", false, false, two(cyc, "{% block a %}{{ super() ~ super() }}{% endblock %}"), "[12]"},
		{"each_loop_iteration", false, false, two(cyc, "{% block a %}{% for i in [1, 2] %}{{ super() ~ '' }}{% endfor %}{% endblock %}"), "[12]"},
		{"untaken_ternary_not_rendered", false, true, two(strictBase, "{% block a %}{{ 'x' if true else super() }}{% endblock %}"), "[x]"},
		{"untaken_elif_not_rendered", false, true, two(strictBase, "{% block a %}{% if true %}y{% elif super() %}n{% endif %}{% endblock %}"), "[y]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sl := loader.NewStringLoader(loader.NewDirectTemplateParser())
			for n, s := range tt.templates {
				sl.AddTemplate(n, s)
			}
			opts := []miya.EnvironmentOption{miya.WithLoader(sl), miya.WithAutoEscape(tt.autoescape)}
			if tt.strict {
				opts = append(opts, miya.WithUndefinedBehavior(runtime.UndefinedStrict))
			}
			tmpl, err := miya.NewEnvironment(opts...).GetTemplate("child")
			if err != nil {
				t.Fatal(err)
			}
			got, err := tmpl.Render(miya.NewContext())
			if err != nil {
				t.Fatalf("render: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}

	// guard: outside every block super() has no parent block to render
	// (Jinja2 3.1.6: UndefinedError 'super' is undefined)
	t.Run("top_level_expression_errors", func(t *testing.T) {
		sl := loader.NewStringLoader(loader.NewDirectTemplateParser())
		sl.AddTemplate("base", "{{ super()|upper }}"+base)
		sl.AddTemplate("child", "{% extends 'base' %}{% block a %}{{ super() ~ '!' }}{% endblock %}")
		tmpl, err := miya.NewEnvironment(miya.WithLoader(sl)).GetTemplate("child")
		if err != nil {
			t.Fatal(err)
		}
		if got, err := tmpl.Render(miya.NewContext()); err == nil {
			t.Errorf("want an error, got %q", got)
		}
	})
}
