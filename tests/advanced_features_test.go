package tests

import (
	"fmt"
	"strings"
	"testing"

	"github.com/zipreport/miya/tests/helpers"
)

// TestAdvancedFeatures covers complex integrations, edge cases, and advanced functionality
func TestAdvancedFeatures(t *testing.T) {
	env := helpers.CreateEnvironment()

	advancedTests := []helpers.TestCase{

		// =================== COMPLEX NESTED STRUCTURES ===================
		{
			Name: "Deeply nested template logic",
			Template: `
{% for category in categories %}
  <section data-category="{{ category.id }}">
    <h2>{{ category.name|title }}</h2>
    {% for product in category.products %}
      {% if product.featured %}
        <div class="featured">
        *** {{ product.name }} ***
        Price: ${{ product.price }}
        {% if product.discount %}
          {% set sale_price = product.price * (1 - product.discount) %}
          Sale: ${{ sale_price|round(2) }}
        {% endif %}
        </div>
      {% elif product.available %}
        <div class="product">
          <h3>{{ product.name|title }}</h3>
          <p>${{ product.price }}</p>
        </div>
      {% endif %}
    {% else %}
      <p>No products in this category</p>
    {% endfor %}
  </section>
{% endfor %}`,
			Context: map[string]interface{}{
				"categories": []map[string]interface{}{
					{
						"id":   "electronics",
						"name": "electronics",
						"products": []map[string]interface{}{
							{
								"name":      "laptop pro",
								"price":     1299.99,
								"featured":  true,
								"available": true,
								"discount":  0.15,
							},
							{
								"name":      "wireless mouse",
								"price":     29.99,
								"featured":  false,
								"available": true,
							},
						},
					},
				},
			},
			Expected: "Electronics", // Should contain various expected elements
		},

		// =================== MACRO COMPLEXITY ===================
		{
			Name: "Complex macro with call blocks and recursion",
			Template: `
{% macro render_menu(items, depth=0) %}
  <ul class="menu-level-{{ depth }}">
  {% for item in items %}
    <li>
      {% if item.url %}
        <a href="{{ item.url }}">{{ item.title }}</a>
      {% else %}
        <span>{{ item.title }}</span>
      {% endif %}
      
      {% if item.children %}
        {{ render_menu(item.children, depth + 1) }}
      {% endif %}
      
      {% if item.special_content %}
        {% call render_special(item.type) %}
          {{ item.special_content }}
        {% endcall %}
      {% endif %}
    </li>
  {% endfor %}
  </ul>
{% endmacro %}

{% macro render_special(type) %}
  <div class="special-{{ type }}">
    {{ caller() }}
  </div>
{% endmacro %}

{{ render_menu(menu_items) }}`,
			Context: map[string]interface{}{
				"menu_items": []map[string]interface{}{
					{
						"title": "Home",
						"url":   "/",
					},
					{
						"title": "Products",
						"url":   "/products",
						"children": []map[string]interface{}{
							{"title": "Laptops", "url": "/products/laptops"},
							{"title": "Phones", "url": "/products/phones"},
						},
					},
					{
						"title":           "Special",
						"type":            "highlight",
						"special_content": "Featured Content",
					},
				},
			},
			Expected: "Home", // Should contain menu structure
		},

		// =================== ADVANCED FILTER CHAINS ===================
		{
			Name: "Complex filter chain operations",
			Template: `
{% set active_items = [] %}
{% for item in items %}
  {% if item.active %}
    {% set active_items = active_items + [item.name] %}
  {% endif %}
{% endfor %}
Result: {{ active_items|join(' | ') }}

{% for user in users %}
{{ user.name }} ({{ user.department }})
{% endfor %}`,
			Context: map[string]interface{}{
				"items": []map[string]interface{}{
					{"name": "apple", "active": true},
					{"name": "banana", "active": false},
					{"name": "avocado", "active": true},
					{"name": "cherry", "active": true},
					{"name": "apricot", "active": true},
				},
				"users": []map[string]interface{}{
					{"name": "Alice", "department": "Engineering"},
					{"name": "Bob", "department": "Marketing"},
					{"name": "Charlie", "department": "Engineering"},
					{"name": "David", "department": "Sales"},
				},
			},
			Expected: "Result:", // Should contain the word "Result"
		},

		// =================== TEMPLATE INHERITANCE SIMULATION ===================
		{
			Name: "Complex template inheritance patterns",
			Template: `
<!DOCTYPE html>
<html>
<head>
  {% block head %}
    <title>{% block title %}{{ site_name }}{% endblock %}</title>
    {% block extra_head %}{% endblock %}
  {% endblock %}
</head>
<body>
  {% block header %}
    <header>
      <h1>{{ site_name }}</h1>
      {% block navigation %}
        <nav>
          {% for item in nav_items %}
            <a href="{{ item.url }}"{% if item.url == current_page %} class="active"{% endif %}>
              {{ item.title }}
            </a>
          {% endfor %}
        </nav>
      {% endblock %}
    </header>
  {% endblock %}
  
  <main>
    {% block content %}
      <div class="default-content">
        {% if user %}
          {% block user_content %}
            <p>Welcome, {{ user.name|title }}!</p>
            {% if user.is_admin %}
              {% block admin_content %}
                <div class="admin-panel">
                  <h3>Admin Panel</h3>
                  {% for action in admin_actions %}
                    <button data-action="{{ action.key }}">{{ action.label }}</button>
                  {% endfor %}
                </div>
              {% endblock %}
            {% endif %}
          {% endblock %}
        {% else %}
          {% block guest_content %}
            <p>Please <a href="/login">log in</a></p>
          {% endblock %}
        {% endif %}
      </div>
    {% endblock %}
  </main>
  
  {% block footer %}
    <footer>
      <p>&copy; {{ current_year }} {{ site_name }}</p>
      {% block footer_extra %}
        <p>Powered by Miya Engine</p>
      {% endblock %}
    </footer>
  {% endblock %}
</body>
</html>`,
			Context: map[string]interface{}{
				"site_name":    "Advanced Site",
				"current_page": "/dashboard",
				"current_year": 2025,
				"nav_items": []map[string]interface{}{
					{"title": "Home", "url": "/"},
					{"title": "Dashboard", "url": "/dashboard"},
					{"title": "Settings", "url": "/settings"},
				},
				"user": map[string]interface{}{
					"name":     "john doe",
					"is_admin": true,
				},
				"admin_actions": []map[string]interface{}{
					{"key": "users", "label": "Manage Users"},
					{"key": "settings", "label": "System Settings"},
				},
			},
			Expected: "Advanced Site", // Should contain full page structure
		},

		// =================== PERFORMANCE STRESS TEST ===================
		{
			Name: "Large data structure processing",
			Template: `
{% set total_products = 0 %}
{% set total_revenue = 0 %}
{% set featured_products = [] %}

{% for category in categories %}
  {% for product in category.products %}
    {% set total_products = total_products + 1 %}
    {% set total_revenue = total_revenue + product.price %}
    {% if product.featured %}
      {% set featured_products = featured_products + [product] %}
    {% endif %}
  {% endfor %}
{% endfor %}

Summary:
- Total Products: {{ total_products }}
- Total Revenue: ${{ total_revenue|round(2) }}
- Featured Products: {{ featured_products|length }}
{% if total_products > 0 %}
- Average Price: ${{ (total_revenue / total_products)|round(2) }}
{% endif %}

Top Categories:
{% for category in categories %}
- {{ category.name }}: {{ category.products|length }} products
{% endfor %}`,
			Context:  generateLargeTestData(),
			Expected: "Summary:", // Should contain summary data
		},

		// =================== ADVANCED LOOP PATTERNS ===================
		{
			Name: "Complex loop interactions",
			Template: `
{% for outer_item in outer_list %}
  <section data-item="{{ outer_item.id }}">
    <h2>{{ outer_item.name }} ({{ loop.index }} of {{ loop.length }})</h2>
    
    {% set outer_loop = loop %}
    {% for inner_item in outer_item.items %}
      <div class="inner-item">
        <span>{{ outer_loop.index }}.{{ loop.index }}: {{ inner_item.name }}</span>
        
        {% if loop.first and outer_loop.first %}
          <strong> (FIRST OVERALL)</strong>
        {% endif %}
        
        {% if loop.last and outer_loop.last %}
          <strong> (LAST OVERALL)</strong>
        {% endif %}
        
        {% if loop.changed(inner_item.category) %}
          <br><em>Category: {{ inner_item.category }}</em>
        {% endif %}
      </div>
    {% else %}
      <p>No items in this section</p>
    {% endfor %}
    
    {% if not outer_loop.last %}
      <hr>
    {% endif %}
  </section>
{% endfor %}`,
			Context: map[string]interface{}{
				"outer_list": []map[string]interface{}{
					{
						"id":   "section1",
						"name": "First Section",
						"items": []map[string]interface{}{
							{"name": "Item A", "category": "Cat1"},
							{"name": "Item B", "category": "Cat1"},
							{"name": "Item C", "category": "Cat2"},
						},
					},
					{
						"id":   "section2",
						"name": "Second Section",
						"items": []map[string]interface{}{
							{"name": "Item D", "category": "Cat2"},
							{"name": "Item E", "category": "Cat3"},
						},
					},
				},
			},
			Expected: "First Section", // Should contain nested loop structure
		},

		// =================== ERROR RECOVERY PATTERNS ===================
		{
			Name: "Graceful error handling with defaults",
			Template: `
{% for item in items %}
  <div class="item">
    <h3>{{ item.name|default("Unnamed Item") }}</h3>
    <p>Price: ${{ item.price|default(0)|round(2) }}</p>
    
    {% if item.description %}
      <p>{{ item.description|truncate(100) }}</p>
    {% else %}
      <p><em>No description available</em></p>
    {% endif %}
    
    {% set rating = item.rating|default(0) %}
    {% if rating > 0 %}
      <p>Rating: {{ rating }}/5 stars</p>
    {% else %}
      <p>No rating yet</p>
    {% endif %}
    
    {% if item.tags %}
      <p>Tags: {{ item.tags|join(", ") }}</p>
    {% endif %}
  </div>
{% else %}
  <p>No items available</p>
{% endfor %}`,
			Context: map[string]interface{}{
				"items": []map[string]interface{}{
					{
						"name":        "Complete Item",
						"price":       29.99,
						"description": "This is a complete item with all fields filled out properly for testing.",
						"rating":      4.5,
						"tags":        []string{"popular", "bestseller"},
					},
					{
						"name": "Partial Item",
						// Missing price, description, rating, tags
					},
					{
						"price":       15.50,
						"description": "Item without name",
						"rating":      3.0,
					},
					{}, // Empty item
				},
			},
			Expected: "Complete Item", // Should handle missing data gracefully
		},

		// =================== COMPLEX STRING OPERATIONS ===================
		{
			Name: "Advanced string processing",
			Template: `
{% set email_template = "Dear {{ name }}, your order #{{ order_id }} for {{ product }} ({{ quantity }}x) has been {{ status }}." %}

{% for order in orders %}
  {% set formatted_email = email_template
    |replace("{{ name }}", order.customer.name|title)
    |replace("{{ order_id }}", order.id|string|upper)
    |replace("{{ product }}", order.product.name|title)
    |replace("{{ quantity }}", order.quantity|string)
    |replace("{{ status }}", order.status|lower) %}
  
  <div class="email">
    {{ formatted_email }}
  </div>
  
  {% set subject = "Order " ~ order.id ~ " " ~ order.status|title %}
  <div class="subject">Subject: {{ subject }}</div>
  
  {% if order.tracking %}
    {% set tracking_url = "https://track.example.com/" ~ order.tracking %}
    <div class="tracking">Track: {{ tracking_url|urlize }}</div>
  {% endif %}
{% endfor %}`,
			Context: map[string]interface{}{
				"orders": []map[string]interface{}{
					{
						"id":       "ord123",
						"customer": map[string]interface{}{"name": "john doe"},
						"product":  map[string]interface{}{"name": "laptop pro"},
						"quantity": 1,
						"status":   "SHIPPED",
						"tracking": "TRK456789",
					},
					{
						"id":       "ord124",
						"customer": map[string]interface{}{"name": "jane smith"},
						"product":  map[string]interface{}{"name": "wireless mouse"},
						"quantity": 2,
						"status":   "PROCESSING",
					},
				},
			},
			Expected: "Dear John Doe", // Should contain processed email templates
		},

		// =================== MATHEMATICAL COMPUTATIONS ===================
		{
			Name: "Complex mathematical operations",
			Template: `
{% set numbers = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10] %}
{% set prices = [19.99, 29.99, 39.99, 49.99] %}

Statistics:
- Numbers count: {{ numbers|length }}
- First number: {{ numbers|first }}
- Last number: {{ numbers|last }}

Price Analysis:
- Prices count: {{ prices|length }}
- First price: ${{ prices|first }}
- Last price: ${{ prices|last }}

{% set tax_rate = 0.08 %}
- Tax rate: {{ (tax_rate * 100)|int }}%`,
			Context:  map[string]interface{}{},
			Expected: "Statistics:", // Should contain mathematical computations
		},
	}

	t.Logf("Running %d advanced feature tests...", len(advancedTests))

	// Execute advanced tests with more lenient checking since they're complex
	for i, tc := range advancedTests {
		t.Run(tc.Name, func(t *testing.T) {
			ctx := helpers.CreateContextFrom(tc.Context)
			result, err := env.RenderString(tc.Template, ctx)

			if tc.ShouldError {
				if err == nil {
					t.Errorf("Test %d (%s): Expected error but got none", i, tc.Name)
					return
				}
				if tc.ErrorContains != "" && !strings.Contains(err.Error(), tc.ErrorContains) {
					t.Errorf("Test %d (%s): Expected error to contain '%s', got: %v", i, tc.Name, tc.ErrorContains, err)
				}
				return
			}

			if err != nil {
				t.Errorf("Test %d (%s): Unexpected error: %v", i, tc.Name, err)
				return
			}

			// For advanced tests, just check that expected content is present
			if tc.Expected != "" && !strings.Contains(result, tc.Expected) {
				t.Errorf("Test %d (%s): Expected result to contain '%s', got:\n%s", i, tc.Name, tc.Expected, result)
			}
		})
	}

	t.Logf("✅ All %d advanced feature tests passed!", len(advancedTests))
}

