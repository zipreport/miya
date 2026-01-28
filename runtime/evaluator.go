package runtime

import (
	"fmt"
	"html"
	"math"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/zipreport/miya/parser"
)

// capitalizeFirst capitalizes the first letter of a string.
// This is used for struct field/method name lookups (e.g., "name" -> "Name").
// Note: strings.Title is deprecated since Go 1.18.
func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError {
		return s
	}
	return string(unicode.ToUpper(r)) + s[size:]
}

// Phase 4b optimization: Pool for loop info maps to reduce allocations
var loopInfoPool = sync.Pool{
	New: func() interface{} {
		return make(map[string]interface{}, 12) // Pre-size for typical loop info
	},
}

// Context interface for runtime package - matches main package interface
type Context interface {
	GetVariable(key string) (interface{}, bool)
	SetVariable(key string, value interface{})
	Clone() Context
	All() map[string]interface{}
}

// FastEvalNode interface for Phase 3c optimization
// Nodes that implement this can evaluate themselves directly
type FastEvalNode interface {
	FastEval(e *DefaultEvaluator, ctx Context) (interface{}, error)
}

// simpleContext is a minimal context implementation for fallback cases
type simpleContext struct {
	variables map[string]interface{}
}

// Cycler represents a cycler object that cycles through values
type Cycler struct {
	Items   []interface{}
	Current int
}

// Next returns the next item in the cycle
func (c *Cycler) Next() interface{} {
	if len(c.Items) == 0 {
		return nil
	}

	// Ensure Current is within bounds (defensive against external modification)
	if c.Current < 0 || c.Current >= len(c.Items) {
		c.Current = 0
	}

	item := c.Items[c.Current]
	c.Current = (c.Current + 1) % len(c.Items)
	return item
}

// GetCurrent returns the current item without advancing
func (c *Cycler) GetCurrent() interface{} {
	if len(c.Items) == 0 {
		return nil
	}
	// Ensure Current is within bounds
	if c.Current < 0 || c.Current >= len(c.Items) {
		return c.Items[0]
	}
	return c.Items[c.Current]
}

// Reset resets the cycler to the beginning
func (c *Cycler) Reset() {
	c.Current = 0
}

// Joiner represents a joiner object that joins values with separators
type Joiner struct {
	Separator string
	Used      bool
}

// Join returns the separator if this is not the first call, empty string otherwise
func (j *Joiner) Join() string {
	if j.Used {
		return j.Separator
	}
	j.Used = true
	return ""
}

// String implements the string interface for template output
func (j *Joiner) String() string {
	return j.Join()
}

// CallableLoop represents a loop object that can be called recursively and also accessed for properties
type CallableLoop struct {
	Info          map[string]interface{}
	RecursiveFunc func(interface{}) (interface{}, error)
}

// Call implements the function call for recursive loops
func (cl *CallableLoop) Call(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("loop() requires exactly one argument, got %d", len(args))
	}
	return cl.RecursiveFunc(args[0])
}

// GetAttribute allows access to loop properties like loop.index, loop.first, etc.
func (cl *CallableLoop) GetAttribute(name string) (interface{}, bool) {
	val, ok := cl.Info[name]
	return val, ok
}

func (c *simpleContext) GetVariable(name string) (interface{}, bool) {
	val, ok := c.variables[name]
	return val, ok
}

func (c *simpleContext) SetVariable(name string, value interface{}) {
	c.variables[name] = value
}

func (c *simpleContext) Clone() Context {
	clone := &simpleContext{variables: make(map[string]interface{})}
	for k, v := range c.variables {
		clone.variables[k] = v
	}
	return clone
}

func (c *simpleContext) All() map[string]interface{} {
	return c.variables
}

// NamespaceInterface defines the interface for namespace objects
type NamespaceInterface interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{})
}

// EnvironmentContext provides access to the template environment
type EnvironmentContext interface {
	Context
	ApplyFilter(name string, value interface{}, args ...interface{}) (interface{}, error)
	ApplyTest(name string, value interface{}, args ...interface{}) (bool, error)
}

// Evaluator interface for runtime operations
type Evaluator interface {
	EvalNode(node parser.Node, ctx Context) (interface{}, error)
}

type DefaultEvaluator struct {
	env              Context
	importSystem     *ImportSystem
	undefinedHandler *UndefinedHandler
}

func NewEvaluator() *DefaultEvaluator {
	return &DefaultEvaluator{
		undefinedHandler: NewUndefinedHandler(UndefinedSilent),
	}
}

// NewStrictEvaluator creates an evaluator with strict undefined handling
func NewStrictEvaluator() *DefaultEvaluator {
	return &DefaultEvaluator{
		undefinedHandler: NewUndefinedHandler(UndefinedStrict),
	}
}

// NewDebugEvaluator creates an evaluator with debug undefined handling
func NewDebugEvaluator() *DefaultEvaluator {
	return &DefaultEvaluator{
		undefinedHandler: NewUndefinedHandler(UndefinedDebug),
	}
}

// SetUndefinedBehavior sets the undefined variable behavior
func (e *DefaultEvaluator) SetUndefinedBehavior(behavior UndefinedBehavior) {
	if e.undefinedHandler == nil {
		e.undefinedHandler = NewUndefinedHandler(behavior)
	} else {
		e.undefinedHandler.SetUndefinedBehavior(behavior)
	}
}

// SetImportSystem sets the import system for handling template imports
func (e *DefaultEvaluator) SetImportSystem(importSystem *ImportSystem) {
	e.importSystem = importSystem
	// Update the import system's evaluator reference so macro calls use this evaluator
	if importSystem != nil {
		importSystem.SetEvaluator(e)
	}
}

func (e *DefaultEvaluator) EvalNode(node parser.Node, ctx Context) (interface{}, error) {
	// Phase 3c optimization: Fast path for nodes that implement FastEval
	if fastNode, ok := node.(FastEvalNode); ok {
		return fastNode.FastEval(e, ctx)
	}

	switch n := node.(type) {
	case *parser.TemplateNode:
		return e.EvalTemplateNode(n, ctx)
	case *parser.TextNode:
		return e.EvalTextNode(n, ctx)
	case *parser.CommentNode:
		return e.EvalCommentNode(n, ctx)
	case *parser.RawNode:
		return e.EvalRawNode(n, ctx)
	case *parser.VariableNode:
		return e.EvalVariableNode(n, ctx)
	case *parser.IdentifierNode:
		return e.EvalIdentifierNode(n, ctx)
	case *parser.LiteralNode:
		return e.EvalLiteralNode(n, ctx)
	case *parser.ListNode:
		return e.EvalListNode(n, ctx)
	case *parser.AttributeNode:
		return e.EvalAttributeNode(n, ctx)
	case *parser.GetItemNode:
		return e.EvalGetItemNode(n, ctx)
	case *parser.FilterNode:
		return e.EvalFilterNode(n, ctx)
	case *parser.BinaryOpNode:
		return e.EvalBinaryOpNode(n, ctx)
	case *parser.UnaryOpNode:
		return e.EvalUnaryOpNode(n, ctx)
	case *parser.IfNode:
		return e.EvalIfNode(n, ctx)
	case *parser.ForNode:
		return e.EvalForNode(n, ctx)
	case *parser.BlockNode:
		return e.EvalBlockNode(n, ctx)
	case *parser.SetNode:
		return e.EvalSetNode(n, ctx)
	case *parser.BlockSetNode:
		return e.EvalBlockSetNode(n, ctx)
	case *parser.CallNode:
		return e.EvalCallNode(n, ctx)
	case *parser.ExtendsNode:
		return e.EvalExtendsNode(n, ctx)
	case *parser.IncludeNode:
		return e.EvalIncludeNode(n, ctx)
	case *parser.SuperNode:
		return e.EvalSuperNode(n, ctx)
	case *parser.MacroNode:
		return e.EvalMacroNode(n, ctx)
	case *parser.TestNode:
		return e.EvalTestNode(n, ctx)
	case *parser.ConditionalNode:
		return e.EvalConditionalNode(n, ctx)
	case *parser.AssignmentNode:
		return e.EvalAssignmentNode(n, ctx)
	case *parser.SliceNode:
		return e.EvalSliceNode(n, ctx)
	case *parser.ComprehensionNode:
		return e.EvalComprehensionNode(n, ctx)
	case *parser.AutoescapeNode:
		return e.EvalAutoescapeNode(n, ctx)
	case *parser.ExtensionNode:
		return e.EvalExtensionNode(n, ctx)
	case *parser.BreakNode:
		return e.EvalBreakNode(n, ctx)
	case *parser.ContinueNode:
		return e.EvalContinueNode(n, ctx)
	case *parser.CallBlockNode:
		return e.EvalCallBlockNode(n, ctx)
	case *parser.ImportNode:
		return e.EvalImportNode(n, ctx)
	case *parser.FromNode:
		return e.EvalFromNode(n, ctx)
	case *parser.WithNode:
		return e.EvalWithNode(n, ctx)
	case *parser.DoNode:
		return e.EvalDoNode(n, ctx)
	case *parser.FilterBlockNode:
		return e.EvalFilterBlockNode(n, ctx)
	default:
		return nil, fmt.Errorf("unsupported node type: %T", node)
	}
}

func (e *DefaultEvaluator) EvalTemplateNode(node *parser.TemplateNode, ctx Context) (interface{}, error) {
	return e.evalNodeList(node.Children, ctx)
}

func (e *DefaultEvaluator) EvalTextNode(node *parser.TextNode, ctx Context) (string, error) {
	return node.Content, nil
}

func (e *DefaultEvaluator) EvalCommentNode(node *parser.CommentNode, ctx Context) (string, error) {
	// Comments produce no output
	return "", nil
}

func (e *DefaultEvaluator) EvalRawNode(node *parser.RawNode, ctx Context) (string, error) {
	// Raw nodes output their content exactly as-is
	return node.Content, nil
}

func (e *DefaultEvaluator) EvalVariableNode(node *parser.VariableNode, ctx Context) (interface{}, error) {
	result, err := e.EvalNode(node.Expression, ctx)
	if err != nil {
		return nil, err
	}

	// Handle nil values
	if result == nil {
		return "", nil
	}

	// Apply auto-escaping if enabled in this context
	if contextWrapper, ok := ctx.(ContextAwareContext); ok && contextWrapper.GetAutoEscaper() != nil {
		escaper := contextWrapper.GetAutoEscaper()
		if escaper.config.Enabled {
			return escaper.Escape(result, contextWrapper.GetEscapeContext()), nil
		}
	} else if autoCtx, ok := ctx.(AutoescapeContext); ok && autoCtx.IsAutoescapeEnabled() {
		// Fallback to old interface for backward compatibility
		if str, ok := result.(string); ok {
			return html.EscapeString(str), nil
		}
	}

	return result, nil
}

func (e *DefaultEvaluator) EvalIdentifierNode(node *parser.IdentifierNode, ctx Context) (interface{}, error) {
	value, ok := ctx.GetVariable(node.Name)
	if !ok {
		// Use undefined handler to determine behavior
		if e.undefinedHandler != nil {
			return e.undefinedHandler.Handle(node.Name, node)
		}
		// Fallback to original behavior
		return nil, NewUndefinedVariableError(node.Name, node)
	}
	return value, nil
}

func (e *DefaultEvaluator) EvalLiteralNode(node *parser.LiteralNode, ctx Context) (interface{}, error) {
	return node.Value, nil
}

func (e *DefaultEvaluator) EvalListNode(node *parser.ListNode, ctx Context) (interface{}, error) {
	result := make([]interface{}, len(node.Elements))
	for i, elem := range node.Elements {
		value, err := e.EvalNode(elem, ctx)
		if err != nil {
			return nil, err
		}
		result[i] = value
	}
	return result, nil
}

