package miya_test

import (
	"fmt"
	miya "github.com/zipreport/miya"
	"strings"
	"testing"

	"github.com/zipreport/miya/extensions"
	"github.com/zipreport/miya/parser"
	"github.com/zipreport/miya/runtime"
)

// Type alias to make function signatures cleaner
type ExtensionNode = parser.ExtensionNode

// Create a simple test extension
type SimpleTestExtension struct {
	*extensions.BaseExtension
}

func NewSimpleTestExtension() *SimpleTestExtension {
	return &SimpleTestExtension{
		BaseExtension: extensions.NewBaseExtension("simple", []string{"hello", "greet"}),
	}
}

func (ste *SimpleTestExtension) ParseTag(tagName string, parser extensions.ExtensionParser) (parser.Node, error) {
	startToken := parser.Current()
	node := parser.NewExtensionNode("simple", tagName, startToken.Line, startToken.Column)

	switch tagName {
	case "hello":
		// {% hello %}
		node.SetEvaluateFunc(func(node *ExtensionNode, ctx interface{}) (interface{}, error) {
			return "Hello, World!", nil
		})

	case "greet":
		// {% greet name %}
		nameArg, err := parser.ParseExpression()
		if err != nil {
			return nil, err
		}
		node.AddArgument(nameArg)

		node.SetEvaluateFunc(func(node *ExtensionNode, ctx interface{}) (interface{}, error) {
			if len(node.Arguments) == 0 {
				return "Hello, Anonymous!", nil
			}

			evaluator := runtime.NewEvaluator()
			runtimeCtx, ok := ctx.(runtime.Context)
			if !ok {
				return nil, fmt.Errorf("invalid context type")
			}
			nameResult, err := evaluator.EvalNode(node.Arguments[0], runtimeCtx)
			if err != nil {
				return nil, err
			}

			return fmt.Sprintf("Hello, %v!", nameResult), nil
		})
	}

	return node, parser.ExpectBlockEnd()
}

func TestBasicExtensionSystem(t *testing.T) {
	// Create extension registry
	registry := extensions.NewRegistry()

	// Register our test extension
	testExt := NewSimpleTestExtension()
	err := registry.Register(testExt)
	if err != nil {
		t.Fatalf("Failed to register extension: %v", err)
	}

	// Test registry functionality
	if !registry.IsCustomTag("hello") {
		t.Error("Expected 'hello' to be recognized as custom tag")
	}

	if !registry.IsCustomTag("greet") {
		t.Error("Expected 'greet' to be recognized as custom tag")
	}

	if registry.IsCustomTag("nonexistent") {
		t.Error("Expected 'nonexistent' to not be recognized as custom tag")
	}

	// Test extension retrieval
	ext, ok := registry.GetExtensionForTag("hello")
	if !ok || ext.Name() != "simple" {
		t.Error("Failed to retrieve extension for 'hello' tag")
	}
}

