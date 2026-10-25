package miya_test

import (
	"strings"
	"testing"

	miya "github.com/zipreport/miya"
	"github.com/zipreport/miya/loader"
)

type macroCase struct {
	name      string
	templates map[string]string // the main template is "i.html"
	ctx       map[string]interface{}
	want      string // exact output; ignored when wantErr is set
	wantErr   string // substring the render error must contain
}

func runMacroCases(t *testing.T, strict bool, cases []macroCase) {
	t.Helper()
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			sl := loader.NewStringLoader(loader.NewDirectTemplateParser())
			for name, src := range c.templates {
				sl.AddTemplate(name, src)
			}
			env := miya.NewEnvironment(miya.WithLoader(sl), miya.WithAutoEscape(false), miya.WithStrictUndefined(strict))
			got, err := func() (string, error) {
				tmpl, err := env.GetTemplate("i.html")
				if err != nil {
					return "", err
				}
				return tmpl.Render(miya.NewContextFrom(c.ctx))
			}()
			if c.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), c.wantErr) {
					t.Fatalf("got %q, %v; want error containing %q", got, err, c.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("render: %v", err)
			}
			if got != c.want {
				t.Errorf("got %q, want %q", got, c.want)
			}
		})
	}
}

func mainOnly(src string) map[string]string { return map[string]string{"i.html": src} }

const mAB = `{% macro m(a, b=2) %}{{ a }}-{{ b }}{% endmacro %}`

func TestMacroArgumentBinding(t *testing.T) {
	runMacroCases(t, false, []macroCase{
		{name: "kw_only", templates: mainOnly(mAB + `{{ m(a=1) }}`), want: `1-2`},
		{name: "positional_and_kw", templates: mainOnly(mAB + `{{ m(1, b=9) }}`), want: `1-9`},
		{name: "kw_reordered", templates: mainOnly(mAB + `{{ m(b=5, a=1) }}`), want: `1-5`},
		{name: "default_sees_earlier", templates: mainOnly(`{% macro m(a, b=a~'x') %}{{ b }}{% endmacro %}{{ m(1) }}`), want: `1x`},
		{name: "default_chain", templates: mainOnly(`{% macro m(a, b=2, c=b+1) %}{{ c }}{% endmacro %}{{ m(1) }}`), want: `3`},
		{name: "default_late_set", templates: mainOnly(`{% macro m(a=z) %}{{ a }}{% endmacro %}{% set z = 'Z2' %}{{ m() }}`), want: `Z2`},
		{name: "missing_param", templates: mainOnly(`{% macro m(a, b) %}[{{ a }}][{{ b }}]{% endmacro %}{{ m(1) }}`), want: `[1][]`},
		{name: "missing_param_defined", templates: mainOnly(`{% macro m(a, b) %}{{ b is defined }}{% endmacro %}{{ m(1) }}`), want: `false`},
		{name: "missing_shadows_outer", templates: mainOnly(`{% set b = 'OUT' %}{% macro m(a, b) %}[{{ b }}]{% endmacro %}{{ m(1) }}`), want: `[]`},
		{name: "surplus_positional", templates: mainOnly(`{% macro m(a) %}{{ a }}{% endmacro %}{{ m(1, 2) }}`), wantErr: `macro "m" takes not more than 1 argument(s)`},
		{name: "duplicate_kw", templates: mainOnly(`{% macro m(a) %}{{ a }}{% endmacro %}{{ m(1, a=2) }}`), wantErr: `macro "m" got multiple values for argument "a"`},
		{name: "unknown_kw", templates: mainOnly(`{% macro m(a) %}{{ a }}{% endmacro %}{{ m(1, z=2) }}`), wantErr: `macro "m" takes no keyword argument "z"`},
		{name: "varargs", templates: mainOnly(`{% macro m() %}{{ varargs|length }}{% endmacro %}{{ m(1, 2) }}`), want: `2`},
		{name: "kwargs", templates: mainOnly(`{% macro m() %}{{ kwargs.x }}{% endmacro %}{{ m(x=1) }}`), want: `1`},
		{name: "no_dynamic_scope", templates: mainOnly(`{% macro m() %}{{ i }}{% endmacro %}{% for i in [1,2] %}{{ m() }}{% endfor %}`), want: ``},
		{name: "sees_later_toplevel_set", templates: mainOnly(`{% macro m() %}{{ v }}{% endmacro %}{% set v = 2 %}{{ m() }}`), ctx: map[string]interface{}{"v": 1}, want: `2`},
		{name: "recursive", templates: mainOnly(`{% macro f(n) %}{% if n > 0 %}{{ n }}{{ f(n-1) }}{% endif %}{% endmacro %}{{ f(3) }}`), want: `321`},
		{name: "print_macro", templates: mainOnly(`{% macro m() %}{% endmacro %}{{ m }}`), want: `<Macro 'm'>`},
		{name: "is_callable", templates: mainOnly(`{% macro m() %}{% endmacro %}{{ m is callable }}`), want: `true`},
		{name: "param_named_kwargs", templates: mainOnly(`{% macro m(kwargs) %}[{{ kwargs }}]{% endmacro %}{{ m(kwargs=3) }}`), want: `[3]`},
		{name: "param_named_kwargs_no_catch", templates: mainOnly(`{% macro m(kwargs) %}[{{ kwargs }}]{% endmacro %}{{ m(1, z=2) }}`), wantErr: `macro "m" takes no keyword argument "z"`},
		{name: "param_named_varargs", templates: mainOnly(`{% macro m(varargs) %}[{{ varargs }}]{% endmacro %}{{ m(1) }}`), want: `[1]`},
		{name: "param_named_varargs_no_catch", templates: mainOnly(`{% macro m(varargs) %}[{{ varargs }}]{% endmacro %}{{ m(1, 2) }}`), wantErr: `macro "m" takes not more than 1 argument(s)`},
	})
}

