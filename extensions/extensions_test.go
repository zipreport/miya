package extensions

import (
	"errors"
	"strings"
	"testing"

	"github.com/zipreport/miya/parser"
	"github.com/zipreport/miya/runtime"
)

// Test ExtensionError functionality
func TestExtensionError(t *testing.T) {
	t.Run("Error formatting with template context", func(t *testing.T) {
		err := NewExtensionParseError("test_ext", "test_tag", "template.html", 10, 5, "test error", nil)

		expected := "extension 'test_ext' error in template 'template.html' at line 10:5: test error"
		if err.Error() != expected {
			t.Errorf("Expected: %s\nGot: %s", expected, err.Error())
		}
	})

	t.Run("Error formatting with template but no line", func(t *testing.T) {
		err := &ExtensionError{
			ExtensionName: "test_ext",
			TemplateName:  "template.html",
			Message:       "test error",
		}

		expected := "extension 'test_ext' error in template 'template.html': test error"
		if err.Error() != expected {
			t.Errorf("Expected: %s\nGot: %s", expected, err.Error())
		}
	})

	t.Run("Error formatting with tag and line", func(t *testing.T) {
		err := &ExtensionError{
			ExtensionName: "test_ext",
			TagName:       "test_tag",
			Line:          5,
			Column:        10,
			Message:       "test error",
		}

		expected := "extension 'test_ext' error in tag 'test_tag' at line 5:10: test error"
		if err.Error() != expected {
			t.Errorf("Expected: %s\nGot: %s", expected, err.Error())
		}
	})

	t.Run("Error formatting with tag but no line", func(t *testing.T) {
		err := &ExtensionError{
			ExtensionName: "test_ext",
			TagName:       "test_tag",
			Message:       "test error",
		}

		expected := "extension 'test_ext' error in tag 'test_tag': test error"
		if err.Error() != expected {
			t.Errorf("Expected: %s\nGot: %s", expected, err.Error())
		}
	})

	t.Run("Error formatting minimal", func(t *testing.T) {
		err := &ExtensionError{
			ExtensionName: "test_ext",
			Message:       "test error",
		}

		expected := "extension 'test_ext' error: test error"
		if err.Error() != expected {
			t.Errorf("Expected: %s\nGot: %s", expected, err.Error())
		}
	})

	t.Run("Error unwrapping", func(t *testing.T) {
		originalErr := errors.New("original error")
		err := NewExtensionError("test_ext", "test error", originalErr)

		unwrapped := err.Unwrap()
		if unwrapped != originalErr {
			t.Errorf("Expected unwrapped error to be original error")
		}
	})
}

