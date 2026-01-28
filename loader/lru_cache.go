package loader

import (
	"sync"
	"time"

	"github.com/zipreport/miya/parser"
)

// LRUCache implements a Least Recently Used cache for templates
type LRUCache struct {
	maxSize int
	items   map[string]*lruItem
	head    *lruItem
	tail    *lruItem
	mutex   sync.RWMutex
	hits    int64
	misses  int64
}

// lruItem represents an item in the LRU cache
type lruItem struct {
	key      string
	template *parser.TemplateNode
	source   *TemplateSource
	expires  time.Time
	prev     *lruItem
	next     *lruItem
}

// NewLRUCache creates a new LRU cache with the specified maximum size
func NewLRUCache(maxSize int) *LRUCache {
	if maxSize <= 0 {
		maxSize = 100 // Default size
	}

	cache := &LRUCache{
		maxSize: maxSize,
		items:   make(map[string]*lruItem),
	}

	// Create sentinel nodes for head and tail
	cache.head = &lruItem{}
	cache.tail = &lruItem{}
	cache.head.next = cache.tail
	cache.tail.prev = cache.head

	return cache
}

// Get retrieves a template from the cache
func (c *LRUCache) Get(key string) (*parser.TemplateNode, *TemplateSource, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	item, exists := c.items[key]
	if !exists {
		c.misses++
		return nil, nil, false
	}

	// Check if item has expired
	if !item.expires.IsZero() && time.Now().After(item.expires) {
		c.remove(item)
		c.misses++
		return nil, nil, false
	}

	// Move to front (most recently used)
	c.moveToFront(item)
	c.hits++
	return item.template, item.source, true
}

// Put adds or updates a template in the cache
func (c *LRUCache) Put(key string, template *parser.TemplateNode, source *TemplateSource, ttl time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Check if item already exists
	if existing, exists := c.items[key]; exists {
		// Update existing item
		existing.template = template
		existing.source = source
		if ttl > 0 {
			existing.expires = time.Now().Add(ttl)
		} else {
			existing.expires = time.Time{}
		}
		c.moveToFront(existing)
		return
	}

	// Create new item
	item := &lruItem{
		key:      key,
		template: template,
		source:   source,
	}

	if ttl > 0 {
		item.expires = time.Now().Add(ttl)
	}

	// Add to cache
	c.items[key] = item
	c.addToFront(item)

	// Check if we need to evict items
	if len(c.items) > c.maxSize {
		c.evictLRU()
	}
}

// Remove removes a specific key from the cache
func (c *LRUCache) Remove(key string) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if item, exists := c.items[key]; exists {
		c.remove(item)
		return true
	}
	return false
}

// Clear removes all items from the cache
func (c *LRUCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.items = make(map[string]*lruItem)
	c.head.next = c.tail
	c.tail.prev = c.head
	c.hits = 0
	c.misses = 0
}

// Size returns the current number of items in the cache
func (c *LRUCache) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return len(c.items)
}

// Stats returns cache statistics
func (c *LRUCache) Stats() CacheStats {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return CacheStats{
		Hits:   c.hits,
		Misses: c.misses,
		Size:   len(c.items),
	}
}

// Keys returns all keys currently in the cache
func (c *LRUCache) Keys() []string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	keys := make([]string, 0, len(c.items))
	for key := range c.items {
		keys = append(keys, key)
	}
	return keys
}

// remove removes an item from the cache (internal method, assumes lock is held)
func (c *LRUCache) remove(item *lruItem) {
	delete(c.items, item.key)
	c.removeFromList(item)
}

// addToFront adds an item to the front of the list (internal method, assumes lock is held)
func (c *LRUCache) addToFront(item *lruItem) {
	item.prev = c.head
	item.next = c.head.next
	c.head.next.prev = item
	c.head.next = item
}

// removeFromList removes an item from the doubly linked list (internal method, assumes lock is held)
func (c *LRUCache) removeFromList(item *lruItem) {
	item.prev.next = item.next
	item.next.prev = item.prev
}

// moveToFront moves an item to the front of the list (internal method, assumes lock is held)
func (c *LRUCache) moveToFront(item *lruItem) {
	c.removeFromList(item)
	c.addToFront(item)
}

