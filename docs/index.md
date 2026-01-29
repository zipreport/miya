# Miya Engine

A high-performance Jinja2-compatible template engine for Go.

[![Go Reference](https://pkg.go.dev/badge/github.com/zipreport/miya.svg)](https://pkg.go.dev/github.com/zipreport/miya)
[![Go Report Card](https://goreportcard.com/badge/github.com/zipreport/miya)](https://goreportcard.com/report/github.com/zipreport/miya)
[![Test](https://github.com/zipreport/miya/actions/workflows/test.yml/badge.svg)](https://github.com/zipreport/miya/actions/workflows/test.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go)

## Features

- **Jinja2 Compatibility**: Parse and render Jinja2 templates with familiar syntax
- **High Performance**: Optimized for speed with template compilation, caching, and AST node pooling
- **Memory Efficient**: Optional `Release()` method for returning pooled nodes to reduce GC pressure
- **Template Inheritance**: Full support for extends, blocks, and includes
- **Rich Filter System**: Built-in filters plus support for custom filters
- **Flexible Loading**: Load templates from filesystem, memory, or embedded resources
- **Security**: Auto-escaping enabled by default to prevent XSS
- **Thread Safe**: Safe for concurrent use in web applications
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

## Supported Syntax

| Syntax | Description |
|--------|-------------|
| `{{ variable }}` | Variable substitution |
| `{% if %}...{% endif %}` | Conditional statements |
| `{% for %}...{% endfor %}` | Loop iteration |
| `{% extends %}` | Template inheritance |
| `{% block %}...{% endblock %}` | Override blocks |
| `{% include %}` | Include templates |
| `{% macro %}...{% endmacro %}` | Reusable macros |
| `{{ value\|filter }}` | Filter application |
| `{# comment #}` | Comments |
| `{%- -%}` / `{{- -}}` | Whitespace control |

## Documentation

### Core Features

- **[Template Inheritance](TEMPLATE_INHERITANCE.md)** - Learn about extends, blocks, and super() for template composition
- **[Control Structures](CONTROL_STRUCTURES_GUIDE.md)** - If/elif/else, for loops, set, and with statements
- **[Filters](FILTERS_GUIDE.md)** - 70+ built-in filters for data transformation
- **[Tests & Operators](TESTS_AND_OPERATORS.md)** - 26+ test functions and all operators

### Advanced Features

- **[Advanced Guide](ADVANCED_FEATURES_GUIDE.md)** - Filter blocks, whitespace control, autoescape
- **[Macros & Includes](MACROS_AND_INCLUDES.md)** - Reusable components and template composition
- **[Global Functions](GLOBAL_FUNCTIONS.md)** - range, dict, cycler, joiner, namespace
- **[Comprehensions](COMPREHENSIONS_GUIDE.md)** - List and dict comprehensions

### Reference

- **[Comprehensive Features](COMPREHENSIVE_FEATURES.md)** - Complete feature overview
- **[Jinja2 Compatibility](JINJA2_VS_MIYA_FEATURE_MATRIX.md)** - Feature comparison matrix
- **[Limitations](MIYA_LIMITATIONS.md)** - Known limitations and workarounds

## Why Miya?

| Feature | Miya | Other Go Template Engines |
|---------|------|---------------------------|
| Jinja2 Syntax | Full compatibility | Partial or none |
| Template Inheritance | Full support | Limited |
| Filters | 70+ built-in | Varies |
| List Comprehensions | Supported | Rare |
| Thread Safety | Yes | Varies |
| Performance | Optimized with pooling | Varies |

## License

MIT License - see [LICENSE](https://github.com/zipreport/miya/blob/master/LICENSE) for details.
