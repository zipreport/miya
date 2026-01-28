package miya_test

import (
	miya "github.com/zipreport/miya"
	"strings"
	"testing"

	"github.com/zipreport/miya/runtime"
)

func TestEnhancedErrorHandling(t *testing.T) {
	env := miya.NewEnvironment(miya.WithStrictUndefined(true)) // Enable strict undefined for error testing

	t.Run("Undefined variable error", func(t *testing.T) {
		template := `{{ undefined_var }}`

		ctx := miya.NewContext()
		_, err := env.RenderString(template, ctx)
		if err == nil {
			t.Fatal("Expected error for undefined variable")
		}

		if runtimeErr, ok := err.(*runtime.RuntimeError); ok {
			if runtimeErr.Type != runtime.ErrorTypeUndefined {
				t.Errorf("Expected UndefinedError, got %s", runtimeErr.Type)
			}

			if !strings.Contains(runtimeErr.Message, "undefined_var") {
				t.Errorf("Expected error message to contain variable name, got: %s", runtimeErr.Message)
			}

			if runtimeErr.Suggestion == "" {
				t.Error("Expected error to have a suggestion")
			}
		} else {
			t.Errorf("Expected RuntimeError, got %T: %v", err, err)
		}
	})

	t.Run("Division by zero error", func(t *testing.T) {
		template := `{{ 10 / 0 }}`

		ctx := miya.NewContext()
		_, err := env.RenderString(template, ctx)
		if err == nil {
			t.Fatal("Expected error for division by zero")
		}

		if runtimeErr, ok := err.(*runtime.RuntimeError); ok {
			if runtimeErr.Type != runtime.ErrorTypeMath {
				t.Errorf("Expected MathError, got %s", runtimeErr.Type)
			}

			if !strings.Contains(strings.ToLower(runtimeErr.Message), "division by zero") {
				t.Errorf("Expected error message about division by zero, got: %s", runtimeErr.Message)
			}
		} else {
			t.Errorf("Expected RuntimeError, got %T: %v", err, err)
		}
	})

	t.Run("Type error for invalid operation", func(t *testing.T) {
		template := `{{ "hello" + 42 }}`

		ctx := miya.NewContext()
		_, err := env.RenderString(template, ctx)
		if err == nil {
			t.Fatal("Expected error for invalid type operation")
		}

		if runtimeErr, ok := err.(*runtime.RuntimeError); ok {
			if runtimeErr.Type != runtime.ErrorTypeType {
				t.Errorf("Expected TypeError, got %s", runtimeErr.Type)
			}

			if !strings.Contains(runtimeErr.Message, "addition") {
				t.Errorf("Expected error message about addition, got: %s", runtimeErr.Message)
			}
		} else {
			t.Errorf("Expected RuntimeError, got %T: %v", err, err)
		}
	})

	t.Run("Error with line and column information", func(t *testing.T) {
		template := `Line 1
Line 2 with {{ bad_var }}
Line 3`

		ctx := miya.NewContext()
		_, err := env.RenderString(template, ctx)
		if err == nil {
			t.Fatal("Expected error")
		}

		if runtimeErr, ok := err.(*runtime.RuntimeError); ok {
			// Check that we have line information
			if runtimeErr.Line <= 0 {
				t.Error("Expected line information in error")
			}

			// Test detailed error output
			detailed := runtimeErr.DetailedError()
			if detailed == "" {
				t.Error("Expected detailed error message")
			}

			if !strings.Contains(detailed, "Template Runtime Error") {
				t.Error("Expected detailed error to have proper header")
			}
		} else {
			t.Errorf("Expected RuntimeError, got %T: %v", err, err)
		}
	})

	t.Run("Unknown operator error", func(t *testing.T) {
		// This would require parser support for unknown operators
		// For now, test with a scenario that produces a runtime error
		template := `{{ 5 % 0 }}` // modulo by zero

		ctx := miya.NewContext()
		_, err := env.RenderString(template, ctx)
		if err == nil {
			t.Fatal("Expected error for modulo by zero")
		}

		// Even if it's not our enhanced error, it should still be an error
		errorStr := err.Error()
		if errorStr == "" {
			t.Error("Expected non-empty error message")
		}
	})
}

func TestErrorSuggestions(t *testing.T) {
	env := miya.NewEnvironment(miya.WithStrictUndefined(true)) // Enable strict undefined for error testing

	testCases := []struct {
		name                 string
		template             string
		expectedErrorType    string
		shouldHaveSuggestion bool
	}{
		{
			name:                 "Undefined variable",
			template:             `{{ missing_var }}`,
			expectedErrorType:    runtime.ErrorTypeUndefined,
			shouldHaveSuggestion: true,
		},
		{
			name:                 "Math error",
			template:             `{{ 1 / 0 }}`,
			expectedErrorType:    runtime.ErrorTypeMath,
			shouldHaveSuggestion: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := miya.NewContext()
			_, err := env.RenderString(tc.template, ctx)

			if err == nil {
				t.Fatal("Expected an error")
			}

			if runtimeErr, ok := err.(*runtime.RuntimeError); ok {
				if runtimeErr.Type != tc.expectedErrorType {
					t.Errorf("Expected %s, got %s", tc.expectedErrorType, runtimeErr.Type)
				}

				if tc.shouldHaveSuggestion && runtimeErr.Suggestion == "" {
					t.Error("Expected error to have a suggestion")
				}
			}
		})
	}
}