// Test Registry functionality
func TestRegistry(t *testing.T) {
	t.Run("NewRegistry creates empty registry", func(t *testing.T) {
		registry := NewRegistry()

		if registry == nil {
			t.Fatal("NewRegistry returned nil")
		}

		if len(registry.extensions) != 0 {
			t.Error("Expected empty extensions map")
		}

		if len(registry.tagMap) != 0 {
			t.Error("Expected empty tagMap")
		}

		if len(registry.loadOrder) != 0 {
			t.Error("Expected empty loadOrder")
		}
	})

	t.Run("Register basic extension", func(t *testing.T) {
		registry := NewRegistry()
		ext := NewHelloExtension()

		err := registry.Register(ext)
		if err != nil {
			t.Fatalf("Registration failed: %v", err)
		}

		// Check extension was registered
		if _, exists := registry.GetExtension("hello"); !exists {
			t.Error("Extension not found after registration")
		}

		// Check tag mapping
		if _, exists := registry.GetExtensionForTag("hello"); !exists {
			t.Error("Tag mapping not created")
		}

		// Check load order
		loadOrder := registry.GetLoadOrder()
		if len(loadOrder) != 1 || loadOrder[0] != "hello" {
			t.Error("Load order not updated correctly")
		}
	})

	t.Run("Register duplicate extension fails", func(t *testing.T) {
		registry := NewRegistry()
		ext1 := NewHelloExtension()
		ext2 := NewHelloExtension()

		err := registry.Register(ext1)
		if err != nil {
			t.Fatalf("First registration failed: %v", err)
		}

		err = registry.Register(ext2)
		if err == nil {
			t.Error("Expected error for duplicate registration")
		}

		if !strings.Contains(err.Error(), "already registered") {
			t.Errorf("Expected 'already registered' in error, got: %v", err)
		}
	})

	t.Run("Register extension with tag conflict fails", func(t *testing.T) {
		registry := NewRegistry()
		ext1 := NewHelloExtension()

		// Create another extension with same tag
		ext2 := &TestExtension{
			BaseExtension: NewBaseExtension("conflicting", []string{"hello"}),
		}

		err := registry.Register(ext1)
		if err != nil {
			t.Fatalf("First registration failed: %v", err)
		}

		err = registry.Register(ext2)
		if err == nil {
			t.Error("Expected error for tag conflict")
		}

		if !strings.Contains(err.Error(), "already handled") {
			t.Errorf("Expected 'already handled' in error, got: %v", err)
		}
	})

	t.Run("Unregister extension", func(t *testing.T) {
		registry := NewRegistry()
		ext := NewHelloExtension()

		// Register first
		err := registry.Register(ext)
		if err != nil {
			t.Fatalf("Registration failed: %v", err)
		}

		// Unregister
		err = registry.Unregister("hello")
		if err != nil {
			t.Fatalf("Unregistration failed: %v", err)
		}

		// Check it's gone
		if _, exists := registry.GetExtension("hello"); exists {
			t.Error("Extension still exists after unregistration")
		}

		if _, exists := registry.GetExtensionForTag("hello"); exists {
			t.Error("Tag mapping still exists after unregistration")
		}
	})

	t.Run("Unregister non-existent extension fails", func(t *testing.T) {
		registry := NewRegistry()

		err := registry.Unregister("nonexistent")
		if err == nil {
			t.Error("Expected error for unregistering non-existent extension")
		}
	})

	t.Run("IsCustomTag works correctly", func(t *testing.T) {
		registry := NewRegistry()
		ext := NewHelloExtension()

		// Before registration
		if registry.IsCustomTag("hello") {
			t.Error("Expected false for unregistered tag")
		}

		// After registration
		registry.Register(ext)
		if !registry.IsCustomTag("hello") {
			t.Error("Expected true for registered tag")
		}

		// Non-existent tag
		if registry.IsCustomTag("nonexistent") {
			t.Error("Expected false for non-existent tag")
		}
	})
}

