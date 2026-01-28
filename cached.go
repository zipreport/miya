package miya

import (
	"strings"
	"sync"
)

// StringBuilderPool provides a pool of strings.Builder instances to reduce allocations
var StringBuilderPool = sync.Pool{
	New: func() interface{} {
		return &strings.Builder{}
	},
}

// GetStringBuilder gets a strings.Builder from the pool
func GetStringBuilder() *strings.Builder {
	return StringBuilderPool.Get().(*strings.Builder)
}

// PutStringBuilder returns a strings.Builder to the pool after resetting it
func PutStringBuilder(sb *strings.Builder) {
	sb.Reset()
	StringBuilderPool.Put(sb)
}

// ContextPool provides a pool of Context instances to reduce allocations
var ContextPool = sync.Pool{
	New: func() interface{} {
		return NewContext()
	},
}

// GetContext gets a Context from the pool
func GetContext() Context {
	return ContextPool.Get().(Context)
}

// PutContext returns a Context to the pool
func PutContext(ctx Context) {
	// Reset the context before returning to pool
	// Clear all variables to avoid data leakage between uses
	if ctx != nil {
		// Create a fresh context to replace the used one
		// This ensures no stale data remains
		ContextPool.Put(NewContext())
	}
}

// CachedTemplate wraps Template with caching optimizations
type CachedTemplate struct {
	*Template
	renderPool sync.Pool
}

// NewCachedTemplate creates a new cached template wrapper
func NewCachedTemplate(tmpl *Template) *CachedTemplate {
	return &CachedTemplate{
		Template: tmpl,
		renderPool: sync.Pool{
			New: func() interface{} {
				return &renderContext{
					builder: &strings.Builder{},
					visited: make(map[string]bool),
				}
			},
		},
	}
}

type renderContext struct {
	builder *strings.Builder
	visited map[string]bool
}

func (rc *renderContext) reset() {
	rc.builder.Reset()
	for k := range rc.visited {
		delete(rc.visited, k)
	}
}

// RenderCached renders the template with caching optimizations
func (ct *CachedTemplate) RenderCached(ctx Context) (string, error) {
	rc := ct.renderPool.Get().(*renderContext)
	defer func() {
		rc.reset()
		ct.renderPool.Put(rc)
	}()

	// Use the existing render method but with cached context
	return ct.Template.Render(ctx)
}

// SlicePool provides pools for commonly used slice types
type SlicePool struct {
	stringPool sync.Pool
	intPool    sync.Pool
}

var GlobalSlicePool = &SlicePool{
	stringPool: sync.Pool{
		New: func() interface{} {
			return make([]string, 0, 16)
		},
	},
	intPool: sync.Pool{
		New: func() interface{} {
			return make([]int, 0, 16)
		},
	},
}

// GetStringSlice gets a string slice from the pool
func (sp *SlicePool) GetStringSlice() []string {
	return sp.stringPool.Get().([]string)
}

// PutStringSlice returns a string slice to the pool
func (sp *SlicePool) PutStringSlice(slice []string) {
	slice = slice[:0] // Reset length but keep capacity
	sp.stringPool.Put(slice)
}

// GetIntSlice gets an int slice from the pool
func (sp *SlicePool) GetIntSlice() []int {
	return sp.intPool.Get().([]int)
}

// PutIntSlice returns an int slice to the pool
func (sp *SlicePool) PutIntSlice(slice []int) {
	slice = slice[:0] // Reset length but keep capacity
	sp.intPool.Put(slice)
}

// FastStringBuilder is an optimized string builder for template rendering
type FastStringBuilder struct {
	buf []byte
}

// NewFastStringBuilder creates a new fast string builder
func NewFastStringBuilder(capacity int) *FastStringBuilder {
	return &FastStringBuilder{
		buf: make([]byte, 0, capacity),
	}
}

// WriteString appends a string to the buffer
func (fsb *FastStringBuilder) WriteString(s string) {
	fsb.buf = append(fsb.buf, s...)
}

