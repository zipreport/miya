package miya_test

import (
	"strings"
	"testing"

	miya "github.com/zipreport/miya"
	"github.com/zipreport/miya/loader"
)

func TestLogicalOperators(t *testing.T) {
	calls := 0
	data := func() map[string]interface{} {
		calls = 0
		return map[string]interface{}{
			"name":  "Ann",
			"empty": "",
			"n":     nil,
			"z":     0,
			"items": nil,
			"f":     func() (interface{}, error) { calls++; return "called", nil },
		}
	}

	cases := []struct {
		name, tmpl, want string
		wantCalls        int
	}{
		{"or_default", `{{ name or 'default' }}`, "Ann", 0},
		{"or_empty", `{{ empty or 'x' }}`, "x", 0},
		{"precedence", `{{ 0 or 1 and 2 }}`, "2", 0},
		{"and_truthy", `{{ name and 'y' }}`, "y", 0},
		{"and_falsy_empty", `{{ empty and 'y' }}`, "", 0},
		{"and_returns_zero", `{{ 1 and 0 }}`, "0", 0},
		{"or_none_arith", `{{ (n or 5) + 1 }}`, "6", 0},
		{"or_empty_list", `{{ [] or 'e' }}`, "e", 0},
		{"or_none_length", `{{ (items or [])|length }}`, "0", 0},
		{"and_guards_division", `{% if z != 0 and 10 / z > 1 %}a{% else %}b{% endif %}`, "b", 0},
		{"and_skips_call", `{{ false and f() }}`, "false", 0},
		{"or_skips_call", `{{ true or f() }}`, "true", 0},
		{"or_undefined", `{{ nope or 'd' }}`, "d", 0},
		{"and_undefined", `{{ nope and 'd' }}|`, "|", 0},
		{"set_or", `{% set x = name or 'd' %}{{ x }}`, "Ann", 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			env := miya.NewEnvironment()
			got, err := env.RenderString(tc.tmpl, miya.NewContextFrom(data()))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
			if calls != tc.wantCalls {
				t.Errorf("f called %d times, want %d", calls, tc.wantCalls)
			}
		})
	}

	strict := []struct {
		name, tmpl, want string
		wantErr          bool
	}{
		{"strict_defined_guard", `{{ a is defined and a.b }}|`, "false|", false},
		{"strict_or_left_truthy", `{{ name or nope }}`, "Ann", false},
		{"strict_or_left_falsy", `{{ empty or nope }}`, "", true},
	}
	for _, tc := range strict {
		t.Run(tc.name, func(t *testing.T) {
			env := miya.NewEnvironment(miya.WithStrictUndefined(true))
			got, err := env.RenderString(tc.tmpl, miya.NewContextFrom(data()))
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

// Each subtest pins current miya behaviour that differs from Jinja2; the plan
// named in the comment must update the expectation when it fixes the cause.
func TestLogicalOperatorsKnownDivergences(t *testing.T) {
	// owner: plan 047 (typed nil becomes none and falsy); Jinja2 would give R.
	t.Run("typed_nil_or", func(t *testing.T) {
		env := miya.NewEnvironment(miya.WithAutoEscape(false))
		got, err := env.RenderString(`{{ tnil or 'R' }}`, miya.NewContextFrom(map[string]interface{}{"tnil": (*struct{ X int })(nil)}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "<nil>" {
			t.Errorf("got %q, want %q", got, "<nil>")
		}
	})

	// owner: plan 027 (Markup); Jinja2 gives R.
	t.Run("safe_empty_or", func(t *testing.T) {
		env := miya.NewEnvironment()
		got, err := env.RenderString(`{{ (''|safe) or 'R' }}`, miya.NewContext())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "" {
			t.Errorf("got %q, want empty", got)
		}
	})

	newStrictEnv := func(child string) *miya.Environment {
		sl := loader.NewStringLoader(loader.NewDirectTemplateParser())
		sl.AddTemplate("base.html", `B[{% block c %}{% endblock %}]`)
		sl.AddTemplate("child.html", child)
		return miya.NewEnvironment(miya.WithLoader(sl), miya.WithStrictUndefined(true))
	}

	// Matches Jinja2 since plan 034: a dynamic extends follows strict mode.
	t.Run("strict_dynamic_extends", func(t *testing.T) {
		env := newStrictEnv(`{% extends nope or 'base.html' %}{% block c %}C{% endblock %}`)
		tpl, err := env.GetTemplate("child.html")
		if err != nil {
			t.Fatalf("GetTemplate: %v", err)
		}
		got, err := tpl.Render(miya.NewContext())
		if err == nil || !strings.Contains(err.Error(), "UndefinedError") || !strings.Contains(err.Error(), "nope") {
			t.Errorf("got %q, %v; want an UndefinedError naming nope", got, err)
		}
	})

	t.Run("strict_dynamic_include", func(t *testing.T) {
		env := newStrictEnv(`{% include nope or 'base.html' %}`)
		tpl, err := env.GetTemplate("child.html")
		if err == nil {
			_, err = tpl.Render(miya.NewContext())
		}
		if err == nil {
			t.Fatal("expected error from strict undefined in include expression")
		}
	})
}
