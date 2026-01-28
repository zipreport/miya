package branching

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/zipreport/miya/runtime"
)

// TestFunc represents a test function
type TestFunc func(value interface{}, args ...interface{}) (bool, error)

// TestRegistry manages template tests
type TestRegistry struct {
	tests map[string]TestFunc
	mutex sync.RWMutex
}

// NewTestRegistry creates a new test registry
func NewTestRegistry() *TestRegistry {
	registry := &TestRegistry{
		tests: make(map[string]TestFunc),
	}

	// Register built-in tests
	registry.registerBuiltinTests()

	return registry
}

// Register registers a test function
func (r *TestRegistry) Register(name string, test TestFunc) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.tests[name]; exists {
		return fmt.Errorf("test %q already registered", name)
	}

	r.tests[name] = test
	return nil
}

// Get retrieves a test function by name
func (r *TestRegistry) Get(name string) (TestFunc, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	test, ok := r.tests[name]
	return test, ok
}

// List returns all registered test names
func (r *TestRegistry) List() []string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	names := make([]string, 0, len(r.tests))
	for name := range r.tests {
		names = append(names, name)
	}
	return names
}

// Apply applies a test to a value
func (r *TestRegistry) Apply(name string, value interface{}, args ...interface{}) (bool, error) {
	test, ok := r.Get(name)
	if !ok {
		return false, fmt.Errorf("unknown test: %s", name)
	}

	return test(value, args...)
}

// registerBuiltinTests registers all built-in test functions
func (r *TestRegistry) registerBuiltinTests() {
	// Basic existence and type tests
	r.tests["defined"] = testDefined
	r.tests["undefined"] = testUndefined
	r.tests["none"] = testNone
	r.tests["boolean"] = testBoolean
	r.tests["string"] = testString
	r.tests["number"] = testNumber
	r.tests["integer"] = testInteger
	r.tests["float"] = testFloat
	r.tests["sequence"] = testSequence
	r.tests["mapping"] = testMapping
	r.tests["iterable"] = testIterable
	r.tests["callable"] = testCallable

	// Numeric tests
	r.tests["even"] = testEven
	r.tests["odd"] = testOdd
	r.tests["divisibleby"] = testDivisibleBy

	// String tests
	r.tests["lower"] = testLower
	r.tests["upper"] = testUpper
	r.tests["startswith"] = testStartsWith
	r.tests["endswith"] = testEndsWith
	r.tests["match"] = testMatch
	r.tests["alpha"] = testAlpha
	r.tests["alnum"] = testAlnum
	r.tests["ascii"] = testAscii

	// Container tests
	r.tests["in"] = testIn
	r.tests["contains"] = testContains
	r.tests["empty"] = testEmpty

	// Identity and comparison tests
	r.tests["sameas"] = testSameAs
	r.tests["escaped"] = testEscaped

	// Comparison operators
	r.tests["eq"] = testEqual
	r.tests["ne"] = testNotEqual
	r.tests["lt"] = testLessThan
	r.tests["le"] = testLessThanOrEqual
	r.tests["gt"] = testGreaterThan
	r.tests["ge"] = testGreaterThanOrEqual
	r.tests["equalto"] = testEqual // alias
}

// Built-in test implementations

// testDefined checks if a value is defined (not nil and not undefined)
func testDefined(value interface{}, args ...interface{}) (bool, error) {
	if value == nil {
		return false, nil
	}
	// Check if value is an Undefined type
	if _, ok := value.(*runtime.Undefined); ok {
		return false, nil
	}
	return true, nil
}

// testUndefined checks if a value is undefined (nil or Undefined type)
func testUndefined(value interface{}, args ...interface{}) (bool, error) {
	if value == nil {
		return true, nil
	}
	// Check if value is an Undefined type
	if _, ok := value.(*runtime.Undefined); ok {
		return true, nil
	}
	return false, nil
}

// testNone checks if a value is None/nil
func testNone(value interface{}, args ...interface{}) (bool, error) {
	return value == nil, nil
}

