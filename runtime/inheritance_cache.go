package runtime

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/zipreport/miya/parser"
)

// InheritanceCacheEntry represents a cached inheritance hierarchy
type InheritanceCacheEntry struct {
	Hierarchy   *InheritanceHierarchy
	CreatedAt   time.Time
	ExpiresAt   time.Time
	AccessCount int64
	LastAccess  time.Time
}

// ResolvedTemplateCacheEntry represents a cached resolved template
//
// Deprecated: resolved templates are not cached; nothing uses this type.
type ResolvedTemplateCacheEntry struct {
	ResolvedAST   *parser.TemplateNode
	ContextHash   string
	TemplateChain []string // Template names in hierarchy
	CreatedAt     time.Time
	ExpiresAt     time.Time
	AccessCount   int64
	LastAccess    time.Time
}

// CacheStats provides cache performance metrics
type CacheStats struct {
	HierarchyCache struct {
		Hits      int64
		Misses    int64
		Entries   int
		TotalSize int64 // Estimated memory usage
		HitRate   float64
		AvgAccess float64
	}
	// Deprecated: resolved templates are not cached; these stats are always zero.
	ResolvedCache struct {
		Hits      int64
		Misses    int64
		Entries   int
		TotalSize int64
		HitRate   float64
		AvgAccess float64
	}
}

// InheritanceCache manages caching for template inheritance
type InheritanceCache struct {
	// Hierarchy cache - stores template inheritance chains
	hierarchyCache map[string]*InheritanceCacheEntry
	hierarchyMutex sync.RWMutex
	hierarchyStats struct {
		hits   int64
		misses int64
	}

	// Configuration
	hierarchyTTL atomic.Int64 // How long to keep hierarchy entries (time.Duration)
	maxEntries   atomic.Int64 // Maximum cache entries

	// Cleanup management
	cleanupMu       sync.Mutex
	lastCleanup     time.Time
	cleanupInterval time.Duration
}

// NewInheritanceCache creates a new inheritance cache
func NewInheritanceCache() *InheritanceCache {
	c := &InheritanceCache{
		hierarchyCache:  make(map[string]*InheritanceCacheEntry),
		lastCleanup:     time.Now(),
		cleanupInterval: 2 * time.Minute,
	}
	c.hierarchyTTL.Store(int64(15 * time.Minute)) // Hierarchies are relatively stable
	c.maxEntries.Store(1000)                      // Reasonable default
	return c
}

// SetHierarchyTTL sets the time-to-live for hierarchy cache entries
func (c *InheritanceCache) SetHierarchyTTL(ttl time.Duration) {
	c.hierarchyTTL.Store(int64(ttl))
}

// SetResolvedTTL does nothing.
//
// Deprecated: resolved templates are not cached.
func (c *InheritanceCache) SetResolvedTTL(ttl time.Duration) {}

// SetMaxEntries sets the maximum number of cache entries
func (c *InheritanceCache) SetMaxEntries(max int) {
	c.maxEntries.Store(int64(max))
}

// GetHierarchy retrieves a cached template hierarchy
func (c *InheritanceCache) GetHierarchy(templateName string) (*InheritanceHierarchy, bool) {
	c.performCleanupIfNeeded()

	c.hierarchyMutex.RLock()
	entry, exists := c.hierarchyCache[templateName]
	c.hierarchyMutex.RUnlock()

	if !exists {
		atomic.AddInt64(&c.hierarchyStats.misses, 1)
		return nil, false
	}

	// Check expiration
	if time.Now().After(entry.ExpiresAt) {
		// Entry expired, remove it
		c.hierarchyMutex.Lock()
		delete(c.hierarchyCache, templateName)
		c.hierarchyMutex.Unlock()
		atomic.AddInt64(&c.hierarchyStats.misses, 1)
		return nil, false
	}

	// Update access statistics
	c.hierarchyMutex.Lock()
	entry.AccessCount++
	entry.LastAccess = time.Now()
	c.hierarchyMutex.Unlock()

	atomic.AddInt64(&c.hierarchyStats.hits, 1)
	return entry.Hierarchy, true
}