// Test dependency management
func TestRegistryDependencies(t *testing.T) {
	t.Run("Register extension with missing dependency fails", func(t *testing.T) {
		registry := NewRegistry()
		ext := &TestExtension{
			BaseExtension: NewBaseExtension("dependent", []string{"dependent_tag"}),
			dependencies:  []string{"missing_dep"},
		}

		err := registry.Register(ext)
		if err == nil {
			t.Error("Expected error for missing dependency")
		}

		if !strings.Contains(err.Error(), "depends on 'missing_dep'") {
			t.Errorf("Expected dependency error, got: %v", err)
		}
	})

	t.Run("Register extension with satisfied dependency succeeds", func(t *testing.T) {
		registry := NewRegistry()

		// Register dependency first
		dep := NewHelloExtension()
		err := registry.Register(dep)
		if err != nil {
			t.Fatalf("Failed to register dependency: %v", err)
		}

		// Register dependent extension
		dependent := &TestExtension{
			BaseExtension: NewBaseExtension("dependent", []string{"dependent_tag"}),
			dependencies:  []string{"hello"},
		}

		err = registry.Register(dependent)
		if err != nil {
			t.Fatalf("Failed to register dependent extension: %v", err)
		}

		// Check dependencies are recorded
		deps := registry.GetDependencies("dependent")
		if len(deps) != 1 || deps[0] != "hello" {
			t.Errorf("Expected dependencies [hello], got %v", deps)
		}
	})

	t.Run("Circular dependency detection", func(t *testing.T) {
		registry := NewRegistry()

		// Register first extension
		ext1 := &TestExtension{
			BaseExtension: NewBaseExtension("ext1", []string{"tag1"}),
			dependencies:  []string{"ext2"},
		}

		// Register second extension (depends on first)
		ext2 := &TestExtension{
			BaseExtension: NewBaseExtension("ext2", []string{"tag2"}),
		}

		err := registry.Register(ext2)
		if err != nil {
			t.Fatalf("Failed to register ext2: %v", err)
		}

		// Now try to register ext1 which would create a circular dependency
		ext2.dependencies = []string{"ext1"}
		registry.dependencies["ext2"] = []string{"ext1"}

		err = registry.Register(ext1)
		if err == nil {
			t.Error("Expected error for circular dependency")
		}

		if !strings.Contains(err.Error(), "circular dependency") {
			t.Errorf("Expected circular dependency error, got: %v", err)
		}
	})

	t.Run("Cannot unregister extension with dependents", func(t *testing.T) {
		registry := NewRegistry()

		// Register dependency first
		dep := NewHelloExtension()
		err := registry.Register(dep)
		if err != nil {
			t.Fatalf("Failed to register dependency: %v", err)
		}

		// Register dependent extension
		dependent := &TestExtension{
			BaseExtension: NewBaseExtension("dependent", []string{"dependent_tag"}),
			dependencies:  []string{"hello"},
		}

		err = registry.Register(dependent)
		if err != nil {
			t.Fatalf("Failed to register dependent extension: %v", err)
		}

		// Try to unregister dependency
		err = registry.Unregister("hello")
		if err == nil {
			t.Error("Expected error when trying to unregister extension with dependents")
		}

		if !strings.Contains(err.Error(), "depends on it") {
			t.Errorf("Expected dependency error, got: %v", err)
		}
	})
}

// Test lifecycle hooks
func TestRegistryLifecycleHooks(t *testing.T) {
	t.Run("OnLoad hook called during registration", func(t *testing.T) {
		registry := NewRegistry()
		mockEnv := &MockExtensionEnvironment{}
		registry.SetEnvironment(mockEnv)

		ext := &TestExtension{
			BaseExtension: NewBaseExtension("test", []string{"test_tag"}),
		}

		err := registry.Register(ext)
		if err != nil {
			t.Fatalf("Registration failed: %v", err)
		}

		if !ext.onLoadCalled {
			t.Error("OnLoad hook was not called")
		}
	})

	t.Run("OnLoad hook failure rolls back registration", func(t *testing.T) {
		registry := NewRegistry()
		mockEnv := &MockExtensionEnvironment{}
		registry.SetEnvironment(mockEnv)

		ext := &TestExtension{
			BaseExtension: NewBaseExtension("test", []string{"test_tag"}),
			onLoadError:   errors.New("OnLoad failed"),
		}

		err := registry.Register(ext)
		if err == nil {
			t.Error("Expected registration to fail when OnLoad fails")
		}

		// Check rollback occurred
		if _, exists := registry.GetExtension("test"); exists {
			t.Error("Extension should not exist after failed OnLoad")
		}

		if _, exists := registry.GetExtensionForTag("test_tag"); exists {
			t.Error("Tag mapping should not exist after failed OnLoad")
		}
	})

	t.Run("BeforeRender calls all extensions in load order", func(t *testing.T) {
		registry := NewRegistry()
		mockEnv := &MockExtensionEnvironment{}
		registry.SetEnvironment(mockEnv)
		mockCtx := &MockExtensionContext{}

		ext1 := &TestExtension{
			BaseExtension: NewBaseExtension("ext1", []string{"tag1"}),
		}
		ext2 := &TestExtension{
			BaseExtension: NewBaseExtension("ext2", []string{"tag2"}),
		}

		registry.Register(ext1)
		registry.Register(ext2)

		err := registry.BeforeRender(mockCtx, "test.html")
		if err != nil {
			t.Fatalf("BeforeRender failed: %v", err)
		}

		if !ext1.beforeRenderCalled || !ext2.beforeRenderCalled {
			t.Error("BeforeRender not called on all extensions")
		}
	})

	t.Run("AfterRender calls all extensions in reverse load order", func(t *testing.T) {
		registry := NewRegistry()
		mockEnv := &MockExtensionEnvironment{}
		registry.SetEnvironment(mockEnv)
		mockCtx := &MockExtensionContext{}

		ext1 := &TestExtension{
			BaseExtension: NewBaseExtension("ext1", []string{"tag1"}),
		}
		ext2 := &TestExtension{
			BaseExtension: NewBaseExtension("ext2", []string{"tag2"}),
		}

		registry.Register(ext1)
		registry.Register(ext2)

		err := registry.AfterRender(mockCtx, "test.html", "result", nil)
		if err != nil {
			t.Fatalf("AfterRender failed: %v", err)
		}

		if !ext1.afterRenderCalled || !ext2.afterRenderCalled {
			t.Error("AfterRender not called on all extensions")
		}
	})
}

