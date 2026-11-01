package miya_test

import (
	"fmt"
	"strings"
	"testing"

	miya "github.com/zipreport/miya"
)

func twentyKeys() map[string]interface{} {
	m := make(map[string]interface{}, 20)
	for i := 0; i < 20; i++ {
		m[fmt.Sprintf("k%02d", i)] = i
	}
	return m
}

func seq(f func(i int) string) string {
	var b strings.Builder
	for i := 0; i < 20; i++ {
		b.WriteString(f(i))
	}
	return b.String()
}

func renderNoPanic(t *testing.T, template string, data map[string]interface{}) (out string, err error) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("panicked: %v", r)
		}
	}()
	env := miya.NewEnvironment(miya.WithAutoEscape(false))
	return env.RenderString(template, miya.NewContextFrom(data))
}

func TestDictKeyIteration(t *testing.T) {
	keysJoined := strings.TrimSuffix(seq(func(i int) string { return fmt.Sprintf("k%02d,", i) }), ",")
	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		want     string
	}{
		{"for_yields_keys", `{% for k in d %}{{ k }},{% endfor %}`, map[string]interface{}{"d": map[string]interface{}{"a": 10}}, "a,"},
		{"for_key_lookup", `{% for k in d %}{{ k }}={{ d[k] }};{% endfor %}`, map[string]interface{}{"d": map[string]interface{}{"a": 1}}, "a=1;"},
		{"for_sorted", `{% for k in d %}{{ k }},{% endfor %}`, map[string]interface{}{"d": twentyKeys()}, keysJoined + ","},
		{"two_vars_sorted", `{% for k, v in d %}{{ k }}={{ v }};{% endfor %}`, map[string]interface{}{"d": twentyKeys()}, seq(func(i int) string { return fmt.Sprintf("k%02d=%d;", i, i) })},
		{"items_sorted", `{% for k, v in d.items() %}{{ k }}={{ v }};{% endfor %}`, map[string]interface{}{"d": twentyKeys()}, seq(func(i int) string { return fmt.Sprintf("k%02d=%d;", i, i) })},
		{"items_comprehension", `{{ [p[0] for p in d.items()]|join(",") }}`, map[string]interface{}{"d": twentyKeys()}, keysJoined},
		{"keys_sorted", `{{ d.keys()|join(",") }}`, map[string]interface{}{"d": twentyKeys()}, keysJoined},
		{"values_in_key_order", `{{ d.keys()|join(",") }}|{{ d.values()|join(",") }}`, map[string]interface{}{"d": map[string]interface{}{"b": 1, "a": 2, "c": 3}}, "a,b,c|2,1,3"},
		{"comprehension_keys", `{{ [k for k in d]|join(",") }}`, map[string]interface{}{"d": twentyKeys()}, keysJoined},
		{"set_unpacks_keys", `{% set a, b = d %}{{ a }}{{ b }}`, map[string]interface{}{"d": map[string]interface{}{"y": 1, "x": 2}}, "xy"},
		{"codepoint_order", `{% for k in d %}{{ k }},{% endfor %}`, map[string]interface{}{"d": map[string]interface{}{"b": 1, "B": 2, "10": 3, "9": 4}}, "10,9,B,b,"},
		{"dictsort_default", `{% for k, v in d|dictsort %}{{ k }}{% endfor %}`, map[string]interface{}{"d": map[string]interface{}{"b": 1, "B": 2, "a": 3, "A": 4}}, "AaBb"},
		{"dictsort_case_sensitive", `{% for k, v in d|dictsort(true) %}{{ k }}{% endfor %}`, map[string]interface{}{"d": map[string]interface{}{"b": 1, "B": 2, "a": 3, "A": 4}}, "ABab"},
		{"dictsort_by_value", `{% for k, v in d|dictsort(false, "value") %}{{ k }}{% endfor %}`, map[string]interface{}{"d": map[string]interface{}{"x": "b", "y": "B", "z": "a"}}, "zyx"},
		{"dictsort_reverse", `{% for k, v in d|dictsort(false, "key", true) %}{{ k }}{% endfor %}`, map[string]interface{}{"d": map[string]interface{}{"b": 1, "B": 2, "a": 3, "A": 4}}, "bBaA"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for n := 0; n < 20; n++ {
				got, err := renderNoPanic(t, tc.template, tc.data)
				if err != nil {
					t.Fatalf("render %q: %v", tc.template, err)
				}
				if got != tc.want {
					t.Fatalf("render %d: got %q, want %q", n, got, tc.want)
				}
			}
		})
	}
}

func TestUnicodeStrings(t *testing.T) {
	const cucar = "açúcar"
	tests := []struct {
		name     string
		template string
		s        string
		want     string
	}{
		{"index", `{{ s[1] }}`, cucar, "ç"},
		{"index_negative", `{{ s[-1] }}`, "açú", "ú"},
		{"index_past_end", `[{{ s[7] }}]`, cucar, "[]"},
		{"slice", `{{ s[0:3] }}`, cucar, "açú"},
		{"slice_tail", `{{ s[-3:] }}`, cucar, "car"},
		{"slice_reverse", `{{ s[::-1] }}`, cucar, "racúça"},
		{"slice_step", `{{ s[1::2] }}`, cucar, "çcr"},
		{"loop_items", `{% for c in s %}[{{ c }}]{% endfor %}`, "açú", "[a][ç][ú]"},
		{"loop_length", `{% for c in s %}{{ loop.length }}{% endfor %}`, cucar, "666666"},
		{"first", `{{ s|first }}`, "ça", "ç"},
		{"last", `{{ s|last }}`, "señor ñ", "ñ"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := renderNoPanic(t, tc.template, map[string]interface{}{"s": tc.s})
			if err != nil {
				t.Fatalf("render %q: %v", tc.template, err)
			}
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestNegativeStepSlices(t *testing.T) {
	data := map[string]interface{}{
		"s":  "açúcar",
		"a":  "abc",
		"l":  []interface{}{1, 2, 3},
		"sl": []string{"x", "y", "z"},
	}
	tests := []struct {
		name     string
		template string
		want     string
	}{
		{"str_start_past_end", `{{ s[10::-1] }}`, "racúça"},
		{"str_stop_before_start", `{{ s[:-10:-1] }}`, "racúça"},
		{"str_start_before_begin", `[{{ s[-10::-1] }}]`, "[]"},
		{"ascii_start_past_end", `{{ a[10::-1] }}`, "cba"},
		{"list_start_past_end", `{{ l[10::-1]|join(",") }}`, "3,2,1"},
		{"list_stop_before_start", `{{ l[:-10:-1]|join(",") }}`, "3,2,1"},
		{"list_start_before_begin", `[{{ l[-10::-1]|join(",") }}]`, "[]"},
		{"reflect_start_past_end", `{{ sl[10::-1]|join(",") }}`, "z,y,x"},
		{"str_inner_negative_step", `{{ s[4:1:-1] }}`, "acú"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := renderNoPanic(t, tc.template, data)
			if err != nil {
				t.Fatalf("render %q: %v", tc.template, err)
			}
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}
