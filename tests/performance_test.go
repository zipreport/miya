package tests

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/zipreport/miya/tests/helpers"
)

// TestPerformanceBenchmarks covers performance testing and benchmarks
func TestPerformanceBenchmarks(t *testing.T) {
	env := helpers.CreateEnvironment()

	// Test template compilation performance
	t.Run("Template Compilation Speed", func(t *testing.T) {
		template := `
		{% for category in categories %}
		  <section>
		    <h2>{{ category.name|title }}</h2>
		    {% for product in category.products %}
		      <div class="product">
		        <h3>{{ product.name }}</h3>
		        <p>${{ product.price|round(2) }}</p>
		        {% if product.tags %}
		          <div class="tags">
		            {% for tag in product.tags %}
		              <span class="tag">{{ tag|upper }}</span>
		            {% endfor %}
		          </div>
		        {% endif %}
		      </div>
		    {% endfor %}
		  </section>
		{% endfor %}`

		start := time.Now()
		for i := 0; i < 100; i++ {
			_, err := env.FromString(template)
			if err != nil {
				t.Fatalf("Template compilation failed: %v", err)
			}
		}
		duration := time.Since(start)

		avgTime := duration / 100
		t.Logf("Average template compilation time: %v", avgTime)

		if avgTime > time.Millisecond*10 {
			t.Errorf("Template compilation too slow: %v (expected < 10ms)", avgTime)
		}
	})

	// Test rendering performance with large datasets
	t.Run("Large Dataset Rendering", func(t *testing.T) {
		template := `
		{% set total = 0 %}
		<div class="results">
		{% for category in categories %}
		  <section data-id="{{ category.id }}">
		    <h2>{{ category.name|upper }}</h2>
		    <div class="products">
		    {% for product in category.products %}
		      {% set total = total + product.price %}
		      <div class="product" data-price="{{ product.price }}">
		        <h3>{{ product.name|title }}</h3>
		        <p class="price">${{ product.price|round(2) }}</p>
		        <p class="desc">{{ product.description|truncate(50) }}</p>
		        {% if product.tags %}
		          {% filter upper %}
		          Tags: {{ product.tags|join(", ") }}
		          {% endfilter %}
		        {% endif %}
		        {% set discount = product.price * 0.1 %}
		        <p class="sale">${{ (product.price - discount)|round(2) }}</p>
		      </div>
		    {% endfor %}
		    </div>
		    <p class="count">Products: {{ category.products|length }}</p>
		  </section>
		{% endfor %}
		<footer>Total Value: ${{ total|round(2) }}</footer>
		</div>`

		// Generate large dataset - 50 categories with 100 products each = 5000 products
		categories := make([]map[string]interface{}, 50)
		for i := 0; i < 50; i++ {
			products := make([]map[string]interface{}, 100)
			for j := 0; j < 100; j++ {
				products[j] = map[string]interface{}{
					"name":        fmt.Sprintf("Product %d-%d", i+1, j+1),
					"price":       float64(10+i+j) * 1.99,
					"description": fmt.Sprintf("Description for product %d-%d with various details and information", i+1, j+1),
					"tags":        []string{"tag1", "tag2", "tag3"},
				}
			}
			categories[i] = map[string]interface{}{
				"id":       fmt.Sprintf("cat_%d", i+1),
				"name":     fmt.Sprintf("category %d", i+1),
				"products": products,
			}
		}

		ctx := helpers.CreateContextFrom(map[string]interface{}{
			"categories": categories,
		})

		// Warm up
		_, err := env.RenderString(template, ctx)
		if err != nil {
			t.Fatalf("Warmup render failed: %v", err)
		}

		// Measure rendering performance
		start := time.Now()
		result, err := env.RenderString(template, ctx)
		duration := time.Since(start)

		if err != nil {
			t.Fatalf("Large dataset render failed: %v", err)
		}

		t.Logf("Large dataset rendering time: %v", duration)
		t.Logf("Output size: %d characters", len(result))
		t.Logf("Rendering speed: %.2f MB/s", float64(len(result))/1024/1024/duration.Seconds())

		if duration > time.Second*5 {
			t.Errorf("Large dataset rendering too slow: %v (expected < 5s)", duration)
		}

		// Verify output contains expected content
		if !strings.Contains(result, "CATEGORY 1") {
			t.Error("Expected output to contain category names")
		}
		if !strings.Contains(result, "Product 1-1") {
			t.Error("Expected output to contain product names")
		}
	})

	// Test memory efficiency
	t.Run("Memory Efficiency", func(t *testing.T) {
		template := `
		{% for i in range(1000) %}
		  <div class="item-{{ i }}">
		    {% set complex_data = "data-" ~ i ~ "-" ~ (i * 2) ~ "-processed" %}
		    <span>{{ complex_data|upper }}</span>
		    {% if i % 10 == 0 %}
		      {% filter upper %}
		      Milestone: {{ i }} / 1000
		      {% endfilter %}
		    {% endif %}
		  </div>
		{% endfor %}`

		ctx := helpers.CreateContextFrom(map[string]interface{}{})

		start := time.Now()
		result, err := env.RenderString(template, ctx)
		duration := time.Since(start)

		if err != nil {
			t.Fatalf("Memory efficiency test failed: %v", err)
		}

		t.Logf("Memory efficiency test time: %v", duration)
		t.Logf("Generated %d characters", len(result))

		if duration > time.Millisecond*500 {
			t.Errorf("Memory efficiency test too slow: %v (expected < 500ms)", duration)
		}

		// Verify some content
		if !strings.Contains(result, "item-0") || !strings.Contains(result, "item-999") {
			t.Error("Expected output to contain first and last items")
		}
	})

	// Test concurrent rendering
	t.Run("Concurrent Rendering Safety", func(t *testing.T) {
		template := `
		<div class="user-{{ user.id }}">
		  <h1>{{ user.name|title }}</h1>
		  <p>Email: {{ user.email|lower }}</p>
		  <p>Age: {{ user.age }}</p>
		  {% if user.preferences %}
		    <div class="preferences">
		    {% for key, value in user.preferences %}
		      <p>{{ key }}: {{ value }}</p>
		    {% endfor %}
		    </div>
		  {% endif %}
		  <p>Generated at: {{ timestamp }}</p>
		</div>`

		// Test concurrent rendering with different data
		const numGoroutines = 10
		const rendersPerGoroutine = 50

		results := make(chan error, numGoroutines)

		for g := 0; g < numGoroutines; g++ {
			go func(goroutineID int) {
				for i := 0; i < rendersPerGoroutine; i++ {
					ctx := helpers.CreateContextFrom(map[string]interface{}{
						"user": map[string]interface{}{
							"id":    goroutineID*1000 + i,
							"name":  fmt.Sprintf("user %d-%d", goroutineID, i),
							"email": fmt.Sprintf("user%d-%d@example.com", goroutineID, i),
							"age":   20 + (goroutineID+i)%50,
							"preferences": map[string]interface{}{
								"theme":    "dark",
								"language": "en",
							},
						},
						"timestamp": time.Now().Unix(),
					})

					result, err := env.RenderString(template, ctx)
					if err != nil {
						results <- fmt.Errorf("goroutine %d, render %d failed: %v", goroutineID, i, err)
						return
					}

					// Verify the result contains expected content
					expectedUser := fmt.Sprintf("user-%d", goroutineID*1000+i)
					if !strings.Contains(result, expectedUser) {
						results <- fmt.Errorf("goroutine %d, render %d: missing expected content", goroutineID, i)
						return
					}
				}
				results <- nil
			}(g)
		}

		// Wait for all goroutines to complete
		for i := 0; i < numGoroutines; i++ {
			err := <-results
			if err != nil {
				t.Errorf("Concurrent rendering error: %v", err)
			}
		}

		t.Logf("Successfully completed %d concurrent renders (%d goroutines × %d renders each)",
			numGoroutines*rendersPerGoroutine, numGoroutines, rendersPerGoroutine)
	})

	// Test filter chain performance
	t.Run("Filter Chain Performance", func(t *testing.T) {
		template := `
		{% for item in items %}
		  {{ item.text|trim|lower|title|replace(" ", "_")|upper|reverse|truncate(20)|center(25) }}
		{% endfor %}`

		items := make([]map[string]interface{}, 1000)
		for i := 0; i < 1000; i++ {
			items[i] = map[string]interface{}{
				"text": fmt.Sprintf("  Sample Text Item Number %d With Extra Spacing  ", i+1),
			}
		}

		ctx := helpers.CreateContextFrom(map[string]interface{}{
			"items": items,
		})

		start := time.Now()
		result, err := env.RenderString(template, ctx)
		duration := time.Since(start)

		if err != nil {
			t.Fatalf("Filter chain performance test failed: %v", err)
		}

		t.Logf("Filter chain performance: %v for 1000 items", duration)
		t.Logf("Average per item: %v", duration/1000)

		if duration > time.Millisecond*200 {
			t.Errorf("Filter chain performance too slow: %v (expected < 200ms)", duration)
		}

		// Verify processing occurred
		if len(result) < 1000 { // Should have substantial output
			t.Error("Expected substantial output from filter chain processing")
		}
	})
}

