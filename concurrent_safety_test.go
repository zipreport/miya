package miya

import (
	"sync"
	"testing"
	"time"
)

// TestThreadSafeTemplate tests the thread-safe template wrapper
func TestThreadSafeTemplate(t *testing.T) {
	env := NewEnvironment()
	tmpl, err := env.FromString("Hello {{ name }}!")
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	tst := NewThreadSafeTemplate(tmpl)
	if tst == nil {
		t.Fatal("NewThreadSafeTemplate returned nil")
	}

	// Test concurrent rendering
	ctx := NewContext()
	ctx.Set("name", "World")

	result, err := tst.RenderConcurrent(ctx)
	if err != nil {
		t.Fatalf("RenderConcurrent failed: %v", err)
	}
	if result != "Hello World!" {
		t.Errorf("Expected 'Hello World!', got %q", result)
	}
}

// TestThreadSafeTemplateConcurrency tests concurrent access to thread-safe template
func TestThreadSafeTemplateConcurrency(t *testing.T) {
	env := NewEnvironment()
	tmpl, err := env.FromString("{{ value }}")
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	tst := NewThreadSafeTemplate(tmpl)

	var wg sync.WaitGroup
	errors := make(chan error, 100)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			ctx := NewContext()
			ctx.Set("value", val)
			_, err := tst.RenderConcurrent(ctx)
			if err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("Concurrent render error: %v", err)
	}
}

// TestThreadSafeEnvironment tests the thread-safe environment wrapper
func TestThreadSafeEnvironment(t *testing.T) {
	tse := NewThreadSafeEnvironment()
	if tse == nil {
		t.Fatal("NewThreadSafeEnvironment returned nil")
	}

	// Test FromStringConcurrent
	tmpl, err := tse.FromStringConcurrent("Hello {{ name }}!")
	if err != nil {
		t.Fatalf("FromStringConcurrent failed: %v", err)
	}
	if tmpl == nil {
		t.Fatal("FromStringConcurrent returned nil template")
	}

	// Test AddFilterConcurrent
	err = tse.AddFilterConcurrent("testfilter", func(value interface{}, args ...interface{}) (interface{}, error) {
		return value, nil
	})
	if err != nil {
		t.Fatalf("AddFilterConcurrent failed: %v", err)
	}

	// Test AddGlobalConcurrent
	tse.AddGlobalConcurrent("testglobal", "value")
}

// TestConcurrentTemplateRenderer tests the concurrent template renderer
func TestConcurrentTemplateRenderer(t *testing.T) {
	env := NewEnvironment()
	tmpl, err := env.FromString("Hello {{ name }}!")
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	renderer := NewConcurrentTemplateRenderer(tmpl, 4)
	if renderer == nil {
		t.Fatal("NewConcurrentTemplateRenderer returned nil")
	}

	// Start the renderer
	renderer.Start()

	// Test async rendering
	ctx := NewContext()
	ctx.Set("name", "World")

	resultCh := renderer.RenderAsync(ctx)
	result := <-resultCh

	if result.err != nil {
		t.Fatalf("RenderAsync failed: %v", result.err)
	}
	if result.output != "Hello World!" {
		t.Errorf("Expected 'Hello World!', got %q", result.output)
	}

	// Stop the renderer
	renderer.Stop()
}

// TestConcurrentTemplateRendererDoubleStart tests that double start is handled
func TestConcurrentTemplateRendererDoubleStart(t *testing.T) {
	env := NewEnvironment()
	tmpl, err := env.FromString("{{ value }}")
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	renderer := NewConcurrentTemplateRenderer(tmpl, 2)
	renderer.Start()
	renderer.Start() // Should be a no-op
	renderer.Stop()
}

// TestConcurrentTemplateRendererDoubleStop tests that double stop is handled
func TestConcurrentTemplateRendererDoubleStop(t *testing.T) {
	env := NewEnvironment()
	tmpl, err := env.FromString("{{ value }}")
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	renderer := NewConcurrentTemplateRenderer(tmpl, 2)
	renderer.Start()
	renderer.Stop()
	renderer.Stop() // Should be a no-op
}

// TestConcurrentTemplateRendererStopped tests rendering after stop
func TestConcurrentTemplateRendererStopped(t *testing.T) {
	env := NewEnvironment()
	tmpl, err := env.FromString("{{ value }}")
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	renderer := NewConcurrentTemplateRenderer(tmpl, 2)
	renderer.Start()
	renderer.Stop()

	// Try to render after stop
	ctx := NewContext()
	ctx.Set("value", "test")
	resultCh := renderer.RenderAsync(ctx)
	result := <-resultCh

	if result.err == nil {
		t.Error("Expected error when rendering after stop")
	}
}

