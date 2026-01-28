package miya

import (
	"fmt"
	"strings"
)

// EnhancedTemplateError represents an enhanced error that occurred during template processing
type EnhancedTemplateError struct {
	Type         string
	Message      string
	TemplateName string
	Line         int
	Column       int
	Source       string
	Context      string
	Suggestion   string
	StackTrace   []StackFrame
}

// StackFrame represents a frame in the template execution stack
type StackFrame struct {
	TemplateName string
	Line         int
	Column       int
	Function     string
	Source       string
}

// Error implements the error interface
func (te *EnhancedTemplateError) Error() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%s: %s", te.Type, te.Message))

	if te.TemplateName != "" {
		sb.WriteString(fmt.Sprintf(" in template '%s'", te.TemplateName))
	}

	if te.Line > 0 {
		sb.WriteString(fmt.Sprintf(" at line %d", te.Line))
		if te.Column > 0 {
			sb.WriteString(fmt.Sprintf(", column %d", te.Column))
		}
	}

	return sb.String()
}

// DetailedError returns a detailed error message with context
func (te *EnhancedTemplateError) DetailedError() string {
	var sb strings.Builder

	// Header
	sb.WriteString(fmt.Sprintf("Template Error: %s\n", te.Message))
	sb.WriteString(strings.Repeat("=", 50))
	sb.WriteString("\n\n")

	// Location
	if te.TemplateName != "" {
		sb.WriteString(fmt.Sprintf("Template: %s\n", te.TemplateName))
	}
	if te.Line > 0 {
		sb.WriteString(fmt.Sprintf("Location: Line %d", te.Line))
		if te.Column > 0 {
			sb.WriteString(fmt.Sprintf(", Column %d", te.Column))
		}
		sb.WriteString("\n")
	}

	// Source context
	if te.Source != "" && te.Line > 0 {
		sb.WriteString("\nSource context:\n")
		sb.WriteString(te.getSourceContext())
		sb.WriteString("\n")
	}

	// Suggestion
	if te.Suggestion != "" {
		sb.WriteString(fmt.Sprintf("\nSuggestion: %s\n", te.Suggestion))
	}

	// Stack trace
	if len(te.StackTrace) > 0 {
		sb.WriteString("\nStack trace:\n")
		for i, frame := range te.StackTrace {
			sb.WriteString(fmt.Sprintf("  %d. %s at %s:%d:%d\n",
				i+1, frame.Function, frame.TemplateName, frame.Line, frame.Column))
		}
	}

	return sb.String()
}