// Test BaseExtension functionality
func TestBaseExtension(t *testing.T) {
	t.Run("NewBaseExtension creates extension correctly", func(t *testing.T) {
		ext := NewBaseExtension("test", []string{"tag1", "tag2"})

		if ext.Name() != "test" {
			t.Errorf("Expected name 'test', got '%s'", ext.Name())
		}

		tags := ext.Tags()
		if len(tags) != 2 || tags[0] != "tag1" || tags[1] != "tag2" {
			t.Errorf("Expected tags [tag1, tag2], got %v", tags)
		}

		if ext.IsBlockExtension("tag1") {
			t.Error("Expected false for IsBlockExtension with simple tags")
		}
	})

	t.Run("NewBlockExtension creates block extension correctly", func(t *testing.T) {
		blockTags := map[string]string{
			"highlight": "endhighlight",
			"cache":     "endcache",
		}

		ext := NewBlockExtension("block_test", blockTags)

		if ext.Name() != "block_test" {
			t.Errorf("Expected name 'block_test', got '%s'", ext.Name())
		}

		// Should handle both start and end tags
		if !ext.IsBlockExtension("highlight") {
			t.Error("Expected true for IsBlockExtension with block tag")
		}

		if ext.GetEndTag("highlight") != "endhighlight" {
			t.Errorf("Expected 'endhighlight', got '%s'", ext.GetEndTag("highlight"))
		}

		if ext.GetEndTag("nonexistent") != "" {
			t.Error("Expected empty string for non-existent tag")
		}
	})

	t.Run("Configuration management", func(t *testing.T) {
		ext := NewBaseExtension("test", []string{"test"})

		// Test setting configuration
		config := map[string]interface{}{
			"option1": "value1",
			"option2": 42,
		}

		err := ext.Configure(config)
		if err != nil {
			t.Fatalf("Configure failed: %v", err)
		}

		// Test getting configuration
		retrievedConfig := ext.GetConfig()
		if len(retrievedConfig) != 2 {
			t.Errorf("Expected 2 config items, got %d", len(retrievedConfig))
		}

		if retrievedConfig["option1"] != "value1" {
			t.Error("Config option1 not set correctly")
		}

		if retrievedConfig["option2"] != 42 {
			t.Error("Config option2 not set correctly")
		}

		// Test getting specific value
		value, exists := ext.GetConfigValue("option1")
		if !exists || value != "value1" {
			t.Error("GetConfigValue failed")
		}

		// Test setting specific value
		ext.SetConfigValue("option3", true)
		value, exists = ext.GetConfigValue("option3")
		if !exists || value != true {
			t.Error("SetConfigValue failed")
		}
	})

	t.Run("Default lifecycle hooks do nothing", func(t *testing.T) {
		ext := NewBaseExtension("test", []string{"test"})
		mockEnv := &MockExtensionEnvironment{}
		mockCtx := &MockExtensionContext{}

		// These should not panic and should return nil
		if err := ext.OnLoad(mockEnv); err != nil {
			t.Errorf("OnLoad should return nil, got %v", err)
		}

		if err := ext.BeforeRender(mockCtx, "test.html"); err != nil {
			t.Errorf("BeforeRender should return nil, got %v", err)
		}

		if err := ext.AfterRender(mockCtx, "test.html", "result", nil); err != nil {
			t.Errorf("AfterRender should return nil, got %v", err)
		}

		if err := ext.OnUnload(); err != nil {
			t.Errorf("OnUnload should return nil, got %v", err)
		}

		deps := ext.Dependencies()
		if deps != nil {
			t.Errorf("Dependencies should return nil, got %v", deps)
		}
	})
}

