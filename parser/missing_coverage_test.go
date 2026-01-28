package parser

import (
	"strings"
	"testing"
)

// Test missing AST node String methods
func TestMissingASTNodeStringMethods(t *testing.T) {
	// Test FilterNode String method
	t.Run("FilterNode String", func(t *testing.T) {
		expr := NewIdentifierNode("name", 1, 1)
		filterNode := NewFilterNode(expr, "upper", []ExpressionNode{}, 1, 1)
		result := filterNode.String()

		if !strings.Contains(result, "Filter") && !strings.Contains(result, "upper") {
			t.Errorf("Expected filter info in result, got: %s", result)
		}

		// Test ExpressionNode interface
		filterNode.ExpressionNode()

		// Test with arguments
		args := []ExpressionNode{NewLiteralNode("test", "test", 1, 1)}
		filterWithArgs := NewFilterNode(expr, "truncate", args, 1, 1)
		result = filterWithArgs.String()

		if !strings.Contains(result, "truncate") {
			t.Errorf("Expected 'truncate' in filter string, got: %s", result)
		}
	})

	// Test BinaryOpNode String method
	t.Run("BinaryOpNode String", func(t *testing.T) {
		left := NewLiteralNode(5, "5", 1, 1)
		right := NewLiteralNode(3, "3", 1, 1)
		binaryOpNode := NewBinaryOpNode(left, "+", right, 1, 1)
		result := binaryOpNode.String()

		if !strings.Contains(result, "+") {
			t.Errorf("Expected '+' operator in result, got: %s", result)
		}

		// Test ExpressionNode interface
		binaryOpNode.ExpressionNode()
	})

	// Test UnaryOpNode String method
	t.Run("UnaryOpNode String", func(t *testing.T) {
		operand := NewIdentifierNode("value", 1, 1)
		unaryOpNode := NewUnaryOpNode("-", operand, 1, 1)
		result := unaryOpNode.String()

		if !strings.Contains(result, "-") {
			t.Errorf("Expected '-' operator in result, got: %s", result)
		}

		// Test ExpressionNode interface
		unaryOpNode.ExpressionNode()
	})

	// Test IfNode String method
	t.Run("IfNode String", func(t *testing.T) {
		condition := NewIdentifierNode("user", 1, 1)

		ifNode := NewIfNode(condition, 1, 1)
		ifNode.Body = []Node{NewTextNode("Hello", 2, 1)}
		result := ifNode.String()

		if !strings.Contains(result, "If") {
			t.Errorf("Expected 'If' in result, got: %s", result)
		}

		// Test StatementNode interface
		ifNode.StatementNode()
	})

	// Test ForNode String method
	t.Run("ForNode String", func(t *testing.T) {
		variables := []string{"item"}
		iterable := NewIdentifierNode("items", 1, 1)

		forNode := NewForNode(variables, iterable, 1, 1)
		forNode.Body = []Node{NewTextNode("Item", 2, 1)}
		result := forNode.String()

		if !strings.Contains(result, "For") {
			t.Errorf("Expected 'For' in result, got: %s", result)
		}

		// Test StatementNode interface
		forNode.StatementNode()
	})

	// Test BlockNode String method
	t.Run("BlockNode String", func(t *testing.T) {
		blockNode := NewBlockNode("header", 1, 1)
		blockNode.Body = []Node{NewTextNode("Block content", 2, 1)}
		result := blockNode.String()

		if !strings.Contains(result, "Block") || !strings.Contains(result, "header") {
			t.Errorf("Expected 'Block' and 'header' in result, got: %s", result)
		}

		// Test StatementNode interface
		blockNode.StatementNode()
	})

	// Test ExtendsNode String method
	t.Run("ExtendsNode String", func(t *testing.T) {
		template := NewLiteralNode("base.html", "base.html", 1, 1)
		extendsNode := NewExtendsNode(template, 1, 1)
		result := extendsNode.String()

		if !strings.Contains(result, "Extends") {
			t.Errorf("Expected 'Extends' in result, got: %s", result)
		}

		// Test StatementNode interface
		extendsNode.StatementNode()
	})

	// Test IncludeNode String method
	t.Run("IncludeNode String", func(t *testing.T) {
		template := NewLiteralNode("partial.html", "partial.html", 1, 1)
		includeNode := NewIncludeNode(template, 1, 1)
		result := includeNode.String()

		if !strings.Contains(result, "Include") {
			t.Errorf("Expected 'Include' in result, got: %s", result)
		}

		// Test StatementNode interface
		includeNode.StatementNode()
	})

	// Test SuperNode String method
	t.Run("SuperNode String", func(t *testing.T) {
		superNode := NewSuperNode(1, 1)
		result := superNode.String()

		if !strings.Contains(result, "Super") {
			t.Errorf("Expected 'Super' in result, got: %s", result)
		}

		// Test ExpressionNode interface
		superNode.ExpressionNode()
	})

	// Test MacroNode String method
	t.Run("MacroNode String", func(t *testing.T) {
		macroNode := NewMacroNode("user_info", 1, 1)
		macroNode.Parameters = []string{"name", "age"}
		macroNode.Body = []Node{NewTextNode("Macro body", 2, 1)}
		result := macroNode.String()

		if !strings.Contains(result, "Macro") || !strings.Contains(result, "user_info") {
			t.Errorf("Expected 'Macro' and 'user_info' in result, got: %s", result)
		}

		// Test StatementNode interface
		macroNode.StatementNode()
	})
}

