package parser

import (
	"fmt"
	"strings"
)

type Node interface {
	String() string
	Line() int
	Column() int
}

type baseNode struct {
	line   int
	column int
}

func (n *baseNode) Line() int {
	return n.line
}

func (n *baseNode) Column() int {
	return n.column
}

// BaseNode is the exported version of baseNode for extensions
type BaseNode struct {
	Line   int
	Column int
}

// TemplateNode represents the root of a template AST
type TemplateNode struct {
	baseNode
	Name     string
	Children []Node
}

func NewTemplateNode(name string, line, column int) *TemplateNode {
	return &TemplateNode{
		baseNode: baseNode{line: line, column: column},
		Name:     name,
		Children: make([]Node, 0),
	}
}

func (n *TemplateNode) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Template(%s)", n.Name))
	for _, child := range n.Children {
		sb.WriteString("\n  ")
		sb.WriteString(strings.ReplaceAll(child.String(), "\n", "\n  "))
	}
	return sb.String()
}

// TextNode represents plain text content
type TextNode struct {
	baseNode
	Content string
}

func NewTextNode(content string, line, column int) *TextNode {
	return &TextNode{
		baseNode: baseNode{line: line, column: column},
		Content:  content,
	}
}

func (n *TextNode) String() string {
	return fmt.Sprintf("Text(%q)", n.Content)
}

// VariableNode represents variable interpolation {{ var }}
type VariableNode struct {
	baseNode
	Expression Node
}

func NewVariableNode(expr Node, line, column int) *VariableNode {
	return &VariableNode{
		baseNode:   baseNode{line: line, column: column},
		Expression: expr,
	}
}

func (n *VariableNode) String() string {
	return fmt.Sprintf("Variable(%s)", n.Expression.String())
}

// ExpressionNode represents various expressions
type ExpressionNode interface {
	Node
	ExpressionNode()
}

// IdentifierNode represents variable names and identifiers
type IdentifierNode struct {
	baseNode
	Name string
}

func NewIdentifierNode(name string, line, column int) *IdentifierNode {
	return &IdentifierNode{
		baseNode: baseNode{line: line, column: column},
		Name:     name,
	}
}

func (n *IdentifierNode) String() string {
	return fmt.Sprintf("Id(%s)", n.Name)
}

func (n *IdentifierNode) ExpressionNode() {}

// LiteralNode represents literal values (strings, numbers, booleans)
type LiteralNode struct {
	baseNode
	Value interface{}
	Raw   string
}

func NewLiteralNode(value interface{}, raw string, line, column int) *LiteralNode {
	return &LiteralNode{
		baseNode: baseNode{line: line, column: column},
		Value:    value,
		Raw:      raw,
	}
}

func (n *LiteralNode) String() string {
	return fmt.Sprintf("Literal(%v)", n.Value)
}

func (n *LiteralNode) ExpressionNode() {}

// ListNode represents list literals [1, 2, 3]
type ListNode struct {
	baseNode
	Elements []ExpressionNode
}

func NewListNode(elements []ExpressionNode, line, column int) *ListNode {
	return &ListNode{
		baseNode: baseNode{line: line, column: column},
		Elements: elements,
	}
}

func (n *ListNode) String() string {
	var elements []string
	for _, elem := range n.Elements {
		elements = append(elements, elem.String())
	}
	return fmt.Sprintf("List([%s])", strings.Join(elements, ", "))
}

func (n *ListNode) ExpressionNode() {}

// AttributeNode represents attribute access (obj.attr)
type AttributeNode struct {
	baseNode
	Object    ExpressionNode
	Attribute string
}

func NewAttributeNode(obj ExpressionNode, attr string, line, column int) *AttributeNode {
	return &AttributeNode{
		baseNode:  baseNode{line: line, column: column},
		Object:    obj,
		Attribute: attr,
	}
}

func (n *AttributeNode) String() string {
	return fmt.Sprintf("Attr(%s.%s)", n.Object.String(), n.Attribute)
}