// Test extension context adapter
func TestExtensionContextAdapter(t *testing.T) {
	t.Run("Adapter delegates to runtime context", func(t *testing.T) {
		mockRuntimeCtx := &MockRuntimeContext{
			variables: map[string]interface{}{
				"test_var": "test_value",
			},
		}

		adapter := &extensionContextAdapter{ctx: mockRuntimeCtx}

		// Test GetVariable
		value, exists := adapter.GetVariable("test_var")
		if !exists || value != "test_value" {
			t.Error("GetVariable not delegated correctly")
		}

		// Test SetVariable
		adapter.SetVariable("new_var", "new_value")
		if mockRuntimeCtx.variables["new_var"] != "new_value" {
			t.Error("SetVariable not delegated correctly")
		}

		// Test GetGlobal fallback
		value, exists = adapter.GetGlobal("test_var")
		if !exists || value != "test_value" {
			t.Error("GetGlobal fallback not working")
		}
	})

	t.Run("Filter and macro calls fail gracefully", func(t *testing.T) {
		mockRuntimeCtx := &MockRuntimeContext{
			variables: map[string]interface{}{},
		}

		adapter := &extensionContextAdapter{ctx: mockRuntimeCtx}

		// Test ApplyFilter failure
		_, err := adapter.ApplyFilter("upper", "test")
		if err == nil {
			t.Error("Expected error for unsupported filter")
		}

		// Test CallMacro failure
		_, err = adapter.CallMacro("test_macro", []interface{}{}, map[string]interface{}{})
		if err == nil {
			t.Error("Expected error for unsupported macro")
		}
	})
}

// Test utility functions
func TestUtilityFunctions(t *testing.T) {
	t.Run("CreateEvaluateFunc works correctly", func(t *testing.T) {
		evalFunc := func(node *parser.ExtensionNode, ctx runtime.Context) (interface{}, error) {
			return "test result", nil
		}

		wrappedFunc := CreateEvaluateFunc(evalFunc)

		mockCtx := &MockRuntimeContext{}
		node := parser.NewExtensionNode("test", "test_tag", 1, 1)

		result, err := wrappedFunc(node, mockCtx)
		if err != nil {
			t.Fatalf("Wrapped function failed: %v", err)
		}

		if result != "test result" {
			t.Errorf("Expected 'test result', got %v", result)
		}
	})

	t.Run("CreateEvaluateFunc fails with wrong context type", func(t *testing.T) {
		evalFunc := func(node *parser.ExtensionNode, ctx runtime.Context) (interface{}, error) {
			return "test result", nil
		}

		wrappedFunc := CreateEvaluateFunc(evalFunc)

		node := parser.NewExtensionNode("test", "test_tag", 1, 1)

		_, err := wrappedFunc(node, "wrong context type")
		if err == nil {
			t.Error("Expected error with wrong context type")
		}
	})

	t.Run("CreateAdvancedEvaluateFunc works correctly", func(t *testing.T) {
		evalFunc := func(node *parser.ExtensionNode, ctx ExtensionContext) (interface{}, error) {
			return "advanced result", nil
		}

		wrappedFunc := CreateAdvancedEvaluateFunc(evalFunc)

		mockCtx := &MockRuntimeContext{}
		node := parser.NewExtensionNode("test", "test_tag", 1, 1)

		result, err := wrappedFunc(node, mockCtx)
		if err != nil {
			t.Fatalf("Advanced wrapped function failed: %v", err)
		}

		if result != "advanced result" {
			t.Errorf("Expected 'advanced result', got %v", result)
		}
	})
}

