# List/Dict Comprehensions - Quick Reference

## Syntax Patterns

### List Comprehensions
```jinja2
{{ [expression for item in iterable] }}
{{ [expression for item in iterable if condition] }}
```

### Dictionary Comprehensions  
```jinja2
{{ {key_expr: value_expr for item in iterable} }}
{{ {key_expr: value_expr for item in iterable if condition} }}
```

## Common Patterns

| Pattern | Example | Result |
|---------|---------|--------|
| **Transform** | `[x * 2 for x in [1,2,3]]` | `[2, 4, 6]` |
| **Filter** | `[x for x in [1,2,3,4] if x > 2]` | `[3, 4]` |
| **Extract** | `[user.name for user in users]` | `["Alice", "Bob"]` |
| **Dict from List** | `{user.id: user.name for user in users}` | `{1: "Alice", 2: "Bob"}` |
| **Transform Dict** | `{k.upper(): v for k,v in dict.items()}` | `{"KEY": "value"}` |
| **Filter Dict** | `{k: v for k,v in dict.items() if v > 10}` | Filtered dict |

## Operators & Functions

| Operation | Syntax | Example |
|-----------|---------|---------|
| **Arithmetic** | `+`, `-`, `*`, `/`, `**`, `%` | `[x ** 2 for x in nums]` |
| **Comparison** | `>`, `<`, `>=`, `<=`, `==`, `!=` | `[x for x in nums if x >= 0]` |
| **Logic** | `and`, `or`, `not` | `[x for x in items if x.active and x.public]` |
| **Membership** | `in`, `not in` | `[x for x in items if x.category in allowed]` |
| **String ops** | `.upper()`, `.lower()`, `.strip()` | `[name.title() for name in names]` |
| **Filters** | `|filter_name` | `[name|title for name in names]` |

## Complex Conditions

```jinja2
<!-- Multiple conditions -->
{{ [item for item in items if item.price > 100 and item.category == "electronics"] }}

<!-- Nested conditions -->
{{ [user for user in users if user.role == "admin" or (user.role == "user" and user.verified)] }}

<!-- Function calls in conditions -->
{{ [post for post in posts if post.published_date > days_ago(7)] }}
```

## Working with Different Data Types

### Lists
```jinja2
{{ [item.upper() for item in ["a", "b", "c"]] }}
<!-- Result: ["A", "B", "C"] -->
```

### Dictionaries
```jinja2
{{ [v for k, v in {"a": 1, "b": 2}.items()] }}
<!-- Result: [1, 2] -->
```

### Objects/Maps
```jinja2
{{ [user.email for user in users if user.active] }}
<!-- Extract emails from active users -->
```

### Nested Structures
```jinja2
{{ [tag.name for post in posts for tag in post.tags] }}
<!-- Flatten all tags from all posts -->
```

## Performance Tips

###  Good Practices
```jinja2
<!-- Simple transformations -->
{{ [user.name for user in users] }}

<!-- Filter before expensive operations -->
{{ [expensive_func(item) for item in items if item.enabled] }}

<!-- Use built-in filters when available -->
{{ users|selectattr("active")|map(attribute="name")|list }}
```

###  Avoid These
```jinja2
<!-- Don't nest too deeply -->
{{ [[f(x) for x in row] for row in [[g(y) for y in col] for col in matrix]] }}

<!-- Don't repeat expensive operations -->
{{ [expensive_func(item) for item in items if expensive_func(item) > threshold] }}

<!-- Don't create huge lists unnecessarily -->
{% for item in [process_all(huge_list)] %}...{% endfor %}
```

## Error Handling

### Common Errors
```jinja2
<!--  Undefined variable -->
{{ [x for x in undefined_list] }}
<!-- Error: undefined variable -->

<!--  Wrong variable scope -->
{{ {x: y for x in items} }}  
<!-- Error: y is not defined -->

<!--  Non-iterable -->
{{ [x for x in 123] }}
<!-- Error: int is not iterable -->
```

### Safe Patterns
```jinja2
<!--  Check if defined -->
{{ [x for x in (items or [])] }}

<!--  Use default filter -->
{{ [x for x in items|default([])] }}

<!--  Proper variable scope -->
{{ {item.key: item.value for item in items} }}
```

## Debugging Tips

### 1. Test Components Separately
```jinja2
<!-- Test the iterable first -->
{{ items }}

<!-- Test the condition -->
{{ [item for item in items] }}

<!-- Test the expression -->
{{ [item.name for item in items] }}
```

### 2. Use Intermediate Variables
```jinja2
<!-- Instead of complex one-liner -->
{% set filtered_items = [item for item in items if item.active] %}
{% set processed_items = [process(item) for item in filtered_items] %}
{{ processed_items }}
```

### 3. Add Debug Output
```jinja2
<!-- Debug the iteration -->
{{ [debug(item) or item.name for item in items] }}
```

## Integration with Filters

### Chaining with Filters
```jinja2
<!-- Apply filters after comprehension -->
{{ [user.name for user in users]|join(", ") }}

<!-- Apply filters within comprehension -->
{{ [name|title for name in raw_names] }}

<!-- Combine both -->
{{ [name|title for name in names if name]|sort|join(", ") }}
```

### Common Filter Combinations
```jinja2
<!-- Count items -->
{{ [1 for item in items if condition]|length }}

<!-- Get unique values -->
{{ [item.category for item in items]|unique }}

<!-- Sum values -->
{{ [item.price for item in cart_items]|sum }}

<!-- Sort results -->
{{ [user.name for user in users]|sort }}
```

## Template Organization

### Keep Templates Clean
```jinja2
<!--  Good: Short and readable -->
{% set active_users = [u for u in users if u.active] %}
<div>{{ active_users|length }} active users</div>

<!--  Avoid: Inline complex comprehensions -->
<div>{{ [u for u in users if u.active and u.last_login > days_ago(30) and u.role in ['admin', 'moderator']]|length }} active users</div>
```

### Use Macros for Reusable Logic
```jinja2
{% macro filter_active_items(items) %}
  {{ [item for item in items if item.active and item.visible] }}
{% endmacro %}

<!-- Use the macro -->
{% set products = filter_active_items(all_products) %}
```

---

## Quick Examples by Use Case

### Data Extraction
```jinja2
{{ [user.email for user in users] }}                    <!-- Extract emails -->
{{ [post.title for post in posts if post.featured] }}   <!-- Extract featured titles -->
{{ [order.total for order in orders]|sum }}             <!-- Sum order totals -->
```

### Data Transformation
```jinja2
{{ [name.title() for name in names] }}                  <!-- Title case names -->
{{ [price * 1.1 for price in prices] }}                 <!-- Add 10% to prices -->
{{ [{"id": item.id, "name": item.name} for item in items] }} <!-- Create new structure -->
```

### Data Filtering
```jinja2
{{ [user for user in users if user.age >= 18] }}        <!-- Adults only -->
{{ [post for post in posts if "python" in post.tags] }} <!-- Posts with Python tag -->
{{ [item for item in items if item.price < 100] }}      <!-- Affordable items -->
```

### Dictionary Operations
```jinja2
{{ {user.id: user.name for user in users} }}            <!-- ID to name mapping -->
{{ {k.upper(): v for k, v in data.items()} }}           <!-- Uppercase keys -->
{{ {k: v for k, v in settings.items() if v is not none} }} <!-- Remove null values -->
```