# Changelog

All notable changes to Miya Engine will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

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
