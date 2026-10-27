package runtime

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/zipreport/miya/parser"
)

// TemplateNamespace represents an imported template's namespace
type TemplateNamespace struct {
	TemplateName string
	Macros       map[string]*TemplateMacro
	Variables    map[string]interface{}
	Context      Context
	evaluator    *DefaultEvaluator // Store evaluator for macro calls
}

// ImportedNamespace is a wrapper for imported templates that supports attribute access
type ImportedNamespace struct {
	namespace *TemplateNamespace
	evaluator *DefaultEvaluator
}

// Get returns an attribute from the namespace (used for macro/variable access).
// A list is returned as a plain slice; RawGet returns the stored value.
func (in *ImportedNamespace) Get(name string) (interface{}, bool) {
	v, ok := in.RawGet(name)
	return Unbox(v), ok
}

// RawGet returns the stored value for the runtime, which can be a list box
// or escaped template output; Go code should use Get.
func (in *ImportedNamespace) RawGet(name string) (interface{}, bool) {
	// Check if it's a macro
	if macro, ok := in.namespace.Macros[name]; ok {
		return macro.bind(in.evaluator), true
	}

	// Check if it's a variable
	if value, ok := in.namespace.Variables[name]; ok {
		return value, true
	}

	// Check special attributes
	if name == "__template__" {
		return in.namespace.TemplateName, true
	}
	if name == "__imported__" {
		return true, true
	}

	return nil, false
}

// Set sets a value in the namespace (implements NamespaceInterface)
func (in *ImportedNamespace) Set(name string, value interface{}) {
	// Allow setting variables in the namespace
	in.namespace.Variables[name] = value
}

// String returns a string representation of the namespace
func (in *ImportedNamespace) String() string {
	return fmt.Sprintf("<ImportedNamespace from '%s' with %d macros, %d variables>",
		in.namespace.TemplateName, len(in.namespace.Macros), len(in.namespace.Variables))
}

// TemplateMacro represents a macro that can be called from templates
type TemplateMacro struct {
	Name       string
	Parameters []string
	Defaults   map[string]interface{}
	Body       []parser.Node
	Context    Context // The context in which the macro was defined
	parser.MacroUsage
}

// bind returns the macro as a callable bound to its definition context.
func (tm *TemplateMacro) bind(evaluator *DefaultEvaluator) *Macro {
	return &Macro{
		Name:       tm.Name,
		Parameters: tm.Parameters,
		Defaults:   tm.Defaults,
		Body:       tm.Body,
		MacroUsage: tm.MacroUsage,
		defCtx:     tm.Context,
		evaluator:  evaluator,
	}
}

// Call executes the macro with the given arguments. callCtx is not used: the
// body runs in the definition context, and a caller is passed in kwargs["caller"].
func (tm *TemplateMacro) Call(evaluator *DefaultEvaluator, callCtx Context, args []interface{}, kwargs map[string]interface{}) (interface{}, error) {
	return tm.bind(evaluator).Call(args, kwargs)
}

// TemplateLoader interface for loading templates
type TemplateLoader interface {
	LoadTemplate(name string) (*parser.TemplateNode, error)
	TemplateExists(name string) bool
}

// ImportSystem handles template imports and namespace management.
// It keeps no per-template state: the evaluator decides when a module is
// reused (see DefaultEvaluator.loadModule), so one ImportSystem can serve
// concurrent renders.
type ImportSystem struct {
	loader TemplateLoader
}

// NewImportSystem creates a new import system
func NewImportSystem(loader TemplateLoader, evaluator *DefaultEvaluator) *ImportSystem {
	// Note: evaluator parameter kept for API compatibility but no longer stored
	return &ImportSystem{
		loader: loader,
	}
}

// LoadTemplateNamespace loads a template and builds its namespace. moduleCtx
// becomes the module's context and is written to (top-level sets and the
// file's own imports land in it), so pass a context you own. The loader must
// return an error for a template that does not exist; TemplateExists is not
// consulted. The module's top-level code runs, so an {% include %} in it calls
// TemplateExists and needs the evaluator's import system: a bare
// NewEvaluator() without one fails on a module that includes.
func (is *ImportSystem) LoadTemplateNamespace(templateName string, moduleCtx Context, evaluator *DefaultEvaluator) (*TemplateNamespace, error) {
	ast, err := is.loader.LoadTemplate(templateName)
	if err != nil {
		return nil, fmt.Errorf("failed to load template %q: %w", templateName, err)
	}
	if ast == nil {
		return nil, fmt.Errorf("failed to load template %q: the loader returned no template", templateName)
	}

	namespace := &TemplateNamespace{
		TemplateName: templateName,
		Macros:       make(map[string]*TemplateMacro),
		Variables:    make(map[string]interface{}),
		Context:      moduleCtx,
	}

	// Extract macros and variables from AST
	err = is.extractNamespaceContent(ast, namespace, evaluator)
	if err != nil {
		return nil, fmt.Errorf("failed to extract namespace from template %q: %w", templateName, err)
	}

	return namespace, nil
}