func (e *DefaultEvaluator) EvalAttributeNode(node *parser.AttributeNode, ctx Context) (interface{}, error) {
	obj, err := e.EvalNode(node.Object, ctx)
	if err != nil {
		return nil, err
	}

	// Handle undefined values with chained access
	if undefined, ok := obj.(*Undefined); ok && e.undefinedHandler != nil {
		return e.undefinedHandler.HandleAttributeAccess(undefined, node.Attribute, node)
	}

	// Check if attribute exists first, then get value
	if e.undefinedHandler != nil && !e.attributeExists(obj, node.Attribute) {
		attrName := fmt.Sprintf("%s.%s", e.getObjectName(obj), node.Attribute)
		return e.undefinedHandler.Handle(attrName, node)
	}

	// Get attribute value - this is safe now since we checked existence
	value := e.getAttribute(obj, node.Attribute)
	return value, nil
}

func (e *DefaultEvaluator) EvalGetItemNode(node *parser.GetItemNode, ctx Context) (interface{}, error) {
	obj, err := e.EvalNode(node.Object, ctx)
	if err != nil {
		return nil, err
	}

	key, err := e.EvalNode(node.Key, ctx)
	if err != nil {
		return nil, err
	}

	// Handle undefined values with item access
	if undefined, ok := obj.(*Undefined); ok && e.undefinedHandler != nil {
		return e.undefinedHandler.HandleItemAccess(undefined, key, node)
	}

	return e.getItem(obj, key)
}

func (e *DefaultEvaluator) EvalFilterNode(node *parser.FilterNode, ctx Context) (interface{}, error) {
	value, err := e.EvalNode(node.Expression, ctx)
	if err != nil {
		return nil, err
	}

	// Evaluate filter arguments with pre-allocated capacity
	args := make([]interface{}, 0, len(node.Arguments))
	for _, arg := range node.Arguments {
		argValue, err := e.EvalNode(arg, ctx)
		if err != nil {
			return nil, err
		}
		args = append(args, argValue)
	}

	// Try to use environment's filter registry if available
	if envCtx, ok := ctx.(EnvironmentContext); ok {
		return envCtx.ApplyFilter(node.FilterName, value, args...)
	}

	// Fallback to basic filters
	return e.applyFilter(node.FilterName, value, args)
}

func (e *DefaultEvaluator) EvalBinaryOpNode(node *parser.BinaryOpNode, ctx Context) (interface{}, error) {
	left, err := e.EvalNode(node.Left, ctx)
	if err != nil {
		return nil, err
	}

	right, err := e.EvalNode(node.Right, ctx)
	if err != nil {
		return nil, err
	}

	return e.applyBinaryOpWithNode(node.Operator, left, right, node)
}

func (e *DefaultEvaluator) EvalUnaryOpNode(node *parser.UnaryOpNode, ctx Context) (interface{}, error) {
	operand, err := e.EvalNode(node.Operand, ctx)
	if err != nil {
		return nil, err
	}

	return e.applyUnaryOp(node.Operator, operand)
}

func (e *DefaultEvaluator) EvalIfNode(node *parser.IfNode, ctx Context) (interface{}, error) {
	condition, err := e.EvalNode(node.Condition, ctx)
	if err != nil {
		return nil, err
	}

	if e.isTruthy(condition) {
		return e.evalNodeList(node.Body, ctx)
	}

	// Check elif conditions
	for _, elif := range node.ElseIfs {
		condition, err := e.EvalNode(elif.Condition, ctx)
		if err != nil {
			return nil, err
		}

		if e.isTruthy(condition) {
			return e.evalNodeList(elif.Body, ctx)
		}
	}

	// Else clause
	if len(node.Else) > 0 {
		return e.evalNodeList(node.Else, ctx)
	}

	return "", nil
}

func (e *DefaultEvaluator) EvalForNode(node *parser.ForNode, ctx Context) (interface{}, error) {
	iterable, err := e.EvalNode(node.Iterable, ctx)
	if err != nil {
		return nil, err
	}

	items, err := e.makeIterableForVariables(iterable, len(node.Variables))
	if err != nil {
		return nil, err
	}

	if len(items) == 0 && len(node.Else) > 0 {
		// Execute else clause if no items
		return e.evalNodeList(node.Else, ctx)
	}

	var results []string
	loopCtx := ctx.Clone()
	loopBroken := false

	// Determine loop depth once for the entire loop
	depth := 1
	if parentLoop, exists := ctx.GetVariable("loop"); exists {
		if parentLoopMap, ok := parentLoop.(map[string]interface{}); ok {
			if parentDepth, ok := parentLoopMap["depth"].(int); ok {
				depth = parentDepth + 1
			}
		}
	}

	// Track previous values for loop.changed() functionality
	var previousChanged map[string]interface{}

	// Pre-filter items if there's a condition to get correct loop indices
	filteredItems := make([]interface{}, 0, len(items))
	if node.Condition != nil {
		for _, item := range items {
			// Create a temporary context to evaluate the condition
			tempCtx := ctx.Clone()

			// Set loop variable(s) for condition evaluation
			if len(node.Variables) == 1 {
				tempCtx.SetVariable(node.Variables[0], item)
			} else {
				// Multiple variable assignment - unpack the item
				unpackedItems, err := e.makeIterable(item)
				if err != nil {
					return nil, fmt.Errorf("cannot unpack non-iterable %T for loop variables", item)
				}

				if len(unpackedItems) != len(node.Variables) {
					return nil, fmt.Errorf("cannot unpack %d values into %d variables", len(unpackedItems), len(node.Variables))
				}

				// Assign each value to its corresponding variable
				for j, variable := range node.Variables {
					tempCtx.SetVariable(variable, unpackedItems[j])
				}
			}

			// Evaluate condition
			conditionResult, err := e.EvalNode(node.Condition, tempCtx)
			if err != nil {
				return nil, err
			}
			if e.isTruthy(conditionResult) {
				filteredItems = append(filteredItems, item)
			}
		}
	} else {
		filteredItems = items
	}

	for i, item := range filteredItems {
		// Set loop variable(s)
		if len(node.Variables) == 1 {
			// Single variable assignment
			loopCtx.SetVariable(node.Variables[0], item)
		} else {
			// Multiple variable assignment - unpack the item
			unpackedItems, err := e.makeIterable(item)
			if err != nil {
				return nil, fmt.Errorf("cannot unpack non-iterable %T for loop variables", item)
			}

			if len(unpackedItems) != len(node.Variables) {
				return nil, fmt.Errorf("cannot unpack %d values into %d variables", len(unpackedItems), len(node.Variables))
			}

			// Assign each value to its corresponding variable
			for j, variable := range node.Variables {
				loopCtx.SetVariable(variable, unpackedItems[j])
			}
		}

		// Use the pre-calculated depth

		// Determine previtem and nextitem
		var previtem, nextitem interface{}
		if i > 0 {
			previtem = filteredItems[i-1]
		}
		if i < len(filteredItems)-1 {
			nextitem = filteredItems[i+1]
		}

		// Create cycle function for this loop
		cycleFunc := func(values ...interface{}) (interface{}, error) {
			if len(values) == 0 {
				return nil, fmt.Errorf("cycle() requires at least one argument")
			}
			cycleIndex := i % len(values)
			return values[cycleIndex], nil
		}

		// Create changed function for tracking value changes between iterations
		changedFunc := func(values ...interface{}) (interface{}, error) {
			if i == 0 {
				// First iteration, nothing to compare to - always considered changed
				previousChanged = make(map[string]interface{})
				for j, val := range values {
					key := fmt.Sprintf("val_%d", j)
					previousChanged[key] = val
				}
				return true, nil
			}

			// Check if any values have changed
			hasChanged := false
			if len(values) != len(previousChanged) {
				hasChanged = true
			} else {
				for j, val := range values {
					key := fmt.Sprintf("val_%d", j)
					if prevVal, exists := previousChanged[key]; !exists || !e.deepEqual(prevVal, val) {
						hasChanged = true
						break
					}
				}
			}

			// Update previous values for next iteration
			previousChanged = make(map[string]interface{})
			for j, val := range values {
				key := fmt.Sprintf("val_%d", j)
				previousChanged[key] = val
			}

			return hasChanged, nil
		}

		// Phase 4b optimization: Get loop info map from pool and reuse it
		var loopInfo map[string]interface{}
		if pooled, ok := loopInfoPool.Get().(map[string]interface{}); ok {
			loopInfo = pooled
			// Clear any stale keys from previous use
			for k := range loopInfo {
				delete(loopInfo, k)
			}
		} else {
			loopInfo = make(map[string]interface{}, 12)
		}

		// Update values (reuse map, don't allocate new one)
		loopInfo["index"] = i + 1
		loopInfo["index0"] = i
		loopInfo["revindex"] = len(filteredItems) - i
		loopInfo["revindex0"] = len(filteredItems) - i - 1
		loopInfo["first"] = i == 0
		loopInfo["last"] = i == len(filteredItems)-1
		loopInfo["length"] = len(filteredItems)
		loopInfo["depth"] = depth
		loopInfo["previtem"] = previtem
		loopInfo["nextitem"] = nextitem
		loopInfo["cycle"] = cycleFunc
		loopInfo["changed"] = changedFunc

		// Add recursive loop function if this is a recursive loop
		if node.Recursive {
			recursiveFunc := func(newIterable interface{}) (interface{}, error) {
				// Handle undefined or nil iterables - just return empty string
				if IsUndefined(newIterable) || newIterable == nil {
					return "", nil
				}

				// Create a new for node with the same structure but new iterable
				literalNode := &parser.LiteralNode{
					Value: newIterable,
					Raw:   fmt.Sprintf("%v", newIterable),
				}
				recursiveNode := &parser.ForNode{
					Variables: node.Variables,
					Iterable:  literalNode,
					Body:      node.Body,
					Else:      node.Else,
					Recursive: true,
				}
				return e.EvalForNode(recursiveNode, loopCtx)
			}

			// Create a callable loop object that supports both property access and function calls
			callableLoop := &CallableLoop{
				Info:          loopInfo,
				RecursiveFunc: recursiveFunc,
			}
			loopCtx.SetVariable("loop", callableLoop)
		} else {
			loopCtx.SetVariable("loop", loopInfo)
		}

		result, err := e.evalNodeList(node.Body, loopCtx)

		// Phase 4b optimization: Clear closure references and return map to pool
		// Clear closures to prevent memory leaks
		delete(loopInfo, "cycle")
		delete(loopInfo, "changed")
		loopInfoPool.Put(loopInfo)

		if err != nil {
			// Check if it's a loop control error
			if loopErr, ok := err.(*LoopControlError); ok {
				// Add any partial result from the current iteration
				if str, ok := result.(string); ok && str != "" {
					results = append(results, str)
				} else if result != nil && ToString(result) != "" {
					results = append(results, ToString(result))
				}

				if loopErr.IsBreak() {
					loopBroken = true
					break
				} else if loopErr.IsContinue() {
					continue
				}
			}
			return nil, err
		}

		if str, ok := result.(string); ok {
			results = append(results, str)
		} else {
			results = append(results, ToString(result))
		}
	}

	// If loop was broken and there are no results, execute else clause
	if loopBroken && len(results) == 0 && len(node.Else) > 0 {
		return e.evalNodeList(node.Else, ctx)
	}

	return strings.Join(results, ""), nil
}

func (e *DefaultEvaluator) EvalBlockNode(node *parser.BlockNode, ctx Context) (interface{}, error) {
	// Block evaluation is handled by the template inheritance system
	// For now, just evaluate the body
	return e.evalNodeList(node.Body, ctx)
}

