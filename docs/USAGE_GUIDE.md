# Usage Guide

This guide covers how to use Miya Engine in your Go applications, from basic rendering to high-concurrency web servers.

---

## Table of Contents

1. [Installation](#installation)
2. [Basic Usage](#basic-usage)
3. [Loading Templates](#loading-templates)
4. [Working with Context](#working-with-context)
5. [Registering Custom Filters, Tests, and Globals](#registering-custom-filters-tests-and-globals)
6. [Environment Configuration](#environment-configuration)
7. [Concurrency and Thread Safety](#concurrency-and-thread-safety)
8. [Web Server Integration](#web-server-integration)
9. [Template Caching](#template-caching)
10. [Memory Management](#memory-management)

---

## Installation

```bash
go get github.com/zipreport/miya
```

Miya has **zero external dependencies** beyond the Go standard library.

---

## Basic Usage

### Render a Template from a String

The simplest way to use Miya is to compile and render a template from a string:

```go
package main

import (
    "fmt"
    "log"

    "github.com/zipreport/miya"
)

func main() {
    // 1. Create an environment (reuse this across your application)
    env := miya.NewEnvironment()

    // 2. Compile a template from a string
    tmpl, err := env.FromString("Hello {{ name }}! You have {{ count }} messages.")
    if err != nil {
        log.Fatal(err)
    }

    // 3. Create a context with template variables
    ctx := miya.NewContext()
    ctx.Set("name", "Alice")
    ctx.Set("count", 5)

    // 4. Render
    output, err := tmpl.Render(ctx)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(output) // Hello Alice! You have 5 messages.
}
```

### One-Shot Rendering

For quick one-off renders, use the convenience methods:

```go
env := miya.NewEnvironment()

// Compile + render in one call
output, err := env.RenderString("Hello {{ name }}!", ctx)
```

Or with the default environment (no loader configuration needed):

```go
output, err := miya.RenderString("Hello {{ name }}!", ctx)
```

---

## Loading Templates

Templates can be loaded from the filesystem, from embedded Go resources, or from in-memory strings.

### Filesystem Loader

Load templates from directories on disk. This is the most common setup for web applications.

```go
package main

import (
    "log"

    "github.com/zipreport/miya"
    "github.com/zipreport/miya/loader"
)

func main() {
    // Create a direct template parser (parses templates independently of the environment)
    templateParser := loader.NewDirectTemplateParser()

    // Create the loader with search paths (tried in order)
    fsLoader := loader.NewFileSystemLoader(
        []string{"templates", "views"},
        templateParser,
    )

    // Create the environment with the loader
    env := miya.NewEnvironment(
        miya.WithLoader(fsLoader),
        miya.WithAutoEscape(true),
        miya.WithTrimBlocks(true),
        miya.WithLstripBlocks(true),
    )

    // Now load templates by name
    tmpl, err := env.GetTemplate("page.html")
    if err != nil {
        log.Fatal(err)
    }

    ctx := miya.NewContext()
    ctx.Set("title", "Home")
    output, _ := tmpl.Render(ctx)
    _ = output
}
```

The filesystem loader:

- Searches paths in order, returning the first match
- Recognizes `.html`, `.htm`, `.jinja`, `.jinja2`, `.j2` extensions by default (configurable via `SetExtensions`)
- Caches parsed templates for 5 minutes
- Prevents directory traversal attacks

### Embedded Filesystem Loader

For self-contained binaries, load templates from Go's `embed.FS`:

```go
import "embed"

//go:embed templates/*
var embeddedTemplates embed.FS

func main() {
    templateParser := loader.NewDirectTemplateParser()
    embedLoader := loader.NewEmbedLoader(embeddedTemplates, "templates", templateParser)

    env := miya.NewEnvironment(
        miya.WithLoader(embedLoader),
    )

    tmpl, _ := env.GetTemplate("page.html")
    // ...
}
```

### String Loader (for Testing)

Load templates from in-memory strings, useful for unit tests:

```go
env := miya.NewEnvironment()
templateParser := loader.NewDirectTemplateParser()

strLoader := loader.NewStringLoader(templateParser)
strLoader.AddTemplate("base.html", `
<!DOCTYPE html>
<html>
<body>{% block content %}{% endblock %}</body>
</html>
`)
strLoader.AddTemplate("page.html", `
{% extends "base.html" %}
{% block content %}<h1>{{ title }}</h1>{% endblock %}
`)

env.SetLoader(strLoader)

tmpl, _ := env.GetTemplate("page.html")
ctx := miya.NewContext()
ctx.Set("title", "Hello")
output, _ := tmpl.Render(ctx)
```

### Chain Loader (Fallback Chain)

Try multiple loaders in sequence:

```go
chainLoader := loader.NewChainLoader(embedLoader, fsLoader, strLoader)
env.SetLoader(chainLoader)
```

The first loader that finds the template wins.

---

## Working with Context

### Creating Contexts

```go
// Empty context
ctx := miya.NewContext()
ctx.Set("name", "Alice")

// From an existing map (the map is copied)
ctx := miya.NewContextFrom(map[string]interface{}{
    "title":  "Home",
    "items":  []string{"a", "b", "c"},
    "count":  42,
})
```

### Supported Value Types

You can pass any Go value into the context. Miya handles:

- **Primitives**: `string`, `bool` and numbers of every Go kind (`int32`, `uint8`, `float32`, ...), including named types such as `type Status string`. Pointers behave as the value they point to; a nil pointer is `none`
- **Collections**: slices, arrays and maps of any element type, including `map[interface{}]interface{}` with string keys (what YAML decoders produce). Maps with string keys iterate in sorted order; maps with other keys iterate bools, then numbers, then strings
- **Structs**: exported fields are accessed by name, and `user.name` finds `Name`. An embedded field is reachable only by its exact type name
- **Nested access**: dot notation works through maps, structs and pointers (e.g., `{{ user.address.city }}`)
- **Methods**: Go methods are **not** callable from templates. Pass computed values as data, or register a global function with `env.AddGlobal`
- **Funcs**: a func-typed value in the data is callable (that is how globals work), so do not pass funcs you would not let template authors call

```go
type User struct {
    Name        string
    Email       string
    Age         int
    DisplayName string // computed by the application
}

ctx := miya.NewContext()
ctx.Set("user", User{Name: "Alice", Email: "alice@example.com", Age: 30, DisplayName: "Alice <alice@example.com>"})
ctx.Set("items", []map[string]interface{}{
    {"name": "Widget", "price": 9.99},
    {"name": "Gadget", "price": 19.99},
})
```

```html+jinja
{{ user.name }}           {# "Alice" — struct field #}
{{ user.displayName }}    {# "Alice <alice@example.com>" — struct field computed by the application #}
{{ items[0].name }}       {# "Widget" — slice index + map access #}
```

### Context Scoping

Context supports nested scopes. When the template engine enters a `{% for %}` or `{% block %}`, it pushes a child scope. Variables set in a child scope do not leak to the parent:

```go
ctx := miya.NewContext()
ctx.Set("outer", "visible")

child := ctx.Push()
child.Set("inner", "only here")

val, _ := child.Get("outer") // "visible" — inherited from parent
val, ok := ctx.Get("inner")  // ok == false — child vars don't leak up

parent := child.Pop()        // returns to parent scope
```

### Global Variables

Variables added to the environment via `AddGlobal` are available in every template without needing to be set in the context:

```go
env := miya.NewEnvironment()
env.AddGlobal("site_name", "My Website")
env.AddGlobal("version", "2.1.0")
```

```html+jinja
<footer>{{ site_name }} v{{ version }}</footer>
```

---

## Registering Custom Filters, Tests, and Globals

### Custom Filters

Filters transform values in templates (`{{ value|filter }}`):

```go
env := miya.NewEnvironment()

// Register a filter: func(value, args...) -> (result, error)
env.AddFilter("currency", func(value interface{}, args ...interface{}) (interface{}, error) {
    // Default symbol
    symbol := "$"
    if len(args) > 0 {
        if s, ok := args[0].(string); ok {
            symbol = s
        }
    }
    return fmt.Sprintf("%s%.2f", symbol, toFloat64(value)), nil
})
```

```html+jinja
{{ 19.9|currency }}          {# $19.90 #}
{{ 19.9|currency("EUR ") }}  {# EUR 19.90 #}
```

### Custom Tests

Tests are used in `{% if value is test %}` expressions:

```go
env.AddTest("palindrome", func(value interface{}, args ...interface{}) (bool, error) {
    s := fmt.Sprintf("%v", value)
    for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
        if s[i] != s[j] {
            return false, nil
        }
    }
    return true, nil
})
```

```html+jinja
{% if word is palindrome %}{{ word }} is a palindrome!{% endif %}
```

### Custom Global Functions

```go
env.AddGlobal("now", func(args ...interface{}) (interface{}, error) {
    return time.Now().Format(time.RFC3339), nil
})
```

```html+jinja
<p>Generated at {{ now() }}</p>
```

---

## Environment Configuration

### Available Options

```go
env := miya.NewEnvironment(
    // Template loader (filesystem, embed, string, chain)
    miya.WithLoader(fsLoader),

    // HTML auto-escaping (default: true)
    // When enabled, {{ value }} output is HTML-escaped automatically.
    // Use the |safe filter to bypass: {{ trusted_html|safe }}
    miya.WithAutoEscape(true),

    // Whitespace control (default: false for both)
    // TrimBlocks: remove first newline after a block tag
    // LstripBlocks: strip leading whitespace from block tag lines
    miya.WithTrimBlocks(true),
    miya.WithLstripBlocks(true),

    // Preserve trailing newline at end of template (default: false)
    miya.WithKeepTrailingNewline(false),

    // Undefined variable behavior (default: silent)
    miya.WithStrictUndefined(true),    // raise error on undefined vars
    // OR
    miya.WithDebugUndefined(true),     // render placeholder for undefined vars
    // OR
    miya.WithUndefinedBehavior(runtime.UndefinedSilent), // empty string (default)
)
```

### Custom Delimiters

If the default `{{ }}` / `{% %}` / `{# #}` delimiters conflict with your content (e.g., rendering Vue.js templates):

```go
env := miya.NewEnvironment()
env.SetDelimiters("[%", "%]", "<%", "%>")       // variable, block
env.SetCommentDelimiters("<#", "#>")             // comment
```

---

## Concurrency and Thread Safety

### What Is Thread-Safe

| Component | Thread-Safe? | Notes |
|-----------|:---:|-------|
| `Environment` | Yes | Template cache, filter/test registries protected by `sync.RWMutex` |
| `Template` | Yes | Multiple goroutines can call `Render()` concurrently on the same template |
| `Context` | **No** | Create a new context per render; do not share across goroutines |
| `FileSystemLoader` | Yes | Cache protected by `sync.RWMutex` |
| `EmbedLoader` | Yes | Cache protected by `sync.RWMutex` |
| `StringLoader` | **No** | Not safe for concurrent `AddTemplate` + reads |

### Safe Pattern: Shared Environment, Per-Request Context

This is the recommended pattern for web servers and any concurrent application. The environment and compiled templates are created once and shared; each render gets its own context:

```go
// At startup — create once, share across goroutines
env := miya.NewEnvironment(miya.WithLoader(fsLoader))

// In HTTP handler — called from multiple goroutines concurrently
func handler(w http.ResponseWriter, r *http.Request) {
    // GetTemplate is safe to call concurrently
    tmpl, err := env.GetTemplate("page.html")
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }

    // Create a NEW context for each request — never share contexts
    ctx := miya.NewContext()
    ctx.Set("user", getCurrentUser(r))
    ctx.Set("path", r.URL.Path)

    // Render is safe to call concurrently on the same template
    output, err := tmpl.Render(ctx)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }

    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    w.Write([]byte(output))
}
```

**Key rules:**

1. Create the `Environment` once at startup
2. Call `env.GetTemplate()` freely from any goroutine — it is read-locked and cached
3. Call `tmpl.Render(ctx)` freely from any goroutine — evaluators are pooled per render
4. **Always create a fresh `Context` for each render** — contexts are not thread-safe

### Changing the Environment While Rendering

An `Environment` is safe for concurrent use: templates may be rendered while `AddGlobal`, `AddFilter`, `AddTest` and `AddExtension` are called. Calls to `AddTest` must not run concurrently with each other. `SetConfig`, `SetLoader` and changes to a loader's templates (such as `StringLoader.AddTemplate`) must happen before rendering starts.

`ThreadSafeEnvironment` adds nothing to this: renders never take its mutexes, so the standard `Environment` is all you need.

### Worker Pool: ConcurrentTemplateRenderer

For high-throughput batch rendering (e.g., generating emails, reports), use the worker pool:

```go
tmpl, _ := env.GetTemplate("email.html")

// Create a pool of 8 render workers
renderer := miya.NewConcurrentTemplateRenderer(tmpl, 8)
renderer.Start()
defer renderer.Stop()

// Async single render
resultCh := renderer.RenderAsync(ctx)
result := <-resultCh // result.output, result.err

// Batch render — all contexts rendered in parallel
contexts := make([]miya.Context, 1000)
for i := range contexts {
    contexts[i] = miya.NewContext()
    contexts[i].Set("recipient", recipients[i])
}

outputs, errs := renderer.RenderBatch(contexts)
for i, output := range outputs {
    if errs[i] != nil {
        log.Printf("render %d failed: %v", i, errs[i])
        continue
    }
    sendEmail(recipients[i], output)
}
```

Workers include panic recovery — a panic in one render does not crash other workers.

Start the renderer once. After `Stop`, `RenderAsync` rejects work with a "renderer is stopped" error, and jobs still queued at `Stop` are answered with the same error. `RenderBatch` does not use the worker pool and is independent of `Start` and `Stop`.

### Rate-Limited Rendering

Limit the number of concurrent renders (e.g., to control memory usage):

```go
tmpl, _ := env.GetTemplate("report.html")

// Allow at most 10 concurrent renders
limiter := miya.NewRateLimitedRenderer(tmpl, 10)

// Render blocks if 10 renders are already in progress
output, err := limiter.Render(ctx)
```

### Context Pooling

For very high-throughput scenarios, reuse context objects to reduce allocations:

```go
pool := miya.NewConcurrentContextPool()

// In hot loop
ctx := pool.Get()
ctx.Set("key", "value")
output, _ := tmpl.Render(ctx)
pool.Put(ctx) // returns a clean context to the pool
```

---

## Web Server Integration

Complete example of a web server using Miya:

```go
package main

import (
    "log"
    "net/http"

    "github.com/zipreport/miya"
    "github.com/zipreport/miya/loader"
)

var env *miya.Environment

func init() {
    templateParser := loader.NewDirectTemplateParser()
    fsLoader := loader.NewFileSystemLoader([]string{"templates"}, templateParser)

    env = miya.NewEnvironment(
        miya.WithLoader(fsLoader),
        miya.WithAutoEscape(true),
        miya.WithTrimBlocks(true),
        miya.WithLstripBlocks(true),
    )

    // Register application-wide globals
    env.AddGlobal("site_name", "My App")
}

func handleHome(w http.ResponseWriter, r *http.Request) {
    tmpl, err := env.GetTemplate("home.html")
    if err != nil {
        http.Error(w, "Template not found", 500)
        return
    }

    ctx := miya.NewContext()
    ctx.Set("title", "Welcome")
    ctx.Set("items", []string{"Go", "Templates", "Miya"})

    output, err := tmpl.Render(ctx)
    if err != nil {
        http.Error(w, "Render error", 500)
        return
    }

    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    w.Write([]byte(output))
}

func main() {
    http.HandleFunc("/", handleHome)
    log.Println("Listening on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

### Writing to an io.Writer

For better performance, write directly to the `http.ResponseWriter` instead of building a string:

```go
func handlePage(w http.ResponseWriter, r *http.Request) {
    tmpl, err := env.GetTemplate("page.html")
    if err != nil {
        http.Error(w, "Template not found", 500)
        return
    }

    ctx := miya.NewContext()
    ctx.Set("title", "Page")

    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    if err := tmpl.RenderTo(w, ctx); err != nil {
        log.Printf("render error: %v", err)
    }
}
```

---

## Template Caching

### How Caching Works

- **Named templates** (`GetTemplate("page.html")`): cached by template name
- **String templates** (`FromString(source)`): cached by FNV-1a content hash
- **Loader-level caching**: `FileSystemLoader` caches parsed ASTs for 5 minutes; `EmbedLoader` caches for 24 hours

### Managing the Cache

```go
// Check how many templates are cached
size := env.GetCacheSize()

// Clear all cached templates (e.g., after editing template files in development)
env.ClearCache()

// Invalidate a single template (also clears its inheritance cache entry)
env.InvalidateTemplate("page.html")

// Clear loader-level cache separately
if cachingLoader, ok := env.GetLoader().(loader.CachingLoader); ok {
    cachingLoader.ClearCache()
}
```

### Inheritance Cache

The chain of a template (e.g., `page.html` extends `layout.html` extends `base.html`) is cached separately, with configurable TTLs. Resolved templates are no longer cached, so the second argument of `ConfigureInheritanceCache` has no effect and `ResolvedCache` always reads zero:

```go
import "time"

env.ConfigureInheritanceCache(
    5*time.Minute,   // hierarchy TTL
    10*time.Minute,  // resolved template TTL (no effect)
    1000,            // max cache entries
)

// Monitor cache performance
stats := env.GetInheritanceCacheStats()
fmt.Printf("Hierarchy - Hits: %d, Misses: %d\n",
    stats.HierarchyCache.Hits, stats.HierarchyCache.Misses)

// Clear when needed
env.ClearInheritanceCache()
```

---

## Memory Management

### Releasing one-off templates

Compiled templates are cached by the environment, by name for `GetTemplate` and by source for `FromString`, and the cache does not expire. If you compile many distinct one-off templates, call `tmpl.Release()` when done: it removes that template from the cache, so it can be garbage collected once you drop it. The template stays usable; compiling the same source again re-parses it. `ClearCache()` drops everything.

```go
tmpl, _ := env.FromString("{{ x }}")
output, _ := tmpl.Render(ctx)

// drop this one-off template from the environment's cache
tmpl.Release()
```

### Context Pooling

For high-throughput rendering, reuse context objects:

```go
pool := miya.NewConcurrentContextPool()

for _, item := range largeDataSet {
    ctx := pool.Get()
    ctx.Set("item", item)
    output, _ := tmpl.Render(ctx)
    pool.Put(ctx)
    process(output)
}
```

---

## See Also

- [Template Inheritance](TEMPLATE_INHERITANCE.md) - extends, blocks, super()
- [Control Structures](CONTROL_STRUCTURES_GUIDE.md) - if, for, set, with
- [Filters](FILTERS_GUIDE.md) - 70+ built-in filters
- [Error Handling](ERROR_HANDLING.md) - error types, debugging, validation
- [Miya Limitations](MIYA_LIMITATIONS.md) - known limitations and workarounds
