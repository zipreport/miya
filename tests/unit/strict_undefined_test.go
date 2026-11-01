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

// TestFilterUndefinedFollowsEnvironment checks that the undefined result of a
// filter (random, first, last, min, max of nothing) takes the environment's
// undefined behavior, as Jinja2's environment.undefined does.
func TestFilterUndefinedFollowsEnvironment(t *testing.T) {
	printing := []string{
		`{{ []|random }}`, `{{ ''|random }}`, `{{ []|first }}`, `{{ []|last }}`,
		`{{ []|min }}`, `{{ []|max }}`, `{{ []|first|upper }}`,
		`{% filter first %}{% endfilter %}`,
	}
	defined := []string{
		`{{ ([]|first) is defined }}`, `{{ ([]|random) is defined }}`,
		`{{ ([]|first) is none }}`,
	}
	modes := map[string]runtime.UndefinedBehavior{
		"silent": runtime.UndefinedSilent, "strict": runtime.UndefinedStrict, "debug": runtime.UndefinedDebug,
	}
	for mode, b := range modes {
		env := miya.NewEnvironment(miya.WithUndefinedBehavior(b))
		for _, tpl := range printing {
			out, err := env.RenderString(tpl, miya.NewContext())
			switch b {
			case runtime.UndefinedStrict:
				if err == nil {
					t.Errorf("%s %s: want error, got %q", mode, tpl, out)
				}
			case runtime.UndefinedDebug:
				if err != nil || !strings.Contains(strings.ToLower(out), "undefined variable") {
					t.Errorf("%s %s: want debug placeholder, got %q, %v", mode, tpl, out, err)
				}
			default:
				if err != nil || out != "" {
					t.Errorf("%s %s: want empty, got %q, %v", mode, tpl, out, err)
				}
			}
		}
		for _, tpl := range defined {
			if out, err := env.RenderString(tpl, miya.NewContext()); err != nil || out != "false" {
				t.Errorf("%s %s: want false, got %q, %v", mode, tpl, out, err)
			}
		}
		if out, err := env.RenderString(`{{ []|first|default('d') }}`, miya.NewContext()); err != nil || out != "d" {
			t.Errorf("%s default: want d, got %q, %v", mode, out, err)
		}
	}
}

// TestStrictFilterUndefinedFailsOnUse checks that a strict undefined filter
// result fails wherever it is used, like an undefined variable: only the
// defined, undefined and none tests and default accept it.
func TestStrictFilterUndefinedFailsOnUse(t *testing.T) {
	strict := miya.NewEnvironment(miya.WithUndefinedBehavior(runtime.UndefinedStrict))
	for _, tpl := range []string{
		`{{ ([]|first) ~ 'a' }}`, `{{ 'a' ~ ([]|first) }}`, `{{ ([]|first) + 'a' }}`,
		`{{ [([]|first)]|join }}`, `{{ "%s"|format([]|first) }}`,
		`{% if []|first %}y{% endif %}`, `{% if 0 %}{% elif []|first %}y{% endif %}`,
		`{{ 'y' if []|first else 'n' }}`, `{% for i in []|first %}y{% endfor %}`,
		`{% set v = []|last %}`, `{{ ([]|first)|upper }}`,
	} {
		out, err := strict.RenderString(tpl, miya.NewContext())
		if err == nil {
			t.Errorf("strict %s: want error, got %q", tpl, out)
		} else if !strings.Contains(strings.ToLower(err.Error()), "undefined") {
			t.Errorf("strict %s: want an undefined error, got %v", tpl, err)
		}
	}
	tolerated := map[string]string{
		`{{ ([]|first) is defined }}`:   "false",
		`{{ ([]|first) is undefined }}`: "true",
		`{{ ([]|first) is none }}`:      "false",
		`{{ []|first|default('d') }}`:   "d",
		`{{ []|first|d('d') }}`:         "d",
		`{{ ([]|first)|d('d') }}`:       "d",
	}
	for tpl, want := range tolerated {
		if out, err := strict.RenderString(tpl, miya.NewContext()); err != nil || out != want {
			t.Errorf("strict %s: want %q, got %q, %v", tpl, want, out, err)
		}
	}
	const uses = `{{ ([]|first) ~ 'a' }}|{{ [([]|first)]|join }}|{% if []|first %}y{% else %}n{% endif %}|{% for i in []|first %}y{% endfor %}|{% set v = []|last %}{{ v is defined }}`
	silent := miya.NewEnvironment()
	if out, err := silent.RenderString(uses, miya.NewContext()); err != nil || out != "a||n||false" {
		t.Errorf("silent: got %q, %v", out, err)
	}
	debug := miya.NewEnvironment(miya.WithUndefinedBehavior(runtime.UndefinedDebug))
	const ph = "{{ undefined variable: first (No first item, sequence was empty.) }}"
	if out, err := debug.RenderString(uses, miya.NewContext()); err != nil || out != ph+"a|"+ph+"|n||false" {
		t.Errorf("debug: got %q, %v", out, err)
	}
}

