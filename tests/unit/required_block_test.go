package miya_test

import (
	"strings"
	"testing"

	miya "github.com/zipreport/miya"
	"github.com/zipreport/miya/loader"
)

// Issue #42: {% block name required %}. Rendering a required block fails
// unless another template in the chain also defines it, and a required block
// may hold only comments and whitespace. Expected values are from Jinja2 3.1.6.
func TestIssue42RequiredBlocks(t *testing.T) {
	render := func(templates map[string]string, name string) (string, error) {
		sl := loader.NewStringLoader(loader.NewDirectTemplateParser())
		for n, s := range templates {
			sl.AddTemplate(n, s)
		}
		env := miya.NewEnvironment(miya.WithLoader(sl))
		tmpl, err := env.GetTemplate(name)
		if err != nil {
			return "", err
		}
		return tmpl.Render(miya.NewContext())
	}

	renders := []struct {
		name      string
		templates map[string]string
		want      string
	}{
		{"override", map[string]string{"base": `[{% block x required %}{% endblock %}]`, "main": `{% extends 'base' %}{% block x %}X{% endblock %}`}, "[X]"},
		{"scoped_required", map[string]string{"base": `[{% block x scoped required %}{% endblock %}]`, "main": `{% extends 'base' %}{% block x %}X{% endblock %}`}, "[X]"},
		{"body_ws_comment", map[string]string{"base": `[{% block x required %}  
 {# c #} {% endblock %}]`, "main": `{% extends 'base' %}{% block x %}X{% endblock %}`}, "[X]"},
		{"body_raw_ws", map[string]string{"base": `[{% block x required %}{% raw %} {% endraw %}{% endblock %}]`, "main": `{% extends 'base' %}{% block x %}X{% endblock %}`}, "[X]"},
		{"super_into_required", map[string]string{"base": `[{% block x required %}{% endblock %}]`, "main": `{% extends 'base' %}{% block x %}X{{ super() }}{% endblock %}`}, "[X]"},
		{"super_in_untaken", map[string]string{"base": `[{% block x required %}{% endblock %}]`, "main": `{% extends 'base' %}{% block x %}X{% if false %}{{ super() }}{% endif %}{% endblock %}`}, "[X]"},
		{"middle_required_override", map[string]string{"base": `[{% block x %}B{% endblock %}]`, "main": `{% extends 'mid' %}{% block x %}X{% endblock %}`, "mid": `{% extends 'base' %}{% block x required %}{% endblock %}`}, "[X]"},
		{"middle_required_no_override", map[string]string{"base": `[{% block x %}B{% endblock %}]`, "main": `{% extends 'mid' %}`, "mid": `{% extends 'base' %}{% block x required %}{% endblock %}`}, "[]"},
		{"grandchild_overrides", map[string]string{"base": `[{% block x required %}{% endblock %}]`, "main": `{% extends 'mid' %}{% block x %}X{% endblock %}`, "mid": `{% extends 'base' %}`}, "[X]"},
		{"untaken_branch", map[string]string{"main": `[{% if false %}{% block x required %}{% endblock %}{% endif %}]`}, "[]"},
		{"in_empty_for", map[string]string{"base": `[{% for i in [] %}{% block x required %}{% endblock %}{% endfor %}]`, "main": `{% extends 'base' %}`}, "[]"},
		{"trim", map[string]string{"base": `[ {%- block x required -%} {%- endblock -%} ]`, "main": `{% extends 'base' %}{% block x %}X{% endblock %}`}, "[X]"},
		{"block_named_required", map[string]string{"main": `[{% block required %}R{% endblock %}]`}, "[R]"},
		{"named_endblock", map[string]string{"base": `[{% block x required %}{% endblock x %}]`, "main": `{% extends 'base' %}{% block x %}X{% endblock %}`}, "[X]"},
		{"include_child", map[string]string{"base": `[{% block x required %}{% endblock %}]`, "c": `{% extends 'base' %}{% block x %}X{% endblock %}`, "main": `a{% include 'c' %}`}, "a[X]"},
		{"super_expr_nested_required", map[string]string{"base": `[{% block outer %}<{% block inner %}b{% endblock %}>{% endblock %}]`, "main": `{% extends 'base' %}{% block outer %}{{ super()|upper }}{% endblock %}{% block inner required %}{% endblock %}`}, "[<>]"},
		{"super_set_nested_required", map[string]string{"base": `[{% block outer %}<{% block inner %}b{% endblock %}>{% endblock %}]`, "main": `{% extends 'base' %}{% block outer %}{% set s = super() %}{{ s }}{% endblock %}{% block inner required %}{% endblock %}`}, "[<>]"},
		{"super_expr_middle_redeclare", map[string]string{"base": `[{% block outer %}<{% block inner required %}{% endblock %}>{% endblock %}]`, "main": `{% extends 'mid' %}{% block outer %}{{ super()|upper }}{% endblock %}`, "mid": `{% extends 'base' %}{% block inner required %}{% endblock %}`}, "[<>]"},
		{"super_inline_nested_required", map[string]string{"base": `[{% block outer %}<{% block inner %}b{% endblock %}>{% endblock %}]`, "main": `{% extends 'base' %}{% block outer %}{{ super() }}{% endblock %}{% block inner required %}{% endblock %}`}, "[<>]"},
		{"hoisted_for_shared", map[string]string{"base": `[{% block x %}B{% endblock %}]`, "main": `{% extends 'base' %}{% for i in [1] %}{% block x required %}{% endblock %}{% endfor %}`}, "[]"},
		{"hoisted_with_shared", map[string]string{"base": `[{% block x %}B{% endblock %}]`, "main": `{% extends 'base' %}{% with a = 1 %}{% block x required %}{% endblock %}{% endwith %}`}, "[]"},
		{"import_hoisted_shared", map[string]string{"base": `[{% block x %}B{% endblock %}]`, "main": `{% import 'mod' as m %}{{ m.f() }}`, "mod": `{% extends 'base' %}{% for i in [1] %}{% block x required %}{% endblock %}{% endfor %}{% macro f() %}F{% endmacro %}`}, "F"},
		{"ws_nbsp_override", map[string]string{"base": "[{% block x required %}\u00a0\u3000\u2028{% endblock %}]", "main": `{% extends 'base' %}{% block x %}X{% endblock %}`}, "[X]"},
		{"dynamic_extends_middle_required", map[string]string{"base": `[{% block x %}B{% endblock %}]`, "main": `{% extends 'mid' ~ '' %}`, "mid": `{% extends 'base' %}{% block x required %}{% endblock %}`}, "[]"},
	}
	for _, tt := range renders {
		t.Run(tt.name, func(t *testing.T) {
			got, err := render(tt.templates, "main")
			if err != nil {
				t.Fatalf("render: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}

	// "syntax" is Jinja2's TemplateSyntaxError, "runtime" its
	// TemplateRuntimeError: Required block 'x' not found
	errs := []struct {
		name      string
		templates map[string]string
		kind      string
	}{
		{"super_and_required_in_root", map[string]string{"main": `[{% block a %}{% if false %}{{ super() }}{% endif %}{% block x required %}{% endblock %}{% endblock %}]`}, "runtime"},
		{"direct", map[string]string{"main": `[{% block x required %}{% endblock %}]`}, "runtime"},
		{"no_override", map[string]string{"base": `[{% block x required %}{% endblock %}]`, "main": `{% extends 'base' %}`}, "runtime"},
		{"required_scoped", map[string]string{"main": `[{% block x required scoped %}{% endblock %}]`}, "syntax"},
		{"body_text", map[string]string{"main": `[{% block x required %}body{% endblock %}]`}, "syntax"},
		{"body_var", map[string]string{"main": `[{% block x required %}{{ '' }}{% endblock %}]`}, "syntax"},
		{"body_set", map[string]string{"main": `[{% block x required %}{% set v = 1 %}{% endblock %}]`}, "syntax"},
		{"body_nested_block", map[string]string{"main": `[{% block x required %}{% block y %}{% endblock %}{% endblock %}]`}, "syntax"},
		{"in_for", map[string]string{"base": `[{% for i in [1] %}{% block x required %}{% endblock %}{% endfor %}]`, "main": `{% extends 'base' %}`}, "runtime"},
		{"required_twice", map[string]string{"main": `[{% block x required required %}{% endblock %}]`}, "syntax"},
		{"include_unoverridden", map[string]string{"base": `[{% block x required %}{% endblock %}]`, "main": `a{% include 'base' %}`}, "runtime"},
		{"import_module", map[string]string{"base": `{% macro f() %}F{% endmacro %}[{% block x required %}{% endblock %}]`, "main": `{% import 'base' as m %}ok{{ m.f() }}`}, "runtime"},
		{"nested_in_override", map[string]string{"base": `[{% block outer %}O{% endblock %}]`, "main": `{% extends 'base' %}{% block outer %}<{% block x required %}{% endblock %}>{% endblock %}`}, "runtime"},
		{"required_other_word", map[string]string{"main": `[{% block x requird %}{% endblock %}]`}, "syntax"},
		{"hoisted_for_alone", map[string]string{"base": `[{% block x %}B{% endblock %}]`, "main": `{% extends 'base' %}{% for i in [1] %}{% block y required %}{% endblock %}{% endfor %}`}, "runtime"},
		{"import_hoisted_alone", map[string]string{"base": `[{% block x %}B{% endblock %}]`, "main": `{% import 'mod' as m %}{{ m.f() }}`, "mod": `{% extends 'base' %}{% for i in [1] %}{% block y required %}{% endblock %}{% endfor %}{% macro f() %}F{% endmacro %}`}, "runtime"},
		{"ws_x1c", map[string]string{"main": "[{% block x required %}\u001c\u001f {% endblock %}]"}, "runtime"},
		{"ws_zwsp", map[string]string{"main": "[{% block x required %}\u200b{% endblock %}]"}, "syntax"},
	}
	for _, tt := range errs {
		t.Run(tt.name, func(t *testing.T) {
			got, err := render(tt.templates, "main")
			if err == nil {
				t.Fatalf("want an error, got %q", got)
			}
			msg := err.Error()
			switch tt.kind {
			case "runtime":
				if !strings.Contains(msg, `required block "`) || !strings.Contains(msg, `" not found`) {
					t.Errorf("error does not say the required block is missing: %v", err)
				}
			case "syntax":
				if !strings.Contains(msg, "parser error") {
					t.Errorf("want a parse error: %v", err)
				}
			}
		})
	}

	// Each chain decides for itself: a child that overrides the block does
	// not make the parent renderable on its own
	t.Run("override_does_not_leak_to_parent", func(t *testing.T) {
		sl := loader.NewStringLoader(loader.NewDirectTemplateParser())
		sl.AddTemplate("base", `[{% block x required %}{% endblock %}]`)
		sl.AddTemplate("c", `{% extends 'base' %}{% block x %}X{% endblock %}`)
		env := miya.NewEnvironment(miya.WithLoader(sl))
		for i := 0; i < 2; i++ {
			c, err := env.GetTemplate("c")
			if err != nil {
				t.Fatal(err)
			}
			if got, err := c.Render(miya.NewContext()); err != nil || got != "[X]" {
				t.Fatalf("child: got %q, %v", got, err)
			}
			base, err := env.GetTemplate("base")
			if err != nil {
				t.Fatal(err)
			}
			if got, err := base.Render(miya.NewContext()); err == nil {
				t.Fatalf("base rendered %q, want the required block error", got)
			}
		}
	})
}
