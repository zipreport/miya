package parser

// This file contains FastEval implementations for Phase 3c optimization
// FastEval allows nodes to evaluate themselves directly without going through
// the type switch in DefaultEvaluator.EvalNode()

// FastEvalNode interface is defined in runtime package to avoid circular dependencies
// We use a local type alias here for convenience
type FastEvaluator interface {
	SetUndefinedBehavior(behavior interface{})
	SetImportSystem(importSystem interface{})
}

// FastEval for TextNode - the most common node type
// Simply returns the text content without any processing
func (n *TextNode) FastEval(e interface{}, ctx interface{}) (interface{}, error) {
	return n.Content, nil
}

// FastEval for LiteralNode - direct value return
func (n *LiteralNode) FastEval(e interface{}, ctx interface{}) (interface{}, error) {
	return n.Value, nil
}

// FastEval for CommentNode - returns empty string
func (n *CommentNode) FastEval(e interface{}, ctx interface{}) (interface{}, error) {
	return "", nil
}

// FastEval for RawNode - returns content as-is
func (n *RawNode) FastEval(e interface{}, ctx interface{}) (interface{}, error) {
	return n.Content, nil
}
