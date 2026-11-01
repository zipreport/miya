package parser

// AST nodes are allocated normally and garbage collected. The Release*
// functions and node pool statistics remain for compatibility and do nothing.

// AcquireLiteralNode returns a new LiteralNode
func AcquireLiteralNode(value interface{}, raw string, line, column int) *LiteralNode {
	n := &LiteralNode{}
	n.baseNode.line = line
	n.baseNode.column = column
	n.Value = value
	n.Raw = raw
	return n
}

// Release*Node is a no-op kept for compatibility.
//
// Deprecated: AST nodes are not pooled; this does nothing.
func ReleaseLiteralNode(n *LiteralNode) {
}

// AcquireIdentifierNode returns a new IdentifierNode
func AcquireIdentifierNode(name string, line, column int) *IdentifierNode {
	n := &IdentifierNode{}
	n.baseNode.line = line
	n.baseNode.column = column
	n.Name = name
	return n
}

// Release*Node is a no-op kept for compatibility.
//
// Deprecated: AST nodes are not pooled; this does nothing.
func ReleaseIdentifierNode(n *IdentifierNode) {
}

// AcquireBinaryOpNode returns a new BinaryOpNode
func AcquireBinaryOpNode(left ExpressionNode, op string, right ExpressionNode, line, column int) *BinaryOpNode {
	n := &BinaryOpNode{}
	n.baseNode.line = line
	n.baseNode.column = column
	n.Left = left
	n.Operator = op
	n.Right = right
	return n
}

// Release*Node is a no-op kept for compatibility.
//
// Deprecated: AST nodes are not pooled; this does nothing.
func ReleaseBinaryOpNode(n *BinaryOpNode) {
}

// AcquireFilterNode returns a new FilterNode
func AcquireFilterNode(expr ExpressionNode, filterName string, args []ExpressionNode, line, column int) *FilterNode {
	n := &FilterNode{}
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

// Release*Node is a no-op kept for compatibility.
//
// Deprecated: AST nodes are not pooled; this does nothing.
func ReleaseFilterNode(n *FilterNode) {
}

// AcquireUnaryOpNode returns a new UnaryOpNode
func AcquireUnaryOpNode(op string, operand ExpressionNode, line, column int) *UnaryOpNode {
	n := &UnaryOpNode{}
	n.baseNode.line = line
	n.baseNode.column = column
	n.Operator = op
	n.Operand = operand
	return n
}

// Release*Node is a no-op kept for compatibility.
//
// Deprecated: AST nodes are not pooled; this does nothing.
func ReleaseUnaryOpNode(n *UnaryOpNode) {
}

// AcquireAttributeNode returns a new AttributeNode
func AcquireAttributeNode(obj ExpressionNode, attr string, line, column int) *AttributeNode {
	n := &AttributeNode{}
	n.baseNode.line = line
	n.baseNode.column = column
	n.Object = obj
	n.Attribute = attr
	return n
}

// Release*Node is a no-op kept for compatibility.
//
// Deprecated: AST nodes are not pooled; this does nothing.
func ReleaseAttributeNode(n *AttributeNode) {
}

// AcquireGetItemNode returns a new GetItemNode
func AcquireGetItemNode(obj, key ExpressionNode, line, column int) *GetItemNode {
	n := &GetItemNode{}
	n.baseNode.line = line
	n.baseNode.column = column
	n.Object = obj
	n.Key = key
	return n
}

// Release*Node is a no-op kept for compatibility.
//
// Deprecated: AST nodes are not pooled; this does nothing.
func ReleaseGetItemNode(n *GetItemNode) {
}

// AcquireCallNode returns a new CallNode
func AcquireCallNode(function ExpressionNode, line, column int) *CallNode {
	n := &CallNode{}
	n.baseNode.line = line
	n.baseNode.column = column
	n.Function = function
	n.Arguments = nil
	if n.Keywords == nil {
		n.Keywords = make(map[string]ExpressionNode)
	}
	return n
}

// Release*Node is a no-op kept for compatibility.
//
// Deprecated: AST nodes are not pooled; this does nothing.
func ReleaseCallNode(n *CallNode) {
}

// Release*Node is a no-op kept for compatibility.
//
// Deprecated: AST nodes are not pooled; this does nothing.
func ReleaseNode(node Node) {
}

// Release*Node is a no-op kept for compatibility.
//
// Deprecated: AST nodes are not pooled; this does nothing.
func ReleaseAST(node Node) {
}

// NodePoolStats is kept for compatibility.
//
// Deprecated: nodes are not pooled; the stats are always zero.
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

// GetNodePoolStats returns zeroed statistics.
//
// Deprecated: nodes are not pooled; the stats are always zero.
func GetNodePoolStats() NodePoolStats {
	return NodePoolStats{}
}
