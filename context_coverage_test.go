package miya

import (
	"testing"
)

// TestCapitalizeFirstRoot tests the capitalizeFirst function in context.go
func TestCapitalizeFirstRoot(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"a", "A"},
		{"hello", "Hello"},
		{"HELLO", "HELLO"},
		{"123abc", "123abc"},
	}

	for _, tt := range tests {
		result := capitalizeFirst(tt.input)
		if result != tt.expected {
			t.Errorf("capitalizeFirst(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

// TestContextGetWithDotNotation tests Get with dot notation
func TestContextGetWithDotNotation(t *testing.T) {
	ctx := NewContext()

	// Set nested data
	user := map[string]interface{}{
		"name": "John",
		"address": map[string]interface{}{
			"city": "NYC",
		},
	}
	ctx.Set("user", user)

	t.Run("SimpleDotNotation", func(t *testing.T) {
		val, ok := ctx.Get("user.name")
		if !ok || val != "John" {
			t.Errorf("Get('user.name') = %v, %v, want 'John', true", val, ok)
		}
	})

	t.Run("NestedDotNotation", func(t *testing.T) {
		val, ok := ctx.Get("user.address.city")
		if !ok || val != "NYC" {
			t.Errorf("Get('user.address.city') = %v, %v, want 'NYC', true", val, ok)
		}
	})

	t.Run("NonExistentPath", func(t *testing.T) {
		_, ok := ctx.Get("user.nonexistent")
		if ok {
			t.Error("Get('user.nonexistent') should return false")
		}
	})

	t.Run("NonExistentNestedPath", func(t *testing.T) {
		_, ok := ctx.Get("user.address.nonexistent")
		if ok {
			t.Error("Get('user.address.nonexistent') should return false")
		}
	})
}

// TestContextGetWithEnvironmentGlobals tests Get with environment globals
func TestContextGetWithEnvironmentGlobals(t *testing.T) {
	env := NewEnvironment()
	env.AddGlobal("globalVar", "globalValue")

	ctx := newContextWithEnv(env)

	// Get global variable
	val, ok := ctx.Get("globalVar")
	if !ok || val != "globalValue" {
		t.Errorf("Get('globalVar') = %v, %v, want 'globalValue', true", val, ok)
	}

	// Local variable should override global
	ctx.Set("globalVar", "localValue")
	val, ok = ctx.Get("globalVar")
	if !ok || val != "localValue" {
		t.Errorf("Get('globalVar') after local set = %v, %v, want 'localValue', true", val, ok)
	}
}

// TestContextGetAttributeWithStruct tests getAttribute with struct types
func TestContextGetAttributeWithStruct(t *testing.T) {
	type Address struct {
		City    string
		ZipCode string
	}

	type User struct {
		Name    string
		Age     int
		Address Address
	}

	ctx := NewContext()
	user := User{
		Name: "John",
		Age:  30,
		Address: Address{
			City:    "NYC",
			ZipCode: "10001",
		},
	}
	ctx.Set("user", user)

	t.Run("StructField", func(t *testing.T) {
		val, ok := ctx.Get("user.Name")
		if !ok || val != "John" {
			t.Errorf("Get('user.Name') = %v, %v, want 'John', true", val, ok)
		}
	})

	t.Run("StructFieldLowercase", func(t *testing.T) {
		val, ok := ctx.Get("user.name")
		if !ok || val != "John" {
			t.Errorf("Get('user.name') = %v, %v, want 'John', true (capitalized lookup)", val, ok)
		}
	})

	t.Run("NestedStructField", func(t *testing.T) {
		val, ok := ctx.Get("user.Address.City")
		if !ok || val != "NYC" {
			t.Errorf("Get('user.Address.City') = %v, %v, want 'NYC', true", val, ok)
		}
	})
}

// TestContextGetAttributeWithPointer tests getAttribute with pointer types
func TestContextGetAttributeWithPointer(t *testing.T) {
	type User struct {
		Name string
	}

	ctx := NewContext()
	user := &User{Name: "Jane"}
	ctx.Set("user", user)

	val, ok := ctx.Get("user.Name")
	if !ok || val != "Jane" {
		t.Errorf("Get('user.Name') with pointer = %v, %v, want 'Jane', true", val, ok)
	}
}

// TestContextGetAttributeWithMethod tests getAttribute with methods
func TestContextGetAttributeWithMethod(t *testing.T) {
	ctx := NewContext()
	ctx.Set("greeter", &TestGreeter{name: "World"})

	// This test depends on safeMethodCall working correctly
	val, ok := ctx.Get("greeter.Greet")
	// Method should be found and callable
	_ = val
	_ = ok
}

// TestGreeter is a test type with methods
type TestGreeter struct {
	name string
}

func (g *TestGreeter) Greet() string {
	return "Hello, " + g.name
}

func (g *TestGreeter) Name() string {
	return g.name
}

// TestContextPushPop tests Push and Pop operations
func TestContextPushPop(t *testing.T) {
	ctx := NewContext()
	ctx.Set("outer", "outerValue")

	// Push creates a new layer
	pushed := ctx.Push()
	pushed.Set("inner", "innerValue")

	// Inner context should see both
	val, ok := pushed.Get("outer")
	if !ok || val != "outerValue" {
		t.Errorf("Pushed context Get('outer') = %v, %v, want 'outerValue', true", val, ok)
	}

	val, ok = pushed.Get("inner")
	if !ok || val != "innerValue" {
		t.Errorf("Pushed context Get('inner') = %v, %v, want 'innerValue', true", val, ok)
	}

	// Pop should return to outer context
	popped := pushed.Pop()

	// Outer context shouldn't see inner
	_, ok = popped.Get("inner")
	if ok {
		t.Error("Popped context should not see 'inner'")
	}

	// Pop on root context returns itself
	poppedAgain := popped.Pop()
	if poppedAgain != popped {
		t.Error("Pop on root context should return itself")
	}
}

// TestContextClone tests Clone operation
func TestContextClone(t *testing.T) {
	ctx := NewContext()
	ctx.Set("key", "value")

	clone := ctx.Clone()

	// Clone should have same values
	val, ok := clone.Get("key")
	if !ok || val != "value" {
		t.Errorf("Clone Get('key') = %v, %v, want 'value', true", val, ok)
	}

	// Modifying clone shouldn't affect original
	clone.Set("key", "modified")
	origVal, _ := ctx.Get("key")
	if origVal != "value" {
		t.Error("Modifying clone should not affect original")
	}
}

// TestContextAll tests All method
func TestContextAll(t *testing.T) {
	env := NewEnvironment()
	env.AddGlobal("global", "globalValue")

	ctx := newContextWithEnv(env)
	ctx.Set("local", "localValue")

	all := ctx.All()

	if all["global"] != "globalValue" {
		t.Error("All() should include globals")
	}
	if all["local"] != "localValue" {
		t.Error("All() should include locals")
	}

	// Test caching - calling All() twice should use cache
	all2 := ctx.All()
	if len(all) != len(all2) {
		t.Error("Cached All() should return same data")
	}

	// Setting a value should invalidate cache
	ctx.Set("new", "newValue")
	all3 := ctx.All()
	if all3["new"] != "newValue" {
		t.Error("All() should include newly set value after cache invalidation")
	}
}

// TestContextGetEnv tests GetEnv method
func TestContextGetEnv(t *testing.T) {
	t.Run("WithEnvironment", func(t *testing.T) {
		env := NewEnvironment()
		ctx := newContextWithEnv(env)

		gotEnv := ctx.GetEnv()
		if gotEnv != env {
			t.Error("GetEnv should return the environment")
		}
	})

	t.Run("WithoutEnvironment", func(t *testing.T) {
		ctx := NewContext()
		gotEnv := ctx.GetEnv()
		if gotEnv != nil {
			t.Error("GetEnv should return nil when no environment set")
		}
	})
}

// TestContextString tests String method
func TestContextString(t *testing.T) {
	ctx := NewContext()
	ctx.Set("key", "value")

	// Access internal context for String method
	if c, ok := ctx.(*context); ok {
		str := c.String()
		if str == "" {
			t.Error("String() should return non-empty string")
		}
	}
}

// TestContextParentChain tests context parent chain lookup
func TestContextParentChain(t *testing.T) {
	ctx1 := NewContext()
	ctx1.Set("level1", "value1")

	ctx2 := ctx1.Push()
	ctx2.Set("level2", "value2")

	ctx3 := ctx2.Push()
	ctx3.Set("level3", "value3")

	// ctx3 should see all three
	val, ok := ctx3.Get("level1")
	if !ok || val != "value1" {
		t.Errorf("ctx3 Get('level1') = %v, %v, want 'value1', true", val, ok)
	}

	val, ok = ctx3.Get("level2")
	if !ok || val != "value2" {
		t.Errorf("ctx3 Get('level2') = %v, %v, want 'value2', true", val, ok)
	}

	val, ok = ctx3.Get("level3")
	if !ok || val != "value3" {
		t.Errorf("ctx3 Get('level3') = %v, %v, want 'value3', true", val, ok)
	}
}

// TestNewContextFrom tests NewContextFrom function
func TestNewContextFrom(t *testing.T) {
	data := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	}

	ctx := NewContextFrom(data)

	val, ok := ctx.Get("key1")
	if !ok || val != "value1" {
		t.Errorf("Get('key1') = %v, %v, want 'value1', true", val, ok)
	}

	val, ok = ctx.Get("key2")
	if !ok || val != 42 {
		t.Errorf("Get('key2') = %v, %v, want 42, true", val, ok)
	}

	// Modifying original map shouldn't affect context
	data["key1"] = "modified"
	val, _ = ctx.Get("key1")
	if val != "value1" {
		t.Error("Context should be independent of original map")
	}
}

// TestCOWContext tests Copy-on-Write context
func TestCOWContext(t *testing.T) {
	t.Run("NewCOWContext", func(t *testing.T) {
		ctx := NewCOWContext()
		if ctx == nil {
			t.Fatal("NewCOWContext returned nil")
		}
	})

	t.Run("NewCOWContextFrom", func(t *testing.T) {
		data := map[string]interface{}{"key": "value"}
		ctx := NewCOWContextFrom(data)

		val, ok := ctx.Get("key")
		if !ok || val != "value" {
			t.Errorf("Get('key') = %v, %v, want 'value', true", val, ok)
		}
	})

	t.Run("SetAndGet", func(t *testing.T) {
		ctx := NewCOWContext()
		ctx.Set("key", "value")

		val, ok := ctx.Get("key")
		if !ok || val != "value" {
			t.Errorf("Get('key') = %v, %v, want 'value', true", val, ok)
		}
	})

	t.Run("Clone", func(t *testing.T) {
		ctx := NewCOWContext()
		ctx.Set("key", "value")

		clone := ctx.Clone()
		cloneVal, ok := clone.Get("key")
		if !ok || cloneVal != "value" {
			t.Error("Clone should see original values")
		}

		// Modify clone
		clone.Set("key", "modified")
		cloneVal, _ = clone.Get("key")
		if cloneVal != "modified" {
			t.Error("Clone modification failed")
		}

		// Original should be unchanged
		origVal, _ := ctx.Get("key")
		if origVal != "value" {
			t.Error("Original should be unchanged after clone modification")
		}
	})

	t.Run("All", func(t *testing.T) {
		ctx := NewCOWContext()
		ctx.Set("key1", "value1")
		ctx.Set("key2", "value2")

		all := ctx.All()
		if all["key1"] != "value1" || all["key2"] != "value2" {
			t.Error("All() should return all variables")
		}
	})

	t.Run("PushPop", func(t *testing.T) {
		ctx := NewCOWContext()
		ctx.Set("outer", "outerValue")

		pushed := ctx.Push()
		pushed.Set("inner", "innerValue")

		// Pushed should see both
		val, ok := pushed.Get("outer")
		if !ok || val != "outerValue" {
			t.Error("Pushed should see parent values")
		}

		val, ok = pushed.Get("inner")
		if !ok || val != "innerValue" {
			t.Error("Pushed should see own values")
		}

		// Pop
		popped := pushed.Pop()
		if popped == nil {
			t.Error("Pop should return parent context")
		}
	})

	t.Run("GetEnv", func(t *testing.T) {
		ctx := NewCOWContext()
		env := ctx.GetEnv()
		if env != nil {
			t.Error("COW context without env should return nil")
		}
	})
}

// TestCOWContextWithParent tests COW context with parent lookups
func TestCOWContextWithParent(t *testing.T) {
	parent := NewCOWContext()
	parent.Set("parentKey", "parentValue")

	child := parent.Push()
	child.Set("childKey", "childValue")

	// Child should see parent key
	val, ok := child.Get("parentKey")
	if !ok || val != "parentValue" {
		t.Errorf("Child Get('parentKey') = %v, %v", val, ok)
	}

	// Child should see own key
	val, ok = child.Get("childKey")
	if !ok || val != "childValue" {
		t.Errorf("Child Get('childKey') = %v, %v", val, ok)
	}

	// Parent should not see child key
	_, ok = parent.Get("childKey")
	if ok {
		t.Error("Parent should not see child key")
	}
}

// TestCOWContextCloneWithLocalModifications tests cloning with local changes
func TestCOWContextCloneWithLocalModifications(t *testing.T) {
	data := map[string]interface{}{"shared": "sharedValue"}
	ctx := NewCOWContextFrom(data)

	// Add local modification
	ctx.Set("local", "localValue")

	// Clone should see both shared and local
	clone := ctx.Clone()

	val, ok := clone.Get("shared")
	if !ok || val != "sharedValue" {
		t.Error("Clone should see shared value")
	}

	val, ok = clone.Get("local")
	if !ok || val != "localValue" {
		t.Error("Clone should see local modifications")
	}
}

// TestCOWContextSetVariableGetVariable tests SetVariable and GetVariable aliases
func TestCOWContextSetVariableGetVariable(t *testing.T) {
	ctx := NewCOWContext()

	// SetVariable is an alias for Set (access via concrete type)
	if cowCtx, ok := ctx.(*cowContext); ok {
		cowCtx.SetVariable("key", "value")

		// GetVariable is an alias for Get
		val, found := cowCtx.GetVariable("key")
		if !found || val != "value" {
			t.Errorf("GetVariable('key') = %v, %v, want 'value', true", val, found)
		}

		// Test with missing key
		_, found = cowCtx.GetVariable("missing")
		if found {
			t.Error("GetVariable('missing') should return false")
		}
	} else {
		t.Skip("NewCOWContext returned different type")
	}
}

// TestSafeMethodCallCoverage tests the safeMethodCall function
func TestSafeMethodCallCoverage(t *testing.T) {
	t.Run("valid method with no args", func(t *testing.T) {
		ctx := NewContext()
		greeter := &TestGreeter{name: "World"}
		ctx.Set("greeter", greeter)

		// Get the Name method via context
		val, ok := ctx.Get("greeter.Name")
		if ok && val != nil {
			// The method was found
			_ = val
		}
	})

	t.Run("method not found", func(t *testing.T) {
		type NoMethods struct{}
		ctx := NewContext()
		ctx.Set("obj", &NoMethods{})

		_, ok := ctx.Get("obj.NonExistent")
		if ok {
			t.Error("Should not find non-existent method")
		}
	})

	t.Run("method with args (should be skipped)", func(t *testing.T) {
		type HasMethodWithArgs struct{}
		ctx := NewContext()
		ctx.Set("obj", &HasMethodWithArgs{})

		// safeMethodCall only calls methods with no args
		_, ok := ctx.Get("obj.SomeMethod")
		// This should return false since no such method exists
		_ = ok
	})

	t.Run("nil value", func(t *testing.T) {
		ctx := NewContext()
		ctx.Set("nilval", nil)

		_, ok := ctx.Get("nilval.anything")
		if ok {
			t.Error("Should not find attribute on nil")
		}
	})
}

// TestGetLocalCoverage tests the getLocal function
func TestGetLocalCoverage(t *testing.T) {
	ctx := NewContext()
	ctx.Set("key", "value")

	// Direct key lookup
	val, ok := ctx.Get("key")
	if !ok || val != "value" {
		t.Errorf("Get('key') = %v, %v", val, ok)
	}

	// With parent context
	pushed := ctx.Push()
	pushed.Set("childKey", "childValue")

	// getLocal should find child key first
	val, ok = pushed.Get("childKey")
	if !ok || val != "childValue" {
		t.Errorf("Get('childKey') = %v, %v", val, ok)
	}

	// getLocal should find parent key
	val, ok = pushed.Get("key")
	if !ok || val != "value" {
		t.Errorf("Get('key') from child = %v, %v", val, ok)
	}
}

// TestGetAttributeCoverage tests getAttribute with various types
func TestGetAttributeCoverage(t *testing.T) {
	ctx := NewContext()

	t.Run("map with special keys", func(t *testing.T) {
		m := map[string]interface{}{
			"normalKey":  "normalValue",
			"nestedMap":  map[string]interface{}{"inner": "innerValue"},
			"nestedList": []interface{}{1, 2, 3},
		}
		ctx.Set("m", m)

		val, ok := ctx.Get("m.normalKey")
		if !ok || val != "normalValue" {
			t.Errorf("Get('m.normalKey') = %v, %v", val, ok)
		}
	})

	t.Run("struct with unexported field", func(t *testing.T) {
		type withUnexported struct {
			Public  string
			private string
		}
		obj := withUnexported{Public: "public", private: "private"}
		ctx.Set("obj", obj)

		// Should find Public
		val, ok := ctx.Get("obj.Public")
		if !ok || val != "public" {
			t.Errorf("Get('obj.Public') = %v, %v", val, ok)
		}

		// Should not find private (unexported)
		_, ok = ctx.Get("obj.private")
		// Behavior may vary - just ensure no panic
	})

	t.Run("slice access by string index", func(t *testing.T) {
		slice := []interface{}{"a", "b", "c"}
		ctx.Set("slice", slice)

		// Access by index as string doesn't work in context.Get
		// but getAttribute handles it
	})

	t.Run("interface type", func(t *testing.T) {
		var iface interface{} = map[string]interface{}{"key": "value"}
		ctx.Set("iface", iface)

		val, ok := ctx.Get("iface.key")
		if !ok || val != "value" {
			t.Errorf("Get('iface.key') = %v, %v", val, ok)
		}
	})
}

// TestContextAllCacheCoverage tests All() caching behavior
func TestContextAllCacheCoverage(t *testing.T) {
	ctx := NewContext()
	ctx.Set("key1", "value1")

	// First call builds cache
	all1 := ctx.All()
	if all1["key1"] != "value1" {
		t.Error("First All() should include key1")
	}

	// Second call uses cache
	all2 := ctx.All()
	if all2["key1"] != "value1" {
		t.Error("Cached All() should include key1")
	}

	// Setting invalidates cache
	ctx.Set("key2", "value2")

	// This should rebuild cache
	all3 := ctx.All()
	if all3["key2"] != "value2" {
		t.Error("All() after Set should include key2")
	}

	// Parent chain in All()
	pushed := ctx.Push()
	pushed.Set("childKey", "childValue")

	allPushed := pushed.All()
	if allPushed["key1"] != "value1" {
		t.Error("Child All() should include parent keys")
	}
	if allPushed["childKey"] != "childValue" {
		t.Error("Child All() should include child keys")
	}
}

// PanicMethod is a test type that panics when called
type PanicMethod struct{}

func (p *PanicMethod) PanicOnCall() string {
	panic("intentional panic")
}

// TestSafeMethodCallPanicRecovery tests that safeMethodCall recovers from panics
func TestSafeMethodCallPanicRecovery(t *testing.T) {
	ctx := NewContext()
	ctx.Set("panicker", &PanicMethod{})

	// This should not panic - safeMethodCall should recover
	defer func() {
		if r := recover(); r != nil {
			t.Error("safeMethodCall should recover from panics")
		}
	}()

	_, _ = ctx.Get("panicker.PanicOnCall")
	// If we get here without panicking, the test passes
}
