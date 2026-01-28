package loader

import (
	"fmt"
	"testing"
	"time"

	"github.com/zipreport/miya/parser"
)

func TestLRUCache(t *testing.T) {
	cache := NewLRUCache(3) // Small cache for testing

	// Create test template and source
	template := &parser.TemplateNode{
		Name:     "test.html",
		Children: []parser.Node{&parser.TextNode{Content: "Test content"}},
	}
	source := &TemplateSource{
		Name:    "test.html",
		Content: "Test content",
		ModTime: time.Now(),
	}

	t.Run("PutAndGet", func(t *testing.T) {
		cache.Put("test.html", template, source, 0)

		retrievedTemplate, retrievedSource, found := cache.Get("test.html")
		if !found {
			t.Error("Expected to find cached item")
		}

		if retrievedTemplate.Name != template.Name {
			t.Errorf("Expected template name '%s', got '%s'", template.Name, retrievedTemplate.Name)
		}

		if retrievedSource.Name != source.Name {
			t.Errorf("Expected source name '%s', got '%s'", source.Name, retrievedSource.Name)
		}
	})

	t.Run("GetNonExistent", func(t *testing.T) {
		_, _, found := cache.Get("nonexistent.html")
		if found {
			t.Error("Expected not to find non-existent item")
		}
	})

	t.Run("LRUEviction", func(t *testing.T) {
		cache.Clear()

		// Fill cache to capacity
		for i := 1; i <= 3; i++ {
			template := &parser.TemplateNode{Name: fmt.Sprintf("template%d.html", i)}
			source := &TemplateSource{Name: fmt.Sprintf("template%d.html", i)}
			cache.Put(fmt.Sprintf("template%d.html", i), template, source, 0)
		}

		// Verify all items are cached
		for i := 1; i <= 3; i++ {
			_, _, found := cache.Get(fmt.Sprintf("template%d.html", i))
			if !found {
				t.Errorf("Expected to find template%d.html", i)
			}
		}

		// Add one more item, should evict the least recently used (template1.html)
		template4 := &parser.TemplateNode{Name: "template4.html"}
		source4 := &TemplateSource{Name: "template4.html"}
		cache.Put("template4.html", template4, source4, 0)

		// template1.html should be evicted
		_, _, found := cache.Get("template1.html")
		if found {
			t.Error("Expected template1.html to be evicted")
		}

		// Others should still be there
		for i := 2; i <= 4; i++ {
			_, _, found := cache.Get(fmt.Sprintf("template%d.html", i))
			if !found {
				t.Errorf("Expected to find template%d.html", i)
			}
		}
	})

	t.Run("TTLExpiration", func(t *testing.T) {
		cache.Clear()

		// Add item with short TTL
		template := &parser.TemplateNode{Name: "expiring.html"}
		source := &TemplateSource{Name: "expiring.html"}
		cache.Put("expiring.html", template, source, 50*time.Millisecond)

		// Should be available immediately
		_, _, found := cache.Get("expiring.html")
		if !found {
			t.Error("Expected to find item before expiration")
		}

		// Wait for expiration
		time.Sleep(100 * time.Millisecond)

		// Should be expired now
		_, _, found = cache.Get("expiring.html")
		if found {
			t.Error("Expected item to be expired")
		}
	})

	t.Run("UpdateExisting", func(t *testing.T) {
		cache.Clear()

		// Add initial item
		template1 := &parser.TemplateNode{Name: "update.html"}
		source1 := &TemplateSource{Name: "update.html", Content: "Original content"}
		cache.Put("update.html", template1, source1, 0)

		// Update with new content
		template2 := &parser.TemplateNode{Name: "update.html"}
		source2 := &TemplateSource{Name: "update.html", Content: "Updated content"}
		cache.Put("update.html", template2, source2, 0)

		// Should get updated content
		_, retrievedSource, found := cache.Get("update.html")
		if !found {
			t.Error("Expected to find updated item")
		}

		if retrievedSource.Content != "Updated content" {
			t.Errorf("Expected updated content, got '%s'", retrievedSource.Content)
		}

		// Should still have only one item in cache
		if cache.Size() != 1 {
			t.Errorf("Expected cache size 1, got %d", cache.Size())
		}
	})

	t.Run("Remove", func(t *testing.T) {
		cache.Clear()

		template := &parser.TemplateNode{Name: "removeme.html"}
		source := &TemplateSource{Name: "removeme.html"}
		cache.Put("removeme.html", template, source, 0)

		// Verify it exists
		_, _, found := cache.Get("removeme.html")
		if !found {
			t.Error("Expected to find item before removal")
		}

		// Remove it
		removed := cache.Remove("removeme.html")
		if !removed {
			t.Error("Expected Remove to return true")
		}

		// Should not exist now
		_, _, found = cache.Get("removeme.html")
		if found {
			t.Error("Expected item to be removed")
		}

		// Removing non-existent item should return false
		removed = cache.Remove("nonexistent.html")
		if removed {
			t.Error("Expected Remove to return false for non-existent item")
		}
	})

	t.Run("Clear", func(t *testing.T) {
		// Add some items
		for i := 1; i <= 5; i++ {
			template := &parser.TemplateNode{Name: fmt.Sprintf("clear%d.html", i)}
			source := &TemplateSource{Name: fmt.Sprintf("clear%d.html", i)}
			cache.Put(fmt.Sprintf("clear%d.html", i), template, source, 0)
		}

		if cache.Size() == 0 {
			t.Error("Expected cache to have items before clear")
		}

		cache.Clear()

		if cache.Size() != 0 {
			t.Errorf("Expected cache size 0 after clear, got %d", cache.Size())
		}

		// Stats should be reset
		stats := cache.Stats()
		if stats.Hits != 0 || stats.Misses != 0 {
			t.Errorf("Expected stats to be reset after clear, got hits=%d, misses=%d", stats.Hits, stats.Misses)
		}
	})

	t.Run("Stats", func(t *testing.T) {
		cache.Clear()

		// Perform some operations to generate stats
		template := &parser.TemplateNode{Name: "stats.html"}
		source := &TemplateSource{Name: "stats.html"}
		cache.Put("stats.html", template, source, 0)

		// Cache hit
		cache.Get("stats.html")

		// Cache miss
		cache.Get("nonexistent.html")

		stats := cache.Stats()
		if stats.Hits != 1 {
			t.Errorf("Expected 1 hit, got %d", stats.Hits)
		}
		if stats.Misses != 1 {
			t.Errorf("Expected 1 miss, got %d", stats.Misses)
		}
		if stats.Size != 1 {
			t.Errorf("Expected size 1, got %d", stats.Size)
		}
	})

	t.Run("Keys", func(t *testing.T) {
		cache.Clear()

		expectedKeys := []string{"key1.html", "key2.html", "key3.html"}
		for _, key := range expectedKeys {
			template := &parser.TemplateNode{Name: key}
			source := &TemplateSource{Name: key}
			cache.Put(key, template, source, 0)
		}

		keys := cache.Keys()
		if len(keys) != len(expectedKeys) {
			t.Errorf("Expected %d keys, got %d", len(expectedKeys), len(keys))
		}

		keySet := make(map[string]bool)
		for _, key := range keys {
			keySet[key] = true
		}

		for _, expectedKey := range expectedKeys {
			if !keySet[expectedKey] {
				t.Errorf("Expected key '%s' not found", expectedKey)
			}
		}
	})
}