func (n *AttributeNode) ExpressionNode() {}

// GetItemNode represents item access (obj[key])
type GetItemNode struct {
	baseNode
	Object ExpressionNode
	Key    ExpressionNode
}

func NewGetItemNode(obj, key ExpressionNode, line, column int) *GetItemNode {
	return &GetItemNode{
		baseNode: baseNode{line: line, column: column},
		Object:   obj,
		Key:      key,
	}
}

func (n *GetItemNode) String() string {
	return fmt.Sprintf("GetItem(%s[%s])", n.Object.String(), n.Key.String())
}

func (n *GetItemNode) ExpressionNode() {}

// FilterNode represents filter application (value|filter)
type FilterNode struct {
	baseNode
	Expression ExpressionNode
	FilterName string
	Arguments  []ExpressionNode
	NamedArgs  map[string]ExpressionNode
}

func NewFilterNode(expr ExpressionNode, filterName string, args []ExpressionNode, line, column int) *FilterNode {
	return &FilterNode{
		baseNode:   baseNode{line: line, column: column},
		Expression: expr,
		FilterName: filterName,
		Arguments:  args,
		NamedArgs:  make(map[string]ExpressionNode),
	}
}

func (n *FilterNode) String() string {
	if len(n.Arguments) > 0 {
		var args []string
		for _, arg := range n.Arguments {
			args = append(args, arg.String())
		}
		return fmt.Sprintf("Filter(%s|%s(%s))", n.Expression.String(), n.FilterName, strings.Join(args, ", "))
	}
	return fmt.Sprintf("Filter(%s|%s)", n.Expression.String(), n.FilterName)
}

func (n *FilterNode) ExpressionNode() {}

// BinaryOpNode represents binary operations (a + b, a > b, etc.)
type BinaryOpNode struct {
	baseNode
	Left     ExpressionNode
	Operator string
	Right    ExpressionNode
}

func NewBinaryOpNode(left ExpressionNode, op string, right ExpressionNode, line, column int) *BinaryOpNode {
	return &BinaryOpNode{
		baseNode: baseNode{line: line, column: column},
		Left:     left,
		Operator: op,
		Right:    right,
	}
}

func (n *BinaryOpNode) String() string {
	return fmt.Sprintf("BinOp(%s %s %s)", n.Left.String(), n.Operator, n.Right.String())
}

func (n *BinaryOpNode) ExpressionNode() {}

// UnaryOpNode represents unary operations (not x, -x)
type UnaryOpNode struct {
	baseNode
	Operator string
	Operand  ExpressionNode
}

func NewUnaryOpNode(op string, operand ExpressionNode, line, column int) *UnaryOpNode {
	return &UnaryOpNode{
		baseNode: baseNode{line: line, column: column},
		Operator: op,
		Operand:  operand,
	}
}

func (n *UnaryOpNode) String() string {
	return fmt.Sprintf("UnaryOp(%s %s)", n.Operator, n.Operand.String())
}

func (n *UnaryOpNode) ExpressionNode() {}

// StatementNode represents template statements
type StatementNode interface {
	Node
	StatementNode()
}

// IfNode represents if/elif/else statements
type IfNode struct {
	baseNode
	Condition ExpressionNode
	Body      []Node
	ElseIfs   []*IfNode
	Else      []Node
}

func NewIfNode(condition ExpressionNode, line, column int) *IfNode {
	return &IfNode{
		baseNode:  baseNode{line: line, column: column},
		Condition: condition,
		Body:      make([]Node, 0),
		ElseIfs:   make([]*IfNode, 0),
		Else:      make([]Node, 0),
	}
}

