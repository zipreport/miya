# Python Jinja2 vs Miya Engine Implementation Differences

## Overview

This document outlines the key differences between Python Jinja2 and the Miya Engine (miya) implementation based on practical testing and implementation of a comprehensive feature showcase.

## Template Engine Comparison

### Python Jinja2
- **Library**: `jinja2` (official Jinja2 implementation)
- **Version**: Latest stable (3.x)
- **Language**: Python
- **Maturity**: Fully mature, extensive ecosystem

### Miya Engine (miya)
- **Library**: `github.com/zipreport/miya`
- **Language**: Go
- **Maturity**: Good core functionality, some limitations
- **Performance**: Potentially faster due to Go's compiled nature

## Core Feature Compatibility

| Feature | Python Jinja2 | Miya Engine | Status | Notes |
|---------|---------------|-----------|---------|-------|
| Variable expressions |  Full |  Full |  Compatible | Same syntax |
| Filters (built-in) |  60+ filters |  ~20 filters |  Partial | Many missing, require custom implementation |
| Control structures |  Full |  Full |  Compatible | if/elif/else, for loops work identically |
| Macros |  Full |  Full |  Compatible | Function definitions and calls work |
| Template inheritance |  Full |  Full |  Compatible | extends/blocks supported |
| Comments |  Full |  Full |  Compatible | {# comment #} syntax |
| Raw blocks |  Full |  Full |  Compatible | {% raw %} blocks work |
| Auto-escaping |  Full |  Full |  Compatible | XSS protection works |
| Tests |  20+ tests |  ~10 tests |  Partial | Basic tests available |
| With statements |  Full |  Full |  Compatible | Context scoping works |

## Major Differences

### 1. Dictionary/Map Iteration

**Python Jinja2:**
```jinja2
{% for key, value in user.preferences.items() %}
    <li><strong>{{ key }}:</strong> {{ value }}</li>
{% endfor %}
```

**Miya Engine:**
```jinja2
<!-- Direct property access required -->
<li><strong>theme:</strong> {{ user.preferences.theme }}</li>
<li><strong>language:</strong> {{ user.preferences.language }}</li>

<!-- OR loop unpacking not supported -->
{% for item in user.preferences %}
    <!-- Only gets values, not key-value pairs -->
{% endfor %}
```

**Impact**: Medium - Requires template restructuring for dictionary iteration

### 2. Date/Time Formatting

**Python Jinja2:**
```jinja2
{{ current_date.strftime("%B %d, %Y") }}
{{ current_date|strftime("%B %d, %Y") }}
```

**Miya Engine:**
```go
// Pre-process in Go code
ctx.Set("formatted_date", now.Format("January 02, 2006"))
```

```jinja2
{{ formatted_date }}
```

**Impact**: Medium - Method calls on objects not supported, requires pre-processing

### 3. Built-in Filters

**Python Jinja2 (Built-in):**
- `truncate`, `wordcount`, `center`, `reverse`, `sum`, `unique`, `urlencode`
- `abs`, `round`, `capitalize`, `random`, `slice`
- 60+ total filters available

**Miya Engine (Built-in):**
- Basic filters: `upper`, `lower`, `title`, `length`, `default`, `join`, `safe`, `escape`
- ~20 total built-in filters

**Custom Implementation Required:**
```go
env.AddFilter("truncate", func(value interface{}, args ...interface{}) (interface{}, error) {
    // Custom implementation needed
})
```

**Impact**: High - Significant development overhead for missing filters

### 4. Built-in Tests

**Python Jinja2:**
```jinja2
{{ 'defined' if var is defined else 'undefined' }}
{{ 'even' if number is even else 'odd' }}
{{ 'string' if value is string else 'not string' }}
{{ 'in list' if item in list else 'not in list' }}
```

**Miya Engine:**
```jinja2
{{ 'defined' if var is defined else 'undefined' }}  <!--  Works -->
{{ 'even' if number is even else 'odd' }}           <!--  Works -->
{{ 'string' if value is string else 'not string' }} <!--  Works -->
{{ 'in list' if item in list else 'not in list' }}  <!--  Works -->
```

**Impact**: Low - Core tests available

### 5. Object Method Calls

**Python Jinja2:**
```jinja2
{{ string_var.upper() }}
{{ list_var.append(item) }}
{{ date_obj.strftime("%Y") }}
```

**Miya Engine:**
```jinja2
{{ string_var | upper }}  <!-- Must use filter syntax -->
<!-- Method calls not supported -->
```

**Impact**: Medium - Requires filter approach instead of method calls

### 6. Advanced Loop Features

**Python Jinja2:**
```jinja2
{% for key, value in dict.items() %}           <!--  Key-value unpacking -->
{% for item in list recursive %}               <!--  Recursive loops -->
{% for item in list if item.condition %}       <!--  Conditional loops -->
```

**Miya Engine:**
```jinja2
<!-- No key-value unpacking -->
{% for item in list %}                         <!--  Basic loops -->
    {% if item.condition %}                    <!--  Manual conditions -->
        {{ item }}
    {% endif %}
{% endfor %}
```

**Impact**: Medium - More verbose template code required

### 7. Template Loading and Inheritance

**Python Jinja2:**
```python
from jinja2 import Environment, FileSystemLoader

env = Environment(loader=FileSystemLoader('templates'))
template = env.get_template('base.html')
```

**Miya Engine:**
```go
env := miya.NewEnvironment()
// File system loading supported but different API
template, err := env.FromString(templateContent)
```

**Impact**: Low - Different API but similar functionality

### 8. Error Handling

**Python Jinja2:**
```python
# Rich error messages with line numbers
# Template debugging tools
# Undefined variable handling options
```

**Miya Engine:**
```go
// Basic error messages
// Limited debugging information
// Undefined handling available
```

**Impact**: Medium - Less detailed error information

## Performance Considerations

### Python Jinja2
- Interpreted language overhead
- Rich feature set may impact performance
- Extensive caching mechanisms
- Mature optimization

### Miya Engine
- Compiled language performance benefits
- Simpler implementation may be faster
- Less feature overhead
- Limited optimization maturity

## Migration Strategy

### 1. Template Audit
- Identify dictionary iterations that need restructuring
- List custom filters/tests that need implementation
- Check for method calls requiring conversion

### 2. Data Pre-processing
- Format dates/times in Go code before template rendering
- Convert dictionary structures for easier template access
- Prepare complex data structures

### 3. Custom Filter Implementation
```go
func addMissingFilters(env *miya.Environment) {
    // Implement required Python filters
    env.AddFilter("truncate", truncateFilter)
    env.AddFilter("wordcount", wordcountFilter)
    // ... etc
}
```

### 4. Template Adaptation
- Replace `dict.items()` with direct property access
- Convert method calls to filter syntax
- Simplify complex loop structures

## Recommendations

### Use Python Jinja2 When:
- Maximum template feature compatibility required
- Complex dictionary operations needed
- Extensive built-in filter usage
- Rich error handling and debugging needed
- Existing Python ecosystem integration

### Use Miya Engine When:
- Performance is critical
- Go ecosystem integration preferred
- Simpler template requirements
- Willing to implement missing features
- Memory efficiency important

## Feature Implementation Status

###  Fully Compatible
- Basic variable expressions
- Control structures (if/for)
- Macros and functions
- Template inheritance
- Comments and raw blocks
- Auto-escaping
- Basic filters and tests

###  Partially Compatible
- Built-in filters (requires custom implementation)
- Dictionary iteration (no key-value unpacking)
- Date/time formatting (no method calls)
- Advanced loop features
- Error handling and debugging

###  Not Supported
- Object method calls in templates
- Dynamic template loading from strings with full Python compatibility
- Some advanced Python-specific features

## Conclusion

Miya Engine provides excellent core Jinja2 functionality with performance benefits, but requires additional development work for full Python Jinja2 compatibility. The choice depends on specific requirements for features versus performance and the willingness to implement missing functionality.

For most web applications, Miya Engine can provide 80-90% of Python Jinja2 functionality with better performance characteristics, making it a viable alternative when the trade-offs are acceptable.