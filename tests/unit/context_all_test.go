package miya_test

import (
	"sync"
	"testing"

	miya "github.com/zipreport/miya"
)

// One Context rendered from several goroutines; run with -race.
func TestContextSharedAcrossRenders(t *testing.T) {
	env := miya.NewEnvironment()
	tmpl, err := env.FromString(`{{ a }}-{{ b }}`)
	if err != nil {
		t.Fatal(err)
	}
	ctx := miya.NewContextFrom(map[string]interface{}{"a": 1, "b": 2})
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				if out, err := tmpl.Render(ctx); err != nil || out != "1-2" {
					t.Errorf("got %q, %v", out, err)
					return
				}
			}
		}()
	}
	wg.Wait()
}

// All() reflects later writes to a parent context.
func TestContextAllSeesParentSet(t *testing.T) {
	parent := miya.NewContext()
	child := parent.Push()
	child.Set("c", 1)
	_ = child.All()
	parent.Set("p", 2)
	if got, ok := child.All()["p"]; !ok || got != 2 {
		t.Fatalf("child.All()[p] = %v, %v; want 2", got, ok)
	}
}