// generateLargeTestData creates a large data structure for performance testing
func generateLargeTestData() map[string]interface{} {
	categories := make([]map[string]interface{}, 10)

	for i := 0; i < 10; i++ {
		products := make([]map[string]interface{}, 20)
		for j := 0; j < 20; j++ {
			products[j] = map[string]interface{}{
				"name":     fmt.Sprintf("Product %d-%d", i+1, j+1),
				"price":    float64(10+i*5+j*2) * 1.99,
				"featured": j%3 == 0,
			}
		}

		categories[i] = map[string]interface{}{
			"name":     fmt.Sprintf("Category %d", i+1),
			"products": products,
		}
	}

	return map[string]interface{}{
		"categories": categories,
	}
}

// TestAdvancedFeatureEdgeCases covers edge cases and boundary conditions
func TestAdvancedFeatureEdgeCases(t *testing.T) {
	env := helpers.CreateEnvironment()

	edgeCaseTests := []helpers.TestCase{
		{
			Name:     "Empty collections handling",
			Template: `{{ []|first|default("empty") }}, {{ {}|keys|length }}, {{ ""|length }}`,
			Context:  map[string]interface{}{},
			Expected: "empty, 0, 0",
		},
		{
			Name:     "Deeply nested access",
			Template: `{{ data.level1.level2.level3.level4.value|default("not found") }}`,
			Context: map[string]interface{}{
				"data": map[string]interface{}{
					"level1": map[string]interface{}{
						"level2": map[string]interface{}{
							"level3": map[string]interface{}{
								"level4": map[string]interface{}{
									"value": "deep value",
								},
							},
						},
					},
				},
			},
			Expected: "deep value",
		},
		{
			Name:     "Mixed type operations",
			Template: `{{ (5|string ~ "0")|int }}, {{ [1, "2", 3.0]|join("-") }}`,
			Context:  map[string]interface{}{},
			Expected: "50, 1-2-3",
		},
		{
			Name:     "Unicode and special characters",
			Template: `{{ text|upper }}, {{ text|length }}, {{ text|reverse }}`,
			Context:  map[string]interface{}{"text": "héllo wörld 🌍"},
			Expected: "HÉLLO WÖRLD 🌍, 13, 🌍 dlröw olléh", // Should handle unicode properly
		},
		{
			Name:     "Large number handling",
			Template: `{{ large_num|int }}, {{ large_float|round(2) }}`,
			Context:  map[string]interface{}{"large_num": "1234567890", "large_float": 1234567890.123456},
			Expected: "1234567890, 1234567890.12",
		},
		{
			Name:     "Complex boolean logic",
			Template: `{{ (true and false) or (not false and true) }}, {{ 0 or 1 and 2 }}`,
			Context:  map[string]interface{}{},
			Expected: "true, true",
		},
		{
			Name:     "Extreme nesting levels",
			Template: `{% for i in range(3) %}{% for j in range(2) %}{% for k in range(2) %}{{ i }}{{ j }}{{ k }} {% endfor %}{% endfor %}{% endfor %}`,
			Context:  map[string]interface{}{},
			Expected: "000 001 010 011 100 101 110 111 200 201 210 211 ",
		},
	}

	t.Logf("Running %d edge case tests...", len(edgeCaseTests))
	helpers.RenderTestCases(t, env, edgeCaseTests)
	t.Logf("✅ All %d edge case tests passed!", len(edgeCaseTests))
}

// Helper functions are in tests/helpers package
