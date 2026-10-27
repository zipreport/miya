package miya_test

import (
	"math"
	"strings"
	"testing"
	"time"

	miya "github.com/zipreport/miya"
	"github.com/zipreport/miya/runtime"
)

type pvUUID [16]byte

func (u pvUUID) String() string { return "uuid-str" }

type pvErr struct{}

func (pvErr) Error() string { return "boom" }

type pvSecret struct{ secret string }

type pvCase struct {
	name       string
	tpl        string
	data       map[string]interface{}
	autoescape bool
	want       string
	errHas     string
	repeat     int
}

func runPvCases(t *testing.T, cases []pvCase) {
	t.Helper()
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			n := tc.repeat
			if n == 0 {
				n = 1
			}
			for i := 0; i < n; i++ {
				env := miya.NewEnvironment(miya.WithAutoEscape(tc.autoescape))
				got, err := renderGo(t, env, tc.tpl, tc.data)
				if tc.errHas != "" {
					if err == nil || !strings.Contains(err.Error(), tc.errHas) {
						t.Fatalf("want error containing %q, got err=%v out=%q", tc.errHas, err, got)
					}
					continue
				}
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got != tc.want {
					t.Fatalf("run %d: got %q, want %q", i, got, tc.want)
				}
			}
		})
	}
}

type pvD = map[string]interface{}

func TestPythonRepr(t *testing.T) {
	s := "x"
	runPvCases(t, []pvCase{
		{name: "list_strings", tpl: `{{ ['a', 'b'] }}`, want: `['a', 'b']`},
		{name: "list_mixed", tpl: `{{ [1, 2.5, true, none] }}`, want: `[1, 2.5, True, None]`},
		{name: "list_nested_data", tpl: `{{ l }}`, data: pvD{"l": []interface{}{1, "x", []interface{}{2, 3}}}, want: `[1, 'x', [2, 3]]`},
		{name: "dict_data", tpl: `{{ d }}`, data: pvD{"d": map[string]interface{}{"k": 1, "j": "v"}}, want: `{'j': 'v', 'k': 1}`, repeat: 20},
		{name: "dict_nested_none", tpl: `{{ d }}`, data: pvD{"d": map[string]interface{}{"k": []interface{}{1, map[string]interface{}{"z": nil}}}}, want: `{'k': [1, {'z': None}]}`},
		{name: "empty", tpl: `{{ [] }}|{% set d = dict() %}{{ d }}`, want: `[]|{}`},
		{name: "quote_double", tpl: `{{ ["it's"] }}`, want: `["it's"]`},
		{name: "quote_both", tpl: `{{ l }}`, data: pvD{"l": []interface{}{`a"b'c`}}, want: `['a"b\'c']`},
		{name: "escapes", tpl: `{{ l }}`, data: pvD{"l": []interface{}{"x\ny", `a\b`, "tab\there"}}, want: `['x\ny', 'a\\b', 'tab\there']`},
		{name: "list_filter", tpl: `{{ l|list }}`, data: pvD{"l": []interface{}{float64(1), float64(2)}}, want: `[1, 2]`},
		{name: "string_filter", tpl: `{{ [1, 2]|string }}`, want: `[1, 2]`},
		{name: "concat", tpl: `{{ ['a'] ~ '' }}`, want: `['a']`},
		{name: "join_nested", tpl: `{{ [[1, 2], [3]]|join(';') }}`, want: `[1, 2];[3]`},
		{name: "typed_slice", tpl: `{{ l }}`, data: pvD{"l": []int{3, 1}}, want: `[3, 1]`},
		{name: "typed_map", tpl: `{{ m }}`, data: pvD{"m": map[string]int{"b": 2, "a": 1}}, want: `{'a': 1, 'b': 2}`, repeat: 20},
		{name: "int_key_map", tpl: `{{ m }}`, data: pvD{"m": map[int]string{10: "a", 2: "b"}}, want: `{2: 'b', 10: 'a'}`},
		{name: "ptr_in_list", tpl: `{{ l }}`, data: pvD{"l": []interface{}{&s, (*string)(nil)}}, want: `['x', None]`},
		{name: "whole_float", tpl: `{{ [1.0, 0.1] }}`, want: `[1, 0.1]`},
		{name: "top_level_unchanged", tpl: `{{ true }}|{{ none }}|{{ 1.0 }}`, want: `true||1`},
		{name: "autoescape_list", tpl: `{{ ['<'] }}`, autoescape: true, want: `[&#39;&lt;&#39;]`},
		{name: "autoescape_data", tpl: `{{ items }}`, data: pvD{"items": []interface{}{"<i>"}}, autoescape: true, want: `[&#39;&lt;i&gt;&#39;]`},
		{name: "autoescape_quotes", tpl: `{{ ["it's", 'q"'] }}`, autoescape: true, want: `[&#34;it&#39;s&#34;, &#39;q&#34;&#39;]`},
		{name: "stringer_top", tpl: `{{ u }}`, data: pvD{"u": pvUUID{1, 2}}, want: `uuid-str`},
		{name: "stringer_in_list", tpl: `{{ [u] }}|{{ l }}`, data: pvD{"u": pvUUID{1, 2}, "l": []pvUUID{{1}}}, want: `[uuid-str]|[uuid-str]`},
		{name: "stringer_set", tpl: `{% set v = u %}{{ v }}|{{ v == u }}`, data: pvD{"u": pvUUID{1, 2}}, want: `uuid-str|true`},
		{name: "error_value", tpl: `{{ e }}|{{ [e] }}`, data: pvD{"e": pvErr{}}, want: `boom|[boom]`},
		{name: "bytes_top", tpl: `{{ b }}`, data: pvD{"b": []byte("hi")}, want: `hi`},
		{name: "bytes_in_list", tpl: `{{ [b] }}`, data: pvD{"b": []byte("hi")}, want: `['hi']`},
		{name: "inf_nan", tpl: `{{ l }}`, data: pvD{"l": []interface{}{math.Inf(1), math.Inf(-1), math.NaN()}}, want: `[inf, -inf, nan]`},
		{name: "inf_top_unchanged", tpl: `{{ x }}`, data: pvD{"x": math.Inf(1)}, want: `+Inf`},
	})
}

