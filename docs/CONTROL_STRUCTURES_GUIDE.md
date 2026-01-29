# Control Structures Guide

Control structures enable conditional logic, loops, and variable management in templates. Miya Engine provides full Jinja2-compatible control flow.

> **Working Example:** See `examples/features/control-structures/` for complete examples.

---

## Table of Contents

1. [Conditional Statements](#conditional-statements)
2. [For Loops](#for-loops)
3. [Inline Conditionals](#inline-conditionals)
4. [Variable Assignment](#variable-assignment)
5. [With Statements](#with-statements)
6. [Complex Control Flow](#complex-control-flow)

---

## Conditional Statements

### If / Elif / Else

Basic conditional branching:

```html+jinja
{% if user.role == "admin" %}
  <p>Admin access granted</p>
{% elif user.role == "moderator" %}
  <p>Moderator access granted</p>
{% elif user.role == "user" %}
  <p>Welcome, user!</p>
{% else %}
  <p>Access restricted</p>
{% endif %}
```

### Simple If Statement

```html+jinja
{% if user.active %}
  <p>Account is ACTIVE</p>
{% else %}
  <p>Account is INACTIVE</p>
{% endif %}
```

### Nested Conditions

```html+jinja
{% if user.active %}
  {% if user.verified %}
    <p> Fully verified and active account</p>
  {% else %}
    <p> Active but needs email verification</p>
  {% endif %}
{% else %}
  <p> Account needs activation</p>
{% endif %}
```

### Multiple Conditions

```html+jinja
{% if user.age >= 18 and user.verified %}
  <p>Access granted</p>
{% endif %}

{% if user.premium or user.trial %}
  <p>Premium features enabled</p>
{% endif %}
```

---

## For Loops

### Basic Iteration

```html+jinja
<ul>
{% for product in products %}
  <li>{{ product.name }} - ${{ product.price }}</li>
{% endfor %}
</ul>
```

### Loop Variables

Miya provides special loop variables:

| Variable | Description | Example Value |
|----------|-------------|---------------|
| `loop.index` | Current iteration (1-indexed) | 1, 2, 3... |
| `loop.index0` | Current iteration (0-indexed) | 0, 1, 2... |
| `loop.first` | True on first iteration | true/false |
| `loop.last` | True on last iteration | true/false |
| `loop.length` | Total items in loop | 10 |
| `loop.revindex` | Iterations remaining (1-indexed) | 10, 9, 8... |
| `loop.revindex0` | Iterations remaining (0-indexed) | 9, 8, 7... |

**Example:**

```html+jinja
<table>
{% for product in products %}
  <tr style="background: {{ 'lightblue' if loop.first else 'lightyellow' if loop.last else 'white' }}">
    <td>{{ loop.index }}</td>
    <td>{{ product.name }}</td>
    <td>{{ loop.revindex }} remaining</td>
  </tr>
{% endfor %}
</table>
```

### Conditional Iteration

Filter items directly in the loop:

```html+jinja
<p>Products over $50:</p>
<ul>
{% for product in products if product.price > 50 %}
  <li>{{ product.name }} - ${{ product.price }}</li>
{% endfor %}
</ul>
```

### Loop with Else

Handle empty collections:

```html+jinja
<ul>
{% for item in items %}
  <li>{{ item }}</li>
{% else %}
  <li><em>No items to display</em></li>
{% endfor %}
</ul>
```

### Nested Loops

Access parent loop variables with `loop.parent.loop`:

```html+jinja
{% for category in categories %}
  <h3>{{ category.name }}</h3>
  <ul>
  {% for item in category.items %}
    <li>{{ item }} (Category {{ loop.parent.loop.index }})</li>
  {% endfor %}
  </ul>
{% endfor %}
```

---

## Inline Conditionals

### Ternary Operator

Concise conditional expressions:

```html+jinja
{{ 'In Stock' if stock > 0 else 'Out of Stock' }}
{{ price if price > 0 else 'N/A' }}
{{ 'SALE!' if discount > 0 else '' }}
```

###  Important Limitation

**Inline conditionals MUST include the `else` clause** in Miya Engine:

```html+jinja
{#  WORKS - includes else #}
{{ 'Badge' if condition else '' }}

{#  FAILS - missing else #}
{{ 'Badge' if condition }}
```

### Chained Ternary

Multiple conditions in sequence:

```html+jinja
{{ 'Premium' if total > 1000 else 'Gold' if total > 500 else 'Silver' if total > 100 else 'Bronze' }}
```

Reads as:
- If total > 1000: "Premium"
- Else if total > 500: "Gold"
- Else if total > 100: "Silver"
- Else: "Bronze"

---

## Variable Assignment

### Set Statement

Assign values to variables:

```html+jinja
{% set greeting = "Hello, " ~ user.name ~ "!" %}
{% set total = 0 %}
{% set product = products[0] %}
```

### Multiple Assignments

```html+jinja
{% set first_product = products[0] %}
{% set second_product = products[1] %}
```

### Variables in Loops

Accumulate values (note: variables in loops are scoped):

```html+jinja
{% set total_price = 0 %}
{% for product in products %}
  {% set total_price = total_price + product.price %}
{% endfor %}
<p>Total: ${{ total_price }}</p>
```

### Block Assignment

Capture multi-line content:

```html+jinja
{% set formatted_html %}
<div class="alert">
  <strong>Important!</strong> You have {{ products|length }} items.
</div>
{% endset %}

{{ formatted_html }}
```

Block assignment captures everything between the tags, including HTML and template expressions.

---

## With Statements

### Variable Scoping

Create local scope for variables:

```html+jinja
{% with total = products|length %}
  <p>You have {{ total }} products in your catalog.</p>
  <p>{{ 'Large' if total > 5 else 'Small' }} catalog</p>
{% endwith %}
{# total is not available here #}
```

### Multiple Variables

```html+jinja
{% with min_price = 10, max_price = 100, currency = 'USD' %}
  <p>Price Range: {{ currency }} {{ min_price }} - {{ currency }} {{ max_price }}</p>
{% endwith %}
```

### Nested With

```html+jinja
{% with category = "Electronics" %}
  {% with count = products|length %}
    <p>{{ category }} category has {{ count }} items</p>
  {% endwith %}
{% endwith %}
```

### When to Use With

- **Clean templates:** Avoid repeating complex expressions
- **Performance:** Calculate expensive operations once
- **Readability:** Give meaningful names to complex filters

**Example:**

```html+jinja
{# Without with #}
<p>Active users: {{ users|selectattr("active")|list|length }}</p>
<p>Found {{ users|selectattr("active")|list|length }} matches</p>

{# With with - cleaner #}
{% with active_count = users|selectattr("active")|list|length %}
  <p>Active users: {{ active_count }}</p>
  <p>Found {{ active_count }} matches</p>
{% endwith %}
```

---

## Complex Control Flow

### Combining Conditions and Loops

```html+jinja
{% for product in products %}
  {% if loop.first %}
    <p><strong>Featured Product:</strong></p>
  {% endif %}

  {% if product.price > 50 %}
    <div>
      <p>{{ product.name }} - ${{ product.price }}</p>
      {% if product.stock < 10 %}
        <p style="color: red;"> Low Stock: {{ product.stock }} remaining</p>
      {% endif %}
    </div>
  {% endif %}

  {% if loop.last %}
    <p><em>End of product list</em></p>
  {% endif %}
{% endfor %}
```

### Complex Logic Example

```html+jinja
{% for user in users %}
  {% if user.active and user.verified %}
    {% with days_since = (today - user.last_login).days %}
      {% if days_since < 7 %}
        <span class="badge-active">{{ user.name }} - Active</span>
      {% elif days_since < 30 %}
        <span class="badge-warning">{{ user.name }} - Inactive</span>
      {% else %}
        <span class="badge-danger">{{ user.name }} - Dormant</span>
      {% endif %}
    {% endwith %}
  {% endif %}
{% endfor %}
```

---

## Practical Examples

### Example 1: User Dashboard

```html+jinja
{% if user.role == "admin" %}
  <h2>Admin Dashboard</h2>
  <ul>
  {% for action in admin_actions %}
    <li><a href="{{ action.url }}">{{ action.name }}</a></li>
  {% endfor %}
  </ul>
{% elif user.role == "user" %}
  <h2>User Dashboard</h2>
  <p>Welcome back, {{ user.name }}!</p>
{% else %}
  <p>Please log in to continue.</p>
{% endif %}
```

### Example 2: Shopping Cart

```html+jinja
{% set cart_total = 0 %}
<h3>Shopping Cart</h3>

{% for item in cart %}
  {% set item_total = item.price * item.quantity %}
  {% set cart_total = cart_total + item_total %}

  <div class="cart-item">
    <h4>{{ item.name }}</h4>
    <p>Price: ${{ item.price }} × {{ item.quantity }} = ${{ item_total }}</p>
  </div>
{% else %}
  <p>Your cart is empty</p>
{% endfor %}

{% if cart|length > 0 %}
  <p><strong>Total: ${{ cart_total }}</strong></p>
{% endif %}
```

### Example 3: Product Filtering

```html+jinja
<h3>Available Products</h3>

{% with in_stock = products|selectattr("stock", "greaterthan", 0)|list %}
  {% if in_stock|length > 0 %}
    <p>Showing {{ in_stock|length }} products in stock</p>
    <ul>
    {% for product in in_stock %}
      <li>
        {{ product.name }} - ${{ product.price }}
        {{ '(Low Stock!)' if product.stock < 5 else '' }}
      </li>
    {% endfor %}
    </ul>
  {% else %}
    <p>No products currently in stock</p>
  {% endif %}
{% endwith %}
```

---

## Best Practices

### 1. Keep Logic Simple

```html+jinja
{#  Good - simple and clear #}
{% if user.active %}
  Welcome back!
{% endif %}

{#  Avoid - too complex for template #}
{% if user.active and user.verified and user.payment_status == "current" and user.subscription_expires > today %}
  ...
{% endif %}
```

**Better approach:** Prepare data in application code, pass simple flags to template.

### 2. Use With for Clarity

```html+jinja
{#  Good - clear and reusable #}
{% with eligible = user.active and user.verified %}
  {% if eligible %}
    Premium features enabled
  {% endif %}
{% endwith %}
```

### 3. Avoid Deep Nesting

```html+jinja
{#  Avoid - hard to read #}
{% if condition1 %}
  {% if condition2 %}
    {% if condition3 %}
      {% for item in items %}
        ...
      {% endfor %}
    {% endif %}
  {% endif %}
{% endif %}

{#  Better - flatten logic #}
{% if condition1 and condition2 and condition3 %}
  {% for item in items %}
    ...
  {% endfor %}
{% endif %}
```

### 4. Always Include Else in Inline Conditionals

```html+jinja
{#  Works in Miya #}
{{ 'Active' if user.active else 'Inactive' }}

{#  Fails in Miya - missing else #}
{{ 'Active' if user.active }}
```

---

## Complete Reference

### Supported Control Structures

| Structure | Syntax | Description |
|-----------|--------|-------------|
| `if` | `{% if condition %}` | Conditional execution |
| `elif` | `{% elif condition %}` | Additional condition |
| `else` | `{% else %}` | Fallback branch |
| `for` | `{% for item in items %}` | Iteration |
| `for...if` | `{% for x in items if condition %}` | Filtered iteration |
| `for...else` | `{% for %}...{% else %}` | Empty handling |
| Inline if | `{{ value if condition else default }}` | Ternary operator |
| `set` | `{% set var = value %}` | Variable assignment |
| `with` | `{% with var = value %}` | Scoped variables |

### Loop Variables

All loop variables available: `index`, `index0`, `first`, `last`, `length`, `revindex`, `revindex0`, `parent.loop`

---

## See Also

- [Working Example](https://github.com/zipreport/miya/tree/master/examples/features/control-structures/) - Complete control structures demo
- [Tests & Operators](TESTS_AND_OPERATORS.md) - Conditional expressions
- [Filters Guide](FILTERS_GUIDE.md) - Data transformation in templates
- [Miya Limitations](MIYA_LIMITATIONS.md) - Known limitations

---

## Summary

**100% Jinja2 Compatible** control structures with one important limitation:

 **Fully Supported:**
- If/elif/else statements
- For loops with all loop variables
- Conditional iteration (`for...if`)
- Loop else clause
- Nested loops with parent access
- Set statements (simple and block)
- With statements (single and multiple variables)
- Nested control structures

 **Important Note:**
- Inline conditionals **MUST** include `else` clause

All control structures work identically to Jinja2, making migration and template sharing seamless.