// testBoolean checks if a value is a boolean
func testBoolean(value interface{}, args ...interface{}) (bool, error) {
	_, ok := value.(bool)
	return ok, nil
}

// testString checks if a value is a string
func testString(value interface{}, args ...interface{}) (bool, error) {
	_, ok := value.(string)
	return ok, nil
}

// testNumber checks if a value is a number (int or float)
func testNumber(value interface{}, args ...interface{}) (bool, error) {
	switch value.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return true, nil
	default:
		return false, nil
	}
}

// testInteger checks if a value is an integer
func testInteger(value interface{}, args ...interface{}) (bool, error) {
	switch value.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return true, nil
	default:
		return false, nil
	}
}

// testFloat checks if a value is a float
func testFloat(value interface{}, args ...interface{}) (bool, error) {
	switch value.(type) {
	case float32, float64:
		return true, nil
	default:
		return false, nil
	}
}

// testSequence checks if a value is a sequence (slice/array)
func testSequence(value interface{}, args ...interface{}) (bool, error) {
	if value == nil {
		return false, nil
	}

	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		return true, nil
	case reflect.String:
		return true, nil // strings are sequences in Jinja2
	default:
		return false, nil
	}
}

// testMapping checks if a value is a mapping (map)
func testMapping(value interface{}, args ...interface{}) (bool, error) {
	if value == nil {
		return false, nil
	}

	rv := reflect.ValueOf(value)
	return rv.Kind() == reflect.Map, nil
}

// testIterable checks if a value is iterable
func testIterable(value interface{}, args ...interface{}) (bool, error) {
	if value == nil {
		return false, nil
	}

	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Slice, reflect.Array, reflect.Map, reflect.String:
		return true, nil
	default:
		return false, nil
	}
}

// testCallable checks if a value is callable (function)
func testCallable(value interface{}, args ...interface{}) (bool, error) {
	if value == nil {
		return false, nil
	}

	rv := reflect.ValueOf(value)
	return rv.Kind() == reflect.Func, nil
}

// testEven checks if a number is even
func testEven(value interface{}, args ...interface{}) (bool, error) {
	num, err := toInt(value)
	if err != nil {
		return false, fmt.Errorf("even test requires an integer, got %T", value)
	}
	return num%2 == 0, nil
}

// testOdd checks if a number is odd
func testOdd(value interface{}, args ...interface{}) (bool, error) {
	num, err := toInt(value)
	if err != nil {
		return false, fmt.Errorf("odd test requires an integer, got %T", value)
	}
	return num%2 != 0, nil
}

// testDivisibleBy checks if a number is divisible by another number
func testDivisibleBy(value interface{}, args ...interface{}) (bool, error) {
	if len(args) != 1 {
		return false, fmt.Errorf("divisibleby test requires exactly one argument")
	}

	num, err := toInt(value)
	if err != nil {
		return false, fmt.Errorf("divisibleby test requires an integer, got %T", value)
	}

	divisor, err := toInt(args[0])
	if err != nil {
		return false, fmt.Errorf("divisibleby test requires an integer divisor, got %T", args[0])
	}

	if divisor == 0 {
		return false, fmt.Errorf("division by zero")
	}

	return num%divisor == 0, nil
}

// testLower checks if a string is all lowercase
func testLower(value interface{}, args ...interface{}) (bool, error) {
	str, ok := value.(string)
	if !ok {
		return false, fmt.Errorf("lower test requires a string, got %T", value)
	}
	return str == strings.ToLower(str), nil
}

// testUpper checks if a string is all uppercase
func testUpper(value interface{}, args ...interface{}) (bool, error) {
	str, ok := value.(string)
	if !ok {
		return false, fmt.Errorf("upper test requires a string, got %T", value)
	}
	return str == strings.ToUpper(str), nil
}

