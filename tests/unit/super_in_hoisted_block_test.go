package miya_test

import (
	"testing"

	miya "github.com/zipreport/miya"
	"github.com/zipreport/miya/loader"
)

// Issue #15: super() in a block inside an extending template's top-level for
// or with renders the parent block, both where the block is hoisted and in
// the parent's slot.
func TestIssue15SuperInHoistedBlock(t *testing.T) {
	const base = "{% block a %}a{% endblock %}"
	const ext = "{% extends 'base' %}"
	two := func(b, child string) map[string]string {
		return map[string]string{"base": b, "child": ext + child}
	}
	// want values are Jinja2 3.1.6's output
	tests := []struct {
		name      string
		templates map[string]string
		want      string
	}{
		{"claim1_for", two(base, "{% for i in [1] %}{% block a %}[{{ super() }}]{% endblock %}{% endfor %}"), "[a][a]"},
		{"claim2_for_expr", two(base, "{% for i in [1] %}{% block a %}{{ super()|upper }}{% endblock %}{% endfor %}"), "AA"},
		{"claim3_with", two(base, "{% with %}{% block a %}[{{ super() }}]{% endblock %}{% endwith %}"), "[a][a]"},
		{"claim4_grandchild_override", map[string]string{
			"base":  base,
			"mid":   "{% extends 'base' %}{% for i in [1] %}{% block a %}m{% endblock %}{% endfor %}",
			"child": "{% extends 'mid' %}{% block a %}[{{ super() }}]{% endblock %}",
		}, "[m][m]"},
		// further shapes
		{"whole_super", two(base, "{% for i in [1] %}{% block a %}{{ super() }}{% endblock %}{% endfor %}"), "aa"},
		{"set_super", two(base, "{% for i in [1] %}{% block a %}{% set x = super() %}[{{ x }}]{% endblock %}{% endfor %}"), "[a][a]"},
		{"two_iterations", two(base, "{% for i in [1, 2] %}{% block a %}[{{ super() }}]{% endblock %}{% endfor %}"), "[a][a][a]"},
		{"parent_with_nested_block", two("{% block a %}<{% block i %}i{% endblock %}>{% endblock %}",
			"{% for x in [1] %}{% block a %}[{{ super() }}]{% endblock %}{% endfor %}"), "[<i>][<i>]"},
		// guard: no super()
		{"no_super", two(base, "{% for i in [1] %}{% block a %}X{% endblock %}{% endfor %}"), "XX"},
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

	// guards: super() outside every block in top-level code has no parent
	// block (Jinja2 3.1.6: UndefinedError 'super' is undefined)
	errs := []struct{ name, child string }{
		{"macro", "{% macro m() %}<{{ super() }}>{% endmacro %}{% block a %}{{ m() }}{% endblock %}"},
		{"call_block", "{% macro m() %}{{ caller() }}{% endmacro %}{% call m() %}<{{ super() }}>{% endcall %}"},
		{"set_block", "{% set x %}<{{ super() }}>{% endset %}{% block a %}{{ x }}{% endblock %}"},
		{"macro_no_blocks", "{% macro m() %}<{{ super() }}>{% endmacro %}{% set y = m() %}"},
	}
	// the same in the root template, which rendered '' before
	rootErrs := []struct{ name, base, child string }{
		{"root_macro", "{% macro m() %}<{{ super() }}>{% endmacro %}{% block a %}{{ m() }}{% endblock %}", ""},
		{"root_set_block", "{% set x %}<{{ super() }}>{% endset %}{% block a %}a{% endblock %}", "{% block a %}X{% endblock %}"},
		{"root_call_block", "{% macro m() %}{{ caller() }}{% endmacro %}{% call m() %}<{{ super() }}>{% endcall %}{% block a %}a{% endblock %}", ""},
		{"root_filter_block", "{% filter upper %}{{ super() }}{% endfilter %}{% block a %}a{% endblock %}", ""},
	}
	for _, tt := range rootErrs {
		t.Run("outside_block_errors_"+tt.name, func(t *testing.T) {
			sl := loader.NewStringLoader(loader.NewDirectTemplateParser())
			sl.AddTemplate("base", tt.base)
			sl.AddTemplate("child", ext+tt.child)
			tmpl, err := miya.NewEnvironment(miya.WithLoader(sl)).GetTemplate("child")
			if err != nil {
				t.Fatal(err)
			}
			if got, err := tmpl.Render(miya.NewContext()); err == nil {
				t.Errorf("want an error, got %q", got)
			}
		})
	}
	for _, tt := range errs {
		t.Run("outside_block_errors_"+tt.name, func(t *testing.T) {
			sl := loader.NewStringLoader(loader.NewDirectTemplateParser())
			sl.AddTemplate("base", base)
			sl.AddTemplate("child", ext+tt.child)
			tmpl, err := miya.NewEnvironment(miya.WithLoader(sl)).GetTemplate("child")
			if err != nil {
				t.Fatal(err)
			}
			if got, err := tmpl.Render(miya.NewContext()); err == nil {
				t.Errorf("want an error, got %q", got)
			}
		})
	}
}