// TestMapUndefinedItemsFollowEnvironment checks that the undefined items map
// makes (first of an empty sequence) take the environment's undefined
// behavior, as a filter's own undefined result does.
func TestMapUndefinedItemsFollowEnvironment(t *testing.T) {
	strict := miya.NewEnvironment(miya.WithUndefinedBehavior(runtime.UndefinedStrict))
	for _, tpl := range []string{
		`{{ [[], [1]]|map('first')|sum }}`, `{{ [[]]|map('first')|join }}`,
		`{{ [[]]|map('first')|first }}`, `{{ ([[]]|map('first')|first) ~ 'a' }}`,
	} {
		out, err := strict.RenderString(tpl, miya.NewContext())
		if err == nil {
			t.Errorf("strict %s: want error, got %q", tpl, out)
		} else if !strings.Contains(strings.ToLower(err.Error()), "undefined") {
			t.Errorf("strict %s: want an undefined error, got %v", tpl, err)
		}
	}
	if out, err := strict.RenderString(`{{ [[]]|map('first')|list|length }}|{{ [[]]|map('first')|first|default('d') }}`, miya.NewContext()); err != nil || out != "1|d" {
		t.Errorf("strict: got %q, %v", out, err)
	}
	const uses = `{{ [[], [1]]|map('first')|sum }}|{{ [[]]|map('first')|join }}|{{ [[]]|map('first')|first }}`
	silent := miya.NewEnvironment()
	if out, err := silent.RenderString(uses, miya.NewContext()); err != nil || out != "1||" {
		t.Errorf("silent: got %q, %v", out, err)
	}
	debug := miya.NewEnvironment(miya.WithUndefinedBehavior(runtime.UndefinedDebug))
	const ph = "{{ undefined variable: first (No first item, sequence was empty.) }}"
	if out, err := debug.RenderString(uses, miya.NewContext()); err != nil || out != "1|"+ph+"|"+ph {
		t.Errorf("debug: got %q, %v", out, err)
	}
}

// TestStrictFilterBlockChainDefault checks that a filter block's chain passes
// a strict undefined result on to default, as Jinja2 does; it fails when the
// chain ends with it or another filter uses it.
func TestStrictFilterBlockChainDefault(t *testing.T) {
	strict := miya.NewEnvironment(miya.WithUndefinedBehavior(runtime.UndefinedStrict))
	for tpl, want := range map[string]string{
		`{% filter first|default('z') %}{% endfilter %}`:   "z",
		`{% filter first|d('z')|upper %}{% endfilter %}`:   "Z",
		`{% filter first|default('z') %}ab{% endfilter %}`: "a",
	} {
		if out, err := strict.RenderString(tpl, miya.NewContext()); err != nil || out != want {
			t.Errorf("strict %s: want %q, got %q, %v", tpl, want, out, err)
		}
	}
	for _, tpl := range []string{`{% filter first %}{% endfilter %}`, `{% filter first|upper %}{% endfilter %}`, `{% filter first|default([])|first|string %}{% endfilter %}`} {
		if out, err := strict.RenderString(tpl, miya.NewContext()); err == nil {
			t.Errorf("strict %s: want error, got %q", tpl, out)
		}
	}
}
