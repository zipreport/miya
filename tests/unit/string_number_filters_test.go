package miya_test

import (
	"math"
	"testing"

	miya "github.com/zipreport/miya"
)

func TestStringNumberFilters(t *testing.T) {
	env := miya.NewEnvironment(miya.WithAutoEscape(false))
	ctx := miya.NewContextFrom(map[string]interface{}{
		"s":    "The quick brown fox jumps over the lazy dog and keeps running",
		"m":    "a\n\nb\n  c",
		"t1":   "a\nb\n",
		"crlf": "a\r\nb",
		"x":    2.5,
		"p":    2.675,
		"nx":   -2.5,
		"nz":   -0.004,
		"big":  1e20,
		"inf":  math.Inf(1),
		"nan":  math.NaN(),
		"neg":  float64(-123456),
		"negi": -123,
		"negf": -1234.5,
	})

	tests := []struct {
		name     string
		template string
		want     string
	}{
		{"tr_leeway_keeps", `{{ 'Hello World'|truncate(8) }}`, "Hello World"},
		{"tr_short", `{{ 'Hello World'|truncate(5) }}`, "He..."},
		{"tr_at_leeway", `{{ 'Hello World'|truncate(9) }}`, "Hello World"},
		{"tr_long", `{{ s|truncate(30) }}`, "The quick brown fox jumps..."},
		{"tr_long_kill", `{{ s|truncate(30, true) }}`, "The quick brown fox jumps o..."},
		{"tr_kill_keeps", `{{ 'Hello World'|truncate(8, true) }}`, "Hello World"},
		{"tr_end_keeps", `{{ 'Hello World'|truncate(8, false, '!') }}`, "Hello World"},
		{"tr_leeway0", `{{ 'hello world foo'|truncate(12, false, '...', 0) }}`, "hello..."},
		{"tr_leeway0_kw", `{{ 'abcdefgh'|truncate(5, leeway=0) }}`, "ab..."},
		{"tr_nospace", `{{ 'helloworldfoo'|truncate(8, false, '...', 0) }}`, "hello..."},
		{"tr_default", `{{ s|truncate|length }}`, "61"},
		{"tr_too_short", `{{ 'abcdef'|truncate(2) }}`, "ERR"},
		{"tr_neg_leeway", `{{ 'abc'|truncate(3, false, '...', -1) }}`, "ERR"},
		{"tr_unicode", `{{ 'ação ação ação'|truncate(8) }}`, "ação..."},
		{"tr_no_floor", `{{ 'A Supercalifragilistic order'|truncate(16) }}`, "A..."},
		{"in_default", `{{ m|indent }}`, "a\n\n    b\n      c"},
		{"in_2", `{{ m|indent(2) }}`, "a\n\n  b\n    c"},
		{"in_first", `{{ m|indent(2, true) }}`, "  a\n\n  b\n    c"},
		{"in_blank", `{{ m|indent(2, false, true) }}`, "a\n  \n  b\n    c"},
		{"in_str", `{{ m|indent('> ') }}`, "a\n\n> b\n>   c"},
		{"in_trail", `{{ t1|indent(2) }}`, "a\n  b\n"},
		{"in_crlf", `{{ crlf|indent(2) }}`, "a\n  b"},
		{"ce_default", `{{ 'x'|center|length }}`, "80"},
		{"ce_odd", `[{{ 'ab'|center(5) }}]`, "[  ab ]"},
		{"ce_even", `[{{ 'x'|center(4) }}]`, "[ x  ]"},
		{"ce_5", `[{{ 'x'|center(5) }}]`, "[  x  ]"},
		{"ro_2dp", `{{ 10|round(2) }}`, "10.00"},
		{"ro_2675", `{{ 2.675|round(2) }}`, "2.68"},
		{"ro_arith", `{{ (p|round(2)) + 1 }}`, "3.68"},
		{"ro_neg_zero", `{{ nz|round(2) }}`, "-0.00"},
		{"ro_half_away", `{{ x|round }}`, "3"},
		{"ro_35", `{{ 3.5|round }}`, "4"},
		{"ro_neg", `{{ nx|round }}`, "-3"},
		{"ro_1005", `{{ 1.005|round(2) }}`, "1.00"},
		{"ro_25_1", `{{ 2.5|round(1) }}`, "2.5"},
		{"ro_1dp_whole", `{{ 70|round(1) }}`, "70"},
		{"ro_float_type", `{{ (2.567|round(1)) + 1 }}`, "3.6"},
		{"ro_compare", `{{ p|round(2) > 2 }}`, "true"},
		{"ro_is_number", `{{ p|round(2) is number }}`, "true"},
		{"ro_eq", `{{ p|round(2) == 2.68 }}`, "true"},
		{"ro_falsy", `{% if 0|round(2) %}t{% else %}f{% endif %}`, "f"},
		{"ro_int", `{{ p|round(2)|int }}`, "2"},
		{"ro_float", `{{ p|round(2)|float }}`, "2.68"},
		{"ro_tojson", `{{ p|round(2)|tojson }}`, "2.68"},
		{"ro_join", `{{ [p|round(2), 10|round(2)]|join(',') }}`, "2.68,10.00"},
		{"ro_concat", `{{ 'p' ~ 10|round(2) }}`, "p10.00"},
		{"ro_format_s", `{{ '%s|%.1f'|format(10|round(2), p|round(2)) }}`, "10.00|2.7"},
		{"ro_test_ge", `{{ p|round(2) is ge(2.68) }}`, "true"},
		{"ro_test_divisible", `{{ 10|round(2) is divisibleby(5) }}`, "true"},
		{"ro_ceil", `{{ 2.1|round(0, 'ceil') }}`, "3"},
		{"ro_ceil2", `{{ 2.001|round(2, 'ceil') }}`, "2.01"},
		{"ro_floor1", `{{ 2.97|round(1, 'floor') }}`, "2.9"},
		{"ro_bad_method", `{{ 2.5|round(0, 'bogus') }}`, "ERR"},
		{"ro_neg_prec", `{{ 1234|round(-2) }}`, "1200"},
		{"ro_neg_prec_half", `{{ 1250.0|round(-2) }}`, "1300"},
		{"ro_prec_cap", `{{ 1.5|round(400) }}`, "1.5"},
		{"ro_prec_cap_320", `{{ 1.5|round(320) }}`, "1.5"},
		{"ro_scaled_inf", `{{ big|round(300) }}`, "1e+20"},
		{"ro_big_exact", `{{ big|round(3) }}`, "100000000000000000000.000"},
		{"ro_inf", `{{ inf|round(2) }}`, "+Inf"},
		{"ro_prec_low", `{{ 5|round(-400) }}`, "0"},
		{"ro_default_falsy", `{{ 0|round(2)|default('-', true) }}`, "-"},
		{"ro_select", `{{ [0|round(2), 1|round(2)]|select|join(',') }}`, "1.00"},
		{"ro_range", `{{ range(3|round(2))|list|length }}`, "3"},
		{"ro_is_empty", `{{ 0|round(2) is empty }}`, "true"},
		{"ro_kw", `{{ 2.567|round(precision=1) }}`, "2.6"},
		{"deviation_round_string", `{{ '2.5'|round }}`, "3"},
		{"int_37", `{{ '3.7'|int }}`, "3"},
		{"int_ws", `{{ ' 7 '|int }}`, "7"},
		{"int_neg", `{{ '-3.7'|int }}`, "-3"},
		{"int_hex_prefix", `{{ '0x1A'|int(0, 16) }}`, "26"},
		{"int_hex_ws", `{{ ' 0x1A '|int(0, 16) }}`, "26"},
		{"int_hex_us", `{{ '0x_1f'|int(0, 16) }}`, "31"},
		{"int_base16", `{{ 'ff'|int(0, 16) }}`, "255"},
		{"int_base0", `{{ '0x1A'|int(0, 0) }}`, "26"},
		{"int_exp", `{{ '1e3'|int }}`, "1000"},
		{"int_us", `{{ '1_000'|int }}`, "1000"},
		{"int_us_bad", `{{ '1_'|int }}`, "0"},
		{"int_bad", `{{ 'x'|int }}`, "0"},
		{"int_bad_def", `{{ 'x'|int(7) }}`, "7"},
		{"int_leading_zero", `{{ '010'|int(0, 0) }}`, "10"},
		{"int_inf", `{{ 'inf'|int }}`, "0"},
		{"int_float", `{{ 3.7|int }}`, "3"},
		{"int_bad_base", `{{ '5'|int(0, 1) }}`, "5"},
		{"int_float_bad_base", `{{ 3.7|int(0, 1) }}`, "3"},
		{"int_nbsp", "{{ ' 7'|int }}", "7"},
		{"int_x1c", "{{ '\x1c7'|int }}", "0"},
		{"int_exp_overflow", `{{ '1e400'|int }}`, "0"},
		{"deviation_int_overflow", `{{ '12345678901234567890'|int }}`, "0"},
		{"deviation_int_float_overflow", `{{ big|int }}`, "0"},
		{"deviation_int_float_inf", `{{ inf|int }}`, "0"},
		{"fl_ws", `{{ ' 2 '|float }}`, "2"},
		{"fl_us", `{{ '1_0.5'|float }}`, "10.5"},
		{"fl_bad_us", `{{ '1._5'|float }}`, "0"},
		{"fl_hex", `{{ '0x10'|float }}`, "0"},
		{"fl_hexfloat", `{{ '0x1p4'|float }}`, "0"},
		{"fl_str", `{{ '3.5'|float }}`, "3.5"},
		{"fl_bad_def", `{{ 'x'|float(1.5) }}`, "1.5"},
		{"fl_overflow", `{{ '1e400'|float }}`, "+Inf"},
		{"fs_1000", `{{ 1000|filesizeformat }}`, "1.0 kB"},
		{"fs_1", `{{ 1|filesizeformat }}`, "1 Byte"},
		{"fs_0", `{{ 0|filesizeformat }}`, "0 Bytes"},
		{"fs_999", `{{ 999|filesizeformat }}`, "999 Bytes"},
		{"fs_frac", `{{ 1.5|filesizeformat }}`, "1 Bytes"},
		{"fs_1500", `{{ 1500|filesizeformat }}`, "1.5 kB"},
		{"fs_mb", `{{ 1500000|filesizeformat }}`, "1.5 MB"},
		{"fs_bin", `{{ 2048|filesizeformat(true) }}`, "2.0 KiB"},
		{"fs_bin_mib", `{{ 1000000000|filesizeformat(true) }}`, "953.7 MiB"},
		{"fs_huge", `{{ (10 ** 30)|filesizeformat }}`, "1000000.0 YB"},
		{"fs_neg", `{{ -5|filesizeformat }}`, "-5 Bytes"},
		{"fs_str", `{{ '2000'|filesizeformat }}`, "2.0 kB"},
		{"fn_neg", `{{ neg|format_number }}`, "-123,456"},
		{"fn_neg_small", `{{ negi|format_number }}`, "-123"},
		{"fn_neg_dec", `{{ negf|format_number(2) }}`, "-1,234.50"},
		{"cu_neg", `{{ neg|currency }}`, "$-123,456.00"},
		{"cu_neg_zero", `{{ nz|currency }}`, "$-0.00"},
		{"cu_inf", `{{ (big ** 20)|currency }}`, "ERR"},
		{"cu_nan", `{{ nan|currency }}`, "ERR"},
		{"fn_inf", `{{ inf|format_number }}`, "ERR"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("panic: %v", r)
				}
			}()
			got, err := env.RenderString(tc.template, ctx)
			if tc.want == "ERR" {
				if err == nil {
					t.Errorf("expected error, got %q", got)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}
