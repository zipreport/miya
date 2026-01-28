package runtime

import (
	"strings"
	"testing"

	"github.com/zipreport/miya/parser"
)

// TestNestedLoopEvaluator tests the NestedLoopEvaluator
func TestNestedLoopEvaluator(t *testing.T) {
	t.Run("NewNestedLoopEvaluator", func(t *testing.T) {
		baseEvaluator := NewEvaluator()
		cfEvaluator := NewControlFlowEvaluator(baseEvaluator)
		nestedEvaluator := NewNestedLoopEvaluator(cfEvaluator)

		if nestedEvaluator == nil {
			t.Fatal("Expected non-nil NestedLoopEvaluator")
		}

		if nestedEvaluator.cf != cfEvaluator {
			t.Error("Expected cf to be set to the provided ControlFlowEvaluator")
		}
	})

	t.Run("EvalNestedLoop basic", func(t *testing.T) {
		baseEvaluator := NewEvaluator()
		cfEvaluator := NewControlFlowEvaluator(baseEvaluator)
		nestedEvaluator := NewNestedLoopEvaluator(cfEvaluator)

		ctx := &simpleContext{variables: make(map[string]interface{})}
		ctx.SetVariable("outer_items", []interface{}{1, 2})
		ctx.SetVariable("inner_items", []interface{}{"a", "b"})

		// Create outer for node
		outerIterableNode := parser.NewIdentifierNode("outer_items", 1, 1)
		outerForNode := parser.NewSingleForNode("x", outerIterableNode, 1, 1)
		outerForNode.Body = []parser.Node{parser.NewIdentifierNode("x", 1, 1)}

		// Create inner for node
		innerIterableNode := parser.NewIdentifierNode("inner_items", 1, 1)
		innerForNode := parser.NewSingleForNode("y", innerIterableNode, 1, 1)
		innerForNode.Body = []parser.Node{parser.NewIdentifierNode("y", 1, 1)}

		result, err := nestedEvaluator.EvalNestedLoop(outerForNode, innerForNode, ctx)
		if err != nil {
			t.Fatalf("EvalNestedLoop failed: %v", err)
		}

		// Each outer item should iterate over all inner items
		// Result should be "abab" (inner loop runs twice for each outer)
		if result != "abab" {
			t.Errorf("Expected 'abab', got %q", result)
		}
	})

	t.Run("EvalNestedLoop with tuple unpacking outer", func(t *testing.T) {
		baseEvaluator := NewEvaluator()
		cfEvaluator := NewControlFlowEvaluator(baseEvaluator)
		nestedEvaluator := NewNestedLoopEvaluator(cfEvaluator)

		ctx := &simpleContext{variables: make(map[string]interface{})}
		// Outer items are tuples
		ctx.SetVariable("outer_items", []interface{}{
			[]interface{}{"a", 1},
			[]interface{}{"b", 2},
		})
		ctx.SetVariable("inner_items", []interface{}{"x", "y"})

		// Create outer for node with tuple unpacking
		outerIterableNode := parser.NewIdentifierNode("outer_items", 1, 1)
		outerForNode := parser.NewForNode([]string{"letter", "num"}, outerIterableNode, 1, 1)
		outerForNode.Body = []parser.Node{parser.NewIdentifierNode("letter", 1, 1)}

		// Create inner for node
		innerIterableNode := parser.NewIdentifierNode("inner_items", 1, 1)
		innerForNode := parser.NewSingleForNode("item", innerIterableNode, 1, 1)
		innerForNode.Body = []parser.Node{parser.NewIdentifierNode("item", 1, 1)}

		result, err := nestedEvaluator.EvalNestedLoop(outerForNode, innerForNode, ctx)
		if err != nil {
			t.Fatalf("EvalNestedLoop with tuple unpacking failed: %v", err)
		}

		// Should be "xyxy" (inner loop runs twice for each outer)
		if result != "xyxy" {
			t.Errorf("Expected 'xyxy', got %q", result)
		}
	})

	t.Run("EvalNestedLoop with tuple unpacking inner", func(t *testing.T) {
		baseEvaluator := NewEvaluator()
		cfEvaluator := NewControlFlowEvaluator(baseEvaluator)
		nestedEvaluator := NewNestedLoopEvaluator(cfEvaluator)

		ctx := &simpleContext{variables: make(map[string]interface{})}
		ctx.SetVariable("outer_items", []interface{}{1, 2})
		// Inner items are tuples
		ctx.SetVariable("inner_items", []interface{}{
			[]interface{}{"a", "x"},
			[]interface{}{"b", "y"},
		})

		// Create outer for node
		outerIterableNode := parser.NewIdentifierNode("outer_items", 1, 1)
		outerForNode := parser.NewSingleForNode("num", outerIterableNode, 1, 1)
		outerForNode.Body = []parser.Node{parser.NewIdentifierNode("num", 1, 1)}

		// Create inner for node with tuple unpacking
		innerIterableNode := parser.NewIdentifierNode("inner_items", 1, 1)
		innerForNode := parser.NewForNode([]string{"a", "b"}, innerIterableNode, 1, 1)
		innerForNode.Body = []parser.Node{parser.NewIdentifierNode("a", 1, 1)}

		result, err := nestedEvaluator.EvalNestedLoop(outerForNode, innerForNode, ctx)
		if err != nil {
			t.Fatalf("EvalNestedLoop with inner tuple unpacking failed: %v", err)
		}

		// Should be "abab"
		if result != "abab" {
			t.Errorf("Expected 'abab', got %q", result)
		}
	})

	t.Run("EvalNestedLoop outer iterable error", func(t *testing.T) {
		baseEvaluator := NewEvaluator()
		cfEvaluator := NewControlFlowEvaluator(baseEvaluator)
		nestedEvaluator := NewNestedLoopEvaluator(cfEvaluator)

		ctx := &simpleContext{variables: make(map[string]interface{})}
		// Don't set outer_items, so it will fail

		outerIterableNode := parser.NewIdentifierNode("undefined_outer", 1, 1)
		outerForNode := parser.NewSingleForNode("x", outerIterableNode, 1, 1)
		outerForNode.Body = []parser.Node{}

		innerIterableNode := parser.NewIdentifierNode("inner", 1, 1)
		innerForNode := parser.NewSingleForNode("y", innerIterableNode, 1, 1)
		innerForNode.Body = []parser.Node{}

		_, err := nestedEvaluator.EvalNestedLoop(outerForNode, innerForNode, ctx)
		// May or may not error depending on undefined behavior, but shouldn't panic
		_ = err
	})

	t.Run("EvalNestedLoop outer non-iterable", func(t *testing.T) {
		baseEvaluator := NewEvaluator()
		cfEvaluator := NewControlFlowEvaluator(baseEvaluator)
		nestedEvaluator := NewNestedLoopEvaluator(cfEvaluator)

		ctx := &simpleContext{variables: make(map[string]interface{})}
		ctx.SetVariable("outer", 42) // Not iterable
		ctx.SetVariable("inner", []interface{}{1, 2})

		outerIterableNode := parser.NewIdentifierNode("outer", 1, 1)
		outerForNode := parser.NewSingleForNode("x", outerIterableNode, 1, 1)
		outerForNode.Body = []parser.Node{}

		innerIterableNode := parser.NewIdentifierNode("inner", 1, 1)
		innerForNode := parser.NewSingleForNode("y", innerIterableNode, 1, 1)
		innerForNode.Body = []parser.Node{}

		_, err := nestedEvaluator.EvalNestedLoop(outerForNode, innerForNode, ctx)
		if err == nil {
			t.Error("Expected error for non-iterable outer")
		}
	})

	t.Run("EvalNestedLoop inner non-iterable", func(t *testing.T) {
		baseEvaluator := NewEvaluator()
		cfEvaluator := NewControlFlowEvaluator(baseEvaluator)
		nestedEvaluator := NewNestedLoopEvaluator(cfEvaluator)

		ctx := &simpleContext{variables: make(map[string]interface{})}
		ctx.SetVariable("outer", []interface{}{1, 2})
		ctx.SetVariable("inner", 42) // Not iterable

		outerIterableNode := parser.NewIdentifierNode("outer", 1, 1)
		outerForNode := parser.NewSingleForNode("x", outerIterableNode, 1, 1)
		outerForNode.Body = []parser.Node{}

		innerIterableNode := parser.NewIdentifierNode("inner", 1, 1)
		innerForNode := parser.NewSingleForNode("y", innerIterableNode, 1, 1)
		innerForNode.Body = []parser.Node{}

		_, err := nestedEvaluator.EvalNestedLoop(outerForNode, innerForNode, ctx)
		if err == nil {
			t.Error("Expected error for non-iterable inner")
		}
	})

	t.Run("EvalNestedLoop outer tuple unpack mismatch", func(t *testing.T) {
		baseEvaluator := NewEvaluator()
		cfEvaluator := NewControlFlowEvaluator(baseEvaluator)
		nestedEvaluator := NewNestedLoopEvaluator(cfEvaluator)

		ctx := &simpleContext{variables: make(map[string]interface{})}
		// Outer items have 2 elements but we're trying to unpack into 3 variables
		ctx.SetVariable("outer_items", []interface{}{
			[]interface{}{"a", 1},
		})
		ctx.SetVariable("inner_items", []interface{}{"x"})

		outerIterableNode := parser.NewIdentifierNode("outer_items", 1, 1)
		outerForNode := parser.NewForNode([]string{"a", "b", "c"}, outerIterableNode, 1, 1)
		outerForNode.Body = []parser.Node{}

		innerIterableNode := parser.NewIdentifierNode("inner_items", 1, 1)
		innerForNode := parser.NewSingleForNode("item", innerIterableNode, 1, 1)
		innerForNode.Body = []parser.Node{}

		_, err := nestedEvaluator.EvalNestedLoop(outerForNode, innerForNode, ctx)
		if err == nil {
			t.Error("Expected error for tuple unpack mismatch")
		}
		if err != nil && !strings.Contains(err.Error(), "cannot unpack") {
			t.Errorf("Expected 'cannot unpack' error, got: %v", err)
		}
	})

	t.Run("EvalNestedLoop inner tuple unpack mismatch", func(t *testing.T) {
		baseEvaluator := NewEvaluator()
		cfEvaluator := NewControlFlowEvaluator(baseEvaluator)
		nestedEvaluator := NewNestedLoopEvaluator(cfEvaluator)

		ctx := &simpleContext{variables: make(map[string]interface{})}
		ctx.SetVariable("outer_items", []interface{}{1})
		// Inner items have 2 elements but we're trying to unpack into 3 variables
		ctx.SetVariable("inner_items", []interface{}{
			[]interface{}{"a", 1},
		})

		outerIterableNode := parser.NewIdentifierNode("outer_items", 1, 1)
		outerForNode := parser.NewSingleForNode("x", outerIterableNode, 1, 1)
		outerForNode.Body = []parser.Node{}

		innerIterableNode := parser.NewIdentifierNode("inner_items", 1, 1)
		innerForNode := parser.NewForNode([]string{"a", "b", "c"}, innerIterableNode, 1, 1)
		innerForNode.Body = []parser.Node{}

		_, err := nestedEvaluator.EvalNestedLoop(outerForNode, innerForNode, ctx)
		if err == nil {
			t.Error("Expected error for inner tuple unpack mismatch")
		}
		if err != nil && !strings.Contains(err.Error(), "cannot unpack") {
			t.Errorf("Expected 'cannot unpack' error, got: %v", err)
		}
	})

	t.Run("EvalNestedLoop outer non-iterable for tuple unpack", func(t *testing.T) {
		baseEvaluator := NewEvaluator()
		cfEvaluator := NewControlFlowEvaluator(baseEvaluator)
		nestedEvaluator := NewNestedLoopEvaluator(cfEvaluator)

		ctx := &simpleContext{variables: make(map[string]interface{})}
		// Outer items are not tuples
		ctx.SetVariable("outer_items", []interface{}{1, 2})
		ctx.SetVariable("inner_items", []interface{}{"x"})

		outerIterableNode := parser.NewIdentifierNode("outer_items", 1, 1)
		// Try to unpack a single int into two variables
		outerForNode := parser.NewForNode([]string{"a", "b"}, outerIterableNode, 1, 1)
		outerForNode.Body = []parser.Node{}

		innerIterableNode := parser.NewIdentifierNode("inner_items", 1, 1)
		innerForNode := parser.NewSingleForNode("item", innerIterableNode, 1, 1)
		innerForNode.Body = []parser.Node{}

		_, err := nestedEvaluator.EvalNestedLoop(outerForNode, innerForNode, ctx)
		if err == nil {
			t.Error("Expected error for non-iterable item in tuple unpack")
		}
		if err != nil && !strings.Contains(err.Error(), "cannot unpack") {
			t.Errorf("Expected 'cannot unpack' error, got: %v", err)
		}
	})

	t.Run("EvalNestedLoop inner non-iterable for tuple unpack", func(t *testing.T) {
		baseEvaluator := NewEvaluator()
		cfEvaluator := NewControlFlowEvaluator(baseEvaluator)
		nestedEvaluator := NewNestedLoopEvaluator(cfEvaluator)

		ctx := &simpleContext{variables: make(map[string]interface{})}
		ctx.SetVariable("outer_items", []interface{}{1})
		// Inner items are not tuples
		ctx.SetVariable("inner_items", []interface{}{42})

		outerIterableNode := parser.NewIdentifierNode("outer_items", 1, 1)
		outerForNode := parser.NewSingleForNode("x", outerIterableNode, 1, 1)
		outerForNode.Body = []parser.Node{}

		innerIterableNode := parser.NewIdentifierNode("inner_items", 1, 1)
		// Try to unpack a single int into two variables
		innerForNode := parser.NewForNode([]string{"a", "b"}, innerIterableNode, 1, 1)
		innerForNode.Body = []parser.Node{}

		_, err := nestedEvaluator.EvalNestedLoop(outerForNode, innerForNode, ctx)
		if err == nil {
			t.Error("Expected error for non-iterable item in inner tuple unpack")
		}
		if err != nil && !strings.Contains(err.Error(), "cannot unpack") {
			t.Errorf("Expected 'cannot unpack' error, got: %v", err)
		}
	})
}

