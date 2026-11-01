package miya_test

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"

	miya "github.com/zipreport/miya"
	"github.com/zipreport/miya/loader"
	"github.com/zipreport/miya/runtime"
)

func newImportEnv(templates map[string]string, autoescape bool) *miya.Environment {
	sl := loader.NewStringLoader(loader.NewDirectTemplateParser())
	for name, src := range templates {
		sl.AddTemplate(name, src)
	}
	env := miya.NewEnvironment(miya.WithLoader(sl), miya.WithAutoEscape(autoescape))
	env.AddGlobal("G", "glob")
	return env
}

func renderImport(env *miya.Environment, ctx map[string]interface{}) (string, error) {
	tmpl, err := env.GetTemplate("i.html")
	if err != nil {
		return "", err
	}
	return tmpl.Render(miya.NewContextFrom(ctx))
}

func TestImportIsolation(t *testing.T) {
	const sV = `{% macro s() %}{{ v }}{% endmacro %}`
	const sX = `{% macro s() %}x:{{ v }}{% endmacro %}`
	const base = `[{% block b %}{% endblock %}]`
	cases := []struct {
		name      string
		templates map[string]string
		ctx       map[string]interface{}
		ae        bool
		want      string
	}{
		{"loop_import", map[string]string{"m.html": sV, "i.html": `{% for v in ['A','B','C'] %}{% import "m.html" as m %}{{ v }}:{{ m.s() }}|{% endfor %}`}, nil, false, `A:|B:|C:|`},
		{"set_two_alias", map[string]string{"m.html": sV, "i.html": `{% set v = 'A' %}{% import "m.html" as m %}A:{{ m.s() }}|{% set v = 'B' %}{% import "m.html" as n %}B:{{ n.s() }}`}, nil, false, `A:|B:`},
		{"loop_from", map[string]string{"m.html": sV, "i.html": `{% for v in ['A','B','C'] %}{% from "m.html" import s %}{{ v }}:{{ s() }}|{% endfor %}`}, nil, false, `A:|B:|C:|`},
		{"loop_include", map[string]string{"m.html": sV, "inc.html": `{% import "m.html" as m %}{{ v }}:{{ m.s() }}|`, "i.html": `{% for v in ['A','B','C'] %}{% include "inc.html" %}{% endfor %}`}, nil, false, `A:|B:|C:|`},
		{"import_in_macro_body", map[string]string{"m.html": sV, "i.html": `{% macro w(v) %}{% import "m.html" as m %}{{ m.s() }}{% endmacro %}{{ w('A') }}|{{ w('B') }}`}, nil, false, `|`},
		{"import_sees_render", map[string]string{"m.html": sX, "i.html": `{% import "m.html" as m %}{{ m.s() }}`}, map[string]interface{}{"v": "SECRET"}, false, `x:`},
		{"from_sees_render", map[string]string{"m.html": sX, "i.html": `{% from "m.html" import s %}{{ s() }}`}, map[string]interface{}{"v": "SECRET"}, false, `x:`},
		{"import_with_context", map[string]string{"m.html": sX, "i.html": `{% import "m.html" as m with context %}{{ m.s() }}`}, map[string]interface{}{"v": "SECRET"}, false, `x:SECRET`},
		{"from_with_context", map[string]string{"m.html": sX, "i.html": `{% from "m.html" import s with context %}{{ s() }}`}, map[string]interface{}{"v": "SECRET"}, false, `x:SECRET`},
		{"import_without_context", map[string]string{"m.html": sX, "i.html": `{% import "m.html" as m without context %}{{ m.s() }}`}, map[string]interface{}{"v": "SECRET"}, false, `x:`},
		{"with_ctx_loop", map[string]string{"m.html": sV, "i.html": `{% for v in ['A','B'] %}{% import "m.html" as m with context %}{{ v }}:{{ m.s() }}|{% endfor %}`}, nil, false, `A:A|B:B|`},
		{"with_ctx_snapshot", map[string]string{"m.html": sV, "i.html": `{% set v = 'A' %}{% import "m.html" as m with context %}{{ m.s() }}|{% set v = 'B' %}{{ m.s() }}`}, nil, false, `A|A`},
		{"with_ctx_toplevel", map[string]string{"m.html": `{% set t = v ~ '!' %}{% macro s() %}{{ t }}{% endmacro %}`, "i.html": `{% import "m.html" as m with context %}{{ m.s() }}|{{ m.t }}`}, map[string]interface{}{"v": "R"}, false, `R!|R!`},
		{"toplevel_set", map[string]string{"m.html": `{% set g = 'G' %}{% macro s() %}g={{ g }}{% endmacro %}`, "i.html": `{% import "m.html" as m %}{{ m.s() }}|{{ m.g }}`}, nil, false, `g=G|G`},
		{"globals_and_filters", map[string]string{"m.html": `{% macro s() %}{{ G }}{{ 'a'|upper }}{{ range(2)|list|length }}{% endmacro %}`, "i.html": `{% import "m.html" as m %}{{ m.s() }}`}, nil, false, `globA2`},
		{"autoescape_block_import", map[string]string{"m.html": `{% macro s() %}{{ v }}{{ G }}{{ 'a'|upper }}{% endmacro %}`, "i.html": `{% autoescape false %}{% import "m.html" as m %}[{{ m.s() }}]{% endautoescape %}`}, map[string]interface{}{"v": "X"}, true, `[globA]`},
		{"extends_from_alias", map[string]string{"m.html": `{% macro s() %}S{% endmacro %}`, "b.html": base, "i.html": `{% extends "b.html" %}{% from "m.html" import s as t %}{% block b %}{{ t() }}{% endblock %}`}, nil, false, `[S]`},
		{"extends_import_with_context", map[string]string{"m.html": sV, "b.html": base, "i.html": `{% extends "b.html" %}{% import "m.html" as m with context %}{% block b %}{{ m.s() }}{% endblock %}`}, map[string]interface{}{"v": "R"}, false, `[R]`},
		{"from_missing_name_unused", map[string]string{"m.html": ``, "i.html": `{% from "m.html" import nothere %}ok`}, nil, false, `ok`},
		{"from_missing_name_printed", map[string]string{"m.html": ``, "i.html": `{% from "m.html" import nothere %}[{{ nothere }}]`}, nil, false, `[]`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := renderImport(newImportEnv(c.templates, c.ae), c.ctx)
			if err != nil {
				t.Fatalf("render error: %v", err)
			}
			if got != c.want {
				t.Fatalf("got %q, want %q", got, c.want)
			}
		})
	}
}

