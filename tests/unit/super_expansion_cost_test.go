package miya_test

import (
	"fmt"
	"strings"
	"testing"

	miya "github.com/zipreport/miya"
	"github.com/zipreport/miya/loader"
)

// superNested returns a base of n nested blocks b0 > b1 > ... each printing
// "x", and a child that overrides every block with body.
func superNested(n int, body string) map[string]string {
	var b, c strings.Builder
	for i := 0; i < n; i++ {
		fmt.Fprintf(&b, "{%% block b%d %%}x", i)
	}
	b.WriteString(strings.Repeat("{% endblock %}", n))
	c.WriteString("{% extends 'base' %}")
	for i := 0; i < n; i++ {
		fmt.Fprintf(&c, "{%% block b%d %%}%s{%% endblock %%}", i, body)
	}
	return map[string]string{"base": b.String(), "c": c.String()}
}

func superRender(t *testing.T, env *miya.Environment, name string) string {
	t.Helper()
	tmpl, err := env.GetTemplate(name)
	if err != nil {
		t.Fatal(err)
	}
	out, err := tmpl.Render(miya.NewContext())
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// Two super() calls per block render as in Jinja2 3.1.6, on the first render
// and from the cached merged template.
func TestSuperTwicePerNestedBlockOutput(t *testing.T) {
	tests := []struct {
		name, body, want string
	}{
		{"one_branch_never_runs", "{% if false %}{{ super() }}{% endif %}{{ super() }}", "xxx"},
		{"both_run", "{{ super() }}|{{ super() }}", "xxx|x|xx|x|xxx|x|xx|x"},
		{"in_else_branch", "{% for i in [1, 2] %}{% if i == 1 %}a{% else %}{{ super() }}{% endif %}{% endfor %}", "axaxax"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := extendsTestEnv(superNested(3, tt.body))
			for i := 0; i < 2; i++ {
				if got := superRender(t, env, "c"); got != tt.want {
					t.Fatalf("render %d: got %q, want %q", i+1, got, tt.want)
				}
			}
		})
	}
}

// A parent block rendered by several super() calls (in a loop, under
// {% autoescape false %}) gets a fresh scope each time, as in Jinja2 3.1.6.
func TestSuperRepeatedParentBlockScope(t *testing.T) {
	templates := map[string]string{
		"base": "{% block a %}{% set n = (n or 0) + 1 %}<{{ n }}{% for i in [1, 2] %}{{ i }}{% endfor %}>{% endblock %}",
		"c":    "{% extends 'base' %}{% block a %}{{ super() }}{% for j in [1] %}{{ super() }}{% endfor %}{% autoescape false %}{{ super() }}{% endautoescape %}{% endblock %}",
	}
	for _, ae := range []bool{false, true} {
		sl := loader.NewStringLoader(loader.NewDirectTemplateParser())
		for k, v := range templates {
			sl.AddTemplate(k, v)
		}
		env := miya.NewEnvironment(miya.WithLoader(sl), miya.WithAutoEscape(ae))
		for i := 0; i < 2; i++ {
			if got, want := superRender(t, env, "c"), "<112><112><112>"; got != want {
				t.Fatalf("autoescape %v render %d: got %q, want %q", ae, i+1, got, want)
			}
		}
	}
}

// Building a child's merged template leaves the parsed templates as they
// were: the parent still renders its own blocks afterwards, and no parsed
// tree changes.
func TestSuperBuildLeavesParsedTemplates(t *testing.T) {
	env := extendsTestEnv(map[string]string{
		"base": "[{% block a %}{% for i in [1] %}{% block b %}B{% endblock %}{% endfor %}{% endblock %}]",
		"c":    "{% extends 'base' %}{% block a %}{{ super() }}{% endblock %}{% block b %}C{% endblock %}",
	})
	fromString, err := env.FromString("{% extends 'base' %}{% block a %}{{ super() }}{% endblock %}{% block b %}F{% endblock %}")
	if err != nil {
		t.Fatal(err)
	}
	parsed := map[string]*miya.Template{"from_string": fromString}
	for _, name := range []string{"base", "c"} {
		tmpl, err := env.GetTemplate(name)
		if err != nil {
			t.Fatal(err)
		}
		parsed[name] = tmpl
	}
	before := map[string]string{}
	for name, tmpl := range parsed {
		before[name] = tmpl.AST().String()
	}
	for i := 0; i < 2; i++ {
		if got := superRender(t, env, "c"); got != "[C]" {
			t.Fatalf("child render %d: got %q, want %q", i+1, got, "[C]")
		}
		if out, err := fromString.Render(miya.NewContext()); err != nil || out != "[F]" {
			t.Fatalf("FromString child render %d: got %q, %v; want %q", i+1, out, err, "[F]")
		}
	}
	if got := superRender(t, env, "base"); got != "[B]" {
		t.Fatalf("base after its children: got %q, want %q", got, "[B]")
	}
	for name, tmpl := range parsed {
		if after := tmpl.AST().String(); after != before[name] {
			t.Errorf("rendering changed the parsed %s template:\nbefore %s\nafter  %s", name, before[name], after)
		}
	}
}

