package miya_test

import (
	miya "github.com/zipreport/miya"
	"testing"
)

func TestBuiltinTestsRuntime(t *testing.T) {
	env := miya.NewEnvironment()

	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected string
	}{
		// Basic existence and type tests
		{"defined test", `{{ name is defined }}`, map[string]interface{}{"name": "John"}, "true"},
		{"undefined test", `{{ undefined_var is defined }}`, map[string]interface{}{}, "false"},
		{"none test", `{{ value is none }}`, map[string]interface{}{"value": nil}, "true"},
		{"string test", `{{ value is string }}`, map[string]interface{}{"value": "hello"}, "true"},
		{"number test", `{{ value is number }}`, map[string]interface{}{"value": 42}, "true"},
		{"integer test", `{{ value is integer }}`, map[string]interface{}{"value": 42}, "true"},
		{"float test", `{{ value is float }}`, map[string]interface{}{"value": 3.14}, "true"},
		{"boolean test", `{{ value is boolean }}`, map[string]interface{}{"value": true}, "true"},

		// Sequence and mapping tests
		{"sequence test", `{{ value is sequence }}`, map[string]interface{}{"value": []string{"a", "b"}}, "true"},
		{"mapping test", `{{ value is mapping }}`, map[string]interface{}{"value": map[string]string{"a": "b"}}, "true"},
		{"iterable test", `{{ value is iterable }}`, map[string]interface{}{"value": []string{"a", "b"}}, "true"},

		// Numeric tests
		{"even test", `{{ value is even }}`, map[string]interface{}{"value": 4}, "true"},
		{"odd test", `{{ value is odd }}`, map[string]interface{}{"value": 5}, "true"},
		{"divisibleby test", `{{ value is divisibleby(3) }}`, map[string]interface{}{"value": 9}, "true"},

		// String tests
		{"lower test", `{{ value is lower }}`, map[string]interface{}{"value": "hello"}, "true"},
		{"upper test", `{{ value is upper }}`, map[string]interface{}{"value": "HELLO"}, "true"},

		// Comparison tests
		{"equalto test", `{{ value is equalto(42) }}`, map[string]interface{}{"value": 42}, "true"},
		{"sameas test", `{{ value is sameas(value2) }}`, map[string]interface{}{"value": 42, "value2": 42}, "true"},

		// Negated tests
		{"not defined test", `{{ undefined_var is not defined }}`, map[string]interface{}{}, "true"},
		{"not none test", `{{ value is not none }}`, map[string]interface{}{"value": "hello"}, "true"},
		{"not even test", `{{ value is not even }}`, map[string]interface{}{"value": 5}, "true"},

		// Tests that should work but might be missing
		{"callable test", `{{ func is callable }}`, map[string]interface{}{"func": func() string { return "test" }}, "true"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := miya.NewContext()
			for k, v := range tt.data {
				ctx.Set(k, v)
			}

			result, err := env.RenderString(tt.template, ctx)

			if err != nil {
				t.Fatalf("Error rendering template for %s: %v", tt.name, err)
			}

			if result != tt.expected {
				t.Errorf("%s: Expected %q, got %q", tt.name, tt.expected, result)
			}
		})
	}
}
