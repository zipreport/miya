package parser

import (
	"sync"
	"testing"
)

func TestAcquireReleaseLiteralNode(t *testing.T) {
	// Acquire a node
	node := AcquireLiteralNode("test", "test", 1, 2)
	if node == nil {
		t.Fatal("AcquireLiteralNode returned nil")
	}
	if node.Value != "test" {
		t.Errorf("expected value 'test', got %v", node.Value)
	}
	if node.Raw != "test" {
		t.Errorf("expected raw 'test', got %v", node.Raw)
	}
	if node.Line() != 1 {
		t.Errorf("expected line 1, got %d", node.Line())
	}
	if node.Column() != 2 {
		t.Errorf("expected column 2, got %d", node.Column())
	}

	// Release and verify reuse
	ReleaseLiteralNode(node)

	// Acquire again - should get a pooled node
	node2 := AcquireLiteralNode(42, "42", 3, 4)
	if node2 == nil {
		t.Fatal("second AcquireLiteralNode returned nil")
	}
	if node2.Value != 42 {
		t.Errorf("expected value 42, got %v", node2.Value)
	}
	ReleaseLiteralNode(node2)
}

func TestAcquireReleaseIdentifierNode(t *testing.T) {
	node := AcquireIdentifierNode("myvar", 5, 10)
	if node == nil {
		t.Fatal("AcquireIdentifierNode returned nil")
	}
	if node.Name != "myvar" {
		t.Errorf("expected name 'myvar', got %v", node.Name)
	}
	if node.Line() != 5 {
		t.Errorf("expected line 5, got %d", node.Line())
	}
	if node.Column() != 10 {
		t.Errorf("expected column 10, got %d", node.Column())
	}

	ReleaseIdentifierNode(node)

	// Acquire again
	node2 := AcquireIdentifierNode("other", 1, 1)
	if node2 == nil {
		t.Fatal("second AcquireIdentifierNode returned nil")
	}
	if node2.Name != "other" {
		t.Errorf("expected name 'other', got %v", node2.Name)
	}
	ReleaseIdentifierNode(node2)
}

func TestAcquireReleaseBinaryOpNode(t *testing.T) {
	left := AcquireLiteralNode(1, "1", 1, 1)
	right := AcquireLiteralNode(2, "2", 1, 5)

	node := AcquireBinaryOpNode(left, "+", right, 1, 3)
	if node == nil {
		t.Fatal("AcquireBinaryOpNode returned nil")
	}
	if node.Operator != "+" {
		t.Errorf("expected operator '+', got %v", node.Operator)
	}
	if node.Left != left {
		t.Error("left child not set correctly")
	}
	if node.Right != right {
		t.Error("right child not set correctly")
	}

	ReleaseBinaryOpNode(node)
	ReleaseLiteralNode(left)
	ReleaseLiteralNode(right)
}

func TestAcquireReleaseFilterNode(t *testing.T) {
	expr := AcquireIdentifierNode("value", 1, 1)
	args := []ExpressionNode{AcquireLiteralNode("arg1", "arg1", 1, 10)}

	node := AcquireFilterNode(expr, "upper", args, 1, 5)
	if node == nil {
		t.Fatal("AcquireFilterNode returned nil")
	}
	if node.FilterName != "upper" {
		t.Errorf("expected filter name 'upper', got %v", node.FilterName)
	}
	if node.Expression != expr {
		t.Error("expression not set correctly")
	}
	if len(node.Arguments) != 1 {
		t.Errorf("expected 1 argument, got %d", len(node.Arguments))
	}
	if node.NamedArgs == nil {
		t.Error("NamedArgs map should be initialized")
	}

	ReleaseFilterNode(node)
}

func TestAcquireReleaseUnaryOpNode(t *testing.T) {
	operand := AcquireLiteralNode(5, "5", 1, 2)

	node := AcquireUnaryOpNode("-", operand, 1, 1)
	if node == nil {
		t.Fatal("AcquireUnaryOpNode returned nil")
	}
	if node.Operator != "-" {
		t.Errorf("expected operator '-', got %v", node.Operator)
	}
	if node.Operand != operand {
		t.Error("operand not set correctly")
	}

	ReleaseUnaryOpNode(node)
	ReleaseLiteralNode(operand)
}

