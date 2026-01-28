package filters

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// FirstFilter returns the first item in a sequence
func FirstFilter(value interface{}, args ...interface{}) (interface{}, error) {
	if value == nil {
		return nil, nil
	}

	switch v := value.(type) {
	case []interface{}:
		if len(v) > 0 {
			return v[0], nil
		}
	case []string:
		if len(v) > 0 {
			return v[0], nil
		}
	case []int:
		if len(v) > 0 {
			return v[0], nil
		}
	case []int64:
		if len(v) > 0 {
			return v[0], nil
		}
	case []float64:
		if len(v) > 0 {
			return v[0], nil
		}
	case []bool:
		if len(v) > 0 {
			return v[0], nil
		}
	case string:
		if len(v) > 0 {
			return string(v[0]), nil
		}
	default:
		// Use reflection for other slice types
		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
			if rv.Len() > 0 {
				return rv.Index(0).Interface(), nil
			}
		}
	}

	return nil, nil
}

// LastFilter returns the last item in a sequence
func LastFilter(value interface{}, args ...interface{}) (interface{}, error) {
	if value == nil {
		return nil, nil
	}

	switch v := value.(type) {
	case []interface{}:
		if len(v) > 0 {
			return v[len(v)-1], nil
		}
	case []string:
		if len(v) > 0 {
			return v[len(v)-1], nil
		}
	case []int:
		if len(v) > 0 {
			return v[len(v)-1], nil
		}
	case []int64:
		if len(v) > 0 {
			return v[len(v)-1], nil
		}
	case []float64:
		if len(v) > 0 {
			return v[len(v)-1], nil
		}
	case []bool:
		if len(v) > 0 {
			return v[len(v)-1], nil
		}
	case string:
		if len(v) > 0 {
			return string(v[len(v)-1]), nil
		}
	default:
		// Use reflection for other slice types
		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
			if rv.Len() > 0 {
				return rv.Index(rv.Len() - 1).Interface(), nil
			}
		}
	}

	return nil, nil
}

// LengthFilter returns the length of a sequence or mapping
func LengthFilter(value interface{}, args ...interface{}) (interface{}, error) {
	if value == nil {
		return 0, nil
	}

	switch v := value.(type) {
	case string:
		return len([]rune(v)), nil // Count Unicode characters, not bytes
	case []interface{}:
		return len(v), nil
	case []string:
		return len(v), nil
	case []int:
		return len(v), nil
	case []int64:
		return len(v), nil
	case []float64:
		return len(v), nil
	case []bool:
		return len(v), nil
	case map[string]interface{}:
		return len(v), nil
	case map[string]string:
		return len(v), nil
	case map[int]interface{}:
		return len(v), nil
	default:
		// Use reflection for other types
		rv := reflect.ValueOf(value)
		switch rv.Kind() {
		case reflect.Slice, reflect.Array, reflect.Map, reflect.Chan:
			return rv.Len(), nil
		case reflect.String:
			return len([]rune(rv.String())), nil
		}
	}

	return 0, fmt.Errorf("object of type %T has no len()", value)
}

// JoinFilter joins sequence elements with separator
func JoinFilter(value interface{}, args ...interface{}) (interface{}, error) {
	separator := ""
	if len(args) > 0 {
		separator = ToString(args[0])
	}

	attribute := ""
	if len(args) > 1 {
		attribute = ToString(args[1])
	}

	items, err := makeStringSlice(value, attribute)
	if err != nil {
		return nil, err
	}

	return strings.Join(items, separator), nil
}

// SortFilter sorts a sequence
func SortFilter(value interface{}, args ...interface{}) (interface{}, error) {
	reverse := false
	caseSensitive := false
	attribute := ""

	if len(args) > 0 {
		reverse = ToBool(args[0])
	}
	if len(args) > 1 {
		caseSensitive = ToBool(args[1])
	}
	if len(args) > 2 {
		attribute = ToString(args[2])
	}

	switch v := value.(type) {
	case []interface{}:
		// Make a copy to avoid modifying original
		sorted := make([]interface{}, len(v))
		copy(sorted, v)

		sort.Slice(sorted, func(i, j int) bool {
			a := sorted[i]
			b := sorted[j]

			// Extract attribute if specified
			if attribute != "" {
				a = extractAttribute(a, attribute)
				b = extractAttribute(b, attribute)
			}

			result := compareValues(a, b, caseSensitive)
			if reverse {
				return result > 0
			}
			return result < 0
		})

		return sorted, nil

	case []string:
		// Make a copy
		sorted := make([]string, len(v))
		copy(sorted, v)

		if caseSensitive {
			sort.Strings(sorted)
		} else {
			sort.Slice(sorted, func(i, j int) bool {
				return strings.ToLower(sorted[i]) < strings.ToLower(sorted[j])
			})
		}

		if reverse {
			reverseStringSlice(sorted)
		}

		return sorted, nil

	default:
		return nil, fmt.Errorf("sort filter requires a sequence")
	}
}

