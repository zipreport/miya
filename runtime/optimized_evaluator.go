package runtime

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/zipreport/miya/parser"
)

// OptimizedEvaluator is a performance-optimized version of the default evaluator
type OptimizedEvaluator struct {
	*DefaultEvaluator
	stringBuilderPool sync.Pool
	contextPool       sync.Pool
}

// NewOptimizedEvaluator creates a new optimized evaluator
func NewOptimizedEvaluator() *OptimizedEvaluator {
	return &OptimizedEvaluator{
		DefaultEvaluator: NewEvaluator(),
		stringBuilderPool: sync.Pool{
			New: func() interface{} {
				return &strings.Builder{}
			},
		},
		contextPool: sync.Pool{
			New: func() interface{} {
				return make(map[string]interface{})
			},
		},
	}
}

// getStringBuilder gets a string builder from the pool
func (e *OptimizedEvaluator) getStringBuilder() *strings.Builder {
	return e.stringBuilderPool.Get().(*strings.Builder)
}

// putStringBuilder returns a string builder to the pool
func (e *OptimizedEvaluator) putStringBuilder(sb *strings.Builder) {
	sb.Reset()
	e.stringBuilderPool.Put(sb)
}

// getContextMap gets a context map from the pool
func (e *OptimizedEvaluator) getContextMap() map[string]interface{} {
	return e.contextPool.Get().(map[string]interface{})
}

// putContextMap returns a context map to the pool
func (e *OptimizedEvaluator) putContextMap(m map[string]interface{}) {
	// Clear the map
	for k := range m {
		delete(m, k)
	}
	e.contextPool.Put(m)
}

// EvalTemplateNodeOptimized is an optimized version of template evaluation
func (e *OptimizedEvaluator) EvalTemplateNodeOptimized(node *parser.TemplateNode, ctx Context) (string, error) {
	if len(node.Children) == 0 {
		return "", nil
	}

	// Use pooled string builder for better memory efficiency
	sb := e.getStringBuilder()
	defer e.putStringBuilder(sb)

	for _, child := range node.Children {
		result, err := e.EvalNode(child, ctx)
		if err != nil {
			return "", err
		}

		// Convert result to string and append
		if result != nil {
			switch v := result.(type) {
			case string:
				sb.WriteString(v)
			case []byte:
				sb.Write(v)
			default:
				sb.WriteString(fmt.Sprintf("%v", v))
			}
		}
	}

	return sb.String(), nil
}

// EvalForNodeOptimized is an optimized version of for loop evaluation
func (e *OptimizedEvaluator) EvalForNodeOptimized(node *parser.ForNode, ctx Context) (string, error) {
	// Evaluate the iterable expression
	iterable, err := e.EvalNode(node.Iterable, ctx)
	if err != nil {
		return "", fmt.Errorf("error evaluating for loop iterable: %v", err)
	}

	// Use pooled string builder
	sb := e.getStringBuilder()
	defer e.putStringBuilder(sb)

	// Convert to slice for iteration using reflection
	items := e.toSlice(iterable)

	// If empty and has else clause
	if len(items) == 0 && len(node.Else) > 0 {
		for _, child := range node.Else {
			result, err := e.EvalNode(child, ctx)
			if err != nil {
				return "", err
			}
			if result != nil {
				sb.WriteString(ToString(result))
			}
		}
		return sb.String(), nil
	}

	// Iterate over items
	for i, item := range items {
		// Create a new scope for loop variables
		if len(node.Variables) == 1 {
			// Single variable assignment
			ctx.SetVariable(node.Variables[0], item)
		} else {
			// Multiple variable assignment - unpack the item
			items := e.toSlice(item)
			if len(items) != len(node.Variables) {
				return "", fmt.Errorf("cannot unpack %d values into %d variables", len(items), len(node.Variables))
			}

			// Assign each value to its corresponding variable
			for j, variable := range node.Variables {
				ctx.SetVariable(variable, items[j])
			}
		}

		// Set loop variables
		ctx.SetVariable("loop", map[string]interface{}{
			"index":     i + 1,
			"index0":    i,
			"revindex":  len(items) - i,
			"revindex0": len(items) - i - 1,
			"first":     i == 0,
			"last":      i == len(items)-1,
			"length":    len(items),
		})

		// Execute body
		for _, child := range node.Body {
			result, err := e.EvalNode(child, ctx)
			if err != nil {
				return "", err
			}
			if result != nil {
				sb.WriteString(ToString(result))
			}
		}
	}

	return sb.String(), nil
}

