package miya_test

import (
	miya "github.com/zipreport/miya"
	"sync"
	"testing"
)

// =============================================================================
// BENCHMARK TESTS - CONSOLIDATED
// =============================================================================
// This file consolidates all benchmark tests from multiple individual files:
// - benchmark_test.go
// - cached_benchmark_test.go
// - concurrent_benchmark_test.go
// - filter_chain_benchmark_test.go
// =============================================================================

// Basic Template Rendering Benchmarks
func BenchmarkSimpleTemplateRendering(b *testing.B) {
	env := miya.NewEnvironment()
	template := "Hello {{ name }}!"
	data := map[string]interface{}{"name": "World"}

	tmpl, err := env.FromString(template)
	if err != nil {
		b.Fatalf("Failed to parse template: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := tmpl.Render(miya.NewContextFrom(data))
		if err != nil {
			b.Fatalf("Failed to render template: %v", err)
		}
	}
}

func BenchmarkComplexTemplateRendering(b *testing.B) {
	env := miya.NewEnvironment()
	template := `
{%- for user in users -%}
<div class="user-{{ loop.index }}">
  <h3>{{ user.name|title }}</h3>
  <p>Email: {{ user.email|lower }}</p>
  <p>Age: {{ user.age }}</p>
  {%- if user.active -%}
  <span class="active">Active</span>
  {%- else -%}
  <span class="inactive">Inactive</span>
  {%- endif -%}
</div>
{%- endfor -%}`

	data := map[string]interface{}{
		"users": []map[string]interface{}{
			{"name": "John", "email": "JOHN@EXAMPLE.COM", "age": 30, "active": true},
			{"name": "Jane", "email": "JANE@EXAMPLE.COM", "age": 25, "active": false},
			{"name": "Bob", "email": "BOB@EXAMPLE.COM", "age": 35, "active": true},
		},
	}

	tmpl, err := env.FromString(template)
	if err != nil {
		b.Fatalf("Failed to parse template: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := tmpl.Render(miya.NewContextFrom(data))
		if err != nil {
			b.Fatalf("Failed to render template: %v", err)
		}
	}
}

// Template Caching Benchmarks
func BenchmarkCachedTemplateParsing(b *testing.B) {
	env := miya.NewEnvironment()
	template := "Cached template with {{ variable }}"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// This should hit the cache after the first parse
		_, err := env.FromString(template)
		if err != nil {
			b.Fatalf("Failed to parse template: %v", err)
		}
	}
}

func BenchmarkCachedTemplateRendering(b *testing.B) {
	env := miya.NewEnvironment()
	template := "Cached rendering {{ name }} - {{ count }}"
	data := map[string]interface{}{"name": "Test", "count": 42}

	tmpl, err := env.FromString(template)
	if err != nil {
		b.Fatalf("Failed to parse template: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := tmpl.Render(miya.NewContextFrom(data))
		if err != nil {
			b.Fatalf("Failed to render template: %v", err)
		}
	}
}

// Inheritance caching benchmark disabled - requires complex setup
// func BenchmarkInheritanceCaching(b *testing.B) { ... }

// Concurrent Template Rendering Benchmarks
func BenchmarkConcurrentTemplateRendering(b *testing.B) {
	env := miya.NewEnvironment()
	template := "Concurrent {{ name }} - {{ value }}"

	tmpl, err := env.FromString(template)
	if err != nil {
		b.Fatalf("Failed to parse template: %v", err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			data := map[string]interface{}{
				"name":  "Test",
				"value": i,
			}
			_, err := tmpl.Render(miya.NewContextFrom(data))
			if err != nil {
				b.Fatalf("Failed to render template: %v", err)
			}
			i++
		}
	})
}

func BenchmarkConcurrentComplexRendering(b *testing.B) {
	env := miya.NewEnvironment()
	template := `
{%- for item in items -%}
<div>{{ item.name|upper }} - {{ item.value|abs }}</div>
{%- endfor -%}`

	tmpl, err := env.FromString(template)
	if err != nil {
		b.Fatalf("Failed to parse template: %v", err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			data := map[string]interface{}{
				"items": []map[string]interface{}{
					{"name": "item1", "value": -10},
					{"name": "item2", "value": -20},
					{"name": "item3", "value": -30},
				},
			}
			_, err := tmpl.Render(miya.NewContextFrom(data))
			if err != nil {
				b.Fatalf("Failed to render template: %v", err)
			}
			i++
		}
	})
}

