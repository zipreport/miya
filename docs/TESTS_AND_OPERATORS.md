# Tests and Operators Guide

Tests and operators enable conditional logic and comparisons in templates. Miya Engine provides comprehensive operator support and 26+ test expressions with 100% Jinja2 compatibility.

> **Working Example:** See `examples/features/tests-operators/` for complete examples.

---

## Table of Contents

1. [Operators](#operators)
   - [Arithmetic Operators](#arithmetic-operators)
   - [Comparison Operators](#comparison-operators)
   - [Logical Operators](#logical-operators)
   - [Membership Operators](#membership-operators)
2. [Test Expressions](#test-expressions)
   - [Type Tests](#type-tests)
   - [Container Tests](#container-tests)
   - [Numeric Tests](#numeric-tests)
   - [String Tests](#string-tests)
   - [Comparison Tests](#comparison-tests)
3. [Practical Examples](#practical-examples)

---

## Operators

### Arithmetic Operators

Perform mathematical operations:

| Operator | Description | Example | Result |
|----------|-------------|---------|--------|
| `+` | Addition | `{{ 5 + 3 }}` | `8` |
| `-` | Subtraction | `{{ 10 - 4 }}` | `6` |
| `*` | Multiplication | `{{ 6 * 7 }}` | `42` |
| `/` | Division | `{{ 15 / 3 }}` | `5.0` |
| `//` | Floor Division | `{{ 17 // 5 }}` | `3` |
| `%` | Modulo | `{{ 17 % 5 }}` | `2` |
| `**` | Power | `{{ 2 ** 8 }}` | `256` |
| `~` | String Concat | `{{ "Hello" ~ " " ~ "World" }}` | `"Hello World"` |

**Examples:**

```jinja2
{# Basic arithmetic #}
<p>Total: ${{ price * quantity }}</p>
<p>Average: {{ total / count }}</p>
<p>Remaining: {{ 100 - used }}</p>

{# Floor division for whole numbers #}
<p>Pages: {{ total_items // items_per_page }}</p>

{# Modulo for even/odd #}
<p class="{{ 'even' if item_num % 2 == 0 else 'odd' }}">

{# Power for calculations #}
<p>Area: {{ radius ** 2 * 3.14159 }}</p>

{# String concatenation #}
<p>{{ first_name ~ " " ~ last_name }}</p>
<p>{{ "Total: $" ~ price|string }}</p>
```

### Comparison Operators

Compare values:

| Operator | Description | Example | Result |
|----------|-------------|---------|--------|
| `==` | Equal to | `{{ 5 == 5 }}` | `true` |
| `!=` | Not equal to | `{{ 5 != 3 }}` | `true` |
| `<` | Less than | `{{ 3 < 5 }}` | `true` |
| `<=` | Less or equal | `{{ 5 <= 5 }}` | `true` |
| `>` | Greater than | `{{ 7 > 3 }}` | `true` |
| `>=` | Greater or equal | `{{ 5 >= 5 }}` | `true` |

**Examples:**

```jinja2
{# Price comparisons #}
{% if product.price < 100 %}
  <span class="affordable">Affordable</span>
{% endif %}

{# Stock checks #}
{% if product.stock > 0 %}
  <button>Add to Cart</button>
{% else %}
  <span>Out of Stock</span>
{% endif %}

{# Age verification #}
{% if user.age >= 18 %}
  <p>Access granted</p>
{% endif %}

{# Equality checks #}
{% if user.role == "admin" %}
  <a href="/admin">Admin Panel</a>
{% endif %}

{# Inequality #}
{% if current_page != total_pages %}
  <a href="?page={{ current_page + 1 }}">Next</a>
{% endif %}
```

### Logical Operators

Combine boolean expressions:

| Operator | Description | Example | Result |
|----------|-------------|---------|--------|
| `and` | Logical AND | `{{ true and true }}` | `true` |
| `or` | Logical OR | `{{ true or false }}` | `true` |
| `not` | Logical NOT | `{{ not false }}` | `true` |

**Examples:**

```jinja2
{# AND - all conditions must be true #}
{% if user.active and user.verified %}
  <p>Full access granted</p>
{% endif %}

{% if product.price < 100 and product.stock > 0 %}
  <button>Buy Now</button>
{% endif %}

{# OR - any condition can be true #}
{% if user.role == "admin" or user.role == "moderator" %}
  <a href="/manage">Management</a>
{% endif %}

{# NOT - negate condition #}
{% if not user.blocked %}
  <p>Welcome!</p>
{% endif %}

{# Complex combinations #}
{% if (user.active and user.verified) or user.role == "admin" %}
  <p>Account is valid</p>
{% endif %}

{% if user.age >= 18 and (user.verified or user.trusted) %}
  <p>Access granted</p>
{% endif %}
```

**Operator Precedence:**
1. `not`
2. `and`
3. `or`

Use parentheses `()` to control order:

```jinja2
{# Different results based on grouping #}
{{ true or false and false }}    → true  (and first)
{{ (true or false) and false }}  → false (or first)
```

### Membership Operators

Test membership in collections:

| Operator | Description | Example | Result |
|----------|-------------|---------|--------|
| `in` | Is member of | `{{ 3 in [1, 2, 3] }}` | `true` |
| `not in` | Is not member | `{{ 10 not in [1, 2, 3] }}` | `true` |

**Examples:**

```jinja2
{# List membership #}
{% if "admin" in user.roles %}
  <p>Admin access</p>
{% endif %}

{# String contains #}
{% if "error" in message|lower %}
  <div class="alert-danger">{{ message }}</div>
{% endif %}

{# Not in #}
{% if product.id not in purchased_ids %}
  <button>Add to Cart</button>
{% else %}
  <span>Already Purchased</span>
{% endif %}

{# Check tag presence #}
{% if "featured" in product.tags %}
  <span class="badge">Featured</span>
{% endif %}
```

---

## Test Expressions

Tests use the `is` syntax to check conditions:

```jinja2
{{ value is test_name }}
{{ value is test_name(argument) }}
{{ value is not test_name }}
```

### Type Tests

Check variable types:

| Test | Description | Example |
|------|-------------|---------|
| `defined` | Variable exists | `{{ user is defined }}` |
| `undefined` | Variable doesn't exist | `{{ foo is undefined }}` |
| `none` | Value is None/null | `{{ value is none }}` |
| `boolean` | Is boolean | `{{ true is boolean }}` |
| `string` | Is string | `{{ "text" is string }}` |
| `number` | Is number | `{{ 42 is number }}` |
| `integer` | Is integer | `{{ 42 is integer }}` |
| `float` | Is float | `{{ 3.14 is float }}` |

**Examples:**

```jinja2
{# Check if variable exists #}
{% if user is defined %}
  <p>Welcome, {{ user.name }}!</p>
{% else %}
  <p>Please log in</p>
{% endif %}

{# Check for null/none #}
{% if error_message is none %}
  <p>No errors</p>
{% else %}
  <div class="alert">{{ error_message }}</div>
{% endif %}

{# Type checking #}
{% if value is string %}
  <p>Text: {{ value }}</p>
{% elif value is number %}
  <p>Number: {{ value }}</p>
{% endif %}

{# Integer vs float #}
{% if price is integer %}
  ${{ price }}
{% else %}
  ${{ price|round(2) }}
{% endif %}
```

### Container Tests

Check container types:

| Test | Description | Example |
|------|-------------|---------|
| `sequence` | Is list/tuple | `{{ [1,2,3] is sequence }}` |
| `mapping` | Is dictionary | `{{ {} is mapping }}` |
| `iterable` | Can be iterated | `{{ "text" is iterable }}` |

**Examples:**

```jinja2
{# Check if iterable #}
{% if items is iterable %}
  <ul>
  {% for item in items %}
    <li>{{ item }}</li>
  {% endfor %}
  </ul>
{% endif %}

{# Dictionary check #}
{% if config is mapping %}
  {% for key, value in config.items() %}
    <p>{{ key }}: {{ value }}</p>
  {% endfor %}
{% endif %}

{# Sequence check #}
{% if data is sequence %}
  <p>List with {{ data|length }} items</p>
{% endif %}
```

### Numeric Tests

Tests for numbers:

| Test | Description | Example |
|------|-------------|---------|
| `even` | Is even number | `{{ 4 is even }}` → `true` |
| `odd` | Is odd number | `{{ 5 is odd }}` → `true` |
| `divisibleby(n)` | Divisible by n | `{{ 15 is divisibleby(3) }}` → `true` |

**Examples:**

```jinja2
{# Alternating row colors #}
{% for item in items %}
  <tr class="{{ 'even-row' if loop.index is even else 'odd-row' }}">
    <td>{{ item }}</td>
  </tr>
{% endfor %}

{# Check divisibility #}
{% if total is divisibleby(10) %}
  <p>Exact multiple of 10</p>
{% endif %}

{# Odd/even logic #}
{% if page_num is odd %}
  <div class="right-page">{{ content }}</div>
{% else %}
  <div class="left-page">{{ content }}</div>
{% endif %}
```

### String Tests

Tests for strings:

| Test | Description | Example |
|------|-------------|---------|
| `lower` | All lowercase | `{{ "hello" is lower }}` → `true` |
| `upper` | All uppercase | `{{ "HELLO" is upper }}` → `true` |
| `startswith(s)` | Starts with string | `{{ "Hello" is startswith("He") }}` → `true` |
| `endswith(s)` | Ends with string | `{{ "World" is endswith("ld") }}` → `true` |

**Examples:**

```jinja2
{# File type checking #}
{% if filename is endswith(".pdf") %}
  <span class="icon-pdf">{{ filename }}</span>
{% elif filename is endswith(".jpg") or filename is endswith(".png") %}
  <img src="{{ filename }}" alt="Image">
{% endif %}

{# URL protocol check #}
{% if url is startswith("https://") %}
  <span class="secure-badge">Secure</span>
{% endif %}

{# Case validation #}
{% if code is upper %}
  <p>Valid product code: {{ code }}</p>
{% else %}
  <p class="error">Code must be uppercase</p>
{% endif %}
```

### Comparison Tests

Alternative comparison syntax:

| Test | Description | Example |
|------|-------------|---------|
| `equalto(value)` | Equal to value | `{{ 5 is equalto(5) }}` → `true` |
| `sameas(value)` | Identity check | `{{ true is sameas(true) }}` → `true` |

**Examples:**

```jinja2
{# Equality test #}
{% if status is equalto("active") %}
  <span class="badge-success">Active</span>
{% endif %}

{# Same as - stricter than == #}
{% if value is sameas(true) %}
  <p>Explicitly true</p>
{% endif %}
```

### Negated Tests

Use `is not` to negate any test:

```jinja2
{# Not defined #}
{% if error is not defined %}
  <p>No errors</p>
{% endif %}

{# Not none #}
{% if user.email is not none %}
  <p>Email: {{ user.email }}</p>
{% endif %}

{# Not even (i.e., odd) #}
{% if number is not even %}
  <p>{{ number }} is odd</p>
{% endif %}

{# Not empty #}
{% if items is not empty %}
  <p>Found {{ items|length }} items</p>
{% endif %}
```

---

## Practical Examples

### Example 1: User Access Control

```jinja2
{% if user is defined and user.active %}
  {% if user.role == "admin" %}
    <div class="admin-panel">
      <h2>Admin Dashboard</h2>
      <p>Full system access</p>
    </div>
  {% elif user.role == "moderator" %}
    <div class="mod-panel">
      <h2>Moderator Tools</h2>
      <p>Content management access</p>
    </div>
  {% elif user.verified and user.age >= 18 %}
    <div class="user-panel">
      <h2>User Dashboard</h2>
      <p>Standard access</p>
    </div>
  {% else %}
    <div class="restricted">
      <p>Limited access - please verify your account</p>
    </div>
  {% endif %}
{% else %}
  <div class="login-prompt">
    <p>Please log in to continue</p>
  </div>
{% endif %}
```

### Example 2: Product Display

```jinja2
{% for product in products %}
  <div class="product-card">
    <h3>{{ product.name }}</h3>
    <p>${{ product.price }}</p>

    {# Stock status #}
    {% if product.stock > 0 %}
      {% if product.stock < 10 %}
        <p class="warning">Only {{ product.stock }} left!</p>
      {% else %}
        <p class="success">In Stock</p>
      {% endif %}
      <button>Add to Cart</button>
    {% else %}
      <p class="error">Out of Stock</p>
      <button disabled>Notify Me</button>
    {% endif %}

    {# Discount badge #}
    {% if product.discount is defined and product.discount > 0 %}
      <span class="badge">{{ product.discount }}% OFF</span>
    {% endif %}

    {# Featured #}
    {% if "featured" in product.tags %}
      <span class="star"> Featured</span>
    {% endif %}
  </div>
{% endfor %}
```

### Example 3: Form Validation Display

```jinja2
<form>
  {# Username field #}
  <div class="form-group">
    <label>Username</label>
    <input type="text" name="username" value="{{ username|default('') }}">
    {% if errors.username is defined %}
      <span class="error">{{ errors.username }}</span>
    {% endif %}
  </div>

  {# Email field #}
  <div class="form-group">
    <label>Email</label>
    <input type="email" name="email" value="{{ email|default('') }}">
    {% if errors.email is defined %}
      <span class="error">{{ errors.email }}</span>
    {% elif email is defined and "@" not in email %}
      <span class="warning">Email looks invalid</span>
    {% endif %}
  </div>

  {# Age field #}
  <div class="form-group">
    <label>Age</label>
    <input type="number" name="age" value="{{ age|default('') }}">
    {% if age is defined %}
      {% if age is not number %}
        <span class="error">Must be a number</span>
      {% elif age < 18 %}
        <span class="error">Must be 18 or older</span>
      {% endif %}
    {% endif %}
  </div>

  <button type="submit">Submit</button>
</form>
```

### Example 4: Data Table with Highlighting

```jinja2
<table>
  <thead>
    <tr>
      <th>ID</th>
      <th>Name</th>
      <th>Amount</th>
      <th>Status</th>
    </tr>
  </thead>
  <tbody>
  {% for row in data %}
    <tr class="{{ 'highlight' if row.amount > 1000 else '' }}
               {{ 'even' if loop.index is even else 'odd' }}">
      <td>{{ row.id }}</td>
      <td>{{ row.name }}</td>
      <td>
        {% if row.amount is divisibleby(100) %}
          <strong>${{ row.amount }}</strong>
        {% else %}
          ${{ row.amount }}
        {% endif %}
      </td>
      <td>
        {% if row.status is equalto("completed") %}
          <span class="badge-success"> Completed</span>
        {% elif row.status is equalto("pending") %}
          <span class="badge-warning"> Pending</span>
        {% else %}
          <span class="badge-danger"> Failed</span>
        {% endif %}
      </td>
    </tr>
  {% endfor %}
  </tbody>
</table>
```

### Example 5: Complex Conditional Logic

```jinja2
{% for order in orders %}
  <div class="order">
    <h3>Order #{{ order.id }}</h3>

    {# Multi-condition status #}
    {% if order.paid and order.shipped and order.delivered %}
      <span class="badge-success"> Complete</span>
    {% elif order.paid and order.shipped %}
      <span class="badge-info"> In Transit</span>
    {% elif order.paid %}
      <span class="badge-warning"> Processing</span>
    {% else %}
      <span class="badge-danger"> Payment Required</span>
    {% endif %}

    {# Priority indicator #}
    {% if order.priority is defined and order.priority is upper %}
      <span class="priority-badge">{{ order.priority }}</span>
    {% endif %}

    {# Total calculation #}
    {% if order.total is number and order.total > 0 %}
      <p>Total: ${{ order.total|round(2) }}</p>

      {% if order.total >= 100 %}
        <p class="success"> Free shipping applied</p>
      {% elif order.total > 50 %}
        <p class="info">Add ${{ (100 - order.total)|round(2) }} for free shipping</p>
      {% endif %}
    {% endif %}
  </div>
{% endfor %}
```

---

## Complete Reference

### Operators Summary

| Category | Operators | Example |
|----------|-----------|---------|
| **Arithmetic** | `+`, `-`, `*`, `/`, `//`, `%`, `**`, `~` | `{{ 5 + 3 }}` |
| **Comparison** | `==`, `!=`, `<`, `<=`, `>`, `>=` | `{{ age >= 18 }}` |
| **Logical** | `and`, `or`, `not` | `{{ a and b }}` |
| **Membership** | `in`, `not in` | `{{ x in list }}` |

### Tests Summary

| Category | Tests | Count |
|----------|-------|-------|
| **Type** | defined, undefined, none, boolean, string, number, integer, float | 8 |
| **Container** | sequence, mapping, iterable, callable | 4 |
| **Numeric** | even, odd, divisibleby | 3 |
| **String** | lower, upper, startswith, endswith, match, alpha, alnum | 7 |
| **Comparison** | equalto, sameas, in, contains | 4 |
| **TOTAL** | | **26+ Tests** |

---

## Best Practices

### 1. Use Tests for Type Checking

```jinja2
{#  Good - explicit type test #}
{% if value is defined and value is not none %}
  {{ value }}
{% endif %}

{#  Avoid - implicit truthiness can be unclear #}
{% if value %}
  {{ value }}
{% endif %}
```

### 2. Combine Operators Clearly

```jinja2
{#  Good - clear with parentheses #}
{% if (user.active and user.verified) or user.role == "admin" %}

{#  Avoid - precedence unclear #}
{% if user.active and user.verified or user.role == "admin" %}
```

### 3. Use Appropriate Tests

```jinja2
{#  Good - use even test #}
{% if number is even %}

{#  Avoid - manual modulo #}
{% if number % 2 == 0 %}
```

---

## See Also

- [Working Example](../examples/features/tests-operators/) - Complete tests and operators demo
- [Control Structures](CONTROL_STRUCTURES_GUIDE.md) - Using tests in conditionals
- [Filters Guide](FILTERS_GUIDE.md) - Data transformation
- [Miya Limitations](MIYA_LIMITATIONS.md) - Known limitations

---

## Summary

**Complete operator and test support - 100% Jinja2 Compatible**

 **All Operators:**
- Arithmetic: `+`, `-`, `*`, `/`, `//`, `%`, `**`, `~`
- Comparison: `==`, `!=`, `<`, `<=`, `>`, `>=`
- Logical: `and`, `or`, `not`
- Membership: `in`, `not in`

 **26+ Tests:**
- Type tests (defined, none, string, number, etc.)
- Container tests (sequence, mapping, iterable)
- Numeric tests (even, odd, divisibleby)
- String tests (lower, upper, startswith, endswith)
- Comparison tests (equalto, sameas)
- All tests support negation with `is not`

All operators and tests work identically to Jinja2, providing powerful conditional logic for templates.
