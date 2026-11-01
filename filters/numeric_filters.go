package filters

import (
	"fmt"
	"math"
	"math/rand/v2"
	"reflect"
	"strconv"
	"strings"

	"github.com/zipreport/miya/runtime"
)

// AbsFilter returns the absolute value of a number
func AbsFilter(value interface{}, args ...interface{}) (interface{}, error) {
	switch v := value.(type) {
	case int:
		if v < 0 {
			return -v, nil
		}
		return v, nil
	case int8:
		if v < 0 {
			return -v, nil
		}
		return v, nil
	case int16:
		if v < 0 {
			return -v, nil
		}
		return v, nil
	case int32:
		if v < 0 {
			return -v, nil
		}
		return v, nil
	case int64:
		if v < 0 {
			return -v, nil
		}
		return v, nil
	case float32:
		return math.Abs(float64(v)), nil
	case float64:
		return math.Abs(v), nil
	default:
		// Try to convert to float
		f, err := ToFloat(value)
		if err != nil {
			return nil, fmt.Errorf("abs filter requires a number: %v", err)
		}
		return math.Abs(f), nil
	}
}

// RoundFilter rounds a number to specified precision. A precision of 1 or more
// returns a runtime.RoundedFloat, a number that prints with that many decimals.
func RoundFilter(value interface{}, args ...interface{}) (interface{}, error) {
	precision := 0
	method := "common"

	if len(args) > 0 && args[0] != nil {
		p, err := ToInt(args[0])
		if err != nil {
			return nil, fmt.Errorf("round precision must be integer: %v", err)
		}
		precision = p
	}
	if len(args) > 1 && args[1] != nil {
		method = ToString(args[1])
	}
	if method != "common" && method != "ceil" && method != "floor" {
		return nil, fmt.Errorf("round: method must be common, ceil or floor")
	}

	f, err := ToFloat(value)
	if err != nil {
		return nil, fmt.Errorf("round filter requires a number: %v", err)
	}

	if precision > 308 || math.IsInf(f, 0) || math.IsNaN(f) {
		return f, nil
	}
	if precision < -308 {
		return math.Copysign(0, f), nil
	}
	m := math.Pow(10, float64(precision))
	scaled := f * m
	if math.IsInf(scaled, 0) {
		return f, nil
	}
	var result float64
	switch {
	case math.Abs(scaled) >= 1<<53:
		result = f
	case method == "ceil":
		result = math.Ceil(scaled) / m
	case method == "floor":
		result = math.Floor(scaled) / m
	default:
		result = math.Round(scaled) / m
	}

	if precision <= 0 || (precision == 1 && result == math.Trunc(result)) {
		return result, nil
	}
	return runtime.RoundedFloat{Value: result, Places: precision}, nil
}

// IntFilter converts value to integer, following Jinja2's do_int.
func IntFilter(value interface{}, args ...interface{}) (interface{}, error) {
	var defaultValue interface{} = 0
	base := 10

	if len(args) > 0 {
		defaultValue = args[0]
	}
	if len(args) > 1 {
		b, err := ToInt(args[1])
		if err != nil {
			return nil, fmt.Errorf("int base must be integer: %v", err)
		}
		base = b
	}

	if r, ok := value.(runtime.RoundedFloat); ok {
		value = r.Value
	}

	value = runtime.BasicValue(value)
	if n, ok := runtime.Number(value); ok {
		value = n
	}

	switch v := value.(type) {
	case int, int8, int16, int32, int64:
		return reflect.ValueOf(v).Int(), nil
	case uint, uint8, uint16, uint32, uint64:
		return int64(reflect.ValueOf(v).Uint()), nil
	case float32, float64:
		return floatToInt(reflect.ValueOf(v).Float(), defaultValue), nil
	case string:
		if n, ok := pyParseInt(v, base); ok {
			return n, nil
		}
		if f, ok := pyParseFloat(v); ok {
			return floatToInt(f, defaultValue), nil
		}
		return defaultValue, nil
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	default:
		return defaultValue, nil
	}
}