func (n *IfNode) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("If(%s)", n.Condition.String()))

	if len(n.Body) > 0 {
		sb.WriteString(" {")
		for _, stmt := range n.Body {
			sb.WriteString("\n  ")
			sb.WriteString(strings.ReplaceAll(stmt.String(), "\n", "\n  "))
		}
		sb.WriteString("\n}")
	}

	for _, elif := range n.ElseIfs {
		sb.WriteString(fmt.Sprintf("\nElif(%s)", elif.Condition.String()))
		if len(elif.Body) > 0 {
			sb.WriteString(" {")
			for _, stmt := range elif.Body {
				sb.WriteString("\n  ")
				sb.WriteString(strings.ReplaceAll(stmt.String(), "\n", "\n  "))
			}
			sb.WriteString("\n}")
		}
	}

	if len(n.Else) > 0 {
		sb.WriteString("\nElse {")
		for _, stmt := range n.Else {
			sb.WriteString("\n  ")
			sb.WriteString(strings.ReplaceAll(stmt.String(), "\n", "\n  "))
		}
		sb.WriteString("\n}")
	}

	return sb.String()
}

func (n *IfNode) StatementNode() {}

// ForNode represents for loops
type ForNode struct {
	baseNode
	Variables []string // Support multiple variables for unpacking
	Iterable  ExpressionNode
	Condition ExpressionNode // Optional conditional for filtered iteration
	Body      []Node
	Else      []Node
	Recursive bool
}

func NewForNode(variables []string, iterable ExpressionNode, line, column int) *ForNode {
	return &ForNode{
		baseNode:  baseNode{line: line, column: column},
		Variables: variables,
		Iterable:  iterable,
		Condition: nil,
		Body:      make([]Node, 0),
		Else:      make([]Node, 0),
	}
}

// NewSingleForNode creates a for node with a single variable (backward compatibility)
func NewSingleForNode(variable string, iterable ExpressionNode, line, column int) *ForNode {
	return NewForNode([]string{variable}, iterable, line, column)
}

func (n *ForNode) String() string {
	var sb strings.Builder
	variables := strings.Join(n.Variables, ", ")
	sb.WriteString(fmt.Sprintf("For(%s in %s)", variables, n.Iterable.String()))

	if len(n.Body) > 0 {
		sb.WriteString(" {")
		for _, stmt := range n.Body {
			sb.WriteString("\n  ")
			sb.WriteString(strings.ReplaceAll(stmt.String(), "\n", "\n  "))
		}
		sb.WriteString("\n}")
	}

	if len(n.Else) > 0 {
		sb.WriteString("\nElse {")
		for _, stmt := range n.Else {
			sb.WriteString("\n  ")
			sb.WriteString(strings.ReplaceAll(stmt.String(), "\n", "\n  "))
		}
		sb.WriteString("\n}")
	}

	return sb.String()
}

func (n *ForNode) StatementNode() {}

// BlockNode represents template blocks
type BlockNode struct {
	baseNode
	Name string
	Body []Node
}

func NewBlockNode(name string, line, column int) *BlockNode {
	return &BlockNode{
		baseNode: baseNode{line: line, column: column},
		Name:     name,
		Body:     make([]Node, 0),
	}
}

func (n *BlockNode) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Block(%s)", n.Name))

	if len(n.Body) > 0 {
		sb.WriteString(" {")
		for _, stmt := range n.Body {
			sb.WriteString("\n  ")
			sb.WriteString(strings.ReplaceAll(stmt.String(), "\n", "\n  "))
		}
		sb.WriteString("\n}")
	}

	return sb.String()
}

func (n *BlockNode) StatementNode() {}

// ExtendsNode represents template inheritance
type ExtendsNode struct {
	baseNode
	Template ExpressionNode
}

func NewExtendsNode(template ExpressionNode, line, column int) *ExtendsNode {
	return &ExtendsNode{
		baseNode: baseNode{line: line, column: column},
		Template: template,
	}
}

func (n *ExtendsNode) String() string {
	return fmt.Sprintf("Extends(%s)", n.Template.String())
}

func (n *ExtendsNode) StatementNode() {}

// IncludeNode represents template inclusion
type IncludeNode struct {
	baseNode
	Template      ExpressionNode
	Context       ExpressionNode // optional
	IgnoreMissing bool
}