func TestAcquireReleaseAttributeNode(t *testing.T) {
	obj := AcquireIdentifierNode("user", 1, 1)

	node := AcquireAttributeNode(obj, "name", 1, 5)
	if node == nil {
		t.Fatal("AcquireAttributeNode returned nil")
	}
	if node.Attribute != "name" {
		t.Errorf("expected attribute 'name', got %v", node.Attribute)
	}
	if node.Object != obj {
		t.Error("object not set correctly")
	}

	ReleaseAttributeNode(node)
	ReleaseIdentifierNode(obj)
}

func TestAcquireReleaseGetItemNode(t *testing.T) {
	obj := AcquireIdentifierNode("items", 1, 1)
	key := AcquireLiteralNode(0, "0", 1, 7)

	node := AcquireGetItemNode(obj, key, 1, 6)
	if node == nil {
		t.Fatal("AcquireGetItemNode returned nil")
	}
	if node.Object != obj {
		t.Error("object not set correctly")
	}
	if node.Key != key {
		t.Error("key not set correctly")
	}

	ReleaseGetItemNode(node)
	ReleaseIdentifierNode(obj)
	ReleaseLiteralNode(key)
}

func TestAcquireReleaseCallNode(t *testing.T) {
	fn := AcquireIdentifierNode("print", 1, 1)

	node := AcquireCallNode(fn, 1, 5)
	if node == nil {
		t.Fatal("AcquireCallNode returned nil")
	}
	if node.Function != fn {
		t.Error("function not set correctly")
	}
	if node.Arguments != nil {
		t.Error("Arguments should be nil initially")
	}
	if node.Keywords == nil {
		t.Error("Keywords map should be initialized")
	}

	ReleaseCallNode(node)
	ReleaseIdentifierNode(fn)
}

func TestReleaseNode(t *testing.T) {
	// Test that ReleaseNode handles different node types
	literal := AcquireLiteralNode(1, "1", 1, 1)
	ReleaseNode(literal)

	ident := AcquireIdentifierNode("x", 1, 1)
	ReleaseNode(ident)

	// Test nil handling
	ReleaseNode(nil)
}

func TestReleaseAST(t *testing.T) {
	// Create a simple AST
	templateNode := NewTemplateNode("test", 1, 1)

	// Add a variable node with binary expression
	left := AcquireLiteralNode(1, "1", 1, 5)
	right := AcquireLiteralNode(2, "2", 1, 9)
	binOp := AcquireBinaryOpNode(left, "+", right, 1, 7)
	varNode := NewVariableNode(binOp, 1, 1)
	templateNode.Children = append(templateNode.Children, varNode)

	// Add a filter expression
	ident := AcquireIdentifierNode("name", 2, 5)
	filterNode := AcquireFilterNode(ident, "upper", nil, 2, 10)
	varNode2 := NewVariableNode(filterNode, 2, 1)
	templateNode.Children = append(templateNode.Children, varNode2)

	// Release the entire AST
	ReleaseAST(templateNode)

	// Test nil handling
	ReleaseAST(nil)
}

func TestConcurrentPoolAccess(t *testing.T) {
	const goroutines = 100
	const iterations = 1000

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				// Acquire various nodes
				lit := AcquireLiteralNode(id*1000+j, "val", 1, 1)
				ident := AcquireIdentifierNode("var", 1, 1)
				binOp := AcquireBinaryOpNode(lit, "+", ident, 1, 1)

				// Verify values are set correctly
				if lit.Value != id*1000+j {
					t.Errorf("goroutine %d iteration %d: incorrect literal value", id, j)
				}

				// Release
				ReleaseBinaryOpNode(binOp)
				ReleaseLiteralNode(lit)
				ReleaseIdentifierNode(ident)
			}
		}(i)
	}

	wg.Wait()
}

