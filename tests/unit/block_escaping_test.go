package miya_test

import (
	"testing"

	miya "github.com/zipreport/miya"
	"github.com/zipreport/miya/loader"
)

func TestBlockEscaping(t *testing.T) {
	p := `{% autoescape false %}{% block b %}{{ x }}{% endblock %}{% endautoescape %}`
	pi := `{% autoescape false %}{{ x }}{% if true %}{% block b %}{{ x }}{% endblock %}{% endif %}{% endautoescape %}`
	inct := `{% autoescape false %}{% include "t" %}{% endautoescape %}`
	rows := []struct {
		name, main string
		tpls       map[string]string
		off        bool
		want       string
	}{
		{"p", p, nil, false, `&lt;`},
		{"pn", p + `{{ x }}`, nil, false, `&lt;&lt;`},
		{"c", `{% extends "p" %}{% block b %}{{ x }}{% endblock %}`, map[string]string{"p": p}, false, `&lt;`},
		{"deviation_super_reescape", `{% extends "p" %}{% block b %}{{ super() }}{{ x }}{% endblock %}`, map[string]string{"p": p}, false, `&lt;&lt;`},
		{"pi", pi, nil, false, `<&lt;`},
		{"ci", `{% extends "pi" %}{% block b %}[{{ x }}]{% endblock %}`, map[string]string{"pi": pi}, false, `<[&lt;]`},
		{"cb", `{% extends "p" %}{% block b %}{% autoescape false %}{{ x }}{% endautoescape %}{% endblock %}`, map[string]string{"p": p}, false, `<`},
		{"p2", `{% autoescape true %}{% block b %}{{ x }}{% endblock %}{% endautoescape %}`, nil, true, `<`},
		{"c2", `{% extends "p2" %}{% block b %}{{ x }}{% endblock %}`, map[string]string{"p2": `{% autoescape true %}{% block b %}{{ x }}{% endblock %}{% endautoescape %}`}, true, `<`},
		{"inct", inct, map[string]string{"t": `{{ x }}`}, false, `&lt;`},
		{"nested", `{% autoescape false %}{% block a %}{{ x }}{% block b %}{{ x }}{% endblock %}{% endblock %}{% endautoescape %}`, nil, false, `&lt;&lt;`},
		{"filter_around", `{% autoescape false %}{% filter upper %}{% block b %}{{ x }}{% endblock %}{% endfilter %}{% endautoescape %}`, nil, false, `&LT;`},
		{"include_block", `{% autoescape false %}{% include "inc" %}{% endautoescape %}`, map[string]string{"inc": `{% block b %}{{ x }}{% endblock %}`}, false, `&lt;`},
		{"include_plain", inct, map[string]string{"t": `{{ x }}`}, false, `&lt;`},
		{"include_without_context", `{% autoescape false %}{% include "t" without context %}{% endautoescape %}`, map[string]string{"t": `{{ "<" }}`}, false, `&lt;`},
		{"hoisted", `{% extends "hp" %}{% autoescape false %}{% for i in [1] %}{% block b %}{{ x }}{% endblock %}{% endfor %}{% endautoescape %}`, map[string]string{"hp": `{% block b %}{% endblock %}`}, false, `&lt;&lt;`},
		{"call_body", `{% macro w() %}{{ caller() }}{% endmacro %}{% autoescape false %}{% call w() %}{% block b %}{{ x }}{% endblock %}{% endcall %}{% endautoescape %}`, nil, false, `&amp;lt;`},
	}
	for _, tc := range rows {
		t.Run(tc.name, func(t *testing.T) {
			l := loader.NewStringLoader(loader.NewDirectTemplateParser())
			for n, s := range tc.tpls {
				l.AddTemplate(n, s)
			}
			l.AddTemplate("main", tc.main)
			env := miya.NewEnvironment(miya.WithLoader(l), miya.WithAutoEscape(!tc.off))
			tmpl, err := env.GetTemplate("main")
			if err != nil {
				t.Fatal(err)
			}
			got, err := tmpl.Render(miya.NewContextFrom(map[string]interface{}{"x": "<"}))
			if err != nil || got != tc.want {
				t.Fatalf("got %q, %v; want %q", got, err, tc.want)
			}
		})
	}
}
