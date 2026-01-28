package runtime

import (
	"crypto/md5"
	"fmt"
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

	// Resolved template cache - stores final resolved templates
	resolvedCache map[string]*ResolvedTemplateCacheEntry
	resolvedMutex sync.RWMutex
	resolvedStats struct {
		hits   int64
		misses int64
	}

	// Configuration
	hierarchyTTL time.Duration // How long to keep hierarchy entries
	resolvedTTL  time.Duration // How long to keep resolved templates
	maxEntries   int           // Maximum cache entries

	// Cleanup management
	lastCleanup     time.Time
	cleanupInterval time.Duration
}

// NewInheritanceCache creates a new inheritance cache
func NewInheritanceCache() *InheritanceCache {
	return &InheritanceCache{
		hierarchyCache:  make(map[string]*InheritanceCacheEntry),
		resolvedCache:   make(map[string]*ResolvedTemplateCacheEntry),
		hierarchyTTL:    15 * time.Minute, // Hierarchies are relatively stable
		resolvedTTL:     5 * time.Minute,  // Resolved templates may vary by context
		maxEntries:      1000,             // Reasonable default
		lastCleanup:     time.Now(),
		cleanupInterval: 2 * time.Minute,
	}
}

// SetHierarchyTTL sets the time-to-live for hierarchy cache entries
func (c *InheritanceCache) SetHierarchyTTL(ttl time.Duration) {
	c.hierarchyTTL = ttl
}

// SetResolvedTTL sets the time-to-live for resolved template cache entries
func (c *InheritanceCache) SetResolvedTTL(ttl time.Duration) {
	c.resolvedTTL = ttl
}

// SetMaxEntries sets the maximum number of cache entries
func (c *InheritanceCache) SetMaxEntries(max int) {
	c.maxEntries = max
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
	if len(c.hierarchyCache) >= c.maxEntries {
		c.evictOldestHierarchy()
	}

	now := time.Now()
	entry := &InheritanceCacheEntry{
		Hierarchy:   hierarchy,
		CreatedAt:   now,
		ExpiresAt:   now.Add(c.hierarchyTTL),
		AccessCount: 1,
		LastAccess:  now,
	}

	c.hierarchyCache[templateName] = entry
}

// GetResolvedTemplate retrieves a cached resolved template
func (c *InheritanceCache) GetResolvedTemplate(templateName string, contextHash string) (*parser.TemplateNode, bool) {
	c.performCleanupIfNeeded()

	cacheKey := c.buildResolvedCacheKey(templateName, contextHash)

	c.resolvedMutex.RLock()
	entry, exists := c.resolvedCache[cacheKey]
	c.resolvedMutex.RUnlock()

	if !exists {
		atomic.AddInt64(&c.resolvedStats.misses, 1)
		return nil, false
	}

	// Check expiration
	if time.Now().After(entry.ExpiresAt) {
		// Entry expired, remove it
		c.resolvedMutex.Lock()
		delete(c.resolvedCache, cacheKey)
		c.resolvedMutex.Unlock()
		atomic.AddInt64(&c.resolvedStats.misses, 1)
		return nil, false
	}

	// Update access statistics
	c.resolvedMutex.Lock()
	entry.AccessCount++
	entry.LastAccess = time.Now()
	c.resolvedMutex.Unlock()

	atomic.AddInt64(&c.resolvedStats.hits, 1)
	return entry.ResolvedAST, true
}

// StoreResolvedTemplate caches a resolved template
func (c *InheritanceCache) StoreResolvedTemplate(templateName string, contextHash string, resolvedAST *parser.TemplateNode, templateChain []string) {
	c.resolvedMutex.Lock()
	defer c.resolvedMutex.Unlock()

	// Check if we need to make space
	if len(c.resolvedCache) >= c.maxEntries {
		c.evictOldestResolved()
	}

	cacheKey := c.buildResolvedCacheKey(templateName, contextHash)
	now := time.Now()

	entry := &ResolvedTemplateCacheEntry{
		ResolvedAST:   resolvedAST,
		ContextHash:   contextHash,
		TemplateChain: templateChain,
		CreatedAt:     now,
		ExpiresAt:     now.Add(c.resolvedTTL),
		AccessCount:   1,
		LastAccess:    now,
	}

	c.resolvedCache[cacheKey] = entry
}

// InvalidateTemplate removes all cache entries related to a template
func (c *InheritanceCache) InvalidateTemplate(templateName string) {
	c.hierarchyMutex.Lock()
	delete(c.hierarchyCache, templateName)
	c.hierarchyMutex.Unlock()

	// Invalidate resolved templates that include this template in their chain
	c.resolvedMutex.Lock()
	keysToDelete := make([]string, 0)
	for key, entry := range c.resolvedCache {
		for _, name := range entry.TemplateChain {
			if name == templateName {
				keysToDelete = append(keysToDelete, key)
				break
			}
		}
	}
	for _, key := range keysToDelete {
		delete(c.resolvedCache, key)
	}
	c.resolvedMutex.Unlock()
}

