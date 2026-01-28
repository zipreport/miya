package whitespace

import (
	"strings"

	"github.com/zipreport/miya/parser"
)

// WhitespaceProcessor handles whitespace control in templates
type WhitespaceProcessor struct {
	trimBlocks          bool
	lstripBlocks        bool
	keepTrailingNewline bool
}

// NewWhitespaceProcessor creates a new whitespace processor
func NewWhitespaceProcessor(trimBlocks, lstripBlocks, keepTrailingNewline bool) *WhitespaceProcessor {
	return &WhitespaceProcessor{
		trimBlocks:          trimBlocks,
		lstripBlocks:        lstripBlocks,
		keepTrailingNewline: keepTrailingNewline,
	}
}

// ProcessNodes processes a list of nodes and applies whitespace control
func (w *WhitespaceProcessor) ProcessNodes(nodes []parser.Node) []parser.Node {
	if len(nodes) == 0 {
		return nodes
	}

	result := make([]parser.Node, 0, len(nodes))

	for i, node := range nodes {
		processed := w.processNode(node, i, nodes)
		if processed != nil {
			result = append(result, processed)
		}
	}

	// Handle trailing newline
	if !w.keepTrailingNewline && len(result) > 0 {
		if textNode, ok := result[len(result)-1].(*parser.TextNode); ok {
			textNode.Content = strings.TrimSuffix(textNode.Content, "\n")
			if textNode.Content == "" {
				result = result[:len(result)-1]
			}
		}
	}

	return result
}

// processNode processes a single node with whitespace control
func (w *WhitespaceProcessor) processNode(node parser.Node, index int, allNodes []parser.Node) parser.Node {
	switch n := node.(type) {
	case *parser.TextNode:
		return w.processTextNode(n, index, allNodes)
	case *parser.CommentNode:
		// Comments are typically removed in output, but we preserve them in AST
		// They don't affect whitespace unless explicitly handled
		return n
	case *parser.RawNode:
		// Raw nodes preserve all content as-is
		return n
	default:
		// For other nodes, recursively process their children
		return w.processCompoundNode(node)
	}
}

// processTextNode applies whitespace rules to text nodes
func (w *WhitespaceProcessor) processTextNode(textNode *parser.TextNode, index int, allNodes []parser.Node) parser.Node {
	content := textNode.Content

	// Apply trim_blocks: remove first newline after block tags
	if w.trimBlocks && index > 0 {
		if w.isBlockStatement(allNodes[index-1]) {
			if strings.HasPrefix(content, "\n") {
				content = content[1:]
			} else if strings.HasPrefix(content, "\r\n") {
				content = content[2:]
			}
		}
	}

	// Apply lstrip_blocks: remove leading whitespace before block tags
	if w.lstripBlocks && index < len(allNodes)-1 {
		if w.isBlockStatement(allNodes[index+1]) {
			// Remove trailing whitespace from this text node
			lines := strings.Split(content, "\n")
			if len(lines) > 0 {
				lastLine := lines[len(lines)-1]
				// Check if last line contains only whitespace (spaces/tabs)
				trimmed := strings.TrimFunc(lastLine, func(r rune) bool {
					return r == ' ' || r == '\t'
				})
				if trimmed == "" && lastLine != "" {
					// Last line is only whitespace, remove it
					if len(lines) == 1 {
						content = ""
					} else {
						content = strings.Join(lines[:len(lines)-1], "\n") + "\n"
					}
				} else if lastLine != "" {
					// Remove trailing spaces/tabs from the last line
					newLastLine := strings.TrimRightFunc(lastLine, func(r rune) bool {
						return r == ' ' || r == '\t'
					})
					if len(lines) == 1 {
						content = newLastLine
					} else {
						lines[len(lines)-1] = newLastLine
						content = strings.Join(lines, "\n")
					}
				}
			}
		}
	}

	if content == "" {
		return nil // Remove empty text nodes
	}

	return &parser.TextNode{
		Content: content,
	}
}

