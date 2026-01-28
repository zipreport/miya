package miya

import (
	"testing"

	"github.com/zipreport/miya/parser"
	"github.com/zipreport/miya/runtime"
	"github.com/zipreport/miya/whitespace"
)

func TestCommentNodes(t *testing.T) {
	evaluator := runtime.NewEvaluator()
	ctx := &TemplateContextAdapter{ctx: NewContext(), env: NewEnvironment()}

	t.Run("Comment node evaluation", func(t *testing.T) {
		commentNode := parser.NewCommentNode("This is a comment", 1, 1)

		result, err := evaluator.EvalCommentNode(commentNode, ctx)
		if err != nil {
			t.Fatalf("Comment evaluation failed: %v", err)
		}

		if result != "" {
			t.Errorf("Expected empty string from comment, got %q", result)
		}
	})

	t.Run("Comment node in template", func(t *testing.T) {
		// Template with text, comment, and more text
		templateNode := parser.NewTemplateNode("test", 1, 1)
		templateNode.Children = []parser.Node{
			parser.NewTextNode("Before comment", 1, 1),
			parser.NewCommentNode("This is a comment", 1, 15),
			parser.NewTextNode("After comment", 1, 40),
		}

		result, err := evaluator.EvalTemplateNode(templateNode, ctx)
		if err != nil {
			t.Fatalf("Template evaluation failed: %v", err)
		}

		expected := "Before commentAfter comment"
		if result != expected {
			t.Errorf("Expected %q, got %q", expected, result)
		}
	})
}

func TestRawNodes(t *testing.T) {
	evaluator := runtime.NewEvaluator()
	ctx := &TemplateContextAdapter{ctx: NewContext(), env: NewEnvironment()}

	t.Run("Raw node evaluation", func(t *testing.T) {
		rawContent := "{{ this should not be evaluated }}"
		rawNode := parser.NewRawNode(rawContent, 1, 1)

		result, err := evaluator.EvalRawNode(rawNode, ctx)
		if err != nil {
			t.Fatalf("Raw evaluation failed: %v", err)
		}

		if result != rawContent {
			t.Errorf("Expected %q, got %q", rawContent, result)
		}
	})

	t.Run("Raw node preserves variables and expressions", func(t *testing.T) {
		rawContent := "Variables: {{ name }}, {{ age }}\nExpressions: {{ 2 + 3 }}"
		rawNode := parser.NewRawNode(rawContent, 1, 1)

		result, err := evaluator.EvalRawNode(rawNode, ctx)
		if err != nil {
			t.Fatalf("Raw evaluation failed: %v", err)
		}

		if result != rawContent {
			t.Errorf("Expected raw content to be preserved exactly, got %q", result)
		}
	})

	t.Run("Raw node in template", func(t *testing.T) {
		ctx.SetVariable("name", "John")

		templateNode := parser.NewTemplateNode("test", 1, 1)
		templateNode.Children = []parser.Node{
			parser.NewTextNode("Hello ", 1, 1),
			parser.NewVariableNode(parser.NewIdentifierNode("name", 1, 7), 1, 7),
			parser.NewTextNode("! Raw: ", 1, 13),
			parser.NewRawNode("{{ name }}", 1, 20),
		}

		result, err := evaluator.EvalTemplateNode(templateNode, ctx)
		if err != nil {
			t.Fatalf("Template evaluation failed: %v", err)
		}

		expected := "Hello John! Raw: {{ name }}"
		if result != expected {
			t.Errorf("Expected %q, got %q", expected, result)
		}
	})
}