func TestStringMethods(t *testing.T) {
	m := map[string]interface{}{"k": "v"}
	runPvCases(t, []pvCase{
		{name: "upper", tpl: `{{ 'abc'.upper() }}`, want: `ABC`},
		{name: "upper_unicode", tpl: `{{ s.upper() }}`, data: pvD{"s": "ça"}, want: "ÇA"},
		{name: "lower", tpl: `{{ 'ABC'.lower() }}`, want: `abc`},
		{name: "strip", tpl: `{{ '  a  '.strip() }}|{{ 'xxaxx'.strip('x') }}|{{ ' a '.lstrip() }}|{{ ' a '.rstrip() }}`, want: `a|a|a | a`},
		{name: "split_sep", tpl: `{{ 'a,b,c'.split(',') }}`, want: `['a', 'b', 'c']`},
		{name: "split_ws", tpl: `{{ s.split() }}`, data: pvD{"s": "  a  b  "}, want: `['a', 'b']`},
		{name: "split_max", tpl: `{{ 'a,b,c'.split(',', 1) }}`, want: `['a', 'b,c']`},
		{name: "split_ws_max", tpl: `{{ 'a b  c '.split(none, 1) }}`, want: `['a', 'b  c ']`},
		{name: "split_trailing", tpl: `{{ 'a,b,'.split(',') }}`, want: `['a', 'b', '']`},
		{name: "split_empty", tpl: `{{ ''.split(',') }}|{{ ''.split() }}`, want: `['']|[]`},
		{name: "split_empty_sep", tpl: `{{ 'abc'.split('') }}`, errHas: "empty separator"},
		{name: "replace", tpl: `{{ 'hello'.replace('l', 'L') }}|{{ 'hello'.replace('l', 'L', 1) }}`, want: `heLLo|heLlo`},
		{name: "starts_ends", tpl: `{{ 'hello'.startswith('he') }}|{{ 'hello'.endswith('lo') }}|{{ 'hello'.startswith('x') }}`, want: `true|true|false`},
		{name: "format_auto", tpl: `{{ '{} and {}'.format('a', 'b') }}`, want: `a and b`},
		{name: "format_index", tpl: `{{ '{0}{1}{0}'.format('a', 'b') }}`, want: `aba`},
		{name: "format_named", tpl: `{{ '{n}!'.format(n=3) }}`, want: `3!`},
		{name: "format_braces", tpl: `{{ f.format('x') }}`, data: pvD{"f": "{{}} {}"}, want: `{} x`},
		{name: "format_spec_rejected", tpl: `{{ '{:>5}'.format('a') }}`, errHas: "format spec"},
		{name: "format_mixed_rejected", tpl: `{{ '{0} {}'.format('a', 'b') }}`, errHas: "field numbering"},
		{name: "title", tpl: `{{ s.title() }}`, data: pvD{"s": "they're 1st"}, want: `They'Re 1St`},
		{name: "capitalize", tpl: `{{ 'hELLO wORLD'.capitalize() }}`, want: `Hello world`},
		{name: "join", tpl: `{{ ', '.join(['a', 'b']) }}|{{ '-'.join('abc') }}`, want: `a, b|a-b-c`},
		{name: "count", tpl: `{{ 'banana'.count('a') }}`, want: `3`},
		{name: "find", tpl: `{{ 'banana'.find('n') }}|{{ 'banana'.find('z') }}|{{ s.find('ú') }}`, data: pvD{"s": "açúcar"}, want: `2|-1|2`},
		{name: "named_string", tpl: `{{ st.upper() }}`, data: pvD{"st": gvStatus("ok")}, want: `OK`},
		{name: "unknown_method", tpl: `{{ 'abc'.nosuch() }}`, errHas: "cannot call"},
		{name: "method_value_print", tpl: `{{ 'x'.upper }}`, want: `<built-in method upper>`},
		{name: "format_item", tpl: `{{ '{0[k]}'.format(m) }}|{{ '{n[k]}'.format(n=m) }}`, data: pvD{"m": m}, want: `v|v`},
		{name: "format_attr_dict", tpl: `{{ '{0.k}'.format(m) }}`, data: pvD{"m": m}, want: `v`},
		{name: "format_conversions", tpl: `{{ '{!r}'.format('a') }}|{{ '{0!r}'.format(l) }}|{{ '{!s}'.format('a') }}`, data: pvD{"l": []interface{}{float64(1), "b"}}, want: `'a'|[1, 'b']|a`},
		{name: "format_index_range", tpl: `{{ '{2}'.format('a') }}`, errHas: "out of range"},
		{name: "format_missing_name", tpl: `{{ '{n}'.format() }}`, errHas: "KeyError: 'n'"},
		{name: "format_missing_key", tpl: `{{ '{0[z]}'.format(m) }}`, data: pvD{"m": m}, errHas: "KeyError: 'z'"},
		{name: "format_unreachable", tpl: `{{ '{0.secret}'.format(u) }}`, data: pvD{"u": pvSecret{"s3"}}, errHas: "has no attribute"},
		{name: "replace_too_large", tpl: `{{ big.replace('', big) }}`, data: pvD{"big": strings.Repeat("a", 10000)}, errHas: "too large"},
		{name: "join_too_large", tpl: `{{ big.join(l) }}`, data: pvD{"big": strings.Repeat("a", 1<<20), "l": repeatStrings("x", 100)}, errHas: "too large"},
	})
}

