package miya_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	miya "github.com/zipreport/miya"
	"github.com/zipreport/miya/loader"
	"github.com/zipreport/miya/runtime"
)

// Expected values are Jinja2 3.1.6's.

const (
	issue35LoopBase  = "{% block a %}A{% if x is not defined %}{% block b %}B{% endblock %}{% endif %}{% endblock %}"
	issue35LoopChild = "{% extends 'base' %}{% block b %}b{% for x in [1] %}{% block a scoped %}a{{ super() }}{% endblock %}{% endfor %}{% endblock %}"
	issue35NsBase    = "{% set ns = namespace(n=0) %}{% block a %}A{% set ns.n = ns.n + 1 %}{% if ns.n < 2 %}{% block b %}B{% endblock %}{% endif %}{% endblock %}"
	issue35NsChild   = "{% extends 'base' %}{% block b %}b{% block a scoped %}a{{ super() }}{% endblock %}{% endblock %}"
	issue35Cycle     = "{% extends 'base' %}{% block b %}{{ super() }}{% block a %}{{ super() }}{% endblock %}{% endblock %}"
)

func issue35Env(templates map[string]string, autoescape bool) *miya.Environment {
	sl := loader.NewStringLoader(loader.NewDirectTemplateParser())
	for k, v := range templates {
		sl.AddTemplate(k, v)
	}
	return miya.NewEnvironment(miya.WithLoader(sl), miya.WithAutoEscape(autoescape))
}

func issue35Render(t *testing.T, env *miya.Environment, name string, data map[string]interface{}) (string, error) {
	t.Helper()
	tmpl, err := env.GetTemplate(name)
	if err != nil {
		t.Fatal(err)
	}
	return tmpl.Render(miya.NewContextFrom(data))
}

// Claim 1: a super() cycle through nested blocks that a condition stops
// renders, here with the condition on a loop variable.
func TestIssue35GuardedCycleInLoop(t *testing.T) {
	tests := []struct {
		name, base, child string
		autoescape        bool
		data              map[string]interface{}
		want              string
	}{
		{"plain", issue35LoopBase, issue35LoopChild, false, nil, "aAbaA"},
		{"super_in_expression", issue35LoopBase, strings.Replace(issue35LoopChild, "a{{ super() }}", "a{{ super()|lower }}{{ super()|length }}", 1), false, nil, "aabaa15"},
		{"autoescape", strings.Replace(issue35LoopBase, "A{% if", "A{{ v }}{% if", 1), issue35LoopChild, true, map[string]interface{}{"v": "<i>"}, "aA&lt;i&gt;baA&lt;i&gt;"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := issue35Env(map[string]string{"base": tt.base, "c": tt.child}, tt.autoescape)
			for i := 0; i < 2; i++ {
				got, err := issue35Render(t, env, "c", tt.data)
				if err != nil || got != tt.want {
					t.Fatalf("render %d: got %q, %v; want %q", i+1, got, err, tt.want)
				}
			}
		})
	}
	t.Run("from_string", func(t *testing.T) {
		env := issue35Env(map[string]string{"base": issue35LoopBase}, false)
		tmpl, err := env.FromString(issue35LoopChild)
		if err != nil {
			t.Fatal(err)
		}
		for i := 0; i < 2; i++ {
			if got, err := tmpl.Render(miya.NewContext()); err != nil || got != "aAbaA" {
				t.Fatalf("render %d: got %q, %v; want %q", i+1, got, err, "aAbaA")
			}
		}
	})
}

// Claim 2: the same, with the condition on a namespace counter, stopping
// after one or three turns.
func TestIssue35GuardedCycleWithNamespace(t *testing.T) {
	for _, tt := range []struct {
		limit, want string
	}{{"2", "aAbaA"}, {"4", "aAbaAbaAbaA"}} {
		t.Run("limit_"+tt.limit, func(t *testing.T) {
			env := issue35Env(map[string]string{"base": strings.Replace(issue35NsBase, "ns.n < 2", "ns.n < "+tt.limit, 1), "c": issue35NsChild}, false)
			got, err := issue35Render(t, env, "c", nil)
			if err != nil || got != tt.want {
				t.Fatalf("got %q, %v; want %q", got, err, tt.want)
			}
		})
	}
}