// floatToInt truncates f toward zero, or returns def when f is NaN, infinite
// or outside the int64 range.
func floatToInt(f float64, def interface{}) interface{} {
	if math.IsNaN(f) || math.IsInf(f, 0) || f < -9223372036854775808 || f >= 9223372036854775808 {
		return def
	}
	return int64(math.Trunc(f))
}

// FloatFilter converts value to float
func FloatFilter(value interface{}, args ...interface{}) (interface{}, error) {
	defaultValue := 0.0

	if len(args) > 0 {
		d, err := ToFloat(args[0])
		if err != nil {
			return nil, fmt.Errorf("float default value must be number: %v", err)
		}
		defaultValue = d
	}

	if r, ok := value.(runtime.RoundedFloat); ok {
		value = r.Value
	}

	value = runtime.BasicValue(value)
	if n, ok := runtime.Number(value); ok {
		value = n
	}

	switch v := value.(type) {
	case int, int8, int16, int32, int64:
		return float64(reflect.ValueOf(v).Int()), nil
	case uint, uint8, uint16, uint32, uint64:
		return float64(reflect.ValueOf(v).Uint()), nil
	case float32, float64:
		return reflect.ValueOf(v).Float(), nil
	case string:
		f, ok := pyParseFloat(v)
		if !ok {
			return defaultValue, nil
		}
		return f, nil
	case bool:
		if v {
			return 1.0, nil
		}
		return 0.0, nil
	default:
		return defaultValue, nil
	}
}

// SumFilter sums numeric values in a sequence
func SumFilter(value interface{}, args ...interface{}) (interface{}, error) {
	start := 0.0
	attribute := ""

	if len(args) > 0 && args[0] != nil {
		switch reflect.ValueOf(args[0]).Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64:
			return nil, fmt.Errorf("sum: the first positional argument is the attribute; pass a start value as start=%v", args[0])
		}
		attribute = ToString(args[0])
	}

	if len(args) > 1 {
		s, err := ToFloat(args[1])
		if err != nil {
			return nil, fmt.Errorf("sum start value must be number: %v", err)
		}
		start = s
	}

	sum := start
	value = mapAsKeys(value)

	switch v := value.(type) {
	case []interface{}:
		for _, item := range v {
			f, err := sumItem(item, attribute)
			if err != nil {
				return nil, err
			}
			sum += f
		}
		return sum, nil

	case []string:
		for _, item := range v {
			f, err := sumItem(item, "")
			if err != nil {
				return nil, err
			}
			sum += f
		}
		return sum, nil

	default:
		// Try reflection for other slice types
		rv := reflect.ValueOf(runtime.Indirect(value))
		if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
			for i := 0; i < rv.Len(); i++ {
				f, err := sumItem(rv.Index(i).Interface(), attribute)
				if err != nil {
					return nil, err
				}
				sum += f
			}
			return sum, nil
		}

		return nil, fmt.Errorf("sum filter requires a sequence")
	}
}

// sumItem is what an item (or its attribute) adds to a sum. Undefined and
// missing attributes add nothing, like miya's lenient undefined in `+`; None
// and other non-numbers are a TypeError, as in Jinja2.
func sumItem(item interface{}, attribute string) (float64, error) {
	if attribute != "" {
		v, found := lookupAttribute(item, attribute)
		if !found {
			return 0, nil
		}
		item = v
	}
	if _, ok := item.(*runtime.Undefined); ok {
		return 0, nil
	}
	f, err := ToFloat(item)
	if item == nil || err != nil {
		name := "NoneType"
		if item != nil {
			name = pyTypeName(item)
		}
		return 0, fmt.Errorf("sum: unsupported operand type(s) for +: 'int' and '%s'", name)
	}
	return f, nil
}

// MinFilter returns the smallest item of a sequence, like Jinja2's min. With
// attribute it still returns the item. None and missing keys are skipped; an
// empty sequence (or only None keys) gives undefined.
func MinFilter(value interface{}, args ...interface{}) (interface{}, error) {
	return extremum("min", -1, value, args)
}

