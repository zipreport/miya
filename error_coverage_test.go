package miya

import (
	"strings"
	"testing"
)

// TestEnhancedTemplateError tests the EnhancedTemplateError type
func TestEnhancedTemplateError(t *testing.T) {
	t.Run("NewEnhancedTemplateError", func(t *testing.T) {
		err := NewEnhancedTemplateError(ErrorTypeSyntax, "test message", "template.html", 10, 5)
		if err == nil {
			t.Fatal("NewEnhancedTemplateError returned nil")
		}
		if err.Type != ErrorTypeSyntax {
			t.Errorf("Type = %q, want %q", err.Type, ErrorTypeSyntax)
		}
		if err.Message != "test message" {
			t.Errorf("Message = %q, want %q", err.Message, "test message")
		}
		if err.TemplateName != "template.html" {
			t.Errorf("TemplateName = %q, want %q", err.TemplateName, "template.html")
		}
		if err.Line != 10 {
			t.Errorf("Line = %d, want %d", err.Line, 10)
		}
		if err.Column != 5 {
			t.Errorf("Column = %d, want %d", err.Column, 5)
		}
	})

	t.Run("Error", func(t *testing.T) {
		err := NewEnhancedTemplateError(ErrorTypeSyntax, "unexpected token", "test.html", 5, 10)
		errStr := err.Error()

		if !strings.Contains(errStr, "SyntaxError") {
			t.Error("Error() should contain error type")
		}
		if !strings.Contains(errStr, "unexpected token") {
			t.Error("Error() should contain message")
		}
		if !strings.Contains(errStr, "test.html") {
			t.Error("Error() should contain template name")
		}
		if !strings.Contains(errStr, "line 5") {
			t.Error("Error() should contain line number")
		}
	})

	t.Run("ErrorWithoutLocation", func(t *testing.T) {
		err := NewEnhancedTemplateError(ErrorTypeRuntime, "some error", "", 0, 0)
		errStr := err.Error()

		if !strings.Contains(errStr, "RuntimeError") {
			t.Error("Error() should contain error type")
		}
		if !strings.Contains(errStr, "some error") {
			t.Error("Error() should contain message")
		}
	})

	t.Run("DetailedError", func(t *testing.T) {
		source := "line1\nline2\nline3\nline4\nline5"
		err := NewEnhancedTemplateError(ErrorTypeSyntax, "test error", "test.html", 3, 5).
			WithSource(source).
			WithSuggestion("Try this instead").
			WithStackFrame("parent.html", 10, 1, "include")

		detailed := err.DetailedError()

		if !strings.Contains(detailed, "Template Error") {
			t.Error("DetailedError should contain header")
		}
		if !strings.Contains(detailed, "test.html") {
			t.Error("DetailedError should contain template name")
		}
		if !strings.Contains(detailed, "Line 3") {
			t.Error("DetailedError should contain line number")
		}
		if !strings.Contains(detailed, "Try this instead") {
			t.Error("DetailedError should contain suggestion")
		}
		if !strings.Contains(detailed, "Stack trace") {
			t.Error("DetailedError should contain stack trace")
		}
	})

	t.Run("WithMethods", func(t *testing.T) {
		err := NewEnhancedTemplateError(ErrorTypeSyntax, "test", "t.html", 1, 1)

		err = err.WithSource("source code")
		if err.Source != "source code" {
			t.Error("WithSource should set Source")
		}

		err = err.WithContext("some context")
		if err.Context != "some context" {
			t.Error("WithContext should set Context")
		}

		err = err.WithSuggestion("a suggestion")
		if err.Suggestion != "a suggestion" {
			t.Error("WithSuggestion should set Suggestion")
		}

		err = err.WithStackFrame("parent.html", 5, 1, "render")
		if len(err.StackTrace) != 1 {
			t.Error("WithStackFrame should add stack frame")
		}
	})

	t.Run("Unwrap", func(t *testing.T) {
		err := NewEnhancedTemplateError(ErrorTypeSyntax, "test", "t.html", 1, 1)
		if err.Unwrap() != nil {
			t.Error("Unwrap should return nil")
		}
	})
}

