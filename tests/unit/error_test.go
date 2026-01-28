package miya_test

import (
	"fmt"
	miya "github.com/zipreport/miya"
	"strings"
	"testing"
)

func TestTemplateError(t *testing.T) {
	tests := []struct {
		name             string
		error            *miya.EnhancedTemplateError
		expectedError    string
		expectedDetailed bool
	}{
		{
			name:             "basic syntax error",
			error:            miya.NewEnhancedTemplateError(miya.ErrorTypeSyntax, "unexpected token", "test.html", 5, 10),
			expectedError:    "SyntaxError: unexpected token in template 'test.html' at line 5, column 10",
			expectedDetailed: true,
		},
		{
			name:             "undefined variable error",
			error:            miya.NewEnhancedTemplateError(miya.ErrorTypeUndefined, "variable not found", "", 0, 0),
			expectedError:    "UndefinedError: variable not found",
			expectedDetailed: false,
		},
		{
			name: "error with source context",
			error: miya.NewEnhancedTemplateError(miya.ErrorTypeSyntax, "missing endif", "test.html", 3, 1).
				WithSource("Line 1\nLine 2\n{% if condition %}\nLine 4\nLine 5").
				WithSuggestion("Add {% endif %} to close the if statement"),
			expectedError:    "SyntaxError: missing endif in template 'test.html' at line 3, column 1",
			expectedDetailed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test basic error message
			if tt.error.Error() != tt.expectedError {
				t.Errorf("expected error message %q, got %q", tt.expectedError, tt.error.Error())
			}

			// Test detailed error message
			detailed := tt.error.DetailedError()
			if tt.expectedDetailed {
				if !strings.Contains(detailed, tt.error.Message) {
					t.Errorf("detailed error should contain message %q", tt.error.Message)
				}
				if tt.error.Suggestion != "" && !strings.Contains(detailed, tt.error.Suggestion) {
					t.Errorf("detailed error should contain suggestion %q", tt.error.Suggestion)
				}
			}
		})
	}
}

func TestSyntaxErrorHelper(t *testing.T) {
	helper := miya.NewSyntaxErrorHelper()

	tests := []struct {
		name             string
		errorMessage     string
		expectSuggestion bool
	}{
		{
			name:             "unknown filter",
			errorMessage:     "unknown filter: nonexistent",
			expectSuggestion: true,
		},
		{
			name:             "unexpected closing brace",
			errorMessage:     "unexpected '}'",
			expectSuggestion: true,
		},
		{
			name:             "random error",
			errorMessage:     "some random error",
			expectSuggestion: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestion := helper.GetSuggestion(tt.errorMessage)
			hasSuggestion := suggestion != ""

			if hasSuggestion != tt.expectSuggestion {
				t.Errorf("expected suggestion: %v, got suggestion: %v (%q)",
					tt.expectSuggestion, hasSuggestion, suggestion)
			}
		})
	}
}

func TestTemplateValidator(t *testing.T) {
	validator := miya.NewTemplateValidator()

	tests := []struct {
		name           string
		templateName   string
		source         string
		expectedErrors int
	}{
		{
			name:           "valid template",
			templateName:   "valid.html",
			source:         "{% if condition %}Hello{% endif %}",
			expectedErrors: 0,
		},
		{
			name:           "unclosed if",
			templateName:   "invalid.html",
			source:         "{% if condition %}Hello",
			expectedErrors: 1,
		},
		{
			name:           "unclosed for",
			templateName:   "invalid.html",
			source:         "{% for item in items %}{{ item }}",
			expectedErrors: 1,
		},
		{
			name:           "multiple errors",
			templateName:   "invalid.html",
			source:         "{% if condition %}{% for item in items %}{{ undefined_var }}",
			expectedErrors: 3, // unclosed if, unclosed for, undefined var
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validator.Validate(tt.templateName, tt.source)

			if len(errors) != tt.expectedErrors {
				t.Errorf("expected %d errors, got %d", tt.expectedErrors, len(errors))
				for _, err := range errors {
					t.Logf("  Error: %s", err.Error())
				}
			}
		})
	}
}

func TestErrorRecovery(t *testing.T) {
	recovery := miya.NewErrorRecovery()
	ctx := miya.NewContext()

	tests := []struct {
		name           string
		error          *miya.EnhancedTemplateError
		expectRecovery bool
		expectedResult string
	}{
		{
			name:           "undefined variable",
			error:          miya.NewEnhancedTemplateError(miya.ErrorTypeUndefined, "variable not found", "", 0, 0),
			expectRecovery: true,
			expectedResult: "",
		},
		{
			name:           "division by zero",
			error:          miya.NewEnhancedTemplateError(miya.ErrorTypeRuntime, "division by zero", "", 0, 0),
			expectRecovery: true,
			expectedResult: "0",
		},
		{
			name:           "syntax error",
			error:          miya.NewEnhancedTemplateError(miya.ErrorTypeSyntax, "unexpected token", "", 0, 0),
			expectRecovery: false,
			expectedResult: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := recovery.Recover(tt.error, ctx)

			if tt.expectRecovery {
				if err != nil {
					t.Errorf("expected recovery, but got error: %v", err)
				}
				if result != tt.expectedResult {
					t.Errorf("expected result %q, got %q", tt.expectedResult, result)
				}
			} else {
				if err == nil {
					t.Errorf("expected no recovery, but got result: %q", result)
				}
			}
		})
	}
}

