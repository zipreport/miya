# Miya Engine Limitations and Differences from Jinja2

This document lists the **known limitations and differences** between Miya Engine and Jinja2 3.1.6. Each item was checked by rendering it with miya and with Jinja2.

> **Applies to:** master after v0.2.0, including the changes listed under [Unreleased] in CHANGELOG.md

miya's conformance suite (`tests/conformance/`) holds 289 small templates written to probe edge cases and known bugs; they are not a sample of real templates. At this commit miya matches Jinja2 3.1.6 on 283 of them (same output, or failing in the same phase for the cases that expect an error), after normalising the printing differences listed below. 4 of the remaining cases are deliberate differences. Every divergence and its reason is listed in `tests/conformance/known_divergences.json`. This is a regression net, not a compatibility percentage.

---

## Table of Contents

1. [Deliberate differences](#deliberate-differences)
2. [Differences that are not deliberate](#differences-that-are-not-deliberate)
3. [Comprehension limitations](#comprehension-limitations)
4. [Memory and untrusted templates](#memory-and-untrusted-templates)
5. [Migration checklist](#migration-checklist)

---

## Deliberate differences

- At the top level, booleans print `true`/`false` and `None` prints empty. Inside lists and dicts they print as in Python (`True`, `False`, `None`).
- Whole floats print without `.0` everywhere, including inside lists and after `string`: `{{ [1.0] }}` prints `[1]`.
- `round` rounds half away from zero and prints exactly n decimals: `2.5|round` prints `3` (Jinja2: `2.0`), `2.5|round(2)` prints `2.50`, `2.675|round(2)` prints `2.68`.
- `default` also replaces `None` and the empty string, not only undefined values.
- Dicts iterate and print in sorted key order, not insertion order.
- Integers are 64-bit: `9223372036854775807 + 1` becomes a float.
- Methods on Go values are not callable from templates (`{{ s.Hello() }}` is an error; fields such as `{{ s.Name }}` work). Pass computed values as data or register a global function. Python string, list and dict methods (`'abc'.upper()`, `d.get('a')`) do work.
- `None` sorts before every other value in `sort`, where Jinja2 raises an error.
- `map` with an unknown filter name yields empty values where Jinja2 raises an error.
- `indent` on a safe value escapes its indent string with autoescape off too, as `center` escapes its fill; Jinja2 inserts it as it is there.
- Map key lookup (`[]`, `.get()`, `in`) on string-key maps, dict literals included, uses the key's string spelling: `d[1]` finds a `"1"` key, `d[none]` a `"None"` key, a `[]byte` key finds a typed string-key map's entry, and an undefined key finds the `""` entry (Jinja2: not found). A list key misses instead of raising, and finds a Go array key with equal elements (Jinja2: TypeError, unhashable). In an interface-key map, a key of a named Go string type is found only by a value of that type, not by a plain string (Jinja2: a `str` subclass key equals the string).

---

## Differences that are not deliberate

These are bugs or gaps, not design decisions.

| Template | miya | Jinja2 |
|----------|------|--------|
| `{{ self.block_name() }}` | error: `cannot call non-function value of type *runtime.Undefined` | renders the block |
| `{% include 'child' %}` where `child` has `{% extends 'base' %}` | inheritance is not resolved: `a{% include %}b` prints `aXb` | `a[X]b` (the base wraps the child's blocks) |
| `{% import 'child' as m %}` where `child` has `{% extends 'base' %}` | only the child's top level runs: the parent's top-level macros and `{% set %}` are not exported (`m.bm is defined` is `false`) | the parent's top level runs too and is exported (`True`) |
| `{{ '%s-%s' % (1, 2) }}` | error: `cannot modulo string and []interface {}` | `1-2` |
| `{% do items.append(1) %}` on a list passed in as render data | error: store it first with `{% set items = items\|list %}` | mutates the list |
| `enumerate(items, start=1)` | `start` is silently ignored; use `enumerate(items, 1)` (`enumerate` does not exist in Jinja2) | `enumerate` is undefined |
| `{% autoescape false %}{% block b %}{{ x\|replace("<", s) }}{% endblock %}{% endautoescape %}` with `x = "<"`, `s = Markup("<b>")` | filters inside the block see escaping as on: `&lt;` (the output is escaped; before, `<b>`) | `&lt;b&gt;` |

Printing differences:

- Tuples print as lists: `{{ (1, 2) }}` prints `[1, 2]`.
- `tojson` uses compact separators: `{"a":1}` where Jinja2 prints `{"a": 1}`.
- `dictsort` pairs print as lists (`[a, 2]`), not tuples.
- `pprint` output differs from Python's.

Workarounds:

```html+jinja
{# string formatting: use the format filter #}
{{ '%s-%s'|format(1, 2) }}      {# 1-2 #}

{# mutating a list from render data #}
{% set numbers = numbers|list %}
{% do numbers.append(1) %}
```

---

## Comprehension limitations

List and dict comprehensions are a miya extension; stock Jinja2 rejects them. Inside that extension, two forms do not parse:

```html+jinja
{{ {k: v for k, v in dict.items()} }}                 {# parser error: expected 'in' in dict comprehension #}
{{ [item for cat in categories for item in cat.items] }}  {# parser error: expected ']' after list comprehension #}
```

Single-variable comprehensions with an `if` clause work:

```html+jinja
{{ [u.name for u in users if u.active] }}
{{ {u.name: u.active for u in users} }}
```

Workaround for nested comprehensions:

```html+jinja
{% set all_items = [] %}
{% for category in categories %}
  {% for item in category['items'] %}
    {% do all_items.append(item) %}
  {% endfor %}
{% endfor %}
```

(`{% do %}` also needs the `jinja2.ext.do` extension in stock Jinja2.)

---

## Memory and untrusted templates

- miya bounds recursion of templates and macros and some single results (64 MiB for `join`, `replace`, `format`, repetition, `center`/`indent`, printed containers), but not a render's total memory. Size-limit errors wrap `runtime.ErrOutputTooLarge`.
- A template can allocate without limit (lists built in loops, `~` accumulation, filters such as `tojson`, `urlize`, `regex_replace`, `wordwrap`), and Go's out-of-memory is fatal.
- Render untrusted templates in a separate process with an OS memory limit (cgroup, container limit, or `ulimit -v`). `GOMEMLIMIT` alone does not prevent the crash.
- A template can also build a value nested tens of thousands of levels deep (`{% set ns.l = [ns.l] %}` in a loop). Printing, comparing or serialising it then overflows the Go stack, which is also fatal. Data from `encoding/json` is limited to 10000 levels by the decoder.
- Render data on its own cannot exceed these limits unless the template amplifies it.

---

## Migration checklist

- [ ] No `self.block()` calls
- [ ] No `{% include %}` of a template that itself uses `{% extends %}`
- [ ] No `'...' % value` formatting; use the `format` filter
- [ ] No `.append()` on lists passed in as render data without `{% set x = x|list %}` first
- [ ] `enumerate()` uses positional arguments only
- [ ] No `.items()` unpacking or multiple `for` clauses in comprehensions
- [ ] No Go method calls on your data
- [ ] Output that depends on the deliberate differences above (boolean and float printing, `round`, `default`, dict order) is checked

---

## See Also

- [Feature Examples](https://github.com/zipreport/miya/tree/master/examples/features/) - Working examples of supported features
- [JINJA2_VS_MIYA_FEATURE_MATRIX.md](JINJA2_VS_MIYA_FEATURE_MATRIX.md) - Complete compatibility matrix
- [COMPREHENSIVE_FEATURES.md](COMPREHENSIVE_FEATURES.md) - All supported features with examples
