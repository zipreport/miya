package miya_test

import (
	"testing"

	miya "github.com/zipreport/miya"
	"github.com/zipreport/miya/loader"
)

// Issue #14: a block has its own variable scope, and super() renders the
// parent block in a scope of its own.
func TestIssue14BlockScope(t *testing.T) {
	two := func(b, child string) map[string]string {
		return map[string]string{"base": b, "child": "{% extends 'base' %}" + child}
	}
	one := func(src string) map[string]string { return map[string]string{"child": src} }
	const setsP = "{% block a %}{% set v = 1 %}p{% endblock %}"
	const readsV = "{% block a %}[{{ v }}]{% endblock %}"
	// want values are Jinja2 3.1.6's output
	tests := []struct {
		name      string
		templates map[string]string
		want      string
	}{
		{"claim1_parent_set_not_in_child", two(setsP, "{% block a %}{{ super() }}[{{ v }}]{% endblock %}"), "p[]"},
		{"claim2_parent_set_not_in_child_expr", two(setsP, "{% block a %}{{ super()|upper }}[{{ v }}]{% endblock %}"), "P[]"},
		{"claim3_child_set_not_in_parent", two(readsV, "{% block a %}{% set v = 1 %}{{ super() }}{% endblock %}"), "[]"},
		{"claim4_child_set_not_in_parent_expr", two(readsV, "{% block a %}{% set v = 1 %}{{ super()|upper }}{% endblock %}"), "[]"},
		{"claim5_parent_set_keeps_child_value", two("{% block a %}{% set v = 2 %}p{% endblock %}", "{% block a %}{% set v = 1 %}{{ super() }}[{{ v }}]{% endblock %}"), "p[1]"},
		{"claim6_no_leak_after_block", one("{% block a %}{% set v = 1 %}{% endblock %}[{{ v }}]"), "[]"},
		{"claim7_child_block_no_leak", two("{% block a %}{% endblock %}[{{ v }}]", "{% block a %}{% set v = 1 %}{% endblock %}"), "[]"},
		// further shapes
		{"no_leak_into_next_block", one("{% block a %}{% set v = 1 %}{% endblock %}{% block b %}[{{ v }}]{% endblock %}"), "[]"},
		{"shadowed_value_restored", one("{% set v = 0 %}{% block a %}{% set v = 1 %}{{ v }}{% endblock %}[{{ v }}]"), "1[0]"},
		{"macro_is_block_local", one("{% block a %}{% macro m() %}m{% endmacro %}{{ m() }}{% endblock %}[{{ 'y' if m is defined else 'n' }}]"), "m[n]"},
		{"three_levels", map[string]string{
			"base":  "{% block a %}{% set v = 3 %}[{{ v }}]{% endblock %}",
			"mid":   "{% extends 'base' %}{% block a %}{% set v = 2 %}{{ super() }}({{ v }}){% endblock %}",
			"child": "{% extends 'mid' %}{% block a %}{% set v = 1 %}{{ super()|upper }}<{{ v }}>{% endblock %}",
		}, "[3](2)<1>"},
		{"middle_set_not_in_grandparent", map[string]string{
			"base":  "{% block a %}[{{ v }}]{% endblock %}",
			"mid":   "{% extends 'base' %}{% block a %}{% set v = 2 %}{{ super() }}{% endblock %}",
			"child": "{% extends 'mid' %}{% block a %}{{ super() }}{% endblock %}",
		}, "[]"},
		{"super_in_for", two("{% block a %}{% set v = 2 %}p{% endblock %}", "{% block a %}{% for i in [1] %}{% set v = 1 %}{{ super() }}{{ v }}{% endfor %}{% endblock %}"), "p1"},
		{"scoped_nested_block_under_super", two("{% block o %}{% set v = 1 %}{% block i scoped %}<{{ v }}>{% endblock %}{% endblock %}", "{% block o %}{{ super() }}{% endblock %}"), "<1>"},
		{"scoped_nested_block_under_super_expr", two("{% block o %}{% set v = 1 %}{% block i scoped %}<{{ v }}>{% endblock %}{% endblock %}", "{% block o %}{{ super()|upper }}{% endblock %}"), "<1>"},
		// guards: what a block still sees and shares
		{"top_level_set_visible", one("{% set v = 1 %}{% block a %}[{{ v }}]{% endblock %}"), "[1]"},
		{"child_top_level_set_visible_to_both", two("{% block a %}{{ v }}{% endblock %}", "{% set v = 1 %}{% block a %}[{{ v }}]{{ super() }}{% endblock %}"), "[1]1"},
		{"namespace_write_shared", one("{% set ns = namespace(v=0) %}{% block a %}{% set ns.v = 1 %}{% endblock %}[{{ ns.v }}]"), "[1]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sl := loader.NewStringLoader(loader.NewDirectTemplateParser())
			for n, s := range tt.templates {
				sl.AddTemplate(n, s)
			}
			tmpl, err := miya.NewEnvironment(miya.WithLoader(sl)).GetTemplate("child")
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
}