func (e *DefaultEvaluator) EvalSetNode(node *parser.SetNode, ctx Context) (interface{}, error) {
	value, err := e.EvalNode(node.Value, ctx)
	if err != nil {
		return nil, err
	}

	if len(node.Targets) == 1 {
		// Single assignment - could be simple variable or attribute
		return e.assignToTarget(node.Targets[0], value, ctx)
	} else {
		// Multiple assignment - unpack the value
		items, err := e.makeIterable(value)
		if err != nil {
			return nil, fmt.Errorf("cannot unpack non-iterable %T for multiple assignment", value)
		}

		if len(items) != len(node.Targets) {
			return nil, fmt.Errorf("cannot unpack %d values into %d variables", len(items), len(node.Targets))
		}

		// Assign each value to its corresponding target
		for i, target := range node.Targets {
			if _, err := e.assignToTarget(target, items[i], ctx); err != nil {
				return nil, err
			}
		}
	}

	return "", nil // Set statements don't produce output
}

// assignToTarget assigns a value to a target expression (identifier, attribute, or subscript)
func (e *DefaultEvaluator) assignToTarget(target parser.ExpressionNode, value interface{}, ctx Context) (interface{}, error) {
	switch t := target.(type) {
	case *parser.IdentifierNode:
		// Simple variable assignment
		ctx.SetVariable(t.Name, value)
		return "", nil

	case *parser.AttributeNode:
		// Attribute assignment: obj.attr = value
		obj, err := e.EvalNode(t.Object, ctx)
		if err != nil {
			return nil, err
		}
		return nil, e.setAttribute(obj, t.Attribute, value)

	case *parser.GetItemNode:
		// Subscript assignment: obj[key] = value
		obj, err := e.EvalNode(t.Object, ctx)
		if err != nil {
			return nil, err
		}
		key, err := e.EvalNode(t.Key, ctx)
		if err != nil {
			return nil, err
		}
		return nil, e.setItem(obj, key, value)

	default:
		return nil, fmt.Errorf("invalid assignment target: %T", target)
	}
}

func (e *DefaultEvaluator) EvalBlockSetNode(node *parser.BlockSetNode, ctx Context) (interface{}, error) {
	// Evaluate the body to get the content for the variable
	bodyResult, err := e.evalNodeList(node.Body, ctx)
	if err != nil {
		return nil, fmt.Errorf("error evaluating block set body: %w", err)
	}

	// Convert the result to string (since block content is usually textual)
	var content string
	if str, ok := bodyResult.(string); ok {
		content = str
	} else {
		content = fmt.Sprintf("%v", bodyResult)
	}

	// Set the variable to the rendered content
	ctx.SetVariable(node.Variable, content)

	return "", nil // Block set statements don't produce output
}

func (e *DefaultEvaluator) EvalCallNode(node *parser.CallNode, ctx Context) (interface{}, error) {
	function, err := e.EvalNode(node.Function, ctx)
	if err != nil {
		return nil, err
	}

	// Evaluate arguments with pre-allocated capacity
	args := make([]interface{}, 0, len(node.Arguments))
	for _, arg := range node.Arguments {
		argValue, err := e.EvalNode(arg, ctx)
		if err != nil {
			return nil, err
		}
		args = append(args, argValue)
	}

	// Evaluate keyword arguments with pre-allocated capacity
	kwargs := make(map[string]interface{}, len(node.Keywords))
	for key, value := range node.Keywords {
		argValue, err := e.EvalNode(value, ctx)
		if err != nil {
			return nil, err
		}
		kwargs[key] = argValue
	}

	return e.callFunctionWithContext(function, args, kwargs, ctx)
}

func (e *DefaultEvaluator) EvalExtendsNode(node *parser.ExtendsNode, ctx Context) (interface{}, error) {
	// Extends nodes are handled by the inheritance resolver, not during runtime evaluation
	// In the runtime context, they should produce no output
	return "", nil
}

func (e *DefaultEvaluator) EvalIncludeNode(node *parser.IncludeNode, ctx Context) (interface{}, error) {
	// Evaluate the template name expression
	templateNameExpr, err := e.EvalNode(node.Template, ctx)
	if err != nil {
		return nil, fmt.Errorf("error evaluating template name in include: %w", err)
	}

	templateName, ok := templateNameExpr.(string)
	if !ok {
		return nil, fmt.Errorf("include template name must be a string, got %T", templateNameExpr)
	}

	// Get the import system
	if e.importSystem == nil {
		return nil, fmt.Errorf("import system not initialized for includes")
	}

	// Check if template exists
	if !e.importSystem.loader.TemplateExists(templateName) {
		if node.IgnoreMissing {
			// If ignore_missing is true, return empty string for missing templates
			return "", nil
		}
		return nil, fmt.Errorf("included template %q not found", templateName)
	}

	// Load the template AST
	templateAST, err := e.importSystem.loader.LoadTemplate(templateName)
	if err != nil {
		if node.IgnoreMissing {
			return "", nil
		}
		return nil, fmt.Errorf("failed to load included template %q: %w", templateName, err)
	}

	// Determine which context to use
	includeCtx := ctx
	if node.Context != nil {
		// If a context expression is provided, evaluate it
		contextValue, err := e.EvalNode(node.Context, ctx)
		if err != nil {
			return nil, fmt.Errorf("error evaluating context for include: %w", err)
		}

		// Create a new context with the provided values
		includeCtx = ctx.Clone()
		if contextMap, ok := contextValue.(map[string]interface{}); ok {
			for key, value := range contextMap {
				includeCtx.SetVariable(key, value)
			}
		}
	}

	// Execute the included template with the appropriate context
	result, err := e.EvalNode(templateAST, includeCtx)
	if err != nil {
		if node.IgnoreMissing {
			return "", nil
		}
		return nil, fmt.Errorf("error executing included template %q: %w", templateName, err)
	}

	return result, nil
}

func (e *DefaultEvaluator) EvalSuperNode(node *parser.SuperNode, ctx Context) (interface{}, error) {
	// Super nodes are handled by the inheritance resolver during template compilation
	// In the runtime context, they should have been replaced with parent block content
	// If we reach here, it means the super() call wasn't resolved properly
	return "", fmt.Errorf("super() call outside of block context")
}

func (e *DefaultEvaluator) EvalMacroNode(node *parser.MacroNode, ctx Context) (interface{}, error) {
	// Create a macro function that can be called with a context parameter
	macroFunc := func(callCtx Context, args ...interface{}) (interface{}, error) {
		// Create a new context for macro execution, inherit from the call context
		// to get access to variables like 'caller' that might be set by call blocks
		macroCtx := callCtx.Clone()

		// Set up macro parameters
		for i, paramName := range node.Parameters {
			if i < len(args) {
				macroCtx.SetVariable(paramName, args[i])
			} else if defaultExpr, hasDefault := node.Defaults[paramName]; hasDefault {
				// Evaluate default value
				defaultVal, err := e.EvalNode(defaultExpr, ctx)
				if err != nil {
					return nil, fmt.Errorf("error evaluating macro parameter default: %w", err)
				}
				macroCtx.SetVariable(paramName, defaultVal)
			} else {
				return nil, fmt.Errorf("missing required macro parameter: %s", paramName)
			}
		}

		// Execute macro body
		result, err := e.evalNodeList(node.Body, macroCtx)
		if err != nil {
			return nil, err
		}

		// Convert result to string
		if result == nil {
			return "", nil
		}
		return ToString(result), nil
	}

	// Register the macro as a variable in the context
	ctx.SetVariable(node.Name, macroFunc)

	// Macro definitions don't produce output during evaluation
	return "", nil
}

func (e *DefaultEvaluator) EvalTestNode(node *parser.TestNode, ctx Context) (interface{}, error) {
	// Special handling for the "defined" test - don't evaluate the expression first
	if node.TestName == "defined" {
		var isDefined bool

		// Check if it's an identifier node - most common case for "defined" test
		if identNode, ok := node.Expression.(*parser.IdentifierNode); ok {
			_, isDefined = ctx.GetVariable(identNode.Name)
		} else {
			// For other expression types, try to evaluate and check for undefined
			value, err := e.EvalNode(node.Expression, ctx)
			if err != nil {
				// If there's an error (e.g., undefined variable), it's not defined
				isDefined = false
			} else {
				// Check if the value is an Undefined type or nil
				isDefined = !IsUndefined(value) && value != nil
			}
		}

		result := isDefined
		if node.Negated {
			result = !result
		}
		return result, nil
	}

	// Evaluate the expression being tested
	value, err := e.EvalNode(node.Expression, ctx)
	if err != nil {
		return nil, err
	}

	// Evaluate test arguments with pre-allocated capacity
	args := make([]interface{}, 0, len(node.Arguments))
	for _, arg := range node.Arguments {
		argValue, err := e.EvalNode(arg, ctx)
		if err != nil {
			return nil, err
		}
		args = append(args, argValue)
	}

	// Try to use environment's test registry if available
	var result bool
	if envCtx, ok := ctx.(EnvironmentContext); ok {
		var testErr error
		result, testErr = envCtx.ApplyTest(node.TestName, value, args...)
		if testErr != nil {
			return nil, testErr
		}
	} else {
		// Fallback to basic tests
		var testErr error
		result, testErr = e.applyTest(node.TestName, value, args)
		if testErr != nil {
			return nil, testErr
		}
	}

	// Apply negation if needed
	if node.Negated {
		result = !result
	}

	return result, nil
}

func (e *DefaultEvaluator) EvalConditionalNode(node *parser.ConditionalNode, ctx Context) (interface{}, error) {
	// Evaluate condition
	condition, err := e.EvalNode(node.Condition, ctx)
	if err != nil {
		return nil, err
	}

	// Return appropriate expression based on condition
	if e.isTruthy(condition) {
		return e.EvalNode(node.TrueExpr, ctx)
	} else {
		return e.EvalNode(node.FalseExpr, ctx)
	}
}

func (e *DefaultEvaluator) EvalAssignmentNode(node *parser.AssignmentNode, ctx Context) (interface{}, error) {
	// Evaluate the value expression
	value, err := e.EvalNode(node.Value, ctx)
	if err != nil {
		return nil, err
	}

	// Handle different target types
	switch target := node.Target.(type) {
	case *parser.IdentifierNode:
		// Simple variable assignment
		ctx.SetVariable(target.Name, value)
		return "", nil // Assignment statements don't produce output

	case *parser.AttributeNode:
		// Attribute assignment (obj.attr = value)
		obj, err := e.EvalNode(target.Object, ctx)
		if err != nil {
			return nil, err
		}
		return nil, e.setAttribute(obj, target.Attribute, value)

	case *parser.GetItemNode:
		// Item assignment (obj[key] = value)
		obj, err := e.EvalNode(target.Object, ctx)
		if err != nil {
			return nil, err
		}
		key, err := e.EvalNode(target.Key, ctx)
		if err != nil {
			return nil, err
		}
		return nil, e.setItem(obj, key, value)

	default:
		return nil, fmt.Errorf("invalid assignment target: %T", target)
	}
}

func (e *DefaultEvaluator) EvalSliceNode(node *parser.SliceNode, ctx Context) (interface{}, error) {
	// Evaluate the object to be sliced
	obj, err := e.EvalNode(node.Object, ctx)
	if err != nil {
		return nil, err
	}

	// Evaluate slice parameters
	var start, end, step *int

	if node.Start != nil {
		startVal, err := e.EvalNode(node.Start, ctx)
		if err != nil {
			return nil, err
		}
		startInt, err := e.toInt(startVal)
		if err != nil {
			return nil, fmt.Errorf("slice start must be an integer: %v", err)
		}
		start = &startInt
	}

	if node.End != nil {
		endVal, err := e.EvalNode(node.End, ctx)
		if err != nil {
			return nil, err
		}
		endInt, err := e.toInt(endVal)
		if err != nil {
			return nil, fmt.Errorf("slice end must be an integer: %v", err)
		}
		end = &endInt
	}

	if node.Step != nil {
		stepVal, err := e.EvalNode(node.Step, ctx)
		if err != nil {
			return nil, err
		}
		stepInt, err := e.toInt(stepVal)
		if err != nil {
			return nil, fmt.Errorf("slice step must be an integer: %v", err)
		}
		if stepInt == 0 {
			return nil, fmt.Errorf("slice step cannot be zero")
		}
		step = &stepInt
	}

	return e.slice(obj, start, end, step)
}

