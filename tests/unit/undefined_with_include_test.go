package miya_test

import (
	"errors"
	"strings"
	"testing"

	miya "github.com/zipreport/miya"
	"github.com/zipreport/miya/loader"
	"github.com/zipreport/miya/runtime"
)

type uwiCase struct {
	name     string
	tmpl     string
	want     string
	contains string
}

func runUWICases(t *testing.T, env *miya.Environment, ctx map[string]interface{}, cases []uwiCase) {
	t.Helper()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := env.RenderString(tc.tmpl, miya.NewContextFrom(ctx))
			checkUWI(t, tc, got, err)
		})
	}
}

func checkUWI(t *testing.T, tc uwiCase, got string, err error) {
	t.Helper()
	if tc.want == "ERR" {
		if err == nil {
			t.Fatalf("%s: expected an error, got %q", tc.tmpl, got)
		}
		if tc.contains != "" && !strings.Contains(err.Error(), tc.contains) {
			t.Fatalf("%s: error %q does not contain %q", tc.tmpl, err, tc.contains)
		}
		return
	}
	if err != nil {
		t.Fatalf("%s: unexpected error: %v", tc.tmpl, err)
	}
	if got != tc.want {
		t.Fatalf("%s: got %q, want %q", tc.tmpl, got, tc.want)
	}
}

func TestStrictUndefinedTolerance(t *testing.T) {
	env := miya.NewEnvironment(miya.WithStrictUndefined(true))
	ctx := map[string]interface{}{
		"o": map[string]interface{}{"k": nil},
		"l": []interface{}{1},
	}
	runUWICases(t, env, ctx, []uwiCase{
		{"default_ident", `{{ a|default('x') }}`, "x", ""},
		{"d_alias", `{{ a|d('x') }}`, "x", ""},
		{"default_attr", `{{ o.zz|default('x') }}`, "x", ""},
		{"default_noarg", `[{{ a|default }}]`, "[]", ""},
		{"null_subscript_default", `[{{ o.k['x']|default('-') }}]`, "[-]", ""},
		{"null_subscript_defined", `{{ o.k['x'] is defined }}`, "false", ""},
		{"default_bool", `{{ a|default('x', true) }}`, "x", ""},
		{"is_undefined", `{{ a is undefined }}`, "true", ""},
		{"attr_is_undefined", `{{ o.zz is undefined }}`, "true", ""},
		{"is_none", `{{ a is none }}`, "false", ""},
		{"nil_attr_defined", `{{ o.k is defined }}`, "true", ""},
		{"item_default", `{{ o['zz']|default('x') }}`, "x", ""},
		{"index_default", `{{ l[5]|default('x') }}`, "x", ""},
		{"index_defined", `{{ l[5] is defined }}`, "false", ""},
		{"if_defined", `{% if a is defined %}y{% else %}n{% endif %}`, "n", ""},
		{"item_print", `{{ o['zz'] }}`, "ERR", ""},
		{"index_print", `{{ l[5] }}`, "ERR", ""},
		{"plain", `{{ a }}`, "ERR", "undefined variable: a"},
		{"chain_default", `{{ a.b|default('x') }}`, "ERR", ""},
		{"chain_defined", `{{ a.b is defined }}`, "ERR", ""},
		{"error_not_swallowed", `{{ (1 // 0) is defined }}`, "ERR", "division by zero"},
		{"filter_on_undefined", `{{ a|upper }}`, "ERR", "undefined variable: a"},
	})
}

