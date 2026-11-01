package miya_test

import (
	"strings"
	"testing"

	miya "github.com/zipreport/miya"
)

type immutabilityStruct struct {
	A string
}

func TestTemplatesCannotMutateData(t *testing.T) {
	env := miya.NewEnvironment(miya.WithAutoEscape(false))

	t.Run("map_attr", func(t *testing.T) {
		d := map[string]interface{}{"k": 0}
		_, err := env.RenderString(`{% set d.k = 1 %}`, miya.NewContextFrom(map[string]interface{}{"d": d}))
		if err == nil || !strings.Contains(err.Error(), "non-namespace object") {
			t.Fatalf("err = %v, want non-namespace object", err)
		}
		if d["k"] != 0 {
			t.Fatalf("caller map modified: %v", d)
		}
	})

	t.Run("map_item", func(t *testing.T) {
		d := map[string]interface{}{"k": 0}
		_, err := env.RenderString(`{% set d['k'] = 2 %}`, miya.NewContextFrom(map[string]interface{}{"d": d}))
		if err == nil || !strings.Contains(err.Error(), "cannot assign to an item") {
			t.Fatalf("err = %v, want cannot assign to an item", err)
		}
		if d["k"] != 0 {
			t.Fatalf("caller map modified: %v", d)
		}
	})

	t.Run("slice_item", func(t *testing.T) {
		l := []interface{}{1}
		_, err := env.RenderString(`{% set l[0] = 9 %}`, miya.NewContextFrom(map[string]interface{}{"l": l}))
		if err == nil || !strings.Contains(err.Error(), "cannot assign to an item") {
			t.Fatalf("err = %v, want cannot assign to an item", err)
		}
		if l[0] != 1 {
			t.Fatalf("caller slice modified: %v", l)
		}
	})

	t.Run("struct_ptr_attr", func(t *testing.T) {
		p := &immutabilityStruct{A: "a"}
		_, err := env.RenderString(`{% set p.A = 'z' %}`, miya.NewContextFrom(map[string]interface{}{"p": p}))
		if err == nil || !strings.Contains(err.Error(), "non-namespace object") {
			t.Fatalf("err = %v, want non-namespace object", err)
		}
		if p.A != "a" {
			t.Fatalf("caller struct modified: %v", p)
		}
	})

	t.Run("template_dict_attr", func(t *testing.T) {
		_, err := env.RenderString(`{% set x = dict() %}{% set x.k = 1 %}{{ x.k }}`, miya.NewContext())
		if err == nil || !strings.Contains(err.Error(), "non-namespace object") {
			t.Fatalf("err = %v, want non-namespace object", err)
		}
	})

	t.Run("context_attr", func(t *testing.T) {
		c := miya.NewContext()
		c.Set("k", 0)
		_, err := env.RenderString(`{% set c.k = 1 %}`, miya.NewContextFrom(map[string]interface{}{"c": c}))
		if err == nil || !strings.Contains(err.Error(), "non-namespace object") {
			t.Fatalf("err = %v, want non-namespace object", err)
		}
		if v, _ := c.Get("k"); v != 0 {
			t.Fatalf("caller context modified: k = %v", v)
		}
	})

	t.Run("namespace_ok", func(t *testing.T) {
		got, err := env.RenderString(`{% set n = namespace(a=1) %}{% set n.a = 2 %}{{ n.a }}`, miya.NewContext())
		if err != nil || got != "2" {
			t.Fatalf("got %q, err %v; want 2", got, err)
		}
	})
}
