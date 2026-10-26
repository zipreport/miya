package miya_test

import (
	"encoding/json"
	"strings"
	"testing"

	miya "github.com/zipreport/miya"
	"github.com/zipreport/miya/filters"
)

const filterKwargsContextJSON = `{"l":[3,1,2],"ld":[{"n":2,"g":"a"},{"n":1,"g":"b"},{"n":3,"g":"a"}],"f":3.14159,"h":3.7,"s":"a\nb","z":0,"dd":{"b":1,"a":2},"one":{"a":1}}`

func filterKwargsContext(t *testing.T) miya.Context {
	t.Helper()
	data := map[string]interface{}{}
	if err := json.Unmarshal([]byte(filterKwargsContextJSON), &data); err != nil {
		t.Fatal(err)
	}
	data["long"] = strings.Repeat("a", 300)
	return miya.NewContextFrom(data)
}

func TestFilterKeywordArgs(t *testing.T) {
	env := miya.NewEnvironment(miya.WithAutoEscape(false))
	err := env.AddFilterWithParams("unsupported_demo", func(v interface{}, args ...interface{}) (interface{}, error) { return v, nil },
		filters.Param{Name: "a"}, filters.Param{Name: "leeway", Unsupported: true})
	if err != nil {
		t.Fatal(err)
	}

	outputs := []struct {
		name     string
		template string
		want     string
	}{
		{"sort_reverse", `{{ l|sort(reverse=true)|join(',') }}`, "3,2,1"},
		{"sort_attribute", `{{ ld|sort(attribute='n')|map(attribute='n')|join(',') }}`, "1,2,3"},
		{"map_attribute", `{{ ld|map(attribute='n')|join(',') }}`, "2,1,3"},
		{"map_default", `{{ ld|map(attribute='zz', default='-')|join(',') }}`, "-,-,-"},
		{"sum_start", `{{ l|sum(start=10) }}`, "16"},
		{"sum_attribute", `{{ ld|sum(attribute='n') }}`, "6"},
		{"sum_positional_attribute", `{{ ld|sum('n') }}`, "6"},
		{"max_positional_case", `{{ l|max(true) }}`, "3"},
		{"min_positional_case", `{{ l|min(true) }}`, "1"},
		{"round_precision", `{{ f|round(precision=1) }}`, "3.1"},
		{"join_d", `{{ l|join(d='-') }}`, "3-1-2"},
		{"join_attribute", `{{ ld|join(',', attribute='n') }}`, "2,1,3"},
		{"default_boolean", `{{ z|default('x', boolean=true) }}`, "x"},
		{"truncate_defaults_gap", `{{ long|truncate(end='+') }}`, strings.Repeat("a", 254) + "+"},
		{"round_method_gap", `{{ h|round(method='floor') }}`, "3"},
		{"indent_first", `{{ s|indent(first=true) }}`, "    a\n    b"},
		{"int_base", `{{ 'ff'|int(base=16) }}`, "255"},
		{"int_base_gap_default", `{{ 'zz'|int(base=16) }}`, "0"},
		{"trim_chars_none", `{{ '  x  '|trim(chars=none) }}`, "x"},
		{"join_repeat_keyword", `{{ l|join(d='-', d='+') }}`, "3+1+2"},
		{"int_default", `{{ 'x'|int(default=7) }}`, "7"},
		{"float_default", `{{ 'x'|float(default=1.5) }}`, "1.5"},
		{"replace_count", `{{ 'aaa'|replace('a', 'b', count=2) }}`, "bba"},
		{"dictsort_reverse", `{% for k, v in dd|dictsort(reverse=true) %}{{ k }}{% endfor %}`, "ba"},
		{"groupby_attribute", `{% for g, items in ld|groupby(attribute='g') %}{{ g }}{{ items|length }};{% endfor %}`, "a2;b1;"},
		{"filesizeformat_binary", `{{ 2048|filesizeformat(binary=true) }}`, "2.0 KiB"},
		{"tojson_indent", `{{ one|tojson(indent=2) }}`, "{\n  \"a\": 1\n}"},
		{"filter_block", `{% filter replace('a', 'b', count=1) %}aaa{% endfilter %}`, "baa"},
	}
	for _, tc := range outputs {
		t.Run(tc.name, func(t *testing.T) {
			got, err := env.RenderString(tc.template, filterKwargsContext(t))
			if err != nil {
				t.Fatalf("render %q: %v", tc.template, err)
			}
			if got != tc.want {
				t.Errorf("%s: got %q, want %q", tc.template, got, tc.want)
			}
		})
	}

	errorCases := []struct {
		name     string
		template string
		contains string
	}{
		{"unknown_kwarg", `{{ l|join(bogus=1) }}`, `unexpected keyword argument "bogus"`},
		{"duplicate_kwarg", `{{ l|join(',', d='-') }}`, `multiple values for argument "d"`},
		{"no_params", `{{ 'a'|upper(x=1) }}`, "does not accept keyword arguments"},
		{"unsupported_param", `{{ 'abc'|unsupported_demo(5, leeway=0) }}`, `keyword argument "leeway" is not supported`},
		{"missing_required", `{{ l|batch(fill_with=0) }}`, `missing required argument "linecount"`},
		{"sum_positional_number", `{{ l|sum(10) }}`, "first positional argument is the attribute"},
		{"positional_after_keyword", `{{ l|join(d='-', ',') }}`, "positional argument follows keyword argument"},
		{"map_unknown_kwarg", `{{ ld|map(attribute='n', x=1)|join(',') }}`, `unexpected keyword argument "x"`},
		{"map_default_alone", `{{ ld|map(default='-')|join(',') }}`, "map: default= needs attribute="},
		{"map_unknown_kwarg_positional", `{{ ld|map('n', x=1)|join(',') }}`, `unexpected keyword argument "x"`},
		{"positional_after_keyword_block", `{% filter replace(count=1, 'a') %}aaa{% endfilter %}`, "positional argument follows keyword argument"},
	}
	for _, tc := range errorCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := env.RenderString(tc.template, filterKwargsContext(t))
			if err == nil {
				t.Fatalf("render %q: expected an error containing %q", tc.template, tc.contains)
			}
			if !strings.Contains(err.Error(), tc.contains) {
				t.Errorf("render %q: error %q does not contain %q", tc.template, err, tc.contains)
			}
		})
	}
}
