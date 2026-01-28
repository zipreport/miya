# Miya Complex Template Examples

This directory contains comprehensive Jinja2 template examples demonstrating most of the features supported by the Miya template engine.

## Directory Structure

```
examples/showcase/
├── README.md                          # This file
├── demo.go                            # Go demo script to render templates
└── templates/
    ├── layouts/
    │   └── base.html                  # Base layout template
    ├── components/
    │   └── macros.html                # Reusable macro library
    └── pages/
        ├── showcase.html              # Feature showcase page
        └── dashboard.html             # Practical dashboard example
```

## Features Demonstrated

### 1. Template Inheritance (`layouts/base.html`)
- Base template with multiple blocks
- Block overriding and extension
- `{{ super() }}` for parent content inclusion
- Nested blocks
- Default block content

### 2. Macros (`components/macros.html`)
-  Simple macros without parameters
-  Macros with required parameters
-  Macros with default parameters
-  Complex macros with conditional logic
-  Macros using filters and tests
-  Macros with loops inside
-  Recursive macros (nested structures)
-  Macros with variable unpacking
-  Macros using namespaces
-  Filter chaining in macros
-  Form input macros
-  Type-testing macros

### 3. Feature Showcase (`pages/showcase.html`)

#### Variables and Filters
- Variable substitution with defaults
- Filter chaining (trim|title|truncate)
- String filters: upper, lower, capitalize, title
- Numeric filters: round, abs, int, float
- Collection filters: sort, reverse, first, last, length, join
- HTML filters: escape, safe, urlencode
- Filter blocks (`{% filter %}`)

#### Control Structures
- If/elif/else conditions
- Ternary operators (`{{ 'value' if condition else 'other' }}`)
- Complex conditional logic

#### Loops and Iteration
- Basic for loops
- Loop variables: `loop.index`, `loop.index0`, `loop.first`, `loop.last`, `loop.revindex`
- `loop.cycle()` for alternating values
- `loop.depth` for nested loops
- Conditional iteration (`{% for item in items if condition %}`)
- Dictionary iteration with `.items()`
- Nested loops

#### Tests and Conditionals
- Type tests: `is string`, `is number`, `is boolean`, `is iterable`, `is mapping`
- State tests: `is defined`, `is none`, `is even`, `is odd`
- Negated tests: `is not`

#### Advanced Filter Features
- Numeric transformations
- Collection operations
- HTML escaping and safety
- Filter blocks for bulk transformations

#### Assignments and Variables
- Simple assignment (`{% set var = value %}`)
- Multiple assignment / tuple unpacking (`{% set x, y, z = coords %}`)
- Namespace for mutable state (`{% set ns = namespace(count=0) %}`)
- Block assignment (assigning HTML blocks to variables)

#### Macros Usage
- Importing macros (`{% from "..." import ... %}`)
- Calling macros with different parameter combinations
- Complex macro examples

#### Data Tables
- Dynamic table generation
- Striped rows
- Loop-based table construction

#### Global Functions
- `range()` - Generate number sequences
- `zip()` - Combine multiple iterables
- `enumerate()` - Get index-value pairs
- `dict()` - Create dictionaries
- `namespace()` - Create mutable namespaces
- `cycler()` - Create cycling values
- `joiner()` - Join with separators

#### Slicing and Indexing
- Array slicing (`arr[:3]`, `arr[-3:]`, `arr[2:5]`)
- Step slicing (`arr[::2]`)
- Reverse slicing (`arr[::-1]`)
- String slicing

#### Whitespace Control
- Inline whitespace control (`{%-` and `-%}`)
- `trim_blocks` and `lstrip_blocks` configuration
- Tight spacing control

#### Comments
- Single-line comments (`{# comment #}`)
- Multi-line comments

#### Expressions and Operators
- Arithmetic: `+`, `-`, `*`, `/`, `//`, `%`, `**`
- Comparison: `>`, `<`, `==`, `!=`, `>=`, `<=`
- Logical: `and`, `or`, `not`
- String concatenation: `~`

#### Do Statements
- Executing expressions for side effects (`{% do list.append(item) %}`)

### 4. Dashboard Example (`pages/dashboard.html`)

A practical example demonstrating:
- Real-world layout structure
- Statistics cards with dynamic data
- Activity feed with loops
- Data tables
- Visual charts using CSS
- Team member cards
- Conditional rendering
- Complex data structures

## Running the Examples

### Using Go

```bash
cd examples/showcase
go run demo.go
```

This will generate:
- `output_feature_showcase.html` - Complete feature demonstration
- `output_dashboard.html` - Practical dashboard example

### Manual Template Rendering

You can also use the templates in your own Go code:

