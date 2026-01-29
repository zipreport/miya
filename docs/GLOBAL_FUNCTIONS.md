# Global Functions Guide

Global functions are built-in functions available in all templates without imports. Miya Engine provides 9 essential global functions with 100% Jinja2 compatibility.

> **Working Example:** See `examples/features/global-functions/` for complete examples.

---

## Table of Contents

1. [range() - Number Sequences](#range-number-sequences)
2. [dict() - Dictionary Constructor](#dict-dictionary-constructor)
3. [cycler() - Cycle Through Values](#cycler-cycle-through-values)
4. [joiner() - Smart Joining](#joiner-smart-joining)
5. [namespace() - Mutable Container](#namespace-mutable-container)
6. [lipsum() - Lorem Ipsum Generator](#lipsum-lorem-ipsum-generator)
7. [zip() - Combine Sequences](#zip-combine-sequences)
8. [enumerate() - Index with Values](#enumerate-index-with-values)
9. [url_for() - URL Generation](#url_for-url-generation)

---

## range() - Number Sequences

### Basic Usage

Generate sequences of numbers:

```jinja2
{# range(n) - generates 0 to n-1 #}
{{ range(5)|list }}
→ [0, 1, 2, 3, 4]

{# range(start, end) - generates start to end-1 #}
{{ range(5, 10)|list }}
→ [5, 6, 7, 8, 9]

{# range(start, end, step) - with custom step #}
{{ range(0, 20, 3)|list }}
→ [0, 3, 6, 9, 12, 15, 18]
```

### Reverse Ranges

Use negative step for countdown:

```jinja2
{{ range(10, 0, -1)|list }}
→ [10, 9, 8, 7, 6, 5, 4, 3, 2, 1]

{{ range(20, 0, -5)|list }}
→ [20, 15, 10, 5]
```

### In Loops

Most common use case:

```jinja2
{# Generate 10 items #}
{% for i in range(10) %}
  <div>Item {{ i }}</div>
{% endfor %}

{# Generate numbered pages #}
{% for page in range(1, 11) %}
  <a href="/page/{{ page }}">Page {{ page }}</a>
{% endfor %}
```

### Practical Examples

**Grid Layout:**
```jinja2
<table>
{% for row in range(5) %}
  <tr>
  {% for col in range(10) %}
    <td>{{ row }},{{ col }}</td>
  {% endfor %}
  </tr>
{% endfor %}
</table>
```

**Pagination:**
```jinja2
<div class="pagination">
{% for page_num in range(1, total_pages + 1) %}
  <a href="?page={{ page_num }}"
     class="{{ 'active' if page_num == current_page else '' }}">
    {{ page_num }}
  </a>
{% endfor %}
</div>
```

---

## dict() - Dictionary Constructor

### Create Dictionaries

Build dictionaries in templates:

```jinja2
{# With keyword arguments #}
{% set user = dict(name="Alice", age=30, role="admin") %}
{{ user.name }}  → Alice
{{ user.age }}   → 30

{# Access with brackets #}
{{ user["role"] }}  → admin
```

### From Key-Value Pairs

```jinja2
{% set config = dict([("key1", "value1"), ("key2", "value2")]) %}
{{ config.key1 }}  → value1
```

### Dynamic Keys

```jinja2
{% set data = dict() %}
{% for i in range(3) %}
  {% set _ = data.update({"item_" ~ i: i * 10}) %}
{% endfor %}
```

### Practical Examples

**User Profile:**
```jinja2
{% set profile = dict(
  username="alice",
  email="alice@example.com",
  verified=True,
  created="2024-01-01"
) %}

<div class="profile">
  <h3>{{ profile.username }}</h3>
  <p>{{ profile.email }}</p>
  {% if profile.verified %}
    <span class="badge">Verified</span>
  {% endif %}
</div>
```

---

## cycler() - Cycle Through Values

### Basic Usage

Cycle through values automatically:

```jinja2
{% set row_class = cycler("odd", "even") %}

<table>
{% for item in items %}
  <tr class="{{ row_class.next() }}">
    <td>{{ item }}</td>
  </tr>
{% endfor %}
</table>
```

**Output:**
```html
<tr class="odd"><td>Item 1</td></tr>
<tr class="even"><td>Item 2</td></tr>
<tr class="odd"><td>Item 3</td></tr>
<tr class="even"><td>Item 4</td></tr>
```

### Multiple Values

```jinja2
{% set colors = cycler("red", "green", "blue") %}

{% for i in range(9) %}
  <div style="color: {{ colors.next() }}">{{ i }}</div>
{% endfor %}
```

Cycles through: red, green, blue, red, green, blue, red, green, blue

### Practical Examples

**Alternating Row Colors:**
```jinja2
{% set row_class = cycler("bg-light", "bg-white") %}

<table>
  <thead>
    <tr><th>Product</th><th>Price</th></tr>
  </thead>
  <tbody>
  {% for product in products %}
    <tr class="{{ row_class.next() }}">
      <td>{{ product.name }}</td>
      <td>${{ product.price }}</td>
    </tr>
  {% endfor %}
  </tbody>
</table>
```

**Status Indicators:**
```jinja2
{% set status_cycle = cycler("pending", "processing", "completed") %}

{% for task in tasks %}
  <div class="task status-{{ status_cycle.next() }}">
    {{ task.name }}
  </div>
{% endfor %}
```

---

## joiner() - Smart Joining

### Basic Usage

Automatically adds separator between items (but not before first):

```jinja2
{% set comma = joiner(", ") %}

{% for item in items %}
  {{ comma() }}{{ item }}
{% endfor %}
```

**Output:** `item1, item2, item3`

### How It Works

```jinja2
{% set sep = joiner("|") %}

{{ sep() }}  → ""       {# First call returns empty string #}
{{ sep() }}  → "|"      {# Subsequent calls return separator #}
{{ sep() }}  → "|"
```

### Practical Examples

**Breadcrumb Navigation:**
```jinja2
{% set separator = joiner(" > ") %}

<nav>
{% for crumb in breadcrumbs %}
  {{ separator() }}<a href="{{ crumb.url }}">{{ crumb.title }}</a>
{% endfor %}
</nav>
```

**Output:** `Home > Products > Electronics > Laptops`

**Tag List:**
```jinja2
{% set comma = joiner(", ") %}

<p>Tags:
{% for tag in tags %}
  {{ comma() }}<span class="tag">{{ tag }}</span>
{% endfor %}
</p>
```

**Conditional Lists:**
```jinja2
{% set sep = joiner(" | ") %}

<p>Actions:
{% if user.can_edit %}
  {{ sep() }}<a href="/edit">Edit</a>
{% endif %}
{% if user.can_delete %}
  {{ sep() }}<a href="/delete">Delete</a>
{% endif %}
{% if user.can_share %}
  {{ sep() }}<a href="/share">Share</a>
{% endif %}
</p>
```

Clean output even when some conditions are false!

---

## namespace() - Mutable Container

### Basic Usage

Create mutable objects for loop scoping:

```jinja2
{% set ns = namespace(count=0, total=0) %}

{% for item in items %}
  {% set ns.count = ns.count + 1 %}
  {% set ns.total = ns.total + item.price %}
{% endfor %}

<p>Items: {{ ns.count }}, Total: ${{ ns.total }}</p>
```

### Why Use namespace()?

Template variables are **immutable** in loop scope:

```jinja2
{#  This doesn't work as expected #}
{% set count = 0 %}
{% for item in items %}
  {% set count = count + 1 %}  {# Creates new local variable #}
{% endfor %}
{{ count }}  → 0  {# Original value unchanged #}

{#  Use namespace instead #}
{% set ns = namespace(count=0) %}
{% for item in items %}
  {% set ns.count = ns.count + 1 %}  {# Modifies namespace attribute #}
{% endfor %}
{{ ns.count }}  → 10  {# Correct value #}
```

### Practical Examples

**Counting:**
```jinja2
{% set stats = namespace(active=0, inactive=0, total=0) %}

{% for user in users %}
  {% set stats.total = stats.total + 1 %}
  {% if user.active %}
    {% set stats.active = stats.active + 1 %}
  {% else %}
    {% set stats.inactive = stats.inactive + 1 %}
  {% endif %}
{% endfor %}

<p>Total: {{ stats.total }}</p>
<p>Active: {{ stats.active }}</p>
<p>Inactive: {{ stats.inactive }}</p>
```

**Finding First Match:**
```jinja2
{% set result = namespace(found=False, match=None) %}

{% for item in items %}
  {% if not result.found and item.price < 50 %}
    {% set result.found = True %}
    {% set result.match = item %}
  {% endif %}
{% endfor %}

{% if result.found %}
  <p>First affordable item: {{ result.match.name }}</p>
{% endif %}
```

**Accumulating Results:**
```jinja2
{% set cart = namespace(subtotal=0, tax=0, total=0) %}

{% for item in cart_items %}
  {% set item_total = item.price * item.quantity %}
  {% set cart.subtotal = cart.subtotal + item_total %}
{% endfor %}

{% set cart.tax = cart.subtotal * 0.1 %}
{% set cart.total = cart.subtotal + cart.tax %}

<p>Subtotal: ${{ cart.subtotal|round(2) }}</p>
<p>Tax (10%): ${{ cart.tax|round(2) }}</p>
<p>Total: ${{ cart.total|round(2) }}</p>
```

---

## lipsum() - Lorem Ipsum Generator

### Basic Usage

Generate placeholder text:

```jinja2
{# Default - one paragraph #}
{{ lipsum() }}

{# Multiple paragraphs #}
{{ lipsum(n=3) }}

{# HTML paragraphs #}
{{ lipsum(n=3, html=True) }}
→ <p>Lorem ipsum...</p><p>Dolor sit...</p><p>Amet...</p>
```

### Practical Examples

**Mockup Pages:**
```jinja2
<article>
  <h1>{{ title|default("Article Title") }}</h1>
  <p class="lead">{{ lipsum(n=1) }}</p>
  {{ lipsum(n=5, html=True) }}
</article>
```

**Placeholder Content:**
```jinja2
{% if product.description %}
  <p>{{ product.description }}</p>
{% else %}
  <p class="text-muted">{{ lipsum(n=1) }}</p>
{% endif %}
```

---

## zip() - Combine Sequences

### Basic Usage

Combine multiple iterables:

```jinja2
{% set names = ["Alice", "Bob", "Charlie"] %}
{% set ages = [30, 25, 35] %}

{% for name, age in zip(names, ages) %}
  <p>{{ name }} is {{ age }} years old</p>
{% endfor %}
```

**Output:**
```
Alice is 30 years old
Bob is 25 years old
Charlie is 35 years old
```

### Multiple Sequences

```jinja2
{% set products = ["Laptop", "Mouse", "Keyboard"] %}
{% set prices = [999, 29, 79] %}
{% set stock = [15, 50, 30] %}

<table>
  <tr><th>Product</th><th>Price</th><th>Stock</th></tr>
{% for product, price, stock in zip(products, prices, stock) %}
  <tr>
    <td>{{ product }}</td>
    <td>${{ price }}</td>
    <td>{{ stock }}</td>
  </tr>
{% endfor %}
</table>
```

### Important Note

`zip()` stops at the shortest sequence:

```jinja2
{% set a = [1, 2, 3, 4, 5] %}
{% set b = ["a", "b", "c"] %}

{% for num, letter in zip(a, b) %}
  {{ num }}:{{ letter }}
{% endfor %}
```

**Output:** `1:a 2:b 3:c` (stops after 3 items)

### Practical Examples

**Parallel Data Display:**
```jinja2
{% for name, email, role in zip(names, emails, roles) %}
  <div class="user-card">
    <h3>{{ name }}</h3>
    <p>{{ email }}</p>
    <span class="badge">{{ role }}</span>
  </div>
{% endfor %}
```

---

## enumerate() - Index with Values

### Basic Usage

Get index and value together:

```jinja2
{% for index, item in enumerate(items) %}
  <p>{{ index }}: {{ item }}</p>
{% endfor %}
```

**Output:**
```
0: Apple
1: Banana
2: Cherry
```

### Custom Start Index

```jinja2
{% for index, item in enumerate(items, 1) %}
  <p>{{ index }}. {{ item }}</p>
{% endfor %}
```

**Output:**
```
1. Apple
2. Banana
3. Cherry
```

** Note:** Use positional argument, not `start=1` keyword.

### Practical Examples

**Numbered List:**
```jinja2
<ol>
{% for num, task in enumerate(tasks, 1) %}
  <li value="{{ num }}">{{ task.title }}</li>
{% endfor %}
</ol>
```

**Table with Row Numbers:**
```jinja2
<table>
  <tr><th>#</th><th>Name</th><th>Email</th></tr>
{% for i, user in enumerate(users, 1) %}
  <tr>
    <td>{{ i }}</td>
    <td>{{ user.name }}</td>
    <td>{{ user.email }}</td>
  </tr>
{% endfor %}
</table>
```

**Alternative Styling:**
```jinja2
{% for index, product in enumerate(products) %}
  <div class="{{ 'even' if index % 2 == 0 else 'odd' }}">
    {{ product.name }}
  </div>
{% endfor %}
```

---

## url_for() - URL Generation

### Basic Usage

Generate URLs from endpoint names:

```jinja2
<a href="{{ url_for('home') }}">Home</a>
<a href="{{ url_for('profile', user_id=123) }}">Profile</a>
<a href="{{ url_for('search', q='laptops', category='electronics') }}">Search</a>
```

### With Query Parameters

```jinja2
{{ url_for('products', page=2, sort='price') }}
→ /products?page=2&sort=price
```

### Framework Integration

**Note:** `url_for()` requires framework integration. Implementation depends on your web framework (Flask, Django, etc.). Miya Engine provides the template function, but URL routing is handled by your application.

**Example with context:**
```jinja2
{# Application provides url_for implementation #}
<nav>
  <a href="{{ url_for('index') }}">Home</a>
  <a href="{{ url_for('blog.list') }}">Blog</a>
  <a href="{{ url_for('blog.post', id=post.id) }}">Read More</a>
</nav>
```

---

## Practical Use Cases

### Use Case 1: Alternating Table Rows

```jinja2
{% set row_class = cycler("row-light", "row-dark") %}

<table>
{% for product in products %}
  <tr class="{{ row_class.next() }}">
    <td>{{ product.name }}</td>
    <td>${{ product.price }}</td>
  </tr>
{% endfor %}
</table>
```

### Use Case 2: Building Navigation

```jinja2
{% set separator = joiner(" > ") %}

<nav class="breadcrumbs">
{% for crumb in breadcrumbs %}
  {{ separator() }}
  {% if loop.last %}
    <span>{{ crumb.title }}</span>
  {% else %}
    <a href="{{ crumb.url }}">{{ crumb.title }}</a>
  {% endif %}
{% endfor %}
</nav>
```

### Use Case 3: Creating Grids

```jinja2
<div class="grid">
{% for row in range(5) %}
  <div class="grid-row">
  {% for col in range(8) %}
    <div class="grid-cell" data-pos="{{ row }},{{ col }}">
      {{ row * 8 + col + 1 }}
    </div>
  {% endfor %}
  </div>
{% endfor %}
</div>
```

### Use Case 4: Counting in Loops

```jinja2
{% set stats = namespace(active=0, total=0) %}

{% for user in users %}
  {% set stats.total = stats.total + 1 %}
  {% if user.active %}
    {% set stats.active = stats.active + 1 %}
  {% endif %}
{% endfor %}

<div class="stats">
  <p>Total Users: {{ stats.total }}</p>
  <p>Active Users: {{ stats.active }} ({{ (stats.active / stats.total * 100)|round }}%)</p>
</div>
```

### Use Case 5: Parallel Iteration

```jinja2
{% for name, age, email in zip(names, ages, emails) %}
  <div class="contact-card">
    <h3>{{ name }}</h3>
    <p>Age: {{ age }}</p>
    <p>Email: {{ email }}</p>
  </div>
{% endfor %}
```

### Use Case 6: Numbered Lists

```jinja2
<h3>Top 10 Products</h3>
<ol class="leaderboard">
{% for rank, product in enumerate(top_products, 1) %}
  <li class="rank-{{ rank }}">
    <span class="rank">{{ rank }}</span>
    <span class="name">{{ product.name }}</span>
    <span class="score">{{ product.rating }}</span>
  </li>
{% endfor %}
</ol>
```

---

## Complete Reference

| Function | Syntax | Description | Returns |
|----------|--------|-------------|---------|
| `range` | `range(n)` | Generate 0 to n-1 | Sequence |
| | `range(start, end)` | Generate start to end-1 | Sequence |
| | `range(start, end, step)` | With custom step | Sequence |
| `dict` | `dict(key=value, ...)` | Create dictionary | Dict |
| | `dict([(k, v), ...])` | From pairs | Dict |
| `cycler` | `cycler(val1, val2, ...)` | Create cycler | Cycler object |
| `joiner` | `joiner(sep)` | Create joiner | Joiner object |
| `namespace` | `namespace(attr=value, ...)` | Mutable container | Namespace object |
| `lipsum` | `lipsum()` | Default lorem ipsum | String |
| | `lipsum(n=N)` | N paragraphs | String |
| | `lipsum(n=N, html=True)` | HTML paragraphs | HTML string |
| `zip` | `zip(seq1, seq2, ...)` | Combine sequences | Iterator |
| `enumerate` | `enumerate(seq)` | Index from 0 | Iterator |
| | `enumerate(seq, start)` | Custom start | Iterator |
| `url_for` | `url_for(endpoint, **params)` | Generate URL | String |

---

## See Also

- [Working Example](https://github.com/zipreport/miya/tree/master/examples/features/global-functions/) - Complete function examples
- [Control Structures](CONTROL_STRUCTURES_GUIDE.md) - Using functions in loops
- [Filters Guide](FILTERS_GUIDE.md) - Data transformation
- [Miya Limitations](MIYA_LIMITATIONS.md) - Known limitations

---

## Summary

**9 Global Functions - 100% Jinja2 Compatible**

All global functions work identically to Jinja2:

 **Fully Supported:**
- `range()` - Number sequences for iteration
- `dict()` - Dictionary construction
- `cycler()` - Cycle through values
- `joiner()` - Smart separator insertion
- `namespace()` - Mutable container for loops
- `lipsum()` - Lorem ipsum generation
- `zip()` - Parallel iteration
- `enumerate()` - Index with values
- `url_for()` - URL generation (requires framework integration)

**Key Benefits:**
- Clean iteration patterns
- Easy counters and accumulators
- Professional placeholder content
- Elegant data combination

All functions are production-ready and perform identically to Jinja2.
