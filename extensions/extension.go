package extensions

import (
	"fmt"

	"github.com/zipreport/miya/lexer"
	"github.com/zipreport/miya/parser"
	"github.com/zipreport/miya/runtime"
)

// ExtensionError provides context-aware error messages for extensions
type ExtensionError struct {
	ExtensionName string
	TagName       string
	TemplateName  string
	Line          int
	Column        int
	Message       string
	Cause         error
}

func (ee *ExtensionError) Error() string {
	if ee.TemplateName != "" && ee.Line > 0 {
		return fmt.Sprintf("extension '%s' error in template '%s' at line %d:%d: %s",
			ee.ExtensionName, ee.TemplateName, ee.Line, ee.Column, ee.Message)
	} else if ee.TemplateName != "" {
		return fmt.Sprintf("extension '%s' error in template '%s': %s",
			ee.ExtensionName, ee.TemplateName, ee.Message)
	} else if ee.TagName != "" && ee.Line > 0 {
		return fmt.Sprintf("extension '%s' error in tag '%s' at line %d:%d: %s",
			ee.ExtensionName, ee.TagName, ee.Line, ee.Column, ee.Message)
	} else if ee.TagName != "" {
		return fmt.Sprintf("extension '%s' error in tag '%s': %s",
			ee.ExtensionName, ee.TagName, ee.Message)
	}
	return fmt.Sprintf("extension '%s' error: %s", ee.ExtensionName, ee.Message)
}

func (ee *ExtensionError) Unwrap() error {
	return ee.Cause
}

// NewExtensionError creates a new extension error with context
func NewExtensionError(extensionName, message string, cause error) *ExtensionError {
	return &ExtensionError{
		ExtensionName: extensionName,
		Message:       message,
		Cause:         cause,
	}
}

// NewExtensionTagError creates a new extension error with tag context
func NewExtensionTagError(extensionName, tagName, message string, cause error) *ExtensionError {
	return &ExtensionError{
		ExtensionName: extensionName,
		TagName:       tagName,
		Message:       message,
		Cause:         cause,
	}
}

// NewExtensionParseError creates a new extension error with parse context
func NewExtensionParseError(extensionName, tagName, templateName string, line, column int, message string, cause error) *ExtensionError {
	return &ExtensionError{
		ExtensionName: extensionName,
		TagName:       tagName,
		TemplateName:  templateName,
		Line:          line,
		Column:        column,
		Message:       message,
		Cause:         cause,
	}
}

// Extension defines the interface for template extensions
type Extension interface {
	// Name returns the unique name of this extension
	Name() string

	// Tags returns a list of tag names this extension handles
	Tags() []string

	// ParseTag is called when one of the extension's tags is encountered during parsing
	// It should parse the tag and return an ExtensionNode
	ParseTag(tagName string, parser ExtensionParser) (parser.Node, error)

	// IsBlockExtension returns true if this extension handles block tags (with end tags)
	IsBlockExtension(tagName string) bool

	// GetEndTag returns the corresponding end tag name for a block tag
	// For example, "highlight" -> "endhighlight"
	GetEndTag(tagName string) string

	// Configure sets configuration options for the extension
	Configure(config map[string]interface{}) error

	// GetConfig returns the current configuration
	GetConfig() map[string]interface{}

	// Lifecycle hooks
	// OnLoad is called when the extension is registered
	OnLoad(env ExtensionEnvironment) error

	// BeforeRender is called before template rendering starts
	BeforeRender(ctx ExtensionContext, templateName string) error

	// AfterRender is called after template rendering completes
	AfterRender(ctx ExtensionContext, templateName string, result interface{}, err error) error

	// OnUnload is called when the extension is being removed (cleanup)
	OnUnload() error

	// Dependencies returns a list of extension names this extension depends on
	Dependencies() []string
}

