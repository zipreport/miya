# Advanced Features Guide

Advanced template features for professionals. Miya Engine provides filter blocks, do statements, whitespace control, raw blocks, and autoescape with 100% Jinja2 compatibility.

> **Working Example:** See `examples/features/advanced/` for complete examples.

---

## Table of Contents

1. [Filter Blocks](#filter-blocks)
2. [Do Statements](#do-statements)
3. [Whitespace Control](#whitespace-control)
4. [Raw Blocks](#raw-blocks)
5. [Autoescape Control](#autoescape-control)
6. [Environment Configuration](#environment-configuration)
7. [Performance & Memory Management](#performance--memory-management)
8. [Practical Examples](#practical-examples)

---

## Filter Blocks

### Basic Usage

Apply filters to entire blocks of content:

```jinja2
{% filter upper %}
This entire block will be uppercase
Including {{ variable }} variables
{% endfilter %}
```

**Output:**
```
THIS ENTIRE BLOCK WILL BE UPPERCASE
INCLUDING VALUE VARIABLES
```

### Single Filter

```jinja2
{% filter upper %}
  hello world
{% endfilter %}
→ HELLO WORLD

{% filter title %}
  this is an important announcement
{% endfilter %}
→ This Is An Important Announcement

{% filter trim %}
   extra whitespace removed
{% endfilter %}
→ extra whitespace removed
```

### Chained Filters

Apply multiple filters in sequence:

```jinja2
{% filter trim|upper|replace("WORLD", "MIYA") %}
   hello world
{% endfilter %}
→ HELLO MIYA

{% filter striptags|upper %}
<p>Remove <strong>HTML</strong> tags</p>
{% endfilter %}
→ REMOVE HTML TAGS
```

### Filter Blocks with Logic

Filters apply to the entire rendered output:

```jinja2
{% filter upper %}
User list:
{% for user in users %}
- {{ user.name }}
{% endfor %}
{% endfilter %}
```

**Output:**
```
USER LIST:
- ALICE
- BOB
- CHARLIE
```

### Nested Filter Blocks

Filter blocks can be nested:

```jinja2
{% filter upper %}
Outer block
{% filter trim %}
  inner block with trim
{% endfilter %}
End outer
{% endfilter %}
```

**Output:**
```
OUTER BLOCK
INNER BLOCK WITH TRIM
END OUTER
```

---

## Do Statements

### Basic Usage

Execute expressions without output:

```jinja2
{% do expression %}
```

The `do` statement evaluates an expression but doesn't print the result.

### Examples

```jinja2
{# Execute without output #}
{% do numbers|sum %}

{# Side effects only #}
{% do data.update({"key": "value"}) %}

{# With filters #}
{% do (100 * 5)|round %}
```

### When to Use

```jinja2
{# Trigger computation without displaying result #}
{% set calculation_result = 0 %}
{% do expensive_function(data) %}

{# Execute for side effects #}
{% do logger.log("Template rendered") %}
```

**Note:** In pure templates, `do` has limited use since templates shouldn't have side effects. Most useful when integrating with systems that track template execution.

---

## Whitespace Control

### The Problem

Templates can generate unwanted whitespace:

```jinja2
<ul>
{% for item in items %}
  <li>{{ item }}</li>
{% endfor %}
</ul>
```

**Output (with extra lines):**
```html
<ul>

  <li>Item 1</li>

  <li>Item 2</li>

</ul>
```

### The Solution: Minus Sign (-)

Use `-` to strip whitespace:

| Syntax | Description |
|--------|-------------|
| `{%-` | Strip whitespace before tag |
| `-%}` | Strip whitespace after tag |
| `{{-` | Strip whitespace before variable |
| `-}}` | Strip whitespace after variable |

### Left Whitespace Control

Remove whitespace before tag:

```jinja2
<ul>
{%- for item in items %}
  <li>{{ item }}</li>
{% endfor %}
</ul>
```

**Output:**
```html
<ul>  <li>Item 1</li>
  <li>Item 2</li>
</ul>
```

### Right Whitespace Control

Remove whitespace after tag:

```jinja2
<ul>
{% for item in items -%}
  <li>{{ item }}</li>
{% endfor %}
</ul>
```

### Both Sides Control

```jinja2
<ul>
{%- for item in items -%}
  <li>{{- item -}}</li>
{%- endfor -%}
</ul>
```

**Output:**
```html
<ul><li>Item1</li><li>Item2</li></ul>
```

### Practical Examples

**Clean Comma-Separated List:**
```jinja2
Items: {%- for item in items -%}
  {{ item }}{% if not loop.last %}, {% endif %}
{%- endfor %}
```

**Output:** `Items: item1, item2, item3`

**Compact Number List:**
```jinja2
Values:
{%- for num in numbers -%}
  {{- num -}}
  {%- if not loop.last -%},{%- endif -%}
{%- endfor %}
```

**Output:** `Values: 1,2,3,4,5`

**Clean HTML:**
```jinja2
<nav>
  {%- for link in links %}
  <a href="{{ link.url }}">{{ link.title }}</a>
  {%- if not loop.last %} | {% endif -%}
  {% endfor -%}
</nav>
```

**Output:**
```html
<nav><a href="/">Home</a> | <a href="/about">About</a> | <a href="/contact">Contact</a></nav>
```

---

## Raw Blocks

### Basic Usage

Prevent template processing for documentation or examples:

```jinja2
{% raw %}
  {% for item in items %}
    {{ item }}
  {% endfor %}
{% endraw %}
```

**Output (literal text):**
```
  {% for item in items %}
    {{ item }}
  {% endfor %}
```

### Use Cases

**1. Documenting Template Syntax:**
```jinja2
<p>To print a variable, use:</p>
<pre>
{% raw %}
{{ variable_name }}
{% endraw %}
</pre>
```

**2. Client-Side Templates:**
```jinja2
<script type="text/template" id="user-template">
{% raw %}
  <div class="user">
    <h3>{{ user.name }}</h3>
    <p>{{ user.email }}</p>
  </div>
{% endraw %}
</script>
```

The template inside `{% raw %}` won't be processed by Miya, allowing it to be used by client-side JavaScript.

**3. Code Examples:**
```jinja2
<h3>Template Example:</h3>
<pre><code>
{% raw %}
{% if user.active %}
  Welcome, {{ user.name }}!
{% endif %}
{% endraw %}
</code></pre>
```

---

## Autoescape Control

### What is Autoescaping?

Autoescaping automatically escapes HTML special characters to prevent XSS attacks:

```jinja2
{# With autoescape ON (default) #}
{{ "<script>alert('XSS')</script>" }}
→ &lt;script&gt;alert('XSS')&lt;/script&gt;

{# With autoescape OFF #}
{{ "<script>alert('XSS')</script>" }}
→ <script>alert('XSS')</script>  {# DANGEROUS! #}
```

### Autoescape Blocks

Control escaping for specific regions:

```jinja2
{# Turn OFF autoescape #}
{% autoescape false %}
  <p>{{ html_content }}</p>
{% endautoescape %}

{# Turn ON autoescape (explicit) #}
{% autoescape true %}
  <p>{{ user_input }}</p>
{% endautoescape %}
```

### Safe Filter

Mark specific strings as safe:

```jinja2
{# Mark as safe - won't be escaped #}
{{ trusted_html|safe }}

{# Force escaping #}
{{ html_string|escape }}
```

### Practical Examples

**Rendering Trusted HTML:**
```jinja2
{# Article body is pre-sanitized HTML #}
{% autoescape false %}
<div class="article-body">
  {{ article.body }}
</div>
{% endautoescape %}

{# Or use safe filter #}
<div class="article-body">
  {{ article.body|safe }}
</div>
```

**Protecting User Input:**
```jinja2
{# Always escape user-provided content #}
{% autoescape true %}
<div class="comment">
  <strong>{{ comment.author }}</strong>
  <p>{{ comment.text }}</p>
</div>
{% endautoescape %}

{# Default behavior is autoescape ON #}
<div class="comment">
  {{ user_comment }}  {# Automatically escaped #}
</div>
```

**Mixed Content:**
```jinja2
{# Escape user content, allow trusted HTML #}
<div class="post">
  <h2>{{ post.title }}</h2>  {# Escaped #}

  {% autoescape false %}
    {{ post.body|safe }}  {# Trusted HTML #}
  {% endautoescape %}

  <div class="comments">
    {% for comment in comments %}
      <p>{{ comment.text }}</p>  {# Escaped again #}
    {% endfor %}
  </div>
</div>
```

### Security Best Practices

```jinja2
{#  SAFE - user input is escaped #}
<p>{{ user.name }}</p>
<p>{{ user.comment }}</p>

{#  SAFE - trusted content explicitly marked #}
<div>{{ article.html|safe }}</div>

{#  DANGEROUS - disabling autoescape for user input #}
{% autoescape false %}
  {{ user_input }}  {# XSS VULNERABILITY! #}
{% endautoescape %}

{#  NEVER DO THIS #}
{{ user_comment|safe }}  {# XSS VULNERABILITY! #}
```

**Rule:** Only use `|safe` or `{% autoescape false %}` for content you **completely control and trust**.

---

## Environment Configuration

### Available Options

Configure template behavior at environment creation:

```go
env := miya.NewEnvironment(
    miya.WithAutoEscape(true),        // Auto-escape HTML (default: true)
    miya.WithStrictUndefined(true),   // Error on undefined vars (default: false)
    miya.WithTrimBlocks(true),        // Remove first newline after block (default: false)
    miya.WithLstripBlocks(true),      // Strip leading spaces before blocks (default: false)
)
```

### AutoEscape

```go
// Enable (default)
miya.WithAutoEscape(true)   // {{ "<script>" }} → &lt;script&gt;

// Disable (be careful!)
miya.WithAutoEscape(false)  // {{ "<script>" }} → <script>
```

### StrictUndefined

```go
// Lenient (default) - undefined variables render as empty
miya.WithStrictUndefined(false)
// {{ undefined_var }} → ""

// Strict - undefined variables cause errors
miya.WithStrictUndefined(true)
// {{ undefined_var }} → ERROR: undefined variable
```

### TrimBlocks

```go
// Enable - remove first newline after block tags
miya.WithTrimBlocks(true)

// Template:
// {% if true %}
// text
// {% endif %}
//
// Output: "text" (no leading newline)
```

### LstripBlocks

```go
// Enable - remove leading whitespace before block tags
miya.WithLstripBlocks(true)

// Template:
//     {% if true %}
//     text
//     {% endif %}
//
// Output: "text" (no leading spaces)
```

### Combining Options

```go
// Production configuration
env := miya.NewEnvironment(
    miya.WithAutoEscape(true),       // Security
    miya.WithStrictUndefined(true),  // Catch errors early
    miya.WithTrimBlocks(true),       // Clean output
    miya.WithLstripBlocks(true),     // Clean output
)

// Development configuration
env := miya.NewEnvironment(
    miya.WithAutoEscape(false),      // Easier debugging
    miya.WithStrictUndefined(false), // Lenient for incomplete data
)
```

---

## Performance & Memory Management

Miya Engine includes several optimizations for high-performance template rendering.

### AST Node Pooling

The parser uses `sync.Pool` to reuse frequently allocated AST nodes, reducing garbage collection pressure during template parsing. This is automatic and requires no configuration.

**Pooled node types:**
- `LiteralNode` - String, number, and boolean literals
- `IdentifierNode` - Variable references
- `BinaryOpNode` - Binary operators (+, -, ==, and, or, etc.)
- `FilterNode` - Filter applications
- `UnaryOpNode` - Unary operators (not, -)
- `AttributeNode` - Attribute access (obj.attr)
- `GetItemNode` - Index access (obj[key])
- `CallNode` - Function calls

### Template.Release()

For high-throughput applications that create many short-lived templates, you can explicitly release pooled AST nodes back to their pools:

```go
// Create and use a template
tmpl, err := env.FromString("Hello {{ name }}!")
if err != nil {
    log.Fatal(err)
}

output, err := tmpl.Render(ctx)
if err != nil {
    log.Fatal(err)
}

// Optional: release pooled nodes for reuse
tmpl.Release()
```

**When to use `Release()`:**
- Processing many unique templates in a loop
- One-time template rendering where the template won't be reused
- Memory-constrained environments
- High-throughput services generating templates dynamically

**When NOT to use `Release()`:**
- Templates cached in the environment (they're reused)
- Long-lived templates rendered multiple times
- When memory pressure isn't a concern

### Evaluator Pooling

The environment automatically pools evaluator instances for reuse across render calls. This is handled internally and requires no user action.

```go
// Evaluators are automatically pooled and reused
for i := 0; i < 1000; i++ {
    output, _ := tmpl.Render(ctx)  // Reuses pooled evaluators
}
```

### Template Caching

Templates are cached by name (for file-based templates) or by content hash (for string templates):

```go
// First call: parses and caches the template
tmpl1, _ := env.FromString("Hello {{ name }}!")

// Second call: returns cached template (no parsing)
tmpl2, _ := env.FromString("Hello {{ name }}!")
// tmpl1 and tmpl2 are the same cached instance
```

### Cache Management

```go
// Clear all cached templates
env.ClearCache()

// Invalidate a specific template (removes from both regular and inheritance caches)
env.InvalidateTemplate("page.html")

// Check cache size
size := env.GetCacheSize()
```

### Best Practices for Performance

```go
// 1. Reuse environments - don't create new ones per request
var env = miya.NewEnvironment(
    miya.WithLoader(loader),
    miya.WithAutoEscape(true),
)

// 2. Let templates cache - call GetTemplate/FromString once
tmpl, _ := env.GetTemplate("page.html")
for _, data := range items {
    ctx := miya.NewContext()
    ctx.Set("item", data)
    output, _ := tmpl.Render(ctx)
}

// 3. Use Release() only for one-off templates
func processUserTemplate(userTemplate string, data interface{}) string {
    tmpl, _ := env.FromString(userTemplate)
    defer tmpl.Release()  // Release since this is a one-time template

    ctx := miya.NewContext()
    ctx.Set("data", data)
    output, _ := tmpl.Render(ctx)
    return output
}
```

---

## Practical Examples

### Example 1: Clean HTML Output

```jinja2
<nav>
  <ul>
    {%- for item in nav_items %}
    <li><a href="{{ item.url }}">{{ item.title }}</a></li>
    {%- endfor %}
  </ul>
</nav>
```

**Output (no extra whitespace):**
```html
<nav>
  <ul><li><a href="/">Home</a></li><li><a href="/about">About</a></li></ul>
</nav>
```

### Example 2: Text Transformation Blocks

```jinja2
{# Convert entire section to uppercase #}
{% filter upper %}
<h1>Important Announcement</h1>
<p>This message will be in uppercase.</p>
<p>Variable {{ name }} will also be uppercase.</p>
{% endfilter %}
```

### Example 3: Safe Content Rendering

```jinja2
<article>
  {# Title is escaped #}
  <h1>{{ article.title }}</h1>

  {# Body is trusted HTML #}
  <div class="content">
    {{ article.body|safe }}
  </div>

  {# User comments are escaped #}
  <div class="comments">
    <h2>Comments</h2>
    {% for comment in comments %}
      <div class="comment">
        <strong>{{ comment.author }}</strong>
        <p>{{ comment.text }}</p>
      </div>
    {% endfor %}
  </div>
</article>
```

### Example 4: Template Documentation

```jinja2
<h2>Template Syntax Guide</h2>

<h3>Variables:</h3>
<pre>{% raw %}{{ variable }}{% endraw %}</pre>

<h3>Loops:</h3>
<pre>
{% raw %}
{% for item in items %}
  {{ item }}
{% endfor %}
{% endraw %}
</pre>

<h3>Conditionals:</h3>
<pre>
{% raw %}
{% if condition %}
  content
{% endif %}
{% endraw %}
</pre>
```

### Example 5: Combining Features

```jinja2
{# Filter block + whitespace control #}
<div class="notice">
{%- filter upper|trim -%}
   This is an important notice
{%- endfilter -%}
</div>

{# Autoescape + filter block #}
{% autoescape false %}
{% filter safe %}
  {{ rich_text_content }}
{% endfilter %}
{% endautoescape %}

{# Whitespace control + conditionals #}
<ul>
{%- for item in items %}
  {%- if item.visible %}
  <li>{{ item.name }}</li>
  {%- endif %}
{%- endfor %}
</ul>
```

---

## Complete Reference

### Feature Summary

| Feature | Syntax | Status | Purpose |
|---------|--------|--------|---------|
| **Filter Blocks** | `{% filter name %}...{% endfilter %}` |  Full | Apply filters to blocks |
| **Chained Filters** | `{% filter f1\|f2 %}...{% endfilter %}` |  Full | Multiple filters |
| **Do Statements** | `{% do expression %}` |  Full | Execute without output |
| **Whitespace Control** | `{%-`, `-%}`, `{{-`, `-}}` |  Full | Control whitespace |
| **Raw Blocks** | `{% raw %}...{% endraw %}` |  Full | Prevent processing |
| **Autoescape** | `{% autoescape bool %}...{% endautoescape %}` |  Full | Control HTML escaping |
| **Safe Filter** | `{{ var\|safe }}` |  Full | Mark as safe HTML |
| **Escape Filter** | `{{ var\|escape }}` |  Full | Force escaping |

### Environment Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `AutoEscape` | bool | `true` | Automatic HTML escaping |
| `StrictUndefined` | bool | `false` | Error on undefined variables |
| `TrimBlocks` | bool | `false` | Remove first newline after blocks |
| `LstripBlocks` | bool | `false` | Strip leading whitespace |

### Memory Management Methods

| Method | Description |
|--------|-------------|
| `Template.Release()` | Return pooled AST nodes for reuse (optional) |
| `Environment.ClearCache()` | Clear all cached templates |
| `Environment.InvalidateTemplate(name)` | Remove specific template from cache |
| `Environment.GetCacheSize()` | Get number of cached templates |

---

## Best Practices

### 1. Use Whitespace Control Sparingly

```jinja2
{#  Good - strategic use #}
Items: {%- for item in items %}{{ item }}{{ ", " if not loop.last }}{% endfor %}

{#  Avoid - excessive #}
{%- for item in items -%}
  {{- item -}}
{%- endfor -%}
```

### 2. Always Escape User Input

```jinja2
{#  SAFE #}
<p>{{ user.comment }}</p>

{#  DANGEROUS #}
<p>{{ user.comment|safe }}</p>
```

### 3. Use Raw for Documentation Only

```jinja2
{#  Good - documenting syntax #}
<pre>{% raw %}{{ variable }}{% endraw %}</pre>

{#  Avoid - for actual content #}
{% raw %}{{ actual_variable_to_render }}{% endraw %}
```

### 4. Filter Blocks for Transformation

```jinja2
{#  Good - transform entire section #}
{% filter upper %}
  {% include "announcement.html" %}
{% endfilter %}

{#  Avoid - single line (use inline filter) #}
{% filter upper %}{{ text }}{% endfilter %}
{# Better: {{ text|upper }} #}
```

---

## See Also

- [Working Example](../examples/features/advanced/) - Complete advanced features demo
- [Filters Guide](FILTERS_GUIDE.md) - Available filters for filter blocks
- [Control Structures](CONTROL_STRUCTURES_GUIDE.md) - Template control flow
- [Security](FILTERS_GUIDE.md#html--security-filters) - HTML escaping and security

---

## Summary

**All Advanced Features: 100% Jinja2 Compatible**

 **Fully Supported:**
- **Filter Blocks** - Apply filters to entire content blocks
- **Chained Filters** - Multiple filters on blocks
- **Do Statements** - Execute without output
- **Whitespace Control** - Fine-grained whitespace management with `-`
- **Raw Blocks** - Prevent template processing
- **Autoescape** - Control HTML escaping for security
- **Environment Configuration** - AutoEscape, StrictUndefined, TrimBlocks, LstripBlocks
- **Performance Optimizations** - AST node pooling, evaluator pooling, template caching

**Key Benefits:**
- Clean, professional output
- Security through autoescaping
- Documentation-friendly with raw blocks
- Powerful block transformations
- High performance with minimal memory overhead

All advanced features work identically to Jinja2, providing professional-grade template control.
