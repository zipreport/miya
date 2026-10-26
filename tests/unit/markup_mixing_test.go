package miya_test

import (
	"fmt"
	"strings"
	"testing"

	miya "github.com/zipreport/miya"
	"github.com/zipreport/miya/runtime"
)

type mixSt int

func (mixSt) String() string { return "<img onerror=x>" }

type mixSp int

func (mixSp) String() string { return " onerror=x " }

type mixFt int

func (mixFt) Format(f fmt.State, _ rune) { _, _ = f.Write([]byte("<img onerror=x>")) }

func mixCtx() miya.Context {
	M := func(s string) runtime.Markup { return runtime.Markup(s) }
	return miya.NewContextFrom(map[string]interface{}{
		"x": "<x>", "q": "it's", "u": "é<",
		"m": M("<b>"), "e": "<s>", "sm": M("<s>"), "card": M("<img src=/a.png alt=photo>"),
		"sep": " onerror=alert(1) ", "br": M("<br>"),
		"fs": M("%s"), "fr": M("%r"), "fa": M("%a"), "fc": M("%c"), "fw": M("%5s|"),
		"ff": M("%.2f"), "fx": M("%x"), "fd": M("%d"), "fka": M("%(a)s"), "fkr": M("%(a)r"),
		"fp": M("%.2s"), "fl": M("%-6s|"),
		"n60": 60, "lt": "<", "f25": 2.5, "n255": 255, "n3": 3,
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
		{"r21", `{{ mab|wordwrap(1, wrapstring=sep) }}`, `a onerror=alert(1) b`},
		{"r22", `{{ pab|wordwrap(1, wrapstring=br) }}`, `a<br>b`},
		{"deviation_safe_input_safe_wrapstring", `{{ miab|wordwrap(1, wrapstring=br) }}`, `<<br>i<br>><br>a<br>b`},
		{"r24", `{{ mab|wordwrap(10) }}`, `a b`},
		{"s1", `{{ fs|format(st) }}`, `&lt;img onerror=x&gt;`},
		{"s2", `{{ fd|format(st) }}`, `1`},
		{"s3", `{{ bx|replace('X', st) }}`, `<b>&lt;img onerror=x&gt;`},
		{"s4", `{{ [m,'c']|join(st) }}`, `<b>&lt;img onerror=x&gt;c`},
		{"s5", `{{ st }}`, `&lt;img onerror=x&gt;`},
		{"s5_formatter", `{{ bx|replace('X', ft) }}`, `<b>&lt;img onerror=x&gt;`},
		{"s6_wordwrap_stringer", `{{ mab|wordwrap(1, wrapstring=sp) }}`, `a onerror=x b`},
		{"kw_truncate", `{{ mb20|truncate(12, true, end=e) }}`, `<b>xxxxxx&lt;s&gt;`},
		{"precision_escapes_first", `{{ fp|format(x) }}`, `&l`},
		{"width_after_escape", `{{ fl|format(lt) }}`, `&lt;  |`},
		{"fb11_guard", `{% filter wordwrap(80) %}<p>{{ x }}</p>{% endfilter %}`, `<p>&lt;x&gt;</p>`},
		{"fb_wordwrap_newline", `{% filter wordwrap(72, true, '\n') %}<p>{{ x }}</p>{% endfilter %}`, `<p>&lt;x&gt;</p>`},
		{"wordwrap_newline_safe", `{{ card|wordwrap(5, false, '\n') }}`, "<img\nsrc=/a.png\nalt=photo>"},
		{"deviation_fb_wordwrap_br", `{% filter wordwrap(72, true, '<br>') %}<p>{{ x }}</p>{% endfilter %}`, `&lt;p&gt;&amp;lt;x&amp;gt;&lt;/p&gt;`},
		{"setblock_length", `{% set n %}20{% endset %}{{ t|truncate(n) }}`, strings.Repeat("y", 17) + `...`},
		{"setblock_width", `{% set n %}5{% endset %}{{ m|wordwrap(n) }}`, `<b>`},
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