func repeatStrings(s string, n int) []interface{} {
	out := make([]interface{}, n)
	for i := range out {
		out[i] = s
	}
	return out
}

func TestSafeReceiverMethods(t *testing.T) {
	x := pvD{"x": "<i>"}
	runPvCases(t, []pvCase{
		{name: "upper", tpl: `{{ (x|safe).upper() }}`, data: x, autoescape: true, want: `<I>`},
		{name: "format_escapes_arg", tpl: `{{ ('<b>{}</b>'|safe).format(x) }}`, data: x, autoescape: true, want: `<b>&lt;i&gt;</b>`},
		{name: "replace_escapes_arg", tpl: `{{ ('<b>'|safe).replace('b', x) }}`, data: x, autoescape: true, want: `<&lt;i&gt;>`},
		{name: "join_escapes_items", tpl: `{{ ('<br>'|safe).join([x, 'a']) }}`, data: x, autoescape: true, want: `&lt;i&gt;<br>a`},
		{name: "strip_title", tpl: `{{ ('<b> '|safe).strip() }}|{{ ('<b>'|safe).title() }}`, data: x, autoescape: true, want: `<b>|<B>`},
		{name: "plain_results", tpl: `{{ ('<b>'|safe).startswith('<') }}|{{ ('<b>'|safe).count('b') }}`, data: x, autoescape: true, want: `true|1`},
		{name: "plain_receiver", tpl: `{{ '<b>{}</b>'.format(x) }}`, data: x, autoescape: true, want: `&lt;b&gt;&lt;i&gt;&lt;/b&gt;`},
		// Jinja2: Markup methods escape their arguments with autoescape off too.
		{name: "format_autoescape_off", tpl: `{{ ('<b>{}</b>'|safe).format(x) }}`, data: x, want: `<b>&lt;i&gt;</b>`},
		{name: "replace_autoescape_off_tag", tpl: `{% autoescape false %}{{ ('<b>'|safe).replace('b', x) }}{% endautoescape %}`, data: x, autoescape: true, want: `<&lt;i&gt;>`},
	})
}

