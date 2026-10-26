# Jinja2 vs Miya (Miya Engine) Feature Matrix

This document provides a comprehensive mapping of all template engine features between Python Jinja2 and Go Miya,
including implementation differences and compatibility notes.

## How compatible is miya?

miya's conformance suite (`tests/conformance/`) holds 289 small templates written to probe edge cases and known bugs; they are not a sample of real templates. At this commit miya matches Jinja2 3.1.6 on 283 of them (same output, or failing in the same phase for the cases that expect an error), after normalising the printing differences listed below. 4 of the remaining cases are deliberate differences. Every divergence and its reason is listed in `tests/conformance/known_divergences.json`. This is a regression net, not a compatibility percentage.

A second suite checks whitespace control (`trim_blocks`, `lstrip_blocks`, `-`/`+` markers, `keep_trailing_newline`) against Jinja2 on 3240 generated templates, with no allowed divergences.

**Deliberate differences** from Jinja2 3.1.6:

- At the top level, booleans print `true`/`false` and `None` prints empty. Inside lists and dicts they print as in Python (`True`, `False`, `None`).
- Whole floats print without `.0` everywhere, including inside lists and after `string`: `{{ 1.0 }}` and `{{ [1.0] }}` print `1` and `[1]`.
- `round` rounds half away from zero and prints exactly the requested number of decimals: `2.5|round` prints `3` (Jinja2: `2.0`), `2.5|round(2)` prints `2.50` and `2.675|round(2)` prints `2.68`.
- `default` also replaces `None` and the empty string, not only undefined values.
- Dicts iterate and print in sorted key order, not insertion order.
- Integers are 64-bit: `9223372036854775807 + 1` becomes a float.
- Methods on Go values are not callable from templates. Pass computed values as data or register a global function.
- `None` sorts before every other value in `sort`, where Jinja2 raises an error.
- `map` with an unknown filter name yields empty values where Jinja2 raises an error.

Differences that are not deliberate, and miya-only extensions, are listed in [MIYA_LIMITATIONS.md](MIYA_LIMITATIONS.md).

## Overview

- **Jinja2**: The original Python template engine
- **Miya**: Go implementation (github.com/zipreport/miya)
- **Status Key**:  Fully Compatible |  Partial/Different |  Not Implemented |  In Progress

##  Documentation Links