// processCompoundNode recursively processes compound nodes
func (w *WhitespaceProcessor) processCompoundNode(node parser.Node) parser.Node {
	switch n := node.(type) {
	case *parser.IfNode:
		newNode := *n
		newNode.Body = w.ProcessNodes(n.Body)
		newNode.Else = w.ProcessNodes(n.Else)
		// Deep copy ElseIfs to avoid modifying original slice
		if len(n.ElseIfs) > 0 {
			newNode.ElseIfs = make([]*parser.IfNode, len(n.ElseIfs))
			for i, elif := range n.ElseIfs {
				newElif := *elif
				newElif.Body = w.ProcessNodes(elif.Body)
				newNode.ElseIfs[i] = &newElif
			}
		}
		return &newNode

	case *parser.ForNode:
		newNode := *n
		newNode.Body = w.ProcessNodes(n.Body)
		newNode.Else = w.ProcessNodes(n.Else)
		return &newNode

	case *parser.BlockNode:
		newNode := *n
		newNode.Body = w.ProcessNodes(n.Body)
		return &newNode

	case *parser.MacroNode:
		newNode := *n
		newNode.Body = w.ProcessNodes(n.Body)
		return &newNode

	default:
		return node
	}
}

// isBlockStatement checks if a node is a block statement that affects whitespace
func (w *WhitespaceProcessor) isBlockStatement(node parser.Node) bool {
	switch node.(type) {
	case *parser.IfNode, *parser.ForNode, *parser.BlockNode, *parser.MacroNode,
		*parser.SetNode, *parser.ExtendsNode, *parser.IncludeNode:
		return true
	default:
		return false
	}
}

// WhitespaceControl represents whitespace control modifiers for individual tags
type WhitespaceControl struct {
	LeftStrip  bool // Strip whitespace before tag (- modifier on left)
	RightStrip bool // Strip whitespace after tag (- modifier on right)
}

// NodeWithWhitespaceControl interface for nodes that support whitespace control
type NodeWithWhitespaceControl interface {
	parser.Node
	GetWhitespaceControl() WhitespaceControl
	SetWhitespaceControl(WhitespaceControl)
}

// ParseWhitespaceControl parses whitespace control from tag content
func ParseWhitespaceControl(content string) (string, WhitespaceControl) {
	control := WhitespaceControl{}

	// Check for left strip (-{%)
	if strings.HasPrefix(content, "-") {
		control.LeftStrip = true
		content = strings.TrimPrefix(content, "-")
		content = strings.TrimSpace(content)
	}

	// Check for right strip (%}-)
	if strings.HasSuffix(content, "-") {
		control.RightStrip = true
		content = strings.TrimSuffix(content, "-")
		content = strings.TrimSpace(content)
	}

	return content, control
}

// ApplyWhitespaceControl applies whitespace stripping based on control modifiers
func ApplyWhitespaceControl(nodes []parser.Node, controls []WhitespaceControl) []parser.Node {
	if len(nodes) == 0 {
		return nodes
	}

	result := make([]parser.Node, 0, len(nodes))

	for i, node := range nodes {
		processed := node

		// Apply whitespace control based on adjacent statement controls
		if textNode, ok := node.(*parser.TextNode); ok {
			content := textNode.Content

			// Check if previous statement has right strip
			if i > 0 && i-1 < len(controls) && controls[i-1].RightStrip {
				content = strings.TrimLeftFunc(content, func(r rune) bool {
					return r == ' ' || r == '\t'
				})
			}

			// Check if next statement has left strip
			if i < len(controls) && controls[i].LeftStrip {
				content = strings.TrimRightFunc(content, func(r rune) bool {
					return r == ' ' || r == '\t'
				})
			}

			if content != textNode.Content {
				processed = &parser.TextNode{Content: content}
			}
		}

		if textNode, ok := processed.(*parser.TextNode); ok && textNode.Content == "" {
			continue // Skip empty text nodes
		}

		result = append(result, processed)
	}

	return result
}
