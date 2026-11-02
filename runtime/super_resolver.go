package runtime

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/zipreport/miya/parser"
)

// superCaptureVar is the variable a super() inside an expression reads its
// parent block from. It is not a valid template identifier.
const superCaptureVar = "super()"

// templateFrameVar holds, in a template's top-level scope, that scope: an
// unscoped block is scoped from it, as a Jinja2 block gets only the template
// context.
const templateFrameVar = "template()"

// blockEntryVar holds, in a block's scope, the context the block was entered
// from. It is not a valid template identifier.
const blockEntryVar = "block()"

// deferredSuper is the parent block content bound to superCaptureVar; it is
// rendered each time super() is evaluated.
type deferredSuper struct {
	node parser.Node
}

var (
	nodeSliceType = reflect.TypeOf([]parser.Node(nil))
	parserPkgPath = reflect.TypeOf(parser.SuperNode{}).PkgPath()
	// nodeFields caches, per AST struct type, the exported fields that can
	// hold a node; the walk runs on every statement of every hierarchy build.
	nodeFields sync.Map // reflect.Type -> []int
)

// superCycle stands where super() would render a parent block that is still
// being resolved further out. When evaluated it renders that block's content,
// so a condition around it decides how deep the recursion goes, as in Jinja2.
type superCycle struct {
	block   string
	content *superContent
}

// superContent is a parent block's resolved content, set once its resolution
// finishes.
type superContent struct {
	node parser.Node
}

// superCycleError is the error of a super() cycle that nothing stopped. The
// outer levels of the cycle return it as it is, so a wrapper inside the cycle
// (a macro, a set body) does not repeat once per level.
type superCycleError struct {
	block string
	err   error
}

func (e *superCycleError) Error() string {
	return fmt.Sprintf("super() in block %q renders itself: %v", e.block, e.err)
}

func (e *superCycleError) Unwrap() error { return e.err }

// superMissing stands where super() would render a parent block that does not
// exist. It fails only when evaluated, so a super() in an untaken branch
// renders, as in Jinja2. A bare super there is undefined instead, so
// {% if super is defined %} guards the call.
type superMissing struct {
	block string
	line  int
	bare  bool
}

func (m *superMissing) String() string { return "super() without parent in block " + m.block }

func (m *superMissing) err() error {
	return fmt.Errorf("there is no parent block called '%s' at line %d", m.block, m.line)
}
func (m *superMissing) Line() int   { return m.line }
func (m *superMissing) Column() int { return 0 }

func (c *superCycle) String() string { return "super() cycle in block " + c.block }
func (c *superCycle) Line() int      { return 0 }
func (c *superCycle) Column() int    { return 0 }

// SuperResolver handles {{ super() }} call resolution during inheritance processing
type SuperResolver struct {
	hierarchy *InheritanceHierarchy
	context   Context
	blockMap  map[string]*parser.BlockNode

	// Parent block cache for efficient super() resolution
	parentBlockCache map[string]*parser.BlockNode
	// parent blocks being resolved, to stop a super() cycle through nested
	// blocks
	resolving map[string]*superContent
}

// NewSuperResolver creates a new super() call resolver
func NewSuperResolver(hierarchy *InheritanceHierarchy, context Context, blockMap map[string]*parser.BlockNode) *SuperResolver {
	return &SuperResolver{
		hierarchy:        hierarchy,
		context:          context,
		blockMap:         blockMap,
		parentBlockCache: make(map[string]*parser.BlockNode),
		resolving:        make(map[string]*superContent),
	}
}

// ResolveSuperCalls processes all {{ super() }} calls in the given AST node
func (s *SuperResolver) ResolveSuperCalls(node parser.Node, currentBlockName string) (parser.Node, error) {
	return s.resolveSuperCallsRecursive(node, currentBlockName, 0, nil)
}

