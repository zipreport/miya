package miya_test

import (
	miya "github.com/zipreport/miya"
	"strings"
	"testing"

	"github.com/zipreport/miya/runtime"
)

func TestStrictUndefinedHandling(t *testing.T) {
	t.Run("Silent undefined behavior (default)", func(t *testing.T) {
		env := miya.NewEnvironment() // Default is silent

		template := `Hello {{ name }}! You are {{ age }} years old.`
		ctx := miya.NewContext()
		// Deliberately not setting name or age

		result, err := env.RenderString(template, ctx)
		if err != nil {
			t.Fatalf("Silent undefined should not error: %v", err)
		}

		expected := `Hello ! You are  years old.`
		if result != expected {
			t.Errorf("Expected %q, got %q", expected, result)
		}
	})

	t.Run("Strict undefined behavior", func(t *testing.T) {
		env := miya.NewEnvironment(miya.WithStrictUndefined(true))

		template := `Hello {{ name }}!`
		ctx := miya.NewContext()
		// Deliberately not setting name

		_, err := env.RenderString(template, ctx)
		if err == nil {
			t.Fatal("Expected error for strict undefined variable")
		}

		if !strings.Contains(err.Error(), "undefined variable: name") {
			t.Errorf("Expected undefined variable error, got: %v", err)
		}
	})

	t.Run("Debug undefined behavior", func(t *testing.T) {
		env := miya.NewEnvironment(miya.WithDebugUndefined(true))

		template := `Hello {{ name }}! You are {{ age }} years old.`
		ctx := miya.NewContext()
		// Deliberately not setting name or age

		result, err := env.RenderString(template, ctx)
		if err != nil {
			t.Fatalf("Debug undefined should not error: %v", err)
		}

		// Should contain debug information about undefined variables
		if !strings.Contains(result, "undefined variable: name") {
			t.Errorf("Expected debug info for 'name', got: %s", result)
		}
		if !strings.Contains(result, "undefined variable: age") {
			t.Errorf("Expected debug info for 'age', got: %s", result)
		}
	})

	t.Run("Strict undefined with attribute access", func(t *testing.T) {
		env := miya.NewEnvironment(miya.WithStrictUndefined(true))

		template := `User name: {{ user.name }}`
		ctx := miya.NewContext()
		// Deliberately not setting user

		_, err := env.RenderString(template, ctx)
		if err == nil {
			t.Fatal("Expected error for strict undefined attribute access")
		}

		if !strings.Contains(err.Error(), "undefined variable: user") {
			t.Errorf("Expected undefined variable error for user, got: %v", err)
		}
	})

	t.Run("Strict undefined with item access", func(t *testing.T) {
		env := miya.NewEnvironment(miya.WithStrictUndefined(true))

		template := `Value: {{ data[0] }}`
		ctx := miya.NewContext()
		// Deliberately not setting data

		_, err := env.RenderString(template, ctx)
		if err == nil {
			t.Fatal("Expected error for strict undefined item access")
		}

		if !strings.Contains(err.Error(), "undefined variable: data") {
			t.Errorf("Expected undefined variable error for data, got: %v", err)
		}
	})

	t.Run("Mixed defined and undefined variables", func(t *testing.T) {
		env := miya.NewEnvironment(miya.WithStrictUndefined(true))

		template := `Hello {{ name }}! Welcome {{ guest }}!`
		ctx := miya.NewContext()
		ctx.Set("name", "Alice") // Set name but not guest

		_, err := env.RenderString(template, ctx)
		if err == nil {
			t.Fatal("Expected error for undefined guest variable")
		}

		if !strings.Contains(err.Error(), "undefined variable: guest") {
			t.Errorf("Expected undefined variable error for guest, got: %v", err)
		}
	})

	t.Run("Undefined behavior configuration", func(t *testing.T) {
		// Test direct behavior configuration
		env := miya.NewEnvironment(miya.WithUndefinedBehavior(runtime.UndefinedDebug))

		template := `Status: {{ status }}`
		ctx := miya.NewContext()

		result, err := env.RenderString(template, ctx)
		if err != nil {
			t.Fatalf("Debug undefined should not error: %v", err)
		}

		if !strings.Contains(result, "undefined variable: status") {
			t.Errorf("Expected debug info for status, got: %s", result)
		}
	})

	t.Run("Chained undefined access in debug mode", func(t *testing.T) {
		env := miya.NewEnvironment(miya.WithDebugUndefined(true))

		template := `Deep value: {{ obj.nested.value }}`
		ctx := miya.NewContext()
		// Not setting obj

		result, err := env.RenderString(template, ctx)
		if err != nil {
			t.Fatalf("Debug undefined should not error: %v", err)
		}

		// Should show the chained access attempt
		t.Logf("Chained undefined result: %s", result)
	})
}

