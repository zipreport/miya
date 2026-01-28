package whitespace

import (
	"testing"

	"github.com/zipreport/miya/parser"
)

// Test WhitespaceProcessor creation and configuration
func TestNewWhitespaceProcessor(t *testing.T) {
	t.Run("Creates processor with correct settings", func(t *testing.T) {
		processor := NewWhitespaceProcessor(true, false, true)

		if !processor.trimBlocks {
			t.Error("Expected trimBlocks to be true")
		}
		if processor.lstripBlocks {
			t.Error("Expected lstripBlocks to be false")
		}
		if !processor.keepTrailingNewline {
			t.Error("Expected keepTrailingNewline to be true")
		}
	})

	t.Run("Creates processor with all options disabled", func(t *testing.T) {
		processor := NewWhitespaceProcessor(false, false, false)

		if processor.trimBlocks {
			t.Error("Expected trimBlocks to be false")
		}
		if processor.lstripBlocks {
			t.Error("Expected lstripBlocks to be false")
		}
		if processor.keepTrailingNewline {
			t.Error("Expected keepTrailingNewline to be false")
		}
	})
}

// Test ProcessNodes functionality
func TestWhitespaceProcessorProcessNodes(t *testing.T) {
	t.Run("Empty nodes returns empty", func(t *testing.T) {
		processor := NewWhitespaceProcessor(false, false, true)
		nodes := []parser.Node{}

		result := processor.ProcessNodes(nodes)

		if len(result) != 0 {
			t.Errorf("Expected 0 nodes, got %d", len(result))
		}
	})

	t.Run("Single text node processed", func(t *testing.T) {
		processor := NewWhitespaceProcessor(false, false, true)
		nodes := []parser.Node{
			&parser.TextNode{Content: "Hello World"},
		}

		result := processor.ProcessNodes(nodes)

		if len(result) != 1 {
			t.Errorf("Expected 1 node, got %d", len(result))
		}

		if textNode, ok := result[0].(*parser.TextNode); ok {
			if textNode.Content != "Hello World" {
				t.Errorf("Expected 'Hello World', got '%s'", textNode.Content)
			}
		} else {
			t.Error("Expected TextNode")
		}
	})

	t.Run("Removes trailing newline when keepTrailingNewline is false", func(t *testing.T) {
		processor := NewWhitespaceProcessor(false, false, false)
		nodes := []parser.Node{
			&parser.TextNode{Content: "Hello World\n"},
		}

		result := processor.ProcessNodes(nodes)

		if len(result) != 1 {
			t.Errorf("Expected 1 node, got %d", len(result))
		}

		if textNode, ok := result[0].(*parser.TextNode); ok {
			if textNode.Content != "Hello World" {
				t.Errorf("Expected 'Hello World', got '%s'", textNode.Content)
			}
		} else {
			t.Error("Expected TextNode")
		}
	})

	t.Run("Keeps trailing newline when keepTrailingNewline is true", func(t *testing.T) {
		processor := NewWhitespaceProcessor(false, false, true)
		nodes := []parser.Node{
			&parser.TextNode{Content: "Hello World\n"},
		}

		result := processor.ProcessNodes(nodes)

		if len(result) != 1 {
			t.Errorf("Expected 1 node, got %d", len(result))
		}

		if textNode, ok := result[0].(*parser.TextNode); ok {
			if textNode.Content != "Hello World\n" {
				t.Errorf("Expected 'Hello World\\n', got '%s'", textNode.Content)
			}
		} else {
			t.Error("Expected TextNode")
		}
	})

	t.Run("Removes empty text nodes after processing", func(t *testing.T) {
		processor := NewWhitespaceProcessor(false, false, false)
		nodes := []parser.Node{
			&parser.TextNode{Content: "\n"},
		}

		result := processor.ProcessNodes(nodes)

		if len(result) != 0 {
			t.Errorf("Expected 0 nodes (empty node removed), got %d", len(result))
		}
	})
}