// ExtensionParser provides parsing utilities for extensions
type ExtensionParser interface {
	// Current returns the current token
	Current() *lexer.Token

	// Advance moves to the next token and returns it
	Advance() *lexer.Token

	// Check returns true if the current token matches the given type
	Check(tokenType lexer.TokenType) bool

	// CheckAny returns true if the current token matches any of the given types
	CheckAny(types ...lexer.TokenType) bool

	// Peek returns the current token without advancing
	Peek() *lexer.Token

	// IsAtEnd returns true if we've reached the end of tokens
	IsAtEnd() bool

	// ParseExpression parses an expression and returns the node
	ParseExpression() (parser.ExpressionNode, error)

	// ParseToEnd parses all tokens until the specified end tag
	ParseToEnd(endTag string) ([]parser.Node, error)

	// ParseBlock parses a block extension body until the end tag
	ParseBlock(endTag string) ([]parser.Node, error)

	// Error creates a parser error with the given message
	Error(message string) error

	// ExpectBlockEnd ensures the current token is a block end (%})
	ExpectBlockEnd() error

	// ExpectEndTag consumes the end tag block
	ExpectEndTag(endTag string) error

	// NewExtensionNode creates a new extension node
	NewExtensionNode(extensionName, tagName string, line, column int) *parser.ExtensionNode

	// ParseArguments parses arguments until block end
	ParseArguments() ([]parser.ExpressionNode, error)
}

// ExtensionContext provides context access for extensions during evaluation
type ExtensionContext interface {
	// GetVariable retrieves a variable from the context
	GetVariable(name string) (interface{}, bool)

	// SetVariable sets a variable in the context
	SetVariable(name string, value interface{})

	// GetGlobal retrieves a global variable
	GetGlobal(name string) (interface{}, bool)

	// ApplyFilter applies a filter to a value
	ApplyFilter(name string, value interface{}, args ...interface{}) (interface{}, error)

	// CallMacro calls a macro
	CallMacro(name string, args []interface{}, kwargs map[string]interface{}) (interface{}, error)
}

// ExtensionEnvironment provides access to the template environment for extensions
type ExtensionEnvironment interface {
	// GetExtension retrieves another extension by name
	GetExtension(name string) (Extension, bool)

	// AddGlobal adds a global variable
	AddGlobal(name string, value interface{})

	// AddFilter adds a custom filter
	AddFilter(name string, filter interface{}) error

	// AddTest adds a custom test
	AddTest(name string, test interface{}) error

	// GetConfig gets environment-level configuration
	GetConfig(key string) (interface{}, bool)

	// SetConfig sets environment-level configuration
	SetConfig(key string, value interface{})
}

// Registry manages all registered extensions
type Registry struct {
	extensions   map[string]Extension
	tagMap       map[string]Extension // Maps tag names to extensions
	loadOrder    []string             // Order in which extensions were loaded
	dependencies map[string][]string  // Extension dependencies
	environment  ExtensionEnvironment // Reference to environment
}

// NewRegistry creates a new extension registry
func NewRegistry() *Registry {
	return &Registry{
		extensions:   make(map[string]Extension),
		tagMap:       make(map[string]Extension),
		loadOrder:    make([]string, 0),
		dependencies: make(map[string][]string),
	}
}

// SetEnvironment sets the environment reference for the registry
func (r *Registry) SetEnvironment(env ExtensionEnvironment) {
	r.environment = env
}

