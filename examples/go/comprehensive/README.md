# Miya Engine Comprehensive Example

This directory contains a complete implementation of a comprehensive Jinja2 template showcase converted from Python to Go.

## Files

- **`comprehensive_example.go`** - Complete Go implementation demonstrating 17+ Jinja2 features  
- **`rendered_output_go.html`** - Generated HTML output (17,080 bytes)
- **`PYTHON_VS_GO_JINJA2.md`** - Detailed comparison between Python and Go implementations

## Features Demonstrated

1. **Variable Expressions & Filters** - Basic variables, filters, defaults
2. **Conditional Statements** - if/elif/else logic  
3. **Loops** - List iteration, loop variables, nested loops, dictionary key-value unpacking
4. **Template Inheritance & Blocks** - Block definitions for template extension
5. **Macros** - Reusable template functions
6. **Variable Assignments** - Dynamic variable creation
7. **Raw Content** - Escaped template syntax
8. **Comments** - Template documentation
9. **Tests** - Boolean checks (defined, number, string, even/odd, divisibleby, etc.)
10. **Whitespace Control** - Strip whitespace with {%- -%} syntax
11. **Complex Expressions** - Math, string operations, comparisons, list slicing
12. **Built-in Filters (70+ Available!)** - Math, string, collection, dict, HTML, date/time, utility filters
13. **Auto-escaping** - XSS protection with autoescape control
14. **Namespace** - Variable scoping across loop iterations  
15. **Call Blocks** - Advanced macro calls with content blocks
16. **With Statements** - Context variable scoping
17. **Complex Data Structures** - Nested objects and arrays

## Running the Example

```bash
go run comprehensive_example.go
```

This will generate `rendered_output_go.html` with the complete showcase (now 17,080 bytes).

## Key Adaptations for Miya Engine

- **Date formatting**: Pre-processed in Go instead of template method calls
- **Dictionary iteration**: Now supports full key-value unpacking like Python
- **Advanced features**: Added whitespace control, namespace, call blocks, enhanced tests
- **Built-in filters**: Discovered 70+ built-in filters including advanced math, collections, HTML, date/time processing
- **Native syntax**: List slicing with [:] syntax, range() function, divisibleby() test
- **Context management**: Proper Go interface usage

## Comparison with Python

See `PYTHON_VS_GO_JINJA2.md` for detailed differences, migration strategies, and compatibility information.

## Output

The generated HTML demonstrates all major Jinja2 features working correctly in Go, producing a comprehensive feature showcase that includes 17+ features with 70+ built-in filters compared to the Python version's 20 features. The Go implementation provides excellent compatibility with Python Jinja2, including advanced features like whitespace control, namespace scoping, call blocks, enhanced template tests, and a comprehensive filter library that rivals Python's capabilities.