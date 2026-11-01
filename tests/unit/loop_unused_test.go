package miya_test

import (
	"testing"

	miya "github.com/zipreport/miya"
	"github.com/zipreport/miya/loader"
)

// A loop body that reaches loop only through a block override, an include or
// an import with context still gets it (miya shows loop there; Jinja2 does not).
func TestLoopReachedIndirectly(t *testing.T) {
	sl := loader.NewStringLoader(loader.NewDirectTemplateParser())
	sl.AddTemplate("inc", "[{{ loop.index }}]")
	sl.AddTemplate("mod", "{% set n = loop.index %}")
	sl.AddTemplate("base", "{% for i in [1, 2] %}{% block b %}{% endblock %}{% endfor %}")
	sl.AddTemplate("block", "{% extends 'base' %}{% block b %}[{{ loop.index }}]{% endblock %}")
	sl.AddTemplate("include", "{% for i in [1, 2] %}{% include 'inc' %}{% endfor %}")
	sl.AddTemplate("import", "{% for i in [1, 2] %}{% import 'mod' as m with context %}[{{ m.n }}]{% endfor %}")
	env := miya.NewEnvironment(miya.WithLoader(sl))
	for _, name := range []string{"block", "include", "import"} {
		t.Run(name, func(t *testing.T) {
			tmpl, err := env.GetTemplate(name)
			if err != nil {
				t.Fatal(err)
			}
			got, err := tmpl.Render(miya.NewContext())
			if err != nil || got != "[1][2]" {
				t.Errorf("got %q, %v; want \"[1][2]\"", got, err)
			}
		})
	}
}
