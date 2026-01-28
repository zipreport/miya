package runtime

import (
	"fmt"
	"strings"

	"github.com/zipreport/miya/parser"
)

// RuntimeError represents an enhanced error that occurred during template evaluation
type RuntimeError struct {
	Type         string
	Message      string
	TemplateName string
	Line         int
	Column       int
	Source       string
	Context      string
	Suggestion   string
	Node         parser.Node // AST node where error occurred
}

// Error implements the error interface
func (re *RuntimeError) Error() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%s: %s", re.Type, re.Message))

	if re.TemplateName != "" {
		sb.WriteString(fmt.Sprintf(" in template '%s'", re.TemplateName))
	}

	if re.Line > 0 {
		sb.WriteString(fmt.Sprintf(" at line %d", re.Line))
		if re.Column > 0 {
			sb.WriteString(fmt.Sprintf(", column %d", re.Column))
		}
	}

	return sb.String()
}

// DetailedError returns a detailed error message with context
func (re *RuntimeError) DetailedError() string {
	var sb strings.Builder

	// Header
	sb.WriteString(fmt.Sprintf("Template Runtime Error: %s\n", re.Message))
	sb.WriteString(strings.Repeat("=", 50))
	sb.WriteString("\n\n")

	// Location
	if re.TemplateName != "" {
		sb.WriteString(fmt.Sprintf("Template: %s\n", re.TemplateName))
	}
	if re.Line > 0 {
		sb.WriteString(fmt.Sprintf("Location: Line %d", re.Line))
		if re.Column > 0 {
			sb.WriteString(fmt.Sprintf(", Column %d", re.Column))
		}
		sb.WriteString("\n")
	}

	// Source context
	if re.Source != "" && re.Line > 0 {
		sb.WriteString("\nSource context:\n")
		sb.WriteString(re.getSourceContext())
		sb.WriteString("\n")
	}

	// Context information
	if re.Context != "" {
		sb.WriteString(fmt.Sprintf("\nContext: %s\n", re.Context))
	}

	// Suggestion
	if re.Suggestion != "" {
		sb.WriteString(fmt.Sprintf("\nSuggestion: %s\n", re.Suggestion))
	}

	return sb.String()
}

// getSourceContext returns the source code context around the error
func (re *RuntimeError) getSourceContext() string {
	if re.Source == "" || re.Line <= 0 {
		return ""
	}

	lines := strings.Split(re.Source, "\n")
	if re.Line > len(lines) {
		return ""
	}

	var sb strings.Builder
	start := max(1, re.Line-2)
	end := min(len(lines), re.Line+2)

	for i := start; i <= end; i++ {
		lineNum := i
		line := lines[i-1]

		if lineNum == re.Line {
			sb.WriteString(fmt.Sprintf("  > %3d | %s\n", lineNum, line))
			if re.Column > 0 {
				// Add pointer to the exact column
				pointer := strings.Repeat(" ", 8+re.Column-1) + "^"
				sb.WriteString(fmt.Sprintf("    %s\n", pointer))
			}
		} else {
			sb.WriteString(fmt.Sprintf("    %3d | %s\n", lineNum, line))
		}
	}

	return sb.String()
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Error type constants
const (
	ErrorTypeUndefined = "UndefinedError"
	ErrorTypeType      = "TypeError"
	ErrorTypeFilter    = "FilterError"
	ErrorTypeTest      = "TestError"
	ErrorTypeRuntime   = "RuntimeError"
	ErrorTypeMath      = "MathError"
	ErrorTypeAccess    = "AccessError"
)

// NewRuntimeError creates a new runtime error with AST node information
func NewRuntimeError(errorType, message string, node parser.Node) *RuntimeError {
	line, column := 0, 0
	if node != nil {
		line = node.Line()
		column = node.Column()
	}

	return &RuntimeError{
		Type:    errorType,
		Message: message,
		Line:    line,
		Column:  column,
		Node:    node,
	}
}

// WithTemplate adds template information to the error
func (re *RuntimeError) WithTemplate(templateName, source string) *RuntimeError {
	re.TemplateName = templateName
	re.Source = source
	return re
}

// WithContext adds context information to the error
func (re *RuntimeError) WithContext(context string) *RuntimeError {
	re.Context = context
	return re
}

// WithSuggestion adds a suggestion to the error
func (re *RuntimeError) WithSuggestion(suggestion string) *RuntimeError {
	re.Suggestion = suggestion
	return re
}

// Common error creation functions
func NewUndefinedVariableError(varName string, node parser.Node) *RuntimeError {
	return NewRuntimeError(ErrorTypeUndefined, fmt.Sprintf("undefined variable: %s", varName), node).
		WithSuggestion(fmt.Sprintf("Check if '%s' is defined in the template context or if it's spelled correctly", varName))
}

func NewFilterError(filterName string, err error, node parser.Node) *RuntimeError {
	message := fmt.Sprintf("error applying filter '%s': %v", filterName, err)
	return NewRuntimeError(ErrorTypeFilter, message, node).
		WithSuggestion(fmt.Sprintf("Check if filter '%s' is registered and arguments are correct", filterName))
}

func NewTestError(testName string, err error, node parser.Node) *RuntimeError {
	message := fmt.Sprintf("error applying test '%s': %v", testName, err)
	return NewRuntimeError(ErrorTypeTest, message, node).
		WithSuggestion(fmt.Sprintf("Check if test '%s' is registered and arguments are correct", testName))
}

func NewTypeError(operation string, value interface{}, node parser.Node) *RuntimeError {
	message := fmt.Sprintf("cannot perform %s on value of type %T", operation, value)
	return NewRuntimeError(ErrorTypeType, message, node).
		WithSuggestion(fmt.Sprintf("Ensure the value supports the %s operation", operation))
}

func NewMathError(operation string, err error, node parser.Node) *RuntimeError {
	message := fmt.Sprintf("math error in %s: %v", operation, err)
	return NewRuntimeError(ErrorTypeMath, message, node).
		WithSuggestion("Check for division by zero or invalid numeric operations")
}

func NewAccessError(attribute string, obj interface{}, node parser.Node) *RuntimeError {
	message := fmt.Sprintf("cannot access attribute '%s' on value of type %T", attribute, obj)
	return NewRuntimeError(ErrorTypeAccess, message, node).
		WithSuggestion(fmt.Sprintf("Check if attribute '%s' exists or if the object is nil", attribute))
}
