package miya_test

import (
	"encoding/json"
	"strings"
	"testing"

	miya "github.com/zipreport/miya"
)

const collectionFilterContextJSON = `{"l":[3,1,2],"nums":[10,9,2,33],"strs":["3","10","2"],"fruits":["apple","cherry","banana"],"mixed":["a","B","b","A"],"ld":[{"n":2,"g":"a"},{"n":1,"g":"b"},{"n":3,"g":"a"}],"dd":{"b":1,"a":2},"seq":[1,2,3,4,5,6,7],"nn":[1,1.0,"1"],"words":["b","a"],"ab":[{"a":1},{"a":0},{"a":2}],"empty":[],
"bl":[true,false,true],"bools":[{"p":true,"i":"a"},{"p":false,"i":"b"},{"p":true,"i":"c"}],"nested":[[2,1],[1,2]],
"ldd":[{"c":{"name":"b"},"n":1},{"c":{"name":"a"},"n":2}],"pairs":[[2,"x"],[1,"y"]],"mk":[{"g":"b","n":1},{"g":"a","n":2},{"g":"a","n":1}],
"ldm":[{"g":"a","i":"x"},{"i":"y"},{"g":"a","i":"z"}],"nd":[{"n":1},{"n":null},{}],"fl":[1,2,3],"ppl":[{"title":"Mr"},{"title":"Dr"}],
"nums2":[1.234,2.345],"nl":[{"i":"a","n":2},{"i":"b","n":null},{"i":"c","n":1},{"i":"d"}],"mx":[{"v":1},{"v":"x"}]}`

func collectionFilterContext(t *testing.T) miya.Context {
	t.Helper()
	data := map[string]interface{}{}
	if err := json.Unmarshal([]byte(collectionFilterContextJSON), &data); err != nil {
		t.Fatal(err)
	}
	return miya.NewContextFrom(data)
}

