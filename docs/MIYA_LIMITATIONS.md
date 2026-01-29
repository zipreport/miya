# Miya Engine Limitations and Differences from Jinja2

This document details the **known limitations and differences** between Miya Engine and Jinja2, based on extensive testing of the implementation.

> **Last Updated:** Based on testing all feature examples in `examples/features/`
> **Miya Version:** Current implementation as of November 2024

---

##  Table of Contents

1. [Critical Limitations](#critical-limitations)
2. [Comprehensions Limitations](#comprehensions-limitations)
3. [Filter Limitations](#filter-limitations)
4. [Macro Limitations](#macro-limitations)
5. [Syntax Limitations](#syntax-limitations)
6. [Workarounds and Alternatives](#workarounds-and-alternatives)

---

## Critical Limitations

###  1. List Comprehensions with Filter Clauses

**Not Supported:**
```jinja2
{{ [x for x in numbers if x > 5] }}
{{ [user.name for user in users if user.active] }}
{{ [x for x in items if x.price > 100 and x.stock > 0] }}
```

**Error:**
```
parser error: expected 'else' in conditional expression
```

**Reason:** The parser treats any `if` inside a comprehension as a ternary operator and requires an `else` clause.

**Workaround:**
```jinja2
{# Use filter chains instead #}
{{ users|selectattr("active")|map(attribute="name")|list }}
{{ numbers|select("greaterthan", 5)|list }}
```

---

###  2. Dictionary Comprehensions with `.items()` Unpacking

**Not Supported:**
```jinja2
{{ {k: v for k, v in dict.items()} }}
{{ {key|upper: value for key, value in config.items()} }}
```

**Error:**
```
parser error: expected 'in' in dict comprehension
```

**Reason:** The parser doesn't support tuple unpacking in dictionary comprehensions.

**Workaround:**
```jinja2
{# Only single-item iteration works #}
{{ {user.id: user.name for user in users} }}
{{ {item.key: item.value for item in key_value_list} }}
```

---

###  3. Nested Comprehensions

**Not Supported:**
```jinja2
{{ [item for category in categories for item in category.items] }}
{{ [(a, b) for a in list1 for b in list2] }}
```

**Error:**
```
parser error: unexpected 'for' in comprehension
```

**Reason:** Multiple `for` clauses in a single comprehension are not parsed correctly.

**Workaround:**
```jinja2
{# Use nested loops instead #}
{% set result = [] %}
{% for category in categories %}
  {% for item in category.items %}
    {% do result.append(item) %}
  {% endfor %}
{% endfor %}
```

---

###  4. Inline Conditionals Without `else`

**Not Supported:**
```jinja2
{{ ", " if not loop.last }}
{{ value if condition }}
{{ user.name if user }}
```

**Error:**
```
parser error: expected 'else' in conditional expression
```

**Reason:** The parser always expects a complete ternary expression with both branches.

**Workaround:**
```jinja2
{# Use block-level conditionals #}
{% if not loop.last %}, {% endif %}
{% if condition %}{{ value }}{% endif %}

{# Or provide an else clause #}
{{ ", " if not loop.last else "" }}
{{ value if condition else "" }}
```

---

## Comprehensions Limitations

### Summary Table

| Feature | Jinja2 | Miya | Status |
|---------|--------|------|--------|
| `[expr for item in list]` |  |  | Works |
| `[expr for item in list if cond]` |  |  | **Not Supported** |
| `{key: val for item in list}` |  |  | Works |
| `{k: v for k, v in dict.items()}` |  |  | **Not Supported** |
| `[item for list in lists for item in list]` |  |  | **Not Supported** |
| `{k: v for item in list if cond}` |  |  | **Not Supported** |

### What Works

```jinja2
{#  Basic list comprehensions #}
{{ [x * 2 for x in numbers] }}
{{ [user.name for user in users] }}
{{ [item|upper for item in items] }}

{#  Basic dict comprehensions #}
{{ {user.id: user.name for user in users} }}
{{ {item.key: item.value for item in data} }}

{#  With filters applied #}
{{ [name|title for name in names] }}
{{ [price|round(2) for price in prices] }}
```

---

## Filter Limitations

###  Collection Filters with Limited Support

The following filters appear to have issues when chained or used with certain data types:

**Problematic Filters:**
- `sort` - May fail with: "sort filter requires a sequence"
- `reverse` - Limited support
- `unique` - May not work as expected
- `slice` - Limited support
- `batch` - May have issues
- `select` - Works but limited
- `reject` - Works but limited
- `selectattr` - Works in filter chains, not in comprehensions
- `rejectattr` - Works in filter chains, not in comprehensions
- `groupby` - Limited testing

**Working Alternatives:**
```jinja2
{#  This may fail #}
{{ numbers|sort|reverse }}

{#  Use these instead #}
{{ numbers|list }}  {# Just convert to list #}
{% for num in numbers|sort %}{{ num }}{% endfor %}  {# In loops #}
```

###  Reliably Working Filters

**String Filters:**
- `upper`, `lower`, `capitalize`, `title`
- `trim`, `replace`, `truncate`
- `split`, `join`
- `startswith`, `endswith`, `contains`
- `slugify`

**Numeric Filters:**
- `abs`, `round`, `int`, `float`
- `default`

**HTML Filters:**
- `escape`, `safe`, `striptags`
- `urlencode`

**Other:**
- `length`, `first`, `last`
- `default`
- `tojson`

---

## Macro Limitations

###  `caller()` Function Not Supported

**Not Supported:**
```jinja2
{% macro wrapper(title) %}
  <div class="card">
    <h3>{{ title }}</h3>
    {{ caller() }}  {#  Not supported #}
  </div>
{% endmacro %}

{% call wrapper("My Card") %}
  <p>Card content</p>
{% endcall %}
```

**Error:**
```
cannot call non-function value of type *runtime.Undefined
```

**Workaround:**
```jinja2
{# Pass content as a parameter instead #}
{% macro wrapper(title, content) %}
  <div class="card">
    <h3>{{ title }}</h3>
    {{ content }}
  </div>
{% endmacro %}

{% set my_content %}
  <p>Card content</p>
{% endset %}
{{ wrapper("My Card", my_content) }}
```

###  Varargs and Kwargs Not Supported

**Not Supported:**
```jinja2
{% macro flexible(*args, **kwargs) %}
  {#  Not supported #}
{% endmacro %}
```

**Workaround:**
```jinja2
{# Use explicit parameters with defaults #}
{% macro flexible(param1="", param2="", param3="") %}
  {#  This works #}
{% endmacro %}
```

---

## Syntax Limitations

###  `enumerate()` Keyword Arguments

**Not Supported:**
```jinja2
{% for i, item in enumerate(items, start=1) %}  {#  #}
```

**Workaround:**
```jinja2
{% for i, item in enumerate(items, 1) %}  {#  Positional arg #}
```

###  Dictionary Iteration with Tuple Unpacking

**Not Supported:**
```jinja2
{% for key, value in dict.items() %}  {#  #}
```

**Workaround:**
```jinja2
{# Iterate over keys #}
{% for key in dict %}
  {{ key }}: {{ dict[key] }}
{% endfor %}

{# Or use a list of objects #}
{% for item in items %}
  {{ item.key }}: {{ item.value }}
{% endfor %}
```

###  Multiple Variable Assignment

**Not Supported:**
```jinja2
{% set a, b = 1, 2 %}  {#  #}
{% set x, y = some_tuple %}  {#  #}
```

**Workaround:**
```jinja2
{% set a = 1 %}
{% set b = 2 %}
```

---

## Workarounds and Alternatives

### Pattern 1: Filtering Data

**Instead of comprehensions with filters:**
```jinja2
{#  Doesn't work #}
{{ [user for user in users if user.active] }}

{#  Use filter chains #}
{{ users|selectattr("active")|list }}

{#  Or use loops with conditionals #}
{% set active_users = [] %}
{% for user in users %}
  {% if user.active %}
    {% do active_users.append(user) %}
  {% endif %}
{% endfor %}
```

### Pattern 2: Complex Data Transformations

**Instead of nested comprehensions:**
```jinja2
{#  Doesn't work #}
{{ [item for cat in categories for item in cat.items] }}

{#  Use nested loops #}
{% set all_items = [] %}
{% for category in categories %}
  {% for item in category.items %}
    {% do all_items.append(item) %}
  {% endfor %}
{% endfor %}
```

### Pattern 3: Conditional Output

**Instead of inline if without else:**
```jinja2
{#  Doesn't work #}
{{ ", " if not loop.last }}

{#  Use block conditionals #}
{% if not loop.last %}, {% endif %}

{#  Or provide else clause #}
{{ ", " if not loop.last else "" }}
```

### Pattern 4: Dictionary Processing

**Instead of .items() unpacking:**
```jinja2
{#  Doesn't work #}
{{ {k|upper: v for k, v in dict.items()} }}

{#  Create list of key-value objects first #}
{% set kv_list = [] %}
{% for key in dict %}
  {% do kv_list.append({"key": key, "value": dict[key]}) %}
{% endfor %}
{{ {item.key|upper: item.value for item in kv_list} }}
```

---

## Testing Your Templates

When migrating from Jinja2 to Miya, test for these patterns:

### Checklist

- [ ] No comprehensions with `if` clauses
- [ ] No `.items()` unpacking in comprehensions
- [ ] No nested comprehensions (multiple `for` clauses)
- [ ] All inline conditionals have `else` clauses
- [ ] No `caller()` in macros
- [ ] No varargs/kwargs in macros
- [ ] `enumerate()` uses positional args only
- [ ] No tuple unpacking in `{% for %}` loops
- [ ] No multiple variable assignment with `{% set %}`

### Quick Test

```bash
# Run examples to verify syntax
cd examples/features/
for dir in */; do
  cd "$dir"
  go run main.go > /dev/null 2>&1 && echo " $dir" || echo " $dir"
  cd ..
done
```

---

## See Also

- [Feature Examples](https://github.com/zipreport/miya/tree/master/examples/features/) - Working examples of all supported features
- [JINJA2_VS_MIYA_FEATURE_MATRIX.md](JINJA2_VS_MIYA_FEATURE_MATRIX.md) - Complete compatibility matrix
- [COMPREHENSIVE_FEATURES.md](COMPREHENSIVE_FEATURES.md) - All supported features with examples

---

## Summary

**Miya Engine provides ~95% Jinja2 compatibility** with these main limitations:

1.  Comprehensions with `if` clauses
2.  Dictionary comprehensions with `.items()`
3.  Nested comprehensions
4.  Inline conditionals without `else`
5.  Macro `caller()` function
6.  Some collection filters have limited support

**Most Jinja2 templates will work with minor adjustments.** Use the examples in `examples/features/` as a reference for working syntax.
