package branching

import (
	"testing"

	"github.com/zipreport/miya/filters"
)

func TestBuiltinTests(t *testing.T) {
	registry := NewTestRegistry()

	t.Run("defined test", func(t *testing.T) {
		result, err := registry.Apply("defined", "hello")
		if err != nil {
			t.Fatalf("defined test failed: %v", err)
		}
		if !result {
			t.Error("expected defined test to return true for non-nil value")
		}

		result, err = registry.Apply("defined", nil)
		if err != nil {
			t.Fatalf("defined test failed: %v", err)
		}
		if result {
			t.Error("expected defined test to return false for nil value")
		}
	})

	t.Run("none test", func(t *testing.T) {
		result, err := registry.Apply("none", nil)
		if err != nil {
			t.Fatalf("none test failed: %v", err)
		}
		if !result {
			t.Error("expected none test to return true for nil value")
		}

		result, err = registry.Apply("none", "hello")
		if err != nil {
			t.Fatalf("none test failed: %v", err)
		}
		if result {
			t.Error("expected none test to return false for non-nil value")
		}
	})

	t.Run("string test", func(t *testing.T) {
		result, err := registry.Apply("string", "hello")
		if err != nil {
			t.Fatalf("string test failed: %v", err)
		}
		if !result {
			t.Error("expected string test to return true for string value")
		}

		result, err = registry.Apply("string", 42)
		if err != nil {
			t.Fatalf("string test failed: %v", err)
		}
		if result {
			t.Error("expected string test to return false for non-string value")
		}
	})

	t.Run("number test", func(t *testing.T) {
		result, err := registry.Apply("number", 42)
		if err != nil {
			t.Fatalf("number test failed: %v", err)
		}
		if !result {
			t.Error("expected number test to return true for int value")
		}

		result, err = registry.Apply("number", 3.14)
		if err != nil {
			t.Fatalf("number test failed: %v", err)
		}
		if !result {
			t.Error("expected number test to return true for float value")
		}

		result, err = registry.Apply("number", "hello")
		if err != nil {
			t.Fatalf("number test failed: %v", err)
		}
		if result {
			t.Error("expected number test to return false for string value")
		}
	})

	t.Run("even test", func(t *testing.T) {
		result, err := registry.Apply("even", 4)
		if err != nil {
			t.Fatalf("even test failed: %v", err)
		}
		if !result {
			t.Error("expected even test to return true for even number")
		}

		result, err = registry.Apply("even", 5)
		if err != nil {
			t.Fatalf("even test failed: %v", err)
		}
		if result {
			t.Error("expected even test to return false for odd number")
		}
	})

	t.Run("odd test", func(t *testing.T) {
		result, err := registry.Apply("odd", 5)
		if err != nil {
			t.Fatalf("odd test failed: %v", err)
		}
		if !result {
			t.Error("expected odd test to return true for odd number")
		}

		result, err = registry.Apply("odd", 4)
		if err != nil {
			t.Fatalf("odd test failed: %v", err)
		}
		if result {
			t.Error("expected odd test to return false for even number")
		}
	})

	t.Run("divisibleby test", func(t *testing.T) {
		result, err := registry.Apply("divisibleby", 10, 2)
		if err != nil {
			t.Fatalf("divisibleby test failed: %v", err)
		}
		if !result {
			t.Error("expected divisibleby test to return true for 10 divisible by 2")
		}

		result, err = registry.Apply("divisibleby", 10, 3)
		if err != nil {
			t.Fatalf("divisibleby test failed: %v", err)
		}
		if result {
			t.Error("expected divisibleby test to return false for 10 not divisible by 3")
		}

		// Test error case
		_, err = registry.Apply("divisibleby", 10, 0)
		if err == nil {
			t.Error("expected divisibleby test to error on division by zero")
		}
	})

	t.Run("lower test", func(t *testing.T) {
		result, err := registry.Apply("lower", "hello")
		if err != nil {
			t.Fatalf("lower test failed: %v", err)
		}
		if !result {
			t.Error("expected lower test to return true for lowercase string")
		}

		result, err = registry.Apply("lower", "HELLO")
		if err != nil {
			t.Fatalf("lower test failed: %v", err)
		}
		if result {
			t.Error("expected lower test to return false for uppercase string")
		}
	})

	t.Run("upper test", func(t *testing.T) {
		result, err := registry.Apply("upper", "HELLO")
		if err != nil {
			t.Fatalf("upper test failed: %v", err)
		}
		if !result {
			t.Error("expected upper test to return true for uppercase string")
		}

		result, err = registry.Apply("upper", "hello")
		if err != nil {
			t.Fatalf("upper test failed: %v", err)
		}
		if result {
			t.Error("expected upper test to return false for lowercase string")
		}
	})

	t.Run("startswith test", func(t *testing.T) {
		result, err := registry.Apply("startswith", "hello world", "hello")
		if err != nil {
			t.Fatalf("startswith test failed: %v", err)
		}
		if !result {
			t.Error("expected startswith test to return true")
		}

		result, err = registry.Apply("startswith", "hello world", "world")
		if err != nil {
			t.Fatalf("startswith test failed: %v", err)
		}
		if result {
			t.Error("expected startswith test to return false")
		}
	})

	t.Run("endswith test", func(t *testing.T) {
		result, err := registry.Apply("endswith", "hello world", "world")
		if err != nil {
			t.Fatalf("endswith test failed: %v", err)
		}
		if !result {
			t.Error("expected endswith test to return true")
		}

		result, err = registry.Apply("endswith", "hello world", "hello")
		if err != nil {
			t.Fatalf("endswith test failed: %v", err)
		}
		if result {
			t.Error("expected endswith test to return false")
		}
	})

	t.Run("match test", func(t *testing.T) {
		result, err := registry.Apply("match", "hello123", `hello\d+`)
		if err != nil {
			t.Fatalf("match test failed: %v", err)
		}
		if !result {
			t.Error("expected match test to return true for matching pattern")
		}

		result, err = registry.Apply("match", "hello", `hello\d+`)
		if err != nil {
			t.Fatalf("match test failed: %v", err)
		}
		if result {
			t.Error("expected match test to return false for non-matching pattern")
		}
	})

	t.Run("sequence test", func(t *testing.T) {
		result, err := registry.Apply("sequence", []int{1, 2, 3})
		if err != nil {
			t.Fatalf("sequence test failed: %v", err)
		}
		if !result {
			t.Error("expected sequence test to return true for slice")
		}

		result, err = registry.Apply("sequence", "hello")
		if err != nil {
			t.Fatalf("sequence test failed: %v", err)
		}
		if !result {
			t.Error("expected sequence test to return true for string")
		}

		result, err = registry.Apply("sequence", 42)
		if err != nil {
			t.Fatalf("sequence test failed: %v", err)
		}
		if result {
			t.Error("expected sequence test to return false for number")
		}
	})

	t.Run("mapping test", func(t *testing.T) {
		result, err := registry.Apply("mapping", map[string]int{"a": 1})
		if err != nil {
			t.Fatalf("mapping test failed: %v", err)
		}
		if !result {
			t.Error("expected mapping test to return true for map")
		}

		result, err = registry.Apply("mapping", []int{1, 2, 3})
		if err != nil {
			t.Fatalf("mapping test failed: %v", err)
		}
		if result {
			t.Error("expected mapping test to return false for slice")
		}
	})

	t.Run("in test", func(t *testing.T) {
		result, err := registry.Apply("in", "hello", "hello world")
		if err != nil {
			t.Fatalf("in test failed: %v", err)
		}
		if !result {
			t.Error("expected in test to return true for substring")
		}

		result, err = registry.Apply("in", 2, []int{1, 2, 3})
		if err != nil {
			t.Fatalf("in test failed: %v", err)
		}
		if !result {
			t.Error("expected in test to return true for element in slice")
		}

		result, err = registry.Apply("in", "foo", map[string]int{"foo": 1})
		if err != nil {
			t.Fatalf("in test failed: %v", err)
		}
		if !result {
			t.Error("expected in test to return true for key in map")
		}
	})
}

