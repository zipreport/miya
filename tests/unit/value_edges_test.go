package miya_test

import (
	"strings"
	"testing"

	miya "github.com/zipreport/miya"
	"github.com/zipreport/miya/loader"
)

func TestValueEdges(t *testing.T) {
	inf := pvD{"w": "inf"}
	cases := []struct {
		name       string
		tpl        string
		data       pvD
		autoescape bool
		strict     bool
		want       string
		errHas     string
		anyErr     bool
	}{
		{name: "u1", tpl: `{{ [nope, 1] }}`, want: `[Undefined, 1]`},
		{name: "u2", tpl: `{{ {'a': nope} }}`, want: `{'a': Undefined}`},
		{name: "u3", tpl: `{{ {'a': nope}|tojson }}`, errHas: "not JSON serializable"},
		{name: "u4", tpl: `{{ [nope]|tojson }}`, errHas: "not JSON serializable"},
		{name: "u5", tpl: `{{ [nope]|length }}`, want: `1`},
		{name: "n1", tpl: `{{ (-7) is integer }}`, want: `true`},
		{name: "n2", tpl: `{{ -7 is integer }}`, want: `true`},
		{name: "n3", tpl: `{{ - x is integer }}`, data: pvD{"x": 7}, want: `true`},
		{name: "n4", tpl: `{{ (-7.0) is float }}`, want: `true`},
		{name: "n5", tpl: `{{ -7 // 2 }}|{{ -7 % 3 }}`, want: `-4|2`},
		{name: "n6", tpl: `{{ (-7)|string }}|{{ [-7] }}|{{ -7 * 1.5 }}`, want: `-7|[-7]|-10.5`},
		{name: "s1", tpl: `{{ w * 2 }}`, data: inf, want: `infinf`},
		{name: "s2", tpl: `{{ w * 2 }}`, data: pvD{"w": "nan"}, want: `nannan`},
		{name: "s3_sub", tpl: `{{ w - 1 }}`, data: inf, anyErr: true},
		{name: "s3_div", tpl: `{{ w / 2 }}`, data: inf, anyErr: true},
		{name: "deviation_inf_plus", tpl: `{{ w + 1 }}`, data: inf, want: `inf1`},
		{name: "s6_abs", tpl: `{{ w|abs }}`, data: inf, errHas: "abs filter requires a number"},
		{name: "deviation_sum_string", tpl: `{{ [w]|sum }}`, data: inf, want: `0`}, // Jinja2: TypeError
		{name: "k5", tpl: `{{ none in {none: 1} }}|{{ none is in {none: 1} }}`, want: `true|true`},
		{name: "k6", tpl: `{{ {none: 1}.get(none) }}`, want: `1`},
		{name: "k7", tpl: `{{ (1.0|round(2)) in {1: 1} }}`, want: `true`},
		{name: "deviation_round_value_attr", tpl: `{{ [2.5|round(1)]|map(attribute='Value')|list }}`, want: `[None]`}, // Jinja2 prints [Undefined]; map(attribute=) yields None for missing attributes
		{name: "k1", tpl: `{{ '<nil>' in ({none: 'x'}|string) }}`, want: `false`},
		{name: "k2", tpl: `{{ {(1.0|round(2)): 'x'}[1] }}`, want: `x`},
		{name: "k3", tpl: `{{ {1: 'x'}[1.0|round(2)] }}`, want: `x`},
		{name: "k4", tpl: `{{ {1: 'a'}[1] }}|{{ {1: 'a'}[1.0] }}`, want: `a|a`},
		{name: "r1", tpl: `{{ (2.5|round(1)).Value }}|{{ (2.5|round(1)).Places }}`, want: `|`},
		{name: "r1_strict", tpl: `{{ (2.5|round(1)).Value }}`, strict: true, errHas: "Value"},
		{name: "deviation_none_key", tpl: `{{ {none: 'x'} }}`, want: `{'None': 'x'}`},
		{name: "deviation_bool_key", tpl: `{{ {true: 'x'}[1] }}`, want: ``},
		{name: "deviation_strict_set", tpl: `{% set a = nope %}`, strict: true, errHas: "nope"},
		{name: "deviation_numeric_data_string_neg", tpl: `{{ w * 2 }}`, data: pvD{"w": "-2.5"}, want: `-5`},
		{name: "deviation_numeric_data_string_exp", tpl: `{{ w * 2 }}`, data: pvD{"w": "1e2"}, want: `200`},
		{name: "deviation_repeat_count_float", tpl: `{{ 'ab' * (2.0|round(0)) }}`, want: `abab`},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			opts := []miya.EnvironmentOption{miya.WithAutoEscape(tc.autoescape)}
			if tc.strict {
				opts = append(opts, miya.WithStrictUndefined(true))
			}
			env := miya.NewEnvironment(opts...)
			got, err := renderGo(t, env, tc.tpl, tc.data)
			switch {
			case tc.errHas != "":
				if err == nil || !strings.Contains(err.Error(), tc.errHas) {
					t.Fatalf("want error containing %q, got err=%v out=%q", tc.errHas, err, got)
				}
			case tc.anyErr:
				if err == nil {
					t.Fatalf("want error, got %q", got)
				}
			default:
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got != tc.want {
					t.Fatalf("got %q, want %q", got, tc.want)
				}
			}
		})
	}
}

