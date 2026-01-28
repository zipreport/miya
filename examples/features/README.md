# Miya Engine Feature Examples

This directory contains comprehensive, organized examples of all Miya Engine features. Each feature group has its own folder with a template file and a corresponding Go program that demonstrates the features.

##  Directory Structure

```
examples/features/
├── inheritance/          - Template inheritance (extends, blocks, super)
├── control-structures/   - Control flow (if, for, set, with)
├── filters/             - All 73+ built-in filters
├── macros-includes/     - Macros, includes, and imports
├── comprehensions/      - List and dictionary comprehensions
├── advanced/            - Filter blocks, do statements, whitespace control
├── tests-operators/     - Test expressions and operators
└── global-functions/    - Global functions (range, dict, cycler, etc.)
```

##  Quick Start

Each example folder contains:
- **template.html** or **child.html** - Template demonstrating the features
- **main.go** - Go program to run the example
- Additional support files (base templates, includes, etc.)

### Running an Example

```bash
# Navigate to any feature folder
cd examples/features/inheritance

# Run the Go program
go run main.go
```

##  Feature Groups

### 1. Template Inheritance (`inheritance/`)

**Features Demonstrated:**
- Template extension with `{% extends %}`
- Block definition and override
- Super calls with `{{ super() }}`
- Multi-level inheritance
- Block name resolution
- Nested block content

**Files:**
- `base.html` - Base template with blocks
- `child.html` - Child template extending base
- `main.go` - Demonstration program

**Run:**
```bash
cd inheritance && go run main.go
```

---

### 2. Control Structures (`control-structures/`)

**Features Demonstrated:**
- **Conditionals:** if/elif/else, inline ternary
- **Loops:** for loops with all loop variables (index, first, last, etc.)
- **Conditional iteration:** `{% for item in items if condition %}`
- **Variable assignment:** `{% set %}`
- **Scoping:** `{% with %}`
- **Dictionary iteration:** `{% for key, value in dict %}`
- **Nested loops and conditions**

**Files:**
- `template.html` - Control structures showcase
- `main.go` - Demonstration program

**Run:**
```bash
cd control-structures && go run main.go
```

---

### 3. Filters (`filters/`)

**Features Demonstrated:**
All 73+ built-in filters organized by category:

- **String (16+):** upper, lower, capitalize, title, trim, replace, truncate, center, wordcount, split, startswith, endswith, contains, slugify, indent, wordwrap
- **Collection (15+):** first, last, length, join, sort, reverse, unique, slice, batch, map, select, reject, selectattr, rejectattr, groupby
- **Numeric (10):** abs, round, int, float, sum, min, max, pow, ceil, floor
- **HTML/Security (7+):** escape, safe, striptags, urlencode, urlize, forceescape
- **Utility (8+):** default, format, tojson, filesizeformat, dictsort, attr, pprint
- **Filter chaining:** `{{ value|filter1|filter2|filter3 }}`
- **Filters with arguments**
- **Filters in conditionals and loops**

**Files:**
- `template.html` - Complete filter showcase
- `main.go` - Demonstration program

**Run:**
```bash
cd filters && go run main.go
```

---

### 4. Macros & Includes (`macros-includes/`)

**Features Demonstrated:**
- **Macros:**
  - Basic macro definition
  - Default parameters
  - Macros with logic
  - Call blocks with `{% call %}`
  - Caller function `{{ caller() }}`
  - Nested macro calls

- **Imports:**
  - Namespace import: `{% import "file.html" as name %}`
  - Selective import: `{% from "file.html" import macro1, macro2 %}`

- **Includes:**
  - Template inclusion: `{% include "template.html" %}`
  - Automatic context passing

**Files:**
- `macros.html` - Macro library
- `includes/header.html` - Header include
- `includes/footer.html` - Footer include
- `template.html` - Main template using macros and includes
- `main.go` - Demonstration program

**Run:**
```bash
cd macros-includes && go run main.go
```

---

### 5. Comprehensions (`comprehensions/`)

**Features Demonstrated:**
- **List comprehensions:**
  - Basic: `[x * 2 for x in numbers]`
  - With conditions: `[x for x in items if x.active]`
  - With filters: `[name|upper for name in names]`
  - Nested comprehensions

- **Dictionary comprehensions:**
  - Basic: `{user.id: user.name for user in users}`
  - With conditions: `{k: v for k, v in dict if condition}`
  - Transform keys/values

- **Integration:**
  - In conditionals
  - In loops
  - With additional filters
  - Complex expressions

**Files:**
- `template.html` - Comprehensions showcase
- `main.go` - Demonstration program

**Run:**
```bash
cd comprehensions && go run main.go
```

---

### 6. Advanced Features (`advanced/`)

**Features Demonstrated:**
- **Filter Blocks:**
  - Single filter: `{% filter upper %}...{% endfilter %}`
  - Chained filters: `{% filter trim|upper|replace(...) %}`
  - Nested filter blocks

