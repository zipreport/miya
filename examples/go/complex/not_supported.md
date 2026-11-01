# Miya Unsupported Features

The maintained list of what miya does differently from Jinja2 is
[docs/MIYA_LIMITATIONS.md](../../../docs/MIYA_LIMITATIONS.md), with the
feature-by-feature table in
[docs/JINJA2_VS_MIYA_FEATURE_MATRIX.md](../../../docs/JINJA2_VS_MIYA_FEATURE_MATRIX.md).
Read those pages first; this page only lists what this example's templates are
most likely to run into.

## Not available

- `{% debug %}` (Jinja2's `jinja2.ext.debug` extension).
- Comprehensions with more than one loop variable (`{k: v for k, v in items}`,
  `[k for k, v in items]`) or more than one `for` clause. Comprehensions with
  one loop variable, list or dict, nested or with an `if`, are a miya extension
  and work; stock Jinja2 rejects all of them.

## Not Jinja2 features either

- `*args`/`**kwargs` in a macro signature. As in Jinja2, extra positional and
  keyword arguments arrive in the implicit `varargs` and `kwargs` variables:
  `{% macro f() %}{{ varargs }}{{ kwargs }}{% endmacro %}{{ f(1, 2, a=3) }}`
  prints `[1, 2]{&#39;a&#39;: 3}` with autoescape on (the default), and
  `[1, 2]{'a': 3}` with it off.
- An `export` modifier on macros.

## Works (listed as unsupported in older versions of this page)

`{% call(item) macro() %}` with arguments, `{% do %}`, recursive macros,
`range(start, stop, step)`, `cycler()`, `joiner()`, `lipsum()`, negative-step
slicing, and `{% break %}`/`{% continue %}`.
