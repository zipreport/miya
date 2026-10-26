package miya_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"os"
	"strings"
	"testing"

	miya "github.com/zipreport/miya"
	"github.com/zipreport/miya/filters"
	"github.com/zipreport/miya/loader"
	"github.com/zipreport/miya/runtime"
)

type markupParityRow struct {
	Name           string                 `json:"name"`
	Tpl            string                 `json:"tpl"`
	Ctx            map[string]interface{} `json:"ctx"`
	Autoescape     *bool                  `json:"autoescape"`
	Files          map[string]string      `json:"files"`
	Want           string                 `json:"want"`
	KnownUnrelated json.RawMessage        `json:"known_unrelated"`
	Deviation      string                 `json:"deviation"`
}

func addArgtype(env *miya.Environment) {
	_ = env.AddFilter("argtype", func(v interface{}, args ...interface{}) (interface{}, error) {
		kind := func(x interface{}) string {
			if _, ok := x.(string); ok {
				return "string"
			}
			return fmt.Sprintf("%T", x)
		}
		var a interface{}
		if len(args) > 0 {
			a = args[0]
		}
		return "value:" + kind(v) + ",arg:" + kind(a), nil
	})
}

func TestMarkupParity(t *testing.T) {
	raw, err := os.ReadFile("testdata/markup_parity.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []markupParityRow
	if err := json.Unmarshal(raw, &rows); err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		t.Run(row.Name, func(t *testing.T) {
			sl := loader.NewStringLoader(loader.NewDirectTemplateParser())
			for name, src := range row.Files {
				sl.AddTemplate(name, src)
			}
			sl.AddTemplate("main.html", row.Tpl)
			ae := true
			if row.Autoescape != nil {
				ae = *row.Autoescape
			}
			env := miya.NewEnvironment(miya.WithLoader(sl), miya.WithAutoEscape(ae))
			addArgtype(env)
			var got string
			tmpl, err := env.GetTemplate("main.html")
			if err == nil {
				got, err = tmpl.Render(miya.NewContextFrom(row.Ctx))
			}
			if err != nil {
				got = "ERROR: " + err.Error()
			}
			if len(row.KnownUnrelated) > 0 && string(row.KnownUnrelated) != "null" && string(row.KnownUnrelated) != "false" {
				if got == row.Want {
					t.Errorf("%s now matches Jinja2; remove known_unrelated from testdata/markup_parity.json", row.Name)
					return
				}
				t.Skip("known_unrelated")
			}
			if got != row.Want {
				t.Errorf("tpl %q\n got: %q\nwant: %q", row.Tpl, got, row.Want)
			}
		})
	}
}

