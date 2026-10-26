package miya_test

import (
	"fmt"
	"strings"
	"testing"

	miya "github.com/zipreport/miya"
	"github.com/zipreport/miya/runtime"
)

func TestMacroListArgs(t *testing.T) {
	const addDef = `{% macro add(l) %}{% do l.append(0) %}{% endmacro %}`
	cases := []struct {
		name, tpl, want string
	}{
		// fixed: a macro changes a list it is given
		{"macro_arg", `{% macro m(l) %}{% do l.append(3) %}{% endmacro %}{% set a=[1] %}{{ m(a) }}{{ a }}`, `[1, 3]`},
		{"macro_kwarg", `{% macro m(l) %}{% do l.append(3) %}{% endmacro %}{% set a=[1] %}{{ m(l=a) }}{{ a }}`, `[1, 3]`},
		{"macro_kwarg_literal", `{% macro m(l) %}{% do l.append(3) %}{{ l }}{% endmacro %}{{ m(l=[1]) }}`, `[1, 3]`},
		{"macro_ns_arg", `{% macro m(l) %}{% do l.append(3) %}{% endmacro %}{% set ns=namespace(l=[1]) %}{{ m(ns.l) }}{{ ns.l }}`, `[1, 3]`},
		{"macro_default_local", `{% macro m(l=[]) %}{% do l.append(3) %}{{ l }}{% endmacro %}{{ m() }}{{ m() }}`, `[3][3]`},
		{"macro_literal_arg", `{% macro m(l) %}{% do l.append(3) %}{{ l }}{% endmacro %}{{ m([1]) }}`, `[1, 3]`},
		{"caller_arg", `{% macro m() %}{{ caller([1]) }}{% endmacro %}{% call(l) m() %}{% do l.append(2) %}{{ l }}{% endcall %}`, `[1, 2]`},
		{"default_alias", `{% macro m(a, b=a) %}{% do b.append(1) %}{% endmacro %}{% set l=[] %}{{ m(l) }}{{ l }}`, `[1]`},
		{"pure_chain_once", `{% macro m(x) %}{% endmacro %}{% set l=[1,2,3] %}{{ m(l.pop().a.b) }}{{ l }}`, `[1, 2]`},

		// kept differences, pinned
		{"deviation_nested_after", `{% set a=[] %}{% set inner=[] %}{% do a.append(inner) %}{% do inner.append(1) %}{{ a }}`, `[[]]`},
		{"deviation_nested_literal", `{% set inner=[] %}{% set a=[inner] %}{% do inner.append(1) %}{{ a }}`, `[[]]`},
		{"deviation_set_from_index", `{% set a=[[1]] %}{% set x=a[0] %}{% do x.append(2) %}{{ a }}`, `[[1]]`},
		{"deviation_self_append", `{% set l = [1] %}{% do l.append(l) %}{{ l }}`, `[1, [1]]`},
		{"deviation_data_arg", `{% macro m(l) %}{% do l.append(1) %}{{ l }}{% endmacro %}{{ m(items) }}{{ items }}`, `[0, 1][0]`},
		{"deviation_loop_snapshot", `{% set l = [1] %}{% for x in l %}{% do l.append(x) %}{% endfor %}{{ l }}`, `[1, 1]`},
		{"deviation_loop_var_arg", addDef + `{% set rows=[[1],[2]] %}{% for r in rows %}{{ add(r) }}{% endfor %}{{ rows }}`, `[[1], [2]]`},
		{"deviation_index_arg", addDef + `{% set a=[[1]] %}{{ add(a[0]) }}{{ a }}`, `[[1]]`},
		{"deviation_varargs_arg", addDef + `{% macro outer() %}{{ add(varargs[0]) }}{{ varargs[0] }}{% endmacro %}{% set l=[] %}{{ outer(l) }}{{ l }}`, `[][]`},
		{"deviation_namespace_arg", `{% set a=[] %}{% set ns=namespace(l=a) %}{% do ns.l.append(1) %}{{ a }}`, `[]`},

		// guards
		{"set_alias", `{% set a=[1] %}{% set b=a %}{% do b.append(2) %}{{ a }}`, `[1, 2]`},
		{"ns_list", `{% set ns=namespace(l=[]) %}{% do ns.l.append(1) %}{% set m=ns.l %}{% do m.append(2) %}{{ ns.l }}`, `[1, 2]`},
		{"data_copy", `{% set c = items|list %}{% do c.append(1) %}{{ c }}{{ items }}`, `[0, 1][0]`},
		{"varargs_list", `{% macro m() %}{{ varargs[0]|length }}{% endmacro %}{% set l=[1,2] %}{{ m(l) }}`, `2`},
		{"pop_in_literal", `{% set l=[1,2,3] %}{% set a=[l.pop()] %}{{ a }}{{ l }}`, `[3][1, 2]`},
		{"macro_arg_side_effect_once", `{% macro m(l) %}{{ l|length }}{% endmacro %}{% set l=[1,2,3] %}{{ m([l.pop()]) }}{{ l }}`, `1[1, 2]`},
	}
	errCases := []struct{ name, tpl string }{
		{"deviation_index_append", `{% set a=[[1]] %}{% do a[0].append(2) %}`},
		{"deviation_loop_var_append", `{% set rows=[[1],[2]] %}{% for r in rows %}{% do r.append(0) %}{% endfor %}`},
		{"deviation_data_list", `{% do items.append(1) %}`},
	}
	const notStored = "can only change a list stored with"
	render := func(t *testing.T, env *miya.Environment, tpl string, data map[string]interface{}) (string, error) {
		return renderGo(t, env, tpl, data)
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			env := miya.NewEnvironment(miya.WithAutoEscape(false))
			got, err := render(t, env, tc.tpl, pvD{"items": []interface{}{0}})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
	for _, tc := range errCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			env := miya.NewEnvironment(miya.WithAutoEscape(false))
			got, err := render(t, env, tc.tpl, pvD{"items": []interface{}{0}})
			if err == nil || !strings.Contains(err.Error(), notStored) {
				t.Fatalf("want error containing %q, got err=%v out=%q", notStored, err, got)
			}
		})
	}

	t.Run("go_context_unboxed", func(t *testing.T) {
		env := miya.NewEnvironment(miya.WithAutoEscape(false))
		env.AddGlobal("gt", func(ctx runtime.Context, args ...interface{}) (interface{}, error) {
			v, _ := ctx.GetVariable("l")
			return fmt.Sprintf("%T", v), nil
		})
		got, err := renderGo(t, env, `{% set l = [1] %}{{ gt() }}`, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "[]interface {}" {
			t.Fatalf("got %q, want []interface {}", got)
		}
	})

	t.Run("namespace_all_unboxed", func(t *testing.T) {
		env := miya.NewEnvironment(miya.WithAutoEscape(false))
		env.AddGlobal("dump", func(args ...interface{}) (interface{}, error) {
			ns, ok := args[0].(*miya.Namespace)
			if !ok {
				return nil, fmt.Errorf("got %T, want *miya.Namespace", args[0])
			}
			g, _ := ns.Get("l")
			return fmt.Sprintf("%T|%T", ns.All()["l"], g), nil
		})
		got, err := renderGo(t, env, `{% set ns = namespace(l=[1]) %}{% do ns.l.append(2) %}{{ dump(ns) }}`, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "[]interface {}|[]interface {}" {
			t.Fatalf("got %q", got)
		}
	})

	t.Run("nested_calls_linear", func(t *testing.T) {
		env := miya.NewEnvironment(miya.WithAutoEscape(false))
		n := 0
		env.AddGlobal("count", func(args ...interface{}) (interface{}, error) {
			n++
			return 1, nil
		})
		_, err := renderGo(t, env, `{% macro m(x) %}{% endmacro %}{{ m(m(m(m(m(count().a.b).a.b).a.b).a.b).a.b) }}`, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if n != 1 {
			t.Fatalf("count() ran %d times, want 1", n)
		}
	})

	t.Run("strict_missing_arg", func(t *testing.T) {
		env := miya.NewEnvironment(miya.WithAutoEscape(false), miya.WithStrictUndefined(true))
		got, err := renderGo(t, env, `{% macro m(l) %}{% endmacro %}{{ m(nope) }}`, nil)
		if err == nil || !strings.Contains(err.Error(), "UndefinedError") {
			t.Fatalf("want UndefinedError, got err=%v out=%q", err, got)
		}
	})

	t.Run("data_arg_capacity", func(t *testing.T) {
		items := make([]interface{}, 1, 4)
		items[0] = 1
		env := miya.NewEnvironment(miya.WithAutoEscape(false))
		_, err := renderGo(t, env, `{% macro m(l) %}{% do l.append(9) %}{% endmacro %}{{ m(items) }}`, pvD{"items": items})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if items[:2][1] != nil {
			t.Fatalf("render data slice was written: %v", items[:2])
		}
	})
}
