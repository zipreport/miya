package parser

import "sync"

// Node pools for frequently allocated AST node types.
// These pools reduce GC pressure by reusing node allocations.
// Nodes should be returned to pools via ReleaseNode() after template execution.

var (
	// literalNodePool pools LiteralNode allocations (most common node type)
	literalNodePool = sync.Pool{
		New: func() interface{} { return &LiteralNode{} },
	}

	// identifierNodePool pools IdentifierNode allocations (variable references)
	identifierNodePool = sync.Pool{
		New: func() interface{} { return &IdentifierNode{} },
	}

	// binaryOpNodePool pools BinaryOpNode allocations (operators)
	binaryOpNodePool = sync.Pool{
		New: func() interface{} { return &BinaryOpNode{} },
	}

	// filterNodePool pools FilterNode allocations (filter applications)
	filterNodePool = sync.Pool{
		New: func() interface{} { return &FilterNode{} },
	}

	// unaryOpNodePool pools UnaryOpNode allocations (unary operators)
	unaryOpNodePool = sync.Pool{
		New: func() interface{} { return &UnaryOpNode{} },
	}

	// attributeNodePool pools AttributeNode allocations (attribute access)
	attributeNodePool = sync.Pool{
		New: func() interface{} { return &AttributeNode{} },
	}

	// getItemNodePool pools GetItemNode allocations (index access)
	getItemNodePool = sync.Pool{
		New: func() interface{} { return &GetItemNode{} },
	}

	// callNodePool pools CallNode allocations (function calls)
	callNodePool = sync.Pool{
		New: func() interface{} { return &CallNode{} },
	}
)

// AcquireLiteralNode gets a LiteralNode from the pool
func AcquireLiteralNode(value interface{}, raw string, line, column int) *LiteralNode {
	n := literalNodePool.Get().(*LiteralNode)
	n.baseNode.line = line
	n.baseNode.column = column
	n.Value = value
	n.Raw = raw
	return n
}

// ReleaseLiteralNode returns a LiteralNode to the pool
func ReleaseLiteralNode(n *LiteralNode) {
	if n == nil {
		return
	}
	// Reset fields to avoid memory leaks
	n.Value = nil
	n.Raw = ""
	n.baseNode.line = 0
	n.baseNode.column = 0
	literalNodePool.Put(n)
}

// AcquireIdentifierNode gets an IdentifierNode from the pool
func AcquireIdentifierNode(name string, line, column int) *IdentifierNode {
	n := identifierNodePool.Get().(*IdentifierNode)
	n.baseNode.line = line
	n.baseNode.column = column
	n.Name = name
	return n
}

// ReleaseIdentifierNode returns an IdentifierNode to the pool
func ReleaseIdentifierNode(n *IdentifierNode) {
	if n == nil {
		return
	}
	n.Name = ""
	n.baseNode.line = 0
	n.baseNode.column = 0
	identifierNodePool.Put(n)
}

// AcquireBinaryOpNode gets a BinaryOpNode from the pool
func AcquireBinaryOpNode(left ExpressionNode, op string, right ExpressionNode, line, column int) *BinaryOpNode {
	n := binaryOpNodePool.Get().(*BinaryOpNode)
	n.baseNode.line = line
	n.baseNode.column = column
	n.Left = left
	n.Operator = op
	n.Right = right
	return n
}

// ReleaseBinaryOpNode returns a BinaryOpNode to the pool
func ReleaseBinaryOpNode(n *BinaryOpNode) {
	if n == nil {
		return
	}
	n.Left = nil
	n.Operator = ""
	n.Right = nil
	n.baseNode.line = 0
	n.baseNode.column = 0
	binaryOpNodePool.Put(n)
}

// AcquireFilterNode gets a FilterNode from the pool
func AcquireFilterNode(expr ExpressionNode, filterName string, args []ExpressionNode, line, column int) *FilterNode {
	n := filterNodePool.Get().(*FilterNode)
	n.baseNode.line = line
	n.baseNode.column = column
	n.Expression = expr
	n.FilterName = filterName
	n.Arguments = args
	if n.NamedArgs == nil {
		n.NamedArgs = make(map[string]ExpressionNode)
	}
	return n
}

// ReleaseFilterNode returns a FilterNode to the pool
func ReleaseFilterNode(n *FilterNode) {
	if n == nil {
		return
	}
	n.Expression = nil
	n.FilterName = ""
	n.Arguments = nil
	// Clear the map instead of reallocating
	for k := range n.NamedArgs {
		delete(n.NamedArgs, k)
	}
	n.baseNode.line = 0
	n.baseNode.column = 0
	filterNodePool.Put(n)
}

// AcquireUnaryOpNode gets a UnaryOpNode from the pool
func AcquireUnaryOpNode(op string, operand ExpressionNode, line, column int) *UnaryOpNode {
	n := unaryOpNodePool.Get().(*UnaryOpNode)
	n.baseNode.line = line
	n.baseNode.column = column
	n.Operator = op
	n.Operand = operand
	return n
}

