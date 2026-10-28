package miya_test

import (
	"fmt"
	"testing"

	miya "github.com/zipreport/miya"
	"github.com/zipreport/miya/runtime"
)

// scanData returns n dicts, each holding a nested list and dict.
func scanData(n int) []interface{} {
	big := make([]interface{}, n)
	for i := range big {
		big[i] = map[string]interface{}{"a": i, "b": []interface{}{"x", "y", map[string]interface{}{"c": "d"}}}
	}
	return big
}

// Escaped output nested in lists and dicts is turned into plain strings only
// for Go code registered by the user: a built-in filter or test call on a big
// value allocates no more than one on a small value, also in a render that
// made escaped output, and an HTML render pays nothing for a user filter
// after another render made escaped output.
func TestEscapedOutputNoScanForBuiltins(t *testing.T) {
	env := miya.NewEnvironment(miya.WithAutoEscape(true))
	_ = env.AddFilter("ufilt", func(v interface{}, _ ...interface{}) (interface{}, error) { return "u", nil })
	parse := func(src string) *miya.Template {
		tmpl, err := env.FromString(src)
		if err != nil {
			t.Fatal(err)
		}
		return tmpl
	}
	// a Go namespace keeps escaped output made by a JS render
	gns := miya.NewNamespace()
	data := func(big []interface{}) map[string]interface{} {
		return map[string]interface{}{"big": big, "gns": gns}
	}
	jsTmpl := parse(`{% macro t() %}x{% endmacro %}{% set gns.v = [t()] %}`)
	builtins := parse(`{% set b = big + gns.v %}{{ b|length }}{{ b|first|tojson|length }}{{ b is sequence }}{{ b|last|length }}{{ gns.v|first is string }}`)
	ev := runtime.NewEvaluator()
	eval := func(tmpl *miya.Template, big []interface{}, ec runtime.EscapeContext) {
		ctx := miya.NewTemplateContextAdapter(miya.NewContextFrom(data(big)), env)
		if _, err := ev.EvalNode(tmpl.AST(), runtime.NewContextWrapper(ctx, runtime.NewAutoEscaper(nil), ec)); err != nil {
			t.Fatal(err)
		}
	}
	eval(jsTmpl, nil, runtime.EscapeContextJS)

	small, large := scanData(10), scanData(10000)
	check := func(name string, run func(big []interface{})) {
		t.Helper()
		a := testing.AllocsPerRun(5, func() { run(small) })
		b := testing.AllocsPerRun(5, func() { run(large) })
		if b > a+5 {
			t.Errorf("%s: %v allocations for 10 items, %v for 10000", name, a, b)
		}
	}
	for _, ec := range []runtime.EscapeContext{runtime.EscapeContextJS, runtime.EscapeContextHTML} {
		check(fmt.Sprintf("builtins_%s", ec), func(big []interface{}) { eval(builtins, big, ec) })
	}
	user := parse(`{{ (big + gns.v)|ufilt }}`)
	check("html_user_filter", func(big []interface{}) {
		if _, err := user.Render(miya.NewContextFrom(data(big))); err != nil {
			t.Fatal(err)
		}
	})
}

// renderIn renders src with data in escape context ec.
func renderIn(t *testing.T, env *miya.Environment, ec runtime.EscapeContext, src string, data map[string]interface{}) string {
	t.Helper()
	tmpl, err := env.FromString(src)
	if err != nil {
		t.Fatal(err)
	}
	ctx := miya.NewTemplateContextAdapter(miya.NewContextFrom(data), env)
	out, err := runtime.NewEvaluator().EvalNode(tmpl.AST(), runtime.NewContextWrapper(ctx, runtime.NewAutoEscaper(nil), ec))
	if err != nil {
		t.Fatal(err)
	}
	return runtime.ToString(out)
}

