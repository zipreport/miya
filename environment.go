package miya

import (
	"fmt"
	"hash/fnv"
	"strings"
	"sync"
	"time"

	"github.com/zipreport/miya/branching"
	"github.com/zipreport/miya/extensions"
	"github.com/zipreport/miya/filters"
	"github.com/zipreport/miya/inheritance"
	"github.com/zipreport/miya/lexer"
	"github.com/zipreport/miya/loader"
	"github.com/zipreport/miya/macros"
	"github.com/zipreport/miya/parser"
	"github.com/zipreport/miya/runtime"
	"github.com/zipreport/miya/whitespace"
)

type Environment struct {
	loader              Loader
	filterRegistry      *filters.FilterRegistry
	inheritanceResolver *inheritance.InheritanceResolver
	macroRegistry       *macros.MacroRegistry
	testRegistry        *branching.TestRegistry
	whitespaceProcessor *whitespace.WhitespaceProcessor
	extensionRegistry   *extensions.Registry
	globals             map[string]interface{}
	tests               map[string]TestFunc // deprecated, use testRegistry
	cache               map[string]*Template
	cacheMutex          sync.RWMutex

	// New inheritance caching system
	inheritanceCache      *runtime.InheritanceCache
	inheritanceProcessor  *runtime.InheritanceProcessor
	inheritanceCacheMutex sync.RWMutex

	// Performance: Evaluator pooling
	evaluatorPool sync.Pool
	importSystem  *runtime.ImportSystem

	autoEscape          bool
	trimBlocks          bool
	lstripBlocks        bool
	keepTrailingNewline bool
	undefinedBehavior   runtime.UndefinedBehavior
	extensionConfig     map[string]interface{} // Extension-specific configuration

	varStartString     string
	varEndString       string
	blockStartString   string
	blockEndString     string
	commentStartString string
	commentEndString   string
}

type EnvironmentOption func(*Environment)

// hashString generates a content-based cache key for template strings
// Uses FNV-1a hash algorithm for fast, uniform distribution
func hashString(s string) string {
	h := fnv.New64a()
	h.Write([]byte(s))
	return fmt.Sprintf("__string__%x", h.Sum64())
}

func NewEnvironment(opts ...EnvironmentOption) *Environment {
	inheritanceCache := runtime.NewInheritanceCache()

	env := &Environment{
		filterRegistry:      filters.NewRegistry(),
		macroRegistry:       macros.NewMacroRegistry(),
		testRegistry:        branching.NewTestRegistry(),
		whitespaceProcessor: whitespace.NewWhitespaceProcessor(false, false, false),
		extensionRegistry:   extensions.NewRegistry(),
		globals:             make(map[string]interface{}),
		tests:               make(map[string]TestFunc), // deprecated
		cache:               make(map[string]*Template),
		inheritanceCache:    inheritanceCache,
		extensionConfig:     make(map[string]interface{}),
		autoEscape:          true,
		undefinedBehavior:   runtime.UndefinedSilent, // Default to silent undefined

		varStartString:     "{{",
		varEndString:       "}}",
		blockStartString:   "{%",
		blockEndString:     "%}",
		commentStartString: "{#",
		commentEndString:   "#}",
	}

	for _, opt := range opts {
		opt(env)
	}

	// Initialize evaluator pool for performance
	env.evaluatorPool = sync.Pool{
		New: func() interface{} {
			eval := runtime.NewEvaluator()
			eval.SetUndefinedBehavior(env.undefinedBehavior)
			return eval
		},
	}

	// Initialize shared import system (reused across renders)
	templateLoader := runtime.NewSimpleTemplateLoader(env)
	env.importSystem = runtime.NewImportSystem(templateLoader, nil)

	// Update whitespace processor with final settings
	env.whitespaceProcessor = whitespace.NewWhitespaceProcessor(
		env.trimBlocks,
		env.lstripBlocks,
		env.keepTrailingNewline,
	)

	// Create inheritance resolver if we have a loader
	if env.loader != nil {
		env.inheritanceResolver = inheritance.NewInheritanceResolver(&loaderAdapter{env.loader})
	}

	registerBuiltinTests(env)
	registerBuiltinGlobals(env)

	// Set up extension registry with environment reference
	env.extensionRegistry.SetEnvironment(env)

	// Note: inheritanceProcessor will be initialized lazily to avoid import cycles

	return env
}

