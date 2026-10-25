package miya_test

import (
	"errors"
	"strings"
	"testing"

	miya "github.com/zipreport/miya"
	"github.com/zipreport/miya/loader"
)

func TestRenderNoPanic(t *testing.T) {
	newEnv := func(opts ...miya.EnvironmentOption) *miya.Environment {
		sl := loader.NewStringLoader(loader.NewDirectTemplateParser())
		sl.AddTemplate("child.html", `{% extends "p.html" %}{% block b %}c{% endblock %}`)
		sl.AddTemplate("p.html", `{% extends 1 %}{% block b %}{% endblock %}`)
		opts = append([]miya.EnvironmentOption{miya.WithLoader(sl), miya.WithAutoEscape(false)}, opts...)
		return miya.NewEnvironment(opts...)
	}

	boom := func(v interface{}, a ...interface{}) (interface{}, error) { panic("boom") }

	var nilVariadic func(...interface{}) (interface{}, error)
	var nilReflect func() string
	var nilIter func() []interface{}

	cases := []struct {
		name     string
		tpl      string
		ctx      map[string]interface{}
		opts     []miya.EnvironmentOption
		render   string // render a named template instead of tpl
		wantOut  *string
		wantErr  string
		notPanic bool
	}{
		{name: "nil_plus_int", tpl: `{{ n + 1 }}`, ctx: map[string]interface{}{"n": nil}, wantErr: "TypeError", notPanic: true},
		{name: "nil_attr_plus_int", tpl: `{{ d.k + 1 }}`, ctx: map[string]interface{}{"d": map[string]interface{}{"k": nil}}, wantErr: "TypeError", notPanic: true},
		{name: "mixed_slices", tpl: `{{ (a + b)|join(',') }}`, ctx: map[string]interface{}{"a": []interface{}{1}, "b": []string{"a"}}, wantOut: sp("1,a")},
		{name: "nil_variadic_func", tpl: `{{ f() }}`, ctx: map[string]interface{}{"f": nilVariadic}, wantErr: "cannot call nil function", notPanic: true},
		{name: "nil_reflect_func", tpl: `{{ f() }}`, ctx: map[string]interface{}{"f": nilReflect}, wantErr: "cannot call nil function", notPanic: true},
		{name: "for_nil_func", tpl: `{% for x in f %}{{ x }}{% endfor %}`, ctx: map[string]interface{}{"f": nilIter}, wantErr: "nil function", notPanic: true},
		{name: "for_nil_func_unpack", tpl: `{% for a, b in f %}{{ a }}{% endfor %}`, ctx: map[string]interface{}{"f": nilIter}, wantErr: "nil function", notPanic: true},
		{name: "for_func_ok", tpl: `{% for x in f %}{{ x }}{% endfor %}`, ctx: map[string]interface{}{"f": func() []interface{} { return []interface{}{1, 2} }}, wantOut: sp("12")},
		{name: "list_concat_too_large", tpl: `{{ (a + b)|length }}`, ctx: map[string]interface{}{"a": make([]bool, 6000000), "b": make([]bool, 6000000)}, wantErr: "too large", notPanic: true},
		{name: "list_concat_ok", tpl: `{{ (a + b)|length }}`, ctx: map[string]interface{}{"a": make([]bool, 3), "b": make([]bool, 3)}, wantOut: sp("6")},
		{name: "trim_delete_last", tpl: "<p>{{ t }}", ctx: map[string]interface{}{"t": "a %}\n\nb\n{% c"}, opts: []miya.EnvironmentOption{miya.WithTrimBlocks(true), miya.WithLstripBlocks(true)}},
		{name: "extends_int", tpl: `{% extends 1 %}`, wantErr: "must be a string", notPanic: true},
		{name: "extends_false", tpl: `{% extends false %}`, wantErr: "must be a string", notPanic: true},
		{name: "extends_none", tpl: `{% extends none %}`, wantErr: "must be a string", notPanic: true},
		{name: "extends_parent_int", render: "child.html", wantErr: "string literal", notPanic: true},
		{name: "range_too_large", tpl: `{{ range(n)|length }}`, ctx: map[string]interface{}{"n": 2000000}, wantErr: "range() too large"},
		{name: "range_at_cap", tpl: `{{ range(n)|length }}`, ctx: map[string]interface{}{"n": 1000000}, wantOut: sp("1000000")},
		{name: "lipsum_too_large", tpl: `{{ lipsum(1001) }}`, wantErr: "lipsum() too large"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			env := newEnv(tc.opts...)
			ctx := miya.NewContext()
			for k, v := range tc.ctx {
				ctx.Set(k, v)
			}
			var out string
			var err error
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Fatalf("panicked: %v", r)
					}
				}()
				if tc.render != "" {
					out, err = env.RenderTemplate(tc.render, ctx)
				} else {
					out, err = env.RenderString(tc.tpl, ctx)
				}
			}()
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("want error containing %q, got err=%v", tc.wantErr, err)
				}
				if tc.notPanic && strings.Contains(err.Error(), "recovered panic") {
					t.Fatalf("error came from the panic backstop: %v", err)
				}
				if strings.Contains(err.Error(), "<<DYNAMIC>>") {
					t.Fatalf("error mentions <<DYNAMIC>>: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantOut != nil && out != *tc.wantOut {
				t.Fatalf("got %q, want %q", out, *tc.wantOut)
			}
		})
	}

	t.Run("call_macro_panic", func(t *testing.T) {
		env := newEnv()
		if err := env.AddFilter("boom", boom); err != nil {
			t.Fatal(err)
		}
		if _, err := env.FromString("{% macro m() %}{{ 1|boom }}{% endmacro %}"); err != nil {
			t.Fatal(err)
		}
		var err error
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("panicked: %v", r)
				}
			}()
			_, err = env.CallMacro("m", miya.NewContext(), nil, nil)
		}()
		if err == nil || !errors.Is(err, miya.ErrRenderPanic) {
			t.Fatalf("want ErrRenderPanic, got %v", err)
		}
	})

	t.Run("panicking_filter", func(t *testing.T) {
		env := newEnv()
		if err := env.AddFilter("boom", boom); err != nil {
			t.Fatal(err)
		}
		var err error
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("panicked: %v", r)
				}
			}()
			_, err = env.RenderString(`{{ 1|boom }}`, miya.NewContext())
		}()
		if err == nil || !strings.Contains(err.Error(), "recovered panic") || !strings.Contains(err.Error(), "boom") {
			t.Fatalf("want recovered panic error mentioning boom, got %v", err)
		}
		if !errors.Is(err, miya.ErrRenderPanic) {
			t.Fatalf("errors.Is(err, ErrRenderPanic) false: %v", err)
		}

		out, err := env.RenderString(`{{ 1 + 1 }}`, miya.NewContext())
		if err != nil || out != "2" {
			t.Fatalf("pool_survives_panic: got %q, %v", out, err)
		}
	})

	t.Run("render_to_panic", func(t *testing.T) {
		env := newEnv()
		if err := env.AddFilter("boom", boom); err != nil {
			t.Fatal(err)
		}
		tmpl, err := env.FromString(`{{ 1|boom }}`)
		if err != nil {
			t.Fatal(err)
		}
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("panicked: %v", r)
				}
			}()
			err = tmpl.RenderTo(&strings.Builder{}, miya.NewContext())
		}()
		if err == nil || !errors.Is(err, miya.ErrRenderPanic) {
			t.Fatalf("want ErrRenderPanic, got %v", err)
		}
	})
}

func sp(s string) *string { return &s }
