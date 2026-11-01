package filters

import (
	"fmt"
	"math"
	"reflect"
	"sync"

	"github.com/zipreport/miya/runtime"
)

type FilterFunc func(value interface{}, args ...interface{}) (interface{}, error)

// TestLookup finds an `is` test by name. Environment wires it to its test
// registry so that select, reject, selectattr and rejectattr can use tests.
type TestLookup func(name string) (func(value interface{}, args ...interface{}) (bool, error), bool)

type FilterRegistry struct {
	filters     map[string]FilterFunc
	params      map[string][]Param
	kwargsNames map[string]bool
	testLookup  TestLookup
	// builtin holds the names whose filter registerBuiltinFilters registered
	builtin map[string]bool
	mutex   sync.RWMutex
}

// SetTestLookup sets how select, reject, selectattr and rejectattr find tests.
func (r *FilterRegistry) SetTestLookup(fn TestLookup) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.testLookup = fn
}

// test finds a test by name; the lookup runs without the lock held.
func (r *FilterRegistry) test(name string) (func(value interface{}, args ...interface{}) (bool, error), error) {
	r.mutex.RLock()
	lookup := r.testLookup
	r.mutex.RUnlock()
	if lookup != nil {
		if t, ok := lookup(name); ok {
			if name == "escaped" {
				return t, nil
			}
			// Tests see plain strings for safe values, as in EvalTestNode.
			return func(value interface{}, args ...interface{}) (bool, error) {
				plain := make([]interface{}, len(args))
				for i, a := range args {
					plain[i] = runtime.Unmark(a)
				}
				return t(runtime.Unmark(value), plain...)
			}, nil
		}
	}
	return nil, fmt.Errorf("no test named %q", name)
}

func NewRegistry() *FilterRegistry {
	registry := &FilterRegistry{
		filters:     make(map[string]FilterFunc),
		params:      make(map[string][]Param, len(builtinParams)),
		kwargsNames: make(map[string]bool, len(kwargsFilters)),
	}
	for name, params := range builtinParams {
		registry.params[name] = params
	}
	for name := range kwargsFilters {
		registry.kwargsNames[name] = true
	}

	// Register all built-in filters
	registry.registerBuiltinFilters()
	registry.builtin = make(map[string]bool, len(registry.filters))
	for name := range registry.filters {
		registry.builtin[name] = true
	}

	return registry
}

// IsBuiltin reports whether the filter registered as name is one of miya's
// own; Replace keeps the mark, Unregister drops it.
func (r *FilterRegistry) IsBuiltin(name string) bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return r.builtin[name]
}

func (r *FilterRegistry) Register(name string, fn FilterFunc) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.filters[name]; exists {
		return fmt.Errorf("filter %q already registered", name)
	}
	r.filters[name] = fn
	return nil
}

// Replace swaps the function of an already registered filter, keeping its
// declared parameters.
func (r *FilterRegistry) Replace(name string, fn FilterFunc) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.filters[name] = fn
}

func (r *FilterRegistry) Get(name string) (FilterFunc, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	fn, ok := r.filters[name]
	return fn, ok
}

func (r *FilterRegistry) Apply(name string, value interface{}, args ...interface{}) (interface{}, error) {
	fn, ok := r.Get(name)
	if !ok {
		return nil, fmt.Errorf("unknown filter: %s", name)
	}
	value, done := undefinedInput(name, value)
	if done {
		return value, nil
	}
	return fn(value, args...)
}

// Built-in filters that see an undefined value as an empty sequence, as
// Jinja2's Undefined iterates as empty. max, min and random of nothing are
// undefined, so they return the value unchanged.
var (
	undefinedAsEmpty = map[string]bool{"length": true, "count": true, "join": true, "sum": true,
		"sort": true, "reverse": true, "unique": true, "list": true, "batch": true, "slice": true,
		"map": true, "select": true, "reject": true, "selectattr": true, "rejectattr": true, "groupby": true}
	undefinedPassThrough = map[string]bool{"max": true, "min": true, "random": true}
)

// undefinedInput substitutes an undefined filter input by name. done reports
// that value is already the filter's result.
func undefinedInput(name string, value interface{}) (result interface{}, done bool) {
	if _, ok := value.(*runtime.Undefined); !ok {
		return value, false
	}
	switch {
	case undefinedPassThrough[name]:
		return value, true
	case undefinedAsEmpty[name]:
		return []interface{}{}, false
	case name == "items":
		return map[string]interface{}{}, false
	}
	return value, false
}

// SetParams declares the parameter names of a registered filter so that it can
// be called with keyword arguments.
func (r *FilterRegistry) SetParams(name string, params ...Param) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.filters[name]; !exists {
		return fmt.Errorf("filter %q is not registered", name)
	}
	if r.kwargsNames[name] {
		return fmt.Errorf("filter %q takes keyword arguments directly; its parameters cannot be declared", name)
	}
	r.params[name] = append([]Param(nil), params...)
	return nil
}