// TestGetSourceContext tests source context extraction
func TestGetSourceContext(t *testing.T) {
	source := "line1\nline2\nline3\nline4\nline5\nline6\nline7"

	t.Run("MiddleLine", func(t *testing.T) {
		err := NewEnhancedTemplateError(ErrorTypeSyntax, "test", "t.html", 4, 3).
			WithSource(source)
		ctx := err.getSourceContext()

		if !strings.Contains(ctx, "line4") {
			t.Error("Should contain target line")
		}
		if !strings.Contains(ctx, ">") {
			t.Error("Should have line pointer")
		}
	})

	t.Run("FirstLine", func(t *testing.T) {
		err := NewEnhancedTemplateError(ErrorTypeSyntax, "test", "t.html", 1, 1).
			WithSource(source)
		ctx := err.getSourceContext()

		if !strings.Contains(ctx, "line1") {
			t.Error("Should contain first line")
		}
	})

	t.Run("LastLine", func(t *testing.T) {
		err := NewEnhancedTemplateError(ErrorTypeSyntax, "test", "t.html", 7, 1).
			WithSource(source)
		ctx := err.getSourceContext()

		if !strings.Contains(ctx, "line7") {
			t.Error("Should contain last line")
		}
	})

	t.Run("NoSource", func(t *testing.T) {
		err := NewEnhancedTemplateError(ErrorTypeSyntax, "test", "t.html", 1, 1)
		ctx := err.getSourceContext()
		if ctx != "" {
			t.Error("Should return empty for no source")
		}
	})

	t.Run("InvalidLine", func(t *testing.T) {
		err := NewEnhancedTemplateError(ErrorTypeSyntax, "test", "t.html", 100, 1).
			WithSource(source)
		ctx := err.getSourceContext()
		if ctx != "" {
			t.Error("Should return empty for invalid line")
		}
	})
}

// TestMinMax tests min and max helper functions
func TestMinMax(t *testing.T) {
	if max(5, 3) != 5 {
		t.Error("max(5, 3) should be 5")
	}
	if max(3, 5) != 5 {
		t.Error("max(3, 5) should be 5")
	}
	if min(5, 3) != 3 {
		t.Error("min(5, 3) should be 3")
	}
	if min(3, 5) != 3 {
		t.Error("min(3, 5) should be 3")
	}
}

// TestErrorHandler tests the ErrorHandler type
func TestErrorHandler(t *testing.T) {
	t.Run("DefaultErrorHandler", func(t *testing.T) {
		handler := DefaultErrorHandler()
		if handler == nil {
			t.Fatal("DefaultErrorHandler returned nil")
		}
		if !handler.ShowSourceContext {
			t.Error("ShowSourceContext should be true by default")
		}
		if !handler.ShowStackTrace {
			t.Error("ShowStackTrace should be true by default")
		}
		if !handler.ShowSuggestions {
			t.Error("ShowSuggestions should be true by default")
		}
	})

	t.Run("FormatEnhancedError", func(t *testing.T) {
		handler := DefaultErrorHandler()
		err := NewEnhancedTemplateError(ErrorTypeSyntax, "test", "t.html", 1, 1).
			WithSource("code").
			WithSuggestion("suggestion")

		formatted := handler.FormatError(err)
		if !strings.Contains(formatted, "Template Error") {
			t.Error("Should format with details")
		}
	})

	t.Run("FormatSimpleError", func(t *testing.T) {
		handler := &ErrorHandler{
			ShowSourceContext: false,
			ShowStackTrace:    false,
			ShowSuggestions:   false,
		}
		err := NewEnhancedTemplateError(ErrorTypeSyntax, "test", "t.html", 1, 1)

		formatted := handler.FormatError(err)
		if strings.Contains(formatted, "Template Error:") {
			t.Error("Should format without details when disabled")
		}
	})

	t.Run("FormatNonTemplateError", func(t *testing.T) {
		handler := DefaultErrorHandler()
		formatted := handler.FormatError(&simpleError{"simple error"})
		if formatted != "simple error" {
			t.Errorf("Expected 'simple error', got %q", formatted)
		}
	})
}

