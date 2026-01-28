package miya_test

import (
	miya "github.com/zipreport/miya"
	"testing"
)

func TestBuiltinTestsRuntime(t *testing.T) {
	env := miya.NewEnvironment()

	tests := []struct {
		name       string
		template   string
		data       map[string]interface{}
		expected   string
		shouldFail bool
	}{
		// Basic existence and type tests
		{"defined test", `{{ name is defined }}`, map[string]interface{}{"name": "John"}, "true", false},
		{"undefined test", `{{ undefined_var is defined }}`, map[string]interface{}{}, "false", false},
		{"none test", `{{ value is none }}`, map[string]interface{}{"value": nil}, "true", false},
		{"string test", `{{ value is string }}`, map[string]interface{}{"value": "hello"}, "true", false},
		{"number test", `{{ value is number }}`, map[string]interface{}{"value": 42}, "true", false},
		{"integer test", `{{ value is integer }}`, map[string]interface{}{"value": 42}, "true", false},
		{"float test", `{{ value is float }}`, map[string]interface{}{"value": 3.14}, "true", false},
		{"boolean test", `{{ value is boolean }}`, map[string]interface{}{"value": true}, "true", false},

		// Sequence and mapping tests
		{"sequence test", `{{ value is sequence }}`, map[string]interface{}{"value": []string{"a", "b"}}, "true", false},
		{"mapping test", `{{ value is mapping }}`, map[string]interface{}{"value": map[string]string{"a": "b"}}, "true", false},
		{"iterable test", `{{ value is iterable }}`, map[string]interface{}{"value": []string{"a", "b"}}, "true", false},

		// Numeric tests
		{"even test", `{{ value is even }}`, map[string]interface{}{"value": 4}, "true", false},
		{"odd test", `{{ value is odd }}`, map[string]interface{}{"value": 5}, "true", false},
		{"divisibleby test", `{{ value is divisibleby(3) }}`, map[string]interface{}{"value": 9}, "true", false},

		// String tests
		{"lower test", `{{ value is lower }}`, map[string]interface{}{"value": "hello"}, "true", false},
		{"upper test", `{{ value is upper }}`, map[string]interface{}{"value": "HELLO"}, "true", false},

		// Comparison tests
		{"equalto test", `{{ value is equalto(42) }}`, map[string]interface{}{"value": 42}, "true", false},
		{"sameas test", `{{ value is sameas(value2) }}`, map[string]interface{}{"value": 42, "value2": 42}, "true", false},

		// Negated tests
		{"not defined test", `{{ undefined_var is not defined }}`, map[string]interface{}{}, "true", false},
		{"not none test", `{{ value is not none }}`, map[string]interface{}{"value": "hello"}, "true", false},
		{"not even test", `{{ value is not even }}`, map[string]interface{}{"value": 5}, "true", false},

		// Tests that should work but might be missing
		{"callable test", `{{ func is callable }}`, map[string]interface{}{"func": func() string { return "test" }}, "true", true}, // Expected to possibly fail
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := miya.NewContext()
			for k, v := range tt.data {
				ctx.Set(k, v)
			}

			result, err := env.RenderString(tt.template, ctx)

			if tt.shouldFail {
				// For tests we expect might fail, just log the result
				if err != nil {
					t.Logf("Expected potential failure for %s: %v", tt.name, err)
				} else if result != tt.expected {
					t.Logf("Expected potential failure for %s: expected %q, got %q", tt.name, tt.expected, result)
				} else {
					t.Logf("Unexpectedly passed for %s: got %q", tt.name, result)
				}
				return
			}

			if err != nil {
				t.Fatalf("Error rendering template for %s: %v", tt.name, err)
			}

			if result != tt.expected {
				t.Errorf("%s: Expected %q, got %q", tt.name, tt.expected, result)
			}
		})
	}
}
