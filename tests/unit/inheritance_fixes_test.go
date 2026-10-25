package miya_test

import (
	"strings"
	"testing"

	miya "github.com/zipreport/miya"
	"github.com/zipreport/miya/loader"
)

// renderInherited loads "main" from templates through a StringLoader and
// renders it; expected values are Jinja2 3.1.6's output.
func renderInherited(t *testing.T, templates map[string]string, data map[string]interface{}) (string, error) {
	t.Helper()
	sl := loader.NewStringLoader(loader.NewDirectTemplateParser())
	for name, src := range templates {
		sl.AddTemplate(name, src)
	}
	env := miya.NewEnvironment(miya.WithLoader(sl), miya.WithAutoEscape(false))
	tmpl, err := env.GetTemplate("main")
	if err != nil {
		return "", err
	}
	return tmpl.Render(miya.NewContextFrom(data))
}

func TestInheritanceFixes(t *testing.T) {
	const ext = "{% extends 'base' %}"
	tests := []struct {
		name      string
		templates map[string]string
		data      map[string]interface{}
		want      string
	}{
		{"block_in_if", map[string]string{"main": ext + "{% block b %}X{% endblock %}", "base": "{% if f %}[{% block b %}d{% endblock %}]{% endif %}"}, map[string]interface{}{"f": true}, "[X]"},
		{"block_in_elif", map[string]string{"main": ext + "{% block b %}X{% endblock %}", "base": "{% if x == 1 %}one{% elif x == 2 %}[{% block b %}d{% endblock %}]{% endif %}"}, map[string]interface{}{"x": 2}, "[X]"},
		{"elif_kept_in_base", map[string]string{"main": ext + "{% block b %}B{% endblock %}", "base": "{% if x == 1 %}one{% elif x == 2 %}two{% else %}other{% endif %}|{% block b %}{% endblock %}"}, map[string]interface{}{"x": 2}, "two|B"},
		{"block_in_for", map[string]string{"main": ext + "{% block b %}X{% endblock %}", "base": "{% for i in [1, 2] %}[{% block b %}d{% endblock %}]{% endfor %}"}, nil, "[X][X]"},
		{"block_in_for_super", map[string]string{"main": ext + "{% block b %}{{ super() }}+{% endblock %}", "base": "{% for i in [1, 2] %}[{% block b %}d{% endblock %}]{% endfor %}"}, nil, "[d+][d+]"},
		{"block_in_with", map[string]string{"main": ext + "{% block b %}X{% endblock %}", "base": "{% with a = 1 %}[{% block b %}d{% endblock %}]{% endwith %}"}, nil, "[X]"},
		{"block_in_filter", map[string]string{"main": ext + "{% block b %}x{% endblock %}", "base": "{% filter upper %}[{% block b %}d{% endblock %}]{% endfilter %}"}, nil, "[X]"},
		{"block_in_set_block", map[string]string{"main": ext + "{% block b %}X{% endblock %}", "base": "{% set v %}[{% block b %}d{% endblock %}]{% endset %}{{ v }}"}, nil, "[X]"},
		{"block_in_macro", map[string]string{"main": ext + "{% block b %}X{% endblock %}", "base": "{% macro m() %}[{% block b %}d{% endblock %}]{% endmacro %}{{ m() }}"}, nil, "[X]"},
		{"block_in_call", map[string]string{"main": ext + "{% block b %}X{% endblock %}", "base": "{% macro m() %}<{{ caller() }}>{% endmacro %}{% call m() %}[{% block b %}d{% endblock %}]{% endcall %}"}, nil, "<[X]>"},
		{"scoped_block", map[string]string{"main": ext + "{% block b %}{{ i }}{% endblock %}", "base": "{% for i in [1, 2] %}[{% block b scoped %}d{% endblock %}]{% endfor %}"}, nil, "[1][2]"},
		{"child_set", map[string]string{"main": ext + "{% set title = 'T' %}", "base": "{{ title }}|{% block b %}{{ title }}{% endblock %}"}, nil, "T|T"},
		{"child_set_after_block", map[string]string{"main": ext + "{% block b %}{{ t }}{% endblock %}{% set t = 'C' %}", "base": "{{ t }}|{% block b %}{% endblock %}"}, nil, "C|C"},
		{"child_set_in_if", map[string]string{"main": ext + "{% if true %}{% set t = 'C' %}{% endif %}", "base": "[{{ t }}]"}, nil, "[C]"},
		{"child_macro", map[string]string{"main": ext + "{% macro m() %}M{% endmacro %}{% block b %}{{ m() }}{% endblock %}", "base": "[{% block b %}{% endblock %}]"}, nil, "[M]"},
		{"parent_set_wins", map[string]string{"main": ext + "{% set t = 'C' %}", "base": "{% set t = 'P' %}{{ t }}|{% block b %}{{ t }}{% endblock %}"}, nil, "P|P"},
		{"child_text_before_extends", map[string]string{"main": "A" + ext, "base": "[base]"}, nil, "A[base]"},
		{"child_text_after_extends", map[string]string{"main": ext + "B{{ x }}", "base": "[base]"}, map[string]interface{}{"x": 1}, "[base]"},
		{"child_block_not_run_at_top_level", map[string]string{"main": ext + "{% block a %}{% set t = 'X' %}{% endblock %}", "base": "[{{ t }}]{% block a %}{% endblock %}"}, nil, "[]"},
		{"three_levels", map[string]string{"main": "{% extends 'mid' %}{% block inner %}C{% endblock %}", "mid": ext + "{% block outer %}<{% block inner %}m{% endblock %}>{% endblock %}", "base": "[{% block outer %}o{% endblock %}]"}, nil, "[<C>]"},
		{"child_output_after_extends_not_run", map[string]string{"main": ext + "{{ nofn() }}{{ 1 / z }}{% for i in [1] %}{{ nofn() }}{% endfor %}{% block b %}B{% endblock %}", "base": "[{% block b %}{% endblock %}]"}, map[string]interface{}{"z": 0}, "[B]"},
		{"include_after_extends", map[string]string{"main": ext + "{% include 'inc' %}", "base": "[base]", "inc": "INC"}, nil, "INC[base]"},
		{"call_after_extends", map[string]string{"main": ext + "{% macro m() %}<{{ caller() }}>{% endmacro %}{% call m() %}T{% endcall %}", "base": "[base]"}, nil, "<T>[base]"},
		{"filter_block_after_extends", map[string]string{"main": ext + "{% filter upper %}x{% block b %}y{% endblock %}{% endfilter %}", "base": "[{% block b %}{% endblock %}]"}, nil, "Y[y]"},
		{"block_in_for_after_extends", map[string]string{"main": ext + "{% for i in [1, 2] %}{% block z %}Z{% endblock %}{% endfor %}", "base": "[base]"}, nil, "ZZ[base]"},
		{"block_in_if_after_extends_skipped", map[string]string{"main": ext + "{% if true %}{% block z %}Z{% endblock %}{% endif %}", "base": "[base]"}, nil, "[base]"},
		{"block_after_extends_takes_grandchild_override", map[string]string{"main": "{% extends 'mid' %}{% block z %}G{% endblock %}", "mid": ext + "{% for i in [1] %}{% block z %}Z{% endblock %}{% endfor %}", "base": "[{% block z %}b{% endblock %}]"}, nil, "G[G]"},
		{"nested_child_block_overrides_base", map[string]string{"main": ext + "{% block outer %}<{% block inner %}I{% endblock %}>{% endblock %}", "base": "[{% block outer %}o{% endblock %}]{% if f %}{% block inner %}i{% endblock %}{% endif %}"}, map[string]interface{}{"f": true}, "[<I>]I"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := renderInherited(t, tt.templates, tt.data)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}

	for name, src := range map[string]string{
		"duplicate_block_sibling": "{% block a %}1{% endblock %}{% block a %}2{% endblock %}",
		"duplicate_block_nested":  "{% block a %}{% block a %}2{% endblock %}{% endblock %}",
	} {
		t.Run(name, func(t *testing.T) {
			_, err := miya.NewEnvironment().FromString(src)
			if err == nil || !strings.Contains(err.Error(), "block 'a' defined twice") {
				t.Errorf("%q: want error \"block 'a' defined twice\", got %v", src, err)
			}
		})
	}

	// A block before {% extends %} in a middle template renders the child's
	// override; rendering the middle template afterwards must show its own.
	t.Run("block_before_extends_in_middle_template", func(t *testing.T) {
		sl := loader.NewStringLoader(loader.NewDirectTemplateParser())
		sl.AddTemplate("main", "{% extends 'mid' %}{% block b %}C{% endblock %}")
		sl.AddTemplate("mid", "{% for i in [1] %}{% block b %}m{% endblock %}{% endfor %}{% extends 'base' %}")
		sl.AddTemplate("base", "[{% block b %}B{% endblock %}]")
		env := miya.NewEnvironment(miya.WithLoader(sl), miya.WithAutoEscape(false))
		for _, step := range []struct{ name, want string }{{"main", "C[C]"}, {"mid", "m[m]"}} {
			tmpl, err := env.GetTemplate(step.name)
			if err != nil {
				t.Fatal(err)
			}
			got, err := tmpl.Render(miya.NewContext())
			if err != nil {
				t.Fatal(err)
			}
			if got != step.want {
				t.Errorf("%s: got %q, want %q", step.name, got, step.want)
			}
		}
	})

	// Overriding must not change the parent templates: render the child, then
	// each parent on its own.
	t.Run("parents_not_modified", func(t *testing.T) {
		sl := loader.NewStringLoader(loader.NewDirectTemplateParser())
		sl.AddTemplate("main", "{% extends 'mid' %}{% block inner %}C{% endblock %}{% block w %}W{% endblock %}")
		sl.AddTemplate("mid", "{% extends 'base' %}{% block outer %}<{% block inner %}m{% endblock %}>{% endblock %}")
		sl.AddTemplate("base", "[{% block outer %}o{% endblock %}]{% if x == 1 %}one{% elif x == 2 %}{% with a = 1 %}({% block w %}w{% endblock %}){% endwith %}{% endif %}")
		env := miya.NewEnvironment(miya.WithLoader(sl), miya.WithAutoEscape(false))
		for _, step := range []struct{ name, want string }{{"main", "[<C>](W)"}, {"mid", "[<m>](w)"}, {"base", "[o](w)"}} {
			tmpl, err := env.GetTemplate(step.name)
			if err != nil {
				t.Fatal(err)
			}
			got, err := tmpl.Render(miya.NewContextFrom(map[string]interface{}{"x": 2}))
			if err != nil {
				t.Fatal(err)
			}
			if got != step.want {
				t.Errorf("%s: got %q, want %q", step.name, got, step.want)
			}
		}
	})

	// A template without a parent takes the replaceSuperInNode path; a
	// recursive loop must stay recursive there.
	t.Run("super_without_parent_keeps_recursive", func(t *testing.T) {
		tree := []interface{}{map[string]interface{}{"v": 1, "c": []interface{}{
			map[string]interface{}{"v": 2, "c": []interface{}{map[string]interface{}{"v": 3}}},
		}}}
		got, err := renderInherited(t, map[string]string{
			"main": "{% for n in tree recursive %}{{ n.v }}{% if n.c %}({{ loop(n.c) }}){% endif %}{% endfor %}{% block b %}{{ super() }}{% endblock %}",
		}, map[string]interface{}{"tree": tree})
		if err != nil {
			t.Fatal(err)
		}
		if got != "1(2(3))" {
			t.Errorf("got %q, want %q", got, "1(2(3))")
		}
	})
}

// sourceLoader serves templates through GetSource only, as zipreport-go's loader does.
type sourceLoader map[string]string

func (m sourceLoader) GetSource(name string) (string, error) {
	if src, ok := m[name]; ok {
		return src, nil
	}
	return "", &loaderError{name}
}
func (m sourceLoader) IsCached(string) bool             { return false }
func (m sourceLoader) ListTemplates() ([]string, error) { return nil, nil }

type loaderError struct{ name string }

func (e *loaderError) Error() string { return "template not found: " + e.name }

// countingStringer counts how often something formats it.
type countingStringer struct{ n *int }

func (c countingStringer) String() string { *c.n++; return "x" }

func TestInheritanceCaches(t *testing.T) {
	t.Run("from_string_templates_do_not_share_a_hierarchy", func(t *testing.T) {
		env := miya.NewEnvironment(miya.WithLoader(sourceLoader{
			"b1": "{% block b %}{% endblock %}|ONE",
			"b2": "{% block b %}{% endblock %}|TWO",
		}), miya.WithAutoEscape(false))
		var got []string
		for _, src := range []string{
			`{% extends "b1" %}{% block b %}one{% endblock %}`,
			`{% extends "b2" %}{% block b %}two{% endblock %}`,
		} {
			tmpl, err := env.FromString(src)
			if err != nil {
				t.Fatal(err)
			}
			out, err := tmpl.Render(nil)
			if err != nil {
				t.Fatal(err)
			}
			got = append(got, out)
		}
		if got[0] != "one|ONE" || got[1] != "two|TWO" {
			t.Errorf("got %q, want [\"one|ONE\" \"two|TWO\"]", got)
		}
	})

	t.Run("clear_cache_reloads_parent", func(t *testing.T) {
		templates := sourceLoader{"base": "[{% block b %}B{% endblock %}]", "main": "{% extends 'base' %}{% block b %}C{% endblock %}"}
		env := miya.NewEnvironment(miya.WithLoader(templates), miya.WithAutoEscape(false))
		render := func() string {
			tmpl, err := env.GetTemplate("main")
			if err != nil {
				t.Fatal(err)
			}
			out, err := tmpl.Render(nil)
			if err != nil {
				t.Fatal(err)
			}
			return out
		}
		if got := render(); got != "[C]" {
			t.Fatalf("first render: got %q, want %q", got, "[C]")
		}
		templates["base"] = "<{% block b %}B{% endblock %}>"
		env.ClearCache()
		if got := render(); got != "<C>" {
			t.Errorf("after ClearCache: got %q, want %q", got, "<C>")
		}
	})

	t.Run("context_is_not_formatted", func(t *testing.T) {
		n := 0
		_, err := renderInherited(t, map[string]string{
			"main": "{% extends 'base' %}{% block b %}C{% endblock %}",
			"base": "[{% block b %}{% endblock %}]",
		}, map[string]interface{}{"unused": countingStringer{&n}})
		if err != nil {
			t.Fatal(err)
		}
		if n != 0 {
			t.Errorf("an unused context value was formatted %d times; want 0", n)
		}
	})
}

// A dynamic {% extends %} follows the environment's undefined behaviour, as
// Jinja2 does: in strict mode an undefined name is an error naming it.
func TestInheritanceDynamicExtendsStrict(t *testing.T) {
	sl := loader.NewStringLoader(loader.NewDirectTemplateParser())
	sl.AddTemplate("base", "B[{% block c %}{% endblock %}]")
	sl.AddTemplate("main", "{% extends nope %}{% block c %}C{% endblock %}")
	env := miya.NewEnvironment(miya.WithLoader(sl), miya.WithStrictUndefined(true))
	tmpl, err := env.GetTemplate("main")
	if err != nil {
		t.Fatal(err)
	}
	out, err := tmpl.Render(miya.NewContext())
	if err == nil || !strings.Contains(err.Error(), "UndefinedError") || !strings.Contains(err.Error(), "nope") {
		t.Errorf("strict {%% extends nope %%}: got %q, %v; want an UndefinedError naming nope", out, err)
	}
}