func TestCollectionFilterSemantics(t *testing.T) {
	env := miya.NewEnvironment(miya.WithAutoEscape(false))

	outputs := []struct {
		name     string
		template string
		want     string
	}{
		{"sort_numbers", `{{ nums|sort|join(',') }}`, "2,9,10,33"},
		{"sort_reverse_numbers", `{{ nums|sort(true)|join(',') }}`, "33,10,9,2"},
		{"sort_strings_ci", `{{ mixed|sort|join(',') }}`, "a,A,B,b"},
		{"sort_strings_cs", `{{ mixed|sort(false, true)|join(',') }}`, "A,B,a,b"},
		{"sort_attribute_numbers", `{{ ld|sort(attribute='n')|map(attribute='g')|join(',') }}`, "b,a,a"},
		{"sort_string_input", `{{ 'cab'|sort|join }}`, "abc"},
		{"dict_input", `{{ dd|sort|join(',') }}`, "a,b"},
		{"unique_ci", `{{ mixed|unique|join(',') }}`, "a,B"},
		{"unique_cs", `{{ mixed|unique(true)|join(',') }}`, "a,B,b,A"},
		{"unique_numbers", `{{ nn|unique|join(',') }}`, "1,1"},
		{"unique_attribute", `{{ ld|unique(attribute='g')|map(attribute='n')|join(',') }}`, "2,1"},
		{"max_strings_numeric_text", `{{ strs|max }}`, "3"},
		{"max_strings", `{{ fruits|max }}`, "cherry"},
		{"min_strings", `{{ fruits|min }}`, "apple"},
		{"max_ci", `{{ ['B','a']|max }}`, "B"},
		{"max_cs", `{{ ['B','a']|max(true) }}`, "a"},
		{"max_attribute_item", `{{ (ld|max(attribute='n')).g }}`, "a"},
		{"min_attribute_item", `{{ (ld|min(attribute='n')).g }}`, "b"},
		{"max_empty", `[{{ empty|max }}]`, "[]"},
		{"slice2", `{% for c in seq|slice(2) %}[{{ c|join(',') }}]{% endfor %}`, "[1,2,3,4][5,6,7]"},
		{"slice3", `{% for c in seq|slice(3) %}[{{ c|join(',') }}]{% endfor %}`, "[1,2,3][4,5][6,7]"},
		{"slice_fill", `{% for c in seq|slice(3, 0) %}[{{ c|join(',') }}]{% endfor %}`, "[1,2,3][4,5,0][6,7,0]"},
		{"slice_len", `{{ l|slice(2)|list|length }}`, "2"},
		{"map_filter", `{{ words|map('upper')|join(',') }}`, "B,A"},
		{"map_filter_args", `{{ ['aa','ba']|map('replace', 'a', 'x')|join(',') }}`, "xx,bx"},
		{"map_first", `{{ dd|dictsort|map('first')|join(',') }}`, "a,b"},
		{"map_attribute_positional", `{{ ld|map('n')|join(',') }}`, "2,1,3"},
		{"select_test", `{{ seq|select('odd')|join(',') }}`, "1,3,5,7"},
		{"select_test_arg", `{{ seq|select('gt', 4)|join(',') }}`, "5,6,7"},
		{"select_alias", `{{ seq|select('>', 4)|join(',') }}`, "5,6,7"},
		{"select_greaterthan", `{{ seq|select('greaterthan', 5)|join(',') }}`, "6,7"},
		{"select_truthy", `{{ [0,1,'',2]|select|join(',') }}`, "1,2"},
		{"reject_test", `{{ seq|reject('odd')|join(',') }}`, "2,4,6"},
		{"reject_truthy", `{{ [0,1,'',2]|reject|list|length }}`, "2"},
		{"selectattr_truthy", `{{ ab|selectattr('a')|list|length }}`, "2"},
		{"selectattr_test", `{{ ld|selectattr('n', 'gt', 1)|map(attribute='n')|join(',') }}`, "2,3"},
		{"selectattr_equalto", `{{ ab|selectattr('a', 'equalto', 1)|list|length }}`, "1"},
		{"selectattr_eqeq", `{{ ab|selectattr('a', '==', 2)|list|length }}`, "1"},
		{"rejectattr_test", `{{ ld|rejectattr('g', 'equalto', 'a')|map(attribute='n')|join(',') }}`, "1"},
		{"rejectattr_truthy", `{{ ab|rejectattr('a')|list|length }}`, "1"},
		{"groupby_attrs", `{% for g in ld|groupby('g') %}{{ g.grouper }}:{{ g.list|map(attribute='n')|join(',') }};{% endfor %}`, "a:2,3;b:1;"},
		{"groupby_unpack", `{% for grouper, items in ld|groupby('g') %}{{ grouper }}{{ items|length }};{% endfor %}`, "a2;b1;"},
		{"groupby_index", `{% for g in ld|groupby('g') %}{{ g[0] }}{% endfor %}`, "ab"},
		{"groupby_numeric_grouper", `{% for g in ld|groupby('n') %}{{ g.grouper is number }};{% endfor %}`, "true;true;true;"},
		{"sort_bools", `{{ bl|sort|join(',') }}`, "false,true,true"},
		{"sort_bool_number", `{{ [true,2,0]|sort|join(',') }}`, "0,true,2"},
		{"groupby_bool", `{% for g in bools|groupby('p') %}{{ g.grouper }}:{{ g.list|map(attribute='i')|join }};{% endfor %}`, "false:b;true:ac;"},
		{"unique_bool_number", `{{ [1,1.0,true]|unique|list|length }}`, "1"},
		{"sort_lists", `{{ nested|sort|map('first')|join(',') }}`, "1,2"},
		{"sort_dotted", `{{ ldd|sort(attribute='c.name')|map(attribute='n')|join(',') }}`, "2,1"},
		{"groupby_dotted", `{% for g in ldd|groupby('c.name') %}{{ g.grouper }};{% endfor %}`, "a;b;"},
		{"sort_integer_path", `{{ pairs|sort(attribute='0')|map(attribute='1')|join(',') }}`, "y,x"},
		{"sort_multi_key", `{{ mk|sort(attribute='g,n')|map(attribute='n')|join(',') }}{{ mk|sort(attribute='g,n')|map(attribute='g')|join(',') }}`, "1,2,1a,a,b"},
		{"groupby_map_grouper", `{{ ld|groupby('g')|map(attribute='grouper')|join(',') }}`, "a,b"},
		{"groupby_selectattr_grouper", `{{ ld|groupby('g')|selectattr('grouper','eq','a')|list|length }}`, "1"},
		{"groupby_sort_grouper", `{{ ld|groupby('g')|sort(attribute='grouper', reverse=true)|map(attribute='grouper')|join(',') }}`, "b,a"},
		{"groupby_map_list_length", `{{ ld|groupby('g')|map(attribute='list')|map('length')|join(',') }}`, "2,1"},
		{"groupby_default", `{% for g in ldm|groupby('g', default='z') %}{{ g.grouper }}:{{ g.list|map(attribute='i')|join }};{% endfor %}`, "a:xz;z:y;"},
		{"map_filter_kwargs", `{{ nums2|map('round', precision=1)|join(',') }}`, "1.2,2.3"},
		{"select_in_numbers", `{{ fl|select('in', [1,2])|join(',') }}`, "1,2"},
		{"selectattr_in", `{{ ld|selectattr('n','in',[1,2])|list|length }}`, "2"},
		{"eq_lists", `{{ [1.0,2.0] is eq([1,2]) }}`, "true"},
		{"selectattr_true_test", `{{ bools|selectattr('p','true')|list|length }}`, "2"},
		{"rejectattr_none_missing", `{{ nd|rejectattr('n','none')|list|length }}`, "2"},
	}
	for _, tc := range outputs {
		t.Run(tc.name, func(t *testing.T) {
			got, err := env.RenderString(tc.template, collectionFilterContext(t))
			if err != nil {
				t.Fatalf("render %q: %v", tc.template, err)
			}
			if got != tc.want {
				t.Errorf("%s: got %q, want %q", tc.template, got, tc.want)
			}
		})
	}

	// Deliberate deviations from Jinja2 (see plans/039): miya-only expectations.
	deviations := []struct {
		name     string
		template string
		want     string
	}{
		{"map_attr_shadows_filter", `{{ ppl|map('title')|join(',') }}`, "Mr,Dr"},
		{"sort_nulls_first", `{{ nl|sort(attribute='n')|map(attribute='i')|join(',') }}`, "b,d,c,a"},
		{"sort_nulls_reverse", `{{ nl|sort(attribute='n', reverse=true)|map(attribute='i')|join(',') }}`, "a,c,b,d"},
		{"groupby_nulls", `{% for g in nl|groupby('n') %}{{ g.grouper }}:{{ g.list|map(attribute='i')|join }};{% endfor %}`, ":bd;1:c;2:a;"},
		{"min_skips_nulls", `{{ (nl|min(attribute='n')).i }}`, "c"},
		{"max_skips_nulls", `{{ (nl|max(attribute='n')).i }}`, "a"},
	}
	for _, tc := range deviations {
		t.Run(tc.name, func(t *testing.T) {
			got, err := env.RenderString(tc.template, collectionFilterContext(t))
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
		{"sort_mixed_error", `{{ [2,'1']|sort|join(',') }}`, "not supported between instances"},
		{"sort_mixed_attribute_error", `{{ mx|sort(attribute='v')|join }}`, `attribute "v"`},
		{"slice_count_too_large", `{{ [1,2]|slice(2000000) }}`, "too large"},
		{"select_unknown_test", `{{ seq|select('nosuchtest')|join(',') }}`, `no test named "nosuchtest"`},
	}
	for _, tc := range errorCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := env.RenderString(tc.template, collectionFilterContext(t))
			if err == nil {
				t.Fatalf("render %q: expected an error containing %q", tc.template, tc.contains)
			}
			if !strings.Contains(err.Error(), tc.contains) {
				t.Errorf("render %q: error %q does not contain %q", tc.template, err, tc.contains)
			}
		})
	}
}
