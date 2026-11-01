package miya_test

import (
	"testing"

	miya "github.com/zipreport/miya"
	"github.com/zipreport/miya/loader"
)

// Issue #40: a block without "scoped" sees only the template's top-level
// variables, not the loop, with or outer block variables around it; "scoped"
// belongs to the block's position, so an override follows the parent's flag.
// Expected values are from Jinja2 3.1.6 (booleans print lowercase in miya).
func TestIssue40UnscopedBlocks(t *testing.T) {
	cases := []struct {
		name      string
		templates map[string]string
		vars      map[string]interface{}
		want      string
	}{
		{"for_unscoped", map[string]string{"main": `{% for i in [1, 2] %}{% block b %}<{{ i }}>{% endblock %}{% endfor %}`}, nil, "<><>"},
		{"for_scoped", map[string]string{"main": `{% for i in [1, 2] %}{% block b scoped %}<{{ i }}>{% endblock %}{% endfor %}`}, nil, "<1><2>"},
		{"loop_defined_unscoped", map[string]string{"main": `{% for i in [1, 2] %}{% block b %}<{{ loop is defined }}>{% endblock %}{% endfor %}`}, nil, "<false><false>"},
		{"loop_scoped", map[string]string{"main": `{% for i in [1, 2] %}{% block b scoped %}<{{ loop.index }}>{% endblock %}{% endfor %}`}, nil, "<1><2>"},
		{"with_unscoped", map[string]string{"main": `{% with v = 5 %}{% block b %}<{{ v }}>{% endblock %}{% endwith %}`}, nil, "<>"},
		{"with_scoped", map[string]string{"main": `{% with v = 5 %}{% block b scoped %}<{{ v }}>{% endblock %}{% endwith %}`}, nil, "<5>"},
		{"outer_set_unscoped", map[string]string{"main": `{% block o %}{% set v = 7 %}[{% block i %}<{{ v }}>{% endblock %}]{% endblock %}`}, nil, "[<>]"},
		{"outer_set_scoped", map[string]string{"main": `{% block o %}{% set v = 7 %}[{% block i scoped %}<{{ v }}>{% endblock %}]{% endblock %}`}, nil, "[<7>]"},
		{"top_set", map[string]string{"main": `{% set v = 3 %}{% block b %}<{{ v }}>{% endblock %}`}, nil, "<3>"},
		{"top_set_after", map[string]string{"main": `{% block b %}<{{ v }}>{% endblock %}{% set v = 3 %}`}, nil, "<>"},
		{"render_var", map[string]string{"main": `{% for name in [1] %}{% block b %}<{{ name }}>{% endblock %}{% endfor %}`}, map[string]interface{}{"name": "N"}, "<N>"},
		{"child_override_in_base_loop", map[string]string{"base": `{% for i in [1, 2] %}{% block b %}<{{ i }}>{% endblock %}{% endfor %}`, "main": `{% extends 'base' %}{% block b %}[{{ i }}]{% endblock %}`}, nil, "[][]"},
		{"child_scoped_override_unscoped_pos", map[string]string{"base": `{% for i in [1, 2] %}{% block b %}<{{ i }}>{% endblock %}{% endfor %}`, "main": `{% extends 'base' %}{% block b scoped %}[{{ i }}]{% endblock %}`}, nil, "[][]"},
		{"child_unscoped_override_scoped_pos", map[string]string{"base": `{% for i in [1, 2] %}{% block b scoped %}<{{ i }}>{% endblock %}{% endfor %}`, "main": `{% extends 'base' %}{% block b %}[{{ i }}]{% endblock %}`}, nil, "[1][2]"},
		{"child_top_set_visible", map[string]string{"base": `{% for i in [1] %}{% block b %}{% endblock %}{% endfor %}`, "main": `{% extends 'base' %}{% set t = 'T' %}{% block b %}[{{ t }}]{% endblock %}`}, nil, "[T]"},
		{"hoisted_for_block", map[string]string{"base": `B`, "main": `{% extends 'base' %}{% for j in [1, 2] %}{% block c %}({{ j }}){% endblock %}{% endfor %}`}, nil, "()()B"},
		{"hoisted_for_block_scoped", map[string]string{"base": `B`, "main": `{% extends 'base' %}{% for j in [1, 2] %}{% block c scoped %}({{ j }}){% endblock %}{% endfor %}`}, nil, "(1)(2)B"},
		{"super_unscoped_in_loop", map[string]string{"base": `{% for i in [1, 2] %}{% block b %}<{{ i }}>{% endblock %}{% endfor %}`, "main": `{% extends 'base' %}{% block b %}[{{ super() }}]{% endblock %}`}, nil, "[<>][<>]"},
		{"super_scoped_pos_in_loop", map[string]string{"base": `{% for i in [1, 2] %}{% block b scoped %}<{{ i }}>{% endblock %}{% endfor %}`, "main": `{% extends 'base' %}{% block b %}[{{ super() }}]{% endblock %}`}, nil, "[<1>][<2>]"},
		{"nested_scoped_in_unscoped_loop", map[string]string{"main": `{% for i in [1] %}{% block o %}{% set v = 7 %}{% block x scoped %}<{{ v }}|{{ i }}>{% endblock %}{% endblock %}{% endfor %}`}, nil, "<7|>"},
		{"include_in_loop_block", map[string]string{"inc": `{% block b %}<{{ i }}>{% endblock %}`, "main": `{% for i in [1, 2] %}{% include 'inc' %}{% endfor %}`}, nil, "<1><2>"},
		{"include_top_set_block", map[string]string{"inc": `{% set v = 4 %}{% block b %}<{{ v }}>{% endblock %}`, "main": `{% include 'inc' %}`}, nil, "<4>"},
		{"macro_arg_block", map[string]string{"main": `{% macro m(a) %}{% block b %}<{{ a }}>{% endblock %}{% endmacro %}{{ m(1) }}`}, nil, "<>"},
		{"call_block", map[string]string{"main": `{% macro m() %}{{ caller() }}{% endmacro %}{% for i in [1] %}{% call m() %}{% block b %}<{{ i }}>{% endblock %}{% endcall %}{% endfor %}`}, nil, "<>"},
		{"block_set_leak", map[string]string{"main": `{% for i in [1, 2] %}{% block b %}{% set i = 9 %}<{{ i }}>{% endblock %}{{ i }}{% endfor %}`}, nil, "<9>1<9>2"},
		{"namespace_write", map[string]string{"main": `{% set ns = namespace(n=0) %}{% for i in [1, 2] %}{% block b %}{% set ns.n = ns.n + 1 %}{% endblock %}{% endfor %}{{ ns.n }}`}, nil, "2"},
		{"import_inside_for", map[string]string{"m": `x`, "main": `{% for i in [1] %}{% import 'm' as mm %}{% block b %}<{{ mm is defined }}>{% endblock %}{% endfor %}`}, nil, "<false>"},
		{"filter_block", map[string]string{"main": `{% for i in [1] %}{% filter upper %}{% block b %}<{{ i }}>{% endblock %}{% endfilter %}{% endfor %}`}, nil, "<>"},
		{"if_in_for", map[string]string{"main": `{% for i in [1, 2] %}{% if true %}{% block b %}<{{ i }}>{% endblock %}{% endif %}{% endfor %}`}, nil, "<><>"},
		{"self_ref_block_in_loop_override", map[string]string{"base": `{% for i in [1, 2] %}{% block b scoped %}<{{ i }}>{% endblock %}{% endfor %}`, "main": `{% extends 'base' %}{% block b scoped %}[{{ i }}]{% endblock %}`}, nil, "[1][2]"},
		{"import_module_block_sees_module_set", map[string]string{"m": `{% set v = 1 %}{% block b %}{% set ns.n = v %}{% endblock %}`, "main": `{% set ns = namespace(n=0) %}{% import 'm' as mm with context %}{{ ns.n }}`}, nil, "1"},
		{"import_module_block_no_loop_var", map[string]string{"m": `{% for i in [5] %}{% block b %}{% set ns.n = i is defined %}{% endblock %}{% endfor %}`, "main": `{% set ns = namespace(n=0) %}{% import 'm' as mm with context %}{{ ns.n }}`}, nil, "false"},
		{"single_template_super_keeps_scoped", map[string]string{"main": `{% for i in [1, 2] %}{% block b scoped %}{% if false %}{{ super() }}{% endif %}<{{ i }}>{% endblock %}{% endfor %}`}, nil, "<1><2>"},
		{"unscoped_in_scoped", map[string]string{"main": `{% for i in [1] %}{% block o scoped %}{% block i %}[{{ i }}]{% endblock %}{% endblock %}{% endfor %}`}, nil, "[1]"},
		{"override_at_scoped_position_nested", map[string]string{"base": `{% for i in [1] %}{% block o scoped %}{% endblock %}{% endfor %}`, "main": `{% extends 'base' %}{% block o %}{% block i %}[{{ i }}]{% endblock %}{% endblock %}`}, nil, "[1]"},
		{"super_nested_in_scoped", map[string]string{"base": `{% for i in [1] %}{% block o scoped %}{% block i %}[{{ i }}]{% endblock %}{% endblock %}{% endfor %}`, "main": `{% extends 'base' %}{% block o %}<{{ super() }}>{% endblock %}`}, nil, "<[1]>"},
		{"super_expr_nested_in_scoped", map[string]string{"base": `{% for i in ['q'] %}{% block o scoped %}{% block i %}[{{ i }}]{% endblock %}{% endblock %}{% endfor %}`, "main": `{% extends 'base' %}{% block o %}<{{ super()|upper }}>{% endblock %}`}, nil, "<[Q]>"},
		{"macro_in_scoped", map[string]string{"main": `{% for i in [1] %}{% block o scoped %}{% macro m() %}{% block b %}[{{ i }}]{% endblock %}{% endmacro %}{{ m() }}{% endblock %}{% endfor %}`}, nil, "[1]"},
		{"caller_in_scoped", map[string]string{"main": `{% macro m() %}{{ caller() }}{% endmacro %}{% for i in [1] %}{% block o scoped %}{% call m() %}{% block b %}[{{ i }}]{% endblock %}{% endcall %}{% endblock %}{% endfor %}`}, nil, "[1]"},
		{"inner_loop_in_scoped", map[string]string{"main": `{% for i in [1] %}{% block o scoped %}{% for j in [2] %}{% block i %}[{{ i }}{{ j }}]{% endblock %}{% endfor %}{% endblock %}{% endfor %}`}, nil, "[1]"},
		{"scoped_in_autoescape_false", map[string]string{"main": `{% autoescape false %}{% for i in [1] %}{% block b scoped %}{{ h }}{{ i }}{% endblock %}{% endfor %}{% endautoescape %}`}, map[string]interface{}{"h": "<b>"}, "&lt;b&gt;1"},
		{"unscoped_in_autoescape_false", map[string]string{"main": `{% autoescape false %}{% block b %}{{ h }}{% endblock %}{% endautoescape %}`}, map[string]interface{}{"h": "<b>"}, "&lt;b&gt;"},
		{"override_scoped_pos_autoescape_false", map[string]string{"base": `{% autoescape false %}{% for i in [1] %}{% block b scoped %}{% endblock %}{% endfor %}{% endautoescape %}`, "main": `{% extends 'base' %}{% block b %}{{ h }}{{ i }}{% endblock %}`}, map[string]interface{}{"h": "<b>"}, "&lt;b&gt;1"},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			sl := loader.NewStringLoader(loader.NewDirectTemplateParser())
			for n, s := range tt.templates {
				sl.AddTemplate(n, s)
			}
			env := miya.NewEnvironment(miya.WithLoader(sl))
			tmpl, err := env.GetTemplate("main")
			if err != nil {
				t.Fatal(err)
			}
			ctx := miya.NewContext()
			for k, v := range tt.vars {
				ctx.Set(k, v)
			}
			got, err := tmpl.Render(ctx)
			if err != nil {
				t.Fatalf("render: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