// Register registers an extension with dependency checking and lifecycle hooks
func (r *Registry) Register(extension Extension) error {
	name := extension.Name()
	if _, exists := r.extensions[name]; exists {
		return NewExtensionError(name, "extension is already registered", nil)
	}

	// Check dependencies first
	deps := extension.Dependencies()
	for _, dep := range deps {
		if _, exists := r.extensions[dep]; !exists {
			return NewExtensionError(name, fmt.Sprintf("depends on '%s' which is not registered", dep), nil)
		}
	}

	// Check for tag name conflicts
	for _, tag := range extension.Tags() {
		if existingExt, exists := r.tagMap[tag]; exists {
			return NewExtensionTagError(name, tag, fmt.Sprintf("tag is already handled by extension '%s'", existingExt.Name()), nil)
		}
	}

	// Check for circular dependencies
	if err := r.checkCircularDependencies(name, deps); err != nil {
		return NewExtensionError(name, "circular dependency detected", err)
	}

	// Register the extension
	r.extensions[name] = extension
	r.dependencies[name] = deps
	r.loadOrder = append(r.loadOrder, name)

	for _, tag := range extension.Tags() {
		r.tagMap[tag] = extension
	}

	// Call OnLoad lifecycle hook
	if r.environment != nil {
		if err := extension.OnLoad(r.environment); err != nil {
			// Rollback registration on error
			delete(r.extensions, name)
			delete(r.dependencies, name)
			r.loadOrder = r.loadOrder[:len(r.loadOrder)-1]
			for _, tag := range extension.Tags() {
				delete(r.tagMap, tag)
			}
			return NewExtensionError(name, "OnLoad lifecycle hook failed", err)
		}
	}

	return nil
}

// checkCircularDependencies checks for circular dependencies
func (r *Registry) checkCircularDependencies(newExtName string, newDeps []string) error {
	// Build dependency graph including the new extension
	graph := make(map[string][]string)
	for ext, deps := range r.dependencies {
		graph[ext] = deps
	}
	graph[newExtName] = newDeps

	// Use DFS to detect cycles
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var hasCycle func(string) bool
	hasCycle = func(node string) bool {
		visited[node] = true
		recStack[node] = true

		for _, dep := range graph[node] {
			if !visited[dep] && hasCycle(dep) {
				return true
			} else if recStack[dep] {
				return true
			}
		}

		recStack[node] = false
		return false
	}

	for ext := range graph {
		if !visited[ext] && hasCycle(ext) {
			return fmt.Errorf("circular dependency detected involving extension '%s'", newExtName)
		}
	}

	return nil
}

// GetExtensionForTag returns the extension that handles the given tag
func (r *Registry) GetExtensionForTag(tagName string) (Extension, bool) {
	ext, ok := r.tagMap[tagName]
	return ext, ok
}

// GetExtension returns the extension with the given name
func (r *Registry) GetExtension(name string) (Extension, bool) {
	ext, ok := r.extensions[name]
	return ext, ok
}

// GetAllExtensions returns all registered extensions
func (r *Registry) GetAllExtensions() []Extension {
	extensions := make([]Extension, 0, len(r.extensions))
	for _, ext := range r.extensions {
		extensions = append(extensions, ext)
	}
	return extensions
}

// IsCustomTag returns true if the tag is handled by an extension
func (r *Registry) IsCustomTag(tagName string) bool {
	_, ok := r.tagMap[tagName]
	return ok
}

// Unregister removes an extension and calls its OnUnload hook
func (r *Registry) Unregister(name string) error {
	ext, exists := r.extensions[name]
	if !exists {
		return NewExtensionError(name, "extension is not registered", nil)
	}

	// Check if other extensions depend on this one
	for extName, deps := range r.dependencies {
		for _, dep := range deps {
			if dep == name {
				return NewExtensionError(name, fmt.Sprintf("cannot unregister: extension '%s' depends on it", extName), nil)
			}
		}
	}

	// Call OnUnload lifecycle hook
	if err := ext.OnUnload(); err != nil {
		return NewExtensionError(name, "OnUnload lifecycle hook failed", err)
	}

	// Remove from registry
	delete(r.extensions, name)
	delete(r.dependencies, name)

	// Remove from load order
	for i, loadedName := range r.loadOrder {
		if loadedName == name {
			r.loadOrder = append(r.loadOrder[:i], r.loadOrder[i+1:]...)
			break
		}
	}

	// Remove tag mappings
	for _, tag := range ext.Tags() {
		delete(r.tagMap, tag)
	}

	return nil
}