func (e *DefaultEvaluator) EvalComprehensionNode(node *parser.ComprehensionNode, ctx Context) (interface{}, error) {
	// Evaluate the iterable
	iterable, err := e.EvalNode(node.Iterable, ctx)
	if err != nil {
		return nil, err
	}

	// Convert to slice for iteration
	items, err := e.makeIterable(iterable)
	if err != nil {
		return nil, err
	}

	if node.IsDict {
		// Dict comprehension
		result := make(map[string]interface{})

		for _, item := range items {
			// Create loop context
			loopCtx := ctx.Clone()
			loopCtx.SetVariable(node.Variable, item)

			// Check condition if present
			if node.Condition != nil {
				conditionResult, err := e.EvalNode(node.Condition, loopCtx)
				if err != nil {
					return nil, err
				}
				if !e.isTruthy(conditionResult) {
					continue
				}
			}

			// Evaluate key and value expressions
			key, err := e.EvalNode(node.KeyExpr, loopCtx)
			if err != nil {
				return nil, err
			}
			value, err := e.EvalNode(node.Expression, loopCtx)
			if err != nil {
				return nil, err
			}

			keyStr := fmt.Sprintf("%v", key)
			result[keyStr] = value
		}

		return result, nil
	} else {
		// List comprehension - pre-allocate with capacity for best case
		result := make([]interface{}, 0, len(items))

		for _, item := range items {
			// Create loop context
			loopCtx := ctx.Clone()
			loopCtx.SetVariable(node.Variable, item)

			// Check condition if present
			if node.Condition != nil {
				conditionResult, err := e.EvalNode(node.Condition, loopCtx)
				if err != nil {
					return nil, err
				}
				if !e.isTruthy(conditionResult) {
					continue
				}
			}

			// Evaluate expression
			value, err := e.EvalNode(node.Expression, loopCtx)
			if err != nil {
				return nil, err
			}

			result = append(result, value)
		}

		return result, nil
	}
}

func (e *DefaultEvaluator) EvalAutoescapeNode(node *parser.AutoescapeNode, ctx Context) (interface{}, error) {
	// Create a new context with the autoescape setting
	// We need to modify the context to track the autoescape state
	autoescapeCtx := &autoescapeContext{
		Context:    ctx,
		autoescape: node.Enabled,
	}

	// Evaluate the body with the new autoescape setting
	return e.evalNodeList(node.Body, autoescapeCtx)
}

// Helper methods

func (e *DefaultEvaluator) evalNodeList(nodes []parser.Node, ctx Context) (interface{}, error) {
	// Pre-allocate with exact capacity (Phase 3a optimization)
	results := make([]string, 0, len(nodes))

	for _, node := range nodes {
		result, err := e.EvalNode(node, ctx)
		if err != nil {
			// Propagate loop control errors up the stack, but include partial results
			if _, ok := err.(*LoopControlError); ok {
				return strings.Join(results, ""), err
			}
			return nil, err
		}

		if str, ok := result.(string); ok {
			results = append(results, str)
		} else if result == nil {
			results = append(results, "")
		} else {
			results = append(results, ToString(result))
		}
	}

	return strings.Join(results, ""), nil
}

// attributeExists checks if an attribute exists on an object (vs. exists but is nil)
func (e *DefaultEvaluator) attributeExists(obj interface{}, attr string) bool {
	if obj == nil {
		return false
	}

	switch v := obj.(type) {
	case NamespaceInterface:
		_, ok := v.Get(attr)
		return ok
	case *Cycler:
		// Handle Cycler methods - this ensures that cycler.next(), cycler.current(), and cycler.reset() work correctly
		// Without this case, attributeExists would return false for these methods, causing undefined behavior
		return attr == "next" || attr == "current" || attr == "reset"
	case *Joiner:
		// Joiner doesn't have specific methods, it's callable directly
		return false
	case *CallableLoop:
		// CallableLoop supports attribute access for loop properties
		_, ok := v.GetAttribute(attr)
		return ok
	case map[string]interface{}:
		// First check for actual keys in the map
		if _, ok := v[attr]; ok {
			return true
		}
		// If key doesn't exist, check for special dictionary methods
		return attr == "items" || attr == "keys" || attr == "values"
	case map[string]string:
		_, ok := v[attr]
		return ok
	default:
		// Use reflection for struct fields and other complex types
		rv := reflect.ValueOf(obj)
		if !rv.IsValid() {
			return false
		}

		// Handle pointers
		for rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface {
			if rv.IsNil() {
				return false
			}
			rv = rv.Elem()
		}

		switch rv.Kind() {
		case reflect.Struct:
			// Check if field exists and is accessible
			field := rv.FieldByName(attr)
			if !field.IsValid() {
				return false
			}
			// Check if field is exported (accessible)
			structType := rv.Type()
			if structField, found := structType.FieldByName(attr); found {
				return structField.PkgPath == "" // Empty PkgPath means exported
			}
			return false
		case reflect.Map:
			// Handle map types we might not have covered
			if rv.Type().Key().Kind() == reflect.String {
				key := reflect.ValueOf(attr)
				return rv.MapIndex(key).IsValid()
			}
			return false
		case reflect.Slice, reflect.Array:
			// For slices/arrays, check if attr is a valid numeric index
			if index, err := strconv.Atoi(attr); err == nil {
				return index >= 0 && index < rv.Len()
			}
			return false
		}

		// For other types, assume attribute doesn't exist
		return false
	}
}

// getObjectName tries to get a meaningful name for an object for error messages
func (e *DefaultEvaluator) getObjectName(obj interface{}) string {
	switch obj.(type) {
	case NamespaceInterface:
		return "namespace"
	case map[string]interface{}:
		return "object"
	case map[string]string:
		return "object"
	default:
		rv := reflect.ValueOf(obj)
		if !rv.IsValid() {
			return "invalid"
		}

		// Handle pointers and interfaces
		for rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface {
			if rv.IsNil() {
				return "nil"
			}
			rv = rv.Elem()
		}

		if rv.IsValid() {
			switch rv.Kind() {
			case reflect.Struct:
				return rv.Type().Name()
			case reflect.Map:
				return "map"
			case reflect.Slice:
				return "list"
			case reflect.Array:
				return "array"
			default:
				return rv.Type().String()
			}
		}
		return "unknown"
	}
}

// DictItems represents a dictionary items iterator for .items() method support
type DictItems struct {
	data map[string]interface{}
}

// String method for DictItems for template rendering
func (d *DictItems) String() string {
	return "[dict_items]"
}

func (e *DefaultEvaluator) getAttribute(obj interface{}, attr string) interface{} {
	if obj == nil {
		return nil
	}

	switch v := obj.(type) {
	case NamespaceInterface:
		val, ok := v.Get(attr)
		if !ok {
			return nil
		}
		return val
	case *Cycler:
		// Handle Cycler methods
		switch attr {
		case "next":
			return func(args ...interface{}) (interface{}, error) {
				return v.Next(), nil
			}
		case "current":
			return func(args ...interface{}) (interface{}, error) {
				return v.GetCurrent(), nil
			}
		case "reset":
			return func(args ...interface{}) (interface{}, error) {
				v.Reset()
				return nil, nil
			}
		}
		return nil
	case *Joiner:
		// Joiner is callable directly
		return func(args ...interface{}) (interface{}, error) {
			return v.Join(), nil
		}
	case *CallableLoop:
		// CallableLoop supports attribute access for loop properties
		val, ok := v.GetAttribute(attr)
		if !ok {
			return nil
		}
		return val
	case map[string]interface{}:
		// First check if the key exists in the map
		if val, exists := v[attr]; exists {
			return val
		}
		// If key doesn't exist, check for special dictionary methods
		switch attr {
		case "items":
			// Return a callable function that returns DictItems (Python-style behavior)
			return func(args ...interface{}) (interface{}, error) {
				return &DictItems{data: v}, nil
			}
		case "keys":
			// Return a callable function that returns the keys
			return func(args ...interface{}) (interface{}, error) {
				keys := make([]interface{}, 0, len(v))
				for k := range v {
					keys = append(keys, k)
				}
				return keys, nil
			}
		case "values":
			// Return a callable function that returns the values
			return func(args ...interface{}) (interface{}, error) {
				values := make([]interface{}, 0, len(v))
				for _, val := range v {
					values = append(values, val)
				}
				return values, nil
			}
		default:
			return nil
		}
	case map[string]string:
		return v[attr]
	default:
		// Use reflection for struct fields and methods
		rv := reflect.ValueOf(obj)
		if rv.Kind() == reflect.Ptr {
			rv = rv.Elem()
		}

		if rv.Kind() == reflect.Struct {
			field := rv.FieldByName(attr)
			if field.IsValid() && field.CanInterface() {
				return field.Interface()
			}

			// Try with capitalized field name
			field = rv.FieldByName(capitalizeFirst(attr))
			if field.IsValid() && field.CanInterface() {
				return field.Interface()
			}
		}

		// Try methods on the original value (including pointer receiver methods)
		origValue := reflect.ValueOf(obj)
		method := origValue.MethodByName(capitalizeFirst(attr))
		if method.IsValid() {
			return method.Interface()
		}

		return nil
	}
}

func (e *DefaultEvaluator) getItem(obj, key interface{}) (interface{}, error) {
	if obj == nil {
		return nil, fmt.Errorf("cannot get item from nil")
	}

	switch v := obj.(type) {
	case map[string]interface{}:
		keyStr := fmt.Sprintf("%v", key)
		return v[keyStr], nil
	case []interface{}:
		keyInt, err := e.toInt(key)
		if err != nil {
			return nil, fmt.Errorf("list index must be integer, got %T", key)
		}
		// Support negative indexing like Python
		if keyInt < 0 {
			keyInt = len(v) + keyInt
		}
		if keyInt < 0 || keyInt >= len(v) {
			// Return undefined for out of bounds access (Jinja2 behavior)
			return NewUndefined(fmt.Sprintf("index[%d]", keyInt), UndefinedSilent, nil), nil
		}
		return v[keyInt], nil
	case string:
		keyInt, err := e.toInt(key)
		if err != nil {
			return nil, fmt.Errorf("string index must be integer, got %T", key)
		}
		// Support negative indexing like Python
		if keyInt < 0 {
			keyInt = len(v) + keyInt
		}
		if keyInt < 0 || keyInt >= len(v) {
			// Return undefined for out of bounds access (Jinja2 behavior)
			return NewUndefined(fmt.Sprintf("index[%d]", keyInt), UndefinedSilent, nil), nil
		}
		return string(v[keyInt]), nil
	default:
		// Try reflection for slice/array access
		rv := reflect.ValueOf(obj)
		if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
			keyInt, err := e.toInt(key)
			if err != nil {
				return nil, fmt.Errorf("index must be integer, got %T", key)
			}
			if keyInt < 0 || keyInt >= rv.Len() {
				// Return undefined for out of bounds access (Jinja2 behavior)
				return NewUndefined(fmt.Sprintf("index[%d]", keyInt), UndefinedSilent, nil), nil
			}
			return rv.Index(keyInt).Interface(), nil
		}

		return nil, fmt.Errorf("object is not subscriptable: %T", obj)
	}
}

