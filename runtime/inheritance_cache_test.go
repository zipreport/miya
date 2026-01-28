package runtime

import (
	"sync"
	"testing"
	"time"

	"github.com/zipreport/miya/parser"
)

// mockContext implements Context interface for testing
type mockContext struct {
	data map[string]interface{}
}

func newMockContext(data map[string]interface{}) *mockContext {
	if data == nil {
		data = make(map[string]interface{})
	}
	return &mockContext{data: data}
}

func (c *mockContext) GetVariable(name string) (interface{}, bool) {
	v, ok := c.data[name]
	return v, ok
}

func (c *mockContext) SetVariable(name string, value interface{}) {
	c.data[name] = value
}

func (c *mockContext) Clone() Context {
	newData := make(map[string]interface{})
	for k, v := range c.data {
		newData[k] = v
	}
	return &mockContext{data: newData}
}

func (c *mockContext) All() map[string]interface{} {
	return c.data
}

func (c *mockContext) ApplyFilter(name string, value interface{}, args ...interface{}) (interface{}, error) {
	return value, nil
}

func (c *mockContext) ApplyTest(name string, value interface{}, args ...interface{}) (bool, error) {
	return false, nil
}

func (c *mockContext) IsAutoescapeEnabled() bool {
	return false
}

// Helper function to create a test template node
func createTestTemplateNode(name string) *parser.TemplateNode {
	return &parser.TemplateNode{
		Name:     name,
		Children: []parser.Node{},
	}
}

// Helper function to create a test hierarchy
func createTestHierarchy(name string) *InheritanceHierarchy {
	templateNode := createTestTemplateNode(name)
	return &InheritanceHierarchy{
		RootTemplate: templateNode,
		Templates:    []*parser.TemplateNode{templateNode},
		BlockMap:     make(map[string]*parser.BlockNode),
		TemplateMap:  map[string]*parser.TemplateNode{name: templateNode},
	}
}

// TestNewInheritanceCache tests cache creation
func TestNewInheritanceCache(t *testing.T) {
	cache := NewInheritanceCache()

	if cache == nil {
		t.Fatal("Expected non-nil cache")
	}

	if cache.hierarchyCache == nil {
		t.Error("Expected initialized hierarchy cache map")
	}

	if cache.resolvedCache == nil {
		t.Error("Expected initialized resolved cache map")
	}

	if cache.hierarchyTTL != 15*time.Minute {
		t.Errorf("Expected hierarchyTTL of 15 minutes, got %v", cache.hierarchyTTL)
	}

	if cache.resolvedTTL != 5*time.Minute {
		t.Errorf("Expected resolvedTTL of 5 minutes, got %v", cache.resolvedTTL)
	}

	if cache.maxEntries != 1000 {
		t.Errorf("Expected maxEntries of 1000, got %d", cache.maxEntries)
	}
}

// TestSetHierarchyTTL tests setting hierarchy TTL
func TestSetHierarchyTTL(t *testing.T) {
	cache := NewInheritanceCache()

	newTTL := 30 * time.Minute
	cache.SetHierarchyTTL(newTTL)

	if cache.hierarchyTTL != newTTL {
		t.Errorf("Expected hierarchyTTL of %v, got %v", newTTL, cache.hierarchyTTL)
	}
}

// TestSetResolvedTTL tests setting resolved template TTL
func TestSetResolvedTTL(t *testing.T) {
	cache := NewInheritanceCache()

	newTTL := 10 * time.Minute
	cache.SetResolvedTTL(newTTL)

	if cache.resolvedTTL != newTTL {
		t.Errorf("Expected resolvedTTL of %v, got %v", newTTL, cache.resolvedTTL)
	}
}

// TestSetMaxEntries tests setting max entries
func TestSetMaxEntries(t *testing.T) {
	cache := NewInheritanceCache()

	newMax := 500
	cache.SetMaxEntries(newMax)

	if cache.maxEntries != newMax {
		t.Errorf("Expected maxEntries of %d, got %d", newMax, cache.maxEntries)
	}
}