func TestDictMethods(t *testing.T) {
	d := func(kv ...interface{}) map[string]interface{} {
		m := map[string]interface{}{}
		for i := 0; i < len(kv); i += 2 {
			m[kv[i].(string)] = kv[i+1]
		}
		return m
	}
	runPvCases(t, []pvCase{
		{name: "get_present", tpl: `{{ d.get('a') }}`, data: pvD{"d": d("a", 1)}, want: `1`},
		{name: "get_missing", tpl: `[{{ d.get('z') }}]`, data: pvD{"d": d("a", 1)}, want: `[]`},
		{name: "get_default", tpl: `{{ d.get('z', 'def') }}`, data: pvD{"d": d("a", 1)}, want: `def`},
		{name: "get_none_value", tpl: `[{{ d.get('a', 'z') }}]`, data: pvD{"d": d("a", nil)}, want: `[]`},
		{name: "get_typed_map", tpl: `{{ m.get('a') }}`, data: pvD{"m": map[string]int{"a": 1}}, want: `1`},
		{name: "key_shadows_method", tpl: `{{ d.get }}`, data: pvD{"d": d("get", "k")}, want: `k`},
		{name: "items_list", tpl: `{{ d.items()|list }}`, data: pvD{"d": d("b", 2, "a", 1)}, want: `[['a', 1], ['b', 2]]`},
		{name: "items_length", tpl: `{{ d.items()|length }}`, data: pvD{"d": d("b", 2, "a", 1)}, want: `2`},
		{name: "items_loop", tpl: `{% for k, v in d.items() %}{{ k }}{{ v }}{% endfor %}`, data: pvD{"d": d("b", 2, "a", 1)}, want: `a1b2`},
		{name: "keys_values", tpl: `{{ d.keys()|list }}|{{ d.values()|list }}`, data: pvD{"d": d("a", 1)}, want: `['a']|[1]`},
		{name: "string_map_get", tpl: `{{ m.get('a') }}|[{{ m.get('z') }}]|{{ m.get('z', 'd') }}`, data: pvD{"m": map[string]string{"a": "x"}}, want: `x|[]|d`},
		{name: "string_map_items", tpl: `{{ m.items()|list }}|{{ m.keys()|list }}|{{ m.values()|list }}`, data: pvD{"m": map[string]string{"b": "y", "a": "x"}}, want: `[['a', 'x'], ['b', 'y']]|['a', 'b']|['x', 'y']`},
	})
}