func TestWhitespaceProcessing(t *testing.T) {
	t.Run("Trim blocks enabled", func(t *testing.T) {
		processor := whitespace.NewWhitespaceProcessor(true, false, false)

		nodes := []parser.Node{
			parser.NewTextNode("Hello", 1, 1),
			&parser.IfNode{}, // Block statement
			parser.NewTextNode("\nWorld", 1, 10),
		}

		result := processor.ProcessNodes(nodes)

		// Should have trimmed the newline after the if block
		if len(result) != 3 {
			t.Fatalf("Expected 3 nodes, got %d", len(result))
		}

		if textNode, ok := result[2].(*parser.TextNode); ok {
			if textNode.Content != "World" {
				t.Errorf("Expected 'World', got %q", textNode.Content)
			}
		} else {
			t.Error("Expected third node to be TextNode")
		}
	})

	t.Run("Lstrip blocks enabled", func(t *testing.T) {
		processor := whitespace.NewWhitespaceProcessor(false, true, false)

		nodes := []parser.Node{
			parser.NewTextNode("Hello   ", 1, 1),
			&parser.IfNode{}, // Block statement
			parser.NewTextNode("World", 1, 10),
		}

		result := processor.ProcessNodes(nodes)

		// Should have stripped trailing whitespace before the if block
		if len(result) != 3 {
			t.Fatalf("Expected 3 nodes, got %d", len(result))
		}

		if textNode, ok := result[0].(*parser.TextNode); ok {
			if textNode.Content != "Hello" {
				t.Errorf("Expected 'Hello', got %q", textNode.Content)
			}
		} else {
			t.Error("Expected first node to be TextNode")
		}
	})

	t.Run("Keep trailing newline disabled", func(t *testing.T) {
		processor := whitespace.NewWhitespaceProcessor(false, false, false)

		nodes := []parser.Node{
			parser.NewTextNode("Hello World\n", 1, 1),
		}

		result := processor.ProcessNodes(nodes)

		// Should have removed trailing newline
		if len(result) != 1 {
			t.Fatalf("Expected 1 node, got %d", len(result))
		}

		if textNode, ok := result[0].(*parser.TextNode); ok {
			if textNode.Content != "Hello World" {
				t.Errorf("Expected 'Hello World', got %q", textNode.Content)
			}
		} else {
			t.Error("Expected node to be TextNode")
		}
	})

	t.Run("Keep trailing newline enabled", func(t *testing.T) {
		processor := whitespace.NewWhitespaceProcessor(false, false, true)

		nodes := []parser.Node{
			parser.NewTextNode("Hello World\n", 1, 1),
		}

		result := processor.ProcessNodes(nodes)

		// Should have kept trailing newline
		if len(result) != 1 {
			t.Fatalf("Expected 1 node, got %d", len(result))
		}

		if textNode, ok := result[0].(*parser.TextNode); ok {
			if textNode.Content != "Hello World\n" {
				t.Errorf("Expected 'Hello World\\n', got %q", textNode.Content)
			}
		} else {
			t.Error("Expected node to be TextNode")
		}
	})

	t.Run("Empty text nodes removed", func(t *testing.T) {
		processor := whitespace.NewWhitespaceProcessor(true, false, false)

		nodes := []parser.Node{
			parser.NewTextNode("Hello", 1, 1),
			&parser.IfNode{},                // Block statement
			parser.NewTextNode("\n", 1, 10), // This will become empty after trim
			parser.NewTextNode("World", 1, 15),
		}

		result := processor.ProcessNodes(nodes)

		// Empty text node should be removed
		if len(result) != 3 {
			t.Fatalf("Expected 3 nodes, got %d", len(result))
		}

		// Check that we have the right nodes
		expectedContents := []string{"Hello", "", "World"} // IfNode produces empty string
		for i, node := range result {
			switch n := node.(type) {
			case *parser.TextNode:
				if i == 0 && n.Content != expectedContents[0] {
					t.Errorf("Node %d: expected %q, got %q", i, expectedContents[0], n.Content)
				}
				if i == 2 && n.Content != expectedContents[2] {
					t.Errorf("Node %d: expected %q, got %q", i, expectedContents[2], n.Content)
				}
			case *parser.IfNode:
				// Expected in position 1
			}
		}
	})
}

