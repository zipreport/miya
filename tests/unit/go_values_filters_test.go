package miya_test

import (
	"testing"

	miya "github.com/zipreport/miya"
	"github.com/zipreport/miya/runtime"
)

type gvfStatus string
type gvfScore float64
type gvfFlag bool
type gvfKey int
type gvfUser struct {
	Name   string
	Age    int
	Status gvfStatus
}
type gvfFn struct{ Fn func() string }

func TestGoValuesInFilters(t *testing.T) {
	f := 2.5
	data := func() map[string]interface{} {
		return map[string]interface{}{
			"users":  []*gvfUser{{"b", 3, "on"}, {"a", 1, "off"}, {"c", 3, "on"}},
			"u":      &gvfUser{Name: "p", Age: 9},
			"num":    &f,
			"score":  gvfScore(2.5),
			"scores": []gvfScore{2.5, 1.5},
			"flags":  []gvfFlag{true, false, true},
			"flag":   gvfFlag(false),
			"flagt":  gvfFlag(true),
			"st":     gvfStatus("on"),
			"sts":    []gvfStatus{"on", "off", "on"},
			"stcase": []gvfStatus{"a", "B"},
			"stdup":  []gvfStatus{"On", "on"},
			"imap":   map[int]string{2: "two", 1: "one", 10: "ten"},
			"kmap":   map[gvfKey]int{2: 20, 1: 10},
			"smap":   map[gvfStatus]int{"b": 2, "a": 1},
			"ms":     map[string]string{"b": "y", "a": "x"},
			"mixed":  map[interface{}]interface{}{1: "a", "b": "c"},
			"sv":     runtime.SafeValue{Value: ""},
			"fs":     gvfFn{Fn: func() string { return "x" }},
		}
	}
	rows := []struct {
		name, tpl, want string
		wantErr         bool
	}{
		{"ptr_map_attr", `{{ users|map(attribute='Name')|join(',') }}`, "b,a,c", false},
		{"ptr_join_attr", `{{ users|join(',', attribute='Name') }}`, "b,a,c", false},
		{"ptr_attr", `{{ u|attr('Name') }}`, "p", false},
		{"ptr_sort_attr", `{{ users|sort(attribute='Age')|map(attribute='Name')|join }}`, "abc", false},
		{"ptr_sum_attr", `{{ users|sum(attribute='Age') }}`, "7", false},
		{"ptr_selectattr", `{{ users|selectattr('Age','gt',1)|map(attribute='Name')|join }}`, "bc", false},
		{"ptr_groupby", `{{ users|groupby('Age')|map(attribute='grouper')|join(',') }}`, "1,3", false},
		{"ptr_unique_named", `{{ users|map(attribute='Status')|unique|join(',') }}`, "on,off", false},
		{"unique_named_case", `{{ stdup|unique|join(',') }}`, "On", false},
		{"ptr_sort_named", `{{ users|sort(attribute='Status')|map(attribute='Name')|join }}`, "abc", false},
		{"ptr_max_attr", `{{ users|max(attribute='Age')|attr('Name') }}`, "b", false},
		{"ptr_number", `{{ num|int }} {{ num|float }} {{ num|round }} {{ num|abs }}`, "2 2.5 3 2.5", false},
		{"named_number", `{{ score|int }} {{ score|float }} {{ scores|sum }} {{ [score, 1]|sum }}`, "2 2.5 4 3.5", false},
		{"named_bool_sum", `{{ flags|sum }}`, "2", false},
		{"named_bool_test", `{{ flag is false }} {{ flag is true }}`, "true false", false},
		{"named_bool_true", `{{ flagt is true }}`, "true", false},
		{"named_string_compare", `{{ [st, 'a']|max }} {{ sts|unique|join(',') }} {{ sts|sort|join(',') }}`, "on on,off off,on,on", false},
		{"named_sort_case", `{{ stcase|sort|join(',') }}`, "a,B", false},
		{"named_string_repr", `{{ '%r'|format(st) }}`, "'on'", false},
		{"imap_dictsort", `{{ imap|dictsort|map('last')|join(',') }}`, "one,two,ten", false},
		{"imap_dictsort_value", `{{ imap|dictsort(by='value')|map('first')|join(',') }}`, "1,10,2", false},
		{"imap_items", `{{ imap|items|map('first')|join(',') }}`, "1,2,10", false},
		{"imap_methods", `{{ imap.keys()|join(',') }} {{ imap.values()|join(',') }} {{ imap.items()|map('last')|join(',') }} {{ imap.get(2) }} {{ imap.get(3, 'd') }}`, "1,2,10 one,two,ten one,two,ten two d", false},
		{"imap_seq", `{{ imap|list|join(',') }} {{ imap|first }} {{ imap|length }}`, "1,2,10 1 3", false},
		{"imap_seq_filters", `{{ imap|sort|join(',') }} {{ imap|max }} {{ imap|unique|list|length }} {{ imap|reverse|join(',') }}`, "1,2,10 10 3 10,2,1", false},
		{"kmap", `{{ kmap|dictsort|map('last')|join(',') }} {{ kmap|list|join(',') }}`, "10,20 1,2", false},
		{"smap_list", `{{ smap|list|join(',') }}`, "a,b", false},
		{"smap_sort", `{{ smap|sort|join }}`, "ab", false},
		{"smap_join", `{{ smap|join(',') }}`, "a,b", false},
		{"smap_first", `{{ smap|first }}`, "a", false},
		{"imap_for", `{% for k in imap %}{{ k }},{% endfor %}`, "1,2,10,", false},
		{"dictsort_numeric_value", `{{ {'a':10,'b':9}|dictsort(by='value')|map('first')|join(',') }}`, "b,a", false},
		{"smap_dictsort", `{{ ms|dictsort|map('last')|join(',') }}`, "x,y", false},
		{"dictsort_mixed_keys", `{{ mixed|dictsort|length }}`, "", true},
		{"undef_select", `{{ ['a', nope, 'b']|select|join(' ') }}`, "a b", false},
		{"undef_reject", `{{ ['a', nope]|reject|list|length }}`, "1", false},
		{"undef_bool_arg", `{{ [3,1,2]|sort(nope)|join }}`, "123", false},
		{"undef_item_attr", `{{ [ug(), nope]|map(attribute='Name')|join(',') }}`, ",", false},
		{"func_field_hidden", `{{ [fs]|map(attribute='Fn')|join(',') }}`, "", false},
		{"safevalue_select", `{{ [sv]|select|list|length }}`, "0", false},
		{"safevalue_if", `{% if sv %}t{% else %}f{% endif %}`, "f", false},
	}
	for _, r := range rows {
		t.Run(r.name, func(t *testing.T) {
			env := miya.NewEnvironment(miya.WithAutoEscape(false))
			env.AddGlobal("ug", func(args ...interface{}) (interface{}, error) {
				return runtime.NewUndefined("p", runtime.UndefinedSilent, nil), nil
			})
			got, err := renderGo(t, env, r.tpl, data())
			if r.wantErr {
				if err == nil {
					t.Fatalf("want an error, got %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("error: %v", err)
			}
			if got != r.want {
				t.Errorf("got %q, want %q", got, r.want)
			}
		})
	}
}
