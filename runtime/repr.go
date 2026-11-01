package runtime

import (
	"fmt"
	"math"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"unsafe"
)

// maxReprBytes bounds the printed form of one container (same bound as the
// filters' maxRepeatBytes, plan 030); output past it ends in "...".
const maxReprBytes = 64 << 20

// reprKey identifies a slice (data pointer and length) or map (pointer, n 0)
// on the current repr path.
type reprKey struct {
	ptr uintptr
	n   int
}

// reprState carries the containers on the current path (to print [...] for
// self-reference, like Python) and the bytes written so far.
type reprState struct {
	stack []reprKey        // containers on the current path
	set   map[reprKey]bool // the same, once the path is deep
	b     []byte
	full  bool
}

// reprContainer prints a slice, array or map the way Python's repr does
// (Decisions 1-3 and 2a-2b in plan 048); ok is false for any other value.
// ToString checks error and fmt.Stringer before calling it.
func reprContainer(v interface{}) (string, bool) {
	v = Indirect(v)
	if v == nil {
		return "", false
	}
	rv := reflect.ValueOf(v)
	s := &reprState{}
	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		if items, ok := v.([]interface{}); ok {
			s.reprItems(items)
		} else {
			s.reprSeq(rv)
		}
	case reflect.Map:
		s.reprMap(rv, v)
	default:
		return "", false
	}
	return string(s.b), true
}

// reprOf is Python's repr of a template value: strings are quoted, containers
// print their elements with repr.
func reprOf(v interface{}) string {
	s := &reprState{}
	s.repr(v)
	return string(s.b)
}

// write appends str unless it would pass the size bound, in which case it
// ends the output with "...".
func (s *reprState) write(str string) {
	if s.full {
		return
	}
	if len(s.b)+len(str) > maxReprBytes {
		s.b = append(s.b, "..."...)
		s.full = true
		return
	}
	s.b = append(s.b, str...)
}

func reprQuote(str string) string {
	q := byte('\'')
	if strings.Contains(str, "'") && !strings.Contains(str, `"`) {
		q = '"'
	}
	var b strings.Builder
	b.Grow(len(str) + 2)
	b.WriteByte(q)
	for _, r := range str {
		switch {
		case r == '\\':
			b.WriteString(`\\`)
		case r == rune(q):
			b.WriteByte('\\')
			b.WriteByte(q)
		case r == '\n':
			b.WriteString(`\n`)
		case r == '\r':
			b.WriteString(`\r`)
		case r == '\t':
			b.WriteString(`\t`)
		default:
			b.WriteRune(r)
		}
	}
	b.WriteByte(q)
	return b.String()
}

// repr writes one element inside a container.
func (s *reprState) repr(v interface{}) {
	if s.full {
		return
	}
	switch x := v.(type) {
	case int:
		if len(s.b) > maxReprBytes-24 {
			s.write(strconv.Itoa(x))
			return
		}
		s.b = strconv.AppendInt(s.b, int64(x), 10)
		return
	case string:
		s.write(reprQuote(x))
		return
	case []interface{}:
		s.reprItems(x)
		return
	}
	v = unbox(v)
	switch sv := v.(type) {
	case SafeValue:
		s.repr(sv.Value)
		return
	case *SafeValue:
		if sv != nil {
			s.repr(sv.Value)
			return
		}
	case Markup:
		s.write(reprQuote(string(sv)))
		return
	}
	v = Indirect(v)
	if v == nil {
		s.write("None")
		return
	}
	if b, ok := v.([]byte); ok {
		s.write(reprQuote(string(b)))
		return
	}
	if ns, ok := v.(ReprNamespace); ok {
		s.reprNamespace(ns, v)
		return
	}
	if _, ok := v.(*Undefined); ok {
		s.write("Undefined")
		return
	}
	if e, ok := v.(error); ok {
		s.write(e.Error())
		return
	}
	if st, ok := v.(fmt.Stringer); ok {
		s.write(st.String())
		return
	}
	switch bv := basicValue(v).(type) {
	case bool:
		if bv {
			s.write("True")
		} else {
			s.write("False")
		}
		return
	case string:
		s.write(reprQuote(bv))
		return
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		s.write(strconv.FormatInt(rv.Int(), 10))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		s.write(strconv.FormatUint(rv.Uint(), 10))
	case reflect.Float32, reflect.Float64:
		f := rv.Float()
		switch {
		case math.IsInf(f, 1):
			s.write("inf")
		case math.IsInf(f, -1):
			s.write("-inf")
		case math.IsNaN(f):
			s.write("nan")
		default:
			bits := 64
			if rv.Kind() == reflect.Float32 {
				bits = 32
			}
			s.write(FormatFloat(f, bits))
		}
	case reflect.Slice, reflect.Array:
		s.reprSeq(rv)
	case reflect.Map:
		s.reprMap(rv, v)
	default:
		s.write(ToString(v))
	}
}

// maxLinearPath is the path depth up to which a self-reference is found by
// scanning the path instead of a map lookup.
const maxLinearPath = 16