// extractNamespaceContent runs the template's top-level code in the module
// context, as Jinja2 does when it builds a module: macros are registered, the
// taken branch of an if runs, and every other statement is evaluated with its
// output discarded. The names the top level assigned are then exported.
func (is *ImportSystem) extractNamespaceContent(node parser.Node, namespace *TemplateNamespace, evaluator *DefaultEvaluator) error {
	tmpl, ok := node.(*parser.TemplateNode)
	if !ok {
		return nil
	}
	nodes := tmpl.Children
	for _, child := range nodes {
		if _, ok := child.(*parser.ExtendsNode); ok {
			// Only the child's own top-level code runs; the parent's does not.
			nodes = topLevelCode(tmpl)
			break
		}
	}
	assigned := make(map[string]bool)
	if err := is.runModuleNodes(nodes, namespace, evaluator, assigned); err != nil {
		return err
	}
	for name := range assigned {
		if strings.HasPrefix(name, "_") {
			continue
		}
		if value, ok := namespace.Context.GetVariable(name); ok {
			namespace.Variables[name] = value
		}
	}
	return nil
}

func (is *ImportSystem) runModuleNodes(nodes []parser.Node, namespace *TemplateNamespace, evaluator *DefaultEvaluator, assigned map[string]bool) error {
	for _, child := range nodes {
		switch n := child.(type) {
		case *parser.MacroNode:
			// Convert defaults from parser expressions to evaluated values
			defaults := make(map[string]interface{})
			for key, expr := range n.Defaults {
				// For now, store the expression itself - in a full implementation
				// we'd evaluate constant expressions
				defaults[key] = expr
			}

			macro := &TemplateMacro{
				Name:       n.Name,
				Parameters: n.Parameters,
				Defaults:   defaults,
				Body:       n.Body,
				Context:    namespace.Context,
				MacroUsage: n.MacroUsage,
			}

			if !strings.HasPrefix(n.Name, "_") {
				namespace.Macros[n.Name] = macro
			}
			// Register the macro in its own module so it can call itself and its siblings.
			namespace.Context.SetVariable(n.Name, macro.bind(evaluator))

		case *parser.SetNode:
			if _, err := evaluator.EvalNode(n, namespace.Context); err != nil {
				names := make([]string, 0, len(n.Targets))
				for _, t := range n.Targets {
					switch t := t.(type) {
					case *parser.IdentifierNode:
						names = append(names, t.Name)
						assigned[t.Name] = true
					case *parser.AttributeNode:
						if o, ok := t.Object.(*parser.IdentifierNode); ok {
							names = append(names, o.Name+"."+t.Attribute)
						}
					}
				}
				return fmt.Errorf("{%% set %s %%}: %w", strings.Join(names, ", "), err)
			}
			for _, t := range n.Targets {
				if id, ok := t.(*parser.IdentifierNode); ok {
					assigned[id.Name] = true
				}
			}

		case *parser.BlockSetNode:
			if _, err := evaluator.EvalNode(n, namespace.Context); err != nil {
				return fmt.Errorf("{%% set %s %%}: %w", n.Variable, err)
			}
			assigned[n.Variable] = true

		case *parser.ImportNode:
			// The file's own imports run in its module context, so its macros can
			// call them. As in Jinja2, they are not exported.
			if _, err := evaluator.EvalNode(n, namespace.Context); err != nil {
				return err
			}
			delete(assigned, n.Alias)

		case *parser.FromNode:
			if _, err := evaluator.EvalNode(n, namespace.Context); err != nil {
				return err
			}
			for _, name := range n.Names {
				if alias, ok := n.Aliases[name]; ok {
					name = alias
				}
				delete(assigned, name)
			}

		case *parser.BlockNode:
			// A block is its own scope: its sets and macros are not exported.
			if _, err := evaluator.EvalNode(n, namespace.Context.Clone()); err != nil {
				return fmt.Errorf("import: %w", err)
			}

		case *parser.IfNode:
			body, err := is.chooseBranch(n, namespace, evaluator)
			if err != nil {
				return fmt.Errorf("import: %w", err)
			}
			if err := is.runModuleNodes(body, namespace, evaluator, assigned); err != nil {
				return err
			}

		default:
			if _, err := evaluator.EvalNode(child, namespace.Context); err != nil {
				return fmt.Errorf("import: %w", err)
			}
		}
	}
	return nil
}

// chooseBranch returns the body of the if/elif/else branch that is taken.
func (is *ImportSystem) chooseBranch(n *parser.IfNode, namespace *TemplateNamespace, evaluator *DefaultEvaluator) ([]parser.Node, error) {
	cond, err := evaluator.EvalNode(n.Condition, namespace.Context)
	if err != nil {
		return nil, err
	}
	if evaluator.isTruthy(cond) {
		return n.Body, nil
	}
	for _, elif := range n.ElseIfs {
		cond, err := evaluator.EvalNode(elif.Condition, namespace.Context)
		if err != nil {
			return nil, err
		}
		if evaluator.isTruthy(cond) {
			return elif.Body, nil
		}
	}
	return n.Else, nil
}