// ReleaseUnaryOpNode returns a UnaryOpNode to the pool
func ReleaseUnaryOpNode(n *UnaryOpNode) {
	if n == nil {
		return
	}
	n.Operator = ""
	n.Operand = nil
	n.baseNode.line = 0
	n.baseNode.column = 0
	unaryOpNodePool.Put(n)
}

// AcquireAttributeNode gets an AttributeNode from the pool
func AcquireAttributeNode(obj ExpressionNode, attr string, line, column int) *AttributeNode {
	n := attributeNodePool.Get().(*AttributeNode)
	n.baseNode.line = line
	n.baseNode.column = column
	n.Object = obj
	n.Attribute = attr
	return n
}

// ReleaseAttributeNode returns an AttributeNode to the pool
func ReleaseAttributeNode(n *AttributeNode) {
	if n == nil {
		return
	}
	n.Object = nil
	n.Attribute = ""
	n.baseNode.line = 0
	n.baseNode.column = 0
	attributeNodePool.Put(n)
}

// AcquireGetItemNode gets a GetItemNode from the pool
func AcquireGetItemNode(obj, key ExpressionNode, line, column int) *GetItemNode {
	n := getItemNodePool.Get().(*GetItemNode)
	n.baseNode.line = line
	n.baseNode.column = column
	n.Object = obj
	n.Key = key
	return n
}

// ReleaseGetItemNode returns a GetItemNode to the pool
func ReleaseGetItemNode(n *GetItemNode) {
	if n == nil {
		return
	}
	n.Object = nil
	n.Key = nil
	n.baseNode.line = 0
	n.baseNode.column = 0
	getItemNodePool.Put(n)
}

// AcquireCallNode gets a CallNode from the pool
func AcquireCallNode(function ExpressionNode, line, column int) *CallNode {
	n := callNodePool.Get().(*CallNode)
	n.baseNode.line = line
	n.baseNode.column = column
	n.Function = function
	n.Arguments = nil
	if n.Keywords == nil {
		n.Keywords = make(map[string]ExpressionNode)
	}
	return n
}

// ReleaseCallNode returns a CallNode to the pool
func ReleaseCallNode(n *CallNode) {
	if n == nil {
		return
	}
	n.Function = nil
	n.Arguments = nil
	for k := range n.Keywords {
		delete(n.Keywords, k)
	}
	n.baseNode.line = 0
	n.baseNode.column = 0
	callNodePool.Put(n)
}

// ReleaseNode releases a node back to its appropriate pool.
// This is a convenience function that handles type dispatching.
// For complex ASTs, use ReleaseAST which recursively releases all nodes.
func ReleaseNode(node Node) {
	if node == nil {
		return
	}

	switch n := node.(type) {
	case *LiteralNode:
		ReleaseLiteralNode(n)
	case *IdentifierNode:
		ReleaseIdentifierNode(n)
	case *BinaryOpNode:
		ReleaseBinaryOpNode(n)
	case *FilterNode:
		ReleaseFilterNode(n)
	case *UnaryOpNode:
		ReleaseUnaryOpNode(n)
	case *AttributeNode:
		ReleaseAttributeNode(n)
	case *GetItemNode:
		ReleaseGetItemNode(n)
	case *CallNode:
		ReleaseCallNode(n)
		// Non-pooled nodes are ignored (they'll be GC'd normally)
	}
}