func TestSimpleExtensionRendering(t *testing.T) {
	// Create a minimal template with custom tag
	template := `{% hello %}`

	// Create tokens
	tokens, err := extensions.CreateTokensFromString(template)
	if err != nil {
		t.Fatalf("Failed to create tokens: %v", err)
	}

	// Create registry and register extension
	registry := extensions.NewRegistry()
	testExt := NewSimpleTestExtension()
	err = registry.Register(testExt)
	if err != nil {
		t.Fatalf("Failed to register extension: %v", err)
	}

	// Parse with extension support
	extParser := extensions.NewExtensionAwareParser(tokens, registry)
	ast, err := extParser.ParseTopLevelPublic()
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	// Create evaluator and context
	evaluator := runtime.NewEvaluator()
	ctx := miya.NewContext()

	// Evaluate
	result, err := evaluator.EvalNode(ast, miya.NewTemplateContextAdapter(ctx, miya.NewEnvironment()))
	if err != nil {
		t.Fatalf("Failed to evaluate template: %v", err)
	}

	expected := "Hello, World!"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestExtensionWithArguments(t *testing.T) {
	// Create template with arguments
	template := `{% greet "Alice" %}`

	// Create tokens
	tokens, err := extensions.CreateTokensFromString(template)
	if err != nil {
		t.Fatalf("Failed to create tokens: %v", err)
	}

	// Create registry and register extension
	registry := extensions.NewRegistry()
	testExt := NewSimpleTestExtension()
	err = registry.Register(testExt)
	if err != nil {
		t.Fatalf("Failed to register extension: %v", err)
	}

	// Parse with extension support
	extParser := extensions.NewExtensionAwareParser(tokens, registry)
	ast, err := extParser.ParseTopLevelPublic()
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	// Create evaluator and context
	evaluator := runtime.NewEvaluator()
	ctx := miya.NewContext()

	// Evaluate
	result, err := evaluator.EvalNode(ast, miya.NewTemplateContextAdapter(ctx, miya.NewEnvironment()))
	if err != nil {
		t.Fatalf("Failed to evaluate template: %v", err)
	}

	expected := "Hello, Alice!"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestMultipleExtensions(t *testing.T) {
	// Create registry
	registry := extensions.NewRegistry()

	// Register multiple extensions
	simpleExt := NewSimpleTestExtension()
	err := registry.Register(simpleExt)
	if err != nil {
		t.Fatalf("Failed to register simple extension: %v", err)
	}

	timestampExt := extensions.NewSimpleTimestampExtension()
	err = registry.Register(timestampExt)
	if err != nil {
		t.Fatalf("Failed to register timestamp extension: %v", err)
	}

	// Verify all tags are registered
	expectedTags := []string{"hello", "greet", "now", "timestamp"}
	for _, tag := range expectedTags {
		if !registry.IsCustomTag(tag) {
			t.Errorf("Expected tag '%s' to be registered", tag)
		}
	}

	// Test that we can get the correct extension for each tag
	for _, tag := range []string{"hello", "greet"} {
		ext, ok := registry.GetExtensionForTag(tag)
		if !ok || ext.Name() != "simple" {
			t.Errorf("Failed to get correct extension for tag '%s'", tag)
		}
	}

	for _, tag := range []string{"now", "timestamp"} {
		ext, ok := registry.GetExtensionForTag(tag)
		if !ok || ext.Name() != "timestamp" {
			t.Errorf("Failed to get correct extension for tag '%s'", tag)
		}
	}
}

func TestExtensionConflicts(t *testing.T) {
	registry := extensions.NewRegistry()

	// Register first extension
	ext1 := NewSimpleTestExtension()
	err := registry.Register(ext1)
	if err != nil {
		t.Fatalf("Failed to register first extension: %v", err)
	}

	// Try to register another extension with the same name
	ext2 := NewSimpleTestExtension()
	err = registry.Register(ext2)
	if err == nil {
		t.Error("Expected error when registering extension with duplicate name")
	}

	// Create a conflicting extension with the same tag
	conflictingExt := NewSimpleTestExtension()
	// Change the name but keep the same tags to create a tag conflict
	conflictingExt.BaseExtension = extensions.NewBaseExtension("conflict", []string{"hello"})

	err = registry.Register(conflictingExt)
	if err == nil {
		t.Error("Expected error when registering extension with conflicting tag")
	}
}

// Test that shows how to integrate extensions into environment
func TestExtensionIntegration(t *testing.T) {
	// This is a conceptual test showing how extensions would integrate
	// In a full implementation, Environment would have an extension registry

	registry := extensions.NewRegistry()
	testExt := NewSimpleTestExtension()
	err := registry.Register(testExt)
	if err != nil {
		t.Fatalf("Failed to register extension: %v", err)
	}

	// In the future, Environment would use the registry during parsing
	// For now, we just verify the extension system components work

	if len(registry.GetAllExtensions()) != 1 {
		t.Error("Expected exactly one registered extension")
	}

	allExts := registry.GetAllExtensions()
	if allExts[0].Name() != "simple" {
		t.Error("Expected extension name to be 'simple'")
	}

	tags := allExts[0].Tags()
	expectedTags := []string{"hello", "greet"}
	if len(tags) != len(expectedTags) {
		t.Error("Extension should have exactly 2 tags")
	}

	for i, tag := range tags {
		if tag != expectedTags[i] {
			t.Errorf("Expected tag %d to be %s, got %s", i, expectedTags[i], tag)
		}
	}
}

// Benchmark extension system overhead
func TestBlockExtensions(t *testing.T) {
	// Test highlight extension
	template := `{% highlight "python" %}
def hello():
    print("world")
{% endhighlight %}`

	// Create tokens
	tokens, err := extensions.CreateTokensFromString(template)
	if err != nil {
		t.Fatalf("Failed to create tokens: %v", err)
	}

	// Create registry and register extension
	registry := extensions.NewRegistry()
	highlightExt := extensions.NewHighlightExtension()
	err = registry.Register(highlightExt)
	if err != nil {
		t.Fatalf("Failed to register highlight extension: %v", err)
	}

	// Parse with extension support
	extParser := extensions.NewExtensionAwareParser(tokens, registry)
	ast, err := extParser.ParseTopLevelPublic()
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	// Create evaluator and context
	evaluator := runtime.NewEvaluator()
	ctx := miya.NewContext()

	// Evaluate
	result, err := evaluator.EvalNode(ast, miya.NewTemplateContextAdapter(ctx, miya.NewEnvironment()))
	if err != nil {
		t.Fatalf("Failed to evaluate template: %v", err)
	}

	expected := `<div class="highlight-python"><pre>
def hello():
    print("world")
</pre></div>`
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestCacheExtension(t *testing.T) {
	// Test cache extension
	template := `{% cache 300 "my-key" %}
Expensive computation: {{ 2 + 2 }}
{% endcache %}`

	// Create tokens
	tokens, err := extensions.CreateTokensFromString(template)
	if err != nil {
		t.Fatalf("Failed to create tokens: %v", err)
	}

	// Create registry and register extension
	registry := extensions.NewRegistry()
	cacheExt := extensions.NewCacheExtension()
	err = registry.Register(cacheExt)
	if err != nil {
		t.Fatalf("Failed to register cache extension: %v", err)
	}

	// Parse with extension support
	extParser := extensions.NewExtensionAwareParser(tokens, registry)
	ast, err := extParser.ParseTopLevelPublic()
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	// Create evaluator and context
	evaluator := runtime.NewEvaluator()
	ctx := miya.NewContext()

	// Evaluate
	result, err := evaluator.EvalNode(ast, miya.NewTemplateContextAdapter(ctx, miya.NewEnvironment()))
	if err != nil {
		t.Fatalf("Failed to evaluate template: %v", err)
	}

	expected := "\nExpensive computation: 4\n"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestBlockExtensionProperties(t *testing.T) {
	highlightExt := extensions.NewHighlightExtension()

	// Test IsBlockExtension
	if !highlightExt.IsBlockExtension("highlight") {
		t.Error("Expected 'highlight' to be a block extension")
	}

	if highlightExt.IsBlockExtension("nonexistent") {
		t.Error("Expected 'nonexistent' to not be a block extension")
	}

	// Test GetEndTag
	endTag := highlightExt.GetEndTag("highlight")
	if endTag != "endhighlight" {
		t.Errorf("Expected end tag 'endhighlight', got %q", endTag)
	}

	endTag = highlightExt.GetEndTag("nonexistent")
	if endTag != "" {
		t.Errorf("Expected empty end tag for nonexistent tag, got %q", endTag)
	}
}

func TestEnvironmentExtensionIntegration(t *testing.T) {
	// Create environment
	env := miya.NewEnvironment()

	// Register highlight extension
	highlightExt := extensions.NewHighlightExtension()
	err := env.AddExtension(highlightExt)
	if err != nil {
		t.Fatalf("Failed to register extension: %v", err)
	}

	// Test that environment recognizes custom tags
	if !env.IsCustomTag("highlight") {
		t.Error("Expected environment to recognize 'highlight' as custom tag")
	}

	if env.IsCustomTag("nonexistent") {
		t.Error("Expected environment to not recognize 'nonexistent' as custom tag")
	}

	// Test rendering with environment
	template := `{% highlight "go" %}
func main() {
    fmt.Println("Hello, World!")
}
{% endhighlight %}`

	result, err := env.RenderString(template, miya.NewContext())
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}

	expected := `<div class="highlight-go"><pre>
func main() {
    fmt.Println("Hello, World!")
}
</pre></div>`

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestEnvironmentStandardTagsWithExtensions(t *testing.T) {
	// Create environment with extensions
	env := miya.NewEnvironment()
	highlightExt := extensions.NewHighlightExtension()
	err := env.AddExtension(highlightExt)
	if err != nil {
		t.Fatalf("Failed to register extension: %v", err)
	}

	// Test that standard tags still work
	template := `{% if true %}Standard tag works{% endif %}`

	result, err := env.RenderString(template, miya.NewContext())
	if err != nil {
		t.Fatalf("Failed to render template with standard tags: %v", err)
	}

	expected := "Standard tag works"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestEnvironmentMixedExtensionsAndStandardTags(t *testing.T) {
	// Create environment with extensions
	env := miya.NewEnvironment()
	highlightExt := extensions.NewHighlightExtension()
	timestampExt := extensions.NewSimpleTimestampExtension()

	err := env.AddExtension(highlightExt)
	if err != nil {
		t.Fatalf("Failed to register highlight extension: %v", err)
	}

	err = env.AddExtension(timestampExt)
	if err != nil {
		t.Fatalf("Failed to register timestamp extension: %v", err)
	}

	// Test template with mixed standard and custom tags
	template := `{% if true %}Standard if tag works.{% endif %} {% now %}`

	result, err := env.RenderString(template, miya.NewContext())
	if err != nil {
		t.Fatalf("Failed to render mixed template: %v", err)
	}

	// Should contain both standard and extension elements
	if !strings.Contains(result, "Standard if tag works") {
		t.Error("Expected result to contain standard if tag output")
	}

	// The {% now %} tag should render a timestamp (should contain year)
	if !strings.Contains(result, "202") { // Should contain part of current year
		t.Errorf("Expected result to contain timestamp, got: %q", result)
	}
}

func TestExtensionConfiguration(t *testing.T) {
	// Test extension configuration
	ext := extensions.NewHighlightExtension()

	// Test initial empty config
	config := ext.GetConfig()
	if len(config) != 0 {
		t.Error("Expected empty initial configuration")
	}

	// Test setting configuration
	newConfig := map[string]interface{}{
		"theme":        "dark",
		"line_numbers": true,
		"max_lines":    100,
	}

	err := ext.Configure(newConfig)
	if err != nil {
		t.Fatalf("Failed to configure extension: %v", err)
	}

	// Test getting configuration
	config = ext.GetConfig()
	if len(config) != 3 {
		t.Errorf("Expected 3 config items, got %d", len(config))
	}

	if config["theme"] != "dark" {
		t.Errorf("Expected theme 'dark', got %v", config["theme"])
	}

	if config["line_numbers"] != true {
		t.Errorf("Expected line_numbers true, got %v", config["line_numbers"])
	}

	if config["max_lines"] != 100 {
		t.Errorf("Expected max_lines 100, got %v", config["max_lines"])
	}
}

func BenchmarkExtensionRegistry(b *testing.B) {
	registry := extensions.NewRegistry()
	testExt := NewSimpleTestExtension()
	timestampExt := extensions.NewSimpleTimestampExtension()

	registry.Register(testExt)
	registry.Register(timestampExt)

	b.ResetTimer()

	b.Run("tag_lookup", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = registry.GetExtensionForTag("hello")
			_, _ = registry.GetExtensionForTag("now")
			_, _ = registry.GetExtensionForTag("nonexistent")
		}
	})

	b.Run("is_custom_tag", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = registry.IsCustomTag("hello")
			_ = registry.IsCustomTag("now")
			_ = registry.IsCustomTag("nonexistent")
		}
	})
}
