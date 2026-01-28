package miya

import (
	"sync"
	"testing"
)

// TestStringBuilderPool tests the StringBuilderPool functions
func TestStringBuilderPool(t *testing.T) {
	t.Run("GetStringBuilder returns a builder", func(t *testing.T) {
		sb := GetStringBuilder()
		if sb == nil {
			t.Fatal("GetStringBuilder returned nil")
		}

		// Write something to verify it works
		sb.WriteString("hello")
		if sb.String() != "hello" {
			t.Errorf("Expected 'hello', got '%s'", sb.String())
		}

		PutStringBuilder(sb)
	})

	t.Run("PutStringBuilder resets the builder", func(t *testing.T) {
		sb := GetStringBuilder()
		sb.WriteString("test data")

		PutStringBuilder(sb)

		// Get another builder (might be the same one from pool)
		sb2 := GetStringBuilder()
		if sb2.Len() != 0 {
			t.Errorf("Expected empty builder, got length %d", sb2.Len())
		}
		PutStringBuilder(sb2)
	})

	t.Run("Concurrent access is safe", func(t *testing.T) {
		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(n int) {
				defer wg.Done()
				sb := GetStringBuilder()
				sb.WriteString("concurrent test")
				_ = sb.String()
				PutStringBuilder(sb)
			}(i)
		}
		wg.Wait()
	})
}

// TestContextPool tests the ContextPool functions
func TestContextPool(t *testing.T) {
	t.Run("GetContext returns a context", func(t *testing.T) {
		ctx := GetContext()
		if ctx == nil {
			t.Fatal("GetContext returned nil")
		}

		// Verify it works
		ctx.Set("key", "value")
		val, ok := ctx.Get("key")
		if !ok || val != "value" {
			t.Errorf("Context not working correctly")
		}

		PutContext(ctx)
	})

	t.Run("PutContext handles nil", func(t *testing.T) {
		// Should not panic
		PutContext(nil)
	})

	t.Run("Concurrent access is safe", func(t *testing.T) {
		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(n int) {
				defer wg.Done()
				ctx := GetContext()
				ctx.Set("key", n)
				PutContext(ctx)
			}(i)
		}
		wg.Wait()
	})
}

// TestCachedTemplate tests the CachedTemplate functionality
func TestCachedTemplate(t *testing.T) {
	t.Run("NewCachedTemplate creates a wrapper", func(t *testing.T) {
		env := NewEnvironment()
		tmpl, err := env.FromString("Hello {{ name }}!")
		if err != nil {
			t.Fatalf("Failed to create template: %v", err)
		}

		cached := NewCachedTemplate(tmpl)
		if cached == nil {
			t.Fatal("NewCachedTemplate returned nil")
		}
		if cached.Template != tmpl {
			t.Error("CachedTemplate does not wrap original template")
		}
	})

	t.Run("RenderCached renders correctly", func(t *testing.T) {
		env := NewEnvironment()
		tmpl, err := env.FromString("Hello {{ name }}!")
		if err != nil {
			t.Fatalf("Failed to create template: %v", err)
		}

		cached := NewCachedTemplate(tmpl)
		ctx := NewContextFrom(map[string]interface{}{"name": "World"})

		result, err := cached.RenderCached(ctx)
		if err != nil {
			t.Fatalf("RenderCached failed: %v", err)
		}
		if result != "Hello World!" {
			t.Errorf("Expected 'Hello World!', got '%s'", result)
		}
	})

	t.Run("RenderCached handles concurrent calls", func(t *testing.T) {
		env := NewEnvironment()
		tmpl, err := env.FromString("Value: {{ n }}")
		if err != nil {
			t.Fatalf("Failed to create template: %v", err)
		}

		cached := NewCachedTemplate(tmpl)

		var wg sync.WaitGroup
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func(n int) {
				defer wg.Done()
				ctx := NewContextFrom(map[string]interface{}{"n": n})
				_, err := cached.RenderCached(ctx)
				if err != nil {
					t.Errorf("RenderCached failed: %v", err)
				}
			}(i)
		}
		wg.Wait()
	})
}