// BeforeRender calls BeforeRender on all registered extensions
func (r *Registry) BeforeRender(ctx ExtensionContext, templateName string) error {
	// Call hooks in load order to respect dependencies
	for _, extName := range r.loadOrder {
		if ext, exists := r.extensions[extName]; exists {
			if err := ext.BeforeRender(ctx, templateName); err != nil {
				return &ExtensionError{
					ExtensionName: extName,
					TemplateName:  templateName,
					Message:       "BeforeRender lifecycle hook failed",
					Cause:         err,
				}
			}
		}
	}
	return nil
}

// AfterRender calls AfterRender on all registered extensions
func (r *Registry) AfterRender(ctx ExtensionContext, templateName string, result interface{}, renderErr error) error {
	// Call hooks in reverse load order for cleanup
	for i := len(r.loadOrder) - 1; i >= 0; i-- {
		extName := r.loadOrder[i]
		if ext, exists := r.extensions[extName]; exists {
			if err := ext.AfterRender(ctx, templateName, result, renderErr); err != nil {
				return &ExtensionError{
					ExtensionName: extName,
					TemplateName:  templateName,
					Message:       "AfterRender lifecycle hook failed",
					Cause:         err,
				}
			}
		}
	}
	return nil
}

// GetLoadOrder returns the order in which extensions were loaded
func (r *Registry) GetLoadOrder() []string {
	result := make([]string, len(r.loadOrder))
	copy(result, r.loadOrder)
	return result
}

// GetDependencies returns the dependencies for a given extension
func (r *Registry) GetDependencies(name string) []string {
	if deps, exists := r.dependencies[name]; exists {
		result := make([]string, len(deps))
		copy(result, deps)
		return result
	}
	return nil
}

// BaseExtension provides a base implementation for common extension functionality
type BaseExtension struct {
	name      string
	tags      []string
	blockTags map[string]string      // maps start tag -> end tag
	config    map[string]interface{} // extension configuration
}

// NewBaseExtension creates a new base extension
func NewBaseExtension(name string, tags []string) *BaseExtension {
	return &BaseExtension{
		name:      name,
		tags:      tags,
		blockTags: make(map[string]string),
		config:    make(map[string]interface{}),
	}
}

// NewBlockExtension creates a new base extension with block tag support
func NewBlockExtension(name string, blockTags map[string]string) *BaseExtension {
	// Extract all tag names (both start and end tags)
	tags := make([]string, 0, len(blockTags)*2)
	for startTag, endTag := range blockTags {
		tags = append(tags, startTag, endTag)
	}

	return &BaseExtension{
		name:      name,
		tags:      tags,
		blockTags: blockTags,
		config:    make(map[string]interface{}),
	}
}

// Name returns the extension name
func (e *BaseExtension) Name() string {
	return e.name
}

// Tags returns the handled tags
func (e *BaseExtension) Tags() []string {
	return e.tags
}

// IsBlockExtension returns true if this extension handles block tags
func (e *BaseExtension) IsBlockExtension(tagName string) bool {
	_, isBlockTag := e.blockTags[tagName]
	return isBlockTag
}

// GetEndTag returns the corresponding end tag name for a block tag
func (e *BaseExtension) GetEndTag(tagName string) string {
	if endTag, exists := e.blockTags[tagName]; exists {
		return endTag
	}
	return ""
}

// Configure sets configuration options for the extension
func (e *BaseExtension) Configure(config map[string]interface{}) error {
	for key, value := range config {
		e.config[key] = value
	}
	return nil
}

// GetConfig returns the current configuration
func (e *BaseExtension) GetConfig() map[string]interface{} {
	result := make(map[string]interface{})
	for key, value := range e.config {
		result[key] = value
	}
	return result
}

// GetConfigValue returns a specific configuration value
func (e *BaseExtension) GetConfigValue(key string) (interface{}, bool) {
	value, exists := e.config[key]
	return value, exists
}

