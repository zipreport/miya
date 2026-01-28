package macros

import (
	"fmt"
	"sync"

	"github.com/zipreport/miya/parser"
	"github.com/zipreport/miya/runtime"
)

// MacroFunc represents a compiled macro function
type MacroFunc func(ctx runtime.Context, args []interface{}, kwargs map[string]interface{}) (interface{}, error)

// Macro represents a template macro
type Macro struct {
	Name       string
	Parameters []string
	Defaults   map[string]interface{}
	Body       []parser.Node
	Template   string // Template name where macro is defined
}

// MacroRegistry manages template macros
type MacroRegistry struct {
	macros map[string]*Macro
	mutex  sync.RWMutex
}

// NewMacroRegistry creates a new macro registry
func NewMacroRegistry() *MacroRegistry {
	return &MacroRegistry{
		macros: make(map[string]*Macro),
	}
}

// Register registers a macro in the registry
func (r *MacroRegistry) Register(name string, macro *Macro) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.macros[name]; exists {
		return fmt.Errorf("macro %q already registered", name)
	}

	r.macros[name] = macro
	return nil
}

// Get retrieves a macro by name
func (r *MacroRegistry) Get(name string) (*Macro, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	macro, ok := r.macros[name]
	return macro, ok
}

// List returns all registered macro names
func (r *MacroRegistry) List() []string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	names := make([]string, 0, len(r.macros))
	for name := range r.macros {
		names = append(names, name)
	}
	return names
}

// Clear removes all macros from the registry
func (r *MacroRegistry) Clear() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.macros = make(map[string]*Macro)
}

// Import imports macros from another template
func (r *MacroRegistry) Import(fromTemplate string, other *MacroRegistry, names []string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	other.mutex.RLock()
	defer other.mutex.RUnlock()

	if len(names) == 0 {
		// Import all macros
		for name, macro := range other.macros {
			if macro.Template == fromTemplate {
				r.macros[name] = macro
			}
		}
	} else {
		// Import specific macros
		for _, name := range names {
			if macro, exists := other.macros[name]; exists && macro.Template == fromTemplate {
				r.macros[name] = macro
			} else {
				return fmt.Errorf("macro %q not found in template %s", name, fromTemplate)
			}
		}
	}

	return nil
}

// MacroExecutor handles macro execution
type MacroExecutor struct {
	evaluator runtime.Evaluator
}

// NewMacroExecutor creates a new macro executor
func NewMacroExecutor(evaluator runtime.Evaluator) *MacroExecutor {
	return &MacroExecutor{
		evaluator: evaluator,
	}
}

// Execute executes a macro with the given arguments
func (e *MacroExecutor) Execute(macro *Macro, ctx runtime.Context, args []interface{}, kwargs map[string]interface{}) (interface{}, error) {
	// Create a new context for macro execution
	macroCtx := ctx.Clone()

	// First, bind keyword arguments (they override positional args)
	for key, value := range kwargs {
		// Check if parameter exists
		found := false
		for _, param := range macro.Parameters {
			if param == key {
				macroCtx.SetVariable(key, value)
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("unknown parameter %q for macro %q", key, macro.Name)
		}
	}

	// Then bind positional parameters (only if not already set by kwargs)
	for i, param := range macro.Parameters {
		if _, alreadySet := macroCtx.GetVariable(param); alreadySet {
			continue // Skip if already set by keyword argument
		}

		if i < len(args) {
			macroCtx.SetVariable(param, args[i])
		} else if defaultValue, hasDefault := macro.Defaults[param]; hasDefault {
			macroCtx.SetVariable(param, defaultValue)
		} else {
			return nil, fmt.Errorf("missing required parameter %q for macro %q", param, macro.Name)
		}
	}

	// Execute macro body
	result := ""
	for _, node := range macro.Body {
		nodeResult, err := e.evaluator.EvalNode(node, macroCtx)
		if err != nil {
			return nil, fmt.Errorf("error executing macro %q: %v", macro.Name, err)
		}
		result += fmt.Sprintf("%v", nodeResult)
	}

	return result, nil
}

// CallMacro is a convenience function to call a macro
func (r *MacroRegistry) CallMacro(name string, ctx runtime.Context, evaluator runtime.Evaluator, args []interface{}, kwargs map[string]interface{}) (interface{}, error) {
	macro, ok := r.Get(name)
	if !ok {
		return nil, fmt.Errorf("macro %q not found", name)
	}

	executor := NewMacroExecutor(evaluator)
	return executor.Execute(macro, ctx, args, kwargs)
}

// MacroContext provides macro-related context operations
type MacroContext struct {
	registry *MacroRegistry
}

// NewMacroContext creates a new macro context
func NewMacroContext() *MacroContext {
	return &MacroContext{
		registry: NewMacroRegistry(),
	}
}

// GetRegistry returns the macro registry
func (mc *MacroContext) GetRegistry() *MacroRegistry {
	return mc.registry
}

// DefineMacro defines a macro from an AST node
func (mc *MacroContext) DefineMacro(node *parser.MacroNode, templateName string) error {
	// Convert defaults from ExpressionNode to interface{}
	defaults := make(map[string]interface{})
	for key, expr := range node.Defaults {
		// For now, we'll store the expression itself
		// In a full implementation, we'd evaluate constant expressions
		defaults[key] = expr
	}

	macro := &Macro{
		Name:       node.Name,
		Parameters: node.Parameters,
		Defaults:   defaults,
		Body:       node.Body,
		Template:   templateName,
	}

	return mc.registry.Register(node.Name, macro)
}

// Macro call helper for templates
func (mc *MacroContext) Call(name string, ctx runtime.Context, evaluator runtime.Evaluator, args []interface{}, kwargs map[string]interface{}) (interface{}, error) {
	return mc.registry.CallMacro(name, ctx, evaluator, args, kwargs)
}