func TestWhitespaceControl(t *testing.T) {
	t.Run("Parse whitespace control - left strip", func(t *testing.T) {
		content, control := whitespace.ParseWhitespaceControl("-if condition")

		if content != "if condition" {
			t.Errorf("Expected 'if condition', got %q", content)
		}

		if !control.LeftStrip {
			t.Error("Expected LeftStrip to be true")
		}

		if control.RightStrip {
			t.Error("Expected RightStrip to be false")
		}
	})

	t.Run("Parse whitespace control - right strip", func(t *testing.T) {
		content, control := whitespace.ParseWhitespaceControl("if condition -")

		if content != "if condition" {
			t.Errorf("Expected 'if condition', got %q", content)
		}

		if control.LeftStrip {
			t.Error("Expected LeftStrip to be false")
		}

		if !control.RightStrip {
			t.Error("Expected RightStrip to be true")
		}
	})

	t.Run("Parse whitespace control - both strips", func(t *testing.T) {
		content, control := whitespace.ParseWhitespaceControl("- if condition -")

		if content != "if condition" {
			t.Errorf("Expected 'if condition', got %q", content)
		}

		if !control.LeftStrip {
			t.Error("Expected LeftStrip to be true")
		}

		if !control.RightStrip {
			t.Error("Expected RightStrip to be true")
		}
	})

	t.Run("Parse whitespace control - no strips", func(t *testing.T) {
		content, control := whitespace.ParseWhitespaceControl("if condition")

		if content != "if condition" {
			t.Errorf("Expected 'if condition', got %q", content)
		}

		if control.LeftStrip {
			t.Error("Expected LeftStrip to be false")
		}

		if control.RightStrip {
			t.Error("Expected RightStrip to be false")
		}
	})

	t.Run("Apply whitespace control", func(t *testing.T) {
		nodes := []parser.Node{
			parser.NewTextNode("Hello   ", 1, 1),
			parser.NewTextNode("   World", 1, 10),
		}

		controls := []whitespace.WhitespaceControl{
			{RightStrip: true}, // First statement strips right
			{},                 // Second statement no control
		}

		result := whitespace.ApplyWhitespaceControl(nodes, controls)

		if len(result) != 2 {
			t.Fatalf("Expected 2 nodes, got %d", len(result))
		}

		// First text node should have right whitespace stripped from next node
		if textNode, ok := result[1].(*parser.TextNode); ok {
			if textNode.Content != "World" {
				t.Errorf("Expected 'World', got %q", textNode.Content)
			}
		} else {
			t.Error("Expected second node to be TextNode")
		}
	})
}

func TestEnvironmentWhitespaceOptions(t *testing.T) {
	t.Run("Environment with trim blocks", func(t *testing.T) {
		env := NewEnvironment(WithTrimBlocks(true))

		if !env.trimBlocks {
			t.Error("Expected trimBlocks to be true")
		}

		if env.whitespaceProcessor == nil {
			t.Error("Expected whitespace processor to be initialized")
		}
	})

	t.Run("Environment with lstrip blocks", func(t *testing.T) {
		env := NewEnvironment(WithLstripBlocks(true))

		if !env.lstripBlocks {
			t.Error("Expected lstripBlocks to be true")
		}
	})

	t.Run("Environment with keep trailing newline", func(t *testing.T) {
		env := NewEnvironment(WithKeepTrailingNewline(true))

		if !env.keepTrailingNewline {
			t.Error("Expected keepTrailingNewline to be true")
		}
	})

	t.Run("Environment with all whitespace options", func(t *testing.T) {
		env := NewEnvironment(
			WithTrimBlocks(true),
			WithLstripBlocks(true),
			WithKeepTrailingNewline(true),
		)

		if !env.trimBlocks {
			t.Error("Expected trimBlocks to be true")
		}

		if !env.lstripBlocks {
			t.Error("Expected lstripBlocks to be true")
		}

		if !env.keepTrailingNewline {
			t.Error("Expected keepTrailingNewline to be true")
		}
	})
}