// toSlice converts various types to a slice of interfaces
func (e *OptimizedEvaluator) toSlice(value interface{}) []interface{} {
	if value == nil {
		return []interface{}{}
	}

	switch v := value.(type) {
	case []interface{}:
		return v
	case []string:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = item
		}
		return result
	case []int:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = item
		}
		return result
	case string:
		result := make([]interface{}, len(v))
		for i, char := range v {
			result[i] = string(char)
		}
		return result
	default:
		// Use reflection as fallback
		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Slice {
			result := make([]interface{}, rv.Len())
			for i := 0; i < rv.Len(); i++ {
				result[i] = rv.Index(i).Interface()
			}
			return result
		}
		return []interface{}{value}
	}
}

// EvalIfNodeOptimized is an optimized version of if statement evaluation
func (e *OptimizedEvaluator) EvalIfNodeOptimized(node *parser.IfNode, ctx Context) (string, error) {
	// Use pooled string builder
	sb := e.getStringBuilder()
	defer e.putStringBuilder(sb)

	// Evaluate condition
	condition, err := e.EvalNode(node.Condition, ctx)
	if err != nil {
		return "", fmt.Errorf("error evaluating if condition: %v", err)
	}

	// Check if condition is truthy
	if e.DefaultEvaluator.isTruthy(condition) {
		// Execute if body
		for _, child := range node.Body {
			result, err := e.EvalNode(child, ctx)
			if err != nil {
				return "", err
			}
			if result != nil {
				sb.WriteString(ToString(result))
			}
		}
		return sb.String(), nil
	}

	// Check elif conditions
	for _, elifNode := range node.ElseIfs {
		condition, err := e.EvalNode(elifNode.Condition, ctx)
		if err != nil {
			return "", fmt.Errorf("error evaluating elif condition: %v", err)
		}

		if e.DefaultEvaluator.isTruthy(condition) {
			for _, child := range elifNode.Body {
				result, err := e.EvalNode(child, ctx)
				if err != nil {
					return "", err
				}
				if result != nil {
					sb.WriteString(ToString(result))
				}
			}
			return sb.String(), nil
		}
	}

	// Execute else body if present
	for _, child := range node.Else {
		result, err := e.EvalNode(child, ctx)
		if err != nil {
			return "", err
		}
		if result != nil {
			sb.WriteString(ToString(result))
		}
	}

	return sb.String(), nil
}

// BatchEvaluation allows evaluating multiple nodes efficiently
type BatchEvaluation struct {
	nodes   []parser.Node
	results []interface{}
	errors  []error
}

// NewBatchEvaluation creates a new batch evaluation
func NewBatchEvaluation(nodes []parser.Node) *BatchEvaluation {
	return &BatchEvaluation{
		nodes:   nodes,
		results: make([]interface{}, len(nodes)),
		errors:  make([]error, len(nodes)),
	}
}

// EvaluateBatch evaluates all nodes in the batch
func (e *OptimizedEvaluator) EvaluateBatch(batch *BatchEvaluation, ctx Context) {
	for i, node := range batch.nodes {
		result, err := e.EvalNode(node, ctx)
		batch.results[i] = result
		batch.errors[i] = err
	}
}

// CachedEvaluator adds caching capabilities to the optimized evaluator
type CachedEvaluator struct {
	*OptimizedEvaluator
	cache     sync.Map
	cacheHits int64
	cacheMiss int64
}

// NewCachedEvaluator creates a new cached evaluator
func NewCachedEvaluator() *CachedEvaluator {
	return &CachedEvaluator{
		OptimizedEvaluator: NewOptimizedEvaluator(),
	}
}

