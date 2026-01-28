package filters

import (
	"fmt"
	"reflect"
	"sync"
)

type FilterFunc func(value interface{}, args ...interface{}) (interface{}, error)

type FilterRegistry struct {
	filters map[string]FilterFunc
	mutex   sync.RWMutex
}

func NewRegistry() *FilterRegistry {
	registry := &FilterRegistry{
		filters: make(map[string]FilterFunc),
	}

	// Register all built-in filters
	registry.registerBuiltinFilters()

	return registry
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
	return fn(value, args...)
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
	r.filters["selectattr"] = SelectAttrFilter
	r.filters["rejectattr"] = RejectAttrFilter
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
	r.filters["map"] = MapFilter
	r.filters["select"] = SelectFilter
	r.filters["reject"] = RejectFilter
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

	switch v := value.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprintf("%v", value)
	}
}

func ToInt(value interface{}) (int, error) {
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
		var i int
		_, err := fmt.Sscanf(v, "%d", &i)
		return i, err
	default:
		return 0, fmt.Errorf("cannot convert %v to int", reflect.TypeOf(value))
	}
}

func ToBool(value interface{}) bool {
	if value == nil {
		return false
	}

	switch v := value.(type) {
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
		case reflect.Slice, reflect.Map, reflect.Array:
			return rv.Len() > 0
		default:
			return true
		}
	}
}
