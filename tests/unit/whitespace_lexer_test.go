package miya_test

import (
	"testing"

	miya "github.com/zipreport/miya"
	"github.com/zipreport/miya/loader"
)

// TestLexerWhitespace pins whitespace control against Jinja2 3.1.6 output.
// Templates go through FromString, the path every loader takes once plan 044
// has merged (Environment.GetTemplate compiles the loader's source).
func TestLexerWhitespace(t *testing.T) {
	type opts struct{ trim, lstrip, keep, delim bool }
	tests := []struct {
		name string
		src  string
		opts opts
		want string
	}{
		{"comment_line_trim", "x = 1;\n{# note #}\ny = 2;", opts{trim: true}, "x = 1;\ny = 2;"},
		{"comment_inline", "Total {# amount #} due", opts{trim: true, lstrip: true}, "Total  due"},
		{"comment_multiline", "<p>a</p>\n{# a\nb #}\n<p>b</p>", opts{trim: true, lstrip: true}, "<p>a</p>\n<p>b</p>"},
		{"marker_in_string", "{% if true %}{{ '-%}' }}{% endif %}", opts{trim: true, lstrip: true}, "-%}"},
		{"comment_in_string", "{{ 'a {# x #} b' }}", opts{trim: true, lstrip: true}, "a {# x #} b"},
		{"data_keeps_blank_line", "{{ d }}", opts{trim: true, lstrip: true}, "x %}\n\ny"},
		{"option_text_kept", "<option>  a\n   b </option>", opts{}, "<option>  a\n   b </option>"},
		{"trailing_newline_removed", "a\n", opts{}, "a"},
		{"trailing_newline_kept", "a\n", opts{keep: true}, "a\n"},
		{"trailing_newline_once", "a\n\n", opts{}, "a\n"},
		{"crlf_normalized", "a\r\nb\rc", opts{}, "a\nb\nc"},
		{"lstrip_first_line", "  {% if true %}x{% endif %}", opts{lstrip: true}, "x"},
		{"lstrip_indented_line", "p\n  {% if true %}x\n  {% endif %}q", opts{lstrip: true}, "p\nx\nq"},
		{"lstrip_not_after_text", "p  {% if true %}x{% endif %}", opts{lstrip: true}, "p  x"},
		{"plus_disables_lstrip", "  {%+ if true %}x{% endif %}", opts{lstrip: true}, "  x"},
		{"plus_disables_trim", "{% if true +%}\nx{% endif %}", opts{trim: true}, "\nx"},
		{"dash_strips_before", "a  \n {%- if true %}x{% endif %}", opts{}, "ax"},
		{"dash_strips_after", "{% if true -%}  \n x{% endif %}", opts{}, "x"},
		{"raw_verbatim", "{% raw %}{{ a }} {% for x in y %}{# c #}{% endraw %}", opts{}, "{{ a }} {% for x in y %}{# c #}"},
		{"raw_dash", "{% raw -%}  {{ a }}  {%- endraw %}", opts{}, "{{ a }}"},
		{"raw_trim_after_endraw", "{% raw %}\n{{ a }}\n{% endraw %}\nz", opts{trim: true}, "\n{{ a }}\nz"},
		{"lstrip_after_var_same_line", "<td>{{ v }}  {% if true %}y{% endif %}</td>", opts{lstrip: true}, "<td>V  y</td>"},
		{"comment_dash_inline", "a {#-#} b", opts{}, "a b"},
		{"raw_whitespace_only_lstrip", "{% raw %}  {% endraw %}", opts{lstrip: true}, "  "},
		{"endraw_plus_trim", "{% raw %}a{% endraw +%}\nz", opts{trim: true}, "a\nz"},
		{"dash_var_before", "a {{- v }}", opts{}, "aV"},
		{"dash_strips_nbsp", "a\u00a0{{- v }}", opts{}, "aV"},
		{"custom_delimiters", "p\n  [% if true %]\n[[ v ]] [# c #]\n  [% endif %]\n[% raw %][[ x ]][% endraw %]", opts{trim: true, lstrip: true, delim: true}, "p\nV [[ x ]]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := miya.NewEnvironment(miya.WithAutoEscape(false), miya.WithTrimBlocks(tt.opts.trim),
				miya.WithLstripBlocks(tt.opts.lstrip), miya.WithKeepTrailingNewline(tt.opts.keep))
			if tt.opts.delim {
				env.SetDelimiters("[[", "]]", "[%", "%]")
				env.SetCommentDelimiters("[#", "#]")
			}
			tmpl, err := env.FromString(tt.src)
			if err != nil {
				t.Fatal(err)
			}
			got, err := tmpl.Render(miya.NewContextFrom(map[string]interface{}{"d": "x %}\n\ny", "v": "V"}))
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}

	t.Run("raw_unclosed", func(t *testing.T) {
		if _, err := miya.NewEnvironment().FromString("{% raw %}a"); err == nil {
			t.Error("want an error for an unclosed raw block")
		}
	})

	t.Run("raw_plus_error", func(t *testing.T) {
		// Jinja2 raises TemplateSyntaxError for {% raw +%}
		if _, err := miya.NewEnvironment().FromString("{% raw +%}x{% endraw %}"); err == nil {
			t.Error("want an error for {% raw +%}")
		}
	})

	t.Run("loader_dash", func(t *testing.T) {
		sl := loader.NewStringLoader(loader.NewDirectTemplateParser())
		sl.AddTemplate("t.html", "a  {{- 1 }}")
		env := miya.NewEnvironment(miya.WithLoader(sl))
		tmpl, err := env.GetTemplate("t.html")
		if err != nil {
			t.Fatal(err)
		}
		got, err := tmpl.Render(nil)
		if err != nil {
			t.Fatal(err)
		}
		if got != "a1" {
			t.Errorf("got %q, want %q", got, "a1")
		}
	})

	// Last: at 78d1bae this input panics inside applyBlockTrimming.
	t.Run("data_with_tags_and_blank_lines", func(t *testing.T) {
		env := miya.NewEnvironment(miya.WithAutoEscape(false), miya.WithTrimBlocks(true), miya.WithLstripBlocks(true))
		tmpl, err := env.FromString("{{ d }}")
		if err != nil {
			t.Fatal(err)
		}
		d := "a %}\n\nb\n{% c"
		got, err := tmpl.Render(miya.NewContextFrom(map[string]interface{}{"d": d}))
		if err != nil {
			t.Fatal(err)
		}
		if got != d {
			t.Errorf("got %q, want %q", got, d)
		}
	})
}