// ClearAll removes all cache entries
func (c *InheritanceCache) ClearAll() {
	c.hierarchyMutex.Lock()
	c.hierarchyCache = make(map[string]*InheritanceCacheEntry)
	c.hierarchyMutex.Unlock()

	c.resolvedMutex.Lock()
	c.resolvedCache = make(map[string]*ResolvedTemplateCacheEntry)
	c.resolvedMutex.Unlock()
}

// GetStats returns cache performance statistics
func (c *InheritanceCache) GetStats() CacheStats {
	c.hierarchyMutex.RLock()
	hierarchyEntries := len(c.hierarchyCache)
	c.hierarchyMutex.RUnlock()
	hierarchyHits := atomic.LoadInt64(&c.hierarchyStats.hits)
	hierarchyMisses := atomic.LoadInt64(&c.hierarchyStats.misses)

	c.resolvedMutex.RLock()
	resolvedEntries := len(c.resolvedCache)
	c.resolvedMutex.RUnlock()
	resolvedHits := atomic.LoadInt64(&c.resolvedStats.hits)
	resolvedMisses := atomic.LoadInt64(&c.resolvedStats.misses)

	stats := CacheStats{}

	// Hierarchy cache stats
	stats.HierarchyCache.Hits = hierarchyHits
	stats.HierarchyCache.Misses = hierarchyMisses
	stats.HierarchyCache.Entries = hierarchyEntries
	if hierarchyHits+hierarchyMisses > 0 {
		stats.HierarchyCache.HitRate = float64(hierarchyHits) / float64(hierarchyHits+hierarchyMisses)
	}

	// Resolved cache stats
	stats.ResolvedCache.Hits = resolvedHits
	stats.ResolvedCache.Misses = resolvedMisses
	stats.ResolvedCache.Entries = resolvedEntries
	if resolvedHits+resolvedMisses > 0 {
		stats.ResolvedCache.HitRate = float64(resolvedHits) / float64(resolvedHits+resolvedMisses)
	}

	return stats
}

// buildResolvedCacheKey creates a cache key for resolved templates
func (c *InheritanceCache) buildResolvedCacheKey(templateName string, contextHash string) string {
	return fmt.Sprintf("%s::%s", templateName, contextHash)
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

// evictOldestResolved removes the least recently used resolved template entry
func (c *InheritanceCache) evictOldestResolved() {
	var oldestKey string
	var oldestTime time.Time = time.Now()

	for key, entry := range c.resolvedCache {
		if entry.LastAccess.Before(oldestTime) {
			oldestTime = entry.LastAccess
			oldestKey = key
		}
	}

	if oldestKey != "" {
		delete(c.resolvedCache, oldestKey)
	}
}

// performCleanupIfNeeded performs periodic cleanup of expired entries
func (c *InheritanceCache) performCleanupIfNeeded() {
	if time.Since(c.lastCleanup) < c.cleanupInterval {
		return
	}

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

	// Cleanup resolved cache
	c.resolvedMutex.Lock()
	resolvedKeysToDelete := make([]string, 0)
	for key, entry := range c.resolvedCache {
		if now.After(entry.ExpiresAt) {
			resolvedKeysToDelete = append(resolvedKeysToDelete, key)
		}
	}
	for _, key := range resolvedKeysToDelete {
		delete(c.resolvedCache, key)
	}
	c.resolvedMutex.Unlock()

	c.lastCleanup = now
}

// ContextHasher creates context hashes for cache keys
type ContextHasher struct{}

// NewContextHasher creates a new context hasher
func NewContextHasher() *ContextHasher {
	return &ContextHasher{}
}

// HashContext creates a hash from context data for cache keying
func (h *ContextHasher) HashContext(context Context) string {
	if context == nil {
		return "empty"
	}

	// Get all context variables and create a deterministic hash
	data := context.All()
	if len(data) == 0 {
		return "empty"
	}

	// Create a simple hash based on key-value pairs
	// Note: This is a simplified approach - a production system might want
	// more sophisticated hashing that handles complex nested structures
	hash := md5.New()

	// Sort keys for consistent hashing
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}

	// Simple deterministic string representation
	for _, key := range keys {
		value := data[key]
		fmt.Fprintf(hash, "%s:%v;", key, value)
	}

	return fmt.Sprintf("%x", hash.Sum(nil))[:16] // Use first 16 characters
}