func TestTestRegistry(t *testing.T) {
	registry := NewTestRegistry()

	t.Run("Register custom test", func(t *testing.T) {
		customTest := func(value interface{}, args ...interface{}) (bool, error) {
			return value == "custom", nil
		}

		err := registry.Register("custom", customTest)
		if err != nil {
			t.Fatalf("Failed to register custom test: %v", err)
		}

		result, err := registry.Apply("custom", "custom")
		if err != nil {
			t.Fatalf("Custom test failed: %v", err)
		}
		if !result {
			t.Error("Expected custom test to return true")
		}

		result, err = registry.Apply("custom", "other")
		if err != nil {
			t.Fatalf("Custom test failed: %v", err)
		}
		if result {
			t.Error("Expected custom test to return false")
		}
	})

	t.Run("Duplicate registration", func(t *testing.T) {
		customTest := func(value interface{}, args ...interface{}) (bool, error) {
			return true, nil
		}

		err := registry.Register("defined", customTest)
		if err == nil {
			t.Error("Expected error when registering duplicate test")
		}
	})

	t.Run("Unknown test", func(t *testing.T) {
		_, err := registry.Apply("nonexistent", "value")
		if err == nil {
			t.Error("Expected error when applying unknown test")
		}
	})

	t.Run("List tests", func(t *testing.T) {
		tests := registry.List()
		if len(tests) == 0 {
			t.Error("Expected at least some built-in tests")
		}

		// Check that common tests are present
		found := make(map[string]bool)
		for _, name := range tests {
			found[name] = true
		}

		expectedTests := []string{"defined", "none", "string", "number", "even", "odd"}
		for _, expected := range expectedTests {
			if !found[expected] {
				t.Errorf("Expected to find built-in test '%s'", expected)
			}
		}
	})
}

