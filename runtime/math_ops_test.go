package runtime

import (
	"testing"
)

// TestEvaluatorMathOperations tests all mathematical operations in the evaluator
func TestEvaluatorMathOperations(t *testing.T) {
	e := NewEvaluator()

	t.Run("add", func(t *testing.T) {
		result, err := e.add(10, 5)
		if err != nil {
			t.Fatalf("add failed: %v", err)
		}
		if result != float64(15) {
			t.Errorf("add(10, 5) = %v, want 15", result)
		}
	})

	t.Run("addOld_strings", func(t *testing.T) {
		result, err := e.addOld("hello", " world")
		if err != nil {
			t.Fatalf("addOld failed: %v", err)
		}
		if result != "hello world" {
			t.Errorf("addOld('hello', ' world') = %v, want 'hello world'", result)
		}
	})

	t.Run("addOld_string_and_number", func(t *testing.T) {
		result, err := e.addOld("count: ", 42)
		if err != nil {
			t.Fatalf("addOld failed: %v", err)
		}
		if result != "count: 42" {
			t.Errorf("addOld('count: ', 42) = %v, want 'count: 42'", result)
		}
	})

	t.Run("addOld_slices", func(t *testing.T) {
		a := []interface{}{1, 2}
		b := []interface{}{3, 4}
		result, err := e.addOld(a, b)
		if err != nil {
			t.Fatalf("addOld failed: %v", err)
		}
		slice, ok := result.([]interface{})
		if !ok || len(slice) != 4 {
			t.Errorf("addOld slices = %v, want [1,2,3,4]", result)
		}
	})

	t.Run("addOld_slice_and_element", func(t *testing.T) {
		a := []interface{}{1, 2}
		result, err := e.addOld(a, 3)
		if err != nil {
			t.Fatalf("addOld failed: %v", err)
		}
		slice, ok := result.([]interface{})
		if !ok || len(slice) != 3 {
			t.Errorf("addOld(slice, element) = %v, want [1,2,3]", result)
		}
	})

	t.Run("addOld_element_and_slice", func(t *testing.T) {
		b := []interface{}{2, 3}
		result, err := e.addOld(1, b)
		if err != nil {
			t.Fatalf("addOld failed: %v", err)
		}
		slice, ok := result.([]interface{})
		if !ok || len(slice) != 3 {
			t.Errorf("addOld(element, slice) = %v, want [1,2,3]", result)
		}
	})

	t.Run("addOld_numbers", func(t *testing.T) {
		result, err := e.addOld(10, 5)
		if err != nil {
			t.Fatalf("addOld failed: %v", err)
		}
		if result != float64(15) {
			t.Errorf("addOld(10, 5) = %v, want 15", result)
		}
	})

	t.Run("divide", func(t *testing.T) {
		result, err := e.divide(10, 2)
		if err != nil {
			t.Fatalf("divide failed: %v", err)
		}
		if result != float64(5) {
			t.Errorf("divide(10, 2) = %v, want 5", result)
		}
	})

	t.Run("divide_by_zero", func(t *testing.T) {
		_, err := e.divide(10, 0)
		if err == nil {
			t.Error("divide by zero should return error")
		}
	})

	t.Run("floorDivide", func(t *testing.T) {
		result, err := e.floorDivide(10, 3)
		if err != nil {
			t.Fatalf("floorDivide failed: %v", err)
		}
		// The result can be int or int64 depending on implementation
		switch v := result.(type) {
		case int64:
			if v != 3 {
				t.Errorf("floorDivide(10, 3) = %v, want 3", result)
			}
		case int:
			if v != 3 {
				t.Errorf("floorDivide(10, 3) = %v, want 3", result)
			}
		default:
			t.Errorf("floorDivide returned unexpected type: %T", result)
		}
	})

	t.Run("floorDivide_by_zero", func(t *testing.T) {
		_, err := e.floorDivide(10, 0)
		if err == nil {
			t.Error("floor divide by zero should return error")
		}
	})

	t.Run("floorDivide_type_error", func(t *testing.T) {
		_, err := e.floorDivide("hello", 2)
		if err == nil {
			t.Error("floor divide with string should return error")
		}
	})
}