func TestListMethods(t *testing.T) {
	const notStored = "can only change a list stored with"
	cases := []pvCase{
		{name: "append", tpl: `{% set l = [1] %}{% do l.append(2) %}{{ l }}`, want: `[1, 2]`},
		{name: "append_expr_none", tpl: `{% set l = [1] %}[{{ l.append(2) }}]{{ l }}`, want: `[][1, 2]`},
		{name: "append_in_loops", tpl: `{% set l = [] %}{% for x in [1, 2] %}{% for y in [3, 4] %}{% do l.append(x * y) %}{% endfor %}{% endfor %}{{ l }}`, want: `[3, 4, 6, 8]`},
		{name: "append_set_underscore", tpl: `{% set l = [] %}{% for x in [1, 2] %}{% set _ = l.append(x) %}{% endfor %}{{ l }}`, want: `[1, 2]`},
		{name: "unbox_read", tpl: `{% set l = [3, 4] %}{{ l|length }}|{{ l[1] }}|{% for x in l %}{{ x }}{% endfor %}|{{ l|join(',') }}`, want: `2|4|34|3,4`},
		{name: "extend_doubling_capped", tpl: `{% set l = [1] %}{% for i in range(40) %}{% do l.extend(l) %}{% endfor %}`, errHas: "too large"},
		{name: "extend", tpl: `{% set l = [1] %}{% do l.extend([2, 3]) %}{{ l }}`, want: `[1, 2, 3]`},
		{name: "pop", tpl: `{% set l = [1, 2, 3] %}{{ l.pop() }}{{ l }}|{% set m = [1, 2, 3] %}{{ m.pop(0) }}{{ m }}`, want: `3[1, 2]|1[2, 3]`},
		{name: "pop_empty", tpl: `{% set l = [] %}{{ l.pop() }}`, errHas: "empty"},
		{name: "index_count", tpl: `{{ [1, 2, 1].index(2) }}|{{ [1, 2, 1].count(1) }}`, want: `1|2`},
		{name: "index_missing", tpl: `{{ [1, 2].index(5) }}`, errHas: "not in list"},
		{name: "count_json_numbers", tpl: `{{ l.count(1) }}`, data: pvD{"l": []interface{}{float64(1), float64(1)}}, want: `2`},
		{name: "alias", tpl: `{% set a = [1] %}{% set b = a %}{% do b.append(2) %}{{ a }}`, want: `[1, 2]`},
		{name: "with_list", tpl: `{% with l = [] %}{% do l.append(1) %}{{ l }}{% endwith %}`, want: `[1]`},
		{name: "caller_list_rejected", tpl: `{% do items.append(9) %}`, errHas: notStored},
		{name: "caller_list_copied", tpl: `{% set items = items|list %}{% do items.append(9) %}{{ items }}`, want: `[1, 9]`},
		{name: "caller_alias_copies", tpl: `{% set mine = items %}{% do mine.append(9) %}{{ mine }}|{{ items }}`, want: `[1, 9]|[1]`},
		{name: "caller_ns_copied", tpl: `{% set ns = namespace(items=items) %}{% do ns.items.append(9) %}{{ ns.items }}|{{ items }}`, want: `[1, 9]|[1]`},
		{name: "ns_accumulator", tpl: `{% set ns = namespace(items=[]) %}{% for x in [1, 2] %}{% do ns.items.append(x) %}{% endfor %}{{ ns.items }}`, want: `[1, 2]`},
		{name: "ns_set_then_append", tpl: `{% set ns = namespace() %}{% set ns.l = [1] %}{% do ns.l.append(2) %}{{ ns.l }}|{{ ns.l|length }}`, want: `[1, 2]|2`},
		{name: "nested_is_snapshot", tpl: `{% set a = [1] %}{% set r = [a] %}{% do a.append(2) %}{{ r }}`, want: `[[1]]`},
		{name: "pop_then_append_safe", tpl: `{% set a = [1, 2, 3] %}{% do a.append(4) %}{% set r = [a] %}{% do a.pop(0) %}{{ r }}|{{ a }}`, want: `[[1, 2, 3, 4]]|[2, 3, 4]`},
		{name: "pop_end_after_write", tpl: `{% set a = [1, 2, 3] %}{% do a.append(4) %}{% set r = [a] %}{% do a.pop() %}{% do a.append(9) %}{{ r }}|{{ a }}`, want: `[[1, 2, 3, 4]]|[1, 2, 3, 9]`},
		{name: "iteration_snapshot", tpl: `{% set l = [1] %}{% for x in l %}{% do l.append(x) %}{% endfor %}{{ l }}`, want: `[1, 1]`},
		{name: "temporary_rejected", tpl: `{% do [1].append(2) %}`, errHas: notStored},
		{name: "loop_item_rejected", tpl: `{% for row in rows %}{% do row.append(1) %}{% endfor %}`, data: pvD{"rows": []interface{}{[]interface{}{1}}}, errHas: notStored},
		{name: "nested_rejected", tpl: `{% set l = [[1]] %}{% do l[0].append(2) %}`, errHas: notStored},
		{name: "macro_arg_append", tpl: `{% macro m(x) %}{% do x.append(1) %}{% endmacro %}{% set a = [] %}{{ m(a) }}{{ a }}`, want: "[1]"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			data := tc.data
			var items []interface{}
			if strings.HasPrefix(tc.name, "caller_") {
				items = make([]interface{}, 1, 4)
				items[0] = 1
				data = pvD{"items": items}
			}
			env := miya.NewEnvironment(miya.WithAutoEscape(false))
			got, err := renderGo(t, env, tc.tpl, data)
			if tc.errHas != "" {
				if err == nil || !strings.Contains(err.Error(), tc.errHas) {
					t.Fatalf("want error containing %q, got err=%v out=%q", tc.errHas, err, got)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			} else if got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
			if items != nil {
				if items[:2][1] != nil || items[0] != 1 {
					t.Fatalf("application slice was written: %v", items[:2])
				}
			}
		})
	}
}

