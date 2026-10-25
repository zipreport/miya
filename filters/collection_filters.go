package filters

import (
	"fmt"
	"math"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/zipreport/miya/runtime"
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
			r, _ := utf8.DecodeRuneInString(v)
			return string(r), nil
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
			r, _ := utf8.DecodeLastRuneInString(v)
			return string(r), nil
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

// maxBatchFill bounds the number of items a batch filter will pad to.
const maxBatchFill = 1000000

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

	if fillWith != nil && size > maxBatchFill {
		return nil, fmt.Errorf("batch size too large to fill")
	}

	switch v := value.(type) {
	case []interface{}:
		var batches [][]interface{}

		for i := 0; i < len(v); i += size {
			end := i + size
			if end > len(v) {
				end = len(v)
			}

			batch := make([]interface{}, end-i, end-i)
			copy(batch, v[i:end])

			// Fill remaining slots if needed
			if fillWith != nil {
				for len(batch) < size {
					batch = append(batch, fillWith)
				}
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
//
// Deprecated: templates use the registry-bound select/reject/selectattr/rejectattr, which apply tests; this function compares values and is kept for direct Go callers.
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
//
// Deprecated: templates use the registry-bound select/reject/selectattr/rejectattr, which apply tests; this function compares values and is kept for direct Go callers.
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
	case map[string]interface{}:
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		result := make([]interface{}, len(keys))
		for i, k := range keys {
			result[i] = k
		}
		return result, nil
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

// SortFilter sorts a sequence like Jinja2's sort: numbers numerically, strings
// case-insensitively unless case_sensitive, attribute may be a path or a
// comma-separated list of paths.
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

	seq, err := toInterfaceSlice(value)
	if err != nil {
		return nil, fmt.Errorf("sort filter requires a sequence")
	}

	attrs := splitAttributes(attribute)
	keys := make([][]interface{}, len(seq))
	for i, item := range seq {
		keys[i] = sortKeys(item, attrs, caseSensitive)
	}

	order := make([]int, len(seq))
	for i := range order {
		order[i] = i
	}
	var firstErr error
	sort.SliceStable(order, func(i, j int) bool {
		if firstErr != nil {
			return false
		}
		c, err := compareKeys(keys[order[i]], keys[order[j]])
		if err != nil {
			firstErr = err
			return false
		}
		if reverse {
			return c > 0
		}
		return c < 0
	})
	if firstErr != nil {
		return nil, wrapAttributeError("sort", attribute, firstErr)
	}

	sorted := make([]interface{}, len(seq))
	for i, idx := range order {
		sorted[i] = seq[idx]
	}
	return sorted, nil
}

// UniqueFilter removes duplicate items from sequence; strings compare
// case-insensitively unless case_sensitive is set.
func UniqueFilter(value interface{}, args ...interface{}) (interface{}, error) {
	caseSensitive := false
	attribute := ""

	if len(args) > 0 {
		caseSensitive = ToBool(args[0])
	}
	if len(args) > 1 {
		attribute = ToString(args[1])
	}

	seq, err := toInterfaceSlice(value)
	if err != nil {
		return nil, fmt.Errorf("unique filter requires a sequence")
	}

	result := make([]interface{}, 0, len(seq))
	seen := make(map[string]bool, len(seq))
	for _, item := range seq {
		key := item
		if attribute != "" {
			key = extractAttribute(item, attribute)
		}
		k := uniqueKey(key, caseSensitive)
		if !seen[k] {
			seen[k] = true
			result = append(result, item)
		}
	}
	return result, nil
}

// SliceFilter splits a sequence into the given number of slices, like
// Jinja2's slice(slices, fill_with=None).
func SliceFilter(value interface{}, args ...interface{}) (interface{}, error) {
	seq, err := toInterfaceSlice(value)
	if err != nil {
		return nil, fmt.Errorf("slice filter requires a sequence")
	}
	if len(args) < 1 {
		return nil, fmt.Errorf("slice filter requires a positive number of slices")
	}
	slices, err := ToInt(args[0])
	if err != nil || slices < 1 {
		return nil, fmt.Errorf("slice filter requires a positive number of slices")
	}
	if slices > maxBatchFill {
		return nil, fmt.Errorf("slice count too large (max %d)", maxBatchFill)
	}
	var fill interface{}
	hasFill := false
	if len(args) > 1 && args[1] != nil {
		fill, hasFill = args[1], true
	}

	perSlice, extra := len(seq)/slices, len(seq)%slices
	offset := 0
	result := make([]interface{}, 0, slices)
	for n := 0; n < slices; n++ {
		start := offset + n*perSlice
		if n < extra {
			offset++
		}
		end := offset + (n+1)*perSlice
		part := append([]interface{}{}, seq[start:end]...)
		if hasFill && n >= extra {
			part = append(part, fill)
		}
		result = append(result, part)
	}
	return result, nil
}

// splitAttributes splits a comma-separated attribute list (Jinja2's
// multi-key sort), dropping blanks.
func splitAttributes(attribute string) []string {
	var attrs []string
	for _, a := range strings.Split(attribute, ",") {
		if a = strings.TrimSpace(a); a != "" {
			attrs = append(attrs, a)
		}
	}
	return attrs
}

// sortKeys computes the sort key of an item once: one value per attribute
// (nil when missing), or the item itself without attributes. Strings are
// lowercased unless caseSensitive.
func sortKeys(item interface{}, attrs []string, caseSensitive bool) []interface{} {
	fold := func(v interface{}) interface{} {
		if s, ok := v.(string); ok && !caseSensitive {
			return strings.ToLower(s)
		}
		return v
	}
	if len(attrs) == 0 {
		return []interface{}{fold(item)}
	}
	keys := make([]interface{}, len(attrs))
	for i, a := range attrs {
		keys[i] = fold(extractAttribute(item, a))
	}
	return keys
}

func compareKeys(a, b []interface{}) (int, error) {
	for i := range a {
		c, err := compareItems(a[i], b[i], true)
		if err != nil {
			return 0, err
		}
		if c != 0 {
			return c, nil
		}
	}
	return 0, nil
}

func wrapAttributeError(filterName, attribute string, err error) error {
	if attribute == "" {
		return fmt.Errorf("%s: %w", filterName, err)
	}
	return fmt.Errorf("%s: attribute %q: %w", filterName, attribute, err)
}

// compareItems orders two values like Jinja2's sort, min and max: numbers
// numerically (any Go numeric type, with true/false as 1/0), strings by code
// point (lowercased first unless caseSensitive), lists element by element.
// None (nil) and missing values sort before everything else and equal each
// other: a deliberate deviation, see the plan's "Why this matters".
// Any other pairing is an error, as in Python.
func compareItems(a, b interface{}, caseSensitive bool) (int, error) {
	if a == nil || b == nil {
		switch {
		case a == nil && b == nil:
			return 0, nil
		case a == nil:
			return -1, nil
		}
		return 1, nil
	}
	if ab, ok := a.(bool); ok {
		a = boolToInt(ab)
	}
	if bb, ok := b.(bool); ok {
		b = boolToInt(bb)
	}
	if cmp, ok := runtime.CompareNumbers(a, b); ok {
		if cmp != 2 {
			return cmp, nil
		}
		aNaN, bNaN := isNaN(a), isNaN(b)
		switch {
		case aNaN && bNaN:
			return 0, nil
		case aNaN:
			return 1, nil
		}
		return -1, nil
	}
	if as, ok := a.(string); ok {
		if bs, ok := b.(string); ok {
			if !caseSensitive {
				as, bs = strings.ToLower(as), strings.ToLower(bs)
			}
			return strings.Compare(as, bs), nil
		}
	}
	av, bv := reflect.ValueOf(a), reflect.ValueOf(b)
	if isListKind(av.Kind()) && isListKind(bv.Kind()) {
		for i := 0; i < av.Len() && i < bv.Len(); i++ {
			c, err := compareItems(av.Index(i).Interface(), bv.Index(i).Interface(), caseSensitive)
			if err != nil {
				return 0, err
			}
			if c != 0 {
				return c, nil
			}
		}
		switch {
		case av.Len() < bv.Len():
			return -1, nil
		case av.Len() > bv.Len():
			return 1, nil
		}
		return 0, nil
	}
	return 0, fmt.Errorf("'<' not supported between instances of '%s' and '%s'", pyTypeName(a), pyTypeName(b))
}

func boolToInt(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

func isNaN(v interface{}) bool {
	f, err := ToFloat(v)
	return err == nil && math.IsNaN(f)
}

func isListKind(k reflect.Kind) bool {
	return k == reflect.Slice || k == reflect.Array
}

func pyTypeName(v interface{}) string {
	switch v.(type) {
	case string:
		return "str"
	case bool:
		return "bool"
	}
	switch k := reflect.ValueOf(v).Kind(); {
	case k >= reflect.Int && k <= reflect.Uint64:
		return "int"
	case k == reflect.Float32 || k == reflect.Float64:
		return "float"
	case k == reflect.Map:
		return "dict"
	case isListKind(k):
		return "list"
	}
	return fmt.Sprintf("%T", v)
}

// uniqueKey maps a value to a string such that two values get the same key
// exactly when Jinja2's unique treats them as equal (booleans count as 1/0,
// numbers compare by value, strings lowercased unless caseSensitive).
func uniqueKey(v interface{}, caseSensitive bool) string {
	if v == nil {
		return "z:"
	}
	if b, ok := v.(bool); ok {
		v = boolToInt(b)
	}
	if _, ok := runtime.CompareNumbers(v, v); ok {
		f, _ := ToFloat(v)
		return "n:" + strconv.FormatFloat(f, 'g', -1, 64)
	}
	if s, ok := v.(string); ok {
		if !caseSensitive {
			s = strings.ToLower(s)
		}
		return "s:" + s
	}
	return "o:" + fmt.Sprintf("%T:%v", v, v)
}

// lookupAttribute resolves a Jinja2 attribute path such as "c.name" or "0"
// on obj: each dot-separated segment is a map key or struct field, and an
// all-digit segment also indexes a list. found is false when any segment is
// missing, so callers can tell a missing value from a None one.
func lookupAttribute(obj interface{}, path string) (value interface{}, found bool) {
	if path == "" {
		return nil, false
	}
	cur := obj
	for _, seg := range strings.Split(path, ".") {
		if seg == "" {
			return nil, false
		}
		cur, found = lookupSegment(cur, seg)
		if !found {
			return nil, false
		}
	}
	return cur, true
}

func lookupSegment(obj interface{}, seg string) (interface{}, bool) {
	if obj == nil {
		return nil, false
	}
	switch v := obj.(type) {
	case runtime.GroupTuple:
		switch seg {
		case "grouper", "0":
			if len(v) > 0 {
				return v[0], true
			}
		case "list", "1":
			if len(v) > 1 {
				return v[1], true
			}
		}
		return nil, false
	case map[string]interface{}:
		val, ok := v[seg]
		return val, ok
	case map[string]string:
		val, ok := v[seg]
		return val, ok
	}

	rv := reflect.ValueOf(obj)
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return nil, false
		}
		rv = rv.Elem()
	}
	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		idx, err := strconv.Atoi(seg)
		if err != nil || idx < 0 || idx >= rv.Len() || strings.TrimLeft(seg, "0123456789") != "" {
			return nil, false
		}
		return rv.Index(idx).Interface(), true
	case reflect.Struct:
		field := rv.FieldByName(seg)
		if field.IsValid() && field.CanInterface() {
			return field.Interface(), true
		}
		field = rv.FieldByName(capitalizeFirst(seg))
		if field.IsValid() && field.CanInterface() {
			return field.Interface(), true
		}
	}
	return nil, false
}

// extractAttribute returns the value at an attribute path, or nil when it is
// missing or None.
func extractAttribute(obj interface{}, attribute string) interface{} {
	v, _ := lookupAttribute(obj, attribute)
	return v
}

// mapFilter is the registered map filter. Besides the attribute forms of
// MapFilter it applies a filter by name: per item, an attribute of that name
// wins, otherwise the filter runs. With keyword arguments the filter always runs.
func (r *FilterRegistry) mapFilter(value interface{}, args ...interface{}) (interface{}, error) {
	seq, err := toInterfaceSlice(value)
	if err != nil {
		return nil, fmt.Errorf("map filter requires a sequence")
	}
	pos := args
	var kw KwArgs
	if n := len(args); n > 0 {
		if k, ok := args[n-1].(KwArgs); ok {
			kw, pos = k, args[:n-1]
		}
	}
	if len(pos) == 0 {
		if kw == nil {
			return nil, fmt.Errorf("map filter requires attribute or filter name")
		}
		return MapFilter(seq, args...)
	}

	name := ToString(pos[0])
	fn, isFilter := r.Get(name)
	if len(kw) > 0 {
		if !isFilter {
			if _, err := MapFilter(seq, args...); err != nil {
				return nil, fmt.Errorf("no filter named %q (%w)", name, err)
			}
			return nil, fmt.Errorf("no filter named %q", name)
		}
		result := make([]interface{}, len(seq))
		for i, item := range seq {
			out, err := r.ApplyKwargs(name, item, pos[1:], kw)
			if err != nil {
				return nil, err
			}
			result[i] = out
		}
		return result, nil
	}

	result := make([]interface{}, len(seq))
	for i, item := range seq {
		if v, found := lookupAttribute(item, name); found {
			result[i] = v
			continue
		}
		if !isFilter {
			result[i] = runtime.NewUndefined(name, runtime.UndefinedSilent, nil)
			continue
		}
		out, err := fn(item, pos[1:]...)
		if err != nil {
			return nil, err
		}
		result[i] = out
	}
	return result, nil
}

func (r *FilterRegistry) selectFilter(value interface{}, args ...interface{}) (interface{}, error) {
	return r.selectItems("select", true, value, args)
}

func (r *FilterRegistry) rejectFilter(value interface{}, args ...interface{}) (interface{}, error) {
	return r.selectItems("reject", false, value, args)
}

func (r *FilterRegistry) selectAttrFilter(value interface{}, args ...interface{}) (interface{}, error) {
	return r.selectAttrItems("selectattr", true, value, args)
}

func (r *FilterRegistry) rejectAttrFilter(value interface{}, args ...interface{}) (interface{}, error) {
	return r.selectAttrItems("rejectattr", false, value, args)
}

// selectItems keeps the items for which the named test (or truthiness, with
// no arguments) equals want.
func (r *FilterRegistry) selectItems(name string, want bool, value interface{}, args []interface{}) (interface{}, error) {
	seq, err := toInterfaceSlice(value)
	if err != nil {
		return nil, fmt.Errorf("%s filter requires a sequence", name)
	}
	var test func(interface{}, ...interface{}) (bool, error)
	if len(args) > 0 {
		if test, err = r.test(ToString(args[0])); err != nil {
			return nil, err
		}
	}
	result := make([]interface{}, 0, len(seq))
	for _, item := range seq {
		ok := ToBool(item)
		if test != nil {
			if ok, err = test(item, args[1:]...); err != nil {
				return nil, err
			}
		}
		if ok == want {
			result = append(result, item)
		}
	}
	return result, nil
}

// selectAttrItems is selectItems on an attribute of each item. A missing
// attribute reaches the test as undefined, a None one as None.
func (r *FilterRegistry) selectAttrItems(name string, want bool, value interface{}, args []interface{}) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("%s filter requires attribute name", name)
	}
	seq, err := toInterfaceSlice(value)
	if err != nil {
		return nil, fmt.Errorf("%s filter requires a sequence", name)
	}
	attr := ToString(args[0])
	var test func(interface{}, ...interface{}) (bool, error)
	if len(args) > 1 {
		if test, err = r.test(ToString(args[1])); err != nil {
			return nil, err
		}
	}
	result := make([]interface{}, 0, len(seq))
	for _, item := range seq {
		v, found := lookupAttribute(item, attr)
		if !found {
			v = runtime.NewUndefined(attr, runtime.UndefinedSilent, nil)
		}
		ok := found && ToBool(v)
		if test != nil {
			if ok, err = test(v, args[2:]...); err != nil {
				return nil, err
			}
		}
		if ok == want {
			result = append(result, item)
		}
	}
	return result, nil
}
