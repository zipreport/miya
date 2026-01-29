# Template Inheritance Guide

Template inheritance is one of Miya Engine's most powerful features, allowing you to create reusable base templates and extend them with specific content.

> **Working Example:** See `examples/features/inheritance/` for a complete, working example.

---

## Table of Contents

1. [Basic Concepts](#basic-concepts)
2. [Creating a Base Template](#creating-a-base-template)
3. [Extending Templates](#extending-templates)
4. [Using super()](#using-super)
5. [Best Practices](#best-practices)
6. [Complete Example](#complete-example)

---

## Basic Concepts

### What is Template Inheritance?

Template inheritance allows you to build a base "skeleton" template containing common elements, with **blocks** that child templates can override.

**Key Components:**
- **Base Template**: Defines the overall structure with blocks
- **Child Template**: Extends the base and overrides specific blocks
- **Blocks**: Named sections that can be overridden
- **super()**: Function to include parent block content

### Syntax

```jinja2
{# base.html #}
{% block blockname %}
  default content
{% endblock %}

{# child.html #}
{% extends "base.html" %}
{% block blockname %}
  override content
{% endblock %}
```

---

## Creating a Base Template

A base template defines the overall structure with placeholder blocks:

**`base.html`:**
```html
<!DOCTYPE html>
<html>
<head>
    <title>{% block title %}Default Title{% endblock %}</title>
    {% block extra_head %}{% endblock %}
</head>
<body>
    <header>
        {% block header %}
        <h1>Welcome</h1>
        {% endblock %}
    </header>

    <main>
        {% block content %}
        <!-- Main content goes here -->
        {% endblock %}
    </main>

    <footer>
        {% block footer %}
        <p>&copy; 2024 Company</p>
        {% endblock %}
    </footer>
</body>
</html>
```

**Block Naming Best Practices:**
- Use descriptive names: `header`, `content`, `sidebar`, `footer`
- Common blocks: `title`, `extra_head`, `extra_css`, `extra_js`
- Be consistent across your application

---

## Extending Templates

Child templates use `{% extends %}` to inherit from a base template:

**`page.html`:**
```jinja2
{% extends "base.html" %}

{% block title %}My Page Title{% endblock %}

{% block extra_head %}
<style>
    /* Page-specific styles */
    .highlight { color: blue; }
</style>
{% endblock %}

{% block content %}
<h2>Page Content</h2>
<p>This overrides the base template's content block.</p>
{% endblock %}
```

### Important Rules

1. **`{% extends %}` must be the first tag** in the child template
2. **Everything outside blocks is ignored** in child templates
3. **Only defined blocks are overridden**; undefined blocks keep base content
4. **Blocks can be nested** within other blocks

---

## Using super()

The `{{ super() }}` function includes the parent block's content:

**Example:**
```jinja2
{% extends "base.html" %}

{% block header %}
    {{ super() }}  {# Includes parent's <h1>Welcome</h1> #}
    <nav>Additional navigation</nav>
{% endblock %}
```

**Output:**
```html
<header>
    <h1>Welcome</h1>
    <nav>Additional navigation</nav>
</header>
```

### Use Cases for super()

1. **Appending to parent content:**
   ```jinja2
   {% block navigation %}
       {{ super() }}
       <a href="/new-link">New Link</a>
   {% endblock %}
   ```

2. **Prepending to parent content:**
   ```jinja2
   {% block footer %}
       <p>Additional info</p>
       {{ super() }}
   {% endblock %}
   ```

3. **Wrapping parent content:**
   ```jinja2
   {% block content %}
       <div class="wrapper">
           {{ super() }}
       </div>
   {% endblock %}
   ```

---

## Best Practices

### 1. Design Block Structure Carefully

**Good block hierarchy:**
```jinja2
{% block page %}
    {% block header %}...{% endblock %}
    {% block content %}
        {% block content_header %}...{% endblock %}
        {% block content_body %}...{% endblock %}
        {% block content_footer %}...{% endblock %}
    {% endblock %}
    {% block footer %}...{% endblock %}
{% endblock %}
```

### 2. Use Empty Blocks for Optional Content

```jinja2
{# Optional blocks that child templates can fill #}
{% block extra_css %}{% endblock %}
{% block extra_js %}{% endblock %}
{% block sidebar %}{% endblock %}
```

### 3. Provide Sensible Defaults

```jinja2
{% block title %}{{ site_name }} - Default Title{% endblock %}
{% block description %}Default site description{% endblock %}
```

### 4. Document Your Blocks

```jinja2
{# base.html #}

{# Block: page_title - Sets the browser title (required) #}
{% block page_title %}Default Title{% endblock %}

{# Block: content - Main page content (required) #}
{% block content %}{% endblock %}

{# Block: sidebar - Optional sidebar content #}
{% block sidebar %}{% endblock %}
```

### 5. Multi-Level Inheritance

You can chain inheritance multiple levels deep:

```
base.html
  └── layout.html (extends base.html)
       └── page.html (extends layout.html)
```

**Example:**
```jinja2
{# base.html - Foundation #}
{% block container %}
    {% block header %}{% endblock %}
    {% block content %}{% endblock %}
{% endblock %}

{# layout.html - Adds structure #}
{% extends "base.html" %}
{% block header %}
    <nav>{{ super() }}</nav>
{% endblock %}

{# page.html - Specific content #}
{% extends "layout.html" %}
{% block content %}
    <h1>My Page</h1>
{% endblock %}
```

---

## Complete Example

Here's a complete working example from `examples/features/inheritance/`:

### Base Template

**`base.html`:**
```html
<!DOCTYPE html>
<html>
<head>
    <title>{% block title %}Default Title{% endblock %}</title>
    {% block extra_head %}{% endblock %}
</head>
<body>
    <header>
        <h1>{% block header %}Welcome to Miya Engine{% endblock %}</h1>
        <nav>
            {% block navigation %}
            <a href="/">Home</a> | <a href="/about">About</a>
            {% endblock %}
        </nav>
    </header>

    <main>
        {% block content %}
        <p>This is the default content from the base template.</p>
        {% endblock %}
    </main>

    <aside>
        {% block sidebar %}
        <h3>Sidebar</h3>
        <p>Default sidebar content</p>
        {% endblock %}
    </aside>

    <footer>
        {% block footer %}
        <p>&copy; 2024 Miya Engine. All rights reserved.</p>
        {% endblock %}
    </footer>
</body>
</html>
```

### Child Template

**`child.html`:**
```jinja2
{% extends "base.html" %}

{% block title %}{{ page_title }} - Miya Engine{% endblock %}

{% block extra_head %}
<style>
    body { font-family: Arial, sans-serif; }
    .highlight { color: #007bff; }
</style>
{% endblock %}

{% block header %}
{{ super() }} - Extended Edition
{% endblock %}

{% block navigation %}
{{ super() }} | <a href="/features">Features</a> | <a href="/docs">Docs</a>
{% endblock %}

{% block content %}
<h2>{{ page_title }}</h2>
<p>{{ description }}</p>

<h3>Key Features Demonstrated:</h3>
<ul>
    {% for feature in features %}
    <li class="highlight">{{ feature }}</li>
    {% endfor %}
</ul>
{% endblock %}

{% block sidebar %}
<h3>Quick Links</h3>
<ul>
    {% for link in quick_links %}
    <li><a href="{{ link.url }}">{{ link.title }}</a></li>
    {% endfor %}
</ul>

<h3>Original Sidebar</h3>
{{ super() }}
{% endblock %}

{% block footer %}
<p>Generated at: {{ timestamp }}</p>
{{ super() }}
{% endblock %}
```

### Go Program

**`main.go`:**
```go
package main

import (
    "fmt"
    "log"
    "time"

    "github.com/zipreport/miya"
    "github.com/zipreport/miya/loader"
    "github.com/zipreport/miya/parser"
)

// SimpleTemplateParser implements loader.TemplateParser interface
type SimpleTemplateParser struct {
    env *miya.Environment
}

func NewSimpleTemplateParser(env *miya.Environment) *SimpleTemplateParser {
    return &SimpleTemplateParser{env: env}
}

func (stp *SimpleTemplateParser) ParseTemplate(name, content string) (*parser.TemplateNode, error) {
    template, err := stp.env.FromString(content)
    if err != nil {
        return nil, err
    }

    if templateNode, ok := template.AST().(*parser.TemplateNode); ok {
        templateNode.Name = name
        return templateNode, nil
    }

    return nil, fmt.Errorf("failed to extract template node")
}

func main() {
    // Create environment with filesystem loader
    env := miya.NewEnvironment(miya.WithAutoEscape(false))
    templateParser := NewSimpleTemplateParser(env)
    fsLoader := loader.NewFileSystemLoader([]string{"."}, templateParser)
    env.SetLoader(fsLoader)

    // Prepare context data
    ctx := miya.NewContext()
    ctx.Set("page_title", "Template Inheritance Showcase")
    ctx.Set("description", "Demonstrating template inheritance with extends, blocks, and super() calls.")
    ctx.Set("features", []string{
        "Template Extension ({% extends %})",
        "Block Definition and Override",
        "Super Calls ({{ super() }})",
        "Multi-level Inheritance",
    })
    ctx.Set("quick_links", []map[string]interface{}{
        {"title": "Documentation", "url": "/docs"},
        {"title": "Examples", "url": "/examples"},
    })
    ctx.Set("timestamp", time.Now().Format("2006-01-02 15:04:05"))

    // Render child template
    tmpl, err := env.GetTemplate("child.html")
    if err != nil {
        log.Fatal(err)
    }

    output, err := tmpl.Render(ctx)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(output)
}
```

### Running the Example

```bash
cd examples/features/inheritance
go run main.go
```

---

## Common Patterns

### Pattern 1: Multi-Page Website

```
templates/
  ├── base.html          # Main layout
  ├── pages/
  │   ├── home.html      # Extends base
  │   ├── about.html     # Extends base
  │   └── contact.html   # Extends base
  └── partials/
      ├── header.html
      └── footer.html
```

### Pattern 2: Dashboard Layout

```jinja2
{# dashboard_base.html #}
{% extends "base.html" %}

{% block content %}
<div class="dashboard">
    <aside class="sidebar">
        {% block dashboard_sidebar %}{% endblock %}
    </aside>
    <main class="dashboard-content">
        {% block dashboard_content %}{% endblock %}
    </main>
</div>
{% endblock %}

{# specific_dashboard.html #}
{% extends "dashboard_base.html" %}
{% block dashboard_content %}
    <!-- Specific dashboard content -->
{% endblock %}
```

### Pattern 3: Email Templates

```jinja2
{# email_base.html #}
<!DOCTYPE html>
<html>
<head>
    <style>
        {% block email_styles %}
        /* Base email styles */
        {% endblock %}
    </style>
</head>
<body>
    <div class="email-header">
        {% block email_header %}Logo{% endblock %}
    </div>
    <div class="email-body">
        {% block email_body %}{% endblock %}
    </div>
    <div class="email-footer">
        {% block email_footer %}Unsubscribe{% endblock %}
    </div>
</body>
</html>
```

---

## Troubleshooting

### Error: "Template Not Found"

**Cause:** The base template path is incorrect.

**Solution:**
```go
// Make sure the loader can find the base template
fsLoader := loader.NewFileSystemLoader([]string{".", "templates"}, templateParser)
```

### Error: "Block Undefined"

**Cause:** Trying to override a block that doesn't exist in the parent.

**Solution:** Check the parent template for the correct block name.

### Empty Output

**Cause:** Content outside blocks in child templates is ignored.

**Solution:**
```jinja2
{#  This is ignored #}
<p>This content is outside any block</p>

{% block content %}
{#  This renders #}
<p>This content is inside a block</p>
{% endblock %}
```

---

## See Also

- [Working Example](https://github.com/zipreport/miya/tree/master/examples/features/inheritance/) - Complete runnable example
- [MIYA_LIMITATIONS.md](MIYA_LIMITATIONS.md) - Known limitations
- [Macros and Includes](MACROS_AND_INCLUDES.md) - Related template organization features

---

## Summary

 **Template Inheritance is Fully Supported** in Miya Engine

**Key Features:**
- `{% extends %}` - Inherit from base templates
- `{% block %}` - Define overridable sections
- `{{ super() }}` - Include parent content
- Multi-level inheritance - Chain multiple templates
- Nested blocks - Blocks within blocks

**100% Jinja2 Compatible** - All inheritance features work identically to Jinja2.
