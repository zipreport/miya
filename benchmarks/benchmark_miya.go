package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/zipreport/miya"
)

func benchmarkSimpleTemplate() map[string]float64 {
	templateStr := "Hello {{ name }}!"
	env := miya.NewEnvironment()

	// Warmup
	tmpl, _ := env.FromString(templateStr)
	ctx := miya.NewContextFrom(map[string]interface{}{"name": "World"})
	for i := 0; i < 100; i++ {
		tmpl.Render(ctx)
	}

	// Benchmark template compilation (cold start)
	compileTimes := []float64{}
	for i := 0; i < 100; i++ {
		start := time.Now()
		env.FromString(templateStr)
		compileTimes = append(compileTimes, float64(time.Since(start).Nanoseconds())/1000.0) // Convert to microseconds
	}

	// Benchmark cached rendering
	tmpl, _ = env.FromString(templateStr)
	renderTimes := []float64{}
	for i := 0; i < 10000; i++ {
		start := time.Now()
		tmpl.Render(ctx)
		renderTimes = append(renderTimes, float64(time.Since(start).Nanoseconds())/1000.0) // Convert to microseconds
	}

	return map[string]float64{
		"compile_avg": mean(compileTimes),
		"compile_med": median(compileTimes),
		"render_avg":  mean(renderTimes),
		"render_med":  median(renderTimes),
	}
}

func benchmarkLoopTemplate() map[string]float64 {
	templateStr := "{% for item in items %}{{ item }} {% endfor %}"
	env := miya.NewEnvironment()

	// Warmup
	tmpl, _ := env.FromString(templateStr)
	ctx := miya.NewContextFrom(map[string]interface{}{
		"items": []string{"apple", "banana", "cherry", "date", "elderberry"},
	})
	for i := 0; i < 100; i++ {
		tmpl.Render(ctx)
	}

	// Benchmark rendering
	tmpl, _ = env.FromString(templateStr)
	renderTimes := []float64{}
	for i := 0; i < 10000; i++ {
		start := time.Now()
		tmpl.Render(ctx)
		renderTimes = append(renderTimes, float64(time.Since(start).Nanoseconds())/1000.0) // Convert to microseconds
	}

	return map[string]float64{
		"render_avg": mean(renderTimes),
		"render_med": median(renderTimes),
	}
}

func benchmarkComplexTemplate() map[string]float64 {
	templateStr := `
{% for user in users %}
  <div class="user">
    <h2>{{ user.name|upper }}</h2>
    <p>Email: {{ user.email }}</p>
    <p>Age: {{ user.age }}</p>
    {% if user.active %}Active{% else %}Inactive{% endif %}
  </div>
{% endfor %}
`
	env := miya.NewEnvironment()

	// Warmup
	tmpl, _ := env.FromString(templateStr)
	ctx := miya.NewContextFrom(map[string]interface{}{
		"users": []map[string]interface{}{
			{"name": "Alice", "email": "alice@example.com", "age": 30, "active": true},
			{"name": "Bob", "email": "bob@example.com", "age": 25, "active": false},
			{"name": "Charlie", "email": "charlie@example.com", "age": 35, "active": true},
		},
	})
	for i := 0; i < 100; i++ {
		tmpl.Render(ctx)
	}

	// Benchmark rendering
	tmpl, _ = env.FromString(templateStr)
	renderTimes := []float64{}
	for i := 0; i < 10000; i++ {
		start := time.Now()
		tmpl.Render(ctx)
		renderTimes = append(renderTimes, float64(time.Since(start).Nanoseconds())/1000.0) // Convert to microseconds
	}

	return map[string]float64{
		"render_avg": mean(renderTimes),
		"render_med": median(renderTimes),
	}
}

func mean(values []float64) float64 {
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func median(values []float64) float64 {
	sorted := make([]float64, len(values))
	copy(sorted, values)
	// Simple bubble sort for median calculation
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	if len(sorted)%2 == 0 {
		return (sorted[len(sorted)/2-1] + sorted[len(sorted)/2]) / 2
	}
	return sorted[len(sorted)/2]
}

func main() {
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("Miya Engine (Go) Performance Benchmark")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()

	fmt.Println("1. Simple Template (Hello {{ name }}!)")
	fmt.Println(strings.Repeat("-", 60))
	simpleResults := benchmarkSimpleTemplate()
	fmt.Printf("   Compilation (cold start):\n")
	fmt.Printf("     Average:  %.2f μs\n", simpleResults["compile_avg"])
	fmt.Printf("     Median:   %.2f μs\n", simpleResults["compile_med"])
	fmt.Printf("   Rendering (cached template):\n")
	fmt.Printf("     Average:  %.2f μs\n", simpleResults["render_avg"])
	fmt.Printf("     Median:   %.2f μs\n", simpleResults["render_med"])
	fmt.Println()

	fmt.Println("2. Loop Template ({% for item in items %})")
	fmt.Println(strings.Repeat("-", 60))
	loopResults := benchmarkLoopTemplate()
	fmt.Printf("   Rendering (5 items):\n")
	fmt.Printf("     Average:  %.2f μs\n", loopResults["render_avg"])
	fmt.Printf("     Median:   %.2f μs\n", loopResults["render_med"])
	fmt.Println()

	fmt.Println("3. Complex Template (nested loops, filters, conditionals)")
	fmt.Println(strings.Repeat("-", 60))
	complexResults := benchmarkComplexTemplate()
	fmt.Printf("   Rendering (3 users):\n")
	fmt.Printf("     Average:  %.2f μs\n", complexResults["render_avg"])
	fmt.Printf("     Median:   %.2f μs\n", complexResults["render_med"])
	fmt.Println()

	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("Summary")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Simple template rendering:  %.2f μs\n", simpleResults["render_avg"])
	fmt.Printf("Loop template rendering:    %.2f μs\n", loopResults["render_avg"])
	fmt.Printf("Complex template rendering: %.2f μs\n", complexResults["render_avg"])
}