func TestStrictUndefinedInComplexTemplates(t *testing.T) {
	t.Run("Strict undefined in conditional blocks", func(t *testing.T) {
		env := miya.NewEnvironment(miya.WithStrictUndefined(true))

		template := `{% if show_greeting %}Hello {{ name }}!{% endif %}`
		ctx := miya.NewContext()
		ctx.Set("show_greeting", true)
		// Not setting name

		_, err := env.RenderString(template, ctx)
		if err == nil {
			t.Fatal("Expected error for undefined name in conditional")
		}

		if !strings.Contains(err.Error(), "undefined variable: name") {
			t.Errorf("Expected undefined variable error for name, got: %v", err)
		}
	})

	t.Run("Strict undefined in loops", func(t *testing.T) {
		env := miya.NewEnvironment(miya.WithStrictUndefined(true))

		template := `{% for item in items %}{{ item.name }}: {{ item.undefined_field }}{% endfor %}`
		ctx := miya.NewContext()
		ctx.Set("items", []map[string]interface{}{
			{"name": "Item1"}, // No undefined_field
			{"name": "Item2"}, // No undefined_field
		})

		_, err := env.RenderString(template, ctx)
		if err == nil {
			t.Fatal("Expected error for undefined field in loop")
		}

		// The error might mention attribute access on undefined
		t.Logf("Loop undefined error: %v", err)
	})

	t.Run("Silent undefined allows template completion", func(t *testing.T) {
		env := miya.NewEnvironment(miya.WithStrictUndefined(false)) // Explicitly silent

		template := `
Name: {{ user.name | default("Unknown") }}
Age: {{ user.age | default("N/A") }}
Status: {{ status | default("Active") }}
`

		ctx := miya.NewContext()
		// Not setting any variables

		result, err := env.RenderString(template, ctx)
		if err != nil {
			t.Fatalf("Silent undefined should complete successfully: %v", err)
		}

		// Should use default values when undefined
		if !strings.Contains(result, "Unknown") {
			t.Errorf("Expected default name 'Unknown', got: %s", result)
		}
		if !strings.Contains(result, "N/A") {
			t.Errorf("Expected default age 'N/A', got: %s", result)
		}
		if !strings.Contains(result, "Active") {
			t.Errorf("Expected default status 'Active', got: %s", result)
		}
	})

	t.Run("Strict undefined with filters", func(t *testing.T) {
		env := miya.NewEnvironment(miya.WithStrictUndefined(true))

		template := `Name: {{ name | upper }}`
		ctx := miya.NewContext()
		// Not setting name

		_, err := env.RenderString(template, ctx)
		if err == nil {
			t.Fatal("Expected error for undefined variable with filter")
		}

		if !strings.Contains(err.Error(), "undefined variable: name") {
			t.Errorf("Expected undefined variable error, got: %v", err)
		}
	})
}

func TestUndefinedBehaviorSwitching(t *testing.T) {
	t.Run("Switch from silent to strict", func(t *testing.T) {
		// Start with silent
		env := miya.NewEnvironment()

		template := `Hello {{ name }}!`
		ctx := miya.NewContext()

		// Should work with silent (default)
		result, err := env.RenderString(template, ctx)
		if err != nil {
			t.Fatalf("Silent undefined should work: %v", err)
		}

		expected := "Hello !"
		if result != expected {
			t.Errorf("Expected %q, got %q", expected, result)
		}

		// Now test with strict environment
		strictEnv := miya.NewEnvironment(miya.WithStrictUndefined(true))
		_, err = strictEnv.RenderString(template, ctx)
		if err == nil {
			t.Fatal("Expected error after switching to strict")
		}
	})

	t.Run("Debug mode shows helpful information", func(t *testing.T) {
		env := miya.NewEnvironment(miya.WithDebugUndefined(true))

		template := `User: {{ user.name }} ({{ user.email }})`
		ctx := miya.NewContext()

		result, err := env.RenderString(template, ctx)
		if err != nil {
			t.Fatalf("Debug undefined should not error: %v", err)
		}

		// Should contain debug info showing the undefined variable access
		if !strings.Contains(result, "undefined variable") {
			t.Errorf("Expected debug information in result, got: %s", result)
		}

		t.Logf("Debug undefined result: %s", result)
	})
}