// Test comparison functions with 0% coverage
func TestComparisonFunctions(t *testing.T) {
	t.Run("testEqual", func(t *testing.T) {
		tests := []struct {
			name     string
			value    interface{}
			args     []interface{}
			expected bool
		}{
			{"equal integers", 5, []interface{}{5}, true},
			{"unequal integers", 5, []interface{}{3}, false},
			{"equal strings", "hello", []interface{}{"hello"}, true},
			{"unequal strings", "hello", []interface{}{"world"}, false},
			{"nil equals nil", nil, []interface{}{nil}, true},
			{"value not nil", 5, []interface{}{nil}, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := testEqual(tt.value, tt.args...)
				if err != nil {
					t.Errorf("testEqual(%v, %v) unexpected error: %v", tt.value, tt.args, err)
				}
				if result != tt.expected {
					t.Errorf("testEqual(%v, %v) = %v, expected %v", tt.value, tt.args, result, tt.expected)
				}
			})
		}
	})

	t.Run("testNotEqual", func(t *testing.T) {
		tests := []struct {
			name     string
			value    interface{}
			args     []interface{}
			expected bool
		}{
			{"equal integers", 5, []interface{}{5}, false},
			{"unequal integers", 5, []interface{}{3}, true},
			{"equal strings", "hello", []interface{}{"hello"}, false},
			{"unequal strings", "hello", []interface{}{"world"}, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := testNotEqual(tt.value, tt.args...)
				if err != nil {
					t.Errorf("testNotEqual(%v, %v) unexpected error: %v", tt.value, tt.args, err)
				}
				if result != tt.expected {
					t.Errorf("testNotEqual(%v, %v) = %v, expected %v", tt.value, tt.args, result, tt.expected)
				}
			})
		}
	})

	t.Run("testLessThan", func(t *testing.T) {
		tests := []struct {
			name     string
			value    interface{}
			args     []interface{}
			expected bool
		}{
			{"5 < 10", 5, []interface{}{10}, true},
			{"10 < 5", 10, []interface{}{5}, false},
			{"5 < 5", 5, []interface{}{5}, false},
			{"string comparison", "a", []interface{}{"b"}, true},
			{"string comparison false", "z", []interface{}{"a"}, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := testLessThan(tt.value, tt.args...)
				if err != nil {
					t.Errorf("test function unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("testLessThan(%v, %v) = %v, expected %v", tt.value, tt.args, result, tt.expected)
				}
			})
		}
	})

	t.Run("testLessThanOrEqual", func(t *testing.T) {
		tests := []struct {
			name     string
			value    interface{}
			args     []interface{}
			expected bool
		}{
			{"5 <= 10", 5, []interface{}{10}, true},
			{"10 <= 5", 10, []interface{}{5}, false},
			{"5 <= 5", 5, []interface{}{5}, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := testLessThanOrEqual(tt.value, tt.args...)
				if err != nil {
					t.Errorf("test function unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("testLessThanOrEqual(%v, %v) = %v, expected %v", tt.value, tt.args, result, tt.expected)
				}
			})
		}
	})

	t.Run("testGreaterThan", func(t *testing.T) {
		tests := []struct {
			name     string
			value    interface{}
			args     []interface{}
			expected bool
		}{
			{"10 > 5", 10, []interface{}{5}, true},
			{"5 > 10", 5, []interface{}{10}, false},
			{"5 > 5", 5, []interface{}{5}, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := testGreaterThan(tt.value, tt.args...)
				if err != nil {
					t.Errorf("test function unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("testGreaterThan(%v, %v) = %v, expected %v", tt.value, tt.args, result, tt.expected)
				}
			})
		}
	})

	t.Run("testGreaterThanOrEqual", func(t *testing.T) {
		tests := []struct {
			name     string
			value    interface{}
			args     []interface{}
			expected bool
		}{
			{"10 >= 5", 10, []interface{}{5}, true},
			{"5 >= 10", 5, []interface{}{10}, false},
			{"5 >= 5", 5, []interface{}{5}, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := testGreaterThanOrEqual(tt.value, tt.args...)
				if err != nil {
					t.Errorf("test function unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("testGreaterThanOrEqual(%v, %v) = %v, expected %v", tt.value, tt.args, result, tt.expected)
				}
			})
		}
	})
}