func (s *reprState) enter(key reprKey) (tracked, cyclic bool) {
	if key.ptr == 0 {
		return false, false
	}
	if s.set != nil {
		if s.set[key] {
			return false, true
		}
	} else {
		for _, k := range s.stack {
			if k == key {
				return false, true
			}
		}
	}
	s.stack = append(s.stack, key)
	if s.set != nil {
		s.set[key] = true
	} else if len(s.stack) > maxLinearPath {
		s.set = make(map[reprKey]bool, 2*maxLinearPath)
		for _, k := range s.stack {
			s.set[k] = true
		}
	}
	return true, false
}

func (s *reprState) leave(key reprKey) {
	s.stack = s.stack[:len(s.stack)-1]
	if s.set != nil {
		delete(s.set, key)
	}
}

// reprItems prints a []interface{}. The list joins the path only when it
// holds something that could contain it again, so lists of plain values (the
// bulk of any large structure) cost no path bookkeeping.
func (s *reprState) reprItems(items []interface{}) {
	start := len(s.b)
	if len(s.b) < maxReprBytes-1 {
		s.b = append(s.b, '[')
	} else {
		s.write("[")
	}
	var key reprKey
	tracked := false
	for i, it := range items {
		if s.full {
			break
		}
		if !tracked {
			switch it.(type) {
			case nil, int, string, bool, float64:
			default:
				if len(items) > 0 {
					key = reprKey{uintptr(unsafe.Pointer(unsafe.SliceData(items))), len(items)}
					var cyclic bool
					tracked, cyclic = s.enter(key)
					if cyclic {
						s.b = s.b[:start]
						s.write("[...]")
						return
					}
				}
			}
		}
		if i > 0 {
			if len(s.b) > maxReprBytes-2 {
				s.write(", ")
				break
			}
			s.b = append(s.b, ',', ' ')
		}
		if n, ok := it.(int); ok && len(s.b) <= maxReprBytes-24 {
			s.b = strconv.AppendInt(s.b, int64(n), 10)
			continue
		}
		s.repr(it)
	}
	s.write("]")
	if tracked {
		s.leave(key)
	}
}

func (s *reprState) reprSeq(rv reflect.Value) {
	var key reprKey
	if rv.Kind() == reflect.Slice && rv.Len() > 0 {
		key = reprKey{rv.Pointer(), rv.Len()}
	}
	tracked, cyclic := s.enter(key)
	if cyclic {
		s.write("[...]")
		return
	}
	s.write("[")
	for i := 0; i < rv.Len() && !s.full; i++ {
		if i > 0 {
			s.write(", ")
		}
		s.repr(rv.Index(i).Interface())
	}
	s.write("]")
	if tracked {
		s.leave(key)
	}
}

func (s *reprState) reprMap(rv reflect.Value, v interface{}) {
	var key reprKey
	if rv.Len() > 0 {
		key = reprKey{ptr: rv.Pointer()}
	}
	tracked, cyclic := s.enter(key)
	if cyclic {
		s.write("{...}")
		return
	}
	s.write("{")
	if m, ok := StringKeyMap(v); ok {
		for i, k := range sortedKeys(m) {
			if s.full {
				break
			}
			if i > 0 {
				s.write(", ")
			}
			s.write(reprQuote(k))
			s.write(": ")
			s.repr(m[k])
		}
	} else {
		for i, p := range sortedMapPairs(rv) {
			if s.full {
				break
			}
			if i > 0 {
				s.write(", ")
			}
			kv := p.([]interface{})
			s.repr(kv[0])
			s.write(": ")
			s.repr(kv[1])
		}
	}
	s.write("}")
	if tracked {
		s.leave(key)
	}
}

// ReprNamespace is implemented by namespace() objects so that repr prints
// their data inside the cycle tracking and size budget.
type ReprNamespace interface {
	ReprData() map[string]interface{}
}

// Repr is Python's repr of a template value, with the cycle tracking and size
// bound of printed containers.
func Repr(v interface{}) string { return reprOf(v) }

func (s *reprState) reprNamespace(ns ReprNamespace, v interface{}) {
	var key reprKey
	if rv := reflect.ValueOf(v); rv.Kind() == reflect.Ptr {
		key = reprKey{ptr: rv.Pointer()}
	}
	tracked, cyclic := s.enter(key)
	if cyclic {
		s.write("namespace(...)")
		return
	}
	s.write("namespace(")
	s.repr(ns.ReprData())
	s.write(")")
	if tracked {
		s.leave(key)
	}
}

// FmtCycles reports whether fmt's %v would recurse forever on v: a non-empty
// slice or map that holds itself through slices, arrays, maps, structs or
// interfaces. fmt prints a nested pointer as an address, so only a top-level
// pointer is followed.
func FmtCycles(v interface{}) bool {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr && !rv.IsNil() {
		rv = rv.Elem()
	}
	var w cycleWalk
	return w.cyclic(rv)
}

// cycleWalk holds the containers on the current path and those already
// walked without finding a cycle.
type cycleWalk struct{ path, done map[reprKey]bool }

