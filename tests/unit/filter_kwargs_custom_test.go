package miya_test

import (
	"fmt"
	"strings"
	"testing"

	miya "github.com/zipreport/miya"
	"github.com/zipreport/miya/filters"
)

func TestFilterKeywordArgsCustomFilter(t *testing.T) {
	tag := func(v interface{}, args ...interface{}) (interface{}, error) {
		return fmt.Sprintf("%v:%v", v, args), nil
	}
	render := func(t *testing.T, env *miya.Environment, tpl string) (string, error) {
		t.Helper()
		return env.RenderString(tpl, miya.NewContext())
	}

	t.Run("undeclared", func(t *testing.T) {
		env := miya.NewEnvironment(miya.WithAutoEscape(false))
		if err := env.AddFilter("tag", tag); err != nil {
			t.Fatal(err)
		}
		_, err := render(t, env, `{{ 'x'|tag(a=1) }}`)
		if err == nil || !strings.Contains(err.Error(), "does not accept keyword arguments") || !strings.Contains(err.Error(), "SetFilterParams") {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("set_filter_params", func(t *testing.T) {
		env := miya.NewEnvironment(miya.WithAutoEscape(false))
		if err := env.AddFilter("tag", tag); err != nil {
			t.Fatal(err)
		}
		if err := env.SetFilterParams("tag", filters.Param{Name: "a"}, filters.Param{Name: "b", Default: "B"}); err != nil {
			t.Fatal(err)
		}
		for tpl, want := range map[string]string{
			`{{ 'x'|tag(b=2) }}`: "x:[<nil> 2]",
			`{{ 'x'|tag(1) }}`:   "x:[1]",
		} {
			got, err := render(t, env, tpl)
			if err != nil || got != want {
				t.Errorf("%s: got %q, %v; want %q", tpl, got, err, want)
			}
		}
		if err := env.SetFilterParams("nosuch", filters.Param{Name: "a"}); err == nil {
			t.Error("SetFilterParams on an unknown filter: expected an error")
		}
	})

	t.Run("add_filter_with_params", func(t *testing.T) {
		env := miya.NewEnvironment(miya.WithAutoEscape(false))
		err := env.AddFilterWithParams("img", tag, filters.Param{Name: "data"}, filters.Param{Name: "alt"}, filters.Param{Name: "width"})
		if err != nil {
			t.Fatal(err)
		}
		for tpl, want := range map[string]string{
			`{{ 'x'|img(width=400) }}`:      "x:[<nil> <nil> 400]",
			`{{ 'x'|img('d', width=400) }}`: "x:[d <nil> 400]",
			`{{ 'x'|img('d', 'a', 400) }}`:  "x:[d a 400]",
		} {
			got, err := render(t, env, tpl)
			if err != nil || got != want {
				t.Errorf("%s: got %q, %v; want %q", tpl, got, err, want)
			}
		}
	})
}