// applyBinaryOpWithNode applies binary operation with enhanced error reporting
func (e *DefaultEvaluator) applyBinaryOpWithNode(op string, left, right interface{}, node parser.Node) (interface{}, error) {
	switch op {
	case "+":
		return e.addWithNode(left, right, node)
	case "-":
		return e.subtractWithNode(left, right, node)
	case "*":
		return e.multiplyWithNode(left, right, node)
	case "/":
		return e.divideWithNode(left, right, node)
	case "//":
		return e.floorDivideWithNode(left, right, node)
	case "%":
		return e.moduloWithNode(left, right, node)
	case "**":
		return e.powerWithNode(left, right, node)
	case "==":
		return e.equal(left, right), nil
	case "!=":
		return !e.equal(left, right), nil
	case "<":
		return e.less(left, right)
	case "<=":
		return e.lessEqual(left, right)
	case ">":
		return e.greater(left, right)
	case ">=":
		return e.greaterEqual(left, right)
	case "and":
		return e.isTruthy(left) && e.isTruthy(right), nil
	case "or":
		return e.isTruthy(left) || e.isTruthy(right), nil
	case "~":
		return e.concatenateWithNode(left, right, node)
	case "in":
		return e.contains(right, left)
	case "not in":
		result, err := e.contains(right, left)
		return !result, err
	default:
		return nil, NewRuntimeError(ErrorTypeRuntime, fmt.Sprintf("unknown binary operator: %s", op), node)
	}
}

// Legacy method for backward compatibility
func (e *DefaultEvaluator) applyBinaryOp(op string, left, right interface{}) (interface{}, error) {
	return e.applyBinaryOpWithNode(op, left, right, nil)
}

func (e *DefaultEvaluator) applyUnaryOp(op string, operand interface{}) (interface{}, error) {
	switch op {
	case "not":
		return !e.isTruthy(operand), nil
	case "-":
		return e.negate(operand)
	case "+":
		return operand, nil
	default:
		return nil, fmt.Errorf("unsupported unary operator: %s", op)
	}
}

func (e *DefaultEvaluator) applyFilter(name string, value interface{}, args []interface{}) (interface{}, error) {
	// This is a fallback implementation - in practice, the environment's filter registry should be used
	// For now, implement basic filters directly
	switch name {
	case "upper":
		return strings.ToUpper(fmt.Sprintf("%v", value)), nil
	case "lower":
		return strings.ToLower(fmt.Sprintf("%v", value)), nil
	case "capitalize":
		s := fmt.Sprintf("%v", value)
		if len(s) == 0 {
			return s, nil
		}
		return strings.ToUpper(s[:1]) + strings.ToLower(s[1:]), nil
	case "trim":
		return strings.TrimSpace(fmt.Sprintf("%v", value)), nil
	case "length":
		return e.length(value)
	case "default":
		if len(args) == 0 {
			return value, nil
		}
		if e.isTruthy(value) {
			return value, nil
		}
		return args[0], nil
	case "escape":
		return e.htmlEscape(fmt.Sprintf("%v", value)), nil
	case "safe":
		// Mark as safe - would need proper implementation
		return value, nil
	default:
		return nil, fmt.Errorf("unknown filter: %s", name)
	}
}

func (e *DefaultEvaluator) callFunction(function interface{}, args []interface{}, kwargs map[string]interface{}) (interface{}, error) {
	return e.callFunctionWithContext(function, args, kwargs, nil)
}

func (e *DefaultEvaluator) callFunctionWithContext(function interface{}, args []interface{}, kwargs map[string]interface{}, ctx Context) (interface{}, error) {
	if function == nil {
		return nil, fmt.Errorf("cannot call nil function")
	}

	// Handle special callable objects
	switch fn := function.(type) {
	case *Joiner:
		// Joiner objects are directly callable
		return fn.Join(), nil
	case *CallableLoop:
		// CallableLoop objects support function calls for recursion
		return fn.Call(args...)
	}

	// Handle different function types
	switch fn := function.(type) {
	case func(Context, ...interface{}) (interface{}, error):
		// Macro function that takes context as first parameter
		if ctx != nil {
			return fn(ctx, args...)
		}
		// If no context provided, use a default empty context
		emptyCtx := &simpleContext{variables: make(map[string]interface{})}
		return fn(emptyCtx, args...)

	case func(...interface{}) (interface{}, error):
		// Variable argument function - append kwargs as last argument if present
		if len(kwargs) > 0 {
			args = append(args, kwargs)
		}
		return fn(args...)

	case func([]interface{}) (interface{}, error):
		return fn(args)

	case func([]interface{}, map[string]interface{}) (interface{}, error):
		return fn(args, kwargs)

	case func() (interface{}, error):
		if len(args) > 0 || len(kwargs) > 0 {
			return nil, fmt.Errorf("function takes no arguments")
		}
		return fn()

	default:
		// Try reflection for more flexible function calls
		fnValue := reflect.ValueOf(function)
		if fnValue.Kind() != reflect.Func {
			return nil, fmt.Errorf("cannot call non-function value of type %T", function)
		}

		fnType := fnValue.Type()

		// Check if it's a macro or other callable
		if fnType.NumIn() == 0 && fnType.NumOut() == 2 {
			// Simple callable with no args that returns (interface{}, error)
			results := fnValue.Call(nil)
			if err, ok := results[1].Interface().(error); ok && err != nil {
				return nil, err
			}
			return results[0].Interface(), nil
		} else if fnType.NumIn() == 0 && fnType.NumOut() == 1 {
			// Simple callable with no args that returns a single value
			results := fnValue.Call(nil)
			return results[0].Interface(), nil
		}

		// For now, return error for unsupported function signatures
		return nil, fmt.Errorf("unsupported function signature: %v", fnType)
	}
}

func (e *DefaultEvaluator) applyTest(name string, value interface{}, args []interface{}) (bool, error) {
	// This is a fallback implementation - in practice, the environment's test registry should be used
	// For now, implement basic tests directly
	switch name {
	case "defined":
		return value != nil, nil
	case "undefined":
		return value == nil, nil
	case "none":
		return value == nil, nil
	case "boolean":
		_, ok := value.(bool)
		return ok, nil
	case "string":
		_, ok := value.(string)
		return ok, nil
	case "number":
		switch value.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
			return true, nil
		default:
			return false, nil
		}
	case "integer":
		switch value.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			return true, nil
		default:
			return false, nil
		}
	case "float":
		switch value.(type) {
		case float32, float64:
			return true, nil
		default:
			return false, nil
		}
	case "even":
		num, err := e.toInt(value)
		if err != nil {
			return false, fmt.Errorf("even test requires an integer, got %T", value)
		}
		return num%2 == 0, nil
	case "odd":
		num, err := e.toInt(value)
		if err != nil {
			return false, fmt.Errorf("odd test requires an integer, got %T", value)
		}
		return num%2 != 0, nil
	case "divisibleby":
		if len(args) != 1 {
			return false, fmt.Errorf("divisibleby test requires exactly one argument")
		}
		num, err := e.toInt(value)
		if err != nil {
			return false, fmt.Errorf("divisibleby test requires an integer, got %T", value)
		}
		divisor, err := e.toInt(args[0])
		if err != nil {
			return false, fmt.Errorf("divisibleby test requires an integer divisor, got %T", args[0])
		}
		if divisor == 0 {
			return false, fmt.Errorf("division by zero")
		}
		return num%divisor == 0, nil
	case "lower":
		str, ok := value.(string)
		if !ok {
			return false, fmt.Errorf("lower test requires a string, got %T", value)
		}
		return str == strings.ToLower(str), nil
	case "upper":
		str, ok := value.(string)
		if !ok {
			return false, fmt.Errorf("upper test requires a string, got %T", value)
		}
		return str == strings.ToUpper(str), nil
	case "startswith":
		if len(args) != 1 {
			return false, fmt.Errorf("startswith test requires exactly one argument")
		}
		str, ok := value.(string)
		if !ok {
			return false, fmt.Errorf("startswith test requires a string, got %T", value)
		}
		prefix, ok := args[0].(string)
		if !ok {
			return false, fmt.Errorf("startswith test requires a string prefix, got %T", args[0])
		}
		return strings.HasPrefix(str, prefix), nil
	case "endswith":
		if len(args) != 1 {
			return false, fmt.Errorf("endswith test requires exactly one argument")
		}
		str, ok := value.(string)
		if !ok {
			return false, fmt.Errorf("endswith test requires a string, got %T", value)
		}
		suffix, ok := args[0].(string)
		if !ok {
			return false, fmt.Errorf("endswith test requires a string suffix, got %T", args[0])
		}
		return strings.HasSuffix(str, suffix), nil
	case "sequence":
		if value == nil {
			return false, nil
		}
		rv := reflect.ValueOf(value)
		switch rv.Kind() {
		case reflect.Slice, reflect.Array:
			return true, nil
		case reflect.String:
			return true, nil // strings are sequences in Jinja2
		default:
			return false, nil
		}
	case "mapping":
		if value == nil {
			return false, nil
		}
		rv := reflect.ValueOf(value)
		return rv.Kind() == reflect.Map, nil
	case "iterable":
		if value == nil {
			return false, nil
		}
		rv := reflect.ValueOf(value)
		switch rv.Kind() {
		case reflect.Slice, reflect.Array, reflect.Map, reflect.String:
			return true, nil
		default:
			return false, nil
		}
	case "in":
		if len(args) != 1 {
			return false, fmt.Errorf("in test requires exactly one argument")
		}
		return e.contains(args[0], value)
	default:
		return false, fmt.Errorf("unknown test: %s", name)
	}
}

func (e *DefaultEvaluator) makeIterable(obj interface{}) ([]interface{}, error) {
	if obj == nil {
		return nil, nil
	}

	// Handle undefined variables as empty iterables
	if _, ok := obj.(*Undefined); ok {
		return []interface{}{}, nil
	}

	// Phase 4a optimization: Fast paths with type assertions BEFORE reflection
	// This covers 90%+ of cases and avoids reflection overhead
	switch v := obj.(type) {
	case []interface{}:
		// Most common case - already correct type
		return v, nil

	case []string:
		// Pre-allocate and convert
		result := make([]interface{}, len(v))
		for i, s := range v {
			result[i] = s
		}
		return result, nil

	case []int:
		result := make([]interface{}, len(v))
		for i, n := range v {
			result[i] = n
		}
		return result, nil

	case []map[string]interface{}:
		result := make([]interface{}, len(v))
		for i, m := range v {
			result[i] = m
		}
		return result, nil

	case map[string]interface{}:
		// Return values only for single variable iteration
		result := make([]interface{}, 0, len(v))
		for _, value := range v {
			result = append(result, value)
		}
		return result, nil

	case *DictItems:
		// Handle DictItems objects - return key-value pairs for iteration
		result := make([]interface{}, 0, len(v.data))
		for key, value := range v.data {
			result = append(result, []interface{}{key, value})
		}
		return result, nil

	case string:
		result := make([]interface{}, len(v))
		for i, r := range v {
			result[i] = string(r)
		}
		return result, nil
	}

	// Slow path: Check if obj is a function that should be called to get an iterable
	if fnValue := reflect.ValueOf(obj); fnValue.Kind() == reflect.Func {
		fnType := fnValue.Type()

		// Check if it's a callable function that returns an iterable
		canCall := false
		var args []reflect.Value

		if fnType.NumIn() == 0 && fnType.NumOut() >= 1 {
			// Pattern: func() T or func() (T, error)
			canCall = true
			args = nil
		} else if fnType.IsVariadic() && fnType.NumOut() >= 1 {
			// Pattern: func(...interface{}) (interface{}, error) - variadic function
			canCall = true
			args = nil // Call with no arguments
		}

		if canCall {
			// Call the function
			results := fnValue.Call(args)
			if len(results) > 0 {
				// Use the first result as the iterable object
				actualObj := results[0].Interface()
				// Handle potential error return
				if len(results) > 1 {
					if err, ok := results[1].Interface().(error); ok && err != nil {
						return nil, err
					}
				}
				// Recursively call makeIterable with the actual object
				return e.makeIterable(actualObj)
			}
		}
		// If we can't call the function or it doesn't fit our pattern, fall through to error
	}

	// Fallback to reflection for uncommon slice/array types
	rv := reflect.ValueOf(obj)
	if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
		result := make([]interface{}, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			result[i] = rv.Index(i).Interface()
		}
		return result, nil
	}

	return nil, fmt.Errorf("object is not iterable: %T", obj)
}