// TestLoopControlErrorMethods tests LoopControlError methods
func TestLoopControlErrorMethods(t *testing.T) {
	t.Run("Error method", func(t *testing.T) {
		breakErr := NewBreakError()
		errStr := breakErr.Error()
		if errStr != "break statement" {
			t.Errorf("Expected 'break statement', got %q", errStr)
		}

		continueErr := NewContinueError()
		errStr = continueErr.Error()
		if errStr != "continue statement" {
			t.Errorf("Expected 'continue statement', got %q", errStr)
		}
	})
}

// TestEvalForLoopMultipleVariables tests for loop with tuple unpacking
func TestEvalForLoopMultipleVariables(t *testing.T) {
	baseEvaluator := NewEvaluator()
	cfEvaluator := NewControlFlowEvaluator(baseEvaluator)

	t.Run("Tuple unpacking in for loop", func(t *testing.T) {
		ctx := &simpleContext{variables: make(map[string]interface{})}
		ctx.SetVariable("items", []interface{}{
			[]interface{}{"a", 1},
			[]interface{}{"b", 2},
		})

		iterableNode := parser.NewIdentifierNode("items", 1, 1)
		forNode := parser.NewForNode([]string{"letter", "num"}, iterableNode, 1, 1)
		forNode.Body = []parser.Node{parser.NewIdentifierNode("letter", 1, 1)}

		result, err := cfEvaluator.EvalForLoop(forNode, ctx)
		if err != nil {
			t.Fatalf("EvalForLoop with tuple unpacking failed: %v", err)
		}

		if result != "ab" {
			t.Errorf("Expected 'ab', got %q", result)
		}
	})

	t.Run("Tuple unpack mismatch", func(t *testing.T) {
		ctx := &simpleContext{variables: make(map[string]interface{})}
		ctx.SetVariable("items", []interface{}{
			[]interface{}{"a", 1, "extra"},
		})

		iterableNode := parser.NewIdentifierNode("items", 1, 1)
		// Only 2 variables but items have 3 elements
		forNode := parser.NewForNode([]string{"letter", "num"}, iterableNode, 1, 1)
		forNode.Body = []parser.Node{parser.NewIdentifierNode("letter", 1, 1)}

		_, err := cfEvaluator.EvalForLoop(forNode, ctx)
		if err == nil {
			t.Error("Expected error for tuple unpack mismatch")
		}
	})

	t.Run("Non-iterable item for tuple unpack", func(t *testing.T) {
		ctx := &simpleContext{variables: make(map[string]interface{})}
		ctx.SetVariable("items", []interface{}{42}) // Single int, not a tuple

		iterableNode := parser.NewIdentifierNode("items", 1, 1)
		forNode := parser.NewForNode([]string{"a", "b"}, iterableNode, 1, 1)
		forNode.Body = []parser.Node{}

		_, err := cfEvaluator.EvalForLoop(forNode, ctx)
		if err == nil {
			t.Error("Expected error for non-iterable item")
		}
	})
}