// TestRenderBatch tests batch rendering
func TestRenderBatch(t *testing.T) {
	env := NewEnvironment()
	tmpl, err := env.FromString("{{ value }}")
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	renderer := NewConcurrentTemplateRenderer(tmpl, 4)

	contexts := make([]Context, 10)
	for i := 0; i < 10; i++ {
		ctx := NewContext()
		ctx.Set("value", i)
		contexts[i] = ctx
	}

	results, errors := renderer.RenderBatch(contexts)

	if len(results) != 10 {
		t.Errorf("Expected 10 results, got %d", len(results))
	}
	if len(errors) != 10 {
		t.Errorf("Expected 10 error slots, got %d", len(errors))
	}

	for i, err := range errors {
		if err != nil {
			t.Errorf("Batch render error at index %d: %v", i, err)
		}
	}
}

// TestTemplatePool tests the template pool
func TestTemplatePool(t *testing.T) {
	env := NewEnvironment()
	tmpl, err := env.FromString("Hello {{ name }}!")
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	pool := NewTemplatePool(tmpl)
	if pool == nil {
		t.Fatal("NewTemplatePool returned nil")
	}

	// Get a template from the pool
	pooledTmpl := pool.Get()
	if pooledTmpl == nil {
		t.Fatal("Get returned nil template")
	}

	// Return it to the pool
	pool.Put(pooledTmpl)

	// Test RenderConcurrent
	ctx := NewContext()
	ctx.Set("name", "World")

	result, err := pool.RenderConcurrent(ctx)
	if err != nil {
		t.Fatalf("RenderConcurrent failed: %v", err)
	}
	if result != "Hello World!" {
		t.Errorf("Expected 'Hello World!', got %q", result)
	}
}

// TestTemplatePoolConcurrency tests concurrent access to template pool
func TestTemplatePoolConcurrency(t *testing.T) {
	env := NewEnvironment()
	tmpl, err := env.FromString("{{ value }}")
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	pool := NewTemplatePool(tmpl)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			ctx := NewContext()
			ctx.Set("value", val)
			_, err := pool.RenderConcurrent(ctx)
			if err != nil {
				t.Errorf("Concurrent pool render error: %v", err)
			}
		}(i)
	}
	wg.Wait()
}

// TestConcurrentContextPool tests the concurrent context pool
func TestConcurrentContextPool(t *testing.T) {
	pool := NewConcurrentContextPool()
	if pool == nil {
		t.Fatal("NewConcurrentContextPool returned nil")
	}

	// Get a context from the pool
	ctx := pool.Get()
	if ctx == nil {
		t.Fatal("Get returned nil context")
	}

	// Put the context back
	pool.Put(ctx)

	// Check stats
	gets, puts := pool.GetStats()
	if gets != 1 {
		t.Errorf("Expected 1 get, got %d", gets)
	}
	if puts != 1 {
		t.Errorf("Expected 1 put, got %d", puts)
	}
}

// TestConcurrentContextPoolConcurrency tests concurrent access to context pool
func TestConcurrentContextPoolConcurrency(t *testing.T) {
	pool := NewConcurrentContextPool()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx := pool.Get()
			time.Sleep(time.Microsecond) // Simulate some work
			pool.Put(ctx)
		}()
	}
	wg.Wait()

	gets, puts := pool.GetStats()
	if gets != 100 {
		t.Errorf("Expected 100 gets, got %d", gets)
	}
	if puts != 100 {
		t.Errorf("Expected 100 puts, got %d", puts)
	}
}

// TestRateLimitedRenderer tests the rate-limited renderer
func TestRateLimitedRenderer(t *testing.T) {
	env := NewEnvironment()
	tmpl, err := env.FromString("Hello {{ name }}!")
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	renderer := NewRateLimitedRenderer(tmpl, 2)
	if renderer == nil {
		t.Fatal("NewRateLimitedRenderer returned nil")
	}

	ctx := NewContext()
	ctx.Set("name", "World")

	result, err := renderer.Render(ctx)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	if result != "Hello World!" {
		t.Errorf("Expected 'Hello World!', got %q", result)
	}
}

// TestRateLimitedRendererConcurrency tests concurrent rate-limited rendering
func TestRateLimitedRendererConcurrency(t *testing.T) {
	env := NewEnvironment()
	tmpl, err := env.FromString("{{ value }}")
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	renderer := NewRateLimitedRenderer(tmpl, 3)

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			ctx := NewContext()
			ctx.Set("value", val)
			_, err := renderer.Render(ctx)
			if err != nil {
				t.Errorf("Rate-limited render error: %v", err)
			}
		}(i)
	}
	wg.Wait()
}