func TestMacroArgumentBindingStrict(t *testing.T) {
	runMacroCases(t, true, []macroCase{
		{name: "missing_param_used", templates: mainOnly(`{% macro m(a, b) %}[{{ b ~ 'x' }}]{% endmacro %}{{ m(1) }}`), wantErr: `undefined variable: b`},
		{name: "missing_param_defined", templates: mainOnly(`{% macro m(a, b) %}{{ b is defined }}{% endmacro %}{{ m(1) }}`), want: `false`},
	})
}

func TestImportedMacroCalls(t *testing.T) {
	hi := `{% macro hi(n, g='Hi') %}{{ g }} {{ n }}{% endmacro %}`
	wrap := `{% macro wrap() %}<{{ caller() }}>{% endmacro %}`
	runMacroCases(t, false, []macroCase{
		{name: "self_recursive", templates: map[string]string{"m.html": `{% macro deep(n) %}{% if n > 0 %}{{ deep(n-1) }}{% else %}end{% endif %}{% endmacro %}`, "i.html": `{% import "m.html" as m %}{{ m.deep(3) }}`}, want: `end`},
		{name: "sibling", templates: map[string]string{"m.html": `{% macro a() %}A{{ b() }}{% endmacro %}{% macro b() %}B{% endmacro %}`, "i.html": `{% import "m.html" as m %}{{ m.a() }}`}, want: `AB`},
		{name: "import_call_block", templates: map[string]string{"w.html": wrap, "i.html": `{% import "w.html" as w %}{% call w.wrap() %}in{% endcall %}`}, want: `<in>`},
		{name: "from_call_block", templates: map[string]string{"w.html": wrap, "i.html": `{% from "w.html" import wrap %}{% call wrap() %}in{% endcall %}`}, want: `<in>`},
		{name: "import_kwargs", templates: map[string]string{"m.html": hi, "i.html": `{% import "m.html" as m %}{{ m.hi(g='G', n='N') }}`}, want: `G N`},
		{name: "from_kwargs", templates: map[string]string{"m.html": hi, "i.html": `{% from "m.html" import hi %}{{ hi('N', g='G') }}`}, want: `G N`},
		{name: "print_imported", templates: map[string]string{"m.html": `{% macro s() %}{% endmacro %}`, "i.html": `{% import "m.html" as m %}{{ m.s }}`}, want: `<Macro 's'>`},
	})
}

