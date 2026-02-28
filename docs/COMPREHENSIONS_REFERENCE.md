# List/Dict Comprehensions - Quick Reference

> **Note:** This reference reflects Miya Engine's actual comprehension support. For the full guide with workarounds, see [COMPREHENSIONS_GUIDE.md](COMPREHENSIONS_GUIDE.md).

## Supported Syntax

### List Comprehensions
```html+jinja
{{ [expression for item in iterable] }}
```

### Dictionary Comprehensions
```html+jinja
{{ {key_expr: value_expr for item in iterable} }}
```

## Supported Patterns

| Pattern | Example | Result |
|---------|---------|--------|
| **Transform** | `[x * 2 for x in [1,2,3]]` | `[2, 4, 6]` |
| **Extract** | `[user.name for user in users]` | `["Alice", "Bob"]` |
| **Dict from List** | `{user.id: user.name for user in users}` | `{1: "Alice", 2: "Bob"}` |
| **With Filters** | `[name\|title for name in names]` | `["Alice", "Bob"]` |
| **With Expressions** | `[x * 2 + 1 for x in nums]` | Transformed list |
| **With Ternary** | `["yes" if x else "no" for x in flags]` | Conditional values |

## Not Supported

| Pattern | Example | Workaround |
|---------|---------|------------|
| **Inline if** | `[x for x in list if cond]` | Use `selectattr`/`select` filters |
| **Tuple unpacking** | `[a ~ b for a, b in zip(l1, l2)]` | Use `{% for %}` loop with unpacking instead |
| **Dict .items()** | `{k: v for k, v in dict.items()}` | Loop over list of objects instead |
| **Nested** | `[x for l in lists for x in l]` | Use nested for loops |

## Operators in Comprehensions

| Operation | Syntax | Example |
|-----------|---------|---------|
| **Arithmetic** | `+`, `-`, `*`, `/`, `**`, `%` | `[x ** 2 for x in nums]` |
| **String concat** | `~` | `[user.name ~ " <" ~ user.email ~ ">" for user in users]` |
| **Ternary** | `a if cond else b` | `["Yes" if x else "No" for x in flags]` |
| **Filters** | `\|filter_name` | `[name\|upper for name in names]` |

## Common Use Cases

### Data Extraction
```html+jinja
{# Extract a property from a list of objects #}
{{ [user.email for user in users] }}

{# Sum values #}
{{ [order.total for order in orders]|sum }}
```

### Data Transformation
```html+jinja
{# Apply arithmetic #}
{{ [price * 1.1 for price in prices] }}

{# Format strings #}
{{ [user.name ~ " <" ~ user.email ~ ">" for user in users] }}

{# Apply filters #}
{{ [name|title for name in raw_names] }}
```

### Creating Lookups
```html+jinja
{# ID to name mapping #}
{% set user_lookup = {user.id: user.name for user in users} %}
{{ user_lookup[123] }}

{# Email to role mapping #}
{% set role_map = {user.email: user.role for user in users} %}
```

### Chaining with Filters
```html+jinja
{# Apply filters after comprehension #}
{{ [user.name for user in users]|join(", ") }}

{# Sort results #}
{{ [user.name for user in users]|sort }}

{# Get unique values #}
{{ [item.category for item in items]|unique }}
```

## Error Handling

### Common Errors
```html+jinja
{# Undefined variable — error in strict mode, empty in silent mode #}
{{ [x for x in undefined_list] }}

{# Wrong variable scope #}
{{ {x: y for x in items} }}  {# Error: y is not defined #}

{# Non-iterable #}
{{ [x for x in 123] }}       {# Error: int is not iterable #}
```

### Safe Patterns
```html+jinja
{# Use default filter for potentially undefined lists #}
{{ [x for x in items|default([])] }}

{# Proper variable scope #}
{{ {item.key: item.value for item in items} }}
```

## Debugging Tips

### 1. Test Components Separately
```html+jinja
{# Test the iterable first #}
{{ items }}

{# Then test simple iteration #}
{{ [item for item in items] }}

{# Then add the expression #}
{{ [item.name for item in items] }}
```

### 2. Use Intermediate Variables
```html+jinja
{# Break complex operations into steps #}
{% set filtered = users|selectattr("active")|list %}
{% set names = [user.name for user in filtered] %}
{{ names|join(", ") }}
```

## Workaround Patterns

For features not directly supported, use filter chains:

```html+jinja
{# Instead of: [x for x in numbers if x > 5] #}
{{ numbers|select("greaterthan", 5)|list }}

{# Instead of: [user.name for user in users if user.active] #}
{{ users|selectattr("active")|map(attribute="name")|list }}

{# Instead of: [user for user in users if user.age >= 18] #}
{{ users|selectattr("age", ">=", 18)|list }}
```

---

## See Also

- [Comprehensions Guide](COMPREHENSIONS_GUIDE.md) - Full guide with examples and workarounds
- [Filters Guide](FILTERS_GUIDE.md) - Using selectattr/select for filtering
- [Control Structures](CONTROL_STRUCTURES_GUIDE.md) - Alternative loop syntax
- [Miya Limitations](MIYA_LIMITATIONS.md) - All known limitations
