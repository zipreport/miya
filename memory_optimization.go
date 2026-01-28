package miya

import (
	"strings"
	"sync"
)

// StringInterner provides string interning to reduce memory usage
type StringInterner struct {
	mutex   sync.RWMutex
	strings map[string]string
}

// NewStringInterner creates a new string interner
func NewStringInterner() *StringInterner {
	return &StringInterner{
		strings: make(map[string]string),
	}
}

// Intern interns a string, returning the canonical instance
func (si *StringInterner) Intern(s string) string {
	si.mutex.RLock()
	if interned, exists := si.strings[s]; exists {
		si.mutex.RUnlock()
		return interned
	}
	si.mutex.RUnlock()

	si.mutex.Lock()
	defer si.mutex.Unlock()

	// Double-check after acquiring write lock
	if interned, exists := si.strings[s]; exists {
		return interned
	}

	// Create a copy to ensure we own the memory
	interned := strings.Clone(s)
	si.strings[interned] = interned
	return interned
}

// Size returns the number of interned strings
func (si *StringInterner) Size() int {
	si.mutex.RLock()
	defer si.mutex.RUnlock()
	return len(si.strings)
}

// Clear clears all interned strings
func (si *StringInterner) Clear() {
	si.mutex.Lock()
	defer si.mutex.Unlock()
	si.strings = make(map[string]string)
}

// Global string interner for template identifiers and common strings
var GlobalStringInterner = NewStringInterner()

// ByteBufferPool provides a pool of byte buffers for efficient string building
type ByteBufferPool struct {
	pool sync.Pool
}

// NewByteBufferPool creates a new byte buffer pool
func NewByteBufferPool() *ByteBufferPool {
	return &ByteBufferPool{
		pool: sync.Pool{
			New: func() interface{} {
				return make([]byte, 0, 1024) // Start with 1KB capacity
			},
		},
	}
}

// Get gets a byte buffer from the pool
func (bbp *ByteBufferPool) Get() []byte {
	return bbp.pool.Get().([]byte)
}

// Put returns a byte buffer to the pool
func (bbp *ByteBufferPool) Put(buf []byte) {
	if cap(buf) > 64*1024 { // Don't pool very large buffers
		return
	}
	buf = buf[:0] // Reset length but keep capacity
	bbp.pool.Put(buf)
}

// Global byte buffer pool
var GlobalByteBufferPool = NewByteBufferPool()

// MemoryEfficientTemplate provides memory-optimized template operations
type MemoryEfficientTemplate struct {
	*Template
	interner *StringInterner
}

// NewMemoryEfficientTemplate creates a memory-efficient template wrapper
func NewMemoryEfficientTemplate(tmpl *Template) *MemoryEfficientTemplate {
	return &MemoryEfficientTemplate{
		Template: tmpl,
		interner: NewStringInterner(),
	}
}

// RenderMemoryEfficient renders the template with memory optimizations
func (met *MemoryEfficientTemplate) RenderMemoryEfficient(ctx Context) (string, error) {
	// Use byte buffer for efficient string building
	buf := GlobalByteBufferPool.Get()
	defer GlobalByteBufferPool.Put(buf)

	// For now, use the standard render method
	// In a full implementation, this would use the byte buffer directly
	return met.Template.Render(ctx)
}

// CachedEnvironmentOptions provides options for memory optimization
type CachedEnvironmentOptions struct {
	UseStringInterning bool
	ByteBufferPoolSize int
	ContextPoolSize    int
	TemplatePoolSize   int
}

// DefaultCachedOptions returns default caching options
func DefaultCachedOptions() *CachedEnvironmentOptions {
	return &CachedEnvironmentOptions{
		UseStringInterning: true,
		ByteBufferPoolSize: 100,
		ContextPoolSize:    50,
		TemplatePoolSize:   20,
	}
}

// MemoryStats provides memory usage statistics
type MemoryStats struct {
	InternedStrings      int
	PooledContexts       int
	PooledTemplates      int
	PooledStringBuilders int
	PooledByteBuffers    int
}

// GetMemoryStats returns current memory optimization statistics
func GetMemoryStats() *MemoryStats {
	return &MemoryStats{
		InternedStrings: GlobalStringInterner.Size(),
		// Pool sizes would need to be tracked separately
	}
}

// SmallStringOptimizer optimizes storage of small strings
type SmallStringOptimizer struct {
	smallStrings map[string]bool // Set of strings <= 32 bytes
	mutex        sync.RWMutex
}

// NewSmallStringOptimizer creates a new small string optimizer
func NewSmallStringOptimizer() *SmallStringOptimizer {
	return &SmallStringOptimizer{
		smallStrings: make(map[string]bool),
	}
}

// OptimizeString optimizes string storage for small strings
func (sso *SmallStringOptimizer) OptimizeString(s string) string {
	if len(s) > 32 {
		return s // Don't optimize large strings
	}

	sso.mutex.RLock()
	if sso.smallStrings[s] {
		sso.mutex.RUnlock()
		return s
	}
	sso.mutex.RUnlock()

	sso.mutex.Lock()
	defer sso.mutex.Unlock()

	// Double-check
	if sso.smallStrings[s] {
		return s
	}

	// Make a copy and track it
	optimized := strings.Clone(s)
	sso.smallStrings[optimized] = true
	return optimized
}