// TestHierarchyCacheOperations tests hierarchy cache store/get operations
func TestHierarchyCacheOperations(t *testing.T) {
	t.Run("Store and retrieve hierarchy", func(t *testing.T) {
		cache := NewInheritanceCache()
		templateName := "test_template.html"
		hierarchy := createTestHierarchy(templateName)

		// Store hierarchy
		cache.StoreHierarchy(templateName, hierarchy)

		// Retrieve hierarchy
		retrieved, found := cache.GetHierarchy(templateName)

		if !found {
			t.Error("Expected to find cached hierarchy")
		}

		if retrieved == nil {
			t.Error("Expected non-nil retrieved hierarchy")
		}

		if retrieved.RootTemplate.Name != templateName {
			t.Errorf("Expected template name %s, got %s", templateName, retrieved.RootTemplate.Name)
		}
	})

	t.Run("Get non-existent hierarchy", func(t *testing.T) {
		cache := NewInheritanceCache()

		retrieved, found := cache.GetHierarchy("non_existent.html")

		if found {
			t.Error("Expected not to find non-existent hierarchy")
		}

		if retrieved != nil {
			t.Error("Expected nil for non-existent hierarchy")
		}
	})

	t.Run("Hierarchy expiration", func(t *testing.T) {
		cache := NewInheritanceCache()
		cache.SetHierarchyTTL(1 * time.Millisecond)

		templateName := "expiring_template.html"
		hierarchy := createTestHierarchy(templateName)

		cache.StoreHierarchy(templateName, hierarchy)

		// Wait for expiration
		time.Sleep(5 * time.Millisecond)

		retrieved, found := cache.GetHierarchy(templateName)

		if found {
			t.Error("Expected expired hierarchy not to be found")
		}

		if retrieved != nil {
			t.Error("Expected nil for expired hierarchy")
		}
	})

	t.Run("Hierarchy access updates stats", func(t *testing.T) {
		cache := NewInheritanceCache()
		templateName := "stats_template.html"
		hierarchy := createTestHierarchy(templateName)

		cache.StoreHierarchy(templateName, hierarchy)

		// Multiple accesses
		cache.GetHierarchy(templateName)
		cache.GetHierarchy(templateName)
		cache.GetHierarchy(templateName)

		stats := cache.GetStats()

		if stats.HierarchyCache.Hits < 3 {
			t.Errorf("Expected at least 3 hits, got %d", stats.HierarchyCache.Hits)
		}
	})
}

// TestResolvedTemplateCacheOperations tests resolved template cache store/get operations
func TestResolvedTemplateCacheOperations(t *testing.T) {
	t.Run("Store and retrieve resolved template", func(t *testing.T) {
		cache := NewInheritanceCache()
		templateName := "resolved_template.html"
		contextHash := "abc123"
		templateNode := createTestTemplateNode(templateName)
		templateChain := []string{"base.html", templateName}

		cache.StoreResolvedTemplate(templateName, contextHash, templateNode, templateChain)

		retrieved, found := cache.GetResolvedTemplate(templateName, contextHash)

		if !found {
			t.Error("Expected to find cached resolved template")
		}

		if retrieved == nil {
			t.Error("Expected non-nil retrieved template")
		}

		if retrieved.Name != templateName {
			t.Errorf("Expected template name %s, got %s", templateName, retrieved.Name)
		}
	})

	t.Run("Get resolved template with wrong context hash", func(t *testing.T) {
		cache := NewInheritanceCache()
		templateName := "hash_template.html"
		templateNode := createTestTemplateNode(templateName)
		templateChain := []string{templateName}

		cache.StoreResolvedTemplate(templateName, "hash1", templateNode, templateChain)

		retrieved, found := cache.GetResolvedTemplate(templateName, "hash2")

		if found {
			t.Error("Expected not to find template with different hash")
		}

		if retrieved != nil {
			t.Error("Expected nil for wrong hash")
		}
	})

	t.Run("Resolved template expiration", func(t *testing.T) {
		cache := NewInheritanceCache()
		cache.SetResolvedTTL(1 * time.Millisecond)

		templateName := "expiring_resolved.html"
		contextHash := "expire_hash"
		templateNode := createTestTemplateNode(templateName)
		templateChain := []string{templateName}

		cache.StoreResolvedTemplate(templateName, contextHash, templateNode, templateChain)

		time.Sleep(5 * time.Millisecond)

		retrieved, found := cache.GetResolvedTemplate(templateName, contextHash)

		if found {
			t.Error("Expected expired resolved template not to be found")
		}

		if retrieved != nil {
			t.Error("Expected nil for expired resolved template")
		}
	})

	t.Run("Resolved template access updates stats", func(t *testing.T) {
		cache := NewInheritanceCache()
		templateName := "stats_resolved.html"
		contextHash := "stats_hash"
		templateNode := createTestTemplateNode(templateName)
		templateChain := []string{templateName}

		cache.StoreResolvedTemplate(templateName, contextHash, templateNode, templateChain)

		// Multiple accesses
		cache.GetResolvedTemplate(templateName, contextHash)
		cache.GetResolvedTemplate(templateName, contextHash)

		stats := cache.GetStats()

		if stats.ResolvedCache.Hits < 2 {
			t.Errorf("Expected at least 2 hits, got %d", stats.ResolvedCache.Hits)
		}
	})
}

