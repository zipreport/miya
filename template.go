package miya

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"regexp"
	"runtime/debug"
	"strings"
	"sync"

	"github.com/zipreport/miya/parser"
	"github.com/zipreport/miya/runtime"
)

// Pre-compiled regexes for performance (Phase 3a optimization)
var (
	htmlOptionRegex = regexp.MustCompile(`(?s)(<option[^>]*>)(.*?)(</option>)`)
	whitespaceRegex = regexp.MustCompile(`\s+`)
)

// Template is a compiled template ready for rendering.
// It holds the parsed AST and a reference to the Environment that created it.
// Templates are safe for concurrent rendering.
type Template struct {
	name     string
	source   string
	env      *Environment
	cacheKey string      // key of this template in env.cache; "" when the Environment did not cache it
	ast      parser.Node // Will be set when parser is implemented

	// Cached inheritance check result (nil = not yet computed)
	hasInheritanceCache *bool
	cacheMu             sync.RWMutex // Protects hasInheritanceCache
}

// ErrRenderPanic is wrapped by the error returned when a render recovers from a
// panic. Its presence means a bug in miya or in a registered filter or function,
// not a template error.
var ErrRenderPanic = errors.New("miya: recovered panic during render")

// Render executes the template with the given context and returns the rendered output as a string.
// It resolves template inheritance, applies whitespace processing, and handles auto-escaping.
func (t *Template) Render(context Context) (output string, err error) {
	defer func() {
		if r := recover(); r != nil {
			output = ""
			err = fmt.Errorf("%w: %v\n%s", ErrRenderPanic, r, debug.Stack())
		}
	}()

	var buf bytes.Buffer
	err = t.RenderTo(&buf, context)
	if err != nil {
		return "", err
	}
	output = buf.String()

	// Apply block trimming if enabled
	if t.env.trimBlocks || t.env.lstripBlocks {
		output = t.env.applyBlockTrimming(output)
	}

	// Apply HTML whitespace normalization for option elements
	output = normalizeHTMLOptionWhitespace(output)

	return output, nil
}

// RenderTo executes the template with the given context and writes the output to w.
// This is more efficient than Render when writing to an io.Writer directly.
func (t *Template) RenderTo(w io.Writer, context Context) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%w: %v\n%s", ErrRenderPanic, r, debug.Stack())
		}
	}()

	if t.ast == nil {
		// If no AST is available, just write the source as-is
		_, err := w.Write([]byte(t.source))
		return err
	}

	// Create context with environment
	ctx := newContextWithEnv(t.env)
	if context != nil {
		for k, v := range context.All() {
			ctx.Set(k, v)
		}
	}

	// Resolve inheritance at render-time if needed
	finalAST := t.ast
	if t.hasInheritanceDirectives() {
		// Use shared inheritance processor from environment (reuses cache)
		processor := t.env.getInheritanceProcessor()
		// Create context adapter for runtime package compatibility
		runtimeCtx := &TemplateContextAdapter{ctx: ctx, env: t.env}
		resolvedAST, err := processor.ResolveInheritance(&templateAdapter{template: t}, runtimeCtx)
		if err != nil {
			return fmt.Errorf("inheritance resolution error: %v", err)
		}
		finalAST = resolvedAST
	}

	// Get evaluator from pool (performance optimization)
	evaluator := t.env.evaluatorPool.Get().(*runtime.DefaultEvaluator)

	// Reset evaluator state for this render
	evaluator.SetUndefinedBehavior(t.env.undefinedBehavior)
	evaluator.SetImportSystem(t.env.importSystem)

	result, err := evaluator.EvalNode(finalAST, &TemplateContextAdapter{ctx: ctx, env: t.env})
	t.env.evaluatorPool.Put(evaluator)
	if err != nil {
		return err
	}

	// Convert result to string
	var resultStr string
	if str, ok := result.(string); ok {
		resultStr = str
	} else {
		resultStr = fmt.Sprintf("%v", result)
	}

	// Write the result
	_, err = w.Write([]byte(resultStr))
	return err
}

// TemplateContextAdapter adapts Context to runtime.Context interface
type TemplateContextAdapter struct {
	ctx Context
	env *Environment
}

// NewTemplateContextAdapter creates a new TemplateContextAdapter
func NewTemplateContextAdapter(ctx Context, env *Environment) *TemplateContextAdapter {
	return &TemplateContextAdapter{ctx: ctx, env: env}
}

func (a *TemplateContextAdapter) GetVariable(name string) (interface{}, bool) {
	return a.ctx.Get(name)
}

func (a *TemplateContextAdapter) SetVariable(name string, value interface{}) {
	a.ctx.Set(name, value)
}

func (a *TemplateContextAdapter) Clone() runtime.Context {
	return &TemplateContextAdapter{ctx: a.ctx.Clone(), env: a.env}
}

// NewModuleContext returns an empty context that shares the environment, for
// templates imported without context.
func (a *TemplateContextAdapter) NewModuleContext() runtime.Context {
	return &TemplateContextAdapter{ctx: newContextWithEnv(a.env), env: a.env}
}

func (a *TemplateContextAdapter) All() map[string]interface{} {
	return a.ctx.All()
}

func (a *TemplateContextAdapter) ApplyFilterKwargs(name string, value interface{}, args []interface{}, kwargs map[string]interface{}) (interface{}, error) {
	return a.env.ApplyFilterKwargs(name, value, args, kwargs)
}

func (a *TemplateContextAdapter) ApplyFilter(name string, value interface{}, args ...interface{}) (interface{}, error) {
	return a.env.ApplyFilter(name, value, args...)
}

func (a *TemplateContextAdapter) ApplyTest(name string, value interface{}, args ...interface{}) (bool, error) {
	return a.env.ApplyTest(name, value, args...)
}