- **[COMPREHENSIVE_FEATURES.md](COMPREHENSIVE_FEATURES.md)** - Complete feature reference with examples
- **[Example Code](https://github.com/zipreport/miya/tree/master/examples/go/complex/)** - Working examples of all features
- **[Template Examples](https://github.com/zipreport/miya/tree/master/examples/go/complex/templates/)** - Professional template organization

---

## 1. Template Syntax & Delimiters

| Feature              | Python Jinja2                 | Go Miya                     | Status | Notes        |
|----------------------|-------------------------------|-------------------------------|--------|--------------|
| Variable expressions | `{{ variable }}`              | `{{ variable }}`              |       | Identical    |
| Statement blocks     | `{% statement %}`             | `{% statement %}`             |       | Identical    |
| Comments             | `{# comment #}`               | `{# comment #}`               |       | Identical    |
| Whitespace control   | `{{- var -}}`, `{%- stmt -%}` | `{{- var -}}`, `{%- stmt -%}` |       | Full support |

---

## 2. Control Structures

### 2.1 Conditional Statements

| Feature             | Python Jinja2                   | Go Miya                       | Status | Notes        |
|---------------------|---------------------------------|---------------------------------|--------|--------------|
| If statement        | `{% if condition %}`            | `{% if condition %}`            |       | Full support |
| Elif statement      | `{% elif condition %}`          | `{% elif condition %}`          |       | Full support |
| Else statement      | `{% else %}`                    | `{% else %}`                    |       | Full support |
| Inline conditionals | `{{ 'yes' if true else 'no' }}` | `{{ 'yes' if true else 'no' }}` |       | Full support |

### 2.2 Loops

| Feature            | Python Jinja2                       | Go Miya                           | Status | Notes                         |
|--------------------|-------------------------------------|-------------------------------------|--------|-------------------------------|
| For loops          | `{% for item in items %}`           | `{% for item in items %}`           |       | Full support                  |
| Variable unpacking | `{% for k, v in dict.items() %}`    | `{% for k, v in dict %}`            |       | miya-only; Jinja2 requires `.items()` (`.items()` also works in miya) |
| Loop variables     | `loop.index`, `loop.first`, etc.    | `loop.index`, `loop.first`, etc.    |       | All loop vars supported       |
| Empty clause       | `{% else %}` in for loops           | `{% else %}` in for loops           |       | Full support                  |
| Loop control       | `{% break %}`, `{% continue %}`     | `{% break %}`, `{% continue %}`     |       | Jinja2 needs the `jinja2.ext.loopcontrols` extension |
| Recursive loops    | `{% for item in items recursive %}` | `{% for item in items recursive %}` |       | Full support                  |

---

## 3. Template Inheritance & Composition

| Feature              | Python Jinja2                         | Go Miya                             | Status | Notes                        |
|----------------------|---------------------------------------|---------------------------------------|--------|------------------------------|
| Template extension   | `{% extends "base.html" %}`           | `{% extends "base.html" %}`           |       | Full support                 |
| Block definition     | `{% block name %}`                    | `{% block name %}`                    |       | Full support                 |
| Block override       | Child template blocks                 | Child template blocks                 |       | Full support                 |
| Super calls          | `{{ super() }}`                       | `{{ super() }}`                       |       | Full support                 |
| Include templates    | `{% include "template.html" %}`       | `{% include "template.html" %}`       |       | Requires proper loader setup |
| Include with context | `{% include "t.html" with context %}` | `{% include "t.html" with context %}` |       | Full support                 |

---

## 4. Import System

| Feature             | Python Jinja2                             | Go Miya                                 | Status | Notes        |
|---------------------|-------------------------------------------|-------------------------------------------|--------|--------------|
| Import templates    | `{% import "macros.html" as m %}`         | `{% import "macros.html" as m %}`         |       | Full support. Imports are without context, as in Jinja2 |
| Selective import    | `{% from "macros.html" import macro %}`   | `{% from "macros.html" import macro %}`   |       | Full support |
| Import with context | `{% import "m.html" as m with context %}` | `{% import "m.html" as m with context %}` |       | Full support |

---

## 5. Macros & Functions

| Feature          | Python Jinja2                            | Go Miya                                | Status | Notes                        |
|------------------|------------------------------------------|------------------------------------------|--------|------------------------------|
| Macro definition | `{% macro name(args) %}`                 | `{% macro name(args) %}`                 |       | Full support                 |
| Macro calls      | `{{ macro_name(args) }}`                 | `{{ macro_name(args) }}`                 |       | Full support                 |
| Call blocks      | `{% call macro() %}content{% endcall %}` | `{% call macro() %}content{% endcall %}` |       | Full support with `caller()`; `{% call(v) macro() %}` also works |
| Caller function  | `{{ caller() }}` in macros               | `{{ caller() }}` in macros               |       | Full support                 |
| Varargs macros   | implicit `varargs` / `kwargs` inside the macro body (`*args` is not Jinja2 syntax) | implicit `varargs` / `kwargs` inside the macro body (`*args` is not miya syntax) |       | Full support |

---

## 6. Variable Assignment

| Feature             | Python Jinja2                      | Go Miya                          | Status | Notes        |
|---------------------|------------------------------------|------------------------------------|--------|--------------|
| Simple assignment   | `{% set var = value %}`            | `{% set var = value %}`            |       | Full support |
| Multiple assignment | `{% set a, b = values %}`          | `{% set a, b = values %}`          |       | Full support |
| Block assignment    | `{% set var %}content{% endset %}` | `{% set var %}content{% endset %}` |       | Full support |
| Namespace scoping   | `{% set ns = namespace() %}`       | `{% set ns = namespace() %}`       |       | Full support |

---

## 7. Context Management

| Feature            | Python Jinja2            | Go Miya                | Status | Notes        |
|--------------------|--------------------------|--------------------------|--------|--------------|
| With statements    | `{% with var = value %}` | `{% with var = value %}` |       | Full support |
| Multiple with vars | `{% with a=1, b=2 %}`    | `{% with a=1, b=2 %}`    |       | Full support |
| Nested with        | Multiple levels          | Multiple levels          |       | Full support |

---

## 8. Special Blocks

| Feature           | Python Jinja2                           | Go Miya                               | Status | Notes                                                     |
|-------------------|-----------------------------------------|-----------------------------------------|--------|-----------------------------------------------------------|
| Raw blocks        | `{% raw %}content{% endraw %}`          | `{% raw %}content{% endraw %}`          |       | Full support                                              |
| Comment blocks    | `{# multiline comments #}`              | `{# multiline comments #}`              |       | Full support                                              |
| Autoescape blocks | `{% autoescape false %}`                | `{% autoescape false %}`                |       | Full support                                              |
| Filter blocks     | `{% filter upper %}text{% endfilter %}` | `{% filter upper %}text{% endfilter %}` |       |  FULLY IMPLEMENTED with chaining support                 |
| Do statements     | `{% do expression %}`                   | `{% do expression %}`                   |       | Jinja2 needs the `jinja2.ext.do` extension |

### 8.1 Filter Blocks Implementation Details

**Filter blocks** in Go Miya provide the same functionality as Python Jinja2.

**Basic Usage:**

```html+jinja
{% filter upper %}
  Hello World
{% endfilter %}
<!-- Output: HELLO WORLD -->
```

**Filter Chaining:**

```html+jinja
{% filter upper|reverse %}
  hello world
{% endfilter %}
<!-- Output: DLROW OLLEH -->
```

**Filters with Arguments:**

```html+jinja
{% filter truncate(10) %}This is a very long text that will be truncated{% endfilter %}
<!-- Output: This... -->
```

**Complex Content Support:**

```html+jinja
{% filter upper %}
  Hello {{ name }}!
  {% for item in items %}
    - {{ item }}
  {% endfor %}
{% endfilter %}
```

**Nested Filter Blocks:**

```html+jinja
{% filter upper %}
  Outer content
  {% filter lower %}
    INNER CONTENT
  {% endfilter %}
{% endfilter %}
<!-- Applies filters from inner to outer -->
```

**Implementation Features:**

-  AST-based parsing with proper error handling
-  Integration with existing filter registry system
-  Support for all built-in and custom filters
-  Filter chaining with proper left-to-right evaluation
-  Complex template content (variables, loops, conditionals)
-  Nested filter blocks with correct scope handling

### 8.2 Do Statements Implementation Details

**Do statements** in Go Miya behave like the `jinja2.ext.do` extension of Python Jinja2 (stock Jinja2 rejects `{% do %}`),
allowing execution of expressions purely for side effects without producing template output.

**Basic Usage:**

```html+jinja
{% set items = [] %}
{% do items.append("hello") %}  {# Python-style method calls; only on a list created with {% set %} #}
{{ items[0] }}  <!-- Output: hello -->

{# Miya Engine equivalent using expressions #}
{% set value = 42 %}
{% do value * 2 + 10 %}  {# Executes expression, no output #}
Original value: {{ value }}  <!-- Output: Original value: 42 -->
```

**Expression Evaluation:**

```html+jinja
{# Arithmetic expressions #}
{% do 5 + 3 * 2 %}

{# Filter applications #}
{% set text = "hello" %}
{% do text|upper %}
Filtered: {{ text|upper }}  <!-- Output: Filtered: HELLO -->

{# Complex expressions #}
{% do (variable + 10) * factor|default(1) %}
```

**Integration with Control Flow:**

```html+jinja
{# In conditionals #}
{% if condition %}
  {% do complex_expression %}
  Content here
{% endif %}

{# In loops #}
{% for item in items %}
  {% do item * 2 + offset %}
  Item: {{ item }}
{% endfor %}

{# With whitespace control #}
Start{%- do expression -%}End  <!-- Output: StartEnd -->
```

**Use Cases:**

```html+jinja
{# Trigger calculations without output #}
{% set counter = 0 %}
{% for i in range(10) %}
  {% do counter + i %}  {# Expression evaluated but no output #}
{% endfor %}
Counter remains: {{ counter }}

{# Apply filters for validation #}
{% set user_input = "  UNSAFE DATA  " %}
{% do user_input|trim|lower %}
Clean input: {{ user_input|trim|lower }}

{# Complex expression evaluation #}
{% do (price * quantity * tax_rate)|round(2) %}
```

**Implementation Features:**

-  Full AST-based parsing with proper error handling
-  Integration with all existing expression types
-  Support for complex expressions (arithmetic, filters, function calls)
-  Proper integration with control flow structures
-  Whitespace control support (`{%- do ... -%}`)
-  Zero impact on template output (true side-effects only)

**Error Handling:**

```html+jinja
{# Undefined variable error propagation #}
{% do undefined_variable %}  <!-- Error: undefined variable -->

{# Invalid expression syntax #}
{% do 5 + %}  <!-- Error: malformed expression -->

{# Missing expression #}
{% do %}  <!-- Error: expected expression -->
```

---

## 9. Expressions & Literals

### 9.1 Basic Literals

| Feature      | Python Jinja2        | Go Miya            | Status | Notes        |
|--------------|----------------------|----------------------|--------|--------------|
| Strings      | `"hello"`, `'hello'` | `"hello"`, `'hello'` |       | Full support |
| Numbers      | `42`, `3.14`         | `42`, `3.14`         |       | Full support |
| Booleans     | `true`, `false`      | `true`, `false`      |       | Full support |
| None/null    | `none`               | `none`               |       | Full support |
| Lists        | `[1, 2, 3]`          | `[1, 2, 3]`          |       | Full support |
| Dictionaries | `{key: value}`       | `{key: value}`       |       | Full support |

### 9.2 Complex Expressions

| Feature                 | Python Jinja2                   | Go Miya                       | Status | Notes                                |
|-------------------------|---------------------------------|---------------------------------|--------|--------------------------------------|
| Attribute access        | `object.attribute`              | `object.attribute`              |       | Full support                         |
| Item access             | `dict['key']`                   | `dict['key']`                   |       | Full support                         |
| Method calls            | `string.upper()`                | `string.upper()`                |       | Python string, list and dict methods are supported (`upper`, `split`, `get`, `append`, ...); Go methods on your data are not callable |
| List slicing            | `list[1:3]`                     | `list[1:3]`                     |       | Native slice syntax support          |
| List comprehensions     | `[x for x in items]`            | `[x for x in items]`            |       | miya-only; not Jinja2 syntax. Nested `for` clauses and `.items()` unpacking in dict comprehensions are not supported |
| Conditional expressions | `value if condition else other` | `value if condition else other` |       | Full support                         |

---

## 10. Operators

### 10.1 Arithmetic Operators

| Feature              | Python Jinja2 | Go Miya | Status | Notes        |
|----------------------|---------------|-----------|--------|--------------|
| Addition             | `+`           | `+`       |       | Full support |
| Subtraction          | `-`           | `-`       |       | Full support |
| Multiplication       | `*`           | `*`       |       | Full support |
| Division             | `/`           | `/`       |       | Full support |
| Floor division       | `//`          | `//`      |       | Full support |
| Modulo               | `%`           | `%`       |       | Full support |
| Power                | `**`          | `**`      |       | Full support |
| String concatenation | `~`           | `~`       |       | Full support |

### 10.2 Comparison Operators

| Feature       | Python Jinja2 | Go Miya | Status | Notes        |
|---------------|---------------|-----------|--------|--------------|
| Equality      | `==`          | `==`      |       | Full support |
| Inequality    | `!=`          | `!=`      |       | Full support |
| Less than     | `<`           | `<`       |       | Full support |
| Less/equal    | `<=`          | `<=`      |       | Full support |
| Greater than  | `>`           | `>`       |       | Full support |
| Greater/equal | `>=`          | `>=`      |       | Full support |

### 10.3 Logical Operators

| Feature | Python Jinja2 | Go Miya | Status | Notes        |
|---------|---------------|-----------|--------|--------------|
| And     | `and`         | `and`     |       | Full support |
| Or      | `or`          | `or`      |       | Full support |
| Not     | `not`         | `not`     |       | Full support |

### 10.4 Membership & Identity

| Feature     | Python Jinja2       | Go Miya           | Status | Notes        |
|-------------|---------------------|---------------------|--------|--------------|
| In operator | `item in sequence`  | `item in sequence`  |       | Full support |
| Is operator | `value is test`     | `value is test`     |       | Full support |
| Is not      | `value is not test` | `value is not test` |       | Full support |

---

## 11. Built-in Filters

### 11.1 String Filters

| Filter        | Python Jinja2 | Go Miya | Status | Notes                      |
|---------------|---------------|-----------|--------|----------------------------|
| upper         |              |          |       | Identical                  |
| lower         |              |          |       | Identical                  |
| capitalize    |              |          |       | Identical                  |
| title         |              |          |       | Identical                  |
| trim/strip    |              |          |       | Identical                  |
| replace       |              |          |       | Identical                  |
| truncate      |              |          |       | Identical                  |
| wordwrap      |              |          |       | Identical                  |
| center        |              |          |       | Identical                  |
| indent        |              |          |       | Identical                  |
| slugify       |              |          |      | Go-specific implementation |
| split         |              |          |       | miya only (not a Jinja2 filter) |
| startswith    |              |          |      | Go-specific implementation |
| endswith      |              |          |      | Go-specific implementation |
| contains      |              |          |      | Go-specific implementation |
| regex_replace |              |          |       | miya only (not a Jinja2 filter) |
| regex_search  |              |          |       | miya only (not a Jinja2 filter) |
| wordcount     |              |          |      | Go-specific implementation |

### 11.2 Math Filters

| Filter | Python Jinja2 | Go Miya | Status | Notes                      |
|--------|---------------|-----------|--------|----------------------------|
| abs    |              |          |       | Identical                  |
| round  |              |          |       | Differs on purpose: half away from zero, prints exactly n decimals |
| int    |              |          |       | Identical                  |
| float  |              |          |       | Identical                  |
| sum    |              |          |       | Identical                  |
| min    |              |          |       | Identical                  |
| max    |              |          |       | Identical                  |
| ceil   |              |          |      | Go-specific implementation |
| floor  |              |          |      | Go-specific implementation |
| pow    |              |          |      | Go-specific implementation |

### 11.3 Collection Filters

| Filter       | Python Jinja2 | Go Miya | Status | Notes             |
|--------------|---------------|-----------|--------|-------------------|
| first        |              |          |       | Identical         |
| last         |              |          |       | Identical         |
| length/count |              |          |       | Identical         |
| join         |              |          |       | Identical         |
| sort         |              |          |       | `None` sorts first (Jinja2 raises an error) |
| reverse      |              |          |       | Identical         |
| unique       |              |          |       | Identical         |
| list         |              |          |       | Identical         |
| slice        |              |          |       | Identical         |
| batch        |              |          |       | Identical         |
| random       |              |          |       | Identical         |
| select       |              |          |       | Identical         |
| reject       |              |          |       | Identical         |
| selectattr   |              |          |       | Identical         |
| rejectattr   |              |          |       | Identical         |
| map          |              |          |       | An unknown filter name yields empty values (Jinja2 raises an error) |
| groupby      |              |          |       | Identical         |
| items        |              |          |       | Dictionary items  |
| keys         |              |          |       | Dictionary keys (miya only) |
| values       |              |          |       | Dictionary values (miya only) |

### 11.4 HTML & URL Filters

| Filter         | Python Jinja2 | Go Miya | Status | Notes     |
|----------------|---------------|-----------|--------|-----------|
| escape/e       |              |          |       | Identical |
| safe           |              |          |       | Identical |
| striptags      |              |          |       | Identical |
| urlencode      |              |          |       | Identical |
| urlize         |              |          |       | Identical |
| urlizetruncate |              |          |       | miya only (not a Jinja2 filter) |
| xmlattr        |              |          |       | Identical |
| forceescape    |              |          |       | Identical |

### 11.5 Date/Time Filters

| Filter        | Python Jinja2 | Go Miya | Status | Notes                      |
|---------------|---------------|-----------|--------|----------------------------|
| strftime      |              |          |       | miya only (not a Jinja2 filter) |
| date          |              |          |      | Go-specific implementation |
| time          |              |          |      | Go-specific implementation |
| datetime      |              |          |      | Go-specific implementation |
| weekday       |              |          |      | Go-specific implementation |
| month_name    |              |          |      | Go-specific implementation |
| age           |              |          |      | Go-specific implementation |
| relative_date |              |          |      | Go-specific implementation |

### 11.6 Utility Filters

| Filter         | Python Jinja2 | Go Miya | Status | Notes                      |
|----------------|---------------|-----------|--------|----------------------------|
| default/d      |              |          |       | Differs on purpose: also replaces `None` and the empty string |
| format         |              |          |       | Identical                  |
| string         |              |          |       | Identical                  |
| attr           |              |          |       | Identical                  |
| filesizeformat |              |          |       | Identical                  |
| pprint         |              |          |       | Output format differs from Python `pprint` |
| tojson         |              |          |       | Compact separators: `{"a":1}` where Jinja2 prints `{"a": 1}` |
| fromjson       |              |          |      | Go-specific implementation |
| dictsort       |              |          |       | Pairs print as lists (`[a, 2]`), where Jinja2 prints tuples |

---

## 12. Built-in Tests (30+ Available)

### 12.1 Type Tests

| Test      | Python Jinja2 | Go Miya | Status | Notes     |
|-----------|---------------|-----------|--------|-----------|
| defined   |              |          |       | Identical |
| undefined |              |          |       | Identical |
| none      |              |          |       | Identical |
| boolean   |              |          |       | Identical |
| string    |              |          |       | Identical |
| number    |              |          |       | Identical |
| integer   |              |          |       | Identical |
| float     |              |          |       | Identical |
| sequence  |              |          |       | Identical |
| mapping   |              |          |       | Identical |
| iterable  |              |          |       | Identical |
| callable  |              |          |       | Identical |

### 12.2 Numeric Tests

| Test        | Python Jinja2 | Go Miya | Status | Notes     |
|-------------|---------------|-----------|--------|-----------|
| even        |              |          |       | Identical |
| odd         |              |          |       | Identical |
| divisibleby |              |          |       | Identical |

### 12.3 String Tests

| Test       | Python Jinja2 | Go Miya | Status | Notes                      |
|------------|---------------|-----------|--------|----------------------------|
| lower      |              |          |       | Identical                  |
| upper      |              |          |       | Identical                  |
| startswith |              |          |       | miya only (not a Jinja2 test) |
| endswith   |              |          |       | miya only (not a Jinja2 test) |
| match      |              |          |       | Regex matching             |
| alpha      |              |          |      | Go-specific implementation |
| alnum      |              |          |      | Go-specific implementation |
| ascii      |              |          |      | Go-specific implementation |

### 12.4 Logical Tests

| Test     | Python Jinja2 | Go Miya | Status | Notes     |
|----------|---------------|-----------|--------|-----------|
| in       |              |          |       | Identical |
| contains |              |          |       | miya only (not a Jinja2 test) |
| sameas   |              |          |       | Identical |
| escaped  |              |          |       | Identical |

### 12.5 Comparison Tests

| Test       | Python Jinja2 | Go Miya | Status | Notes     |
|------------|---------------|-----------|--------|-----------|
| eq/equalto |              |          |       | Identical |
| ne         |              |          |       | Identical |
| lt         |              |          |       | Identical |
| le         |              |          |       | Identical |
| gt         |              |          |       | Identical |
| ge         |              |          |       | Identical |

---

## 13. Global Functions

| Function    | Python Jinja2 | Go Miya | Status | Notes                                |
|-------------|---------------|-----------|--------|--------------------------------------|
| range()     |              |          |       | Identical behavior                   |
| lipsum()    |              |          |       | Lorem ipsum generator                |
| dict()      |              |          |       | Dictionary constructor               |
| cycler()    |              |          |       | Cycle through values                 |
| joiner()    |              |          |       | Join values with separator           |
| namespace() |              |          |       | Create namespace object              |
| url_for()   |              |          |       | URL generation (framework dependent) |
| zip()       |              |          |       | Zip multiple sequences (miya only; not a Jinja2 global) |
| enumerate() |              |          |       | Enumerate with index (miya only); the `start=` keyword is ignored, use `enumerate(items, 1)` |

---

## 14. Advanced Features

### 14.1 Extensions System

| Feature                | Python Jinja2 | Go Miya | Status | Notes                  |
|------------------------|---------------|-----------|--------|------------------------|
| Custom extensions      |              |          |       | Full extension API     |
| Extension dependencies |              |          |       | Dependency resolution  |
| Custom tags            |              |          |       | Custom parser nodes    |
| Extension lifecycle    |              |          |       | OnLoad, OnRender hooks |
| Parser extension       |              |          |       | Custom syntax parsing  |

### 14.2 Error Handling

| Feature              | Python Jinja2 | Go Miya | Status | Notes                     |
|----------------------|---------------|-----------|--------|---------------------------|
| Template exceptions  |              |          |       | Detailed error messages   |
| Line number tracking |              |          |       | Source location in errors |
| Error suggestions    |              |          |      | Go-specific enhancement   |
| Debug mode           |              |          |       | Enhanced debugging        |
| Variable watching    |              |          |      | Go-specific enhancement   |

### 14.3 Performance Features

| Feature                   | Python Jinja2 | Go Miya | Status | Notes                     |
|---------------------------|---------------|-----------|--------|---------------------------|
| Template caching          |              |          |       | LRU cache with TTL        |
| Compiled templates        |              |          |       | AST caching               |
| Concurrent safety         |              |          |       | An `Environment` is safe for concurrent use; configure it (`SetConfig`, `SetLoader`, loader templates) before rendering starts |

### 14.4 Loader System

| Feature            | Python Jinja2 | Go Miya | Status | Notes                     |
|--------------------|---------------|-----------|--------|---------------------------|
| FileSystemLoader   |              |          |       | Load from filesystem      |
| StringLoader       |              |          |       | Load from strings         |
| EmbedLoader        |              |          |      | Go embed.FS support       |
| Template discovery |              |          |      | Advanced template finding |
| LRU caching loader |              |          |      | Performance optimization  |

---

## 15. Implementation Differences

### 15.1 Type System

| Aspect         | Python Jinja2    | Go Miya         | Notes                                      |
|----------------|------------------|-------------------|--------------------------------------------|
| Dynamic typing | Native Python    | Interface{} based | Go uses static typing with interfaces      |
| Method calls   | `string.upper()` | Python string, list and dict methods | Go methods on caller data are not callable from templates; pass computed values as data or register a global function |
| Duck typing    | Full support     | Limited           | Go's type system limitations               |
| Reflection     | Runtime          | Compile-time      | Different reflection models                |

### 15.2 Performance Characteristics

| Aspect               | Python Jinja2   | Go Miya         | Notes                            |
|----------------------|-----------------|-------------------|----------------------------------|
| Template compilation | Runtime         | AST-based         | Go uses parsed AST trees         |
| Concurrency          | GIL limitations | Native goroutines | An `Environment` is safe for concurrent use |

### 15.3 Ecosystem Integration

| Aspect             | Python Jinja2      | Go Miya           | Notes                          |
|--------------------|--------------------|---------------------|--------------------------------|
| Web frameworks     | Flask, Django      | Gin, Echo, etc.     | Different framework ecosystems |
| Standard library   | Rich Python stdlib | Go's focused stdlib | Different library philosophies |
| Package management | pip/poetry         | go mod              | Different dependency systems   |

---

## 16. Migration Guide

### 16.1 Python to Go Migration

**Behaves as in Jinja2** (see the conformance numbers at the top for what is checked):

- Basic template syntax
- Control structures (if/for/with)
- Template inheritance (but see [Limitations](MIYA_LIMITATIONS.md) for `self.block()` and `include` of a child template)
- Macros, `caller()`, imports and includes
- Filter blocks (`{% filter %}`)
- Most filters and tests
- Variable assignments, including `{% set a, b = ... %}`

**Requires Adaptation:**

- Methods on Go values: `obj.Method()` is not callable; pass computed values as data or register a global function (Python string, list and dict methods such as `s.upper()` do work)
- Python-specific filters that miya lacks, or that print differently (see the list at the top and [Limitations](MIYA_LIMITATIONS.md)) → Use Go equivalents
- Custom extensions → Rewrite using Go extension API
- Date/time handling → Use Go time.Time objects

**Not Available:**

- Some Python-specific global functions

### 16.2 Best Practices for Go Miya

1. **Use Go types effectively**: Pass Go structs, maps, and slices
2. **Leverage concurrency**: An `Environment` is safe for concurrent use; configure it (`SetConfig`, `SetLoader`, loader templates) before rendering starts
3. **Cache templates**: Use loader caching for performance
4. **Error handling**: Implement proper error checking
5. **Extensions**: Use Go's type system for custom functionality
