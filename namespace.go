package miya

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/zipreport/miya/runtime"
)

// Namespace represents a Jinja2 namespace object
// Namespaces are mutable objects that allow sharing data between scopes
type Namespace struct {
	mu   sync.RWMutex
	data map[string]interface{}
}

// NewNamespace creates a new namespace object
func NewNamespace() *Namespace {
	return &Namespace{
		data: make(map[string]interface{}),
	}
}

// Get retrieves a value from the namespace
func (ns *Namespace) Get(key string) (interface{}, bool) {
	ns.mu.RLock()
	defer ns.mu.RUnlock()
	val, ok := ns.data[key]
	return val, ok
}

// Set sets a value in the namespace
func (ns *Namespace) Set(key string, value interface{}) {
	ns.mu.Lock()
	defer ns.mu.Unlock()
	ns.data[key] = value
}

// Has checks if a key exists in the namespace
func (ns *Namespace) Has(key string) bool {
	ns.mu.RLock()
	defer ns.mu.RUnlock()
	_, ok := ns.data[key]
	return ok
}

// All returns a copy of all data in the namespace
func (ns *Namespace) All() map[string]interface{} {
	ns.mu.RLock()
	defer ns.mu.RUnlock()

	result := make(map[string]interface{})
	for k, v := range ns.data {
		result[k] = v
	}
	return result
}

// Delete removes a key from the namespace
func (ns *Namespace) Delete(key string) {
	ns.mu.Lock()
	defer ns.mu.Unlock()
	delete(ns.data, key)
}

// Clear removes all data from the namespace
func (ns *Namespace) Clear() {
	ns.mu.Lock()
	defer ns.mu.Unlock()
	ns.data = make(map[string]interface{})
}

// String returns a string representation of the namespace
func (ns *Namespace) String() string {
	ns.mu.RLock()
	defer ns.mu.RUnlock()
	return fmt.Sprintf("namespace(%v)", ns.data)
}

// namespaceFunction is the global namespace() function for templates
func namespaceFunction(args ...interface{}) (interface{}, error) {
	ns := NewNamespace()

	// If keyword arguments are provided, initialize the namespace with them
	if len(args) == 1 {
		if kwargs, ok := args[0].(map[string]interface{}); ok {
			for k, v := range kwargs {
				ns.Set(k, v)
			}
		}
	}

	return ns, nil
}

// NamespaceWrapper wraps a namespace to provide dot-notation access in templates
type NamespaceWrapper struct {
	*Namespace
}

// GetAttr provides attribute-style access to namespace values
func (nw *NamespaceWrapper) GetAttr(name string) (interface{}, error) {
	val, ok := nw.Get(name)
	if !ok {
		// Return nil for missing attributes (Jinja2 behavior)
		return nil, nil
	}
	return val, nil
}

// SetAttr provides attribute-style setting of namespace values
func (nw *NamespaceWrapper) SetAttr(name string, value interface{}) error {
	nw.Set(name, value)
	return nil
}

// rangeFunction implements the range() global function
// It generates a sequence of numbers similar to Python's range()
func rangeFunction(args ...interface{}) (interface{}, error) {
	var start, stop, step int

	switch len(args) {
	case 1:
		// range(stop)
		stop = toInt(args[0])
		start = 0
		step = 1
	case 2:
		// range(start, stop)
		start = toInt(args[0])
		stop = toInt(args[1])
		step = 1
	case 3:
		// range(start, stop, step)
		start = toInt(args[0])
		stop = toInt(args[1])
		step = toInt(args[2])
		if step == 0 {
			return nil, fmt.Errorf("range() step argument must not be zero")
		}
	default:
		return nil, fmt.Errorf("range() takes 1 to 3 arguments, got %d", len(args))
	}

	// Generate the sequence
	var result []interface{}
	if step > 0 {
		for i := start; i < stop; i += step {
			result = append(result, i)
		}
	} else {
		for i := start; i > stop; i += step {
			result = append(result, i)
		}
	}

	return result, nil
}