// Test string validation functions
func TestStringValidationFunctions(t *testing.T) {
	t.Run("testAlpha", func(t *testing.T) {
		tests := []struct {
			name      string
			value     interface{}
			expected  bool
			expectErr bool
		}{
			{"all letters", "hello", true, false},
			{"uppercase letters", "HELLO", true, false},
			{"mixed case", "HelloWorld", true, false},
			{"with numbers", "hello123", false, false},
			{"with spaces", "hello world", false, false},
			{"empty string", "", false, false},
			{"non-string", 123, false, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := testAlpha(tt.value)
				if tt.expectErr {
					if err == nil {
						t.Errorf("testAlpha(%v) expected error but got none", tt.value)
					}
				} else {
					if err != nil {
						t.Errorf("testAlpha(%v) unexpected error: %v", tt.value, err)
					}
					if result != tt.expected {
						t.Errorf("testAlpha(%v) = %v, expected %v", tt.value, result, tt.expected)
					}
				}
			})
		}
	})

	t.Run("testAlnum", func(t *testing.T) {
		tests := []struct {
			name      string
			value     interface{}
			expected  bool
			expectErr bool
		}{
			{"all letters", "hello", true, false},
			{"all numbers", "12345", true, false},
			{"mixed alphanumeric", "hello123", true, false},
			{"with special chars", "hello@123", false, false},
			{"with spaces", "hello 123", false, false},
			{"empty string", "", false, false},
			{"non-string", 123, false, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := testAlnum(tt.value)
				if tt.expectErr {
					if err == nil {
						t.Errorf("testAlnum(%v) expected error but got none", tt.value)
					}
				} else {
					if err != nil {
						t.Errorf("testAlnum(%v) unexpected error: %v", tt.value, err)
					}
					if result != tt.expected {
						t.Errorf("testAlnum(%v) = %v, expected %v", tt.value, result, tt.expected)
					}
				}
			})
		}
	})

	t.Run("testAscii", func(t *testing.T) {
		tests := []struct {
			name      string
			value     interface{}
			expected  bool
			expectErr bool
		}{
			{"ASCII letters", "hello", true, false},
			{"ASCII with numbers", "hello123", true, false},
			{"ASCII special chars", "hello!@#", true, false},
			{"empty string", "", true, false}, // empty is valid ASCII
			{"unicode chars", "hello世界", false, false},
			{"emoji", "hello😀", false, false},
			{"non-string", 123, false, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := testAscii(tt.value)
				if tt.expectErr {
					if err == nil {
						t.Errorf("testAscii(%v) expected error but got none", tt.value)
					}
				} else {
					if err != nil {
						t.Errorf("testAscii(%v) unexpected error: %v", tt.value, err)
					}
					if result != tt.expected {
						t.Errorf("testAscii(%v) = %v, expected %v", tt.value, result, tt.expected)
					}
				}
			})
		}
	})
}

