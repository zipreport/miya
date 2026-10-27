package runtime

import (
	"fmt"

	"github.com/zipreport/miya/parser"
)

// SuperResolver handles {{ super() }} call resolution during inheritance processing
type SuperResolver struct {
	hierarchy *InheritanceHierarchy
	context   Context
	blockMap  map[string]*parser.BlockNode

	// Parent block cache for efficient super() resolution
	parentBlockCache map[string]*parser.BlockNode
}

// NewSuperResolver creates a new super() call resolver
func NewSuperResolver(hierarchy *InheritanceHierarchy, context Context, blockMap map[string]*parser.BlockNode) *SuperResolver {
	return &SuperResolver{
		hierarchy:        hierarchy,
		context:          context,
		blockMap:         blockMap,
		parentBlockCache: make(map[string]*parser.BlockNode),
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

	switch n := node.(type) {
	case *parser.SuperNode:
		// Found a direct {{ super() }} call - resolve it
		resolved, err := s.resolveSuperCall(currentBlockName, depth)
		if err != nil {
			return nil, err
		}
		return s.withParentEscaping(resolved, n, enclosing), nil

	case *parser.VariableNode:
		// Check if this is {{ super() }} or {{ super }} wrapped in VariableNode
		if _, isSuperNode := n.Expression.(*parser.SuperNode); isSuperNode {
			// Replace the VariableNode containing SuperNode with resolved content
			resolved, err := s.resolveSuperCall(currentBlockName, depth)
			if err != nil {
				return nil, err
			}
			return s.withParentEscaping(resolved, n, enclosing), nil
		}
		// Not a super call - return as-is
		return node, nil

	case *parser.BlockNode:
		// Process block contents with the block name as context
		block := parser.CloneStatement(n).(*parser.BlockNode)
		for i, child := range block.Body {
			resolved, err := s.resolveSuperCallsRecursive(child, n.Name, depth, enclosing)
			if err != nil {
				return nil, err
			}
			block.Body[i] = resolved
		}
		return block, nil

	default:
		if len(parser.ChildBodies(node)) == 0 {
			return node, nil
		}
		clone := parser.CloneStatement(node)
		if ae, ok := clone.(*parser.AutoescapeNode); ok {
			on := ae.Enabled
			enclosing = &on
		}
		for _, body := range parser.ChildBodies(clone) {
			for i, child := range *body {
				resolved, err := s.resolveSuperCallsRecursive(child, currentBlockName, depth, enclosing)
				if err != nil {
					return nil, err
				}
				(*body)[i] = resolved
			}
		}
		return clone, nil
	}
}

// withParentEscaping wraps resolved parent content in an {% autoescape %}
// node with the render's base setting when a child {% autoescape %} around the
// super() call would otherwise change how the parent block escapes.
func (s *SuperResolver) withParentEscaping(resolved parser.Node, at parser.Node, enclosing *bool) parser.Node {
	if enclosing == nil {
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
	if blockName == "" {
		// In base template with no inheritance, super() should return empty string
		// instead of erroring - this allows templates to be more flexible
		return &parser.TextNode{Content: ""}, nil
	}

	// Create a unique cache key that includes depth to avoid infinite recursion
	cacheKey := fmt.Sprintf("%s_depth_%d", blockName, depth)

	// Check cache first
	if cached, exists := s.parentBlockCache[cacheKey]; exists {
		// Return a deep copy of cached content
		return s.cloneNodeForSuper(cached)
	}

	// Find parent block in inheritance hierarchy - skip 'depth' number of blocks
	parentBlock, err := s.findParentBlockAtDepth(blockName, depth)
	if err != nil {
		return nil, err
	}

	if parentBlock == nil {
		// No parent block found - return empty text node
		return &parser.TextNode{Content: ""}, nil
	}

	// Clone the parent block first
	clonedParent, err := s.cloneNodeForSuper(parentBlock)
	if err != nil {
		return nil, err
	}

	// Process cloned parent block content for nested super() calls
	resolvedParent, err := s.resolveSuperCallsRecursive(clonedParent, blockName, depth+1, nil)
	if err != nil {
		return nil, err
	}

	// Cache the resolved parent block (not the original)
	if resolvedParentBlock, ok := resolvedParent.(*parser.BlockNode); ok {
		s.parentBlockCache[cacheKey] = resolvedParentBlock
	}

	return resolvedParent, nil
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

// cloneNodeForSuper creates a deep copy of a node for super() resolution
func (s *SuperResolver) cloneNodeForSuper(node parser.Node) (parser.Node, error) {
	switch n := node.(type) {
	case *parser.BlockNode:
		// For super() calls, we want the block's body, not the block wrapper
		if len(n.Body) == 0 {
			return &parser.TextNode{Content: ""}, nil
		}

		// If single child, return it directly
		if len(n.Body) == 1 {
			return s.cloneNodeForSuper(n.Body[0])
		}

		// Multiple children - wrap in a container
		clonedChildren := make([]parser.Node, len(n.Body))
		for i, child := range n.Body {
			cloned, err := s.cloneNodeForSuper(child)
			if err != nil {
				return nil, err
			}
			clonedChildren[i] = cloned
		}

		return &parser.TemplateNode{
			Name:     "super_content",
			Children: clonedChildren,
		}, nil

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