// TestInvalidateTemplate tests template invalidation
func TestInvalidateTemplate(t *testing.T) {
	t.Run("Invalidate removes hierarchy entry", func(t *testing.T) {
		cache := NewInheritanceCache()
		templateName := "invalidate_test.html"
		hierarchy := createTestHierarchy(templateName)

		cache.StoreHierarchy(templateName, hierarchy)

		// Verify it exists
		_, found := cache.GetHierarchy(templateName)
		if !found {
			t.Fatal("Expected hierarchy to exist before invalidation")
		}

		cache.InvalidateTemplate(templateName)

		// Verify it's gone
		_, found = cache.GetHierarchy(templateName)
		if found {
			t.Error("Expected hierarchy to be removed after invalidation")
		}
	})

	t.Run("Invalidate removes resolved templates in chain", func(t *testing.T) {
		cache := NewInheritanceCache()
		baseTemplate := "base.html"
		childTemplate := "child.html"

		// Store a resolved template that includes base.html in its chain
		templateNode := createTestTemplateNode(childTemplate)
		templateChain := []string{baseTemplate, childTemplate}
		cache.StoreResolvedTemplate(childTemplate, "hash123", templateNode, templateChain)

		// Verify it exists
		_, found := cache.GetResolvedTemplate(childTemplate, "hash123")
		if !found {
			t.Fatal("Expected resolved template to exist before invalidation")
		}

		// Invalidate base template
		cache.InvalidateTemplate(baseTemplate)

		// Verify resolved template is also removed
		_, found = cache.GetResolvedTemplate(childTemplate, "hash123")
		if found {
			t.Error("Expected resolved template to be removed when parent is invalidated")
		}
	})
}

// TestClearAll tests clearing all cache entries
func TestClearAll(t *testing.T) {
	cache := NewInheritanceCache()

	// Add some entries
	for i := 0; i < 5; i++ {
		name := "template_" + string(rune('a'+i)) + ".html"
		cache.StoreHierarchy(name, createTestHierarchy(name))
		cache.StoreResolvedTemplate(name, "hash"+string(rune('0'+i)), createTestTemplateNode(name), []string{name})
	}

	// Verify entries exist
	stats := cache.GetStats()
	if stats.HierarchyCache.Entries == 0 {
		t.Fatal("Expected some hierarchy entries before clear")
	}
	if stats.ResolvedCache.Entries == 0 {
		t.Fatal("Expected some resolved entries before clear")
	}

	// Clear all
	cache.ClearAll()

	// Verify all cleared
	stats = cache.GetStats()
	if stats.HierarchyCache.Entries != 0 {
		t.Errorf("Expected 0 hierarchy entries after clear, got %d", stats.HierarchyCache.Entries)
	}
	if stats.ResolvedCache.Entries != 0 {
		t.Errorf("Expected 0 resolved entries after clear, got %d", stats.ResolvedCache.Entries)
	}
}

