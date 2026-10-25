package miya_test

import (
	"fmt"
	"math"
	"strings"
	"sync"
	"testing"
	"time"

	miya "github.com/zipreport/miya"
)

type gvStatus string
type gvFlag bool
type gvCount int
type gvUser struct {
	Name  string
	Email string
	age   int
}
type gvInner struct{ X int }
type gvOuter struct{ *gvInner }
type gvLocked struct {
	*sync.Mutex
	Name string
}
type gvFns struct{ Fn func() string }
type gvSelf *gvSelf // a pointer type that points to itself

func (u gvUser) DisplayName() string { return "D:" + u.Name }

func renderGo(t *testing.T, env *miya.Environment, tpl string, data map[string]interface{}) (out string, err error) {
	t.Helper()
	type result struct {
		out string
		err error
	}
	done := make(chan result, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				done <- result{"", fmt.Errorf("panic: %v", r)}
			}
		}()
		tmpl, err := env.FromString(tpl)
		if err != nil {
			done <- result{"", err}
			return
		}
		o, err := tmpl.Render(miya.NewContextFrom(data))
		done <- result{o, err}
	}()
	select {
	case res := <-done:
		return res.out, res.err
	case <-time.After(5 * time.Second):
		t.Fatalf("render did not finish (infinite loop?)")
		return "", nil
	}
}

func gvIntKeys() (map[int]string, string) {
	m := map[int]string{}
	var sb strings.Builder
	for i := 1; i <= 20; i++ {
		m[i] = fmt.Sprint("v", i)
		fmt.Fprintf(&sb, "%d,", i)
	}
	return m, sb.String()
}

func gvMixedKeys() map[interface{}]interface{} {
	return map[interface{}]interface{}{"1": "s", 1: "n", true: "b", "a": "x"}
}

func gvNanKeys() map[float64]string {
	m := map[float64]string{}
	m[1] = "a"
	m[math.NaN()] = "b"
	m[math.NaN()] = "c"
	return m
}

