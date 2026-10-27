package miya_test

import (
	"errors"
	"runtime/debug"
	"strings"
	"testing"

	miya "github.com/zipreport/miya"
	"github.com/zipreport/miya/loader"
	"github.com/zipreport/miya/runtime"
)

func newRecursionEnv(templates map[string]string) *miya.Environment {
	sl := loader.NewStringLoader(loader.NewDirectTemplateParser())
	for name, src := range templates {
		sl.AddTemplate(name, src)
	}
	return miya.NewEnvironment(miya.WithLoader(sl), miya.WithAutoEscape(false))
}

func TestRecursionLimit(t *testing.T) {
	defer debug.SetMaxStack(debug.SetMaxStack(64 << 20))

	errCases := []struct {
		name      string
		templates map[string]string
		entry     string
	}{
		{"self_include", map[string]string{"index.html": `x{% include "index.html" %}`}, "index.html"},
		{"mutual_include", map[string]string{
			"a.html": `{% include "b.html" %}`,
			"b.html": `{% include "a.html" %}`,
		}, "a.html"},
		{"ignore_missing", map[string]string{"index.html": `x{% include "index.html" ignore missing %}`}, "index.html"},
		{"local_macro", map[string]string{"index.html": `{% macro f(n) %}{{ f(n) }}{% endmacro %}{{ f(1) }}`}, "index.html"},
		{"imported_macro", map[string]string{
			"m.html":     `{% macro apply(f) %}{{ f(f) }}{% endmacro %}`,
			"index.html": `{% import "m.html" as m %}{{ m.apply(m.apply) }}`,
		}, "index.html"},
		{"macro_default", map[string]string{"index.html": `{% macro g(a=g()) %}{% endmacro %}{{ g() }}`}, "index.html"},
		{"recursive_loop", map[string]string{"index.html": `{% for x in [1] recursive %}{{ loop([1]) }}{% endfor %}`}, "index.html"},
		{"with_block", map[string]string{"index.html": `{% macro f(n) %}{{ f(n) }}{% endmacro %}{% with y = f(1) %}{{ y }}{% endwith %}`}, "index.html"},
		{"child_block_include", map[string]string{
			"base.html":  `{% block b %}{% endblock %}`,
			"child.html": `{% extends "base.html" %}{% block b %}{% include "child.html" %}{% endblock %}`,
		}, "child.html"},
	}
	for _, tc := range errCases {
		t.Run(tc.name, func(t *testing.T) {
			env := newRecursionEnv(tc.templates)
			out, err := env.RenderTemplate(tc.entry, miya.NewContext())
			if !errors.Is(err, runtime.ErrMaxRecursion) {
				t.Fatalf("expected ErrMaxRecursion, got err=%v (output len %d)", err, len(out))
			}
			// the limit is reported once, not wrapped at every level
			if n := len(err.Error()); n > 1024 {
				t.Fatalf("error message is %d bytes: %.300s...", n, err.Error())
			}
		})
	}

	t.Run("legit_depth", func(t *testing.T) {
		env := newRecursionEnv(nil)
		out, err := env.RenderString(`{% macro f(n) %}{% if n %}({{ f(n - 1) }}){% endif %}{% endmacro %}{{ f(100) }}`, miya.NewContext())
		if err != nil {
			t.Fatal(err)
		}
		want := strings.Repeat("(", 100) + strings.Repeat(")", 100)
		if out != want {
			t.Fatalf("got %q", out)
		}
	})

	t.Run("legit_tree_loop", func(t *testing.T) {
		env := newRecursionEnv(nil)
		ctx := miya.NewContext()
		ctx.Set("items", []interface{}{
			map[string]interface{}{"name": "a", "children": []interface{}{
				map[string]interface{}{"name": "b", "children": []interface{}{
					map[string]interface{}{"name": "c"},
				}},
				map[string]interface{}{"name": "d"},
			}},
			map[string]interface{}{"name": "e"},
		})
		out, err := env.RenderString(`{% for n in items recursive %}[{{ n.name }}{% if n.children %}{{ loop(n.children) }}{% endif %}]{% endfor %}`, ctx)
		if err != nil {
			t.Fatal(err)
		}
		if want := "[a[b[c]][d]][e]"; out != want {
			t.Fatalf("got %q want %q", out, want)
		}
	})

	t.Run("after_error", func(t *testing.T) {
		env := newRecursionEnv(map[string]string{"index.html": `x{% include "index.html" %}`})
		if _, err := env.RenderTemplate("index.html", miya.NewContext()); !errors.Is(err, runtime.ErrMaxRecursion) {
			t.Fatalf("expected ErrMaxRecursion, got %v", err)
		}
		out, err := env.RenderString(`{{ 1 }}`, miya.NewContext())
		if err != nil || out != "1" {
			t.Fatalf("got %q, %v", out, err)
		}
	})
}