// TestRenderContext tests the renderContext helper
func TestRenderContext(t *testing.T) {
	t.Run("reset clears builder and visited map", func(t *testing.T) {
		rc := &renderContext{
			builder: GetStringBuilder(),
			visited: make(map[string]bool),
		}

		rc.builder.WriteString("some content")
		rc.visited["key1"] = true
		rc.visited["key2"] = true

		rc.reset()

		if rc.builder.Len() != 0 {
			t.Errorf("Builder not reset, length: %d", rc.builder.Len())
		}
		if len(rc.visited) != 0 {
			t.Errorf("Visited map not cleared, length: %d", len(rc.visited))
		}

		PutStringBuilder(rc.builder)
	})
}

// TestSlicePool tests the SlicePool functionality
func TestSlicePool(t *testing.T) {
	t.Run("GetStringSlice returns a slice", func(t *testing.T) {
		slice := GlobalSlicePool.GetStringSlice()
		if slice == nil {
			t.Fatal("GetStringSlice returned nil")
		}
		if len(slice) != 0 {
			t.Errorf("Expected empty slice, got length %d", len(slice))
		}

		// Use it
		slice = append(slice, "a", "b", "c")
		if len(slice) != 3 {
			t.Errorf("Expected length 3, got %d", len(slice))
		}

		GlobalSlicePool.PutStringSlice(slice)
	})

	t.Run("PutStringSlice resets the slice", func(t *testing.T) {
		slice := GlobalSlicePool.GetStringSlice()
		slice = append(slice, "x", "y", "z")
		GlobalSlicePool.PutStringSlice(slice)

		// Get another slice
		slice2 := GlobalSlicePool.GetStringSlice()
		if len(slice2) != 0 {
			t.Errorf("Expected empty slice, got length %d", len(slice2))
		}
		GlobalSlicePool.PutStringSlice(slice2)
	})

	t.Run("GetIntSlice returns a slice", func(t *testing.T) {
		slice := GlobalSlicePool.GetIntSlice()
		if slice == nil {
			t.Fatal("GetIntSlice returned nil")
		}
		if len(slice) != 0 {
			t.Errorf("Expected empty slice, got length %d", len(slice))
		}

		// Use it
		slice = append(slice, 1, 2, 3, 4, 5)
		if len(slice) != 5 {
			t.Errorf("Expected length 5, got %d", len(slice))
		}

		GlobalSlicePool.PutIntSlice(slice)
	})

	t.Run("PutIntSlice resets the slice", func(t *testing.T) {
		slice := GlobalSlicePool.GetIntSlice()
		slice = append(slice, 10, 20, 30)
		GlobalSlicePool.PutIntSlice(slice)

		// Get another slice
		slice2 := GlobalSlicePool.GetIntSlice()
		if len(slice2) != 0 {
			t.Errorf("Expected empty slice, got length %d", len(slice2))
		}
		GlobalSlicePool.PutIntSlice(slice2)
	})

	t.Run("Concurrent access is safe", func(t *testing.T) {
		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(2)
			go func() {
				defer wg.Done()
				s := GlobalSlicePool.GetStringSlice()
				s = append(s, "test")
				GlobalSlicePool.PutStringSlice(s)
			}()
			go func() {
				defer wg.Done()
				s := GlobalSlicePool.GetIntSlice()
				s = append(s, 42)
				GlobalSlicePool.PutIntSlice(s)
			}()
		}
		wg.Wait()
	})
}

