# List and Dictionary Comprehensions

List and dictionary comprehensions provide a concise way to create new collections by iterating over existing ones and applying transformations or filtering conditions. This feature is fully implemented in the Miya Engine template engine.

## Table of Contents

- [Overview](#overview)
- [List Comprehensions](#list-comprehensions)
- [Dictionary Comprehensions](#dictionary-comprehensions)
- [Conditional Filtering](#conditional-filtering)
- [Complex Expressions](#complex-expressions)
- [Nested Comprehensions](#nested-comprehensions)
- [Performance Considerations](#performance-considerations)
- [Examples](#examples)
- [Limitations](#limitations)

## Overview

Comprehensions allow you to create new lists or dictionaries by:
1. Iterating over an existing iterable (list, dictionary, etc.)
2. Applying a transformation expression to each element
3. Optionally filtering elements with a condition

**Basic Syntax:**
```jinja2
{{ [expression for variable in iterable] }}
{{ [expression for variable in iterable if condition] }}
{{ {key_expr: value_expr for variable in iterable} }}
{{ {key_expr: value_expr for variable in iterable if condition} }}
```

## List Comprehensions

### Basic List Comprehension

Transform each element in a list:

```jinja2
<!-- Template -->
{{ [x * 2 for x in numbers] }}

<!-- Context: numbers = [1, 2, 3, 4] -->
<!-- Output: [2, 4, 6, 8] -->
```

### String Transformations

Apply string operations:

```jinja2
<!-- Template -->
{{ [name.upper() for name in users] }}

<!-- Context: users = ["alice", "bob", "charlie"] -->
<!-- Output: ["ALICE", "BOB", "CHARLIE"] -->
```

### Accessing Object Properties

Extract properties from objects:

```jinja2
<!-- Template -->
{{ [user.name for user in users] }}

<!-- Context: users = [{"name": "Alice", "age": 25}, {"name": "Bob", "age": 30}] -->
<!-- Output: ["Alice", "Bob"] -->
```

## Dictionary Comprehensions

### Basic Dictionary Comprehension

Create dictionaries from iterables:

```jinja2
<!-- Template -->
{{ {user.name: user.age for user in users} }}

<!-- Context: users = [{"name": "Alice", "age": 25}, {"name": "Bob", "age": 30}] -->
<!-- Output: {"Alice": 25, "Bob": 30} -->
```

### Key Transformations

Transform keys while preserving values:

```jinja2
<!-- Template -->
{{ {key.upper(): value for key, value in items.items()} }}

<!-- Context: items = {"name": "Alice", "role": "admin"} -->
<!-- Output: {"NAME": "Alice", "ROLE": "admin"} -->
```

### Value Transformations

Transform values while preserving keys:

```jinja2
<!-- Template -->
{{ {key: value * 2 for key, value in scores.items()} }}

<!-- Context: scores = {"alice": 85, "bob": 92, "charlie": 78} -->
<!-- Output: {"alice": 170, "bob": 184, "charlie": 156} -->
```

## Conditional Filtering

### List Comprehensions with Conditions

Filter elements based on conditions:

```jinja2
<!-- Template -->
{{ [x for x in numbers if x > 5] }}

<!-- Context: numbers = [1, 6, 3, 8, 2, 9] -->
<!-- Output: [6, 8, 9] -->
```

### Complex Filtering Conditions

Use complex boolean expressions:

```jinja2
<!-- Template -->
{{ [user.name for user in users if user.age >= 18 and user.active] }}

<!-- Context: users = [
  {"name": "Alice", "age": 25, "active": true},
  {"name": "Bob", "age": 16, "active": true},
  {"name": "Charlie", "age": 30, "active": false}
] -->
<!-- Output: ["Alice"] -->
```

### Dictionary Comprehensions with Conditions

Filter dictionary entries:

```jinja2
<!-- Template -->
{{ {key: value for key, value in scores.items() if value > 80} }}

<!-- Context: scores = {"alice": 85, "bob": 75, "charlie": 92} -->
<!-- Output: {"alice": 85, "charlie": 92} -->
```

## Complex Expressions

### Mathematical Operations

Perform calculations in comprehensions:

```jinja2
<!-- Template -->
{{ [x**2 + 1 for x in range(1, 5)] }}

<!-- Context: range function available -->
<!-- Output: [2, 5, 10, 17] -->
```

### String Formatting

Format strings within comprehensions:

```jinja2
<!-- Template -->
{{ [user.name + " (" + user.role + ")" for user in users] }}

<!-- Context: users = [{"name": "Alice", "role": "admin"}, {"name": "Bob", "role": "user"}] -->
<!-- Output: ["Alice (admin)", "Bob (user)"] -->
```

### Filter Applications

Apply filters within comprehensions:

```jinja2
<!-- Template -->
{{ [name|title for name in names] }}

<!-- Context: names = ["alice", "bob", "charlie"] -->
<!-- Output: ["Alice", "Bob", "Charlie"] -->
```

## Nested Comprehensions

### List of Lists

Create nested list structures:

```jinja2
<!-- Template -->
{{ [[x * y for x in row] for y in [1, 2, 3]] }}

<!-- Context: row = [1, 2, 3] -->
<!-- Output: [[1, 2, 3], [2, 4, 6], [3, 6, 9]] -->
```

### Flattening Nested Lists

Flatten nested structures:

```jinja2
<!-- Template -->
{{ [item for sublist in nested_list for item in sublist] }}

<!-- Context: nested_list = [[1, 2], [3, 4], [5, 6]] -->
<!-- Output: [1, 2, 3, 4, 5, 6] -->
```

## Performance Considerations

### Memory Usage

Comprehensions create new collections in memory:

```jinja2
<!-- Efficient: Direct iteration -->
{% for user in users %}
  {{ user.name }}
{% endfor %}

<!-- Less efficient: Creates intermediate list -->
{% for name in [user.name for user in users] %}
  {{ name }}
{% endfor %}
```

### Complex Filtering

For complex filtering, consider using filters instead:

```jinja2
<!-- Comprehension approach -->
{{ [user for user in users if user.age > 18 and user.department == "engineering"] }}

<!-- Filter approach (may be more readable) -->
{{ users|selectattr("age", "gt", 18)|selectattr("department", "eq", "engineering")|list }}
```

## Examples

### Real-world Use Cases

#### 1. Navigation Menu Generation

```jinja2
<!-- Template -->
<nav>
  {% for item in [{"url": page.url, "title": page.title, "active": page.slug == current_page} for page in pages if page.visible] %}
    <a href="{{ item.url }}" {% if item.active %}class="active"{% endif %}>
      {{ item.title }}
    </a>
  {% endfor %}
</nav>
```

#### 2. Form Field Processing

```jinja2
<!-- Template -->
{% set required_fields = [field.name for field in form.fields if field.required] %}
{% set field_errors = {field.name: field.errors for field in form.fields if field.errors} %}

<form>
  {% for field in form.fields %}
    <div class="field{% if field.name in field_errors %} error{% endif %}">
      <label for="{{ field.name }}">
        {{ field.label }}
        {% if field.name in required_fields %}<span class="required">*</span>{% endif %}
      </label>
      {{ field.render() }}
      {% if field.name in field_errors %}
        <div class="error-message">{{ field_errors[field.name]|join(", ") }}</div>
      {% endif %}
    </div>
  {% endfor %}
</form>
```

#### 3. Data Aggregation

```jinja2
<!-- Template -->
{% set total_sales = [sale.amount for sale in sales if sale.status == "completed"]|sum %}
{% set sales_by_category = {category: [sale.amount for sale in sales if sale.category == category]|sum for category in categories} %}

<div class="dashboard">
  <h2>Total Sales: ${{ total_sales }}</h2>
  
  <div class="category-breakdown">
    {% for category, amount in sales_by_category.items() %}
      <div class="category">
        <span>{{ category|title }}</span>
        <span>${{ amount }}</span>
      </div>
    {% endfor %}
  </div>
</div>
```

#### 4. Tag Cloud Generation

```jinja2
<!-- Template -->
{% set tag_counts = {tag.name: tag.count for tag in tags if tag.count > 0} %}
{% set max_count = tag_counts.values()|max %}

<div class="tag-cloud">
  {% for tag_name, count in tag_counts.items() %}
    {% set size = (count / max_count * 100)|round %}
    <span class="tag" style="font-size: {{ size }}%">
      {{ tag_name }}
      <small>({{ count }})</small>
    </span>
  {% endfor %}
</div>
```

## Limitations

### Current Limitations

1. **Variable Scope**: Comprehension variables don't leak into the outer scope
2. **Error Handling**: Errors in comprehension expressions can be harder to debug
3. **Memory Usage**: Large comprehensions can consume significant memory

### Not Supported

1. **Nested Variable Assignment**: Cannot assign to variables within comprehensions
2. **Multiple Iterables**: Single iterable per comprehension (no zip-like functionality)
3. **Generator Expressions**: Always creates full collections, not lazy generators

### Error Examples

```jinja2
<!-- This will cause an error -->
{{ [x for x in undefined_variable] }}

<!-- This will also cause an error -->
{{ {key: value for item in items} }}  <!-- key and value not defined -->
```

## Best Practices

### 1. Keep Comprehensions Simple

```jinja2
<!-- Good: Simple and readable -->
{{ [user.name for user in users] }}

<!-- Avoid: Too complex -->
{{ [user.profile.personal_info.display_name or user.username or "Anonymous" for user in users if user.is_active and not user.is_banned and user.profile] }}
```

### 2. Use Meaningful Variable Names

```jinja2
<!-- Good -->
{{ [product.name for product in products if product.in_stock] }}

<!-- Avoid -->
{{ [x.name for x in products if x.in_stock] }}
```

### 3. Consider Readability vs. Conciseness

```jinja2
<!-- Sometimes a traditional loop is more readable -->
{% set active_user_names = [] %}
{% for user in users %}
  {% if user.is_active and user.profile %}
    {% set _ = active_user_names.append(user.profile.display_name or user.username) %}
  {% endif %}
{% endfor %}

<!-- Versus -->
{{ [user.profile.display_name or user.username for user in users if user.is_active and user.profile] }}
```

### 4. Use Filters for Common Operations

```jinja2
<!-- Instead of -->
{{ [user for user in users if user.age >= 18] }}

<!-- Consider -->
{{ users|selectattr("age", "ge", 18)|list }}
```

---

**Note**: List and dictionary comprehensions are a powerful feature that can make templates more concise and expressive. However, they should be used judiciously to maintain template readability and performance.