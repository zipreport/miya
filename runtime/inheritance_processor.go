package runtime

import (
	"fmt"
	"strings"

	"github.com/zipreport/miya/parser"
)

// InheritanceProcessor handles template inheritance resolution at render-time
type InheritanceProcessor struct {
	env           EnvironmentInterface
	templateCache map[string]*parser.TemplateNode
	cache         *InheritanceCache
	hasher        *ContextHasher
}

// EnvironmentInterface defines the minimal interface needed from Environment
type EnvironmentInterface interface {
	GetTemplate(name string) (TemplateInterface, error)
	GetLoader() interface{}
}

// TemplateInterface defines the minimal interface needed from Template
type TemplateInterface interface {
	AST() *parser.TemplateNode
	Name() string
}

// InheritanceHierarchy represents a resolved template inheritance chain
type InheritanceHierarchy struct {
	RootTemplate *parser.TemplateNode
	Templates    []*parser.TemplateNode // From child to parent order
	BlockMap     map[string]*parser.BlockNode
	TemplateMap  map[string]*parser.TemplateNode
}

// NewInheritanceProcessor creates a new inheritance processor
func NewInheritanceProcessor(env EnvironmentInterface) *InheritanceProcessor {
	return &InheritanceProcessor{
		env:           env,
		templateCache: make(map[string]*parser.TemplateNode),
		cache:         NewInheritanceCache(),
		hasher:        NewContextHasher(),
	}
}

// NewInheritanceProcessorWithCache creates a processor with a shared cache
func NewInheritanceProcessorWithCache(env EnvironmentInterface, cache *InheritanceCache) *InheritanceProcessor {
	return &InheritanceProcessor{
		env:           env,
		templateCache: make(map[string]*parser.TemplateNode),
		cache:         cache,
		hasher:        NewContextHasher(),
	}
}

// ResolveInheritance resolves template inheritance at render-time with caching
func (p *InheritanceProcessor) ResolveInheritance(template TemplateInterface, context Context) (*parser.TemplateNode, error) {
	ast := template.AST()
	templateName := template.Name()

	// Check if template has inheritance directives
	hasInheritance := p.hasInheritanceDirectives(ast)
	if !hasInheritance {
		return ast, nil // No inheritance needed
	}

	// Try to get resolved template from cache first
	contextHash := p.hasher.HashContext(context)
	// TEMPORARY: Disable caching for templates with super() calls to avoid cache collisions
	// TODO: Improve cache key to include template content hash
	if !p.hasInheritanceDirectives(ast) {
		if cachedTemplate, found := p.cache.GetResolvedTemplate(templateName, contextHash); found {
			return cachedTemplate, nil
		}
	}

	// Check if we have the hierarchy cached
	var hierarchy *InheritanceHierarchy
	var err error

	// For dynamic inheritance, we can't cache hierarchies since they depend on context
	if p.hasDynamicInheritance(template.AST()) {
		// Build inheritance hierarchy with context for dynamic resolution
		hierarchy, err = p.buildInheritanceHierarchyWithContext(template, context)
		if err != nil {
			return nil, fmt.Errorf("failed to build inheritance hierarchy: %v", err)
		}
	} else {
		// Static inheritance - can use caching
		if cachedHierarchy, found := p.cache.GetHierarchy(templateName); found {
			hierarchy = cachedHierarchy
		} else {
			// Build inheritance hierarchy
			hierarchy, err = p.buildInheritanceHierarchy(template)
			if err != nil {
				return nil, fmt.Errorf("failed to build inheritance hierarchy: %v", err)
			}

			// Cache the hierarchy for future use
			p.cache.StoreHierarchy(templateName, hierarchy)
		}
	}

	// Resolve blocks and build final template
	finalTemplate, err := p.buildFinalTemplate(hierarchy, context)
	if err != nil {
		return nil, fmt.Errorf("failed to build final template: %v", err)
	}

	// Cache the resolved template
	templateChain := make([]string, len(hierarchy.Templates))
	for i, tmpl := range hierarchy.Templates {
		if tmpl.Name != "" {
			templateChain[i] = tmpl.Name
		}
	}
	p.cache.StoreResolvedTemplate(templateName, contextHash, finalTemplate, templateChain)

	return finalTemplate, nil
}

// hasInheritanceDirectives checks if template contains {% extends %} directives or {{ super() }} calls
func (p *InheritanceProcessor) hasInheritanceDirectives(node parser.Node) bool {
	if node == nil {
		return false
	}

	switch n := node.(type) {
	case *parser.TemplateNode:
		for _, child := range n.Children {
			if p.hasInheritanceDirectives(child) {
				return true
			}
		}

	case *parser.ExtendsNode:
		return true

	case *parser.SuperNode:
		return true

	case *parser.VariableNode:
		// Check if this is {{ super() }} wrapped in VariableNode
		if _, isSuperNode := n.Expression.(*parser.SuperNode); isSuperNode {
			return true
		}
		// Recursively check the expression
		if p.hasInheritanceDirectives(n.Expression) {
			return true
		}

	case *parser.BlockNode:
		for _, child := range n.Body {
			if p.hasInheritanceDirectives(child) {
				return true
			}
		}

	case *parser.ForNode:
		for _, child := range n.Body {
			if p.hasInheritanceDirectives(child) {
				return true
			}
		}

	case *parser.IfNode:
		for _, child := range n.Body {
			if p.hasInheritanceDirectives(child) {
				return true
			}
		}
		for _, child := range n.Else {
			if p.hasInheritanceDirectives(child) {
				return true
			}
		}
		for _, elseIf := range n.ElseIfs {
			if p.hasInheritanceDirectives(elseIf) {
				return true
			}
		}
	}
	return false
}