// SetConfigValue sets a specific configuration value
func (e *BaseExtension) SetConfigValue(key string, value interface{}) {
	e.config[key] = value
}

// Lifecycle hook implementations (default no-op implementations)
// OnLoad is called when the extension is registered
func (e *BaseExtension) OnLoad(env ExtensionEnvironment) error {
	return nil // Default implementation does nothing
}

// BeforeRender is called before template rendering starts
func (e *BaseExtension) BeforeRender(ctx ExtensionContext, templateName string) error {
	return nil // Default implementation does nothing
}

// AfterRender is called after template rendering completes
func (e *BaseExtension) AfterRender(ctx ExtensionContext, templateName string, result interface{}, err error) error {
	return nil // Default implementation does nothing
}

// OnUnload is called when the extension is being removed (cleanup)
func (e *BaseExtension) OnUnload() error {
	return nil // Default implementation does nothing
}

// Dependencies returns a list of extension names this extension depends on
func (e *BaseExtension) Dependencies() []string {
	return nil // Default implementation has no dependencies
}

// CreateEvaluateFunc creates an evaluation function for extension nodes
func CreateEvaluateFunc(evalFunc func(*parser.ExtensionNode, runtime.Context) (interface{}, error)) func(*parser.ExtensionNode, interface{}) (interface{}, error) {
	return func(node *parser.ExtensionNode, ctx interface{}) (interface{}, error) {
		runtimeCtx, ok := ctx.(runtime.Context)
		if !ok {
			return nil, fmt.Errorf("invalid context type for extension evaluation")
		}
		return evalFunc(node, runtimeCtx)
	}
}

// CreateAdvancedEvaluateFunc creates an evaluation function with enhanced context access
func CreateAdvancedEvaluateFunc(evalFunc func(*parser.ExtensionNode, ExtensionContext) (interface{}, error)) func(*parser.ExtensionNode, interface{}) (interface{}, error) {
	return func(node *parser.ExtensionNode, ctx interface{}) (interface{}, error) {
		runtimeCtx, ok := ctx.(runtime.Context)
		if !ok {
			return nil, fmt.Errorf("invalid context type for extension evaluation")
		}

		// Create extension context adapter
		extCtx := &extensionContextAdapter{ctx: runtimeCtx}
		return evalFunc(node, extCtx)
	}
}

// extensionContextAdapter adapts runtime.Context to ExtensionContext
type extensionContextAdapter struct {
	ctx runtime.Context
}

func (eca *extensionContextAdapter) GetVariable(name string) (interface{}, bool) {
	return eca.ctx.GetVariable(name)
}

func (eca *extensionContextAdapter) SetVariable(name string, value interface{}) {
	eca.ctx.SetVariable(name, value)
}

func (eca *extensionContextAdapter) GetGlobal(name string) (interface{}, bool) {
	// Try to get global through environment if available
	if envCtx, ok := eca.ctx.(interface {
		GetGlobal(string) (interface{}, bool)
	}); ok {
		return envCtx.GetGlobal(name)
	}
	return eca.ctx.GetVariable(name) // Fallback to regular variable lookup
}

func (eca *extensionContextAdapter) ApplyFilter(name string, value interface{}, args ...interface{}) (interface{}, error) {
	// Try to apply filter through environment if available
	if envCtx, ok := eca.ctx.(interface {
		ApplyFilter(string, interface{}, ...interface{}) (interface{}, error)
	}); ok {
		return envCtx.ApplyFilter(name, value, args...)
	}
	return nil, fmt.Errorf("filter support not available in this context")
}

func (eca *extensionContextAdapter) CallMacro(name string, args []interface{}, kwargs map[string]interface{}) (interface{}, error) {
	// Try to call macro through environment if available
	if envCtx, ok := eca.ctx.(interface {
		CallMacro(string, []interface{}, map[string]interface{}) (interface{}, error)
	}); ok {
		return envCtx.CallMacro(name, args, kwargs)
	}
	return nil, fmt.Errorf("macro support not available in this context")
}