func NewIncludeNode(template ExpressionNode, line, column int) *IncludeNode {
	return &IncludeNode{
		baseNode:      baseNode{line: line, column: column},
		Template:      template,
		IgnoreMissing: false,
	}
}

func (n *IncludeNode) String() string {
	if n.Context != nil {
		return fmt.Sprintf("Include(%s with %s)", n.Template.String(), n.Context.String())
	}
	return fmt.Sprintf("Include(%s)", n.Template.String())
}

func (n *IncludeNode) StatementNode() {}

// SuperNode represents super() calls in template inheritance
type SuperNode struct {
	baseNode
}

func NewSuperNode(line, column int) *SuperNode {
	return &SuperNode{
		baseNode: baseNode{line: line, column: column},
	}
}

func (n *SuperNode) String() string {
	return "Super()"
}

func (n *SuperNode) ExpressionNode() {}

// MacroNode represents macro definitions
type MacroNode struct {
	baseNode
	Name       string
	Parameters []string
	Defaults   map[string]ExpressionNode
	Body       []Node
}

func NewMacroNode(name string, line, column int) *MacroNode {
	return &MacroNode{
		baseNode:   baseNode{line: line, column: column},
		Name:       name,
		Parameters: make([]string, 0),
		Defaults:   make(map[string]ExpressionNode),
		Body:       make([]Node, 0),
	}
}

func (n *MacroNode) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Macro(%s(", n.Name))

	for i, param := range n.Parameters {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(param)
		if def, ok := n.Defaults[param]; ok {
			sb.WriteString("=")
			sb.WriteString(def.String())
		}
	}

	sb.WriteString("))")

	if len(n.Body) > 0 {
		sb.WriteString(" {")
		for _, stmt := range n.Body {
			sb.WriteString("\n  ")
			sb.WriteString(strings.ReplaceAll(stmt.String(), "\n", "\n  "))
		}
		sb.WriteString("\n}")
	}

	return sb.String()
}

func (n *MacroNode) StatementNode() {}

// SetNode represents variable assignment (supports both single and multiple assignment)
type SetNode struct {
	baseNode
	Targets []ExpressionNode // Support multiple targets for tuple unpacking (can be identifiers or attributes)
	Value   ExpressionNode   // The value expression to assign
}

func NewSetNode(variable string, value ExpressionNode, line, column int) *SetNode {
	return &SetNode{
		baseNode: baseNode{line: line, column: column},
		Targets:  []ExpressionNode{NewIdentifierNode(variable, line, column)},
		Value:    value,
	}
}

func NewMultiSetNode(variables []string, value ExpressionNode, line, column int) *SetNode {
	targets := make([]ExpressionNode, len(variables))
	for i, v := range variables {
		targets[i] = NewIdentifierNode(v, line, column)
	}
	return &SetNode{
		baseNode: baseNode{line: line, column: column},
		Targets:  targets,
		Value:    value,
	}
}

func NewSetNodeWithTargets(targets []ExpressionNode, value ExpressionNode, line, column int) *SetNode {
	return &SetNode{
		baseNode: baseNode{line: line, column: column},
		Targets:  targets,
		Value:    value,
	}
}

func (n *SetNode) String() string {
	targetStrs := make([]string, len(n.Targets))
	for i, t := range n.Targets {
		targetStrs[i] = t.String()
	}
	if len(n.Targets) == 1 {
		return fmt.Sprintf("Set(%s = %s)", targetStrs[0], n.Value.String())
	}
	return fmt.Sprintf("Set(%s = %s)", strings.Join(targetStrs, ", "), n.Value.String())
}

func (n *SetNode) StatementNode() {}

// BlockSetNode represents block assignment ({% set var %}content{% endset %})
type BlockSetNode struct {
	baseNode
	Variable string
	Body     []Node // The content between {% set var %} and {% endset %}
}

func NewBlockSetNode(variable string, body []Node, line, column int) *BlockSetNode {
	return &BlockSetNode{
		baseNode: baseNode{line: line, column: column},
		Variable: variable,
		Body:     body,
	}
}

func (n *BlockSetNode) String() string {
	return fmt.Sprintf("BlockSet(%s) { ... }", n.Variable)
}

func (n *BlockSetNode) StatementNode() {}

// CallNode represents function/macro calls
type CallNode struct {
	baseNode
	Function  ExpressionNode
	Arguments []ExpressionNode
	Keywords  map[string]ExpressionNode
}

func NewCallNode(function ExpressionNode, line, column int) *CallNode {
	return &CallNode{
		baseNode:  baseNode{line: line, column: column},
		Function:  function,
		Arguments: make([]ExpressionNode, 0),
		Keywords:  make(map[string]ExpressionNode),
	}
}

func (n *CallNode) String() string {
	var args []string

	for _, arg := range n.Arguments {
		args = append(args, arg.String())
	}

	for key, value := range n.Keywords {
		args = append(args, fmt.Sprintf("%s=%s", key, value.String()))
	}

	return fmt.Sprintf("Call(%s(%s))", n.Function.String(), strings.Join(args, ", "))
}

func (n *CallNode) ExpressionNode() {}

// CallBlockNode represents call blocks {% call macro_name() %}content{% endcall %}
type CallBlockNode struct {
	baseNode
	Call ExpressionNode // The function/macro call
	Body []Node         // The content between {% call %} and {% endcall %}
}

func NewCallBlockNode(call ExpressionNode, body []Node, line, column int) *CallBlockNode {
	return &CallBlockNode{
		baseNode: baseNode{line: line, column: column},
		Call:     call,
		Body:     body,
	}
}

func (n *CallBlockNode) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("CallBlock(%s)", n.Call.String()))

	if len(n.Body) > 0 {
		sb.WriteString(" {")
		for _, stmt := range n.Body {
			sb.WriteString("\n  ")
			sb.WriteString(strings.ReplaceAll(stmt.String(), "\n", "\n  "))
		}
		sb.WriteString("\n}")
	}

	return sb.String()
}