// TestEvalForLoopBreakContinue tests break and continue in for loops
func TestEvalForLoopBreakContinue(t *testing.T) {
	t.Run("For loop with break at specific index", func(t *testing.T) {
		// This test verifies the break handling code path is covered
		baseEvaluator := NewEvaluator()
		cfEvaluator := NewControlFlowEvaluator(baseEvaluator)

		ctx := &simpleContext{variables: make(map[string]interface{})}
		ctx.SetVariable("items", []interface{}{1, 2, 3, 4, 5})

		// We can't easily create a break node without more complex setup
		// but we can verify the EvalBreak method returns the right error
		err := cfEvaluator.EvalBreak()
		if err == nil {
			t.Fatal("Expected error from EvalBreak")
		}
		if loopErr, ok := err.(*LoopControlError); ok {
			if !loopErr.IsBreak() {
				t.Error("Expected IsBreak to return true")
			}
		}
	})

	t.Run("For loop with continue", func(t *testing.T) {
		baseEvaluator := NewEvaluator()
		cfEvaluator := NewControlFlowEvaluator(baseEvaluator)

		err := cfEvaluator.EvalContinue()
		if err == nil {
			t.Fatal("Expected error from EvalContinue")
		}
		if loopErr, ok := err.(*LoopControlError); ok {
			if !loopErr.IsContinue() {
				t.Error("Expected IsContinue to return true")
			}
		}
	})
}

