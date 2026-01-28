package miya_test

import (
	miya "github.com/zipreport/miya"
	"testing"
)

func TestStringTests(t *testing.T) {
	env := miya.NewEnvironment()

	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected string
	}{
		// Alpha tests
		{
			"alpha test - alphabetic only",
			`{{ text is alpha }}`,
			map[string]interface{}{"text": "HelloWorld"},
			"true",
		},
		{
			"alpha test - with numbers",
			`{{ text is alpha }}`,
			map[string]interface{}{"text": "Hello123"},
			"false",
		},
		{
			"alpha test - with spaces",
			`{{ text is alpha }}`,
			map[string]interface{}{"text": "Hello World"},
			"false",
		},
		{
			"alpha test - empty string",
			`{{ text is alpha }}`,
			map[string]interface{}{"text": ""},
			"false",
		},
		{
			"alpha test - mixed case",
			`{{ text is alpha }}`,
			map[string]interface{}{"text": "AbCdEf"},
			"true",
		},

		// Alnum tests
		{
			"alnum test - alphanumeric",
			`{{ text is alnum }}`,
			map[string]interface{}{"text": "Hello123"},
			"true",
		},
		{
			"alnum test - only letters",
			`{{ text is alnum }}`,
			map[string]interface{}{"text": "HelloWorld"},
			"true",
		},
		{
			"alnum test - only numbers",
			`{{ text is alnum }}`,
			map[string]interface{}{"text": "123456"},
			"true",
		},
		{
			"alnum test - with special chars",
			`{{ text is alnum }}`,
			map[string]interface{}{"text": "Hello-World"},
			"false",
		},
		{
			"alnum test - with spaces",
			`{{ text is alnum }}`,
			map[string]interface{}{"text": "Hello 123"},
			"false",
		},
		{
			"alnum test - empty string",
			`{{ text is alnum }}`,
			map[string]interface{}{"text": ""},
			"false",
		},

		// ASCII tests
		{
			"ascii test - basic ascii",
			`{{ text is ascii }}`,
			map[string]interface{}{"text": "Hello World 123!@#"},
			"true",
		},
		{
			"ascii test - with unicode",
			`{{ text is ascii }}`,
			map[string]interface{}{"text": "Hello 🌍"},
			"false",
		},
		{
			"ascii test - with accented chars",
			`{{ text is ascii }}`,
			map[string]interface{}{"text": "Café"},
			"false",
		},
		{
			"ascii test - empty string",
			`{{ text is ascii }}`,
			map[string]interface{}{"text": ""},
			"true",
		},
		{
			"ascii test - extended ascii",
			`{{ text is ascii }}`,
			map[string]interface{}{"text": "Hello\x80"},
			"false",
		},

		// Negated tests
		{
			"not alpha test",
			`{{ text is not alpha }}`,
			map[string]interface{}{"text": "Hello123"},
			"true",
		},
		{
			"not alnum test",
			`{{ text is not alnum }}`,
			map[string]interface{}{"text": "Hello-World"},
			"true",
		},
		{
			"not ascii test",
			`{{ text is not ascii }}`,
			map[string]interface{}{"text": "Café"},
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

func TestStringTestsErrorCases(t *testing.T) {
	env := miya.NewEnvironment()

	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
	}{
		{
			"alpha test with non-string",
			`{{ value is alpha }}`,
			map[string]interface{}{"value": 123},
		},
		{
			"alnum test with non-string",
			`{{ value is alnum }}`,
			map[string]interface{}{"value": []string{"hello"}},
		},
		{
			"ascii test with non-string",
			`{{ value is ascii }}`,
			map[string]interface{}{"value": true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := miya.NewContext()
			for k, v := range tt.data {
				ctx.Set(k, v)
			}

			_, err := env.RenderString(tt.template, ctx)
			if err == nil {
				t.Error("Expected error for non-string input, but got none")
			}
		})
	}
}