// testStartsWith checks if a string starts with a prefix
func testStartsWith(value interface{}, args ...interface{}) (bool, error) {
	if len(args) != 1 {
		return false, fmt.Errorf("startswith test requires exactly one argument")
	}

	str, ok := value.(string)
	if !ok {
		return false, fmt.Errorf("startswith test requires a string, got %T", value)
	}

	prefix, ok := args[0].(string)
	if !ok {
		return false, fmt.Errorf("startswith test requires a string prefix, got %T", args[0])
	}

	return strings.HasPrefix(str, prefix), nil
}

// testEndsWith checks if a string ends with a suffix
func testEndsWith(value interface{}, args ...interface{}) (bool, error) {
	if len(args) != 1 {
		return false, fmt.Errorf("endswith test requires exactly one argument")
	}

	str, ok := value.(string)
	if !ok {
		return false, fmt.Errorf("endswith test requires a string, got %T", value)
	}

	suffix, ok := args[0].(string)
	if !ok {
		return false, fmt.Errorf("endswith test requires a string suffix, got %T", args[0])
	}

	return strings.HasSuffix(str, suffix), nil
}

// testMatch checks if a string matches a regular expression
func testMatch(value interface{}, args ...interface{}) (bool, error) {
	if len(args) != 1 {
		return false, fmt.Errorf("match test requires exactly one argument")
	}

	str, ok := value.(string)
	if !ok {
		return false, fmt.Errorf("match test requires a string, got %T", value)
	}

	pattern, ok := args[0].(string)
	if !ok {
		return false, fmt.Errorf("match test requires a string pattern, got %T", args[0])
	}

	matched, err := regexp.MatchString(pattern, str)
	if err != nil {
		return false, fmt.Errorf("invalid regular expression: %v", err)
	}

	return matched, nil
}

// testIn checks if a value is in a container
func testIn(value interface{}, args ...interface{}) (bool, error) {
	if len(args) != 1 {
		return false, fmt.Errorf("in test requires exactly one argument")
	}

	container := args[0]
	return contains(container, value)
}

// testContains checks if a container contains a value
func testContains(value interface{}, args ...interface{}) (bool, error) {
	if len(args) != 1 {
		return false, fmt.Errorf("contains test requires exactly one argument")
	}

	item := args[0]
	return contains(value, item)
}

// testEmpty checks if a value is empty (empty string, empty container, zero, etc.)
func testEmpty(value interface{}, args ...interface{}) (bool, error) {
	if value == nil {
		return true, nil
	}

	switch v := value.(type) {
	case string:
		return len(v) == 0, nil
	case []interface{}:
		return len(v) == 0, nil
	case []string:
		return len(v) == 0, nil
	case map[string]interface{}:
		return len(v) == 0, nil
	case map[string]string:
		return len(v) == 0, nil
	case int:
		return v == 0, nil
	case int8:
		return v == 0, nil
	case int16:
		return v == 0, nil
	case int32:
		return v == 0, nil
	case int64:
		return v == 0, nil
	case uint:
		return v == 0, nil
	case uint8:
		return v == 0, nil
	case uint16:
		return v == 0, nil
	case uint32:
		return v == 0, nil
	case uint64:
		return v == 0, nil
	case float32:
		return v == 0.0, nil
	case float64:
		return v == 0.0, nil
	case bool:
		return !v, nil
	default:
		// Use reflection for other types
		rv := reflect.ValueOf(value)
		switch rv.Kind() {
		case reflect.Slice, reflect.Array, reflect.Map, reflect.Chan:
			return rv.Len() == 0, nil
		case reflect.String:
			return rv.String() == "", nil
		case reflect.Ptr, reflect.Interface:
			return rv.IsNil(), nil
		default:
			return false, nil
		}
	}
}

