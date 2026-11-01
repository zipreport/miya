package runtime

import (
	"errors"
	"fmt"

	"github.com/zipreport/miya/parser"
)

// errNeedsContext stops a static hierarchy build at a template whose parent
// depends on the render: its {% extends %} is inside an {% if %}.
var errNeedsContext = errors.New("the parent template depends on the render context")

// holdsExtends reports whether an {% extends %} sits in a branch of n, directly
// or in a nested {% if %}, where Jinja2 runs it when the branch is taken.
func holdsExtends(n *parser.IfNode) bool {
	in := func(body []parser.Node) bool {
		for _, c := range body {
			switch c := c.(type) {
			case *parser.ExtendsNode:
				return true
			case *parser.IfNode:
				if holdsExtends(c) {
					return true
				}
			}
		}
		return false
	}
	if in(n.Body) || in(n.Else) {
		return true
	}
	for _, elif := range n.ElseIfs {
		if in(elif.Body) {
			return true
		}
	}
	return false
}

// hasConditionalExtends reports whether a top-level {% if %} of tmpl holds an
// {% extends %}.
func hasConditionalExtends(tmpl *parser.TemplateNode) bool {
	for _, c := range tmpl.Children {
		if n, ok := c.(*parser.IfNode); ok && holdsExtends(n) {
			return true
		}
	}
	return false
}

// parentExtends returns tmpl as it runs for ctx and the {% extends %} at its top
// level, or nil when it has none. Each top-level {% if %} that holds an
// {% extends %} is replaced by the branch ctx takes, so the {% extends %} Jinja2
// would run sits at the top level; the branches not taken are kept under an
// {% if false %}, as Jinja2 registers every block of a template. A second
// {% extends %} is an error, as in Jinja2.
func parentExtends(tmpl *parser.TemplateNode, ctx Context, eval Evaluator) (*parser.TemplateNode, *parser.ExtendsNode, error) {
	if hasConditionalExtends(tmpl) {
		nodes, err := takeExtendsBranches(tmpl.Children, ctx, eval)
		if err != nil {
			return nil, nil, err
		}
		flat := *tmpl
		flat.Children = nodes
		tmpl = &flat
	}
	var ext *parser.ExtendsNode
	for _, c := range tmpl.Children {
		if e, ok := c.(*parser.ExtendsNode); ok {
			if ext != nil {
				return nil, nil, fmt.Errorf("template %q: extended multiple times", tmpl.Name)
			}
			ext = e
		}
	}
	return tmpl, ext, nil
}

func takeExtendsBranches(nodes []parser.Node, ctx Context, eval Evaluator) ([]parser.Node, error) {
	var out []parser.Node
	for _, n := range nodes {
		ifn, ok := n.(*parser.IfNode)
		if !ok || !holdsExtends(ifn) {
			out = append(out, n)
			continue
		}
		bodies := [][]parser.Node{ifn.Body}
		conds := []parser.ExpressionNode{ifn.Condition}
		for _, elif := range ifn.ElseIfs {
			bodies = append(bodies, elif.Body)
			conds = append(conds, elif.Condition)
		}
		taken := -1
		for i, cond := range conds {
			v, err := eval.EvalNode(cond, ctx)
			if err != nil {
				return nil, err
			}
			if (&DefaultEvaluator{}).isTruthy(v) {
				taken = i
				break
			}
		}
		bodies = append(bodies, ifn.Else)
		if taken < 0 {
			taken = len(bodies) - 1
		}
		branch, err := takeExtendsBranches(bodies[taken], ctx, eval)
		if err != nil {
			return nil, err
		}
		out = append(out, branch...)
		var rest []parser.Node
		for i, b := range bodies {
			if i != taken {
				rest = append(rest, b...)
			}
		}
		if len(rest) > 0 {
			hold := parser.NewIfNode(parser.NewLiteralNode(false, "false", ifn.Line(), ifn.Column()), ifn.Line(), ifn.Column())
			hold.Body = rest
			out = append(out, hold)
		}
	}
	return out, nil
}