func (n *CallBlockNode) StatementNode() {}

// WithNode represents with statements {% with var = expr %}...{% endwith %}
type WithNode struct {
	baseNode
	Assignments map[string]ExpressionNode // Variable assignments like var1=expr1, var2=expr2
	Body        []Node                    // The content between {% with %} and {% endwith %}
}

func NewWithNode(assignments map[string]ExpressionNode, body []Node, line, column int) *WithNode {
	return &WithNode{
		baseNode:    baseNode{line: line, column: column},
		Assignments: assignments,
		Body:        body,
	}
}

func (n *WithNode) String() string {
	var sb strings.Builder
	sb.WriteString("With(")

	var assignments []string
	for key, value := range n.Assignments {
		assignments = append(assignments, fmt.Sprintf("%s=%s", key, value.String()))
	}
	sb.WriteString(strings.Join(assignments, ", "))
	sb.WriteString(")")

	if len(n.Body) > 0 {
		sb.WriteString(" {")
		for _, stmt := range n.Body {
			sb.WriteString("\n  ")
			sb.WriteString(strings.ReplaceAll(stmt.String(), "\n", "\n  "))
		}
		sb.WriteString("\n}")
	}

	return sb.String()
}

func (n *WithNode) StatementNode() {}

// TestNode represents test expressions (variable is defined, value is none)
type TestNode struct {
	baseNode
	Expression ExpressionNode
	TestName   string
	Arguments  []ExpressionNode
	Negated    bool // for "is not" tests
}

func NewTestNode(expr ExpressionNode, testName string, line, column int) *TestNode {
	return &TestNode{
		baseNode:   baseNode{line: line, column: column},
		Expression: expr,
		TestName:   testName,
		Arguments:  make([]ExpressionNode, 0),
		Negated:    false,
	}
}

func (n *TestNode) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Test(%s is ", n.Expression.String()))

	if n.Negated {
		sb.WriteString("not ")
	}

	sb.WriteString(n.TestName)

	if len(n.Arguments) > 0 {
		sb.WriteString("(")
		for i, arg := range n.Arguments {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(arg.String())
		}
		sb.WriteString(")")
	}

	sb.WriteString(")")
	return sb.String()
}