// Test more missing String methods
func TestMoreMissingStringMethods(t *testing.T) {
	// Test SetNode String method
	t.Run("SetNode String", func(t *testing.T) {
		targets := []ExpressionNode{
			NewIdentifierNode("x", 1, 1),
			NewIdentifierNode("y", 1, 1),
		}
		value := NewLiteralNode(42, "42", 1, 1)
		setNode := NewSetNodeWithTargets(targets, value, 1, 1)
		result := setNode.String()

		if !strings.Contains(result, "Set") {
			t.Errorf("Expected 'Set' in result, got: %s", result)
		}

		// Test StatementNode interface
		setNode.StatementNode()
	})

	// Test CallNode String method
	t.Run("CallNode String", func(t *testing.T) {
		expr := NewIdentifierNode("my_macro", 1, 1)
		callNode := NewCallNode(expr, 1, 1)
		callNode.Arguments = []ExpressionNode{NewLiteralNode("arg1", "arg1", 1, 1)}
		result := callNode.String()

		if !strings.Contains(result, "Call") {
			t.Errorf("Expected 'Call' in result, got: %s", result)
		}

		// Test ExpressionNode interface
		callNode.ExpressionNode()
	})

	// Test ConditionalNode String method
	t.Run("ConditionalNode String", func(t *testing.T) {
		condition := NewIdentifierNode("user", 1, 1)
		trueExpr := NewLiteralNode("yes", "yes", 1, 1)
		falseExpr := NewLiteralNode("no", "no", 1, 1)

		conditionalNode := NewConditionalNode(condition, trueExpr, falseExpr, 1, 1)
		result := conditionalNode.String()

		if !strings.Contains(result, "Conditional") {
			t.Errorf("Expected 'Conditional' in result, got: %s", result)
		}

		// Test ExpressionNode interface
		conditionalNode.ExpressionNode()
	})

	// Test SliceNode String method
	t.Run("SliceNode String", func(t *testing.T) {
		obj := NewIdentifierNode("items", 1, 1)

		sliceNode := NewSliceNode(obj, 1, 1)
		sliceNode.Start = NewLiteralNode(0, "0", 1, 1)
		sliceNode.End = NewLiteralNode(5, "5", 1, 1)
		sliceNode.Step = NewLiteralNode(2, "2", 1, 1)
		result := sliceNode.String()

		if !strings.Contains(result, "Slice") {
			t.Errorf("Expected 'Slice' in result, got: %s", result)
		}

		// Test ExpressionNode interface
		sliceNode.ExpressionNode()
	})

	// Test RawNode String method
	t.Run("RawNode String", func(t *testing.T) {
		rawNode := NewRawNode("raw content here", 1, 1)
		result := rawNode.String()

		if !strings.Contains(result, "Raw") {
			t.Errorf("Expected 'Raw' in result, got: %s", result)
		}

		// Test StatementNode interface - but RawNode doesn't implement it
		// rawNode.StatementNode() // This would cause compile error
	})

	// Test ExtensionNode String method
	t.Run("ExtensionNode String", func(t *testing.T) {
		extNode := NewExtensionNode("test", "test_tag", 1, 1)
		result := extNode.String()

		if !strings.Contains(result, "Extension") && !strings.Contains(result, "test") {
			t.Errorf("Expected extension info in result, got: %s", result)
		}

		// Test StatementNode interface
		extNode.StatementNode()
	})
}

// Test ExpressionNode interface methods
func TestExpressionNodeInterfaceMethods(t *testing.T) {
	t.Run("ExpressionNode methods", func(t *testing.T) {
		// Test that calling ExpressionNode() doesn't crash
		nodes := []ExpressionNode{
			NewIdentifierNode("test", 1, 1),
			NewLiteralNode("value", "value", 1, 1),
			NewListNode([]ExpressionNode{}, 1, 1),
			NewAttributeNode(NewIdentifierNode("obj", 1, 1), "attr", 1, 1),
			NewGetItemNode(NewIdentifierNode("obj", 1, 1), NewLiteralNode(0, "0", 1, 1), 1, 1),
		}

		for _, node := range nodes {
			// Should not panic
			node.ExpressionNode()
		}
	})
}
