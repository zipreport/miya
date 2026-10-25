package miya_test

import (
	"errors"
	"strings"
	"testing"

	miya "github.com/zipreport/miya"
	"github.com/zipreport/miya/runtime"
)

type namedString string

type stringer struct{ s string }

func (s stringer) String() string { return s.s }

func TestAutoescapeOutput(t *testing.T) {
	u := "<script>'x'&</script>"
	esc := "&lt;script&gt;&#39;x&#39;&amp;&lt;/script&gt;"

	type data = map[string]interface{}
	tests := []struct {
		name        string
		tmpl        string
		data        data
		off         bool
		want        string
		contains    []string
		notContains []string
		errContains string
		check       bool // true: use contains/notContains/errContains instead of want
	}{
		{name: "named_string", tmpl: `{{ v }}`, data: data{"v": namedString(u)}, want: esc},
		{name: "error_value", tmpl: `{{ v }}`, data: data{"v": errors.New(u)}, want: esc},
		{name: "stringer", tmpl: `{{ v }}`, data: data{"v": stringer{u}}, want: esc},
		{name: "list", tmpl: `{{ v }}`, data: data{"v": []interface{}{u}}, want: "[" + esc + "]"},
		{name: "map", tmpl: `{{ v }}`, data: data{"v": map[string]interface{}{"k": u}}, check: true, contains: []string{esc}, notContains: []string{"<script>"}},
		{name: "number_bool", tmpl: `{{ n }}{{ b }}`, data: data{"n": 3.5, "b": true}, want: "3.5true"},
		{name: "string_unchanged", tmpl: `{{ v }}`, data: data{"v": u}, want: esc},
		{name: "safe_raw", tmpl: `{{ v|safe }}`, data: data{"v": u}, want: u},
		{name: "tojson_raw", tmpl: `{{ v|tojson }}`, data: data{"v": u}, want: `"\u003cscript\u003e\u0027x\u0027\u0026\u003c/script\u003e"`},
		{name: "tojson_apostrophe", tmpl: `{{ "a'b"|tojson }}`, want: `"a\u0027b"`},
		{name: "xmlattr", tmpl: `{{ v|xmlattr }}`, data: data{"v": map[string]interface{}{"k": "<v\">"}}, want: ` k="&lt;v&#34;&gt;"`},
		{name: "xmlattr_bad_key", tmpl: `{{ v|xmlattr }}`, data: data{"v": map[string]interface{}{"a b": 1}}, check: true, errContains: "invalid character in attribute name"},
		{name: "urlize_text", tmpl: `{{ 'go http://x.com <b>'|urlize }}`, want: `go <a href="http://x.com">http://x.com</a> &lt;b&gt;`},
		{name: "urlize_safe_input", tmpl: `{{ v|safe|urlize }}`, data: data{"v": "go http://x.com <b>"}, want: `go <a href="http://x.com">http://x.com</a> <b>`},
		{name: "urlizetruncate_text", tmpl: `{{ 'go http://x.com <b>'|urlizetruncate(40) }}`, want: `go <a href="http://x.com">http://x.com</a> &lt;b&gt;`},
		{name: "nl2br", tmpl: `{{ v|nl2br }}`, data: data{"v": "<s>\n"}, want: `&lt;s&gt;<br>`},
		{name: "truncatehtml_short", tmpl: `{{ v|truncatehtml(100) }}`, data: data{"v": u}, want: esc},
		{name: "truncatehtml_short_safe", tmpl: `{{ v|safe|truncatehtml(100) }}`, data: data{"v": "<b>x</b>"}, want: `<b>x</b>`},
		{name: "truncatehtml_decoded", tmpl: `{{ "&lt;s&gt; long text here"|truncatehtml(5) }}`, check: true, contains: []string{"&lt;s&gt;"}, notContains: []string{"<s>"}},
		{name: "autoescape_off", tmpl: `{{ v }}`, data: data{"v": namedString(u)}, off: true, want: u},
		{name: "autoescape_block", tmpl: `{% autoescape true %}{{ v }}{% endautoescape %}`, data: data{"v": namedString(u)}, off: true, want: esc},
		{name: "autoescape_block_for", tmpl: `{% autoescape true %}{% for i in [1] %}{{ v }}{% endfor %}{% endautoescape %}`, data: data{"v": u}, off: true, want: esc},
		{name: "autoescape_block_with", tmpl: `{% autoescape true %}{% with a = 1 %}{{ v }}{% endwith %}{% endautoescape %}`, data: data{"v": u}, off: true, want: esc},
		{name: "autoescape_off_block_for", tmpl: `{% autoescape false %}{% for i in [1] %}{{ v }}{% endfor %}{% endautoescape %}`, data: data{"v": u}, want: u},
		{name: "safe_pointer", tmpl: `{{ v }}`, data: data{"v": &runtime.SafeValue{Value: "<g>"}}, want: `<g>`},
		{name: "bytes_plain", tmpl: `{{ v }}`, data: data{"v": []byte("<b>")}, want: `&lt;b&gt;`},
		{name: "bytes_safe", tmpl: `{{ v|safe }}`, data: data{"v": []byte("<b>")}, want: `<b>`},
		{name: "urlize_nofollow_twice", tmpl: `{{ 'a http://x.com b http://y.com'|urlize(true, true) }}`, want: `a <a href="http://x.com" rel="nofollow">http://x.com</a> b <a href="http://y.com" rel="nofollow">http://y.com</a>`},
		{name: "truncatehtml_decoded_safe", tmpl: `{{ "&lt;s&gt; long text here"|truncatehtml(5)|safe }}`, check: true, contains: []string{"&lt;s&gt;"}, notContains: []string{"<s>"}},
		{name: "filter_block_pinned", tmpl: `{% filter upper %}<b>x</b>{% endfilter %}`, want: `<B>X</B>`},
		// Known hole owned by plan 027 (Markup): filter-block arguments are not escaped.
		{name: "filter_block_arg_pinned", tmpl: `{% filter replace('X', v) %}X{% endfilter %}`, data: data{"v": "<s>"}, want: `<s>`},
		// Known over-escape owned by plan 027 (Markup): macro output is escaped a second time.
		{name: "macro_nonstring_pinned", tmpl: `{% macro m(a) %}{{ a }}{% endmacro %}{{ m(v) }}`, data: data{"v": namedString("<n>")}, want: `&amp;lt;n&amp;gt;`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := miya.NewEnvironment(miya.WithAutoEscape(!tt.off))
			ctx := miya.NewContext()
			for k, v := range tt.data {
				ctx.Set(k, v)
			}
			got, err := env.RenderString(tt.tmpl, ctx)
			if tt.check {
				if tt.errContains != "" {
					if err == nil || !strings.Contains(err.Error(), tt.errContains) {
						t.Fatalf("want error containing %q, got err=%v out=%q", tt.errContains, err, got)
					}
					return
				}
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				for _, s := range tt.contains {
					if !strings.Contains(got, s) {
						t.Errorf("output %q does not contain %q", got, s)
					}
				}
				for _, s := range tt.notContains {
					if strings.Contains(got, s) {
						t.Errorf("output %q contains %q", got, s)
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
