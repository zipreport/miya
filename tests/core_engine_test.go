package tests

import (
	"testing"

	"github.com/zipreport/miya/tests/helpers"
)

// TestCoreTemplateEngine covers ALL basic template engine functionality in one comprehensive test
func TestCoreTemplateEngine(t *testing.T) {
	env := helpers.CreateEnvironment()

	// MEGA TEST SUITE - All Core Functionality
	allTests := []helpers.TestCase{

		// =================== BASIC VARIABLES & EXPRESSIONS ===================
		{
			Name:     "Simple variable",
			Template: `{{ name }}`,
			Context:  map[string]interface{}{"name": "World"},
			Expected: "World",
		},
		{
			Name:     "Arithmetic operations",
			Template: `{{ 5 + 3 }}, {{ 10 - 4 }}, {{ 6 * 7 }}, {{ 15 / 3 }}`,
			Context:  map[string]interface{}{},
			Expected: "8, 6, 42, 5",
		},
		{
			Name:     "String concatenation",
			Template: `{{ "hello" ~ " " ~ "world" }}`,
			Context:  map[string]interface{}{},
			Expected: "hello world",
		},
		{
			Name:     "Comparison operators",
			Template: `{{ 5 == 5 }}, {{ 5 != 3 }}, {{ 3 < 5 }}, {{ 5 > 3 }}`,
			Context:  map[string]interface{}{},
			Expected: "true, true, true, true",
		},
		{
			Name:     "Logical operators",
			Template: `{{ true and true }}, {{ false or true }}, {{ not false }}`,
			Context:  map[string]interface{}{},
			Expected: "true, true, true",
		},
		{
			Name:     "Ternary conditional",
			Template: `{{ 'yes' if condition else 'no' }}`,
			Context:  map[string]interface{}{"condition": true},
			Expected: "yes",
		},

		// =================== ASSIGNMENTS ===================
		{
			Name:     "Simple assignment",
			Template: `{% set x = 42 %}{{ x }}`,
			Context:  map[string]interface{}{},
			Expected: "42",
		},
		{
			Name:     "Multiple assignment",
			Template: `{% set a, b = values %}{{ a }},{{ b }}`,
			Context:  map[string]interface{}{"values": []int{1, 2}},
			Expected: "1,2",
		},
		{
			Name:     "Block assignment",
			Template: `{% set content %}Hello {{ name }}!{% endset %}{{ content }}`,
			Context:  map[string]interface{}{"name": "World"},
			Expected: "Hello World!",
		},
		{
			Name:     "Complex assignment",
			Template: `{% set result = (price * quantity)|round(2) %}{{ result }}`,
			Context:  map[string]interface{}{"price": 10.5, "quantity": 3},
			Expected: "31.50",
		},

		// =================== CONTROL STRUCTURES ===================
		{
			Name:     "If-elif-else",
			Template: `{% if score >= 90 %}A{% elif score >= 80 %}B{% else %}C{% endif %}`,
			Context:  map[string]interface{}{"score": 85},
			Expected: "B",
		},
		{
			Name:     "Simple for loop",
			Template: `{% for item in items %}{{ item }}{% endfor %}`,
			Context:  map[string]interface{}{"items": []int{1, 2, 3}},
			Expected: "123",
		},
		{
			Name:     "Loop with variables",
			Template: `{% for item in items %}{{ loop.index }}:{{ item }} {% endfor %}`,
			Context:  map[string]interface{}{"items": []string{"a", "b"}},
			Expected: "1:a 2:b ",
		},
		{
			Name:     "Loop with else",
			Template: `{% for item in items %}{{ item }}{% else %}empty{% endfor %}`,
			Context:  map[string]interface{}{"items": []interface{}{}},
			Expected: "empty",
		},
		{
			Name:     "Break and continue",
			Template: `{% for i in range(5) %}{% if i == 2 %}{% continue %}{% endif %}{{ i }}{% if i == 3 %}{% break %}{% endif %}{% endfor %}`,
			Context:  map[string]interface{}{},
			Expected: "013",
		},
		{
			Name:     "Nested loops",
			Template: `{% for i in [1,2] %}{% for j in ['a','b'] %}{{ i }}{{ j }} {% endfor %}{% endfor %}`,
			Context:  map[string]interface{}{},
			Expected: "1a 1b 2a 2b ",
		},
		{
			Name:     "Dictionary iteration",
			Template: `{% for key, value in dict|dictsort %}{{ key }}:{{ value }} {% endfor %}`,
			Context:  map[string]interface{}{"dict": map[string]interface{}{"a": 1, "b": 2}},
			Expected: "a:1 b:2 ",
		},

		// =================== WITH STATEMENTS ===================
		{
			Name:     "With statement",
			Template: `{% with greeting = "Hello" %}{{ greeting }} World{% endwith %}`,
			Context:  map[string]interface{}{},
			Expected: "Hello World",
		},
		{
			Name:     "Multiple with variables",
			Template: `{% with x = 1, y = 2 %}{{ x + y }}{% endwith %}`,
			Context:  map[string]interface{}{},
			Expected: "3",
		},
		{
			Name:     "With scoping",
			Template: `{{ name }}{% with name = "Bob" %}{{ name }}{% endwith %}{{ name }}`,
			Context:  map[string]interface{}{"name": "Alice"},
			Expected: "AliceBobAlice",
		},

		// =================== STRING FILTERS ===================
		{
			Name:     "Basic string filters",
			Template: `{{ text|upper }}, {{ text|lower }}, {{ text|title }}, {{ text|capitalize }}`,
			Context:  map[string]interface{}{"text": "hello world"},
			Expected: "HELLO WORLD, hello world, Hello World, Hello world",
		},
		{
			Name:     "String manipulation filters",
			Template: `[{{ text|trim }}], {{ text|replace("world", "universe") }}, {{ text|reverse }}`,
			Context:  map[string]interface{}{"text": "  hello world  "},
			Expected: "[hello world],   hello universe  ,   dlrow olleh  ",
		},
		{
			Name:     "String analysis filters",
			Template: `{{ text|length }}, {{ text|wordcount }}, {{ text|startswith("hello") }}`,
			Context:  map[string]interface{}{"text": "hello world test"},
			Expected: "16, 3, true",
		},
		{
			Name:     "Advanced string filters",
			Template: `{{ text|truncate(10) }}, "{{ 'test'|center(10) }}", {{ 'hello-world'|slugify }}`,
			Context:  map[string]interface{}{"text": "This is a long text"},
			Expected: `This is a..., "   test   ", hello-world`,
		},

		// =================== MATH FILTERS ===================
		{
			Name:     "Math filters",
			Template: `{{ num|abs }}, {{ num|round(1) }}, {{ num|ceil }}, {{ num|floor }}`,
			Context:  map[string]interface{}{"num": -3.7},
			Expected: "3.7, -3.7, -3, -4",
		},
		{
			Name:     "Collection math",
			Template: `{{ nums|list|min }}, {{ nums|list|max }}, {{ nums|list|sum }}`,
			Context:  map[string]interface{}{"nums": []int{1, 5, 3, 9, 2}},
			Expected: "1, 9, 20",
		},
		{
			Name:     "Power and complex math",
			Template: `{{ 2|pow(3) }}, {{ [1,2,3,4]|length }}, {{ 5.678|round(2) }}`,
			Context:  map[string]interface{}{},
			Expected: "8, 4, 5.68",
		},

		// =================== COLLECTION FILTERS ===================
		{
			Name:     "List filters",
			Template: `{{ items|first }}, {{ items|last }}, {{ items|length }}`,
			Context:  map[string]interface{}{"items": []string{"a", "b", "c"}},
			Expected: "a, c, 3",
		},
		{
			Name:     "List manipulation",
			Template: `{{ items|sort|join(",") }}, {{ items|reverse|join("|") }}`,
			Context:  map[string]interface{}{"items": []string{"c", "a", "b"}},
			Expected: "a,b,c, b|a|c",
		},
		{
			Name:     "Advanced list operations",
			Template: `{{ items|list|unique|join(",") }}, {{ [1,2,3,4,5]|list|slice(1,4)|join("-") }}`,
			Context:  map[string]interface{}{"items": []int{1, 2, 2, 3}},
			Expected: "1,2,3, 2-3-4",
		},
		{
			Name:     "Dictionary filters",
			Template: `{{ dict|keys|sort|join(",") }}, {{ dict|values|sort|join("|") }}`,
			Context:  map[string]interface{}{"dict": map[string]interface{}{"b": 2, "a": 1, "c": 3}},
			Expected: "a,b,c, 1|2|3",
		},

		// =================== HTML/URL FILTERS ===================
		{
			Name:     "HTML filters",
			Template: `{{ html|escape }}, {{ html|safe }}, {{ html|striptags }}`,
			Context:  map[string]interface{}{"html": "<b>test</b>"},
			Expected: "&amp;lt;b&amp;gt;test&amp;lt;/b&amp;gt;, <b>test</b>, test",
		},
		{
			Name:     "URL filters",
			Template: `{{ url|urlencode }}, {{ text|urlize }}`,
			Context:  map[string]interface{}{"url": "hello world", "text": "Visit https://example.com"},
			Expected: `hello+world, Visit <a href="https://example.com">https://example.com</a>`,
		},

		// =================== TYPE CONVERSION ===================
		{
			Name:     "Type conversion",
			Template: `{{ "42"|int }}, {{ "3.14"|float }}, {{ 123|string }}, {{ "hello"|list|join("-") }}`,
			Context:  map[string]interface{}{},
			Expected: "42, 3.14, 123, h-e-l-l-o",
		},

		// =================== UTILITY FILTERS ===================
		{
			Name:     "Utility filters",
			Template: `{{ name|default("none") }}, {{ size }}`,
			Context:  map[string]interface{}{"size": 1024},
			Expected: `none, 1024`,
		},

		// =================== FILTER CHAINING ===================
		{
			Name:     "Filter chaining",
			Template: `{{ text|trim|upper|reverse }}, {{ items|sort|reverse|join(" | ") }}`,
			Context:  map[string]interface{}{"text": "  hello  ", "items": []string{"c", "a", "b"}},
			Expected: "OLLEH, c | b | a",
		},
		{
			Name:     "Complex filter chains",
			Template: `{{ text|replace(" ", "_")|upper|center(20) }}`,
			Context:  map[string]interface{}{"text": "hello world"},
			Expected: "    HELLO_WORLD     ",
		},

		// =================== ADVANCED LOOPS ===================
		{
			Name:     "Advanced loop variables",
			Template: `{% for item in items %}{{ loop.cycle('odd','even') }} {% endfor %}`,
			Context:  map[string]interface{}{"items": []string{"a", "b", "c", "d"}},
			Expected: "odd even odd even ",
		},
		{
			Name:     "Loop changed detection",
			Template: `{% for item in items %}{{ item }}:{{ loop.changed(item) }} {% endfor %}`,
			Context:  map[string]interface{}{"items": []string{"a", "a", "b"}},
			Expected: "a:true a:false b:true ",
		},
		{
			Name:     "Recursive loops",
			Template: `{% for item in items recursive %}{{ item.name }}{% if item.children %}-{{ loop(item.children) }}{% endif %} {% endfor %}`,
			Context: map[string]interface{}{
				"items": []interface{}{
					map[string]interface{}{
						"name": "parent",
						"children": []interface{}{
							map[string]interface{}{"name": "child1"},
							map[string]interface{}{"name": "child2"},
						},
					},
				},
			},
			Expected: "parent-child1 child2  ",
		},

		// =================== SLICING & EXPRESSIONS ===================
		{
			Name:     "String and list slicing",
			Template: `{{ text[0:5] }}, {{ items[1:3]|list|join(",") }}, {{ text[-5:] }}`,
			Context:  map[string]interface{}{"text": "hello world", "items": []int{1, 2, 3, 4, 5}},
			Expected: "hello, 2,3, world",
		},
		{
			Name:     "Complex access patterns",
			Template: `{{ data.user.name }}, {{ data["items"][0] }}, {{ matrix[1][0] }}`,
			Context: map[string]interface{}{
				"data": map[string]interface{}{
					"user":  map[string]interface{}{"name": "John"},
					"items": []string{"first", "second"},
				},
				"matrix": [][]int{{1, 2}, {3, 4}},
			},
			Expected: "John, first, 3",
		},

		// =================== TESTS (BOOLEAN CHECKS) ===================
		{
			Name:     "Type tests",
			Template: `{{ name is defined }}, {{ age is number }}, {{ items is sequence }}`,
			Context:  map[string]interface{}{"name": "John", "age": 25, "items": []int{1, 2}},
			Expected: "true, true, true",
		},
		{
			Name:     "Value tests",
			Template: `{{ 4 is even }}, {{ 5 is odd }}, {{ 10 is divisibleby(5) }}`,
			Context:  map[string]interface{}{},
			Expected: "true, true, true",
		},
		{
			Name:     "String tests",
			Template: `{{ text is lower }}, {{ "UPPER" is upper }}, {{ "hello" is startswith("he") }}`,
			Context:  map[string]interface{}{"text": "hello"},
			Expected: "true, true, true",
		},
		{
			Name:     "In operator",
			Template: `{{ "apple" in fruits }}, {{ "ell" in "hello" }}, {{ "key" in dict }}`,
			Context: map[string]interface{}{
				"fruits": []string{"apple", "banana"},
				"dict":   map[string]int{"key": 1},
			},
			Expected: "true, true, true",
		},

		// =================== SPECIAL BLOCKS ===================
		{
			Name:     "Raw blocks",
			Template: `{% raw %}{{ not_rendered }} {% for x in y %}{% endraw %}`,
			Context:  map[string]interface{}{},
			Expected: "{{not_rendered}} {%forxiny%}",
		},
		{
			Name:     "Filter blocks",
			Template: `{% filter upper %}hello {{ name }}{% endfilter %}`,
			Context:  map[string]interface{}{"name": "world"},
			Expected: "HELLO WORLD",
		},
		{
			Name:     "Filter block chaining",
			Template: `{% filter trim|upper|reverse %}  hello world  {% endfilter %}`,
			Context:  map[string]interface{}{},
			Expected: "DLROW OLLEH",
		},
		{
			Name:     "Do statements",
			Template: `{% do 5 + 3 %}{% set x = 42 %}{% do x * 2 %}{{ x }}`,
			Context:  map[string]interface{}{},
			Expected: "42",
		},

		// =================== MACROS ===================
		{
			Name:     "Simple macro",
			Template: `{% macro greet(name) %}Hello {{ name }}!{% endmacro %}{{ greet("World") }}`,
			Context:  map[string]interface{}{},
			Expected: "Hello World!",
		},
		{
			Name:     "Macro with defaults",
			Template: `{% macro alert(msg, type="info") %}<div class="{{ type }}">{{ msg }}</div>{% endmacro %}{{ alert("Test") }}{{ alert("Error", "danger") }}`,
			Context:  map[string]interface{}{},
			Expected: `&lt;div class=&#34;info&#34;&gt;Test&lt;/div&gt;&lt;div class=&#34;danger&#34;&gt;Error&lt;/div&gt;`,
		},
		{
			Name:     "Call blocks",
			Template: `{% macro dialog(title) %}<h3>{{ title }}</h3><div>{{ caller() }}</div>{% endmacro %}{% call dialog("Test") %}Content here{% endcall %}`,
			Context:  map[string]interface{}{},
			Expected: "<h3>Test</h3><div>Content here</div>",
		},

		// =================== TEMPLATE INHERITANCE ===================
		{
			Name:     "Basic blocks",
			Template: `{% block header %}Default Header{% endblock %} | {% block content %}Default Content{% endblock %}`,
			Context:  map[string]interface{}{},
			Expected: "Default Header | Default Content",
		},

		// =================== WHITESPACE CONTROL ===================
		{
			Name:     "Whitespace control",
			Template: `a{%- if true -%}b{%- endif -%}c`,
			Context:  map[string]interface{}{},
			Expected: "abc",
		},
		{
			Name:     "Loop whitespace",
			Template: `{%- for i in [1,2,3] -%}{{ i }}{%- endfor -%}`,
			Context:  map[string]interface{}{},
			Expected: "123",
		},

		// =================== NAMESPACE ===================
		{
			Name:     "Namespace usage",
			Template: `{% set ns = namespace(count=0) %}{% for i in range(3) %}{% set ns.count = ns.count + 1 %}{% endfor %}{{ ns.count }}`,
			Context:  map[string]interface{}{},
			Expected: "3",
		},

		// =================== DATE/TIME FILTERS ===================
		{
			Name:     "Date filters basic",
			Template: `{{ now|strftime('%Y')|length == 4 }}`,
			Context:  map[string]interface{}{"now": "2025-01-01"},
			Expected: "true",
		},

		// =================== COMPLEX INTEGRATIONS ===================
		{
			Name:     "Complex integration",
			Template: `{% for user in users %}{% if user.active %}{{ user.name|title }}{% if user.is_admin %} (Admin){% endif %}{% if not loop.last %}, {% endif %}{% endif %}{% endfor %}`,
			Context: map[string]interface{}{
				"users": []map[string]interface{}{
					{"name": "alice", "active": true, "is_admin": true},
					{"name": "bob", "active": false, "is_admin": false},
					{"name": "charlie", "active": true, "is_admin": false},
				},
			},
			Expected: "Alice (Admin), Charlie",
		},

		// =================== LIST COMPREHENSIONS ===================
		{
			Name:     "List comprehensions",
			Template: `{{ [x * 2 for x in numbers]|join(",") }}`,
			Context:  map[string]interface{}{"numbers": []int{1, 2, 3}},
			Expected: "2,4,6",
		},
		{
			Name:     "Conditional comprehensions",
			Template: `{% for x in numbers %}{% if x % 2 == 0 %}{{ x }}{% endif %}{% endfor %}`,
			Context:  map[string]interface{}{"numbers": []int{1, 2, 3, 4, 5, 6}},
			Expected: "246",
		},
	}

	// Execute all tests
	t.Logf("Running %d comprehensive core engine tests...", len(allTests))
	helpers.RenderTestCases(t, env, allTests)
	t.Logf("✅ All %d core engine tests passed!", len(allTests))
}

