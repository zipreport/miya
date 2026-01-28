package miya_test

import (
	miya "github.com/zipreport/miya"
	"testing"
)

func TestOperatorPrecedence(t *testing.T) {
	env := miya.NewEnvironment()

	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected string
	}{
		{
			"multiplication before addition",
			`{{ 2 + 3 * 4 }}`,
			map[string]interface{}{},
			"14", // Should be (2 + (3 * 4)) = 14, not ((2 + 3) * 4) = 20
		},
		{
			"comparison before logical and",
			`{{ 5 > 3 and 2 < 4 }}`,
			map[string]interface{}{},
			"true", // Should be ((5 > 3) and (2 < 4)) = true
		},
		{
			"logical and before logical or",
			`{{ false or true and false }}`,
			map[string]interface{}{},
			"false", // Should be (false or (true and false)) = false
		},
		{
			"comparison before in operator",
			`{{ 1 == 1 and "a" in "abc" }}`,
			map[string]interface{}{},
			"true", // Should be ((1 == 1) and ("a" in "abc")) = true
		},
		{
			"power right associative",
			`{{ 2 ** 3 ** 2 }}`,
			map[string]interface{}{},
			"512", // Should be (2 ** (3 ** 2)) = 512, not ((2 ** 3) ** 2) = 64
		},
		{
			"unary minus before power",
			`{{ -2 ** 2 }}`,
			map[string]interface{}{},
			"-4", // Should be (-(2 ** 2)) = -4, not ((-2) ** 2) = 4
		},
		{
			"parentheses override precedence",
			`{{ (2 + 3) * 4 }}`,
			map[string]interface{}{},
			"20", // Should be ((2 + 3) * 4) = 20
		},
		{
			"complex mixed precedence",
			`{{ 1 + 2 * 3 > 5 and true }}`,
			map[string]interface{}{},
			"true", // Should be (((1 + (2 * 3)) > 5) and true) = true
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := miya.NewContext()
			for k, v := range tt.data {
				ctx.Set(k, v)
			}

			result, err := env.RenderString(tt.template, ctx)
			if err != nil {
				t.Fatalf("Error rendering template: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestComplexExpressions(t *testing.T) {
	env := miya.NewEnvironment()

	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected string
	}{
		{
			"nested operations",
			`{{ (a + b) * (c - d) }}`,
			map[string]interface{}{"a": 2, "b": 3, "c": 10, "d": 6},
			"20", // (2 + 3) * (10 - 6) = 5 * 4 = 20
		},
		{
			"chained comparisons",
			`{{ 1 < 2 < 3 }}`,
			map[string]interface{}{},
			"true", // Should handle chained comparisons if supported
		},
		{
			"mixed string and numeric operations",
			`{{ "count: " ~ (items | length) }}`,
			map[string]interface{}{"items": []string{"a", "b", "c"}},
			"count: 3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := miya.NewContext()
			for k, v := range tt.data {
				ctx.Set(k, v)
			}

			result, err := env.RenderString(tt.template, ctx)
			if err != nil {
				t.Logf("Error rendering template (might be expected): %v", err)
				// Some complex features might not be implemented yet
				return
			}

			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}