// GetImportedNamespace returns an ImportedNamespace wrapper for the namespace
func (is *ImportSystem) GetImportedNamespace(namespace *TemplateNamespace, evaluator *DefaultEvaluator) *ImportedNamespace {
	// Set the evaluator in the namespace
	namespace.evaluator = evaluator

	return &ImportedNamespace{
		namespace: namespace,
		evaluator: evaluator,
	}
}

// GetNamespaceMap returns a map representation of the namespace for template use
func (is *ImportSystem) GetNamespaceMap(namespace *TemplateNamespace, evaluator *DefaultEvaluator) map[string]interface{} {
	result := make(map[string]interface{})

	// Add macros as callable functions
	for name, macro := range namespace.Macros {
		result[name] = macro.bind(evaluator)
	}

	// Add variables
	for name, value := range namespace.Variables {
		result[name] = value
	}

	// Add template metadata
	result["__template__"] = namespace.TemplateName
	result["__imported__"] = true

	return result
}

// SimpleTemplateLoader implements TemplateLoader for basic file loading
type SimpleTemplateLoader struct {
	environment interface{} // Reference to environment for template loading
}

// NewSimpleTemplateLoader creates a new simple template loader
func NewSimpleTemplateLoader(env interface{}) *SimpleTemplateLoader {
	return &SimpleTemplateLoader{environment: env}
}

// LoadTemplate loads a template AST by name
func (stl *SimpleTemplateLoader) LoadTemplate(name string) (*parser.TemplateNode, error) {
	// Try to use GetTemplate method to get the already-parsed template
	envValue := reflect.ValueOf(stl.environment)
	if !envValue.IsValid() {
		return nil, fmt.Errorf("invalid environment reference")
	}

	// Try to get the GetTemplate method directly
	getTemplateMethod := envValue.MethodByName("GetTemplate")
	if getTemplateMethod.IsValid() {
		// Call GetTemplate(name) to get the parsed template
		results := getTemplateMethod.Call([]reflect.Value{reflect.ValueOf(name)})
		if len(results) == 2 {
			// Check for error first
			if err := results[1].Interface(); err != nil {
				if e, ok := err.(error); ok {
					return nil, e
				}
			}

			// Get the template object
			if template := results[0].Interface(); template != nil {
				// Try to get the AST from the template
				templateValue := reflect.ValueOf(template)
				astMethod := templateValue.MethodByName("AST")
				if astMethod.IsValid() {
					astResults := astMethod.Call([]reflect.Value{})
					if len(astResults) == 1 {
						if ast, ok := astResults[0].Interface().(*parser.TemplateNode); ok {
							return ast, nil
						}
					}
				}
			}
		}
	}

	// Fallback: Try to use loader directly if available
	getLoaderMethod := envValue.MethodByName("GetLoader")
	if !getLoaderMethod.IsValid() {
		return nil, fmt.Errorf("environment does not have GetLoader method")
	}

	// Call GetLoader() to get the loader
	loaderResults := getLoaderMethod.Call([]reflect.Value{})
	if len(loaderResults) != 1 {
		return nil, fmt.Errorf("GetLoader returned unexpected number of values")
	}

	loaderValue := loaderResults[0]
	if !loaderValue.IsValid() || loaderValue.IsNil() {
		return nil, fmt.Errorf("environment loader is nil")
	}

	// Check if it's an AdvancedLoader with LoadTemplate method
	loadTemplateMethod := loaderValue.MethodByName("LoadTemplate")
	if loadTemplateMethod.IsValid() {
		results := loadTemplateMethod.Call([]reflect.Value{reflect.ValueOf(name)})
		if len(results) == 2 {
			// Check for error
			if err := results[1].Interface(); err != nil {
				if e, ok := err.(error); ok {
					return nil, e
				}
			}

			// Get the AST
			if ast, ok := results[0].Interface().(*parser.TemplateNode); ok {
				return ast, nil
			}
		}
	}

	// Last resort: Get source and return empty template
	getSourceMethod := loaderValue.MethodByName("GetSource")
	if !getSourceMethod.IsValid() {
		return nil, fmt.Errorf("loader does not have GetSource method")
	}

	// Call GetSource to get template content
	results := getSourceMethod.Call([]reflect.Value{reflect.ValueOf(name)})
	if len(results) != 2 {
		return nil, fmt.Errorf("GetSource returned unexpected number of values")
	}

	// Check for error
	if err := results[1].Interface(); err != nil {
		if e, ok := err.(error); ok {
			return nil, e
		}
	}

	// For now, return an empty template node
	// In a real implementation, we'd need to parse the source
	template := &parser.TemplateNode{
		Name:     name,
		Children: []parser.Node{},
	}

	return template, nil
}

// TemplateExists checks if a template exists
func (stl *SimpleTemplateLoader) TemplateExists(name string) bool {
	// Try to load the template to check if it exists
	_, err := stl.LoadTemplate(name)
	return err == nil
}