// dictFunction implements the dict() global function
// It creates a dictionary from key-value pairs
func dictFunction(args ...interface{}) (interface{}, error) {
	result := make(map[string]interface{})

	// Handle different argument patterns
	if len(args) == 0 {
		return result, nil
	}

	// If single argument is a map, use it as kwargs
	if len(args) == 1 {
		if kwargs, ok := args[0].(map[string]interface{}); ok {
			for k, v := range kwargs {
				result[k] = v
			}
			return result, nil
		}
	}

	// Otherwise, expect pairs of arguments
	if len(args)%2 != 0 {
		return nil, fmt.Errorf("dict() requires an even number of arguments when not using keyword arguments")
	}

	for i := 0; i < len(args); i += 2 {
		key, ok := args[i].(string)
		if !ok {
			return nil, fmt.Errorf("dict() keys must be strings, got %T", args[i])
		}
		result[key] = args[i+1]
	}

	return result, nil
}

// cyclerFunction creates a cycler object that cycles through values
func cyclerFunction(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("cycler() requires at least one argument")
	}

	return &runtime.Cycler{
		Items:   args,
		Current: 0,
	}, nil
}

// joinerFunction creates a joiner object that joins values with separators
func joinerFunction(args ...interface{}) (interface{}, error) {
	separator := ", " // default separator
	if len(args) > 0 {
		if sep, ok := args[0].(string); ok {
			separator = sep
		}
	}

	return &runtime.Joiner{
		Separator: separator,
		Used:      false,
	}, nil
}

// lipsumFunction generates Lorem Ipsum text
func lipsumFunction(args ...interface{}) (interface{}, error) {
	paragraphs := 5
	sentences := true
	minSentences := 4
	maxSentences := 6

	// Parse arguments
	if len(args) > 0 {
		if p, ok := args[0].(int); ok {
			paragraphs = p
		}
	}
	if len(args) > 1 {
		if s, ok := args[1].(bool); ok {
			sentences = s
		}
	}
	if len(args) > 2 {
		if min, ok := args[2].(int); ok {
			minSentences = min
		}
	}
	if len(args) > 3 {
		if max, ok := args[3].(int); ok {
			maxSentences = max
		}
	}

	// Lorem ipsum words
	words := []string{
		"lorem", "ipsum", "dolor", "sit", "amet", "consectetur", "adipiscing", "elit",
		"sed", "do", "eiusmod", "tempor", "incididunt", "ut", "labore", "et", "dolore",
		"magna", "aliqua", "enim", "ad", "minim", "veniam", "quis", "nostrud",
		"exercitation", "ullamco", "laboris", "nisi", "aliquip", "ex", "ea", "commodo",
		"consequat", "duis", "aute", "irure", "in", "reprehenderit", "voluptate",
		"velit", "esse", "cillum", "fugiat", "nulla", "pariatur", "excepteur", "sint",
		"occaecat", "cupidatat", "non", "proident", "sunt", "culpa", "qui", "officia",
		"deserunt", "mollit", "anim", "id", "est", "laborum",
	}

	var result strings.Builder

	for p := 0; p < paragraphs; p++ {
		if p > 0 {
			result.WriteString("\n\n")
		}

		if sentences {
			// Generate sentences
			numSentences := minSentences + (p % (maxSentences - minSentences + 1))
			for s := 0; s < numSentences; s++ {
				if s > 0 {
					result.WriteString(" ")
				}

				// Generate a sentence (5-15 words)
				sentenceLength := 5 + (s % 11)
				for w := 0; w < sentenceLength; w++ {
					if w > 0 {
						result.WriteString(" ")
					}
					wordIndex := (p*numSentences*sentenceLength + s*sentenceLength + w) % len(words)
					word := words[wordIndex]
					if w == 0 {
						// Capitalize first word
						word = capitalizeFirst(word)
					}
					result.WriteString(word)
				}
				result.WriteString(".")
			}
		} else {
			// Generate word list
			wordCount := 50 + (p * 10)
			for w := 0; w < wordCount; w++ {
				if w > 0 {
					result.WriteString(" ")
				}
				wordIndex := (p*wordCount + w) % len(words)
				result.WriteString(words[wordIndex])
			}
		}
	}

	return result.String(), nil
}

