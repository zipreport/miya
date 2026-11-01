package miya_test

import (
	"runtime/debug"
	"strings"
	"testing"

	miya "github.com/zipreport/miya"
)

type cycHolder struct{ L []interface{} }

type cycHidden struct {
	n int
	l []interface{}
}

// Self-containing lists in render data must not overflow the stack.
func TestCyclicRenderData(t *testing.T) {
	defer debug.SetMaxStack(debug.SetMaxStack(64 << 20))

	selfList := func() []interface{} {
		l := []interface{}{1, nil}
		l[1] = l
		return l
	}
	cyc, cyc2 := selfList(), selfList()
	cycMap := map[string]interface{}{"x": 1}
	cycMap["self"] = cycMap
	data := map[string]interface{}{
		"cyc": cyc, "cyc2": cyc2,
		"d":       map[string]interface{}{"a": 1},
		"holder":  cycHolder{L: cyc},
		"hidden":  &cycHidden{n: 7, l: cyc},
		"plain":   cycHolder{L: []interface{}{1, "x"}},
		"tm":      map[string]int{"a": 1},
		"cycmap":  cycMap,
		"pholder": &cycHolder{L: cyc},
	}
	rows := []struct{ name, tpl, want string }{
		{"print", `{{ cyc }}`, `[1, [...]]`},
		{"getitem", `{{ d[cyc] }}`, ``},
		{"in_dict", `{{ "y" if cyc in d else "n" }}`, `n`},
		{"dict_literal_key", `{{ {cyc: 1} }}`, `{'[1 [...]]': 1}`},
		{"equal", `{{ "y" if cyc == cyc2 else "n" }}`, `y`},
		{"not_equal", `{{ "y" if cyc == [1, [2]] else "n" }}`, `n`},
		{"in_list", `{{ "y" if cyc2 in [cyc] else "n" }}`, `y`},
		{"in_string", `{{ "y" if cyc in "[1 [...]]" else "n" }}`, `y`},
		{"typed_map_getitem", `{{ tm[cyc] }}`, ``},
		{"struct", `{{ holder }}`, `{[1 [...]]}`},
		{"struct_in_list", `{{ [holder] }}`, `[{[1 [...]]}]`},
		{"unexported_ptr", `{{ hidden }}`, `&{7 [1 [...]]}`},
		{"plain_struct_unchanged", `{{ plain }}`, `{[1 x]}`},
		{"is_lower", `{{ cyc is lower }}|{{ cyc is upper }}|{{ cycmap is lower }}`, `false|false|true`},
		{"struct_ptr_in_string", `{{ pholder is in "&{[1 [...]]}" }}|{{ plain is in "{[1 x]}" }}`, `true|true`},
		{"undefined_getitem", `[{{ missing[cyc] }}][{{ missing[cycmap] }}]`, `[][]`},
	}
	env := miya.NewEnvironment(miya.WithAutoEscape(false))
	for _, tc := range rows {
		t.Run(tc.name, func(t *testing.T) {
			out, err := env.RenderString(tc.tpl, miya.NewContextFrom(data))
			if err != nil {
				t.Fatal(err)
			}
			if out != tc.want {
				t.Errorf("got %q, want %q", out, tc.want)
			}
		})
	}

	t.Run("recursive_loop", func(t *testing.T) {
		// errors (a list item is not iterable); it must not crash
		_, _ = env.RenderString(`{% for x in [cyc] recursive %}{{ loop(x) }}{% endfor %}`, miya.NewContextFrom(data))
	})

	t.Run("struct_pointer_cycle_unchanged", func(t *testing.T) {
		type node struct {
			N    int
			Next *node
		}
		n := &node{N: 1}
		n.Next = n
		out, err := env.RenderString(`{{ n }}`, miya.NewContextFrom(map[string]interface{}{"n": n}))
		if err != nil || !strings.HasPrefix(out, "&{1 0x") {
			t.Fatalf("got %q, %v", out, err)
		}
	})
}