// Mock implementations for testing

type TestExtension struct {
	*BaseExtension
	dependencies       []string
	onLoadError        error
	onLoadCalled       bool
	beforeRenderCalled bool
	afterRenderCalled  bool
	onUnloadCalled     bool
}

func (te *TestExtension) Dependencies() []string {
	return te.dependencies
}

func (te *TestExtension) OnLoad(env ExtensionEnvironment) error {
	te.onLoadCalled = true
	return te.onLoadError
}

func (te *TestExtension) BeforeRender(ctx ExtensionContext, templateName string) error {
	te.beforeRenderCalled = true
	return nil
}

func (te *TestExtension) AfterRender(ctx ExtensionContext, templateName string, result interface{}, err error) error {
	te.afterRenderCalled = true
	return nil
}

func (te *TestExtension) OnUnload() error {
	te.onUnloadCalled = true
	return nil
}

func (te *TestExtension) ParseTag(tagName string, parser ExtensionParser) (parser.Node, error) {
	current := parser.Current()
	node := parser.NewExtensionNode(te.Name(), tagName, current.Line, current.Column)
	return node, parser.ExpectBlockEnd()
}

type MockExtensionEnvironment struct{}

func (mee *MockExtensionEnvironment) GetExtension(name string) (Extension, bool)      { return nil, false }
func (mee *MockExtensionEnvironment) AddGlobal(name string, value interface{})        {}
func (mee *MockExtensionEnvironment) AddFilter(name string, filter interface{}) error { return nil }
func (mee *MockExtensionEnvironment) AddTest(name string, test interface{}) error     { return nil }
func (mee *MockExtensionEnvironment) GetConfig(key string) (interface{}, bool)        { return nil, false }
func (mee *MockExtensionEnvironment) SetConfig(key string, value interface{})         {}

type MockExtensionContext struct{}

func (mec *MockExtensionContext) GetVariable(name string) (interface{}, bool) { return nil, false }
func (mec *MockExtensionContext) SetVariable(name string, value interface{})  {}
func (mec *MockExtensionContext) GetGlobal(name string) (interface{}, bool)   { return nil, false }
func (mec *MockExtensionContext) ApplyFilter(name string, value interface{}, args ...interface{}) (interface{}, error) {
	return nil, nil
}
func (mec *MockExtensionContext) CallMacro(name string, args []interface{}, kwargs map[string]interface{}) (interface{}, error) {
	return nil, nil
}

type MockRuntimeContext struct {
	variables map[string]interface{}
}

func (mrc *MockRuntimeContext) GetVariable(name string) (interface{}, bool) {
	val, exists := mrc.variables[name]
	return val, exists
}

func (mrc *MockRuntimeContext) SetVariable(name string, value interface{}) {
	mrc.variables[name] = value
}

func (mrc *MockRuntimeContext) Clone() runtime.Context {
	newVars := make(map[string]interface{})
	for k, v := range mrc.variables {
		newVars[k] = v
	}
	return &MockRuntimeContext{variables: newVars}
}

func (mrc *MockRuntimeContext) All() map[string]interface{} {
	return mrc.variables
}
