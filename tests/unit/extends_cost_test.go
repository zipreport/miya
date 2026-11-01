package miya_test

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"testing"

	miya "github.com/zipreport/miya"
	"github.com/zipreport/miya/loader"
)

func extendsTestBody() string {
	var b strings.Builder
	for i := 0; i < 40; i++ {
		fmt.Fprintf(&b, "<p>{{ x }} %d {%% if x %%}y{%% endif %%}</p>\n", i)
	}
	return b.String()
}

func extendsTestEnv(templates map[string]string) *miya.Environment {
	sl := loader.NewStringLoader(loader.NewDirectTemplateParser())
	for k, v := range templates {
		sl.AddTemplate(k, v)
	}
	return miya.NewEnvironment(miya.WithLoader(sl))
}

func renderAllocs(t *testing.T, env *miya.Environment, name string) (uint64, string) {
	t.Helper()
	tmpl, err := env.GetTemplate(name)
	if err != nil {
		t.Fatal(err)
	}
	ctx := miya.NewContextFrom(map[string]interface{}{"x": 1})
	out, err := tmpl.Render(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	for i := 0; i < 20; i++ {
		if _, err := tmpl.Render(ctx); err != nil {
			t.Fatal(err)
		}
	}
	runtime.ReadMemStats(&after)
	return (after.TotalAlloc - before.TotalAlloc) / 20, out
}

// Rendering a template that extends another costs about what rendering the
// same page as one template costs: the merged template is built once.
func TestExtendsRenderCostCloseToFlat(t *testing.T) {
	body := extendsTestBody()
	flat, flatOut := renderAllocs(t, extendsTestEnv(map[string]string{
		"flat.html": "<html>" + body + "<main>M" + body + "</main></html>",
	}), "flat.html")
	ext, extOut := renderAllocs(t, extendsTestEnv(map[string]string{
		"base.html":  "<html>" + body + "<main>{% block main %}M{% endblock %}</main></html>",
		"child.html": "{% extends 'base.html' %}{% set unused = 1 %}{% block main %}{{ super() }}" + body + "{% endblock %}",
	}), "child.html")
	if flatOut != extOut {
		t.Fatalf("outputs differ:\nflat %q\next  %q", flatOut, extOut)
	}
	if ratio := float64(ext) / float64(flat); ratio > 1.3 {
		t.Fatalf("an extending template allocates %.2fx a flat one with the same output (flat %d B, extends %d B); the merged template must be reused", ratio, flat, ext)
	}
}

// Renders racing on an extending template's first use all get the right page.
func TestExtendsConcurrentFirstRender(t *testing.T) {
	env := extendsTestEnv(map[string]string{
		"base.html":  "[{% block a %}A{% endblock %}|{% block b %}B{% endblock %}]",
		"child.html": "{% extends 'base.html' %}{% block b %}<{{ super() }}{{ x }}>{% endblock %}",
	})
	tmpl, err := env.GetTemplate("child.html")
	if err != nil {
		t.Fatal(err)
	}
	start := make(chan struct{})
	var wg sync.WaitGroup
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			<-start
			for i := 0; i < 50; i++ {
				got, err := tmpl.Render(miya.NewContextFrom(map[string]interface{}{"x": g}))
				if want := fmt.Sprintf("[A|<B%d>]", g); err != nil || got != want {
					t.Errorf("got %q, %v; want %q", got, err, want)
					return
				}
			}
		}(g)
	}
	close(start)
	wg.Wait()
}

// Invalidating an extending template drops its merged template with it.
func TestExtendsInvalidateReloads(t *testing.T) {
	sl := loader.NewStringLoader(loader.NewDirectTemplateParser())
	sl.AddTemplate("inv_base.html", "[{% block b %}B{% endblock %}]")
	sl.AddTemplate("inv_child.html", "{% extends 'inv_base.html' %}{% block b %}1{% endblock %}")
	env := miya.NewEnvironment(miya.WithLoader(sl))
	render := func() string {
		t.Helper()
		tmpl, err := env.GetTemplate("inv_child.html")
		if err != nil {
			t.Fatal(err)
		}
		out, err := tmpl.Render(miya.NewContext())
		if err != nil {
			t.Fatal(err)
		}
		return out
	}
	if got := render() + render(); got != "[1][1]" {
		t.Fatalf("got %q, want [1][1]", got)
	}
	sl.AddTemplate("inv_child.html", "{% extends 'inv_base.html' %}{% block b %}2{% endblock %}")
	env.InvalidateTemplate("inv_child.html")
	if got := render(); got != "[2]" {
		t.Fatalf("after InvalidateTemplate: got %q, want [2]", got)
	}
}
