package miya

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"
	"unicode/utf8"
)

// capitalizeFirst capitalizes the first letter of a string.
// This is used for struct field/method name lookups (e.g., "name" -> "Name").
// Note: strings.Title is deprecated since Go 1.18.
func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError {
		return s
	}
	return string(unicode.ToUpper(r)) + s[size:]
}

type Context interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{})
	Push() Context
	Pop() Context
	All() map[string]interface{}
	GetEnv() *Environment
	Clone() Context
}

type LoopInfo struct {
	Index     int
	Index0    int
	RevIndex  int
	RevIndex0 int
	First     bool
	Last      bool
	Length    int
}

type context struct {
	parent *context
	data   map[string]interface{}
	env    *Environment

	// Phase 3b optimization: Cache for All() method
	cachedAll     map[string]interface{}
	allCacheValid bool
}

func NewContext() Context {
	return &context{
		data: make(map[string]interface{}, 8), // Phase 4a: Pre-size for typical usage
	}
}

func NewContextFrom(data map[string]interface{}) Context {
	// Phase 4a: Pre-size based on input size
	ctx := &context{
		data: make(map[string]interface{}, len(data)),
	}

	// Copy data to avoid mutations
	for k, v := range data {
		ctx.data[k] = v
	}

	return ctx
}

func newContextWithEnv(env *Environment) Context {
	ctx := &context{
		data: make(map[string]interface{}, 8), // Phase 4a: Pre-size for typical usage
		env:  env,
	}

	// Phase 3b optimization: Don't copy globals - use lazy lookup instead
	// Globals will be looked up via env.globals when not found in local data

	return ctx
}

func (c *context) Get(key string) (interface{}, bool) {
	// Phase 4a optimization: Fast path for keys without dots (90%+ of cases)
	if !strings.ContainsRune(key, '.') {
		// Simple key - no split needed
		current := c
		for current != nil {
			if val, ok := current.data[key]; ok {
				return val, true
			}
			current = current.parent
		}

		// Check environment globals
		if c.env != nil {
			if val, ok := c.env.globals[key]; ok {
				return val, true
			}
		}

		return nil, false
	}

	// Slow path: Handle dot notation (e.g., "user.name")
	parts := strings.Split(key, ".")

	current := c
	for current != nil {
		val, ok := current.getLocal(parts)
		if ok {
			return val, true
		}
		current = current.parent
	}

	// Check environment globals (only for simple keys)
	if c.env != nil && len(parts) == 1 {
		if val, ok := c.env.globals[key]; ok {
			return val, true
		}
	}

	return nil, false
}

func (c *context) getLocal(parts []string) (interface{}, bool) {
	if len(parts) == 0 {
		return nil, false
	}

	val, ok := c.data[parts[0]]
	if !ok {
		return nil, false
	}

	// Navigate through nested structures
	for i := 1; i < len(parts); i++ {
		val = c.getAttribute(val, parts[i])
		if val == nil {
			return nil, false
		}
	}

	return val, true
}

func (c *context) getAttribute(obj interface{}, attr string) interface{} {
	if obj == nil {
		return nil
	}

	switch v := obj.(type) {
	case map[string]interface{}:
		return v[attr]
	case map[string]string:
		return v[attr]
	default:
		// Use reflection for struct fields
		rv := reflect.ValueOf(obj)
		if rv.Kind() == reflect.Ptr {
			rv = rv.Elem()
		}

		if rv.Kind() == reflect.Struct {
			field := rv.FieldByName(attr)
			if field.IsValid() && field.CanInterface() {
				return field.Interface()
			}

			// Try with capitalized field name
			field = rv.FieldByName(capitalizeFirst(attr))
			if field.IsValid() && field.CanInterface() {
				return field.Interface()
			}
		}

		// Try method call (with panic recovery)
		if rv.Kind() == reflect.Struct || rv.Kind() == reflect.Ptr {
			if result := safeMethodCall(rv, attr); result != nil {
				return result
			}
			if result := safeMethodCall(rv, capitalizeFirst(attr)); result != nil {
				return result
			}
		}

		return nil
	}
}

// safeMethodCall safely calls a method on a reflect.Value with panic recovery.
// Returns nil if the method doesn't exist, has wrong signature, or panics.
func safeMethodCall(rv reflect.Value, methodName string) (result interface{}) {
	defer func() {
		if r := recover(); r != nil {
			// Method panicked, return nil
			result = nil
		}
	}()

	method := rv.MethodByName(methodName)
	if !method.IsValid() {
		return nil
	}

	// Only call methods with no arguments
	methodType := method.Type()
	if methodType.NumIn() != 0 {
		return nil
	}

	results := method.Call(nil)
	if len(results) > 0 {
		return results[0].Interface()
	}
	return nil
}

func (c *context) Set(key string, value interface{}) {
	c.data[key] = value
	// Invalidate All() cache (Phase 3b optimization)
	c.allCacheValid = false
}

func (c *context) Push() Context {
	return &context{
		parent: c,
		data:   make(map[string]interface{}, 8), // Phase 4a: Pre-size for typical child context
		env:    c.env,
	}
}

func (c *context) Pop() Context {
	if c.parent != nil {
		return c.parent
	}
	return c
}

func (c *context) All() map[string]interface{} {
	// Use cached version if valid (Phase 3b optimization)
	if c.allCacheValid {
		return c.cachedAll
	}

	result := make(map[string]interface{})

	// Phase 3b optimization: Start with environment globals (if present)
	if c.env != nil {
		for k, v := range c.env.globals {
			result[k] = v
		}
	}

	// Collect all values from parent contexts (overrides globals)
	var collect func(*context)
	collect = func(ctx *context) {
		if ctx == nil {
			return
		}
		if ctx.parent != nil {
			collect(ctx.parent)
		}
		for k, v := range ctx.data {
			result[k] = v
		}
	}

	collect(c)

	// Cache the result (Phase 3b optimization)
	c.cachedAll = result
	c.allCacheValid = true

	return result
}

func (c *context) GetEnv() *Environment {
	return c.env
}

func (c *context) Clone() Context {
	// Phase 4a: Pre-size based on current data + room for growth
	clone := &context{
		parent: c.parent,
		data:   make(map[string]interface{}, len(c.data)+4),
		env:    c.env,
	}

	// Copy data
	for k, v := range c.data {
		clone.data[k] = v
	}

	return clone
}

func (c *context) String() string {
	return fmt.Sprintf("Context(%v)", c.All())
}