func (e *Environment) GetTemplate(name string) (*Template, error) {
	e.cacheMutex.RLock()
	if tmpl, ok := e.cache[name]; ok {
		e.cacheMutex.RUnlock()
		return tmpl, nil
	}
	e.cacheMutex.RUnlock()

	if e.loader == nil {
		return nil, fmt.Errorf("no loader configured for environment")
	}

	// Try to load from advanced loader first if available
	if advancedLoader, ok := e.loader.(loader.AdvancedLoader); ok {
		templateNode, err := advancedLoader.LoadTemplate(name)
		if err != nil {
			return nil, fmt.Errorf("failed to load template %q: %w", name, err)
		}

		// NOTE: Old inheritance resolution removed - now handled at render-time
		// This allows templates to preserve their raw AST with ExtendsNode and SuperNode
		// for processing by the new runtime inheritance system

		tmpl := &Template{
			name: name,
			env:  e,
			ast:  templateNode,
		}

		e.cacheMutex.Lock()
		e.cache[name] = tmpl
		e.cacheMutex.Unlock()

		return tmpl, nil
	}

	// Fallback to basic loader
	source, err := e.loader.GetSource(name)
	if err != nil {
		return nil, fmt.Errorf("failed to load template %q: %w", name, err)
	}

	tmpl, err := e.compile(name, source)
	if err != nil {
		return nil, err
	}

	e.cacheMutex.Lock()
	e.cache[name] = tmpl
	e.cacheMutex.Unlock()

	return tmpl, nil
}

// hasInlineWhitespaceControl checks if template contains inline whitespace control syntax
func (e *Environment) hasInlineWhitespaceControl(source string) bool {
	// Fast path: if no '-' exists, no whitespace control syntax is present
	if strings.IndexByte(source, '-') == -1 {
		return false
	}
	// Only do detailed checks if '-' was found
	return strings.Contains(source, "{%-") ||
		strings.Contains(source, "-%}") ||
		strings.Contains(source, "{{-") ||
		strings.Contains(source, "-}}") ||
		strings.Contains(source, "{#-") ||
		strings.Contains(source, "-#}")
}

func (e *Environment) FromString(source string) (*Template, error) {
	// Generate cache key from content hash (Phase 2 optimization)
	cacheKey := hashString(source)

	// Check cache first
	e.cacheMutex.RLock()
	if tmpl, ok := e.cache[cacheKey]; ok {
		e.cacheMutex.RUnlock()
		return tmpl, nil
	}
	e.cacheMutex.RUnlock()

	// Compile if not cached (use "<string>" as display name)
	tmpl, err := e.compile("<string>", source)
	if err != nil {
		return nil, err
	}

	// Store in cache with hash key
	e.cacheMutex.Lock()
	e.cache[cacheKey] = tmpl
	e.cacheMutex.Unlock()

	return tmpl, nil
}

func (e *Environment) SetLoader(loader Loader) {
	e.loader = loader
	// Create inheritance resolver if we have a loader
	if loader != nil {
		e.inheritanceResolver = inheritance.NewInheritanceResolver(&loaderAdapter{loader})
	} else {
		e.inheritanceResolver = nil
	}
}

func (e *Environment) AddFilterLegacy(name string, filter FilterFunc) error {
	return e.filterRegistry.Register(name, filters.FilterFunc(filter))
}

func (e *Environment) GetFilter(name string) (FilterFunc, bool) {
	f, ok := e.filterRegistry.Get(name)
	if !ok {
		return nil, false
	}
	return FilterFunc(f), true
}

func (e *Environment) ApplyFilter(name string, value interface{}, args ...interface{}) (interface{}, error) {
	return e.filterRegistry.Apply(name, value, args...)
}

func (e *Environment) AddGlobal(name string, value interface{}) {
	e.globals[name] = value
}

// AddExtension registers an extension with the environment
func (e *Environment) AddExtension(extension extensions.Extension) error {
	return e.extensionRegistry.Register(extension)
}

// GetExtensionRegistry returns the extension registry
func (e *Environment) GetExtensionRegistry() *extensions.Registry {
	return e.extensionRegistry
}

// IsCustomTag returns true if the tag is handled by an extension
func (e *Environment) IsCustomTag(tagName string) bool {
	return e.extensionRegistry.IsCustomTag(tagName)
}

// GetExtensionForTag returns the extension that handles the given tag
func (e *Environment) GetExtensionForTag(tagName string) (extensions.Extension, bool) {
	return e.extensionRegistry.GetExtensionForTag(tagName)
}