// Global small string optimizer
var GlobalSmallStringOptimizer = NewSmallStringOptimizer()

// PrecomputedHash stores precomputed hash values to avoid recomputation
type PrecomputedHash struct {
	mutex  sync.RWMutex
	hashes map[string]uint64
}

// NewPrecomputedHash creates a new precomputed hash store
func NewPrecomputedHash() *PrecomputedHash {
	return &PrecomputedHash{
		hashes: make(map[string]uint64),
	}
}

// GetHash gets or computes a hash for a string
func (ph *PrecomputedHash) GetHash(s string) uint64 {
	ph.mutex.RLock()
	if hash, exists := ph.hashes[s]; exists {
		ph.mutex.RUnlock()
		return hash
	}
	ph.mutex.RUnlock()

	ph.mutex.Lock()
	defer ph.mutex.Unlock()

	// Double-check
	if hash, exists := ph.hashes[s]; exists {
		return hash
	}

	// Simple hash function (in production, use a better one)
	hash := uint64(0)
	for _, b := range []byte(s) {
		hash = hash*31 + uint64(b)
	}

	ph.hashes[s] = hash
	return hash
}

// Global precomputed hash store
var GlobalPrecomputedHash = NewPrecomputedHash()

// MemoryProfiler provides memory profiling capabilities
type MemoryProfiler struct {
	allocations   map[string]int64
	deallocations map[string]int64
	mutex         sync.RWMutex
}

// NewMemoryProfiler creates a new memory profiler
func NewMemoryProfiler() *MemoryProfiler {
	return &MemoryProfiler{
		allocations:   make(map[string]int64),
		deallocations: make(map[string]int64),
	}
}

// RecordAllocation records a memory allocation
func (mp *MemoryProfiler) RecordAllocation(category string, size int64) {
	mp.mutex.Lock()
	defer mp.mutex.Unlock()
	mp.allocations[category] += size
}

// RecordDeallocation records a memory deallocation
func (mp *MemoryProfiler) RecordDeallocation(category string, size int64) {
	mp.mutex.Lock()
	defer mp.mutex.Unlock()
	mp.deallocations[category] += size
}

// GetStats returns memory profiling statistics
func (mp *MemoryProfiler) GetStats() map[string]int64 {
	mp.mutex.RLock()
	defer mp.mutex.RUnlock()

	stats := make(map[string]int64)
	for category, allocs := range mp.allocations {
		deallocs := mp.deallocations[category]
		stats[category] = allocs - deallocs
	}
	return stats
}

// Global memory profiler
var GlobalMemoryProfiler = NewMemoryProfiler()

// MemoryOptimizedContext is a context implementation optimized for memory usage
type MemoryOptimizedContext struct {
	data     map[string]interface{}
	parent   *MemoryOptimizedContext
	interner *StringInterner
}

// NewMemoryOptimizedContext creates a new memory-optimized context
func NewMemoryOptimizedContext() *MemoryOptimizedContext {
	return &MemoryOptimizedContext{
		data:     make(map[string]interface{}),
		interner: GlobalStringInterner,
	}
}

// Get retrieves a variable value
func (moc *MemoryOptimizedContext) Get(key string) (interface{}, bool) {
	// Intern the key to reduce memory usage for repeated lookups
	internedKey := moc.interner.Intern(key)

	if val, ok := moc.data[internedKey]; ok {
		return val, true
	}

	if moc.parent != nil {
		return moc.parent.Get(internedKey)
	}

	return nil, false
}

// Set sets a variable value
func (moc *MemoryOptimizedContext) Set(key string, value interface{}) {
	internedKey := moc.interner.Intern(key)
	moc.data[internedKey] = value
}

// Push creates a new scope
func (moc *MemoryOptimizedContext) Push() Context {
	return &MemoryOptimizedContext{
		data:     make(map[string]interface{}),
		parent:   moc,
		interner: moc.interner,
	}
}

// Pop returns to parent scope
func (moc *MemoryOptimizedContext) Pop() Context {
	if moc.parent != nil {
		return moc.parent
	}
	return moc
}

// All returns all variables
func (moc *MemoryOptimizedContext) All() map[string]interface{} {
	result := make(map[string]interface{})

	// Collect from parent first
	if moc.parent != nil {
		for k, v := range moc.parent.All() {
			result[k] = v
		}
	}

	// Override with local variables
	for k, v := range moc.data {
		result[k] = v
	}

	return result
}

// GetEnv returns the environment (placeholder implementation)
func (moc *MemoryOptimizedContext) GetEnv() *Environment {
	return nil // Would need to be properly implemented
}

// Clone creates a copy of the context
func (moc *MemoryOptimizedContext) Clone() Context {
	clone := &MemoryOptimizedContext{
		data:     make(map[string]interface{}),
		interner: moc.interner,
	}

	// Copy all variables
	for k, v := range moc.All() {
		clone.data[k] = v
	}

	return clone
}
