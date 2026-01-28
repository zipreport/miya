# CLAUDE.md - Miya Engine Development Guide

This document provides essential context for Claude Code instances working in this repository.

## Project Overview

**Miya Engine** is a high-performance Jinja2-compatible template engine written in Go. It provides full compatibility with Python's Jinja2 templating syntax while leveraging Go's performance characteristics.

- **Module**: `github.com/zipreport/miya`
- **Go Version**: 1.24.12+
- **License**: MIT

## Build & Test Commands

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests for a specific package
go test ./filters
go test ./parser
go test ./runtime

# Run tests with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Build the package
go build ./...

# Format code (run before commits)
gofmt -w .

# Static analysis
go vet ./...

# Run benchmarks
go test -bench=. ./...
```

## Architecture Overview

### Template Processing Pipeline

```
Template Source → Lexer → Parser → AST → Runtime Evaluator → Output
```

1. **Lexer** (`lexer/`): Tokenizes template source into tokens (text, variables, blocks, comments)
2. **Parser** (`parser/`): Builds an Abstract Syntax Tree (AST) from tokens
3. **Runtime** (`runtime/`): Evaluates the AST with context data to produce output

### Core Components

| Package | Purpose |
|---------|---------|
| `miya` (root) | Main API: `Environment`, `Template`, `Context` |
| `lexer/` | Tokenization of template source |
| `parser/` | AST construction from tokens |
| `runtime/` | Template evaluation, inheritance processing |
| `filters/` | Built-in filter functions (100+ filters) |
| `branching/` | Test functions for conditionals (`is`, `is not`) |
| `loader/` | Template loading from filesystem/memory |
| `inheritance/` | Template inheritance (`{% extends %}`, `{% block %}`) |
| `macros/` | Macro definition and invocation |
| `extensions/` | Custom tag/extension system |
| `whitespace/` | Whitespace control processing |

### Key Types

```go
// Environment - configured template compilation environment
env := miya.NewEnvironment(
    miya.WithLoader(loader),
    miya.WithAutoEscape(true),
    miya.WithTrimBlocks(true),
)

// Template - compiled template ready for rendering
tmpl, err := env.FromString("Hello {{ name }}!")
tmpl, err := env.GetTemplate("page.html")

// Context - variable data for template rendering
ctx := miya.NewContext()
ctx.Set("name", "World")

// Render
output, err := tmpl.Render(ctx)
```

## Code Patterns & Conventions

### Performance Optimizations

1. **Pre-compiled regex patterns**: Define regex at package level, not in functions
   ```go
   var rePattern = regexp.MustCompile(`pattern`)  // Good

   func process() {
       re := regexp.MustCompile(`pattern`)  // Bad - recompiles each call
   }
   ```

2. **sync.Pool for evaluators**: The environment pools evaluators for reuse

3. **sync.Pool for AST nodes**: The parser pools frequently allocated node types to reduce GC pressure. See `parser/node_pool.go` for pooled types:
   - `LiteralNode`, `IdentifierNode`, `BinaryOpNode`, `FilterNode`
   - `UnaryOpNode`, `AttributeNode`, `GetItemNode`, `CallNode`
   - Use `AcquireXXXNode()` in parser, `Template.Release()` to return nodes to pools

4. **Template caching**: Templates are cached by name/content hash

5. **Lazy initialization**: Inheritance processor initializes on first use

### Error Handling

- Use `fmt.Errorf` with `%w` for error wrapping
- Return descriptive errors with context (template name, line number when available)
- Filters return `(interface{}, error)` - always check errors

### Panic Recovery

- Concurrent operations use panic recovery to prevent cascade failures
- Reflection-based method calls use `safeMethodCall()` with recovery
- Worker pools recover from panics in individual render jobs

### Thread Safety

- `Environment.cache` protected by `sync.RWMutex`
- `ConcurrentTemplateRenderer` for parallel rendering
- `sync.Map` used for high-contention concurrent access

## Test Organization

```
tests/
├── unit/           # Unit tests for individual components
├── integration/    # Full template rendering tests
├── jinja/          # Jinja2 compatibility tests
├── performance/    # Benchmark tests
└── fixtures/       # Test template files
```

### Running Specific Test Categories

```bash
# Unit tests only
go test ./tests/unit/...

# Integration tests
go test ./tests/integration/...

# Jinja compatibility tests
go test ./tests/jinja/...
```

## Common Development Tasks

### Adding a New Filter

1. Add filter function in `filters/` (appropriate category file)
2. Register in `filters/registry.go` `NewRegistry()` function
3. Add tests in `tests/unit/` or relevant test file

```go
// In filters/string_filters.go
func MyFilter(value interface{}, args ...interface{}) (interface{}, error) {
    s := ToString(value)
    // ... filter logic
    return result, nil
}

// In filters/registry.go NewRegistry()
r.Register("myfilter", MyFilter)
```

### Adding a New Test Function

1. Add test function in `branching/tests.go`
2. Register in `branching/registry.go`
3. Use with `{% if value is mytest %}` syntax

### Adding Template Tags via Extensions

See `extensions/` package for the extension API. Extensions can add:
- Custom tags (`{% mytag %}`)
- Custom filters
- Custom tests

## File Naming Conventions

- `*_test.go` - Test files
- `*_bench_test.go` - Benchmark files
- Snake_case for multi-word files (e.g., `string_filters.go`)

## Important Implementation Notes

### Map Iteration Order

Go maps have non-deterministic iteration order. When testing template output involving maps, check for multiple valid orderings or sort keys first.

### Undefined Variables

Three undefined behaviors available:
- `UndefinedSilent` (default): Returns empty string
- `UndefinedStrict`: Returns error
- `UndefinedDebug`: Returns debug placeholder

### Auto-escaping

HTML auto-escaping is enabled by default. Use `| safe` filter or `SafeValue` type to bypass.

### Whitespace Control

- `{%-` / `-%}` - Strip whitespace around blocks
- `{{-` / `-}}` - Strip whitespace around variables
- `trimBlocks` / `lstripBlocks` environment options

## Dependencies

This project has no external dependencies beyond the Go standard library.

## Quick Reference: Jinja2 Syntax

```jinja
{# Comments #}
{{ variable }}
{{ variable | filter }}
{{ variable | filter(arg1, arg2) }}

{% if condition %}...{% elif %}...{% else %}...{% endif %}
{% for item in items %}...{% else %}...{% endfor %}
{% set variable = value %}
{% include "template.html" %}
{% extends "base.html" %}
{% block name %}...{% endblock %}
{% macro name(args) %}...{% endmacro %}
{% call macro() %}...{% endcall %}
{% with variable = value %}...{% endwith %}
{% do expression %}
```
