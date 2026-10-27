package miya

import (
	"testing"

	"github.com/zipreport/miya/loader"
)

// GetTemplate goes to the loader even for the name FromString caches under.
func TestGetTemplateIgnoresFromStringCache(t *testing.T) {
	src := "from string"
	key := hashString(src)
	sl := loader.NewStringLoader(loader.NewDirectTemplateParser())
	sl.AddTemplate(key, "from loader")
	env := NewEnvironment(WithLoader(sl))
	if _, err := env.FromString(src); err != nil {
		t.Fatal(err)
	}
	tmpl, err := env.GetTemplate(key)
	if err != nil {
		t.Fatal(err)
	}
	if out, _ := tmpl.Render(NewContext()); out != "from loader" {
		t.Fatalf("GetTemplate(%q) rendered %q, want the loader's template", key, out)
	}
	if tmpl, _ := env.FromString(src); tmpl.source != src {
		t.Fatalf("FromString returned the template of %q", tmpl.source)
	}
}

// A FromString cache hit must be for the same source, not just the same hash.
func TestFromStringHashCollision(t *testing.T) {
	env := NewEnvironment()
	a, err := env.FromString("A")
	if err != nil {
		t.Fatal(err)
	}
	// simulate "B" hashing like "A"
	env.cacheMutex.Lock()
	env.stringCache[hashString("B")] = a
	env.cacheMutex.Unlock()

	b, err := env.FromString("B")
	if err != nil {
		t.Fatal(err)
	}
	if out, _ := b.Render(NewContext()); out != "B" {
		t.Fatalf("FromString(\"B\") rendered %q", out)
	}
}
