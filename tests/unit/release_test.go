package miya_test

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	miya "github.com/zipreport/miya"
	"github.com/zipreport/miya/loader"
	"github.com/zipreport/miya/parser"
)

// GetSource-only, so GetTemplate takes its compile path (not loader.AdvancedLoader).
type releaseMapLoader map[string]string

func (l releaseMapLoader) GetSource(name string) (string, error) {
	if s, ok := l[name]; ok {
		return s, nil
	}
	return "", fmt.Errorf("template %q not found", name)
}
func (l releaseMapLoader) IsCached(string) bool             { return false }
func (l releaseMapLoader) ListTemplates() ([]string, error) { return nil, nil }

var releaseBaseChild = map[string]string{
	"base.html":  `[{% block a %}X{% endblock %}|{{ 1 + 1 }}]{% block b %}{% endblock %}`,
	"child.html": `{% extends "base.html" %}{% block b %}C{% endblock %}`,
}

func newReleaseEnv(templates map[string]string) *miya.Environment {
	sl := loader.NewStringLoader(loader.NewDirectTemplateParser())
	for name, src := range templates {
		sl.AddTemplate(name, src)
	}
	return miya.NewEnvironment(miya.WithLoader(sl), miya.WithAutoEscape(false))
}

func releaseRender(t *testing.T, tm *miya.Template, want string) {
	t.Helper()
	out, err := tm.Render(miya.NewContext())
	if err != nil {
		t.Fatalf("render error: %v (output %q)", err, out)
	}
	if out != want {
		t.Fatalf("got %q, want %q", out, want)
	}
}

func releaseMust(t *testing.T, tm *miya.Template, err error) *miya.Template {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
	return tm
}

func releaseCacheSize(t *testing.T, env *miya.Environment, want int) {
	t.Helper()
	if got := env.GetCacheSize(); got != want {
		t.Fatalf("cache size = %d, want %d", got, want)
	}
}

