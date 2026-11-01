package runtime

import (
	"fmt"
	"math"
	"reflect"
	"sort"
)

// maxDeref bounds pointer and interface chains. A value that is still a
// pointer after maxDeref steps (a pointer type that points to itself) is
// treated as nil.
const maxDeref = 64

// Indirect follows pointers so templates see the value a pointer refers to:
// a nil pointer becomes nil (none), and a pointer to a basic value, slice,
// array or map becomes that value. Pointers to structs are kept, so field
// access still works. Pointers to funcs, channels and unsafe pointers are
// kept as they are, so a *func is never called and a *chan is never read.
func Indirect(v interface{}) interface{} {
	switch v.(type) {
	case nil, string, bool, int, int64, float64, []interface{}, map[string]interface{}:
		return v
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr {
		return v
	}
	for i := 0; i < maxDeref; i++ {
		if rv.Kind() != reflect.Ptr {
			return rv.Interface()
		}
		if rv.IsNil() {
			return nil
		}
		switch rv.Elem().Kind() {
		case reflect.Struct, reflect.Func, reflect.Chan, reflect.UnsafePointer:
			return rv.Interface()
		}
		rv = rv.Elem()
	}
	return nil
}

// basicValue converts a named bool or string type (type Flag bool, type Status
// string) to plain bool or string, after Indirect. Other values are unchanged.
func basicValue(v interface{}) interface{} {
	v = Indirect(v)
	switch v.(type) {
	case nil, string, bool:
		return v
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Bool:
		return rv.Bool()
	case reflect.String:
		return rv.String()
	}
	return v
}

// StringKeyMap converts a map whose keys are all strings to
// map[string]interface{}: a map whose key kind is string (map[string]int,
// map[Status]T), or a map with interface keys whose every key holds a string
// (what YAML decoders produce). ok is false for any other value.
func StringKeyMap(v interface{}) (map[string]interface{}, bool) {
	if m, ok := v.(map[string]interface{}); ok {
		return m, true
	}
	rv := reflect.ValueOf(Indirect(v))
	if rv.Kind() != reflect.Map {
		return nil, false
	}
	kk := rv.Type().Key().Kind()
	if kk != reflect.String && kk != reflect.Interface {
		return nil, false
	}
	result := make(map[string]interface{}, rv.Len())
	iter := rv.MapRange()
	for iter.Next() {
		k := iter.Key()
		if kk == reflect.Interface {
			k = k.Elem()
			if k.Kind() != reflect.String {
				return nil, false
			}
		}
		result[k.String()] = iter.Value().Interface()
	}
	return result, true
}

// stringMapKey spells key for a string-kind map key: DictKey of the dereferenced value; a []byte key is its text.
func stringMapKey(key interface{}) string {
	key = Indirect(key)
	if _, ok := key.([]byte); ok {
		return ToString(key)
	}
	return DictKey(key)
}

// mapKey converts a template value to a key of type kt: any value for
// string-kind keys (spelled by DictKey, so none is "None" as for dicts; a
// []byte key is its text); a number for integer, unsigned and float keys (a
// float must be integral and in range for integer keys; unsigned keys reject
// negative numbers); true, false, or a number that is 0 or 1 for bool keys; a
// list or array of the same length whose items convert to the element type
// and equal their conversion for array keys; for interface keys, the value
// itself if it is comparable (nil is the zero key); for any other kind, a
// comparable value of exactly type kt. ok is false when no key of type kt can
// be built.
func mapKey(key interface{}, kt reflect.Type) (reflect.Value, bool) {
	switch kt.Kind() {
	case reflect.String:
		return reflect.ValueOf(stringMapKey(key)).Convert(kt), true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, u, f, kind, ok := numericValue(Indirect(key))
		if !ok {
			return reflect.Value{}, false
		}
		switch kind {
		case 'u':
			if u > math.MaxInt64 {
				return reflect.Value{}, false
			}
			i = int64(u)
		case 'f':
			if f != math.Trunc(f) || f < -(1<<63) || f >= 1<<63 {
				return reflect.Value{}, false
			}
			i = int64(f)
		}
		out := reflect.New(kt).Elem()
		if out.OverflowInt(i) {
			return reflect.Value{}, false
		}
		out.SetInt(i)
		return out, true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		i, u, f, kind, ok := numericValue(Indirect(key))
		if !ok {
			return reflect.Value{}, false
		}
		switch kind {
		case 'i':
			if i < 0 {
				return reflect.Value{}, false
			}
			u = uint64(i)
		case 'f':
			if f != math.Trunc(f) || f < 0 || f >= 1<<64 {
				return reflect.Value{}, false
			}
			u = uint64(f)
		}
		out := reflect.New(kt).Elem()
		if out.OverflowUint(u) {
			return reflect.Value{}, false
		}
		out.SetUint(u)
		return out, true
	case reflect.Float32, reflect.Float64:
		i, u, f, kind, ok := numericValue(Indirect(key))
		if !ok {
			return reflect.Value{}, false
		}
		switch kind {
		case 'i':
			f = float64(i)
		case 'u':
			f = float64(u)
		}
		out := reflect.New(kt).Elem()
		if out.OverflowFloat(f) {
			return reflect.Value{}, false
		}
		out.SetFloat(f)
		return out, true
	case reflect.Bool:
		var b bool
		if bv, ok := basicValue(Unmark(key)).(bool); ok {
			b = bv
		} else {
			i, u, f, kind, ok := numericValue(Indirect(key))
			if !ok {
				return reflect.Value{}, false
			}
			var n float64
			switch kind {
			case 'i':
				n = float64(i)
			case 'u':
				n = float64(u)
			case 'f':
				n = f
			}
			if n != 0 && n != 1 {
				return reflect.Value{}, false
			}
			b = n == 1
		}
		return reflect.ValueOf(b).Convert(kt), true
	case reflect.Array:
		lv := reflect.ValueOf(Indirect(key))
		if !lv.IsValid() || (lv.Kind() != reflect.Slice && lv.Kind() != reflect.Array) || lv.Len() != kt.Len() {
			return reflect.Value{}, false
		}
		out := reflect.New(kt).Elem()
		for j := 0; j < kt.Len(); j++ {
			elem := lv.Index(j).Interface()
			ev, ok := mapKey(elem, kt.Elem())
			if !ok || !Equal(elem, ev.Interface()) {
				return reflect.Value{}, false
			}
			out.Index(j).Set(ev)
		}
		return out, true
	case reflect.Interface:
		kv := reflect.ValueOf(key)
		if !kv.IsValid() {
			return reflect.Zero(kt), true
		}
		if !kv.Comparable() || !kv.Type().AssignableTo(kt) {
			return reflect.Value{}, false
		}
		return kv, true
	}
	kv := reflect.ValueOf(key)
	if !kv.IsValid() || kv.Type() != kt || !kv.Comparable() {
		return reflect.Value{}, false
	}
	return kv, true
}

// mapLookup finds key in any Go map with a bounded number of MapIndex
// calls: interfaceMapLookup for interface keys, mapKey for the rest.
func mapLookup(rv reflect.Value, key interface{}) (reflect.Value, bool) {
	kt := rv.Type().Key()
	if kt.Kind() == reflect.Interface {
		return interfaceMapLookup(rv, key)
	}
	k, ok := mapKey(key, kt)
	if !ok {
		return reflect.Value{}, false
	}
	if val := rv.MapIndex(k); val.IsValid() {
		return val, true
	}
	return reflect.Value{}, false
}

// MapLookup is mapLookup for callers outside runtime: ok is false when m is
// not a map or has no entry for key.
func MapLookup(m, key interface{}) (interface{}, bool) {
	rv := reflect.ValueOf(Indirect(m))
	if rv.Kind() != reflect.Map {
		return nil, false
	}
	val, ok := mapLookup(rv, key)
	if !ok {
		return nil, false
	}
	return val.Interface(), true
}

func mapKeyRank(k interface{}) int {
	switch basicValue(k).(type) {
	case bool:
		return 0
	case string:
		return 2
	}
	if _, _, _, _, ok := numericValue(k); ok {
		return 1
	}
	return 3
}

func isNaNValue(k interface{}) bool {
	_, _, f, kind, ok := numericValue(k)
	return ok && kind == 'f' && f != f
}

// mapKeyLess is the strict order sortedMapKeys uses.
func mapKeyLess(a, b interface{}) bool {
	ra, rb := mapKeyRank(a), mapKeyRank(b)
	if ra != rb {
		return ra < rb
	}
	switch ra {
	case 0:
		ab, bb := basicValue(a).(bool), basicValue(b).(bool)
		if ab != bb {
			return !ab
		}
	case 1:
		an, bn := isNaNValue(a), isNaNValue(b)
		if an != bn {
			return bn
		}
		if !an {
			if cmp, ok := compareNumbers(a, b); ok && cmp != 0 {
				return cmp < 0
			}
		}
	case 2:
		as, bs := basicValue(a).(string), basicValue(b).(string)
		if as != bs {
			return as < bs
		}
	default:
		as, bs := ToString(a), ToString(b)
		if as != bs {
			return as < bs
		}
	}
	if ta, tb := fmt.Sprintf("%T", a), fmt.Sprintf("%T", b); ta != tb {
		return ta < tb
	}
	return ToString(a) < ToString(b)
}

// sortedMapKeys returns the keys of a map with non-string keys in a total
// order, so the same data renders the same way every time:
//  1. type rank: bool < number < string < other;
//  2. value: bools false < true; numbers by compareNumbers, with NaN after
//     every other number; strings by code point; others by ToString;
//  3. on a tie (1 and 1.0, or two NaN keys), by fmt.Sprintf("%T", k);
//  4. then by ToString(k).
func sortedMapKeys(rv reflect.Value) []interface{} {
	keys := make([]interface{}, 0, rv.Len())
	iter := rv.MapRange()
	for iter.Next() {
		keys = append(keys, iter.Key().Interface())
	}
	sort.SliceStable(keys, func(i, j int) bool { return mapKeyLess(keys[i], keys[j]) })
	return keys
}

// sortedMapPairs returns [key, value] pairs of a map in sortedMapKeys order.
// Values are read while ranging, since a NaN key cannot be looked up.
func sortedMapPairs(rv reflect.Value) []interface{} {
	pairs := make([]interface{}, 0, rv.Len())
	iter := rv.MapRange()
	for iter.Next() {
		pairs = append(pairs, []interface{}{iter.Key().Interface(), iter.Value().Interface()})
	}
	sort.SliceStable(pairs, func(i, j int) bool {
		return mapKeyLess(pairs[i].([]interface{})[0], pairs[j].([]interface{})[0])
	})
	return pairs
}

// isDictMethodName reports whether attr names a dict method that the
// map[string]interface{} case of getAttribute provides (get, items, keys, values).
func isDictMethodName(attr string) bool {
	return attr == "get" || attr == "items" || attr == "keys" || attr == "values"
}

// UsesReflectLookup reports whether attribute access on obj goes through
// lookupReflect: true unless obj is one of the runtime's own types handled
// by the fast cases of attributeExists and getAttribute.
func UsesReflectLookup(obj interface{}) bool {
	switch obj.(type) {
	case nil, NamespaceInterface, *Cycler, *Joiner, *CallableLoop,
		map[string]interface{}, map[string]string, GroupTuple:
		return false
	}
	return true
}

// Number returns v as int64, uint64 (only above MaxInt64) or float64 when it
// is a number of any Go kind, a pointer to one, a named numeric type, a bool
// (0 or 1, as in Python) or a RoundedFloat. ok is false otherwise.
func Number(v interface{}) (n interface{}, ok bool) {
	i, u, f, kind, ok := numericValue(v)
	if !ok {
		return nil, false
	}
	switch kind {
	case 'i':
		return i, true
	case 'u':
		if u <= math.MaxInt64 {
			return int64(u), true
		}
		return u, true
	}
	return f, true
}

// BasicValue converts a named bool or string type, or a pointer to one, to a
// plain bool or string. Other values are returned unchanged.
func BasicValue(v interface{}) interface{} { return basicValue(v) }

// SortedMapKeys returns the keys of any Go map in the order a {% for %} loop
// visits them. ok is false when v is not a map.
func SortedMapKeys(v interface{}) ([]interface{}, bool) {
	if m, ok := StringKeyMap(v); ok {
		keys := sortedKeys(m)
		out := make([]interface{}, len(keys))
		for i, k := range keys {
			out[i] = k
		}
		return out, true
	}
	rv := reflect.ValueOf(Indirect(v))
	if rv.Kind() != reflect.Map {
		return nil, false
	}
	return sortedMapKeys(rv), true
}

// SortedMapPairs returns the [key, value] pairs of any Go map in the order a
// {% for %} loop visits the keys. ok is false when v is not a map.
func SortedMapPairs(v interface{}) ([]interface{}, bool) {
	if m, ok := StringKeyMap(v); ok {
		keys := sortedKeys(m)
		out := make([]interface{}, len(keys))
		for i, k := range keys {
			out[i] = []interface{}{k, m[k]}
		}
		return out, true
	}
	rv := reflect.ValueOf(Indirect(v))
	if rv.Kind() != reflect.Map {
		return nil, false
	}
	return sortedMapPairs(rv), true
}

// interfaceMapLookup finds key in a map with interface keys the way a template
// compares keys, with a bounded number of MapIndex calls and no scan: the key
// itself, a string as Markup, the same number in another numeric type (a bool
// is 0 or 1), or its DictKey spelling.
func interfaceMapLookup(rv reflect.Value, key interface{}) (reflect.Value, bool) {
	kt := rv.Type().Key()
	try := func(c interface{}) (reflect.Value, bool) {
		cv := reflect.ValueOf(c)
		if !cv.IsValid() || !cv.Type().AssignableTo(kt) || !cv.Comparable() {
			return reflect.Value{}, false
		}
		if val := rv.MapIndex(cv); val.IsValid() {
			return val, true
		}
		return reflect.Value{}, false
	}
	key = Indirect(key)
	if key == nil {
		if val := rv.MapIndex(reflect.Zero(kt)); val.IsValid() {
			return val, true
		}
	} else if val, ok := try(key); ok {
		return val, true
	}
	var cands []interface{}
	if sv := reflect.ValueOf(Unmark(key)); sv.IsValid() && sv.Kind() == reflect.String {
		cands = append(cands, Markup(sv.String()))
	}
	if b, ok := key.(bool); ok {
		n := 0
		if b {
			n = 1
		}
		cands = append(cands, n, int64(n), float64(n))
	} else if i, u, f, kind, ok := numericValue(key); ok {
		switch kind {
		case 'i':
			cands = append(cands, int(i), i, float64(i))
			if i >= 0 {
				cands = append(cands, uint64(i))
			}
			if i == 0 || i == 1 {
				cands = append(cands, i == 1)
			}
		case 'u':
			cands = append(cands, u, float64(u))
			if u <= math.MaxInt64 {
				cands = append(cands, int(u), int64(u))
			}
			if u <= 1 {
				cands = append(cands, u == 1)
			}
		case 'f':
			if f == math.Trunc(f) && f >= -(1<<63) && f < 1<<63 {
				cands = append(cands, int(f), int64(f))
				if f >= 0 {
					cands = append(cands, uint64(f))
				}
				if f == 0 || f == 1 {
					cands = append(cands, f == 1)
				}
			} else if f == math.Trunc(f) && f >= 0 && f < 1<<64 {
				cands = append(cands, uint64(f))
			}
			cands = append(cands, f)
		}
	}
	for _, c := range cands {
		if val, ok := try(c); ok {
			return val, true
		}
	}
	return try(DictKey(key))
}

// dictMethodOver builds the dict method attr (get, items, keys, values) of a
// Go map, in the order of sortedMapPairs. get looks the key up with mapLookup,
// without scanning or copying the map.
func dictMethodOver(rv reflect.Value, attr string) interface{} {
	if attr == "get" {
		return &builtinMethod{name: "get", fn: func(args []interface{}, kwargs map[string]interface{}) (interface{}, error) {
			if len(kwargs) > 0 {
				return nil, fmt.Errorf("get() takes no keyword arguments")
			}
			if len(args) < 1 || len(args) > 2 {
				return nil, fmt.Errorf("get() takes 1 or 2 arguments (%d given)", len(args))
			}
			if val, ok := mapLookup(rv, args[0]); ok {
				return val.Interface(), nil
			}
			if len(args) == 2 {
				return args[1], nil
			}
			return nil, nil
		}}
	}
	return func(args ...interface{}) (interface{}, error) {
		pairs := sortedMapPairs(rv)
		if attr == "items" {
			return pairs, nil
		}
		pick := 0
		if attr == "values" {
			pick = 1
		}
		out := make([]interface{}, len(pairs))
		for i, p := range pairs {
			out[i] = p.([]interface{})[pick]
		}
		return out, nil
	}
}