// TestEvaluatorFastToString tests the fastToString function
func TestEvaluatorFastToString(t *testing.T) {
	e := NewEvaluator()

	tests := []struct {
		input    interface{}
		expected string
	}{
		{"hello", "hello"},
		{123, "123"},
		{45.67, "45.67"},
		{true, "true"},
		{false, "false"},
		{nil, "<nil>"},
	}

	for _, tt := range tests {
		result := e.fastToString(tt.input)
		if result != tt.expected {
			t.Errorf("fastToString(%v) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

// TestEvaluatorHtmlEscape tests the htmlEscape function
func TestEvaluatorHtmlEscape(t *testing.T) {
	e := NewEvaluator()

	tests := []struct {
		input    string
		expected string
	}{
		{"<script>", "&lt;script&gt;"},
		{"a & b", "a &amp; b"},
		{"\"quoted\"", "&quot;quoted&quot;"},
		{"'single'", "&#x27;single&#x27;"},
		{"normal text", "normal text"},
	}

	for _, tt := range tests {
		result := e.htmlEscape(tt.input)
		if result != tt.expected {
			t.Errorf("htmlEscape(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

// TestFormatIntegerWithCommas tests integer formatting with commas
func TestFormatIntegerWithCommas(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{0, "0"},
		{100, "100"},
		{1000, "1,000"},
		{1000000, "1,000,000"},
		{-1000, "-1,000"},
	}

	for _, tt := range tests {
		result := formatIntegerWithCommas(tt.input)
		if result != tt.expected {
			t.Errorf("formatIntegerWithCommas(%d) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

// TestFormatIntegerPartWithCommas tests integer part formatting
func TestFormatIntegerPartWithCommas(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1000", "1,000"},
		{"100", "100"},
		{"1234567", "1,234,567"},
		{"12", "12"},
		{"1", "1"},
	}

	for _, tt := range tests {
		result := formatIntegerPartWithCommas(tt.input)
		if result != tt.expected {
			t.Errorf("formatIntegerPartWithCommas(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

// TestApplyBinaryOp tests binary operations
func TestApplyBinaryOp(t *testing.T) {
	e := NewEvaluator()

	tests := []struct {
		name     string
		op       string
		a        interface{}
		b        interface{}
		expected interface{}
		hasError bool
	}{
		{"add numbers", "+", 1, 2, float64(3), false},
		{"subtract", "-", 5, 3, float64(2), false},
		{"multiply", "*", 4, 3, float64(12), false},
		{"divide", "/", 10, 2, float64(5), false},
		{"modulo", "%", 10, 3, nil, false}, // Returns int, check separately
		{"power", "**", 2, 3, float64(8), false},
		{"floor divide", "//", 10, 3, nil, false}, // Returns int, check separately
		{"string concat", "~", "hello", "world", "helloworld", false},
		{"equal", "==", 1, 1, true, false},
		{"not equal", "!=", 1, 2, true, false},
		{"less than", "<", 1, 2, true, false},
		{"less than equal", "<=", 1, 1, true, false},
		{"greater than", ">", 2, 1, true, false},
		{"greater than equal", ">=", 2, 2, true, false},
		{"contains", "in", "a", []interface{}{"a", "b"}, true, false},
		{"not in", "not in", "c", []interface{}{"a", "b"}, true, false},
		{"and", "and", true, false, false, false},
		{"or", "or", true, false, true, false},
		{"unknown op", "??", 1, 2, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := e.applyBinaryOp(tt.op, tt.a, tt.b)
			if tt.hasError {
				if err == nil {
					t.Errorf("expected error for op %q", tt.op)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				// Skip comparison for nil expected (checked separately for type-flexible results)
				if tt.expected != nil && result != tt.expected {
					t.Errorf("%s: got %v (%T), want %v (%T)", tt.name, result, result, tt.expected, tt.expected)
				}
			}
		})
	}
}

// TestCallFunction tests calling Go functions from templates
func TestCallFunction(t *testing.T) {
	e := NewEvaluator()

	t.Run("variadic function", func(t *testing.T) {
		fn := func(args ...interface{}) (interface{}, error) {
			if len(args) >= 2 {
				a, _ := args[0].(int)
				b, _ := args[1].(int)
				return a + b, nil
			}
			return 0, nil
		}
		result, err := e.callFunction(fn, []interface{}{1, 2}, nil)
		if err != nil {
			t.Fatalf("callFunction failed: %v", err)
		}
		if result != 3 {
			t.Errorf("callFunction = %v, want 3", result)
		}
	})

	t.Run("function with no args", func(t *testing.T) {
		fn := func() (interface{}, error) {
			return "hello", nil
		}
		result, err := e.callFunction(fn, nil, nil)
		if err != nil {
			t.Fatalf("callFunction failed: %v", err)
		}
		if result != "hello" {
			t.Errorf("callFunction = %v, want 'hello'", result)
		}
	})

	t.Run("slice function", func(t *testing.T) {
		fn := func(args []interface{}) (interface{}, error) {
			return len(args), nil
		}
		result, err := e.callFunction(fn, []interface{}{1, 2, 3}, nil)
		if err != nil {
			t.Fatalf("callFunction failed: %v", err)
		}
		if result != 3 {
			t.Errorf("callFunction = %v, want 3", result)
		}
	})

	t.Run("function with args and kwargs", func(t *testing.T) {
		fn := func(args []interface{}, kwargs map[string]interface{}) (interface{}, error) {
			return len(args) + len(kwargs), nil
		}
		kwargs := map[string]interface{}{"a": 1, "b": 2}
		result, err := e.callFunction(fn, []interface{}{3}, kwargs)
		if err != nil {
			t.Fatalf("callFunction failed: %v", err)
		}
		if result != 3 {
			t.Errorf("callFunction = %v, want 3", result)
		}
	})

	t.Run("non-function", func(t *testing.T) {
		_, err := e.callFunction("not a function", nil, nil)
		if err == nil {
			t.Error("expected error for non-function")
		}
	})

	t.Run("nil function", func(t *testing.T) {
		_, err := e.callFunction(nil, nil, nil)
		if err == nil {
			t.Error("expected error for nil function")
		}
	})

	t.Run("Joiner", func(t *testing.T) {
		joiner := &Joiner{Separator: ", ", Used: true}
		result, err := e.callFunction(joiner, nil, nil)
		if err != nil {
			t.Fatalf("callFunction with Joiner failed: %v", err)
		}
		if result != ", " {
			t.Errorf("callFunction with Joiner = %v, want ', '", result)
		}
	})
}

// TestSetAttributeAndItem tests setting attributes and items
func TestSetAttributeAndItem(t *testing.T) {
	e := NewEvaluator()

	t.Run("setAttribute on map", func(t *testing.T) {
		m := map[string]interface{}{"existing": "value"}
		err := e.setAttribute(m, "new_key", "new_value")
		if err != nil {
			t.Fatalf("setAttribute failed: %v", err)
		}
		if m["new_key"] != "new_value" {
			t.Error("setAttribute did not set value")
		}
	})

	t.Run("setItem on slice", func(t *testing.T) {
		s := []interface{}{"a", "b", "c"}
		err := e.setItem(s, 1, "X")
		if err != nil {
			t.Fatalf("setItem failed: %v", err)
		}
		if s[1] != "X" {
			t.Error("setItem did not set value")
		}
	})

	t.Run("setItem out of bounds", func(t *testing.T) {
		s := []interface{}{"a"}
		err := e.setItem(s, 10, "X")
		if err == nil {
			t.Error("expected error for out of bounds")
		}
	})

	t.Run("setItem on map", func(t *testing.T) {
		m := map[string]interface{}{}
		err := e.setItem(m, "key", "value")
		if err != nil {
			t.Fatalf("setItem on map failed: %v", err)
		}
		if m["key"] != "value" {
			t.Error("setItem did not set map value")
		}
	})
}