// toInt converts an interface to int
func toInt(v interface{}) int {
	switch val := v.(type) {
	case int:
		return val
	case int64:
		return int(val)
	case float64:
		return int(val)
	case string:
		// Try to parse string as int
		var i int
		fmt.Sscanf(val, "%d", &i)
		return i
	default:
		return 0
	}
}

// zipFunction combines multiple iterables into tuples
func zipFunction(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return []interface{}{}, nil
	}

	// Convert all arguments to slices
	slices := make([][]interface{}, len(args))
	minLen := -1

	for i, arg := range args {
		slice, err := makeIterable(arg)
		if err != nil {
			return nil, fmt.Errorf("zip() argument %d is not iterable: %v", i+1, err)
		}
		slices[i] = slice
		if minLen == -1 || len(slice) < minLen {
			minLen = len(slice)
		}
	}

	// Create result tuples
	result := make([]interface{}, minLen)
	for i := 0; i < minLen; i++ {
		tuple := make([]interface{}, len(slices))
		for j, slice := range slices {
			tuple[j] = slice[i]
		}
		result[i] = tuple
	}

	return result, nil
}

// enumerateFunction returns pairs of (index, item) for an iterable
func enumerateFunction(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("enumerate() requires at least one argument")
	}

	start := 0
	if len(args) > 1 {
		if s, ok := args[1].(int); ok {
			start = s
		}
	}

	slice, err := makeIterable(args[0])
	if err != nil {
		return nil, fmt.Errorf("enumerate() argument is not iterable: %v", err)
	}

	result := make([]interface{}, len(slice))
	for i, item := range slice {
		result[i] = []interface{}{start + i, item}
	}

	return result, nil
}

// urlForFunction generates URLs for endpoints (basic implementation)
func urlForFunction(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("url_for() requires at least one argument (endpoint)")
	}

	endpoint, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("url_for() endpoint must be a string")
	}

	// Basic implementation - in a real application, this would use a URL router
	// For now, we'll create simple URLs based on the endpoint name
	url := "/" + endpoint

	// Handle additional arguments as URL parameters
	if len(args) > 1 {
		params := make([]string, 0)
		for i := 1; i < len(args); i += 2 {
			if i+1 < len(args) {
				key := fmt.Sprintf("%v", args[i])
				value := fmt.Sprintf("%v", args[i+1])
				params = append(params, fmt.Sprintf("%s=%s", key, value))
			}
		}
		if len(params) > 0 {
			url += "?" + strings.Join(params, "&")
		}
	}

	return url, nil
}

// Helper function to convert interface to iterable slice
func makeIterable(value interface{}) ([]interface{}, error) {
	if value == nil {
		return []interface{}{}, nil
	}

	switch v := value.(type) {
	case []interface{}:
		return v, nil
	case []string:
		result := make([]interface{}, len(v))
		for i, s := range v {
			result[i] = s
		}
		return result, nil
	case []int:
		result := make([]interface{}, len(v))
		for i, n := range v {
			result[i] = n
		}
		return result, nil
	case string:
		result := make([]interface{}, len(v))
		for i, r := range []rune(v) {
			result[i] = string(r)
		}
		return result, nil
	default:
		// Try reflection for other slice types
		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
			result := make([]interface{}, rv.Len())
			for i := 0; i < rv.Len(); i++ {
				result[i] = rv.Index(i).Interface()
			}
			return result, nil
		}
		return nil, fmt.Errorf("value is not iterable: %T", value)
	}
}