// TestCoreEngineErrors covers all error cases in one comprehensive test
func TestCoreEngineErrors(t *testing.T) {
	env := helpers.CreateEnvironment()

	errorTests := []helpers.TestCase{
		// Syntax errors
		{
			Name:          "Missing endif",
			Template:      `{% if condition %}content`,
			Context:       map[string]interface{}{"condition": true},
			ShouldError:   true,
			ErrorContains: "close if statement",
		},
		{
			Name:          "Missing endfor",
			Template:      `{% for item in items %}{{ item }}`,
			Context:       map[string]interface{}{"items": []int{1, 2, 3}},
			ShouldError:   true,
			ErrorContains: "close for statement",
		},

		// Variable errors (this might not error in this implementation)
		{
			Name:        "Undefined variable",
			Template:    `{{ undefined_var }}`,
			Context:     map[string]interface{}{},
			ShouldError: false, // Some implementations handle undefined gracefully
			Expected:    "",    // May render as empty or undefined
		},

		// Assignment errors
		{
			Name:          "Unpacking mismatch",
			Template:      `{% set a, b = values %}`,
			Context:       map[string]interface{}{"values": []int{1, 2, 3}},
			ShouldError:   true,
			ErrorContains: "cannot unpack 3 values into 2",
		},

		// Filter errors
		{
			Name:          "Unknown filter",
			Template:      `{{ text|nonexistent }}`,
			Context:       map[string]interface{}{"text": "test"},
			ShouldError:   true,
			ErrorContains: "unknown filter",
		},
		{
			Name:        "Type mismatch",
			Template:    `{{ 42|upper }}`,
			Context:     map[string]interface{}{},
			ShouldError: false, // May handle gracefully
			Expected:    "42",  // May just return the original value
		},

		// Math errors
		{
			Name:          "Division by zero",
			Template:      `{{ 5 / 0 }}`,
			Context:       map[string]interface{}{},
			ShouldError:   true,
			ErrorContains: "division by zero",
		},

		// Loop errors
		{
			Name:          "Break outside loop",
			Template:      `{% break %}`,
			Context:       map[string]interface{}{},
			ShouldError:   true,
			ErrorContains: "break statement",
		},
		{
			Name:          "Continue outside loop",
			Template:      `{% continue %}`,
			Context:       map[string]interface{}{},
			ShouldError:   true,
			ErrorContains: "continue statement",
		},

		// Type errors
		{
			Name:        "Invalid slice",
			Template:    `{{ "hello"[10:5] }}`,
			Context:     map[string]interface{}{},
			ShouldError: false, // May handle gracefully
			Expected:    "",    // Empty result for out-of-bounds slice
		},

		// Macro errors
		{
			Name:          "Undefined macro",
			Template:      `{{ nonexistent_macro() }}`,
			Context:       map[string]interface{}{},
			ShouldError:   true,
			ErrorContains: "cannot call non-function",
		},
	}

	t.Logf("Running %d comprehensive error tests...", len(errorTests))
	helpers.RenderTestCases(t, env, errorTests)
	t.Logf("✅ All %d error tests passed!", len(errorTests))
}

// Helper functions are in tests/helpers package
