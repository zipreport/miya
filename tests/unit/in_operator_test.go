package miya_test

import (
	miya "github.com/zipreport/miya"
	"testing"
)

func TestInOperatorParsing(t *testing.T) {
	env := miya.NewEnvironment()

	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected string
	}{
		{
			"basic in operator",
			`{{ "apple" in fruits }}`,
			map[string]interface{}{"fruits": []string{"apple", "banana", "cherry"}},
			"true",
		},
		{
			"basic not in operator",
			`{{ "grape" not in fruits }}`,
			map[string]interface{}{"fruits": []string{"apple", "banana", "cherry"}},
			"true",
		},
		{
			"in operator with variable",
			`{{ item in fruits }}`,
			map[string]interface{}{
				"item":   "banana",
				"fruits": []string{"apple", "banana", "cherry"},
			},
			"true",
		},
		{
			"not in operator with variable",
			`{{ item not in fruits }}`,
			map[string]interface{}{
				"item":   "grape",
				"fruits": []string{"apple", "banana", "cherry"},
			},
			"true",
		},
		{
			"in operator with string",
			`{{ "app" in word }}`,
			map[string]interface{}{"word": "apple"},
			"true",
		},
		{
			"in operator with map",
			`{{ "key1" in data }}`,
			map[string]interface{}{"data": map[string]string{"key1": "value1", "key2": "value2"}},
			"true",
		},
		{
			"not in operator false case",
			`{{ "apple" not in fruits }}`,
			map[string]interface{}{"fruits": []string{"apple", "banana", "cherry"}},
			"false",
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

func TestInOperatorPrecedence(t *testing.T) {
	env := miya.NewEnvironment()

	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected string
	}{
		{
			"in operator with and",
			`{{ "apple" in fruits and "banana" in fruits }}`,
			map[string]interface{}{"fruits": []string{"apple", "banana", "cherry"}},
			"true",
		},
		{
			"in operator with or",
			`{{ "grape" in fruits or "apple" in fruits }}`,
			map[string]interface{}{"fruits": []string{"apple", "banana", "cherry"}},
			"true",
		},
		{
			"in operator with comparison",
			`{{ "apple" in fruits == true }}`,
			map[string]interface{}{"fruits": []string{"apple", "banana", "cherry"}},
			"true",
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
