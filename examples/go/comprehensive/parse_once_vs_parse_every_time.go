//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"os"
	"time"

	miya "github.com/zipreport/miya"
	"github.com/zipreport/miya/loader"
	"github.com/zipreport/miya/parser"
)

func main() {
	fmt.Println("🔬 Miya Engine: Parse-Once vs Parse-Every-Time Comparison")
	fmt.Println("============================================================")

	// Template to test with
	templateSource := `<!DOCTYPE html>
<html>
<head><title>{{ title }}</title></head>
<body>
	<h1>{{ title }}</h1>
	<p>Hello {{ user.name }}!</p>
	<ul>
	{% for item in items %}
		<li>{{ item }} ({{ loop.index }})</li>
	{% endfor %}
	</ul>
	{% filter upper %}
	This text will be uppercased: {{ message }}
	{% endfilter %}
</body>
</html>`

	// Sample data
	createData := func() miya.Context {
		ctx := miya.NewContext()
		ctx.Set("title", "Performance Test")
		ctx.Set("user", map[string]interface{}{"name": "Alice"})
		ctx.Set("items", []string{"apple", "banana", "cherry"})
		ctx.Set("message", "Hello World")
		return ctx
	}

	fmt.Println()

	// Test 1: Parse-Every-Time (RenderString approach)
	fmt.Println("🔄 TEST 1: Parse-Every-Time Approach (RenderString)")
	testParseEveryTime(templateSource, createData)

	fmt.Println()

	// Test 2: Parse-Once with Manual Template Object
	fmt.Println("⚡ TEST 2: Parse-Once with Manual Template Caching")
	testParseOnceManual(templateSource, createData)

	fmt.Println()

	// Test 3: Parse-Once with CachedEnvironment
	fmt.Println("🚀 TEST 3: Parse-Once with CachedEnvironment")
	testParseOnceCached(templateSource, createData)

	fmt.Println()

	// Test 4: Parse-Once with FileSystem Loader (like Python)
	fmt.Println("📁 TEST 4: Parse-Once with FileSystem Loader (Python-like)")
	testParseOnceFileSystem(createData)
}

func testParseEveryTime(templateSource string, createData func() miya.Context) {
	env := miya.NewEnvironment()
	numRenders := 5

	fmt.Printf("   Rendering template %d times with RenderString()...\n", numRenders)

	var totalRenderTime time.Duration
	var results []string

	for i := 0; i < numRenders; i++ {
		ctx := createData()

		renderStart := time.Now()
		result, err := env.RenderString(templateSource, ctx)
		renderTime := time.Since(renderStart)
		totalRenderTime += renderTime

		if err != nil {
			fmt.Printf("   ❌ Render %d failed: %v\n", i+1, err)
			return
		}

		results = append(results, result)
		fmt.Printf("   🔄 Render %d: %.2fms (includes parsing)\n", i+1, float64(renderTime.Nanoseconds())/1e6)
	}

	avgTime := float64(totalRenderTime.Nanoseconds()) / float64(numRenders) / 1e6
	fmt.Printf("   📊 Average time: %.2fms per render\n", avgTime)
	fmt.Printf("   📏 Output size: %d bytes\n", len(results[0]))
}

func testParseOnceManual(templateSource string, createData func() miya.Context) {
	env := miya.NewEnvironment()
	numRenders := 5

	// Parse once
	fmt.Println("   🔧 Parsing template once with FromString()...")
	parseStart := time.Now()
	template, err := env.FromString(templateSource)
	parseTime := time.Since(parseStart)

	if err != nil {
		fmt.Printf("   ❌ Parse failed: %v\n", err)
		return
	}

	fmt.Printf("   ✅ Template parsed in %.2fms\n", float64(parseTime.Nanoseconds())/1e6)

	// Render multiple times
	fmt.Printf("   Rendering parsed template %d times...\n", numRenders)

	var totalRenderTime time.Duration
	var results []string

	for i := 0; i < numRenders; i++ {
		ctx := createData()

		renderStart := time.Now()
		result, err := template.Render(ctx)
		renderTime := time.Since(renderStart)
		totalRenderTime += renderTime

		if err != nil {
			fmt.Printf("   ❌ Render %d failed: %v\n", i+1, err)
			return
		}

		results = append(results, result)
		fmt.Printf("   ⚡ Render %d: %.2fms (no parsing overhead)\n", i+1, float64(renderTime.Nanoseconds())/1e6)
	}

	avgRenderTime := float64(totalRenderTime.Nanoseconds()) / float64(numRenders) / 1e6
	totalWithParsing := float64(parseTime.Nanoseconds())/1e6 + float64(totalRenderTime.Nanoseconds())/1e6

	fmt.Printf("   📊 Parse time: %.2fms (one-time cost)\n", float64(parseTime.Nanoseconds())/1e6)
	fmt.Printf("   📊 Average render time: %.2fms per render\n", avgRenderTime)
	fmt.Printf("   📊 Total time: %.2fms (%.2fms amortized per render)\n", totalWithParsing, totalWithParsing/float64(numRenders))
	fmt.Printf("   📏 Output size: %d bytes\n", len(results[0]))
}

