package miya_test

import (
	"strings"
	"testing"
	"time"

	miya "github.com/zipreport/miya"
)

func TestMapKeyRule(t *testing.T) {
	one := "1"
	data := map[string]interface{}{
		"ms":   map[string]string{"None": "N", "": "E", "1": "one"},
		"ifm":  map[interface{}]interface{}{nil: "nil", 1: "one", "a": "A"},
		"yam":  map[interface{}]interface{}{"None": "N", "1": "one", "true": "T"},
		"d":    map[string]interface{}{"None": "N", "": "E"},
		"p":    &one,
		"bs":   []byte("1"),
		"ifm1": map[interface{}]interface{}{1: "one"},
		"ift":  map[interface{}]interface{}{true: "T"},
	}
	rows := []struct{ name, tpl, want string }{
		// Jinja2: [|None|False]
		{"deviation_ms_none", `[{{ ms[none] }}|{{ ms.get(none) }}|{{ none in ms }}]`, `[N|N|true]`},
		{"d_none", `[{{ d[none] }}|{{ d.get(none) }}|{{ none in d }}]`, `[N|N|true]`},
		// Jinja2: [|one|False] (1 is not "1")
		{"deviation_ms_int", `[{{ ms[1] }}|{{ ms.get(1) }}|{{ 1 in ms }}]`, `[one|one|true]`},
		{"ifm_none", `[{{ ifm[none] }}|{{ ifm.get(none) }}|{{ none in ifm }}]`, `[nil|nil|true]`},
		{"ifm_float", `[{{ ifm[1.0] }}|{{ ifm.get(1.0) }}|{{ 1.0 in ifm }}]`, `[one|one|true]`},
		{"yam_none", `[{{ yam[none] }}|{{ yam.get(none) }}|{{ none in yam }}]`, `[N|N|true]`},
		{"yam_int", `[{{ yam[1] }}|{{ yam.get(1) }}|{{ 1 in yam }}]`, `[one|one|true]`},
		{"ptr_key", `{{ ms[p] }}|{{ p in ms }}`, `one|true`},
		{"ptr_get", `[{{ ms[p] }}|{{ ms.get(p) }}|{{ p in ms }}]`, `[one|one|true]`},
		{"bytes_key", `{{ ms[bs] }}`, `one`},
		{"mixed_num", `{{ ifm[1.0] }}|{{ ifm.get(1.0) }}|{{ 1.0 in ifm }}`, `one|one|true`},
		{"mixed_miss", `[{{ ifm[2] }}|{{ ifm.get(2) }}|{{ 2 in ifm }}]`, `[||false]`},
		{"ptr_dict_literal", `{{ {p: 1}|list }}`, `['1']`},
		{"bool_int_key", `{{ ifm1[true] }}|{{ ift[1] }}`, `one|T`},
	}
	env := miya.NewEnvironment()
	for _, tc := range rows {
		t.Run(tc.name, func(t *testing.T) {
			tmpl, err := env.FromString(tc.tpl)
			if err != nil {
				t.Fatal(err)
			}
			got, err := tmpl.Render(miya.NewContextFrom(data))
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestMapLookupCost(t *testing.T) {
	bigi := map[interface{}]interface{}{"s": "S"}
	bigt := map[int]int{}
	for i := 0; i < 20000; i++ {
		bigi[i] = i
		bigt[i] = i
	}
	data := map[string]interface{}{"bigi": bigi, "bigt": bigt}
	ops := []struct{ name, expr string }{
		{"miss", `bigX[-1]`},
		{"in", `-1 in bigX`},
		{"get_miss", `bigX.get(-1)`},
		{"get_hit", `bigX.get(5)`},
	}
	env := miya.NewEnvironment()
	for _, m := range []string{"bigi", "bigt"} {
		for _, op := range ops {
			t.Run(m+"_"+op.name, func(t *testing.T) {
				src := "{% for i in range(2000) %}{{ " + strings.ReplaceAll(op.expr, "bigX", m) + " }}{% endfor %}"
				tmpl, err := env.FromString(src)
				if err != nil {
					t.Fatal(err)
				}
				done := make(chan error, 1)
				start := time.Now()
				go func() {
					_, err := tmpl.Render(miya.NewContextFrom(data))
					done <- err
				}()
				select {
				case err := <-done:
					if err != nil {
						t.Fatal(err)
					}
					t.Logf("took %v", time.Since(start))
				case <-time.After(300 * time.Millisecond):
					t.Fatal("2000 lookups took over 300ms: the lookup scans the map")
				}
			})
		}
	}
}
