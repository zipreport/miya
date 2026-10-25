package miya_test

import (
	"encoding/json"
	"math"
	"testing"

	miya "github.com/zipreport/miya"
)

type Money float64

func numbersContext() miya.Context {
	return miya.NewContextFrom(map[string]interface{}{
		"zero": float64(0),
		"one":  float64(1),
		"i32":  int32(5),
		"u8":   uint8(7),
		"f32":  float32(0),
		"big1": int64(9007199254740993),
		"big2": int64(9007199254740992),
		"m":    Money(5),
		"nan":  math.NaN(),
	})
}

func TestNumericComparison(t *testing.T) {
	env := miya.NewEnvironment()

	tests := []struct {
		name     string
		expr     string
		expected string
	}{
		{"r01", `(1 - 1) > 0`, "F"},
		{"r02", `zero > 0`, "F"},
		{"r03", `(one - 1) > 0`, "F"},
		{"r04", `(2 - 1) <= 1`, "T"},
		{"r05", `zero == 0`, "T"},
		{"r06", `zero != 0`, "F"},
		{"r07", `0 in [zero]`, "T"},
		{"r08", `(1 - 1) == 0`, "T"},
		{"r09", `[1, 2] == [1.0, 2.0]`, "T"},
		{"r10", `i32 == 5`, "T"},
		{"r11", `i32 > 4`, "T"},
		{"r12", `u8 < 3`, "F"},
		{"r13", `i32 == u8`, "F"},
		{"r14", `f32 == 0`, "T"},
		{"r15", `m > 4`, "T"},
		{"r16", `big1 > big2`, "T"},
		{"r17", `big1 == big2`, "F"},
		{"r18", `"10" < "9"`, "T"},
		{"r19", `5 >= 5`, "T"},
		{"r20", `0 > 0`, "F"},
		{"r21", `-1 < u8`, "T"},
		{"r22", `u8 > -1`, "T"},
		{"r23", `nan > 1`, "F"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tpl := `{% if ` + tc.expr + ` %}T{% else %}F{% endif %}`
			result, err := env.RenderString(tpl, numbersContext())
			if err != nil {
				t.Fatalf("render %q: %v", tpl, err)
			}
			if result != tc.expected {
				t.Errorf("%s: got %q, want %q", tc.expr, result, tc.expected)
			}
		})
	}
}

func TestNumericComparisonStringNumberErrors(t *testing.T) {
	env := miya.NewEnvironment()
	_, err := env.RenderString(`{% if "5" < 6 %}T{% endif %}`, miya.NewContext())
	if err == nil {
		t.Fatal(`expected an error comparing "5" < 6`)
	}
}

func TestNumberOutput(t *testing.T) {
	env := miya.NewEnvironment(miya.WithAutoEscape(true))

	tests := []struct {
		name     string
		template string
		value    interface{}
		expected string
	}{
		{"whole_float", `{{ v }}`, float64(10023), "10023"},
		{"year", `{{ v }}`, 2024.0, "2024"},
		{"1234.5", `{{ v }}`, 1234.5, "1234.5"},
		{"pi", `{{ v }}`, 3.14159, "3.14159"},
		{"half", `{{ v }}`, 0.5, "0.5"},
		{"1234567.89", `{{ v }}`, 1234567.89, "1234567.89"},
		{"neg", `{{ v }}`, -1234567.5, "-1234567.5"},
		{"1e20", `{{ v }}`, 1e20, "1e+20"},
		{"1e-05", `{{ v }}`, 0.00001, "1e-05"},
		{"float32", `{{ v }}`, float32(1500), "1500"},
		{"int", `{{ v }}`, int(10023), "10023"},
		{"concat", `{{ "T: " ~ v }}`, 2500000.0, "T: 2500000"},
		{"string_filter", `{{ v|string }}`, 1234567.89, "1234567.89"},
		{"join_filter", `{{ [v, 1]|join(",") }}`, 1234567.89, "1234567.89,1"},
		{"escape_filter", `{{ v|e }}`, 1234567.89, "1234567.89"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := miya.NewContextFrom(map[string]interface{}{"v": tc.value})
			result, err := env.RenderString(tc.template, ctx)
			if err != nil {
				t.Fatalf("render: %v", err)
			}
			if result != tc.expected {
				t.Errorf("got %q, want %q", result, tc.expected)
			}
		})
	}

	t.Run("json_decoded", func(t *testing.T) {
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(`{"n": 10023}`), &data); err != nil {
			t.Fatal(err)
		}
		result, err := env.RenderString(`{{ n }}`, miya.NewContextFrom(data))
		if err != nil {
			t.Fatalf("render: %v", err)
		}
		if result != "10023" {
			t.Errorf("got %q, want %q", result, "10023")
		}
	})
}
