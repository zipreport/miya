package parser

import (
	"testing"

	"github.com/zipreport/miya/lexer"
)

// CopyStatement copies the statement and its elif branches but shares every
// statement list, so setting a body of the copy leaves n unchanged.
func TestCopyStatement(t *testing.T) {
	tokens, err := lexer.NewLexer("{% if a %}A{% elif b %}B{% else %}C{% endif %}", nil).Tokenize()
	if err != nil {
		t.Fatal(err)
	}
	tmpl, err := NewParser(tokens).Parse()
	if err != nil {
		t.Fatal(err)
	}
	orig := tmpl.Children[0].(*IfNode)
	cp := CopyStatement(orig).(*IfNode)
	if cp == orig || cp.ElseIfs[0] == orig.ElseIfs[0] {
		t.Fatal("CopyStatement must return a new IfNode with new elif branches")
	}
	if &cp.Body[0] != &orig.Body[0] || &cp.ElseIfs[0].Body[0] != &orig.ElseIfs[0].Body[0] || &cp.Else[0] != &orig.Else[0] {
		t.Fatal("CopyStatement must share the statement lists")
	}
	for _, body := range ChildBodies(cp) {
		*body = nil
	}
	if len(orig.Body) != 1 || len(orig.ElseIfs[0].Body) != 1 || len(orig.Else) != 1 {
		t.Fatalf("setting the copy's bodies changed the original: %d %d %d", len(orig.Body), len(orig.ElseIfs[0].Body), len(orig.Else))
	}
	clone := CloneStatement(orig).(*IfNode)
	if &clone.Body[0] == &orig.Body[0] || &clone.ElseIfs[0].Body[0] == &orig.ElseIfs[0].Body[0] || &clone.Else[0] == &orig.Else[0] {
		t.Fatal("CloneStatement must copy the statement lists")
	}
	text := orig.Body[0]
	if CopyStatement(text) != text {
		t.Fatal("a leaf statement is returned as is")
	}
}

// Every statement type with bodies is copied, so a resolver that sets the
// copy's bodies never writes into a parsed template.
func TestCopyStatementCopiesEveryBodyType(t *testing.T) {
	for _, n := range []Node{&TemplateNode{}, &BlockNode{}, &IfNode{}, &ForNode{}, &MacroNode{}, &BlockSetNode{},
		&CallBlockNode{}, &WithNode{}, &AutoescapeNode{}, &FilterBlockNode{}, &ExtensionNode{}, &DiscardNode{}} {
		if len(ChildBodies(n)) == 0 {
			t.Fatalf("%T has no bodies", n)
		}
		if CopyStatement(n) == n {
			t.Errorf("CopyStatement(%T) returned n itself", n)
		}
	}
}
