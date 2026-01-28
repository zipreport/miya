package miya

// Phase 5b: Copy-on-Write (COW) Context Implementation
//
// This provides a COW wrapper around the standard context to avoid
// deep copying on every Clone() call. Instead, we share the data map
// and only copy when modifications occur.

import (
	"sync"
)

// cowContext implements a copy-on-write context
type cowContext struct {
	// shared points to the shared data map (read-only until write)
	shared map[string]interface{}

	// local holds modifications made to this context (write layer)
	local map[string]interface{}

	// parent context for variable lookup hierarchy
	parent Context

	// env is the template environment
	env *Environment

	// mu protects concurrent access if needed
	mu sync.RWMutex

	// isShared indicates if the shared map is still being shared with parent
	isShared bool
}

// newCOWContext creates a new COW context
func newCOWContext(parent Context, env *Environment) *cowContext {
	return &cowContext{
		shared:   make(map[string]interface{}, 8),
		local:    nil, // No local modifications yet
		parent:   parent,
		env:      env,
		isShared: false,
	}
}

// newCOWContextFrom creates a COW context from existing data
func newCOWContextFrom(data map[string]interface{}, parent Context, env *Environment) *cowContext {
	// Start with data as shared (read-only)
	return &cowContext{
		shared:   data,
		local:    nil,
		parent:   parent,
		env:      env,
		isShared: true, // Mark as shared since we're using external data
	}
}

// Get retrieves a variable value
func (c *cowContext) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Check local modifications first
	if c.local != nil {
		if val, ok := c.local[key]; ok {
			return val, true
		}
	}

	// Check shared data
	if val, ok := c.shared[key]; ok {
		return val, true
	}

	// Check parent context
	if c.parent != nil {
		return c.parent.Get(key)
	}

	// Check environment globals
	if c.env != nil {
		if val, ok := c.env.globals[key]; ok {
			return val, true
		}
	}

	return nil, false
}

// Set modifies a variable (triggers copy-on-write)
func (c *cowContext) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// If we're sharing data and don't have local modifications yet, initialize local
	if c.local == nil {
		c.local = make(map[string]interface{}, 4)
	}

	// Write to local layer
	c.local[key] = value
}

// Clone creates a copy-on-write clone
func (c *cowContext) Clone() Context {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Create new COW context that shares our data
	clone := &cowContext{
		parent:   c.parent,
		env:      c.env,
		local:    nil,  // New clone has no local modifications
		isShared: true, // It's sharing our data
	}

	// If we have local modifications, merge them into shared for the clone
	if c.local != nil {
		// Create merged view for clone to share
		merged := make(map[string]interface{}, len(c.shared)+len(c.local))

		// Copy shared first
		for k, v := range c.shared {
			merged[k] = v
		}

		// Apply local modifications
		for k, v := range c.local {
			merged[k] = v
		}

		clone.shared = merged
	} else {
		// No local modifications, just share our shared map
		clone.shared = c.shared
	}

	return clone
}

// All returns all variables (merged view)
func (c *cowContext) All() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Build merged view
	result := make(map[string]interface{}, len(c.shared)+len(c.local))

	// Copy shared
	for k, v := range c.shared {
		result[k] = v
	}

	// Apply local modifications
	if c.local != nil {
		for k, v := range c.local {
			result[k] = v
		}
	}

	return result
}

// Push creates a new context layer (for blocks)
func (c *cowContext) Push() Context {
	// Create new COW context with current as parent
	return &cowContext{
		shared:   make(map[string]interface{}, 4),
		local:    nil,
		parent:   c,
		env:      c.env,
		isShared: false,
	}
}

// Pop returns the parent context (for blocks)
func (c *cowContext) Pop() Context {
	return c.parent
}

// GetEnv returns the environment
func (c *cowContext) GetEnv() *Environment {
	return c.env
}

// SetVariable is an alias for Set for compatibility
func (c *cowContext) SetVariable(key string, value interface{}) {
	c.Set(key, value)
}

// GetVariable is an alias for Get for compatibility
func (c *cowContext) GetVariable(key string) (interface{}, bool) {
	return c.Get(key)
}

// NewCOWContextFrom creates a new COW context from a map
func NewCOWContextFrom(data map[string]interface{}) Context {
	return newCOWContextFrom(data, nil, nil)
}

// NewCOWContext creates a new empty COW context
func NewCOWContext() Context {
	return newCOWContext(nil, nil)
}