// ReverseFilter reverses a sequence
func ReverseFilter(value interface{}, args ...interface{}) (interface{}, error) {
	switch v := value.(type) {
	case []interface{}:
		reversed := make([]interface{}, len(v))
		for i, item := range v {
			reversed[len(v)-1-i] = item
		}
		return reversed, nil

	case []string:
		reversed := make([]string, len(v))
		for i, item := range v {
			reversed[len(v)-1-i] = item
		}
		return reversed, nil

	case string:
		runes := []rune(v)
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		return string(runes), nil

	default:
		return nil, fmt.Errorf("reverse filter requires a sequence")
	}
}

// UniqueFilter removes duplicate items from sequence
func UniqueFilter(value interface{}, args ...interface{}) (interface{}, error) {
	caseSensitive := true
	attribute := ""

	if len(args) > 0 {
		caseSensitive = ToBool(args[0])
	}
	if len(args) > 1 {
		attribute = ToString(args[1])
	}

	switch v := value.(type) {
	case []interface{}:
		// Pre-allocate with capacity for typical case (most items unique)
		result := make([]interface{}, 0, len(v))
		seen := make(map[string]bool, len(v))

		for _, item := range v {
			key := item
			if attribute != "" {
				key = extractAttribute(item, attribute)
			}

			keyStr := ToString(key)
			if !caseSensitive {
				keyStr = strings.ToLower(keyStr)
			}

			if !seen[keyStr] {
				seen[keyStr] = true
				result = append(result, item)
			}
		}

		return result, nil

	case []string:
		// Pre-allocate with capacity
		result := make([]string, 0, len(v))
		seen := make(map[string]bool, len(v))

		for _, item := range v {
			key := item
			if !caseSensitive {
				key = strings.ToLower(item)
			}

			if !seen[key] {
				seen[key] = true
				result = append(result, item)
			}
		}

		return result, nil

	default:
		return nil, fmt.Errorf("unique filter requires a sequence")
	}
}

// SliceFilter returns a slice of the sequence
func SliceFilter(value interface{}, args ...interface{}) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("slice filter requires at least one argument")
	}

	start, err := ToInt(args[0])
	if err != nil {
		return nil, fmt.Errorf("slice start must be integer: %v", err)
	}

	var end *int
	if len(args) > 1 {
		if args[1] != nil {
			e, err := ToInt(args[1])
			if err != nil {
				return nil, fmt.Errorf("slice end must be integer: %v", err)
			}
			end = &e
		}
	}

	switch v := value.(type) {
	case []interface{}:
		length := len(v)
		if start < 0 {
			start = length + start
		}
		if start < 0 {
			start = 0
		}
		if start >= length {
			return []interface{}{}, nil
		}

		endIdx := length
		if end != nil {
			endIdx = *end
			if endIdx < 0 {
				endIdx = length + endIdx
			}
			if endIdx > length {
				endIdx = length
			}
		}

		if endIdx <= start {
			return []interface{}{}, nil
		}

		return v[start:endIdx], nil

	case []string:
		length := len(v)
		if start < 0 {
			start = length + start
		}
		if start < 0 {
			start = 0
		}
		if start >= length {
			return []string{}, nil
		}

		endIdx := length
		if end != nil {
			endIdx = *end
			if endIdx < 0 {
				endIdx = length + endIdx
			}
			if endIdx > length {
				endIdx = length
			}
		}

		if endIdx <= start {
			return []string{}, nil
		}

		return v[start:endIdx], nil

	case string:
		runes := []rune(v)
		length := len(runes)
		if start < 0 {
			start = length + start
		}
		if start < 0 {
			start = 0
		}
		if start >= length {
			return "", nil
		}

		endIdx := length
		if end != nil {
			endIdx = *end
			if endIdx < 0 {
				endIdx = length + endIdx
			}
			if endIdx > length {
				endIdx = length
			}
		}

		if endIdx <= start {
			return "", nil
		}

		return string(runes[start:endIdx]), nil

	default:
		return nil, fmt.Errorf("slice filter requires a sequence")
	}
}