// resolveSuperCallsRecursive recursively processes super() calls with depth tracking
func (s *SuperResolver) resolveSuperCallsRecursive(node parser.Node, currentBlockName string, depth int, enclosing *bool) (parser.Node, error) {
	if depth > 10 {
		return nil, fmt.Errorf("super() call depth exceeded (possible infinite recursion)")
	}

	if currentBlockName == "" && isWholeSuper(node) {
		// outside a block super() has no parent block; it fails when evaluated
		return node, nil
	}

	switch n := node.(type) {
	case *parser.SuperNode:
		// Found a direct {{ super() }} call - resolve it
		resolved, err := s.resolveSuperCall(currentBlockName, depth)
		if err != nil {
			return nil, err
		}
		return s.withParentEscaping(atSuper(resolved, n), n, enclosing), nil

	case *parser.VariableNode:
		// Check if this is {{ super() }} or {{ super }} wrapped in VariableNode
		if sn, isSuperNode := n.Expression.(*parser.SuperNode); isSuperNode {
			// Replace the VariableNode containing SuperNode with resolved content
			resolved, err := s.resolveSuperCall(currentBlockName, depth)
			if err != nil {
				return nil, err
			}
			return s.withParentEscaping(atSuper(resolved, sn), n, enclosing), nil
		}
		// Not a super call - return as-is
		return node, nil

	case *parser.BlockNode:
		// A block renders its most-derived definition, also inside parent
		// block content, as Jinja2 does; its super() starts that block's chain
		scoped := n.Scoped
		if derived, ok := s.blockMap[n.Name]; ok {
			n = derived
		}
		block := parser.CopyStatement(n).(*parser.BlockNode)
		block.Scoped = scoped
		body, _, err := s.resolveBody(block.Body, n.Name, 0, enclosing)
		if err != nil {
			return nil, err
		}
		block.Body = body
		return block, nil

	default:
		bodies := parser.ChildBodies(node)
		if len(bodies) == 0 {
			return node, nil
		}
		if ae, ok := node.(*parser.AutoescapeNode); ok {
			on := ae.Enabled
			enclosing = &on
		}
		var resolved [][]parser.Node
		for i, body := range bodies {
			r, changed, err := s.resolveBody(*body, currentBlockName, depth, enclosing)
			if err != nil {
				return nil, err
			}
			if changed && resolved == nil {
				resolved = make([][]parser.Node, len(bodies))
				for j := 0; j < i; j++ {
					resolved[j] = *bodies[j]
				}
			}
			if resolved != nil {
				resolved[i] = r
			}
		}
		if resolved == nil {
			return node, nil
		}
		clone := parser.CopyStatement(node)
		for i, body := range parser.ChildBodies(clone) {
			*body = resolved[i]
		}
		return clone, nil
	}
}

// resolveBody resolves super() calls in a statement list. Before a statement
// whose expressions call super() it binds superCaptureVar to the parent block
// content, which EvalSuperNode renders where super() is evaluated.
// It returns body itself, and false, when no statement changed.
func (s *SuperResolver) resolveBody(body []parser.Node, currentBlockName string, depth int, enclosing *bool) ([]parser.Node, bool, error) {
	var out []parser.Node
	for i, child := range body {
		var bind parser.Node
		// outside a block super() has no parent to bind; it fails when evaluated
		if currentBlockName != "" && !isWholeSuper(child) && expressionsCallSuper(reflect.ValueOf(child)) {
			parent, err := s.resolveSuperCall(currentBlockName, depth)
			if err != nil {
				return nil, false, err
			}
			if m, ok := parent.(*superMissing); ok {
				m.line = child.Line()
			}
			parent = s.withParentEscaping(parent, child, enclosing)
			value := parser.NewLiteralNode(&deferredSuper{node: parent}, "", child.Line(), child.Column())
			bind = parser.NewSetNode(superCaptureVar, value, child.Line(), child.Column())
		}
		resolved, err := s.resolveSuperCallsRecursive(child, currentBlockName, depth, enclosing)
		if err != nil {
			return nil, false, err
		}
		if out == nil && (bind != nil || replaced(resolved, child)) {
			out = make([]parser.Node, i, len(body)+1)
			copy(out, body[:i])
		}
		if out != nil {
			if bind != nil {
				out = append(out, bind)
			}
			out = append(out, resolved)
		}
	}
	if out == nil {
		return body, false, nil
	}
	return out, true, nil
}

// replaced reports whether the resolver returned a different node for child.
// The resolver replaces only pointer nodes; a node of another type (only an
// extension can make one) may not be comparable, so it is not compared.
func replaced(resolved, child parser.Node) bool {
	if t := reflect.TypeOf(child); t == nil || t.Kind() != reflect.Ptr {
		return false
	}
	return resolved != child
}

// atSuper records on a missing parent block where its super sits, and whether
// it is called.
func atSuper(resolved parser.Node, sn *parser.SuperNode) parser.Node {
	if m, ok := resolved.(*superMissing); ok {
		m.line = sn.Line()
		m.bare = sn.Bare
	}
	return resolved
}

// isWholeSuper reports whether n is a {{ super() }} statement, which is
// replaced by the parent block content inline.
func isWholeSuper(n parser.Node) bool {
	switch n := n.(type) {
	case *parser.SuperNode:
		return true
	case *parser.VariableNode:
		_, ok := n.Expression.(*parser.SuperNode)
		return ok
	}
	return false
}