func TestSilentUndefinedValues(t *testing.T) {
	env := miya.NewEnvironment()
	runUWICases(t, env, map[string]interface{}{}, []uwiCase{
		{"length", `[{{ a|length }}]`, "[0]", ""},
		{"count", `[{{ a|count }}]`, "[0]", ""},
		{"join", `[{{ a|join(',') }}]`, "[]", ""},
		{"sum", `[{{ a|sum }}]`, "[0]", ""},
		{"sort", `[{{ a|sort|join }}]`, "[]", ""},
		{"reverse", `[{{ a|reverse|join }}]`, "[]", ""},
		{"unique", `[{{ a|unique|join }}]`, "[]", ""},
		{"max", `[{{ a|max }}]`, "[]", ""},
		{"min", `[{{ a|min }}]`, "[]", ""},
		{"list", `{{ a|list|length }}`, "0", ""},
		{"batch", `{{ a|batch(2)|list|length }}`, "0", ""},
		{"map", `{{ a|map('upper')|list|length }}`, "0", ""},
		{"select", `{{ a|select|list|length }}`, "0", ""},
		{"reject", `{{ a|reject|list|length }}`, "0", ""},
		{"selectattr", `{{ a|selectattr('x')|list|length }}`, "0", ""},
		{"rejectattr", `{{ a|rejectattr('x')|list|length }}`, "0", ""},
		{"groupby", `{{ a|groupby('x')|list|length }}`, "0", ""},
		{"items", `[{% for k, v in a|items %}{{ k }}{% endfor %}]`, "[]", ""},
		{"random", `[{{ a|random }}]`, "[]", ""},
		{"in", `{{ 1 in a }}`, "false", ""},
		{"not_in", `{{ 1 not in a }}`, "true", ""},
		{"length_gt", `{{ a|length > 0 }}`, "false", ""},
		{"is_iterable", `{{ a is iterable }}`, "true", ""},
		{"is_sequence", `{{ a is sequence }}`, "true", ""},
		{"tojson", `{{ a|tojson }}`, "ERR", "Undefined"},
		{"first", `[{{ a|first }}]`, "[]", ""},
		{"upper", `[{{ a|upper }}]`, "[]", ""},
		{"for_else", `{% for x in a %}x{% else %}E{% endfor %}`, "E", ""},
		{"is_mapping", `{{ a is mapping }}`, "false", ""},
		{"deviation_chain", `[{{ a.b }}]`, "[]", ""},
		{"deviation_int", `{{ a|int }}`, "0", ""},
	})
}

func TestNoneIsDefined(t *testing.T) {
	env := miya.NewEnvironment()
	ctx := map[string]interface{}{
		"n": nil,
		"o": map[string]interface{}{"k": nil},
		"s": "",
		"z": 0,
		"l": []interface{}{1},
		"rows": []interface{}{
			map[string]interface{}{"e": "a"},
			map[string]interface{}{"e": nil},
			map[string]interface{}{},
		},
	}
	runUWICases(t, env, ctx, []uwiCase{
		{"none_defined", `{{ n is defined }}`, "true", ""},
		{"none_undefined", `{{ n is undefined }}`, "false", ""},
		{"none_attr_defined", `{{ o.k is defined }}`, "true", ""},
		{"none_item_defined", `{{ o['k'] is defined }}`, "true", ""},
		{"deviation_default_none", `[{{ n|default('x') }}]`, "[x]", ""},
		{"deviation_default_attr_none", `[{{ o.k|default('x') }}]`, "[x]", ""},
		{"deviation_d_none", `[{{ n|d('x') }}]`, "[x]", ""},
		{"deviation_default_empty", `[{{ s|default('x') }}]`, "[x]", ""},
		{"default_zero_kept", `[{{ z|default('x') }}]`, "[0]", ""},
		{"default_rows", `{% for r in rows %}[{{ r.e|default('-') }}]{% endfor %}`, "[a][-][-]", ""},
		{"default_zero_arith", `[{{ n|default(0) + 1 }}]`, "[1]", ""},
		{"first_empty_default", `[{{ []|first|default('x') }}]`, "[x]", ""},
		{"attr_missing_default", `[{{ o|attr('zz')|default('x') }}]`, "[x]", ""},
		{"none_subscript_defined", `{{ n['x'] is defined }}`, "false", ""},
		{"int_subscript_defined", `{{ l[0]['x'] is defined }}`, "false", ""},
		{"list_strkey_defined", `{{ l['x'] is defined }}`, "false", ""},
		{"none_default_bool", `[{{ n|default('x', true) }}]`, "[x]", ""},
		{"empty_default_bool", `[{{ s|default('x', true) }}]`, "[x]", ""},
		{"zero_default_bool", `[{{ z|default('x', true) }}]`, "[x]", ""},
		{"missing_item_is_none", `{{ o['zz'] is none }}`, "false", ""},
		{"missing_item_defined", `{{ o['zz'] is defined }}`, "false", ""},
		{"none_is_none", `{{ n is none }}`, "true", ""},
	})
}