// ExtensionEnvironment interface implementation
// GetExtension retrieves another extension by name
func (e *Environment) GetExtension(name string) (extensions.Extension, bool) {
	return e.extensionRegistry.GetExtension(name)
}

// AddFilter adds a custom filter (for extensions and legacy support)
func (e *Environment) AddFilter(name string, filter interface{}) error {
	// Handle legacy FilterFunc type
	if filterFunc, ok := filter.(FilterFunc); ok {
		return e.filterRegistry.Register(name, filters.FilterFunc(filterFunc))
	}

	// Handle filters.FilterFunc type
	if filterFunc, ok := filter.(filters.FilterFunc); ok {
		return e.filterRegistry.Register(name, filterFunc)
	}

	// For other types, try to convert to filters.FilterFunc
	if filterFunc, ok := filter.(func(interface{}, ...interface{}) (interface{}, error)); ok {
		return e.filterRegistry.Register(name, filters.FilterFunc(filterFunc))
	}

	return fmt.Errorf("unsupported filter type: %T", filter)
}

// AddTest adds a custom test (for extensions and legacy support)
func (e *Environment) AddTest(name string, test interface{}) error {
	// Handle legacy TestFunc type
	if testFunc, ok := test.(TestFunc); ok {
		// Legacy TestFunc - wrap it for the new registry
		wrappedTest := func(value interface{}, args ...interface{}) (bool, error) {
			return testFunc(value, args...)
		}

		// Register in new registry
		if err := e.testRegistry.Register(name, branching.TestFunc(wrappedTest)); err != nil {
			return err
		}

		// Also store in legacy map for backwards compatibility
		if _, exists := e.tests[name]; exists {
			return fmt.Errorf("test %q already registered", name)
		}
		e.tests[name] = testFunc
		return nil
	}

	// Handle branching.TestFunc type
	if testFunc, ok := test.(branching.TestFunc); ok {
		return e.testRegistry.Register(name, testFunc)
	}

	// For other types, try to convert to branching.TestFunc
	if testFunc, ok := test.(func(interface{}, ...interface{}) (bool, error)); ok {
		return e.testRegistry.Register(name, branching.TestFunc(testFunc))
	}

	return fmt.Errorf("unsupported test type: %T", test)
}

// GetConfig gets environment-level configuration for extensions
func (e *Environment) GetConfig(key string) (interface{}, bool) {
	value, exists := e.extensionConfig[key]
	return value, exists
}

// SetConfig sets environment-level configuration for extensions
func (e *Environment) SetConfig(key string, value interface{}) {
	e.extensionConfig[key] = value
}

// GetTest retrieves a test function by name
func (e *Environment) GetTest(name string) (branching.TestFunc, bool) {
	return e.testRegistry.Get(name)
}

// ListTests returns all available test names
func (e *Environment) ListTests() []string {
	return e.testRegistry.List()
}

// ApplyTest applies a test to a value
func (e *Environment) ApplyTest(name string, value interface{}, args ...interface{}) (bool, error) {
	return e.testRegistry.Apply(name, value, args...)
}

func (e *Environment) SetDelimiters(varStart, varEnd, blockStart, blockEnd string) {
	e.varStartString = varStart
	e.varEndString = varEnd
	e.blockStartString = blockStart
	e.blockEndString = blockEnd
}

func (e *Environment) SetCommentDelimiters(start, end string) {
	e.commentStartString = start
	e.commentEndString = end
}

// GetLoader returns the current template loader
func (e *Environment) GetLoader() Loader {
	return e.loader
}

// ListFilters returns all available filter names
func (e *Environment) ListFilters() []string {
	return e.filterRegistry.List()
}

// Macro management methods

// GetMacro retrieves a macro by name
func (e *Environment) GetMacro(name string) (*macros.Macro, bool) {
	return e.macroRegistry.Get(name)
}

// ListMacros returns all available macro names
func (e *Environment) ListMacros() []string {
	return e.macroRegistry.List()
}

// CallMacro calls a macro with the given arguments
func (e *Environment) CallMacro(name string, ctx Context, args []interface{}, kwargs map[string]interface{}) (interface{}, error) {
	// Create runtime context adapter
	runtimeCtx := &TemplateContextAdapter{ctx: ctx, env: e}

	// Create evaluator
	evaluator := runtime.NewEvaluator()

	// Set up import system for the evaluator
	loader := runtime.NewSimpleTemplateLoader(e)
	importSystem := runtime.NewImportSystem(loader, evaluator)
	evaluator.SetImportSystem(importSystem)

	return e.macroRegistry.CallMacro(name, runtimeCtx, evaluator, args, kwargs)
}