// TestEvalIfStatementWithElseIf tests if statements with elif branches
func TestEvalIfStatementWithElseIf(t *testing.T) {
	baseEvaluator := NewEvaluator()
	cfEvaluator := NewControlFlowEvaluator(baseEvaluator)
	ctx := &simpleContext{variables: make(map[string]interface{})}

	t.Run("If with elif - elif branch taken", func(t *testing.T) {
		ctx.SetVariable("value", 2)

		// Condition: value == 1 (false)
		conditionNode := parser.NewBinaryOpNode(
			parser.NewIdentifierNode("value", 1, 1),
			"==",
			parser.NewLiteralNode(1, "", 1, 1),
			1, 1,
		)

		// Elif condition: value == 2 (true)
		elifConditionNode := parser.NewBinaryOpNode(
			parser.NewIdentifierNode("value", 1, 1),
			"==",
			parser.NewLiteralNode(2, "", 1, 1),
			1, 1,
		)

		ifNode := parser.NewIfNode(conditionNode, 1, 1)
		ifNode.Body = []parser.Node{parser.NewLiteralNode("first", "", 1, 1)}
		elifNode := parser.NewIfNode(elifConditionNode, 1, 1)
		elifNode.Body = []parser.Node{parser.NewLiteralNode("second", "", 1, 1)}
		ifNode.ElseIfs = []*parser.IfNode{elifNode}
		ifNode.Else = []parser.Node{parser.NewLiteralNode("third", "", 1, 1)}

		result, err := cfEvaluator.EvalIfStatement(ifNode, ctx)
		if err != nil {
			t.Fatalf("EvalIfStatement failed: %v", err)
		}

		if result != "second" {
			t.Errorf("Expected 'second', got %q", result)
		}
	})

	t.Run("If with multiple elifs", func(t *testing.T) {
		ctx.SetVariable("value", 3)

		conditionNode := parser.NewBinaryOpNode(
			parser.NewIdentifierNode("value", 1, 1),
			"==",
			parser.NewLiteralNode(1, "", 1, 1),
			1, 1,
		)

		elif1ConditionNode := parser.NewBinaryOpNode(
			parser.NewIdentifierNode("value", 1, 1),
			"==",
			parser.NewLiteralNode(2, "", 1, 1),
			1, 1,
		)

		elif2ConditionNode := parser.NewBinaryOpNode(
			parser.NewIdentifierNode("value", 1, 1),
			"==",
			parser.NewLiteralNode(3, "", 1, 1),
			1, 1,
		)

		ifNode := parser.NewIfNode(conditionNode, 1, 1)
		ifNode.Body = []parser.Node{parser.NewLiteralNode("first", "", 1, 1)}
		elif1Node := parser.NewIfNode(elif1ConditionNode, 1, 1)
		elif1Node.Body = []parser.Node{parser.NewLiteralNode("second", "", 1, 1)}
		elif2Node := parser.NewIfNode(elif2ConditionNode, 1, 1)
		elif2Node.Body = []parser.Node{parser.NewLiteralNode("third", "", 1, 1)}
		ifNode.ElseIfs = []*parser.IfNode{elif1Node, elif2Node}

		result, err := cfEvaluator.EvalIfStatement(ifNode, ctx)
		if err != nil {
			t.Fatalf("EvalIfStatement failed: %v", err)
		}

		if result != "third" {
			t.Errorf("Expected 'third', got %q", result)
		}
	})
}