// ReleaseAST recursively releases all pooled nodes in an AST back to their pools.
// This should be called when a template is no longer needed to enable node reuse.
// Non-pooled node types are ignored and will be garbage collected normally.
func ReleaseAST(node Node) {
	if node == nil {
		return
	}

	switch n := node.(type) {
	case *TemplateNode:
		for _, child := range n.Children {
			ReleaseAST(child)
		}
		// TemplateNode itself is not pooled

	case *TextNode:
		// TextNode is not pooled

	case *VariableNode:
		ReleaseAST(n.Expression)
		// VariableNode itself is not pooled

	case *LiteralNode:
		ReleaseLiteralNode(n)

	case *IdentifierNode:
		ReleaseIdentifierNode(n)

	case *BinaryOpNode:
		ReleaseAST(n.Left)
		ReleaseAST(n.Right)
		ReleaseBinaryOpNode(n)

	case *UnaryOpNode:
		ReleaseAST(n.Operand)
		ReleaseUnaryOpNode(n)

	case *FilterNode:
		ReleaseAST(n.Expression)
		for _, arg := range n.Arguments {
			ReleaseAST(arg)
		}
		for _, arg := range n.NamedArgs {
			ReleaseAST(arg)
		}
		ReleaseFilterNode(n)

	case *AttributeNode:
		ReleaseAST(n.Object)
		ReleaseAttributeNode(n)

	case *GetItemNode:
		ReleaseAST(n.Object)
		ReleaseAST(n.Key)
		ReleaseGetItemNode(n)

	case *CallNode:
		ReleaseAST(n.Function)
		for _, arg := range n.Arguments {
			ReleaseAST(arg)
		}
		for _, arg := range n.Keywords {
			ReleaseAST(arg)
		}
		ReleaseCallNode(n)

	case *ListNode:
		for _, elem := range n.Elements {
			ReleaseAST(elem)
		}
		// ListNode itself is not pooled

	case *IfNode:
		ReleaseAST(n.Condition)
		for _, child := range n.Body {
			ReleaseAST(child)
		}
		for _, elif := range n.ElseIfs {
			ReleaseAST(elif)
		}
		for _, child := range n.Else {
			ReleaseAST(child)
		}
		// IfNode itself is not pooled

	case *ForNode:
		ReleaseAST(n.Iterable)
		if n.Condition != nil {
			ReleaseAST(n.Condition)
		}
		for _, child := range n.Body {
			ReleaseAST(child)
		}
		for _, child := range n.Else {
			ReleaseAST(child)
		}
		// ForNode itself is not pooled

	case *BlockNode:
		for _, child := range n.Body {
			ReleaseAST(child)
		}
		// BlockNode itself is not pooled

	case *SetNode:
		for _, target := range n.Targets {
			ReleaseAST(target)
		}
		ReleaseAST(n.Value)
		// SetNode itself is not pooled

	case *BlockSetNode:
		for _, child := range n.Body {
			ReleaseAST(child)
		}
		// BlockSetNode itself is not pooled

	case *MacroNode:
		for _, defaultExpr := range n.Defaults {
			ReleaseAST(defaultExpr)
		}
		for _, child := range n.Body {
			ReleaseAST(child)
		}
		// MacroNode itself is not pooled

	case *TestNode:
		ReleaseAST(n.Expression)
		for _, arg := range n.Arguments {
			ReleaseAST(arg)
		}
		// TestNode itself is not pooled

	case *ConditionalNode:
		ReleaseAST(n.Condition)
		ReleaseAST(n.TrueExpr)
		ReleaseAST(n.FalseExpr)
		// ConditionalNode itself is not pooled

	case *SliceNode:
		ReleaseAST(n.Object)
		if n.Start != nil {
			ReleaseAST(n.Start)
		}
		if n.End != nil {
			ReleaseAST(n.End)
		}
		if n.Step != nil {
			ReleaseAST(n.Step)
		}
		// SliceNode itself is not pooled

	case *ComprehensionNode:
		ReleaseAST(n.Expression)
		ReleaseAST(n.Iterable)
		if n.Condition != nil {
			ReleaseAST(n.Condition)
		}
		if n.KeyExpr != nil {
			ReleaseAST(n.KeyExpr)
		}
		// ComprehensionNode itself is not pooled

	case *WithNode:
		for _, expr := range n.Assignments {
			ReleaseAST(expr)
		}
		for _, child := range n.Body {
			ReleaseAST(child)
		}
		// WithNode itself is not pooled

	case *CallBlockNode:
		ReleaseAST(n.Call)
		for _, child := range n.Body {
			ReleaseAST(child)
		}
		// CallBlockNode itself is not pooled

	case *ExtendsNode:
		ReleaseAST(n.Template)
		// ExtendsNode itself is not pooled

	case *IncludeNode:
		ReleaseAST(n.Template)
		if n.Context != nil {
			ReleaseAST(n.Context)
		}
		// IncludeNode itself is not pooled

	case *ImportNode:
		ReleaseAST(n.Template)
		// ImportNode itself is not pooled

	case *FromNode:
		ReleaseAST(n.Template)
		// FromNode itself is not pooled

	case *DoNode:
		ReleaseAST(n.Expression)
		// DoNode itself is not pooled

	case *AutoescapeNode:
		for _, child := range n.Body {
			ReleaseAST(child)
		}
		// AutoescapeNode itself is not pooled

	case *FilterBlockNode:
		for _, child := range n.Body {
			ReleaseAST(child)
		}
		// FilterBlockNode itself is not pooled

	case *ExtensionNode:
		for _, arg := range n.Arguments {
			ReleaseAST(arg)
		}
		for _, child := range n.Body {
			ReleaseAST(child)
		}
		// ExtensionNode itself is not pooled

		// Nodes that don't contain child nodes and are not pooled:
		// SuperNode, BreakNode, ContinueNode, CommentNode, RawNode, AssignmentNode
	}
}

// NodePoolStats returns statistics about node pool usage.
// This is useful for debugging and performance tuning.
type NodePoolStats struct {
	LiteralNodes    int
	IdentifierNodes int
	BinaryOpNodes   int
	FilterNodes     int
	UnaryOpNodes    int
	AttributeNodes  int
	GetItemNodes    int
	CallNodes       int
}

// Note: sync.Pool doesn't expose internal statistics,
// so this function is a placeholder for future instrumentation.
// For now, pool effectiveness should be measured via benchmarks.
func GetNodePoolStats() NodePoolStats {
	return NodePoolStats{}
}