// Test other utility functions
func TestUtilityFunctions(t *testing.T) {
	t.Run("testContains", func(t *testing.T) {
		tests := []struct {
			name     string
			value    interface{}
			args     []interface{}
			expected bool
		}{
			{"string contains", "hello world", []interface{}{"world"}, true},
			{"string not contains", "hello world", []interface{}{"foo"}, false},
			{"slice contains", []interface{}{1, 2, 3}, []interface{}{2}, true},
			{"slice not contains", []interface{}{1, 2, 3}, []interface{}{5}, false},
			{"map contains key", map[string]interface{}{"a": 1, "b": 2}, []interface{}{"a"}, true},
			{"map not contains key", map[string]interface{}{"a": 1, "b": 2}, []interface{}{"c"}, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := testContains(tt.value, tt.args...)
				if err != nil {
					t.Errorf("testContains(%v, %v) unexpected error: %v", tt.value, tt.args, err)
				}
				if result != tt.expected {
					t.Errorf("testContains(%v, %v) = %v, expected %v", tt.value, tt.args, result, tt.expected)
				}
			})
		}
	})

	t.Run("testSameAs", func(t *testing.T) {
		obj1 := &struct{ val int }{val: 1}
		obj2 := &struct{ val int }{val: 1}

		tests := []struct {
			name     string
			value    interface{}
			args     []interface{}
			expected bool
		}{
			{"same object", obj1, []interface{}{obj1}, true},
			{"different objects", obj1, []interface{}{obj2}, false},
			{"nil same as nil", nil, []interface{}{nil}, true},
			{"value types equal", 5, []interface{}{5}, true}, // value types are compared by value
			{"strings equal", "hello", []interface{}{"hello"}, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := testSameAs(tt.value, tt.args...)
				if err != nil {
					t.Errorf("testSameAs(%v, %v) unexpected error: %v", tt.value, tt.args, err)
				}
				if result != tt.expected {
					t.Errorf("testSameAs(%v, %v) = %v, expected %v", tt.value, tt.args, result, tt.expected)
				}
			})
		}
	})

	t.Run("testEscaped", func(t *testing.T) {
		tests := []struct {
			name     string
			value    interface{}
			expected bool
		}{
			{"SafeValue is escaped", filters.SafeValue{Value: "safe"}, true},
			{"string not escaped", "unsafe", false},
			{"number not escaped", 123, false},
			{"nil not escaped", nil, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := testEscaped(tt.value)
				if err != nil {
					t.Errorf("test function unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("testEscaped(%v) = %v, expected %v", tt.value, result, tt.expected)
				}
			})
		}
	})
}