func (w *cycleWalk) cyclic(rv reflect.Value) bool {
	switch rv.Kind() {
	case reflect.Interface:
		return !rv.IsNil() && w.cyclic(rv.Elem())
	case reflect.Struct:
		for i := 0; i < rv.NumField(); i++ {
			if w.cyclic(rv.Field(i)) {
				return true
			}
		}
	case reflect.Array:
		if canHoldContainer(rv.Type().Elem()) {
			for i := 0; i < rv.Len(); i++ {
				if w.cyclic(rv.Index(i)) {
					return true
				}
			}
		}
	case reflect.Slice, reflect.Map:
		t := rv.Type()
		if rv.Len() == 0 || !canHoldContainer(t.Elem()) && (t.Kind() == reflect.Slice || !canHoldContainer(t.Key())) {
			return false
		}
		key := reprKey{ptr: rv.Pointer()}
		if t.Kind() == reflect.Slice {
			key.n = rv.Len()
		}
		if w.path[key] {
			return true
		}
		if w.done[key] {
			return false
		}
		if w.path == nil {
			w.path, w.done = map[reprKey]bool{}, map[reprKey]bool{}
		}
		w.path[key] = true
		found := false
		if t.Kind() == reflect.Slice {
			for i := 0; i < rv.Len() && !found; i++ {
				found = w.cyclic(rv.Index(i))
			}
		} else {
			for it := rv.MapRange(); !found && it.Next(); {
				found = w.cyclic(it.Key()) || w.cyclic(it.Value())
			}
		}
		delete(w.path, key)
		w.done[key] = true
		return found
	}
	return false
}

// canHoldContainer reports whether a value of type t can hold a slice or map
// that fmt would print.
func canHoldContainer(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Interface, reflect.Struct, reflect.Array, reflect.Slice, reflect.Map:
		return true
	}
	return false
}

// SprintV is fmt.Sprintf("%v", v), except that a value fmt would print
// forever (FmtCycles) prints a slice or map met again inside itself as [...]
// or map[...].
func SprintV(v interface{}) string {
	if !FmtCycles(v) {
		return fmt.Sprintf("%v", v)
	}
	s := &reprState{}
	s.fmtValue(reflect.ValueOf(v), 0)
	return string(s.b)
}

func (s *reprState) fmtValue(rv reflect.Value, depth int) {
	if s.full {
		return
	}
	if !rv.IsValid() {
		s.write("<nil>")
		return
	}
	if depth > 0 && rv.CanInterface() && !(rv.Kind() == reflect.Ptr && rv.IsNil()) {
		switch x := rv.Interface().(type) {
		case error:
			s.write(x.Error())
			return
		case fmt.Stringer:
			s.write(x.String())
			return
		}
	}
	switch rv.Kind() {
	case reflect.Interface:
		s.fmtValue(rv.Elem(), depth+1)
	case reflect.Ptr:
		if rv.IsNil() {
			s.write("<nil>")
			return
		}
		if depth == 0 {
			switch rv.Elem().Kind() {
			case reflect.Array, reflect.Slice, reflect.Struct, reflect.Map:
				s.write("&")
				s.fmtValue(rv.Elem(), depth+1)
				return
			}
		}
		s.write("0x" + strconv.FormatUint(uint64(rv.Pointer()), 16))
	case reflect.Struct:
		s.write("{")
		for i := 0; i < rv.NumField(); i++ {
			if i > 0 {
				s.write(" ")
			}
			s.fmtValue(rv.Field(i), depth+1)
		}
		s.write("}")
	case reflect.Map:
		key := reprKey{ptr: rv.Pointer()}
		tracked, cyclic := s.enter(key)
		if cyclic {
			s.write("map[...]")
			return
		}
		type entry struct {
			k string
			v reflect.Value
		}
		entries := make([]entry, 0, rv.Len())
		for it := rv.MapRange(); it.Next(); {
			ks := &reprState{}
			ks.fmtValue(it.Key(), depth+1)
			entries = append(entries, entry{string(ks.b), it.Value()})
		}
		sort.Slice(entries, func(i, j int) bool { return entries[i].k < entries[j].k })
		s.write("map[")
		for i, en := range entries {
			if i > 0 {
				s.write(" ")
			}
			s.write(en.k + ":")
			s.fmtValue(en.v, depth+1)
		}
		s.write("]")
		if tracked {
			s.leave(key)
		}
	case reflect.Slice, reflect.Array:
		var key reprKey
		if rv.Kind() == reflect.Slice && rv.Len() > 0 {
			key = reprKey{rv.Pointer(), rv.Len()}
		}
		tracked, cyclic := s.enter(key)
		if cyclic {
			s.write("[...]")
			return
		}
		s.write("[")
		for i := 0; i < rv.Len() && !s.full; i++ {
			if i > 0 {
				s.write(" ")
			}
			s.fmtValue(rv.Index(i), depth+1)
		}
		s.write("]")
		if tracked {
			s.leave(key)
		}
	default:
		s.write(fmt.Sprintf("%v", rv))
	}
}