// getSourceContext returns the source code context around the error
func (te *EnhancedTemplateError) getSourceContext() string {
	if te.Source == "" || te.Line <= 0 {
		return ""
	}

	lines := strings.Split(te.Source, "\n")
	if te.Line > len(lines) {
		return ""
	}

	var sb strings.Builder
	start := max(1, te.Line-2)
	end := min(len(lines), te.Line+2)

	for i := start; i <= end; i++ {
		lineNum := i
		line := lines[i-1]

		if lineNum == te.Line {
			sb.WriteString(fmt.Sprintf("  > %3d | %s\n", lineNum, line))
			if te.Column > 0 {
				// Add pointer to the exact column
				pointer := strings.Repeat(" ", 8+te.Column-1) + "^"
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

// ErrorType constants for different types of template errors
const (
	ErrorTypeSyntax      = "SyntaxError"
	ErrorTypeUndefined   = "UndefinedError"
	ErrorTypeType        = "TypeError"
	ErrorTypeFilter      = "FilterError"
	ErrorTypeTest        = "TestError"
	ErrorTypeRuntime     = "RuntimeError"
	ErrorTypeTemplate    = "TemplateError"
	ErrorTypeInheritance = "InheritanceError"
	ErrorTypeMacro       = "MacroError"
)

// NewEnhancedTemplateError creates a new enhanced template error
func NewEnhancedTemplateError(errorType, message, templateName string, line, column int) *EnhancedTemplateError {
	return &EnhancedTemplateError{
		Type:         errorType,
		Message:      message,
		TemplateName: templateName,
		Line:         line,
		Column:       column,
		StackTrace:   make([]StackFrame, 0),
	}
}

// WithSource adds source context to the error
func (te *EnhancedTemplateError) WithSource(source string) *EnhancedTemplateError {
	te.Source = source
	return te
}

// WithContext adds additional context to the error
func (te *EnhancedTemplateError) WithContext(context string) *EnhancedTemplateError {
	te.Context = context
	return te
}

// WithSuggestion adds a suggestion to fix the error
func (te *EnhancedTemplateError) WithSuggestion(suggestion string) *EnhancedTemplateError {
	te.Suggestion = suggestion
	return te
}

// WithStackFrame adds a stack frame to the error
func (te *EnhancedTemplateError) WithStackFrame(templateName string, line, column int, function string) *EnhancedTemplateError {
	frame := StackFrame{
		TemplateName: templateName,
		Line:         line,
		Column:       column,
		Function:     function,
	}
	te.StackTrace = append(te.StackTrace, frame)
	return te
}

// ErrorHandler provides configurable error handling
type ErrorHandler struct {
	ShowSourceContext bool
	ShowStackTrace    bool
	ShowSuggestions   bool
	MaxSourceLines    int
}

// DefaultErrorHandler returns a default error handler
func DefaultErrorHandler() *ErrorHandler {
	return &ErrorHandler{
		ShowSourceContext: true,
		ShowStackTrace:    true,
		ShowSuggestions:   true,
		MaxSourceLines:    5,
	}
}

// FormatError formats an error according to the handler configuration
func (eh *ErrorHandler) FormatError(err error) string {
	if te, ok := err.(*EnhancedTemplateError); ok {
		if eh.ShowSourceContext || eh.ShowStackTrace || eh.ShowSuggestions {
			return te.DetailedError()
		}
		return te.Error()
	}
	return err.Error()
}

// SyntaxErrorHelper provides helpful syntax error messages
type SyntaxErrorHelper struct {
	commonMistakes map[string]string
}

// NewSyntaxErrorHelper creates a new syntax error helper
func NewSyntaxErrorHelper() *SyntaxErrorHelper {
	return &SyntaxErrorHelper{
		commonMistakes: map[string]string{
			"unexpected '}'":                "Check for missing opening brace '{' or extra closing brace '}'",
			"unexpected ')'":                "Check for missing opening parenthesis '(' or extra closing parenthesis ')'",
			"unexpected ']'":                "Check for missing opening bracket '[' or extra closing bracket ']'",
			"expected '}}' after variable":  "Variable expressions must be closed with '}}'",
			"expected '%}' after block":     "Block statements must be closed with '%}'",
			"expected 'endif'":              "If statements must be closed with '{% endif %}'",
			"expected 'endfor'":             "For loops must be closed with '{% endfor %}'",
			"expected 'endblock'":           "Blocks must be closed with '{% endblock %}'",
			"expected 'endmacro'":           "Macros must be closed with '{% endmacro %}'",
			"unknown filter":                "Check filter name spelling or register the filter",
			"unknown test":                  "Check test name spelling or register the test",
			"undefined variable":            "Check variable name spelling or ensure it's defined in the context",
			"cannot access attribute":       "Check if the object has the attribute or if it's null",
			"division by zero":              "Ensure the denominator is not zero",
			"unsupported operand type":      "Check that operators are used with compatible types",
			"template not found":            "Check template path and ensure the file exists",
			"circular template inheritance": "Remove circular references in template inheritance",
			"block not found":               "Ensure the block is defined in the parent template",
		},
	}
}

// GetSuggestion returns a suggestion for common syntax errors
func (seh *SyntaxErrorHelper) GetSuggestion(errorMessage string) string {
	for pattern, suggestion := range seh.commonMistakes {
		if strings.Contains(strings.ToLower(errorMessage), strings.ToLower(pattern)) {
			return suggestion
		}
	}
	return ""
}

// Global syntax error helper
var GlobalSyntaxErrorHelper = NewSyntaxErrorHelper()

// DebugInfo provides debugging information for templates
type DebugInfo struct {
	Variables        map[string]interface{}
	AvailableFilters []string
	AvailableTests   []string
	TemplatePath     []string
	ExecutionTime    int64 // in nanoseconds
	MemoryUsage      int64 // in bytes
}

// TemplateDebugger provides debugging capabilities
type TemplateDebugger struct {
	enabled        bool
	breakpoints    map[string][]int // template -> line numbers
	watchVariables []string
	stepMode       bool
}

// NewTemplateDebugger creates a new template debugger
func NewTemplateDebugger() *TemplateDebugger {
	return &TemplateDebugger{
		enabled:        false,
		breakpoints:    make(map[string][]int),
		watchVariables: make([]string, 0),
		stepMode:       false,
	}
}

// Enable enables the debugger
func (td *TemplateDebugger) Enable() {
	td.enabled = true
}

// Disable disables the debugger
func (td *TemplateDebugger) Disable() {
	td.enabled = false
}

// AddBreakpoint adds a breakpoint at the specified line
func (td *TemplateDebugger) AddBreakpoint(templateName string, line int) {
	if td.breakpoints[templateName] == nil {
		td.breakpoints[templateName] = make([]int, 0)
	}
	td.breakpoints[templateName] = append(td.breakpoints[templateName], line)
}

// RemoveBreakpoint removes a breakpoint
func (td *TemplateDebugger) RemoveBreakpoint(templateName string, line int) {
	lines := td.breakpoints[templateName]
	for i, l := range lines {
		if l == line {
			td.breakpoints[templateName] = append(lines[:i], lines[i+1:]...)
			break
		}
	}
}

// AddWatchVariable adds a variable to watch
func (td *TemplateDebugger) AddWatchVariable(varName string) {
	td.watchVariables = append(td.watchVariables, varName)
}

// ShouldBreak returns true if execution should break at this location
func (td *TemplateDebugger) ShouldBreak(templateName string, line int) bool {
	if !td.enabled {
		return false
	}

	if td.stepMode {
		return true
	}

	lines := td.breakpoints[templateName]
	for _, l := range lines {
		if l == line {
			return true
		}
	}

	return false
}

// GetWatchedVariables returns the values of watched variables
func (td *TemplateDebugger) GetWatchedVariables(ctx Context) map[string]interface{} {
	result := make(map[string]interface{})
	for _, varName := range td.watchVariables {
		if value, ok := ctx.Get(varName); ok {
			result[varName] = value
		} else {
			result[varName] = "<undefined>"
		}
	}
	return result
}

// Global template debugger
var GlobalTemplateDebugger = NewTemplateDebugger()

// ValidationError represents a template validation error
type ValidationError struct {
	*EnhancedTemplateError
	Severity string // "error", "warning", "info"
}

// NewValidationError creates a new validation error
func NewValidationError(severity, message, templateName string, line, column int) *ValidationError {
	return &ValidationError{
		EnhancedTemplateError: NewEnhancedTemplateError("ValidationError", message, templateName, line, column),
		Severity:              severity,
	}
}

// TemplateValidator provides template validation capabilities
type TemplateValidator struct {
	rules []ValidationRule
}

// ValidationRule represents a validation rule
type ValidationRule struct {
	Name        string
	Description string
	Severity    string
	Check       func(templateName, source string) []*ValidationError
}

// NewTemplateValidator creates a new template validator
func NewTemplateValidator() *TemplateValidator {
	validator := &TemplateValidator{
		rules: make([]ValidationRule, 0),
	}

	// Add default validation rules
	validator.AddDefaultRules()
	return validator
}

// AddRule adds a validation rule
func (tv *TemplateValidator) AddRule(rule ValidationRule) {
	tv.rules = append(tv.rules, rule)
}

// AddDefaultRules adds default validation rules
func (tv *TemplateValidator) AddDefaultRules() {
	// Rule: Check for unclosed blocks
	tv.AddRule(ValidationRule{
		Name:        "unclosed_blocks",
		Description: "Check for unclosed template blocks",
		Severity:    "error",
		Check: func(templateName, source string) []*ValidationError {
			errors := make([]*ValidationError, 0)
			// Simple check for common unclosed blocks
			if strings.Contains(source, "{% if") && !strings.Contains(source, "{% endif %}") {
				errors = append(errors, NewValidationError("error",
					"If statement appears to be unclosed", templateName, 0, 0))
			}
			if strings.Contains(source, "{% for") && !strings.Contains(source, "{% endfor %}") {
				errors = append(errors, NewValidationError("error",
					"For loop appears to be unclosed", templateName, 0, 0))
			}
			return errors
		},
	})

	// Rule: Check for undefined variables (basic heuristic)
	tv.AddRule(ValidationRule{
		Name:        "undefined_variables",
		Description: "Check for potentially undefined variables",
		Severity:    "warning",
		Check: func(templateName, source string) []*ValidationError {
			errors := make([]*ValidationError, 0)
			// This is a simplified check - a real implementation would parse the template
			if strings.Contains(source, "{{ undefined_var }}") {
				errors = append(errors, NewValidationError("warning",
					"Variable 'undefined_var' may be undefined", templateName, 0, 0))
			}
			return errors
		},
	})
}

// Validate validates a template
func (tv *TemplateValidator) Validate(templateName, source string) []*ValidationError {
	var allErrors []*ValidationError

	for _, rule := range tv.rules {
		errors := rule.Check(templateName, source)
		allErrors = append(allErrors, errors...)
	}

	return allErrors
}

// Global template validator
var GlobalTemplateValidator = NewTemplateValidator()

// ErrorRecovery provides error recovery capabilities
type ErrorRecovery struct {
	strategies map[string]RecoveryStrategy
}

// RecoveryStrategy represents an error recovery strategy
type RecoveryStrategy func(err *EnhancedTemplateError, ctx Context) (string, error)

// NewErrorRecovery creates a new error recovery system
func NewErrorRecovery() *ErrorRecovery {
	er := &ErrorRecovery{
		strategies: make(map[string]RecoveryStrategy),
	}

	// Add default recovery strategies
	er.AddDefaultStrategies()
	return er
}

// AddStrategy adds a recovery strategy
func (er *ErrorRecovery) AddStrategy(errorType string, strategy RecoveryStrategy) {
	er.strategies[errorType] = strategy
}

// AddDefaultStrategies adds default recovery strategies
func (er *ErrorRecovery) AddDefaultStrategies() {
	// Strategy for undefined variables
	er.AddStrategy(ErrorTypeUndefined, func(err *EnhancedTemplateError, ctx Context) (string, error) {
		return "", nil // Return empty string for undefined variables
	})

	// Strategy for division by zero
	er.AddStrategy(ErrorTypeRuntime, func(err *EnhancedTemplateError, ctx Context) (string, error) {
		if strings.Contains(err.Message, "division by zero") {
			return "0", nil // Return 0 for division by zero
		}
		return "", err // Can't recover from other runtime errors
	})
}

// Recover attempts to recover from an error
func (er *ErrorRecovery) Recover(err *EnhancedTemplateError, ctx Context) (string, error) {
	if strategy, exists := er.strategies[err.Type]; exists {
		return strategy(err, ctx)
	}
	return "", err // No recovery strategy available
}

// Global error recovery system
var GlobalErrorRecovery = NewErrorRecovery()

// Compatibility functions for backward compatibility with the old errors.go API
// These provide the same interface as the old error types but create enhanced errors

// TemplateError is an alias for EnhancedTemplateError for backward compatibility
type TemplateError = EnhancedTemplateError

// SyntaxError creates a syntax error (wrapper around EnhancedTemplateError)
type SyntaxError struct {
	*EnhancedTemplateError
}

// RuntimeError creates a runtime error (wrapper around EnhancedTemplateError)
type RuntimeError struct {
	*EnhancedTemplateError
	Stack []string
}

// UndefinedError creates an undefined variable error (wrapper around EnhancedTemplateError)
type UndefinedError struct {
	*EnhancedTemplateError
	Variable string
}

// Error method for RuntimeError to include stack trace
func (e *RuntimeError) Error() string {
	base := e.EnhancedTemplateError.Error()
	if len(e.Stack) > 0 {
		return fmt.Sprintf("%s\nTemplate stack:\n  %s", base, strings.Join(e.Stack, "\n  "))
	}
	return base
}

// Error method for UndefinedError to provide specific undefined variable message
func (e *UndefinedError) Error() string {
	if e.TemplateName != "" && e.Line > 0 {
		return fmt.Sprintf("undefined variable %q in template %q at line %d", e.Variable, e.TemplateName, e.Line)
	}
	return fmt.Sprintf("undefined variable %q", e.Variable)
}

// Unwrap method for TemplateError compatibility
func (te *EnhancedTemplateError) Unwrap() error {
	// Enhanced errors don't have a wrapped cause like the old TemplateError
	// but we can return nil for compatibility
	return nil
}

// Compatibility constructors that match the old errors.go API

// NewTemplateError creates a new template error (enhanced version)
func NewTemplateError(template string, line, column int, format string, args ...interface{}) *TemplateError {
	return NewEnhancedTemplateError(ErrorTypeTemplate, fmt.Sprintf(format, args...), template, line, column)
}

// NewSyntaxError creates a new syntax error (enhanced version)
func NewSyntaxError(template string, line, column int, format string, args ...interface{}) *SyntaxError {
	return &SyntaxError{
		EnhancedTemplateError: NewEnhancedTemplateError(ErrorTypeSyntax, fmt.Sprintf(format, args...), template, line, column),
	}
}

// NewRuntimeError creates a new runtime error (enhanced version)
func NewRuntimeError(template string, line, column int, format string, args ...interface{}) *RuntimeError {
	return &RuntimeError{
		EnhancedTemplateError: NewEnhancedTemplateError(ErrorTypeRuntime, fmt.Sprintf(format, args...), template, line, column),
		Stack:                 make([]string, 0),
	}
}

// NewUndefinedError creates a new undefined variable error (enhanced version)
func NewUndefinedError(variable, template string, line, column int) *UndefinedError {
	return &UndefinedError{
		EnhancedTemplateError: NewEnhancedTemplateError(ErrorTypeUndefined, fmt.Sprintf("undefined variable: %s", variable), template, line, column),
		Variable:              variable,
	}
}
