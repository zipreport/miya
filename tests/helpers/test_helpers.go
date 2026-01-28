package helpers

import (
	"testing"

	jinja2 "github.com/zipreport/miya"
)

// TestCase represents a standard test case structure
type TestCase struct {
	Name          string
	Template      string
	Context       map[string]interface{}
	Expected      string
	ShouldError   bool
	ErrorContains string
}

// RenderTestCase executes a single test case with standard validation
func RenderTestCase(t *testing.T, env *jinja2.Environment, tc TestCase) {
	t.Run(tc.Name, func(t *testing.T) {
		ctx := CreateContext(tc.Context)
		result, err := env.RenderString(tc.Template, ctx)

		if tc.ShouldError {
			if err == nil {
				t.Errorf("Expected error but got none")
				return
			}
			if tc.ErrorContains != "" && !contains(err.Error(), tc.ErrorContains) {
				t.Errorf("Expected error to contain '%s', got: %v", tc.ErrorContains, err)
			}
			return
		}

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if result != tc.Expected {
			t.Errorf("Expected '%s', got '%s'", tc.Expected, result)
		}
	})
}

// RenderTestCases executes multiple test cases
func RenderTestCases(t *testing.T, env *jinja2.Environment, cases []TestCase) {
	for _, tc := range cases {
		RenderTestCase(t, env, tc)
	}
}

// CreateContext creates a jinja2.Context from a map
func CreateContext(data map[string]interface{}) jinja2.Context {
	ctx := jinja2.NewContext()
	for key, value := range data {
		ctx.Set(key, value)
	}
	return ctx
}

// CreateEnvironment creates a standard test environment
func CreateEnvironment() *jinja2.Environment {
	return jinja2.NewEnvironment()
}

// contains is a simple string contains check
func contains(s, substr string) bool {
	return len(substr) == 0 || len(s) >= len(substr) && (s == substr || containsAt(s, substr, 0, len(s)-len(substr)+1))
}

func containsAt(s, substr string, start, end int) bool {
	for i := start; i < end; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// AssertError checks that an error occurred with optional message validation
func AssertError(t *testing.T, err error, expectedContains string) {
	t.Helper()
	if err == nil {
		t.Error("Expected error but got none")
		return
	}
	if expectedContains != "" && !contains(err.Error(), expectedContains) {
		t.Errorf("Expected error to contain '%s', got: %v", expectedContains, err)
	}
}

// AssertNoError checks that no error occurred
func AssertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

// AssertEqual checks string equality with better error messages
func AssertEqual(t *testing.T, expected, actual string) {
	t.Helper()
	if expected != actual {
		t.Errorf("Expected '%s', got '%s'", expected, actual)
	}
}