// TestFastStringBuilder tests the FastStringBuilder functionality
func TestFastStringBuilder(t *testing.T) {
	t.Run("NewFastStringBuilder creates builder with capacity", func(t *testing.T) {
		fsb := NewFastStringBuilder(100)
		if fsb == nil {
			t.Fatal("NewFastStringBuilder returned nil")
		}
		if fsb.Len() != 0 {
			t.Errorf("Expected length 0, got %d", fsb.Len())
		}
	})

	t.Run("WriteString appends strings", func(t *testing.T) {
		fsb := NewFastStringBuilder(64)
		fsb.WriteString("Hello")
		fsb.WriteString(" ")
		fsb.WriteString("World")

		if fsb.String() != "Hello World" {
			t.Errorf("Expected 'Hello World', got '%s'", fsb.String())
		}
		if fsb.Len() != 11 {
			t.Errorf("Expected length 11, got %d", fsb.Len())
		}
	})

	t.Run("WriteByte appends bytes", func(t *testing.T) {
		fsb := NewFastStringBuilder(64)
		fsb.WriteString("AB")
		err := fsb.WriteByte('C')
		if err != nil {
			t.Errorf("WriteByte returned error: %v", err)
		}

		if fsb.String() != "ABC" {
			t.Errorf("Expected 'ABC', got '%s'", fsb.String())
		}
	})

	t.Run("Reset clears the buffer", func(t *testing.T) {
		fsb := NewFastStringBuilder(64)
		fsb.WriteString("some content")
		fsb.Reset()

		if fsb.Len() != 0 {
			t.Errorf("Expected length 0 after reset, got %d", fsb.Len())
		}
		if fsb.String() != "" {
			t.Errorf("Expected empty string after reset, got '%s'", fsb.String())
		}
	})

	t.Run("Handles large content", func(t *testing.T) {
		fsb := NewFastStringBuilder(16) // Start small
		for i := 0; i < 1000; i++ {
			fsb.WriteString("x")
		}
		if fsb.Len() != 1000 {
			t.Errorf("Expected length 1000, got %d", fsb.Len())
		}
	})
}

// TestFastStringBuilderPool tests the FastStringBuilder pool
func TestFastStringBuilderPool(t *testing.T) {
	t.Run("GetFastStringBuilder returns a builder", func(t *testing.T) {
		fsb := GetFastStringBuilder()
		if fsb == nil {
			t.Fatal("GetFastStringBuilder returned nil")
		}
		fsb.WriteString("test")
		PutFastStringBuilder(fsb)
	})

	t.Run("PutFastStringBuilder resets the builder", func(t *testing.T) {
		fsb := GetFastStringBuilder()
		fsb.WriteString("content")
		PutFastStringBuilder(fsb)

		fsb2 := GetFastStringBuilder()
		if fsb2.Len() != 0 {
			t.Errorf("Expected empty builder, got length %d", fsb2.Len())
		}
		PutFastStringBuilder(fsb2)
	})

	t.Run("Large builders are not returned to pool", func(t *testing.T) {
		fsb := GetFastStringBuilder()
		// Write more than 64KB
		largeContent := make([]byte, 65*1024)
		for i := range largeContent {
			largeContent[i] = 'x'
		}
		fsb.WriteString(string(largeContent))

		// This should not panic and should not return to pool
		PutFastStringBuilder(fsb)
	})

	t.Run("Concurrent access is safe", func(t *testing.T) {
		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				fsb := GetFastStringBuilder()
				fsb.WriteString("concurrent test")
				_ = fsb.String()
				PutFastStringBuilder(fsb)
			}()
		}
		wg.Wait()
	})
}

// TestTemplateCache tests the LRU template cache
func TestTemplateCache(t *testing.T) {
	t.Run("NewTemplateCache creates empty cache", func(t *testing.T) {
		cache := NewTemplateCache(10)
		if cache == nil {
			t.Fatal("NewTemplateCache returned nil")
		}
		if len(cache.cache) != 0 {
			t.Errorf("Expected empty cache, got %d items", len(cache.cache))
		}
	})

	t.Run("Put and Get work correctly", func(t *testing.T) {
		cache := NewTemplateCache(10)
		env := NewEnvironment()

		tmpl, _ := env.FromString("Hello {{ name }}")
		cache.Put("key1", tmpl)

		retrieved, exists := cache.Get("key1")
		if !exists {
			t.Fatal("Expected to find template in cache")
		}
		if retrieved != tmpl {
			t.Error("Retrieved template doesn't match stored template")
		}
	})

	t.Run("Get returns false for missing keys", func(t *testing.T) {
		cache := NewTemplateCache(10)

		_, exists := cache.Get("nonexistent")
		if exists {
			t.Error("Expected exists to be false for missing key")
		}
	})

	t.Run("LRU eviction works", func(t *testing.T) {
		cache := NewTemplateCache(3)
		env := NewEnvironment()

		tmpl1, _ := env.FromString("Template 1")
		tmpl2, _ := env.FromString("Template 2")
		tmpl3, _ := env.FromString("Template 3")
		tmpl4, _ := env.FromString("Template 4")

		cache.Put("key1", tmpl1)
		cache.Put("key2", tmpl2)
		cache.Put("key3", tmpl3)

		// Access key1 to make it recently used
		cache.Get("key1")

		// Add key4, should evict key2 (least recently used)
		cache.Put("key4", tmpl4)

		// key2 should be evicted
		_, exists := cache.Get("key2")
		if exists {
			t.Error("key2 should have been evicted")
		}

		// key1 should still exist
		_, exists = cache.Get("key1")
		if !exists {
			t.Error("key1 should still exist")
		}
	})

	t.Run("Put updates existing key", func(t *testing.T) {
		cache := NewTemplateCache(10)
		env := NewEnvironment()

		tmpl1, _ := env.FromString("Original")
		tmpl2, _ := env.FromString("Updated")

		cache.Put("key", tmpl1)
		cache.Put("key", tmpl2)

		retrieved, _ := cache.Get("key")
		if retrieved != tmpl2 {
			t.Error("Template should have been updated")
		}
	})

	t.Run("Concurrent access is safe", func(t *testing.T) {
		cache := NewTemplateCache(100)
		env := NewEnvironment()

		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(2)
			go func(n int) {
				defer wg.Done()
				tmpl, _ := env.FromString("Template")
				cache.Put(string(rune('a'+n%26)), tmpl)
			}(i)
			go func(n int) {
				defer wg.Done()
				cache.Get(string(rune('a' + n%26)))
			}(i)
		}
		wg.Wait()
	})
}