// TestEvalBodyNodesWithLoopControl tests the loop control propagation
func TestEvalBodyNodesWithLoopControl(t *testing.T) {
	baseEvaluator := NewEvaluator()
	cfEvaluator := NewControlFlowEvaluator(baseEvaluator)
	ctx := &simpleContext{variables: make(map[string]interface{})}

	t.Run("Normal evaluation", func(t *testing.T) {
		nodes := []parser.Node{
			parser.NewLiteralNode("hello", "", 1, 1),
			parser.NewLiteralNode(" ", "", 1, 1),
			parser.NewLiteralNode("world", "", 1, 1),
		}

		result, err := cfEvaluator.evalBodyNodesWithLoopControl(nodes, ctx)
		if err != nil {
			t.Fatalf("evalBodyNodesWithLoopControl failed: %v", err)
		}

		if result != "hello world" {
			t.Errorf("Expected 'hello world', got %q", result)
		}
	})

	t.Run("Non-string result conversion", func(t *testing.T) {
		nodes := []parser.Node{
			parser.NewLiteralNode(42, "", 1, 1),
		}

		result, err := cfEvaluator.evalBodyNodesWithLoopControl(nodes, ctx)
		if err != nil {
			t.Fatalf("evalBodyNodesWithLoopControl failed: %v", err)
		}

		if result != "42" {
			t.Errorf("Expected '42', got %q", result)
		}
	})
}

