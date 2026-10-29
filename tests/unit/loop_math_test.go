package miya_test

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"testing"

	miya "github.com/zipreport/miya"
	"github.com/zipreport/miya/loader"
)

func loopMathContext() miya.Context {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(`{"tree":[{"c":[{"c":[{"c":[]}]}]}],"vt":[{"v":"a","c":[{"v":"b","c":[{"v":"c","c":[]}]}]}],"l":[1,[2,3],4],"x":-5.5,"ten":10,"one":1,"s":"10","f":3.7}`), &data); err != nil {
		panic(err)
	}
	data["i8"] = int8(-7)
	return miya.NewContextFrom(data)
}

func TestLoopScopeAndDepth(t *testing.T) {
	env := miya.NewEnvironment(miya.WithAutoEscape(false))

	tests := []struct {
		name     string
		template string
		want     string
	}{
		{"else_after_filter", `{% for i in [1,2,3] if i > 5 %}a{% else %}e{% endfor %}`, "e"},
		{"else_filtered_nonempty", `{% for i in [1,2,3] if i > 2 %}{{ i }}{% else %}e{% endfor %}`, "3"},
		{"set_not_persist", `{% for i in [1,2] %}{{ c|default('u') }}{% set c = i %}{% endfor %}`, "uu"},
		{"set_reads_outer", `{% set x = 1 %}{% for i in [1,2] %}{% set x = x + 1 %}{{ x }}{% endfor %}{{ x }}`, "221"},
		{"set_in_if", `{% for i in [1,2,3] %}{% if i == 1 %}{% set t = 'a' %}{% endif %}[{{ t }}]{% endfor %}`, "[a][][]"},
		{"namespace_persists", `{% set ns = namespace(c=0) %}{% for i in [1,2,3] %}{% set ns.c = ns.c + i %}{% endfor %}{{ ns.c }}`, "6"},
		{"depth_nested", `{% for a in [[1],[2]] %}{% for b in a %}{{ loop.depth }}{% endfor %}{% endfor %}`, "11"},
		{"depth_recursive", `{% for n in tree recursive %}{{ loop.depth }}{% if n.c %}({{ loop(n.c) }}){% endif %}{% endfor %}`, "1(2(3))"},
		{"depth0_recursive", `{% for n in tree recursive %}{{ loop.depth0 }}{% if n.c %}({{ loop(n.c) }}){% endif %}{% endfor %}`, "0(1(2))"},
		{"recursive_keeps_filter", `{% for i in l if i != 3 recursive %}{% if i is iterable %}({{ loop(i) }}){% else %}{{ i }}{% endif %}{% endfor %}`, "1(2)4"},
		{"rec_set_not_inherited", `{% for n in vt recursive %}[{{ x|default('u') }}]{% set x = n.v %}{% if n.c %}{{ loop(n.c) }}{% endif %}{% endfor %}`, "[-5.5][-5.5][-5.5]"},
		{"outer_loop_after_inner", `{% for a in [1,2] %}{% for b in [1] %}{% endfor %}{{ loop.index }}{% endfor %}`, "12"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := env.RenderString(tt.template, loopMathContext())
			if err != nil {
				t.Fatalf("render error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPythonFloorDivMod(t *testing.T) {
	env := miya.NewEnvironment(miya.WithAutoEscape(false))

	tests := []struct {
		name     string
		template string
		want     string
	}{
		{"floordiv_neg", `{{ -7 // 2 }}`, "-4"},
		{"floordiv_negdivisor", `{{ 7 // -2 }}`, "-4"},
		{"mod_neg", `{{ -7 % 3 }}`, "2"},
		{"mod_negdivisor", `{{ 7 % -3 }}`, "-2"},
		{"int_result", `{{ (7 // 2) ~ '|' ~ (7 % 2) }}`, "3|1"},
		{"exact_mod", `{{ 6 % 3 }}`, "0"},
		{"fmod", `{{ 7.5 % 2 }}`, "1.5"},
		{"fmod_neg", `{{ -7.5 % 2 }}`, "0.5"},
		{"ffloordiv", `{{ x // 2 }}`, "-3"},
		{"ffloordiv_small", `{{ ten // 0.3 }}`, "33"},
		{"ffloordiv_tenth", `{{ one // 0.1 }}`, "9"},
		{"ffloordiv_negtenth", `{{ -1 // 0.1 }}`, "-10"},
		{"negzero_floordiv", `{{ 0.0 // -1 }}`, "-0"},
		{"floordiv_float_not_integer", `{{ (ten // 2) is integer }}`, "false"},
		{"int8_floordiv", `{{ i8 // 2 }}`, "-4"},
		{"int8_mod", `{{ i8 % 3 }}`, "2"},
		{"format_d_json", `{{ "%02d:%02d"|format(ten*75 // 60, ten*75 % 60) }}`, "12:30"},
		{"format_i_float", `{{ "%i|%5d"|format(ten, ten) }}`, "10|   10"},
		{"format_d_truncates", `{{ "%d"|format(f) }}`, "3"},
		{"format_f_untouched", `{{ "%.1f"|format(f) }}`, "3.7"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := env.RenderString(tt.template, loopMathContext())
			if err != nil {
				t.Fatalf("render error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}

	errTests := []struct {
		name     string
		template string
		zero     bool
	}{
		{"floordiv_zero", `{{ 1 // 0 }}`, true},
		{"ffloordiv_zero", `{{ 1.5 // 0 }}`, true},
		{"mod_zero", `{{ 1 % 0 }}`, true},
		{"fmod_zero", `{{ 1 % 0.0 }}`, true},
		{"str_floordiv", `{{ s // 3 }}`, false},
		{"str_mod", `{{ s % 3 }}`, false},
	}

	for _, tt := range errTests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := env.RenderString(tt.template, loopMathContext())
			if err == nil {
				t.Fatalf("expected error, got %q", got)
			}
			if tt.zero && !strings.Contains(err.Error(), "by zero") {
				t.Errorf("error %q does not mention division by zero", err)
			}
		})
	}
}

func TestLoopUnderExtends(t *testing.T) {
	tests := []struct {
		name     string
		template string
		want     string
	}{
		{"filter_under_extends", `{% extends 'base' %}{% block b %}{% for i in [1,2,3] if i > 1 %}{{ i }}{% endfor %}{% endblock %}`, "[23]"},
		{"else_after_filter_under_extends", `{% extends 'base' %}{% block b %}{% for i in [1,2,3] if i > 5 %}a{% else %}e{% endfor %}{% endblock %}`, "[e]"},
		{"recursive_under_extends", `{% extends 'base' %}{% block b %}{% for n in tree recursive %}{{ loop.depth }}{% if n.c %}({{ loop(n.c) }}){% endif %}{% endfor %}{% endblock %}`, "[1(2(3))]"},
		{"super_without_parent_keeps_filter", `{% for i in [1,2,3] if i > 1 %}{{ i }}{% endfor %}{% block b %}{{ super() }}{% endblock %}`, "23"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sl := loader.NewStringLoader(loader.NewDirectTemplateParser())
			sl.AddTemplate("base", "[{% block b %}{% endblock %}]")
			sl.AddTemplate("main", tt.template)
			env := miya.NewEnvironment(miya.WithAutoEscape(false), miya.WithLoader(sl))
			tmpl, err := env.GetTemplate("main")
			if err != nil {
				t.Fatalf("GetTemplate error: %v", err)
			}
			got, err := tmpl.Render(loopMathContext())
			if err != nil {
				t.Fatalf("render error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func loopAllocPerRender(t *testing.T, vars int) uint64 {
	t.Helper()
	env := miya.NewEnvironment(miya.WithAutoEscape(false))
	tmpl, err := env.FromString("{% for r in rows %}{% set y = r %}{{ y }}{% endfor %}")
	if err != nil {
		t.Fatal(err)
	}
	rows := make([]interface{}, 10000)
	for i := range rows {
		rows[i] = float64(i)
	}
	data := map[string]interface{}{"rows": rows}
	for i := 0; i < vars; i++ {
		data[fmt.Sprintf("v%d", i)] = i
	}
	out, err := tmpl.Render(miya.NewContextFrom(data))
	if err != nil || !strings.HasSuffix(out, "9999") {
		t.Fatalf("render failed: err=%v, output ends %q", err, out[max(0, len(out)-10):])
	}
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	for i := 0; i < 3; i++ {
		if _, err := tmpl.Render(miya.NewContextFrom(data)); err != nil {
			t.Fatal(err)
		}
	}
	runtime.ReadMemStats(&after)
	return (after.TotalAlloc - before.TotalAlloc) / 3
}

func TestLoopScopeCostIndependentOfContextWidth(t *testing.T) {
	narrow, wide := loopAllocPerRender(t, 5), loopAllocPerRender(t, 200)
	if narrow == 0 {
		t.Fatal("no allocation measured for the narrow context")
	}
	if ratio := float64(wide) / float64(narrow); ratio > 1.5 {
		t.Fatalf("a 10k-row loop allocates %.1fx more with 200 variables in scope than with 5 (narrow %d B, wide %d B); each iteration must not copy the context", ratio, narrow, wide)
	}
}

func BenchmarkLoopScope10kRows(b *testing.B) {
	env := miya.NewEnvironment(miya.WithAutoEscape(false))
	tmpl, err := env.FromString("{% for r in rows %}{% set y = r %}{{ y }}{% endfor %}")
	if err != nil {
		b.Fatal(err)
	}
	rows := make([]interface{}, 10000)
	for i := range rows {
		rows[i] = float64(i)
	}
	data := map[string]interface{}{"rows": rows}
	for i := 0; i < 50; i++ {
		data[fmt.Sprintf("v%d", i)] = i
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := tmpl.Render(miya.NewContextFrom(data)); err != nil {
			b.Fatal(err)
		}
	}
}

// loop.changed compares like ==, so 1 and 1.0 are the same value, as in Jinja2.
func TestLoopChanged(t *testing.T) {
	env := miya.NewEnvironment(miya.WithAutoEscape(false))

	tests := []struct {
		name     string
		template string
		want     string
	}{
		{"int_float", `{% for x in [1, 1.0, 2] %}{{ loop.changed(x) }} {% endfor %}`, "true false true "},
		{"two_args", `{% for x in [1, 1.0, 1.5, 1.5] %}{{ loop.changed(x, 'k') }} {% endfor %}`, "true false true false "},
		{"lists", `{% for l in [[1, 2], [1.0, 2.0], [1, 3]] %}{{ loop.changed(l) }} {% endfor %}`, "true false true "},
		{"bool_int", `{% for x in [true, 1] %}{{ loop.changed(x) }} {% endfor %}`, "true false "},
		{"safe_string", `{% for s in ['a', 'a'|safe, 'b'] %}{{ loop.changed(s) }} {% endfor %}`, "true false true "},
		{"dicts", `{% for d in [{'a': 1}, {'a': 1}, {'a': 2}] %}{{ loop.changed(d) }} {% endfor %}`, "true false true "},
		{"string_vs_number", `{% for x in ['1', 1] %}{{ loop.changed(x) }} {% endfor %}`, "true true "},
		{"none_then_zero", `{% for x in [none, none, 0] %}{{ loop.changed(x) }} {% endfor %}`, "true false true "},
		{"same_strings", `{% for s in ['a', 'a', 'b'] %}{{ loop.changed(s) }} {% endfor %}`, "true false true "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := env.RenderString(tt.template, miya.NewContext())
			if err != nil {
				t.Fatalf("render error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

// '5' + 1 is 6 and 'a' + 1 is 'a1': a deliberate difference from Jinja2,
// which raises TypeError for both (docs/MIYA_LIMITATIONS.md). Two strings
// always concatenate, as in Jinja2: '5' + '1' is '51'. Other operators still
// coerce two numeric strings ('5' - '1' is 4; Jinja2 raises TypeError).
func TestStringNumberAddition(t *testing.T) {
	tests := []struct {
		name     string
		template string
		strict   bool
		want     string
		wantErr  bool
	}{
		{"numeric_string_plus_int", `{{ '5' + 1 }}`, false, "6", false},
		{"int_plus_numeric_string", `{{ 1 + '5' }}`, false, "6", false},
		{"two_numeric_strings", `{{ '5' + '1' }}`, false, "51", false},
		{"two_data_strings", `{{ s + t }}`, false, "51", false},
		{"exponent_strings", `{{ '1e3' + '1' }}`, false, "1e31", false},
		{"namespace_accumulates_text", `{% set ns = namespace(a='0') %}{% for x in ['1', '2'] %}{% set ns.a = ns.a + x %}{% endfor %}{{ ns.a }}`, false, "012", false},
		{"strict_two_numeric_strings", `{{ '5' + '1' }}`, true, "51", false},
		{"data_string_plus_int", `{{ s + 1 }}`, false, "6", false},
		{"text_plus_int", `{{ 'a' + 1 }}`, false, "a1", false},
		{"int_plus_text", `{{ 1 + 'a' }}`, false, "1a", false},
		{"two_texts", `{{ 'a' + 'b' }}`, false, "ab", false},
		{"strict_numeric_string", `{{ '5' + 1 }}`, true, "6", false},
		{"strict_text_plus_int", `{{ 'a' + 1 }}`, true, "", true},
		{"two_data_strings_subtract", `{{ s - t }}`, false, "4", false},
		{"two_data_strings_multiply", `{{ s * t }}`, false, "5", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := miya.NewEnvironment(miya.WithAutoEscape(false), miya.WithStrictUndefined(tt.strict))
			got, err := env.RenderString(tt.template, miya.NewContextFrom(map[string]interface{}{"s": "5", "t": "1"}))
			if tt.wantErr {
				if err == nil {
					t.Fatalf("want an error, got %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("render error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