// Test type check functions with 0% coverage
func TestTypeCheckFunctions(t *testing.T) {
	t.Run("testUndefined", func(t *testing.T) {
		tests := []struct {
			name     string
			value    interface{}
			expected bool
		}{
			{"nil is undefined", nil, true},
			{"string is defined", "hello", false},
			{"int is defined", 42, false},
			{"empty string is defined", "", false},
			{"zero is defined", 0, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := testUndefined(tt.value)
				if err != nil {
					t.Errorf("testUndefined(%v) unexpected error: %v", tt.value, err)
				}
				if result != tt.expected {
					t.Errorf("testUndefined(%v) = %v, expected %v", tt.value, result, tt.expected)
				}
			})
		}
	})

	t.Run("testBoolean", func(t *testing.T) {
		tests := []struct {
			name     string
			value    interface{}
			expected bool
		}{
			{"true is boolean", true, true},
			{"false is boolean", false, true},
			{"string is not boolean", "true", false},
			{"int is not boolean", 1, false},
			{"nil is not boolean", nil, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := testBoolean(tt.value)
				if err != nil {
					t.Errorf("testBoolean(%v) unexpected error: %v", tt.value, err)
				}
				if result != tt.expected {
					t.Errorf("testBoolean(%v) = %v, expected %v", tt.value, result, tt.expected)
				}
			})
		}
	})

	t.Run("testInteger", func(t *testing.T) {
		tests := []struct {
			name     string
			value    interface{}
			expected bool
		}{
			{"int is integer", 42, true},
			{"int8 is integer", int8(42), true},
			{"int16 is integer", int16(42), true},
			{"int32 is integer", int32(42), true},
			{"int64 is integer", int64(42), true},
			{"uint is integer", uint(42), true},
			{"uint8 is integer", uint8(42), true},
			{"uint16 is integer", uint16(42), true},
			{"uint32 is integer", uint32(42), true},
			{"uint64 is integer", uint64(42), true},
			{"float32 is not integer", float32(42.5), false},
			{"float64 is not integer", float64(42.5), false},
			{"string is not integer", "42", false},
			{"nil is not integer", nil, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := testInteger(tt.value)
				if err != nil {
					t.Errorf("testInteger(%v) unexpected error: %v", tt.value, err)
				}
				if result != tt.expected {
					t.Errorf("testInteger(%v) = %v, expected %v", tt.value, result, tt.expected)
				}
			})
		}
	})

	t.Run("testFloat", func(t *testing.T) {
		tests := []struct {
			name     string
			value    interface{}
			expected bool
		}{
			{"float32 is float", float32(3.14), true},
			{"float64 is float", float64(3.14), true},
			{"int is not float", 42, false},
			{"string is not float", "3.14", false},
			{"nil is not float", nil, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := testFloat(tt.value)
				if err != nil {
					t.Errorf("testFloat(%v) unexpected error: %v", tt.value, err)
				}
				if result != tt.expected {
					t.Errorf("testFloat(%v) = %v, expected %v", tt.value, result, tt.expected)
				}
			})
		}
	})

	t.Run("testIterable", func(t *testing.T) {
		tests := []struct {
			name     string
			value    interface{}
			expected bool
		}{
			{"slice is iterable", []int{1, 2, 3}, true},
			{"array is iterable", [3]int{1, 2, 3}, true},
			{"map is iterable", map[string]int{"a": 1}, true},
			{"string is iterable", "hello", true},
			{"int is not iterable", 42, false},
			{"nil is not iterable", nil, false},
			{"bool is not iterable", true, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := testIterable(tt.value)
				if err != nil {
					t.Errorf("testIterable(%v) unexpected error: %v", tt.value, err)
				}
				if result != tt.expected {
					t.Errorf("testIterable(%v) = %v, expected %v", tt.value, result, tt.expected)
				}
			})
		}
	})

	t.Run("testCallable", func(t *testing.T) {
		testFunc := func() {}
		testFuncWithArgs := func(a, b int) int { return a + b }

		tests := []struct {
			name     string
			value    interface{}
			expected bool
		}{
			{"function is callable", testFunc, true},
			{"function with args is callable", testFuncWithArgs, true},
			{"string is not callable", "hello", false},
			{"int is not callable", 42, false},
			{"nil is not callable", nil, false},
			{"slice is not callable", []int{1, 2}, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := testCallable(tt.value)
				if err != nil {
					t.Errorf("testCallable(%v) unexpected error: %v", tt.value, err)
				}
				if result != tt.expected {
					t.Errorf("testCallable(%T) = %v, expected %v", tt.value, result, tt.expected)
				}
			})
		}
	})

	t.Run("testEmpty", func(t *testing.T) {
		tests := []struct {
			name     string
			value    interface{}
			expected bool
		}{
			// nil
			{"nil is empty", nil, true},
			// strings
			{"empty string is empty", "", true},
			{"non-empty string is not empty", "hello", false},
			// slices
			{"empty slice is empty", []interface{}{}, true},
			{"non-empty slice is not empty", []interface{}{1, 2}, false},
			{"empty string slice is empty", []string{}, true},
			{"non-empty string slice is not empty", []string{"a"}, false},
			// maps
			{"empty map is empty", map[string]interface{}{}, true},
			{"non-empty map is not empty", map[string]interface{}{"a": 1}, false},
			{"empty string map is empty", map[string]string{}, true},
			{"non-empty string map is not empty", map[string]string{"a": "b"}, false},
			// integers
			{"zero int is empty", 0, true},
			{"non-zero int is not empty", 42, false},
			{"zero int8 is empty", int8(0), true},
			{"zero int16 is empty", int16(0), true},
			{"zero int32 is empty", int32(0), true},
			{"zero int64 is empty", int64(0), true},
			// unsigned integers
			{"zero uint is empty", uint(0), true},
			{"non-zero uint is not empty", uint(42), false},
			{"zero uint8 is empty", uint8(0), true},
			{"zero uint16 is empty", uint16(0), true},
			{"zero uint32 is empty", uint32(0), true},
			{"zero uint64 is empty", uint64(0), true},
			// floats
			{"zero float32 is empty", float32(0.0), true},
			{"non-zero float32 is not empty", float32(3.14), false},
			{"zero float64 is empty", float64(0.0), true},
			{"non-zero float64 is not empty", float64(3.14), false},
			// booleans
			{"false is empty", false, true},
			{"true is not empty", true, false},
			// reflection fallback for other slice types
			{"empty int slice is empty", []int{}, true},
			{"non-empty int slice is not empty", []int{1, 2, 3}, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := testEmpty(tt.value)
				if err != nil {
					t.Errorf("testEmpty(%v) unexpected error: %v", tt.value, err)
				}
				if result != tt.expected {
					t.Errorf("testEmpty(%v) = %v, expected %v", tt.value, result, tt.expected)
				}
			})
		}
	})
}