// testSameAs checks if two values are the same object (identity comparison)
func testSameAs(value interface{}, args ...interface{}) (bool, error) {
	if len(args) != 1 {
		return false, fmt.Errorf("sameas test requires exactly one argument")
	}

	other := args[0]

	// For basic types, sameas is the same as equality
	// For reference types, we check if they point to the same memory location
	if value == nil && other == nil {
		return true, nil
	}
	if value == nil || other == nil {
		return false, nil
	}

	// Use reflect to check if they are the same object for reference types
	v1 := reflect.ValueOf(value)
	v2 := reflect.ValueOf(other)

	// If types don't match, they can't be the same object
	if v1.Type() != v2.Type() {
		return false, nil
	}

	// For pointer, slice, map, chan, func types check if they point to the same location
	switch v1.Kind() {
	case reflect.Ptr, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func:
		if !v1.IsValid() || !v2.IsValid() {
			return v1.IsValid() == v2.IsValid(), nil
		}
		return v1.Pointer() == v2.Pointer(), nil
	default:
		// For value types, sameas is equivalent to equality
		return reflect.DeepEqual(value, other), nil
	}
}

// testEscaped checks if a value is marked as escaped/safe
func testEscaped(value interface{}, args ...interface{}) (bool, error) {
	// Check if the value is wrapped in a SafeValue (from filters package)
	// We need to check the type name since we can't import the filters package here
	// due to potential circular imports
	valueType := reflect.TypeOf(value)
	if valueType != nil {
		// Check if the type name suggests it's a safe/escaped value
		typeName := valueType.String()
		if strings.Contains(typeName, "SafeValue") || strings.Contains(typeName, "Safe") {
			return true, nil
		}
	}

	// By default, values are not considered escaped
	return false, nil
}

// testEqual checks if two values are equal
func testEqual(value interface{}, args ...interface{}) (bool, error) {
	if len(args) != 1 {
		return false, fmt.Errorf("eq test requires exactly one argument")
	}

	other := args[0]
	return reflect.DeepEqual(value, other), nil
}

// testNotEqual checks if two values are not equal
func testNotEqual(value interface{}, args ...interface{}) (bool, error) {
	if len(args) != 1 {
		return false, fmt.Errorf("ne test requires exactly one argument")
	}

	other := args[0]
	return !reflect.DeepEqual(value, other), nil
}

// testLessThan checks if value < other
func testLessThan(value interface{}, args ...interface{}) (bool, error) {
	if len(args) != 1 {
		return false, fmt.Errorf("lt test requires exactly one argument")
	}

	return compareValues(value, args[0], func(cmp int) bool { return cmp < 0 })
}

// testLessThanOrEqual checks if value <= other
func testLessThanOrEqual(value interface{}, args ...interface{}) (bool, error) {
	if len(args) != 1 {
		return false, fmt.Errorf("le test requires exactly one argument")
	}

	return compareValues(value, args[0], func(cmp int) bool { return cmp <= 0 })
}

// testGreaterThan checks if value > other
func testGreaterThan(value interface{}, args ...interface{}) (bool, error) {
	if len(args) != 1 {
		return false, fmt.Errorf("gt test requires exactly one argument")
	}

	return compareValues(value, args[0], func(cmp int) bool { return cmp > 0 })
}

// testGreaterThanOrEqual checks if value >= other
func testGreaterThanOrEqual(value interface{}, args ...interface{}) (bool, error) {
	if len(args) != 1 {
		return false, fmt.Errorf("ge test requires exactly one argument")
	}

	return compareValues(value, args[0], func(cmp int) bool { return cmp >= 0 })
}

// Helper functions

// toInt converts a value to int
func toInt(value interface{}) (int, error) {
	switch v := value.(type) {
	case int:
		return v, nil
	case int8:
		return int(v), nil
	case int16:
		return int(v), nil
	case int32:
		return int(v), nil
	case int64:
		return int(v), nil
	case uint:
		return int(v), nil
	case uint8:
		return int(v), nil
	case uint16:
		return int(v), nil
	case uint32:
		return int(v), nil
	case uint64:
		return int(v), nil
	case float32:
		return int(v), nil
	case float64:
		return int(v), nil
	case string:
		return strconv.Atoi(v)
	default:
		return 0, fmt.Errorf("cannot convert %T to int", value)
	}
}