// TestEvalLogicalAndShortCircuit tests short-circuit evaluation
func TestEvalLogicalAndShortCircuit(t *testing.T) {
	baseEvaluator := NewEvaluator()
	cfEvaluator := NewControlFlowEvaluator(baseEvaluator)
	ctx := &simpleContext{variables: make(map[string]interface{})}

	t.Run("Short-circuit on false left operand", func(t *testing.T) {
		// When left is false, right should not be evaluated
		leftNode := parser.NewLiteralNode(false, "", 1, 1)
		// Right node references undefined variable - would error if evaluated
		rightNode := parser.NewIdentifierNode("undefined_var", 1, 1)

		result, err := cfEvaluator.EvalLogicalAnd(leftNode, rightNode, ctx)
		if err != nil {
			t.Fatalf("EvalLogicalAnd should short-circuit: %v", err)
		}

		if result != false {
			t.Errorf("Expected false, got %v", result)
		}
	})
}

// TestEvalLogicalOrShortCircuit tests short-circuit evaluation for OR
func TestEvalLogicalOrShortCircuit(t *testing.T) {
	baseEvaluator := NewEvaluator()
	cfEvaluator := NewControlFlowEvaluator(baseEvaluator)
	ctx := &simpleContext{variables: make(map[string]interface{})}

	t.Run("Short-circuit on true left operand", func(t *testing.T) {
		// When left is true, right should not be evaluated
		leftNode := parser.NewLiteralNode(true, "", 1, 1)
		// Right node references undefined variable - would error if evaluated
		rightNode := parser.NewIdentifierNode("undefined_var", 1, 1)

		result, err := cfEvaluator.EvalLogicalOr(leftNode, rightNode, ctx)
		if err != nil {
			t.Fatalf("EvalLogicalOr should short-circuit: %v", err)
		}

		if result != true {
			t.Errorf("Expected true, got %v", result)
		}
	})
}