// MaxFilter returns the largest item of a sequence; see MinFilter.
func MaxFilter(value interface{}, args ...interface{}) (interface{}, error) {
	return extremum("max", 1, value, args)
}

func extremum(name string, want int, value interface{}, args []interface{}) (interface{}, error) {
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
		return nil, fmt.Errorf("%s filter requires a sequence", name)
	}

	var best, bestKey interface{}
	found := false
	for _, item := range seq {
		key := item
		if attribute != "" {
			key = extractAttribute(item, attribute)
		}
		if key == nil {
			continue
		}
		if !found {
			best, bestKey, found = item, key, true
			continue
		}
		c, err := compareItems(key, bestKey, caseSensitive)
		if err != nil {
			return nil, wrapAttributeError(name, attribute, err)
		}
		if c == want {
			best, bestKey = item, key
		}
	}
	if !found {
		return emptyUndefined("No aggregated item, sequence was empty."), nil
	}
	return best, nil
}

// RandomFilter returns a random item of a sequence or character of a string;
// an empty one gives undefined, as in Jinja2.
func RandomFilter(value interface{}, args ...interface{}) (interface{}, error) {
	var items []interface{}
	switch v := runtime.BasicValue(value).(type) {
	case string:
		runes := []rune(v)
		if len(runes) == 0 {
			return emptyUndefined("No random item, sequence was empty."), nil
		}
		return string(runes[rand.IntN(len(runes))]), nil
	case []interface{}:
		items = v
	default:
		if !isListKind(reflect.ValueOf(v).Kind()) {
			return nil, fmt.Errorf("random filter requires a sequence")
		}
		items, _ = toInterfaceSlice(v)
	}
	if len(items) == 0 {
		return emptyUndefined("No random item, sequence was empty."), nil
	}
	return items[rand.IntN(len(items))], nil
}

// emptyUndefined is the undefined result of a filter given an empty sequence;
// the evaluator gives it the environment's undefined behavior.
func emptyUndefined(hint string) *runtime.Undefined {
	u := runtime.NewUndefined("", runtime.UndefinedSilent, nil)
	u.Hint = hint
	return u
}

// CeilFilter returns the ceiling (smallest integer greater than or equal to) of a number
func CeilFilter(value interface{}, args ...interface{}) (interface{}, error) {
	f, err := ToFloat(value)
	if err != nil {
		return nil, fmt.Errorf("ceil filter requires a number: %v", err)
	}

	return math.Ceil(f), nil
}

// FloorFilter returns the floor (largest integer less than or equal to) of a number
func FloorFilter(value interface{}, args ...interface{}) (interface{}, error) {
	f, err := ToFloat(value)
	if err != nil {
		return nil, fmt.Errorf("floor filter requires a number: %v", err)
	}

	return math.Floor(f), nil
}

// PowFilter returns the power of a number (value ** exponent)
func PowFilter(value interface{}, args ...interface{}) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("pow filter requires exponent argument")
	}

	base, err := ToFloat(value)
	if err != nil {
		return nil, fmt.Errorf("pow filter base must be number: %v", err)
	}

	exponent, err := ToFloat(args[0])
	if err != nil {
		return nil, fmt.Errorf("pow filter exponent must be number: %v", err)
	}

	return math.Pow(base, exponent), nil
}

// CurrencyFilter formats a number as currency with proper thousands separators
func CurrencyFilter(value interface{}, args ...interface{}) (interface{}, error) {
	// Default currency options
	symbol := "$"
	decimals := 2
	separator := ","

	// Parse arguments
	if len(args) > 0 {
		symbol = ToString(args[0])
	}
	if len(args) > 1 {
		d, err := ToInt(args[1])
		if err != nil {
			return nil, fmt.Errorf("currency decimals must be integer: %v", err)
		}
		decimals = d
	}
	if len(args) > 2 {
		separator = ToString(args[2])
	}

	// Convert to float
	f, err := ToFloat(value)
	if err != nil {
		return nil, fmt.Errorf("currency filter requires a number: %v", err)
	}
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return nil, fmt.Errorf("currency filter requires a finite number")
	}

	// Format with specified decimals
	formatted := fmt.Sprintf("%."+fmt.Sprintf("%d", decimals)+"f", f)

	// Split integer and decimal parts
	parts := strings.Split(formatted, ".")
	sign, integerPart := "", parts[0]
	if strings.HasPrefix(integerPart, "-") {
		sign, integerPart = "-", integerPart[1:]
	}

	// Add thousands separators to integer part
	if len(integerPart) > 3 {
		var result strings.Builder
		for i, digit := range integerPart {
			if i > 0 && (len(integerPart)-i)%3 == 0 {
				result.WriteString(separator)
			}
			result.WriteRune(digit)
		}
		integerPart = result.String()
	}
	integerPart = sign + integerPart

	// Reconstruct the number
	if decimals > 0 {
		return symbol + integerPart + "." + parts[1], nil
	}
	return symbol + integerPart, nil
}

