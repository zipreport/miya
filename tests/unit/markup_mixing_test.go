package miya_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	miya "github.com/zipreport/miya"
	"github.com/zipreport/miya/runtime"
)

type mixSt int

func (mixSt) String() string { return "<img onerror=x>" }

type mixSp int

func (mixSp) String() string { return " <x> " }

type mixFt int

func (mixFt) Format(f fmt.State, _ rune) { _, _ = f.Write([]byte("<img onerror=x>")) }

func mixCtx() miya.Context {
	M := func(s string) runtime.Markup { return runtime.Markup(s) }
	return miya.NewContextFrom(map[string]interface{}{
		"x": "<x>", "q": "it's", "u": "é<",
		"m": M("<b>"), "e": "<s>", "sm": M("<s>"), "card": M("<img src=/a.png alt=photo>"),
		"sep": " onerror=alert(1) ", "sepx": " <x> ", "br": M("<br>"),
		"fs": M("%s"), "fr": M("%r"), "fa": M("%a"), "fc": M("%c"), "fw": M("%5s|"),
		"ff": M("%.2f"), "fx": M("%x"), "fd": M("%d"), "fka": M("%(a)s"), "fkr": M("%(a)r"),
		"fp": M("%.2s"), "fl": M("%-6s|"),
		"n60": 60, "lt": "<", "bs": `\`, "f25": 2.5, "n255": 255, "n3": 3,
		"bx": M("<b>X"), "mb20": M("<b>" + strings.Repeat("x", 20)), "p20": strings.Repeat("x", 20),
		"mb3": M("<b>xxx"), "mab": M("a b"), "pab": "a b", "miab": M("<i>a b"),
		"st": mixSt(1), "sp": mixSp(1), "ft": mixFt(1), "t": strings.Repeat("y", 40),
	})
}

func TestMarkupMixing(t *testing.T) {
	env := miya.NewEnvironment(miya.WithAutoEscape(true))
	rows := []struct{ name, tpl, want string }{
		{"r1", `{{ fs|format(x) }}`, `&lt;x&gt;`},
		{"r2", `{{ fr|format(x) }}`, `&#39;&lt;x&gt;&#39;`},
		{"r3", `{{ fr|format(q) }}`, `&#34;it&#39;s&#34;`},
		{"r4", `{{ fa|format(u) }}`, `&#39;\xe9&lt;&#39;`},
		{"r5", `{{ fc|format(n60) }}`, `ERR:%c`},
		{"r6", `{{ fc|format(lt) }}`, `ERR:%c`},
		{"r7", `{{ fs|format(m) }}`, `<b>`},
		{"r8", `{{ fw|format(lt) }}`, ` &lt;|`},
		{"r9", `{{ ff|format(f25) }}`, `2.50`},
		{"deviation_hex_in_safe_format", `{{ fx|format(n255) }}`, `ff`},
		{"r11", `{{ fd|format(n3) }}`, `3`},
		{"r12", `{{ fka|format(a=x) }}`, `&lt;x&gt;`},
		{"r13", `{{ fkr|format(a=x) }}`, `&#39;&lt;x&gt;&#39;`},
		{"r14", `{{ bx|replace('X', x) }}`, `<b>&lt;x&gt;`},
		{"r15", `{{ mb20|truncate(12, true, e) }}`, `<b>xxxxxx&lt;s&gt;`},
		{"r16", `{{ p20|truncate(12, true, e) }}`, `xxxxxxxxx&lt;s&gt;`},
		{"r17", `{{ p20|truncate(12, true, sm) }}`, `xxxxxxxxx<s>`},
		{"r18", `{{ mb20|truncate(12, true, sm) }}`, `<b>xxxxxx<s>`},
		{"r19", `{{ mb3|truncate(12) }}`, `<b>xxx`},
		{"r20", `{{ card|wordwrap(5, false, sep) }}`, `&lt;img onerror=alert(1) src=/a.png onerror=alert(1) alt=photo&gt;`},
		{"r21", `{{ mab|wordwrap(1, wrapstring=sepx) }}`, `a &lt;x&gt; b`},
		{"r22", `{{ pab|wordwrap(1, wrapstring=br) }}`, `a<br>b`},
		{"deviation_safe_input_safe_wrapstring", `{{ miab|wordwrap(1, wrapstring=br) }}`, `<<br>i<br>><br>a<br>b`},
		{"r24", `{{ mab|wordwrap(10) }}`, `a b`},
		{"s1", `{{ fs|format(st) }}`, `&lt;img onerror=x&gt;`},
		{"s2", `{{ fd|format(st) }}`, `1`},
		{"s3", `{{ bx|replace('X', st) }}`, `<b>&lt;img onerror=x&gt;`},
		{"s4", `{{ [m,'c']|join(st) }}`, `<b>&lt;img onerror=x&gt;c`},
		{"s5", `{{ st }}`, `&lt;img onerror=x&gt;`},
		{"s5_formatter", `{{ bx|replace('X', ft) }}`, `<b>&lt;img onerror=x&gt;`},
		{"s6_wordwrap_stringer", `{{ mab|wordwrap(1, wrapstring=sp) }}`, `a &lt;x&gt; b`},
		{"kw_truncate", `{{ mb20|truncate(12, true, end=e) }}`, `<b>xxxxxx&lt;s&gt;`},
		{"precision_escapes_first", `{{ fp|format(x) }}`, `&l`},
		{"width_after_escape", `{{ fl|format(lt) }}`, `&lt;  |`},
		{"fb11_guard", `{% filter wordwrap(80) %}<p>{{ x }}</p>{% endfilter %}`, `<p>&lt;x&gt;</p>`},
		{"fb_wordwrap_newline", `{% filter wordwrap(72, true, '\n') %}<p>{{ x }}</p>{% endfilter %}`, `<p>&lt;x&gt;</p>`},
		// Jinja2: "&lt;img\nsrc=/a.png\nalt=photo&gt;"
		{"deviation_wordwrap_newline_safe", `{{ card|wordwrap(5, false, '\n') }}`, "<img\nsrc=/a.png\nalt=photo>"},
		{"deviation_fb_wordwrap_br", `{% filter wordwrap(72, true, '<br>') %}<p>{{ x }}</p>{% endfilter %}`, `&lt;p&gt;&amp;lt;x&amp;gt;&lt;/p&gt;`},
		// Jinja2: error: TypeError
		{"deviation_setblock_length", `{% set n %}20{% endset %}{{ t|truncate(n) }}`, strings.Repeat("y", 17) + `...`},
		// Jinja2: error: TypeError
		{"deviation_setblock_width", `{% set n %}5{% endset %}{{ m|wordwrap(n) }}`, `<b>`},
	}
	for _, tc := range rows {
		t.Run(tc.name, func(t *testing.T) {
			tmpl, err := env.FromString(tc.tpl)
			var got string
			if err == nil {
				got, err = tmpl.Render(mixCtx())
			}
			if strings.HasPrefix(tc.want, "ERR:") {
				if err == nil || !strings.Contains(err.Error(), tc.want[4:]) {
					t.Fatalf("want error containing %q, got %q, err=%v", tc.want[4:], got, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

// Outside the HTML family these filters must not HTML-escape: the context
// escaper runs once on the result.
func TestMarkupMixingNonHTML(t *testing.T) {
	rows := []struct{ name, tpl, want string }{
		{"plain", `{{ x }}|{{ q }}`, `\u003cx\u003e|it\'s`},
		{"format", `{{ fs|format(x) }}`, `\u003cx\u003e`},
		{"format_safe_arg", `{{ fs|format(m) }}`, `\u003cb\u003e`},
		{"truncate", `{{ p20|truncate(12, true, sm) }}`, `xxxxxxxxx\u003cs\u003e`},
		{"wordwrap", `{{ card|wordwrap(5, false, sep) }}`, `\u003cimg onerror=alert(1) src=/a.png onerror=alert(1) alt=photo\u003e`},
		{"replace_safe", `{{ ("Hi NAME"|safe)|replace("NAME", bs) }}`, `Hi \\`},
		{"replace_safe_tag", `{{ m|replace("b", x) }}`, `\u003c\u003cx\u003e\u003e`},
		{"method_replace", `{{ ("Hi NAME"|safe).replace("NAME", bs) }}`, `Hi \\`},
		{"method_format", `{{ ("{}"|safe).format(x) }}`, `\u003cx\u003e`},
		{"method_join", `{{ (","|safe).join([x, x]) }}`, `\u003cx\u003e,\u003cx\u003e`},
		{"method_upper", `{{ m.upper() }}`, `\u003cB\u003e`},
		{"method_split", `{{ m.split()[0] }}`, `\u003cb\u003e`},
		// With autoescape off the method result is still plain text here, so
		// the JS escaper runs on it once it is printed with autoescape on.
		{"method_format_autoescape_off", `{% autoescape false %}{% set s = ("{}"|safe).format(bs) %}{% endautoescape %}"{{ s }}"`, `"\\"`},
		{"method_replace_autoescape_off", `{% autoescape false %}{% set s = ("A"|safe).replace("A", bs) %}{% endautoescape %}"{{ s }}"`, `"\\"`},
		{"filter_replace_autoescape_off", `{% autoescape false %}{% set s = ("A"|safe)|replace("A", bs) %}{% endautoescape %}"{{ s }}"`, `"\\"`},
	}
	for _, tc := range rows {
		t.Run(tc.name, func(t *testing.T) {
			env := miya.NewEnvironment(miya.WithAutoEscape(true))
			tmpl, err := env.FromString(tc.tpl)
			if err != nil {
				t.Fatal(err)
			}
			ctx := miya.NewTemplateContextAdapter(mixCtx(), env)
			out, err := runtime.NewEvaluator().EvalNode(tmpl.AST(), runtime.NewContextWrapper(ctx, runtime.NewAutoEscaper(nil), runtime.EscapeContextJS))
			if err != nil {
				t.Fatal(err)
			}
			got := runtime.ToString(out)
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

// With autoescape off in an HTML-family context a filter keeps Markup as
// markupsafe does, except where Jinja2's filter checks autoescape itself;
// the value set there is printed later with autoescape on.
func TestMarkupFiltersAutoescapeOff(t *testing.T) {
	env := miya.NewEnvironment(miya.WithAutoEscape(true))
	const pre = `{% set ns = namespace() %}{% autoescape false %}{% set ns.s = `
	const post = ` %}{% endautoescape %}{{ ns.s }}`
	rows := []struct{ name, expr, want string }{
		{"upper", `m|upper`, `<B>`},
		{"upper_plain", `x|upper`, `&lt;X&gt;`},
		{"center", `m|center(7)`, `  <b>  `},
		{"trim", `m|trim`, `<b>`},
		{"indent", `m|indent(2, true)`, `  <b>`},
		{"indent_arg_escaped", `("a\nb"|safe)|indent("<i>")`, "a\n&lt;i&gt;b"},
		{"trim_arg_not_escaped", `("<b><"|safe)|trim("<")`, `b>`},
		{"reverse", `m|reverse`, `>b<`},
		{"default", `m|default(x)`, `<b>`},
		{"escape_safe", `m|escape`, `<b>`},
		{"escape_plain_once", `x|e`, `&lt;x&gt;`},
		{"format_escapes_arg", `("%s<b>"|safe)|format(x)`, `&lt;x&gt;<b>`},
		{"truncate", `m|truncate(10)`, `<b>`},
		{"map_upper", `(["<b>"|safe]|map("upper")|list)[0]`, `<B>`},
		// default returns its argument as it is, not marked safe
		{"default_arg_not_marked", `""|safe|default(x, true)`, `&lt;x&gt;`},
		{"d_arg_not_marked", `""|safe|d(x, true)`, `&lt;x&gt;`},
		{"map_default_arg_not_marked", `[""|safe]|map("default", x, true)|first`, `&lt;x&gt;`},
		{"default_safe_arg", `""|safe|default(m, true)`, `<b>`},
		// markupsafe's Markup.center escapes the fill whatever autoescape is
		{"center_fill_escaped", `m|center(11, x)`, strings.Repeat(`&lt;x&gt;`, 4) + `<b>` + strings.Repeat(`&lt;x&gt;`, 4)},
		{"map_center_fill_escaped", `[m]|map("center", 5, x)|first`, `&lt;x&gt;<b>&lt;x&gt;`},
		// Jinja2's do_replace, do_title, do_urlize and do_join return plain
		// text when autoescape is off
		{"replace_plain", `m|replace("b", "<i>")`, `&lt;&lt;i&gt;&gt;`},
		{"replace_plain_arg", `m|replace("b", x)`, `&lt;&lt;x&gt;&gt;`},
		{"title_plain", `m|title`, `&lt;B&gt;`},
		{"urlize_plain", `m|urlize`, `&lt;b&gt;`},
		{"join_plain", `[m, x]|join`, `&lt;b&gt;&lt;x&gt;`},
	}
	for _, tc := range rows {
		t.Run(tc.name, func(t *testing.T) {
			tmpl, err := env.FromString(pre + tc.expr + post)
			if err != nil {
				t.Fatal(err)
			}
			got, err := tmpl.Render(mixCtx())
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

// default on a safe input returns its argument as it is, keeping the
// argument's own safeness and type, as Jinja2's do_default does.
func TestMarkupDefaultArgument(t *testing.T) {
	env := miya.NewEnvironment(miya.WithAutoEscape(true))
	rows := []struct{ name, tpl, want string }{
		{"printed", `{{ ""|safe|default(x, true) }}`, `&lt;x&gt;`},
		{"length", `{{ ""|safe|default(x, true)|length }}`, `3`},
		{"list", `{% for u in ""|safe|default(["a", "<b>"], true) %}[{{ u }}]{% endfor %}`, `[a][&lt;b&gt;]`},
		{"safe_arg", `{{ ""|safe|default(m, true) }}`, `<b>`},
		{"input", `{{ m|default(x) }}`, `<b>`},
	}
	for _, tc := range rows {
		t.Run(tc.name, func(t *testing.T) {
			got, err := env.RenderString(tc.tpl, mixCtx())
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

// Template output already escaped for a non-HTML context (macro, caller,
// block set, filter block) is not escaped again when printed there.
func TestEscapedOutputNonHTML(t *testing.T) {
	js, css, url, jsn := runtime.EscapeContextJS, runtime.EscapeContextCSS, runtime.EscapeContextURL, runtime.EscapeContextJSON
	const m2 = `{% macro m2(v) %}{{ v }}{% endmacro %}`
	rows := []struct {
		name string
		ec   runtime.EscapeContext
		tpl  string
		want string
	}{
		{"js_filter_upper", js, `{% filter upper %}<{{ x }}{{ bs }}{% endfilter %}`, `<\u003cX\u003e\\`},
		{"js_filter_trim", js, `{% filter trim %}  {{ bs }}  {% endfilter %}`, `\\`},
		{"js_filter_safe_escaped_once", js, `{% filter safe %}{{ x }}{{ bs }}{% endfilter %}`, `\u003cx\u003e\\`},
		{"js_macro", js, m2 + `"{{ m2(bs) }}"`, `"\\"`},
		{"js_macro_tag", js, m2 + `"{{ m2(x) }}"`, `"\u003cx\u003e"`},
		{"js_caller", js, `{% macro c() %}[{{ caller() }}]{% endmacro %}"{% call c() %}{{ bs }}{% endcall %}"`, `"[\\]"`},
		{"js_block_set", js, `{% set s %}{{ bs }}{% endset %}"{{ s }}"`, `"\\"`},
		{"js_concat", js, `{% set s %}{{ bs }}{% endset %}{{ s ~ x }}`, `\\\u003cx\u003e`},
		{"js_upper_keeps_escapes", js, `{% set s %}{{ x }}{% endset %}{{ s|upper }}|{{ s.upper() }}`, `\u003cX\u003e|\u003cX\u003e`},
		{"js_last_escaped_again", js, `{% set s %}{{ bs }}{% endset %}{{ s|last }}`, `\\`},
		{"js_length", js, `{% set s %}{{ bs }}{% endset %}{{ s|length }}`, `2`},
		{"js_autoescape_off", js, `{% set s %}{{ bs }}{% endset %}{% autoescape false %}{{ s }}{% endautoescape %}`, `\\`},
		{"css_macro", css, m2 + `"{{ m2(bs) }}"`, `"\\"`},
		{"css_filter_upper", css, `{% filter upper %}{{ bs }}a{% endfilter %}`, `\\A`},
		{"url_block_set", url, `{% set s %}{{ x }}{% endset %}{{ s }}`, `%3Cx%3E`},
		{"url_macro", url, m2 + `{{ m2(bs) }}`, `%5C`},
		{"url_filter_lower", url, `{% filter lower %}{{ x }}{% endfilter %}`, `%3cx%3e`},
		{"json_block_set", jsn, `{% set s %}{{ bs }}{% endset %}"{{ s }}"`, `"\\"`},
		// default's argument is escaped once, and stays a list
		{"js_default_arg_escaped_once", js, `{% set s %}{% endset %}'{{ s|default(x, true) }}'`, `'\u003cx\u003e'`},
		{"js_default_list_arg", js, `{% set s %}{% endset %}{% for u in s|default(["a", "b"], true) %}[{{ u }}]{% endfor %}`, `[a][b]`},
		{"js_default_arg_length", js, `{% set s %}{% endset %}{{ s|default(x, true)|length }}`, `3`},
		{"js_default_keeps_input", js, `{% set s %}{{ bs }}{% endset %}"{{ s|default(x) }}"`, `"\\"`},
		{"js_default_safe_arg", js, `{% set s %}{% endset %}'{{ s|default(bs|safe, true) }}'`, `'\\'`},
		// title's word boundaries would fall inside escape sequences: it
		// sees the text as plain text
		{"js_title_plain", js, `{% set s %}{{ x }}{% endset %}{{ s|title }}|{{ s.title() }}`, `\\u003cx\\u003e|\\U003Cx\\U003E`},
		// stripping would drop the space that ends a CSS escape
		{"css_strip_plain", css, `{% set s %}{{ "a\n" }}{% endset %}{{ s|trim }}2|{{ s|strip }}2|{{ s.strip() }}2|{{ s.rstrip() }}2`, `a\\A2|a\\A2|a\\A2|a\\A2`},
		// an item or a slice can split an escape sequence: plain text
		{"js_item_plain", js, `{% set s %}{{ bs }}{% endset %}'{{ s[0] }}'`, `'\\'`},
		{"js_slice_plain", js, `{% set s %}{{ bs }}{% endset %}'{{ s[:1] }}'`, `'\\'`},
		// trim and strip with arguments can split an escape sequence too
		{"js_trim_args_plain", js, `{% set s %}{{ x }}{% endset %}'{{ s|trim("e") }}'`, `'\\u003cx\\u003'`},
		{"js_strip_method_args_plain", js, `{% set s %}{{ x }}{% endset %}'{{ s.rstrip("e") }}'`, `'\\u003cx\\u003'`},
		// the arguments are escaped for the text's context even with
		// autoescape off
		{"js_args_escaped_autoescape_off", js, `{% set s %}{{ x }}{% endset %}{% autoescape false %}'{{ s|center(15, "'") }}'{% endautoescape %}`, `'\'\u003cx\u003e\''`},
		// a safe argument is HTML-safe, not JS-safe
		{"js_center_safe_fill", js, `{% set s %}{{ x }}{% endset %}'{{ s|center(15, bs|safe) }}'`, `'\\\u003cx\u003e\\'`},
		{"js_indent_safe_arg", js, `{% set s %}{{ x }}{% endset %}'{{ s|indent(q|safe, true) }}'`, `'it\'s\u003cx\u003e'`},
	}
	render := func(t *testing.T, ec runtime.EscapeContext, tpl string, env *miya.Environment) string {
		t.Helper()
		tmpl, err := env.FromString(tpl)
		if err != nil {
			t.Fatal(err)
		}
		ctx := miya.NewTemplateContextAdapter(mixCtx(), env)
		out, err := runtime.NewEvaluator().EvalNode(tmpl.AST(), runtime.NewContextWrapper(ctx, runtime.NewAutoEscaper(nil), ec))
		if err != nil {
			t.Fatal(err)
		}
		return runtime.ToString(out)
	}
	for _, tc := range rows {
		t.Run(tc.name, func(t *testing.T) {
			if got := render(t, tc.ec, tc.tpl, miya.NewEnvironment(miya.WithAutoEscape(true))); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
	t.Run("go_filter_gets_string", func(t *testing.T) {
		env := miya.NewEnvironment(miya.WithAutoEscape(true))
		_ = env.AddFilter("kind", func(v interface{}, a ...interface{}) (interface{}, error) {
			if l, ok := v.([]interface{}); ok {
				v = l[0]
			}
			return fmt.Sprintf("%T", v), nil
		})
		if got := render(t, js, `{% set s %}{{ x }}{% endset %}{{ s|kind }}|{{ [s]|kind }}`, env); got != "string|string" {
			t.Errorf("got %q", got)
		}
	})
	t.Run("other_context_plain", func(t *testing.T) {
		// output escaped for JS, kept in a Go namespace, is plain text in
		// a CSS render
		env := miya.NewEnvironment(miya.WithAutoEscape(true))
		ns := miya.NewNamespace()
		run := func(ec runtime.EscapeContext, tpl string) string {
			tmpl, err := env.FromString(tpl)
			if err != nil {
				t.Fatal(err)
			}
			ctx := miya.NewTemplateContextAdapter(miya.NewContextFrom(map[string]interface{}{"x": "<x>", "gns": ns}), env)
			out, err := runtime.NewEvaluator().EvalNode(tmpl.AST(), runtime.NewContextWrapper(ctx, runtime.NewAutoEscaper(nil), ec))
			if err != nil {
				t.Fatal(err)
			}
			return runtime.ToString(out)
		}
		run(js, `{% set s %}{{ x }}{% endset %}{% set gns.v = s %}`)
		if got := run(css, `{{ gns.v }}|{{ gns.v|upper }}`); got != `\\u003cx\\u003e|\\U003CX\\U003E` {
			t.Errorf("got %q", got)
		}
	})
	t.Run("go_code_gets_strings_nested", func(t *testing.T) {
		// leaves lists the types of the strings in v, also nested ones
		var leaves func(v interface{}) string
		leaves = func(v interface{}) string {
			switch c := v.(type) {
			case []interface{}:
				out := ""
				for _, it := range c {
					out += leaves(it)
				}
				return out
			case map[string]interface{}:
				out := ""
				for _, it := range c {
					out += leaves(it)
				}
				return out
			}
			if reflect.ValueOf(v).Kind() == reflect.String {
				return fmt.Sprintf("%T,", v)
			}
			return ""
		}
		all := func(args []interface{}) string {
			out := ""
			for _, a := range args {
				out += leaves(a)
			}
			return out
		}
		env := miya.NewEnvironment(miya.WithAutoEscape(true))
		_ = env.AddFilter("kind", func(v interface{}, a ...interface{}) (interface{}, error) {
			return leaves(v) + all(a), nil
		})
		_ = env.AddTest("kindt", func(v interface{}, a ...interface{}) (bool, error) {
			return leaves(v)+all(a) == "string,string,", nil
		})
		ns := miya.NewNamespace()
		data := map[string]interface{}{
			"x":     "<x>",
			"kindf": func(a ...interface{}) (interface{}, error) { return all(a), nil },
			"keep":  runtime.KeepMarkupFunc(func(a ...interface{}) (interface{}, error) { return all(a), nil }),
			"ctxf": func(ctx runtime.Context, _ ...interface{}) (interface{}, error) {
				l, _ := ctx.GetVariable("l")
				return leaves(l) + leaves(ctx.All()["l"]), nil
			},
			"gons": ns,
		}
		const set = `{% set s %}{{ x }}{% endset %}{% set l = [[s], {'a': s}] %}`
		rows := map[string]string{
			`{{ [[s]]|kind }}`:                      "string,",
			`{{ {'a': s}|kind }}`:                   "string,",
			`{{ {'a': [s]}|kind }}`:                 "string,",
			`{{ 1|kind([s], {'a': s}) }}`:           "string,string,",
			`{{ [s] is kindt([s]) }}`:               "true",
			`{{ kindf([s], {'a': [s]}) }}`:          "string,string,",
			`{{ keep(s, [s]) }}`:                    "string,string,",
			`{{ ctxf() }}`:                          "string,string,string,string,",
			`{% set gons.v = l %}{{ gons.v|kind }}`: "string,string,",
		}
		for tpl, want := range rows {
			tmpl, err := env.FromString(set + tpl)
			if err != nil {
				t.Fatal(err)
			}
			ctx := miya.NewTemplateContextAdapter(miya.NewContextFrom(data), env)
			out, err := runtime.NewEvaluator().EvalNode(tmpl.AST(), runtime.NewContextWrapper(ctx, runtime.NewAutoEscaper(nil), js))
			if got := runtime.ToString(out); err != nil || got != want {
				t.Errorf("%s: got %q, %v, want %q", tpl, got, err, want)
			}
		}
		if v, _ := ns.Get("v"); leaves(v) != "string,string," {
			t.Errorf("Namespace.Get: got %q", leaves(v))
		}
	})
}
