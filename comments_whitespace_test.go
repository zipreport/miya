package miya

import (
	"testing"

	"github.com/zipreport/miya/parser"
	"github.com/zipreport/miya/runtime"
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

func TestEnvironmentWhitespaceOptions(t *testing.T) {
	t.Run("Environment with trim blocks", func(t *testing.T) {
		env := NewEnvironment(WithTrimBlocks(true))

		if !env.trimBlocks {
			t.Error("Expected trimBlocks to be true")
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