// Claim 3: a cycle that nothing stops still fails when rendered (Jinja2
// raises RecursionError).
func TestIssue35UnconditionalCycleErrors(t *testing.T) {
	env := issue35Env(map[string]string{"base": "{% block a %}A{% block b %}B{% endblock %}{% endblock %}", "c": issue35Cycle}, false)
	got, err := issue35Render(t, env, "c", nil)
	if err == nil || !strings.Contains(err.Error(), `super() in block "a" renders itself`) {
		t.Fatalf("got %q, %v; want the super() cycle error", got, err)
	}
}

// Claim 5: that error comes from following the cycle until the recursion
// limit, so callers can match it as one.
func TestIssue35UnconditionalCycleIsMaxRecursion(t *testing.T) {
	env := issue35Env(map[string]string{"base": "{% block a %}A{% block b %}B{% endblock %}{% endblock %}", "c": issue35Cycle}, false)
	_, err := issue35Render(t, env, "c", nil)
	if !errors.Is(err, runtime.ErrMaxRecursion) {
		t.Fatalf("got %v; want an error matching runtime.ErrMaxRecursion", err)
	}
}

// Guard: when the loop variable is already defined, the condition stops the
// recursion at the first turn.
func TestIssue35GuardedCycleStopsAtOnce(t *testing.T) {
	env := issue35Env(map[string]string{"base": issue35LoopBase, "c": issue35LoopChild}, false)
	got, err := issue35Render(t, env, "c", map[string]interface{}{"x": 1})
	if err != nil || got != "aA" {
		t.Fatalf("got %q, %v; want %q", got, err, "aA")
	}
}

// Guard: a render that fails on a super() cycle leaves the recursion count as
// it was, so the next renders' macro calls still work.
func TestIssue35CycleErrorLeavesNoDepth(t *testing.T) {
	env := issue35Env(map[string]string{
		"base": "{% block a %}A{% block b %}B{% endblock %}{% endblock %}",
		"c":    issue35Cycle,
		"ok":   "{% macro m(n) %}{% if n %}{{ m(n - 1) }}{% endif %}M{% endmacro %}{{ m(150) | length }}",
	}, false)
	for i := 0; i < 3; i++ {
		if _, err := issue35Render(t, env, "c", nil); err == nil {
			t.Fatalf("round %d: want the cycle error", i+1)
		}
		if got, err := issue35Render(t, env, "ok", nil); err != nil || got != "151" {
			t.Fatalf("round %d: got %q, %v; want %q", i+1, got, err, "151")
		}
	}
}

// issue35Deep returns templates for depth n: an outer block a around n+1
// nested blocks; the child overrides each but the last with two super()
// calls, and the last with a block a whose super() never runs.
func issue35Deep(n int) map[string]string {
	var b, c strings.Builder
	b.WriteString("{% block a %}")
	for i := 0; i <= n; i++ {
		fmt.Fprintf(&b, "{%% block b%d %%}x", i)
	}
	b.WriteString(strings.Repeat("{% endblock %}", n+2))
	c.WriteString("{% extends 'base' %}")
	for i := 0; i < n; i++ {
		fmt.Fprintf(&c, "{%% block b%d %%}{{ super() }}{{ super() }}{%% endblock %%}", i)
	}
	fmt.Fprintf(&c, "{%% block b%d %%}{%% block a %%}{%% if false %%}{{ super() }}{%% endif %%}{%% endblock %%}{%% endblock %%}", n)
	return map[string]string{"base": b.String(), "c": c.String()}
}

