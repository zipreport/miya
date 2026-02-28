# Error Handling & Debugging

Miya Engine provides a comprehensive error system with detailed error messages, source context, suggestions, and debugging tools.

---

## Table of Contents

1. [Error Types](#error-types)
2. [Undefined Variable Behaviors](#undefined-variable-behaviors)
3. [Common Error Scenarios](#common-error-scenarios)
4. [Error Handling in Go Code](#error-handling-in-go-code)
5. [Enhanced Error Details](#enhanced-error-details)
6. [Template Validation](#template-validation)
7. [Debugging Templates](#debugging-templates)

---

## Error Types

Miya defines 9 error type categories, each targeting a specific class of problem:

| Error Type | Constant | Description |
|------------|----------|-------------|
| **SyntaxError** | `ErrorTypeSyntax` | Invalid template syntax (unclosed tags, bad expressions) |
| **UndefinedError** | `ErrorTypeUndefined` | Accessing an undefined variable (in strict mode) |
| **TypeError** | `ErrorTypeType` | Operation applied to an incompatible type |
| **FilterError** | `ErrorTypeFilter` | Error in filter execution (unknown filter, bad arguments) |
| **TestError** | `ErrorTypeTest` | Error in test execution (unknown test, bad arguments) |
| **RuntimeError** | `ErrorTypeRuntime` | General runtime evaluation error |
| **TemplateError** | `ErrorTypeTemplate` | Template loading or processing error |
| **InheritanceError** | `ErrorTypeInheritance` | Template inheritance error (circular extends, missing blocks) |
| **MacroError** | `ErrorTypeMacro` | Macro definition or invocation error |

The runtime package also defines additional specialized types:

| Error Type | Constant | Description |
|------------|----------|-------------|
| **MathError** | `runtime.ErrorTypeMath` | Division by zero or invalid numeric operations |
| **AccessError** | `runtime.ErrorTypeAccess` | Cannot access attribute on an object |

---

## Undefined Variable Behaviors

Miya supports three modes for handling undefined variables, configurable at environment creation:

### Silent (Default)

Undefined variables return an empty string. No error is raised.

```go
env := miya.NewEnvironment() // UndefinedSilent is the default
```

```html+jinja
{# name is not defined #}
Hello {{ name }}!
{# Output: "Hello !" #}
```

### Strict

Undefined variables raise an `UndefinedError`. Use this in development to catch typos and missing context variables.

```go
env := miya.NewEnvironment(miya.WithStrictUndefined(true))
```

```html+jinja
{# name is not defined #}
Hello {{ name }}!
{# Error: UndefinedError: undefined variable: name #}
```

### Debug

Undefined variables render as a debug placeholder showing the variable name. Useful for identifying missing variables without breaking rendering.

```go
env := miya.NewEnvironment(miya.WithDebugUndefined(true))
```

```html+jinja
{# name is not defined #}
Hello {{ name }}!
{# Output: "Hello {{ name }}!" (placeholder shown) #}
```

You can also set the behavior explicitly:

```go
env := miya.NewEnvironment(
    miya.WithUndefinedBehavior(runtime.UndefinedStrict),
)
```

---

## Common Error Scenarios

### Syntax Errors

Unclosed tags and malformed expressions:

```html+jinja
{# Missing endif #}
{% if user.active %}
  <p>Active</p>

{# Missing closing delimiter #}
{{ user.name

{# Unclosed block #}
{% block content %}
  <p>Hello</p>
```

### Filter Errors

Unknown or misused filters:

```html+jinja
{# Unknown filter #}
{{ name|nonexistent_filter }}

{# Wrong argument type #}
{{ "hello"|round(2) }}
```

### Type Errors

Operations applied to incompatible types:

```html+jinja
{# Cannot add string and int #}
{{ "hello" + 5 }}

{# Iterating over a non-iterable #}
{% for item in 42 %}{{ item }}{% endfor %}
```

### Inheritance Errors

Template inheritance issues:

```html+jinja
{# Template not found #}
{% extends "nonexistent_base.html" %}

{# Circular inheritance #}
{# a.html extends b.html, b.html extends a.html #}
```

### Attribute Access Errors

Accessing attributes that don't exist:

```html+jinja
{# user is a string, not an object #}
{{ user.name }}

{# Accessing attribute on nil #}
{{ none.something }}
```

---

## Error Handling in Go Code

### Basic Error Checking

Template compilation and rendering return errors that should always be checked:

```go
// Compilation errors (syntax issues)
tmpl, err := env.FromString("{{ unclosed")
if err != nil {
    log.Printf("Template compilation failed: %v", err)
}

// Rendering errors (runtime issues)
output, err := tmpl.Render(ctx)
if err != nil {
    log.Printf("Template rendering failed: %v", err)
}

// Template loading errors
tmpl, err := env.GetTemplate("missing.html")
if err != nil {
    log.Printf("Template not found: %v", err)
}
```

### Inspecting Enhanced Errors

Miya errors carry additional context that can be extracted:

```go
output, err := tmpl.Render(ctx)
if err != nil {
    // Check if it's an enhanced template error
    if te, ok := err.(*miya.EnhancedTemplateError); ok {
        fmt.Printf("Type: %s\n", te.Type)
        fmt.Printf("Message: %s\n", te.Message)
        fmt.Printf("Template: %s\n", te.TemplateName)
        fmt.Printf("Line: %d, Column: %d\n", te.Line, te.Column)

        // Get detailed error with source context
        fmt.Println(te.DetailedError())
    }
}
```

### Using the Error Handler

The `ErrorHandler` provides configurable error formatting:

```go
handler := miya.DefaultErrorHandler()
handler.ShowSourceContext = true
handler.ShowStackTrace = true
handler.ShowSuggestions = true

output, err := tmpl.Render(ctx)
if err != nil {
    formatted := handler.FormatError(err)
    log.Println(formatted)
}
```

### Error Recovery

Miya provides an error recovery system for graceful degradation:

```go
recovery := miya.NewErrorRecovery()

// Add custom recovery strategy
recovery.AddStrategy(miya.ErrorTypeUndefined,
    func(err *miya.EnhancedTemplateError, ctx miya.Context) (string, error) {
        return "(missing)", nil // Show placeholder instead of failing
    },
)
```

---

## Enhanced Error Details

Enhanced template errors include:

- **Source context**: Surrounding lines of template code with the error line highlighted
- **Suggestions**: Automated fix suggestions for common mistakes
- **Stack traces**: Template execution stack for nested template calls

Example error output:

```
Template Error: undefined variable: user_name
==================================================

Template: profile.html
Location: Line 15, Column 8

Source context:
     13 | <div class="profile">
     14 |   <h1>User Profile</h1>
  >  15 |   <p>{{ user_name }}</p>
             ^
     16 |   <p>{{ email }}</p>
     17 | </div>

Suggestion: Check if 'user_name' is defined in the template context or if it's spelled correctly
```

### Common Syntax Error Suggestions

Miya's `SyntaxErrorHelper` maps common error patterns to helpful suggestions:

| Error Pattern | Suggestion |
|---------------|------------|
| `unexpected '}'` | Check for missing opening brace or extra closing brace |
| `expected 'endif'` | If statements must be closed with `{% endif %}` |
| `expected 'endfor'` | For loops must be closed with `{% endfor %}` |
| `unknown filter` | Check filter name spelling or register the filter |
| `undefined variable` | Check variable name spelling or ensure it's in the context |
| `template not found` | Check template path and ensure the file exists |
| `circular template inheritance` | Remove circular references in template inheritance |
| `division by zero` | Ensure the denominator is not zero |

---

## Template Validation

Miya includes a template validator that checks for common mistakes before rendering:

```go
validator := miya.NewTemplateValidator()

errors := validator.Validate("page.html", templateSource)
for _, err := range errors {
    fmt.Printf("[%s] %s: %s\n", err.Severity, err.Type, err.Message)
}
```

Built-in validation rules check for:
- Unclosed `{% if %}` blocks
- Unclosed `{% for %}` loops
- Potentially undefined variables

You can add custom validation rules:

```go
validator.AddRule(miya.ValidationRule{
    Name:        "no_raw_html",
    Description: "Warn about unescaped HTML output",
    Severity:    "warning",
    Check: func(templateName, source string) []*miya.ValidationError {
        // Custom validation logic
        return nil
    },
})
```

---

## Debugging Templates

### Using Debug Undefined Mode

The easiest way to find missing variables is to enable debug undefined mode:

```go
env := miya.NewEnvironment(miya.WithDebugUndefined(true))
```

This renders missing variables as visible placeholders instead of empty strings, making it easy to spot which variables are not being passed to the template.

### Template Debugger

Miya includes a template debugger with breakpoints and variable watching:

```go
debugger := miya.NewTemplateDebugger()
debugger.Enable()

// Set breakpoints
debugger.AddBreakpoint("page.html", 15)

// Watch variables
debugger.AddWatchVariable("user")
debugger.AddWatchVariable("items")

// Check watched variables at a breakpoint
if debugger.ShouldBreak("page.html", 15) {
    vars := debugger.GetWatchedVariables(ctx)
    for name, value := range vars {
        fmt.Printf("  %s = %v\n", name, value)
    }
}
```

### Debugging Tips

1. **Start with strict mode** during development to catch undefined variables early
2. **Use the `default` filter** to provide fallbacks for optional variables: `{{ name|default("Anonymous") }}`
3. **Check `is defined`** before accessing optional variables: `{% if user is defined %}...{% endif %}`
4. **Inspect context data** by rendering `{{ variable }}` directly to see its value
5. **Break complex expressions** into intermediate `{% set %}` variables for easier debugging
6. **Use the validator** to catch structural issues before rendering

---

## See Also

- [Miya Limitations](MIYA_LIMITATIONS.md) - Known limitations and workarounds
- [Advanced Features](ADVANCED_FEATURES_GUIDE.md) - Autoescape control, whitespace handling
- [Comprehensions Guide](COMPREHENSIONS_GUIDE.md) - Comprehension-specific error patterns