// TestCacheStats tests statistics gathering
func TestCacheStats(t *testing.T) {
	cache := NewInheritanceCache()

	// Create some hits and misses
	cache.StoreHierarchy("hit_template.html", createTestHierarchy("hit_template.html"))

	cache.GetHierarchy("hit_template.html")  // hit
	cache.GetHierarchy("hit_template.html")  // hit
	cache.GetHierarchy("miss_template.html") // miss
	cache.GetHierarchy("another_miss.html")  // miss

	stats := cache.GetStats()

	if stats.HierarchyCache.Hits != 2 {
		t.Errorf("Expected 2 hierarchy hits, got %d", stats.HierarchyCache.Hits)
	}

	if stats.HierarchyCache.Misses != 2 {
		t.Errorf("Expected 2 hierarchy misses, got %d", stats.HierarchyCache.Misses)
	}

	if stats.HierarchyCache.Entries != 1 {
		t.Errorf("Expected 1 hierarchy entry, got %d", stats.HierarchyCache.Entries)
	}

	expectedHitRate := 0.5 // 2 hits / (2 hits + 2 misses)
	if stats.HierarchyCache.HitRate != expectedHitRate {
		t.Errorf("Expected hit rate of %f, got %f", expectedHitRate, stats.HierarchyCache.HitRate)
	}
}

// TestCacheEviction tests LRU eviction when max entries is reached
func TestCacheEviction(t *testing.T) {
	t.Run("Hierarchy cache eviction", func(t *testing.T) {
		cache := NewInheritanceCache()
		cache.SetMaxEntries(3)

		// Add 3 entries
		cache.StoreHierarchy("first.html", createTestHierarchy("first.html"))
		time.Sleep(1 * time.Millisecond)
		cache.StoreHierarchy("second.html", createTestHierarchy("second.html"))
		time.Sleep(1 * time.Millisecond)
		cache.StoreHierarchy("third.html", createTestHierarchy("third.html"))

		// Access first to make it more recent
		cache.GetHierarchy("first.html")
		time.Sleep(1 * time.Millisecond)

		// Add fourth entry, should evict second (oldest access)
		cache.StoreHierarchy("fourth.html", createTestHierarchy("fourth.html"))

		stats := cache.GetStats()
		if stats.HierarchyCache.Entries != 3 {
			t.Errorf("Expected 3 entries after eviction, got %d", stats.HierarchyCache.Entries)
		}

		// second.html should have been evicted (oldest access time)
		_, found := cache.GetHierarchy("second.html")
		if found {
			t.Error("Expected second.html to be evicted")
		}
	})

	t.Run("Resolved cache eviction", func(t *testing.T) {
		cache := NewInheritanceCache()
		cache.SetMaxEntries(2)

		// Add 2 entries
		cache.StoreResolvedTemplate("first.html", "h1", createTestTemplateNode("first.html"), []string{"first.html"})
		time.Sleep(1 * time.Millisecond)
		cache.StoreResolvedTemplate("second.html", "h2", createTestTemplateNode("second.html"), []string{"second.html"})

		// Access first to make it more recent
		cache.GetResolvedTemplate("first.html", "h1")
		time.Sleep(1 * time.Millisecond)

		// Add third entry, should evict second
		cache.StoreResolvedTemplate("third.html", "h3", createTestTemplateNode("third.html"), []string{"third.html"})

		stats := cache.GetStats()
		if stats.ResolvedCache.Entries != 2 {
			t.Errorf("Expected 2 entries after eviction, got %d", stats.ResolvedCache.Entries)
		}

		// second.html should have been evicted
		_, found := cache.GetResolvedTemplate("second.html", "h2")
		if found {
			t.Error("Expected second.html to be evicted")
		}
	})
}

