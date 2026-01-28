package parser

import (
	"strings"
	"testing"
)

// Simple tests to improve coverage of AST String methods and basic functionality
func TestBasicASTNodes(t *testing.T) {
	// Test baseNode methods
	t.Run("baseNode methods", func(t *testing.T) {
		node := &baseNode{line: 5, column: 10}

		if node.Line() != 5 {
			t.Errorf("Expected line 5, got %d", node.Line())
		}

		if node.Column() != 10 {
			t.Errorf("Expected column 10, got %d", node.Column())
		}
	})

	// Test TextNode String method
	t.Run("TextNode String", func(t *testing.T) {
		textNode := NewTextNode("Hello World", 1, 1)
		result := textNode.String()

		expected := "Text(\"Hello World\")"
		if result != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result)
		}
	})

	// Test IdentifierNode
	t.Run("IdentifierNode", func(t *testing.T) {
		idNode := NewIdentifierNode("test_var", 1, 1)
		result := idNode.String()

		if !strings.Contains(result, "test_var") {
			t.Errorf("Expected 'test_var' in result, got: %s", result)
		}

		// Test ExpressionNode interface method
		idNode.ExpressionNode()
	})

	// Test LiteralNode
	t.Run("LiteralNode", func(t *testing.T) {
		literalNode := NewLiteralNode("hello", "hello", 1, 1)
		result := literalNode.String()

		if !strings.Contains(result, "hello") {
			t.Errorf("Expected 'hello' in result, got: %s", result)
		}

		// Test ExpressionNode interface method
		literalNode.ExpressionNode()
	})

	// Test VariableNode
	t.Run("VariableNode", func(t *testing.T) {
		expr := NewIdentifierNode("name", 1, 1)
		varNode := NewVariableNode(expr, 1, 1)
		result := varNode.String()

		if !strings.Contains(result, "Variable") {
			t.Errorf("Expected 'Variable' in result, got: %s", result)
		}

		if !strings.Contains(result, "name") {
			t.Errorf("Expected 'name' in result, got: %s", result)
		}
	})

	// Test TemplateNode with children
	t.Run("TemplateNode with children", func(t *testing.T) {
		template := NewTemplateNode("test.html", 1, 1)
		template.Children = []Node{
			NewTextNode("Hello", 1, 1),
			NewTextNode("World", 1, 7),
		}

		result := template.String()

		if !strings.Contains(result, "Template(test.html)") {
			t.Errorf("Expected Template(test.html) in result, got: %s", result)
		}

		if !strings.Contains(result, "Text(\"Hello\")") {
			t.Errorf("Expected Text(\"Hello\") in result, got: %s", result)
		}

		if !strings.Contains(result, "Text(\"World\")") {
			t.Errorf("Expected Text(\"World\") in result, got: %s", result)
		}
	})
}

// Test more AST node types to increase coverage
func TestAdditionalASTNodes(t *testing.T) {
	// Test AttributeNode if it exists
	t.Run("AttributeNode", func(t *testing.T) {
		obj := NewIdentifierNode("user", 1, 1)
		attrNode := NewAttributeNode(obj, "name", 1, 1)
		result := attrNode.String()

		if !strings.Contains(result, "user") || !strings.Contains(result, "name") {
			t.Errorf("Expected 'user' and 'name' in result, got: %s", result)
		}

		// Test ExpressionNode interface
		attrNode.ExpressionNode()
	})

	// Test GetItemNode if it exists
	t.Run("GetItemNode", func(t *testing.T) {
		obj := NewIdentifierNode("items", 1, 1)
		key := NewLiteralNode(0, "0", 1, 1)
		getItemNode := NewGetItemNode(obj, key, 1, 1)
		result := getItemNode.String()

		if !strings.Contains(result, "items") {
			t.Errorf("Expected 'items' in result, got: %s", result)
		}

		// Test ExpressionNode interface
		getItemNode.ExpressionNode()
	})

	// Test ListNode if it exists
	t.Run("ListNode", func(t *testing.T) {
		elements := []ExpressionNode{
			NewLiteralNode(1, "1", 1, 1),
			NewLiteralNode(2, "2", 1, 1),
		}
		listNode := NewListNode(elements, 1, 1)
		result := listNode.String()

		if !strings.Contains(result, "List") {
			t.Errorf("Expected 'List' in result, got: %s", result)
		}

		// Test ExpressionNode interface
		listNode.ExpressionNode()
	})
}

// Test ExtensionNode Evaluate method (our addition)
func TestExtensionNodeEvaluate(t *testing.T) {
	extNode := NewExtensionNode("test", "test_tag", 1, 1)

	// Test without evaluate function
	_, err := extNode.Evaluate("test context")
	if err == nil {
		t.Error("Expected error when no evaluate function is set")
	}

	// Test with evaluate function
	extNode.SetEvaluateFunc(func(node *ExtensionNode, ctx interface{}) (interface{}, error) {
		return "test result", nil
	})

	result, err := extNode.Evaluate("test context")
	if err != nil {
		t.Errorf("Evaluate failed: %v", err)
	}

	if result != "test result" {
		t.Errorf("Expected 'test result', got %v", result)
	}
}

// Test various node constructors that may not be covered
func TestNodeConstructors(t *testing.T) {
	// Test creating various nodes to ensure constructors are called
	t.Run("Create various nodes", func(t *testing.T) {
		// These should not panic and should create valid nodes
		template := NewTemplateNode("test", 1, 1)
		if template == nil {
			t.Error("NewTemplateNode returned nil")
		}

		text := NewTextNode("test", 1, 1)
		if text == nil {
			t.Error("NewTextNode returned nil")
		}

		id := NewIdentifierNode("var", 1, 1)
		if id == nil {
			t.Error("NewIdentifierNode returned nil")
		}

		literal := NewLiteralNode("value", "value", 1, 1)
		if literal == nil {
			t.Error("NewLiteralNode returned nil")
		}

		variable := NewVariableNode(id, 1, 1)
		if variable == nil {
			t.Error("NewVariableNode returned nil")
		}

		// Test that they have correct line/column
		if template.Line() != 1 || template.Column() != 1 {
			t.Error("Template node line/column not set correctly")
		}

		if text.Line() != 1 || text.Column() != 1 {
			t.Error("Text node line/column not set correctly")
		}
	})
}