// Building the merged template of a hierarchy whose overrides each call
// super() twice costs allocations linear in the nesting depth (see #36: each
// level used to double the tree, 27M allocations at depth 20).
func TestSuperTwicePerNestedBlockBuildsLinearly(t *testing.T) {
	build := func(depth int) float64 {
		templates := superNested(depth, "{% if false %}{{ super() }}{% endif %}{{ super() }}")
		var out string
		allocs := testing.AllocsPerRun(1, func() {
			out = superRender(t, extendsTestEnv(templates), "c")
		})
		if want := strings.Repeat("x", depth); out != want {
			t.Fatalf("depth %d: got %q, want %q", depth, out, want)
		}
		return allocs
	}
	a10, a20 := build(10), build(20)
	if a20 > 30000 || a20 > 3*a10 {
		t.Fatalf("building a hierarchy with two super() calls per block took %.0f allocations at depth 10 and %.0f at depth 20; it must be linear in depth (about 1500 and 3000)", a10, a20)
	}
}

// A FromString child is rebuilt on every render. A super() call adds work
// for the blocks it renders, not a copy of every statement in the template.
func TestSuperCallCostPerRenderFromString(t *testing.T) {
	const stmts = 2000
	loops := strings.Repeat("{% for i in [1] %}{% set q = i %}{% endfor %}", stmts)
	env := extendsTestEnv(map[string]string{
		"base":  "[{% block a %}x{% endblock %}]",
		"large": "[{% block a %}" + loops + "{% endblock %}]",
	})
	allocs := func(t *testing.T, src, want string) float64 {
		t.Helper()
		tmpl, err := env.FromString(src)
		if err != nil {
			t.Fatal(err)
		}
		ctx := miya.NewContext()
		out, err := tmpl.Render(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if out != want {
			t.Fatalf("got %q, want %q", out, want)
		}
		return testing.AllocsPerRun(10, func() { _, _ = tmpl.Render(ctx) })
	}
	tests := []struct {
		name, without, with, wantWithout, wantWith string
		maxExtra                                   float64
	}{
		// top-level statements of the child, outside any block
		{"hoisted", "{% extends 'base' %}" + loops + "{% block a %}y{% endblock %}",
			"{% extends 'base' %}" + loops + "{% block a %}{{ super() }}{% endblock %}", "[y]", "[x]", 3 * stmts},
		// statements in the overriding block, around the super() call
		{"in_child_block", "{% extends 'base' %}{% block a %}" + loops + "{% endblock %}",
			"{% extends 'base' %}{% block a %}" + loops + "{{ super() }}{% endblock %}", "[]", "[x]", 3000},
		// statements in the parent block super() renders
		{"in_parent_block", "{% extends 'large' %}",
			"{% extends 'large' %}{% block a %}{{ super() }}{% endblock %}", "[]", "[]", 1000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			without := allocs(t, tt.without, tt.wantWithout)
			with := allocs(t, tt.with, tt.wantWith)
			if with-without > tt.maxExtra {
				t.Fatalf("super() adds %.0f allocations per render (%.0f without, %.0f with); want at most %.0f", with-without, without, with, tt.maxExtra)
			}
			if tt.name == "hoisted" && with > 1.2*without {
				t.Fatalf("super() costs %.2fx per render (%.0f without, %.0f with); want at most 1.2x", with/without, without, with)
			}
		})
	}
}
