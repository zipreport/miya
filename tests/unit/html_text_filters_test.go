package miya_test

import (
	"testing"

	miya "github.com/zipreport/miya"
)

func TestHTMLTextFilters(t *testing.T) {
	env := miya.NewEnvironment(miya.WithAutoEscape(false))
	ctx := miya.NewContextFrom(map[string]interface{}{
		"h":   "<p>a</p>  <p>b</p>\n<br/>c",
		"d1":  map[string]interface{}{"a b": "x/y"},
		"d2":  map[string]interface{}{"b": 1, "a": 2},
		"ms":  map[string]string{"k": "v w"},
		"nb":  "see\u00a0http://x.com",
		"tn":  "hello\u00a0world",
		"tv":  "a\vb",
		"fsx": "a\x1cb",
		"wn":  "a\u00a0b",
		"ls":  "x http://a.com\u2028y",
	})

	tests := []struct {
		name string
		tmpl string
		want string
	}{
		{"ue_space", `{{ 'a b/c'|urlencode }}`, "a%20b/c"},
		{"ue_amp", `{{ 'a&b=c?d'|urlencode }}`, "a%26b%3Dc%3Fd"},
		{"ue_dict", `{{ d1|urlencode }}`, "a+b=x%2Fy"},
		{"deviation_ue_dict_order", `{{ d2|urlencode }}`, "a=2&b=1"},
		{"ue_pairs", `{{ [['a', 1], ['b', 'x y']]|urlencode }}`, "a=1&b=x+y"},
		{"ue_number", `{{ 5|urlencode }}`, "5"},
		{"ue_bad", `{{ [1, 2]|urlencode }}`, "ERR"},
		{"ue_unicode", `{{ 'ação'|urlencode }}`, "a%C3%A7%C3%A3o"},
		{"ue_tilde", `{{ 'a~b_c.d-e'|urlencode }}`, "a~b_c.d-e"},
		{"ue_dotdot", `{{ '../admin'|urlencode }}`, "../admin"},
		{"ue_map_string", `{{ ms|urlencode }}`, "k=v+w"},
		{"deviation_ue_bool", `{{ {'a': true}|urlencode }}`, "a=true"},
		{"deviation_ue_str_pair", `{{ ['ab']|urlencode }}`, "ERR"},
		{"st_collapse", `{{ '<p>a</p>  <p>b</p>'|striptags }}`, "a b"},
		{"st_newlines", `{{ h|striptags }}`, "a b c"},
		{"st_entities", `{{ '<b>a &amp; b</b>'|striptags }}`, "a & b"},
		{"st_entity_after", `{{ 'a&lt;b&gt;'|striptags }}`, "a<b>"},
		{"st_comment", `{{ 'a<!-- x -->b'|striptags }}`, "ab"},
		{"st_unclosed", `{{ 'a<b'|striptags }}`, "a<b"},
		{"st_pyspace", `{{ fsx|striptags }}`, "a b"},
		{"ti_basic", `{{ 'hello world'|title }}`, "Hello World"},
		{"ti_mixed", `{{ 'hELLO wORLD'|title }}`, "Hello World"},
		{"ti_acme", `{{ 'ACME corp'|title }}`, "Acme Corp"},
		{"ti_dash", `{{ 'a-b c_d'|title }}`, "A-B C_d"},
		{"ti_brackets", `{{ 'a(b) c[d] e{f} g<h'|title }}`, "A(B) C[D] E{F} G<H"},
		{"ti_dot", `{{ 'a.b c,d'|title }}`, "A.b C,d"},
		{"ti_unicode", `{{ 'ação über'|title }}`, "Ação Über"},
		{"ti_apos", `{{ "they're bill's"|title }}`, "They're Bill's"},
		{"ti_nbsp", `{{ tn|title }}`, "Hello\u00a0World"},
		{"ti_vt", `{{ tv|title }}`, "A\vB"},
		{"wc_basic", `{{ 'a b  c'|wordcount }}`, "3"},
		{"wc_punct", `{{ 'hello, world! foo-bar'|wordcount }}`, "4"},
		{"wc_apos", `{{ "don't stop"|wordcount }}`, "3"},
		{"wc_num", `{{ 'a 1 2.5'|wordcount }}`, "4"},
		{"wc_underscore", `{{ 'a_b-c d'|wordcount }}`, "3"},
		{"wc_unicode", `{{ 'ação, já!'|wordcount }}`, "2"},
		{"wc_nbsp", `{{ wn|wordcount }}`, "2"},
		{"wc_empty", `{{ ''|wordcount }}`, "0"},
		{"uz_basic", `{{ 'go http://x.com now'|urlize }}`, `go <a href="http://x.com" rel="noopener">http://x.com</a> now`},
		{"uz_dot", `{{ 'see http://x.com.'|urlize }}`, `see <a href="http://x.com" rel="noopener">http://x.com</a>.`},
		{"uz_paren", `{{ '(http://x.com)'|urlize }}`, `(<a href="http://x.com" rel="noopener">http://x.com</a>)`},
		{"uz_comma", `{{ 'a http://x.com, b'|urlize }}`, `a <a href="http://x.com" rel="noopener">http://x.com</a>, b`},
		{"uz_balanced", `{{ 'http://x.com/(a)'|urlize }}`, `<a href="http://x.com/(a)" rel="noopener">http://x.com/(a)</a>`},
		{"uz_balanced_twice", `{{ 'http://x.com/((a)b))'|urlize }}`, `<a href="http://x.com/((a)b))" rel="noopener">http://x.com/((a)b))</a>`},
		{"uz_balanced_outer", `{{ '(http://x.com/a(b))'|urlize }}`, `(<a href="http://x.com/a(b)" rel="noopener">http://x.com/a(b)</a>)`},
		{"uz_www", `{{ 'www.x.com'|urlize }}`, `<a href="https://www.x.com" rel="noopener">www.x.com</a>`},
		{"uz_email", `{{ 'mail a@b.com'|urlize }}`, `mail <a href="mailto:a@b.com">a@b.com</a>`},
		{"uz_email_dot", `{{ 'a@b.com.'|urlize }}`, `<a href="mailto:a@b.com">a@b.com</a>.`},
		{"uz_mailto", `{{ 'mailto:a@b.com'|urlize }}`, `<a href="mailto:a@b.com">a@b.com</a>`},
		{"uz_bare_domain", `{{ 'x.com'|urlize }}`, `x.com`},
		{"uz_ftp", `{{ 'ftp://x.com'|urlize }}`, `ftp://x.com`},
		{"uz_ip", `{{ 'http://1.2.3.4:80/x'|urlize }}`, `<a href="http://1.2.3.4:80/x" rel="noopener">http://1.2.3.4:80/x</a>`},
		{"uz_trim", `{{ 'http://example.com/abcdef'|urlize(10) }}`, `<a href="http://example.com/abcdef" rel="noopener">http://exa...</a>`},
		{"uz_trim_exact", `{{ 'http://x.com'|urlize(12) }}`, `<a href="http://x.com" rel="noopener">http://x.com</a>`},
		{"uz_trim_runes", `{{ 'http://ação.com/abcdef'|urlize(12) }}`, `<a href="http://ação.com/abcdef" rel="noopener">http://ação....</a>`},
		{"uz_no_default_trim", `{{ 'http://example.com/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa'|urlize }}`, `<a href="http://example.com/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" rel="noopener">http://example.com/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa</a>`},
		{"uz_nbsp", `{{ nb|urlize }}`, "see\u00a0<a href=\"http://x.com\" rel=\"noopener\">http://x.com</a>"},
		{"uz_line_sep", `{{ ls|urlize }}`, "x <a href=\"http://a.com\" rel=\"noopener\">http://a.com</a>\u2028y"},
		{"uz_nofollow", `{{ 'http://x.com'|urlize(nofollow=true) }}`, `<a href="http://x.com" rel="nofollow noopener">http://x.com</a>`},
		{"uz_target", `{{ 'http://x.com'|urlize(target='_blank') }}`, `<a href="http://x.com" rel="noopener" target="_blank">http://x.com</a>`},
		{"uz_target_escape", `{{ 'http://x.com'|urlize(target='"x') }}`, `<a href="http://x.com" rel="noopener" target="&#34;x">http://x.com</a>`},
		{"uz_target_empty", `{{ 'http://x.com'|urlize(target='') }}`, `<a href="http://x.com" rel="noopener">http://x.com</a>`},
		{"uz_rel_escape", `{{ 'http://x.com'|urlize(rel='a"b') }}`, `<a href="http://x.com" rel="a&#34;b noopener">http://x.com</a>`},
		{"uz_trim_bool", `{{ 'http://example.com/abcdef'|urlize(true) }}`, `<a href="http://example.com/abcdef" rel="noopener">h...</a>`},
		{"uz_nofollow_all", `{{ 'a http://x.com b http://y.com'|urlize(nofollow=true) }}`, `a <a href="http://x.com" rel="nofollow noopener">http://x.com</a> b <a href="http://y.com" rel="nofollow noopener">http://y.com</a>`},
		{"uz_rel", `{{ 'http://x.com'|urlize(rel='me') }}`, `<a href="http://x.com" rel="me noopener">http://x.com</a>`},
		{"uz_rel_sorted", `{{ 'http://x.com'|urlize(rel='zz') }}`, `<a href="http://x.com" rel="noopener zz">http://x.com</a>`},
		{"uz_rel_union", `{{ 'http://x.com'|urlize(rel='zz aa', nofollow=true) }}`, `<a href="http://x.com" rel="aa nofollow noopener zz">http://x.com</a>`},
		{"uz_query", `{{ 'https://x.com/a?b=1&c=2'|urlize }}`, `<a href="https://x.com/a?b=1&amp;c=2" rel="noopener">https://x.com/a?b=1&amp;c=2</a>`},
		{"uz_escape", `{{ 'go <b>http://x.com</b>'|urlize }}`, `go &lt;b&gt;http://x.com&lt;/b&gt;`},
		{"uz_angle", `{{ '<http://x.com>'|urlize }}`, `&lt;<a href="http://x.com" rel="noopener">http://x.com</a>&gt;`},
		{"uz_extra", `{{ 'tel:123'|urlize(extra_schemes=['tel:']) }}`, `<a href="tel:123" rel="noopener">tel:123</a>`},
		{"uz_target_filter", `{{ 'http://example.com/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa'|urlizetarget('_blank') }}`, `<a href="http://example.com/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" rel="noopener" target="_blank">http://example.com/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa</a>`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tmpl, err := env.FromString(tc.tmpl)
			var got string
			if err == nil {
				got, err = tmpl.Render(ctx)
			}
			if tc.want == "ERR" {
				if err == nil {
					t.Fatalf("expected an error, got %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("got  %q\nwant %q", got, tc.want)
			}
		})
	}
}