func TestRepetition(t *testing.T) {
	runPvCases(t, []pvCase{
		{name: "str_int", tpl: `{{ 'a' * 3 }}`, want: `aaa`},
		{name: "int_str", tpl: `{{ 3 * 'ab' }}`, want: `ababab`},
		{name: "zero_neg", tpl: `[{{ 'a' * 0 }}|{{ 'a' * -1 }}]`, want: `[|]`},
		{name: "json_number", tpl: `{{ '-' * n }}`, data: pvD{"n": float64(3)}, want: `---`},
		{name: "fractional", tpl: `{{ 'ab' * 2.5 }}`, errHas: "integer"},
		{name: "literal_numeric", tpl: `{{ '0' * 3 }}|{{ '3' * 2 }}`, want: `000|33`},
		{name: "zero_pad_idiom", tpl: `{{ '0' * (6 - s|length) ~ s }}`, data: pvD{"s": "42"}, want: `000042`},
		{name: "data_numeric_kept", tpl: `{{ n * 2 }}`, data: pvD{"n": "3"}, want: `6`},
		{name: "data_inf_repeats", tpl: `{{ w * 2 }}`, data: pvD{"w": "inf"}, want: `infinf`},
		{name: "data_text_repeats", tpl: `{{ w * 2 }}`, data: pvD{"w": "ab"}, want: `abab`},
		{name: "list", tpl: `{{ [1] * 2 }}|{{ 2 * [1] }}|{{ [1, 2] * 0 }}`, want: `[1, 1]|[1, 1]|[]`},
		{name: "too_large", tpl: `{{ 'a' * 100000000 }}`, errHas: "too large"},
		{name: "list_too_large", tpl: `{{ [1] * 2000000 }}`, errHas: "too large"},
	})
}

