package miya_test

import (
	"testing"

	miya "github.com/zipreport/miya"
)

func TestMapKeyRule(t *testing.T) {
	one := "1"
	data := map[string]interface{}{
		"ms":  map[string]string{"None": "N", "": "E", "1": "one"},
		"ifm": map[interface{}]interface{}{nil: "nil", 1: "one", "a": "A"},
		"yam": map[interface{}]interface{}{"None": "N", "1": "one", "true": "T"},
		"d":   map[string]interface{}{"None": "N", "": "E"},
		"p":   &one,
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
