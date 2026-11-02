package runtime

import (
	"reflect"
	"testing"

	"github.com/zipreport/miya/parser"
)

// holderNode is an expression node type an extension could define.
type holderNode struct{ Super parser.Node }

func (h *holderNode) String() string  { return "holder" }
func (h *holderNode) Line() int       { return 1 }
func (h *holderNode) Column() int     { return 1 }
func (h *holderNode) ExpressionNode() {}

// Issue #38 guard: looking for super() in a statement does not walk an
// extension node's properties, which may refer back to the node, nor node
// types defined outside the parser; it still looks at the tag's arguments.
func TestIssue38ExpressionsCallSuperSkipsExtensionValues(t *testing.T) {
	self := parser.NewExtensionNode("x", "x", 1, 1)
	self.SetProperty("self", self)
	props := parser.NewExtensionNode("x", "x", 1, 1)
	props.SetProperty("s", parser.NewSuperNode(1, 1))
	own := parser.NewExtensionNode("x", "x", 1, 1)
	own.Arguments = []parser.ExpressionNode{&holderNode{Super: parser.NewSuperNode(1, 1)}}
	args := parser.NewExtensionNode("x", "x", 1, 1)
	args.Arguments = []parser.ExpressionNode{parser.NewSuperNode(1, 1)}
	for _, tt := range []struct {
		name string
		node parser.Node
		want bool
	}{
		{"cyclic_property", self, false},
		{"super_in_property", props, false},
		{"super_in_own_node_type", own, false},
		{"super_in_argument", args, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := expressionsCallSuper(reflect.ValueOf(tt.node)); got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}
