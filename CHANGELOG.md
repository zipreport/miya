# Changelog

All notable changes to Miya Engine will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

### Added

- `miya.ErrRenderPanic`, wrapped by the error `Render`, `RenderTo` and `Environment.CallMacro` return when they recover from a panic.

### Fixed

- Bad data and malformed templates return errors instead of panicking or crashing the process: `nil` in `+`, `+` on lists of different element types, nil functions (called or iterated), `{% extends %}` with a non-string (also in a parent template), mismatched `-%}`/`-}}`, negative `truncate`/`truncatehtml` length, huge `center`/`pad_left`/`pad_right`/`indent`/`batch` sizes, NaN and infinite `currency`, a crash in the trim-blocks pass, and templates nested more than 500 levels deep (previously a fatal stack overflow). `Render`, `RenderTo` and `CallMacro` recover any other panic as an error wrapping `ErrRenderPanic`; `inheritance.InheritanceEvaluator` does not.
- `Template.Release()` no longer corrupts cached templates. A released `FromString` template used to render its raw source, a released `GetTemplate` template rendered nothing, and releasing a base template changed its children's output.
- Dict literals keep their contents: `{% set d = {'k': 'v'} %}{{ d.k }}` prints `v` (it printed nothing, and every dict literal was empty). Each evaluation builds a new dict, so writes into one no longer carry over to the next render. Keys are stored as strings, as for every miya dict, so `{1: 'a'}[1]` and `{1: 'a'}['1']` find the same entry, and a list key such as `{[1, 2]: 'x'}` is stored under `"[1 2]"` instead of raising an error as Jinja2 does.
- With autoescape on, values that are not strings (named string types, errors, `fmt.Stringer`s, lists, maps, structs) are escaped on output; a `*runtime.SafeValue` is treated as safe; `{{ bytes }}` prints a `[]byte` as text (escaped) instead of `[60 98 62]`. `urlize`, `urlizetruncate`, `nl2br` and `truncatehtml` escape their input text. `xmlattr` output is no longer double-escaped. `tojson` escapes `'`. `{% autoescape %}` blocks keep their setting inside `for` loops and `{% with %}`. `urlize(..., nofollow=true)` no longer adds an extra `nofollow` to every link after the first.
- Concurrent renders on one `Environment` that import templates no longer race or crash (`concurrent map writes`).
- `{% from x import name %}` for a missing name gives an undefined value whose call fails with "does not export the requested name", not placeholder text.
- Macros see their own file's top-level `{% set %}` variables, including block sets (`{% set x %}...{% endset %}`).
- An imported template's own `import`/`from` statements run, so its macros can call the helpers it imports.
- A self-import or mutual import fails with `runtime.ErrMaxRecursion` instead of returning a half-built module.
- `from ... import a as b` works in templates that use `{% extends %}`.
- `with context` / `without context` parse, including after a trailing comma in `from ... import a, with context`.

### Changed

- **Breaking:** `range()` over more than 1,000,000 items, `lipsum()` over 1000 paragraphs or sentences, list `+` producing more than 10,000,000 items, `batch` filling past 1,000,000 items, padding filters past 64 MiB and `indent` output past 64 MiB return an error; templates nested deeper than 500 levels fail to parse. A parent template whose `{% extends %}` is not a string literal is an error (it used to look up a template named `<<DYNAMIC>>`). These limits stop single huge allocations; they do not bound a render's memory or time.
- `Template.Release()` now removes the template from the environment's template cache and leaves it usable. A later `FromString`/`GetTemplate` of the same source or name parses it again. The parser no longer pools AST nodes.
- **Breaking:** a filter, global or data value that relied on a non-string type to bypass autoescape (for example Go's `html/template.HTML`) is now escaped; return `runtime.SafeValue` instead. `filters.SafeValue` is now an alias of `runtime.SafeValue`. `xmlattr` rejects attribute names containing whitespace, `/`, `>` or `=`. `nl2br` and `urlize` escape their input text even with autoescape off (filters cannot see the setting; `urlize` matches Jinja2). A truncated `truncatehtml` result is escaped text, including the `end` argument. Non-string values passed through macros, `caller()`, imports and `{% set %}` blocks are now escaped twice (as plain strings already were) until the Markup change lands.
- **Breaking:** `and`/`or` return one of their operands and skip the right side when the left decides, as in Jinja2: `{{ name or 'default' }}` prints the name, not `true`; `{% if x != 0 and 10 / x %}` no longer raises; templates that printed `{{ a or b }}` expecting `true`/`false` now print a value.
- **Breaking:** imported templates run without the importer's context, as in Jinja2. `{% import %}` and `{% from %}` macros no longer see render variables or variables set before the import. Write `{% import "x.html" as x with context %}` (or `from ... import ... with context`) to keep the old visibility. Each import builds its own namespace, so a loop no longer reuses the first iteration's values.
- **Breaking:** importing a template that does not exist, or one with a syntax error, is an error. Previously both rendered silently.
- **Breaking:** a top-level `{% set %}` in an imported template that fails to evaluate (for example a strict-mode undefined variable) fails the render. Previously the variable silently held the text `[Variable x from file]`.
- Within one render, every import of a template *without* context shares one module, as in Jinja2 (an imported `namespace()` counter keeps counting across imports in a loop). Imports with context are built fresh each time. Unlike Jinja2, modules are never shared between renders.
- `{{ x is defined }}` is false when `x` holds an undefined value (`{% set x = nope %}`), as in Jinja2. In strict mode, printing a name that a `{% from %}` import did not find raises an error.
- `runtime.ImportSystem.LoadTemplateNamespace` uses the context it is given as the module context, without cloning it, writes to it, and expects the loader to return an error for a missing template (`TemplateExists` is not consulted). An import on an evaluator that has no import system (for example one an extension creates with `runtime.NewEvaluator()`) is an error instead of binding placeholder macros.

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
