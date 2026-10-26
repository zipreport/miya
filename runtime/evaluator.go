package runtime

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
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

// Context interface for runtime package - matches main package interface
type Context interface {
	GetVariable(key string) (interface{}, bool)
	SetVariable(key string, value interface{})
	Clone() Context
	All() map[string]interface{}
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
	// Safe marks a separator that was given as Markup.
	Safe bool
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

// NamespaceMarker marks a type as a namespace() object. Embed it to let templates
// assign attributes on the type with {% set x.attr = ... %}.
type NamespaceMarker struct{}

func (NamespaceMarker) miyaNamespace() {}

// assignableNamespace is satisfied only by types that embed NamespaceMarker.
type assignableNamespace interface {
	NamespaceInterface
	miyaNamespace()
}

// EnvironmentContext provides access to the template environment
type EnvironmentContext interface {
	Context
	ApplyFilter(name string, value interface{}, args ...interface{}) (interface{}, error)
	ApplyTest(name string, value interface{}, args ...interface{}) (bool, error)
}

// KwargsFilterApplier is implemented by contexts that can apply a filter with
// keyword arguments. It is separate from EnvironmentContext so existing
// implementations of that interface keep compiling.
type KwargsFilterApplier interface {
	ApplyFilterKwargs(name string, value interface{}, args []interface{}, kwargs map[string]interface{}) (interface{}, error)
}

// Evaluator interface for runtime operations
type Evaluator interface {
	EvalNode(node parser.Node, ctx Context) (interface{}, error)
}

// ErrMaxRecursion is returned when includes, macro calls and recursive loops
// nest deeper than maxRecursionDepth.
var ErrMaxRecursion = errors.New("maximum template recursion depth exceeded")

// ErrTemplateNotFound is matched (errors.Is) by every error that means a
// template does not exist: the loader could not provide its source. A
// template that exists but fails to compile is not "not found".
var ErrTemplateNotFound = errors.New("template not found")

// TemplateNotFoundError reports which template could not be found. It
// matches ErrTemplateNotFound with errors.Is and unwraps to the loader's error.
type TemplateNotFoundError struct {
	Name string
	Err  error
}

func (e *TemplateNotFoundError) Error() string {
	return fmt.Sprintf("failed to load template %q: %v", e.Name, e.Err)
}
func (e *TemplateNotFoundError) Unwrap() error        { return e.Err }
func (e *TemplateNotFoundError) Is(target error) bool { return target == ErrTemplateNotFound }

const maxRecursionDepth = 200

type DefaultEvaluator struct {
	env              Context
	importSystem     *ImportSystem
	undefinedHandler *UndefinedHandler
	depth            atomic.Int32
	// importMemo holds the modules imported without context during the current
	// render, so every such import of a template shares one module, as in Jinja2.
	// SetImportSystem clears it; renders never share it.
	importMemo map[string]*TemplateNamespace
}

// enter counts one level of template recursion; call the returned func when leaving.
// The counter is atomic because one evaluator's depth may be touched from
// goroutines started by Go callables. Imported macro closures are no longer
// shared across renders.
func (e *DefaultEvaluator) enter(what string) (func(), error) {
	if e.depth.Add(1) > maxRecursionDepth {
		e.depth.Add(-1)
		return func() {}, fmt.Errorf("%w (%d levels) at %s", ErrMaxRecursion, maxRecursionDepth, what)
	}
	return func() { e.depth.Add(-1) }, nil
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
	e.importMemo = nil
}

func (e *DefaultEvaluator) EvalNode(node parser.Node, ctx Context) (interface{}, error) {
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
	case *parser.DictNode:
		return e.EvalDictNode(n, ctx)
	case *parser.AttributeNode:
		return e.EvalAttributeNode(n, ctx)
	case *parser.GetItemNode:
		return e.EvalGetItemNode(n, ctx)
	case *parser.FilterNode:
		return e.EvalFilterNode(n, ctx)
	case *parser.BinaryOpNode:
		return e.EvalBinaryOpNode(n, ctx)
	case *parser.CompareNode:
		return e.EvalCompareNode(n, ctx)
	case *parser.UnaryOpNode:
		return e.EvalUnaryOpNode(n, ctx)
	case *parser.IfNode:
		return e.EvalIfNode(n, ctx)
	case *parser.ForNode:
		return e.EvalForNode(n, ctx)
	case *parser.BlockNode:
		return e.EvalBlockNode(n, ctx)
	case *parser.DiscardNode:
		if _, err := e.evalNodeList(n.Body, ctx); err != nil {
			return nil, err
		}
		return "", nil
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

	// A strict undefined value stored in a variable (such as a name a
	// {% from %} import did not find) raises when printed, as in Jinja2.
	if u, ok := result.(*Undefined); ok && u.Behavior == UndefinedStrict {
		if u.Hint != "" {
			return nil, fmt.Errorf("undefined variable %s: %s", u.Name, u.Hint)
		}
		return nil, u.Error()
	}

	// Apply auto-escaping if enabled in this context
	if on, _, _ := resolveEscape(ctx); on {
		return escapeFor(ctx, result), nil
	}

	return result, nil
}

func (e *DefaultEvaluator) EvalIdentifierNode(node *parser.IdentifierNode, ctx Context) (interface{}, error) {
	value, ok := ctx.GetVariable(node.Name)
	if u, isUndef := value.(*Undefined); ok && isUndef && e.undefinedHandler != nil && e.undefinedHandler.GetUndefinedBehavior() == UndefinedStrict {
		// A bound undefined (a macro parameter that was not passed, or a name
		// a {% from %} import did not find) fails on use, keeping its hint.
		if u.Hint != "" {
			return nil, fmt.Errorf("undefined variable %s: %s", u.Name, u.Hint)
		}
		ok = false
	}
	if !ok {
		// Use undefined handler to determine behavior
		if e.undefinedHandler != nil {
			return e.undefinedHandler.Handle(node.Name, node)
		}
		// Fallback to original behavior
		return nil, NewUndefinedVariableError(node.Name, node)
	}
	return Unbox(value), nil
}

// evalLookup evaluates expr like EvalNode, except that when expr is a
// variable, attribute or item lookup that finds nothing it returns an
// *Undefined instead of failing, even in strict mode. Jinja2's StrictUndefined
// can be passed to default and to the defined/undefined/none tests; only
// using it fails. Errors from evaluating the base of the lookup still fail.
func (e *DefaultEvaluator) evalLookup(expr parser.Node, ctx Context) (interface{}, error) {
	behavior := UndefinedSilent
	if e.undefinedHandler != nil {
		behavior = e.undefinedHandler.GetUndefinedBehavior()
	}
	switch n := expr.(type) {
	case *parser.IdentifierNode:
		if v, ok := ctx.GetVariable(n.Name); ok {
			return Unbox(v), nil
		}
		return NewUndefined(n.Name, behavior, n), nil
	case *parser.AttributeNode:
		if isListMethodName(n.Attribute) {
			if raw, ok := e.rawOperand(n.Object, ctx); ok {
				if b, ok := raw.(*listBox); ok {
					return b.method(e, n.Attribute), nil
				}
			}
		}
		obj, err := e.EvalNode(n.Object, ctx)
		if err != nil {
			return nil, err
		}
		if u, ok := obj.(*Undefined); ok && e.undefinedHandler != nil {
			return e.undefinedHandler.HandleAttributeAccess(u, n.Attribute, n)
		}
		if m, ok := e.builtinMethodFor(obj, n.Attribute); ok {
			return m, nil
		}
		missing := NewUndefined(fmt.Sprintf("%s.%s", e.getObjectName(obj), n.Attribute), behavior, n)
		if UsesReflectLookup(obj) {
			value, ok := e.lookupReflect(obj, n.Attribute)
			if !ok {
				return missing, nil
			}
			return value, nil
		}
		if !e.attributeExists(obj, n.Attribute) {
			return missing, nil
		}
		return e.getAttribute(obj, n.Attribute), nil
	case *parser.GetItemNode:
		obj, err := e.EvalNode(n.Object, ctx)
		if err != nil {
			return nil, err
		}
		key, err := e.EvalNode(n.Key, ctx)
		if err != nil {
			return nil, err
		}
		if u, ok := obj.(*Undefined); ok && e.undefinedHandler != nil {
			return e.undefinedHandler.HandleItemAccess(u, key, n)
		}
		v, gerr := e.getItem(obj, key)
		if gerr != nil {
			// Jinja2's getitem returns Undefined when the base is None, is not
			// subscriptable, or rejects the key type; it never raises here.
			return NewUndefined(fmt.Sprintf("%s[%v]", e.getObjectName(obj), key), behavior, n), nil
		}
		return v, nil
	}
	return e.EvalNode(expr, ctx)
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

func (e *DefaultEvaluator) EvalDictNode(node *parser.DictNode, ctx Context) (interface{}, error) {
	result := make(map[string]interface{}, len(node.Keys))
	for i, keyNode := range node.Keys {
		key, err := e.EvalNode(keyNode, ctx)
		if err != nil {
			return nil, err
		}
		value, err := e.EvalNode(node.Values[i], ctx)
		if err != nil {
			return nil, err
		}
		result[DictKey(key)] = value
	}
	return result, nil
}

func (e *DefaultEvaluator) EvalAttributeNode(node *parser.AttributeNode, ctx Context) (interface{}, error) {
	if isListMethodName(node.Attribute) {
		if raw, ok := e.rawOperand(node.Object, ctx); ok {
			if b, ok := raw.(*listBox); ok {
				return b.method(e, node.Attribute), nil
			}
		}
	}

	obj, err := e.EvalNode(node.Object, ctx)
	if err != nil {
		return nil, err
	}

	// Handle undefined values with chained access
	if undefined, ok := obj.(*Undefined); ok && e.undefinedHandler != nil {
		return e.undefinedHandler.HandleAttributeAccess(undefined, node.Attribute, node)
	}

	if m, ok := e.builtinMethodFor(obj, node.Attribute); ok {
		return m, nil
	}

	if UsesReflectLookup(obj) {
		value, ok := e.lookupReflect(obj, node.Attribute)
		if !ok && e.undefinedHandler != nil {
			attrName := fmt.Sprintf("%s.%s", e.getObjectName(obj), node.Attribute)
			return e.undefinedHandler.Handle(attrName, node)
		}
		return value, nil
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

	v, err := e.getItem(obj, key)
	if err != nil {
		return nil, err
	}
	if _, missing := v.(*Undefined); missing && e.undefinedHandler != nil {
		// a missing key or index: strict mode fails, other modes get their own Undefined
		return e.undefinedHandler.Handle(fmt.Sprintf("%s[%v]", e.getObjectName(obj), key), node)
	}
	return v, nil
}

func (e *DefaultEvaluator) EvalFilterNode(node *parser.FilterNode, ctx Context) (interface{}, error) {
	var value interface{}
	var err error
	if node.FilterName == "default" || node.FilterName == "d" {
		value, err = e.evalLookup(node.Expression, ctx)
	} else {
		value, err = e.EvalNode(node.Expression, ctx)
	}
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

	if len(node.NamedArgs) > 0 {
		names := make([]string, 0, len(node.NamedArgs))
		for name := range node.NamedArgs {
			names = append(names, name)
		}
		sort.Strings(names)
		kwargs := make(map[string]interface{}, len(names))
		for _, name := range names {
			v, err := e.EvalNode(node.NamedArgs[name], ctx)
			if err != nil {
				return nil, err
			}
			kwargs[name] = v
		}
		applier, ok := filterContext(ctx).(KwargsFilterApplier)
		if !ok {
			return nil, fmt.Errorf("filter %q: keyword arguments need an environment", node.FilterName)
		}
		// The keywords travel as a trailing KwArgs so the markup policy treats
		// them like positional arguments.
		return e.applyFilterWithMarkup(node.FilterName, value, append(args[:len(args):len(args)], KwArgs(kwargs)), ctx,
			func(v interface{}, a []interface{}) (interface{}, error) {
				n := len(a) - 1
				return applier.ApplyFilterKwargs(node.FilterName, v, a[:n], map[string]interface{}(a[n].(KwArgs)))
			})
	}

	return e.applyFilterWithMarkup(node.FilterName, value, args, ctx, func(v interface{}, a []interface{}) (interface{}, error) {
		// Try to use environment's filter registry if available
		if envCtx, ok := envContext(ctx); ok {
			return envCtx.ApplyFilter(node.FilterName, v, a...)
		}
		// Fallback to basic filters
		return e.applyFilter(node.FilterName, v, a)
	})
}

func (e *DefaultEvaluator) EvalBinaryOpNode(node *parser.BinaryOpNode, ctx Context) (interface{}, error) {
	if node.Operator == "and" || node.Operator == "or" {
		left, err := e.EvalNode(node.Left, ctx)
		if err != nil {
			return nil, err
		}
		// Jinja2: `or` keeps a truthy left, `and` keeps a falsy left; otherwise the
		// right operand is the result and is only evaluated now.
		if e.isTruthy(left) == (node.Operator == "or") {
			return left, nil
		}
		return e.EvalNode(node.Right, ctx)
	}

	left, err := e.EvalNode(node.Left, ctx)
	if err != nil {
		return nil, err
	}

	right, err := e.EvalNode(node.Right, ctx)
	if err != nil {
		return nil, err
	}

	switch node.Operator {
	case "~":
		if markupContext(ctx) {
			if m, ok := concatMarkup(ctx, left, right); ok {
				return m, nil
			}
		}
		left, right = Unmark(left), Unmark(right)
	case "+":
		if !markupContext(ctx) {
			left, right = Unmark(left), Unmark(right)
			break
		}
		_, ls := basicValue(left).(string)
		_, rs := basicValue(right).(string)
		_, lm := AsMarkup(left)
		_, rm := AsMarkup(right)
		if (lm || rm) && (ls || lm) && (rs || rm) {
			m, _ := concatMarkup(ctx, left, right)
			return m, nil
		}
		left, right = Unmark(left), Unmark(right)
	case "*":
		_, lm := AsMarkup(left)
		_, rm := AsMarkup(right)
		if !markupContext(ctx) {
			left, right = Unmark(left), Unmark(right)
		} else if lm != rm {
			r, err := e.applyBinaryOpWithNode(node.Operator, Unmark(left), Unmark(right), node)
			if s, ok := r.(string); ok && err == nil {
				return Markup(s), nil
			}
			return r, err
		}
	case "==", "!=", "<", "<=", ">", ">=", "in", "not in":
		left, right = Unmark(left), Unmark(right)
	}
	return e.applyBinaryOpWithNode(node.Operator, left, right, node)
}

// EvalCompareNode evaluates a comparison chain such as a < b < c. Each
// operand is evaluated at most once, and evaluation stops at the first
// comparison that is false, as in Jinja2.
func (e *DefaultEvaluator) EvalCompareNode(node *parser.CompareNode, ctx Context) (interface{}, error) {
	left, err := e.EvalNode(node.Expr, ctx)
	if err != nil {
		return nil, err
	}
	for _, op := range node.Ops {
		right, err := e.EvalNode(op.Right, ctx)
		if err != nil {
			return nil, err
		}
		result, err := e.applyBinaryOpWithNode(op.Operator, Unmark(left), Unmark(right), node)
		if err != nil {
			return nil, err
		}
		if !e.isTruthy(result) {
			return false, nil
		}
		left = right
	}
	return true, nil
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

// childScoper is implemented by contexts that can create a cheap child
// scope (reads fall through to the parent, writes stay in the child).
type childScoper interface {
	Child() Context
}

func (e *DefaultEvaluator) EvalForNode(node *parser.ForNode, ctx Context) (interface{}, error) {
	return e.evalFor(node, ctx, 1)
}

func (e *DefaultEvaluator) evalFor(node *parser.ForNode, ctx Context, depth int) (interface{}, error) {
	iterable, err := e.EvalNode(node.Iterable, ctx)
	if err != nil {
		return nil, err
	}

	items, err := e.makeIterableForVariables(iterable, len(node.Variables))
	if err != nil {
		return nil, err
	}

	var results []string
	loopBroken := false

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

	if len(filteredItems) == 0 && len(node.Else) > 0 {
		// Jinja2 runs else when the loop body never ran, after the if-filter.
		return e.evalNodeList(node.Else, ctx)
	}

	for i, item := range filteredItems {
		var loopCtx Context
		if c, ok := ctx.(childScoper); ok {
			loopCtx = c.Child()
		} else {
			loopCtx = ctx.Clone()
		}

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

		// Every iteration gets its own map: a template can keep a reference to
		// `loop`, so it must never be shared or recycled.
		loopInfo := make(map[string]interface{}, 12)

		// Fill in the loop values
		loopInfo["index"] = i + 1
		loopInfo["index0"] = i
		loopInfo["revindex"] = len(filteredItems) - i
		loopInfo["revindex0"] = len(filteredItems) - i - 1
		loopInfo["first"] = i == 0
		loopInfo["last"] = i == len(filteredItems)-1
		loopInfo["length"] = len(filteredItems)
		loopInfo["depth"] = depth
		loopInfo["depth0"] = depth - 1
		loopInfo["previtem"] = previtem
		loopInfo["nextitem"] = nextitem
		if markupContext(ctx) {
			loopInfo["cycle"] = KeepMarkupFunc(cycleFunc)
		} else {
			loopInfo["cycle"] = cycleFunc
		}
		loopInfo["changed"] = changedFunc

		// Add recursive loop function if this is a recursive loop
		if node.Recursive {
			recursiveFunc := func(newIterable interface{}) (interface{}, error) {
				leave, err := e.enter("recursive loop")
				if err != nil {
					return nil, err
				}
				defer leave()

				// Handle undefined or nil iterables - just return empty string
				if IsUndefined(newIterable) || newIterable == nil {
					return markIfSafe("", ctx), nil
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
					Condition: node.Condition,
				}
				r, err := e.evalFor(recursiveNode, ctx, depth+1)
				if err != nil {
					return r, err
				}
				// The iteration output is template output of the loop's own context;
				// the call site must not decide again.
				return markIfSafe(ToString(r), ctx), nil
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
	return e.evalNodeList(node.Body, withBaseAutoescape(ctx))
}

func (e *DefaultEvaluator) EvalSetNode(node *parser.SetNode, ctx Context) (interface{}, error) {
	value, err := e.EvalNode(node.Value, ctx)
	if err != nil {
		return nil, err
	}

	if len(node.Targets) == 1 {
		// Single assignment - could be simple variable or attribute
		return e.assignToTarget(node.Targets[0], value, node.Value, ctx)
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
			if _, err := e.assignToTarget(target, items[i], nil, ctx); err != nil {
				return nil, err
			}
		}
	}

	return "", nil // Set statements don't produce output
}

// assignToTarget assigns a value to a target expression (identifier, attribute, or subscript)
func (e *DefaultEvaluator) assignToTarget(target parser.ExpressionNode, value interface{}, valueExpr parser.ExpressionNode, ctx Context) (interface{}, error) {
	switch t := target.(type) {
	case *parser.IdentifierNode:
		// Simple variable assignment
		ctx.SetVariable(t.Name, e.bindValue(valueExpr, value, ctx))
		return "", nil

	case *parser.AttributeNode:
		// Attribute assignment: obj.attr = value
		obj, err := e.EvalNode(t.Object, ctx)
		if err != nil {
			return nil, err
		}
		if _, ok := obj.(assignableNamespace); ok {
			value = e.bindValue(valueExpr, value, ctx)
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

	// Set the variable to the rendered content (Markup where escaping is on, as in Jinja2)
	ctx.SetVariable(node.Variable, markIfSafe(content, ctx))

	return "", nil // Block set statements don't produce output
}

func (e *DefaultEvaluator) EvalCallNode(node *parser.CallNode, ctx Context) (interface{}, error) {
	return e.evalCall(node, ctx, nil)
}

func (e *DefaultEvaluator) evalCall(node *parser.CallNode, ctx Context, caller *Macro) (interface{}, error) {
	function, err := e.EvalNode(node.Function, ctx)
	if err != nil {
		return nil, err
	}

	// A macro keeps the identity of the lists it is given
	_, isMacro := function.(*Macro)
	evalArg := e.EvalNode
	if isMacro {
		evalArg = func(n parser.Node, c Context) (interface{}, error) {
			if x, ok := n.(parser.ExpressionNode); ok {
				return e.callArg(x, c)
			}
			return e.EvalNode(n, c)
		}
	}

	// Evaluate arguments with pre-allocated capacity
	args := make([]interface{}, 0, len(node.Arguments))
	for _, arg := range node.Arguments {
		argValue, err := evalArg(arg, ctx)
		if err != nil {
			return nil, err
		}
		args = append(args, argValue)
	}

	// Evaluate keyword arguments with pre-allocated capacity
	kwargs := make(map[string]interface{}, len(node.Keywords))
	for key, value := range node.Keywords {
		argValue, err := evalArg(value, ctx)
		if err != nil {
			return nil, err
		}
		kwargs[key] = argValue
	}
	if caller != nil {
		kwargs["caller"] = caller
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

	var names []string
	switch v := templateNameExpr.(type) {
	case string:
		names = []string{v}
	case nil, *Undefined:
		// no name: nothing to find
	case []string:
		names = v
	case []interface{}:
		for _, item := range v {
			name, ok := item.(string)
			if !ok {
				return nil, fmt.Errorf("include template name must be a string or a list of strings, got %T in list", item)
			}
			names = append(names, name)
		}
	default:
		return nil, fmt.Errorf("include template name must be a string or a list of strings, got %T", templateNameExpr)
	}

	// Get the import system
	if e.importSystem == nil {
		return nil, fmt.Errorf("import system not initialized for includes")
	}

	// Load the first template that exists
	var templateAST *parser.TemplateNode
	var templateName string
	var lastErr error
	for _, name := range names {
		ast, err := e.importSystem.loader.LoadTemplate(name)
		if err == nil {
			templateAST, templateName = ast, name
			break
		}
		if !e.notFound(err, name) {
			return nil, fmt.Errorf("failed to load included template %q: %w", name, err)
		}
		lastErr = err
	}
	if templateAST == nil {
		if node.IgnoreMissing {
			return "", nil
		}
		switch len(names) {
		case 0:
			return nil, fmt.Errorf("none of the included templates %q were found: %w", names, ErrTemplateNotFound)
		case 1:
			return nil, fmt.Errorf("included template %q not found: %w", names[0], lastErr)
		}
		return nil, fmt.Errorf("none of the included templates %q were found: %w", names, lastErr)
	}

	// The include runs in a copy of the includer's scope: it reads the
	// includer's variables, and its own {% set %} stays inside.
	var includeCtx Context
	if node.WithoutContext {
		includeCtx = importContext(ctx, false)
	} else {
		if c, ok := ctx.(childScoper); ok {
			includeCtx = c.Child()
		} else {
			includeCtx = ctx.Clone()
		}
		if node.Context != nil {
			// If a context expression is provided, evaluate it
			contextValue, err := e.EvalNode(node.Context, ctx)
			if err != nil {
				return nil, fmt.Errorf("error evaluating context for include: %w", err)
			}

			// Add the provided values to the include's scope
			if contextMap, ok := contextValue.(map[string]interface{}); ok {
				for key, value := range contextMap {
					includeCtx.SetVariable(key, value)
				}
			}
		}
	}

	includeCtx = withBaseAutoescape(includeCtx)

	leave, err := e.enter("include " + strconv.Quote(templateName))
	if err != nil {
		return nil, err
	}
	defer leave()

	// Execute the included template with the appropriate context
	result, err := e.EvalNode(templateAST, includeCtx)
	if err != nil {
		return nil, fmt.Errorf("error executing included template %q: %w", templateName, err)
	}

	return result, nil
}

// notFound reports whether err means that template name itself does not
// exist. A TemplateNotFoundError for another name (a missing parent of an
// included template) is a real error, as in Jinja2. A custom TemplateLoader
// that does not use TemplateNotFoundError keeps the old rule: it is "not
// found" when the loader says the template does not exist.
func (e *DefaultEvaluator) notFound(err error, name string) bool {
	var nf *TemplateNotFoundError
	if errors.As(err, &nf) {
		return nf.Name == name
	}
	if _, builtin := e.importSystem.loader.(*SimpleTemplateLoader); builtin {
		return false
	}
	return !e.importSystem.loader.TemplateExists(name)
}

func (e *DefaultEvaluator) EvalSuperNode(node *parser.SuperNode, ctx Context) (interface{}, error) {
	// Super nodes are handled by the inheritance resolver during template compilation
	// In the runtime context, they should have been replaced with parent block content
	// If we reach here, it means the super() call wasn't resolved properly
	return "", fmt.Errorf("super() call outside of block context")
}

func (e *DefaultEvaluator) EvalMacroNode(node *parser.MacroNode, ctx Context) (interface{}, error) {
	defaults := make(map[string]interface{}, len(node.Defaults))
	for name, expr := range node.Defaults {
		defaults[name] = expr
	}
	// The macro body runs in the context it is defined in, as in Jinja2.
	ctx.SetVariable(node.Name, &Macro{
		Name:       node.Name,
		Parameters: node.Parameters,
		Defaults:   defaults,
		Body:       node.Body,
		MacroUsage: node.MacroUsage,
		defCtx:     ctx,
		evaluator:  e,
	})

	// Macro definitions don't produce output during evaluation
	return "", nil
}

func (e *DefaultEvaluator) EvalTestNode(node *parser.TestNode, ctx Context) (interface{}, error) {
	var value interface{}
	var err error
	switch node.TestName {
	case "defined", "undefined", "none":
		value, err = e.evalLookup(node.Expression, ctx)
	default:
		value, err = e.EvalNode(node.Expression, ctx)
	}
	if err != nil {
		return nil, err
	}
	if node.TestName == "defined" {
		result := !IsUndefined(value)
		if node.Negated {
			result = !result
		}
		return result, nil
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

	if node.TestName != "escaped" {
		value = Unmark(value)
		args = unmarkArgs(args)
	}

	// Try to use environment's test registry if available
	var result bool
	if envCtx, ok := envContext(ctx); ok {
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
	}
	if node.FalseExpr == nil {
		// {{ a if b }} with b false is undefined, as in Jinja2
		if e.undefinedHandler != nil {
			return e.undefinedHandler.Handle("inline if-expression without else", node)
		}
		return nil, nil
	}
	return e.EvalNode(node.FalseExpr, ctx)
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

			keyStr := DictKey(key)
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
	if _, ok := e.builtinMethodFor(obj, attr); ok {
		return true
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
		return isDictMethodName(attr)
	case map[string]string:
		if _, ok := v[attr]; ok {
			return true
		}
		return isDictMethodName(attr)
	case GroupTuple:
		return attr == "grouper" || attr == "list"
	default:
		_, ok := e.lookupReflect(obj, attr)
		return ok
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
		for i := 0; rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface; i++ {
			if rv.IsNil() || i >= maxDeref {
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
//
// Deprecated: nothing creates DictItems since d.items() returns a list; it will be removed in a future release.
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
	if m, ok := e.builtinMethodFor(obj, attr); ok {
		return m
	}

	switch v := obj.(type) {
	case NamespaceInterface:
		val, ok := v.Get(attr)
		if !ok {
			return nil
		}
		return Unbox(val)
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
		case "get":
			return mapGetMethod(v)
		case "items":
			// Return a callable function that returns the [key, value] pairs
			return func(args ...interface{}) (interface{}, error) {
				sorted := sortedKeys(v)
				pairs := make([]interface{}, 0, len(sorted))
				for _, k := range sorted {
					pairs = append(pairs, []interface{}{k, v[k]})
				}
				return pairs, nil
			}
		case "keys":
			// Return a callable function that returns the keys
			return func(args ...interface{}) (interface{}, error) {
				sorted := sortedKeys(v)
				keys := make([]interface{}, 0, len(sorted))
				for _, k := range sorted {
					keys = append(keys, k)
				}
				return keys, nil
			}
		case "values":
			// Return a callable function that returns the values
			return func(args ...interface{}) (interface{}, error) {
				values := make([]interface{}, 0, len(v))
				for _, k := range sortedKeys(v) {
					values = append(values, v[k])
				}
				return values, nil
			}
		default:
			return nil
		}
	case map[string]string:
		if val, ok := v[attr]; ok {
			return val
		}
		if attr == "get" {
			return dictMethodOver(reflect.ValueOf(v), "get")
		}
		if isDictMethodName(attr) {
			return e.getAttribute(stringMap(v), attr)
		}
		return ""
	case GroupTuple:
		switch attr {
		case "grouper":
			return v[0]
		case "list":
			return v[1]
		}
		return nil
	default:
		val, _ := e.lookupReflect(obj, attr)
		return val
	}
}

// lookupReflect finds attr on a Go value that is not one of the runtime's own
// types: a struct field (exact name; then, for non-embedded exported fields
// that do not hold a func, the capitalised name), an entry of a map with
// string-kind or interface keys (never a func-valued entry), a dict method of
// a map whose keys are all strings (keys/values/items), or a slice element by
// numeric attr. It never returns a Go method.
func (e *DefaultEvaluator) lookupReflect(obj interface{}, attr string) (interface{}, bool) {
	if _, ok := Indirect(obj).(RoundedFloat); ok {
		return nil, false
	}
	rv := reflect.ValueOf(obj)
	for i := 0; ; i++ {
		if !rv.IsValid() {
			return nil, false
		}
		if rv.Kind() != reflect.Ptr && rv.Kind() != reflect.Interface {
			break
		}
		if i >= maxDeref || rv.IsNil() {
			return nil, false
		}
		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.Struct:
		if v, ok := structField(rv, attr, false); ok {
			return v, true
		}
		if capital := capitalizeFirst(attr); capital != attr {
			return structField(rv, capital, true)
		}
	case reflect.Map:
		kt := rv.Type().Key()
		var key reflect.Value
		switch kt.Kind() {
		case reflect.String:
			key = reflect.ValueOf(attr).Convert(kt)
		case reflect.Interface:
			key = reflect.ValueOf(attr)
			if !key.Type().AssignableTo(kt) {
				return nil, false
			}
		default:
			if isDictMethodName(attr) {
				return dictMethodOver(rv, attr), true
			}
			return nil, false
		}
		if val := rv.MapIndex(key); val.IsValid() {
			inner := val
			if inner.Kind() == reflect.Interface {
				inner = inner.Elem()
			}
			if inner.Kind() == reflect.Func {
				return nil, false
			}
			return val.Interface(), true
		}
		if isDictMethodName(attr) {
			if attr == "get" {
				return dictMethodOver(rv, attr), true
			}
			if m, ok := StringKeyMap(obj); ok && e.attributeExists(m, attr) {
				return e.getAttribute(m, attr), true
			}
			if kt.Kind() == reflect.Interface {
				return dictMethodOver(rv, attr), true
			}
		}
	case reflect.Slice, reflect.Array:
		if index, err := strconv.Atoi(attr); err == nil && index >= 0 && index < rv.Len() {
			return rv.Index(index).Interface(), true
		}
	}
	return nil, false
}

// structField reads the exported field name of the struct rv. Fields promoted
// through a nil embedded pointer are not found. With fallback set (the
// capitalised name for a lower-case attribute), embedded fields and func
// fields are not found: those are reachable only by their exact name.
func structField(rv reflect.Value, name string, fallback bool) (interface{}, bool) {
	sf, ok := rv.Type().FieldByName(name)
	if !ok || sf.PkgPath != "" {
		return nil, false
	}
	if fallback && (sf.Anonymous || sf.Type.Kind() == reflect.Func) {
		return nil, false
	}
	f, err := rv.FieldByIndexErr(sf.Index)
	if err != nil || !f.CanInterface() {
		return nil, false
	}
	return f.Interface(), true
}

// DictKey is the string a dict literal, subscript, `in` test or .get() uses
// for key.
func DictKey(key interface{}) string {
	key = Indirect(key)
	switch k := key.(type) {
	case nil:
		return "None"
	case RoundedFloat:
		return fmt.Sprintf("%v", k.Value)
	}
	return fmt.Sprintf("%v", key)
}

func (e *DefaultEvaluator) getItem(obj, key interface{}) (interface{}, error) {
	if obj == nil {
		return nil, fmt.Errorf("cannot get item from nil")
	}
	key = Unmark(key)
	if m, ok := AsMarkup(obj); ok {
		r, err := e.getItem(string(m), key)
		if s, ok := r.(string); ok {
			return Markup(s), err
		}
		return r, err
	}

	switch v := obj.(type) {
	case map[string]interface{}:
		keyStr := DictKey(key)
		val, ok := v[keyStr]
		if !ok {
			return NewUndefined(keyStr, UndefinedSilent, nil), nil
		}
		return val, nil
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
		n := keyInt
		if n < 0 {
			n += utf8.RuneCountInString(v)
		}
		if n >= 0 {
			for _, r := range v {
				if n == 0 {
					return string(r), nil
				}
				n--
			}
		}
		// Return undefined for out of bounds access (Jinja2 behavior)
		return NewUndefined(fmt.Sprintf("index[%d]", keyInt), UndefinedSilent, nil), nil
	default:
		// Try reflection for slice, array and map access
		obj = Indirect(obj)
		if obj == nil {
			return nil, fmt.Errorf("cannot get item from nil")
		}
		rv := reflect.ValueOf(obj)
		switch rv.Kind() {
		case reflect.Slice, reflect.Array:
			keyInt, err := e.toInt(key)
			if err != nil {
				return nil, fmt.Errorf("index must be integer, got %T", key)
			}
			// Support negative indexing like Python
			if keyInt < 0 {
				keyInt += rv.Len()
			}
			if keyInt < 0 || keyInt >= rv.Len() {
				// Return undefined for out of bounds access (Jinja2 behavior)
				return NewUndefined(fmt.Sprintf("index[%d]", keyInt), UndefinedSilent, nil), nil
			}
			return rv.Index(keyInt).Interface(), nil
		case reflect.Map:
			if val, ok := mapLookup(rv, key); ok {
				return val.Interface(), nil
			}
			return NewUndefined(fmt.Sprintf("%v", key), UndefinedSilent, nil), nil
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
		if !e.isTruthy(left) {
			return left, nil
		}
		return right, nil
	case "or":
		if e.isTruthy(left) {
			return left, nil
		}
		return right, nil
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
		var defaultValue interface{} = ""
		if len(args) > 0 {
			defaultValue = args[0]
		}
		if len(args) > 1 && e.isTruthy(args[1]) {
			if e.isTruthy(value) {
				return value, nil
			}
			return defaultValue, nil
		}
		if _, undef := value.(*Undefined); undef || value == nil || value == "" {
			return defaultValue, nil
		}
		return value, nil
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
	if fv := reflect.ValueOf(function); fv.Kind() == reflect.Func && fv.IsNil() {
		return nil, fmt.Errorf("cannot call nil function")
	}

	// Handle special callable objects
	switch fn := function.(type) {
	case *Joiner:
		// Joiner objects are directly callable
		if fn.Safe && markupContext(ctx) {
			return Markup(fn.Join()), nil
		}
		return fn.Join(), nil
	case *CallableLoop:
		// CallableLoop objects support function calls for recursion
		r, err := fn.Call(args...)
		if err != nil {
			return r, err
		}
		// recursiveFunc already marked the output with the loop's own context
		return r, nil
	case *Macro:
		return fn.Call(args, kwargs)
	case *builtinMethod:
		return fn.fn(args, kwargs)
	case *Undefined:
		if fn.Hint != "" {
			return nil, fmt.Errorf("cannot call undefined %s: %s", fn.Name, fn.Hint)
		}
	case KeepMarkupFunc:
		if len(kwargs) > 0 {
			args = append(args, kwargs)
		}
		return fn(args...)
	}

	// Go functions get plain strings for safe values
	args, kwargs = unmarkCall(args, kwargs)

	// Handle different function types
	switch fn := function.(type) {
	case func(Context, ...interface{}) (interface{}, error):
		// Macro function that takes context as first parameter
		if ctx != nil {
			return fn(newGoContext(ctx), args...)
		}
		// If no context provided, use a default empty context
		emptyCtx := &simpleContext{variables: make(map[string]interface{})}
		return fn(newGoContext(emptyCtx), args...)

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

// sortedKeys returns the keys of m in sorted (code-point) order. Go maps
// have no insertion order, so miya walks dicts by sorted key (Jinja2 uses
// insertion order).
func sortedKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func (e *DefaultEvaluator) makeIterable(obj interface{}) ([]interface{}, error) {
	if obj == nil {
		return nil, nil
	}
	// Iterating Markup yields plain characters, as in Jinja2
	obj = Unmark(obj)
	obj = Indirect(obj)
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
		// Return keys in sorted order (Python semantics)
		result := make([]interface{}, 0, len(v))
		for _, key := range sortedKeys(v) {
			result = append(result, key)
		}
		return result, nil

	case *DictItems:
		// Handle DictItems objects - return key-value pairs for iteration
		result := make([]interface{}, 0, len(v.data))
		for _, key := range sortedKeys(v.data) {
			result = append(result, []interface{}{key, v.data[key]})
		}
		return result, nil

	case string:
		result := make([]interface{}, 0, utf8.RuneCountInString(v))
		for _, r := range v {
			result = append(result, string(r))
		}
		return result, nil
	}

	// Slow path: Check if obj is a function that should be called to get an iterable
	if fnValue := reflect.ValueOf(obj); fnValue.Kind() == reflect.Func {
		if fnValue.IsNil() {
			return nil, fmt.Errorf("cannot iterate over nil function")
		}
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

	// Typed maps: string keys iterate like a dict, other keys in a fixed order
	if m, ok := StringKeyMap(obj); ok {
		return e.makeIterable(m)
	}
	rv := reflect.ValueOf(obj)
	if rv.Kind() == reflect.Map {
		return sortedMapKeys(rv), nil
	}

	// Fallback to reflection for uncommon slice/array types
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
// For maps with multiple variables, returns key-value pairs; otherwise returns keys
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
		if fnValue.IsNil() {
			return nil, fmt.Errorf("cannot iterate over nil function")
		}
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
		for _, key := range sortedKeys(v.data) {
			result = append(result, []interface{}{key, v.data[key]})
		}
		return result, nil
	case map[string]interface{}:
		if numVariables == 2 {
			// Return key-value pairs for unpacking
			// Pre-allocate with known capacity to avoid reallocations
			result := make([]interface{}, 0, len(v))
			for _, key := range sortedKeys(v) {
				result = append(result, []interface{}{key, v[key]})
			}
			return result, nil
		} else if numVariables == 1 {
			// Return keys in sorted order for single variable
			// Pre-allocate with known capacity to avoid reallocations
			result := make([]interface{}, 0, len(v))
			for _, key := range sortedKeys(v) {
				result = append(result, key)
			}
			return result, nil
		} else {
			return nil, fmt.Errorf("cannot unpack map into %d variables (expected 1 or 2)", numVariables)
		}
	default:
		if m, ok := StringKeyMap(obj); ok {
			return e.makeIterableForVariables(m, numVariables)
		}
		if rv := reflect.ValueOf(Indirect(obj)); rv.Kind() == reflect.Map {
			switch numVariables {
			case 1:
				return sortedMapKeys(rv), nil
			case 2:
				return sortedMapPairs(rv), nil
			}
			return nil, fmt.Errorf("cannot unpack map into %d variables (expected 1 or 2)", numVariables)
		}
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
	case RoundedFloat:
		return v.Value != 0
	case bool:
		return v
	case string:
		return v != ""
	case Markup:
		return v != ""
	case SafeValue:
		return e.isTruthy(v.Value)
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
		if rv.Kind() == reflect.Ptr {
			inner := Indirect(obj)
			if inner == nil {
				return false
			}
			if reflect.ValueOf(inner).Kind() != reflect.Ptr {
				return e.isTruthy(inner)
			}
			return true
		}
		switch rv.Kind() {
		case reflect.Bool:
			return rv.Bool()
		case reflect.String:
			return rv.Len() > 0
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return rv.Int() != 0
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			return rv.Uint() != 0
		case reflect.Float32, reflect.Float64:
			return rv.Float() != 0
		case reflect.Slice, reflect.Map, reflect.Array:
			return rv.Len() > 0
		default:
			return true
		}
	}
}

// maxConcatLength bounds the list a single + can build.
const maxConcatLength = 10000000

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
	aStr, aStrOk := basicValue(a).(string)
	bStr, bStrOk := basicValue(b).(string)
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
	if a != nil && b != nil && reflect.TypeOf(a).Kind() == reflect.Slice && reflect.TypeOf(b).Kind() == reflect.Slice {
		av := reflect.ValueOf(a)
		bv := reflect.ValueOf(b)
		if av.Len()+bv.Len() > maxConcatLength {
			return nil, &SizeError{Msg: fmt.Sprintf("list concatenation too large: %d items (limit %d)", av.Len()+bv.Len(), maxConcatLength)}
		}
		if av.Type() != bv.Type() {
			merged := make([]interface{}, 0, av.Len()+bv.Len())
			for i := 0; i < av.Len(); i++ {
				merged = append(merged, av.Index(i).Interface())
			}
			for i := 0; i < bv.Len(); i++ {
				merged = append(merged, bv.Index(i).Interface())
			}
			return merged, nil
		}
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
	// A string literal operand repeats even when it looks like a number
	if op, ok := node.(*parser.BinaryOpNode); ok {
		isStrLit := func(n parser.ExpressionNode) bool {
			lit, ok := n.(*parser.LiteralNode)
			if !ok {
				return false
			}
			_, isStr := lit.Value.(string)
			return isStr
		}
		if isStrLit(op.Left) {
			return repeatValue(a, b)
		}
		if isStrLit(op.Right) {
			return repeatValue(b, a)
		}
	}
	return e.multiply(a, b)
}

func (e *DefaultEvaluator) moduloWithNode(a, b interface{}, node parser.Node) (interface{}, error) {
	return e.modulo(a, b)
}

func (e *DefaultEvaluator) powerWithNode(a, b interface{}, node parser.Node) (interface{}, error) {
	return e.power(a, b)
}

func (e *DefaultEvaluator) concatenateWithNode(a, b interface{}, node parser.Node) (interface{}, error) {
	return ToString(a) + ToString(b), nil
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
	if seq, count, ok := sequenceOperands(a, b); ok {
		return repeatValue(seq, count)
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
	ai, bi, af, bf, isInt, ok := pyOperands(a, b)
	if !ok {
		return nil, NewTypeError("floor division", a, node).WithContext(fmt.Sprintf("cannot floor divide %T and %T", a, b))
	}
	if (isInt && bi == 0) || (!isInt && bf == 0) {
		return nil, NewMathError("floor division", fmt.Errorf("division by zero"), node)
	}
	if isInt {
		q := ai / bi
		if ai%bi != 0 && (ai < 0) != (bi < 0) {
			q--
		}
		return int(q), nil
	}
	floordiv, _ := pyFloatDivMod(af, bf)
	return floordiv, nil
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

	ai, bi, af, bf, isInt, ok := pyOperands(a, b)
	if !ok {
		return nil, fmt.Errorf("cannot modulo %T and %T", a, b)
	}
	if (isInt && bi == 0) || (!isInt && bf == 0) {
		return nil, fmt.Errorf("modulo by zero")
	}
	if isInt {
		r := ai % bi
		if r != 0 && (r < 0) != (bi < 0) {
			r += bi
		}
		return int(r), nil
	}
	_, mod := pyFloatDivMod(af, bf)
	return mod, nil
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
	if i, u, f, kind, ok := numericValue(a); ok {
		switch kind {
		case 'i':
			if i == math.MinInt64 {
				return -float64(i), nil
			}
			return -i, nil
		case 'u':
			if u <= math.MaxInt64 {
				return -int64(u), nil
			}
			return -float64(u), nil
		case 'f':
			return -f, nil
		}
	}
	aFloat, err := e.toFloat(a)
	if err == nil {
		return -aFloat, nil
	}
	return nil, fmt.Errorf("cannot negate %T", a)
}

// numericValue reports a value of any Go numeric kind (named types included).
// Signed integers return kind 'i', unsigned integers 'u' and floats 'f', so
// integers can be compared exactly and only floats lose precision. Pointers
// are followed, and bools count as 0 and 1 as in Python. String, nil and
// every other kind return ok=false.
func numericValue(v interface{}) (i int64, u uint64, f float64, kind byte, ok bool) {
	if v == nil {
		return
	}
	v = Indirect(v)
	if v == nil {
		return
	}
	if r, isR := v.(RoundedFloat); isR {
		return 0, 0, r.Value, 'f', true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Bool:
		if rv.Bool() {
			return 1, 0, 0, 'i', true
		}
		return 0, 0, 0, 'i', true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int(), 0, 0, 'i', true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return 0, rv.Uint(), 0, 'u', true
	case reflect.Float32, reflect.Float64:
		return 0, 0, rv.Float(), 'f', true
	}
	return
}

// pyOperands classifies a and b for Python's // and %: both integers (kind
// 'i', or 'u' that fits in int64) use exact int64 arithmetic; anything else
// numeric uses float64. Non-numbers (including numeric strings) are not ok.
func pyOperands(a, b interface{}) (ai, bi int64, af, bf float64, isInt, ok bool) {
	i1, u1, f1, k1, ok1 := numericValue(a)
	i2, u2, f2, k2, ok2 := numericValue(b)
	if !ok1 || !ok2 {
		return
	}
	ok = true
	if k1 == 'u' && u1 <= math.MaxInt64 {
		i1, k1 = int64(u1), 'i'
	}
	if k2 == 'u' && u2 <= math.MaxInt64 {
		i2, k2 = int64(u2), 'i'
	}
	if k1 == 'i' && k2 == 'i' {
		return i1, i2, 0, 0, true, true
	}
	af = pyFloatOf(i1, u1, f1, k1)
	bf = pyFloatOf(i2, u2, f2, k2)
	return
}

func pyFloatOf(i int64, u uint64, f float64, kind byte) float64 {
	switch kind {
	case 'i':
		return float64(i)
	case 'u':
		return float64(u)
	}
	return f
}

// pyFloatDivMod is CPython's float_divmod: floor division and a modulo with
// the sign of the divisor, corrected so that 1 // 0.1 == 9.
func pyFloatDivMod(vx, wx float64) (floordiv, mod float64) {
	mod = math.Mod(vx, wx)
	div := (vx - mod) / wx
	if mod != 0 {
		if (wx < 0) != (mod < 0) {
			mod += wx
			div -= 1.0
		}
	} else {
		mod = math.Copysign(0, wx)
	}
	if div != 0 {
		floordiv = math.Floor(div)
		if div-floordiv > 0.5 {
			floordiv += 1.0
		}
	} else {
		floordiv = math.Copysign(0, vx/wx)
	}
	return floordiv, mod
}

// compareNumbers orders two numbers of any Go numeric type. ok is false unless
// both values are numbers.
func compareNumbers(a, b interface{}) (cmp int, ok bool) {
	ai, au, af, ak, aok := numericValue(a)
	bi, bu, bf, bk, bok := numericValue(b)
	if !aok || !bok {
		return 0, false
	}
	switch {
	case ak == 'i' && bk == 'i':
		switch {
		case ai < bi:
			return -1, true
		case ai > bi:
			return 1, true
		}
		return 0, true
	case ak == 'u' && bk == 'u':
		switch {
		case au < bu:
			return -1, true
		case au > bu:
			return 1, true
		}
		return 0, true
	case ak == 'i' && bk == 'u':
		if ai < 0 {
			return -1, true
		}
		return compareNumbers(uint64(ai), bu)
	case ak == 'u' && bk == 'i':
		if bi < 0 {
			return 1, true
		}
		return compareNumbers(au, uint64(bi))
	}
	if ak == 'i' {
		af = float64(ai)
	} else if ak == 'u' {
		af = float64(au)
	}
	if bk == 'i' {
		bf = float64(bi)
	} else if bk == 'u' {
		bf = float64(bu)
	}
	switch {
	case af < bf:
		return -1, true
	case af > bf:
		return 1, true
	case af == bf:
		return 0, true
	}
	// NaN is unordered: report it as neither less nor equal.
	return 2, true
}

// CompareNumbers orders two numbers of any Go numeric type; ok is false
// unless both are numbers. cmp is 2 when either side is NaN. Filters use it
// to sort and compare like Jinja2.
func CompareNumbers(a, b interface{}) (cmp int, ok bool) { return compareNumbers(a, b) }

// Equal reports whether a == b holds in a template: numbers by value, lists
// element by element. Tests such as eq and in use it so they agree with ==.
func Equal(a, b interface{}) bool { var e DefaultEvaluator; return e.equal(a, b) }

// Comparison operations
func (e *DefaultEvaluator) equal(a, b interface{}) bool {
	a, b = Unmark(a), Unmark(b)
	if cmp, ok := compareNumbers(a, b); ok {
		return cmp == 0
	}
	a, b = basicValue(a), basicValue(b)
	if a != nil && b != nil {
		av, bv := reflect.ValueOf(a), reflect.ValueOf(b)
		ak, bk := av.Kind(), bv.Kind()
		if (ak == reflect.Slice || ak == reflect.Array) && (bk == reflect.Slice || bk == reflect.Array) {
			if av.Len() != bv.Len() {
				return false
			}
			if ak == reflect.Slice && bk == reflect.Slice && av.Len() > 0 && av.Pointer() == bv.Pointer() {
				return true
			}
			for i := 0; i < av.Len(); i++ {
				if !e.equal(av.Index(i).Interface(), bv.Index(i).Interface()) {
					return false
				}
			}
			return true
		}
	}
	return reflect.DeepEqual(a, b)
}

func (e *DefaultEvaluator) less(a, b interface{}) (bool, error) {
	if cmp, ok := compareNumbers(a, b); ok {
		return cmp < 0, nil
	}

	// String comparison
	if aStr, ok := basicValue(a).(string); ok {
		if bStr, ok := basicValue(b).(string); ok {
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
	return e.less(b, a)
}

func (e *DefaultEvaluator) greaterEqual(a, b interface{}) (bool, error) {
	greater, err := e.less(b, a)
	if err != nil {
		return false, err
	}
	return greater || e.equal(a, b), nil
}

func (e *DefaultEvaluator) contains(container, item interface{}) (bool, error) {
	rawItem := item
	container, item = basicValue(Indirect(container)), basicValue(Indirect(item))
	switch v := container.(type) {
	case *Undefined:
		return false, nil
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
		_, ok := v[DictKey(item)]
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
			_, found := mapLookup(rv, rawItem)
			return found, nil
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
	case Markup:
		return e.length(string(v))
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
	case RoundedFloat:
		return int(v.Value), nil
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	case string:
		return strconv.Atoi(v)
	default:
		i, u, f, kind, ok := numericValue(obj)
		if ok {
			switch kind {
			case 'i':
				return int(i), nil
			case 'u':
				if u <= math.MaxInt {
					return int(u), nil
				}
			case 'f':
				return int(f), nil
			}
		}
		return 0, fmt.Errorf("cannot convert %T to int", obj)
	}
}

func (e *DefaultEvaluator) toFloat(obj interface{}) (float64, error) {
	switch v := obj.(type) {
	case RoundedFloat:
		return v.Value, nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case float64:
		return v, nil
	case string:
		f, err := strconv.ParseFloat(v, 64)
		if err == nil && (math.IsInf(f, 0) || math.IsNaN(f)) {
			return 0, fmt.Errorf("cannot convert %q to float", v)
		}
		return f, err
	default:
		i, u, f, kind, ok := numericValue(obj)
		if ok {
			switch kind {
			case 'i':
				return float64(i), nil
			case 'u':
				return float64(u), nil
			case 'f':
				return f, nil
			}
		}
		return 0, fmt.Errorf("cannot convert %T to float", obj)
	}
}

// setAttribute sets an attribute on an object (for assignment expressions).
// Only namespace() objects accept attribute assignment.
func (e *DefaultEvaluator) setAttribute(obj interface{}, attr string, value interface{}) error {
	if obj == nil {
		return fmt.Errorf("cannot set attribute on nil object")
	}

	switch v := obj.(type) {
	case assignableNamespace:
		v.Set(attr, value)
		return nil
	}

	return fmt.Errorf("cannot assign attribute on non-namespace object (%T)", obj)
}

// setItem rejects item assignment: templates cannot modify maps or slices.
func (e *DefaultEvaluator) setItem(obj, key, value interface{}) error {
	if obj == nil {
		return fmt.Errorf("cannot set item on nil object")
	}
	return fmt.Errorf("cannot assign to an item of %T; use namespace() for mutable state", obj)
}

// slice implements Python-style slicing
func (e *DefaultEvaluator) slice(obj interface{}, start, end, step *int) (interface{}, error) {
	if obj == nil {
		return nil, fmt.Errorf("cannot slice nil")
	}
	if m, ok := AsMarkup(obj); ok {
		return Markup(e.sliceString(string(m), start, end, step)), nil
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

// sliceBounds resolves Python slice bounds for a sequence of length n, as
// CPython's PySlice_AdjustIndices does. Iterate i from start while
// i < stop (step > 0) or i > stop (step < 0). The caller rejects step 0.
func sliceBounds(n int, start, end, step *int) (int, int, int) {
	stepVal := 1
	if step != nil {
		stepVal = *step
	}
	lower, upper := 0, n
	if stepVal < 0 {
		lower, upper = -1, n-1
	}
	adjust := func(p *int, def int) int {
		if p == nil {
			return def
		}
		v := *p
		if v < 0 {
			v += n
			if v < lower {
				v = lower
			}
		} else if v > upper {
			v = upper
		}
		return v
	}
	if stepVal < 0 {
		return adjust(start, upper), adjust(end, lower), stepVal
	}
	return adjust(start, lower), adjust(end, upper), stepVal
}

func (e *DefaultEvaluator) sliceString(s string, start, end, step *int) string {
	r := []rune(s)
	startVal, endVal, stepVal := sliceBounds(len(r), start, end, step)

	var result strings.Builder
	for i := startVal; (stepVal > 0 && i < endVal) || (stepVal < 0 && i > endVal); i += stepVal {
		result.WriteRune(r[i])
	}
	return result.String()
}

func (e *DefaultEvaluator) sliceSlice(s []interface{}, start, end, step *int) []interface{} {
	startVal, endVal, stepVal := sliceBounds(len(s), start, end, step)

	result := []interface{}{}
	for i := startVal; (stepVal > 0 && i < endVal) || (stepVal < 0 && i > endVal); i += stepVal {
		result = append(result, s[i])
	}
	return result
}

func (e *DefaultEvaluator) sliceReflect(rv reflect.Value, start, end, step *int) interface{} {
	startVal, endVal, stepVal := sliceBounds(rv.Len(), start, end, step)

	result := reflect.MakeSlice(reflect.SliceOf(rv.Type().Elem()), 0, 0)
	for i := startVal; (stepVal > 0 && i < endVal) || (stepVal < 0 && i > endVal); i += stepVal {
		result = reflect.Append(result, rv.Index(i))
	}
	return result.Interface()
}

// autoescapeContext wraps a Context to track autoescape state
type autoescapeContext struct {
	Context
	autoescape bool
}

// goContext is the context a Go function receives: it never hands out a
// listBox. Unwrap lets filterContext and resolveEscape find the real context.
type goContext struct{ inner Context }

func (g *goContext) GetVariable(key string) (interface{}, bool) {
	v, ok := g.inner.GetVariable(key)
	return Unbox(v), ok
}

func (g *goContext) SetVariable(key string, value interface{}) { g.inner.SetVariable(key, value) }

func (g *goContext) Clone() Context { return newGoContext(g.inner.Clone()) }

// All returns a new map: the inner map may be shared or cached.
func (g *goContext) All() map[string]interface{} {
	all := g.inner.All()
	out := make(map[string]interface{}, len(all))
	for k, v := range all {
		out[k] = Unbox(v)
	}
	return out
}

func (g *goContext) Unwrap() Context { return g.inner }

// goAutoescapeContext also reports the inner context's autoescape setting.
type goAutoescapeContext struct{ goContext }

func (g *goAutoescapeContext) Clone() Context { return newGoContext(g.inner.Clone()) }

func (g *goAutoescapeContext) IsAutoescapeEnabled() bool {
	return g.inner.(AutoescapeContext).IsAutoescapeEnabled()
}

// goEnvContext also forwards the environment-aware interfaces.
type goEnvContext struct{ goAutoescapeContext }

func (g *goEnvContext) Clone() Context { return newGoContext(g.inner.Clone()) }

func (g *goEnvContext) ApplyFilter(name string, value interface{}, args ...interface{}) (interface{}, error) {
	return g.inner.(EnvironmentContext).ApplyFilter(name, value, args...)
}

func (g *goEnvContext) ApplyTest(name string, value interface{}, args ...interface{}) (bool, error) {
	return g.inner.(EnvironmentContext).ApplyTest(name, value, args...)
}

func (g *goEnvContext) ApplyFilterKwargs(name string, value interface{}, args []interface{}, kwargs map[string]interface{}) (interface{}, error) {
	return g.inner.(KwargsFilterApplier).ApplyFilterKwargs(name, value, args, kwargs)
}

// newGoContext wraps ctx for a Go function. Go cannot add methods
// conditionally, so the wrapper forwards the optional interfaces only for the
// combinations the library's own contexts have; any other is reachable
// through Unwrap.
func newGoContext(ctx Context) Context {
	_, esc := ctx.(AutoescapeContext)
	_, env := ctx.(EnvironmentContext)
	_, kw := ctx.(KwargsFilterApplier)
	g := goContext{inner: ctx}
	switch {
	case esc && env && kw:
		return &goEnvContext{goAutoescapeContext{g}}
	case esc && !env && !kw:
		return &goAutoescapeContext{g}
	}
	return &g
}

// filterContext peels context wrappers (autoescapeContext, ContextWrapper) so
// that filter and test lookups reach the environment-aware context underneath.
func filterContext(ctx Context) Context {
	for c := ctx; c != nil; {
		if _, ok := c.(EnvironmentContext); ok {
			return c
		}
		u, ok := c.(interface{ Unwrap() Context })
		if !ok {
			break
		}
		c = u.Unwrap()
	}
	return ctx
}

// Unwrap returns the wrapped context.
func (ac *autoescapeContext) Unwrap() Context { return ac.Context }

// IsAutoescapeEnabled returns whether autoescape is enabled in this context
func (ac *autoescapeContext) IsAutoescapeEnabled() bool {
	return ac.autoescape
}

// Clone keeps the block's autoescape setting for loops and with-blocks inside it.
func (ac *autoescapeContext) Clone() Context {
	return &autoescapeContext{Context: ac.Context.Clone(), autoescape: ac.autoescape}
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
	call, ok := node.Call.(*parser.CallNode)
	if !ok {
		// The parser rejects this; only a hand-built AST gets here.
		return nil, fmt.Errorf("{%% call %%} requires a call expression, got %s", node.Call.String())
	}
	defaults := make(map[string]interface{}, len(node.Defaults))
	for name, expr := range node.Defaults {
		defaults[name] = expr
	}
	// caller() renders the block body in the call site's context.
	caller := &Macro{
		Name:       "caller",
		Parameters: node.Parameters,
		Defaults:   defaults,
		Body:       node.Body,
		MacroUsage: node.MacroUsage,
		defCtx:     ctx,
		evaluator:  e,
	}
	return e.evalCall(call, ctx, caller)
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

	namespace, err := e.loadModule(templateName, ctx, node.WithContext)
	if err != nil {
		return nil, err
	}
	ctx.SetVariable(node.Alias, e.importSystem.GetImportedNamespace(namespace, e))

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

	namespace, err := e.loadModule(templateName, ctx, node.WithContext)
	if err != nil {
		return nil, err
	}
	namespaceMap := e.importSystem.GetNamespaceMap(namespace, e)

	// Import specific items from the namespace
	for _, name := range node.Names {
		// Determine the variable name (use alias if provided, otherwise original name)
		varName := name
		if alias, hasAlias := node.Aliases[name]; hasAlias {
			varName = alias
		}

		if strings.HasPrefix(name, "_") {
			return nil, fmt.Errorf("names starting with an underline can not be imported: %q", name)
		}

		// Get the item from the namespace
		if item, exists := namespaceMap[name]; exists {
			ctx.SetVariable(varName, item)
		} else {
			// As in Jinja2, a missing name is undefined; calling it is an error.
			ctx.SetVariable(varName, &Undefined{
				Name:     name,
				Behavior: e.undefinedHandler.GetUndefinedBehavior(),
				Hint:     fmt.Sprintf("the template %q does not export the requested name %q", templateName, name),
			})
		}
	}

	return "", nil // From statements don't produce output
}

// EvalWithNode evaluates with statements ({% with var=expr %}...{% endwith %})
func (e *DefaultEvaluator) EvalWithNode(node *parser.WithNode, ctx Context) (interface{}, error) {
	// Create a new context scope for the with block
	withCtx := ctx.Clone()

	names, exprs := node.Names, node.Exprs
	if len(names) != len(exprs) || (len(names) == 0 && len(node.Assignments) > 0) {
		// an AST built without the parser: fall back to sorted keys
		names = make([]string, 0, len(node.Assignments))
		for n := range node.Assignments {
			names = append(names, n)
		}
		sort.Strings(names)
		exprs = make([]parser.ExpressionNode, len(names))
		for i, n := range names {
			exprs[i] = node.Assignments[n]
		}
	}

	// Jinja2 evaluates every value in the scope outside the with block, in
	// order (duplicates too), and only then binds the names; the last one wins.
	values := make([]interface{}, len(names))
	for i, varName := range names {
		value, err := e.EvalNode(exprs[i], ctx)
		if err != nil {
			return nil, fmt.Errorf("error evaluating with assignment %s: %w", varName, err)
		}
		values[i] = e.bindValue(exprs[i], value, ctx)
	}
	for i, varName := range names {
		withCtx.SetVariable(varName, values[i])
	}

	// Evaluate the body with the new context
	result, err := e.evalNodeList(node.Body, withCtx)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// loadModule returns the namespace of an imported template. An import without
// context reuses the module already built for that template in this render, as
// Jinja2 does; an import with context is built fresh every time.
func (e *DefaultEvaluator) loadModule(templateName string, ctx Context, withContext bool) (*TemplateNamespace, error) {
	if e.importSystem == nil {
		return nil, fmt.Errorf("cannot import %q: the evaluator has no import system (it was not created by an Environment render)", templateName)
	}
	if !withContext {
		if ns, ok := e.importMemo[templateName]; ok {
			return ns, nil
		}
	}
	leave, err := e.enter("import " + strconv.Quote(templateName))
	if err != nil {
		return nil, err
	}
	defer leave()
	ns, err := e.importSystem.LoadTemplateNamespace(templateName, importContext(ctx, withContext), e)
	if err != nil {
		return nil, err
	}
	if !withContext {
		if e.importMemo == nil {
			e.importMemo = make(map[string]*TemplateNamespace)
		}
		e.importMemo[templateName] = ns
	}
	return ns, nil
}

// ModuleContexter is implemented by contexts that can create an empty context
// sharing their environment (globals, filters, tests). Imports without context
// use it so the imported template sees no render variables.
type ModuleContexter interface {
	NewModuleContext() Context
}

// stripAutoescapeBlocks drops the importer's {% autoescape %} block wrappers: the imported
// template escapes by its own setting, also when it is imported with context.
func stripAutoescapeBlocks(ctx Context) Context {
	switch c := ctx.(type) {
	case *autoescapeContext:
		return stripAutoescapeBlocks(c.Context)
	case *ContextWrapper:
		return &ContextWrapper{Context: stripAutoescapeBlocks(c.Context), escapeContext: c.escapeContext, autoEscaper: c.autoEscaper}
	}
	return ctx
}

// importContext returns the context an imported template runs in: a copy of
// ctx for "with context", otherwise an empty context with ctx's environment.
func importContext(ctx Context, withContext bool) Context {
	if withContext {
		return stripAutoescapeBlocks(ctx).Clone()
	}
	switch c := ctx.(type) {
	case ModuleContexter:
		return c.NewModuleContext()
	case *autoescapeContext:
		// The importer's {% autoescape %} block does not apply to the imported template.
		return importContext(c.Context, false)
	case *ContextWrapper:
		return &ContextWrapper{Context: importContext(c.Context, false), escapeContext: c.escapeContext, autoEscaper: c.autoEscaper}
	}
	return &simpleContext{variables: make(map[string]interface{})}
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

	// The body is template output: Markup where escaping is on
	body := bodyContent.String()
	var current interface{} = markIfSafe(body, ctx)

	// Apply each filter in the chain to the previous result as it is
	for _, filter := range node.FilterChain {
		contentLiteral := parser.NewLiteralNode(current, ToString(current), filter.Line(), filter.Column())

		// Create a new filter node with the current content as the expression
		filterNode := parser.NewFilterNode(contentLiteral, filter.FilterName, filter.Arguments, filter.Line(), filter.Column())
		filterNode.NamedArgs = filter.NamedArgs

		// Evaluate the filter
		result, err := e.EvalFilterNode(filterNode, ctx)
		if err != nil {
			return nil, fmt.Errorf("error applying filter '%s' in filter block: %w", filter.FilterName, err)
		}
		current = result
	}

	// Markup is written as is; anything else is escaped like {{ }} would
	if m, ok := AsMarkup(current); ok {
		return string(m), nil
	}
	return escapeFor(ctx, current), nil
}