type simpleError struct {
	msg string
}

func (e *simpleError) Error() string {
	return e.msg
}

// TestSyntaxErrorHelper tests the SyntaxErrorHelper type
func TestSyntaxErrorHelper(t *testing.T) {
	helper := NewSyntaxErrorHelper()

	t.Run("KnownPatterns", func(t *testing.T) {
		patterns := []string{
			"unexpected '}'",
			"expected 'endif'",
			"unknown filter",
			"undefined variable",
			"template not found",
		}

		for _, pattern := range patterns {
			suggestion := helper.GetSuggestion(pattern)
			if suggestion == "" {
				t.Errorf("Expected suggestion for %q", pattern)
			}
		}
	})

	t.Run("UnknownPattern", func(t *testing.T) {
		suggestion := helper.GetSuggestion("completely random error")
		if suggestion != "" {
			t.Error("Should return empty for unknown pattern")
		}
	})

	t.Run("CaseInsensitive", func(t *testing.T) {
		suggestion := helper.GetSuggestion("UNEXPECTED '}'")
		if suggestion == "" {
			t.Error("Should match case-insensitively")
		}
	})
}

// TestGlobalSyntaxErrorHelper tests the global syntax error helper
func TestGlobalSyntaxErrorHelper(t *testing.T) {
	if GlobalSyntaxErrorHelper == nil {
		t.Error("GlobalSyntaxErrorHelper should not be nil")
	}
}

// TestTemplateDebugger tests the TemplateDebugger type
func TestTemplateDebugger(t *testing.T) {
	t.Run("NewTemplateDebugger", func(t *testing.T) {
		debugger := NewTemplateDebugger()
		if debugger == nil {
			t.Fatal("NewTemplateDebugger returned nil")
		}
		if debugger.enabled {
			t.Error("Debugger should be disabled by default")
		}
	})

	t.Run("EnableDisable", func(t *testing.T) {
		debugger := NewTemplateDebugger()
		debugger.Enable()
		if !debugger.enabled {
			t.Error("Enable should enable the debugger")
		}
		debugger.Disable()
		if debugger.enabled {
			t.Error("Disable should disable the debugger")
		}
	})

	t.Run("Breakpoints", func(t *testing.T) {
		debugger := NewTemplateDebugger()
		debugger.Enable()

		debugger.AddBreakpoint("test.html", 10)
		debugger.AddBreakpoint("test.html", 20)

		if !debugger.ShouldBreak("test.html", 10) {
			t.Error("Should break at breakpoint")
		}
		if debugger.ShouldBreak("test.html", 15) {
			t.Error("Should not break at non-breakpoint")
		}

		debugger.RemoveBreakpoint("test.html", 10)
		if debugger.ShouldBreak("test.html", 10) {
			t.Error("Should not break after removing breakpoint")
		}
	})

	t.Run("WatchVariables", func(t *testing.T) {
		debugger := NewTemplateDebugger()
		debugger.AddWatchVariable("foo")
		debugger.AddWatchVariable("bar")

		ctx := NewContext()
		ctx.Set("foo", "fooValue")

		watched := debugger.GetWatchedVariables(ctx)
		if watched["foo"] != "fooValue" {
			t.Error("Should include defined variable")
		}
		if watched["bar"] != "<undefined>" {
			t.Error("Should mark undefined variable")
		}
	})

	t.Run("StepMode", func(t *testing.T) {
		debugger := NewTemplateDebugger()
		debugger.Enable()
		debugger.stepMode = true

		if !debugger.ShouldBreak("any.html", 1) {
			t.Error("Should break on any line in step mode")
		}
	})

	t.Run("DisabledDebugger", func(t *testing.T) {
		debugger := NewTemplateDebugger()
		debugger.AddBreakpoint("test.html", 10)

		if debugger.ShouldBreak("test.html", 10) {
			t.Error("Should not break when disabled")
		}
	})
}