// BatchFilter creates batches of items
func BatchFilter(value interface{}, args ...interface{}) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("batch filter requires size argument")
	}

	size, err := ToInt(args[0])
	if err != nil {
		return nil, fmt.Errorf("batch size must be integer: %v", err)
	}

	if size <= 0 {
		return nil, fmt.Errorf("batch size must be positive")
	}

	fillWith := interface{}(nil)
	if len(args) > 1 {
		fillWith = args[1]
	}

	switch v := value.(type) {
	case []interface{}:
		var batches [][]interface{}

		for i := 0; i < len(v); i += size {
			end := i + size
			if end > len(v) {
				end = len(v)
			}

			batch := make([]interface{}, size)
			copy(batch, v[i:end])

			// Fill remaining slots if needed
			if fillWith != nil {
				for j := end - i; j < size; j++ {
					batch[j] = fillWith
				}
			} else {
				batch = batch[:end-i]
			}

			batches = append(batches, batch)
		}

		return batches, nil

	default:
		return nil, fmt.Errorf("batch filter requires a sequence")
	}
}

// ListFilter converts value to list
func ListFilter(value interface{}, args ...interface{}) (interface{}, error) {
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
	case string:
		result := make([]interface{}, len([]rune(v)))
		for i, r := range []rune(v) {
			result[i] = string(r)
		}
		return result, nil
	case map[string]interface{}:
		result := make([]interface{}, 0, len(v))
		for k := range v {
			result = append(result, k)
		}
		sort.Slice(result, func(i, j int) bool {
			return ToString(result[i]) < ToString(result[j])
		})
		return result, nil
	default:
		// Try reflection
		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
			result := make([]interface{}, rv.Len())
			for i := 0; i < rv.Len(); i++ {
				result[i] = rv.Index(i).Interface()
			}
			return result, nil
		}

		// Single item becomes single-item list
		return []interface{}{value}, nil
	}
}

// SelectAttrFilter selects items with a specific attribute value
func SelectAttrFilter(value interface{}, args ...interface{}) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("selectattr filter requires attribute name")
	}

	attrName := ToString(args[0])
	var testValue interface{}

	if len(args) > 1 {
		testValue = args[1]
	}

	switch v := value.(type) {
	case []interface{}:
		var result []interface{}

		for _, item := range v {
			attr := extractAttribute(item, attrName)

			if len(args) == 1 {
				// Just check if attribute is truthy
				if ToBool(attr) {
					result = append(result, item)
				}
			} else {
				// Check if attribute equals test value
				if reflect.DeepEqual(attr, testValue) {
					result = append(result, item)
				}
			}
		}

		return result, nil

	default:
		return nil, fmt.Errorf("selectattr filter requires a sequence")
	}
}

// RejectAttrFilter rejects items with a specific attribute value
func RejectAttrFilter(value interface{}, args ...interface{}) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("rejectattr filter requires attribute name")
	}

	attrName := ToString(args[0])
	var testValue interface{}

	if len(args) > 1 {
		testValue = args[1]
	}

	switch v := value.(type) {
	case []interface{}:
		var result []interface{}

		for _, item := range v {
			attr := extractAttribute(item, attrName)

			if len(args) == 1 {
				// Just check if attribute is falsy
				if !ToBool(attr) {
					result = append(result, item)
				}
			} else {
				// Check if attribute doesn't equal test value
				if !reflect.DeepEqual(attr, testValue) {
					result = append(result, item)
				}
			}
		}

		return result, nil

	default:
		return nil, fmt.Errorf("rejectattr filter requires a sequence")
	}
}

// Helper functions

func makeStringSlice(value interface{}, attribute string) ([]string, error) {
	switch v := value.(type) {
	case []interface{}:
		result := make([]string, len(v))
		for i, item := range v {
			if attribute != "" {
				item = extractAttribute(item, attribute)
			}
			result[i] = ToString(item)
		}
		return result, nil

	case []string:
		if attribute == "" {
			return v, nil
		}
		// If attribute is specified, we can't extract it from strings
		return nil, fmt.Errorf("cannot extract attribute from string sequence")

	default:
		return nil, fmt.Errorf("join filter requires a sequence")
	}
}

func extractAttribute(obj interface{}, attribute string) interface{} {
	if obj == nil {
		return nil
	}

	switch v := obj.(type) {
	case map[string]interface{}:
		return v[attribute]
	case map[string]string:
		return v[attribute]
	default:
		// Use reflection for struct fields
		rv := reflect.ValueOf(obj)
		if rv.Kind() == reflect.Ptr {
			rv = rv.Elem()
		}

		if rv.Kind() == reflect.Struct {
			field := rv.FieldByName(attribute)
			if field.IsValid() && field.CanInterface() {
				return field.Interface()
			}

			// Try with capitalized field name
			field = rv.FieldByName(capitalizeFirst(attribute))
			if field.IsValid() && field.CanInterface() {
				return field.Interface()
			}
		}

		return nil
	}
}