// FormatNumberFilter formats a number with thousands separators
func FormatNumberFilter(value interface{}, args ...interface{}) (interface{}, error) {
	// Default formatting options
	decimals := -1 // -1 means preserve original decimals
	separator := ","

	// Parse arguments
	if len(args) > 0 {
		d, err := ToInt(args[0])
		if err != nil {
			return nil, fmt.Errorf("format_number decimals must be integer: %v", err)
		}
		decimals = d
	}
	if len(args) > 1 {
		separator = ToString(args[1])
	}

	// Convert to float
	f, err := ToFloat(value)
	if err != nil {
		return nil, fmt.Errorf("format_number filter requires a number: %v", err)
	}
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return nil, fmt.Errorf("format_number filter requires a finite number")
	}

	// Format with specified or preserved decimals
	var formatted string
	if decimals >= 0 {
		formatted = fmt.Sprintf("%."+fmt.Sprintf("%d", decimals)+"f", f)
	} else {
		formatted = strconv.FormatFloat(f, 'f', -1, 64)
	}

	// Split integer and decimal parts
	parts := strings.Split(formatted, ".")
	sign, integerPart := "", parts[0]
	if strings.HasPrefix(integerPart, "-") {
		sign, integerPart = "-", integerPart[1:]
	}

	// Add thousands separators to integer part
	if len(integerPart) > 3 {
		var result strings.Builder
		for i, digit := range integerPart {
			if i > 0 && (len(integerPart)-i)%3 == 0 {
				result.WriteString(separator)
			}
			result.WriteRune(digit)
		}
		integerPart = result.String()
	}
	integerPart = sign + integerPart

	// Reconstruct the number
	if len(parts) > 1 {
		return integerPart + "." + parts[1], nil
	}
	return integerPart, nil
}

func parseFiniteFloat(s string) (float64, error) {
	f, err := strconv.ParseFloat(s, 64)
	if err == nil && (math.IsInf(f, 0) || math.IsNaN(f)) {
		return 0, fmt.Errorf("cannot convert %q to float", s)
	}
	return f, err
}

// Helper function for numeric filters
func ToFloat(value interface{}) (float64, error) {
	switch v := value.(type) {
	case runtime.RoundedFloat:
		return v.Value, nil
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
		return parseFiniteFloat(v)
	case bool:
		if v {
			return 1.0, nil
		}
		return 0.0, nil
	default:
		switch bv := runtime.BasicValue(value).(type) {
		case string:
			return parseFiniteFloat(bv)
		case bool:
			if bv {
				return 1.0, nil
			}
			return 0.0, nil
		}
		if n, ok := numberNoBool(value); ok {
			switch x := n.(type) {
			case int64:
				return float64(x), nil
			case uint64:
				return float64(x), nil
			case float64:
				return x, nil
			}
		}
		return 0, fmt.Errorf("cannot convert %T to float64", value)
	}
}

// numberNoBool is runtime.Number without bool kinds, for the ToInt and
// ToFloat fallbacks: a bool is not an integer in ToInt.
func numberNoBool(value interface{}) (interface{}, bool) {
	v := runtime.Indirect(value)
	if v == nil || reflect.TypeOf(v).Kind() == reflect.Bool {
		return nil, false
	}
	return runtime.Number(v)
}
