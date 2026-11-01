package parser

import "fmt"

// ChildBodies returns pointers to every statement list directly inside n: the
// body of a block statement, and every branch of an if, including each elif.
// Leaf statements and expressions have none.
func ChildBodies(n Node) []*[]Node {
	switch n := n.(type) {
	case *TemplateNode:
		return []*[]Node{&n.Children}
	case *BlockNode:
		return []*[]Node{&n.Body}
	case *IfNode:
		bodies := []*[]Node{&n.Body}
		for _, elif := range n.ElseIfs {
			bodies = append(bodies, ChildBodies(elif)...)
		}
		return append(bodies, &n.Else)
	case *ForNode:
		return []*[]Node{&n.Body, &n.Else}
	case *MacroNode:
		return []*[]Node{&n.Body}
	case *BlockSetNode:
		return []*[]Node{&n.Body}
	case *CallBlockNode:
		return []*[]Node{&n.Body}
	case *WithNode:
		return []*[]Node{&n.Body}
	case *AutoescapeNode:
		return []*[]Node{&n.Body}
	case *FilterBlockNode:
		return []*[]Node{&n.Body}
	case *ExtensionNode:
		return []*[]Node{&n.Body}
	case *DiscardNode:
		return []*[]Node{&n.Body}
	}
	return nil
}

// CloneStatement returns a copy of n whose statement lists, at every depth, are
// new slices, so the copy's bodies can be changed without changing n.
// Expressions and leaf statements are shared with n.
func CloneStatement(n Node) Node {
	var c Node
	switch n := n.(type) {
	case *TemplateNode:
		cp := *n
		c = &cp
	case *BlockNode:
		cp := *n
		c = &cp
	case *IfNode:
		cp := *n
		cp.ElseIfs = make([]*IfNode, len(n.ElseIfs))
		for i, elif := range n.ElseIfs {
			e := *elif
			cp.ElseIfs[i] = &e
		}
		c = &cp
	case *ForNode:
		cp := *n
		c = &cp
	case *MacroNode:
		cp := *n
		c = &cp
	case *BlockSetNode:
		cp := *n
		c = &cp
	case *CallBlockNode:
		cp := *n
		c = &cp
	case *WithNode:
		cp := *n
		c = &cp
	case *AutoescapeNode:
		cp := *n
		c = &cp
	case *FilterBlockNode:
		cp := *n
		c = &cp
	case *ExtensionNode:
		cp := *n
		c = &cp
	case *DiscardNode:
		cp := *n
		c = &cp
	default:
		return n
	}
	for _, body := range ChildBodies(c) {
		cloned := make([]Node, len(*body))
		for i, child := range *body {
			cloned[i] = CloneStatement(child)
		}
		*body = cloned
	}
	return c
}

// WalkStatements calls fn for each statement in nodes, in source order, and
// descends into a statement's bodies when fn returns true.
func WalkStatements(nodes []Node, fn func(Node) bool) {
	for _, n := range nodes {
		if fn(n) {
			for _, body := range ChildBodies(n) {
				WalkStatements(*body, fn)
			}
		}
	}
}

// CheckDuplicateBlocks returns an error when two {% block %} definitions in
// t share a name, as Jinja2 does.
func CheckDuplicateBlocks(t *TemplateNode) error {
	seen := map[string]bool{}
	var err error
	WalkStatements(t.Children, func(n Node) bool {
		if block, ok := n.(*BlockNode); ok && err == nil {
			if seen[block.Name] {
				err = fmt.Errorf("block '%s' defined twice at line %d", block.Name, block.Line())
			}
			seen[block.Name] = true
		}
		return err == nil
	})
	return err
}