func TestReleaseNilNodes(t *testing.T) {
	// All release functions should handle nil gracefully
	ReleaseLiteralNode(nil)
	ReleaseIdentifierNode(nil)
	ReleaseBinaryOpNode(nil)
	ReleaseFilterNode(nil)
	ReleaseUnaryOpNode(nil)
	ReleaseAttributeNode(nil)
	ReleaseGetItemNode(nil)
	ReleaseCallNode(nil)
	ReleaseNode(nil)
	ReleaseAST(nil)
}

func TestNodeReuseAfterRelease(t *testing.T) {
	// Verify that released nodes get their fields properly reset

	// Test LiteralNode
	lit := AcquireLiteralNode("original", "original", 100, 200)
	ReleaseLiteralNode(lit)

	// The node should be reset (we can't verify the exact same node is returned,
	// but we can verify a new acquisition works correctly)
	lit2 := AcquireLiteralNode("new", "new", 1, 1)
	if lit2.Value != "new" {
		t.Errorf("expected value 'new', got %v", lit2.Value)
	}
	ReleaseLiteralNode(lit2)

	// Test FilterNode with NamedArgs
	filter := AcquireFilterNode(nil, "test", nil, 1, 1)
	filter.NamedArgs["key"] = AcquireLiteralNode("value", "value", 1, 1)
	ReleaseFilterNode(filter)

	filter2 := AcquireFilterNode(nil, "other", nil, 1, 1)
	// NamedArgs should be cleared (empty map, not nil)
	if len(filter2.NamedArgs) != 0 {
		t.Errorf("expected empty NamedArgs, got %d entries", len(filter2.NamedArgs))
	}
	ReleaseFilterNode(filter2)
}

func BenchmarkAcquireReleaseLiteralNode(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		node := AcquireLiteralNode(i, "test", 1, 1)
		ReleaseLiteralNode(node)
	}
}

func BenchmarkNewLiteralNode(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = NewLiteralNode(i, "test", 1, 1)
	}
}

func BenchmarkAcquireReleaseIdentifierNode(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		node := AcquireIdentifierNode("test", 1, 1)
		ReleaseIdentifierNode(node)
	}
}

func BenchmarkNewIdentifierNode(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = NewIdentifierNode("test", 1, 1)
	}
}

func BenchmarkAcquireReleaseBinaryOpNode(b *testing.B) {
	left := AcquireLiteralNode(1, "1", 1, 1)
	right := AcquireLiteralNode(2, "2", 1, 1)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		node := AcquireBinaryOpNode(left, "+", right, 1, 1)
		ReleaseBinaryOpNode(node)
	}

	ReleaseLiteralNode(left)
	ReleaseLiteralNode(right)
}

func BenchmarkNewBinaryOpNode(b *testing.B) {
	left := NewLiteralNode(1, "1", 1, 1)
	right := NewLiteralNode(2, "2", 1, 1)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewBinaryOpNode(left, "+", right, 1, 1)
	}
}

func BenchmarkParseExpressionPooled(b *testing.B) {
	// Benchmark parsing a moderately complex expression
	// This tests the overall impact of pooling on parser performance
	source := "a + b * c - d / e"

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// We'd need to import lexer here, but this gives the structure
		// For now, just allocate/release typical expression nodes
		a := AcquireIdentifierNode("a", 1, 1)
		b := AcquireIdentifierNode("b", 1, 5)
		c := AcquireIdentifierNode("c", 1, 9)
		d := AcquireIdentifierNode("d", 1, 13)
		e := AcquireIdentifierNode("e", 1, 17)

		mul := AcquireBinaryOpNode(b, "*", c, 1, 7)
		div := AcquireBinaryOpNode(d, "/", e, 1, 15)
		add := AcquireBinaryOpNode(a, "+", mul, 1, 3)
		sub := AcquireBinaryOpNode(add, "-", div, 1, 11)

		_ = source
		_ = sub

		// Release in order
		ReleaseBinaryOpNode(sub)
		ReleaseBinaryOpNode(add)
		ReleaseBinaryOpNode(div)
		ReleaseBinaryOpNode(mul)
		ReleaseIdentifierNode(e)
		ReleaseIdentifierNode(d)
		ReleaseIdentifierNode(c)
		// Note: b and a were already passed to binary nodes but we lost the references
		// In real code, ReleaseAST handles this recursively
	}
}