// ClearMacros clears all registered macros
func (e *Environment) ClearMacros() {
	e.macroRegistry.Clear()
}

// registerMacrosFromAST walks the AST and registers any macro definitions found
func (e *Environment) registerMacrosFromAST(node parser.Node, templateName string) {
	switch n := node.(type) {
	case *parser.TemplateNode:
		for _, child := range n.Children {
			e.registerMacrosFromAST(child, templateName)
		}
	case *parser.MacroNode:
		// Convert defaults from ExpressionNode to interface{}
		defaults := make(map[string]interface{})
		for key, expr := range n.Defaults {
			// For now, we'll store the expression itself
			// In a full implementation, we'd evaluate constant expressions
			defaults[key] = expr
		}

		macro := &macros.Macro{
			Name:       n.Name,
			Parameters: n.Parameters,
			Defaults:   defaults,
			Body:       n.Body,
			Template:   templateName,
		}

		// Register the macro (ignore errors for now)
		e.macroRegistry.Register(n.Name, macro)
	case *parser.IfNode:
		for _, child := range n.Body {
			e.registerMacrosFromAST(child, templateName)
		}
		for _, elifNode := range n.ElseIfs {
			for _, child := range elifNode.Body {
				e.registerMacrosFromAST(child, templateName)
			}
		}
		for _, child := range n.Else {
			e.registerMacrosFromAST(child, templateName)
		}
	case *parser.ForNode:
		for _, child := range n.Body {
			e.registerMacrosFromAST(child, templateName)
		}
		for _, child := range n.Else {
			e.registerMacrosFromAST(child, templateName)
		}
	case *parser.BlockNode:
		for _, child := range n.Body {
			e.registerMacrosFromAST(child, templateName)
		}
		// Other node types don't need special handling for macro registration
	}
}

func (e *Environment) compile(name, source string) (*Template, error) {
	// Apply whitespace preprocessing if whitespace control is enabled
	preprocessedSource := source
	if e.trimBlocks || e.lstripBlocks || e.hasInlineWhitespaceControl(source) {
		processor := whitespace.NewAdvancedWhitespaceProcessor(e.trimBlocks, e.lstripBlocks, e.keepTrailingNewline)
		preprocessedSource = processor.ProcessTemplate(source)
	}

	// Create lexer configuration from environment settings
	lexerConfig := &lexer.LexerConfig{
		VarStartString:     e.varStartString,
		VarEndString:       e.varEndString,
		BlockStartString:   e.blockStartString,
		BlockEndString:     e.blockEndString,
		CommentStartString: e.commentStartString,
		CommentEndString:   e.commentEndString,
		TrimBlocks:         e.trimBlocks,
		LstripBlocks:       e.lstripBlocks,
	}

	// Tokenize the preprocessed source
	l := lexer.NewLexer(preprocessedSource, lexerConfig)
	tokens, err := l.Tokenize()
	if err != nil {
		return nil, fmt.Errorf("lexer error in template %s: %v", name, err)
	}

	// Parse the tokens into an AST
	var ast *parser.TemplateNode

	// Use extension-aware parser if we have extensions registered
	if len(e.extensionRegistry.GetAllExtensions()) > 0 {
		extParser := extensions.NewExtensionAwareParser(tokens, e.extensionRegistry)
		var err error
		ast, err = extParser.Parse()
		if err != nil {
			return nil, fmt.Errorf("parser error in template %s: %v", name, err)
		}
	} else {
		// Use standard parser if no extensions
		p := parser.NewParser(tokens)
		var err error
		ast, err = p.Parse()
		if err != nil {
			return nil, fmt.Errorf("parser error in template %s: %v", name, err)
		}
	}

	// Set the template name in the AST
	ast.Name = name

	// Register any macros found in the template
	e.registerMacrosFromAST(ast, name)

	// NOTE: Inheritance resolution moved to render-time to avoid circular dependencies
	// Templates are now compiled with their raw AST preserved, allowing proper
	// template hierarchy loading without compilation-time circular references

	return &Template{
		name:   name,
		source: source,
		env:    e,
		ast:    ast,
	}, nil
}

