# Comprehensions Guide

List and dictionary comprehensions provide concise syntax for creating collections. Miya Engine supports basic comprehensions with some limitations compared to Jinja2.

> **Working Example:** See `examples/features/comprehensions/` for complete examples.

---

## Table of Contents

1. [List Comprehensions](#list-comprehensions)
2. [Dictionary Comprehensions](#dictionary-comprehensions)
3. [Limitations](#limitations)
4. [Workarounds](#workarounds)
5. [Practical Examples](#practical-examples)

---

## List Comprehensions

### Basic Syntax

Transform collections in a single line:

```jinja2
{{ [expression for item in iterable] }}
```

### Simple Transformations

```jinja2
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

```jinja2
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

```jinja2
{# Calculate totals #}
{{ [price * quantity for price, quantity in zip(prices, quantities)] }}

{# Apply discount #}
{{ [price * 0.9 for price in prices] }}
→ [8.99, 17.99, 26.99]

{# Complex expressions #}
{{ [price * qty * (1 - discount/100) for price, qty, discount in product_data] }}
```

### String Operations

```jinja2
{# Concatenate strings #}
{{ [first ~ " " ~ last for first, last in zip(first_names, last_names)] }}
→ ["Alice Smith", "Bob Jones", "Charlie Brown"]

{# Format user display #}
{{ [user.name ~ " (" ~ user.role ~ ")" for user in users] }}
→ ["Alice (admin)", "Bob (user)", "Charlie (moderator)"]
```

---

## Dictionary Comprehensions

### Basic Syntax

Create dictionaries from iterables:

```jinja2
{{ {key_expr: value_expr for item in iterable} }}
```

### Simple Mappings

```jinja2
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

```jinja2
{# Uppercase keys #}
{{ {key|upper: value for key, value in data} }}

{# Format values #}
{{ {product.id: "$" ~ product.price for product in products} }}
→ {1: "$999", 2: "$29", 3: "$79"}
```

---

## Limitations

###  Inline If Clauses Not Supported

**Jinja2 syntax (NOT in Miya):**
```jinja2
{#  Does NOT work in Miya #}
{{ [x for x in numbers if x > 5] }}
{{ [user.name for user in users if user.active] }}
{{ {k: v for k, v in dict.items() if v > 0} }}
```

**Error:** Comprehensions with inline `if` clauses are not supported in Miya Engine.

###  Dict Unpacking with .items() Not Supported

**Jinja2 syntax (NOT in Miya):**
```jinja2
{#  Does NOT work in Miya #}
{{ {k: v for k, v in data.items()} }}
{{ {k.upper(): v * 2 for k, v in config.items()} }}
```

**Error:** Tuple unpacking in dict comprehensions is not supported.

###  Nested Comprehensions Not Supported

**Jinja2 syntax (NOT in Miya):**
```jinja2
{#  Does NOT work in Miya #}
{{ [item for sublist in lists for item in sublist] }}
{{ [x * y for x in range(3) for y in range(3)] }}
```

**Error:** Multiple `for` clauses are not supported.

---

## Workarounds

### Alternative 1: Use Template Filters

Instead of inline `if`, use filters to pre-filter data:

```jinja2
{#  Not supported #}
{{ [x for x in numbers if x > 5] }}

{#  Use select filter #}
{{ numbers|select("greaterthan", 5)|list }}

{#  Not supported #}
{{ [user.name for user in users if user.active] }}

{#  Use selectattr filter #}
{{ users|selectattr("active")|map(attribute="name")|list }}

{#  Not supported #}
{{ [user for user in users if user.age >= 18] }}

{#  Use selectattr + list #}
{{ users|selectattr("age", ">=", 18)|list }}
```

### Alternative 2: Use For Loops

For complex filtering, use traditional loops:

```jinja2
{#  Not supported #}
{{ [x * 2 for x in numbers if x is even] }}

{#  Use for loop with condition #}
{% set result = [] %}
{% for x in numbers %}
  {% if x is even %}
    {% set _ = result.append(x * 2) %}
  {% endif %}
{% endfor %}
{{ result }}

{# Or use loop filtering #}
{% for x in numbers if x is even %}
  {{ x * 2 }}{{ ", " if not loop.last else "" }}
{% endfor %}
```

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
```jinja2
{# Simple comprehension on pre-filtered data #}
{{ [user.name for user in active_users] }}
```

### Alternative 4: Use Set with Block Assignment

```jinja2
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

### Example 1: Extract Names

```jinja2
{# Simple property extraction #}
{% set user_names = [user.name for user in users] %}
<p>Users: {{ user_names|join(", ") }}</p>

{# With filter #}
{% set user_names = [user.name|title for user in users] %}
<p>Users: {{ user_names|join(", ") }}</p>
```

### Example 2: Calculate Totals

```jinja2
{# Calculate line totals #}
{% set line_totals = [item.price * item.qty for item in cart] %}
<p>Line Totals: {{ line_totals }}</p>
<p>Cart Total: ${{ line_totals|sum|round(2) }}</p>

{# With discount #}
{% set discounted = [price * 0.9 for price in prices] %}
<p>Discounted Prices: {{ discounted }}</p>
```

### Example 3: Create Lookups

```jinja2
{# Create ID lookup dictionary #}
{% set user_lookup = {user.id: user.name for user in users} %}

{# Use the lookup #}
<p>User 123: {{ user_lookup[123] }}</p>

{# Email to role mapping #}
{% set role_map = {user.email: user.role for user in users} %}
<p>alice@example.com is: {{ role_map["alice@example.com"] }}</p>
```

### Example 4: Format Display Names

```jinja2
{# Create formatted names #}
{% set display_names = [user.name ~ " <" ~ user.email ~ ">" for user in users] %}

<select name="recipient">
{% for name in display_names %}
  <option>{{ name }}</option>
{% endfor %}
</select>
```

### Example 5: Nested Data (Use Loops)

```jinja2
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

```jinja2
{# First filter with selectattr, then comprehend #}
{% set active_users = users|selectattr("active")|list %}
{% set active_emails = [user.email for user in active_users] %}

<p>Active user emails: {{ active_emails|join(", ") }}</p>

{# Or chain it all #}
{{ users|selectattr("active")|map(attribute="email")|join(", ") }}
```

### Example 7: Conditional Value Transformation

```jinja2
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

### Example 8: Working with Zip

```jinja2
{# Combine multiple lists #}
{% set names = ["Alice", "Bob", "Charlie"] %}
{% set scores = [95, 87, 92] %}

{% set results = [
  name ~ ": " ~ score
  for name, score in zip(names, scores)
] %}

<ul>
{% for result in results %}
  <li>{{ result }}</li>
{% endfor %}
</ul>
```

---

## Feature Comparison

| Feature | Syntax | Status | Alternative |
|---------|--------|--------|-------------|
| Basic List | `[expr for x in list]` |  Supported | - |
| List with Filter | `[expr for x in list if cond]` |  Not supported | Use `selectattr`/`select` filters |
| Basic Dict | `{expr: expr for x in list}` |  Supported | - |
| Dict with .items() | `{k: v for k, v in dict.items()}` |  Not supported | Loop over list instead |
| Dict with Filter | `{k: v for x in list if cond}` |  Not supported | Pre-filter with `selectattr` |
| Nested | `[x for list in lists for x in list]` |  Not supported | Use nested loops |
| With Filters | `[x|filter for x in list]` |  Supported | - |
| With Expressions | `[x * 2 for x in list]` |  Supported | - |
| With Zip | `[a + b for a, b in zip(l1, l2)]` |  Supported | - |

---

## Best Practices

### 1. Use Filters for Filtering

```jinja2
{#  Good - use filters #}
{{ users|selectattr("active")|map(attribute="name")|list }}

{#  Avoid - inline if not supported #}
{{ [user.name for user in users if user.active] }}
```

### 2. Keep Comprehensions Simple

```jinja2
{#  Good - simple transformation #}
{{ [x * 2 for x in numbers] }}

{#  Avoid - too complex #}
{{ [complex_function(x, y, z) for x, y, z in zip(a, b, c) if condition] }}
```

### 3. Pre-filter in Application Code

```jinja2
{#  Good - filter in Go, comprehend in template #}
# Go: ctx.Set("active_users", filterActiveUsers(users))
{{ [user.name for user in active_users] }}

{#  Avoid - trying to filter in comprehension #}
{{ [user.name for user in users if user.active] }}
```

### 4. Use Descriptive Variable Names

```jinja2
{#  Good - clear purpose #}
{% set user_emails = [user.email for user in users] %}

{#  Avoid - unclear #}
{% set x = [u.e for u in users] %}
```

---

## Summary

**Comprehensions in Miya Engine: ~70% Jinja2 Compatible**

 **Fully Supported:**
- Basic list comprehensions `[expr for x in list]`
- Basic dict comprehensions `{k_expr: v_expr for x in list}`
- Comprehensions with filters `[x|filter for x in list]`
- Comprehensions with expressions `[x * 2 for x in list]`
- Comprehensions with zip/enumerate
- Nested property access `[user.name for user in users]`
- String concatenation in comprehensions
- Ternary operators in comprehensions

 **Not Supported:**
- Inline if clauses `[x for x in list if condition]`
- Dict comprehensions with .items() unpacking `{k: v for k, v in dict.items()}`
- Nested comprehensions `[x for list in lists for x in list]`
- Complex conditional filtering in comprehensions

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
