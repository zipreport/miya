package inheritance

import (
	"fmt"

	"github.com/zipreport/miya/parser"
	"github.com/zipreport/miya/runtime"
)

// TemplateLoader interface for loading templates by name
type TemplateLoader interface {
	LoadTemplate(name string) (*parser.TemplateNode, error)
	ResolveTemplateName(name string) string
}

// InheritanceResolver handles template inheritance logic
type InheritanceResolver struct {
	loader TemplateLoader
}

// NewInheritanceResolver creates a new inheritance resolver
func NewInheritanceResolver(loader TemplateLoader) *InheritanceResolver {
	return &InheritanceResolver{
		loader: loader,
	}
}

// BlockContext holds information about blocks in the inheritance chain
type BlockContext struct {
	Name     string
	Body     []parser.Node
	Parent   *BlockContext
	Template string
}

// TemplateContext represents a template in the inheritance chain
type TemplateContext struct {
	Name      string
	Template  *parser.TemplateNode
	Parent    *TemplateContext
	Blocks    map[string]*BlockContext
	SuperCall bool
}

// ResolveInheritance resolves template inheritance for a given template
func (r *InheritanceResolver) ResolveInheritance(template *parser.TemplateNode, templateName string) (*TemplateContext, error) {
	ctx := &TemplateContext{
		Name:     templateName,
		Template: template,
		Blocks:   make(map[string]*BlockContext),
	}

	// Collect blocks from this template
	if err := r.collectBlocks(template.Children, ctx, templateName); err != nil {
		return nil, err
	}

	// Check if template extends another template
	extendsNode := r.findExtendsNode(template.Children)
	if extendsNode != nil {
		parentName := r.evaluateTemplateName(extendsNode.Template)
		parentTemplate, err := r.loader.LoadTemplate(parentName)
		if err != nil {
			return nil, fmt.Errorf("failed to load parent template %s: %v", parentName, err)
		}

		// Recursively resolve parent inheritance
		parentCtx, err := r.ResolveInheritance(parentTemplate, parentName)
		if err != nil {
			return nil, err
		}

		ctx.Parent = parentCtx

		// Merge blocks from parent, child blocks override parent blocks
		for name, parentBlock := range parentCtx.Blocks {
			if childBlock, exists := ctx.Blocks[name]; exists {
				// Child block overrides parent, but keep parent as super
				childBlock.Parent = parentBlock
			} else {
				// Use parent block as-is
				ctx.Blocks[name] = parentBlock
			}
		}
	}

	return ctx, nil
}

// BuildFinalTemplate constructs the final template by resolving inheritance
func (r *InheritanceResolver) BuildFinalTemplate(ctx *TemplateContext) (*parser.TemplateNode, error) {
	// Start with the root template (top of inheritance chain)
	rootCtx := ctx
	for rootCtx.Parent != nil {
		rootCtx = rootCtx.Parent
	}

	// Create final template based on root template
	finalTemplate := &parser.TemplateNode{
		Name:     ctx.Name,
		Children: make([]parser.Node, 0),
	}

	// Process root template nodes, replacing blocks with resolved versions
	finalNodes, err := r.processNodes(rootCtx.Template.Children, ctx)
	if err != nil {
		return nil, err
	}

	finalTemplate.Children = finalNodes
	return finalTemplate, nil
}

// processNodes processes a list of nodes, resolving blocks and includes
func (r *InheritanceResolver) processNodes(nodes []parser.Node, ctx *TemplateContext) ([]parser.Node, error) {
	var result []parser.Node

	for _, node := range nodes {
		switch n := node.(type) {
		case *parser.BlockNode:
			// Replace with resolved block content
			if blockCtx, exists := ctx.Blocks[n.Name]; exists {
				resolvedNodes, err := r.resolveBlockContent(blockCtx, ctx)
				if err != nil {
					return nil, err
				}
				result = append(result, resolvedNodes...)
			} else {
				// Block not found, use original
				result = append(result, node)
			}

		case *parser.ExtendsNode:
			// Skip extends nodes in final template
			continue

		case *parser.IncludeNode:
			// Resolve include
			includedNodes, err := r.resolveInclude(n, ctx)
			if err != nil {
				return nil, err
			}
			result = append(result, includedNodes...)

		default:
			// Process child nodes recursively for compound nodes
			processedNode, err := r.processNode(node, ctx)
			if err != nil {
				return nil, err
			}
			result = append(result, processedNode)
		}
	}

	return result, nil
}

