# Macros and Includes Guide

Macros enable template reusability through parameterized components. Includes allow template composition. Miya Engine provides Jinja2-compatible macros and includes with one limitation.

> **Working Example:** See `examples/features/macros-includes/` for complete examples.

---

## Table of Contents

1. [Macros Basics](#macros-basics)
2. [Macro Parameters](#macro-parameters)
3. [Importing Macros](#importing-macros)
4. [Template Includes](#template-includes)
5. [Practical Examples](#practical-examples)
6. [Limitations](#limitations)

---

## Macros Basics

### What are Macros?

Macros are reusable template functions that accept parameters and return rendered content.

### Basic Macro Definition

**macros.html:**
```html+jinja
{% macro render_input(name, type="text", placeholder="") %}
<div class="form-group">
  <input type="{{ type }}"
         name="{{ name }}"
         placeholder="{{ placeholder }}"
         class="form-control">
</div>
{% endmacro %}
```

### Using a Macro

```html+jinja
{% import "macros.html" as forms %}

{{ forms.render_input("username", placeholder="Enter username") }}
{{ forms.render_input("email", type="email") }}
{{ forms.render_input("password", type="password") }}
```

**Output:**
```html
<div class="form-group">
  <input type="text" name="username" placeholder="Enter username" class="form-control">
</div>
<div class="form-group">
  <input type="email" name="email" placeholder="" class="form-control">
</div>
<div class="form-group">
  <input type="password" name="password" placeholder="" class="form-control">
</div>
```

---

## Macro Parameters

### Default Parameters

```html+jinja
{% macro render_button(text, type="button", class="btn-primary", disabled=False) %}
<button type="{{ type }}"
        class="btn {{ class }}"
        {{ 'disabled' if disabled else '' }}>
  {{ text }}
</button>
{% endmacro %}
```

**Usage:**
```html+jinja
{{ forms.render_button("Click Me") }}
→ <button type="button" class="btn btn-primary">Click Me</button>

{{ forms.render_button("Save", type="submit", class="btn-success") }}
→ <button type="submit" class="btn btn-success">Save</button>

{{ forms.render_button("Disabled", disabled=True) }}
→ <button type="button" class="btn btn-primary" disabled>Disabled</button>
```

### Required Parameters

Parameters without defaults are required:

```html+jinja
{% macro render_field(name, label, type="text", required=False, help_text="") %}
<div class="form-field">
  <label>
    {{ label }}
    {% if required %}<span class="required">*</span>{% endif %}
  </label>
  <input type="{{ type }}" name="{{ name }}" class="form-control">
  {% if help_text %}
  <small class="help-text">{{ help_text }}</small>
  {% endif %}
</div>
{% endmacro %}
```

**Usage:**
```html+jinja
{{ forms.render_field("full_name", label="Full Name", required=True) }}
{{ forms.render_field("age", label="Age", type="number", help_text="Must be 18 or older") }}
```

### Variable Arguments

Access all macro variables:

```html+jinja
{% macro render_table_row(cells, header=False) %}
<tr>
  {% for cell in cells %}
    {% if header %}
    <th>{{ cell }}</th>
    {% else %}
    <td>{{ cell }}</td>
    {% endif %}
  {% endfor %}
</tr>
{% endmacro %}
```

---

## Importing Macros

### Method 1: Import with Namespace

Import all macros from a file into a namespace:

```html+jinja
{% import "macros.html" as forms %}

{{ forms.render_input("username") }}
{{ forms.render_button("Submit") }}
{{ forms.render_field("email", label="Email") }}
```

**Advantages:**
- Clean namespace separation
- No naming conflicts
- Clear source of macros

### Method 2: From (Selective Import)

Import specific macros directly:

```html+jinja
{% from "macros.html" import render_badge, render_list %}

{{ render_badge("New", type="info") }}
{{ render_list(items) }}
```

**Advantages:**
- Shorter syntax for frequently used macros
- Import only what you need
- Direct access without namespace prefix

### Method 3: Import Multiple with Aliases

```html+jinja
{% from "macros.html" import render_input as input, render_button as btn %}

{{ input("username") }}
{{ btn("Submit") }}
```

---

## Template Includes

### Basic Include

Include templates to compose pages:

```html+jinja
{% include "includes/header.html" %}

<h1>Main Content</h1>

{% include "includes/footer.html" %}
```

### Context Inheritance

Included templates have access to the current context:

**includes/footer.html:**
```html+jinja
<footer>
  <p>&copy; {{ year|default(2024) }} {{ company|default("Miya Engine") }}</p>
  <p>
    {% for link in footer_links %}
    <a href="{{ link.url }}">{{ link.title }}</a>{% if not loop.last %} | {% endif %}
    {% endfor %}
  </p>
</footer>
```

**Usage:**
```html+jinja
{# Variables are accessible in included template #}
{% set year = 2024 %}
{% set company = "My Company" %}
{% set footer_links = [...] %}

{% include "includes/footer.html" %}
```

### Include with Context Control

**Default behavior:** Full context access
```html+jinja
{% include "template.html" %}
```

**With only specific variables:**
```html+jinja
{% include "template.html" with context %}  {# explicit, same as default #}
```

### Conditional Includes

```html+jinja
{% if user.role == "admin" %}
  {% include "admin/navbar.html" %}
{% else %}
  {% include "user/navbar.html" %}
{% endif %}
```

---

## Practical Examples

### Example 1: Form Components Library

**macros.html:**
```html+jinja
{% macro render_input(name, type="text", placeholder="") %}
<div class="form-group">
  <input type="{{ type }}" name="{{ name }}" placeholder="{{ placeholder }}" class="form-control">
</div>
{% endmacro %}

{% macro render_button(text, type="button", class="btn-primary", disabled=False) %}
<button type="{{ type }}" class="btn {{ class }}" {{ 'disabled' if disabled else '' }}>
  {{ text }}
</button>
{% endmacro %}

{% macro render_field(name, label, type="text", required=False, help_text="") %}
<div class="form-field">
  <label>
    {{ label }}
    {% if required %}<span class="required">*</span>{% endif %}
  </label>
  <input type="{{ type }}" name="{{ name }}" class="form-control">
  {% if help_text %}
  <small class="help-text">{{ help_text }}</small>
  {% endif %}
</div>
{% endmacro %}
```

**Usage:**
```html+jinja
{% import "macros.html" as forms %}

<form>
  {{ forms.render_field("full_name", label="Full Name", required=True) }}
  {{ forms.render_field("email", label="Email", type="email", required=True) }}
  {{ forms.render_field("age", label="Age", type="number", help_text="Must be 18+") }}
  {{ forms.render_button("Submit", type="submit") }}
</form>
```

### Example 2: UI Components

**macros.html:**
```html+jinja
{% macro render_badge(text, type="info") %}
<span class="badge badge-{{ type }}">{{ text }}</span>
{% endmacro %}

{% macro render_list(items, ordered=False, class="") %}
{% if ordered %}
<ol class="{{ class }}">
{% else %}
<ul class="{{ class }}">
{% endif %}
  {% for item in items %}
  <li>{{ item }}</li>
  {% endfor %}
{% if ordered %}
</ol>
{% else %}
</ul>
{% endif %}
{% endmacro %}

{% macro render_card(title, content) %}
<div class="card">
  <div class="card-header">{{ title }}</div>
  <div class="card-body">{{ content }}</div>
</div>
{% endmacro %}
```

### Example 3: Data Tables

**macros.html:**
```html+jinja
{% macro render_table_row(cells, header=False) %}
<tr>
  {% for cell in cells %}
    {% if header %}
    <th>{{ cell }}</th>
    {% else %}
    <td>{{ cell }}</td>
    {% endif %}
  {% endfor %}
</tr>
{% endmacro %}
```

**Usage:**
```html+jinja
{% import "macros.html" as ui %}

<table>
  {{ ui.render_table_row(["Name", "Price", "Stock"], header=True) }}
  {% for product in products %}
    {{ ui.render_table_row([product.name, "$" ~ product.price, product.stock]) }}
  {% endfor %}
</table>
```

### Example 4: Navigation Components

**includes/header.html:**
```html+jinja
<header class="site-header">
  <h1>{{ site_title|default("My Site") }}</h1>
  <nav>
    <ul>
    {% for item in nav_items %}
      <li><a href="{{ item.url }}">{{ item.title }}</a></li>
    {% endfor %}
    </ul>
  </nav>
</header>
```

**page.html:**
```html+jinja
{% set site_title = "My Awesome Site" %}
{% set nav_items = [
  {"title": "Home", "url": "/"},
  {"title": "About", "url": "/about"},
  {"title": "Contact", "url": "/contact"}
] %}

{% include "includes/header.html" %}

<main>
  <h2>Welcome!</h2>
</main>

{% include "includes/footer.html" %}
```

### Example 5: Macros in Loops

```html+jinja
{% import "macros.html" as ui %}

<h3>Product Catalog</h3>
{% for product in products %}
  {{ ui.render_card(
    title=product.name,
    content="Price: $" ~ product.price ~ "<br>Stock: " ~ product.stock
  ) }}
{% endfor %}
```

### Example 6: Nested Macro Calls

Macros can call other macros:

**macros.html:**
```html+jinja
{% macro render_icon(name) %}
<i class="icon-{{ name }}"></i>
{% endmacro %}

{% macro render_button_with_icon(text, icon, type="button") %}
<button type="{{ type }}" class="btn">
  {{ render_icon(icon) }} {{ text }}
</button>
{% endmacro %}
```

**Usage:**
```html+jinja
{% import "macros.html" as ui %}
{{ ui.render_button_with_icon("Save", "save", type="submit") }}
```

---

## Limitations

###  Caller Function Not Supported

Jinja2's `caller()` function for macro call blocks is **NOT supported** in Miya Engine.

**Jinja2 (Not in Miya):**
```html+jinja
{#  This does NOT work in Miya #}
{% macro render_dialog(title) %}
<div class="dialog">
  <h3>{{ title }}</h3>
  <div class="content">
    {{ caller() }}  {# NOT SUPPORTED #}
  </div>
</div>
{% endmacro %}

{% call render_dialog("Hello") %}
  This content would be passed to caller()
{% endcall %}
```

**Workaround:** Pass content as a parameter:

```html+jinja
{#  Works in Miya #}
{% macro render_dialog(title, content) %}
<div class="dialog">
  <h3>{{ title }}</h3>
  <div class="content">{{ content }}</div>
</div>
{% endmacro %}

{% set dialog_content %}
  This content is passed as a parameter
{% endset %}

{{ render_dialog("Hello", dialog_content) }}
```

Or use direct macro invocation:

```html+jinja
{% macro render_dialog(title, content) %}
<div class="dialog">
  <h3>{{ title }}</h3>
  <div class="content">{{ content }}</div>
</div>
{% endmacro %}

{{ render_dialog("Hello", "<p>Direct HTML content</p>") }}
```

---

## Best Practices

### 1. Organize Macros by Purpose

```
templates/
  macros/
    forms.html      - Form components
    ui.html         - UI elements (badges, cards, etc.)
    tables.html     - Table components
    navigation.html - Nav components
```

### 2. Use Descriptive Names

```html+jinja
{#  Good #}
{% macro render_user_profile_card(user) %}

{#  Avoid #}
{% macro card(u) %}
```

### 3. Provide Sensible Defaults

```html+jinja
{% macro render_button(text, type="button", class="btn-primary") %}
  {# Sensible defaults make macros easy to use #}
{% endmacro %}
```

### 4. Document Complex Macros

```html+jinja
{#
  Renders a data table with sortable columns.

  Parameters:
    - headers: list of column headers
    - rows: list of data rows
    - sortable: boolean, enable sorting (default: false)
#}
{% macro render_data_table(headers, rows, sortable=False) %}
  ...
{% endmacro %}
```

### 5. Keep Macros Focused

Each macro should do one thing well:

```html+jinja
{#  Good - focused macro #}
{% macro render_input(name, type="text") %}
  ...
{% endmacro %}

{#  Avoid - too many responsibilities #}
{% macro render_entire_form_with_validation_and_submission(...) %}
  ...
{% endmacro %}
```

---

## Complete Reference

### Import Syntax

| Syntax | Description | Usage |
|--------|-------------|-------|
| `{% import "file.html" as name %}` | Import with namespace | `{{ name.macro() }}` |
| `{% from "file.html" import macro %}` | Direct import | `{{ macro() }}` |
| `{% from "file.html" import macro as m %}` | Import with alias | `{{ m() }}` |
| `{% from "file.html" import m1, m2 %}` | Import multiple | `{{ m1() }} {{ m2() }}` |

### Include Syntax

| Syntax | Description |
|--------|-------------|
| `{% include "file.html" %}` | Include with full context |
| `{% include "file.html" with context %}` | Explicit context (same as default) |

### Macro Features

 **Supported:**
- Macro definitions
- Default parameters
- Named parameters
- Variable number of parameters (as lists)
- Nested macro calls
- Macros in loops
- Import with namespace
- Selective import with `from`
- Template includes with context

 **Not Supported:**
- `caller()` function for call blocks
- `{% call %}` statement

---

## See Also

- [Working Example](https://github.com/zipreport/miya/tree/master/examples/features/macros-includes/) - Complete macro examples
- [Template Inheritance](TEMPLATE_INHERITANCE.md) - Extending templates
- [Control Structures](CONTROL_STRUCTURES_GUIDE.md) - Logic in templates
- [Miya Limitations](MIYA_LIMITATIONS.md) - Known limitations

---

## Summary

Macros and includes provide powerful template reusability:

** Fully Supported (95% Jinja2 Compatible):**
- Macro definitions with parameters
- Default parameter values
- Import with namespace
- Selective imports with `from`
- Template includes with context
- Nested macro calls
- Macros in loops

** Known Limitation:**
- `caller()` function not supported (use parameters instead)

Macros enable building component libraries for consistent, maintainable templates across your application.