func TestGoValues(t *testing.T) {
	s := "x"
	n := 5
	b := false
	var self gvSelf
	self = &self
	var iself interface{}
	iself = &iself
	intKeys, intKeysWant := gvIntKeys()

	tests := []struct {
		name  string
		tpl   string
		data  map[string]interface{}
		want  string
		times int
	}{
		{name: "map_attr", tpl: `{{ m.a }}`, data: map[string]interface{}{"m": map[string]int{"a": 1}}, want: "1"},
		{name: "map_item", tpl: `{{ m['a'] }}`, data: map[string]interface{}{"m": map[string]int{"a": 1}}, want: "1"},
		{name: "map_iter_keys", tpl: `{% for k in m %}{{ k }},{% endfor %}`, data: map[string]interface{}{"m": map[string]int{"b": 2, "a": 1}}, want: "a,b,", times: 20},
		{name: "map_iter_pairs", tpl: `{% for k, v in m %}{{ k }}={{ v }};{% endfor %}`, data: map[string]interface{}{"m": map[string]int{"b": 2, "a": 1}}, want: "a=1;b=2;", times: 20},
		{name: "map_keys_method", tpl: `{{ m.keys()|join(',') }}`, data: map[string]interface{}{"m": map[string]int{"b": 2, "a": 1}}, want: "a,b"},
		{name: "map_named_key_attr", tpl: `{{ m.a }}`, data: map[string]interface{}{"m": map[gvStatus]string{"a": "A"}}, want: "A"},
		{name: "map_named_key_item", tpl: `{{ m['a'] }}`, data: map[string]interface{}{"m": map[gvStatus]string{"a": "A"}}, want: "A"},
		{name: "map_int_key_item", tpl: `{{ m[1] }}`, data: map[string]interface{}{"m": map[int]string{1: "one", 2: "two"}}, want: "one"},
		{name: "map_int_key_iter", tpl: `{% for k in m %}{{ k }},{% endfor %}`, data: map[string]interface{}{"m": intKeys}, want: intKeysWant, times: 20},
		{name: "map_mixed_key_iter", tpl: `{% for k in m %}{{ k }},{% endfor %}`, data: map[string]interface{}{"m": gvMixedKeys()}, want: "true,1,1,a,", times: 20},
		{name: "map_nan_keys", tpl: `{% for k in m %}{{ k }},{% endfor %}`, data: map[string]interface{}{"m": gvNanKeys()}, want: "1,NaN,NaN,", times: 20},
		{name: "map_struct_value", tpl: `{{ m.a.Name }}`, data: map[string]interface{}{"m": map[string]gvUser{"a": {Name: "Ann"}}}, want: "Ann"},
		{name: "yaml_attr", tpl: `{{ m.name }}`, data: map[string]interface{}{"m": map[interface{}]interface{}{"name": "Y"}}, want: "Y"},
		{name: "yaml_items", tpl: `{% for k, v in m.items() %}{{ k }}={{ v }};{% endfor %}`, data: map[string]interface{}{"m": map[interface{}]interface{}{"b": 2, "a": 1}}, want: "a=1;b=2;", times: 20},
		{name: "unhashable_item", tpl: `[{{ m[[1]] }}]`, data: map[string]interface{}{"m": map[interface{}]interface{}{"a": 1}}, want: "[]"},
		{name: "unhashable_in_map", tpl: `{{ [1] in m }}`, data: map[string]interface{}{"m": map[interface{}]interface{}{"a": 1}}, want: "false"},
		{name: "uint_key_negative", tpl: `[{{ m[-1] }}]`, data: map[string]interface{}{"m": map[uint64]string{math.MaxUint64: "max"}}, want: "[]"},
		{name: "slice_neg_index", tpl: `{{ l[-1] }}`, data: map[string]interface{}{"l": []int{3, 1, 2}}, want: "2"},
		{name: "array_neg_index", tpl: `{{ a[-1] }}`, data: map[string]interface{}{"a": [3]int{7, 8, 9}}, want: "9"},
		{name: "slice_join", tpl: `{{ l|join(',') }}`, data: map[string]interface{}{"l": []int{3, 1, 2}}, want: "3,1,2"},
		{name: "array_join", tpl: `{{ a|join(',') }}`, data: map[string]interface{}{"a": [3]int{7, 8, 9}}, want: "7,8,9"},
		{name: "slice_reverse", tpl: `{{ l|reverse|join(',') }}`, data: map[string]interface{}{"l": []int{3, 1, 2}}, want: "2,1,3"},
		{name: "slice_batch", tpl: `{{ l|batch(2)|list|length }}`, data: map[string]interface{}{"l": []int{3, 1, 2}}, want: "2"},
		{name: "join_attr_typed_map", tpl: `{{ lm|join(',', 'a') }}`, data: map[string]interface{}{"lm": []map[string]int{{"a": 1}, {"a": 2}}}, want: "1,2"},
		{name: "dictsort_typed", tpl: `{% for k, v in m|dictsort %}{{ k }}{{ v }}{% endfor %}`, data: map[string]interface{}{"m": map[string]int{"b": 2, "a": 1}}, want: "a1b2"},
		{name: "items_filter_typed", tpl: `{% for k, v in m|items %}{{ k }}{{ v }}{% endfor %}`, data: map[string]interface{}{"m": map[string]int{"b": 2, "a": 1}}, want: "a1b2"},
		{name: "int32_add", tpl: `{{ n + 1 }}`, data: map[string]interface{}{"n": int32(5)}, want: "6"},
		{name: "int32_sub", tpl: `{{ n - 1 }}`, data: map[string]interface{}{"n": int32(5)}, want: "4"},
		{name: "int32_mul", tpl: `{{ n * 2 }}`, data: map[string]interface{}{"n": int32(5)}, want: "10"},
		{name: "int32_div", tpl: `{{ n / 2 }}`, data: map[string]interface{}{"n": int32(5)}, want: "2.5"},
		{name: "int32_pow", tpl: `{{ n ** 2 }}`, data: map[string]interface{}{"n": int32(5)}, want: "25"},
		{name: "int32_neg", tpl: `{{ -n }}`, data: map[string]interface{}{"n": int32(5)}, want: "-5"},
		{name: "uint8_add", tpl: `{{ n + 1 }}`, data: map[string]interface{}{"n": uint8(5)}, want: "6"},
		{name: "float32_add", tpl: `{{ n + 1 }}`, data: map[string]interface{}{"n": float32(5.5)}, want: "6.5"},
		{name: "named_int_add", tpl: `{{ n + 1 }}`, data: map[string]interface{}{"n": gvCount(5)}, want: "6"},
		{name: "bool_add", tpl: `{{ true + 1 }}`, want: "2"},
		{name: "bool_mul", tpl: `{{ true * 3 }}`, want: "3"},
		{name: "bool_eq_int", tpl: `{{ true == 1 }}|{{ false == 0 }}`, want: "true|true"},
		{name: "ptr_string_print", tpl: `{{ p }}`, data: map[string]interface{}{"p": &s}, want: "x"},
		{name: "ptr_string_filter", tpl: `{{ p|upper }}`, data: map[string]interface{}{"p": &s}, want: "X"},
		{name: "ptr_string_eq", tpl: `{{ p == 'x' }}`, data: map[string]interface{}{"p": &s}, want: "true"},
		{name: "ptr_string_in", tpl: `{{ 'x' in p }}`, data: map[string]interface{}{"p": &s}, want: "true"},
		{name: "ptr_string_is_string", tpl: `{{ p is string }}`, data: map[string]interface{}{"p": &s}, want: "true"},
		{name: "ptr_int_add", tpl: `{{ p + 1 }}`, data: map[string]interface{}{"p": &n}, want: "6"},
		{name: "ptr_int_cmp", tpl: `{{ p > 3 }}`, data: map[string]interface{}{"p": &n}, want: "true"},
		{name: "ptr_bool_false", tpl: `{% if p %}T{% else %}F{% endif %}`, data: map[string]interface{}{"p": &b}, want: "F"},
		{name: "nil_ptr_print", tpl: `[{{ p }}]`, data: map[string]interface{}{"p": (*string)(nil)}, want: "[]"},
		{name: "nil_ptr_falsy", tpl: `{% if p %}T{% else %}F{% endif %}`, data: map[string]interface{}{"p": (*string)(nil)}, want: "F"},
		{name: "nil_ptr_is_none", tpl: `{{ p is none }}`, data: map[string]interface{}{"p": (*string)(nil)}, want: "true"},
		{name: "nil_struct_ptr_none", tpl: `{{ u is none }}`, data: map[string]interface{}{"u": (*gvUser)(nil)}, want: "true"},
		{name: "cyclic_ptr_print", tpl: `[{{ p }}]`, data: map[string]interface{}{"p": self}, want: "[]"},
		{name: "cyclic_iface_attr", tpl: `[{{ x.a }}]`, data: map[string]interface{}{"x": iself}, want: "[]"},
		{name: "named_bool_false_falsy", tpl: `{% if f %}T{% else %}F{% endif %}`, data: map[string]interface{}{"f": gvFlag(false)}, want: "F"},
		{name: "named_bool_not", tpl: `{{ not f }}`, data: map[string]interface{}{"f": gvFlag(false)}, want: "true"},
		{name: "named_bool_eq", tpl: `{{ f == false }}`, data: map[string]interface{}{"f": gvFlag(false)}, want: "true"},
		{name: "named_bool_default_bool", tpl: `[{{ f|default('d', true) }}]`, data: map[string]interface{}{"f": gvFlag(false)}, want: "[d]"},
		{name: "named_string_empty_falsy", tpl: `{% if st %}T{% else %}F{% endif %}`, data: map[string]interface{}{"st": gvStatus("")}, want: "F"},
		{name: "named_string_eq", tpl: `{{ st == 'ok' }}`, data: map[string]interface{}{"st": gvStatus("ok")}, want: "true"},
		{name: "named_string_concat", tpl: `{{ st + st }}`, data: map[string]interface{}{"st": gvStatus("ok")}, want: "okok"},
		{name: "named_string_is_string", tpl: `{{ st is string }}`, data: map[string]interface{}{"st": gvStatus("ok")}, want: "true"},
		{name: "named_int_zero_falsy", tpl: `{% if c %}T{% else %}F{% endif %}`, data: map[string]interface{}{"c": gvCount(0)}, want: "F"},
		{name: "named_int_is_number", tpl: `{{ c is number }}|{{ c is integer }}`, data: map[string]interface{}{"c": gvCount(5)}, want: "true|true"},
		{name: "struct_lower_field", tpl: `{{ u.name }}`, data: map[string]interface{}{"u": gvUser{Name: "Ann"}}, want: "Ann"},
		{name: "struct_lower_field2", tpl: `{{ u.email }}`, data: map[string]interface{}{"u": gvUser{Email: "a@x"}}, want: "a@x"},
		{name: "struct_ptr_lower_field", tpl: `{{ u.name }}`, data: map[string]interface{}{"u": &gvUser{Name: "Ann"}}, want: "Ann"},
		{name: "struct_lower_is_defined", tpl: `{{ u.name is defined }}`, data: map[string]interface{}{"u": gvUser{Name: "Ann"}}, want: "true"},
		{name: "nil_embedded_attr", tpl: `[{{ o.X }}]`, data: map[string]interface{}{"o": gvOuter{}}, want: "[]"},
		{name: "nil_embedded_join_attr", tpl: `[{{ l|join(',', 'X') }}]`, data: map[string]interface{}{"l": []interface{}{gvOuter{}}}, want: "[]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := miya.NewEnvironment(miya.WithAutoEscape(false))
			times := tt.times
			if times == 0 {
				times = 1
			}
			for i := 0; i < times; i++ {
				out, err := renderGo(t, env, tt.tpl, tt.data)
				if err != nil {
					t.Fatalf("render %d: unexpected error: %v", i, err)
				}
				if out != tt.want {
					t.Fatalf("render %d: got %q, want %q", i, out, tt.want)
				}
			}
		})
	}

	t.Run("typed_nil_like_nil", func(t *testing.T) {
		env := miya.NewEnvironment(miya.WithAutoEscape(false))
		tpl := `[{{ p|default('d') }}]|{% if p %}T{% else %}F{% endif %}|{{ p is none }}`
		typed, err := renderGo(t, env, tpl, map[string]interface{}{"p": (*string)(nil)})
		if err != nil {
			t.Fatalf("typed nil: %v", err)
		}
		untyped, err := renderGo(t, env, tpl, map[string]interface{}{"p": nil})
		if err != nil {
			t.Fatalf("nil: %v", err)
		}
		if typed != untyped {
			t.Fatalf("typed nil %q differs from nil %q", typed, untyped)
		}
	})
}