- **Do Statements:**
  - Execute without output: `{% do expression %}`
  - With filters: `{% do value|filter %}`

- **Whitespace Control:**
  - Left strip: `{%- statement %}`
  - Right strip: `{% statement -%}`
  - Both sides: `{%- statement -%}`

- **Raw Blocks:**
  - Escape template syntax: `{% raw %}...{% endraw %}`

- **Autoescape:**
  - Control HTML escaping
  - Safe and escape filters

- **Environment Configuration:**
  - AutoEscape, StrictUndefined, TrimBlocks, LstripBlocks

**Files:**
- `template.html` - Advanced features showcase
- `main.go` - Demonstration program

**Run:**
```bash
cd advanced && go run main.go
```

---

### 7. Tests & Operators (`tests-operators/`)

**Features Demonstrated:**
- **Operators (19):**
  - Arithmetic: +, -, *, /, //, %, **, ~
  - Comparison: ==, !=, <, <=, >, >=
  - Logical: and, or, not
  - Membership: in, not in

- **Tests (26+):**
  - Type tests: defined, undefined, none, boolean, string, number, integer, float
  - Container tests: sequence, mapping, iterable, callable
  - Numeric tests: even, odd, divisibleby
  - String tests: lower, upper, startswith, endswith, match, alpha, alnum
  - Comparison tests: equalto, sameas, in, contains
  - Negated tests: All tests support `is not`

**Files:**
- `template.html` - Tests and operators showcase
- `main.go` - Demonstration program

**Run:**
```bash
cd tests-operators && go run main.go
```

---

### 8. Global Functions (`global-functions/`)

**Features Demonstrated:**
- **range()** - Number sequences
- **dict()** - Dictionary constructor
- **cycler()** - Cycle through values
- **joiner()** - Smart joining
- **namespace()** - Mutable container for loops
- **lipsum()** - Lorem ipsum generator
- **zip()** - Combine sequences
- **enumerate()** - Iterate with index
- **url_for()** - URL generation

**Files:**
- `template.html` - Global functions showcase
- `main.go` - Demonstration program

**Run:**
```bash
cd global-functions && go run main.go
```

---

##  Feature Coverage

### Complete Feature List

| Category | Features | Count | Status |
|----------|----------|-------|--------|
| Template Inheritance | extends, blocks, super | 3 |  100% |
| Control Structures | if/elif/else, for, set, with | 4+ |  100% |
| Filters | String, Collection, Numeric, HTML, Utility | 73+ |  100% |
| Macros & Includes | macros, call, import, include | 4 |  90% |
| Comprehensions | List, Dict | 2 |  100% |
| Advanced Features | Filter blocks, do, whitespace, raw, autoescape | 5 |  100% |
| Tests | Type, Container, Numeric, String, Comparison | 26+ |  95% |
| Operators | Arithmetic, Comparison, Logical, Membership | 19 |  100% |
| Global Functions | range, dict, cycler, joiner, etc. | 9 |  100% |

**Overall Compatibility: 95.4% with Jinja2**

##  Learning Path

### For Beginners
1. Start with **control-structures** - Learn basic flow control
2. Move to **filters** - Understand data transformation
3. Try **template inheritance** - Learn template organization

### For Intermediate Users
4. Explore **macros-includes** - Reusable components
5. Study **comprehensions** - Advanced data processing
6. Review **tests-operators** - Complex conditions

### For Advanced Users
7. Master **advanced** - Filter blocks, whitespace control
8. Utilize **global-functions** - Helper functions
9. Build real applications combining all features

##  Common Patterns

### Pattern 1: Reusable Layout
```
base.html (inheritance/)
  └── page.html extends base
      └── Override blocks as needed
```

### Pattern 2: Component Library
```
macros.html (macros-includes/)
  └── Define reusable UI components
  └── Import in templates
```

### Pattern 3: Data Processing
```
Use comprehensions + filters for data transformation
{{ [user.name|upper for user in users if user.active]|join(", ") }}
```

##  Tips

1. **Start Simple**: Run each example as-is before modifying
2. **Read Comments**: Template files contain inline documentation
3. **Check Output**: Compare expected vs actual output
4. **Experiment**: Modify context data in main.go to see different results
5. **Combine Features**: Most features work together seamlessly

##  Troubleshooting

### Template Not Found
- Ensure you're in the correct directory when running
- Loader looks for templates relative to current directory

### Import Errors
- Check that `github.com/zipreport/miya` is installed
- Run `go mod tidy` if needed

### Rendering Errors
- Check the error message for line numbers
- Verify your template syntax
- Ensure all variables in context are defined

##  Additional Resources

- **Main README**: `../../README.md` - Project overview
- **Documentation**: `../../docs/` - Detailed feature documentation
- **Other Examples**: `../../examples/go/` - More complex examples

##  Contributing

Want to add more examples? Feel free to:
1. Create a new feature folder
2. Add template and Go program
3. Update this README
4. Submit a pull request

##  License

Copyright 2024 Joao Pinheiro

MIT License - see the LICENSE file in the project root for details.
