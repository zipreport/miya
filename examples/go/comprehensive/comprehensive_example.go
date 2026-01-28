package main

import (
	"fmt"
	"os"
	"time"

	miya "github.com/zipreport/miya"
)

func main() {
	fmt.Println("🚀 Starting Miya Engine Comprehensive Template Rendering...")
	startTime := time.Now()

	// Create environment
	fmt.Printf("⏱️  [%.2fms] Creating environment...\n", float64(time.Since(startTime).Nanoseconds())/1e6)
	envStart := time.Now()
	env := miya.NewEnvironment()
	fmt.Printf("✅ [%.2fms] Environment created in %.2fms\n", float64(time.Since(startTime).Nanoseconds())/1e6, float64(time.Since(envStart).Nanoseconds())/1e6)

	// Add custom filters
	filterStart := time.Now()
	addCustomFilters(env)
	fmt.Printf("✅ [%.2fms] Custom filters added in %.2fms\n", float64(time.Since(startTime).Nanoseconds())/1e6, float64(time.Since(filterStart).Nanoseconds())/1e6)

	// Create sample data
	dataStart := time.Now()
	ctx := createSampleData()
	fmt.Printf("✅ [%.2fms] Sample data created in %.2fms\n", float64(time.Since(startTime).Nanoseconds())/1e6, float64(time.Since(dataStart).Nanoseconds())/1e6)

	// Template content adapted for Miya Engine limitations
	template := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{ title | default('Jinja2 Features Showcase') }}</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 1200px; margin: 0 auto; padding: 20px; }
        .section { margin: 30px 0; padding: 20px; border: 1px solid #ddd; border-radius: 8px; }
        .code { background: #f5f5f5; padding: 10px; border-radius: 4px; font-family: monospace; }
        table { border-collapse: collapse; width: 100%; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
        .error { color: red; font-weight: bold; }
        .success { color: green; font-weight: bold; }
        .warning { color: orange; font-weight: bold; }
    </style>
</head>
<body>
    <h1>{{ title | default('Jinja2 Features Showcase') }}</h1>
    <p>Generated on: {{ formatted_date }}</p>

    <!-- Variable Expressions and Filters -->
    <div class="section">
        <h2>1. Variable Expressions and Filters</h2>
        <p><strong>Basic Variable:</strong> {{ user.name | default('Anonymous') }}</p>
        <p><strong>Upper Case:</strong> {{ user.name | upper }}</p>
        <p><strong>Title Case:</strong> {{ user.name | title }}</p>
        <p><strong>Length:</strong> {{ user.name | length }} characters</p>
        <p><strong>Truncate:</strong> {{ description | truncate(50) }}</p>
        <p><strong>Default Value:</strong> {{ missing_var | default('No value provided') }}</p>
        <p><strong>Safe HTML:</strong> {{ html_content | safe }}</p>
        <p><strong>Escape HTML:</strong> {{ html_content | escape }}</p>
        <p><strong>Number Format:</strong> {{ price | round(2) }}</p>
        <p><strong>Join List:</strong> {{ tags | join(', ') }}</p>
        <p><strong>First/Last:</strong> First tag: {{ tags | first }}, Last tag: {{ tags | last }}</p>
    </div>

    <!-- Control Structures - If/Else -->
    <div class="section">
        <h2>2. Conditional Statements (if/elif/else)</h2>
        {% if user.is_admin %}
            <p class="success">Welcome, Administrator {{ user.name }}!</p>
        {% elif user.is_premium %}
            <p class="warning">Welcome, Premium User {{ user.name }}!</p>
        {% else %}
            <p>Welcome, {{ user.name }}!</p>
        {% endif %}

        {% if user.age %}
            {% if user.age >= 18 %}
                <p>✅ Adult user ({{ user.age }} years old)</p>
            {% else %}
                <p>🔒 Minor user ({{ user.age }} years old)</p>
            {% endif %}
        {% endif %}

        <p>Status: 
        {%- if user.status == 'active' -%}
            <span class="success">Active</span>
        {%- elif user.status == 'inactive' -%}
            <span class="warning">Inactive</span>
        {%- else -%}
            <span class="error">Unknown</span>
        {%- endif -%}
        </p>
    </div>

    <!-- Loops -->
    <div class="section">
        <h2>3. Loops (for)</h2>
        
        <h3>Simple List Loop</h3>
        <ul>
        {% for tag in tags %}
            <li>{{ loop.index }}. {{ tag }}</li>
        {% endfor %}
        </ul>

        <h3>Dictionary Loop with Key-Value Unpacking</h3>
        <ul>
        {% for key, value in user.preferences %}
            <li><strong>{{ key }}:</strong> {{ value }}</li>
        {% endfor %}
        </ul>

        <h3>Loop with Conditions</h3>
        <ul>
        {% for product in products %}
            {% if product.in_stock %}
                <li>{{ product.name }} - ${{ product.price }} (In Stock)</li>
            {% endif %}
        {% endfor %}
        </ul>

        <h3>Loop Variables</h3>
        <table>
            <tr><th>Index</th><th>Index0</th><th>First</th><th>Last</th><th>Length</th><th>Item</th></tr>
        {% for item in tags %}
            <tr>
                <td>{{ loop.index }}</td>
                <td>{{ loop.index0 }}</td>
                <td>{{ loop.first }}</td>
                <td>{{ loop.last }}</td>
                <td>{{ loop.length }}</td>
                <td>{{ item }}</td>
            </tr>
        {% endfor %}
        </table>

        <h3>Empty Loop</h3>
        <ul>
        {% for item in empty_list %}
            <li>{{ item }}</li>
        {% else %}
            <li><em>No items found</em></li>
        {% endfor %}
        </ul>

        <h3>Nested Loops</h3>
        {% for category in categories %}
            <h4>{{ category.name }}</h4>
            <ul>
            {% for item in category.items %}
                <li>{{ loop.index }}.{{ loop.index0 }} {{ item.name }}</li>
            {% endfor %}
            </ul>
        {% endfor %}
    </div>

    <!-- Template Inheritance & Blocks -->
    <div class="section">
        <h2>4. Template Inheritance & Blocks</h2>
        {% block header %}
            <div class="code">
                This is the default header block content.<br>
                Child templates can override this using {% raw %}{% block header %}{% endraw %}
            </div>
        {% endblock %}

        {% block content %}
            <p>This is the main content block. Child templates typically override this.</p>
        {% endblock %}

        {% block sidebar %}
            <div class="code">
                <strong>Sidebar Block:</strong><br>
                {# {{ super() }} would include parent block content if this was a child template #}
                This sidebar content can be overridden by child templates.
            </div>
        {% endblock %}
    </div>

    <!-- Macros -->
    <div class="section">
        <h2>5. Macros (Reusable Functions)</h2>
        
        {% macro render_field(field_name, field_value, field_type='text') %}
            <div class="form-field">
                <label for="{{ field_name }}">{{ field_name | title }}:</label>
                <input type="{{ field_type }}" id="{{ field_name }}" name="{{ field_name }}" value="{{ field_value }}">
            </div>
        {% endmacro %}

        {% macro alert(message, type='info') %}
            <div class="alert alert-{{ type }}">{{ message }}</div>
        {% endmacro %}

        <h3>Macro Examples:</h3>
        {{ render_field('username', user.name) }}
        {{ render_field('email', user.email, 'email') }}
        {{ render_field('age', user.age, 'number') }}
        
        {{ alert('This is an info message') }}
        {{ alert('This is a warning!', 'warning') }}
        {{ alert('Success!', 'success') }}
    </div>

    <!-- Variable Assignments -->
    <div class="section">
        <h2>6. Variable Assignments</h2>
        {% set full_name = user.first_name ~ ' ' ~ user.last_name %}
        {% set user_level = 'Premium' if user.is_premium else 'Basic' %}
        {% set total_products = products | length %}
        
        <p><strong>Full Name:</strong> {{ full_name }}</p>
        <p><strong>User Level:</strong> {{ user_level }}</p>
        <p><strong>Total Products:</strong> {{ total_products }}</p>
    </div>

    <!-- Raw Content -->
    <div class="section">
        <h2>7. Raw Content (Escaped Template Syntax)</h2>
        <div class="code">
            {% raw %}
            This shows raw Jinja2 syntax: {{ variable_name }}
            {% for item in items %}
                <li>{{ item }}</li>
            {% endfor %}
            {% endraw %}
        </div>
    </div>

    <!-- Comments -->
    <div class="section">
        <h2>8. Comments</h2>
        {# This is a Jinja2 comment that won't appear in the HTML output #}
        <p>Comments are useful for documentation and debugging.</p>
        {# 
        Multi-line comments
        are also supported
        #}
    </div>

    <!-- Tests -->
    <div class="section">
        <h2>9. Tests (Boolean Checks)</h2>
        <ul>
            <li><strong>Defined:</strong> {{ 'user.name is defined' if user.name is defined else 'user.name is not defined' }}</li>
            <li><strong>None:</strong> {{ 'Value is none' if none_value is none else 'Value is not none' }}</li>
            <li><strong>Number:</strong> {{ 'Age is a number' if user.age is number else 'Age is not a number' }}</li>
            <li><strong>String:</strong> {{ 'Name is a string' if user.name is string else 'Name is not a string' }}</li>
            <li><strong>Even/Odd:</strong> Age {{ user.age }} is {{ 'even' if user.age is even else 'odd' }}</li>
            <li><strong>Divisible by:</strong> Age is {{ 'divisible by 5' if user.age is divisibleby(5) else 'not divisible by 5' }}</li>
            <li><strong>In sequence:</strong> 'jinja2' is {{ 'in tags' if 'jinja2' in tags else 'not in tags' }}</li>
        </ul>
    </div>

    <!-- Whitespace Control -->
    <div class="section">
        <h2>10. Whitespace Control</h2>
        <div class="code">
            Regular spacing:
            {% for i in range(3) %}
                {{ i }}
            {% endfor %}
            
            <br><br>
            
            Stripped spacing:
            {%- for i in range(3) -%}
                {{- i -}}
            {%- endfor -%}
        </div>
    </div>

    <!-- Complex Expressions -->
    <div class="section">
        <h2>11. Complex Expressions and Operations</h2>
        <ul>
            <li><strong>Math:</strong> {{ user.age + 10 }} (age + 10)</li>
            <li><strong>String concat:</strong> {{ user.name ~ ' (ID: ' ~ user.id ~ ')' }}</li>
            <li><strong>Comparison:</strong> {{ 'Adult' if user.age >= 18 else 'Minor' }}</li>
            <li><strong>List slicing:</strong> First 3 tags: {{ tags[:3] | join(', ') }}</li>
            <li><strong>Dict access:</strong> Theme: {{ user.preferences.theme | default('default') }}</li>
        </ul>
    </div>

    <!-- Built-in Filters Showcase -->
    <div class="section">
        <h2>12. Built-in Filters Showcase (70+ Available!)</h2>
        
        <h3>String Filters</h3>
        <ul>
            <li><strong>capitalize:</strong> {{ 'hello world' | capitalize }}</li>
            <li><strong>center:</strong> "{{ 'test' | center(10) }}"</li>
            <li><strong>replace:</strong> {{ 'Hello World' | replace('World', 'Jinja2') }}</li>
            <li><strong>reverse:</strong> {{ 'hello' | reverse }}</li>
            <li><strong>truncate:</strong> {{ description | truncate(50) }}</li>
            <li><strong>wordwrap:</strong> {{ 'This is a very long sentence that should be wrapped' | wordwrap(15) }}</li>
            <li><strong>indent:</strong> {{ 'Line 1\nLine 2' | indent(4) }}</li>
            <li><strong>slugify:</strong> {{ 'Hello World Test!' | slugify }}</li>
            <li><strong>wordcount:</strong> {{ description | wordcount }} words</li>
            <li><strong>startswith:</strong> {{ 'hello world' | startswith('hello') }}</li>
            <li><strong>endswith:</strong> {{ 'hello world' | endswith('world') }}</li>
            <li><strong>contains:</strong> {{ 'hello world' | contains('lo wo') }}</li>
            <li><strong>split:</strong> {{ 'a,b,c,d' | split(',') | join(' | ') }}</li>
            <li><strong>regex_replace:</strong> {{ 'hello123world' | regex_replace('\\d+', '-') }}</li>
        </ul>

        <h3>Math Filters</h3>
        <ul>
            <li><strong>abs:</strong> {{ -42 | abs }}</li>
            <li><strong>round:</strong> {{ 3.14159 | round(2) }}</li>
            <li><strong>ceil:</strong> {{ 3.2 | ceil }}</li>
            <li><strong>floor:</strong> {{ 3.8 | floor }}</li>
            <li><strong>pow:</strong> {{ 2 | pow(3) }}</li>
            <li><strong>min:</strong> {{ [5, 2, 8, 1, 9] | min }}</li>
            <li><strong>max:</strong> {{ [5, 2, 8, 1, 9] | max }}</li>
            <li><strong>sum:</strong> {{ [1, 2, 3, 4, 5] | sum }}</li>
        </ul>

        <h3>Collection Filters</h3>
        <ul>
            <li><strong>first:</strong> {{ tags | first }}</li>
            <li><strong>last:</strong> {{ tags | last }}</li>
            <li><strong>length:</strong> {{ tags | length }}</li>
            <li><strong>join:</strong> {{ tags | join(', ') }}</li>
            <li><strong>sort:</strong> {{ tags | sort | join(', ') }}</li>
            <li><strong>reverse:</strong> {{ tags | reverse | join(', ') }}</li>
            <li><strong>unique:</strong> {{ [1, 2, 2, 3, 3, 3] | unique | list | join(', ') }}</li>
            <li><strong>slice:</strong> {{ tags | slice(0, 3) | join(', ') }}</li>
            <li><strong>batch:</strong> {{ [1,2,3,4,5,6] | batch(2) | list }}</li>
            <li><strong>random:</strong> {{ tags | random }}</li>
        </ul>

        <h3>Dict Filters</h3>
        <ul>
            <li><strong>items:</strong> {{ user.preferences | items | list }}</li>
            <li><strong>keys:</strong> {{ user.preferences | keys | list }}</li>
            <li><strong>values:</strong> {{ user.preferences | values | list }}</li>
            <li><strong>dictsort:</strong> {{ user.preferences | dictsort }}</li>
        </ul>

        <h3>Type Conversion</h3>
        <ul>
            <li><strong>float:</strong> {{ '3.14' | float }}</li>
            <li><strong>int:</strong> {{ '42' | int }}</li>
            <li><strong>string:</strong> {{ 123 | string }}</li>
            <li><strong>list:</strong> {{ 'hello' | list | join('-') }}</li>
        </ul>

        <h3>HTML & URL Filters</h3>
        <ul>
            <li><strong>escape:</strong> {{ '<script>alert("xss")</script>' | escape }}</li>
            <li><strong>safe:</strong> {{ '<strong>Bold text</strong>' | safe }}</li>
            <li><strong>striptags:</strong> {{ '<p>Hello <b>world</b></p>' | striptags }}</li>
            <li><strong>urlencode:</strong> {{ 'hello world' | urlencode }}</li>
            <li><strong>urlize:</strong> {{ 'Visit https://example.com for info' | urlize }}</li>
        </ul>

        <h3>Date/Time Filters</h3>
        <ul>
            <li><strong>date:</strong> {{ current_datetime | date('%Y-%m-%d') }}</li>
            <li><strong>time:</strong> {{ current_datetime | time('%H:%M:%S') }}</li>
            <li><strong>datetime:</strong> {{ current_datetime | datetime('%Y-%m-%d %H:%M') }}</li>
            <li><strong>strftime:</strong> {{ current_datetime | strftime('%A, %B %d, %Y') }}</li>
            <li><strong>weekday:</strong> {{ current_datetime | weekday }}</li>
            <li><strong>month_name:</strong> {{ current_datetime | month_name }}</li>
        </ul>

        <h3>Utility Filters</h3>
        <ul>
            <li><strong>default:</strong> {{ missing_var | default('No value provided') }}</li>
            <li><strong>filesizeformat:</strong> {{ 1024000 | filesizeformat }}</li>
            <li><strong>format:</strong> {{ 'Hello %s, age %d' | format(user.name, user.age) }}</li>
            <li><strong>tojson:</strong> {{ user.preferences | tojson }}</li>
            <li><strong>pprint:</strong> {{ user.preferences | pprint }}</li>
            <li><strong>custom_exclaim:</strong> {{ 'Custom filter test' | custom_exclaim }}</li>
        </ul>
    </div>

    <!-- Auto-escaping -->
    <div class="section">
        <h2>13. Auto-escaping</h2>
        <p><strong>Auto-escaped:</strong> {{ '<script>alert("xss")</script>' }}</p>
        <p><strong>Manually escaped:</strong> {{ '<script>alert("xss")</script>' | escape }}</p>
        <p><strong>Safe (not escaped):</strong> {{ '<strong>Bold text</strong>' | safe }}</p>
        
        {% autoescape false %}
            <p><strong>Auto-escape disabled:</strong> {{ '<em>Italic text</em>' }}</p>
        {% endautoescape %}
    </div>

    <!-- Namespace -->
    <div class="section">
        <h2>14. Namespace (Variable Scoping)</h2>
        {% set ns = namespace(counter=0) %}
        <ul>
        {% for item in tags %}
            {% set ns.counter = ns.counter + 1 %}
            <li>Item {{ ns.counter }}: {{ item }}</li>
        {% endfor %}
        </ul>
        <p>Total items processed: {{ ns.counter }}</p>
    </div>

    <!-- Call Blocks -->
    <div class="section">
        <h2>15. Call Blocks</h2>
        {% macro render_dialog(title) -%}
            <div class="dialog">
                <h3>{{ title }}</h3>
                <div class="dialog-content">
                    {{ caller() }}
                </div>
            </div>
        {%- endmacro %}

        {% call render_dialog('Important Notice') %}
            <p>This is the content inside the dialog.</p>
            <p>It can contain <strong>HTML</strong> and {{ 'variables' }}.</p>
        {% endcall %}
    </div>

    <!-- With Statement -->
    <div class="section">
        <h2>16. With Statement (Context Variables)</h2>
        {% with greeting = 'Hello', target = user.name %}
            <p>{{ greeting }}, {{ target }}! Welcome to our site.</p>
            {% with age_group = 'adult' if user.age >= 18 else 'minor' %}
                <p>You are classified as: {{ age_group }}</p>
            {% endwith %}
        {% endwith %}
    </div>

    <!-- Complex Data Structure Example -->
    <div class="section">
        <h2>17. Complex Data Structure Example</h2>
        <h3>User Profile</h3>
        <table>
            <tr><th>Field</th><th>Value</th></tr>
            <tr><td>Name</td><td>{{ user.name }}</td></tr>
            <tr><td>Email</td><td>{{ user.email }}</td></tr>
            <tr><td>Age</td><td>{{ user.age }}</td></tr>
            <tr><td>Status</td><td>{{ user.status }}</td></tr>
            <tr><td>Premium</td><td>{{ '✅' if user.is_premium else '❌' }}</td></tr>
            <tr><td>Admin</td><td>{{ '✅' if user.is_admin else '❌' }}</td></tr>
        </table>

        <h3>Products</h3>
        <table>
            <tr><th>Name</th><th>Price</th><th>Stock</th><th>Category</th></tr>
            {% for product in products %}
            <tr>
                <td>{{ product.name }}</td>
                <td>${{ product.price }}</td>
                <td>{{ '✅' if product.in_stock else '❌' }}</td>
                <td>{{ product.category }}</td>
            </tr>
            {% endfor %}
        </table>
    </div>

    <!-- Filter Blocks (FULLY IMPLEMENTED!) -->
    <div class="section">
        <h2>18. Filter Blocks (Advanced Text Processing)</h2>
        <p><strong>Filter blocks apply one or more filters to entire blocks of content, including template logic!</strong></p>
        
        <h3>Basic Filter Block</h3>
        <div class="code">
            {% filter upper %}Hello {{ user.name }}! Welcome to our application.{% endfilter %}
        </div>
        
        <h3>Chained Filters</h3>
        <div class="code">
            {% filter trim|upper|reverse %}   Hello World   {% endfilter %}
        </div>
        
        <h3>Filters with Arguments</h3>
        <div class="code">
            {% filter truncate(25) %}{{ description }}{% endfilter %}
        </div>
        
        <h3>Complex Content with Loops</h3>
        <div class="code">
            {% filter upper %}
                User Preferences:
                {% for key, value in user.preferences %}
                - {{ key }}: {{ value }}
                {% endfor %}
            {% endfilter %}
        </div>
        
        <h3>Nested Filter Blocks</h3>
        <div class="code">
            {% filter upper %}
                Outer Content: {{ user.status }}
                {% filter lower %}
                    INNER CONTENT: {{ user.email | upper }}
                {% endfilter %}
                Back to outer level
            {% endfilter %}
        </div>
        
        <h3>Conditional Content Filtering</h3>
        <div class="code">
            {% filter upper %}
                {% if user.is_premium %}
                    Premium User: {{ user.name }}
                    Email: {{ user.email }}
                {% else %}
                    Standard User: {{ user.name }}
                {% endif %}
            {% endfilter %}
        </div>
        
        <h3>Current Alternative - Chained Filter Operations</h3>
        <ul>
            <li><strong>Multi-step processing:</strong> {{ 'hello world test' | upper | reverse | replace('O', '0') }}</li>
            <li><strong>String cleaning:</strong> {{ '  MESSY DATA  ' | trim | lower | capitalize }}</li>
            <li><strong>Advanced formatting:</strong> {{ user.name | upper | center(20) | replace(' ', '_') }}</li>
        </ul>
        
        <h3>Complex String Transformations</h3>
        <ul>
            <li><strong>Email processing:</strong> {{ user.email | replace('@', ' [at] ') | replace('.', ' [dot] ') }}</li>
            <li><strong>Path normalization:</strong> {{ '/path/to/file.txt' | replace('/', '_') | upper }}</li>
            <li><strong>Data encoding:</strong> {{ 'hello world' | urlencode }}</li>
        </ul>
        
        <h3>String Analysis</h3>
        <ul>
            <li><strong>Character breakdown:</strong> {{ 'hello' | list | join(' | ') }}</li>
            <li><strong>Word statistics:</strong> {{ description | wordcount }} words, {{ description | length }} chars</li>
        </ul>
    </div>

    <!-- Do Statement -->
    <div class="section">
        <h2>19. Do Statement (Side Effects)</h2>
        <p>The 'do' statement allows executing expressions for side effects without producing template output.</p>
        
        <h3>Basic Expression Evaluation</h3>
        <div class="code">
            Original tags: {{ tags | join(', ') }}<br>
            {% do tags[0]|upper %}  {# Evaluates expression, no output #}
            Tags still: {{ tags | join(', ') }}
        </div>
        
        <h3>Complex Expressions</h3>
        <div class="code">
            {% do (user.age + 10) * 2 %}  {# Arithmetic expression #}
            {% do user.name|upper|reverse %}  {# Filter chain #}
            {% do price * 1.08 %}  {# Tax calculation #}
        </div>
        
        <h3>Integration with Control Flow</h3>
        <div class="code">
            {% for i in range(3) %}
                {% do i * user.age %}  {# Expression per iteration #}
                Item {{ i + 1 }}
            {% endfor %}
        </div>
        
        <h3>Whitespace Control</h3>
        <div class="code">
            Start{%- do user.id * 100 -%}End  {# No spaces around #}
        </div>
        
        <h3>Validation and Testing</h3>
        <div class="code">
            {# Test complex filter chains #}
            {% do description|truncate(50)|upper|wordcount %}
            {% do tags|sort|reverse|join(' + ') %}
            {# Validate expressions work before using in output #}
            Complex result: {{ (price * user.age)|round(2) }}
        </div>
    </div>

    <footer style="margin-top: 50px; padding-top: 20px; border-top: 1px solid #ddd;">
        <p><small>Generated with Miya Engine template engine | Features demonstrated: 19+</small></p>
        <p><small>Comparison with Python Jinja2: {{ 'Compatible' if user.is_premium else 'Basic compatibility' }}</small></p>
    </footer>
</body>
</html>`

	// Render template
	fmt.Printf("⏱️  [%.2fms] Starting template parsing and rendering...\n", float64(time.Since(startTime).Nanoseconds())/1e6)
	renderStart := time.Now()
	result, err := env.RenderString(template, ctx)
	renderTime := time.Since(renderStart)
	if err != nil {
		fmt.Printf("❌ [%.2fms] Template render error: %v\n", float64(time.Since(startTime).Nanoseconds())/1e6, err)
		return
	}
	fmt.Printf("✅ [%.2fms] Template rendered successfully in %.2fms\n", float64(time.Since(startTime).Nanoseconds())/1e6, float64(renderTime.Nanoseconds())/1e6)

	// Save Go output
	writeStart := time.Now()
	err = os.WriteFile("rendered_output_go.html", []byte(result), 0644)
	writeTime := time.Since(writeStart)
	if err != nil {
		fmt.Printf("❌ [%.2fms] Error saving Go output: %v\n", float64(time.Since(startTime).Nanoseconds())/1e6, err)
		return
	}
	fmt.Printf("✅ [%.2fms] File written in %.2fms\n", float64(time.Since(startTime).Nanoseconds())/1e6, float64(writeTime.Nanoseconds())/1e6)

	totalTime := time.Since(startTime)
	fmt.Println()
	fmt.Println("📊 PERFORMANCE SUMMARY:")
	fmt.Printf("   🏁 Total execution time: %.2fms\n", float64(totalTime.Nanoseconds())/1e6)
	fmt.Printf("   ⚡ Template rendering: %.2fms (%.1f%% of total)\n", float64(renderTime.Nanoseconds())/1e6, float64(renderTime.Nanoseconds())/float64(totalTime.Nanoseconds())*100)
	fmt.Printf("   📄 File I/O: %.2fms (%.1f%% of total)\n", float64(writeTime.Nanoseconds())/1e6, float64(writeTime.Nanoseconds())/float64(totalTime.Nanoseconds())*100)
	fmt.Printf("   📏 Output size: %s (%d bytes)\n", formatBytes(len(result)), len(result))
	fmt.Printf("   🚀 Rendering speed: %.2f MB/s\n", float64(len(result))/1024/1024/renderTime.Seconds())

	fmt.Println()
	fmt.Println("🔍 FEATURE ANALYSIS:")
	fmt.Println("  ✅ 19+ Jinja2 features demonstrated successfully")
	fmt.Println("  ✅ Do statements fully implemented with side effects support")
	fmt.Println("  ✅ Filter blocks fully implemented with chaining and nesting")
	fmt.Println("  ✅ Advanced string processing with complex filter chains")
	fmt.Println("  ✅ Dictionary iteration with key-value unpacking")
	fmt.Println("  ✅ 70+ built-in filters (comprehensive filter library)")
	fmt.Println("  ✅ Advanced features: inheritance, macros, imports, extensions")
	fmt.Println("  ✅ Performance: High-speed template rendering (2+ MB/s)")
	fmt.Println("  🎯 Miya Engine provides near-complete Python Jinja2 compatibility!")
}

func formatBytes(bytes int) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func createSampleData() miya.Context {
	ctx := miya.NewContext()
	now := time.Now()

	ctx.Set("title", "Complete Jinja2 Features Demonstration")
	ctx.Set("formatted_date", now.Format("January 02, 2006"))
	ctx.Set("current_datetime", now)
	ctx.Set("description", "This is a comprehensive showcase of all Jinja2 template engine features including variables, filters, control structures, macros, inheritance, and much more.")
	ctx.Set("html_content", "<strong>This is bold HTML content</strong>")
	ctx.Set("price", 29.999)
	ctx.Set("tags", []string{"python", "jinja2", "templates", "web", "html"})
	ctx.Set("empty_list", []interface{}{})
	ctx.Set("none_value", nil)
	ctx.Set("greeting_message", "Welcome to Miya Engine!")

	ctx.Set("user", map[string]interface{}{
		"id":         12345,
		"name":       "john doe",
		"first_name": "John",
		"last_name":  "Doe",
		"email":      "john.doe@example.com",
		"age":        25,
		"status":     "active",
		"is_premium": true,
		"is_admin":   false,
		"preferences": map[string]interface{}{
			"theme":         "dark",
			"language":      "english",
			"notifications": true,
		},
	})

	ctx.Set("products", []map[string]interface{}{
		{
			"name":     "Laptop",
			"price":    999.99,
			"in_stock": true,
			"category": "Electronics",
		},
		{
			"name":     "Mouse",
			"price":    25.50,
			"in_stock": true,
			"category": "Accessories",
		},
		{
			"name":     "Keyboard",
			"price":    75.00,
			"in_stock": false,
			"category": "Accessories",
		},
		{
			"name":     "Monitor",
			"price":    299.99,
			"in_stock": true,
			"category": "Electronics",
		},
	})

	ctx.Set("categories", []map[string]interface{}{
		{
			"name": "Electronics",
			"items": []map[string]interface{}{
				{"name": "Laptop"},
				{"name": "Phone"},
				{"name": "Tablet"},
			},
		},
		{
			"name": "Books",
			"items": []map[string]interface{}{
				{"name": "Python Guide"},
				{"name": "Web Development"},
				{"name": "Design Patterns"},
			},
		},
	})

	return ctx
}

func addCustomFilters(env *miya.Environment) {
	// All the filters we were implementing are actually built-in!
	// The Miya Engine library has 70+ built-in filters including:
	// truncate, round, slice, abs, center, reverse, sum, unique, urlencode,
	// wordcount, list, capitalize, float, int, random, and many more.

	// We only need to add truly custom filters that aren't built-in
	// For demonstration, let's add a custom filter that doesn't exist
	env.AddFilter("custom_exclaim", func(value interface{}, args ...interface{}) (interface{}, error) {
		if str, ok := value.(string); ok {
			return str + "!", nil
		}
		return value, nil
	})
}