// RegisterWithParams registers a filter and declares its parameters under one
// lock.
func (r *FilterRegistry) RegisterWithParams(name string, fn FilterFunc, params ...Param) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.filters[name]; exists {
		return fmt.Errorf("filter %q already registered", name)
	}
	r.filters[name] = fn
	r.params[name] = append([]Param(nil), params...)
	return nil
}

// ApplyKwargs applies a filter with keyword arguments, binding them to the
// filter's declared parameters. The filter is called without the lock held.
func (r *FilterRegistry) ApplyKwargs(name string, value interface{}, args []interface{}, kwargs map[string]interface{}) (interface{}, error) {
	if len(kwargs) == 0 {
		return r.Apply(name, value, args...)
	}

	r.mutex.RLock()
	fn, ok := r.filters[name]
	isKwargs := r.kwargsNames[name]
	params, hasParams := r.params[name]
	r.mutex.RUnlock()

	if !ok {
		return nil, fmt.Errorf("unknown filter: %s", name)
	}
	value, done := undefinedInput(name, value)
	if done {
		return value, nil
	}
	if isKwargs {
		return fn(value, append(args[:len(args):len(args)], KwArgs(kwargs))...)
	}
	if !hasParams {
		return nil, fmt.Errorf("filter %q does not accept keyword arguments; declare its parameters with Environment.SetFilterParams or AddFilterWithParams", name)
	}
	bound, err := bindKwargs(name, params, args, kwargs)
	if err != nil {
		return nil, err
	}
	return fn(value, bound...)
}

func (r *FilterRegistry) List() []string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var names []string
	for name := range r.filters {
		names = append(names, name)
	}
	return names
}

func (r *FilterRegistry) Unregister(name string) bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.filters[name]; exists {
		delete(r.filters, name)
		delete(r.params, name)
		delete(r.kwargsNames, name)
		delete(r.builtin, name)
		return true
	}
	return false
}

// registerBuiltinFilters registers all built-in filters
func (r *FilterRegistry) registerBuiltinFilters() {
	// String filters
	r.filters["upper"] = UpperFilter
	r.filters["lower"] = LowerFilter
	r.filters["capitalize"] = CapitalizeFilter
	r.filters["title"] = TitleFilter
	r.filters["trim"] = TrimFilter
	r.filters["strip"] = TrimFilter // alias
	r.filters["lstrip"] = LstripFilter
	r.filters["rstrip"] = RstripFilter
	r.filters["replace"] = ReplaceFilter
	r.filters["truncate"] = TruncateFilter
	r.filters["wordwrap"] = WordwrapFilter
	r.filters["center"] = CenterFilter
	r.filters["indent"] = IndentFilter
	r.filters["regex_replace"] = RegexReplaceFilter
	r.filters["regex_search"] = RegexSearchFilter
	r.filters["regex_findall"] = RegexFindallFilter
	r.filters["split"] = SplitFilter
	r.filters["startswith"] = StartswithFilter
	r.filters["endswith"] = EndswithFilter
	r.filters["contains"] = ContainsFilter
	r.filters["slugify"] = SlugifyFilter
	r.filters["pad_left"] = PadLeftFilter
	r.filters["pad_right"] = PadRightFilter
	r.filters["wordcount"] = WordcountFilter

	// HTML/Security filters
	r.filters["escape"] = EscapeFilter
	r.filters["e"] = EscapeFilter // alias
	r.filters["safe"] = SafeFilter
	r.filters["forceescape"] = ForceEscapeFilter
	r.filters["urlencode"] = URLEncodeFilter
	r.filters["urlize"] = UrlizeFilter
	r.filters["urlizetruncate"] = UrlizeTruncateFilter
	r.filters["urlizetarget"] = UrlizeTargetFilter
	r.filters["truncatehtml"] = TruncateHTMLFilter
	r.filters["autoescape"] = AutoEscapeFilter
	r.filters["marksafe"] = MarkSafeFilter
	r.filters["xmlattr"] = XMLAttrFilter
	r.filters["striptags"] = StripTagsFilter
	r.filters["nl2br"] = NL2BRFilter

	// Collection filters
	r.filters["first"] = FirstFilter
	r.filters["last"] = LastFilter
	r.filters["length"] = LengthFilter
	r.filters["count"] = LengthFilter // alias
	r.filters["join"] = JoinFilter
	r.filters["sort"] = SortFilter
	r.filters["reverse"] = ReverseFilter
	r.filters["unique"] = UniqueFilter
	r.filters["slice"] = SliceFilter
	r.filters["batch"] = BatchFilter
	r.filters["list"] = ListFilter
	r.filters["selectattr"] = r.selectAttrFilter
	r.filters["rejectattr"] = r.rejectAttrFilter
	r.filters["items"] = ItemsFilter
	r.filters["keys"] = KeysFilter
	r.filters["values"] = ValuesFilter
	r.filters["zip"] = ZipFilter

	// Numeric filters
	r.filters["abs"] = AbsFilter
	r.filters["round"] = RoundFilter
	r.filters["int"] = IntFilter
	r.filters["float"] = FloatFilter
	r.filters["sum"] = SumFilter
	r.filters["min"] = MinFilter
	r.filters["max"] = MaxFilter
	r.filters["ceil"] = CeilFilter
	r.filters["floor"] = FloorFilter
	r.filters["pow"] = PowFilter
	r.filters["random"] = RandomFilter
	r.filters["currency"] = CurrencyFilter
	r.filters["format_number"] = FormatNumberFilter

	// Utility filters
	r.filters["default"] = DefaultFilter
	r.filters["d"] = DefaultFilter // alias
	r.filters["map"] = r.mapFilter
	r.filters["select"] = r.selectFilter
	r.filters["reject"] = r.rejectFilter
	r.filters["attr"] = AttrFilter
	r.filters["format"] = FormatFilter
	r.filters["filesizeformat"] = FileSizeFormatFilter
	r.filters["pprint"] = PPrintFilter
	r.filters["dictsort"] = DictSortFilter
	r.filters["groupby"] = GroupByFilter

	// Date/Time filters
	r.filters["date"] = DateFilter
	r.filters["time"] = TimeFilter
	r.filters["datetime"] = DatetimeFilter
	r.filters["strftime"] = StrftimeFilter
	r.filters["timestamp"] = TimestampFilter
	r.filters["age"] = AgeFilter
	r.filters["relative_date"] = RelativeDateFilter
	r.filters["weekday"] = WeekdayFilter
	r.filters["month_name"] = MonthNameFilter

	// Utility filters
	r.filters["string"] = StringFilter
	r.filters["tojson"] = ToJSONFilter
	r.filters["fromjson"] = FromJSONFilter
}