func BenchmarkConcurrentCaching(b *testing.B) {
	env := miya.NewEnvironment()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			template := "Concurrent cached template {{ value }}"
			tmpl, err := env.FromString(template)
			if err != nil {
				b.Fatalf("Failed to parse template: %v", err)
			}

			data := map[string]interface{}{"value": i}
			_, err = tmpl.Render(miya.NewContextFrom(data))
			if err != nil {
				b.Fatalf("Failed to render template: %v", err)
			}
			i++
		}
	})
}

// Filter Chain Performance Benchmarks
func BenchmarkSimpleFilter(b *testing.B) {
	env := miya.NewEnvironment()
	template := "{{ value|upper }}"
	data := map[string]interface{}{"value": "hello world"}

	tmpl, err := env.FromString(template)
	if err != nil {
		b.Fatalf("Failed to parse template: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := tmpl.Render(miya.NewContextFrom(data))
		if err != nil {
			b.Fatalf("Failed to render template: %v", err)
		}
	}
}

func BenchmarkFilterChain(b *testing.B) {
	env := miya.NewEnvironment()
	template := "{{ value|upper|trim|reverse }}"
	data := map[string]interface{}{"value": "  hello world  "}

	tmpl, err := env.FromString(template)
	if err != nil {
		b.Fatalf("Failed to parse template: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := tmpl.Render(miya.NewContextFrom(data))
		if err != nil {
			b.Fatalf("Failed to render template: %v", err)
		}
	}
}

func BenchmarkLongFilterChain(b *testing.B) {
	env := miya.NewEnvironment()
	template := "{{ value|upper|trim|lower|title|strip|capitalize|reverse }}"
	data := map[string]interface{}{"value": "  hello world test  "}

	tmpl, err := env.FromString(template)
	if err != nil {
		b.Fatalf("Failed to parse template: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := tmpl.Render(miya.NewContextFrom(data))
		if err != nil {
			b.Fatalf("Failed to render template: %v", err)
		}
	}
}

func BenchmarkCollectionFilter(b *testing.B) {
	env := miya.NewEnvironment()
	template := "{{ items|sort|reverse|join(', ') }}"
	data := map[string]interface{}{
		"items": []string{"banana", "apple", "cherry", "date"},
	}

	tmpl, err := env.FromString(template)
	if err != nil {
		b.Fatalf("Failed to parse template: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := tmpl.Render(miya.NewContextFrom(data))
		if err != nil {
			b.Fatalf("Failed to render template: %v", err)
		}
	}
}

func BenchmarkComplexFilterChain(b *testing.B) {
	env := miya.NewEnvironment()
	template := "{{ users|selectattr('active')|map('name')|sort|join(' | ') }}"
	data := map[string]interface{}{
		"users": []map[string]interface{}{
			{"name": "John", "active": true},
			{"name": "Jane", "active": false},
			{"name": "Bob", "active": true},
			{"name": "Alice", "active": true},
		},
	}

	tmpl, err := env.FromString(template)
	if err != nil {
		b.Fatalf("Failed to parse template: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := tmpl.Render(miya.NewContextFrom(data))
		if err != nil {
			b.Fatalf("Failed to render template: %v", err)
		}
	}
}

// Memory Performance Benchmarks
func BenchmarkMemoryUsage(b *testing.B) {
	env := miya.NewEnvironment()
	template := `
{%- for i in range(100) -%}
<div class="item-{{ i }}">
  <p>{{ text|upper }}</p>
  <span>{{ i * 2 }}</span>
</div>
{%- endfor -%}`

	data := map[string]interface{}{"text": "benchmark test"}

	tmpl, err := env.FromString(template)
	if err != nil {
		b.Fatalf("Failed to parse template: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := tmpl.Render(miya.NewContextFrom(data))
		if err != nil {
			b.Fatalf("Failed to render template: %v", err)
		}
	}
}

// Thread Safety Benchmarks
func BenchmarkThreadSafety(b *testing.B) {
	env := miya.NewEnvironment()
	template := "Thread safe {{ id }} - {{ value }}"

	tmpl, err := env.FromString(template)
	if err != nil {
		b.Fatalf("Failed to parse template: %v", err)
	}

	var wg sync.WaitGroup
	numGoroutines := 10

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(numGoroutines)

		for j := 0; j < numGoroutines; j++ {
			go func(id, iter int) {
				defer wg.Done()

				data := map[string]interface{}{
					"id":    id,
					"value": iter,
				}

				_, err := tmpl.Render(miya.NewContextFrom(data))
				if err != nil {
					b.Errorf("Failed to render template: %v", err)
				}
			}(j, i)
		}

		wg.Wait()
	}
}