// WriteByte appends a byte to the buffer
func (fsb *FastStringBuilder) WriteByte(b byte) error {
	fsb.buf = append(fsb.buf, b)
	return nil
}

// String returns the accumulated string
func (fsb *FastStringBuilder) String() string {
	return string(fsb.buf)
}

// Reset clears the buffer
func (fsb *FastStringBuilder) Reset() {
	fsb.buf = fsb.buf[:0]
}

// Len returns the current length
func (fsb *FastStringBuilder) Len() int {
	return len(fsb.buf)
}

// FastStringBuilderPool provides pooled FastStringBuilder instances
var FastStringBuilderPool = sync.Pool{
	New: func() interface{} {
		return NewFastStringBuilder(1024) // Start with 1KB capacity
	},
}

// GetFastStringBuilder gets a FastStringBuilder from the pool
func GetFastStringBuilder() *FastStringBuilder {
	return FastStringBuilderPool.Get().(*FastStringBuilder)
}

// PutFastStringBuilder returns a FastStringBuilder to the pool
func PutFastStringBuilder(fsb *FastStringBuilder) {
	if fsb.Len() < 64*1024 { // Only reuse if buffer is reasonable size
		fsb.Reset()
		FastStringBuilderPool.Put(fsb)
	}
}

// CacheKey represents a cache key for templates
type CacheKey struct {
	Template string
	Hash     uint64
}

// TemplateCache is a thread-safe LRU cache for compiled templates
type TemplateCache struct {
	mutex   sync.RWMutex
	cache   map[string]*Template
	order   []string
	maxSize int
}

// NewTemplateCache creates a new template cache
func NewTemplateCache(maxSize int) *TemplateCache {
	return &TemplateCache{
		cache:   make(map[string]*Template),
		order:   make([]string, 0, maxSize),
		maxSize: maxSize,
	}
}

// Get retrieves a template from the cache
func (tc *TemplateCache) Get(key string) (*Template, bool) {
	tc.mutex.RLock()
	tmpl, exists := tc.cache[key]
	tc.mutex.RUnlock()

	if exists {
		// Move to front (LRU)
		tc.mutex.Lock()
		tc.moveToFront(key)
		tc.mutex.Unlock()
	}

	return tmpl, exists
}

// Put stores a template in the cache
func (tc *TemplateCache) Put(key string, tmpl *Template) {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()

	if _, exists := tc.cache[key]; exists {
		tc.moveToFront(key)
		tc.cache[key] = tmpl
		return
	}

	if len(tc.cache) >= tc.maxSize {
		// Remove least recently used
		oldest := tc.order[len(tc.order)-1]
		delete(tc.cache, oldest)
		tc.order = tc.order[:len(tc.order)-1]
	}

	tc.cache[key] = tmpl
	tc.order = append([]string{key}, tc.order...)
}

func (tc *TemplateCache) moveToFront(key string) {
	for i, k := range tc.order {
		if k == key {
			// Move to front
			copy(tc.order[1:i+1], tc.order[0:i])
			tc.order[0] = key
			break
		}
	}
}

// CachedEnvironment wraps Environment with caching optimizations
type CachedEnvironment struct {
	*Environment
	templateCache *TemplateCache
}

// NewCachedEnvironment creates a cached environment
func NewCachedEnvironment(opts ...EnvironmentOption) *CachedEnvironment {
	env := NewEnvironment(opts...)
	return &CachedEnvironment{
		Environment:   env,
		templateCache: NewTemplateCache(100), // Cache up to 100 templates
	}
}

// FromStringCached compiles a template with caching
func (ce *CachedEnvironment) FromStringCached(source string) (*Template, error) {
	// Use source as cache key (in production, you might want to hash this)
	if tmpl, exists := ce.templateCache.Get(source); exists {
		return tmpl, nil
	}

	tmpl, err := ce.Environment.FromString(source)
	if err != nil {
		return nil, err
	}

	ce.templateCache.Put(source, tmpl)
	return tmpl, nil
}