func TestCallBlockCallers(t *testing.T) {
	runMacroCases(t, false, []macroCase{
		{name: "caller_args", templates: mainOnly(`{% macro m() %}{{ caller(1) }}{% endmacro %}{% call(x) m() %}[{{ x }}]{% endcall %}`), want: `[1]`},
		{name: "caller_kwargs", templates: mainOnly(`{% macro m() %}{{ caller(y=2) }}{% endmacro %}{% call(x=1, y=0) m() %}[{{ x }}{{ y }}]{% endcall %}`), want: `[12]`},
		{name: "caller_default", templates: mainOnly(`{% macro m() %}{{ caller() }}{% endmacro %}{% call(x='d') m() %}[{{ x }}]{% endcall %}`), want: `[d]`},
		{name: "caller_per_call", templates: mainOnly(`{% macro m() %}{% for i in [1,2] %}{{ caller(i) }}{% endfor %}{% endmacro %}{% call(i) m() %}<{{ i }}>{% endcall %}`), want: `<1><2>`},
		{name: "caller_sees_call_site", templates: mainOnly(`{% macro m() %}{{ caller() }}{% endmacro %}{% for i in [1,2] %}{% call m() %}{{ i }}{% endcall %}{% endfor %}`), want: `12`},
		{name: "caller_not_macro_scope", templates: mainOnly(`{% macro m(q) %}{{ caller() }}{% endmacro %}{% call m(5) %}[{{ q }}]{% endcall %}`), want: `[]`},
		{name: "caller_not_inherited", templates: mainOnly(`{% macro a() %}{{ caller is defined }}{% endmacro %}{% macro w() %}{{ a() }}{{ caller() }}{% endmacro %}{% call w() %}x{% endcall %}`), want: `falsex`},
		{name: "nested_macro_own_caller", templates: mainOnly(`{% macro w() %}{% macro a() %}{{ caller is defined }}{% endmacro %}{{ a() }}{{ caller() }}{% endmacro %}{% call w() %}x{% endcall %}`), want: `falsex`},
		{name: "nested_call", templates: mainOnly(`{% macro inner() %}[{{ caller() }}]{% endmacro %}{% macro outer() %}{% call inner() %}x{% endcall %}{% endmacro %}{{ outer() }}`), want: `[x]`},
		{name: "caller_unused", templates: mainOnly(`{% macro m() %}M{% endmacro %}{% call m() %}x{% endcall %}`), wantErr: `macro "m" was called with a caller but does not use caller()`},
		{name: "caller_param_default", templates: mainOnly(`{% macro m(caller='d') %}[{{ caller }}]{% endmacro %}{{ m() }}`), want: `[d]`},
		{name: "caller_param_bound_by_call", templates: mainOnly(`{% macro m(caller='d') %}[{{ caller() }}]{% endmacro %}{% call m() %}X{% endcall %}`), want: `[X]`},
		{name: "caller_param_needs_default", templates: mainOnly(`{% macro m(caller) %}{% endmacro %}`), wantErr: `special "caller" argument must be omitted or be given a default`},
		{name: "call_requires_call", templates: mainOnly(`{% macro w() %}{{ caller() }}{% endmacro %}{% call w()|upper %}x{% endcall %}`), wantErr: `expected call`},
		{name: "nested_caller_undefined", templates: mainOnly(`{% macro w() %}<{{ caller() }}>{% endmacro %}{% macro o() %}{% call w() %}{{ caller() }}{% endcall %}{% endmacro %}{% call o() %}X{% endcall %}`), wantErr: `no caller defined`},
	})
}

// TestMacroAutoescapeBlock runs with autoescape on: a {% autoescape false %}
// block must still apply inside macro and caller() bodies defined in it.
func TestMacroAutoescapeBlock(t *testing.T) {
	cases := []struct{ name, src, want string }{
		{"caller_in_block", `{% autoescape false %}{% macro w() %}{{ caller() }}{% endmacro %}{% call w() %}<p>{{ v }}{% endcall %}{% endautoescape %}`, `<p><b>`},
		{"macro_in_block", `{% autoescape false %}{% macro w(x) %}{{ x }}{% endmacro %}{{ w(v) }}{% endautoescape %}`, `<b>`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			env := miya.NewEnvironment(miya.WithAutoEscape(true))
			got, err := env.RenderString(c.src, miya.NewContextFrom(map[string]interface{}{"v": "<b>"}))
			if err != nil {
				t.Fatalf("render: %v", err)
			}
			if got != c.want {
				t.Errorf("got %q, want %q", got, c.want)
			}
		})
	}
}

// TestMacroRecursionError checks that runaway recursion reports the depth
// limit once instead of wrapping it at every macro level.
func TestMacroRecursionError(t *testing.T) {
	env := miya.NewEnvironment()
	_, err := env.RenderString(`{% macro f(n) %}{{ f(n+1) }}{% endmacro %}{{ f(0) }}`, miya.NewContext())
	if err == nil || !strings.Contains(err.Error(), "maximum template recursion depth exceeded") {
		t.Fatalf("want a recursion-depth error, got %v", err)
	}
	if n := strings.Count(err.Error(), "error executing macro"); n > 1 {
		t.Errorf("error wraps %d macro levels, want at most 1: %.200s", n, err.Error())
	}
}