// contains checks if a container contains an item
func contains(container, item interface{}) (bool, error) {
	if container == nil {
		return false, nil
	}

	switch v := container.(type) {
	case string:
		itemStr := fmt.Sprintf("%v", item)
		return strings.Contains(v, itemStr), nil
	case []interface{}:
		for _, elem := range v {
			if reflect.DeepEqual(elem, item) {
				return true, nil
			}
		}
		return false, nil
	case map[string]interface{}:
		itemStr := fmt.Sprintf("%v", item)
		_, ok := v[itemStr]
		return ok, nil
	default:
		// Use reflection for other slice/array types
		rv := reflect.ValueOf(container)
		switch rv.Kind() {
		case reflect.Slice, reflect.Array:
			for i := 0; i < rv.Len(); i++ {
				if reflect.DeepEqual(rv.Index(i).Interface(), item) {
					return true, nil
				}
			}
			return false, nil
		case reflect.Map:
			itemValue := reflect.ValueOf(item)
			return rv.MapIndex(itemValue).IsValid(), nil
		default:
			return false, fmt.Errorf("cannot check containment in %T", container)
		}
	}
}

// compareValues compares two values and applies a comparison function
func compareValues(v1, v2 interface{}, compFunc func(int) bool) (bool, error) {
	// Handle nil values
	if v1 == nil && v2 == nil {
		return compFunc(0), nil
	}
	if v1 == nil {
		return compFunc(-1), nil
	}
	if v2 == nil {
		return compFunc(1), nil
	}

	// Try numeric comparison first
	if num1, err1 := toFloat(v1); err1 == nil {
		if num2, err2 := toFloat(v2); err2 == nil {
			if num1 < num2 {
				return compFunc(-1), nil
			} else if num1 > num2 {
				return compFunc(1), nil
			} else {
				return compFunc(0), nil
			}
		}
	}

	// Try string comparison
	if str1, ok1 := v1.(string); ok1 {
		if str2, ok2 := v2.(string); ok2 {
			if str1 < str2 {
				return compFunc(-1), nil
			} else if str1 > str2 {
				return compFunc(1), nil
			} else {
				return compFunc(0), nil
			}
		}
	}

	// For other types, only equality comparison makes sense
	if reflect.DeepEqual(v1, v2) {
		return compFunc(0), nil
	}

	// If we can't compare, return an error for ordering comparisons
	// but allow equality comparisons to work
	if compFunc(-1) != compFunc(1) { // This is an ordering comparison
		return false, fmt.Errorf("cannot compare %T and %T", v1, v2)
	}

	// For equality comparison of uncomparable types, they're not equal
	return compFunc(1), nil // Any non-zero value works for inequality
}

// toFloat converts a value to float64 for numeric comparison
func toFloat(value interface{}) (float64, error) {
	switch v := value.(type) {
	case int:
		return float64(v), nil
	case int8:
		return float64(v), nil
	case int16:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case uint:
		return float64(v), nil
	case uint8:
		return float64(v), nil
	case uint16:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case float32:
		return float64(v), nil
	case float64:
		return v, nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to float", value)
	}
}

// testAlpha checks if a string contains only alphabetic characters
func testAlpha(value interface{}, args ...interface{}) (bool, error) {
	str, ok := value.(string)
	if !ok {
		return false, fmt.Errorf("alpha test requires a string, got %T", value)
	}

	if str == "" {
		return false, nil
	}

	for _, r := range str {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')) {
			return false, nil
		}
	}
	return true, nil
}

// testAlnum checks if a string contains only alphanumeric characters
func testAlnum(value interface{}, args ...interface{}) (bool, error) {
	str, ok := value.(string)
	if !ok {
		return false, fmt.Errorf("alnum test requires a string, got %T", value)
	}

	if str == "" {
		return false, nil
	}

	for _, r := range str {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')) {
			return false, nil
		}
	}
	return true, nil
}

// testAscii checks if a string contains only ASCII characters
func testAscii(value interface{}, args ...interface{}) (bool, error) {
	str, ok := value.(string)
	if !ok {
		return false, fmt.Errorf("ascii test requires a string, got %T", value)
	}

	for _, r := range str {
		if r > 127 {
			return false, nil
		}
	}
	return true, nil
}