func (n *TestNode) ExpressionNode() {}

// ConditionalNode represents ternary conditional expressions (condition ? true_expr : false_expr)
type ConditionalNode struct {
	baseNode
	Condition ExpressionNode
	TrueExpr  ExpressionNode
	FalseExpr ExpressionNode
}

func NewConditionalNode(condition, trueExpr, falseExpr ExpressionNode, line, column int) *ConditionalNode {
	return &ConditionalNode{
		baseNode:  baseNode{line: line, column: column},
		Condition: condition,
		TrueExpr:  trueExpr,
		FalseExpr: falseExpr,
	}
}

func (n *ConditionalNode) String() string {
	return fmt.Sprintf("Conditional(%s ? %s : %s)", n.Condition.String(), n.TrueExpr.String(), n.FalseExpr.String())
}

func (n *ConditionalNode) ExpressionNode() {}

// AssignmentNode represents assignment expressions (set a = b)
type AssignmentNode struct {
	baseNode
	Target ExpressionNode // Can be identifier or attribute/item access
	Value  ExpressionNode
}

func NewAssignmentNode(target, value ExpressionNode, line, column int) *AssignmentNode {
	return &AssignmentNode{
		baseNode: baseNode{line: line, column: column},
		Target:   target,
		Value:    value,
	}
}

func (n *AssignmentNode) String() string {
	return fmt.Sprintf("Assignment(%s = %s)", n.Target.String(), n.Value.String())
}

func (n *AssignmentNode) ExpressionNode() {}

// SliceNode represents slice expressions (array[start:end:step])
type SliceNode struct {
	baseNode
	Object ExpressionNode
	Start  ExpressionNode // optional
	End    ExpressionNode // optional
	Step   ExpressionNode // optional
}

func NewSliceNode(object ExpressionNode, line, column int) *SliceNode {
	return &SliceNode{
		baseNode: baseNode{line: line, column: column},
		Object:   object,
	}
}

func (n *SliceNode) String() string {
	var parts []string
	parts = append(parts, n.Object.String())
	parts = append(parts, "[")

	if n.Start != nil {
		parts = append(parts, n.Start.String())
	}
	parts = append(parts, ":")

	if n.End != nil {
		parts = append(parts, n.End.String())
	}

	if n.Step != nil {
		parts = append(parts, ":")
		parts = append(parts, n.Step.String())
	}

	parts = append(parts, "]")
	return fmt.Sprintf("Slice(%s)", strings.Join(parts, ""))
}

func (n *SliceNode) ExpressionNode() {}

// ComprehensionNode represents list/dict comprehensions
type ComprehensionNode struct {
	baseNode
	Expression ExpressionNode
	Variable   string
	Iterable   ExpressionNode
	Condition  ExpressionNode // optional filter condition
	IsDict     bool           // true for dict comprehensions
	KeyExpr    ExpressionNode // for dict comprehensions
}

func NewComprehensionNode(expr ExpressionNode, variable string, iterable ExpressionNode, line, column int) *ComprehensionNode {
	return &ComprehensionNode{
		baseNode:   baseNode{line: line, column: column},
		Expression: expr,
		Variable:   variable,
		Iterable:   iterable,
		IsDict:     false,
	}
}

func (n *ComprehensionNode) String() string {
	if n.IsDict {
		result := fmt.Sprintf("DictComp({%s: %s for %s in %s", n.KeyExpr.String(), n.Expression.String(), n.Variable, n.Iterable.String())
		if n.Condition != nil {
			result += fmt.Sprintf(" if %s", n.Condition.String())
		}
		return result + "})"
	}

	result := fmt.Sprintf("ListComp([%s for %s in %s", n.Expression.String(), n.Variable, n.Iterable.String())
	if n.Condition != nil {
		result += fmt.Sprintf(" if %s", n.Condition.String())
	}
	return result + "])"
}

