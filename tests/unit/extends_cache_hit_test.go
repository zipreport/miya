package miya_test

import (
	"strings"
	"testing"

	miya "github.com/zipreport/miya"
	"github.com/zipreport/miya/loader"
)

const (
	cacheHitBase   = `<{% block a %}A{{ v }}{% endblock %}|{% for i in [1,2] %}{% block b scoped %}b{{ i }}{% endblock %}{% endfor %}>`
	cacheHitMid    = `{% extends 'base' %}{% set midv = 'M' %}{% block a %}m[{{ super() }}]{% endblock %}`
	cacheHitLeaf   = `{% extends 'mid' %}{% macro w(x) %}({{ x }}){% endmacro %}{% block a %}{{ w(v) }}{{ super() }}{% endblock %}{% block b %}B{{ i }}{{ super() }}{% endblock %}`
	cacheHitSupOut = `{% extends 'base' %}{{ super() }}`
)

func cacheHitEnv(ae bool) *miya.Environment {
	sl := loader.NewStringLoader(loader.NewDirectTemplateParser())
	sl.AddTemplate("base", cacheHitBase)
	sl.AddTemplate("mid", cacheHitMid)
	sl.AddTemplate("leaf", cacheHitLeaf)
	sl.AddTemplate("sup_out", cacheHitSupOut)
	return miya.NewEnvironment(miya.WithLoader(sl), miya.WithAutoEscape(ae))
}

func cacheHitRender(t *testing.T, env *miya.Environment, name, v string) string {
	t.Helper()
	tmpl, err := env.GetTemplate(name)
	if err != nil {
		t.Fatal(err)
	}
	out, err := tmpl.Render(miya.NewContextFrom(map[string]interface{}{"v": v}))
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func TestExtendsCacheHitMatchesCold(t *testing.T) {
	for _, ae := range []bool{true, false} {
		want := map[string]string{
			"<1>": `<(<1>)m[A<1>]|B1b1B2b2>`,
			"2":   `<(2)m[A2]|B1b1B2b2>`,
		}
		name := "ae_off"
		if ae {
			want["<1>"] = `<(&lt;1&gt;)m[A&lt;1&gt;]|B1b1B2b2>`
			name = "ae_on"
		}
		t.Run(name, func(t *testing.T) {
			env := cacheHitEnv(ae)
			for _, v := range []string{"<1>", "2", "<1>"} {
				if got := cacheHitRender(t, env, "leaf", v); got != want[v] {
					t.Errorf("cached render v=%q: got %q, want %q", v, got, want[v])
				}
			}
			for _, v := range []string{"<1>", "2"} {
				cold := cacheHitRender(t, cacheHitEnv(ae), "leaf", v)
				if cold != want[v] {
					t.Errorf("cold render v=%q: got %q, want %q", v, cold, want[v])
				}
			}
		})
	}
}

func TestExtendsBuildErrorNotCached(t *testing.T) {
	env := cacheHitEnv(true)
	tmpl, err := env.GetTemplate("sup_out")
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		out, err := tmpl.Render(miya.NewContext())
		if err == nil || !strings.Contains(err.Error(), "super()") {
			t.Fatalf("render %d: want super() error, got %v", i+1, err)
		}
		if out != "" {
			t.Errorf("render %d: want empty output, got %q", i+1, out)
		}
	}
}

func TestMacroMissingParamShadowsOuter(t *testing.T) {
	env := miya.NewEnvironment()
	out, err := env.RenderString(`{% set p = 'outer' %}{% macro m(p) %}[{{ p }}]{% endmacro %}{{ m() }}|{{ p }}`, nil)
	if err != nil {
		t.Fatal(err)
	}
	if out != "[]|outer" {
		t.Errorf("got %q, want %q", out, "[]|outer")
	}
}