// BenchmarkBasicRendering benchmarks basic template rendering
func BenchmarkBasicRendering(b *testing.B) {
	env := helpers.CreateEnvironment()
	template := `Hello {{ name }}! You have {{ count }} messages.`
	ctx := helpers.CreateContextFrom(map[string]interface{}{
		"name":  "User",
		"count": 5,
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := env.RenderString(template, ctx)
		if err != nil {
			b.Fatalf("Benchmark render failed: %v", err)
		}
	}
}

// BenchmarkComplexRendering benchmarks complex template rendering
func BenchmarkComplexRendering(b *testing.B) {
	env := helpers.CreateEnvironment()
	template := `
	{% for item in items %}
	  <div class="item">
	    <h3>{{ item.name|title }}</h3>
	    <p>${{ item.price|round(2) }}</p>
	    {% if item.tags %}
	      {% for tag in item.tags %}
	        <span class="tag">{{ tag|upper }}</span>
	      {% endfor %}
	    {% endif %}
	  </div>
	{% endfor %}`

	items := make([]map[string]interface{}, 100)
	for i := 0; i < 100; i++ {
		items[i] = map[string]interface{}{
			"name":  fmt.Sprintf("item %d", i),
			"price": float64(i) * 1.99,
			"tags":  []string{"tag1", "tag2", "tag3"},
		}
	}

	ctx := helpers.CreateContextFrom(map[string]interface{}{
		"items": items,
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := env.RenderString(template, ctx)
		if err != nil {
			b.Fatalf("Complex benchmark render failed: %v", err)
		}
	}
}

// BenchmarkFilterChains benchmarks filter chain performance
func BenchmarkFilterChains(b *testing.B) {
	env := helpers.CreateEnvironment()
	template := `{{ text|trim|upper|reverse|truncate(20)|center(25)|replace(" ", "_") }}`
	ctx := helpers.CreateContextFrom(map[string]interface{}{
		"text": "  sample text for filter chain benchmarking  ",
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := env.RenderString(template, ctx)
		if err != nil {
			b.Fatalf("Filter chain benchmark failed: %v", err)
		}
	}
}

// BenchmarkTemplateCompilation benchmarks template compilation
func BenchmarkTemplateCompilation(b *testing.B) {
	env := helpers.CreateEnvironment()
	template := `
	{% for category in categories %}
	  <section>
	    <h2>{{ category.name }}</h2>
	    {% for product in category.products %}
	      <div>
	        <h3>{{ product.name }}</h3>
	        <p>${{ product.price }}</p>
	        {% if product.available %}
	          <span class="available">Available</span>
	        {% endif %}
	      </div>
	    {% endfor %}
	  </section>
	{% endfor %}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := env.FromString(template)
		if err != nil {
			b.Fatalf("Template compilation benchmark failed: %v", err)
		}
	}
}

// TestPerformanceRegression tests for performance regressions
func TestPerformanceRegression(t *testing.T) {
	env := helpers.CreateEnvironment()

	// Define performance baselines (these could be adjusted based on hardware)
	baselines := map[string]time.Duration{
		"simple_render":    time.Microsecond * 100,
		"complex_render":   time.Millisecond * 10,
		"filter_chain":     time.Microsecond * 500,
		"large_dataset":    time.Second * 2,
		"template_compile": time.Millisecond * 5,
	}

	tests := []struct {
		name     string
		template string
		context  map[string]interface{}
		baseline time.Duration
	}{
		{
			name:     "simple_render",
			template: `Hello {{ name }}!`,
			context:  map[string]interface{}{"name": "World"},
			baseline: baselines["simple_render"],
		},
		{
			name: "complex_render",
			template: `
			{% for user in users %}
			  <div>
			    <h3>{{ user.name|title }}</h3>
			    {% if user.active %}
			      <p>Status: {{ user.status|upper }}</p>
			      {% for role in user.roles %}
			        <span>{{ role }}</span>
			      {% endfor %}
			    {% endif %}
			  </div>
			{% endfor %}`,
			context: map[string]interface{}{
				"users": []map[string]interface{}{
					{"name": "alice", "active": true, "status": "online", "roles": []string{"admin", "user"}},
					{"name": "bob", "active": false, "status": "offline", "roles": []string{"user"}},
					{"name": "charlie", "active": true, "status": "busy", "roles": []string{"moderator", "user"}},
				},
			},
			baseline: baselines["complex_render"],
		},
		{
			name:     "filter_chain",
			template: `{{ text|trim|upper|reverse|truncate(10)|center(15) }}`,
			context:  map[string]interface{}{"text": "  hello world  "},
			baseline: baselines["filter_chain"],
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := helpers.CreateContextFrom(test.context)

			// Warmup
			_, err := env.RenderString(test.template, ctx)
			if err != nil {
				t.Fatalf("Warmup failed: %v", err)
			}

			// Measure performance
			const iterations = 10
			start := time.Now()
			for i := 0; i < iterations; i++ {
				_, err := env.RenderString(test.template, ctx)
				if err != nil {
					t.Fatalf("Performance test failed: %v", err)
				}
			}
			avgTime := time.Since(start) / iterations

			t.Logf("%s average time: %v (baseline: %v)", test.name, avgTime, test.baseline)

			// Allow 3x baseline for performance regression detection
			maxAllowed := test.baseline * 3
			if avgTime > maxAllowed {
				t.Errorf("Performance regression detected: %v > %v (3x baseline)", avgTime, maxAllowed)
			}
		})
	}
}

// Helper functions are in tests/helpers package