// makeIterableForVariables creates an iterable based on the number of variables
// For maps with multiple variables, returns key-value pairs; otherwise returns values
func (e *DefaultEvaluator) makeIterableForVariables(obj interface{}, numVariables int) ([]interface{}, error) {
	if obj == nil {
		return nil, nil
	}

	// Handle undefined variables as empty iterables
	if _, ok := obj.(*Undefined); ok {
		return []interface{}{}, nil
	}

	// Check if obj is a function that should be called to get an iterable
	if fnValue := reflect.ValueOf(obj); fnValue.Kind() == reflect.Func {
		fnType := fnValue.Type()

		// Check if it's a callable function that returns an iterable
		// Handle both func() T and func(...interface{}) (interface{}, error) patterns
		canCall := false
		var args []reflect.Value

		if fnType.NumIn() == 0 && fnType.NumOut() >= 1 {
			// Pattern: func() T or func() (T, error)
			canCall = true
			args = nil
		} else if fnType.IsVariadic() && fnType.NumOut() >= 1 {
			// Pattern: func(...interface{}) (interface{}, error) - variadic function
			canCall = true
			args = nil // Call with no arguments
		}

		if canCall {
			// Call the function
			results := fnValue.Call(args)
			if len(results) > 0 {
				// Use the first result as the iterable object
				actualObj := results[0].Interface()
				// Handle potential error return
				if len(results) > 1 {
					if err, ok := results[1].Interface().(error); ok && err != nil {
						return nil, err
					}
				}
				// Recursively call makeIterableForVariables with the actual object
				return e.makeIterableForVariables(actualObj, numVariables)
			}
		}
		// If we can't call the function or it doesn't fit our pattern, fall through to error
	}

	switch v := obj.(type) {
	case *DictItems:
		// Handle .items() method result - always returns key-value pairs
		// Pre-allocate with known capacity to avoid reallocations
		result := make([]interface{}, 0, len(v.data))
		for key, value := range v.data {
			result = append(result, []interface{}{key, value})
		}
		return result, nil
	case map[string]interface{}:
		if numVariables == 2 {
			// Return key-value pairs for unpacking
			// Pre-allocate with known capacity to avoid reallocations
			result := make([]interface{}, 0, len(v))
			for key, value := range v {
				result = append(result, []interface{}{key, value})
			}
			return result, nil
		} else if numVariables == 1 {
			// Return values only for single variable
			// Pre-allocate with known capacity to avoid reallocations
			result := make([]interface{}, 0, len(v))
			for _, value := range v {
				result = append(result, value)
			}
			return result, nil
		} else {
			return nil, fmt.Errorf("cannot unpack map into %d variables (expected 1 or 2)", numVariables)
		}
	default:
		// For non-maps, use the standard makeIterable method
		return e.makeIterable(obj)
	}
}

func (e *DefaultEvaluator) isTruthy(obj interface{}) bool {
	if obj == nil {
		return false
	}

	// Check for undefined values
	if IsUndefined(obj) {
		return false
	}

	// Fast path: direct type comparisons without reflection
	switch v := obj.(type) {
	case bool:
		return v
	case string:
		return v != ""
	case int:
		return v != 0
	case int8:
		return v != 0
	case int16:
		return v != 0
	case int32:
		return v != 0
	case int64:
		return v != 0
	case uint:
		return v != 0
	case uint8:
		return v != 0
	case uint16:
		return v != 0
	case uint32:
		return v != 0
	case uint64:
		return v != 0
	case float32:
		return v != 0
	case float64:
		return v != 0
	case []interface{}:
		return len(v) > 0
	case map[string]interface{}:
		return len(v) > 0
	default:
		// Fallback to reflection for other types
		rv := reflect.ValueOf(obj)
		switch rv.Kind() {
		case reflect.Slice, reflect.Map, reflect.Array:
			return rv.Len() > 0
		default:
			return true
		}
	}
}

// Arithmetic operations
// Enhanced addition with better error reporting
func (e *DefaultEvaluator) addWithNode(a, b interface{}, node parser.Node) (interface{}, error) {
	// Handle undefined values gracefully - treat undefined as 0 in arithmetic operations
	if IsUndefined(a) {
		a = 0
	}
	if IsUndefined(b) {
		b = 0
	}

	// Try numeric addition first
	aFloat, aErr := e.toFloat(a)
	bFloat, bErr := e.toFloat(b)
	if aErr == nil && bErr == nil {
		return aFloat + bFloat, nil
	}

	// Try string concatenation
	aStr, aStrOk := a.(string)
	bStr, bStrOk := b.(string)
	if aStrOk && bStrOk {
		return aStr + bStr, nil
	}

	// Try mixed-type string concatenation (graceful handling)
	// Only allow mixed-type concatenation in non-strict mode
	isStrictMode := e.undefinedHandler != nil && e.undefinedHandler.GetUndefinedBehavior() == UndefinedStrict
	if !isStrictMode {
		// If one operand is string, convert the other to string and concatenate
		if aStrOk {
			return aStr + ToString(b), nil
		}
		if bStrOk {
			return ToString(a) + bStr, nil
		}
	}

	// Try slice concatenation
	if reflect.TypeOf(a).Kind() == reflect.Slice && reflect.TypeOf(b).Kind() == reflect.Slice {
		av := reflect.ValueOf(a)
		bv := reflect.ValueOf(b)
		result := reflect.MakeSlice(av.Type(), 0, av.Len()+bv.Len())
		result = reflect.AppendSlice(result, av)
		result = reflect.AppendSlice(result, bv)
		return result.Interface(), nil
	}

	return nil, NewTypeError("addition", a, node).WithContext(fmt.Sprintf("cannot add %T and %T", a, b))
}

func (e *DefaultEvaluator) add(a, b interface{}) (interface{}, error) {
	return e.addWithNode(a, b, nil)
}

// Placeholder enhanced operations (simplified for now)
func (e *DefaultEvaluator) subtractWithNode(a, b interface{}, node parser.Node) (interface{}, error) {
	return e.subtract(a, b)
}

func (e *DefaultEvaluator) multiplyWithNode(a, b interface{}, node parser.Node) (interface{}, error) {
	return e.multiply(a, b)
}

func (e *DefaultEvaluator) moduloWithNode(a, b interface{}, node parser.Node) (interface{}, error) {
	return e.modulo(a, b)
}

func (e *DefaultEvaluator) powerWithNode(a, b interface{}, node parser.Node) (interface{}, error) {
	return e.power(a, b)
}

func (e *DefaultEvaluator) concatenateWithNode(a, b interface{}, node parser.Node) (interface{}, error) {
	return fmt.Sprintf("%v%v", a, b), nil
}

// Legacy method for existing code
func (e *DefaultEvaluator) addOld(a, b interface{}) (interface{}, error) {
	// String concatenation
	if aStr, ok := a.(string); ok {
		return aStr + fmt.Sprintf("%v", b), nil
	}
	if bStr, ok := b.(string); ok {
		return fmt.Sprintf("%v", a) + bStr, nil
	}

	// Array/slice concatenation
	if aSlice, ok := a.([]interface{}); ok {
		if bSlice, ok := b.([]interface{}); ok {
			result := make([]interface{}, len(aSlice)+len(bSlice))
			copy(result, aSlice)
			copy(result[len(aSlice):], bSlice)
			return result, nil
		}
		// Concatenate single element to slice
		result := make([]interface{}, len(aSlice)+1)
		copy(result, aSlice)
		result[len(aSlice)] = b
		return result, nil
	}
	if bSlice, ok := b.([]interface{}); ok {
		// Concatenate single element to beginning of slice
		result := make([]interface{}, 1+len(bSlice))
		result[0] = a
		copy(result[1:], bSlice)
		return result, nil
	}

	// Numeric addition
	aFloat, aErr := e.toFloat(a)
	bFloat, bErr := e.toFloat(b)
	if aErr == nil && bErr == nil {
		return aFloat + bFloat, nil
	}

	return nil, fmt.Errorf("cannot add %T and %T", a, b)
}

func (e *DefaultEvaluator) subtract(a, b interface{}) (interface{}, error) {
	// Handle undefined values gracefully - treat undefined as 0 in arithmetic operations
	if IsUndefined(a) {
		a = 0
	}
	if IsUndefined(b) {
		b = 0
	}

	aFloat, aErr := e.toFloat(a)
	bFloat, bErr := e.toFloat(b)
	if aErr == nil && bErr == nil {
		return aFloat - bFloat, nil
	}
	return nil, fmt.Errorf("cannot subtract %T and %T", a, b)
}

func (e *DefaultEvaluator) multiply(a, b interface{}) (interface{}, error) {
	// Handle undefined values gracefully - treat undefined as 0 in arithmetic operations
	if IsUndefined(a) {
		a = 0
	}
	if IsUndefined(b) {
		b = 0
	}

	aFloat, aErr := e.toFloat(a)
	bFloat, bErr := e.toFloat(b)
	if aErr == nil && bErr == nil {
		return aFloat * bFloat, nil
	}
	return nil, fmt.Errorf("cannot multiply %T and %T", a, b)
}

// Enhanced math operations with better error reporting
func (e *DefaultEvaluator) divideWithNode(a, b interface{}, node parser.Node) (interface{}, error) {
	// Handle undefined values gracefully - treat undefined as 0 in arithmetic operations
	if IsUndefined(a) {
		a = 0
	}
	if IsUndefined(b) {
		b = 0
	}

	aFloat, aErr := e.toFloat(a)
	bFloat, bErr := e.toFloat(b)
	if aErr == nil && bErr == nil {
		if bFloat == 0 {
			return nil, NewMathError("division", fmt.Errorf("division by zero"), node)
		}
		return aFloat / bFloat, nil
	}
	return nil, NewTypeError("division", a, node).WithContext(fmt.Sprintf("cannot divide %T and %T", a, b))
}

func (e *DefaultEvaluator) divide(a, b interface{}) (interface{}, error) {
	return e.divideWithNode(a, b, nil)
}

func (e *DefaultEvaluator) floorDivideWithNode(a, b interface{}, node parser.Node) (interface{}, error) {
	aInt, aErr := e.toInt(a)
	bInt, bErr := e.toInt(b)
	if aErr == nil && bErr == nil {
		if bInt == 0 {
			return nil, NewMathError("floor division", fmt.Errorf("division by zero"), node)
		}
		return aInt / bInt, nil
	}
	return nil, NewTypeError("floor division", a, node).WithContext(fmt.Sprintf("cannot floor divide %T and %T", a, b))
}

func (e *DefaultEvaluator) floorDivide(a, b interface{}) (interface{}, error) {
	return e.floorDivideWithNode(a, b, nil)
}

func (e *DefaultEvaluator) modulo(a, b interface{}) (interface{}, error) {
	// Handle undefined values gracefully - treat undefined as 0 in arithmetic operations
	if IsUndefined(a) {
		a = 0
	}
	if IsUndefined(b) {
		b = 0
	}

	aInt, aErr := e.toInt(a)
	bInt, bErr := e.toInt(b)
	if aErr == nil && bErr == nil {
		if bInt == 0 {
			return nil, fmt.Errorf("modulo by zero")
		}
		return aInt % bInt, nil
	}
	return nil, fmt.Errorf("cannot modulo %T and %T", a, b)
}

