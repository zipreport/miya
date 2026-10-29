package miya_test

import (
	"strings"
	"testing"

	"github.com/zipreport/miya"
)

func TestRenderAsyncResultFromOutside(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		env := miya.NewEnvironment()
		tmpl, err := env.FromString("Hello {{ name }}!")
		if err != nil {
			t.Fatal(err)
		}
		r := miya.NewConcurrentTemplateRenderer(tmpl, 2)
		r.Start()
		defer r.Stop()
		ctx := miya.NewContext()
		ctx.Set("name", "World")
		res := <-r.RenderAsync(ctx)
		if res.Err != nil {
			t.Fatalf("unexpected error: %v", res.Err)
		}
		if res.Output != "Hello World!" {
			t.Fatalf("Output = %q", res.Output)
		}
	})

	t.Run("error", func(t *testing.T) {
		env := miya.NewEnvironment(miya.WithStrictUndefined(true))
		tmpl, err := env.FromString("{{ x.y.z }}")
		if err != nil {
			t.Fatal(err)
		}
		r := miya.NewConcurrentTemplateRenderer(tmpl, 2)
		r.Start()
		defer r.Stop()
		res := <-r.RenderAsync(miya.NewContext())
		if res.Err == nil {
			t.Fatal("expected an error")
		}
		if res.Output != "" {
			t.Fatalf("Output = %q", res.Output)
		}
	})

	t.Run("not_started", func(t *testing.T) {
		env := miya.NewEnvironment()
		tmpl, err := env.FromString("x")
		if err != nil {
			t.Fatal(err)
		}
		r := miya.NewConcurrentTemplateRenderer(tmpl, 1)
		res := <-r.RenderAsync(miya.NewContext())
		if res.Err == nil || !strings.Contains(res.Err.Error(), "not started") {
			t.Fatalf("Err = %v", res.Err)
		}
	})
}