func TestImportCrossRender(t *testing.T) {
	env := newImportEnv(map[string]string{
		"m.html": `{% macro s() %}{{ v }}{% endmacro %}`,
		"i.html": `{% import "m.html" as m %}<{{ v }}:{{ m.s() }}>`,
	}, false)
	for _, v := range []string{"A", "B", "C"} {
		got, err := renderImport(env, map[string]interface{}{"v": v})
		want := "<" + v + ":>"
		if err != nil || got != want {
			t.Fatalf("v=%s: got %q, %v; want %q", v, got, err, want)
		}
	}
}

func TestImportErrors(t *testing.T) {
	t.Run("missing_import", func(t *testing.T) {
		got, err := renderImport(newImportEnv(map[string]string{"i.html": `{% import "nope.html" as m %}ok`}, false), nil)
		if err == nil || got != "" {
			t.Fatalf("got %q, %v; want an error and empty output", got, err)
		}
	})
	t.Run("missing_from", func(t *testing.T) {
		got, err := renderImport(newImportEnv(map[string]string{"i.html": `{% from "nope.html" import s %}ok`}, false), nil)
		if err == nil || got != "" {
			t.Fatalf("got %q, %v; want an error and empty output", got, err)
		}
	})
	t.Run("missing_name_called", func(t *testing.T) {
		env := newImportEnv(map[string]string{"m.html": `{% macro s() %}S{% endmacro %}`, "i.html": `{% from "m.html" import nothere %}[{{ nothere() }}]`}, false)
		_, err := renderImport(env, nil)
		if err == nil || !strings.Contains(err.Error(), `does not export the requested name "nothere"`) {
			t.Fatalf("got %v; want a does-not-export error", err)
		}
	})
	t.Run("bad_context_keyword", func(t *testing.T) {
		env := newImportEnv(map[string]string{"m.html": ``, "i.html": `{% import "m.html" as m with foo %}`}, false)
		_, err := env.GetTemplate("i.html")
		if err == nil || !strings.Contains(err.Error(), "expected 'context'") {
			t.Fatalf("got %v; want an expected 'context' parse error", err)
		}
	})
}