// processNode processes a single node recursively
func (r *InheritanceResolver) processNode(node parser.Node, ctx *TemplateContext) (parser.Node, error) {
	switch n := node.(type) {
	case *parser.IfNode:
		processedBody, err := r.processNodes(n.Body, ctx)
		if err != nil {
			return nil, err
		}
		processedElse, err := r.processNodes(n.Else, ctx)
		if err != nil {
			return nil, err
		}

		newNode := *n
		newNode.Body = processedBody
		newNode.Else = processedElse

		// Process elif conditions
		for i, elif := range n.ElseIfs {
			processedElifBody, err := r.processNodes(elif.Body, ctx)
			if err != nil {
				return nil, err
			}
			newNode.ElseIfs[i].Body = processedElifBody
		}

		return &newNode, nil

	case *parser.ForNode:
		processedBody, err := r.processNodes(n.Body, ctx)
		if err != nil {
			return nil, err
		}
		processedElse, err := r.processNodes(n.Else, ctx)
		if err != nil {
			return nil, err
		}

		newNode := *n
		newNode.Body = processedBody
		newNode.Else = processedElse
		return &newNode, nil

	default:
		// Return node as-is for leaf nodes
		return node, nil
	}
}

// resolveBlockContent resolves the content of a block, handling super() calls
func (r *InheritanceResolver) resolveBlockContent(blockCtx *BlockContext, templateCtx *TemplateContext) ([]parser.Node, error) {
	// Process the block body, looking for super() calls
	return r.processNodesWithSuper(blockCtx.Body, blockCtx, templateCtx)
}

// processNodesWithSuper processes nodes while handling super() calls
func (r *InheritanceResolver) processNodesWithSuper(nodes []parser.Node, blockCtx *BlockContext, templateCtx *TemplateContext) ([]parser.Node, error) {
	var result []parser.Node

	for _, node := range nodes {
		switch node.(type) {
		case *parser.SuperNode:
			// Replace super() with parent block content
			if blockCtx.Parent != nil {
				parentContent, err := r.resolveBlockContent(blockCtx.Parent, templateCtx)
				if err != nil {
					return nil, err
				}
				result = append(result, parentContent...)
			}
			// If no parent block, super() produces nothing

		default:
			// Process recursively for compound nodes
			processedNode, err := r.processNodeWithSuper(node, blockCtx, templateCtx)
			if err != nil {
				return nil, err
			}
			result = append(result, processedNode)
		}
	}

	return result, nil
}

// processNodeWithSuper processes a single node while handling super() calls
func (r *InheritanceResolver) processNodeWithSuper(node parser.Node, blockCtx *BlockContext, templateCtx *TemplateContext) (parser.Node, error) {
	switch n := node.(type) {
	case *parser.IfNode:
		processedBody, err := r.processNodesWithSuper(n.Body, blockCtx, templateCtx)
		if err != nil {
			return nil, err
		}
		processedElse, err := r.processNodesWithSuper(n.Else, blockCtx, templateCtx)
		if err != nil {
			return nil, err
		}

		newNode := *n
		newNode.Body = processedBody
		newNode.Else = processedElse

		// Process elif conditions
		for i, elif := range n.ElseIfs {
			processedElifBody, err := r.processNodesWithSuper(elif.Body, blockCtx, templateCtx)
			if err != nil {
				return nil, err
			}
			newNode.ElseIfs[i].Body = processedElifBody
		}

		return &newNode, nil

	case *parser.ForNode:
		processedBody, err := r.processNodesWithSuper(n.Body, blockCtx, templateCtx)
		if err != nil {
			return nil, err
		}
		processedElse, err := r.processNodesWithSuper(n.Else, blockCtx, templateCtx)
		if err != nil {
			return nil, err
		}

		newNode := *n
		newNode.Body = processedBody
		newNode.Else = processedElse
		return &newNode, nil

	default:
		// Return node as-is for leaf nodes
		return node, nil
	}
}

// resolveInclude resolves an include statement
func (r *InheritanceResolver) resolveInclude(includeNode *parser.IncludeNode, ctx *TemplateContext) ([]parser.Node, error) {
	templateName := r.evaluateTemplateName(includeNode.Template)

	includedTemplate, err := r.loader.LoadTemplate(templateName)
	if err != nil {
		if includeNode.IgnoreMissing {
			// Return empty if template is missing and ignore_missing is true
			return []parser.Node{}, nil
		}
		return nil, fmt.Errorf("failed to load included template %s: %v", templateName, err)
	}

	// Process included template (no inheritance resolution for includes)
	return r.processNodes(includedTemplate.Children, ctx)
}