func WithLoader(loader Loader) EnvironmentOption {
	return func(e *Environment) {
		e.loader = loader
		// Create inheritance resolver if we have a loader
		if loader != nil {
			e.inheritanceResolver = inheritance.NewInheritanceResolver(&loaderAdapter{loader})
		}
	}
}

func WithAutoEscape(enabled bool) EnvironmentOption {
	return func(e *Environment) {
		e.autoEscape = enabled
	}
}

func WithTrimBlocks(enabled bool) EnvironmentOption {
	return func(e *Environment) {
		e.trimBlocks = enabled
	}
}

func WithLstripBlocks(enabled bool) EnvironmentOption {
	return func(e *Environment) {
		e.lstripBlocks = enabled
	}
}

func WithKeepTrailingNewline(enabled bool) EnvironmentOption {
	return func(e *Environment) {
		e.keepTrailingNewline = enabled
	}
}

// WithStrictUndefined enables strict undefined variable handling
func WithStrictUndefined(enabled bool) EnvironmentOption {
	return func(e *Environment) {
		if enabled {
			e.undefinedBehavior = runtime.UndefinedStrict
		} else {
			e.undefinedBehavior = runtime.UndefinedSilent
		}
	}
}

// WithDebugUndefined enables debug undefined variable handling
func WithDebugUndefined(enabled bool) EnvironmentOption {
	return func(e *Environment) {
		if enabled {
			e.undefinedBehavior = runtime.UndefinedDebug
		} else {
			e.undefinedBehavior = runtime.UndefinedSilent
		}
	}
}

// WithUndefinedBehavior sets the undefined variable behavior
func WithUndefinedBehavior(behavior runtime.UndefinedBehavior) EnvironmentOption {
	return func(e *Environment) {
		e.undefinedBehavior = behavior
	}
}

// Additional Environment methods

// ClearCache clears the template cache
func (e *Environment) ClearCache() {
	e.cacheMutex.Lock()
	defer e.cacheMutex.Unlock()
	e.cache = make(map[string]*Template)
}

// GetCacheSize returns the number of cached templates
func (e *Environment) GetCacheSize() int {
	e.cacheMutex.RLock()
	defer e.cacheMutex.RUnlock()
	return len(e.cache)
}

// Inheritance cache management methods

// ClearInheritanceCache clears all inheritance-related caches
func (e *Environment) ClearInheritanceCache() {
	e.inheritanceCacheMutex.Lock()
	defer e.inheritanceCacheMutex.Unlock()
	e.inheritanceCache.ClearAll()
}

// InvalidateTemplate removes template from both regular and inheritance caches
func (e *Environment) InvalidateTemplate(templateName string) {
	// Remove from regular template cache
	e.cacheMutex.Lock()
	delete(e.cache, templateName)
	e.cacheMutex.Unlock()

	// Remove from inheritance cache
	e.inheritanceCacheMutex.Lock()
	e.inheritanceCache.InvalidateTemplate(templateName)
	e.inheritanceCacheMutex.Unlock()
}

// GetInheritanceCacheStats returns inheritance cache performance statistics
func (e *Environment) GetInheritanceCacheStats() runtime.CacheStats {
	e.inheritanceCacheMutex.RLock()
	defer e.inheritanceCacheMutex.RUnlock()
	return e.inheritanceCache.GetStats()
}

// ConfigureInheritanceCache allows tuning of inheritance cache settings
func (e *Environment) ConfigureInheritanceCache(hierarchyTTL, resolvedTTL time.Duration, maxEntries int) {
	e.inheritanceCacheMutex.Lock()
	defer e.inheritanceCacheMutex.Unlock()
	e.inheritanceCache.SetHierarchyTTL(hierarchyTTL)
	e.inheritanceCache.SetResolvedTTL(resolvedTTL)
	e.inheritanceCache.SetMaxEntries(maxEntries)
}

// getInheritanceProcessor returns the inheritance processor, initializing it if needed
func (e *Environment) getInheritanceProcessor() *runtime.InheritanceProcessor {
	e.inheritanceCacheMutex.Lock()
	defer e.inheritanceCacheMutex.Unlock()

	if e.inheritanceProcessor == nil {
		// Initialize processor with shared cache (lazy initialization to avoid import cycles)
		e.inheritanceProcessor = runtime.NewInheritanceProcessorWithCache(&environmentAdapter{env: e}, e.inheritanceCache)
	}

	return e.inheritanceProcessor
}