```go
package main

import (
    "fmt"
    "github.com/zipreport/miya"
    "github.com/zipreport/miya/loader"
    "github.com/zipreport/miya/parser"
)

// SimpleTemplateParser implements loader.TemplateParser
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

func main() {
    // Create environment first
    env := miya.NewEnvironment(
        miya.WithAutoEscape(true),
    )

    // Create template parser and filesystem loader
    templateParser := &SimpleTemplateParser{env: env}
    fsLoader := loader.NewFileSystemLoader([]string{"templates"}, templateParser)
    env.SetLoader(fsLoader)

    // Create context
    ctx := miya.NewContextFrom(map[string]interface{}{
        "site_name": "My Site",
        "username": "alice",
        // ... more context variables
    })

    // Load and render template
    tmpl, err := env.GetTemplate("pages/showcase.html")
    if err != nil {
        panic(err)
    }

    output, err := tmpl.Render(ctx)
    if err != nil {
        panic(err)
    }

    fmt.Println(output)
}
```

## Template Features Reference

### Template Syntax

| Feature | Syntax | Example |
|---------|--------|---------|
| Variable | `{{ var }}` | `{{ username }}` |
| Comment | `{# comment #}` | `{# TODO: update #}` |
| Statement | `{% statement %}` | `{% if condition %}` |
| Whitespace control | `{%-`, `-%}` | `{%- for item in items -%}` |

### Control Flow

```jinja2
{# If/elif/else #}
{% if condition %}
    ...
{% elif other_condition %}
    ...
{% else %}
    ...
{% endif %}

{# For loop #}
{% for item in items %}
    {{ item }}
{% endfor %}

{# For loop with condition #}
{% for item in items if item.active %}
    {{ item }}
{% endfor %}

{# While loop #}
{% while condition %}
    ...
{% endwhile %}
```

### Filters

```jinja2
{{ value|filter }}
{{ value|filter(arg) }}
{{ value|filter1|filter2|filter3 }}

{% filter upper %}
    content
{% endfilter %}
```

### Tests

```jinja2
{% if value is defined %}...{% endif %}
{% if value is none %}...{% endif %}
{% if value is string %}...{% endif %}
{% if value is number %}...{% endif %}
{% if value is even %}...{% endif %}
{% if value is odd %}...{% endif %}
```

### Macros

```jinja2
{% macro name(param1, param2='default') %}
    {{ param1 }} {{ param2 }}
{% endmacro %}

{{ name('value') }}
{{ name('value', 'custom') }}
```

### Inheritance

```jinja2
{# base.html #}
{% block content %}Default{% endblock %}

{# child.html #}
{% extends "base.html" %}
{% block content %}
    {{ super() }}
    Additional content
{% endblock %}
```

## Best Practices

1. **Use Template Inheritance** - Create base layouts for consistent structure
2. **Create Reusable Macros** - Avoid duplication with component macros
3. **Enable Auto-escaping** - Prevent XSS attacks by escaping HTML by default
4. **Use Whitespace Control** - Keep output clean with `{%-` and `-%}`
5. **Leverage Filters** - Transform data in templates rather than in code
6. **Test Your Templates** - Verify template rendering with different data sets
7. **Document Complex Logic** - Use comments to explain non-obvious template code
8. **Keep Business Logic in Code** - Templates should focus on presentation

## Common Patterns

### Conditional CSS Classes

```jinja2
<div class="item {% if item.active %}active{% endif %}">
```

### Loop with Index

```jinja2
{% for item in items %}
    {{ loop.index }}. {{ item }}
{% endfor %}
```

### Safe HTML Output

```jinja2
{{ user_content|escape }}  {# Escaped by default if auto-escape is on #}
{{ trusted_html|safe }}    {# Mark as safe HTML #}
```

### Default Values

```jinja2
{{ username|default('Guest') }}
{{ user.name|default('Unknown', true) }}  {# true = use default for empty strings #}
```

### Nested Data Access

```jinja2
{{ user.profile.address.city }}
{{ data['key']['nested_key'] }}
```

## Performance Tips

1. **Use Template Caching** - Templates are automatically cached after first compilation
2. **Minimize Filter Chains** - Long filter chains can impact performance
3. **Pre-process Data** - Complex calculations should happen in Go code, not templates
4. **Avoid Deep Recursion** - Recursive macros can be slow for deep structures
5. **Use Whitespace Control** - Reduces output size

## Troubleshooting

### Template Not Found
- Check loader configuration
- Verify template path is relative to loader's base directory
- Ensure file extension matches loader configuration

### Variable Undefined
- Check context contains the variable
- Use `|default()` filter for optional variables
- Enable debug mode for better error messages

### Macro Not Found
- Verify macro is imported correctly
- Check macro name spelling
- Ensure macro is defined before use

### Unexpected Output
- Check whitespace control settings
- Verify auto-escape configuration
- Review filter application order

## Contributing

Found a bug or want to add more examples? Contributions are welcome!

## License

MIT License - See main project LICENSE file
