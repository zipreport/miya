package miya_test

import (
	"testing"

	miya "github.com/zipreport/miya"
	"github.com/zipreport/miya/loader"
)

// Issue #4: InvalidateTemplate on a parent reaches the templates that extend it.
func TestIssue4InvalidateParentReachesChildren(t *testing.T) {
	setup := func(templates map[string]string) (*loader.StringLoader, *miya.Environment) {
		sl := loader.NewStringLoader(loader.NewDirectTemplateParser())
		for n, s := range templates {
			sl.AddTemplate(n, s)
		}
		return sl, miya.NewEnvironment(miya.WithLoader(sl))
	}
	render := func(t *testing.T, env *miya.Environment, name string) string {
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
	two := map[string]string{
		"base.html":  "[old {% block b %}{% endblock %}]",
		"child.html": "{% extends 'base.html' %}{% block b %}C{% endblock %}",
		"other.html": "other",
	}
	three := map[string]string{
		"g.html":    "<g1 {% block b %}{% endblock %}>",
		"m.html":    "{% extends 'g.html' %}{% block b %}m1 {{ super() }}{% endblock %}",
		"leaf.html": "{% extends 'm.html' %}{% block b %}C {{ super() }}{% endblock %}",
	}

	t.Run("parent", func(t *testing.T) {
		sl, env := setup(two)
		if got := render(t, env, "child.html"); got != "[old C]" {
			t.Fatalf("first render: %q", got)
		}
		sl.AddTemplate("base.html", "[new {% block b %}{% endblock %}]")
		env.InvalidateTemplate("base.html")
		if got := render(t, env, "child.html"); got != "[new C]" {
			t.Errorf("after InvalidateTemplate(base.html): got %q, want %q", got, "[new C]")
		}
	})

	t.Run("grandparent", func(t *testing.T) {
		sl, env := setup(three)
		if got := render(t, env, "leaf.html"); got != "<g1 C m1 >" {
			t.Fatalf("first render: %q", got)
		}
		sl.AddTemplate("g.html", "<g2 {% block b %}{% endblock %}>")
		env.InvalidateTemplate("g.html")
		if got := render(t, env, "leaf.html"); got != "<g2 C m1 >" {
			t.Errorf("after InvalidateTemplate(g.html): got %q, want %q", got, "<g2 C m1 >")
		}
	})

	t.Run("middle", func(t *testing.T) {
		sl, env := setup(three)
		render(t, env, "leaf.html")
		sl.AddTemplate("m.html", "{% extends 'g.html' %}{% block b %}m2 {{ super() }}{% endblock %}")
		env.InvalidateTemplate("m.html")
		if got := render(t, env, "leaf.html"); got != "<g1 C m2 >" {
			t.Errorf("after InvalidateTemplate(m.html): got %q, want %q", got, "<g1 C m2 >")
		}
	})

	// guard: invalidating an unrelated template keeps the child's cached
	// hierarchy (a changed but not invalidated parent stays unseen)
	t.Run("unrelated_keeps_cache", func(t *testing.T) {
		sl, env := setup(two)
		render(t, env, "child.html")
		sl.AddTemplate("base.html", "[new {% block b %}{% endblock %}]")
		env.InvalidateTemplate("other.html")
		if n := env.GetInheritanceCacheStats().HierarchyCache.Entries; n != 1 {
			t.Errorf("after InvalidateTemplate(other.html): %d cached hierarchies, want 1 (child.html's)", n)
		}
		if got := render(t, env, "child.html"); got != "[old C]" {
			t.Errorf("after InvalidateTemplate(other.html): got %q, want the cached %q", got, "[old C]")
		}
	})

	// guard: invalidating the child itself still reloads it
	t.Run("child_itself", func(t *testing.T) {
		sl, env := setup(two)
		render(t, env, "child.html")
		sl.AddTemplate("child.html", "{% extends 'base.html' %}{% block b %}D{% endblock %}")
		env.InvalidateTemplate("child.html")
		if got := render(t, env, "child.html"); got != "[old D]" {
			t.Errorf("after InvalidateTemplate(child.html): got %q, want %q", got, "[old D]")
		}
	})
}