func (e *DefaultEvaluator) power(a, b interface{}) (interface{}, error) {
	// Simple power implementation using math.Pow
	aFloat, aErr := e.toFloat(a)
	bFloat, bErr := e.toFloat(b)
	if aErr == nil && bErr == nil {
		result := math.Pow(aFloat, bFloat)
		return result, nil
	}
	return nil, fmt.Errorf("power operator requires numeric operands")
}

func (e *DefaultEvaluator) negate(a interface{}) (interface{}, error) {
	aFloat, err := e.toFloat(a)
	if err == nil {
		return -aFloat, nil
	}
	return nil, fmt.Errorf("cannot negate %T", a)
}

// Comparison operations
func (e *DefaultEvaluator) equal(a, b interface{}) bool {
	return reflect.DeepEqual(a, b)
}

func (e *DefaultEvaluator) less(a, b interface{}) (bool, error) {
	aFloat, aErr := e.toFloat(a)
	bFloat, bErr := e.toFloat(b)
	if aErr == nil && bErr == nil {
		return aFloat < bFloat, nil
	}

	// String comparison
	if aStr, ok := a.(string); ok {
		if bStr, ok := b.(string); ok {
			return aStr < bStr, nil
		}
	}

	return false, fmt.Errorf("cannot compare %T and %T", a, b)
}

func (e *DefaultEvaluator) lessEqual(a, b interface{}) (bool, error) {
	less, err := e.less(a, b)
	if err != nil {
		return false, err
	}
	return less || e.equal(a, b), nil
}

func (e *DefaultEvaluator) greater(a, b interface{}) (bool, error) {
	lessEq, err := e.lessEqual(a, b)
	if err != nil {
		return false, err
	}
	return !lessEq, nil
}

func (e *DefaultEvaluator) greaterEqual(a, b interface{}) (bool, error) {
	less, err := e.less(a, b)
	if err != nil {
		return false, err
	}
	return !less, nil
}

func (e *DefaultEvaluator) contains(container, item interface{}) (bool, error) {
	switch v := container.(type) {
	case string:
		itemStr := e.fastToString(item)
		return strings.Contains(v, itemStr), nil
	case []interface{}:
		for _, elem := range v {
			if e.equal(elem, item) {
				return true, nil
			}
		}
		return false, nil
	case map[string]interface{}:
		itemStr := e.fastToString(item)
		_, ok := v[itemStr]
		return ok, nil
	default:
		// Use reflection to handle any slice/array type
		rv := reflect.ValueOf(container)
		switch rv.Kind() {
		case reflect.Slice, reflect.Array:
			for i := 0; i < rv.Len(); i++ {
				elem := rv.Index(i).Interface()
				if e.equal(elem, item) {
					return true, nil
				}
			}
			return false, nil
		case reflect.Map:
			// Handle any map type
			keys := rv.MapKeys()
			itemValue := reflect.ValueOf(item)
			for _, key := range keys {
				if reflect.DeepEqual(key.Interface(), item) {
					return true, nil
				}
			}
			// Also check if item can be used as a key directly
			if itemValue.IsValid() && itemValue.Type().AssignableTo(rv.Type().Key()) {
				return rv.MapIndex(itemValue).IsValid(), nil
			}
			return false, nil
		default:
			return false, fmt.Errorf("cannot check containment in %T", container)
		}
	}
}

// fastToString converts a value to string without fmt.Sprintf overhead
func (e *DefaultEvaluator) fastToString(v interface{}) string {
	switch s := v.(type) {
	case string:
		return s
	case int:
		return strconv.Itoa(s)
	case int64:
		return strconv.FormatInt(s, 10)
	case float64:
		return strconv.FormatFloat(s, 'f', -1, 64)
	case bool:
		if s {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprintf("%v", v)
	}
}

// deepEqual performs deep equality comparison for loop.changed() functionality
func (e *DefaultEvaluator) deepEqual(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// For basic types, use standard equality
	if reflect.TypeOf(a) != reflect.TypeOf(b) {
		return false
	}

	switch va := a.(type) {
	case string, int, int64, float64, bool:
		return a == b
	case []interface{}:
		vb, ok := b.([]interface{})
		if !ok || len(va) != len(vb) {
			return false
		}
		for i, item := range va {
			if !e.deepEqual(item, vb[i]) {
				return false
			}
		}
		return true
	case map[string]interface{}:
		vb, ok := b.(map[string]interface{})
		if !ok || len(va) != len(vb) {
			return false
		}
		for key, value := range va {
			if bValue, exists := vb[key]; !exists || !e.deepEqual(value, bValue) {
				return false
			}
		}
		return true
	default:
		// For complex types, use reflection
		return reflect.DeepEqual(a, b)
	}
}

func (e *DefaultEvaluator) length(obj interface{}) (int, error) {
	if obj == nil {
		return 0, nil
	}

	switch v := obj.(type) {
	case string:
		return len(v), nil
	case []interface{}:
		return len(v), nil
	case map[string]interface{}:
		return len(v), nil
	default:
		rv := reflect.ValueOf(obj)
		switch rv.Kind() {
		case reflect.Slice, reflect.Array, reflect.Map:
			return rv.Len(), nil
		default:
			return 0, fmt.Errorf("object has no length: %T", obj)
		}
	}
}

func (e *DefaultEvaluator) htmlEscape(s string) string {
	// Fast path: check if escaping is needed
	needsEscape := false
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '&', '<', '>', '"', '\'':
			needsEscape = true
			break
		}
		if needsEscape {
			break
		}
	}
	if !needsEscape {
		return s
	}

	// Single-pass escaping with strings.Builder
	var buf strings.Builder
	buf.Grow(len(s) + len(s)/8) // Estimate ~12% growth for escapes
	for _, r := range s {
		switch r {
		case '&':
			buf.WriteString("&amp;")
		case '<':
			buf.WriteString("&lt;")
		case '>':
			buf.WriteString("&gt;")
		case '"':
			buf.WriteString("&quot;")
		case '\'':
			buf.WriteString("&#x27;")
		default:
			buf.WriteRune(r)
		}
	}
	return buf.String()
}

// Type conversion helpers
func (e *DefaultEvaluator) toInt(obj interface{}) (int, error) {
	switch v := obj.(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	case string:
		return strconv.Atoi(v)
	default:
		return 0, fmt.Errorf("cannot convert %T to int", obj)
	}
}

func (e *DefaultEvaluator) toFloat(obj interface{}) (float64, error) {
	switch v := obj.(type) {
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case float64:
		return v, nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to float", obj)
	}
}

// setAttribute sets an attribute on an object (for assignment expressions)
func (e *DefaultEvaluator) setAttribute(obj interface{}, attr string, value interface{}) error {
	if obj == nil {
		return fmt.Errorf("cannot set attribute on nil object")
	}

	switch v := obj.(type) {
	case NamespaceInterface:
		v.Set(attr, value)
		return nil
	case map[string]interface{}:
		v[attr] = value
		return nil
	}

	rv := reflect.ValueOf(obj)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	if rv.Kind() == reflect.Struct {
		field := rv.FieldByName(attr)
		if !field.IsValid() {
			// Try with capitalized field name
			field = rv.FieldByName(capitalizeFirst(attr))
		}

		if field.IsValid() && field.CanSet() {
			valueRv := reflect.ValueOf(value)
			if valueRv.Type().ConvertibleTo(field.Type()) {
				field.Set(valueRv.Convert(field.Type()))
				return nil
			}
			return fmt.Errorf("cannot assign %T to %s field", value, field.Type())
		}
		return fmt.Errorf("cannot set field %s on struct", attr)
	}

	return fmt.Errorf("cannot set attribute on %T", obj)
}

// setItem sets an item in a container (for assignment expressions)
func (e *DefaultEvaluator) setItem(obj, key, value interface{}) error {
	if obj == nil {
		return fmt.Errorf("cannot set item on nil object")
	}

	switch v := obj.(type) {
	case map[string]interface{}:
		keyStr := fmt.Sprintf("%v", key)
		v[keyStr] = value
		return nil
	case []interface{}:
		keyInt, err := e.toInt(key)
		if err != nil {
			return fmt.Errorf("list index must be integer, got %T", key)
		}
		// Support negative indexing like Python
		if keyInt < 0 {
			keyInt = len(v) + keyInt
		}
		if keyInt < 0 || keyInt >= len(v) {
			return fmt.Errorf("list index out of range: %d", keyInt)
		}
		v[keyInt] = value
		return nil
	default:
		// Try reflection for slice assignment
		rv := reflect.ValueOf(obj)
		if rv.Kind() == reflect.Slice {
			keyInt, err := e.toInt(key)
			if err != nil {
				return fmt.Errorf("index must be integer, got %T", key)
			}
			if keyInt < 0 || keyInt >= rv.Len() {
				return fmt.Errorf("index out of range: %d", keyInt)
			}

			valueRv := reflect.ValueOf(value)
			elemType := rv.Type().Elem()
			if valueRv.Type().ConvertibleTo(elemType) {
				rv.Index(keyInt).Set(valueRv.Convert(elemType))
				return nil
			}
			return fmt.Errorf("cannot assign %T to slice element of type %s", value, elemType)
		}

		return fmt.Errorf("object is not subscriptable for assignment: %T", obj)
	}
}

// slice implements Python-style slicing
func (e *DefaultEvaluator) slice(obj interface{}, start, end, step *int) (interface{}, error) {
	if obj == nil {
		return nil, fmt.Errorf("cannot slice nil")
	}

	switch v := obj.(type) {
	case string:
		return e.sliceString(v, start, end, step), nil
	case []interface{}:
		return e.sliceSlice(v, start, end, step), nil
	default:
		// Try reflection for slice types
		rv := reflect.ValueOf(obj)
		if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
			return e.sliceReflect(rv, start, end, step), nil
		}

		return nil, fmt.Errorf("object is not sliceable: %T", obj)
	}
}

func (e *DefaultEvaluator) sliceString(s string, start, end, step *int) string {
	length := len(s)

	// Default values
	startVal := 0
	endVal := length
	stepVal := 1

	if start != nil {
		startVal = *start
		if startVal < 0 {
			startVal += length
		}
		if startVal < 0 {
			startVal = 0
		}
		if startVal > length {
			startVal = length
		}
	}

	if end != nil {
		endVal = *end
		if endVal < 0 {
			endVal += length
		}
		if endVal < 0 {
			endVal = 0
		}
		if endVal > length {
			endVal = length
		}
	}

	if step != nil {
		stepVal = *step
	}

	if stepVal == 1 {
		// Simple case
		if startVal >= endVal {
			return ""
		}
		return s[startVal:endVal]
	}

	// Step-wise slicing
	var result strings.Builder
	if stepVal > 0 {
		for i := startVal; i < endVal; i += stepVal {
			result.WriteByte(s[i])
		}
	} else {
		// Negative step
		if start == nil {
			startVal = length - 1
		}
		if end == nil {
			endVal = -1
		}
		for i := startVal; i > endVal; i += stepVal {
			result.WriteByte(s[i])
		}
	}

	return result.String()
}