// collectBlocks finds all block nodes in the template and creates block contexts
func (r *InheritanceResolver) collectBlocks(nodes []parser.Node, ctx *TemplateContext, templateName string) error {
	for _, node := range nodes {
		if err := r.collectBlocksFromNode(node, ctx, templateName); err != nil {
			return err
		}
	}
	return nil
}

// collectBlocksFromNode recursively collects blocks from a node
func (r *InheritanceResolver) collectBlocksFromNode(node parser.Node, ctx *TemplateContext, templateName string) error {
	switch n := node.(type) {
	case *parser.BlockNode:
		blockCtx := &BlockContext{
			Name:     n.Name,
			Body:     n.Body,
			Template: templateName,
		}
		ctx.Blocks[n.Name] = blockCtx

	case *parser.IfNode:
		if err := r.collectBlocks(n.Body, ctx, templateName); err != nil {
			return err
		}
		if err := r.collectBlocks(n.Else, ctx, templateName); err != nil {
			return err
		}
		for _, elif := range n.ElseIfs {
			if err := r.collectBlocks(elif.Body, ctx, templateName); err != nil {
				return err
			}
		}

	case *parser.ForNode:
		if err := r.collectBlocks(n.Body, ctx, templateName); err != nil {
			return err
		}
		if err := r.collectBlocks(n.Else, ctx, templateName); err != nil {
			return err
		}
	}

	return nil
}

// findExtendsNode finds the extends node in a template
func (r *InheritanceResolver) findExtendsNode(nodes []parser.Node) *parser.ExtendsNode {
	for _, node := range nodes {
		if extendsNode, ok := node.(*parser.ExtendsNode); ok {
			return extendsNode
		}
	}
	return nil
}

// evaluateTemplateName evaluates a template name expression to a string
func (r *InheritanceResolver) evaluateTemplateName(expr parser.Node) string {
	// For now, assume template names are literal strings
	if literal, ok := expr.(*parser.LiteralNode); ok {
		if str, ok := literal.Value.(string); ok {
			return str
		}
	}
	// Fallback to string conversion
	return fmt.Sprintf("%v", expr)
}

// MemoryTemplateLoader is a simple in-memory template loader for testing
type MemoryTemplateLoader struct {
	templates map[string]*parser.TemplateNode
}

// NewMemoryTemplateLoader creates a new memory template loader
func NewMemoryTemplateLoader() *MemoryTemplateLoader {
	return &MemoryTemplateLoader{
		templates: make(map[string]*parser.TemplateNode),
	}
}

// AddTemplate adds a template to the loader
func (m *MemoryTemplateLoader) AddTemplate(name string, template *parser.TemplateNode) {
	m.templates[name] = template
}

// LoadTemplate loads a template by name
func (m *MemoryTemplateLoader) LoadTemplate(name string) (*parser.TemplateNode, error) {
	template, exists := m.templates[name]
	if !exists {
		return nil, fmt.Errorf("template not found: %s", name)
	}
	return template, nil
}

// ResolveTemplateName resolves a template name (identity function for memory loader)
func (m *MemoryTemplateLoader) ResolveTemplateName(name string) string {
	return name
}

// InheritanceEvaluator extends the runtime evaluator with inheritance support
type InheritanceEvaluator struct {
	evaluator runtime.Evaluator
	resolver  *InheritanceResolver
}

// NewInheritanceEvaluator creates a new inheritance-aware evaluator
func NewInheritanceEvaluator(loader TemplateLoader) *InheritanceEvaluator {
	return &InheritanceEvaluator{
		evaluator: runtime.NewEvaluator(),
		resolver:  NewInheritanceResolver(loader),
	}
}

// EvaluateTemplate evaluates a template with inheritance resolution
func (e *InheritanceEvaluator) EvaluateTemplate(template *parser.TemplateNode, templateName string, ctx runtime.Context) (string, error) {
	// Resolve inheritance
	inheritanceCtx, err := e.resolver.ResolveInheritance(template, templateName)
	if err != nil {
		return "", err
	}

	// Build final template
	finalTemplate, err := e.resolver.BuildFinalTemplate(inheritanceCtx)
	if err != nil {
		return "", err
	}

	// Evaluate final template
	result, err := e.evaluator.EvalNode(finalTemplate, ctx)
	if err != nil {
		return "", err
	}

	if str, ok := result.(string); ok {
		return str, nil
	}

	return fmt.Sprintf("%v", result), nil
}

// Helper function to create template name from string
func CreateTemplateNameNode(name string) parser.Node {
	return &parser.LiteralNode{Value: name}
}
