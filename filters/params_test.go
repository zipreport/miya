package filters

import (
	"reflect"
	"strings"
	"sync"
	"testing"
)

func TestFilterParamBinding(t *testing.T) {
	params := []Param{
		{Name: "a", Required: true},
		{Name: "b", Default: "B"},
		{Name: "c", Unsupported: true},
		{Name: "d", Default: 4},
	}

	tests := []struct {
		name    string
		args    []interface{}
		kw      map[string]interface{}
		want    []interface{}
		wantErr string
	}{
		{"gap_default", []interface{}{1}, map[string]interface{}{"b": 2}, []interface{}{1, 2}, ""},
		{"gap_fill", []interface{}{1}, map[string]interface{}{"d": 9}, []interface{}{1, "B", nil, 9}, ""},
		{"missing_required", nil, map[string]interface{}{"b": 2}, nil, `missing required argument "a"`},
		{"unexpected", nil, map[string]interface{}{"x": 1}, nil, `unexpected keyword argument "x"`},
		{"unsupported", []interface{}{1}, map[string]interface{}{"c": 1}, nil, "not supported"},
		{"duplicate", []interface{}{1, 2}, map[string]interface{}{"b": 3}, nil, "multiple values"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := bindKwargs("f", params, tc.args, tc.kw)
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("got err %v, want containing %q", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("got %#v, want %#v", got, tc.want)
			}
		})
	}

	t.Run("deterministic", func(t *testing.T) {
		for i := 0; i < 20; i++ {
			_, err := bindKwargs("f", params, []interface{}{1}, map[string]interface{}{"y": 1, "x": 1})
			if err == nil || !strings.Contains(err.Error(), `"x"`) {
				t.Fatalf("run %d: got %v, want error naming \"x\"", i, err)
			}
		}
	})

	fn := func(v interface{}, args ...interface{}) (interface{}, error) { return args, nil }

	t.Run("unregister_drops_params", func(t *testing.T) {
		r := NewRegistry()
		r.Unregister("sort")
		if err := r.Register("sort", fn); err != nil {
			t.Fatal(err)
		}
		_, err := r.ApplyKwargs("sort", []interface{}{1}, nil, map[string]interface{}{"reverse": true})
		if err == nil || !strings.Contains(err.Error(), "does not accept keyword arguments") {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("setparams_rejects_kwargs_filter", func(t *testing.T) {
		err := NewRegistry().SetParams("map", Param{Name: "x"})
		if err == nil || !strings.Contains(err.Error(), "takes keyword arguments directly") {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("register_with_params", func(t *testing.T) {
		r := NewRegistry()
		if err := r.RegisterWithParams("tag", fn, Param{Name: "a"}); err != nil {
			t.Fatal(err)
		}
		got, err := r.ApplyKwargs("tag", "v", nil, map[string]interface{}{"a": 1})
		if err != nil {
			t.Fatal(err)
		}
		if want := []interface{}{1}; !reflect.DeepEqual(got, want) {
			t.Errorf("got %#v, want %#v", got, want)
		}
	})
}

func TestFilterParamsConcurrent(t *testing.T) {
	r := NewRegistry()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			if err := r.SetParams("join", Param{Name: "d"}, Param{Name: "attribute"}); err != nil {
				t.Error(err)
				return
			}
		}
	}()
	for g := 0; g < 4; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 1000; i++ {
				if _, err := r.ApplyKwargs("join", []interface{}{1, 2}, nil, map[string]interface{}{"d": "-"}); err != nil {
					t.Error(err)
					return
				}
			}
		}()
	}
	wg.Wait()
}
