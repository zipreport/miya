# Comprehensions Guide

List and dictionary comprehensions provide concise syntax for creating collections. Comprehensions are a miya extension: Jinja2 has no list or dictionary comprehensions (`[x for x in y]` is a syntax error there). Miya supports the forms below and a few limitations.

> **Working Example:** See `examples/features/comprehensions/` for complete examples.

---

## Table of Contents

1. [List Comprehensions](#list-comprehensions)
2. [Dictionary Comprehensions](#dictionary-comprehensions)
3. [Filtering with `if`](#filtering-with-if)
4. [Limitations](#limitations)
5. [Workarounds](#workarounds)
6. [Practical Examples](#practical-examples)

---

## List Comprehensions

### Basic Syntax

Transform collections in a single line:

```html+jinja
{{ [expression for item in iterable] }}
```

### Simple Transformations

```html+jinja
{# Double each number #}
{{ [x * 2 for x in numbers] }}
→ [2, 4, 6, 8, 10]

{# Convert to uppercase #}
{{ [name|upper for name in names] }}
→ ["ALICE", "BOB", "CHARLIE"]

{# Extract property #}
{{ [user.name for user in users] }}
→ ["Alice", "Bob", "Charlie"]
```

### With Filters

Apply template filters in comprehensions:

```html+jinja
{# Title case all names #}
{{ [name|title for name in names] }}
→ ["Alice", "Bob", "Charlie"]

{# Format prices #}
{{ ["$" ~ (price|round(2)) for price in prices] }}
→ ["$9.99", "$19.99", "$29.99"]

{# Slugify titles #}
{{ [title|slugify for title in titles] }}
→ ["hello-world", "product-name", "user-guide"]

{# Chain multiple filters #}
{{ [name|trim|upper for name in names] }}
→ ["ALICE", "BOB", "CHARLIE"]
```

### Arithmetic Operations

```html+jinja
{# Apply discount #}
{{ [price * 0.9 for price in prices] }}
→ [8.99, 17.99, 26.99]

{# Calculate with expressions #}
{{ [item.price * item.qty for item in cart] }}
```

### String Operations

```html+jinja
{# Format user display #}
{{ [user.name ~ " (" ~ user.role ~ ")" for user in users] }}
→ ["Alice (admin)", "Bob (user)", "Charlie (moderator)"]

{# Concatenate with filter #}
{{ [name|upper for name in names] }}
→ ["ALICE", "BOB", "CHARLIE"]
```

---

## Dictionary Comprehensions

### Basic Syntax

Create dictionaries from iterables:

```html+jinja
{{ {key_expr: value_expr for item in iterable} }}
```

### Simple Mappings

```html+jinja
{# Create ID to name mapping #}
{{ {user.id: user.name for user in users} }}
→ {1: "Alice", 2: "Bob", 3: "Charlie"}

{# Email to role mapping #}
{{ {user.email: user.role for user in users} }}
→ {"alice@ex.com": "admin", "bob@ex.com": "user"}

{# Product SKU to name #}
{{ {product.sku: product.name for product in products} }}
→ {"LAP001": "Laptop", "MOU002": "Mouse"}
```

### With Transformations

```html+jinja
{# Format values #}
{{ {product.id: "$" ~ product.price for product in products} }}
→ {'1': '$999', '2': '$29', '3': '$79'}
```

Dictionary comprehension keys become strings: `{x: x * 2 for x in [1, 2]}` prints `{'1': 2, '2': 4}` (`&#39;` instead of `'` when autoescape is on).

---

## Filtering with `if`

Both list and dictionary comprehensions accept an `if` clause:

```html+jinja
{{ [x * 2 for x in [1, 2, 3, 4] if x is even] }}
→ [4, 8]

{{ [x for x in [1, 2, 3, 4] if x > 2] }}
→ [3, 4]

{{ {x: x * 2 for x in [1, 2, 3, 4] if x > 2} }}
→ {'3': 6, '4': 8}
```

---

## Limitations

###  Dict Unpacking with .items() Not Supported

**Python syntax (not supported in Miya):**
```html+jinja
{#  Does NOT work in Miya #}
{{ {k: v for k, v in data.items()} }}
{{ {k.upper(): v * 2 for k, v in config.items()} }}
```

**Error:** Tuple unpacking in dict comprehensions is not supported.

###  Nested Comprehensions Not Supported

**Python syntax (not supported in Miya):**
```html+jinja
{#  Does NOT work in Miya #}
{{ [item for sublist in lists for item in sublist] }}
{{ [x * y for x in range(3) for y in range(3)] }}
```

**Error:** Multiple `for` clauses are not supported.

---

## Workarounds

### Alternative 1: Use Template Filters

Filters can pre-filter data as well as an `if` clause:

```html+jinja
{{ numbers|select("greaterthan", 5)|list }}

{{ users|selectattr("active")|map(attribute="name")|list }}

{{ users|selectattr("age", ">=", 18)|list }}
```

### Alternative 2: Use For Loops

For complex filtering, use traditional loops:

```html+jinja
{#  Use for loop with condition #}
{% set result = [] %}
{% for x in numbers %}
  {% if x is even %}
    {% set _ = result.append(x * 2) %}
  {% endif %}
{% endfor %}
{{ result }}
→ [4, 8]

{# Or use loop filtering #}
{% for x in numbers if x is even %}
  {{ x * 2 }}{{ ", " if not loop.last else "" }}
{% endfor %}
→ 4, 8  (whitespace aside)
```

`result.append(...)` works on a list the template created (with `set`, `with` or `namespace()`). To change a list passed in as render data, copy it first: `{% set numbers = numbers|list %}`.

### Alternative 3: Prepare Data in Application

The best approach for complex filtering:

**In Go code:**
```go
// Filter data before passing to template
activeUsers := []User{}
for _, user := range users {
    if user.Active && user.Age >= 18 {
        activeUsers = append(activeUsers, user)
    }
}
ctx.Set("active_users", activeUsers)
```

**In template:**
```html+jinja
{# Simple comprehension on pre-filtered data #}
{{ [user.name for user in active_users] }}
```

### Alternative 4: Use Set with Block Assignment

```html+jinja
{# Collect filtered results #}
{% set filtered_names %}
  {% for user in users if user.active %}
    {{ user.name }}{{ "," if not loop.last else "" }}
  {% endfor %}
{% endset %}

{# Split back to list if needed #}
{{ filtered_names|trim|split(",") }}
```

---

## Practical Examples

### Real-World: Shopping Cart Summary

```html+jinja
{# Calculate line totals using list comprehension #}
{% set cart_items = [{"name": item.product.name, "price": item.price, "qty": item.quantity, "total": item.price * item.quantity} for item in cart.items] %}
{% set subtotal = [item.total for item in cart_items]|sum %}
{% set tax = subtotal * 0.08 %}

<div class="cart-summary">
  <h3>Order Summary</h3>
  {% for item in cart_items %}
    <div class="cart-item">
      <span>{{ item.name }} ({{ item.qty }}x)</span>
      <span>${{ item.total }}</span>
    </div>
  {% endfor %}
  <div class="totals">
    <div>Subtotal: ${{ subtotal }}</div>
    <div>Tax: ${{ tax|round(2) }}</div>
    <div class="total">Total: ${{ (subtotal + tax)|round(2) }}</div>
  </div>
</div>
```

### Real-World: Dynamic Form Fields

```html+jinja
{# Extract required fields and field types using comprehensions #}
{% set required_fields = [field.name for field in form_schema.fields] %}
{% set field_types = {field.name: field.type for field in form_schema.fields} %}

<form id="dynamic-form" method="post">
  {% for field in form_schema.fields %}
    <div class="form-group">
      <label for="{{ field.name }}">{{ field.label }}</label>
      <input type="{{ field_types[field.name] }}"
             name="{{ field.name }}"
             id="{{ field.name }}">
    </div>
  {% endfor %}
  <button type="submit">Submit</button>
</form>
```

### Example 1: Extract Names

```html+jinja
{# Simple property extraction #}
{% set user_names = [user.name for user in users] %}
<p>Users: {{ user_names|join(", ") }}</p>

{# With filter #}
{% set user_names = [user.name|title for user in users] %}
<p>Users: {{ user_names|join(", ") }}</p>
```

### Example 2: Calculate Totals

```html+jinja
{# Calculate line totals #}
{% set line_totals = [item.price * item.qty for item in cart] %}
<p>Line Totals: {{ line_totals }}</p>
<p>Cart Total: ${{ line_totals|sum|round(2) }}</p>

{# With discount #}
{% set discounted = [price * 0.9 for price in prices] %}
<p>Discounted Prices: {{ discounted }}</p>
```

### Example 3: Create Lookups

```html+jinja
{# Create ID lookup dictionary #}
{% set user_lookup = {user.id: user.name for user in users} %}

{# Use the lookup #}
<p>User 123: {{ user_lookup[123] }}</p>

{# Email to role mapping #}
{% set role_map = {user.email: user.role for user in users} %}
<p>alice@example.com is: {{ role_map["alice@example.com"] }}</p>
```

### Example 4: Format Display Names

```html+jinja
{# Create formatted names #}
{% set display_names = [user.name ~ " <" ~ user.email ~ ">" for user in users] %}

<select name="recipient">
{% for name in display_names %}
  <option>{{ name }}</option>
{% endfor %}
</select>
```

### Example 5: Nested Data (Use Loops)

```html+jinja
{#  Nested comprehension not supported #}
{# {{ [item.name for category in categories for item in category.items] }} #}

{#  Use nested loops instead #}
{% for category in categories %}
  <h3>{{ category.name }}</h3>
  <ul>
    {{ [item.name for item in category.items]|join(", ") }}
  </ul>
{% endfor %}
```

### Example 6: Filter Then Comprehend

```html+jinja
{# First filter with selectattr, then comprehend #}
{% set active_users = users|selectattr("active")|list %}
{% set active_emails = [user.email for user in active_users] %}

<p>Active user emails: {{ active_emails|join(", ") }}</p>

{# Or chain it all #}
{{ users|selectattr("active")|map(attribute="email")|join(", ") }}
```

### Example 7: Conditional Value Transformation

```html+jinja
{# Using ternary in comprehension #}
{% set statuses = [
  "Active" if user.active else "Inactive"
  for user in users
] %}

<ul>
{% for status in statuses %}
  <li>{{ status }}</li>
{% endfor %}
</ul>
```

### Example 8: Working with Zip (Use For Loop)

Tuple unpacking (`for a, b in zip(...)`) is not supported in comprehensions. Use a `{% for %}` loop instead:

```html+jinja
{# Combine multiple lists — use for loop (supports unpacking) #}
{% set names = ["Alice", "Bob", "Charlie"] %}
{% set scores = [95, 87, 92] %}

<ul>
{% for name, score in zip(names, scores) %}
  <li>{{ name }}: {{ score }}</li>
{% endfor %}
</ul>
```

---

## Feature Comparison

| Feature | Syntax | Status | Alternative |
|---------|--------|--------|-------------|
| Basic List | `[expr for x in list]` |  Supported | - |
| List with Filter | `[expr for x in list if cond]` |  Supported | - |
| Basic Dict | `{expr: expr for x in list}` |  Supported | - |
| Dict with .items() | `{k: v for k, v in dict.items()}` |  Not supported | Loop over list instead |
| Dict with Filter | `{k: v for x in list if cond}` |  Supported | - |
| Nested | `[x for list in lists for x in list]` |  Not supported | Use nested loops |
| With Filters | `[x|filter for x in list]` |  Supported | - |
| With Expressions | `[x * 2 for x in list]` |  Supported | - |
| With Zip (unpacking) | `[a + b for a, b in zip(l1, l2)]` |  Not supported | Use `{% for %}` loop with unpacking |
| With Ternary | `["yes" if x else "no" for x in list]` |  Supported | - |

---

## Best Practices

### 1. Use Filters for Filtering

```html+jinja
{#  Good - use filters #}
{{ users|selectattr("active")|map(attribute="name")|list }}

{#  Also works #}
{{ [user.name for user in users if user.active] }}
```

### 2. Keep Comprehensions Simple

```html+jinja
{#  Good - simple transformation #}
{{ [x * 2 for x in numbers] }}

{#  Avoid - too complex (tuple unpacking is not supported) #}
{# {{ [complex_function(x, y, z) for x, y, z in zip(a, b, c) if condition] }} #}
```

### 3. Pre-filter in Application Code

```html+jinja
{#  Good - filter in Go, comprehend in template #}
# Go: ctx.Set("active_users", filterActiveUsers(users))
{{ [user.name for user in active_users] }}
```

### 4. Use Descriptive Variable Names

```html+jinja
{#  Good - clear purpose #}
{% set user_emails = [user.email for user in users] %}

{#  Avoid - unclear #}
{% set x = [u.e for u in users] %}
```

---

## Summary

**Comprehensions in Miya Engine**

 **Fully Supported:**
- Basic list comprehensions `[expr for x in list]`
- Basic dict comprehensions `{k_expr: v_expr for x in list}`
- Comprehensions with filters `[x|filter for x in list]`
- Comprehensions with expressions `[x * 2 for x in list]`
- Comprehensions with zip/enumerate
- Nested property access `[user.name for user in users]`
- String concatenation in comprehensions
- Ternary operators in comprehensions
- `if` clauses `[x for x in list if condition]` and `{k: v for x in list if condition}`

 **Not Supported:**
- Dict comprehensions with .items() unpacking `{k: v for k, v in dict.items()}`
- Nested comprehensions `[x for list in lists for x in list]`

**Workarounds Available:**
- Use `selectattr`/`select`/`reject` filters for filtering
- Use traditional for loops for complex logic
- Pre-filter data in application code
- Use filter chains: `items|selectattr("active")|map(attribute="name")|list`

Miya's comprehension support covers the most common use cases. For advanced filtering, use the powerful filter system or prepare data in your application.

---

## See Also

- [Working Example](https://github.com/zipreport/miya/tree/master/examples/features/comprehensions/) - Complete comprehensions demo
- [Filters Guide](FILTERS_GUIDE.md) - Using selectattr/select for filtering
- [Control Structures](CONTROL_STRUCTURES_GUIDE.md) - Alternative loop syntax
- [Miya Limitations](MIYA_LIMITATIONS.md) - All known limitations