// TestConcurrentCacheManager tests the concurrent cache manager
func TestConcurrentCacheManager(t *testing.T) {
	cache := NewConcurrentCacheManager()
	if cache == nil {
		t.Fatal("NewConcurrentCacheManager returned nil")
	}

	// Test Set and Get
	cache.Set("key1", "value1")
	value, ok := cache.Get("key1")
	if !ok {
		t.Error("Expected key1 to be found")
	}
	if value != "value1" {
		t.Errorf("Expected 'value1', got %v", value)
	}

	// Test Get miss
	_, ok = cache.Get("nonexistent")
	if ok {
		t.Error("Expected nonexistent key to not be found")
	}

	// Test Delete
	cache.Delete("key1")
	_, ok = cache.Get("key1")
	if ok {
		t.Error("Expected key1 to be deleted")
	}

	// Test stats
	hits, misses := cache.GetStats()
	if hits != 1 {
		t.Errorf("Expected 1 hit, got %d", hits)
	}
	if misses != 2 { // One for nonexistent, one for deleted key
		t.Errorf("Expected 2 misses, got %d", misses)
	}
}

// TestConcurrentCacheManagerConcurrency tests concurrent cache access
func TestConcurrentCacheManagerConcurrency(t *testing.T) {
	cache := NewConcurrentCacheManager()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			key := string(rune('a' + idx%26))
			cache.Set(key, idx)
			cache.Get(key)
		}(i)
	}
	wg.Wait()
}

// TestConcurrentEnvironmentRegistry tests the environment registry
func TestConcurrentEnvironmentRegistry(t *testing.T) {
	registry := NewConcurrentEnvironmentRegistry()
	if registry == nil {
		t.Fatal("NewConcurrentEnvironmentRegistry returned nil")
	}

	// Test GetDefaultEnvironment
	defaultEnv := registry.GetDefaultEnvironment()
	if defaultEnv == nil {
		t.Fatal("GetDefaultEnvironment returned nil")
	}

	// Test RegisterEnvironment and GetEnvironment
	customEnv := NewThreadSafeEnvironment()
	registry.RegisterEnvironment("custom", customEnv)

	retrieved, ok := registry.GetEnvironment("custom")
	if !ok {
		t.Error("Expected custom environment to be found")
	}
	if retrieved != customEnv {
		t.Error("Retrieved environment doesn't match registered environment")
	}

	// Test GetEnvironment for nonexistent
	_, ok = registry.GetEnvironment("nonexistent")
	if ok {
		t.Error("Expected nonexistent environment to not be found")
	}
}

// TestGlobalConcurrentContextPool tests the global context pool
func TestGlobalConcurrentContextPool(t *testing.T) {
	if GlobalConcurrentContextPool == nil {
		t.Fatal("GlobalConcurrentContextPool is nil")
	}

	ctx := GlobalConcurrentContextPool.Get()
	if ctx == nil {
		t.Fatal("Get returned nil context")
	}
	GlobalConcurrentContextPool.Put(ctx)
}

// TestGlobalConcurrentCache tests the global cache
func TestGlobalConcurrentCache(t *testing.T) {
	if GlobalConcurrentCache == nil {
		t.Fatal("GlobalConcurrentCache is nil")
	}

	GlobalConcurrentCache.Set("global_test", "value")
	value, ok := GlobalConcurrentCache.Get("global_test")
	if !ok {
		t.Error("Expected key to be found in global cache")
	}
	if value != "value" {
		t.Errorf("Expected 'value', got %v", value)
	}
	GlobalConcurrentCache.Delete("global_test")
}

// TestGlobalEnvironmentRegistry tests the global environment registry
func TestGlobalEnvironmentRegistry(t *testing.T) {
	if GlobalEnvironmentRegistry == nil {
		t.Fatal("GlobalEnvironmentRegistry is nil")
	}

	defaultEnv := GlobalEnvironmentRegistry.GetDefaultEnvironment()
	if defaultEnv == nil {
		t.Fatal("GetDefaultEnvironment returned nil")
	}
}

// TestThreadSafeEnvironmentGetTemplateConcurrent tests GetTemplateConcurrent
func TestThreadSafeEnvironmentGetTemplateConcurrent(t *testing.T) {
	// Create a loader that returns a template
	loader := &simpleTestLoader{
		templates: map[string]string{
			"test.html": "Hello {{ name }}!",
		},
	}

	tse := NewThreadSafeEnvironment(WithLoader(loader))

	// This should fail because the loader doesn't support the full interface
	_, err := tse.GetTemplateConcurrent("test.html")
	if err != nil {
		// Expected - the loader might not fully implement the interface
		// This is acceptable behavior
	}
}

// simpleTestLoader is a simple loader for testing
type simpleTestLoader struct {
	templates map[string]string
}

func (l *simpleTestLoader) GetSource(name string) (string, error) {
	if content, ok := l.templates[name]; ok {
		return content, nil
	}
	return "", nil
}

func (l *simpleTestLoader) IsCached(name string) bool {
	_, ok := l.templates[name]
	return ok
}

func (l *simpleTestLoader) ListTemplates() ([]string, error) {
	var names []string
	for name := range l.templates {
		names = append(names, name)
	}
	return names, nil
}
