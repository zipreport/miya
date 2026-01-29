# Filters Guide

Filters transform variables and expressions in templates. Miya Engine supports 70+ built-in filters with Jinja2 compatibility.

> **Working Example:** See `examples/features/filters/` for complete examples.

---

## Table of Contents

1. [Basic Usage](#basic-usage)
2. [String Filters](#string-filters)
3. [Collection Filters](#collection-filters)
4. [Numeric Filters](#numeric-filters)
5. [HTML & Security Filters](#html-security-filters)
6. [Utility Filters](#utility-filters)
7. [Filter Chaining](#filter-chaining)
8. [Limitations](#limitations)

---

## Basic Usage

### Syntax

```jinja2
{{ variable|filter }}
{{ variable|filter(arg1, arg2) }}
{{ variable|filter1|filter2|filter3 }}
```

### Examples

```jinja2
{{ "hello"|upper }}                    → HELLO
{{ 3.14159|round(2) }}                 → 3.14
{{ items|join(", ") }}                 → item1, item2, item3
{{ text|truncate(20) }}                → Truncated text...
{{ numbers|length }}                   → 5
```

---

## String Filters

### Case Conversion

| Filter | Example | Result |
|--------|---------|--------|
| `upper` | `{{"hello"\|upper}}` | `HELLO` |
| `lower` | `{{"HELLO"\|lower}}` | `hello` |
| `capitalize` | `{{"hello world"\|capitalize}}` | `Hello world` |
| `title` | `{{"hello world"\|title}}` | `Hello World` |

```jinja2
{{ "miya engine"|upper }}              → MIYA ENGINE
{{ "MIYA ENGINE"|lower }}              → miya engine
{{ "hello world"|title }}              → Hello World
{{ "miya engine"|capitalize }}         → Miya engine
```

### String Manipulation

| Filter | Description | Example |
|--------|-------------|---------|
| `trim` | Remove whitespace | `{{"  text  "\|trim}}` → `text` |
| `replace` | Replace substring | `{{"hello"\|replace("l", "L")}}` → `heLLo` |
| `truncate` | Truncate to length | `{{"long text"\|truncate(5)}}` → `lo...` |
| `center` | Center in width | `{{"x"\|center(5, "-")}}` → `--x--` |
| `slugify` | URL-safe slug | `{{"Hello World!"\|slugify}}` → `hello-world` |

**Examples:**
```jinja2
{{ "   hello world   "|trim }}         → "hello world"
{{ "Hello World"|replace("World", "Miya") }}  → "Hello Miya"
{{ long_text|truncate(30) }}           → "This is a very long text..."
{{ "Title"|center(20, "-") }}          → "-------Title--------"
{{ "Product Name!"|slugify }}          → "product-name"
```

### String Analysis

| Filter | Description | Example |
|--------|-------------|---------|
| `wordcount` | Count words | `{{"hello world"\|wordcount}}` → `2` |
| `startswith` | Check start | `{{"hello"\|startswith("he")}}` → `true` |
| `endswith` | Check end | `{{"hello"\|endswith("lo")}}` → `true` |
| `contains` | Check substring | `{{"hello"\|contains("ell")}}` → `true` |
| `split` | Split to list | `{{"a,b,c"\|split(",")}}` → `["a","b","c"]` |

**Examples:**
```jinja2
{{ "The quick brown fox"|wordcount }}  → 4
{{ "filename.txt"|endswith(".txt") }}  → true
{{ "hello world"|contains("world") }}  → true
{{ "apple,banana,cherry"|split(",") }} → ["apple", "banana", "cherry"]
```

---

## Collection Filters

###  Important Limitations

Some collection filters have limited support in Miya. **Reliable filters:**

-  `length`, `first`, `last`, `join`
-  `list` (convert to list)
-  `sort`, `reverse`, `unique`, `slice`, `batch` - May have issues

### Basic Collection Filters

| Filter | Description | Example |
|--------|-------------|---------|
| `first` | Get first item | `{{[1,2,3]\|first}}` → `1` |
| `last` | Get last item | `{{[1,2,3]\|last}}` → `3` |
| `length` | Get count | `{{[1,2,3]\|length}}` → `3` |
| `join` | Join to string | `{{["a","b"]\|join(",")}}` → `"a,b"` |
| `list` | Convert to list | `{{range(5)\|list}}` → `[0,1,2,3,4]` |

**Examples:**
```jinja2
{{ numbers|first }}                    → 1
{{ numbers|last }}                     → 10
{{ items|length }}                     → 5
{{ fruits|join(", ") }}                → "apple, banana, cherry"
{{ range(5)|list }}                    → [0, 1, 2, 3, 4]
```

### Attribute Extraction

| Filter | Description | Example |
|--------|-------------|---------|
| `map(attribute)` | Extract attribute | `{{users\|map(attribute="name")}}` |
| `selectattr` | Filter by attribute | `{{users\|selectattr("active")}}` |
| `rejectattr` | Reject by attribute | `{{users\|rejectattr("active")}}` |

**Examples:**
```jinja2
{# Extract all user names #}
{{ users|map(attribute="name")|join(", ") }}
→ "Alice, Bob, Charlie"

{# Get only active users #}
{{ users|selectattr("active")|list }}

{# Get inactive users #}
{{ users|rejectattr("active")|list }}
```

---

## Numeric Filters

| Filter | Description | Example |
|--------|-------------|---------|
| `abs` | Absolute value | `{{-42\|abs}}` → `42` |
| `round` | Round number | `{{3.14159\|round(2)}}` → `3.14` |
| `int` | Convert to int | `{{"123"\|int}}` → `123` |
| `float` | Convert to float | `{{"99.99"\|float}}` → `99.99` |
| `pow` | Power operation | `{{2\|pow(8)}}` → `256` |

**Examples:**
```jinja2
{{ -42|abs }}                          → 42
{{ 3.14159|round(2) }}                 → 3.14
{{ "123"|int + 7 }}                    → 130
{{ "99.99"|float }}                    → 99.99
{{ 2|pow(8) }}                         → 256
```

### Aggregate Functions

 **Note:** `sum`, `min`, `max` may have limited support. Test in your use case.

```jinja2
{# May work in some contexts #}
{{ numbers|sum }}
{{ numbers|min }}
{{ numbers|max }}
```

---

## HTML & Security Filters

### Essential Security Filters

| Filter | Description | Example |
|--------|-------------|---------|
| `escape` | Escape HTML | `{{"<script>"\|escape}}` → `&lt;script&gt;` |
| `safe` | Mark as safe | `{{html\|safe}}` - Renders HTML |
| `striptags` | Remove tags | `{{"<b>text</b>"\|striptags}}` → `text` |
| `urlencode` | URL encode | `{{"hello world"\|urlencode}}` → `hello%20world` |

**Examples:**
```jinja2
{# Auto-escape user input #}
{{ user_input|escape }}
→ &lt;script&gt;alert('XSS')&lt;/script&gt;

{# Allow trusted HTML #}
{{ trusted_html|safe }}
→ <strong>Bold Text</strong>

{# Strip HTML tags #}
{{ "<p>Hello <b>World</b></p>"|striptags }}
→ "Hello World"

{# URL encoding #}
{{ "hello world!"|urlencode }}
→ "hello%20world%21"
```

### When to Use

- **`escape`**: Always escape untrusted user input
- **`safe`**: Only for content you control and trust
- **`striptags`**: Extract text from HTML
- **`urlencode`**: Encode URL parameters

---

## Utility Filters

### Default Values

```jinja2
{{ undefined_var|default("fallback") }}     → "fallback"
{{ none_var|default("default") }}           → "default"
{{ ""|default("empty string") }}            → "empty string"

{# Alias: d #}
{{ var|d("default") }}                      → "default"
```

### Data Formatting

| Filter | Description | Example |
|--------|-------------|---------|
| `format` | String formatting | `{{"Hello {0}"\|format("World")}}` |
| `tojson` | Convert to JSON | `{{data\|tojson}}` |
| `filesizeformat` | Format file size | `{{1536\|filesizeformat}}` → `1.5 KB` |

**Examples:**
```jinja2
{{ "Hello {0}, you have {1} messages"|format(name, count) }}
→ "Hello Alice, you have 5 messages"

{{ {"id": 123, "name": "Product"}|tojson }}
→ '{"id":123,"name":"Product"}'

{{ 1536|filesizeformat }}
→ "1.5 KB"

{{ 1048576|filesizeformat }}
→ "1.0 MB"
```

---

## Filter Chaining

### Multiple Filters

Apply multiple filters in sequence:

```jinja2
{{ "  hello world  "|trim|upper|replace("WORLD", "MIYA") }}
→ "HELLO MIYA"

{{ long_text|truncate(30)|title }}
→ "This Is A Very Long Tex..."

{{ users|selectattr("active")|map(attribute="name")|join(", ")|upper }}
→ "ALICE, CHARLIE, DIANA"
```

### Order Matters

```jinja2
{# Different results based on order #}
{{ "hello"|upper|replace("L", "X") }}      → "HEXXO"
{{ "hello"|replace("l", "X")|upper }}      → "HEXXO"
```

### Chaining with Arguments

```jinja2
{{ value|default("N/A")|upper|center(20, "-") }}
→ "--------N/A---------"

{{ prices|map(attribute="amount")|sum|round(2) }}
→ 299.98
```

---

## Practical Examples

### Example 1: User Display Names

```jinja2
{# Format user names #}
{% for user in users %}
  {{ user.name|title }} ({{ user.role|capitalize }})
{% endfor %}

{# Extract and join active user emails #}
{{ users|selectattr("active")|map(attribute="email")|join(", ") }}
```

### Example 2: Price Formatting

```jinja2
{# Format currency #}
${{ price|round(2) }}

{# Apply discount #}
Original: ${{ price|round(2) }}
Discounted: ${{ (price * 0.9)|round(2) }}
```

### Example 3: Safe HTML Rendering

```jinja2
{# Escape user content #}
<div class="comment">
  {{ user_comment|escape }}
</div>

{# Render trusted HTML #}
<div class="content">
  {{ article_html|safe }}
</div>
```

### Example 4: List Processing

```jinja2
{# Join list items #}
Tags: {{ tags|join(", ") }}

{# Get active user count #}
Active Users: {{ users|selectattr("active")|list|length }}

{# First and last items #}
First: {{ items|first }}, Last: {{ items|last }}
```

---

## Limitations

###  Filters with Known Issues

Based on testing, these filters may not work reliably:

```jinja2
{# May fail with "requires a sequence" error #}
{{ numbers|sort }}
{{ numbers|reverse }}
{{ items|unique }}
{{ numbers|slice(3) }}
{{ range(10)|batch(3) }}
```

###  Working Alternatives

Instead of problematic filters, use:

```jinja2
{# Use in loops #}
{% for num in numbers %}{{ num }}{% endfor %}

{# Use filter chains #}
{{ items|selectattr("active")|map(attribute="name")|join(", ") }}

{# Convert to list #}
{{ range(10)|list }}
```

### Comprehension Filters Don't Work

```jinja2
{#  This doesn't work #}
{{ [x for x in numbers if x > 5] }}

{#  Use filter chains instead #}
{{ numbers|select("greaterthan", 5)|list }}
```

---

## Complete Filter Reference

### String Filters (16+)
 `upper`, `lower`, `capitalize`, `title`, `trim`, `replace`, `truncate`, `center`, `wordcount`, `split`, `startswith`, `endswith`, `contains`, `slugify`, `indent`, `wordwrap`

### Collection Filters (8)
 `first`, `last`, `length`, `join`, `list`, `map`, `selectattr`, `rejectattr`

 `sort`, `reverse`, `unique`, `slice`, `batch` - Limited support

### Numeric Filters (5)
 `abs`, `round`, `int`, `float`, `pow`

 `sum`, `min`, `max` - May have issues

### HTML/Security Filters (4)
 `escape`, `safe`, `striptags`, `urlencode`

### Utility Filters (5)
 `default`, `format`, `tojson`, `filesizeformat`, `dictsort`

---

## See Also

- [Working Example](https://github.com/zipreport/miya/tree/master/examples/features/filters/) - Complete filter examples
- [MIYA_LIMITATIONS.md](MIYA_LIMITATIONS.md) - Known filter limitations
- [Filter Blocks](FILTER_BLOCKS_IMPLEMENTATION.md) - Apply filters to entire blocks

---

## Summary

**70+ filters available** with most Jinja2 filters working identically.

**Reliably Working:**
- All string filters (upper, lower, trim, replace, etc.)
- Basic collection filters (first, last, length, join)
- Numeric filters (abs, round, int, float)
- Security filters (escape, safe, striptags)
- Utility filters (default, format, tojson)

**Use with Caution:**
- Collection transformation filters (sort, reverse, unique)
- Aggregate functions (sum, min, max)

**Test your specific use case** - run `examples/features/filters/` to see what works.
