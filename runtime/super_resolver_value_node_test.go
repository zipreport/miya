package runtime

import (
	"testing"

	"github.com/zipreport/miya/parser"
)

// valueNode and anyNode are node types whose values cannot be compared, as
// an extension may return from ParseTag: valueNode by its type, anyNode by
// what its field holds.
type valueNode struct{ parts []string }

func (valueNode) String() string { return "value" }
func (valueNode) Line() int      { return 1 }
func (valueNode) Column() int    { return 1 }

type anyNode struct{ v interface{} }

func (anyNode) String() string { return "value" }
func (anyNode) Line() int      { return 1 }
func (anyNode) Column() int    { return 1 }

// The resolver keeps a node it cannot compare as it is.
func TestResolveBodyKeepsUncomparableNode(t *testing.T) {
	for _, n := range []parser.Node{valueNode{parts: []string{"a"}}, anyNode{v: []string{"a"}}} {
		s := NewSuperResolver(&InheritanceHierarchy{}, nil, nil)
		body := []parser.Node{n}
		out, changed, err := s.resolveBody(body, "", 0, nil)
		if err != nil || changed || len(out) != 1 || &out[0] != &body[0] {
			t.Fatalf("%T: got %v, %v, %v; want the body unchanged", n, out, changed, err)
		}
		if c := parser.CloneStatement(n); c.String() != "value" {
			t.Fatalf("%T: CloneStatement returned %v", n, c)
		}
	}
}
