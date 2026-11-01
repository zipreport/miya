package miya_test

import (
	"strings"
	"testing"

	miya "github.com/zipreport/miya"
)

func runWordwrapRows(t *testing.T, env *miya.Environment, ctx miya.Context, rows []struct{ name, tpl, want string }) {
	t.Helper()
	for _, tc := range rows {
		t.Run(tc.name, func(t *testing.T) {
			tmpl, err := env.FromString(tc.tpl)
			var got string
			if err == nil {
				got, err = tmpl.Render(ctx)
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

func TestWordwrapFilter(t *testing.T) {
	env := miya.NewEnvironment(miya.WithAutoEscape(false))
	ctx := miya.NewContextFrom(map[string]interface{}{
		"s":     "The quick brown fox jumps over the lazy dog and keeps running",
		"para":  "abc\n\ndef ghi",
		"t1":    "a\nb\n",
		"nl":    "\n",
		"vt":    "ab\vcd ef",
		"crlf":  "a\r\nb c",
		"long":  "The quick brown fox jumps over the lazy dog and keeps running past the old mill by the river",
		"w80":   "abcdefghi abcdefghi abcdefghi abcdefghi abcdefghi abcdefghi abcdefghi abcdefghij",
		"nbsp":  "foo   bar",
		"nnbsp": "foo   bar",
	})
	rows := []struct{ name, tpl, want string }{
		{"ww_long", `{{ 'abcdefgh ij'|wordwrap(3) }}`, "abc\ndef\ngh\nij"},
		{"ww_default", `{{ long|wordwrap }}`, "The quick brown fox jumps over the lazy dog and keeps running past the old mill\nby the river"},
		{"ww_default_edge", `{{ w80|wordwrap }}`, "abcdefghi abcdefghi abcdefghi abcdefghi abcdefghi abcdefghi abcdefghi\nabcdefghij"},
		{"ww_kw_default_width", `{{ w80|wordwrap(break_on_hyphens=false) }}`, "abcdefghi abcdefghi abcdefghi abcdefghi abcdefghi abcdefghi abcdefghi\nabcdefghij"},
		{"ww_10", `{{ s|wordwrap(10) }}`, "The quick\nbrown fox\njumps over\nthe lazy\ndog and\nkeeps\nrunning"},
		{"ww_nobreak", `{{ 'abcdefgh ij'|wordwrap(3, false) }}`, "abcdefgh\nij"},
		{"ww_hyphen", `{{ 'well-known co-operation x'|wordwrap(6) }}`, "well-\nknown\nco-ope\nration\nx"},
		{"ww_selfexp", `{{ 'self-explanatory'|wordwrap(8) }}`, "self-exp\nlanatory"},
		{"ww_nohyphen", `{{ 'well-known x'|wordwrap(6, true, nl, false) }}`, "well-k\nnown x"},
		{"ww_short_hyphen", `{{ 'a-b'|wordwrap(2) }}`, "a-\nb"},
		{"ww_hyphen_inner", `{{ 'a-bc-de'|wordwrap(4) }}`, "a-\nbc-\nde"},
		{"ww_hyphen_two_letters", `{{ 'x a-bcd'|wordwrap(6) }}`, "x\na-bcd"},
		{"ww_hyphen_refine", `{{ '1-2-3'|wordwrap(3) }}`, "1-\n2-3"},
		{"ww_hyphen_lead", `{{ '-abc-def'|wordwrap(4) }}`, "-abc\n-def"},
		{"ww_hyphen_digits", `{{ '1-2-3'|wordwrap(2) }}`, "1-\n2-\n3"},
		{"ww_emdash", `{{ 'abc--def ghi'|wordwrap(4) }}`, "abc\n--\ndef\nghi"},
		{"ww_spaces", `{{ 'abc  def'|wordwrap(20) }}`, "abc  def"},
		{"ww_ws_runs", `{{ '  x   y  '|wordwrap(3) }}`, "  x\ny"},
		{"ww_drop_ws", `{{ 'abc d'|wordwrap(3) }}`, "abc\nd"},
		{"ww_space_left", `{{ 'abc defgh'|wordwrap(4) }}`, "abc\ndefg\nh"},
		{"ww_letter_no", `{{ 'm²-based'|wordwrap(4) }}`, "m²-b\nased"},
		{"ww_letter_no_3", `{{ 'm²-based'|wordwrap(3) }}`, "m²-\nbas\ned"},
		{"ww_letter_no_6", `{{ 'x²-y²-z'|wordwrap(6) }}`, "x²-\ny²-z"},
		{"ww_unicode", `{{ 'ação ação'|wordwrap(4) }}`, "ação\nação"},
		{"ww_paragraphs", `{{ para|wordwrap(5) }}`, "abc\n\ndef\nghi"},
		{"ww_trailing_nl", `{{ t1|wordwrap(5) }}`, "a\nb"},
		{"ww_vt", `{{ vt|wordwrap(5) }}`, "ab\ncd ef"},
		{"ww_crlf", `{{ crlf|wordwrap(3) }}`, "a\nb c"},
		{"ww_wrapstr", `{{ s|wordwrap(20, true, '|') }}`, "The quick brown fox|jumps over the lazy|dog and keeps|running"},
		{"ww_kw", `{{ 'well-known x'|wordwrap(6, break_on_hyphens=false) }}`, "well-k\nnown x"},
		{"ww_kw_wrapstring", `{{ 'ab cd'|wordwrap(2, wrapstring='<br>') }}`, "ab<br>cd"},
		{"ww_lead", `[{{ '  lead'|wordwrap(20) }}]`, "[  lead]"},
		{"ww_zero", `{{ 'abc'|wordwrap(0) }}`, "ERR:invalid width 0 (must be > 0)"},
		{"ww_empty_zero", `{{ ''|wordwrap(0) }}`, ""},
		{"ww_nbsp", `{{ nbsp|wordwrap(4) }}`, "foo\n bar"},
		{"ww_nnbsp", `{{ nnbsp|wordwrap(4) }}`, "foo\n bar"},
	}
	runWordwrapRows(t, env, ctx, rows)
}

func TestWordwrapMarkup(t *testing.T) {
	env := miya.NewEnvironment(miya.WithAutoEscape(true))
	ctx := miya.NewContextFrom(map[string]interface{}{
		"lt":  "ab cd <i>",
		"amp": "a&b cd",
	})
	// Plan 027 is in the base: a safe input stays Markup through wordwrap.
	rows := []struct{ name, tpl, want string }{
		{"wwm_safe_wrapstring", `{{ lt|wordwrap(5, wrapstring='<br>'|safe) }}`, "ab cd<br>&lt;i&gt;"},
		{"wwm_plain_wrapstring", `{{ lt|wordwrap(5, wrapstring='<br>') }}`, "ab cd&lt;br&gt;&lt;i&gt;"},
		{"wwm_default", `{{ lt|wordwrap(5) }}`, "ab cd\n&lt;i&gt;"},
		// Jinja2: "ab cd\n&lt;i&gt;"
		{"deviation_wwm_safe_input", `{{ lt|safe|wordwrap(5) }}`, "ab cd\n<i>"},
		// Jinja2: "ab cd<br>&lt;i&gt;"
		{"deviation_wwm_safe_both", `{{ lt|safe|wordwrap(5, wrapstring='<br>'|safe) }}`, "ab cd<br><i>"},
		{"wwm_amp", `{{ amp|wordwrap(5, wrapstring='<br>'|safe) }}`, "a&amp;b<br>cd"},
	}
	runWordwrapRows(t, env, ctx, rows)
}
