# Python Jinja2 vs Miya

This page used to carry its own comparison table; it described an older miya
(about 20 filters, no method calls on lists, no `.items()` loops) and is
replaced by the maintained pages:

- [docs/JINJA2_VS_MIYA_FEATURE_MATRIX.md](../../../docs/JINJA2_VS_MIYA_FEATURE_MATRIX.md):
  feature-by-feature comparison with Jinja2 3.1.6.
- [docs/MIYA_LIMITATIONS.md](../../../docs/MIYA_LIMITATIONS.md): deliberate
  differences, known gaps, and a migration checklist.
- [docs/FILTERS_GUIDE.md](../../../docs/FILTERS_GUIDE.md): the built-in filters.

One difference that matters when porting templates from Python: Go methods on
values passed from Go are never callable from a template, so
`{{ date.strftime("%Y") }}` on a `time.Time` does not work. Use a filter
(`{{ date|strftime("%Y") }}`) or format the value in Go before rendering.