// Test trim_blocks functionality
func TestTrimBlocks(t *testing.T) {
	t.Run("Trims newline after block statement", func(t *testing.T) {
		processor := NewWhitespaceProcessor(true, false, true)

		// Mock IfNode as block statement
		ifNode := &parser.IfNode{}
		nodes := []parser.Node{
			ifNode,
			&parser.TextNode{Content: "\nHello"},
		}

		result := processor.ProcessNodes(nodes)

		if len(result) != 2 {
			t.Errorf("Expected 2 nodes, got %d", len(result))
		}

		if textNode, ok := result[1].(*parser.TextNode); ok {
			if textNode.Content != "Hello" {
				t.Errorf("Expected 'Hello', got '%s'", textNode.Content)
			}
		} else {
			t.Error("Expected TextNode as second result")
		}
	})

	t.Run("Trims CRLF after block statement", func(t *testing.T) {
		processor := NewWhitespaceProcessor(true, false, true)

		// Mock ForNode as block statement
		forNode := &parser.ForNode{}
		nodes := []parser.Node{
			forNode,
			&parser.TextNode{Content: "\r\nHello"},
		}

		result := processor.ProcessNodes(nodes)

		if len(result) != 2 {
			t.Errorf("Expected 2 nodes, got %d", len(result))
		}

		if textNode, ok := result[1].(*parser.TextNode); ok {
			if textNode.Content != "Hello" {
				t.Errorf("Expected 'Hello', got '%s'", textNode.Content)
			}
		} else {
			t.Error("Expected TextNode as second result")
		}
	})

	t.Run("Does not trim when trimBlocks is false", func(t *testing.T) {
		processor := NewWhitespaceProcessor(false, false, true)

		ifNode := &parser.IfNode{}
		nodes := []parser.Node{
			ifNode,
			&parser.TextNode{Content: "\nHello"},
		}

		result := processor.ProcessNodes(nodes)

		if textNode, ok := result[1].(*parser.TextNode); ok {
			if textNode.Content != "\nHello" {
				t.Errorf("Expected '\\nHello', got '%s'", textNode.Content)
			}
		}
	})
}

// Test lstrip_blocks functionality
func TestLstripBlocks(t *testing.T) {
	t.Run("Strips leading whitespace before block statement", func(t *testing.T) {
		processor := NewWhitespaceProcessor(false, true, true)

		ifNode := &parser.IfNode{}
		nodes := []parser.Node{
			&parser.TextNode{Content: "Hello\n    "},
			ifNode,
		}

		result := processor.ProcessNodes(nodes)

		if len(result) != 2 {
			t.Errorf("Expected 2 nodes, got %d", len(result))
		}

		if textNode, ok := result[0].(*parser.TextNode); ok {
			if textNode.Content != "Hello\n" {
				t.Errorf("Expected 'Hello\\n', got '%s'", textNode.Content)
			}
		} else {
			t.Error("Expected TextNode as first result")
		}
	})

	t.Run("Removes whitespace-only line before block", func(t *testing.T) {
		processor := NewWhitespaceProcessor(false, true, true)

		blockNode := &parser.BlockNode{}
		nodes := []parser.Node{
			&parser.TextNode{Content: "Hello\n\t  "},
			blockNode,
		}

		result := processor.ProcessNodes(nodes)

		if textNode, ok := result[0].(*parser.TextNode); ok {
			if textNode.Content != "Hello\n" {
				t.Errorf("Expected 'Hello\\n', got '%s'", textNode.Content)
			}
		}
	})

	t.Run("Handles single line whitespace-only content", func(t *testing.T) {
		processor := NewWhitespaceProcessor(false, true, true)

		macroNode := &parser.MacroNode{}
		nodes := []parser.Node{
			&parser.TextNode{Content: "   \t  "},
			macroNode,
		}

		result := processor.ProcessNodes(nodes)

		// Should remove empty text node
		if len(result) != 1 {
			t.Errorf("Expected 1 node after removing empty text, got %d", len(result))
		}

		if _, ok := result[0].(*parser.MacroNode); !ok {
			t.Error("Expected MacroNode to remain")
		}
	})
}

