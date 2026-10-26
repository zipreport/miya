package miya_test

import (
	"math"
	"strings"
	"testing"

	miya "github.com/zipreport/miya"
)

func TestFormatFilter(t *testing.T) {
	env := miya.NewEnvironment(miya.WithAutoEscape(false))
	ctx := miya.NewContextFrom(map[string]interface{}{
		"n": float64(10023), "p": 2.675, "tiny": 1e-10, "inf": math.Inf(1), "nan": math.NaN(),
		"wf": 255.0, "w8": 8.0, "w65": 65.0, "w5": 5.0, "w2": 2.0,
		"ap": "it's", "nls": "a\nb", "q": "a\"b'c", "rc": "a\x01\tb\\",
		"big": 1e20, "np": "a ​ \u0080\u009fzé\u007fÿ",
	})

	tests := []struct {
		name string
		tpl  string
		want string
	}{
		{"fo_d", `{{ '%d'|format(n) }}`, "10023"},
		{"fo_ss", `{{ '%s-%s'|format(1, 2) }}`, "1-2"},
		{"fo_2f", `{{ '%.2f'|format(3.14159) }}`, "3.14"},
		{"fo_f_int", `{{ '%.2f'|format(10) }}`, "10.00"},
		{"fo_d_float", `{{ '%d'|format(3.9) }}`, "3"},
		{"fo_s_float", `{{ '%s'|format(2.5) }}`, "2.5"},
		{"fo_05d", `{{ '%05d'|format(42) }}`, "00042"},
		{"fo_02d", `{{ '%02d'|format(3) }}`, "03"},
		{"fo_pct", `{{ '%d%%'|format(5) }}`, "5%"},
		{"fo_x", `{{ '%x'|format(255) }}`, "ff"},
		{"fo_flags", `{{ '%5.1f|%-5d|%+d|% d'|format(3.14159, 42, 5, 7) }}`, "  3.1|42   |+5| 7"},
		{"fo_alt", `{{ '%#x %#o %X'|format(255, 8, 255) }}`, "0xff 0o10 FF"},
		{"fo_alt_zero_x", `{{ '%#08x'|format(255) }}`, "0x0000ff"},
		{"fo_alt_zero_o", `{{ '%#08o'|format(8) }}`, "0o000010"},
		{"fo_alt_f_g", `{{ '%#.0f|%#g'|format(3.0, 3.0) }}`, "3.|3.00000"},
		{"fo_c", `{{ '%c'|format(65) }}`, "A"},
		{"fo_c_str", `{{ '%c'|format('a') }}`, "a"},
		{"fo_c_long", `{{ '%c'|format('ab') }}`, "ERR:%c requires an int or a unicode character"},
		{"fo_r", `{{ '%r'|format('a') }}`, "'a'"},
		{"fo_r_apos", `{{ '%r'|format(ap) }}`, "\"it's\""},
		{"fo_r_nl", `{{ '%r'|format(nls) }}`, "'a\\nb'"},
		{"fo_r_quotes", `{{ '%r'|format(q) }}`, "'a\"b\\'c'"},
		{"fo_r_ctrl", `{{ '%r'|format(rc) }}`, "'a\\x01\\tb\\\\'"},
		{"fo_a", `{{ '%a'|format('é') }}`, "'\\xe9'"},
		{"fo_g", `{{ '%g %g %g'|format(0.0001, 123456789.0, 100000.0) }}`, "0.0001 1.23457e+08 100000"},
		{"fo_G", `{{ '%G'|format(tiny) }}`, "1E-10"},
		{"fo_e", `{{ '%e'|format(12345.678) }}`, "1.234568e+04"},
		{"fo_e_flags", `{{ '%+.2e'|format(12345.678) }}`, "+1.23e+04"},
		{"fo_inf", `{{ '%f|%F|%e'|format(inf, inf, inf) }}`, "inf|INF|inf"},
		{"fo_inf_flags", `{{ '%+f|%05f|%-6f|'|format(inf, inf, nan) }}`, "+inf|00inf|nan   |"},
		{"fo_nan", `{{ '%f|%g'|format(nan, nan) }}`, "nan|nan"},
		{"fo_d_inf", `{{ '%d'|format(inf) }}`, "ERR:cannot convert float infinity to integer"},
		{"fo_d_nan", `{{ '%d'|format(nan) }}`, "ERR:cannot convert float NaN to integer"},
		{"fo_x_frac", `{{ '%x'|format(2.5) }}`, "ERR:%x format: an integer is required, not float"},
		{"fo_iu", `{{ '%i|%u'|format(3.7, 2) }}`, "3|2"},
		{"fo_d_bool", `{{ '%d'|format(true) }}`, "1"},
		{"fo_len_mod", `{{ '%ld|%hd'|format(5, 6) }}`, "5|6"},
		{"fo_left_plus", `{{ '%-+6d|'|format(5) }}`, "+5    |"},
		{"fo_space_zero", `{{ '% 05d'|format(42) }}`, " 0042"},
		{"fo_star", `{{ '%*d'|format(5, 3) }}`, "    3"},
		{"fo_star_neg", `{{ '%*d|'|format(-5, 3) }}`, "3    |"},
		{"fo_star_missing", `{{ '%*d'|format(5) }}`, "ERR:not enough arguments for format string"},
		{"fo_prec_s", `{{ '%.3s'|format('abcdef') }}`, "abc"},
		{"fo_prec_s2", `{{ '%.2s|%5.1s|'|format('abc', 'xyz') }}`, "ab|    x|"},
		{"fo_0s", `{{ '%05s'|format('ab') }}`, "   ab"},
		{"fo_unicode_width", `{{ '%5s|%-5s|'|format('ção', 'ção') }}`, "  ção|ção  |"},
		{"fo_kw", `{{ '%(a)s-%(b)d'|format(a=1, b=2) }}`, "1-2"},
		{"fo_kw_pct", `{{ '%(a)s %%'|format(a=1) }}`, "1 %"},
		{"fo_kw_unused", `{{ 'x'|format(a=1) }}`, "x"},
		{"fo_kw_missing", `{{ '%(b)s'|format(a=1) }}`, "ERR:'b'"},
		{"fo_kw_few", `{{ '%s %s'|format(a=1) }}`, "ERR:not enough arguments for format string"},
		{"fo_key_incomplete", `{{ '%(a'|format(a=1) }}`, "ERR:incomplete format key"},
		{"fo_needs_mapping", `{{ '%(a)s'|format(1) }}`, "ERR:format requires a mapping"},
		{"fo_plain", `{{ 'abc'|format }}`, "abc"},
		{"fo_few", `{{ '%s %s'|format(1) }}`, "ERR:not enough arguments for format string"},
		{"fo_many", `{{ '%s'|format('a', 'b') }}`, "ERR:not all arguments converted during string formatting"},
		{"fo_d_str", `{{ '%d'|format('3') }}`, "ERR:%d format: a real number is required, not str"},
		{"fo_incomplete", `{{ '100%'|format }}`, "ERR:incomplete format"},
		{"fo_unknown", `{{ '%q'|format(1) }}`, "ERR:unsupported format character 'q' (0x71) at index 1"},
		{"fo_big_d", `{{ '%d|%d'|format(big, -big) }}`, "100000000000000000000|-100000000000000000000"},
		{"fo_prec0_zero", `{{ '%.0d|%#.0x|%#.0o'|format(0, 0, 0) }}`, "0|0x0|0o0"},
		{"fo_zero_prec_d", `{{ '%08.3d'|format(5) }}`, "00000005"},
		{"fo_zero_c_r_a", `{{ '%05c|%05r|%05a'|format('a', 'a', 'a') }}`, "    a|  'a'|  'a'"},
		{"fo_star_prec_neg", `{{ '%.*f|'|format(-2, 3.14159) }}`, "3|"},
		{"fo_r_nonprint", `{{ '%r'|format(np) }}`, "'a\\xa0\\u200b\\u2028\\x80\\x9f\\ue000zé\\x7fÿ'"},
		{"fo_c_range", `{{ '%c'|format(1114112) }}`, "ERR:%c arg not in range(0x110000)"},
		{"fo_go_O", `{{ '%O'|format(8) }}`, "ERR:unsupported format character 'O' (0x4f) at index 1"},
		{"fo_pct_width", `{{ '%5%'|format(1) }}`, "ERR:unsupported format character '%' (0x25) at index 2"},
		{"fo_kw_then_unnamed", `{{ '%(a)s %s'|format(a=1) }}`, "ERR:not enough arguments for format string"},
		{"deviation_fo_big_x", `{{ '%x'|format(big) }}`, "56bc75e2d63100000"},
		{"deviation_fo_int_whole_float", `{{ '%x|%o|%X|%c'|format(wf, w8, wf, w65) }}`, "ff|10|FF|A"},
		{"deviation_fo_star_float", `{{ '%*d|%.*f'|format(w5, 3, w2, 3.14159) }}`, "    3|3.14"},
		{"deviation_fo_s_none", `{{ '%s'|format(none) }}`, ""},
		{"deviation_fo_s_bool", `{{ '%s'|format(true) }}`, "true"},
		{"deviation_fo_rounded", `{{ '%s|%r|%.1f|%d'|format(10|round(2), 10|round(2), p|round(2), p|round(2)) }}`, "10.00|10.00|2.7|2"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tmpl, err := env.FromString(tc.tpl)
			if err == nil {
				var got string
				got, err = tmpl.Render(ctx)
				if !strings.HasPrefix(tc.want, "ERR:") {
					if err != nil {
						t.Fatalf("unexpected error: %v", err)
					}
					if got != tc.want {
						t.Errorf("got %q, want %q", got, tc.want)
					}
					return
				}
			}
			if !strings.HasPrefix(tc.want, "ERR:") {
				t.Fatalf("unexpected error: %v", err)
			}
			if err == nil {
				t.Fatalf("expected error containing %q", tc.want[4:])
			}
			if !strings.Contains(err.Error(), tc.want[4:]) {
				t.Errorf("error %q does not contain %q", err.Error(), tc.want[4:])
			}
		})
	}
}

