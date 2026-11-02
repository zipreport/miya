package miya

import (
	"fmt"
	"strings"
	"testing"

	"github.com/zipreport/miya"
	"github.com/zipreport/miya/loader"
)

// issue37Chain renders tN, where t0 is [{% block a %}x{% endblock %}] and
// every tI extends t(I-1) and wraps super() in < >.
func issue37Chain(t *testing.T, n int) (string, error) {
	t.Helper()
	sl := loader.NewStringLoader(loader.NewDirectTemplateParser())
	sl.AddTemplate("t0", `[{% block a %}x{% endblock %}]`)
	for i := 1; i <= n; i++ {
		sl.AddTemplate(fmt.Sprintf("t%d", i), fmt.Sprintf(`{%% extends 't%d' %%}{%% block a %%}<{{ super() }}>{%% endblock %%}`, i-1))
	}
	env := miya.NewEnvironment(miya.WithLoader(sl))
	tmpl, err := env.GetTemplate(fmt.Sprintf("t%d", n))
	if err != nil {
		t.Fatal(err)
	}
	return tmpl.Render(miya.NewContext())
}

func issue37Want(n int) string {
	return "[" + strings.Repeat("<", n) + "x" + strings.Repeat(">", n) + "]"
}

// Claim 1: the issue's chain of 12 templates (t0 to t11) renders as in Jinja2.
func TestIssue37ElevenLevelSuperChainRenders(t *testing.T) {
	got, err := issue37Chain(t, 11)
	if err != nil || got != issue37Want(11) {
		t.Fatalf("got %q, %v; want %q", got, err, issue37Want(11))
	}
}

// Claim 2: a chain of 31 templates (t0 to t30) renders with 30 pairs of < >.
func TestIssue37ThirtyLevelSuperChainRenders(t *testing.T) {
	got, err := issue37Chain(t, 30)
	if err != nil || got != issue37Want(30) {
		t.Fatalf("got %q, %v; want %q", got, err, issue37Want(30))
	}
}

// Guard: the chain of 11 templates (t0 to t10) that rendered before still
// renders.
func TestIssue37TenLevelSuperChainRenders(t *testing.T) {
	got, err := issue37Chain(t, 10)
	if err != nil || got != issue37Want(10) {
		t.Fatalf("got %q, %v; want %q", got, err, issue37Want(10))
	}
}

func issue37CycleErr(t *testing.T, templates map[string]string, name string) error {
	t.Helper()
	sl := loader.NewStringLoader(loader.NewDirectTemplateParser())
	for k, v := range templates {
		sl.AddTemplate(k, v)
	}
	env := miya.NewEnvironment(miya.WithLoader(sl))
	tmpl, err := env.GetTemplate(name)
	if err != nil {
		t.Fatal(err)
	}
	got, err := tmpl.Render(miya.NewContext())
	if err == nil {
		t.Fatalf("got %q, nil; want the super() cycle error", got)
	}
	return err
}

const issue37Base = "{% block a %}A{% block b %}B{% endblock %}{% endblock %}"

// Claim 3: the cycle error names the template and line of the super() that
// closes the cycle.
func TestIssue37CycleErrorNamesTemplateAndLine(t *testing.T) {
	err := issue37CycleErr(t, map[string]string{
		"base": issue37Base,
		"c":    "{% extends 'base' %}\n{% block b %}{{ super() }}\n\n{% block a %}{{ super() }}{% endblock %}{% endblock %}",
	}, "c")
	if want := `super() in block "a" renders itself in template "c" at line 4`; !strings.Contains(err.Error(), want) {
		t.Fatalf("got %v; want it to contain %q", err, want)
	}
}

// Guard: a super() inside an expression closing the cycle is located too.
func TestIssue37CycleErrorInExpressionNamesTemplateAndLine(t *testing.T) {
	err := issue37CycleErr(t, map[string]string{
		"base": issue37Base,
		"c":    "{% extends 'base' %}\n{% block b %}{{ super() }}\n\n\n{% block a %}{% set x = super() %}{{ x }}{% endblock %}{% endblock %}",
	}, "c")
	if want := `super() in block "a" renders itself in template "c" at line 5`; !strings.Contains(err.Error(), want) {
		t.Fatalf("got %v; want it to contain %q", err, want)
	}
}

// Guard: when the super() closing the cycle is in a middle template, that
// template is named, not the one being rendered.
func TestIssue37CycleErrorNamesMiddleTemplate(t *testing.T) {
	err := issue37CycleErr(t, map[string]string{
		"base": issue37Base,
		"mid":  "{% extends 'base' %}\n\n{% block b %}{{ super() }}{% block a %}{{ super() }}{% endblock %}{% endblock %}",
		"c":    "{% extends 'mid' %}{% block z %}{% endblock %}",
	}, "c")
	if want := `super() in block "a" renders itself in template "mid" at line 3`; !strings.Contains(err.Error(), want) {
		t.Fatalf("got %v; want it to contain %q", err, want)
	}
}

// Guard: when the recursion limit is hit inside the cycle by a macro call,
// the error still names the template and line of the super().
func TestIssue37CycleErrorThroughMacroNamesTemplateAndLine(t *testing.T) {
	err := issue37CycleErr(t, map[string]string{
		"base": issue37Base,
		"c":    "{% extends 'base' %}{% macro w() %}{{ caller() }}{% endmacro %}\n{% block b %}{% block a %}{% call w() %}\n{{ super() }}{% endcall %}{% endblock %}{% endblock %}",
	}, "c")
	if want := `super() in block "a" renders itself in template "c" at line 3`; !strings.Contains(err.Error(), want) {
		t.Fatalf("got %v; want it to contain %q", err, want)
	}
}