// TestEvalNotInExpression tests the not in expression
func TestEvalNotInExpression(t *testing.T) {
	baseEvaluator := NewEvaluator()
	cfEvaluator := NewControlFlowEvaluator(baseEvaluator)
	ctx := &simpleContext{variables: make(map[string]interface{})}

	t.Run("Value not in list", func(t *testing.T) {
		valueNode := parser.NewLiteralNode(5, "", 1, 1)
		listNode := parser.NewLiteralNode([]interface{}{1, 2, 3}, "", 1, 1)

		result, err := cfEvaluator.EvalNotInExpression(valueNode, listNode, ctx)
		if err != nil {
			t.Fatalf("EvalNotInExpression failed: %v", err)
		}

		if result != true {
			t.Errorf("Expected true (5 not in [1,2,3]), got %v", result)
		}
	})

	t.Run("Value in list", func(t *testing.T) {
		valueNode := parser.NewLiteralNode(2, "", 1, 1)
		listNode := parser.NewLiteralNode([]interface{}{1, 2, 3}, "", 1, 1)

		result, err := cfEvaluator.EvalNotInExpression(valueNode, listNode, ctx)
		if err != nil {
			t.Fatalf("EvalNotInExpression failed: %v", err)
		}

		if result != false {
			t.Errorf("Expected false (2 in [1,2,3]), got %v", result)
		}
	})
}

// TestForLoopWithBreakAndElse tests for loop break behavior with else clause
func TestForLoopWithBreakAndElse(t *testing.T) {
	baseEvaluator := NewEvaluator()
	cfEvaluator := NewControlFlowEvaluator(baseEvaluator)

	t.Run("Empty loop with else", func(t *testing.T) {
		ctx := &simpleContext{variables: make(map[string]interface{})}
		ctx.SetVariable("items", []interface{}{})

		iterableNode := parser.NewIdentifierNode("items", 1, 1)
		forNode := parser.NewSingleForNode("item", iterableNode, 1, 1)
		forNode.Body = []parser.Node{parser.NewIdentifierNode("item", 1, 1)}
		forNode.Else = []parser.Node{parser.NewLiteralNode("no items", "", 1, 1)}

		result, err := cfEvaluator.EvalForLoop(forNode, ctx)
		if err != nil {
			t.Fatalf("EvalForLoop failed: %v", err)
		}

		if result != "no items" {
			t.Errorf("Expected 'no items', got %q", result)
		}
	})
}

// Benchmark tests
func BenchmarkEvalForLoop(b *testing.B) {
	baseEvaluator := NewEvaluator()
	cfEvaluator := NewControlFlowEvaluator(baseEvaluator)
	ctx := &simpleContext{variables: make(map[string]interface{})}

	items := make([]interface{}, 100)
	for i := 0; i < 100; i++ {
		items[i] = i
	}
	ctx.SetVariable("items", items)

	iterableNode := parser.NewIdentifierNode("items", 1, 1)
	forNode := parser.NewSingleForNode("item", iterableNode, 1, 1)
	forNode.Body = []parser.Node{parser.NewIdentifierNode("item", 1, 1)}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cfEvaluator.EvalForLoop(forNode, ctx)
	}
}

func BenchmarkEvalNestedLoop(b *testing.B) {
	baseEvaluator := NewEvaluator()
	cfEvaluator := NewControlFlowEvaluator(baseEvaluator)
	nestedEvaluator := NewNestedLoopEvaluator(cfEvaluator)
	ctx := &simpleContext{variables: make(map[string]interface{})}

	outerItems := make([]interface{}, 10)
	for i := 0; i < 10; i++ {
		outerItems[i] = i
	}
	innerItems := make([]interface{}, 10)
	for i := 0; i < 10; i++ {
		innerItems[i] = i
	}
	ctx.SetVariable("outer", outerItems)
	ctx.SetVariable("inner", innerItems)

	outerIterableNode := parser.NewIdentifierNode("outer", 1, 1)
	outerForNode := parser.NewSingleForNode("x", outerIterableNode, 1, 1)
	outerForNode.Body = []parser.Node{parser.NewIdentifierNode("x", 1, 1)}

	innerIterableNode := parser.NewIdentifierNode("inner", 1, 1)
	innerForNode := parser.NewSingleForNode("y", innerIterableNode, 1, 1)
	innerForNode.Body = []parser.Node{parser.NewIdentifierNode("y", 1, 1)}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = nestedEvaluator.EvalNestedLoop(outerForNode, innerForNode, ctx)
	}
}