// IsAutoescapeEnabled returns the environment's autoescape setting
func (a *TemplateContextAdapter) IsAutoescapeEnabled() bool {
	return a.env.autoEscape
}

// Name returns the template name (filename or "<string>" for string templates).
func (t *Template) Name() string {
	return t.name
}

// Source returns the original template source code.
func (t *Template) Source() string {
	return t.source
}

// AST returns the parsed abstract syntax tree of the template.
func (t *Template) AST() parser.Node {
	return t.ast
}

// normalizeHTMLOptionWhitespace compresses whitespace within HTML option elements
func normalizeHTMLOptionWhitespace(html string) string {
	// Use pre-compiled regex (Phase 3a optimization)
	return htmlOptionRegex.ReplaceAllStringFunc(html, func(match string) string {
		parts := htmlOptionRegex.FindStringSubmatch(match)
		if len(parts) != 4 {
			return match
		}

		openTag := parts[1]
		content := parts[2]
		closeTag := parts[3]

		// Use pre-compiled whitespace regex (Phase 3a optimization)
		normalized := whitespaceRegex.ReplaceAllString(strings.TrimSpace(content), " ")

		return openTag + normalized + closeTag
	})
}

// GetASTAsTemplateNode returns the AST as a TemplateNode for inheritance processing
func (t *Template) GetASTAsTemplateNode() *parser.TemplateNode {
	if templateNode, ok := t.ast.(*parser.TemplateNode); ok {
		return templateNode
	}
	return nil
}

func (t *Template) SetAST(ast parser.Node) {
	t.ast = ast
	// Invalidate inheritance cache when AST changes
	t.cacheMu.Lock()
	t.hasInheritanceCache = nil
	t.cacheMu.Unlock()
}

// Release removes this template from its Environment's template cache, so
// the Environment no longer keeps it alive. Use it for one-off templates,
// such as many distinct FromString sources, to keep the cache from growing.
//
// The template stays usable, and anyone else holding it is unaffected; a
// later FromString or GetTemplate of the same source or name parses it
// again. Release only removes the cache entry if it still points to this
// template, so it is safe to call more than once, concurrently, or after
// ClearCache.
func (t *Template) Release() {
	if t.env == nil || t.cacheKey == "" {
		return
	}
	e := t.env
	e.cacheMutex.Lock()
	if cached, ok := e.cache[t.cacheKey]; ok && cached == t {
		delete(e.cache, t.cacheKey)
	}
	e.cacheMutex.Unlock()
}

// hasInheritanceDirectives checks if template contains {% extends %} directives
// Result is cached after first computation to avoid AST traversal on every render
func (t *Template) hasInheritanceDirectives() bool {
	// First, try to read cached result with read lock
	t.cacheMu.RLock()
	if t.hasInheritanceCache != nil {
		result := *t.hasInheritanceCache
		t.cacheMu.RUnlock()
		return result
	}
	t.cacheMu.RUnlock()

	// Compute result (outside lock to avoid holding lock during traversal)
	result := t.hasInheritanceDirectivesInNode(t.ast)

	// Cache the result with write lock
	t.cacheMu.Lock()
	// Double-check in case another goroutine computed it
	if t.hasInheritanceCache == nil {
		t.hasInheritanceCache = &result
	} else {
		result = *t.hasInheritanceCache
	}
	t.cacheMu.Unlock()

	return result
}

// hasInheritanceDirectivesInNode recursively checks for inheritance directives
func (t *Template) hasInheritanceDirectivesInNode(node parser.Node) bool {
	if node == nil {
		return false
	}

	switch n := node.(type) {
	case *parser.TemplateNode:
		for _, child := range n.Children {
			if t.hasInheritanceDirectivesInNode(child) {
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
		if t.hasInheritanceDirectivesInNode(n.Expression) {
			return true
		}

	case *parser.BlockNode:
		for _, child := range n.Body {
			if t.hasInheritanceDirectivesInNode(child) {
				return true
			}
		}

	case *parser.ForNode:
		for _, child := range n.Body {
			if t.hasInheritanceDirectivesInNode(child) {
				return true
			}
		}

	case *parser.IfNode:
		for _, child := range n.Body {
			if t.hasInheritanceDirectivesInNode(child) {
				return true
			}
		}
		for _, child := range n.Else {
			if t.hasInheritanceDirectivesInNode(child) {
				return true
			}
		}
		for _, elseIf := range n.ElseIfs {
			if t.hasInheritanceDirectivesInNode(elseIf) {
				return true
			}
		}

	// Add more node types that might contain super() calls
	case parser.ExpressionNode:
		// Handle any other expression nodes generically
		// This catches cases where super() might be nested in other expressions
		if _, isSuperNode := n.(*parser.SuperNode); isSuperNode {
			return true
		}
	}
	return false
}

// environmentAdapter adapts Environment to runtime.EnvironmentInterface
type environmentAdapter struct {
	env *Environment
}

func (a *environmentAdapter) GetTemplate(name string) (runtime.TemplateInterface, error) {
	template, err := a.env.GetTemplate(name)
	if err != nil {
		return nil, err
	}
	return &templateAdapter{template: template}, nil
}

func (a *environmentAdapter) GetLoader() interface{} {
	return a.env.GetLoader()
}

// templateAdapter adapts Template to runtime.TemplateInterface
type templateAdapter struct {
	template *Template
}

func (a *templateAdapter) AST() *parser.TemplateNode {
	return a.template.GetASTAsTemplateNode()
}

func (a *templateAdapter) Name() string {
	return a.template.Name()
}

type Node interface {
	// Will be expanded when we implement the parser
	String() string
}