// evictLRU removes the least recently used item (internal method, assumes lock is held)
func (c *LRUCache) evictLRU() {
	if len(c.items) == 0 {
		return
	}

	// Remove the item at the tail (least recently used)
	lru := c.tail.prev
	if lru != c.head {
		c.remove(lru)
	}
}

// LRUFileSystemLoader extends FileSystemLoader with LRU caching
type LRUFileSystemLoader struct {
	*FileSystemLoader
	lruCache *LRUCache
	cacheTTL time.Duration
}

// NewLRUFileSystemLoader creates a filesystem loader with LRU caching
func NewLRUFileSystemLoader(searchPaths []string, parser TemplateParser, maxCacheSize int, cacheTTL time.Duration) *LRUFileSystemLoader {
	fsLoader := NewFileSystemLoader(searchPaths, parser)

	return &LRUFileSystemLoader{
		FileSystemLoader: fsLoader,
		lruCache:         NewLRUCache(maxCacheSize),
		cacheTTL:         cacheTTL,
	}
}

// LoadTemplate loads and parses a template with LRU caching
func (l *LRUFileSystemLoader) LoadTemplate(name string) (*parser.TemplateNode, error) {
	// Try to get from LRU cache first
	if template, _, found := l.lruCache.Get(name); found {
		return template, nil
	}

	// Load from filesystem
	source, err := l.GetSourceWithMetadata(name)
	if err != nil {
		return nil, err
	}

	template, err := l.parser.ParseTemplate(name, source.Content)
	if err != nil {
		return nil, err
	}

	// Cache the template
	l.lruCache.Put(name, template, source, l.cacheTTL)

	return template, nil
}

// IsCached checks if a template is cached in the LRU cache
func (l *LRUFileSystemLoader) IsCached(name string) bool {
	_, _, found := l.lruCache.Get(name)
	return found
}

// ClearCache clears the LRU cache
func (l *LRUFileSystemLoader) ClearCache() {
	l.lruCache.Clear()
	l.FileSystemLoader.ClearCache() // Also clear the parent cache
}

// GetCacheStats returns LRU cache statistics
func (l *LRUFileSystemLoader) GetCacheStats() CacheStats {
	return l.lruCache.Stats()
}

// SetCacheTTL sets the time-to-live for cached templates
func (l *LRUFileSystemLoader) SetCacheTTL(ttl time.Duration) {
	l.cacheTTL = ttl
}

// GetCacheKeys returns all keys currently in the LRU cache
func (l *LRUFileSystemLoader) GetCacheKeys() []string {
	return l.lruCache.Keys()
}

// EvictExpiredItems manually removes expired items from the cache
func (l *LRUFileSystemLoader) EvictExpiredItems() int {
	keys := l.lruCache.Keys()
	evicted := 0

	for _, key := range keys {
		// Try to get the item, which will automatically remove it if expired
		if _, _, found := l.lruCache.Get(key); !found {
			evicted++
		}
	}

	return evicted
}

// LRU Cache cleanup helper
type CacheCleanup struct {
	cache    *LRUCache
	interval time.Duration
	stopCh   chan struct{}
	stopped  bool
	mutex    sync.Mutex
}

// NewCacheCleanup creates a new cache cleanup helper
func NewCacheCleanup(cache *LRUCache, interval time.Duration) *CacheCleanup {
	return &CacheCleanup{
		cache:    cache,
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

// Start begins the cleanup routine
func (cc *CacheCleanup) Start() {
	cc.mutex.Lock()
	if cc.stopped {
		cc.mutex.Unlock()
		return
	}
	cc.mutex.Unlock()

	go func() {
		ticker := time.NewTicker(cc.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				cc.cleanup()
			case <-cc.stopCh:
				return
			}
		}
	}()
}

// Stop stops the cleanup routine
func (cc *CacheCleanup) Stop() {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()

	if !cc.stopped {
		close(cc.stopCh)
		cc.stopped = true
	}
}

// cleanup removes expired items from the cache
func (cc *CacheCleanup) cleanup() {
	keys := cc.cache.Keys()

	for _, key := range keys {
		// Getting an expired item will automatically remove it
		cc.cache.Get(key)
	}
}