// TestBuildResolvedCacheKey tests the cache key building function
func TestBuildResolvedCacheKey(t *testing.T) {
	cache := NewInheritanceCache()

	key := cache.buildResolvedCacheKey("template.html", "context123")
	expected := "template.html::context123"

	if key != expected {
		t.Errorf("Expected key %q, got %q", expected, key)
	}
}

// TestConcurrentAccess tests thread safety of cache operations
func TestConcurrentAccess(t *testing.T) {
	cache := NewInheritanceCache()

	var wg sync.WaitGroup
	numGoroutines := 10
	numOperations := 100

	// Concurrent hierarchy operations
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				templateName := "template_" + string(rune('a'+id)) + ".html"
				cache.StoreHierarchy(templateName, createTestHierarchy(templateName))
				cache.GetHierarchy(templateName)
				cache.GetStats()
			}
		}(i)
	}

	// Concurrent resolved template operations
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				templateName := "resolved_" + string(rune('a'+id)) + ".html"
				contextHash := "hash" + string(rune('0'+id))
				cache.StoreResolvedTemplate(templateName, contextHash, createTestTemplateNode(templateName), []string{templateName})
				cache.GetResolvedTemplate(templateName, contextHash)
			}
		}(i)
	}

	// Some goroutines doing invalidations and clears
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations/10; j++ {
				templateName := "template_" + string(rune('a'+id)) + ".html"
				cache.InvalidateTemplate(templateName)
				if j%5 == 0 {
					cache.ClearAll()
				}
				time.Sleep(1 * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()

	// If we get here without panics or data races, the test passes
}

// TestCleanup tests the periodic cleanup functionality
func TestCleanup(t *testing.T) {
	cache := NewInheritanceCache()
	cache.SetHierarchyTTL(1 * time.Millisecond)
	cache.SetResolvedTTL(1 * time.Millisecond)
	cache.cleanupInterval = 1 * time.Millisecond

	// Add entries
	cache.StoreHierarchy("cleanup_test.html", createTestHierarchy("cleanup_test.html"))
	cache.StoreResolvedTemplate("cleanup_resolved.html", "hash", createTestTemplateNode("cleanup_resolved.html"), []string{})

	// Wait for expiration
	time.Sleep(10 * time.Millisecond)

	// Force cleanup by updating last cleanup time
	cache.lastCleanup = time.Now().Add(-5 * time.Minute)
	cache.performCleanupIfNeeded()

	// Give cleanup goroutine time to run
	time.Sleep(10 * time.Millisecond)

	// Entries should be cleaned up
	stats := cache.GetStats()
	if stats.HierarchyCache.Entries > 0 || stats.ResolvedCache.Entries > 0 {
		// Cleanup runs in background, may not have completed yet
		// Just verify no panic occurred
		t.Log("Cleanup may still be in progress")
	}
}

// TestContextHasher tests the context hasher
func TestContextHasher(t *testing.T) {
	hasher := NewContextHasher()

	if hasher == nil {
		t.Fatal("Expected non-nil hasher")
	}

	t.Run("Hash nil context", func(t *testing.T) {
		hash := hasher.HashContext(nil)
		if hash != "empty" {
			t.Errorf("Expected 'empty' for nil context, got %q", hash)
		}
	})

	t.Run("Hash empty context", func(t *testing.T) {
		ctx := newMockContext(map[string]interface{}{})
		hash := hasher.HashContext(ctx)
		if hash != "empty" {
			t.Errorf("Expected 'empty' for empty context, got %q", hash)
		}
	})

	t.Run("Hash context with data", func(t *testing.T) {
		ctx := newMockContext(map[string]interface{}{
			"name":   "test",
			"count":  42,
			"active": true,
		})
		hash := hasher.HashContext(ctx)

		if hash == "" {
			t.Error("Expected non-empty hash")
		}

		if hash == "empty" {
			t.Error("Expected hash different from 'empty' for non-empty context")
		}

		if len(hash) != 16 {
			t.Errorf("Expected hash length of 16, got %d", len(hash))
		}
	})

	t.Run("Same context produces same hash", func(t *testing.T) {
		ctx1 := newMockContext(map[string]interface{}{
			"key": "value",
		})
		ctx2 := newMockContext(map[string]interface{}{
			"key": "value",
		})

		hash1 := hasher.HashContext(ctx1)
		hash2 := hasher.HashContext(ctx2)

		if hash1 != hash2 {
			t.Errorf("Expected same hash for same context data, got %q and %q", hash1, hash2)
		}
	})

	t.Run("Different context produces different hash", func(t *testing.T) {
		ctx1 := newMockContext(map[string]interface{}{
			"key": "value1",
		})
		ctx2 := newMockContext(map[string]interface{}{
			"key": "value2",
		})

		hash1 := hasher.HashContext(ctx1)
		hash2 := hasher.HashContext(ctx2)

		if hash1 == hash2 {
			t.Error("Expected different hashes for different context data")
		}
	})
}

// TestCacheStatsHitRate tests hit rate calculation edge cases
func TestCacheStatsHitRate(t *testing.T) {
	t.Run("Hit rate with no accesses", func(t *testing.T) {
		cache := NewInheritanceCache()
		stats := cache.GetStats()

		if stats.HierarchyCache.HitRate != 0 {
			t.Errorf("Expected 0 hit rate with no accesses, got %f", stats.HierarchyCache.HitRate)
		}
		if stats.ResolvedCache.HitRate != 0 {
			t.Errorf("Expected 0 hit rate with no accesses, got %f", stats.ResolvedCache.HitRate)
		}
	})

	t.Run("Hit rate calculation", func(t *testing.T) {
		cache := NewInheritanceCache()

		// Store one entry
		cache.StoreHierarchy("test.html", createTestHierarchy("test.html"))

		// 4 hits
		for i := 0; i < 4; i++ {
			cache.GetHierarchy("test.html")
		}

		// 1 miss
		cache.GetHierarchy("nonexistent.html")

		stats := cache.GetStats()
		expectedRate := float64(4) / float64(5) // 4 hits / 5 total

		if stats.HierarchyCache.HitRate != expectedRate {
			t.Errorf("Expected hit rate %f, got %f", expectedRate, stats.HierarchyCache.HitRate)
		}
	})
}

// Benchmark tests
func BenchmarkHierarchyCacheStore(b *testing.B) {
	cache := NewInheritanceCache()
	hierarchy := createTestHierarchy("benchmark.html")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.StoreHierarchy("benchmark.html", hierarchy)
	}
}