func testParseOnceCached(templateSource string, createData func() miya.Context) {
	cachedEnv := miya.NewCachedEnvironment()
	numRenders := 5

	fmt.Printf("   Using CachedEnvironment for %d renders...\n", numRenders)

	var totalTime time.Duration
	var firstRenderTime time.Duration
	var subsequentRenderTime time.Duration
	var results []string

	for i := 0; i < numRenders; i++ {
		ctx := createData()

		start := time.Now()

		// This will parse on first call, use cache on subsequent calls
		template, err := cachedEnv.FromStringCached(templateSource)
		if err != nil {
			fmt.Printf("   ❌ Template creation failed: %v\n", err)
			return
		}

		result, err := template.Render(ctx)
		renderTime := time.Since(start)
		totalTime += renderTime

		if err != nil {
			fmt.Printf("   ❌ Render %d failed: %v\n", i+1, err)
			return
		}

		results = append(results, result)

		if i == 0 {
			firstRenderTime = renderTime
			fmt.Printf("   🔧 First render: %.2fms (includes parsing + caching)\n", float64(renderTime.Nanoseconds())/1e6)
		} else {
			subsequentRenderTime += renderTime
			fmt.Printf("   🚀 Render %d: %.2fms (cached template)\n", i+1, float64(renderTime.Nanoseconds())/1e6)
		}
	}

	avgSubsequentTime := float64(subsequentRenderTime.Nanoseconds()) / float64(numRenders-1) / 1e6
	avgTotalTime := float64(totalTime.Nanoseconds()) / float64(numRenders) / 1e6

	fmt.Printf("   📊 First render (with parsing): %.2fms\n", float64(firstRenderTime.Nanoseconds())/1e6)
	fmt.Printf("   📊 Average subsequent renders: %.2fms\n", avgSubsequentTime)
	fmt.Printf("   📊 Average overall: %.2fms per render\n", avgTotalTime)
	fmt.Printf("   📏 Output size: %d bytes\n", len(results[0]))
}

// SimpleTemplateParser implements loader.TemplateParser interface
type SimpleTemplateParser struct {
	env *miya.Environment
}

func NewSimpleTemplateParser(env *miya.Environment) *SimpleTemplateParser {
	return &SimpleTemplateParser{env: env}
}

func (stp *SimpleTemplateParser) ParseTemplate(name, content string) (*parser.TemplateNode, error) {
	// Use the environment's compile method to parse the template
	template, err := stp.env.FromString(content)
	if err != nil {
		return nil, err
	}

	// Extract the AST from the template
	if templateNode, ok := template.AST().(*parser.TemplateNode); ok {
		templateNode.Name = name
		return templateNode, nil
	}

	return nil, fmt.Errorf("failed to extract template node from parsed template")
}

func testParseOnceFileSystem(createData func() miya.Context) {
	// Create a temporary template file
	templateContent := `<!DOCTYPE html>
<html>
<head><title>{{ title }} - FileSystem Test</title></head>
<body>
	<h1>{{ title }} (Loaded from File)</h1>
	<p>Hello {{ user.name }}! This template was loaded from filesystem.</p>
	<ul>
	{% for item in items %}
		<li>{{ item }} - File-based template</li>
	{% endfor %}
	</ul>
	{% filter upper %}
	File system template with filter: {{ message }}
	{% endfilter %}
</body>
</html>`

	// Write template to file
	err := writeFile("test_template.html", templateContent)

	if err != nil {
		fmt.Printf("   ❌ Failed to create template file: %v\n", err)
		fmt.Println("   ℹ️  Skipping filesystem test")
		return
	}

	defer removeFile("test_template.html") // Clean up

	// Create environment with filesystem loader
	env := miya.NewEnvironment()
	templateParser := NewSimpleTemplateParser(env)
	fsLoader := loader.NewFileSystemLoader([]string{"."}, templateParser)
	env.SetLoader(fsLoader)

	numRenders := 5
	fmt.Printf("   Using FileSystem Loader for %d renders...\n", numRenders)

	var totalTime time.Duration
	var firstLoadTime time.Duration
	var results []string

	for i := 0; i < numRenders; i++ {
		ctx := createData()

		start := time.Now()

		// This uses environment's built-in cache (like Python Jinja2)
		template, err := env.GetTemplate("test_template.html")
		if err != nil {
			fmt.Printf("   ❌ Template load failed: %v\n", err)
			return
		}

		result, err := template.Render(ctx)
		totalRenderTime := time.Since(start)
		totalTime += totalRenderTime

		if err != nil {
			fmt.Printf("   ❌ Render %d failed: %v\n", i+1, err)
			return
		}

		results = append(results, result)

		if i == 0 {
			firstLoadTime = totalRenderTime
			fmt.Printf("   📁 First load: %.2fms (includes file I/O + parsing)\n", float64(totalRenderTime.Nanoseconds())/1e6)
		} else {
			fmt.Printf("   🚀 Render %d: %.2fms (cached template)\n", i+1, float64(totalRenderTime.Nanoseconds())/1e6)
		}
	}

	avgTime := float64(totalTime.Nanoseconds()) / float64(numRenders) / 1e6

	fmt.Printf("   📊 First load (with file I/O + parsing): %.2fms\n", float64(firstLoadTime.Nanoseconds())/1e6)
	fmt.Printf("   📊 Average time: %.2fms per render\n", avgTime)
	fmt.Printf("   📏 Output size: %d bytes\n", len(results[0]))
}

// Helper functions for file operations
func writeFile(name, content string) error {
	return os.WriteFile(name, []byte(content), 0644)
}

func removeFile(name string) error {
	return os.Remove(name)
}