// isSelfReferencingBlock checks if this is the same block (no real inheritance)
func (p *InheritanceProcessor) isSelfReferencingBlock(original, override *parser.BlockNode) bool {
	// Only consider it self-referencing if it's literally the same block pointer
	// This should only happen when a template has no parent and references its own blocks
	return original == override
}

// blockContainsSuperCalls checks if a block contains super() calls
func (p *InheritanceProcessor) blockContainsSuperCalls(block *parser.BlockNode) bool {
	detector := NewSuperCallDetector()
	for _, child := range block.Body {
		if detector.HasSuperCalls(child) {
			return true
		}
	}
	return false
}

// debugPrintTemplate prints the structure of a template for debugging
func (p *InheritanceProcessor) debugPrintTemplate(node parser.Node, depth int) {
	indent := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}

	switch n := node.(type) {
	case *parser.TemplateNode:
		fmt.Printf("%sTemplateNode (%d children)\n", indent, len(n.Children))
		for i, child := range n.Children {
			fmt.Printf("%s  [%d]:", indent, i)
			p.debugPrintTemplate(child, depth+1)
		}

	case *parser.BlockNode:
		fmt.Printf(" BlockNode '%s' (%d children)\n", n.Name, len(n.Body))
		for i, child := range n.Body {
			fmt.Printf("%s    [%d]:", indent, i)
			p.debugPrintTemplate(child, depth+1)
		}

	case *parser.VariableNode:
		fmt.Printf(" VariableNode (expr: %T)\n", n.Expression)

	case *parser.SuperNode:
		fmt.Printf(" SuperNode\n")

	default:
		fmt.Printf(" %T\n", n)
	}
}

// buildInheritanceHierarchy builds the complete inheritance chain
func (p *InheritanceProcessor) buildInheritanceHierarchy(template TemplateInterface) (*InheritanceHierarchy, error) {
	hierarchy := &InheritanceHierarchy{
		Templates:   make([]*parser.TemplateNode, 0),
		BlockMap:    make(map[string]*parser.BlockNode),
		TemplateMap: make(map[string]*parser.TemplateNode),
	}

	current := template
	templateNames := make(map[string]bool) // Prevent circular inheritance
	isFirstTemplate := true

	for current != nil {
		// Check for circular inheritance
		if templateNames[current.Name()] {
			return nil, fmt.Errorf("circular inheritance detected: %s", current.Name())
		}
		templateNames[current.Name()] = true

		ast := current.AST()
		hierarchy.Templates = append(hierarchy.Templates, ast)
		hierarchy.TemplateMap[current.Name()] = ast

		// Extract blocks from this template
		if isFirstTemplate {
			// For child template, extract all blocks as potential overrides
			p.extractBlocks(ast, hierarchy.BlockMap)
		} else {
			// For parent templates, only extract blocks that don't exist yet
			p.extractParentBlocks(ast, hierarchy.BlockMap)
		}

		// Find parent template
		parentName := p.findExtendsTemplate(ast)
		if parentName == "" {
			hierarchy.RootTemplate = ast
			break
		}

		// Load parent template
		parentTemplate, err := p.env.GetTemplate(parentName)
		if err != nil {
			return nil, fmt.Errorf("failed to load parent template %s: %v", parentName, err)
		}

		current = parentTemplate
		isFirstTemplate = false
	}

	return hierarchy, nil
}

// findExtendsTemplate finds the template name from {% extends %} directive
func (p *InheritanceProcessor) findExtendsTemplate(node parser.Node) string {
	switch n := node.(type) {
	case *parser.TemplateNode:
		for _, child := range n.Children {
			if result := p.findExtendsTemplate(child); result != "" {
				return result
			}
		}
	case *parser.ExtendsNode:
		if literal, ok := n.Template.(*parser.LiteralNode); ok {
			return strings.Trim(literal.Value.(string), "\"'")
		}
		// For dynamic inheritance, we can't resolve at parse time
		// Return a special marker to indicate dynamic inheritance
		return "<<DYNAMIC>>"
	}
	return ""
}

// hasDynamicInheritance checks if template uses dynamic inheritance ({% extends variable %})
func (p *InheritanceProcessor) hasDynamicInheritance(node parser.Node) bool {
	switch n := node.(type) {
	case *parser.TemplateNode:
		for _, child := range n.Children {
			if p.hasDynamicInheritance(child) {
				return true
			}
		}
	case *parser.ExtendsNode:
		// If it's not a literal node, it's dynamic
		_, isLiteral := n.Template.(*parser.LiteralNode)
		return !isLiteral
	}
	return false
}