func TestLRUFileSystemLoader(t *testing.T) {
	templatesDir := createTestTemplates(t)
	parser := &MockParser{}

	// Create LRU filesystem loader
	loader := NewLRUFileSystemLoader([]string{templatesDir}, parser, 2, 5*time.Minute)

	t.Run("LoadAndCache", func(t *testing.T) {
		// Load template
		template, err := loader.LoadTemplate("base.html")
		if err != nil {
			t.Fatalf("Failed to load template: %v", err)
		}

		if template.Name != "base.html" {
			t.Errorf("Expected template name 'base.html', got '%s'", template.Name)
		}

		// Should be cached
		if !loader.IsCached("base.html") {
			t.Error("Template should be cached")
		}

		// Load again, should come from cache
		template2, err := loader.LoadTemplate("base.html")
		if err != nil {
			t.Fatalf("Failed to load cached template: %v", err)
		}

		if template2.Name != template.Name {
			t.Error("Cached template should be identical")
		}
	})

	t.Run("LRUEviction", func(t *testing.T) {
		loader.ClearCache()

		// Load templates to fill cache (max size is 2)
		loader.LoadTemplate("base.html")
		loader.LoadTemplate("page.html")

		// Both should be cached
		if !loader.IsCached("base.html") || !loader.IsCached("page.html") {
			t.Error("Both templates should be cached")
		}

		// Load third template, should evict least recently used
		loader.LoadTemplate("partial.jinja")

		// partial.jinja should be cached
		if !loader.IsCached("partial.jinja") {
			t.Error("partial.jinja should be cached")
		}

		// One of the previous templates should be evicted
		cached := 0
		if loader.IsCached("base.html") {
			cached++
		}
		if loader.IsCached("page.html") {
			cached++
		}

		if cached != 1 {
			t.Errorf("Expected exactly 1 of the first 2 templates to remain cached, got %d", cached)
		}
	})

	t.Run("CacheStats", func(t *testing.T) {
		loader.ClearCache()

		// Load template (cache miss)
		loader.LoadTemplate("base.html")

		// Load again (cache hit)
		loader.LoadTemplate("base.html")

		stats := loader.GetCacheStats()
		if stats.Hits < 1 {
			t.Errorf("Expected at least 1 cache hit, got %d", stats.Hits)
		}
		if stats.Size < 1 {
			t.Errorf("Expected cache size at least 1, got %d", stats.Size)
		}
	})

	t.Run("ClearCache", func(t *testing.T) {
		// Load some templates
		loader.LoadTemplate("base.html")
		loader.LoadTemplate("page.html")

		// Clear cache
		loader.ClearCache()

		stats := loader.GetCacheStats()
		if stats.Size != 0 {
			t.Errorf("Expected cache size 0 after clear, got %d", stats.Size)
		}
	})

	t.Run("GetCacheKeys", func(t *testing.T) {
		loader.ClearCache()

		// Load some templates
		loader.LoadTemplate("base.html")
		loader.LoadTemplate("page.html")

		keys := loader.GetCacheKeys()
		if len(keys) != 2 {
			t.Errorf("Expected 2 cache keys, got %d", len(keys))
		}
	})

	t.Run("SetCacheTTL", func(t *testing.T) {
		loader.SetCacheTTL(10 * time.Second)
		// Can't easily test the TTL change without waiting, but at least verify it doesn't crash
	})

	t.Run("EvictExpiredItems", func(t *testing.T) {
		loader.ClearCache()

		// Set very short TTL
		loader.SetCacheTTL(50 * time.Millisecond)

		// Load template
		loader.LoadTemplate("base.html")

		// Wait for expiration
		time.Sleep(100 * time.Millisecond)

		// Evict expired items
		evicted := loader.EvictExpiredItems()
		if evicted < 1 {
			t.Errorf("Expected at least 1 item to be evicted, got %d", evicted)
		}

		// Reset TTL
		loader.SetCacheTTL(5 * time.Minute)
	})
}

