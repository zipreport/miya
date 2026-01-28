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
	return s.resolveSuperCallsRecursive(node, currentBlockName, 0)
}

// resolveSuperCallsRecursive recursively processes super() calls with depth tracking
func (s *SuperResolver) resolveSuperCallsRecursive(node parser.Node, currentBlockName string, depth int) (parser.Node, error) {
	if depth > 10 {
		return nil, fmt.Errorf("super() call depth exceeded (possible infinite recursion)")
	}

	switch n := node.(type) {
	case *parser.SuperNode:
		// Found a direct {{ super() }} call - resolve it
		return s.resolveSuperCall(currentBlockName, depth)

	case *parser.VariableNode:
		// Check if this is {{ super() }} or {{ super }} wrapped in VariableNode
		if _, isSuperNode := n.Expression.(*parser.SuperNode); isSuperNode {
			// Replace the VariableNode containing SuperNode with resolved content
			resolved, err := s.resolveSuperCall(currentBlockName, depth)
			if err != nil {
				return nil, err
			}
			return resolved, nil
		}
		// Not a super call - return as-is
		return node, nil

	case *parser.TemplateNode:
		// Process all children in template
		newChildren := make([]parser.Node, len(n.Children))
		for i, child := range n.Children {
			resolved, err := s.resolveSuperCallsRecursive(child, currentBlockName, depth)
			if err != nil {
				return nil, err
			}
			newChildren[i] = resolved
		}

		return &parser.TemplateNode{
			Name:     n.Name,
			Children: newChildren,
		}, nil

	case *parser.BlockNode:
		// Process block contents with the block name as context
		newBody := make([]parser.Node, len(n.Body))
		for i, child := range n.Body {
			resolved, err := s.resolveSuperCallsRecursive(child, n.Name, depth)
			if err != nil {
				return nil, err
			}
			newBody[i] = resolved
		}

		return &parser.BlockNode{
			Name: n.Name,
			Body: newBody,
		}, nil

	case *parser.IfNode:
		// Process if statement bodies
		newBody := make([]parser.Node, len(n.Body))
		for i, child := range n.Body {
			resolved, err := s.resolveSuperCallsRecursive(child, currentBlockName, depth)
			if err != nil {
				return nil, err
			}
			newBody[i] = resolved
		}

		newElse := make([]parser.Node, len(n.Else))
		for i, child := range n.Else {
			resolved, err := s.resolveSuperCallsRecursive(child, currentBlockName, depth)
			if err != nil {
				return nil, err
			}
			newElse[i] = resolved
		}

		return &parser.IfNode{
			Condition: n.Condition,
			Body:      newBody,
			Else:      newElse,
		}, nil

	case *parser.ForNode:
		// Process for loop bodies
		newBody := make([]parser.Node, len(n.Body))
		for i, child := range n.Body {
			resolved, err := s.resolveSuperCallsRecursive(child, currentBlockName, depth)
			if err != nil {
				return nil, err
			}
			newBody[i] = resolved
		}

		newElse := make([]parser.Node, len(n.Else))
		for i, child := range n.Else {
			resolved, err := s.resolveSuperCallsRecursive(child, currentBlockName, depth)
			if err != nil {
				return nil, err
			}
			newElse[i] = resolved
		}

		return &parser.ForNode{
			Variables: n.Variables,
			Iterable:  n.Iterable,
			Body:      newBody,
			Else:      newElse,
		}, nil

	default:
		// For other node types, return as-is (no super() calls possible)
		return node, nil
	}
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
	resolvedParent, err := s.resolveSuperCallsRecursive(clonedParent, blockName, depth+1)
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

// findParentBlock searches the inheritance hierarchy for the parent block (legacy method)
func (s *SuperResolver) findParentBlock(blockName string) (*parser.BlockNode, error) {
	return s.findParentBlockAtDepth(blockName, 0)
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
	switch n := node.(type) {
	case *parser.TemplateNode:
		for _, child := range n.Children {
			if block := s.findBlockRecursive(child, blockName); block != nil {
				return block
			}
		}
	case *parser.BlockNode:
		if n.Name == blockName {
			return n
		}
		// Also check nested blocks
		for _, child := range n.Body {
			if block := s.findBlockRecursive(child, blockName); block != nil {
				return block
			}
		}
	case *parser.IfNode:
		for _, child := range n.Body {
			if block := s.findBlockRecursive(child, blockName); block != nil {
				return block
			}
		}
		for _, child := range n.Else {
			if block := s.findBlockRecursive(child, blockName); block != nil {
				return block
			}
		}
	case *parser.ForNode:
		for _, child := range n.Body {
			if block := s.findBlockRecursive(child, blockName); block != nil {
				return block
			}
		}
		for _, child := range n.Else {
			if block := s.findBlockRecursive(child, blockName); block != nil {
				return block
			}
		}
	}

	return nil
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

	case *parser.TextNode:
		return &parser.TextNode{Content: n.Content}, nil

	case *parser.VariableNode:
		if _, isSuperNode := n.Expression.(*parser.SuperNode); isSuperNode {
		}
		return &parser.VariableNode{Expression: n.Expression}, nil

	case *parser.IfNode:
		clonedBody := make([]parser.Node, len(n.Body))
		for i, child := range n.Body {
			cloned, err := s.cloneNodeForSuper(child)
			if err != nil {
				return nil, err
			}
			clonedBody[i] = cloned
		}

		clonedElse := make([]parser.Node, len(n.Else))
		for i, child := range n.Else {
			cloned, err := s.cloneNodeForSuper(child)
			if err != nil {
				return nil, err
			}
			clonedElse[i] = cloned
		}

		return &parser.IfNode{
			Condition: n.Condition,
			Body:      clonedBody,
			Else:      clonedElse,
		}, nil

	case *parser.ForNode:
		clonedBody := make([]parser.Node, len(n.Body))
		for i, child := range n.Body {
			cloned, err := s.cloneNodeForSuper(child)
			if err != nil {
				return nil, err
			}
			clonedBody[i] = cloned
		}

		clonedElse := make([]parser.Node, len(n.Else))
		for i, child := range n.Else {
			cloned, err := s.cloneNodeForSuper(child)
			if err != nil {
				return nil, err
			}
			clonedElse[i] = cloned
		}

		return &parser.ForNode{
			Variables: n.Variables,
			Iterable:  n.Iterable,
			Body:      clonedBody,
			Else:      clonedElse,
		}, nil

	default:
		// For unknown types, return as-is
		return node, nil
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
	// Clone the template
	newTemplate := &parser.TemplateNode{
		Name:     template.Name,
		Children: make([]parser.Node, len(template.Children)),
	}

	// Process each child node
	for i, child := range template.Children {
		resolvedChild, err := s.resolveNodeWithBlockContext(child, "")
		if err != nil {
			return nil, err
		}
		newTemplate.Children[i] = resolvedChild
	}

	return newTemplate, nil
}

// resolveNodeWithBlockContext processes a node while tracking block context
func (s *SuperResolver) resolveNodeWithBlockContext(node parser.Node, currentBlockName string) (parser.Node, error) {
	if node == nil {
		return nil, nil
	}

	switch n := node.(type) {
	case *parser.BlockNode:
		// Process block content with the block name as context
		newBody := make([]parser.Node, len(n.Body))
		for i, child := range n.Body {
			resolvedChild, err := s.resolveNodeWithBlockContext(child, n.Name)
			if err != nil {
				return nil, err
			}
			newBody[i] = resolvedChild
		}

		return &parser.BlockNode{
			Name: n.Name,
			Body: newBody,
		}, nil

	case *parser.SuperNode:
		// Resolve super() call with current block context
		return s.resolveSuperCall(currentBlockName, 0)

	case *parser.VariableNode:
		// Check if this is {{ super() }} wrapped in VariableNode
		if _, isSuperNode := n.Expression.(*parser.SuperNode); isSuperNode {
			return s.resolveSuperCall(currentBlockName, 0)
		}
		// For other variable nodes, return as-is
		return node, nil

	default:
		// For other node types, return as-is or process recursively if needed
		return node, nil
	}
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
	switch n := node.(type) {
	case *parser.SuperNode:
		return true

	case *parser.VariableNode:
		// Check if this is {{ super() }} or {{ super }} wrapped in VariableNode
		if _, isSuperNode := n.Expression.(*parser.SuperNode); isSuperNode {
			return true
		}
		return false

	case *parser.TemplateNode:
		for _, child := range n.Children {
			if d.hasSuperCallsRecursive(child) {
				return true
			}
		}

	case *parser.BlockNode:
		for _, child := range n.Body {
			if d.hasSuperCallsRecursive(child) {
				return true
			}
		}

	case *parser.IfNode:
		for _, child := range n.Body {
			if d.hasSuperCallsRecursive(child) {
				return true
			}
		}
		for _, child := range n.Else {
			if d.hasSuperCallsRecursive(child) {
				return true
			}
		}

	case *parser.ForNode:
		for _, child := range n.Body {
			if d.hasSuperCallsRecursive(child) {
				return true
			}
		}
		for _, child := range n.Else {
			if d.hasSuperCallsRecursive(child) {
				return true
			}
		}
	}

	return false
}
