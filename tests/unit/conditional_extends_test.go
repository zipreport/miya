package miya_test

import (
	"strings"
	"testing"

	miya "github.com/zipreport/miya"
	"github.com/zipreport/miya/loader"
)

// Issue #28: an {% extends %} inside {% if %} is run when its branch is taken,
// a second {% extends %} is an error, and import follows the {% extends %} the
// module runs. Expected values are from Jinja2 3.1.6.
func TestIssue28ConditionalExtends(t *testing.T) {
	newEnv := func(templates map[string]string) *miya.Environment {
		sl := loader.NewStringLoader(loader.NewDirectTemplateParser())
		for n, s := range templates {
			sl.AddTemplate(n, s)
		}
		return miya.NewEnvironment(miya.WithLoader(sl))
	}
	render := func(env *miya.Environment, name string, vars map[string]interface{}) (string, error) {
		tmpl, err := env.GetTemplate("main")
		if err != nil {
			return "", err
		}
		ctx := miya.NewContext()
		ctx.Set("name", name)
		for k, v := range vars {
			ctx.Set(k, v)
		}
		return tmpl.Render(ctx)
	}
	const a = `A[{% block x %}{% endblock %}]`
	const b = `B[{% block x %}{% endblock %}]`

	renders := []struct {
		name, main string
		more       map[string]string
		want       string
	}{
		{"claim1_if_else_extends", `{% if name == 'p' %}{% extends 'a' %}{% else %}{% extends 'b' %}{% endif %}`, nil, "A[]"},
		{"claim1_if_true_extends", `{% if true %}{% extends 'b' %}{% endif %}M`, nil, "B[]"},
		{"first_true_branch_wins", `{% if true %}{% extends 'b' %}{% elif true %}{% extends 'nope' %}{% endif %}`, nil, "B[]"},
		{"import_parent_condition_without_context", `{% import 'i' as m %}ok`,
			map[string]string{"i": `{% extends 'j' %}`, "j": `{% if name == 'p' %}{% extends 'nope' %}{% endif %}`}, "ok"},
		{"else_branch_taken", `{% if name == 'q' %}{% extends 'a' %}{% else %}{% extends 'b' %}{% endif %}{% block x %}X{% endblock %}`, nil, "B[X]"},
		{"no_branch_extends", `{% if name == 'q' %}{% extends 'b' %}{% endif %}M{% block x %}X{% endblock %}`, nil, "MX"},
		{"output_before_if_kept", `pre{% if true %}{% extends 'b' %}{% endif %}M{% block x %}X{% endblock %}`, nil, "preB[X]"},
		{"nested_if_in_elif", `{% if false %}{% elif name == 'p' %}{% if true %}{% extends 'a' %}{% endif %}{% endif %}{% block x %}X{% endblock %}`, nil, "A[X]"},
		{"block_in_untaken_branch_overrides", `{% if true %}{% extends 'b' %}{% else %}{% block x %}E{% endblock %}{% endif %}`, nil, "B[E]"},
		{"code_in_taken_branch_runs", `{% if name == 'p' %}{% extends 'b' %}{% set v = 1 %}{% endif %}{% block x %}{{ v }}{% endblock %}`, nil, "B[1]"},
		{"parent_with_conditional_extends", `{% extends 'm' %}{% block x %}C{% endblock %}`,
			map[string]string{"m": `{% if name == 'p' %}{% extends 'a' %}{% endif %}{% block x %}M{% endblock %}`}, "A[C]"},
		{"import_untaken_missing_parent", `{% import 'i' as m %}ok{{ m.f() }}`,
			map[string]string{"i": `{% if false %}{% extends 'nope' %}{% endif %}{% macro f() %}F{% endmacro %}`}, "okF"},
		{"from_conditional_parent", `{% from 'i' import f %}{{ f() }}`,
			map[string]string{"i": `{% if true %}{% extends 'b' %}{% endif %}{% macro f() %}F{% endmacro %}`}, "F"},
		{"claim4_import_dynamic_parent", `{% import 'i' as m with context %}ok{{ m.f() }}`,
			map[string]string{"i": `{% extends name %}{% macro f() %}F{% endmacro %}`, "p": "P"}, "okF"},
		{"import_parent_untaken_extends", `{% import 'i' as m %}ok`,
			map[string]string{"i": `{% extends 'j' %}`, "j": `{% if false %}{% extends 'nope' %}{% endif %}`}, "ok"},
	}
	for _, tt := range renders {
		t.Run(tt.name, func(t *testing.T) {
			templates := map[string]string{"main": tt.main, "a": a, "b": b}
			for k, v := range tt.more {
				templates[k] = v
			}
			env := newEnv(templates)
			for i := 0; i < 2; i++ {
				got, err := render(env, "p", nil)
				if err != nil {
					t.Fatalf("render %d: %v", i, err)
				}
				if got != tt.want {
					t.Errorf("render %d: got %q, want %q", i, got, tt.want)
				}
			}
		})
	}

	errs := []struct {
		name, main string
		more       map[string]string
		contains   string
	}{
		{"claim2_extended_twice", `{% extends 'b' %}{% extends 'a' %}`, nil, "extended multiple times"},
		{"claim2_extended_twice_conditional", `{% if true %}{% extends 'b' %}{% endif %}{% extends 'a' %}`, nil, "extended multiple times"},
		{"claim2_import_extended_twice", `{% import 'i' as m %}ok`,
			map[string]string{"i": `{% extends 'b' %}{% extends 'a' %}`}, "extended multiple times"},
		{"import_parent_extended_twice", `{% import 'i' as m %}ok`,
			map[string]string{"i": `{% extends 'j' %}`, "j": `{% extends 'a' %}{% extends 'b' %}`}, "extended multiple times"},
		{"claim3_import_non_string_parent", `{% import 'i' as m %}ok`,
			map[string]string{"i": `{% extends 1 %}`}, "must be a string"},
		{"claim4_import_dynamic_missing_parent", `{% import 'i' as m with context %}ok`,
			map[string]string{"i": `{% extends name %}`}, `failed to load template "p"`},
		{"import_conditional_missing_parent", `{% import 'i' as m %}ok`,
			map[string]string{"i": `{% if true %}{% extends 'nope' %}{% endif %}`}, `failed to load template "nope"`},
		{"import_parent_condition_with_context", `{% import 'i' as m with context %}ok`,
			map[string]string{"i": `{% extends 'j' %}`, "j": `{% if name == 'p' %}{% extends 'nope' %}{% endif %}`}, `failed to load template "nope"`},
		{"import_parent_conditional_missing", `{% import 'i' as m %}ok`,
			map[string]string{"i": `{% extends 'j' %}`, "j": `{% if true %}{% extends 'nope' %}{% endif %}`}, `failed to load template "nope"`},
	}
	for _, tt := range errs {
		t.Run(tt.name, func(t *testing.T) {
			templates := map[string]string{"main": tt.main, "a": a, "b": b}
			for k, v := range tt.more {
				templates[k] = v
			}
			got, err := render(newEnv(templates), "p", nil)
			if err == nil {
				t.Fatalf("want an error, got %q", got)
			}
			if !strings.Contains(err.Error(), tt.contains) {
				t.Errorf("error does not say %q: %v", tt.contains, err)
			}
		})
	}

	// The parent is picked per render, so one compiled template follows each
	// render's context
	t.Run("parent_per_render", func(t *testing.T) {
		for _, main := range []string{
			`{% if standalone %}{% else %}{% extends 'b' %}{% endif %}M{% block x %}X{% endblock %}`,
			`{% extends 'm' %}{% block x %}C{% endblock %}`,
		} {
			env := newEnv(map[string]string{"main": main, "a": a, "b": b,
				"m": `{% if standalone %}{% else %}{% extends 'b' %}{% endif %}M{% block x %}X{% endblock %}`})
			want := map[bool]string{false: "B[X]", true: "MX"}
			if strings.Contains(main, "'m'") {
				want = map[bool]string{false: "B[C]", true: "MC"}
			}
			for _, standalone := range []bool{false, true, false, true} {
				got, err := render(env, "p", map[string]interface{}{"standalone": standalone})
				if err != nil {
					t.Fatalf("%s standalone=%v: %v", main, standalone, err)
				}
				if got != want[standalone] {
					t.Errorf("%s standalone=%v: got %q, want %q", main, standalone, got, want[standalone])
				}
			}
		}
	})

	// On import, a condition around an {% extends %} is evaluated once
	t.Run("import_condition_evaluated_once", func(t *testing.T) {
		env := newEnv(map[string]string{"main": `{% import 'i' as m with context %}{{ m.g() }}`, "b": b,
			"i": `{% if f() %}{% extends 'b' %}{% endif %}{% macro g() %}G{% endmacro %}`})
		calls := 0
		f := func() bool { calls++; return false }
		got, err := render(env, "p", map[string]interface{}{"f": f})
		if err != nil || got != "G" {
			t.Fatalf("got %q, %v; want \"G\"", got, err)
		}
		if calls != 1 {
			t.Errorf("f called %d times, want 1", calls)
		}
	})

	// A FromString template is never looked up in the hierarchy cache
	t.Run("from_string_skips_hierarchy_cache", func(t *testing.T) {
		env := newEnv(map[string]string{"b": b})
		for i := 0; i < 3; i++ {
			got, err := env.RenderString(`{% extends 'b' %}{% block x %}X{% endblock %}`, miya.NewContext())
			if err != nil || got != "B[X]" {
				t.Fatalf("got %q, %v", got, err)
			}
		}
		if st := env.GetInheritanceCacheStats().HierarchyCache; st.Misses != 0 || st.Hits != 0 {
			t.Errorf("hierarchy cache lookups: %+v", st)
		}
	})
}
