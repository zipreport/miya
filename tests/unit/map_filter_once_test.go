package miya_test

import (
	"testing"

	miya "github.com/zipreport/miya"
)

func TestMapFilterOnce(t *testing.T) {
	rows := []struct {
		name, tpl string
		calls     int
	}{
		{"numbers", `{{ [1,2]|map('cnt')|join }}`, 2},
		{"lists", `{{ [[1],[2]]|map('cnt')|list }}`, 2},
		{"strings", `{{ ['a','b']|map('cnt')|join }}`, 2},
		{"safe_mix", `{% set s = '<b>'|safe %}{{ [s, 'b']|map('cnt')|join }}`, 2},
	}
	for _, tc := range rows {
		t.Run(tc.name, func(t *testing.T) {
			env := miya.NewEnvironment(miya.WithAutoEscape(true))
			n := 0
			env.AddFilter("cnt", func(v interface{}, args ...interface{}) (interface{}, error) {
				n++
				return v, nil
			})
			tmpl, err := env.FromString(tc.tpl)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := tmpl.Render(miya.NewContext()); err != nil {
				t.Fatal(err)
			}
			if n != tc.calls {
				t.Errorf("filter ran %d times, want %d", n, tc.calls)
			}
		})
	}

	t.Run("digit_filter_registered", func(t *testing.T) {
		env := miya.NewEnvironment(miya.WithAutoEscape(true))
		n := 0
		env.AddFilter("0", func(v interface{}, args ...interface{}) (interface{}, error) {
			n++
			return v, nil
		})
		tmpl, err := env.FromString(`{{ [[v,2]]|map('0')|join }}`)
		if err != nil {
			t.Fatal(err)
		}
		got, err := tmpl.Render(miya.NewContextFrom(map[string]interface{}{"v": "<b>"}))
		if err != nil {
			t.Fatal(err)
		}
		if got != "&lt;b&gt;" || n != 0 {
			t.Errorf("got %q after %d filter calls, want %q after 0", got, n, "&lt;b&gt;")
		}
	})
}