// cacheKey generates a cache key for a node and context
func (e *CachedEvaluator) cacheKey(node parser.Node, ctx Context) string {
	// Simple cache key - in production you might want more sophisticated hashing
	return fmt.Sprintf("%T:%p", node, node)
}

// EvalNodeCached evaluates a node with caching
func (e *CachedEvaluator) EvalNodeCached(node parser.Node, ctx Context) (interface{}, error) {
	// Only cache certain types of nodes that are expensive and likely to be repeated
	switch node.(type) {
	case *parser.FilterNode, *parser.TestNode:
		key := e.cacheKey(node, ctx)
		if cached, ok := e.cache.Load(key); ok {
			e.cacheHits++
			return cached, nil
		}

		result, err := e.EvalNode(node, ctx)
		if err == nil {
			e.cache.Store(key, result)
		}
		e.cacheMiss++
		return result, err
	default:
		return e.EvalNode(node, ctx)
	}
}

// GetCacheStats returns cache performance statistics
func (e *CachedEvaluator) GetCacheStats() (hits, misses int64) {
	return e.cacheHits, e.cacheMiss
}

// ClearCache clears the evaluation cache
func (e *CachedEvaluator) ClearCache() {
	e.cache.Range(func(key, value interface{}) bool {
		e.cache.Delete(key)
		return true
	})
	e.cacheHits = 0
	e.cacheMiss = 0
}

// MemoryEfficientContext is a context implementation optimized for memory usage
type MemoryEfficientContext struct {
	variables map[string]interface{}
	stack     []map[string]interface{}
	pool      *sync.Pool
}

// NewMemoryEfficientContext creates a new memory-efficient context
func NewMemoryEfficientContext() *MemoryEfficientContext {
	return &MemoryEfficientContext{
		variables: make(map[string]interface{}),
		stack:     make([]map[string]interface{}, 0, 4), // Pre-allocate for common case
		pool: &sync.Pool{
			New: func() interface{} {
				return make(map[string]interface{})
			},
		},
	}
}

// GetVariable retrieves a variable value
func (c *MemoryEfficientContext) GetVariable(key string) (interface{}, bool) {
	// Check stack from top to bottom
	for i := len(c.stack) - 1; i >= 0; i-- {
		if val, ok := c.stack[i][key]; ok {
			return val, true
		}
	}

	// Check base variables
	val, ok := c.variables[key]
	return val, ok
}

// SetVariable sets a variable value
func (c *MemoryEfficientContext) SetVariable(key string, value interface{}) {
	if len(c.stack) > 0 {
		// Set in top scope
		c.stack[len(c.stack)-1][key] = value
	} else {
		c.variables[key] = value
	}
}

// PushScope creates a new variable scope
func (c *MemoryEfficientContext) PushScope() {
	scope := c.pool.Get().(map[string]interface{})
	c.stack = append(c.stack, scope)
}

// PopScope removes the top variable scope
func (c *MemoryEfficientContext) PopScope() {
	if len(c.stack) > 0 {
		scope := c.stack[len(c.stack)-1]
		c.stack = c.stack[:len(c.stack)-1]

		// Clear and return to pool
		for k := range scope {
			delete(scope, k)
		}
		c.pool.Put(scope)
	}
}

// Clone creates a copy of the context
func (c *MemoryEfficientContext) Clone() Context {
	newCtx := NewMemoryEfficientContext()

	// Copy base variables
	for k, v := range c.variables {
		newCtx.variables[k] = v
	}

	// Copy stack
	for _, scope := range c.stack {
		newScope := newCtx.pool.Get().(map[string]interface{})
		for k, v := range scope {
			newScope[k] = v
		}
		newCtx.stack = append(newCtx.stack, newScope)
	}

	return newCtx
}

// All returns all variables (flattened)
func (c *MemoryEfficientContext) All() map[string]interface{} {
	result := make(map[string]interface{})

	// Start with base variables
	for k, v := range c.variables {
		result[k] = v
	}

	// Override with stack variables (bottom to top)
	for _, scope := range c.stack {
		for k, v := range scope {
			result[k] = v
		}
	}

	return result
}
