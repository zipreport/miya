package runtime

import (
	"reflect"
	"testing"

	"github.com/zipreport/miya/parser"
)

// MockContext implements the Context interface for testing
type MockContext struct {
	data map[string]interface{}
}

func NewMockContext() *MockContext {
	return &MockContext{
		data: make(map[string]interface{}),
	}
}

func (c *MockContext) GetVariable(key string) (interface{}, bool) {
	val, ok := c.data[key]
	return val, ok
}

func (c *MockContext) SetVariable(key string, value interface{}) {
	c.data[key] = value
}

func (c *MockContext) Clone() Context {
	newCtx := &MockContext{data: make(map[string]interface{})}
	for k, v := range c.data {
		newCtx.data[k] = v
	}
	return newCtx
}

// Legacy methods for backwards compatibility
func (c *MockContext) Get(key string) (interface{}, bool) {
	return c.GetVariable(key)
}

func (c *MockContext) Set(key string, value interface{}) {
	c.SetVariable(key, value)
}

func (c *MockContext) All() map[string]interface{} {
	return c.data
}

func TestEvaluatorTextNode(t *testing.T) {
	eval := NewEvaluator()
	ctx := NewMockContext()

	node := parser.NewTextNode("Hello World", 1, 1)

	result, err := eval.EvalTextNode(node, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != "Hello World" {
		t.Errorf("expected 'Hello World', got %q", result)
	}
}

func TestEvaluatorLiteralNode(t *testing.T) {
	eval := NewEvaluator()
	ctx := NewMockContext()

	tests := []struct {
		name     string
		value    interface{}
		expected interface{}
	}{
		{"string", "hello", "hello"},
		{"integer", 42, 42},
		{"float", 3.14, 3.14},
		{"boolean", true, true},
		{"nil", nil, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := parser.NewLiteralNode(tt.value, "", 1, 1)

			result, err := eval.EvalLiteralNode(node, ctx)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestEvaluatorIdentifierNode(t *testing.T) {
	eval := NewStrictEvaluator() // Use strict evaluator to get errors for undefined variables
	ctx := NewMockContext()

	// Set up context
	ctx.Set("name", "John")
	ctx.Set("age", 30)

	tests := []struct {
		name     string
		ident    string
		expected interface{}
		hasError bool
	}{
		{"existing variable", "name", "John", false},
		{"existing number", "age", 30, false},
		{"undefined variable", "missing", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := parser.NewIdentifierNode(tt.ident, 1, 1)

			result, err := eval.EvalIdentifierNode(node, ctx)

			if tt.hasError {
				if err == nil {
					t.Errorf("expected error, got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestEvaluatorBinaryOpNode(t *testing.T) {
	eval := NewEvaluator()
	ctx := NewMockContext()

	tests := []struct {
		name     string
		left     interface{}
		op       string
		right    interface{}
		expected interface{}
		hasError bool
	}{
		{"addition int", 5, "+", 3, 8.0, false},
		{"addition string", "hello", "+", " world", "hello world", false},
		{"subtraction", 10, "-", 4, 6.0, false},
		{"multiplication", 6, "*", 7, 42.0, false},
		{"division", 15, "/", 3, 5.0, false},
		{"division by zero", 10, "/", 0, nil, true},
		{"equality true", 5, "==", 5, true, false},
		{"equality false", 5, "==", 3, false, false},
		{"inequality", 5, "!=", 3, true, false},
		{"less than", 3, "<", 5, true, false},
		{"greater than", 7, ">", 5, true, false},
		{"string concatenation", "hello", "~", "world", "helloworld", false},
		{"logical and true", true, "and", true, true, false},
		{"logical and false", true, "and", false, false, false},
		{"logical or", false, "or", true, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			leftNode := parser.NewLiteralNode(tt.left, "", 1, 1)
			rightNode := parser.NewLiteralNode(tt.right, "", 1, 1)
			node := parser.NewBinaryOpNode(leftNode, tt.op, rightNode, 1, 1)

			result, err := eval.EvalBinaryOpNode(node, ctx)

			if tt.hasError {
				if err == nil {
					t.Errorf("expected error, got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("expected %v (%T), got %v (%T)", tt.expected, tt.expected, result, result)
			}
		})
	}
}

func TestEvaluatorUnaryOpNode(t *testing.T) {
	eval := NewEvaluator()
	ctx := NewMockContext()

	tests := []struct {
		name     string
		op       string
		operand  interface{}
		expected interface{}
		hasError bool
	}{
		{"not true", "not", true, false, false},
		{"not false", "not", false, true, false},
		{"not empty string", "not", "", true, false},
		{"not non-empty string", "not", "hello", false, false},
		{"negation positive", "-", 5, -5.0, false},
		{"negation negative", "-", -3, 3.0, false},
		{"plus", "+", 42, 42, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			operandNode := parser.NewLiteralNode(tt.operand, "", 1, 1)
			node := parser.NewUnaryOpNode(tt.op, operandNode, 1, 1)

			result, err := eval.EvalUnaryOpNode(node, ctx)

			if tt.hasError {
				if err == nil {
					t.Errorf("expected error, got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestEvaluatorFilterNode(t *testing.T) {
	eval := NewEvaluator()
	ctx := NewMockContext()

	tests := []struct {
		name       string
		value      interface{}
		filterName string
		args       []interface{}
		expected   interface{}
		hasError   bool
	}{
		{"upper filter", "hello", "upper", nil, "HELLO", false},
		{"lower filter", "WORLD", "lower", nil, "world", false},
		{"capitalize filter", "jinja", "capitalize", nil, "Jinja", false},
		{"trim filter", "  spaced  ", "trim", nil, "spaced", false},
		{"default filter with value", "exists", "default", []interface{}{"fallback"}, "exists", false},
		{"default filter without value", "", "default", []interface{}{"fallback"}, "fallback", false},
		{"length filter string", "hello", "length", nil, 5, false},
		{"length filter slice", []interface{}{1, 2, 3}, "length", nil, 3, false},
		{"unknown filter", "test", "unknown", nil, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valueNode := parser.NewLiteralNode(tt.value, "", 1, 1)
			var argNodes []parser.ExpressionNode

			for _, arg := range tt.args {
				argNodes = append(argNodes, parser.NewLiteralNode(arg, "", 1, 1))
			}

			node := parser.NewFilterNode(valueNode, tt.filterName, argNodes, 1, 1)

			result, err := eval.EvalFilterNode(node, ctx)

			if tt.hasError {
				if err == nil {
					t.Errorf("expected error, got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestEvaluatorForNode(t *testing.T) {
	eval := NewEvaluator()
	ctx := NewMockContext()

	// Set up iterable data
	items := []interface{}{"a", "b", "c"}
	ctx.Set("items", items)

	// Create nodes
	variableNode := parser.NewIdentifierNode("items", 1, 1)
	forNode := parser.NewSingleForNode("item", variableNode, 1, 1)

	// Add body - just output the item
	identNode := parser.NewIdentifierNode("item", 1, 1)
	varNode := parser.NewVariableNode(identNode, 1, 1)
	forNode.Body = append(forNode.Body, varNode)

	result, err := eval.EvalForNode(forNode, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "abc"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestEvaluatorIfNode(t *testing.T) {
	eval := NewEvaluator()
	ctx := NewMockContext()

	tests := []struct {
		name      string
		condition interface{}
		expected  string
	}{
		{"true condition", true, "yes"},
		{"false condition", false, ""},
		{"truthy string", "hello", "yes"},
		{"falsy string", "", ""},
		{"truthy number", 42, "yes"},
		{"falsy number", 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conditionNode := parser.NewLiteralNode(tt.condition, "", 1, 1)
			ifNode := parser.NewIfNode(conditionNode, 1, 1)

			// Add body - output "yes"
			textNode := parser.NewTextNode("yes", 1, 1)
			ifNode.Body = append(ifNode.Body, textNode)

			result, err := eval.EvalIfNode(ifNode, ctx)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestEvaluatorSetNode(t *testing.T) {
	eval := NewEvaluator()
	ctx := NewMockContext()

	valueNode := parser.NewLiteralNode("test value", "", 1, 1)
	setNode := parser.NewSetNode("myvar", valueNode, 1, 1)

	result, err := eval.EvalSetNode(setNode, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Set should return empty string
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}

	// Check that variable was set
	value, ok := ctx.Get("myvar")
	if !ok {
		t.Errorf("variable was not set")
	}

	if value != "test value" {
		t.Errorf("expected 'test value', got %q", value)
	}
}

func TestEvaluatorTruthy(t *testing.T) {
	eval := NewEvaluator()

	tests := []struct {
		name     string
		value    interface{}
		expected bool
	}{
		{"nil", nil, false},
		{"true", true, true},
		{"false", false, false},
		{"empty string", "", false},
		{"non-empty string", "hello", true},
		{"zero int", 0, false},
		{"non-zero int", 42, true},
		{"zero float", 0.0, false},
		{"non-zero float", 3.14, true},
		{"empty slice", []interface{}{}, false},
		{"non-empty slice", []interface{}{1, 2, 3}, true},
		{"empty map", map[string]interface{}{}, false},
		{"non-empty map", map[string]interface{}{"key": "value"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := eval.isTruthy(tt.value)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestEvaluatorTestNode(t *testing.T) {
	eval := NewEvaluator()
	ctx := NewMockContext()

	t.Run("defined test", func(t *testing.T) {
		// Test with defined variable
		ctx.SetVariable("test_var", "hello")

		testNode := &parser.TestNode{
			Expression: parser.NewIdentifierNode("test_var", 1, 1),
			TestName:   "defined",
			Arguments:  []parser.ExpressionNode{},
			Negated:    false,
		}

		result, err := eval.EvalTestNode(testNode, ctx)
		if err != nil {
			t.Fatalf("EvalTestNode failed: %v", err)
		}

		if result != true {
			t.Errorf("Expected defined test to return true, got %v", result)
		}

		// Test with undefined variable - should return false, not error
		testNode.Expression = parser.NewIdentifierNode("undefined_var", 1, 1)

		result, err = eval.EvalTestNode(testNode, ctx)
		if err != nil {
			t.Errorf("Unexpected error for undefined variable in defined test: %v", err)
		}
		if result != false {
			t.Errorf("Expected defined test to return false for undefined variable, got %v", result)
		}
	})

	t.Run("none test", func(t *testing.T) {
		ctx.SetVariable("null_var", nil)

		testNode := &parser.TestNode{
			Expression: parser.NewIdentifierNode("null_var", 1, 1),
			TestName:   "none",
			Arguments:  []parser.ExpressionNode{},
			Negated:    false,
		}

		result, err := eval.EvalTestNode(testNode, ctx)
		if err != nil {
			t.Fatalf("EvalTestNode failed: %v", err)
		}

		if result != true {
			t.Errorf("Expected none test to return true for nil value, got %v", result)
		}
	})

	t.Run("string test", func(t *testing.T) {
		ctx.SetVariable("string_var", "hello")

		testNode := &parser.TestNode{
			Expression: parser.NewIdentifierNode("string_var", 1, 1),
			TestName:   "string",
			Arguments:  []parser.ExpressionNode{},
			Negated:    false,
		}

		result, err := eval.EvalTestNode(testNode, ctx)
		if err != nil {
			t.Fatalf("EvalTestNode failed: %v", err)
		}

		if result != true {
			t.Errorf("Expected string test to return true for string value, got %v", result)
		}

		// Test with non-string
		ctx.SetVariable("int_var", 42)
		testNode.Expression = parser.NewIdentifierNode("int_var", 1, 1)

		result, err = eval.EvalTestNode(testNode, ctx)
		if err != nil {
			t.Fatalf("EvalTestNode failed: %v", err)
		}

		if result != false {
			t.Errorf("Expected string test to return false for non-string value, got %v", result)
		}
	})

	t.Run("even test", func(t *testing.T) {
		ctx.SetVariable("even_num", 4)

		testNode := &parser.TestNode{
			Expression: parser.NewIdentifierNode("even_num", 1, 1),
			TestName:   "even",
			Arguments:  []parser.ExpressionNode{},
			Negated:    false,
		}

		result, err := eval.EvalTestNode(testNode, ctx)
		if err != nil {
			t.Fatalf("EvalTestNode failed: %v", err)
		}

		if result != true {
			t.Errorf("Expected even test to return true for even number, got %v", result)
		}

		// Test with odd number
		ctx.SetVariable("odd_num", 5)
		testNode.Expression = parser.NewIdentifierNode("odd_num", 1, 1)

		result, err = eval.EvalTestNode(testNode, ctx)
		if err != nil {
			t.Fatalf("EvalTestNode failed: %v", err)
		}

		if result != false {
			t.Errorf("Expected even test to return false for odd number, got %v", result)
		}
	})

	t.Run("divisibleby test with arguments", func(t *testing.T) {
		ctx.SetVariable("number", 10)

		testNode := &parser.TestNode{
			Expression: parser.NewIdentifierNode("number", 1, 1),
			TestName:   "divisibleby",
			Arguments: []parser.ExpressionNode{
				parser.NewLiteralNode(2, "", 1, 1),
			},
			Negated: false,
		}

		result, err := eval.EvalTestNode(testNode, ctx)
		if err != nil {
			t.Fatalf("EvalTestNode failed: %v", err)
		}

		if result != true {
			t.Errorf("Expected divisibleby test to return true for 10 divisible by 2, got %v", result)
		}

		// Test with non-divisible number
		testNode.Arguments[0] = parser.NewLiteralNode(3, "", 1, 1)

		result, err = eval.EvalTestNode(testNode, ctx)
		if err != nil {
			t.Fatalf("EvalTestNode failed: %v", err)
		}

		if result != false {
			t.Errorf("Expected divisibleby test to return false for 10 not divisible by 3, got %v", result)
		}
	})

	t.Run("negated test", func(t *testing.T) {
		ctx.SetVariable("test_var", "hello")

		testNode := &parser.TestNode{
			Expression: parser.NewIdentifierNode("test_var", 1, 1),
			TestName:   "none",
			Arguments:  []parser.ExpressionNode{},
			Negated:    true, // "is not none"
		}

		result, err := eval.EvalTestNode(testNode, ctx)
		if err != nil {
			t.Fatalf("EvalTestNode failed: %v", err)
		}

		if result != true {
			t.Errorf("Expected negated none test to return true for non-nil value, got %v", result)
		}
	})

	t.Run("startswith test with arguments", func(t *testing.T) {
		ctx.SetVariable("text", "hello world")

		testNode := &parser.TestNode{
			Expression: parser.NewIdentifierNode("text", 1, 1),
			TestName:   "startswith",
			Arguments: []parser.ExpressionNode{
				parser.NewLiteralNode("hello", "", 1, 1),
			},
			Negated: false,
		}

		result, err := eval.EvalTestNode(testNode, ctx)
		if err != nil {
			t.Fatalf("EvalTestNode failed: %v", err)
		}

		if result != true {
			t.Errorf("Expected startswith test to return true, got %v", result)
		}

		// Test with non-matching prefix
		testNode.Arguments[0] = parser.NewLiteralNode("world", "", 1, 1)

		result, err = eval.EvalTestNode(testNode, ctx)
		if err != nil {
			t.Fatalf("EvalTestNode failed: %v", err)
		}

		if result != false {
			t.Errorf("Expected startswith test to return false, got %v", result)
		}
	})

	t.Run("unknown test", func(t *testing.T) {
		ctx.SetVariable("test_var", "hello")

		testNode := &parser.TestNode{
			Expression: parser.NewIdentifierNode("test_var", 1, 1),
			TestName:   "nonexistent",
			Arguments:  []parser.ExpressionNode{},
			Negated:    false,
		}

		_, err := eval.EvalTestNode(testNode, ctx)
		if err == nil {
			t.Error("Expected error for unknown test, but got none")
		}

		if err.Error() != "unknown test: nonexistent" {
			t.Errorf("Expected specific error message, got: %v", err)
		}
	})
}