// TestGlobalTemplateDebugger tests the global template debugger
func TestGlobalTemplateDebugger(t *testing.T) {
	if GlobalTemplateDebugger == nil {
		t.Error("GlobalTemplateDebugger should not be nil")
	}
}

// TestTemplateValidator tests the TemplateValidator type
func TestTemplateValidator(t *testing.T) {
	t.Run("NewTemplateValidator", func(t *testing.T) {
		validator := NewTemplateValidator()
		if validator == nil {
			t.Fatal("NewTemplateValidator returned nil")
		}
		if len(validator.rules) == 0 {
			t.Error("Validator should have default rules")
		}
	})

	t.Run("ValidateUnclosedIf", func(t *testing.T) {
		validator := NewTemplateValidator()
		errors := validator.Validate("test.html", "{% if true %}")

		if len(errors) == 0 {
			t.Error("Should detect unclosed if")
		}
	})

	t.Run("ValidateUnclosedFor", func(t *testing.T) {
		validator := NewTemplateValidator()
		errors := validator.Validate("test.html", "{% for i in items %}")

		if len(errors) == 0 {
			t.Error("Should detect unclosed for")
		}
	})

	t.Run("ValidateClosedBlocks", func(t *testing.T) {
		validator := NewTemplateValidator()
		errors := validator.Validate("test.html", "{% if true %}{% endif %}")

		hasIfError := false
		for _, err := range errors {
			if strings.Contains(err.Message, "If statement") {
				hasIfError = true
			}
		}
		if hasIfError {
			t.Error("Should not report closed if as unclosed")
		}
	})

	t.Run("AddRule", func(t *testing.T) {
		validator := NewTemplateValidator()
		initialRules := len(validator.rules)

		validator.AddRule(ValidationRule{
			Name:     "test_rule",
			Severity: "error",
			Check: func(templateName, source string) []*ValidationError {
				return nil
			},
		})

		if len(validator.rules) != initialRules+1 {
			t.Error("AddRule should add a rule")
		}
	})
}

// TestValidationError tests the ValidationError type
func TestValidationError(t *testing.T) {
	err := NewValidationError("warning", "test warning", "test.html", 5, 10)
	if err == nil {
		t.Fatal("NewValidationError returned nil")
	}
	if err.Severity != "warning" {
		t.Errorf("Severity = %q, want 'warning'", err.Severity)
	}
}

// TestGlobalTemplateValidator tests the global template validator
func TestGlobalTemplateValidator(t *testing.T) {
	if GlobalTemplateValidator == nil {
		t.Error("GlobalTemplateValidator should not be nil")
	}
}