// Claim 4: content holding a super() cycle back to an outer block is built
// once, so the build is linear in depth (it doubled per level).
func TestIssue35CycleContentBuildsLinearly(t *testing.T) {
	build := func(depth int) float64 {
		templates := issue35Deep(depth)
		var out string
		var err error
		allocs := testing.AllocsPerRun(1, func() {
			out, err = issue35Render(t, issue35Env(templates, false), "c", nil)
		})
		if err != nil || out != "" {
			t.Fatalf("depth %d: got %q, %v; want %q", depth, out, err, "")
		}
		return allocs
	}
	a9, a18 := build(9), build(18)
	if a18 > 50000 || a18 > 3*a9 {
		t.Fatalf("building took %.0f allocations at depth 9 and %.0f at depth 18; it must be linear in depth", a9, a18)
	}
}

// Claims 1 and 2 through multi-level chains: a guarded cycle whose content is
// cached and reused where it is no cycle, and cycles that pass through a
// middle template's overrides.
func TestIssue35GuardedCycleChains(t *testing.T) {
	tests := []struct {
		name      string
		templates map[string]string
		want      string
	}{
		{"content_reused_outside_cycle", map[string]string{
			"base":  "{% set ns = namespace(n=0) %}{% block a %}A{% set ns.n = ns.n + 1 %}{% if ns.n < 3 %}{% block b %}B{% endblock %}{% endif %}{% endblock %}|{% block c %}{% endblock %}",
			"mid":   "{% extends 'base' %}{% block c %}{% block b %}u({% block a %}ua{{ super() }}{% endblock %}){% endblock %}{% endblock %}",
			"child": "{% extends 'mid' %}{% block b %}c{{ super() }}{% endblock %}",
		}, "uaAcu(uaAcu(uaA))|cu(uaA)"},
		{"mid_overrides_outer_block", map[string]string{
			"base":  issue35LoopBase,
			"mid":   "{% extends 'base' %}{% block a %}M{{ super() }}{% endblock %}",
			"child": strings.Replace(issue35LoopChild, "'base'", "'mid'", 1),
		}, "aMAbaMA"},
		{"mid_holds_cycle", map[string]string{
			"base":  issue35LoopBase,
			"mid":   strings.Replace(issue35LoopChild, "a{{ super() }}", "m{{ super() }}", 1),
			"child": "{% extends 'mid' %}{% block a %}c{{ super() }}{% endblock %}",
		}, "cmAbcmA"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := issue35Render(t, issue35Env(tt.templates, false), "child", nil)
			if err != nil || got != tt.want {
				t.Fatalf("got %q, %v; want %q", got, err, tt.want)
			}
		})
	}
}

// Claim 3 when the cycle runs through a macro call or a set body: the error
// still names the cycle, and is not wrapped once per level.
func TestIssue35CycleErrorThroughWrappers(t *testing.T) {
	base := "{% block a %}A{% block b %}B{% endblock %}{% endblock %}"
	for _, tt := range []struct{ name, child string }{
		{"call_block", "{% extends 'base' %}{% macro w() %}{{ caller() }}{% endmacro %}{% block b %}{% block a %}{% call w() %}{{ super() }}{% endcall %}{% endblock %}{% endblock %}"},
		{"set_body", "{% extends 'base' %}{% block b %}{% block a %}{% set s %}{{ super() }}{% endset %}{{ s }}{% endblock %}{% endblock %}"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			_, err := issue35Render(t, issue35Env(map[string]string{"base": base, "c": tt.child}, false), "c", nil)
			if err == nil || !strings.Contains(err.Error(), `super() in block "a" renders itself`) {
				t.Fatalf("got %v; want the super() cycle error", err)
			}
			if n := len(err.Error()); n > 1000 {
				t.Fatalf("the error is %d bytes; want it wrapped once, not once per level: %.300s", n, err.Error())
			}
		})
	}
}
