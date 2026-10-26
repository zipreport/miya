# Changelog

All notable changes to Miya Engine will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

### Added

- `miya.ErrRenderPanic`, wrapped by the error `Render`, `RenderTo` and `Environment.CallMacro` return when they recover from a panic.
- `Environment.SetFilterParams(name, params...)`, `Environment.AddFilterWithParams(name, fn, params...)` and `filters.Param` declare a custom filter's parameter names so templates can pass them as keyword arguments. Keywords reach the filter as ordinary positional arguments; an unfilled earlier position receives its `Param.Default`, which is `nil` when unset. Also `filters.KwArgs` and `runtime.KwargsFilterApplier`.
- `runtime.InheritanceProcessor.ResolveInheritanceWith`, which takes the evaluator for dynamic `{% extends %}`.
- `runtime.CompareNumbers`, `runtime.Equal`, `runtime.GroupTuple`, `filters.TestLookup` and `(*filters.FilterRegistry).SetTestLookup`.
- `runtime.Markup`, `runtime.AsMarkup`, `runtime.Unmark`, `runtime.EscapeMarkup`, `runtime.KeepMarkupFunc` and `runtime.KwArgs` (`filters.KwArgs` is now an alias of it). Go code that type-asserted `filters.SafeValue` on a `|safe` result should use `runtime.AsMarkup`.
- String methods `upper`, `lower`, `strip`, `lstrip`, `rstrip`, `split`, `replace`, `startswith`, `endswith`, `format`, `title`, `capitalize`, `join`, `count`, `find`. `str.format` supports `{0[key]}`, `{0.attr}`, `!r` and `!s`. On a safe (`Markup`) string they return `Markup` and `format`, `replace` and `join` escape the arguments that are not safe.
- `d.get(key, default)`, and `get`/`items`/`keys`/`values` on string-to-string maps.
- List methods `index` and `count`, plus `append`, `extend` and `pop` on lists stored with `{% set %}`, `{% with %}` or in a `namespace()` (`{% do ns.items.append(x) %}`). The mutating methods copy the list on the first write and never change the application's data. On any other list they fail with a hint to store it first (`{% set items = items|list %}`).
- `'ab' * 3` and `[1] * 2` repeat.
- `runtime.ErrTemplateNotFound` and `runtime.TemplateNotFoundError` (with the missing template's `Name`), returned for templates the loader cannot find. A custom `runtime.TemplateLoader` that returns its own errors keeps the old rule (`TemplateExists` decides).
- `WithTruncateLeeway(n)` environment option: the default `leeway` of `truncate` (Jinja2's `truncate.leeway` policy). Also `filters.TruncateFilterWithLeeway`, `(*filters.FilterRegistry).Replace` and `runtime.RoundedFloat`.

### Fixed

- Typed Go maps, slices and arrays work in templates: `m.key`, `m['key']`, `m[1]`, `for k in m`, `for k, v in m`, `m.keys()`, negative indexes, and the `join`, `reverse`, `batch`, `items`, `keys`, `values` and `dictsort` filters. Maps with string keys decoded from YAML (`map[interface{}]interface{}`) work too. Maps with non-string keys iterate in a fixed order (bools, numbers, then strings).
- Maps with a named key type, unhashable keys in `in` or `m[...]`, and struct fields reached through a nil embedded pointer no longer crash the render; a pointer that points to itself no longer hangs it.
- Arithmetic (`+ - * / **`, unary `-`) works with every Go numeric type and named numeric types.
- `user.name` finds the struct field `Name`, as documented.
- Indexing, slicing, iterating, `first` and `last` on strings work on characters instead of UTF-8 bytes: `{{ "açúcar"[1] }}` prints `ç`, not `Ã`.
- Slices with a negative step follow Python's bounds: `{{ s[10::-1] }}` no longer panics on a string or list shorter than 11, and `s[:-10:-1]` includes the first element.
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
- Filters honour keyword arguments, bound by name to Jinja2's parameters: `sort(reverse=true)`, `map(attribute="name", default="-")`, `sum(attribute="price", start=10)`, `round(precision=1)`, `join(d=", ", attribute="n")`, `int(base=16)`, `truncate(length=…)` and more. They used to be silently ignored.
- `trim(chars=none)` strips whitespace, as in Jinja2.
- Custom filters and tests, and keyword arguments to filters, work inside `{% autoescape %}` blocks. A custom filter or test used there failed with `unknown filter` / `unknown test`.
- `'%d'|format(x)` and `'%i'|format(x)` accept float operands, such as JSON numbers, truncating them as Python does.
- `{% for … if … %}…{% else %}` runs the else block when the filter leaves no items.
- `loop.depth` counts recursion levels only (nested ordinary loops are each depth 1; recursive calls go 1, 2, 3…), and `loop.depth0` exists.
- Recursive loops apply their `if` filter at every level.
- A `for` loop in a template without a parent that calls `super()` keeps its `if` filter and `recursive` flag. (Under `{% extends %}`, the same needs plan 034's inheritance fixes.)
- The `loop` object is no longer reused across loops and renders (a stored `loop` showed another loop's values, and concurrent renders raced on it); `AddGlobal`, `AddExtension` and `ConfigureInheritanceCache` no longer race with renders; an extension is visible only after its `OnLoad` succeeds; inheritance-cache cleanup and loader statistics are race-free; `NewTemplateCache` with a negative size no longer panics; `ConcurrentTemplateRenderer` no longer panics on a second `Stop`, rejects `RenderAsync` before `Start`, and answers jobs pending at `Stop`.
- Templates from miya's loaders (`FileSystemLoader`, `EmbedLoader`, `StringLoader`, `ChainLoader`, `LRUFileSystemLoader`) are compiled with the environment's settings: `trim_blocks`, `lstrip_blocks`, `{{-`/`-}}`, custom delimiters and extensions used to be ignored for them (they were, and still are, honoured for GetSource-only loaders).
- **Security:** `FileSystemLoader` with `followLinks` false (the default) reads templates through `os.Root`, so no path or symlink can leave the search path. Before, only a symlink in the last path component was rejected, so `linkdir/x.html` could read files outside the template folder (on Windows a directory junction was not detected at all).
- Template names are checked per segment, like Jinja2's `split_template_path`: `a..b.html` is a valid name, any `..` segment is rejected (`a/../b` used to be accepted), and on POSIX a backslash is an ordinary character.
- `LoadTemplate` on a loader without a parser returns an error instead of panicking.
- `FileSystemLoader.LoadTemplate` re-parses a template whose file modification time changed. This affects direct `LoadTemplate` callers only: an `Environment` caches compiled templates itself and does not re-read files (use `ClearCache`/`InvalidateTemplate`).
- Macro keyword arguments bind by name for local, imported and from-imported macros; defaults can use earlier parameters; `varargs` and `kwargs` work; a parameter without a default and not passed is undefined instead of an error.
- Imported macros can call themselves and macros from their own file, and work in `{% call %}`.
- `{% call(args) macro() %}` parses and passes arguments to the block; `caller()` is rendered on each call.
- An `{% autoescape %}` block's setting applies inside macro and call-block bodies defined in it.
- Runaway macro recursion reports the depth limit once instead of wrapping it at every level.
- Data containing `%}` followed by blank lines no longer panics with trim_blocks.
- `{% elif %}` branches of a base template were dropped when a child extended it.
- Two `FromString` templates extending different parents rendered the first one's output.
- `ClearCache` now also clears the inheritance cache, so changed parent templates are reloaded when the loader serves the new source. A loader's own cache (`FileSystemLoader`, `EmbedLoader`, `LRUFileSystemLoader`) is separate: call its `ClearCache` too. `InvalidateTemplate(parent)` does not reach the cached hierarchies of templates that extend `parent`; call `ClearCache` after changing a parent.
- Rendering a template that uses `extends` no longer formats the whole context on every render (9.8 ms → 8.5 µs with a 20 000-row context).
- `map("filtername", args...)` applies the filter to each item, keyword arguments included.
- `groupby` groups have `.grouper` and `.list`, the grouper keeps its type, and `groupby(attr, default=...)` works; `map(attribute="grouper")`, `selectattr("grouper", ...)` and `sort(attribute="grouper")` work on groups.
- `sort(attribute="a,b")` sorts by several attributes.
- The `true` and `false` tests exist (`selectattr("paid", "true")`); the `in` test compares numbers by value.
- Collection filters accept strings, dicts (their keys, sorted) and typed Go slices.
- The `eq`/`equalto`/`ne` tests compare like `==` (JSON `1` equals `1`, lists element by element); the aliases `==`, `!=`, `<`, `<=`, `>`, `>=`, `lessthan`, `greaterthan` exist.
- `{% filter %}` blocks no longer write filter arguments unescaped (`{% filter format(u) %}%s{% endfilter %}`).
- **Breaking:** in non-HTML escape contexts (`ContextWrapper` with JS, CSS, URL or JSON), only `|safe`/`|marksafe` results bypass the context escaper. `nl2br`, `urlize`, `|e`, `tojson` and the other HTML filters are escaped for the context; `{{ d|tojson }}` in a JS context, raw before, is now JS-escaped. Use `|tojson|safe` to emit JSON raw there.
- An `{% autoescape %}` block inside a non-HTML `ContextWrapper` keeps that context's escaper.
- In XML/XHTML contexts, concatenation, `join` and filter arguments escape with the XML escaper.
- An imported template (also `with context`) escapes by its own setting, not the importer's `{% autoescape %}` block.
- A recursive `loop()` result is safe or plain according to the context of the loop itself, not of the call site (`{% autoescape true %}{{ loop(x.c) }}{% endautoescape %}` in an autoescape-off loop escapes the inner output).
- In non-HTML escape contexts, Markup items in a list returned by a filter (`[u]|map('e')`) are plain strings, so the context escaper runs on them.
- `joiner()` keeps a safe separator (`joiner('<br>'|safe)`) instead of falling back to `, `.
- Tests take one argument without parentheses (`x is divisibleby 3`); `is in`, `is true`, `is false`, `is filter` and `is test` work; a test registered as `filter` or `test` replaces the built-in meaning.
- Keywords can be attribute names (`d.if`, `d.block`) and, except `and`/`or`/`not`/`is`/`in`/`if`/`else`, variable names in expressions and `{% set %}` (not yet `for` targets, macro parameters or keyword-argument names); identifiers may be non-ASCII (not NFKC-normalized); `b.0` and `L.0.1` index.
- Number literals accept `_`, `0x`/`0o`/`0b` and exponents (`1e3`); adjacent string literals join; `(1, 2)`, `()`, `set a, b = 1, 2`, `set (a, b) = ...` and `for v in 1, 2` parse; `{{ a if b }}` without `else` is undefined when `b` is false; `[x for x in items if x > 5]` (and the dict form) filter.
- A comparison chain evaluates each operand once and stops at the first false comparison.
- `include` accepts Jinja2's `ignore missing` before `with context`/`without context`; `without context` hides the render's variables.
- Deeply nested `not`/`-` chains return `template nesting too deep` instead of crashing.
- Long flat chains (`1 + 1 + ...`, `a or b or ...`, `~`, `**`, `x|f|f|...`, `a.b.b...`) count towards the 500-level parse limit and return `template nesting too deep` instead of overflowing the stack when the template is rendered.
- Printing a self-referencing list or dict prints `[...]` instead of crashing the process. Printing one container is capped at 64 MiB.
- Strict mode: `default`, `d`, `is defined`, `is undefined` and `is none` accept undefined values; `{{ a|default }}` prints empty.
- `x['k'] is defined` and `x['k']|default(…)` no longer fail when `x` is `None` or not subscriptable; the item counts as undefined, as in Jinja2.
- `include none ignore missing` renders nothing.
- `is defined` is true for a key whose value is `None`, and `None is undefined` is false.
- An undefined value is an empty sequence for `length`, `join`, `sum`, `sort`, `list`, `in`, `items` and the other collection filters (by filter name, so a custom filter registered under one of these names also receives `[]`); `max`/`min`/`random` of it are undefined; `tojson` of it is an error instead of internal fields.
- `include` accepts a list of names and renders the first that exists.
- `{% with %}` without assignments is accepted.
- `lower`/`upper` tests follow Python (`'' is lower` is false, numbers are neither); `true is number`; a dict is a sequence.
- `int` parses `'3.7'`, `' 7 '`, `'1e3'`, `'1_000'` and base prefixes like Jinja2, and returns the default for floats outside the 64-bit range (they printed `-9223372036854775808`); an invalid base falls back like Jinja2 instead of erroring; `float` accepts surrounding whitespace and returns `inf` for `'1e400'`.
- `format_number` returns an error for infinite or NaN values instead of printing `+,Inf`.
- `center` defaults to 80 and pads like Python.
- `format_number` and `currency` no longer put a separator after the minus sign (`-123456|format_number` printed `-,123,456`).
- `}}`, `%}` and custom end delimiters no longer end a tag while a `(`, `[` or `{` opened in it is still open, as in Jinja2: `{{ {'a': {'b': 2}}.a.b }}` renders `2`, and a block end of `}` works with dict literals.

### Changed

- **Breaking:** `wordwrap(width, break_long_words, wrapstring, break_on_hyphens)` follows Python's `textwrap`, as Jinja2 does. Long words are broken (pass `break_long_words=false` to keep them), words split after hyphens, runs of spaces inside a line are kept, and `width` defaults to 79. The second argument is `break_long_words`; it used to be `break_on_hyphens`.
- **Breaking:** a nil pointer prints as empty, is false in `{% if %}`, is `none`, and is treated like `none` by `default`; it printed `<nil>` and was true. A non-nil pointer to a string, number, bool, slice or map behaves as the value it points to (it printed a memory address). Zero values of named bool, string and number types (`type Flag bool`) are false in `{% if %}` and in `default(x, true)`, named bools and strings compare equal to plain ones, and named strings and numbers pass `is string` / `is number`.
- **Breaking:** booleans are numbers in arithmetic and comparison, as in Python: `true + 1` is `2` and `true == 1` is true (both used to fail or be false).
- Go methods are not callable from templates. The docs used to promise zero-argument method calls, which never worked; they now say so. Pass computed values as data, or register a global function.
- **Breaking:** iterating a dict yields its keys, as in Jinja2: `{% for k in d %}` (and comprehensions, `{% set a, b = d %}`) used to yield the values. Use `d.values()` for values. `{% for k, v in d %}` still yields (key, value) pairs. Loops that unpack list-of-dict items into two variables (`{% for a, b in list_of_dicts %}`) now get each dict's two keys, as in Jinja2; use `item.field` instead.
- Dicts are walked in sorted key order (`for`, `.items()`, `.keys()`, `.values()`), so output no longer changes between renders. The order is code-point order, as Python's `sorted()`: uppercase before lowercase, accented letters after `z`, and `"10"` before `"9"`. This is a deliberate difference from Jinja2, which uses insertion order: Go maps have none.
- `dictsort` is deterministic: keys (or values) that compare equal case-insensitively are ordered by code point, then by key. Jinja2 orders such ties by insertion order.
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
- **Breaking:** a keyword argument a filter cannot honour is an error instead of being ignored: an unknown name, a Jinja2 parameter miya does not implement yet (`truncate(leeway=…)`, `indent(blank=…)`, `groupby(default=…)`), or any keyword on a filter without declared parameters, including custom filters. Declare a custom filter's parameters with `Environment.SetFilterParams` or `Environment.AddFilterWithParams`; the error message says so.
- **Breaking:** `sum`, `min` and `max` take positional arguments in Jinja2's order: `sum(attribute, start)` and `min/max(case_sensitive, attribute)`. `l|sum(10)` used to mean start=10 and is now an error; write `l|sum(start=10)`. `l|max("n")` used to mean attribute="n"; write `l|max(attribute="n")` (until plan 039 lands, this still returns the attribute's value rather than the item, as before).
- **Breaking:** a positional filter argument after a keyword argument (`join(d="-", ",")`) is a syntax error, as in Jinja2. A repeated keyword still binds the last value, as in Jinja2.
- **Breaking:** `//` and `%` follow Python: they floor instead of truncating (`-7 // 2` is `-4`, `-7 % 3` is `2`), keep float operands (`7.5 % 2` is `1.5`, `10 // 0.3` is `33`), accept every Go numeric type, and reject strings (`"10" // 3` used to be `3`). With JSON data, where every number is a float, `n // 2` is now a float, so `(n // 2) is integer` is false, as in Jinja2.
- **Breaking:** each for-loop iteration starts from the scope outside the loop, as in Jinja2: a `{% set %}` in the loop body no longer carries over to the next iteration, and a recursive `loop(...)` call no longer sees the calling iteration's `set`s. Use `namespace()` to accumulate across iterations.
- **Breaking:** assigning an attribute (`{% set x.y = ... %}`) works only on `namespace()` objects, and item assignment (`{% set x[k] = ... %}`) is an error, as in Jinja2; templates can no longer modify the caller's maps, slices, structs, or Go values that merely have `Get`/`Set` methods (embed `runtime.NamespaceMarker` to opt a type in). `NewTemplateCache(0)` disables the cache and a negative size means unbounded.
- **Breaking:** `Environment.GetTemplate` reads every template through `GetSource` and compiles it itself; an `AdvancedLoader`'s `LoadTemplate` is no longer used by the environment. Custom `AdvancedLoader`s must return the template text from `GetSource` for every name `LoadTemplate` accepts.
- **Breaking:** a lexer configuration set on a loader's parser (`loader.NewDirectTemplateParserWithConfig`, `SetLexerConfig`) no longer affects rendering through an `Environment`. Set delimiters and whitespace options on the `Environment` instead.
- **Breaking:** with `followLinks` false, a symlink inside the search path is followed only if it is relative and resolves inside the search path. An absolute symlink, or one that leaves the folder, is "template not found" (set `SetFollowLinks(true)` to follow every link, as before for directories). A relative symlinked *file* inside the folder is now accepted; it used to be rejected.
- **Breaking:** in a `ChainLoader`, a template that exists in an earlier loader but does not compile is an error; it used to fall through to a later loader's copy.
- `LRUFileSystemLoader`'s size and TTL bound only its own `LoadTemplate` cache; they do not bound what an `Environment` caches.
- **Breaking:** a macro body no longer sees the caller's local variables (loop variables, and `{% set %}` inside the calling macro or loop body). As in Jinja2, it sees its parameters and the scope where it was defined. Pass values as arguments. A `{% set %}` inside a `{% block %}` is still visible to macros called from that block, because miya blocks share the top-level context (unchanged).
- **Breaking:** an unknown keyword argument, a keyword that repeats a positional argument, or more positional arguments than parameters is an error, unless the macro uses `kwargs`/`varargs`.
- **Breaking:** a missing argument without a default is undefined, not an error.
- **Breaking:** `{% call %}` on a macro that never calls `caller()` is an error.
- **Breaking:** `caller()` inside a call block nested in a macro no longer reaches the outer macro's caller; it fails with `no caller defined`, as in Jinja2.
- **Breaking:** `{% call %}` must be followed by a call (`{% call w()|upper %}` is a parse error, `expected call`), and a macro or call-block parameter named `caller` must have a default.
- **Breaking:** a parameter named `varargs` or `kwargs` is an ordinary parameter and turns the catch-all off, as in Jinja2.
- **Breaking (Go):** macros are `*runtime.Macro` values (`Call(args, kwargs)`), so Go code that type-asserted a macro to a `func` must use the new type. `{{ m }}` prints `<Macro 'm'>`. Each lookup of an imported macro returns a new value, so `m.s is sameas m.s` is now false.
- **Breaking (Go):** a Go function used in `{% call %}` receives the caller differently. A `func(...interface{})` gets a trailing `map[string]interface{}` holding `caller`; a `func(Context, ...interface{})` no longer finds `caller` in its context.
- **Breaking:** whitespace control is done by the lexer and matches Jinja2:
  - a comment on its own line no longer joins the lines around it, and inline comments keep the spaces around them;
  - `-`, trim_blocks and lstrip_blocks no longer act inside string literals, `{% raw %}` or data;
  - lstrip_blocks applies on the first line;
  - `{%+`, `+%}`, `{#+` and `+#}` are supported;
  - `\r\n` and `\r` in templates become `\n`;
  - one trailing newline of the template is removed unless keep_trailing_newline is set (before, only when another whitespace option was on);
  - `-` markers now work in templates loaded through `StringLoader` or `FileSystemLoader`.
- **Breaking:** `{% raw %}` content is output verbatim; it used to lose spaces, quotes and comments.
- **Breaking:** `Render` no longer rewrites its output: whitespace inside `<option>` is kept, and data containing `%}` or `{%` keeps its blank lines and indentation. `Render` and `RenderTo` now produce the same output.
- **Breaking:** the `whitespace` package is removed.
- **Breaking:** CRLF and CR line endings in templates always become LF. There is no option to keep them (miya has no `newline_sequence`), so templates that must emit CRLF (MIME, CSV, ICS) need it produced from data or a filter.
- **Breaking:** code that builds its own `lexer.LexerConfig` — `loader.NewDirectTemplateParserWithConfig`, `DirectTemplateParser.SetLexerConfig`, `extensions.CreateTokensFromString` — now gets that config's trim_blocks, lstrip_blocks and keep_trailing_newline applied (they were ignored), CRLF normalized, and the template's trailing newline removed unless the config sets `KeepTrailingNewline`. Templates rendered through an `Environment` use the environment's options instead.
- `{% raw +%}` is rejected with an error, as in Jinja2.
- **Breaking:** blocks inside `if`/`elif`/`else`, `for`, `with`, `macro`, `filter`, `set` blocks, `call` and `autoescape` can be overridden by a child template, and `super()` works there. A block nested inside a child's block override also overrides the parent's block of that name.
- **Breaking:** a child template's top-level code runs before its parent's, as in Jinja2: its `set`, `macro`, `import` and `if` statements take effect, and text before `{% extends %}` is output. After `{% extends %}`, `{{ }}` and text are still not output (and not evaluated), but `{% include %}`, `{% call %}`, `filter` blocks and blocks inside `for`/`with`/`autoescape` now run and print before the parent's output; a missing include there is now an error.
- **Breaking:** a root template's `{% import %}`/`{% from %}` is no longer moved to the top: a block that uses an import written after it fails, as in Jinja2. Move imports above the blocks that use them.
- **Breaking:** two blocks with the same name in one template are an error (`block 'a' defined twice`), also when they sit in different `if`/`else` branches, as in Jinja2; such templates now fail to load.
- **Breaking:** a dynamic `{% extends expr %}` follows the environment's undefined behaviour: in strict mode an undefined name in `expr` is an `UndefinedError`.
- `{% block name scoped %}` is accepted.
- **Breaking:** `sort`, `min`, `max` and `groupby` compare numbers numerically (`true`/`false` count as 1/0), strings case-insensitively (unless `case_sensitive`) and lists element by element, like Jinja2; comparing a number with a string is an error that names the attribute. `[10, 9, 2]|sort` used to give `10, 2, 9`.
- `None` (JSON `null`) and missing attributes never fail a sort: `sort` and `groupby` put them first, as one group (last with `reverse`), and `min`/`max` skip them. This deliberately differs from Jinja2, which raises `TypeError`.
- **Breaking:** `unique` is case-insensitive by default and keeps `1` and `"1"` apart.
- **Breaking:** `min`/`max` with `attribute=` return the item, not the attribute; on an empty list they return undefined (prints nothing) instead of failing.
- **Breaking:** `slice(n, fill_with)` splits a list into n columns, as in Jinja2. It used to slice from index n.
- **Breaking:** `select`, `reject`, `selectattr` and `rejectattr` apply the named test (`select("odd")`, `selectattr("n", "gt", 1)`); they used to ignore it. An unknown test name is an error.
- **Breaking:** `map("x")` applies the filter `x` to items that have no attribute `x` (`words|map("upper")`). Items that have the attribute still map it (a miya extension; Jinja2 always applies the filter). With keyword arguments, `map("round", precision=1)` always applies the filter.
- **Breaking:** attribute names are paths, as in Jinja2: `attribute="c.name"` looks up `name` inside `c`, and `"0"` indexes a list. A map key that itself contains a `.` is no longer found.
- `==` and the `eq` test still do not equate `true` with `1` (Python does), while `sort` and `unique` treat booleans as 0/1.
- **Breaking (Go API):** `SortFilter` and `UniqueFilter` return `[]interface{}` for `[]string` input; `GroupByFilter` returns `[]interface{}` of `runtime.GroupTuple` instead of `[][]interface{}`.
- **Breaking:** under autoescape, the output of macros, `caller()`, recursive `loop()`, `{% set %}...{% endset %}` and `{% filter %}` blocks is safe Markup, as in Jinja2. It is no longer escaped a second time, so templates can drop `|safe` on macro calls (keeping it is harmless).
- **Breaking:** `|safe`, `|e`/`|escape`, `|forceescape`, `|urlize`, `|nl2br`, `|truncatehtml`, `|xmlattr`, `|tojson` (and the other safe-producing filters) return `runtime.Markup` instead of `filters.SafeValue` or `string`. `{{ "<"|e }}` is no longer double-escaped.
- **Breaking:** a `{% filter %}` block's result is escaped unless it is Markup. Jinja2 writes it raw. A custom filter that produces HTML for a filter block must return `runtime.Markup` or `runtime.SafeValue`.
- Filters applied to a safe value escape their arguments where the result stays safe (`replace`, `format`, `center`, `default`, ...), as markupsafe does: `{{ ('<b>%s</b>'|safe)|format(u) }}` escapes `u`.
- A macro's output is safe only if autoescape was on where the macro is defined. Jinja2 decides at the call site.
- Go filters, tests and Go global functions receive a plain `string` for safe values; macros and the built-in `namespace`, `dict`, `cycler` and `joiner` keep receiving Markup.
- A Go filter that returns `runtime.SafeValue` or `*runtime.SafeValue` still works; the value becomes Markup.
- **Breaking:** operator precedence follows Jinja2: `-2 ** 2` is `4`, `2 ** 3 ** 2` is `64`, `-x|abs` is `(-x)|abs`, and comparisons chain (`a < b < c` is `a < b and b < c`, so `"apple" in fruits == true` is `false`).
- **Breaking:** `~` binds tighter than `+` and `-`: `'Page ' ~ n + 1` is `('Page ' ~ n) + 1`, which prints `Page 11` for `n = 1` (it printed `Page 2`), and `a + b ~ c` is `a + (b ~ c)`. Write `'Page ' ~ (n + 1)`.
- **Breaking:** `is` applies to the operand right before it: `x + 1 is divisibleby(3)` is `x + (1 is divisibleby(3))` and `a == 1 is sameas(true)` is `a == (1 is sameas(true))`. Write `(x + 1) is divisibleby(3)`.
- **Breaking (extensions):** `ParserAdapter.ParseArguments` uses the new expression grammar: adjacent string arguments join (`{% tag "a" "b" %}` gets one argument, `"ab"`), and keywords such as `as` are names.
- To find affected templates: `grep -nE '~[^}%]*[-+]|[-+][^}%]*~' -r templates/` (`~` mixed with `+`/`-`) and `grep -nE '[-+*/%<>=][^}%]* is ' -r templates/` (`is` after an operator).
- **Breaking:** lists and dicts print like Python: `['a', 'b']`, `{'k': 1}`, with `True`/`False`/`None` inside them (top-level `true`/`false`/empty are unchanged). Dicts print in sorted key order. Under autoescape the printed form is HTML-escaped, as in Jinja2.
- **Breaking:** values with an `Error()` or `String()` method print through it inside lists and dicts too, and a `[]byte` inside a list prints as a quoted string.
- **Breaking:** `*` with a string literal repeats it, as in Python, even when it looks like a number: `'0' * 3` is `000` and `'3' * 2` is `33` (it used to be `6`). A numeric string from data still multiplies as a number.
- `d.items()` returns a list of `[key, value]` pairs (in sorted key order), so `|list`, `|length` and printing work.
- **Breaking:** `{% include %}` renders in a copy of the includer's context: `{% set %}` inside the included template no longer changes the includer.
- **Breaking:** `include … ignore missing` ignores only a template that does not exist. An included template that fails to compile or to render, or that includes a missing template itself, is an error.
- **Breaking:** `{% with a = …, b = … %}` evaluates every value in the scope outside the block, as Jinja2 does: `b = a` sees the outer `a`. The order used to be random.
- **Breaking:** `lt`/`le`/`gt`/`ge` tests fail on `None` and on a string compared with a number; `even`/`odd`/`divisibleby` fail on strings and no longer truncate floats.
- **Breaking:** in strict mode, `o['missing']` and an out-of-range index are errors, and so is `a.b is defined` when `a` is undefined.
- **Breaking:** `truncate` follows Jinja2. Strings up to 5 characters over the length are kept whole (`leeway`), so output can be up to `length + 5` characters; pass `leeway=0` where the text must fit, such as a table cell in a PDF, or set it for the whole environment with `miya.WithTruncateLeeway(0)`. The end marker counts toward the length, the default length is 255, and a length shorter than the marker is an error. The word cut no longer stops at half the length: `'A Supercalifragilistic order'|truncate(16)` is `A...`.
- **Breaking:** `indent(width, first, blank)`: the third argument is Jinja2's `blank` (it used to be an indent string; pass a string as `width` instead), blank lines are not indented unless `blank`, and `width` defaults to 4.
- **Breaking:** `round` with a precision returns a number instead of a string.
  - It prints exactly as before: `10|round(2)` is `10.00`, and rounding is half away from zero.
  - It now works in comparisons, arithmetic, `is number` and `|int`.
  - `tojson` emits `2.68`, not `"2.68"`.
  - Go code calling `filters.RoundFilter` gets a `runtime.RoundedFloat`; so do custom filters and global functions that receive a rounded value.
  - Templates that treated the result as a string change: `x|round(2) == '0.00'` is now false (compare with `0`), `x|round(2) < 'a'` is an error, `|length` is an error, `|first` and `|list` no longer split it into characters, and `is string`/`is iterable`/`is sequence` turn false.
  - `inf|round(2)|tojson` is an error (JSON has no infinity; it was `"+Inf"`).
  - `round(0)` (or no precision) of a value of 1e16 or more prints in exponent form, like any float (`1e20|round` prints `1e+20`; it printed `100000000000000000000`).
  - An unknown `method` is an error. A precision above 308, or one that overflows the value, returns it unchanged (it printed `NaN` or `+Inf`), and a value whose scaled form is beyond 2^53 is no longer distorted (`1e20|round(3)` printed `99999999999999983616.000`).
- `filesizeformat` prints `kB` and `1 Byte`, truncates below 1 kB, and goes up to YB/YiB.
- Unbalanced brackets are now reported by the lexer, with Jinja2's wording: `{{ (1 }}` gives `lexer error in template ...: unexpected '}', expected ')'` (was `parser error ...: expected ')' after expression`), and a stray closer such as `{{ x) }}` or the typo `{{ x }<div>` gives `unexpected ')'` / `unexpected '}'`. A bracket left open at the end of the template, or at an end delimiter that cannot close it (such as `%>`), is reported where it was opened: `unexpected '%>', expected ')' (opened at line 2, column ...)`. Code that matches on the old error text needs updating.
- Custom extension tags whose arguments contain unbalanced punctuation (for example `:-)`) now fail in the lexer. Jinja2 rejects them too.
- Error columns on lines after the first are one higher than the actual column (an existing quirk, unchanged here); line numbers are exact.

### Deprecated

- `parser.ReleaseAST`, `parser.ReleaseNode`, the `parser.Release*Node` functions, `parser.GetNodePoolStats` and `parser.NodePoolStats` do nothing.
- `filters.SelectFilter`, `RejectFilter`, `SelectAttrFilter` and `RejectAttrFilter`: templates no longer use them (the registered filters apply tests); they keep their old value-comparing behaviour for direct Go callers.

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