func TestReleaseIsSafe(t *testing.T) {
	t.Run("fromstring_release", func(t *testing.T) {
		env := newReleaseEnv(nil)
		a, err := env.FromString("{{ 1 + 2 }}")
		a = releaseMust(t, a, err)
		releaseRender(t, a, "3")
		a.Release()
		a.Release()
		b, err := env.FromString("{{ 1 + 2 }}")
		b = releaseMust(t, b, err)
		releaseRender(t, b, "3")
		releaseRender(t, a, "3")
		if a == b {
			t.Fatal("expected a re-parse after Release")
		}
		releaseCacheSize(t, env, 1)
	})

	t.Run("fromstring_evicts", func(t *testing.T) {
		env := newReleaseEnv(nil)
		a, err := env.FromString("{{ 1 + 2 }}")
		a = releaseMust(t, a, err)
		releaseCacheSize(t, env, 1)
		a.Release()
		releaseCacheSize(t, env, 0)
		releaseRender(t, a, "3")
	})

	t.Run("stale_pointer_keeps_new_entry", func(t *testing.T) {
		env := newReleaseEnv(nil)
		a, err := env.FromString("x")
		a = releaseMust(t, a, err)
		env.ClearCache()
		b, err := env.FromString("x")
		b = releaseMust(t, b, err)
		a.Release()
		c, err := env.FromString("x")
		c = releaseMust(t, c, err)
		if env.GetCacheSize() != 1 || c != b {
			t.Fatalf("size=%d c==b %v", env.GetCacheSize(), c == b)
		}
	})

	t.Run("gettemplate_release", func(t *testing.T) {
		env := newReleaseEnv(map[string]string{"p.html": "{{ 1 + 2 }}"})
		p, err := env.GetTemplate("p.html")
		p = releaseMust(t, p, err)
		p.Release()
		q, err := env.GetTemplate("p.html")
		q = releaseMust(t, q, err)
		releaseRender(t, q, "3")
		releaseRender(t, p, "3")
		if p == q {
			t.Fatal("expected a re-parse after Release")
		}
	})

	t.Run("gettemplate_compile_path_evicts", func(t *testing.T) {
		env := miya.NewEnvironment(miya.WithLoader(releaseMapLoader{"p.html": "{{ 1 + 2 }}"}), miya.WithAutoEscape(false))
		p, err := env.GetTemplate("p.html")
		p = releaseMust(t, p, err)
		releaseCacheSize(t, env, 1)
		p.Release()
		releaseCacheSize(t, env, 0)
		releaseRender(t, p, "3")
	})

	t.Run("pool_copy_release", func(t *testing.T) {
		env := newReleaseEnv(nil)
		a, err := env.FromString("{{ 1 + 2 }}")
		a = releaseMust(t, a, err)
		cp := miya.NewTemplatePool(a).Get()
		cp.Release()
		releaseRender(t, a, "3")
		releaseCacheSize(t, env, 1)
	})

	fillStrings := func(t *testing.T, env *miya.Environment) {
		for i := 0; i < 50; i++ {
			if _, err := env.FromString(fmt.Sprintf("{{ %d - 9 }}{{ y }}", i)); err != nil {
				t.Fatal(err)
			}
		}
	}

	t.Run("base_release", func(t *testing.T) {
		env := newReleaseEnv(releaseBaseChild)
		child, err := env.GetTemplate("child.html")
		child = releaseMust(t, child, err)
		releaseRender(t, child, "[X|2]C")
		base, err := env.GetTemplate("base.html")
		releaseMust(t, base, err).Release()
		c2, err := env.GetTemplate("child.html")
		releaseMust(t, c2, err).Release()
		fillStrings(t, env)
		c3, err := env.GetTemplate("child.html")
		c3 = releaseMust(t, c3, err)
		releaseRender(t, c3, "[X|2]C")
	})

	t.Run("parser_release_ast", func(t *testing.T) {
		env := newReleaseEnv(releaseBaseChild)
		child, err := env.GetTemplate("child.html")
		child = releaseMust(t, child, err)
		releaseRender(t, child, "[X|2]C")
		base, err := env.GetTemplate("base.html")
		base = releaseMust(t, base, err)
		parser.ReleaseAST(base.AST())
		fillStrings(t, env)
		releaseRender(t, child, "[X|2]C")
	})

	t.Run("concurrent", func(t *testing.T) {
		env := newReleaseEnv(releaseBaseChild)
		var wg sync.WaitGroup
		for g := 0; g < 8; g++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < 200; i++ {
					tm, err := env.FromString(fmt.Sprintf("{{ %d }}", i%5))
					if err != nil {
						t.Errorf("FromString: %v", err)
						return
					}
					out, err := tm.Render(miya.NewContext())
					if err != nil || out != fmt.Sprint(i%5) {
						t.Errorf("got %q err %v, want %d", out, err, i%5)
					}
					tm.Release()
					c, err := env.GetTemplate("child.html")
					if err != nil {
						t.Errorf("GetTemplate: %v", err)
						return
					}
					out, err = c.Render(miya.NewContext())
					if err != nil || out != "[X|2]C" {
						t.Errorf("child %q err %v", out, err)
					}
					c.Release()
				}
			}()
		}
		wg.Wait()
	})
}

func TestReleaseBoundsCache(t *testing.T) {
	env := miya.NewEnvironment()
	for i := 0; i < 5000; i++ {
		tm, err := env.FromString(fmt.Sprintf("%d {{ x }}", i))
		if err != nil {
			t.Fatal(err)
		}
		ctx := miya.NewContext()
		ctx.Set("x", 1)
		if _, err := tm.Render(ctx); err != nil {
			t.Fatal(err)
		}
		if i == 0 && env.GetCacheSize() != 1 {
			t.Fatalf("cache size = %d before Release, want 1", env.GetCacheSize())
		}
		tm.Release()
	}
	releaseCacheSize(t, env, 0)
}

var _ = strings.Repeat