func TestFormatFilterMarkup(t *testing.T) {
	env := miya.NewEnvironment(miya.WithAutoEscape(true))
	ctx := miya.NewContextFrom(map[string]interface{}{
		"fs": "<b>%s</b>", "fk": "<b>%(n)s</b>", "lt": "<i>", "p": 2.675,
	})

	tests := []struct {
		name string
		tpl  string
		want string
	}{
		{"fm_plain", `{{ fs|format(lt) }}`, "&lt;b&gt;&lt;i&gt;&lt;/b&gt;"},
		{"fm_safe", `{{ fs|safe|format(lt) }}`, "<b>&lt;i&gt;</b>"},
		{"fm_safe_arg", `{{ fs|safe|format(lt|safe) }}`, "<b><i></b>"},
		{"fm_kw", `{{ fk|safe|format(n=lt) }}`, "<b>&lt;i&gt;</b>"},
		{"fm_kw_safe_arg", `{{ fk|safe|format(n=lt|safe) }}`, "<b><i></b>"},
		{"fm_kw_plain", `{{ fk|format(n=lt) }}`, "&lt;b&gt;&lt;i&gt;&lt;/b&gt;"},
		{"fm_rounded", `{{ ('%.1f'|safe)|format(p|round(2)) }}`, "2.7"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tmpl, err := env.FromString(tc.tpl)
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			got, err := tmpl.Render(ctx)
			if err != nil {
				t.Fatalf("render: %v", err)
			}
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}