// Test block statement detection
func TestIsBlockStatement(t *testing.T) {
	processor := NewWhitespaceProcessor(false, false, true)

	testCases := []struct {
		name     string
		node     parser.Node
		expected bool
	}{
		{"IfNode is block statement", &parser.IfNode{}, true},
		{"ForNode is block statement", &parser.ForNode{}, true},
		{"BlockNode is block statement", &parser.BlockNode{}, true},
		{"MacroNode is block statement", &parser.MacroNode{}, true},
		{"SetNode is block statement", &parser.SetNode{}, true},
		{"ExtendsNode is block statement", &parser.ExtendsNode{}, true},
		{"IncludeNode is block statement", &parser.IncludeNode{}, true},
		{"TextNode is not block statement", &parser.TextNode{}, false},
		{"CommentNode is not block statement", &parser.CommentNode{}, false},
		{"VariableNode is not block statement", &parser.VariableNode{}, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := processor.isBlockStatement(tc.node)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

// Test compound node processing
func TestProcessCompoundNode(t *testing.T) {
	t.Run("Processes IfNode children recursively", func(t *testing.T) {
		processor := NewWhitespaceProcessor(false, false, false)

		ifNode := &parser.IfNode{
			Body: []parser.Node{
				&parser.TextNode{Content: "Body content\n"},
			},
			Else: []parser.Node{
				&parser.TextNode{Content: "Else content\n"},
			},
			ElseIfs: []*parser.IfNode{
				{
					Body: []parser.Node{
						&parser.TextNode{Content: "Elif content\n"},
					},
				},
			},
		}

		result := processor.processCompoundNode(ifNode)

		processedIf, ok := result.(*parser.IfNode)
		if !ok {
			t.Fatal("Expected IfNode")
		}

		// Check body processing (should have trailing newline removed)
		if len(processedIf.Body) != 1 {
			t.Error("Expected 1 body node")
		}
		if bodyText, ok := processedIf.Body[0].(*parser.TextNode); ok {
			if bodyText.Content != "Body content" {
				t.Errorf("Expected 'Body content', got '%s'", bodyText.Content)
			}
		}

		// Check else processing
		if len(processedIf.Else) != 1 {
			t.Error("Expected 1 else node")
		}
		if elseText, ok := processedIf.Else[0].(*parser.TextNode); ok {
			if elseText.Content != "Else content" {
				t.Errorf("Expected 'Else content', got '%s'", elseText.Content)
			}
		}

		// Check elif processing
		if len(processedIf.ElseIfs) != 1 {
			t.Error("Expected 1 elif")
		}
		if len(processedIf.ElseIfs[0].Body) != 1 {
			t.Error("Expected 1 elif body node")
		}
		if elifText, ok := processedIf.ElseIfs[0].Body[0].(*parser.TextNode); ok {
			if elifText.Content != "Elif content" {
				t.Errorf("Expected 'Elif content', got '%s'", elifText.Content)
			}
		}
	})

	t.Run("Processes ForNode children recursively", func(t *testing.T) {
		processor := NewWhitespaceProcessor(false, false, false)

		forNode := &parser.ForNode{
			Body: []parser.Node{
				&parser.TextNode{Content: "Loop body\n"},
			},
			Else: []parser.Node{
				&parser.TextNode{Content: "Loop else\n"},
			},
		}

		result := processor.processCompoundNode(forNode)

		processedFor, ok := result.(*parser.ForNode)
		if !ok {
			t.Fatal("Expected ForNode")
		}

		if len(processedFor.Body) != 1 {
			t.Error("Expected 1 body node")
		}
		if bodyText, ok := processedFor.Body[0].(*parser.TextNode); ok {
			if bodyText.Content != "Loop body" {
				t.Errorf("Expected 'Loop body', got '%s'", bodyText.Content)
			}
		}
	})

	t.Run("Returns other nodes unchanged", func(t *testing.T) {
		processor := NewWhitespaceProcessor(false, false, true)

		textNode := &parser.TextNode{Content: "unchanged"}
		result := processor.processCompoundNode(textNode)

		if result != textNode {
			t.Error("Expected same TextNode instance")
		}
	})
}

// Test special node handling
func TestSpecialNodeHandling(t *testing.T) {
	t.Run("CommentNode preserved unchanged", func(t *testing.T) {
		processor := NewWhitespaceProcessor(true, true, false)

		commentNode := &parser.CommentNode{Content: "test comment"}
		nodes := []parser.Node{commentNode}

		result := processor.ProcessNodes(nodes)

		if len(result) != 1 {
			t.Errorf("Expected 1 node, got %d", len(result))
		}

		if result[0] != commentNode {
			t.Error("Expected same CommentNode instance")
		}
	})

	t.Run("RawNode preserved unchanged", func(t *testing.T) {
		processor := NewWhitespaceProcessor(true, true, false)

		rawNode := &parser.RawNode{Content: "  raw content  \n"}
		nodes := []parser.Node{rawNode}

		result := processor.ProcessNodes(nodes)

		if len(result) != 1 {
			t.Errorf("Expected 1 node, got %d", len(result))
		}

		if result[0] != rawNode {
			t.Error("Expected same RawNode instance")
		}
	})
}