func TestIsTestSemantics(t *testing.T) {
	env := miya.NewEnvironment()
	ctx := map[string]interface{}{
		"f": 4.0, "x": 4.5, "g": -3.0, "n": nil, "b": true,
		"d": map[string]interface{}{"a": 1},
	}
	runUWICases(t, env, ctx, []uwiCase{
		{"even_float", `{{ f is even }}`, "true", ""},
		{"even_frac", `{{ x is even }}`, "false", ""},
		{"odd_negfloat", `{{ g is odd }}`, "true", ""},
		{"divisible_float", `{{ 7.5 is divisibleby(2.5) }}`, "true", ""},
		{"even_string", `{{ '4' is even }}`, "ERR", ""},
		{"lower_number", `{{ 5 is lower }}`, "false", ""},
		{"upper_number", `{{ 5 is upper }}`, "false", ""},
		{"lower_empty", `{{ '' is lower }}`, "false", ""},
		{"upper_empty", `{{ '' is upper }}`, "false", ""},
		{"lower_digits", `{{ '12' is lower }}`, "false", ""},
		{"lower_mixed", `{{ 'abc1' is lower }}`, "true", ""},
		{"lower_unicode", `{{ 'ação' is lower }}`, "true", ""},
		{"upper_unicode", `{{ 'AÇÃO' is upper }}`, "true", ""},
		{"lower_none", `{{ n is lower }}`, "false", ""},
		{"number_bool", `{{ b is number }}`, "true", ""},
		{"integer_bool", `{{ b is integer }}`, "false", ""},
		{"sequence_dict", `{{ d is sequence }}`, "true", ""},
		{"sequence_number", `{{ 3 is sequence }}`, "false", ""},
		{"lt_none", `{{ n is lt(3) }}`, "ERR", ""},
		{"gt_none", `{{ n is gt(3) }}`, "ERR", ""},
		{"gt_string_number", `{{ '10' is gt(9) }}`, "ERR", ""},
		{"gt_strings", `{{ 'b' is gt('a') }}`, "true", ""},
		{"lt_numbers", `{{ 2 is lt(3) }}`, "true", ""},
		{"eq_none", `{{ n is eq(n) }}`, "true", ""},
	})
}

func TestWithScope(t *testing.T) {
	env := miya.NewEnvironment()
	runUWICases(t, env, map[string]interface{}{}, []uwiCase{
		{"outer_scope", `{% set a = 10 %}{% with a = 1, b = a %}{{ b }}{% endwith %}`, "10", ""},
		{"no_outer", `{% with a = 1, b = a %}[{{ b }}]{% endwith %}`, "[]", ""},
		{"chain5", `{% set a = 0 %}{% with a = 1, b = a + 1, c = b, d = c, e = d %}{{ a }}{{ b }}{{ c }}{{ d }}{{ e }}{% endwith %}`, "11", ""},
		{"eval_order", `{% set c = cycler('x', 'y', 'z') %}{% with b = c.next(), a = c.next() %}{{ b }}{{ a }}{% endwith %}`, "xy", ""},
		{"reads_outer", `{% set x = 5 %}{% with y = x + 1 %}{{ y }}{% endwith %}`, "6", ""},
		{"shadow", `{% set a = 10 %}{% with a = a + 1 %}{{ a }}{% endwith %}{{ a }}`, "1110", ""},
		{"not_leak", `{% with y = 1 %}{% endwith %}[{{ y }}]`, "[]", ""},
		{"bare", `{% with %}x{% endwith %}`, "x", ""},
		{"bare_set_scoped", `{% with %}{% set z = 1 %}{% endwith %}[{{ z }}]`, "[]", ""},
		{"duplicate", `{% with a = 1, a = 2 %}{{ a }}{% endwith %}`, "2", ""},
		{"duplicate_both", `{% set c = cycler('x', 'y') %}{% with a = c.next(), a = c.next() %}{{ a }}{% endwith %}`, "y", ""},
	})
}

type uwiIncludeCase struct {
	name      string
	templates map[string]string
	ctx       map[string]interface{}
	want      string
	contains  string
}

func newIncludeEnv(templates map[string]string) *miya.Environment {
	sl := loader.NewStringLoader(loader.NewDirectTemplateParser())
	for name, src := range templates {
		sl.AddTemplate(name, src)
	}
	return miya.NewEnvironment(miya.WithLoader(sl))
}