func (n *ComprehensionNode) ExpressionNode() {}

// CommentNode represents template comments {# comment #}
type CommentNode struct {
	baseNode
	Content string
}

func NewCommentNode(content string, line, column int) *CommentNode {
	return &CommentNode{
		baseNode: baseNode{line: line, column: column},
		Content:  content,
	}
}

func (n *CommentNode) String() string {
	return fmt.Sprintf("Comment(%q)", n.Content)
}

// RawNode represents raw blocks {% raw %}...{% endraw %}
type RawNode struct {
	baseNode
	Content string
}

func NewRawNode(content string, line, column int) *RawNode {
	return &RawNode{
		baseNode: baseNode{line: line, column: column},
		Content:  content,
	}
}

func (n *RawNode) String() string {
	return fmt.Sprintf("Raw(%q)", n.Content)
}

// AutoescapeNode represents autoescape blocks {% autoescape true %}...{% endautoescape %}
type AutoescapeNode struct {
	baseNode
	Enabled bool // true for autoescape on, false for off
	Body    []Node
}

func NewAutoescapeNode(enabled bool, line, column int) *AutoescapeNode {
	return &AutoescapeNode{
		baseNode: baseNode{line: line, column: column},
		Enabled:  enabled,
		Body:     make([]Node, 0),
	}
}

func (n *AutoescapeNode) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Autoescape(%v)", n.Enabled))

	if len(n.Body) > 0 {
		sb.WriteString(" {")
		for _, stmt := range n.Body {
			sb.WriteString("\n  ")
			sb.WriteString(strings.ReplaceAll(stmt.String(), "\n", "\n  "))
		}
		sb.WriteString("\n}")
	}

	return sb.String()
}

func (n *AutoescapeNode) StatementNode() {}

// FilterBlockNode represents filter blocks {% filter upper %}...{% endfilter %}
type FilterBlockNode struct {
	baseNode
	FilterChain []FilterNode // Chain of filters to apply
	Body        []Node       // Template content to filter
}

func NewFilterBlockNode(filterChain []FilterNode, line, column int) *FilterBlockNode {
	return &FilterBlockNode{
		baseNode:    baseNode{line: line, column: column},
		FilterChain: filterChain,
		Body:        make([]Node, 0),
	}
}

func (n *FilterBlockNode) String() string {
	var sb strings.Builder
	sb.WriteString("FilterBlock(")

	// Show filter chain
	if len(n.FilterChain) > 0 {
		var filters []string
		for _, filter := range n.FilterChain {
			filters = append(filters, filter.FilterName)
		}
		sb.WriteString(strings.Join(filters, "|"))
	}

	// Show body
	if len(n.Body) > 0 {
		sb.WriteString(" {")
		for _, stmt := range n.Body {
			sb.WriteString("\n  ")
			sb.WriteString(strings.ReplaceAll(stmt.String(), "\n", "\n  "))
		}
		sb.WriteString("\n}")
	}

	sb.WriteString(")")
	return sb.String()
}

func (n *FilterBlockNode) StatementNode() {}

// BreakNode represents break statements in loops
type BreakNode struct {
	baseNode
}

func NewBreakNode(line, column int) *BreakNode {
	return &BreakNode{
		baseNode: baseNode{line: line, column: column},
	}
}

func (n *BreakNode) String() string {
	return "Break()"
}

func (n *BreakNode) StatementNode() {}

// ContinueNode represents continue statements in loops
type ContinueNode struct {
	baseNode
}

func NewContinueNode(line, column int) *ContinueNode {
	return &ContinueNode{
		baseNode: baseNode{line: line, column: column},
	}
}

func (n *ContinueNode) String() string {
	return "Continue()"
}

func (n *ContinueNode) StatementNode() {}

// ExtensionNode represents a custom tag node in the AST
type ExtensionNode struct {
	baseNode
	ExtensionName string
	TagName       string
	Arguments     []ExpressionNode
	Body          []Node
	Properties    map[string]interface{}
	EvaluateFunc  func(node *ExtensionNode, ctx interface{}) (interface{}, error)
}

