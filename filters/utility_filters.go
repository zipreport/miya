package filters

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/zipreport/miya/runtime"
)

// DefaultFilter returns default value if input is falsy
func DefaultFilter(value interface{}, args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return value, nil
	}

	defaultValue := args[0]
	boolean := false

	if len(args) > 1 {
		boolean = ToBool(args[1])
	}

	// Check if value is undefined
	if _, ok := value.(*runtime.Undefined); ok {
		return defaultValue, nil
	}

	if boolean {
		// Use boolean test
		if ToBool(value) {
			return value, nil
		}
	} else {
		// Use undefined test (nil, empty string, etc.)
		if value != nil {
			if s, ok := value.(string); ok && s != "" {
				return value, nil
			} else if !ok {
				return value, nil
			}
		}
	}

	return defaultValue, nil
}

// MapFilter applies an attribute or filter to each item
func MapFilter(value interface{}, args ...interface{}) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("map filter requires attribute or filter name")
	}

	attrOrFilter := ToString(args[0])

	switch v := value.(type) {
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			// Try as attribute first
			attr := extractAttribute(item, attrOrFilter)
			if attr != nil {
				result[i] = attr
			} else {
				// Could try as filter here
				result[i] = item
			}
		}
		return result, nil

	default:
		return nil, fmt.Errorf("map filter requires a sequence")
	}
}

// SelectFilter selects items that pass a test
func SelectFilter(value interface{}, args ...interface{}) (interface{}, error) {
	// This is a simplified version - full implementation would support test functions
	switch v := value.(type) {
	case []interface{}:
		var result []interface{}
		for _, item := range v {
			if ToBool(item) {
				result = append(result, item)
			}
		}
		return result, nil

	default:
		return nil, fmt.Errorf("select filter requires a sequence")
	}
}

// RejectFilter rejects items that pass a test
func RejectFilter(value interface{}, args ...interface{}) (interface{}, error) {
	// This is a simplified version - full implementation would support test functions
	switch v := value.(type) {
	case []interface{}:
		var result []interface{}
		for _, item := range v {
			if !ToBool(item) {
				result = append(result, item)
			}
		}
		return result, nil

	default:
		return nil, fmt.Errorf("reject filter requires a sequence")
	}
}

// AttrFilter extracts an attribute from an object
func AttrFilter(value interface{}, args ...interface{}) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("attr filter requires attribute name")
	}

	attrName := ToString(args[0])
	return extractAttribute(value, attrName), nil
}

// PPrintFilter formats value for pretty printing
func PPrintFilter(value interface{}, args ...interface{}) (interface{}, error) {
	verbose := false
	if len(args) > 0 {
		verbose = ToBool(args[0])
	}

	if verbose {
		// More detailed representation
		return fmt.Sprintf("%#v", value), nil
	}

	// Simple pretty print using strings.Builder for efficiency
	switch v := value.(type) {
	case map[string]interface{}:
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		var result strings.Builder
		result.WriteByte('{')
		for i, k := range keys {
			if i > 0 {
				result.WriteString(", ")
			}
			result.WriteString(k)
			result.WriteString(": ")
			fmt.Fprintf(&result, "%v", v[k])
		}
		result.WriteByte('}')
		return result.String(), nil

	case []interface{}:
		var result strings.Builder
		result.WriteByte('[')
		for i, item := range v {
			if i > 0 {
				result.WriteString(", ")
			}
			fmt.Fprintf(&result, "%v", item)
		}
		result.WriteByte(']')
		return result.String(), nil

	default:
		return fmt.Sprintf("%v", value), nil
	}
}

// DictSortFilter sorts a dictionary by keys or values
func DictSortFilter(value interface{}, args ...interface{}) (interface{}, error) {
	caseSensitive := false
	byKey := true
	reverse := false

	if len(args) > 0 {
		caseSensitive = ToBool(args[0])
	}
	if len(args) > 1 {
		byKey = ToString(args[1]) == "key"
	}
	if len(args) > 2 {
		reverse = ToBool(args[2])
	}

	switch v := value.(type) {
	case map[string]interface{}:
		type kv struct {
			key   string
			value interface{}
		}

		var pairs []kv
		for k, val := range v {
			pairs = append(pairs, kv{key: k, value: val})
		}

		sort.Slice(pairs, func(i, j int) bool {
			var a, b string
			if byKey {
				a, b = pairs[i].key, pairs[j].key
			} else {
				a, b = ToString(pairs[i].value), ToString(pairs[j].value)
			}

			if !caseSensitive {
				a, b = strings.ToLower(a), strings.ToLower(b)
			}

			if reverse {
				return a > b
			}
			return a < b
		})

		result := make([][]interface{}, len(pairs))
		for i, pair := range pairs {
			result[i] = []interface{}{pair.key, pair.value}
		}

		return result, nil

	default:
		return nil, fmt.Errorf("dictsort filter requires a mapping")
	}
}

// GroupByFilter groups sequence items by attribute
func GroupByFilter(value interface{}, args ...interface{}) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("groupby filter requires attribute name")
	}

	attrName := ToString(args[0])

	switch v := value.(type) {
	case []interface{}:
		groups := make(map[string][]interface{})

		for _, item := range v {
			key := ToString(extractAttribute(item, attrName))
			groups[key] = append(groups[key], item)
		}

		// Convert to list of [key, items] pairs
		var result [][]interface{}
		keys := make([]string, 0, len(groups))
		for k := range groups {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, key := range keys {
			result = append(result, []interface{}{key, groups[key]})
		}

		return result, nil

	default:
		return nil, fmt.Errorf("groupby filter requires a sequence")
	}
}

// ToJSONFilter converts value to JSON string.
// Returns a SafeValue to prevent HTML auto-escaping, matching Jinja2's
// tojson behavior which returns Markup. This is necessary for safe
// embedding of JSON in <script> blocks where HTML entities are not decoded.
func ToJSONFilter(value interface{}, args ...interface{}) (interface{}, error) {
	indent := 0
	if len(args) > 0 {
		i, err := ToInt(args[0])
		if err == nil && i > 0 {
			indent = i
		}
	}

	var data []byte
	var err error

	if indent > 0 {
		data, err = json.MarshalIndent(value, "", strings.Repeat(" ", indent))
	} else {
		data, err = json.Marshal(value)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to convert to JSON: %v", err)
	}

	return runtime.SafeValue{Value: string(data)}, nil
}

// FromJSONFilter parses JSON string to value
func FromJSONFilter(value interface{}, args ...interface{}) (interface{}, error) {
	s := ToString(value)
	if s == "" {
		return nil, nil
	}

	var result interface{}
	err := json.Unmarshal([]byte(s), &result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	return result, nil
}

// Helper functions are defined in filter.go