// Test toInt conversion function
func TestToIntConversion(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected int
		hasError bool
	}{
		{"int", 42, 42, false},
		{"int8", int8(42), 42, false},
		{"int16", int16(42), 42, false},
		{"int32", int32(42), 42, false},
		{"int64", int64(42), 42, false},
		{"uint", uint(42), 42, false},
		{"uint8", uint8(42), 42, false},
		{"uint16", uint16(42), 42, false},
		{"uint32", uint32(42), 42, false},
		{"uint64", uint64(42), 42, false},
		{"float32", float32(42.9), 42, false},
		{"float64", float64(42.9), 42, false},
		{"string number", "42", 42, false},
		{"invalid string", "not a number", 0, true},
		{"bool", true, 0, true},
		{"slice", []int{1, 2}, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := toInt(tt.value)
			if tt.hasError {
				if err == nil {
					t.Errorf("toInt(%v) expected error but got none", tt.value)
				}
			} else {
				if err != nil {
					t.Errorf("toInt(%v) unexpected error: %v", tt.value, err)
				}
				if result != tt.expected {
					t.Errorf("toInt(%v) = %v, expected %v", tt.value, result, tt.expected)
				}
			}
		})
	}
}

// Test conversion functions
func TestConversionFunctions(t *testing.T) {
	t.Run("toFloat", func(t *testing.T) {
		tests := []struct {
			name     string
			value    interface{}
			expected float64
			hasError bool
		}{
			{"int to float", 42, 42.0, false},
			{"int64 to float", int64(42), 42.0, false},
			{"float32 to float", float32(42.5), 42.5, false},
			{"float64 unchanged", 42.5, 42.5, false},
			{"string to float", "42.5", 42.5, false},
			{"invalid string", "not a number", 0, true},
			{"bool to float", true, 0, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := toFloat(tt.value)
				if tt.hasError {
					if err == nil {
						t.Errorf("toFloat(%v) expected error but got none", tt.value)
					}
				} else {
					if err != nil {
						t.Errorf("toFloat(%v) unexpected error: %v", tt.value, err)
					}
					if result != tt.expected {
						t.Errorf("toFloat(%v) = %v, expected %v", tt.value, result, tt.expected)
					}
				}
			})
		}
	})

	t.Run("compareValues", func(t *testing.T) {
		tests := []struct {
			name      string
			a         interface{}
			b         interface{}
			compFunc  func(int) bool
			expected  bool
			expectErr bool
		}{
			{"equal ints", 5, 5, func(c int) bool { return c == 0 }, true, false},
			{"a < b ints", 3, 5, func(c int) bool { return c < 0 }, true, false},
			{"a > b ints", 7, 5, func(c int) bool { return c > 0 }, true, false},
			{"equal floats", 5.5, 5.5, func(c int) bool { return c == 0 }, true, false},
			{"a < b floats", 3.5, 5.5, func(c int) bool { return c < 0 }, true, false},
			{"a > b floats", 7.5, 5.5, func(c int) bool { return c > 0 }, true, false},
			{"mixed int float equal", 5, 5.0, func(c int) bool { return c == 0 }, true, false},
			{"mixed int float less", 3, 5.0, func(c int) bool { return c < 0 }, true, false},
			{"equal strings", "hello", "hello", func(c int) bool { return c == 0 }, true, false},
			{"a < b strings", "apple", "banana", func(c int) bool { return c < 0 }, true, false},
			{"a > b strings", "zebra", "apple", func(c int) bool { return c > 0 }, true, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := compareValues(tt.a, tt.b, tt.compFunc)
				if tt.expectErr {
					if err == nil {
						t.Errorf("compareValues(%v, %v) expected error but got none", tt.a, tt.b)
					}
				} else {
					if err != nil {
						t.Errorf("compareValues(%v, %v) unexpected error: %v", tt.a, tt.b, err)
					}
					if result != tt.expected {
						t.Errorf("compareValues(%v, %v) = %v, expected %v", tt.a, tt.b, result, tt.expected)
					}
				}
			})
		}
	})
}