func TestGoValueGuards(t *testing.T) {
	f := func() string { return "called" }

	tests := []struct {
		name   string
		tpl    string
		data   map[string]interface{}
		want   string
		errHas string
	}{
		{name: "map_missing_key", tpl: `[{{ m.zz }}]`, data: map[string]interface{}{"m": map[string]int{"a": 1}}, want: "[]"},
		{name: "slice_neg_out", tpl: `[{{ l[-4] }}]`, data: map[string]interface{}{"l": []int{3, 1, 2}}, want: "[]"},
		{name: "struct_unexported", tpl: `[{{ u.age }}]`, data: map[string]interface{}{"u": gvUser{age: 3}}, want: "[]"},
		{name: "method_not_callable", tpl: `{{ u.DisplayName() }}`, data: map[string]interface{}{"u": gvUser{Name: "Ann"}}, errHas: "cannot call"},
		{name: "method_not_attr", tpl: `[{{ u.DisplayName }}]`, data: map[string]interface{}{"u": gvUser{Name: "Ann"}}, want: "[]"},
		{name: "func_field_exact_callable", tpl: `{{ o.Fn() }}`, data: map[string]interface{}{"o": gvFns{Fn: f}}, want: "called"},
		{name: "func_field_lower_not_callable", tpl: `{{ o.fn() }}`, data: map[string]interface{}{"o": gvFns{Fn: f}}, errHas: "cannot call"},
		{name: "func_typed_map_not_callable", tpl: `{{ m.f() }}`, data: map[string]interface{}{"m": map[string]func() string{"f": f}}, errHas: "cannot call"},
		{name: "ptr_func_not_iterable", tpl: `{% for x in pf %}{{ x }}{% endfor %}`, data: map[string]interface{}{"pf": &f}, errHas: "not iterable"},
		{name: "embedded_lower_name_hidden", tpl: `[{{ lk.mutex }}]`, data: map[string]interface{}{"lk": gvLocked{Mutex: &sync.Mutex{}, Name: "n"}}, want: "[]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := miya.NewEnvironment(miya.WithAutoEscape(false))
			out, err := renderGo(t, env, tt.tpl, tt.data)
			if tt.errHas != "" {
				if err == nil || !strings.Contains(err.Error(), tt.errHas) {
					t.Fatalf("got out=%q err=%v, want error containing %q", out, err, tt.errHas)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if out != tt.want {
				t.Fatalf("got %q, want %q", out, tt.want)
			}
		})
	}
}

func gvRows(n int) []interface{} {
	rows := make([]interface{}, n)
	for i := range rows {
		rows[i] = map[string]interface{}{"a": i, "b": "x", "c": 1.5, "d": true, "e": "y"}
	}
	return rows
}

func gvBench(b *testing.B, tpl string, data map[string]interface{}) {
	env := miya.NewEnvironment(miya.WithAutoEscape(false))
	tmpl, err := env.FromString(tpl)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := tmpl.Render(miya.NewContextFrom(data)); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGoValuesJSONRows(b *testing.B) {
	gvBench(b, `{% for r in rows %}{{ r.a }}{{ r.b }}{{ r.c }}{{ r.d }}{{ r.e }}{% endfor %}`,
		map[string]interface{}{"rows": gvRows(10000)})
}

func BenchmarkGoValuesStructPtrRows(b *testing.B) {
	type row struct{ A, B, C, D, E int }
	rows := make([]interface{}, 10000)
	for i := range rows {
		rows[i] = &row{i, i, i, i, i}
	}
	gvBench(b, `{% for r in rows %}{{ r.A }}{{ r.B }}{{ r.C }}{{ r.D }}{{ r.E }}{% endfor %}`,
		map[string]interface{}{"rows": rows})
}

func BenchmarkGoValuesTypedMapMiss(b *testing.B) {
	maps := make([]interface{}, 1000)
	for i := range maps {
		m := make(map[string]int, 50)
		for j := 0; j < 50; j++ {
			m[fmt.Sprint("k", j)] = j
		}
		maps[i] = m
	}
	gvBench(b, `{% for m in maps %}{{ m.zz is defined }}{% endfor %}`,
		map[string]interface{}{"maps": maps})
}