func TestIncludeScope(t *testing.T) {
	p := func(s string) map[string]string { return map[string]string{"p": s} }
	with := func(main string, others map[string]string) map[string]string {
		m := map[string]string{"main": main}
		for k, v := range others {
			m[k] = v
		}
		return m
	}
	cases := []uwiIncludeCase{
		{"set_not_leak", with(`{% include 'p' %}[{{ z }}]`, p(`{% set z = 1 %}`)), nil, "[]", ""},
		{"set_not_overwrite", with(`{% set z = 0 %}{% include 'p' %}[{{ z }}]`, p(`{% set z = 1 %}`)), nil, "[0]", ""},
		{"set_local", with(`{% set w = 1 %}{% include 'p' %}{{ w }}`, p(`{% set w = 2 %}{{ w }}`)), nil, "21", ""},
		{"sees_context", with(`{% set w = 'W' %}{% include 'p' %}`, p(`{{ v }}{{ w }}`)), map[string]interface{}{"v": "V"}, "VW", ""},
		{"loop_var", with(`{% for i in [1,2] %}{% include 'p' %}{% endfor %}`, p(`{{ i }}`)), nil, "12", ""},
		{"missing_ignored", with(`a{% include 'nope' ignore missing %}b`, nil), nil, "ab", ""},
		{"missing_error", with(`a{% include 'nope' %}b`, nil), nil, "ERR", "nope"},
		{"runtime_error_not_ignored", with(`a{% include 'bad' ignore missing %}b`, map[string]string{"bad": `{{ 1 // 0 }}`}), nil, "ERR", "division by zero"},
		{"syntax_error_not_ignored", with(`a{% include 'bad' ignore missing %}b`, map[string]string{"bad": `{% if %}`}), nil, "ERR", ""},
		{"nested_missing_not_ignored", with(`a{% include 'p' ignore missing %}b`, p(`{% include 'nope' %}`)), nil, "ERR", "nope"},
		{"list_first_found", with(`a{% include ['nope', 'p'] ignore missing %}b`, p(`P`)), nil, "aPb", ""},
		{"list_first_found_noignore", with(`a{% include ['nope', 'p'] %}b`, p(`P`)), nil, "aPb", ""},
		{"list_none_ignored", with(`a{% include ['nope', 'nope2'] ignore missing %}b`, nil), nil, "ab", ""},
		{"list_none_error", with(`a{% include ['nope', 'nope2'] %}b`, nil), nil, "ERR", "nope2"},
		{"none_name_ignored", with(`a{% include none ignore missing %}b`, nil), nil, "ab", ""},
		{"none_name_error", with(`a{% include none %}b`, nil), nil, "ERR", ""},
		{"parent_missing_not_ignored", with(`a{% include 'c' ignore missing %}b`, map[string]string{"c": `{% extends 'nope' %}`}), nil, "ERR", "nope"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			env := newIncludeEnv(tc.templates)
			tmpl, err := env.GetTemplate("main")
			var got string
			if err == nil {
				got, err = tmpl.Render(miya.NewContextFrom(tc.ctx))
			}
			checkUWI(t, uwiCase{tc.name, tc.templates["main"], tc.want, tc.contains}, got, err)
		})
	}
}

func TestIncludeNotFoundSentinel(t *testing.T) {
	env := newIncludeEnv(map[string]string{"main": `a{% include 'nope' %}b`})
	tmpl, err := env.GetTemplate("main")
	if err != nil {
		t.Fatal(err)
	}
	_, err = tmpl.Render(miya.NewContext())
	if !errors.Is(err, runtime.ErrTemplateNotFound) {
		t.Fatalf("want ErrTemplateNotFound, got %v", err)
	}
	var nf *runtime.TemplateNotFoundError
	if !errors.As(err, &nf) || nf.Name != "nope" {
		t.Fatalf("want TemplateNotFoundError for nope, got %v", err)
	}

	env = newIncludeEnv(map[string]string{"main": `{% include 'bad' %}`, "bad": `{% if %}`})
	tmpl, err = env.GetTemplate("main")
	if err == nil {
		_, err = tmpl.Render(miya.NewContext())
	}
	if err == nil || errors.Is(err, runtime.ErrTemplateNotFound) {
		t.Fatalf("a template that does not compile is not 'not found': %v", err)
	}
}
