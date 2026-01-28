# Miya Engine Examples

This directory contains examples demonstrating the features and capabilities of the Miya template engine.

## Directory Structure

```
examples/
├── README.md              # This file
├── tutorial/              # Step-by-step beginner tutorial (START HERE)
│   ├── step1_hello_world.go
│   ├── step2_variables_filters.go
│   ├── step3_control_flow.go
│   ├── step4_inheritance.go
│   ├── step5_macros.go
│   ├── step6_filesystem.go
│   └── templates/         # Template files for step 6
├── features/              # Feature-specific examples (learning)
│   ├── inheritance/       # Template inheritance
│   ├── control-structures/ # If/for/set/with
│   ├── filters/           # Built-in filters
│   ├── macros-includes/   # Macros and includes
│   ├── comprehensions/    # List/dict comprehensions
│   ├── advanced/          # Filter blocks, do statements
│   ├── tests-operators/   # Tests and operators
│   └── global-functions/  # range, cycler, namespace, etc.
├── go/                    # Advanced Go examples
│   ├── basic/             # Fundamental features
│   ├── advanced/          # Advanced features
│   ├── complex/           # Complex template organization
│   ├── comprehensive/     # Full feature demonstration
│   └── web-server/        # HTTP server example
└── showcase/              # Complex template showcase
    ├── demo.go            # Demo runner
    └── templates/         # Professional template organization
```

## Getting Started

### New to Miya? Start with the Tutorial

The `tutorial/` directory contains a 6-step guide that teaches you Miya from scratch:

```bash
cd examples/tutorial

# Step 1: Basic template rendering
go run step1_hello_world.go

# Step 2: Variables and filters
go run step2_variables_filters.go

# Step 3: Conditionals and loops
go run step3_control_flow.go

# Step 4: Template inheritance
go run step4_inheritance.go

# Step 5: Macros and components
go run step5_macros.go

# Step 6: Loading templates from disk
go run step6_filesystem.go
```

Each step builds on the previous one and includes detailed comments explaining the concepts.

### Feature Examples (`features/`)

Best for learning specific features. Each directory focuses on one aspect:

```bash
# Learn template inheritance
cd examples/features/inheritance
go run main.go

# Learn control structures
cd examples/features/control-structures
go run main.go

# Learn filters
cd examples/features/filters
go run main.go
```

### Go Examples (`go/`)

Progressive complexity examples:

**1. Basic Example (`go/basic/`)**
```bash
cd examples/go/basic
go run main.go
```
Demonstrates:
- Variable substitution
- Filters (upper, lower, title, truncate, join, etc.)
- Control structures (if/elif/else)
- Loops with loop variables
- Macros

**2. Advanced Example (`go/advanced/`)**
```bash
cd examples/go/advanced
go run main.go
```
Demonstrates:
- Template inheritance (extends/blocks)
- Recursive loops for hierarchical data
- Custom filters and tests
- Auto-escaping configuration
- Advanced features (set, with, namespace, cycler, raw blocks)

**3. Web Server Example (`go/web-server/`)**
```bash
cd examples/go/web-server
go run main.go
# Open http://localhost:8080 in your browser
```
Demonstrates:
- HTTP server with template rendering
- Template inheritance for layout
- Dynamic content from Go structs
- Product catalog with categories
- User session simulation
- XSS protection with auto-escaping

**4. Comprehensive Example (`go/comprehensive/`)**
```bash
cd examples/go/comprehensive
go run comprehensive_example.go
```
Full feature demonstration with all Miya capabilities.

### Showcase (`showcase/`)

Complex, real-world template examples:

```bash
cd examples/showcase
go run demo.go
```

Features professional template organization with layouts, components, and pages.

## Features Demonstrated

### Template Syntax
- **Variables:** `{{ variable }}`, `{{ object.property }}`
- **Filters:** `{{ text|upper }}`, `{{ price|round(2) }}`
- **Tests:** `{% if value is defined %}`, `{% if number is even %}`
- **Comments:** `{# This is a comment #}`

### Control Structures
- **Conditionals:** `{% if %}`, `{% elif %}`, `{% else %}`, `{% endif %}`
- **Loops:** `{% for item in items %}`, with `loop.index`, `loop.first`, `loop.last`
- **Recursive loops:** `{% for item in tree recursive %}`

### Template Composition
- **Inheritance:** `{% extends "base.html" %}`
- **Blocks:** `{% block content %}...{% endblock %}`
- **Includes:** `{% include "partial.html" %}`
- **Macros:** `{% macro name(args) %}...{% endmacro %}`

### Built-in Filters
- **String:** upper, lower, title, capitalize, trim, truncate
- **Lists:** join, first, last, length, sort, reverse
- **Numbers:** round, int, float, abs
- **HTML:** escape, safe, urlencode

### Built-in Tests
- **Type checks:** defined, none, string, number
- **Comparisons:** even, odd, divisibleby
- **String tests:** lower, upper, startswith, endswith
- **Collections:** in, sequence, mapping

### Global Functions
- `range()` - Generate number sequences
- `cycler()` - Cycle through values
- `namespace()` - Create namespaces for variables
- `dict()` - Create dictionaries

## Creating Custom Filters and Tests

```go
// Add custom filter
env.AddFilter("reverse", func(value interface{}, args ...interface{}) (interface{}, error) {
    // Implementation
    return reversedValue, nil
})

// Add custom test
env.AddTest("palindrome", func(value interface{}, args ...interface{}) (bool, error) {
    // Implementation
    return isPalindrome, nil
})
```

## Template Loading Options

```go
// From string (simplest approach)
env := miya.NewEnvironment()
result, err := env.RenderString(templateString, context)

// From filesystem (requires a template parser)
// First, create a parser that implements loader.TemplateParser
type SimpleTemplateParser struct {
    env *miya.Environment
}

func (p *SimpleTemplateParser) ParseTemplate(name, content string) (*parser.TemplateNode, error) {
    tmpl, err := p.env.FromString(content)
    if err != nil {
        return nil, err
    }
    if node, ok := tmpl.AST().(*parser.TemplateNode); ok {
        node.Name = name
        return node, nil
    }
    return nil, fmt.Errorf("failed to extract template node")
}

// Then use it with the filesystem loader
env := miya.NewEnvironment()
templateParser := &SimpleTemplateParser{env: env}
fsLoader := loader.NewFileSystemLoader([]string{"templates"}, templateParser)
env.SetLoader(fsLoader)

tmpl, err := env.GetTemplate("template.html")
result, err := tmpl.Render(context)

// In-memory templates with StringLoader (for template inheritance without files)
env := miya.NewEnvironment()
templateParser := &SimpleTemplateParser{env: env}
stringLoader := loader.NewStringLoader(templateParser)
stringLoader.AddTemplate("base.html", baseTemplateContent)
stringLoader.AddTemplate("child.html", childTemplateContent)
env.SetLoader(stringLoader)

tmpl, err := env.GetTemplate("child.html")
result, err := tmpl.Render(context)
```

## Security Features

- **Auto-escaping:** Enabled by default to prevent XSS attacks
- **Safe filter:** Mark content as safe when needed
- **Escape filter:** Explicitly escape content

## Contributing

Feel free to add more examples demonstrating additional features or use cases!