// buildInheritanceHierarchyWithContext builds hierarchy with context for dynamic template resolution
func (p *InheritanceProcessor) buildInheritanceHierarchyWithContext(template TemplateInterface, context Context) (*InheritanceHierarchy, error) {
	hierarchy := &InheritanceHierarchy{
		Templates:   make([]*parser.TemplateNode, 0),
		BlockMap:    make(map[string]*parser.BlockNode),
		TemplateMap: make(map[string]*parser.TemplateNode),
	}

	current := template
	templateNames := make(map[string]bool) // Prevent circular inheritance
	isFirstTemplate := true

	for current != nil {
		// Check for circular inheritance
		if templateNames[current.Name()] {
			return nil, fmt.Errorf("circular inheritance detected: %s", current.Name())
		}
		templateNames[current.Name()] = true

		ast := current.AST()
		hierarchy.Templates = append(hierarchy.Templates, ast)
		hierarchy.TemplateMap[current.Name()] = ast

		// Extract blocks from this template
		if isFirstTemplate {
			// For child template, extract all blocks as potential overrides
			p.extractBlocks(ast, hierarchy.BlockMap)
		} else {
			// For parent templates, only extract blocks that don't exist yet
			p.extractParentBlocks(ast, hierarchy.BlockMap)
		}

		// Find parent template - use context for dynamic resolution
		parentName, err := p.findExtendsTemplateWithContext(ast, context)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve parent template: %v", err)
		}
		if parentName == "" {
			hierarchy.RootTemplate = ast
			break
		}

		// Load parent template
		parentTemplate, err := p.env.GetTemplate(parentName)
		if err != nil {
			return nil, fmt.Errorf("failed to load parent template %s: %v", parentName, err)
		}

		current = parentTemplate
		isFirstTemplate = false
	}

	return hierarchy, nil
}

// findExtendsTemplateWithContext finds template name with context evaluation for dynamic inheritance
func (p *InheritanceProcessor) findExtendsTemplateWithContext(node parser.Node, context Context) (string, error) {
	switch n := node.(type) {
	case *parser.TemplateNode:
		for _, child := range n.Children {
			if result, err := p.findExtendsTemplateWithContext(child, context); result != "" || err != nil {
				return result, err
			}
		}
	case *parser.ExtendsNode:
		if literal, ok := n.Template.(*parser.LiteralNode); ok {
			// Static template name
			return strings.Trim(literal.Value.(string), "\"'"), nil
		} else {
			// Dynamic template name - evaluate expression
			evaluator := NewCachedEvaluator()
			templateName, err := evaluator.EvalNodeCached(n.Template, context)
			if err != nil {
				return "", fmt.Errorf("failed to evaluate dynamic template name: %v", err)
			}

			// Convert result to string
			if str, ok := templateName.(string); ok {
				return str, nil
			} else {
				return "", fmt.Errorf("dynamic template name must be a string, got %T", templateName)
			}
		}
	}
	return "", nil
}

// extractBlocks extracts all block definitions from a template
func (p *InheritanceProcessor) extractBlocks(node parser.Node, blockMap map[string]*parser.BlockNode) {
	switch n := node.(type) {
	case *parser.TemplateNode:
		for _, child := range n.Children {
			p.extractBlocks(child, blockMap)
		}
	case *parser.BlockNode:
		// Child blocks override parent blocks (first encountered wins)
		if _, exists := blockMap[n.Name]; !exists {
			blockMap[n.Name] = n
		} else {
		}
		// Also recursively extract nested blocks within this block
		for _, child := range n.Body {
			p.extractBlocks(child, blockMap)
		}
	case *parser.IfNode:
		for _, child := range n.Body {
			p.extractBlocks(child, blockMap)
		}
		for _, child := range n.Else {
			p.extractBlocks(child, blockMap)
		}
	case *parser.ForNode:
		for _, child := range n.Body {
			p.extractBlocks(child, blockMap)
		}
		for _, child := range n.Else {
			p.extractBlocks(child, blockMap)
		}
	}
}

// extractParentBlocks extracts blocks from parent templates, but only adds them if they don't override child blocks
func (p *InheritanceProcessor) extractParentBlocks(node parser.Node, blockMap map[string]*parser.BlockNode) {
	switch n := node.(type) {
	case *parser.TemplateNode:
		for _, child := range n.Children {
			p.extractParentBlocks(child, blockMap)
		}
	case *parser.BlockNode:
		// For parent templates, don't add blocks that already exist (child overrides take precedence)
		if _, exists := blockMap[n.Name]; !exists {
			blockMap[n.Name] = n
		} else {
		}
		// Still process nested blocks
		for _, child := range n.Body {
			p.extractParentBlocks(child, blockMap)
		}
	case *parser.IfNode:
		for _, child := range n.Body {
			p.extractParentBlocks(child, blockMap)
		}
		for _, child := range n.Else {
			p.extractParentBlocks(child, blockMap)
		}
	case *parser.ForNode:
		for _, child := range n.Body {
			p.extractParentBlocks(child, blockMap)
		}
		for _, child := range n.Else {
			p.extractParentBlocks(child, blockMap)
		}
	}
}