// TestErrorRecovery tests the ErrorRecovery type
func TestErrorRecovery(t *testing.T) {
	t.Run("NewErrorRecovery", func(t *testing.T) {
		recovery := NewErrorRecovery()
		if recovery == nil {
			t.Fatal("NewErrorRecovery returned nil")
		}
		if len(recovery.strategies) == 0 {
			t.Error("Should have default strategies")
		}
	})

	t.Run("RecoverUndefined", func(t *testing.T) {
		recovery := NewErrorRecovery()
		err := NewEnhancedTemplateError(ErrorTypeUndefined, "undefined var", "t.html", 1, 1)
		ctx := NewContext()

		result, recoverErr := recovery.Recover(err, ctx)
		if recoverErr != nil {
			t.Errorf("Should recover from undefined error: %v", recoverErr)
		}
		if result != "" {
			t.Errorf("Should return empty string, got %q", result)
		}
	})

	t.Run("RecoverDivisionByZero", func(t *testing.T) {
		recovery := NewErrorRecovery()
		err := NewEnhancedTemplateError(ErrorTypeRuntime, "division by zero", "t.html", 1, 1)
		ctx := NewContext()

		result, recoverErr := recovery.Recover(err, ctx)
		if recoverErr != nil {
			t.Errorf("Should recover from division by zero: %v", recoverErr)
		}
		if result != "0" {
			t.Errorf("Should return '0', got %q", result)
		}
	})

	t.Run("CannotRecoverOtherRuntime", func(t *testing.T) {
		recovery := NewErrorRecovery()
		err := NewEnhancedTemplateError(ErrorTypeRuntime, "other error", "t.html", 1, 1)
		ctx := NewContext()

		_, recoverErr := recovery.Recover(err, ctx)
		if recoverErr == nil {
			t.Error("Should not recover from other runtime errors")
		}
	})

	t.Run("NoStrategyForType", func(t *testing.T) {
		recovery := NewErrorRecovery()
		err := NewEnhancedTemplateError(ErrorTypeInheritance, "test", "t.html", 1, 1)
		ctx := NewContext()

		_, recoverErr := recovery.Recover(err, ctx)
		if recoverErr == nil {
			t.Error("Should not recover without strategy")
		}
	})

	t.Run("AddStrategy", func(t *testing.T) {
		recovery := NewErrorRecovery()
		recovery.AddStrategy(ErrorTypeFilter, func(err *EnhancedTemplateError, ctx Context) (string, error) {
			return "default", nil
		})

		err := NewEnhancedTemplateError(ErrorTypeFilter, "test", "t.html", 1, 1)
		result, recoverErr := recovery.Recover(err, NewContext())

		if recoverErr != nil {
			t.Error("Should recover with custom strategy")
		}
		if result != "default" {
			t.Errorf("Expected 'default', got %q", result)
		}
	})
}

// TestGlobalErrorRecovery tests the global error recovery
func TestGlobalErrorRecovery(t *testing.T) {
	if GlobalErrorRecovery == nil {
		t.Error("GlobalErrorRecovery should not be nil")
	}
}

// TestCompatibilityErrors tests backward compatibility error types
func TestCompatibilityErrors(t *testing.T) {
	t.Run("NewTemplateError", func(t *testing.T) {
		err := NewTemplateError("test.html", 5, 10, "error: %s", "message")
		if err == nil {
			t.Fatal("NewTemplateError returned nil")
		}
		if !strings.Contains(err.Message, "error: message") {
			t.Error("Message should include formatted string")
		}
	})

	t.Run("NewSyntaxError", func(t *testing.T) {
		err := NewSyntaxError("test.html", 5, 10, "syntax: %s", "issue")
		if err == nil {
			t.Fatal("NewSyntaxError returned nil")
		}
		if err.Type != ErrorTypeSyntax {
			t.Error("Type should be SyntaxError")
		}
	})

	t.Run("NewRuntimeError", func(t *testing.T) {
		err := NewRuntimeError("test.html", 5, 10, "runtime: %s", "issue")
		if err == nil {
			t.Fatal("NewRuntimeError returned nil")
		}

		// Test Error method with stack
		err.Stack = []string{"frame1", "frame2"}
		errStr := err.Error()
		if !strings.Contains(errStr, "Template stack") {
			t.Error("RuntimeError.Error() should include stack")
		}
	})

	t.Run("NewUndefinedError", func(t *testing.T) {
		err := NewUndefinedError("myvar", "test.html", 5, 10)
		if err == nil {
			t.Fatal("NewUndefinedError returned nil")
		}
		if err.Variable != "myvar" {
			t.Errorf("Variable = %q, want 'myvar'", err.Variable)
		}

		errStr := err.Error()
		if !strings.Contains(errStr, "myvar") {
			t.Error("Error() should include variable name")
		}
		if !strings.Contains(errStr, "test.html") {
			t.Error("Error() should include template name")
		}
	})

	t.Run("UndefinedErrorWithoutLocation", func(t *testing.T) {
		err := NewUndefinedError("myvar", "", 0, 0)
		errStr := err.Error()
		if !strings.Contains(errStr, "myvar") {
			t.Error("Error() should include variable name")
		}
	})
}