// ListTemplates returns all available template names from the loader
func (e *Environment) ListTemplates() ([]string, error) {
	if e.loader == nil {
		return nil, fmt.Errorf("no loader configured")
	}
	return e.loader.ListTemplates()
}

// RenderString is a convenience method to compile and render a template from string
func (e *Environment) RenderString(source string, context Context) (string, error) {
	template, err := e.FromString(source)
	if err != nil {
		return "", err
	}
	return template.Render(context)
}

// RenderTemplate is a convenience method to load and render a template by name
func (e *Environment) RenderTemplate(name string, context Context) (string, error) {
	template, err := e.GetTemplate(name)
	if err != nil {
		return "", err
	}
	return template.Render(context)
}

// applyBlockTrimming applies block trimming rules to output
func (e *Environment) applyBlockTrimming(output string) string {
	if !e.trimBlocks && !e.lstripBlocks {
		return output
	}

	lines := strings.Split(output, "\n")
	for i, line := range lines {
		if e.lstripBlocks {
			// Remove leading whitespace from lines that start with block tags
			if strings.Contains(line, e.blockStartString) {
				lines[i] = strings.TrimLeft(line, " \t")
			}
		}
		if e.trimBlocks {
			// Remove trailing newline after block tags
			if strings.Contains(line, e.blockEndString) {
				if i < len(lines)-1 && lines[i+1] == "" {
					lines = append(lines[:i+1], lines[i+2:]...)
				}
			}
		}
	}

	return strings.Join(lines, "\n")
}

// loaderAdapter adapts loader.Loader to inheritance.TemplateLoader interface
type loaderAdapter struct {
	loader Loader
}

func (la *loaderAdapter) LoadTemplate(name string) (*parser.TemplateNode, error) {
	// Try advanced loader first
	if advancedLoader, ok := la.loader.(loader.AdvancedLoader); ok {
		return advancedLoader.LoadTemplate(name)
	}

	// Fallback to basic loader - for now just return a simple template
	source, err := la.loader.GetSource(name)
	if err != nil {
		return nil, err
	}

	// Create a simple template node with the source as text
	return &parser.TemplateNode{
		Name:     name,
		Children: []parser.Node{&parser.TextNode{Content: source}},
	}, nil
}

func (la *loaderAdapter) ResolveTemplateName(name string) string {
	// Try advanced loader first
	if advancedLoader, ok := la.loader.(loader.AdvancedLoader); ok {
		return advancedLoader.ResolveTemplateName(name)
	}
	// Simple fallback - just return the name as-is
	return name
}

func registerBuiltinTests(env *Environment) {
	// Will be populated when we implement tests
}

// Global environment instance for convenience functions
var defaultEnv *Environment

// init initializes the default environment
func init() {
	defaultEnv = NewEnvironment()
}

// SetDefaultLoader sets the loader for the default environment
func SetDefaultLoader(loader Loader) {
	defaultEnv.SetLoader(loader)
}

// FromString is a convenience function that uses the default environment
func FromString(source string) (*Template, error) {
	return defaultEnv.FromString(source)
}

// GetTemplate is a convenience function that uses the default environment
func GetTemplate(name string) (*Template, error) {
	return defaultEnv.GetTemplate(name)
}

// RenderString is a convenience function using the default environment
func RenderString(source string, context Context) (string, error) {
	return defaultEnv.RenderString(source, context)
}

// RenderTemplate is a convenience function using the default environment
func RenderTemplate(name string, context Context) (string, error) {
	return defaultEnv.RenderTemplate(name, context)
}

// registerBuiltinGlobals registers all built-in global functions
func registerBuiltinGlobals(env *Environment) {
	// namespace() function
	env.AddGlobal("namespace", namespaceFunction)

	// range() function
	env.AddGlobal("range", rangeFunction)

	// dict() function
	env.AddGlobal("dict", dictFunction)

	// cycler() function
	env.AddGlobal("cycler", cyclerFunction)

	// joiner() function
	env.AddGlobal("joiner", joinerFunction)

	// lipsum() function
	env.AddGlobal("lipsum", lipsumFunction)

	// zip() function
	env.AddGlobal("zip", zipFunction)

	// enumerate() function
	env.AddGlobal("enumerate", enumerateFunction)

	// url_for() function
	env.AddGlobal("url_for", urlForFunction)
}