func TestErrorHandler(t *testing.T) {
	handler := miya.DefaultErrorHandler()

	tests := []struct {
		name           string
		error          error
		expectedLength int // Approximate length check
	}{
		{
			name: "template error with details",
			error: miya.NewEnhancedTemplateError(miya.ErrorTypeSyntax, "test error", "test.html", 1, 1).
				WithSource("test source").
				WithSuggestion("test suggestion"),
			expectedLength: 100, // Should be longer due to details
		},
		{
			name:           "regular error",
			error:          fmt.Errorf("regular error"),
			expectedLength: 10, // Should be short
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatted := handler.FormatError(tt.error)

			if len(formatted) < tt.expectedLength {
				t.Errorf("expected formatted error to be at least %d characters, got %d",
					tt.expectedLength, len(formatted))
			}

			t.Logf("Formatted error:\n%s", formatted)
		})
	}
}

func TestDebugTracer(t *testing.T) {
	tracer := miya.NewDebugTracer()

	// Test basic functionality
	tracer.Enable()
	tracer.SetLevel(miya.DebugLevelDetailed)

	// Record some events
	variables := map[string]interface{}{
		"name": "test",
		"age":  25,
	}

	tracer.TraceEvent("template_start", "test.html", 1, 1, "Starting template", variables)
	tracer.TraceEvent("variable_access", "test.html", 5, 10, "Accessing variable", nil)
	tracer.TraceEvent("template_end", "test.html", 0, 0, "Template finished", nil)

	// Test event retrieval
	events := tracer.GetEvents()
	if len(events) != 3 {
		t.Errorf("expected 3 events, got %d", len(events))
	}

	// Test summary
	summary := tracer.GetSummary()
	if !strings.Contains(summary, "3 events") {
		t.Errorf("summary should mention 3 events")
	}

	// Test detailed log
	detailed := tracer.GetDetailedLog()
	if !strings.Contains(detailed, "template_start") {
		t.Errorf("detailed log should contain event types")
	}

	// Test filtering
	tracer.Clear()
	tracer.AddFilter("template_start")

	tracer.TraceEvent("template_start", "test.html", 1, 1, "Starting", nil)
	tracer.TraceEvent("variable_access", "test.html", 5, 10, "Accessing", nil)

	events = tracer.GetEvents()
	if len(events) != 1 {
		t.Errorf("expected 1 filtered event, got %d", len(events))
	}

	tracer.Disable()
}

func TestPerformanceProfiler(t *testing.T) {
	profiler := miya.NewPerformanceProfiler()
	profiler.Enable()

	// Record some measurements
	done1 := profiler.StartMeasurement("template_render")
	// Simulate work
	done1()

	done2 := profiler.StartMeasurement("filter_apply")
	// Simulate work
	done2()

	done3 := profiler.StartMeasurement("template_render")
	// Simulate work
	done3()

	// Get measurements
	measurements := profiler.GetMeasurements()

	if len(measurements) != 2 {
		t.Errorf("expected 2 measurement types, got %d", len(measurements))
	}

	templateMeasurement := measurements["template_render"]
	if templateMeasurement == nil {
		t.Errorf("expected template_render measurement")
	} else if templateMeasurement.Count != 2 {
		t.Errorf("expected 2 template_render measurements, got %d", templateMeasurement.Count)
	}

	// Test report generation
	report := profiler.GetReport()
	if !strings.Contains(report, "template_render") {
		t.Errorf("report should contain measurement names")
	}

	// Test clear
	profiler.Clear()
	measurements = profiler.GetMeasurements()
	if len(measurements) != 0 {
		t.Errorf("expected 0 measurements after clear, got %d", len(measurements))
	}

	profiler.Disable()
}

func TestInteractiveDebugger(t *testing.T) {
	debugger := miya.NewInteractiveDebugger()
	debugger.Enable()

	// Test breakpoint management
	debugger.SetBreakpoint("test.html", 5)
	debugger.SetBreakpoint("test.html", 10)
	debugger.SetBreakpoint("other.html", 3)

	// Test watch variables
	debugger.Watch("name")
	debugger.Watch("age")

	// Create test context
	ctx := miya.NewContext()
	ctx.Set("name", "John")
	ctx.Set("age", 30)

	// Test watched values
	values := debugger.GetWatchedValues(ctx)

	if len(values) != 2 {
		t.Errorf("expected 2 watched values, got %d", len(values))
	}

	if values["name"] != "John" {
		t.Errorf("expected name='John', got %v", values["name"])
	}

	if values["age"] != 30 {
		t.Errorf("expected age=30, got %v", values["age"])
	}

	debugger.Disable()
}