func TestCacheCleanup(t *testing.T) {
	cache := NewLRUCache(10)
	cleanup := NewCacheCleanup(cache, 100*time.Millisecond)

	t.Run("StartAndStop", func(t *testing.T) {
		// Start cleanup
		cleanup.Start()

		// Add an item with short TTL
		template := &parser.TemplateNode{Name: "cleanup.html"}
		source := &TemplateSource{Name: "cleanup.html"}
		cache.Put("cleanup.html", template, source, 150*time.Millisecond)

		// Wait for cleanup cycle
		time.Sleep(200 * time.Millisecond)

		// Item should be cleaned up
		_, _, found := cache.Get("cleanup.html")
		if found {
			t.Error("Expected expired item to be cleaned up")
		}

		// Stop cleanup
		cleanup.Stop()
	})

	t.Run("MultipleStartStop", func(t *testing.T) {
		// Multiple starts should not cause issues
		cleanup.Start()
		cleanup.Start()

		// Multiple stops should not cause issues
		cleanup.Stop()
		cleanup.Stop()
	})
}

func TestLRUCacheEdgeCases(t *testing.T) {
	t.Run("ZeroSizeCache", func(t *testing.T) {
		cache := NewLRUCache(0) // Should default to 100

		template := &parser.TemplateNode{Name: "test.html"}
		source := &TemplateSource{Name: "test.html"}
		cache.Put("test.html", template, source, 0)

		_, _, found := cache.Get("test.html")
		if !found {
			t.Error("Expected to find item in cache with default size")
		}
	})

	t.Run("NegativeSizeCache", func(t *testing.T) {
		cache := NewLRUCache(-10) // Should default to 100

		template := &parser.TemplateNode{Name: "test.html"}
		source := &TemplateSource{Name: "test.html"}
		cache.Put("test.html", template, source, 0)

		_, _, found := cache.Get("test.html")
		if !found {
			t.Error("Expected to find item in cache with default size")
		}
	})

	t.Run("SingleItemCache", func(t *testing.T) {
		cache := NewLRUCache(1)

		template1 := &parser.TemplateNode{Name: "test1.html"}
		source1 := &TemplateSource{Name: "test1.html"}
		cache.Put("test1.html", template1, source1, 0)

		template2 := &parser.TemplateNode{Name: "test2.html"}
		source2 := &TemplateSource{Name: "test2.html"}
		cache.Put("test2.html", template2, source2, 0)

		// First item should be evicted
		_, _, found := cache.Get("test1.html")
		if found {
			t.Error("Expected first item to be evicted")
		}

		// Second item should be present
		_, _, found = cache.Get("test2.html")
		if !found {
			t.Error("Expected second item to be present")
		}
	})

	t.Run("ConcurrentAccess", func(t *testing.T) {
		cache := NewLRUCache(100)

		// Test concurrent puts and gets
		done := make(chan bool, 10)

		for i := 0; i < 10; i++ {
			go func(id int) {
				for j := 0; j < 100; j++ {
					key := fmt.Sprintf("concurrent_%d_%d.html", id, j)
					template := &parser.TemplateNode{Name: key}
					source := &TemplateSource{Name: key}

					cache.Put(key, template, source, 0)
					cache.Get(key)
				}
				done <- true
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}

		// Cache should still be functional
		stats := cache.Stats()
		if stats.Size <= 0 {
			t.Error("Expected cache to have items after concurrent access")
		}
	})
}