func compareValues(a, b interface{}, caseSensitive bool) int {
	// Convert to strings for comparison
	aStr := ToString(a)
	bStr := ToString(b)

	if !caseSensitive {
		aStr = strings.ToLower(aStr)
		bStr = strings.ToLower(bStr)
	}

	if aStr < bStr {
		return -1
	} else if aStr > bStr {
		return 1
	}
	return 0
}

func reverseStringSlice(slice []string) {
	for i, j := 0, len(slice)-1; i < j; i, j = i+1, j-1 {
		slice[i], slice[j] = slice[j], slice[i]
	}
}

// ItemsFilter returns (key, value) pairs for dictionaries
func ItemsFilter(value interface{}, args ...interface{}) (interface{}, error) {
	if value == nil {
		return []interface{}{}, nil
	}

	switch v := value.(type) {
	case map[string]interface{}:
		var items []interface{}
		for key, val := range v {
			items = append(items, []interface{}{key, val})
		}
		// Sort by key for consistent output
		sort.Slice(items, func(i, j int) bool {
			return items[i].([]interface{})[0].(string) < items[j].([]interface{})[0].(string)
		})
		return items, nil

	case map[string]string:
		var items []interface{}
		for key, val := range v {
			items = append(items, []interface{}{key, val})
		}
		// Sort by key for consistent output
		sort.Slice(items, func(i, j int) bool {
			return items[i].([]interface{})[0].(string) < items[j].([]interface{})[0].(string)
		})
		return items, nil

	default:
		return nil, fmt.Errorf("items filter requires a mapping")
	}
}

// KeysFilter returns keys from a dictionary
func KeysFilter(value interface{}, args ...interface{}) (interface{}, error) {
	if value == nil {
		return []interface{}{}, nil
	}

	switch v := value.(type) {
	case map[string]interface{}:
		var keys []interface{}
		for key := range v {
			keys = append(keys, key)
		}
		// Sort for consistent output
		sort.Slice(keys, func(i, j int) bool {
			return keys[i].(string) < keys[j].(string)
		})
		return keys, nil

	case map[string]string:
		var keys []interface{}
		for key := range v {
			keys = append(keys, key)
		}
		// Sort for consistent output
		sort.Slice(keys, func(i, j int) bool {
			return keys[i].(string) < keys[j].(string)
		})
		return keys, nil

	default:
		return nil, fmt.Errorf("keys filter requires a mapping")
	}
}

// ValuesFilter returns values from a dictionary
func ValuesFilter(value interface{}, args ...interface{}) (interface{}, error) {
	if value == nil {
		return []interface{}{}, nil
	}

	switch v := value.(type) {
	case map[string]interface{}:
		var values []interface{}
		// Get keys first and sort them for consistent order
		var keys []string
		for key := range v {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		// Add values in sorted key order
		for _, key := range keys {
			values = append(values, v[key])
		}
		return values, nil

	case map[string]string:
		var values []interface{}
		// Get keys first and sort them for consistent order
		var keys []string
		for key := range v {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		// Add values in sorted key order
		for _, key := range keys {
			values = append(values, v[key])
		}
		return values, nil

	default:
		return nil, fmt.Errorf("values filter requires a mapping")
	}
}

// ZipFilter combines multiple sequences into tuples
func ZipFilter(value interface{}, args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("zip filter requires at least one argument")
	}

	// Convert first sequence to []interface{}
	firstSeq, err := toInterfaceSlice(value)
	if err != nil {
		return nil, fmt.Errorf("zip filter requires sequences: %v", err)
	}

	// Convert all argument sequences
	sequences := [][]interface{}{firstSeq}
	for i, arg := range args {
		seq, err := toInterfaceSlice(arg)
		if err != nil {
			return nil, fmt.Errorf("zip filter argument %d requires sequence: %v", i+1, err)
		}
		sequences = append(sequences, seq)
	}

	// Find minimum length
	minLen := len(firstSeq)
	for _, seq := range sequences[1:] {
		if len(seq) < minLen {
			minLen = len(seq)
		}
	}

	// Create tuples
	var result []interface{}
	for i := 0; i < minLen; i++ {
		tuple := make([]interface{}, len(sequences))
		for j, seq := range sequences {
			tuple[j] = seq[i]
		}
		result = append(result, tuple)
	}

	return result, nil
}

// Helper function to convert various types to []interface{}
func toInterfaceSlice(value interface{}) ([]interface{}, error) {
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
	case string:
		result := make([]interface{}, len([]rune(v)))
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

		return nil, fmt.Errorf("expected sequence, got %T", value)
	}
}