func ToString(value interface{}) string {
	if value == nil {
		return ""
	}
	value = runtime.Indirect(value)
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case float64:
		return runtime.FormatFloat(v, 64)
	case float32:
		return runtime.FormatFloat(float64(v), 32)
	case fmt.Stringer:
		return v.String()
	default:
		return runtime.ToString(value)
	}
}

func ToInt(value interface{}) (int, error) {
	switch v := value.(type) {
	case runtime.RoundedFloat:
		return intFromFloat(v.Value)
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
		return intFromUint(uint64(v))
	case uint8:
		return int(v), nil
	case uint16:
		return int(v), nil
	case uint32:
		return int(v), nil
	case uint64:
		return intFromUint(v)
	case float32:
		return intFromFloat(float64(v))
	case float64:
		return intFromFloat(v)
	case string:
		var i int
		_, err := fmt.Sscanf(v, "%d", &i)
		return i, err
	default:
		if n, ok := numberNoBool(value); ok {
			switch x := n.(type) {
			case int64:
				return int(x), nil
			case uint64:
				return intFromUint(x)
			case float64:
				return intFromFloat(x)
			}
		}
		return 0, fmt.Errorf("cannot convert %v to int", reflect.TypeOf(value))
	}
}

// intFromFloat truncates f toward zero; a value outside int's range (or NaN) is
// an error instead of wrapping.
func intFromFloat(f float64) (int, error) {
	if !(f >= math.MinInt && f < -math.MinInt) {
		return 0, fmt.Errorf("cannot convert %v to int: out of range", f)
	}
	return int(f), nil
}

func intFromUint(u uint64) (int, error) {
	if u > math.MaxInt {
		return 0, fmt.Errorf("cannot convert %d to int: out of range", u)
	}
	return int(u), nil
}

func ToBool(value interface{}) bool {
	value = runtime.Indirect(value)
	if value == nil {
		return false
	}

	if runtime.IsUndefined(value) {
		return false
	}

	switch v := value.(type) {
	case runtime.SafeValue:
		return ToBool(v.Value)
	case *runtime.SafeValue:
		return ToBool(v.Value)
	case runtime.RoundedFloat:
		return v.Value != 0
	case bool:
		return v
	case string:
		return v != ""
	case int, int8, int16, int32, int64:
		return reflect.ValueOf(v).Int() != 0
	case uint, uint8, uint16, uint32, uint64:
		return reflect.ValueOf(v).Uint() != 0
	case float32, float64:
		return reflect.ValueOf(v).Float() != 0
	default:
		// Check if it's a slice, map, or array
		rv := reflect.ValueOf(value)
		switch rv.Kind() {
		case reflect.Bool:
			return rv.Bool()
		case reflect.String:
			return rv.Len() > 0
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return rv.Int() != 0
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			return rv.Uint() != 0
		case reflect.Float32, reflect.Float64:
			return rv.Float() != 0
		case reflect.Slice, reflect.Map, reflect.Array:
			return rv.Len() > 0
		default:
			return true
		}
	}
}
