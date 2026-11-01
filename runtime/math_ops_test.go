package runtime

import (
	"testing"
)

// TestEvaluatorMathOperations tests all mathematical operations in the evaluator
func TestEvaluatorMathOperations(t *testing.T) {
	e := NewEvaluator()

	t.Run("add", func(t *testing.T) {
		result, err := e.addWithNode(10, 5, nil)
		if err != nil {
			t.Fatalf("add failed: %v", err)
		}
		if result != float64(15) {
			t.Errorf("add(10, 5) = %v, want 15", result)
		}
	})

	t.Run("divide", func(t *testing.T) {
		result, err := e.divideWithNode(10, 2, nil)
		if err != nil {
			t.Fatalf("divide failed: %v", err)
		}
		if result != float64(5) {
			t.Errorf("divide(10, 2) = %v, want 5", result)
		}
	})

	t.Run("divide_by_zero", func(t *testing.T) {
		_, err := e.divideWithNode(10, 0, nil)
		if err == nil {
			t.Error("divide by zero should return error")
		}
	})

	t.Run("floorDivide", func(t *testing.T) {
		result, err := e.floorDivideWithNode(10, 3, nil)
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
		_, err := e.floorDivideWithNode(10, 0, nil)
		if err == nil {
			t.Error("floor divide by zero should return error")
		}
	})

	t.Run("floorDivide_type_error", func(t *testing.T) {
		_, err := e.floorDivideWithNode("hello", 2, nil)
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
			result, err := e.applyBinaryOpWithNode(tt.op, tt.a, tt.b, nil)
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
		result, err := e.callFunctionWithContext(fn, []interface{}{1, 2}, nil, nil)
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
		result, err := e.callFunctionWithContext(fn, nil, nil, nil)
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
		result, err := e.callFunctionWithContext(fn, []interface{}{1, 2, 3}, nil, nil)
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
		result, err := e.callFunctionWithContext(fn, []interface{}{3}, kwargs, nil)
		if err != nil {
			t.Fatalf("callFunction failed: %v", err)
		}
		if result != 3 {
			t.Errorf("callFunction = %v, want 3", result)
		}
	})

	t.Run("non-function", func(t *testing.T) {
		_, err := e.callFunctionWithContext("not a function", nil, nil, nil)
		if err == nil {
			t.Error("expected error for non-function")
		}
	})

	t.Run("nil function", func(t *testing.T) {
		_, err := e.callFunctionWithContext(nil, nil, nil, nil)
		if err == nil {
			t.Error("expected error for nil function")
		}
	})

	t.Run("Joiner", func(t *testing.T) {
		joiner := &Joiner{Separator: ", ", Used: true}
		result, err := e.callFunctionWithContext(joiner, nil, nil, nil)
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
		if err := e.setAttribute(m, "new_key", "new_value"); err == nil {
			t.Fatal("expected error assigning an attribute on a map")
		}
		if _, ok := m["new_key"]; ok || len(m) != 1 {
			t.Errorf("map was modified: %v", m)
		}
	})

	t.Run("setItem on slice", func(t *testing.T) {
		s := []interface{}{"a", "b", "c"}
		if err := e.setItem(s, 1, "X"); err == nil {
			t.Fatal("expected error assigning an item on a slice")
		}
		if s[1] != "b" {
			t.Errorf("slice was modified: %v", s)
		}
	})

	t.Run("setItem on map", func(t *testing.T) {
		m := map[string]interface{}{}
		if err := e.setItem(m, "key", "value"); err == nil {
			t.Fatal("expected error assigning an item on a map")
		}
		if len(m) != 0 {
			t.Errorf("map was modified: %v", m)
		}
	})
}

func TestApplyBinaryOpLogicalOperands(t *testing.T) {
	e := NewEvaluator()
	tests := []struct {
		op          string
		left, right interface{}
		want        interface{}
	}{
		{"or", "", "x", "x"},
		{"or", "a", "x", "a"},
		{"and", 0, "x", 0},
		{"and", 1, "x", "x"},
	}
	for _, tt := range tests {
		got, err := e.applyBinaryOpWithNode(tt.op, tt.left, tt.right, nil)
		if err != nil {
			t.Fatalf("%v %s %v: %v", tt.left, tt.op, tt.right, err)
		}
		if got != tt.want {
			t.Errorf("%#v %s %#v = %#v, want %#v", tt.left, tt.op, tt.right, got, tt.want)
		}
	}
}
