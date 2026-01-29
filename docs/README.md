# Miya Engine Documentation

Complete documentation for Miya Engine - a blazing-fast Jinja2-compatible template engine for Go.

---

##  Quick Start

1. **Installation**: `go get github.com/zipreport/miya`
2. **Basic Usage**: See main [README.md](../README.md) for quick start
3. **Working Examples**: Check `examples/features/` for complete working demos

---

##  Core Guides

### Essential Features

| Guide | Description | Compatibility |
|-------|-------------|---------------|
| **[Template Inheritance](TEMPLATE_INHERITANCE.md)** | Extends, blocks, super() |  100% Jinja2 |
| **[Control Structures](CONTROL_STRUCTURES_GUIDE.md)** | If/elif/else, for loops, set, with |  100% Jinja2 |
| **[Filters](FILTERS_GUIDE.md)** | 70+ filters for data transformation |  100% Jinja2 |
| **[Tests & Operators](TESTS_AND_OPERATORS.md)** | 26+ tests, all operators |  100% Jinja2 |

### Advanced Features

| Guide | Description | Compatibility |
|-------|-------------|---------------|
| **[Macros & Includes](MACROS_AND_INCLUDES.md)** | Reusable components, template composition |  95% (no caller()) |
| **[Global Functions](GLOBAL_FUNCTIONS.md)** | range, dict, cycler, joiner, namespace, etc. |  100% Jinja2 |
| **[Comprehensions](COMPREHENSIONS_GUIDE.md)** | List and dict comprehensions |  70% (no inline if) |
| **[Advanced Features](ADVANCED_FEATURES_GUIDE.md)** | Filter blocks, whitespace control, autoescape, performance |  100% Jinja2 |

### Reference

| Document | Description |
|----------|-------------|
| **[Miya Limitations](MIYA_LIMITATIONS.md)** | Complete list of unsupported features and workarounds |
| **[Feature Matrix](JINJA2_VS_MIYA_FEATURE_MATRIX.md)** | Jinja2 vs Miya compatibility comparison |
| **[Comprehensive Features](COMPREHENSIVE_FEATURES.md)** | Complete feature overview |

---

##  Feature Documentation

### 1. Template Inheritance

Learn how to build reusable template hierarchies:

```jinja2
{# base.html #}
<!DOCTYPE html>
<html>
<head>
  <title>{% block title %}Default{% endblock %}</title>
</head>
<body>
  {% block content %}{% endblock %}
</body>
</html>

{# page.html #}
{% extends "base.html" %}
{% block title %}My Page{% endblock %}
{% block content %}
  {{ super() }}  {# Include parent content #}
  <h1>Hello World!</h1>
{% endblock %}
```

 **[Read Full Guide: Template Inheritance](TEMPLATE_INHERITANCE.md)**

**Topics Covered:**
- Extends and blocks
- Block overriding
- super() for parent content
- Multi-level inheritance
- Block scoping

---

### 2. Control Structures

Master template control flow:

```jinja2
{# Conditionals #}
{% if user.role == "admin" %}
  <p>Admin Panel</p>
{% elif user.role == "moderator" %}
  <p>Moderator Tools</p>
{% else %}
  <p>User Dashboard</p>
{% endif %}

{# Loops #}
{% for product in products if product.stock > 0 %}
  <div>{{ product.name }} - ${{ product.price }}</div>
{% endfor %}

{# Variable scoping #}
{% with total = items|length %}
  <p>{{ total }} items in cart</p>
{% endwith %}
```

 **[Read Full Guide: Control Structures](CONTROL_STRUCTURES_GUIDE.md)**

**Topics Covered:**
- If/elif/else statements
- For loops with loop variables
- Inline conditionals (ternary)
- Set statements
- With statements for scoping
- Complex control flow

---

### 3. Filters

Transform data with 70+ built-in filters:

```jinja2
{# String filters #}
{{ "hello world"|title }}              → Hello World
{{ "  text  "|trim|upper }}            → TEXT

{# Collection filters #}
{{ numbers|first }}                    → 1
{{ items|join(", ") }}                 → item1, item2, item3

{# Numeric filters #}
{{ 3.14159|round(2) }}                 → 3.14
{{ -42|abs }}                          → 42

{# Security filters #}
{{ user_input|escape }}                → &lt;script&gt;...
{{ trusted_html|safe }}                → <strong>Bold</strong>
```

 **[Read Full Guide: Filters](FILTERS_GUIDE.md)**

**Topics Covered:**
- String filters (upper, lower, trim, replace, etc.)
- Collection filters (first, last, join, map, etc.)
- Numeric filters (abs, round, int, float)
- HTML/Security filters (escape, safe, striptags)
- Filter chaining
- Known limitations

---

### 4. Tests & Operators

Powerful conditional expressions:

```jinja2
{# Tests #}
{% if user is defined and user.active %}
  Welcome!
{% endif %}

{% if value is number and value > 100 %}
  Large number
{% endif %}

{# Operators #}
{{ 5 + 3 }}                            → 8
{{ "Hello" ~ " " ~ "World" }}          → Hello World
{{ 4 is even }}                        → true
{{ "admin" in user.roles }}            → true
```

 **[Read Full Guide: Tests & Operators](TESTS_AND_OPERATORS.md)**

**Topics Covered:**
- Arithmetic operators (+, -, *, /, //, %, **)
- Comparison operators (==, !=, <, >, <=, >=)
- Logical operators (and, or, not)
- Membership operators (in, not in)
- 26+ test expressions
- Type tests, container tests, numeric tests

---

### 5. Macros & Includes

Build reusable components:

```jinja2
{# Define macro #}
{% macro render_input(name, type="text") %}
<input type="{{ type }}" name="{{ name }}" class="form-control">
{% endmacro %}

{# Use macro #}
{% import "forms.html" as forms %}
{{ forms.render_input("email", type="email") }}

{# Include templates #}
{% include "header.html" %}
<main>Content here</main>
{% include "footer.html" %}
```

 **[Read Full Guide: Macros & Includes](MACROS_AND_INCLUDES.md)**

**Topics Covered:**
- Macro definitions and parameters
- Import with namespace
- Selective imports (from)
- Template includes
- Practical component libraries
- Limitation: caller() not supported

---

### 6. Global Functions

Essential template utilities:

```jinja2
{# range() - number sequences #}
{% for i in range(10) %}{{ i }}{% endfor %}

{# cycler() - alternate values #}
{% set row_color = cycler("odd", "even") %}
{% for item in items %}
  <tr class="{{ row_color.next() }}">{{ item }}</tr>
{% endfor %}

{# namespace() - mutable counter #}
{% set ns = namespace(count=0) %}
{% for item in items %}
  {% set ns.count = ns.count + 1 %}
{% endfor %}
Total: {{ ns.count }}

{# zip() - parallel iteration #}
{% for name, age in zip(names, ages) %}
  {{ name }}: {{ age }}
{% endfor %}
```

 **[Read Full Guide: Global Functions](GLOBAL_FUNCTIONS.md)**

**Topics Covered:**
- range() - number sequences
- dict() - dictionary construction
- cycler() - cycle through values
- joiner() - smart separators
- namespace() - mutable containers
- lipsum() - lorem ipsum generator
- zip() - combine sequences
- enumerate() - index with values
- url_for() - URL generation

---

### 7. Comprehensions

Concise collection creation:

```jinja2
{# List comprehensions #}
{{ [x * 2 for x in numbers] }}
{{ [user.name for user in users] }}
{{ [name|upper for name in names] }}

{# Dictionary comprehensions #}
{{ {user.id: user.name for user in users} }}
{{ {product.sku: product.price for product in products} }}

{# Note: Inline if not supported - use filters instead #}
{{ users|selectattr("active")|map(attribute="name")|list }}
```

 **[Read Full Guide: Comprehensions](COMPREHENSIONS_GUIDE.md)**

**Topics Covered:**
- Basic list comprehensions
- Dictionary comprehensions
- Using filters in comprehensions
- Limitations (no inline if, no .items() unpacking)
- Workarounds with selectattr/select filters

---

### 8. Advanced Features

Professional template control:

```jinja2
{# Filter blocks #}
{% filter upper %}
  This entire block becomes uppercase
{% endfilter %}

{# Whitespace control #}
<ul>
{%- for item in items %}
  <li>{{ item }}</li>
{%- endfor %}
</ul>

{# Raw blocks for documentation #}
{% raw %}
  Template: {{ variable }}
{% endraw %}

{# Autoescape control #}
{% autoescape false %}
  {{ trusted_html }}
{% endautoescape %}
```

 **[Read Full Guide: Advanced Features](ADVANCED_FEATURES_GUIDE.md)**

**Topics Covered:**
- Filter blocks
- Do statements
- Whitespace control with `-`
- Raw blocks
- Autoescape control
- Environment configuration

---

##  Quick Reference

### Most Common Operations

```jinja2
{# Variables #}
{{ variable }}
{{ user.name }}
{{ items[0] }}

{# Conditionals #}
{% if condition %}...{% endif %}
{{ 'yes' if condition else 'no' }}

{# Loops #}
{% for item in items %}{{ item }}{% endfor %}
{% for key, value in dict.items() %}...{% endfor %}

{# Filters #}
{{ text|upper }}
{{ items|join(", ") }}
{{ value|default("N/A") }}

{# Tests #}
{% if value is defined %}...{% endif %}
{% if number is even %}...{% endif %}

{# Inheritance #}
{% extends "base.html" %}
{% block content %}...{% endblock %}

{# Includes #}
{% include "header.html" %}

{# Macros #}
{% import "forms.html" as forms %}
{{ forms.input("name") }}
```

---

##  Important Limitations

Miya Engine is ~95% compatible with Jinja2. Key differences:

| Feature | Jinja2 | Miya Engine | Workaround |
|---------|--------|-------------|------------|
| List comprehension with if |  Supported |  Not supported | Use `selectattr` filter |
| Dict comprehension with .items() |  Supported |  Not supported | Loop over list instead |
| Inline conditional without else |  Supported |  Not supported | Always include `else` |
| Macro caller() |  Supported |  Not supported | Pass content as parameter |
| Nested comprehensions |  Supported |  Not supported | Use nested loops |

 **[Read Full Details: Miya Limitations](MIYA_LIMITATIONS.md)**

---

##  Documentation Structure

```
docs/
├── README.md                          # This file - documentation index
│
├── Core Feature Guides/
│   ├── TEMPLATE_INHERITANCE.md        # Extends, blocks, super()
│   ├── CONTROL_STRUCTURES_GUIDE.md    # If/for/set/with statements
│   ├── FILTERS_GUIDE.md               # 70+ built-in filters
│   ├── TESTS_AND_OPERATORS.md         # 26+ tests, all operators
│   ├── MACROS_AND_INCLUDES.md         # Reusable components
│   ├── GLOBAL_FUNCTIONS.md            # range, cycler, namespace, etc.
│   ├── COMPREHENSIONS_GUIDE.md        # List/dict comprehensions
│   └── ADVANCED_FEATURES_GUIDE.md     # Filter blocks, whitespace, etc.
│
├── Reference Documentation/
│   ├── MIYA_LIMITATIONS.md            # Complete limitations list
│   ├── JINJA2_VS_MIYA_FEATURE_MATRIX.md  # Compatibility matrix
│   ├── COMPREHENSIVE_FEATURES.md      # Complete feature overview
│   ├── FILTER_BLOCKS_IMPLEMENTATION.md   # Filter blocks details
│   └── DO_STATEMENTS_IMPLEMENTATION.md   # Do statement details
│
└── Legacy Docs/ (may be outdated)
    ├── LIST_DICT_COMPREHENSIONS.md    # Old comprehensions doc
    ├── comprehensions_examples.md
    ├── comprehensions_reference.md
    ├── filters/
    ├── flow_control/
    ├── inheritance/
    └── tests/
```

---

##  Learning Path

### Beginner

1. Start with [Template Inheritance](TEMPLATE_INHERITANCE.md)
2. Learn [Control Structures](CONTROL_STRUCTURES_GUIDE.md)
3. Master [Filters](FILTERS_GUIDE.md)
4. Explore [Working Examples](https://github.com/zipreport/miya/tree/master/examples/features/)

### Intermediate

1. Study [Tests & Operators](TESTS_AND_OPERATORS.md)
2. Build components with [Macros & Includes](MACROS_AND_INCLUDES.md)
3. Use [Global Functions](GLOBAL_FUNCTIONS.md)
4. Review [Miya Limitations](MIYA_LIMITATIONS.md)

### Advanced

1. Master [Comprehensions](COMPREHENSIONS_GUIDE.md)
2. Learn [Advanced Features](ADVANCED_FEATURES_GUIDE.md)
3. Study [Feature Matrix](JINJA2_VS_MIYA_FEATURE_MATRIX.md)
4. Build production templates with all features

---

##  External Resources

- **Main Repository**: [github.com/zipreport/miya](https://github.com/zipreport/miya)
- **Working Examples**: See `examples/features/` directory
- **Jinja2 Documentation**: [jinja.palletsprojects.com](https://jinja.palletsprojects.com)
- **Go Package**: [pkg.go.dev/github.com/zipreport/miya](https://pkg.go.dev/github.com/zipreport/miya)

---

##  Contributing

When adding new documentation:

1. Follow the existing structure and naming conventions
2. Include practical examples alongside syntax explanations
3. Add cross-references to related features
4. Update this index file with new documentation
5. Test all code examples
6. Document any limitations or differences from Jinja2

---

##  Support

- **Issues**: Report bugs on GitHub
- **Examples**: Check `examples/` for working code
- **Tests**: Review test files for usage patterns
- **Questions**: Open a GitHub issue

---

##  Compatibility Summary

**Miya Engine: ~95% Jinja2 Compatible**

 **100% Compatible:**
- Template inheritance (extends, blocks, super)
- Control structures (if/for/set/with)
- All filters (70+)
- All tests (26+)
- All operators
- Global functions (9)
- Advanced features (filter blocks, whitespace, etc.)
- Performance optimizations (AST node pooling, evaluator pooling, caching)

 **Partial Compatibility:**
- Macros (95% - no caller())
- Comprehensions (70% - no inline if, no nested)

 **Not Supported:**
- Import from parent scope
- Async/await
- Some advanced Jinja2 extensions

See [MIYA_LIMITATIONS.md](MIYA_LIMITATIONS.md) for complete details.

---

**Last Updated**: 2024
**Version**: Based on Miya Engine latest
**Jinja2 Version**: Compatible with Jinja2 3.x syntax