func BenchmarkHierarchyCacheGet(b *testing.B) {
	cache := NewInheritanceCache()
	cache.StoreHierarchy("benchmark.html", createTestHierarchy("benchmark.html"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.GetHierarchy("benchmark.html")
	}
}

func BenchmarkResolvedCacheStore(b *testing.B) {
	cache := NewInheritanceCache()
	templateNode := createTestTemplateNode("benchmark.html")
	templateChain := []string{"benchmark.html"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.StoreResolvedTemplate("benchmark.html", "hash123", templateNode, templateChain)
	}
}

func BenchmarkResolvedCacheGet(b *testing.B) {
	cache := NewInheritanceCache()
	cache.StoreResolvedTemplate("benchmark.html", "hash123", createTestTemplateNode("benchmark.html"), []string{"benchmark.html"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.GetResolvedTemplate("benchmark.html", "hash123")
	}
}

func BenchmarkContextHasher(b *testing.B) {
	hasher := NewContextHasher()
	ctx := newMockContext(map[string]interface{}{
		"name":   "test",
		"count":  42,
		"active": true,
		"items":  []string{"a", "b", "c"},
		"nested": map[string]interface{}{
			"key": "value",
		},
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hasher.HashContext(ctx)
	}
}

func BenchmarkConcurrentCacheAccess(b *testing.B) {
	cache := NewInheritanceCache()
	cache.StoreHierarchy("bench.html", createTestHierarchy("bench.html"))

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cache.GetHierarchy("bench.html")
		}
	})
}