// StoreHierarchy caches a template hierarchy
func (c *InheritanceCache) StoreHierarchy(templateName string, hierarchy *InheritanceHierarchy) {
	c.hierarchyMutex.Lock()
	defer c.hierarchyMutex.Unlock()

	// Check if we need to make space
	if len(c.hierarchyCache) >= int(c.maxEntries.Load()) {
		c.evictOldestHierarchy()
	}

	now := time.Now()
	entry := &InheritanceCacheEntry{
		Hierarchy:   hierarchy,
		CreatedAt:   now,
		ExpiresAt:   now.Add(time.Duration(c.hierarchyTTL.Load())),
		AccessCount: 1,
		LastAccess:  now,
	}

	c.hierarchyCache[templateName] = entry
}

// InvalidateTemplate removes all cache entries related to a template: its own
// hierarchy and the hierarchy of every template whose chain contains it.
func (c *InheritanceCache) InvalidateTemplate(templateName string) {
	c.hierarchyMutex.Lock()
	for key, entry := range c.hierarchyCache {
		if key == templateName {
			delete(c.hierarchyCache, key)
		} else if entry.Hierarchy != nil {
			if _, inChain := entry.Hierarchy.TemplateMap[templateName]; inChain {
				delete(c.hierarchyCache, key)
			}
		}
	}
	c.hierarchyMutex.Unlock()
}

// ClearAll removes all cache entries
func (c *InheritanceCache) ClearAll() {
	c.hierarchyMutex.Lock()
	c.hierarchyCache = make(map[string]*InheritanceCacheEntry)
	c.hierarchyMutex.Unlock()
}

// GetStats returns cache performance statistics
func (c *InheritanceCache) GetStats() CacheStats {
	c.hierarchyMutex.RLock()
	hierarchyEntries := len(c.hierarchyCache)
	c.hierarchyMutex.RUnlock()
	hierarchyHits := atomic.LoadInt64(&c.hierarchyStats.hits)
	hierarchyMisses := atomic.LoadInt64(&c.hierarchyStats.misses)

	stats := CacheStats{}

	// Hierarchy cache stats
	stats.HierarchyCache.Hits = hierarchyHits
	stats.HierarchyCache.Misses = hierarchyMisses
	stats.HierarchyCache.Entries = hierarchyEntries
	if hierarchyHits+hierarchyMisses > 0 {
		stats.HierarchyCache.HitRate = float64(hierarchyHits) / float64(hierarchyHits+hierarchyMisses)
	}

	return stats
}

// evictOldestHierarchy removes the least recently used hierarchy entry
func (c *InheritanceCache) evictOldestHierarchy() {
	var oldestKey string
	var oldestTime time.Time = time.Now()

	for key, entry := range c.hierarchyCache {
		if entry.LastAccess.Before(oldestTime) {
			oldestTime = entry.LastAccess
			oldestKey = key
		}
	}

	if oldestKey != "" {
		delete(c.hierarchyCache, oldestKey)
	}
}

// performCleanupIfNeeded performs periodic cleanup of expired entries
func (c *InheritanceCache) performCleanupIfNeeded() {
	c.cleanupMu.Lock()
	if time.Since(c.lastCleanup) < c.cleanupInterval {
		c.cleanupMu.Unlock()
		return
	}
	c.lastCleanup = time.Now()
	c.cleanupMu.Unlock()

	go c.cleanup() // Run cleanup in background
}

// cleanup removes expired cache entries
func (c *InheritanceCache) cleanup() {
	now := time.Now()

	// Cleanup hierarchy cache
	c.hierarchyMutex.Lock()
	hierarchyKeysToDelete := make([]string, 0)
	for key, entry := range c.hierarchyCache {
		if now.After(entry.ExpiresAt) {
			hierarchyKeysToDelete = append(hierarchyKeysToDelete, key)
		}
	}
	for _, key := range hierarchyKeysToDelete {
		delete(c.hierarchyCache, key)
	}
	c.hierarchyMutex.Unlock()
}