func TestImportConcurrentRenders(t *testing.T) {
	env := newImportEnv(map[string]string{
		"m.html": `{% macro s(x) %}{{ x }}{{ v }}{% endmacro %}`,
		"i.html": `{% import "m.html" as m %}{% from "m.html" import s %}{{ v }}:{{ m.s(v) }}:{{ s(v) }}`,
	}, false)
	const goroutines, iterations = 8, 200
	start := make(chan struct{})
	var wg sync.WaitGroup
	errs := make(chan error, goroutines)
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			<-start
			for i := 0; i < iterations; i++ {
				v := fmt.Sprintf("g%d-%d", g, i)
				got, err := renderImport(env, map[string]interface{}{"v": v})
				if err != nil {
					errs <- err
					return
				}
				if want := v + ":" + v + ":" + v; got != want {
					errs <- fmt.Errorf("got %q, want %q", got, want)
					return
				}
			}
		}(g)
	}
	close(start)
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatalf("concurrent render: %v", err)
	}
}

func newStrictImportEnv(templates map[string]string) *miya.Environment {
	sl := loader.NewStringLoader(loader.NewDirectTemplateParser())
	for name, src := range templates {
		sl.AddTemplate(name, src)
	}
	return miya.NewEnvironment(miya.WithLoader(sl), miya.WithStrictUndefined(true))
}

func TestImportModules(t *testing.T) {
	const h = `{% macro x() %}H{% endmacro %}`
	const counter = `{% set ns = namespace(c=0) %}{% macro s() %}{% set ns.c = ns.c + 1 %}{{ ns.c }}{% endmacro %}`
	cases := []struct {
		name      string
		templates map[string]string
		want      string
	}{
		{"nested_import", map[string]string{"h.html": h, "m.html": `{% import 'h.html' as h %}{% macro s() %}[{{ h.x() }}]{% endmacro %}`, "i.html": `{% import 'm.html' as m %}{{ m.s() }}`}, `[H]`},
		{"nested_from", map[string]string{"h.html": h, "m.html": `{% from 'h.html' import x %}{% macro s() %}[{{ x() }}]{% endmacro %}`, "i.html": `{% import 'm.html' as m %}{{ m.s() }}`}, `[H]`},
		{"nested_not_exported", map[string]string{"h.html": h, "n.html": `{% import 'h.html' as h %}{% from 'h.html' import x %}{% macro s() %}{{ h.x() }}{{ x() }}{% endmacro %}`, "i.html": `{% import 'n.html' as n %}[{{ n.h is defined }}|{{ n.x is defined }}|{{ n.s() }}]`}, `[false|false|HH]`},
		{"block_set", map[string]string{"m.html": `{% set x %}X{% endset %}{% macro s() %}[{{ x }}]{% endmacro %}`, "i.html": `{% import 'm.html' as m %}{{ m.s() }}|{{ m.x }}`}, `[X]|X`},
		{"from_missing_defined", map[string]string{"m.html": ``, "i.html": `{% from "m.html" import nothere %}{{ nothere is defined }}`}, `false`},
		{"defined_on_undefined_variable", map[string]string{"i.html": `{% set x = nope %}{{ x is defined }}|{{ x is undefined }}|{% set n = none %}{{ n is defined }}`}, `false|true|true`},
		{"from_trailing_comma", map[string]string{"m.html": `{% macro s() %}S{{ v }}{% endmacro %}`, "i.html": `{% set v = 'V' %}{% from "m.html" import s, with context %}{{ s() }}`}, `SV`},
		{"module_shared_in_loop", map[string]string{"m.html": counter, "i.html": `{% for i in [1,2,3] %}{% import 'm.html' as m %}{{ m.s() }}{% endfor %}`}, `123`},
		{"module_shared_with_include", map[string]string{"m.html": counter, "k.html": `{% import 'm.html' as m %}{{ m.s() }}`, "i.html": `{% import 'm.html' as m %}{{ m.s() }}{% include 'k.html' %}`}, `12`},
		{"module_shared_import_and_from", map[string]string{"m.html": counter, "i.html": `{% import 'm.html' as m %}{% from 'm.html' import s %}{{ m.s() }}{{ s() }}`}, `12`},
		{"with_context_after_plain", map[string]string{"m.html": `{% macro s() %}{{ v }}{% endmacro %}`, "i.html": `{% import 'm.html' as a %}{% set v = 'B' %}{% import 'm.html' as b with context %}[{{ b.s() }}]`}, `[B]`},
		{"plain_after_with_context", map[string]string{"m.html": `{% macro s() %}{{ v }}{% endmacro %}`, "i.html": `{% set v = 'B' %}{% import 'm.html' as b with context %}{% import 'm.html' as a %}[{{ a.s() }}]`}, `[]`},
		{"with_context_not_shared", map[string]string{"m.html": counter, "i.html": `{% for i in [1,2] %}{% import 'm.html' as m with context %}{{ m.s() }}{% endfor %}`}, `11`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := renderImport(newImportEnv(c.templates, false), nil)
			if err != nil {
				t.Fatalf("render error: %v", err)
			}
			if got != c.want {
				t.Fatalf("got %q, want %q", got, c.want)
			}
		})
	}
}

