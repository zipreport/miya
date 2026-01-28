# Miya Engine

A high-performance Jinja2-compatible template engine implementation in Go.

## Features

- **Jinja2 Compatibility**: Parse and render Jinja2 templates with familiar syntax
- **High Performance**: Optimized for speed with template compilation, caching, and AST node pooling
- **Memory Efficient**: Optional `Release()` method for returning pooled nodes to reduce GC pressure
- **Template Inheritance**: Full support for extends, blocks, and includes
- **Rich Filter System**: Built-in filters plus support for custom filters
- **Flexible Loading**: Load templates from filesystem, memory, or embedded resources
- **Security**: Auto-escaping enabled by default to prevent XSS
- **Developer Friendly**: Clear error messages with line numbers and debugging support

## Installation

```bash
go get github.com/zipreport/miya
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"

    "github.com/zipreport/miya"
)

func main() {
    env := miya.NewEnvironment()

    tmpl, err := env.FromString("Hello {{ name }}!")
    if err != nil {
        log.Fatal(err)
    }

    ctx := miya.NewContext()
    ctx.Set("name", "World")

    output, err := tmpl.Render(ctx)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(output) // Output: Hello World!
}
```

## Supported Features

### Core Features
- Variable substitution: `{{ variable }}`
- Control structures: `{% if %}`, `{% for %}`, etc.
- Template inheritance: `{% extends %}`, `{% block %}`
- Includes: `{% include %}`
- Macros: `{% macro %}`
- Filters: `{{ value|upper|trim }}`
- **List/Dict comprehensions**: `{{ [x*2 for x in items] }}`
- **Filter blocks**: `{% filter upper %}content{% endfilter %}`
- Do statements: `{% do expression %}`
- Comments: `{# comment #}`
- Whitespace control: `{%- -%}`
- Custom delimiters

### Built-in Filters
- String: `capitalize`, `lower`, `upper`, `title`, `trim`, `replace`
- HTML: `escape`, `safe`, `urlencode`
- Collections: `first`, `last`, `length`, `join`, `sort`, `reverse`
- Numbers: `abs`, `round`, `int`, `float`

## Advanced Examples

### List/Dict Comprehensions
Create collections with concise syntax:

```go
template := `
Active users: {{ [user.name for user in users if user.active] }}
User lookup: {{ {user.id: user.name for user in users} }}
Cart total: ${{ [item.price * item.qty for item in cart]|sum }}
`

ctx := miya.NewContext()
ctx.Set("users", []map[string]interface{}{
    {"id": 1, "name": "Alice", "active": true},
    {"id": 2, "name": "Bob", "active": false},
    {"id": 3, "name": "Charlie", "active": true},
})
ctx.Set("cart", []map[string]interface{}{
    {"price": 10.99, "qty": 2},
    {"price": 5.50, "qty": 1},
})

result, _ := env.RenderString(template, ctx)
// Output: Active users: [Alice, Charlie]
//         User lookup: {1: Alice, 2: Bob, 3: Charlie}  
//         Cart total: $27.48
```

### Filter Blocks
Apply filters to entire blocks of content:

```go
template := `{% filter upper %}
Hello {{ name }}!
Your items:
{% for item in items %}
- {{ item }}
{% endfor %}
{% endfilter %}`

ctx := miya.NewContext()
ctx.Set("name", "Alice")
ctx.Set("items", []string{"apple", "banana"})

result, _ := env.RenderString(template, ctx)
// Output: HELLO ALICE!\nYOUR ITEMS:\n- APPLE\n- BANANA
```

### Chained Filters
```go
template := `{% filter trim|upper|reverse %}   hello world   {% endfilter %}`
result, _ := env.RenderString(template, ctx)
// Output: DLROW OLLEH
```

### Do Statements
Execute expressions for side effects:

```go
template := `{% do complex_calculation %}Result: {{ result }}`
```

### Memory Management

Miya uses AST node pooling to reduce allocations during template parsing. For high-throughput applications, you can explicitly release pooled nodes when a template is no longer needed:

```go
tmpl, _ := env.FromString("Hello {{ name }}!")
output, _ := tmpl.Render(ctx)

// Optional: release pooled AST nodes for reuse
tmpl.Release()
```

**Note:** Calling `Release()` is optional. If not called, nodes will be garbage collected normally. Use it in scenarios where you're creating many short-lived templates and want to minimize GC pressure.

## Documentation

See [docs/](docs/) for detailed documentation.

## More Examples

Check out the [examples/](examples/) directory for comprehensive usage examples.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see [LICENSE](LICENSE) for details.
