# Changelog

All notable changes to Miya Engine will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

### Fixed

- `Template.Release()` no longer corrupts cached templates. A released `FromString` template used to render its raw source, a released `GetTemplate` template rendered nothing, and releasing a base template changed its children's output.

### Changed

- `Template.Release()` now removes the template from the environment's template cache and leaves it usable. A later `FromString`/`GetTemplate` of the same source or name parses it again. The parser no longer pools AST nodes.

### Deprecated

- `parser.ReleaseAST`, `parser.ReleaseNode`, the `parser.Release*Node` functions, `parser.GetNodePoolStats` and `parser.NodePoolStats` do nothing.

## [v0.2.0]

### Changed

- **Breaking:** the `==`, `!=`, `<`, `<=`, `>`, `>=`, `in` operators compare numbers of any Go numeric type correctly. Previously a float64 (e.g. from JSON) compared wrongly against an int literal: `0 == 0.0` was false and `(1 - 1) > 0` was true. Integers compare exactly. Comparing a string with a number using `<`/`>` is now an error, as in Jinja2, and numeric strings compare as strings (`"10" < "9"`).
- **Breaking:** floats print like Jinja2 in output, `~` concatenation, and the `string`, `join` and `e` filters: no thousands separators, no rounding to two decimals, no scientific notation below 1e16. Whole floats still print without a decimal part. Use `format_number` or `currency` for formatted amounts.

### Fixed

- Includes, macro calls (including default arguments) and recursive loops nested more than 200 levels deep return `runtime.ErrMaxRecursion` instead of crashing the process with a stack overflow. `ignore missing` does not hide it.
- Removed two test-only output rewrites: author links were stripped from templates named `blog_post` (or whose output contains `blog-post`), and empty imported macros returned debug text.

## [v0.1.1]

### Fixed

- `tojson` filter now returns `runtime.SafeValue` instead of a plain string, preventing HTML auto-escaping of JSON output. This matches Jinja2's `tojson` behavior which returns `Markup` (a safe string). Previously, when auto-escaping was enabled, `{{ data | tojson }}` inside `<script>` blocks would produce HTML-escaped entities (e.g., `"` became `&#34;`), resulting in invalid JavaScript since browsers do not decode HTML entities in raw text elements like `<script>`.
- Fixed data race in concurrent template rendering.

## [v0.1.0] - Initial Release

### Added

- **Template engine core**: Lexer, parser, AST, and runtime evaluator pipeline.
- **Template inheritance**: `{% extends %}`, `{% block %}`, `{{ super() }}` with multi-level support.
- **Control structures**: `{% if %}` / `{% elif %}` / `{% else %}`, `{% for %}` with `loop` variable, `{% set %}`, `{% with %}`.
- **Expressions**: Arithmetic, comparison, logical, membership (`in`/`not in`), string concatenation (`~`), ternary (`a if cond else b`), power (`**`), floor division (`//`).
- **70+ built-in filters**: String, numeric, collection, HTML/security, date, and encoding filters.
- **26+ built-in tests**: Type tests (`number`, `string`, `mapping`, etc.), value tests (`defined`, `none`, `true`, `false`), numeric tests (`even`, `odd`, `divisibleby`), container tests (`iterable`, `sequence`), and string tests (`upper`, `lower`).
- **Macros**: `{% macro %}`, `{% import %}`, `{% from ... import %}` with default parameters.
- **Template includes**: `{% include %}` with `ignore missing` support.
- **List and dictionary comprehensions**: `[expr for x in list]`, `{k: v for x in list}`, with filter and expression support.
- **Global functions**: `range()`, `dict()`, `cycler()`, `joiner()`, `namespace()`, `lipsum()`, `zip()`, `enumerate()`, `url_for()`.
- **Filter blocks**: `{% filter upper %}...{% endfilter %}`.
- **Do statements**: `{% do expression %}`.
- **Raw blocks**: `{% raw %}...{% endraw %}`.
- **Autoescape control**: `{% autoescape %}...{% endautoescape %}` with HTML auto-escaping enabled by default.
- **Whitespace control**: `{%-`, `-%}`, `{{-`, `-}}` inline trimming, `trimBlocks` and `lstripBlocks` environment options.
- **Custom delimiters**: Configurable variable, block, and comment delimiters.
- **Undefined variable behaviors**: Silent (default), Strict, and Debug modes.
- **Template caching**: Content-hash-based caching with `sync.RWMutex` protection.
- **AST node pooling**: `sync.Pool` for frequently allocated node types to reduce GC pressure.
- **Evaluator pooling**: Reusable evaluators via `sync.Pool` for concurrent rendering.
- **Inheritance caching**: Configurable TTL-based cache for resolved template hierarchies.
- **Extension system**: Custom tags and extensions via the `extensions` package.
- **Error system**: Enhanced errors with source context, suggestions, stack traces, and 9 error type categories.
- **Template validation**: Rule-based validator for common template mistakes.
- **Template debugging**: Breakpoints, variable watching, and step mode.
- **Filesystem and memory loaders**: Load templates from disk or in-memory maps.
- **Concurrent rendering**: `ConcurrentTemplateRenderer` with worker pools and panic recovery.
- **Zero external dependencies**: Built entirely on the Go standard library.
