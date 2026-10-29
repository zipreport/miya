package miya_test

import (
	"strings"
	"testing"

	"github.com/zipreport/miya"
	"github.com/zipreport/miya/loader"
)

func TestIncludeExtends(t *testing.T) {
	tpls := map[string]string{
		"base": `[{% block x %}B{{ v }}{% endblock %}]`,
		"c":    `{% extends 'base' %}{% block x %}X{{ super() }}{% endblock %}`,
		"d":    `{% extends layout %}{% block x %}D{% endblock %}`,
		"e":    `{% extends 'nope' %}`,
		"f":    `{% extends 'base' %}{% block x %}{{ u }}{% endblock %}`,
		"p":    `<{{ v }}>`,
		"m":    `a{% include 'c' %}b`,
		"m2":   `{% for i in [1,2,3] %}{% include 'c' %}{% endfor %}`,
		"m3":   `{% include 'd' %}`,
		"m4":   `a{% include 'e' ignore missing %}b`,
		"m5":   `{% include 'f' %}`,
		"m6":   `{% autoescape false %}{% include 'f' %}{% endautoescape %}`,
		"m7":   `{% include 'c' without context %}`,
		"m8":   `x{% include 'p' %}y`,
		"ab":   `{% block b %}<i>{{ u }}</i>{% endblock %}`,
		"ac":   `{% extends 'ab' %}{% block b %}{% autoescape false %}{{ super() }}{% endautoescape %}{% endblock %}`,
		"m9":   `{% include 'd' without context %}`,
		"m10":  `{% autoescape false %}{% include 'ac' %}{% endautoescape %}`,
	}
	type tc struct {
		name    string
		seq     []string
		on, off []string
		errSub  string
	}
	cases := []tc{
		{name: "resolved", seq: []string{"m"}, on: []string{"a[XBV]b"}, off: []string{"a[XBV]b"}},
		{name: "in_loop", seq: []string{"m2"}, on: []string{"[XBV][XBV][XBV]"}, off: []string{"[XBV][XBV][XBV]"}},
		{name: "dynamic_parent", seq: []string{"m3"}, on: []string{"[D]"}, off: []string{"[D]"}},
		{name: "missing_parent_ignore_missing", seq: []string{"m4"}, errSub: "nope"},
		{name: "escaped_block", seq: []string{"m5"}, on: []string{"[&lt;u&gt;]"}, off: []string{"[<u>]"}},
		{name: "autoescape_false_block_includer", seq: []string{"m6"}, on: []string{"[&lt;u&gt;]"}, off: []string{"[<u>]"}},
		{name: "without_context", seq: []string{"m7"}, on: []string{"[XB]"}, off: []string{"[XB]"}},
		{name: "direct_and_included_agree", seq: []string{"c", "m", "c"},
			on: []string{"[XBV]", "a[XBV]b", "[XBV]"}, off: []string{"[XBV]", "a[XBV]b", "[XBV]"}},
		{name: "plain_include_unchanged", seq: []string{"m8"}, on: []string{"x<V>y"}, off: []string{"x<V>y"}},
		// The parent is resolved with the include's own context (issue #5).
		// Jinja2 3.1.6: UndefinedError 'layout' is undefined; the substring
		// tells the parent-name error apart from a missing-template error.
		{name: "without_context_dynamic_parent", seq: []string{"m9"}, errSub: "dynamic template name"},
		{name: "autoescape_false_includer_super", seq: []string{"m10"}, on: []string{"<i>&lt;u&gt;</i>"}, off: []string{"<i><u></i>"}},
	}
	for _, c := range cases {
		for _, ae := range []bool{true, false} {
			name := c.name + "/ae_off"
			want := c.off
			if ae {
				name, want = c.name+"/ae_on", c.on
			}
			t.Run(name, func(t *testing.T) {
				sl := loader.NewStringLoader(loader.NewDirectTemplateParser())
				for n, s := range tpls {
					sl.AddTemplate(n, s)
				}
				env := miya.NewEnvironment(miya.WithLoader(sl), miya.WithAutoEscape(ae))
				for i, entry := range c.seq {
					tmpl, err := env.GetTemplate(entry)
					if err != nil {
						t.Fatal(err)
					}
					ctx := miya.NewContext()
					ctx.Set("v", "V")
					ctx.Set("layout", "base")
					ctx.Set("u", "<u>")
					got, err := tmpl.Render(ctx)
					if c.errSub != "" {
						if err == nil || !strings.Contains(err.Error(), c.errSub) {
							t.Fatalf("want error containing %q, got %q, %v", c.errSub, got, err)
						}
						continue
					}
					if err != nil {
						t.Fatalf("render %s: %v", entry, err)
					}
					if got != want[i] {
						t.Fatalf("render %s: got %q, want %q", entry, got, want[i])
					}
				}
			})
		}
	}
}