// TestCachedEnvironment tests the CachedEnvironment functionality
func TestCachedEnvironment(t *testing.T) {
	t.Run("NewCachedEnvironment creates environment", func(t *testing.T) {
		env := NewCachedEnvironment()
		if env == nil {
			t.Fatal("NewCachedEnvironment returned nil")
		}
		if env.Environment == nil {
			t.Error("Environment is nil")
		}
		if env.templateCache == nil {
			t.Error("templateCache is nil")
		}
	})

	t.Run("NewCachedEnvironment accepts options", func(t *testing.T) {
		env := NewCachedEnvironment(WithAutoEscape(true))
		if env == nil {
			t.Fatal("NewCachedEnvironment returned nil")
		}
	})

	t.Run("FromStringCached compiles and caches", func(t *testing.T) {
		env := NewCachedEnvironment()

		source := "Hello {{ name }}!"

		// First call - should compile
		tmpl1, err := env.FromStringCached(source)
		if err != nil {
			t.Fatalf("FromStringCached failed: %v", err)
		}

		// Second call - should return cached
		tmpl2, err := env.FromStringCached(source)
		if err != nil {
			t.Fatalf("FromStringCached failed: %v", err)
		}

		// Should be the same template instance
		if tmpl1 != tmpl2 {
			t.Error("Expected same template instance from cache")
		}
	})

	t.Run("FromStringCached returns error for invalid template", func(t *testing.T) {
		env := NewCachedEnvironment()

		_, err := env.FromStringCached("{% if %}")
		if err == nil {
			t.Error("Expected error for invalid template")
		}
	})

	t.Run("Cached template renders correctly", func(t *testing.T) {
		env := NewCachedEnvironment()

		tmpl, err := env.FromStringCached("Hello {{ name }}!")
		if err != nil {
			t.Fatalf("FromStringCached failed: %v", err)
		}

		ctx := NewContextFrom(map[string]interface{}{"name": "World"})
		result, err := tmpl.Render(ctx)
		if err != nil {
			t.Fatalf("Render failed: %v", err)
		}
		if result != "Hello World!" {
			t.Errorf("Expected 'Hello World!', got '%s'", result)
		}
	})

	t.Run("Concurrent FromStringCached is safe", func(t *testing.T) {
		env := NewCachedEnvironment()

		var wg sync.WaitGroup
		sources := []string{
			"Template A: {{ a }}",
			"Template B: {{ b }}",
			"Template C: {{ c }}",
		}

		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(n int) {
				defer wg.Done()
				source := sources[n%len(sources)]
				_, err := env.FromStringCached(source)
				if err != nil {
					t.Errorf("FromStringCached failed: %v", err)
				}
			}(i)
		}
		wg.Wait()
	})
}