func NewExtensionNode(extensionName, tagName string, line, column int) *ExtensionNode {
	return &ExtensionNode{
		baseNode:      baseNode{line: line, column: column},
		ExtensionName: extensionName,
		TagName:       tagName,
		Arguments:     make([]ExpressionNode, 0),
		Body:          make([]Node, 0),
		Properties:    make(map[string]interface{}),
	}
}

func (n *ExtensionNode) String() string {
	return fmt.Sprintf("Extension(%s:%s)", n.ExtensionName, n.TagName)
}

func (n *ExtensionNode) StatementNode() {}

// AddArgument adds an argument to the extension node
func (n *ExtensionNode) AddArgument(arg ExpressionNode) {
	n.Arguments = append(n.Arguments, arg)
}

// AddBodyNode adds a node to the body
func (n *ExtensionNode) AddBodyNode(node Node) {
	n.Body = append(n.Body, node)
}

// SetProperty sets a custom property
func (n *ExtensionNode) SetProperty(key string, value interface{}) {
	n.Properties[key] = value
}

// GetProperty gets a custom property
func (n *ExtensionNode) GetProperty(key string) (interface{}, bool) {
	value, ok := n.Properties[key]
	return value, ok
}

// SetEvaluateFunc sets the evaluation function
func (n *ExtensionNode) SetEvaluateFunc(fn func(*ExtensionNode, interface{}) (interface{}, error)) {
	n.EvaluateFunc = fn
}

// Evaluate evaluates the extension node with the given context
func (n *ExtensionNode) Evaluate(ctx interface{}) (interface{}, error) {
	if n.EvaluateFunc == nil {
		return nil, fmt.Errorf("no evaluate function set for extension %s", n.ExtensionName)
	}
	return n.EvaluateFunc(n, ctx)
}

// ImportNode represents import statements ({% import 'template.html' as name %})
type ImportNode struct {
	baseNode
	Template ExpressionNode // The template to import
	Alias    string         // The alias name for the imported template
}

func NewImportNode(line, column int, template ExpressionNode, alias string) *ImportNode {
	return &ImportNode{
		baseNode: baseNode{line: line, column: column},
		Template: template,
		Alias:    alias,
	}
}

func (n *ImportNode) String() string {
	return fmt.Sprintf("Import(%s as %s)", n.Template.String(), n.Alias)
}

func (n *ImportNode) StatementNode() {}

// FromNode represents from-import statements ({% from 'template.html' import item1, item2 %})
type FromNode struct {
	baseNode
	Template ExpressionNode    // The template to import from
	Names    []string          // The names to import
	Aliases  map[string]string // Optional aliases for imported names (name -> alias)
}

func NewFromNode(line, column int, template ExpressionNode, names []string, aliases map[string]string) *FromNode {
	if aliases == nil {
		aliases = make(map[string]string)
	}
	return &FromNode{
		baseNode: baseNode{line: line, column: column},
		Template: template,
		Names:    names,
		Aliases:  aliases,
	}
}

func (n *FromNode) String() string {
	names := make([]string, len(n.Names))
	for i, name := range n.Names {
		if alias, ok := n.Aliases[name]; ok {
			names[i] = fmt.Sprintf("%s as %s", name, alias)
		} else {
			names[i] = name
		}
	}
	return fmt.Sprintf("From(%s import %s)", n.Template.String(), strings.Join(names, ", "))
}

func (n *FromNode) StatementNode() {}

// DoNode represents a do statement that executes an expression for side effects only
type DoNode struct {
	baseNode
	Expression ExpressionNode
}

func NewDoNode(expr ExpressionNode, line, column int) *DoNode {
	return &DoNode{
		baseNode:   baseNode{line: line, column: column},
		Expression: expr,
	}
}

func (n *DoNode) String() string {
	return fmt.Sprintf("Do(%s)", n.Expression.String())
}

func (n *DoNode) StatementNode() {}