// expressionsCallSuper reports whether a SuperNode is reachable from v
// through exported fields, without descending into statement lists.
func expressionsCallSuper(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Interface, reflect.Ptr:
		if v.IsNil() {
			return false
		}
		switch n := v.Interface().(type) {
		case *parser.SuperNode:
			return true
		case *parser.ExtensionNode:
			// its properties hold the extension's own values, not expressions
			for _, a := range n.Arguments {
				if expressionsCallSuper(reflect.ValueOf(a)) {
					return true
				}
			}
			return false
		case *parser.IdentifierNode, *parser.TextNode, *parser.LiteralNode:
			return false
		}
		return expressionsCallSuper(v.Elem())
	case reflect.Struct:
		// a node type defined outside the parser, such as an extension's,
		// may refer back to itself
		if v.Type().PkgPath() != parserPkgPath {
			return false
		}
		for _, i := range fieldsHoldingNodes(v.Type()) {
			if expressionsCallSuper(v.Field(i)) {
				return true
			}
		}
	case reflect.Slice, reflect.Array:
		if v.Type() == nodeSliceType {
			return false
		}
		for i := 0; i < v.Len(); i++ {
			if expressionsCallSuper(v.Index(i)) {
				return true
			}
		}
	case reflect.Map:
		iter := v.MapRange()
		for iter.Next() {
			if expressionsCallSuper(iter.Value()) {
				return true
			}
		}
	}
	return false
}

// overridesCallSuper reports whether a block that overrides the root's calls
// super() inside an expression. Only those blocks have a parent block to
// render, so the root's own blocks (often most of the template) are skipped.
func overridesCallSuper(h *InheritanceHierarchy) bool {
	rootBlocks := make(map[*parser.BlockNode]bool)
	parser.WalkStatements(h.RootTemplate.Children, func(n parser.Node) bool {
		if b, ok := n.(*parser.BlockNode); ok {
			rootBlocks[b] = true
		}
		return true
	})
	for _, b := range h.BlockMap {
		if !rootBlocks[b] && superInExpressions(b) {
			return true
		}
	}
	return false
}

// superInExpressions reports whether any statement under node calls super()
// inside an expression, which HasSuperCalls does not look for.
func superInExpressions(node parser.Node) bool {
	found := false
	parser.WalkStatements([]parser.Node{node}, func(n parser.Node) bool {
		if !found && !isWholeSuper(n) && expressionsCallSuper(reflect.ValueOf(n)) {
			found = true
		}
		return !found
	})
	return found
}

// fieldsHoldingNodes returns the indexes of t's exported fields whose kind can
// reach a node.
func fieldsHoldingNodes(t reflect.Type) []int {
	if f, ok := nodeFields.Load(t); ok {
		return f.([]int)
	}
	var idx []int
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		switch f.Type.Kind() {
		case reflect.Interface, reflect.Ptr, reflect.Struct, reflect.Slice, reflect.Array, reflect.Map:
			idx = append(idx, i)
		}
	}
	nodeFields.Store(t, idx)
	return idx
}

// withParentEscaping wraps resolved parent content in an {% autoescape %}
// node with the render's base setting when a child {% autoescape %} around the
// super() call would otherwise change how the parent block escapes.
func (s *SuperResolver) withParentEscaping(resolved parser.Node, at parser.Node, enclosing *bool) parser.Node {
	if _, missing := resolved.(*superMissing); enclosing == nil || missing {
		return resolved
	}
	on, _, _ := resolveEscape(s.context)
	if *enclosing == on {
		return resolved
	}
	w := parser.NewAutoescapeNode(on, at.Line(), at.Column())
	w.Body = []parser.Node{resolved}
	return w
}

// resolveSuperCall resolves a single {{ super() }} call to parent block content
func (s *SuperResolver) resolveSuperCall(blockName string, depth int) (parser.Node, error) {
	// Create a unique cache key that includes depth to avoid infinite recursion
	cacheKey := fmt.Sprintf("%s_depth_%d", blockName, depth)

	// Check cache first
	if cached, exists := s.parentBlockCache[cacheKey]; exists {
		// Share the cached content
		return s.cloneNodeForSuper(cached)
	}

	// Find parent block in inheritance hierarchy - skip 'depth' number of blocks
	parentBlock, err := s.findParentBlockAtDepth(blockName, depth)
	if err != nil {
		return nil, err
	}

	if parentBlock == nil {
		return &superMissing{block: blockName}, nil
	}

	if content, ok := s.resolving[cacheKey]; ok {
		return &superCycle{block: blockName, content: content}, nil
	}
	content := &superContent{}
	s.resolving[cacheKey] = content
	defer delete(s.resolving, cacheKey)

	// Process the parent block content for nested super() calls
	resolvedParentBlock := parser.CopyStatement(parentBlock).(*parser.BlockNode)
	body, _, err := s.resolveBody(resolvedParentBlock.Body, blockName, depth+1, nil)
	if err != nil {
		return nil, err
	}
	resolvedParentBlock.Body = body

	// A cycle back to this block renders the same content, so the content
	// does not depend on where it was resolved and is always cached
	s.parentBlockCache[cacheKey] = resolvedParentBlock
	resolved, err := s.cloneNodeForSuper(resolvedParentBlock)
	content.node = resolved
	return resolved, err
}

