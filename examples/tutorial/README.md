# Miya Engine Tutorial

A step-by-step guide to using the Miya template engine, from basic rendering to advanced patterns.

## Overview

This tutorial consists of 6 progressive steps:

| Step | Topic | What You'll Learn |
|------|-------|-------------------|
| 1 | Hello World | Basic template rendering from strings |
| 2 | Variables & Filters | Passing data and transforming output |
| 3 | Control Flow | Conditionals and loops |
| 4 | Template Inheritance | Building reusable layouts |
| 5 | Macros & Components | Creating reusable template functions |
| 6 | Filesystem Loading | Loading templates from disk |

## Prerequisites

- Go 1.18 or later
- Miya Engine installed: `go get github.com/zipreport/miya`

## Running the Examples

Each step is self-contained in its own directory. Run any step from the project root:

```bash
go run ./examples/tutorial/step1_hello_world
go run ./examples/tutorial/step2_variables_filters
go run ./examples/tutorial/step3_control_flow
go run ./examples/tutorial/step4_inheritance
go run ./examples/tutorial/step5_macros
go run ./examples/tutorial/step6_filesystem
```

---

## Step 1: Hello World

**Directory:** `step1_hello_world/`

Learn the basics of creating an environment and rendering a simple template.

**Key concepts:**
- Creating a Miya environment
- Rendering templates from strings
- Basic variable substitution with `{{ }}`

**Example output:**
```
Hello, World!
Hello, Alice!
```

---

## Step 2: Variables & Filters

**Directory:** `step2_variables_filters/`

Learn how to pass complex data and transform it with filters.

**Key concepts:**
- Passing maps and structs to templates
- Accessing nested properties (`user.name`, `user.address.city`)
- Using filters (`|upper`, `|lower`, `|title`, `|default`)
- Chaining multiple filters (`|trim|upper`)

**Example output:**
```
User: ALICE JOHNSON
Email: alice@example.com
City: San Francisco
Status: Active (default applied)
```

---

## Step 3: Control Flow

**Directory:** `step3_control_flow/`

Learn conditional rendering and iteration.

**Key concepts:**
- If/elif/else statements
- For loops with `loop` variables (index, first, last)
- Inline conditionals (ternary): `{{ 'yes' if condition else 'no' }}`
- Empty list handling with `{% else %}` in for loops

**Example output:**
```
Premium user: Alice
Items: Apple, Banana, Cherry
Total: 3 items
```

---

## Step 4: Template Inheritance

**Directory:** `step4_inheritance/`

Learn how to create reusable layouts with blocks.

**Key concepts:**
- Base templates with `{% block %}` definitions
- Child templates with `{% extends %}`
- Overriding blocks
- Using `{{ super() }}` to include parent content

**Example output:**
```html
<!DOCTYPE html>
<html>
<head><title>My Page</title></head>
<body>
  <header>Site Header</header>
  <main>Welcome to my page!</main>
  <footer>Copyright 2024</footer>
</body>
</html>
```

---

## Step 5: Macros & Components

**Directory:** `step5_macros/`

Learn how to create reusable template components.

**Key concepts:**
- Defining macros with `{% macro %}`
- Macro parameters with defaults
- Importing macros with `{% import %}` and `{% from %}`
- Building a component library

**Example output:**
```html
<button class="btn btn-primary">Submit</button>
<button class="btn btn-secondary">Cancel</button>
<input type="email" name="email" class="form-control" required>
```

---

## Step 6: Filesystem Loading

**Directory:** `step6_filesystem/`

Learn how to load templates from the filesystem for real-world applications.

**Key concepts:**
- Using `FileSystemLoader` to load templates from disk
- Organizing templates in directories (layouts/, pages/, includes/)
- Combining inheritance, includes, and macros with file-based templates
- Template caching for performance
- Error handling for missing templates

**Template structure:**
```
templates/
├── layouts/
│   └── base.html           # Base layout template
├── pages/
│   ├── home.html           # Extends base.html
│   └── dashboard.html      # Uses includes
└── includes/
    ├── user_info.html      # User info partial
    ├── stats.html          # Statistics widget
    └── orders_table.html   # Orders table partial
```

**Example output:**
```html
<!DOCTYPE html>
<html>
<head><title>Dashboard - My Website</title></head>
<body>
  <header>...</header>
  <main>
    <div class="user-info">Charlie (admin)</div>
    <div class="stats">1250 users, 847 orders...</div>
    <table>Recent orders...</table>
  </main>
</body>
</html>
```

---

## Next Steps

After completing this tutorial:

1. **Explore more examples:** Check `examples/features/` for in-depth feature demos
2. **Read the documentation:** See `docs/` for complete feature reference
3. **Build something:** Try creating a simple web page with templates

## Quick Reference

### Template Syntax

```jinja2
{{ variable }}              {# Output a variable #}
{{ variable|filter }}       {# Apply a filter #}
{% if condition %}...{% endif %}
{% for item in list %}...{% endfor %}
{% block name %}...{% endblock %}
{% macro name(args) %}...{% endmacro %}
{# This is a comment #}
```

### Common Filters

| Filter | Example | Result |
|--------|---------|--------|
| `upper` | `{{ "hello"\|upper }}` | `HELLO` |
| `lower` | `{{ "HELLO"\|lower }}` | `hello` |
| `title` | `{{ "hello world"\|title }}` | `Hello World` |
| `trim` | `{{ "  hi  "\|trim }}` | `hi` |
| `default` | `{{ missing\|default("N/A") }}` | `N/A` |
| `length` | `{{ [1,2,3]\|length }}` | `3` |
| `join` | `{{ ["a","b"]\|join(", ") }}` | `a, b` |
| `first` | `{{ [1,2,3]\|first }}` | `1` |
| `last` | `{{ [1,2,3]\|last }}` | `3` |

### Loop Variables

| Variable | Description |
|----------|-------------|
| `loop.index` | Current iteration (1-indexed) |
| `loop.index0` | Current iteration (0-indexed) |
| `loop.first` | True if first iteration |
| `loop.last` | True if last iteration |
| `loop.length` | Total number of items |