func (e *DefaultEvaluator) sliceSlice(s []interface{}, start, end, step *int) []interface{} {
	length := len(s)

	// Default values
	startVal := 0
	endVal := length
	stepVal := 1

	if start != nil {
		startVal = *start
		if startVal < 0 {
			startVal += length
		}
		if startVal < 0 {
			startVal = 0
		}
		if startVal > length {
			startVal = length
		}
	}

	if end != nil {
		endVal = *end
		if endVal < 0 {
			endVal += length
		}
		if endVal < 0 {
			endVal = 0
		}
		if endVal > length {
			endVal = length
		}
	}

	if step != nil {
		stepVal = *step
	}

	if stepVal == 1 {
		// Simple case
		if startVal >= endVal {
			return []interface{}{}
		}
		return s[startVal:endVal]
	}

	// Step-wise slicing
	var result []interface{}
	if stepVal > 0 {
		for i := startVal; i < endVal; i += stepVal {
			result = append(result, s[i])
		}
	} else {
		// Negative step
		if start == nil {
			startVal = length - 1
		}
		if end == nil {
			endVal = -1
		}
		for i := startVal; i > endVal; i += stepVal {
			result = append(result, s[i])
		}
	}

	return result
}

func (e *DefaultEvaluator) sliceReflect(rv reflect.Value, start, end, step *int) interface{} {
	length := rv.Len()

	// Default values
	startVal := 0
	endVal := length
	stepVal := 1

	if start != nil {
		startVal = *start
		if startVal < 0 {
			startVal += length
		}
		if startVal < 0 {
			startVal = 0
		}
		if startVal > length {
			startVal = length
		}
	}

	if end != nil {
		endVal = *end
		if endVal < 0 {
			endVal += length
		}
		if endVal < 0 {
			endVal = 0
		}
		if endVal > length {
			endVal = length
		}
	}

	if step != nil {
		stepVal = *step
	}

	// Create result slice
	elemType := rv.Type().Elem()
	result := reflect.MakeSlice(reflect.SliceOf(elemType), 0, 0)

	if stepVal > 0 {
		for i := startVal; i < endVal; i += stepVal {
			result = reflect.Append(result, rv.Index(i))
		}
	} else {
		// Negative step
		if start == nil {
			startVal = length - 1
		}
		if end == nil {
			endVal = -1
		}
		for i := startVal; i > endVal; i += stepVal {
			result = reflect.Append(result, rv.Index(i))
		}
	}

	return result.Interface()
}

// autoescapeContext wraps a Context to track autoescape state
type autoescapeContext struct {
	Context
	autoescape bool
}

// IsAutoescapeEnabled returns whether autoescape is enabled in this context
func (ac *autoescapeContext) IsAutoescapeEnabled() bool {
	return ac.autoescape
}

// AutoescapeContext interface for contexts that support autoescape state
type AutoescapeContext interface {
	Context
	IsAutoescapeEnabled() bool
}

// EvalExtensionNode evaluates extension nodes
func (e *DefaultEvaluator) EvalExtensionNode(node *parser.ExtensionNode, ctx Context) (interface{}, error) {
	if node.EvaluateFunc == nil {
		return nil, fmt.Errorf("extension node '%s:%s' has no evaluation function", node.ExtensionName, node.TagName)
	}
	return node.EvaluateFunc(node, ctx)
}

// EvalBreakNode evaluates break statements
func (e *DefaultEvaluator) EvalBreakNode(node *parser.BreakNode, ctx Context) (interface{}, error) {
	// Get the control flow evaluator
	cfEvaluator := NewControlFlowEvaluator(e)
	return nil, cfEvaluator.EvalBreak()
}

// EvalContinueNode evaluates continue statements
func (e *DefaultEvaluator) EvalContinueNode(node *parser.ContinueNode, ctx Context) (interface{}, error) {
	// Get the control flow evaluator
	cfEvaluator := NewControlFlowEvaluator(e)
	return nil, cfEvaluator.EvalContinue()
}

// EvalCallBlockNode evaluates call blocks ({% call macro() %}content{% endcall %})
func (e *DefaultEvaluator) EvalCallBlockNode(node *parser.CallBlockNode, ctx Context) (interface{}, error) {
	// Render the block content to create the "caller" function
	blockContent, err := e.evalNodeList(node.Body, ctx)
	if err != nil {
		return nil, err
	}

	// Convert block content to string
	var blockStr string
	if blockContent != nil {
		blockStr = fmt.Sprintf("%v", blockContent)
	}

	// Create a "caller" function that returns the block content
	callerFunc := func(args ...interface{}) (interface{}, error) {
		return blockStr, nil
	}

	// Create a new context that includes the caller function
	callCtx := ctx.Clone()
	callCtx.SetVariable("caller", callerFunc)

	// Evaluate the call expression with the extended context
	result, err := e.EvalNode(node.Call, callCtx)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// EvalImportNode evaluates import statements ({% import 'template' as name %})
func (e *DefaultEvaluator) EvalImportNode(node *parser.ImportNode, ctx Context) (interface{}, error) {
	// Evaluate the template expression to get the template name
	templateExpr, err := e.EvalNode(node.Template, ctx)
	if err != nil {
		return nil, fmt.Errorf("error evaluating template expression in import: %w", err)
	}

	templateName, ok := templateExpr.(string)
	if !ok {
		return nil, fmt.Errorf("template name must be a string, got %T", templateExpr)
	}

	// Use the new import system if available
	if e.importSystem != nil {
		namespace, err := e.importSystem.LoadTemplateNamespace(templateName, ctx)
		if err != nil {
			return nil, fmt.Errorf("error loading template %q: %w", templateName, err)
		}

		// Create an ImportedNamespace wrapper
		importedNS := e.importSystem.GetImportedNamespace(namespace)
		ctx.SetVariable(node.Alias, importedNS)
	} else {
		// Fallback to the old placeholder system
		templateNamespace, err := e.loadTemplateNamespace(templateName, ctx)
		if err != nil {
			return nil, fmt.Errorf("error loading template %q: %w", templateName, err)
		}
		ctx.SetVariable(node.Alias, templateNamespace)
	}

	return "", nil // Import statements don't produce output
}

// EvalFromNode evaluates from-import statements ({% from 'template' import name1, name2 %})
func (e *DefaultEvaluator) EvalFromNode(node *parser.FromNode, ctx Context) (interface{}, error) {

	// Evaluate the template expression to get the template name
	templateExpr, err := e.EvalNode(node.Template, ctx)
	if err != nil {
		return nil, fmt.Errorf("error evaluating template expression in from: %w", err)
	}

	templateName, ok := templateExpr.(string)
	if !ok {
		return nil, fmt.Errorf("template name must be a string, got %T", templateExpr)
	}

	var namespaceMap map[string]interface{}

	// Use the new import system if available
	if e.importSystem != nil {
		namespace, err := e.importSystem.LoadTemplateNamespace(templateName, ctx)
		if err != nil {
			return nil, fmt.Errorf("error loading template %q: %w", templateName, err)
		}
		namespaceMap = e.importSystem.GetNamespaceMap(namespace)
	} else {
		// Fallback to the old placeholder system
		templateNamespace, err := e.loadTemplateNamespace(templateName, ctx)
		if err != nil {
			return nil, fmt.Errorf("error loading template %q: %w", templateName, err)
		}
		var ok bool
		namespaceMap, ok = templateNamespace.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("template namespace is not a map")
		}
	}

	// Import specific items from the namespace
	for _, name := range node.Names {
		// Determine the variable name (use alias if provided, otherwise original name)
		varName := name
		if alias, hasAlias := node.Aliases[name]; hasAlias {
			varName = alias
		}

		// Get the item from the namespace
		if item, exists := namespaceMap[name]; exists {
			ctx.SetVariable(varName, item)
		} else {
			// Create a placeholder if the item doesn't exist
			placeholderFunc := func(args ...interface{}) (interface{}, error) {
				return fmt.Sprintf("[Undefined macro %s from %s]", name, templateName), nil
			}
			ctx.SetVariable(varName, placeholderFunc)
		}
	}

	return "", nil // From statements don't produce output
}

// EvalWithNode evaluates with statements ({% with var=expr %}...{% endwith %})
func (e *DefaultEvaluator) EvalWithNode(node *parser.WithNode, ctx Context) (interface{}, error) {
	// Create a new context scope for the with block
	withCtx := ctx.Clone()

	// Evaluate and set assignments in order (so later assignments can reference earlier ones)
	// Note: In Go, map iteration order is not guaranteed, but for most use cases this will work
	// In a production implementation, we'd need to preserve assignment order in the parser
	for varName, expr := range node.Assignments {
		value, err := e.EvalNode(expr, withCtx) // Use with context so assignments can reference each other
		if err != nil {
			return nil, fmt.Errorf("error evaluating with assignment %s: %v", varName, err)
		}
		withCtx.SetVariable(varName, value)
	}

	// Evaluate the body with the new context
	result, err := e.evalNodeList(node.Body, withCtx)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// loadTemplateNamespace loads a template and extracts its macros and variables
func (e *DefaultEvaluator) loadTemplateNamespace(templateName string, ctx Context) (interface{}, error) {
	// We need access to the environment to load templates
	// This is a limitation of the current design - the runtime evaluator doesn't have direct access to the environment
	// For now, we'll return a basic namespace with placeholder functionality

	// Try to get environment from context if available
	if envCtx, ok := ctx.(EnvironmentContext); ok {
		// In a full implementation, we would:
		// 1. Load the template using envCtx.LoadTemplate(templateName)
		// 2. Parse the template AST to find MacroNodes
		// 3. Create callable functions for each macro
		// 4. Return a namespace containing all the macros
		_ = envCtx // Use the context to avoid unused variable warning
	}

	// For now, create a basic namespace regardless of context availability
	return e.createPlaceholderNamespace(templateName), nil
}

// createPlaceholderNamespace creates a placeholder namespace for import functionality
func (e *DefaultEvaluator) createPlaceholderNamespace(templateName string) map[string]interface{} {
	return map[string]interface{}{
		"__template__": templateName,
		"__imported__": true,
		// Add some common macro names as placeholders
		"render_field": func(args ...interface{}) (interface{}, error) {
			return fmt.Sprintf("[Macro render_field from %s]", templateName), nil
		},
		"input": func(args ...interface{}) (interface{}, error) {
			return fmt.Sprintf("[Macro input from %s]", templateName), nil
		},
		"button": func(args ...interface{}) (interface{}, error) {
			return fmt.Sprintf("[Macro button from %s]", templateName), nil
		},
	}
}

// EvalDoNode evaluates a do statement by executing the expression for side effects
func (e *DefaultEvaluator) EvalDoNode(node *parser.DoNode, ctx Context) (interface{}, error) {
	// Execute the expression for its side effects
	_, err := e.EvalNode(node.Expression, ctx)
	if err != nil {
		return nil, err
	}

	// Do statements produce no output
	return "", nil
}

// EvalFilterBlockNode evaluates a filter block by rendering the body content and applying filters
func (e *DefaultEvaluator) EvalFilterBlockNode(node *parser.FilterBlockNode, ctx Context) (interface{}, error) {
	// First, render the body content
	var bodyContent strings.Builder
	for _, bodyNode := range node.Body {
		result, err := e.EvalNode(bodyNode, ctx)
		if err != nil {
			return nil, err
		}

		// Convert result to string
		resultStr := ToString(result)
		bodyContent.WriteString(resultStr)
	}

	// Get the rendered body content as starting value
	currentContent := bodyContent.String()

	// Apply each filter in the chain to the content
	for _, filter := range node.FilterChain {
		// Create a literal node with the current content as input to the filter
		contentLiteral := parser.NewLiteralNode(currentContent, currentContent, filter.Line(), filter.Column())

		// Create a new filter node with the current content as the expression
		filterNode := parser.NewFilterNode(contentLiteral, filter.FilterName, filter.Arguments, filter.Line(), filter.Column())
		filterNode.NamedArgs = filter.NamedArgs

		// Evaluate the filter
		result, err := e.EvalFilterNode(filterNode, ctx)
		if err != nil {
			return nil, fmt.Errorf("error applying filter '%s' in filter block: %v", filter.FilterName, err)
		}

		// Convert result back to string for the next filter in chain
		currentContent = ToString(result)
	}

	return currentContent, nil
}