// findParentBlockAtDepth searches the inheritance hierarchy for the parent block at a specific depth
func (s *SuperResolver) findParentBlockAtDepth(blockName string, depth int) (*parser.BlockNode, error) {
	// Build block hierarchy from templates (child to parent order)
	blockHierarchy := s.buildBlockHierarchy(blockName)

	// Calculate which block to return based on depth
	// depth 0 = child block (we want parent), depth 1 = parent block (we want grandparent), etc.
	targetIndex := depth + 1

	if len(blockHierarchy) <= targetIndex {
		return nil, nil // No parent block at this depth
	}

	// Return the parent block at the target depth
	return blockHierarchy[targetIndex], nil
}

// buildBlockHierarchy builds a hierarchy of blocks with the same name
func (s *SuperResolver) buildBlockHierarchy(blockName string) []*parser.BlockNode {
	var hierarchy []*parser.BlockNode

	// Go through templates from child to parent
	for _, template := range s.hierarchy.Templates {
		block := s.findBlockInTemplate(template, blockName)
		if block != nil {
			hierarchy = append(hierarchy, block)
		}
	}

	return hierarchy
}

// findBlockInTemplate finds a block with the given name in a template
func (s *SuperResolver) findBlockInTemplate(template *parser.TemplateNode, blockName string) *parser.BlockNode {
	return s.findBlockRecursive(template, blockName)
}

// findBlockRecursive recursively searches for a block in AST nodes
func (s *SuperResolver) findBlockRecursive(node parser.Node, blockName string) *parser.BlockNode {
	var found *parser.BlockNode
	parser.WalkStatements([]parser.Node{node}, func(n parser.Node) bool {
		if block, ok := n.(*parser.BlockNode); ok && block.Name == blockName {
			found = block
		}
		return found == nil
	})
	return found
}

// cloneNodeForSuper wraps a resolved parent block's body in a block named
// superCaptureVar. Every super() that renders the block shares the body: the
// merged template is not changed once built.
func (s *SuperResolver) cloneNodeForSuper(node parser.Node) (parser.Node, error) {
	switch n := node.(type) {
	case *parser.BlockNode:
		// For super() calls, we want the block's body in a block named
		// superCaptureVar, which renders it in the parent block's own scope
		if len(n.Body) == 0 {
			return &parser.TextNode{Content: ""}, nil
		}

		parent := parser.NewBlockNode(superCaptureVar, n.Line(), n.Column())
		parent.Body = n.Body
		return parent, nil

	default:
		return parser.CloneStatement(node), nil
	}
}

// ProcessTemplateWithSuperCalls processes an entire template to resolve super() calls
func (s *SuperResolver) ProcessTemplateWithSuperCalls(template *parser.TemplateNode) (*parser.TemplateNode, error) {
	// Process the template with block-aware super resolution
	processed, err := s.resolveTemplateWithBlockContext(template)
	if err != nil {
		return nil, err
	}

	return processed, nil
}

// resolveTemplateWithBlockContext processes template while maintaining block context
func (s *SuperResolver) resolveTemplateWithBlockContext(template *parser.TemplateNode) (*parser.TemplateNode, error) {
	resolved, err := s.resolveSuperCallsRecursive(template, "", 0, nil)
	if err != nil {
		return nil, err
	}
	return resolved.(*parser.TemplateNode), nil
}

// SuperCallDetector checks if a node contains super() calls
type SuperCallDetector struct{}

// NewSuperCallDetector creates a new super() call detector
func NewSuperCallDetector() *SuperCallDetector {
	return &SuperCallDetector{}
}

// HasSuperCalls checks if the given node contains any {{ super() }} calls
func (d *SuperCallDetector) HasSuperCalls(node parser.Node) bool {
	return d.hasSuperCallsRecursive(node)
}

// hasSuperCallsRecursive recursively checks for super() calls
func (d *SuperCallDetector) hasSuperCallsRecursive(node parser.Node) bool {
	found := false
	parser.WalkStatements([]parser.Node{node}, func(n parser.Node) bool {
		switch n := n.(type) {
		case *parser.SuperNode:
			found = true
		case *parser.VariableNode:
			if _, isSuper := n.Expression.(*parser.SuperNode); isSuper {
				found = true
			}
		}
		return !found
	})
	return found
}