// A safe operand of ~ or + with output escaped for a non-HTML context is
// HTML-safe only: it is escaped for the context like plain text.
func TestConcatEscapedSafeOperand(t *testing.T) {
	env := miya.NewEnvironment(miya.WithAutoEscape(true))
	data := map[string]interface{}{"m": runtime.Markup(`ok<b>'"\`), "p": `ok<b>'"\`, "a": `9<'"&\9`}
	const macro = `{% macro t() %}T{% endmacro %}`
	pairs := [][2]string{
		{`{{ t() ~ m }}`, `{{ t() ~ p }}`},
		{`{{ m ~ t() }}`, `{{ p ~ t() }}`},
		{`{{ t() + m }}`, `{{ t() + p }}`},
		{`{{ m + t() }}`, `{{ p + t() }}`},
		{`{{ t() ~ (a|safe) }}`, `{{ t() ~ a }}`},
		{`{{ (a|safe) ~ t() }}`, `{{ a ~ t() }}`},
	}
	for _, ec := range []runtime.EscapeContext{runtime.EscapeContextJS, runtime.EscapeContextCSS, runtime.EscapeContextURL, runtime.EscapeContextJSON} {
		for _, p := range pairs {
			got := renderIn(t, env, ec, macro+p[0], data)
			want := renderIn(t, env, ec, macro+p[1], data)
			if got != want {
				t.Errorf("%s %s: got %q, want %q", ec, p[0], got, want)
			}
		}
	}
}

// indent escapes its indent string for a safe value even with autoescape
// off, as center escapes its fill (Jinja2 leaves it raw).
func TestIndentStringEscapedForSafeValueAutoescapeOff(t *testing.T) {
	env := miya.NewEnvironment(miya.WithAutoEscape(true))
	data := func() map[string]interface{} {
		return map[string]interface{}{"m": runtime.Markup("ok TRUST\nok TRUST"), "a": `9<'"&\9`, "ns": miya.NewNamespace()}
	}
	for _, expr := range []string{`m|indent(a)`, `m|indent(width=a)`, `[m]|map('indent', a)|first`} {
		on := `{% set ns.v = ` + expr + ` %}{{ ns.v }}`
		off := `{% autoescape false %}{% set ns.v = ` + expr + ` %}{% endautoescape %}{{ ns.v }}`
		want := renderIn(t, env, runtime.EscapeContextHTML, on, data())
		if got := renderIn(t, env, runtime.EscapeContextHTML, off, data()); got != want {
			t.Errorf("%s: got %q, want %q", expr, got, want)
		}
		tmpl, err := env.FromString(off)
		if err != nil {
			t.Fatal(err)
		}
		if got, err := tmpl.Render(miya.NewContextFrom(data())); err != nil || got != want {
			t.Errorf("%s: Render got %q, %v, want %q", expr, got, err, want)
		}
	}
}

// map gives a filter escaped output as the text it is, as a filter applied to
// it directly gets it.
func TestMapOverEscapedOutput(t *testing.T) {
	env := miya.NewEnvironment(miya.WithAutoEscape(true))
	data := map[string]interface{}{"a": `9<'"&\9`}
	const macros = `{% macro t() %}ok<b>'"\{% endmacro %}{% macro et() %}{% endmacro %}`
	pairs := [][2]string{
		{`{{ [t()]|map('reverse')|first }}`, `{{ t()|reverse }}`},
		{`{{ [et()]|map('default', a)|first }}`, `{{ et()|default(a) }}`},
		{`{{ [t()]|map('length')|first }}`, `{{ t()|length }}`},
		{`{{ [t()]|map('title')|first }}`, `{{ t()|title }}`},
	}
	for _, ec := range []runtime.EscapeContext{runtime.EscapeContextJS, runtime.EscapeContextCSS, runtime.EscapeContextURL, runtime.EscapeContextJSON} {
		for _, p := range pairs {
			got := renderIn(t, env, ec, macros+p[0], data)
			want := renderIn(t, env, ec, macros+p[1], data)
			if got != want {
				t.Errorf("%s %s: got %q, want %q", ec, p[0], got, want)
			}
		}
	}
}