// TestPythonReprCycle must stay last: before the fix, a cyclic value is a
// fatal stack overflow that kills the test binary.
func TestPythonReprCycle(t *testing.T) {
	render := func(t *testing.T, tpl string, data map[string]interface{}) (string, error) {
		t.Helper()
		return renderGo(t, miya.NewEnvironment(miya.WithAutoEscape(false)), tpl, data)
	}
	t.Run("self", func(t *testing.T) {
		cyc := []interface{}{nil}
		cyc[0] = cyc
		out, err := render(t, `{{ l }}`, pvD{"l": cyc})
		if err != nil || out != "[[...]]" {
			t.Fatalf("got %q, err=%v", out, err)
		}
	})
	t.Run("fanout_two", func(t *testing.T) {
		cyc := []interface{}{nil, nil}
		cyc[0] = cyc
		cyc[1] = cyc
		out, err := render(t, `{{ l }}`, pvD{"l": cyc})
		if err != nil || out != "[[...], [...]]" {
			t.Fatalf("got %q, err=%v", out, err)
		}
	})
	t.Run("through_safe_value", func(t *testing.T) {
		l := []interface{}{nil}
		l[0] = runtime.SafeValue{Value: l}
		out, err := render(t, `{{ l }}`, pvD{"l": l})
		if err != nil || out != "[[...]]" {
			t.Fatalf("got %q, err=%v", out, err)
		}
	})
	t.Run("self_append_snapshot", func(t *testing.T) {
		out, err := render(t, `{% set l = [1] %}{% do l.append(l) %}{{ l }}`, nil)
		if err != nil || out != "[1, [1]]" {
			t.Fatalf("got %q, err=%v", out, err)
		}
	})
	t.Run("doubling_budget", func(t *testing.T) {
		// printing 64 MiB takes seconds under the race detector, so this
		// subtest has a longer limit than renderGo's 5 seconds
		type res struct {
			out string
			err error
		}
		done := make(chan res, 1)
		go func() {
			env := miya.NewEnvironment(miya.WithAutoEscape(false))
			out, err := env.RenderString(`{% set ns = namespace(l=[1]) %}{% for i in range(40) %}{% set ns.l = [ns.l, ns.l] %}{% endfor %}{{ ns.l }}`, miya.NewContext())
			done <- res{out, err}
		}()
		select {
		case r := <-done:
			if r.err != nil {
				t.Fatalf("unexpected error: %v", r.err)
			}
			if len(r.out) > 64<<20+3 || !strings.HasSuffix(r.out, "...") {
				t.Fatalf("len=%d, suffix=%q", len(r.out), r.out[max(0, len(r.out)-10):])
			}
		case <-time.After(60 * time.Second):
			t.Fatalf("timed out")
		}
	})
}

func TestNamespacePrinting(t *testing.T) {
	t.Run("namespace_self_reference_prints", func(t *testing.T) {
		env := miya.NewEnvironment(miya.WithAutoEscape(false))
		out, err := renderGo(t, env, `{% set ns = namespace(a=1) %}{% set ns.me = ns %}{{ ns }}`, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if out != "namespace({'a': 1, 'me': namespace(...)})" {
			t.Fatalf("got %q", out)
		}
	})
	t.Run("namespace_plain", func(t *testing.T) {
		env := miya.NewEnvironment(miya.WithAutoEscape(false))
		out, err := renderGo(t, env, `{% set ns = namespace(a=[1, 'x']) %}{{ ns }}`, nil)
		if err != nil || out != "namespace({'a': [1, 'x']})" {
			t.Fatalf("got %q err=%v", out, err)
		}
	})
}