func TestSuperEscaping(t *testing.T) {
	const base = `{% block b %}<i>{{ v }}</i>{% endblock %}`
	const s1 = `{% extends 'base' %}{% block b %}{% autoescape false %}{{ super() }}{% endautoescape %}{% endblock %}`
	cases := []struct {
		name  string
		child string
		ae    bool
		want  string
	}{
		{"S1", s1, true, `<i>&lt;x&gt;</i>`},
		{"S2", `{% extends 'base' %}{% block b %}{% autoescape true %}{{ super() }}{% endautoescape %}{% endblock %}`, false, `<i><x></i>`},
		{"S3", `{% extends 'base' %}{% block b %}[{{ super() }}]{% endblock %}`, true, `[<i>&lt;x&gt;</i>]`},
		{"S4", s1, false, `<i><x></i>`},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			l := loader.NewStringLoader(loader.NewDirectTemplateParser())
			l.AddTemplate("base", base)
			l.AddTemplate("child", c.child)
			env := miya.NewEnvironment(miya.WithLoader(l), miya.WithAutoEscape(c.ae))
			tmpl, err := env.GetTemplate("child")
			if err != nil {
				t.Fatal(err)
			}
			got, err := tmpl.Render(miya.NewContextFrom(map[string]interface{}{"v": "<x>"}))
			if err != nil || got != c.want {
				t.Fatalf("got %q, %v; want %q", got, err, c.want)
			}
		})
	}
}

// deviation_super_parent_tag: Jinja2 3.1.6 prints [&lt;i&gt;&amp;lt;x&amp;gt;&lt;/i&gt;]; miya does not double-escape.
func TestSuperParentTag(t *testing.T) {
	l := loader.NewStringLoader(loader.NewDirectTemplateParser())
	l.AddTemplate("base", `{% autoescape false %}{% block b %}<i>{{ v }}</i>{% endblock %}{% endautoescape %}`)
	l.AddTemplate("child", `{% extends 'base' %}{% block b %}[{{ super() }}]{% endblock %}`)
	env := miya.NewEnvironment(miya.WithLoader(l), miya.WithAutoEscape(true))
	tmpl, err := env.GetTemplate("child")
	if err != nil {
		t.Fatal(err)
	}
	got, err := tmpl.Render(miya.NewContextFrom(map[string]interface{}{"v": "<x>"}))
	t.Logf("deviation_super_parent_tag: %q %v", got, err)
	if err != nil || got != `[<i>&lt;x&gt;</i>]` {
		t.Fatalf("got %q, %v", got, err)
	}
}