func TestImportModuleFreshPerRender(t *testing.T) {
	env := newImportEnv(map[string]string{
		"m.html": `{% set ns = namespace(c=0) %}{% macro s() %}{% set ns.c = ns.c + 1 %}{{ ns.c }}{% endmacro %}`,
		"i.html": `{% for i in [1,2] %}{% import 'm.html' as m %}{{ m.s() }}{% endfor %}`,
	}, false)
	for i := 0; i < 3; i++ {
		got, err := renderImport(env, nil)
		if err != nil || got != "12" {
			t.Fatalf("render %d: got %q, %v; want \"12\"", i, got, err)
		}
	}
}

func TestImportStrict(t *testing.T) {
	t.Run("toplevel_set_error", func(t *testing.T) {
		env := newStrictImportEnv(map[string]string{"m.html": `{% set z = v %}`, "i.html": `{% import 'm.html' as m %}{{ m.z }}`})
		got, err := renderImport(env, map[string]interface{}{"v": "V"})
		if err == nil || !strings.Contains(err.Error(), "{% set z %}") || !strings.Contains(err.Error(), "undefined variable: v") {
			t.Fatalf("got %q, %v; want an undefined-variable error for v from {%% set z %%}", got, err)
		}
	})
	t.Run("from_missing_printed", func(t *testing.T) {
		env := newStrictImportEnv(map[string]string{"m.html": ``, "i.html": `{% from "m.html" import nothere %}[{{ nothere }}]`})
		got, err := renderImport(env, nil)
		if err == nil || !strings.Contains(err.Error(), `does not export the requested name "nothere"`) {
			t.Fatalf("got %q, %v; want a does-not-export error", got, err)
		}
	})
	t.Run("from_missing_defined", func(t *testing.T) {
		env := newStrictImportEnv(map[string]string{"m.html": ``, "i.html": `{% from "m.html" import nothere %}{{ nothere is defined }}`})
		got, err := renderImport(env, nil)
		if err != nil || got != "false" {
			t.Fatalf("got %q, %v; want \"false\"", got, err)
		}
	})
}

func TestImportRecursion(t *testing.T) {
	cases := map[string]map[string]string{
		"self":   {"m.html": `{% import 'm.html' as me %}{% macro s() %}S{% endmacro %}`, "i.html": `{% import 'm.html' as m %}{{ m.s() }}`},
		"mutual": {"a.html": `{% import 'b.html' as b %}{% macro s() %}A{% endmacro %}`, "b.html": `{% import 'a.html' as a %}{% macro t() %}B{% endmacro %}`, "i.html": `{% import 'a.html' as m %}{{ m.s() }}`},
	}
	for name, templates := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := renderImport(newImportEnv(templates, false), nil)
			if !errors.Is(err, runtime.ErrMaxRecursion) {
				t.Fatalf("got %v, want ErrMaxRecursion", err)
			}
		})
	}
}

func benchImport(b *testing.B, main string) {
	env := newImportEnv(map[string]string{
		"m.html": `{% set t = 'T' %}{% macro s(x) %}{{ x }}{{ t }}{% endmacro %}`,
		"i.html": main,
	}, false)
	rows := make([]interface{}, 10000)
	for i := range rows {
		rows[i] = i
	}
	tmpl, err := env.GetTemplate("i.html")
	if err != nil {
		b.Fatal(err)
	}
	ctx := map[string]interface{}{"rows": rows}
	// Guard the budget against a benchmark that renders nothing.
	if out, err := tmpl.Render(miya.NewContextFrom(ctx)); err != nil || !strings.HasSuffix(out, "9999T") {
		b.Fatalf("render: %v, output ends %q", err, out[max(0, len(out)-10):])
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := tmpl.Render(miya.NewContextFrom(ctx)); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkImportTopLevel(b *testing.B) {
	benchImport(b, `{% import 'm.html' as m %}{% for r in rows %}{{ m.s(r) }}{% endfor %}`)
}

func BenchmarkImportInLoop(b *testing.B) {
	benchImport(b, `{% for r in rows %}{% import 'm.html' as m %}{{ m.s(r) }}{% endfor %}`)
}