// ctxU ends with a backslash: in a JS string literal an unescaped one breaks out.
const ctxU = `</script>"<s>'\`

func renderInContext(t *testing.T, ec runtime.EscapeContext, src string) string {
	t.Helper()
	env := miya.NewEnvironment(miya.WithAutoEscape(true))
	tmpl, err := env.FromString(src)
	if err != nil {
		t.Fatal(err)
	}
	ctx := miya.NewTemplateContextAdapter(miya.NewContextFrom(map[string]interface{}{"u": ctxU}), env)
	out, err := runtime.NewEvaluator().EvalNode(tmpl.AST(), runtime.NewContextWrapper(ctx, runtime.NewAutoEscaper(nil), ec))
	if err != nil {
		t.Fatal(err)
	}
	return runtime.ToString(out)
}

func TestMarkupEscapeContexts(t *testing.T) {
	esc := runtime.NewAutoEscaper(nil)
	js := runtime.EscapeContextJS
	xml := runtime.EscapeContextXML
	m := "{% macro b(x) %}<b>{{ x }}</b>{% endmacro %}"
	tj, _ := filters.ToJSONFilter(ctxU)
	tojson := runtime.ToString(tj)

	tests := []struct {
		name      string
		ec        runtime.EscapeContext
		src, want string
	}{
		{"js_e", js, `{{ u|e }}`, esc.Escape(html.EscapeString(ctxU), js)},
		{"js_safe", js, `{{ u|safe }}`, ctxU},
		{"js_safe_concat", js, `{{ (u|safe) ~ u }}`, esc.Escape(ctxU+ctxU, js)},
		{"js_macro", js, m + `{{ b(u) }}`, esc.Escape("<b>"+esc.Escape(ctxU, js)+"</b>", js)},
		{"js_nl2br", js, `{{ u|nl2br }}`, esc.Escape(html.EscapeString(ctxU), js)},
		{"js_urlize", js, `{{ u|urlize }}`, esc.Escape(html.EscapeString(ctxU), js)},
		{"js_tojson", js, `{{ u|tojson }}`, esc.Escape(tojson, js)},
		{"js_ae_block_var", js, `{% autoescape true %}{{ u }}{% endautoescape %}`, esc.Escape(ctxU, js)},
		{"js_ae_block_set", js, `{% autoescape true %}{% set q %}{{ u }}{% endset %}{{ q }}{% endautoescape %}`, esc.Escape(esc.Escape(ctxU, js), js)},
		{"js_ae_off_block", js, `{% autoescape false %}{{ u }}{% endautoescape %}`, ctxU},
		{"js_filterblock", js, `{% filter upper %}{{ u }}{% endfilter %}`, esc.Escape(strings.ToUpper(esc.Escape(ctxU, js)), js)},
		{"xml_macro", xml, m + `{{ b(u) }}`, "<b>" + esc.Escape(ctxU, xml) + "</b>"},
		{"xml_concat", xml, `{{ (u|safe) ~ u }}`, ctxU + esc.Escape(ctxU, xml)},
		{"xml_join", xml, `{{ [u|safe, u]|join }}`, ctxU + esc.Escape(ctxU, xml)},
		{"xml_filterblock_replace", xml, `{% filter replace('X', u) %}<i>X</i>{% endfilter %}`, "<i>" + esc.Escape(ctxU, xml) + "</i>"},
		{"js_map_e_loop", js, `{% for x in [u]|map('e') %}{{ x }}{% endfor %}`, esc.Escape(html.EscapeString(ctxU), js)},
		{"js_map_e_index", js, `{{ ([u]|map('e'))[0] }}`, esc.Escape(html.EscapeString(ctxU), js)},
		{"js_map_tojson", js, `{{ ([u]|map('tojson'))[0] }}`, esc.Escape(tojson, js)},
		{"html_map_e_join", runtime.EscapeContextHTML, `{{ [u]|map('e')|join }}`, html.EscapeString(ctxU)},
		{"html_e", runtime.EscapeContextHTML, `{{ u|e }}`, html.EscapeString(ctxU)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := renderInContext(t, tt.ec, tt.src); got != tt.want {
				t.Errorf("src %q\n got: %q\nwant: %q", tt.src, got, tt.want)
			}
		})
	}
}

func renderAE(t *testing.T, ae bool, src string, data map[string]interface{}, setup func(*miya.Environment)) string {
	t.Helper()
	env := miya.NewEnvironment(miya.WithAutoEscape(ae))
	if setup != nil {
		setup(env)
	}
	tmpl, err := env.FromString(src)
	if err != nil {
		t.Fatal(err)
	}
	out, err := tmpl.Render(miya.NewContextFrom(data))
	if err != nil {
		t.Fatalf("render %q: %v", src, err)
	}
	return out
}

func TestMarkupOutputIsEscaped(t *testing.T) {
	data := map[string]interface{}{"n": namedString("<s>"), "st": stringer{"<s>"}, "er": errors.New("<s>")}
	for _, name := range []string{"n", "st", "er"} {
		t.Run(name, func(t *testing.T) {
			if got := renderAE(t, true, "{{ "+name+" }}", data, nil); got != "&lt;s&gt;" {
				t.Errorf("got %q want %q", got, "&lt;s&gt;")
			}
		})
	}
}

func TestMarkupGoInterop(t *testing.T) {
	const m = "{% macro b(x) %}<b>{{ x }}</b>{% endmacro %}"
	tests := []struct {
		name, src, want string
		data            map[string]interface{}
		setup           func(*miya.Environment)
	}{
		{"go_callable_gets_string", m + `{{ kind(b("x")) }}`, "string", nil, func(env *miya.Environment) {
			env.AddGlobal("kind", func(args ...interface{}) (interface{}, error) { return fmt.Sprintf("%T", args[0]), nil })
		}},
		{"safevalue_filter_normalised", `{{ ""|legacy|upper }}`, "<B>X</B>", nil, func(env *miya.Environment) {
			_ = env.AddFilter("legacy", func(v interface{}, a ...interface{}) (interface{}, error) {
				return runtime.SafeValue{Value: "<b>x</b>"}, nil
			})
		}},
		{"go_test_gets_string", `{{ ('<b>'|safe) is isplain }}`, "true", nil, func(env *miya.Environment) {
			_ = env.AddTest("isplain", func(v interface{}, args ...interface{}) (bool, error) { _, ok := v.(string); return ok, nil })
		}},
		{"go_test_gets_string_select", `{{ [('<b>'|safe)]|select('isplain')|list|length }}`, "1", nil, func(env *miya.Environment) {
			_ = env.AddTest("isplain", func(v interface{}, args ...interface{}) (bool, error) { _, ok := v.(string); return ok, nil })
		}},
		{"joiner_safe_sep", `{% set j = joiner("<br>"|safe) %}{% for i in [1,2,3] %}{{ j() }}{{ i }}{% endfor %}`, "1<br>2<br>3", nil, nil},
		{"escaped_test", m + `{{ "T" if b("x") is escaped else "F" }}`, "T", nil, nil},
		{"forceescape_length", `{{ u|forceescape|length }}`, "9", map[string]interface{}{"u": "<s>"}, nil},
		{"format_named_arg", `{{ ("<b>%s</b>"|safe)|format(n) }}`, "<b>&lt;s&gt;</b>", map[string]interface{}{"n": namedString("<s>")}, nil},
		{"safe_pointer_upper", `{{ v|upper }}`, "<G>", map[string]interface{}{"v": &runtime.SafeValue{Value: "<g>"}}, nil},
		{"selectattr_markup", `{{ items|selectattr('t', 'equalto', 'x')|list|length }}`, "1",
			map[string]interface{}{"items": []map[string]interface{}{{"t": runtime.Markup("x")}}}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := renderAE(t, true, tt.src, tt.data, tt.setup); got != tt.want {
				t.Errorf("got %q want %q", got, tt.want)
			}
		})
	}
}

func TestMarkupAutoescapeBlocks(t *testing.T) {
	data := map[string]interface{}{"u": "<s>"}
	tests := []struct{ name, src, want string }{
		{"for_in_block", `{% autoescape true %}{% for i in [1] %}{{ u }}{% endfor %}{% endautoescape %}`, "&lt;s&gt;"},
		{"filter_in_block", `{% autoescape true %}{{ u|upper }}{% endautoescape %}`, "&lt;S&gt;"},
		// urlencode is not a built-in fallback filter: it needs the environment behind the block
		{"env_filter_in_block", `{% autoescape true %}{{ u|urlencode }}{% endautoescape %}`, "%3Cs%3E"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := renderAE(t, false, tt.src, data, nil); got != tt.want {
				t.Errorf("got %q want %q", got, tt.want)
			}
		})
	}
}

func TestMarkupFilterBlocks(t *testing.T) {
	t.Run("title_keeps_markup", func(t *testing.T) {
		got := renderAE(t, true, `{% filter title %}<b>x</b>{% endfilter %}`, nil, nil)
		if strings.Contains(got, "&lt;") || !strings.HasPrefix(strings.ToLower(got), "<b>") {
			t.Errorf("got %q", got)
		}
	})
}

func TestMarkupRecursiveLoopBlocks(t *testing.T) {
	tree := []interface{}{map[string]interface{}{"n": "<a>", "c": []interface{}{map[string]interface{}{"n": "<b>", "c": []interface{}{}}}}}
	data := map[string]interface{}{"tree": tree}
	const head = `{% for x in tree recursive %}[{{ x.n }}`
	const tail = `]{% endfor %}`
	tests := []struct {
		name string
		ae   bool
		src  string
		want string
	}{
		{"recursive_loop_ae_off_body_in_ae_block", false, head + `{% autoescape true %}{{ loop(x.c) }}{% endautoescape %}` + tail, `[<a>[&lt;b&gt;]]`},
		{"recursive_loop_ae_on_body_in_ae_off_block", true, head + `{% autoescape false %}{{ loop(x.c) }}{% endautoescape %}` + tail, `[&lt;a&gt;[&lt;b&gt;]]`},
		{"recursive_loop_concat", false, head + `{% autoescape true %}{{ loop(x.c) ~ '' }}{% endautoescape %}` + tail, `[<a>[&lt;b&gt;]]`},
		{"recursive_loop_set", false, head + `{% autoescape true %}{% set r = loop(x.c) %}{{ r }}{% endautoescape %}` + tail, `[<a>[&lt;b&gt;]]`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := renderAE(t, tt.ae, tt.src, data, nil); got != tt.want {
				t.Errorf("got %q want %q", got, tt.want)
			}
		})
	}
}