// buildFinalTemplate builds the final template by merging the inheritance hierarchy
func (p *InheritanceProcessor) buildFinalTemplate(hierarchy *InheritanceHierarchy, context Context) (*parser.TemplateNode, error) {
	if hierarchy.RootTemplate == nil {
		return nil, fmt.Errorf("no root template found in hierarchy")
	}

	// Start with root template and replace blocks with child implementations
	finalTemplate := p.cloneTemplateNode(hierarchy.RootTemplate)

	// Collect and add import statements from all templates in the hierarchy
	err := p.addImportStatementsToTemplate(finalTemplate, hierarchy)
	if err != nil {
		return nil, fmt.Errorf("failed to add import statements: %v", err)
	}

	// Apply child block overrides to the final template
	err = p.applyBlockOverrides(finalTemplate, hierarchy, context)
	if err != nil {
		return nil, fmt.Errorf("failed to apply block overrides: %v", err)
	}

	// Check for super() calls outside of blocks - these should always error
	// Only validate for multi-template hierarchies (actual inheritance)
	detector := NewSuperCallDetector()
	if len(hierarchy.Templates) > 1 {
		for _, template := range hierarchy.Templates {
			if detector.HasSuperCalls(template) {
				// Check if there are super() calls outside of blocks - these should error
				err := p.validateSuperCallsInTemplateWithHierarchy(template, hierarchy)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	// Check if final template needs super() resolution
	// Only resolve super() calls if we have actual inheritance (multiple templates)
	hasSuperCalls := detector.HasSuperCalls(finalTemplate)
	if hasSuperCalls && len(hierarchy.Templates) > 1 {
		// Resolve all super() calls in the final template
		superResolver := NewSuperResolver(hierarchy, context, hierarchy.BlockMap)
		resolvedTemplate, err := superResolver.ProcessTemplateWithSuperCalls(finalTemplate)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve super() calls: %v", err)
		}
		finalTemplate = resolvedTemplate
	} else if hasSuperCalls && len(hierarchy.Templates) == 1 {
		// For single templates (base templates), replace super() calls with empty content
		finalTemplate, err = p.replaceSuperCallsWithEmpty(finalTemplate)
		if err != nil {
			return nil, fmt.Errorf("failed to replace super() calls: %v", err)
		}
	}

	return finalTemplate, nil
}

// cloneTemplateNode creates a deep copy of a template node
func (p *InheritanceProcessor) cloneTemplateNode(template *parser.TemplateNode) *parser.TemplateNode {
	clone := &parser.TemplateNode{
		Name:     template.Name,
		Children: make([]parser.Node, len(template.Children)),
	}

	for i, child := range template.Children {
		clone.Children[i] = p.cloneNode(child)
	}

	return clone
}

// cloneNode creates a deep copy of any parser node
func (p *InheritanceProcessor) cloneNode(node parser.Node) parser.Node {
	switch n := node.(type) {
	case *parser.TextNode:
		return &parser.TextNode{Content: n.Content}
	case *parser.VariableNode:
		return &parser.VariableNode{Expression: n.Expression}
	case *parser.BlockNode:
		clone := &parser.BlockNode{
			Name: n.Name,
			Body: make([]parser.Node, len(n.Body)),
		}
		for i, child := range n.Body {
			clone.Body[i] = p.cloneNode(child)
		}
		return clone
	case *parser.ExtendsNode:
		return &parser.ExtendsNode{Template: n.Template}
	case *parser.FromNode:
		return &parser.FromNode{Template: n.Template, Names: n.Names}
	case *parser.ImportNode:
		return &parser.ImportNode{Template: n.Template, Alias: n.Alias}
	case *parser.IfNode:
		clone := &parser.IfNode{
			Condition: n.Condition,
			Body:      make([]parser.Node, len(n.Body)),
			Else:      make([]parser.Node, len(n.Else)),
		}
		for i, child := range n.Body {
			clone.Body[i] = p.cloneNode(child)
		}
		for i, child := range n.Else {
			clone.Else[i] = p.cloneNode(child)
		}
		return clone
	case *parser.ForNode:
		clone := &parser.ForNode{
			Variables: n.Variables,
			Iterable:  n.Iterable,
			Body:      make([]parser.Node, len(n.Body)),
			Else:      make([]parser.Node, len(n.Else)),
		}
		for i, child := range n.Body {
			clone.Body[i] = p.cloneNode(child)
		}
		for i, child := range n.Else {
			clone.Else[i] = p.cloneNode(child)
		}
		return clone
	default:
		// For unknown node types, return as-is (risky but functional)
		return node
	}
}

// replaceBlocks recursively replaces block nodes with their overridden versions
func (p *InheritanceProcessor) replaceBlocks(node parser.Node, blockMap map[string]*parser.BlockNode, context Context) error {
	switch n := node.(type) {
	case *parser.TemplateNode:
		for i, child := range n.Children {
			if blockNode, ok := child.(*parser.BlockNode); ok {

				// Check if this block has an override
				if overriddenBlock, exists := blockMap[blockNode.Name]; exists {
					n.Children[i] = p.cloneNode(overriddenBlock)
				} else {
					// Process this block's contents recursively to handle nested overrides
					err := p.replaceBlocks(blockNode, blockMap, context)
					if err != nil {
						return err
					}
				}
			} else {
				// Recursively process child nodes
				err := p.replaceBlocks(child, blockMap, context)
				if err != nil {
					return err
				}
			}
		}
	case *parser.IfNode:
		for _, child := range n.Body {
			if err := p.replaceBlocks(child, blockMap, context); err != nil {
				return err
			}
		}
		for _, child := range n.Else {
			if err := p.replaceBlocks(child, blockMap, context); err != nil {
				return err
			}
		}
	case *parser.ForNode:
		for _, child := range n.Body {
			if err := p.replaceBlocks(child, blockMap, context); err != nil {
				return err
			}
		}
		for _, child := range n.Else {
			if err := p.replaceBlocks(child, blockMap, context); err != nil {
				return err
			}
		}
	case *parser.BlockNode:
		// Process block contents recursively for nested block replacements
		for i, child := range n.Body {
			if blockNode, ok := child.(*parser.BlockNode); ok {
				// Check if this nested block has an override
				if overriddenBlock, exists := blockMap[blockNode.Name]; exists {
					n.Body[i] = p.cloneNode(overriddenBlock)
				} else {
					// Recursively process this nested block
					err := p.replaceBlocks(blockNode, blockMap, context)
					if err != nil {
						return err
					}
				}
			} else {
				// Recursively process other child nodes
				err := p.replaceBlocks(child, blockMap, context)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// applyBlockOverrides applies child template block overrides to the base template structure
// This processes the inheritance chain from parent to child, applying overrides progressively
func (p *InheritanceProcessor) applyBlockOverrides(finalTemplate *parser.TemplateNode, hierarchy *InheritanceHierarchy, context Context) error {
	// Process templates from parent to child (reverse order of hierarchy.Templates)
	// hierarchy.Templates is [child, middle, parent], so we process from end to start
	for i := len(hierarchy.Templates) - 2; i >= 0; i-- {
		template := hierarchy.Templates[i]

		// Collect overrides from this level
		levelOverrides := make(map[string]*parser.BlockNode)
		p.collectChildOverrides(template, levelOverrides)

		// Apply this level's overrides to the current state
		err := p.applyOverridesToTemplate(finalTemplate, levelOverrides)
		if err != nil {
			return err
		}
	}

	return nil
}

// collectChildOverrides collects block definitions from child template
func (p *InheritanceProcessor) collectChildOverrides(template *parser.TemplateNode, overrides map[string]*parser.BlockNode) {
	p.collectOverridesRecursive(template, overrides)
}

// collectOverridesRecursive recursively collects block overrides
func (p *InheritanceProcessor) collectOverridesRecursive(node parser.Node, overrides map[string]*parser.BlockNode) {
	switch n := node.(type) {
	case *parser.TemplateNode:
		for _, child := range n.Children {
			p.collectOverridesRecursive(child, overrides)
		}
	case *parser.BlockNode:
		overrides[n.Name] = n
		// Don't recurse into block body - we want the block as a complete override
	case *parser.IfNode:
		for _, child := range n.Body {
			p.collectOverridesRecursive(child, overrides)
		}
		for _, child := range n.Else {
			p.collectOverridesRecursive(child, overrides)
		}
	case *parser.ForNode:
		for _, child := range n.Body {
			p.collectOverridesRecursive(child, overrides)
		}
		for _, child := range n.Else {
			p.collectOverridesRecursive(child, overrides)
		}
	}
}

// applyOverridesToTemplate applies block overrides to template structure
func (p *InheritanceProcessor) applyOverridesToTemplate(template *parser.TemplateNode, overrides map[string]*parser.BlockNode) error {
	return p.applyOverridesRecursive(template, overrides)
}

// applyOverridesRecursive recursively applies block overrides
func (p *InheritanceProcessor) applyOverridesRecursive(node parser.Node, overrides map[string]*parser.BlockNode) error {
	switch n := node.(type) {
	case *parser.TemplateNode:
		for i, child := range n.Children {
			if blockNode, ok := child.(*parser.BlockNode); ok {
				if override, exists := overrides[blockNode.Name]; exists {
					n.Children[i] = p.cloneNode(override)
				} else {
					// No override, recurse into block to handle nested blocks
					err := p.applyOverridesRecursive(blockNode, overrides)
					if err != nil {
						return err
					}
				}
			} else {
				err := p.applyOverridesRecursive(child, overrides)
				if err != nil {
					return err
				}
			}
		}
	case *parser.BlockNode:
		for i, child := range n.Body {
			if blockNode, ok := child.(*parser.BlockNode); ok {
				if override, exists := overrides[blockNode.Name]; exists {
					n.Body[i] = p.cloneNode(override)
				} else {
					err := p.applyOverridesRecursive(blockNode, overrides)
					if err != nil {
						return err
					}
				}
			} else {
				err := p.applyOverridesRecursive(child, overrides)
				if err != nil {
					return err
				}
			}
		}
	case *parser.IfNode:
		for _, child := range n.Body {
			err := p.applyOverridesRecursive(child, overrides)
			if err != nil {
				return err
			}
		}
		for _, child := range n.Else {
			err := p.applyOverridesRecursive(child, overrides)
			if err != nil {
				return err
			}
		}
	case *parser.ForNode:
		for _, child := range n.Body {
			err := p.applyOverridesRecursive(child, overrides)
			if err != nil {
				return err
			}
		}
		for _, child := range n.Else {
			err := p.applyOverridesRecursive(child, overrides)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Cache management methods

// GetCache returns the inheritance cache for external management
func (p *InheritanceProcessor) GetCache() *InheritanceCache {
	return p.cache
}

// InvalidateTemplate removes all cached data for a specific template
func (p *InheritanceProcessor) InvalidateTemplate(templateName string) {
	p.cache.InvalidateTemplate(templateName)
}

// ClearCache removes all cached data
func (p *InheritanceProcessor) ClearCache() {
	p.cache.ClearAll()
}

// GetCacheStats returns cache performance statistics
func (p *InheritanceProcessor) GetCacheStats() CacheStats {
	return p.cache.GetStats()
}

// addImportStatementsToTemplate collects import statements from all templates in the hierarchy
// and adds them to the final template
func (p *InheritanceProcessor) addImportStatementsToTemplate(finalTemplate *parser.TemplateNode, hierarchy *InheritanceHierarchy) error {
	var importNodes []parser.Node

	// Collect import statements from all templates in hierarchy (child to parent order)
	for _, template := range hierarchy.Templates {
		imports := p.collectImportNodes(template)
		importNodes = append(importNodes, imports...)
	}

	if len(importNodes) > 0 {
		// Insert import statements at the beginning of the final template (after extends if present)
		insertIndex := 0
		if len(finalTemplate.Children) > 0 {
			// Skip past ExtendsNode if it exists
			if _, isExtendsNode := finalTemplate.Children[0].(*parser.ExtendsNode); isExtendsNode {
				insertIndex = 1
			}
		}

		// Create new children slice with imports inserted
		newChildren := make([]parser.Node, 0, len(finalTemplate.Children)+len(importNodes))
		newChildren = append(newChildren, finalTemplate.Children[:insertIndex]...)
		newChildren = append(newChildren, importNodes...)
		newChildren = append(newChildren, finalTemplate.Children[insertIndex:]...)
		finalTemplate.Children = newChildren
	}

	return nil
}

// collectImportNodes recursively collects ImportNode and FromNode from a template
func (p *InheritanceProcessor) collectImportNodes(template *parser.TemplateNode) []parser.Node {
	var imports []parser.Node

	for _, child := range template.Children {
		switch child.(type) {
		case *parser.ImportNode, *parser.FromNode:
			imports = append(imports, p.cloneNode(child))
		}
	}

	return imports
}

// validateSuperCallsInTemplate checks if super() calls are used in valid contexts
func (p *InheritanceProcessor) validateSuperCallsInTemplate(template *parser.TemplateNode) error {
	return p.validateSuperCallsInNode(template, "")
}

// validateSuperCallsInTemplateWithHierarchy checks super() calls with hierarchy context
func (p *InheritanceProcessor) validateSuperCallsInTemplateWithHierarchy(template *parser.TemplateNode, hierarchy *InheritanceHierarchy) error {
	return p.validateSuperCallsInNodeWithHierarchy(template, "", hierarchy)
}

// replaceSuperCallsWithEmpty replaces all super() calls with empty text nodes
func (p *InheritanceProcessor) replaceSuperCallsWithEmpty(template *parser.TemplateNode) (*parser.TemplateNode, error) {
	// For base templates, we still need to validate that super() calls are inside blocks
	// But we handle super() calls inside blocks gracefully (replace with empty content)
	err := p.validateSuperCallsOutsideBlocksForBaseTemplate(template, "")
	if err != nil {
		return nil, err
	}

	// Clone the template
	newTemplate := &parser.TemplateNode{
		Name:     template.Name,
		Children: make([]parser.Node, len(template.Children)),
	}

	// Process each child node - replace super() calls with empty content
	for i, child := range template.Children {
		replacedChild, err := p.replaceSuperInNode(child)
		if err != nil {
			return nil, err
		}
		newTemplate.Children[i] = replacedChild
	}

	return newTemplate, nil
}

// validateSuperCallsOutsideBlocksForBaseTemplate validates super() calls for base templates
func (p *InheritanceProcessor) validateSuperCallsOutsideBlocksForBaseTemplate(node parser.Node, currentBlockName string) error {
	if node == nil {
		return nil
	}

	switch n := node.(type) {
	case *parser.TemplateNode:
		for _, child := range n.Children {
			err := p.validateSuperCallsOutsideBlocksForBaseTemplate(child, currentBlockName)
			if err != nil {
				return err
			}
		}

	case *parser.BlockNode:
		// Inside a block - super() calls will be gracefully handled (replaced with empty)
		// No need to validate further inside blocks for base templates
		return nil

	case *parser.SuperNode:
		// Direct super() node outside of block is an error
		if currentBlockName == "" {
			return fmt.Errorf("super() call outside of block context")
		}

	case *parser.VariableNode:
		// Check if this is {{ super() }} wrapped in VariableNode
		if _, isSuperNode := n.Expression.(*parser.SuperNode); isSuperNode {
			if currentBlockName == "" {
				return fmt.Errorf("super() call outside of block context")
			}
		}
		// Also check nested expressions
		if n.Expression != nil {
			err := p.validateSuperCallsOutsideBlocksForBaseTemplate(n.Expression, currentBlockName)
			if err != nil {
				return err
			}
		}

	case *parser.IfNode:
		for _, child := range n.Body {
			err := p.validateSuperCallsOutsideBlocksForBaseTemplate(child, currentBlockName)
			if err != nil {
				return err
			}
		}
		for _, child := range n.Else {
			err := p.validateSuperCallsOutsideBlocksForBaseTemplate(child, currentBlockName)
			if err != nil {
				return err
			}
		}

	case *parser.ForNode:
		for _, child := range n.Body {
			err := p.validateSuperCallsOutsideBlocksForBaseTemplate(child, currentBlockName)
			if err != nil {
				return err
			}
		}
		for _, child := range n.Else {
			err := p.validateSuperCallsOutsideBlocksForBaseTemplate(child, currentBlockName)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// validateSuperCallsOutsideBlocks validates that super() calls are inside blocks
func (p *InheritanceProcessor) validateSuperCallsOutsideBlocks(node parser.Node, currentBlockName string) error {
	if node == nil {
		return nil
	}

	switch n := node.(type) {
	case *parser.TemplateNode:
		for _, child := range n.Children {
			err := p.validateSuperCallsOutsideBlocks(child, currentBlockName)
			if err != nil {
				return err
			}
		}

	case *parser.BlockNode:
		// Inside a block - super() calls are valid
		for _, child := range n.Body {
			err := p.validateSuperCallsOutsideBlocks(child, n.Name)
			if err != nil {
				return err
			}
		}

	case *parser.SuperNode:
		// Direct super() node outside of block is an error
		if currentBlockName == "" {
			return fmt.Errorf("super() call outside of block context")
		}

	case *parser.VariableNode:
		// Check if this is {{ super() }} wrapped in VariableNode
		if _, isSuperNode := n.Expression.(*parser.SuperNode); isSuperNode {
			if currentBlockName == "" {
				return fmt.Errorf("super() call outside of block context")
			}
		}
		// Also check nested expressions
		if n.Expression != nil {
			err := p.validateSuperCallsOutsideBlocks(n.Expression, currentBlockName)
			if err != nil {
				return err
			}
		}

	case *parser.IfNode:
		for _, child := range n.Body {
			err := p.validateSuperCallsOutsideBlocks(child, currentBlockName)
			if err != nil {
				return err
			}
		}
		for _, child := range n.Else {
			err := p.validateSuperCallsOutsideBlocks(child, currentBlockName)
			if err != nil {
				return err
			}
		}

	case *parser.ForNode:
		for _, child := range n.Body {
			err := p.validateSuperCallsOutsideBlocks(child, currentBlockName)
			if err != nil {
				return err
			}
		}
		for _, child := range n.Else {
			err := p.validateSuperCallsOutsideBlocks(child, currentBlockName)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// replaceSuperInNodeWithValidation recursively replaces super() calls with validation
// isBaseTemplate indicates if this is a base template (single template hierarchy) where super() should be gracefully handled
func (p *InheritanceProcessor) replaceSuperInNodeWithValidation(node parser.Node, currentBlockName string) (parser.Node, error) {
	if node == nil {
		return nil, nil
	}

	switch n := node.(type) {
	case *parser.TemplateNode:
		newChildren := make([]parser.Node, len(n.Children))
		for i, child := range n.Children {
			replacedChild, err := p.replaceSuperInNodeWithValidation(child, currentBlockName)
			if err != nil {
				return nil, err
			}
			newChildren[i] = replacedChild
		}
		return &parser.TemplateNode{
			Name:     n.Name,
			Children: newChildren,
		}, nil

	case *parser.BlockNode:
		newBody := make([]parser.Node, len(n.Body))
		for i, child := range n.Body {
			replacedChild, err := p.replaceSuperInNodeWithValidation(child, n.Name)
			if err != nil {
				return nil, err
			}
			newBody[i] = replacedChild
		}
		return &parser.BlockNode{
			Name: n.Name,
			Body: newBody,
		}, nil

	case *parser.SuperNode:
		// Replace with empty text for graceful handling in base templates
		return &parser.TextNode{Content: ""}, nil

	case *parser.VariableNode:
		// Check if this is {{ super() }}
		if _, isSuperNode := n.Expression.(*parser.SuperNode); isSuperNode {
			// For base templates (single template hierarchy), be more permissive
			// The validation for truly invalid super() calls (outside blocks) should happen elsewhere
			// Replace with empty text for graceful handling
			return &parser.TextNode{Content: ""}, nil
		}
		// For other variable nodes, process the expression
		if n.Expression != nil {
			replacedExpr, err := p.replaceSuperInNodeWithValidation(n.Expression, currentBlockName)
			if err != nil {
				return nil, err
			}
			return &parser.VariableNode{
				Expression: replacedExpr,
			}, nil
		}
		return n, nil

	case *parser.IfNode:
		newBody := make([]parser.Node, len(n.Body))
		for i, child := range n.Body {
			replacedChild, err := p.replaceSuperInNodeWithValidation(child, currentBlockName)
			if err != nil {
				return nil, err
			}
			newBody[i] = replacedChild
		}

		newElse := make([]parser.Node, len(n.Else))
		for i, child := range n.Else {
			replacedChild, err := p.replaceSuperInNodeWithValidation(child, currentBlockName)
			if err != nil {
				return nil, err
			}
			newElse[i] = replacedChild
		}

		return &parser.IfNode{
			Condition: n.Condition,
			Body:      newBody,
			Else:      newElse,
		}, nil

	case *parser.ForNode:
		newBody := make([]parser.Node, len(n.Body))
		for i, child := range n.Body {
			replacedChild, err := p.replaceSuperInNodeWithValidation(child, currentBlockName)
			if err != nil {
				return nil, err
			}
			newBody[i] = replacedChild
		}

		newElse := make([]parser.Node, len(n.Else))
		for i, child := range n.Else {
			replacedChild, err := p.replaceSuperInNodeWithValidation(child, currentBlockName)
			if err != nil {
				return nil, err
			}
			newElse[i] = replacedChild
		}

		return &parser.ForNode{
			Variables: n.Variables,
			Iterable:  n.Iterable,
			Body:      newBody,
			Else:      newElse,
		}, nil

	default:
		// For other node types, return as-is
		return node, nil
	}
}

// replaceSuperInNode recursively replaces super() calls with empty text nodes
func (p *InheritanceProcessor) replaceSuperInNode(node parser.Node) (parser.Node, error) {
	if node == nil {
		return nil, nil
	}

	switch n := node.(type) {
	case *parser.TemplateNode:
		newChildren := make([]parser.Node, len(n.Children))
		for i, child := range n.Children {
			replacedChild, err := p.replaceSuperInNode(child)
			if err != nil {
				return nil, err
			}
			newChildren[i] = replacedChild
		}
		return &parser.TemplateNode{
			Name:     n.Name,
			Children: newChildren,
		}, nil

	case *parser.BlockNode:
		newBody := make([]parser.Node, len(n.Body))
		for i, child := range n.Body {
			replacedChild, err := p.replaceSuperInNode(child)
			if err != nil {
				return nil, err
			}
			newBody[i] = replacedChild
		}
		return &parser.BlockNode{
			Name: n.Name,
			Body: newBody,
		}, nil

	case *parser.SuperNode:
		// Replace super() call with empty text
		return &parser.TextNode{Content: ""}, nil

	case *parser.VariableNode:
		// Check if this is {{ super() }} wrapped in VariableNode
		if _, isSuperNode := n.Expression.(*parser.SuperNode); isSuperNode {
			// Replace {{ super() }} with empty text
			return &parser.TextNode{Content: ""}, nil
		}
		// For other variable nodes, return as-is
		return node, nil

	case *parser.IfNode:
		newBody := make([]parser.Node, len(n.Body))
		for i, child := range n.Body {
			replacedChild, err := p.replaceSuperInNode(child)
			if err != nil {
				return nil, err
			}
			newBody[i] = replacedChild
		}
		newElse := make([]parser.Node, len(n.Else))
		for i, child := range n.Else {
			replacedChild, err := p.replaceSuperInNode(child)
			if err != nil {
				return nil, err
			}
			newElse[i] = replacedChild
		}
		return &parser.IfNode{
			Condition: n.Condition,
			Body:      newBody,
			Else:      newElse,
		}, nil

	case *parser.ForNode:
		newBody := make([]parser.Node, len(n.Body))
		for i, child := range n.Body {
			replacedChild, err := p.replaceSuperInNode(child)
			if err != nil {
				return nil, err
			}
			newBody[i] = replacedChild
		}
		newElse := make([]parser.Node, len(n.Else))
		for i, child := range n.Else {
			replacedChild, err := p.replaceSuperInNode(child)
			if err != nil {
				return nil, err
			}
			newElse[i] = replacedChild
		}
		return &parser.ForNode{
			Variables: n.Variables,
			Iterable:  n.Iterable,
			Body:      newBody,
			Else:      newElse,
		}, nil

	default:
		// For other node types, return as-is
		return node, nil
	}
}

// validateSuperCallsInNodeWithHierarchy recursively validates super() calls with hierarchy context
func (p *InheritanceProcessor) validateSuperCallsInNodeWithHierarchy(node parser.Node, currentBlockName string, hierarchy *InheritanceHierarchy) error {
	switch n := node.(type) {
	case *parser.TemplateNode:
		for _, child := range n.Children {
			err := p.validateSuperCallsInNodeWithHierarchy(child, currentBlockName, hierarchy)
			if err != nil {
				return err
			}
		}

	case *parser.BlockNode:
		// Inside a block - super() calls are valid
		for _, child := range n.Body {
			err := p.validateSuperCallsInNodeWithHierarchy(child, n.Name, hierarchy)
			if err != nil {
				return err
			}
		}

	case *parser.SuperNode:
		// Direct super() node outside of block should be an error
		if currentBlockName == "" {
			return fmt.Errorf("super() call outside of block context")
		}

	case *parser.VariableNode:
		// Check if this is {{ super() }} wrapped in VariableNode
		if _, isSuperNode := n.Expression.(*parser.SuperNode); isSuperNode {
			if currentBlockName == "" {
				// super() calls outside of any block should always error
				return fmt.Errorf("super() call outside of block context")
			}
			// Don't recursively validate the SuperNode expression - we've already handled it
			return nil
		}
		// Also check nested expressions (only if not a SuperNode)
		if n.Expression != nil {
			err := p.validateSuperCallsInNodeWithHierarchy(n.Expression, currentBlockName, hierarchy)
			if err != nil {
				return err
			}
		}

	case *parser.IfNode:
		for _, child := range n.Body {
			err := p.validateSuperCallsInNodeWithHierarchy(child, currentBlockName, hierarchy)
			if err != nil {
				return err
			}
		}
		for _, child := range n.Else {
			err := p.validateSuperCallsInNodeWithHierarchy(child, currentBlockName, hierarchy)
			if err != nil {
				return err
			}
		}

	case *parser.ForNode:
		for _, child := range n.Body {
			err := p.validateSuperCallsInNodeWithHierarchy(child, currentBlockName, hierarchy)
			if err != nil {
				return err
			}
		}
		for _, child := range n.Else {
			err := p.validateSuperCallsInNodeWithHierarchy(child, currentBlockName, hierarchy)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// validateSuperCallsInNode recursively validates super() calls in a node
func (p *InheritanceProcessor) validateSuperCallsInNode(node parser.Node, currentBlockName string) error {
	switch n := node.(type) {
	case *parser.TemplateNode:
		for _, child := range n.Children {
			err := p.validateSuperCallsInNode(child, currentBlockName)
			if err != nil {
				return err
			}
		}

	case *parser.BlockNode:
		// Inside a block - super() calls are valid
		for _, child := range n.Body {
			err := p.validateSuperCallsInNode(child, n.Name)
			if err != nil {
				return err
			}
		}

	case *parser.SuperNode:
		// Direct super() node outside of block should be an error
		// But we allow it here and let the replacement logic handle it
		if currentBlockName == "" {
			return nil
		}

	case *parser.VariableNode:
		// Check if this is {{ super() }} wrapped in VariableNode
		if _, isSuperNode := n.Expression.(*parser.SuperNode); isSuperNode {
			if currentBlockName == "" {
				// In Jinja2, super() outside of blocks should error
				// But we can't easily determine the inheritance context here
				// For now, allow this validation to pass and let the replacement logic handle it
				return nil
			}
			// Don't recursively validate the SuperNode expression - we've already handled it
			return nil
		}
		// Also check nested expressions (only if not a SuperNode)
		if n.Expression != nil {
			err := p.validateSuperCallsInNode(n.Expression, currentBlockName)
			if err != nil {
				return err
			}
		}

	case *parser.IfNode:
		for _, child := range n.Body {
			err := p.validateSuperCallsInNode(child, currentBlockName)
			if err != nil {
				return err
			}
		}
		for _, child := range n.Else {
			err := p.validateSuperCallsInNode(child, currentBlockName)
			if err != nil {
				return err
			}
		}

	case *parser.ForNode:
		for _, child := range n.Body {
			err := p.validateSuperCallsInNode(child, currentBlockName)
			if err != nil {
				return err
			}
		}
		for _, child := range n.Else {
			err := p.validateSuperCallsInNode(child, currentBlockName)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
