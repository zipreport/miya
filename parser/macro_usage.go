package parser

import (
	"fmt"
	"reflect"
)

// findMacroUsage reports whether body references varargs, kwargs or caller
// anywhere, including nested blocks, macros and call blocks. A declared
// parameter with one of those names is an ordinary parameter, as in Jinja2.
func findMacroUsage(params []string, defaults map[string]ExpressionNode, body []Node) (MacroUsage, error) {
	var u MacroUsage
	for _, n := range body {
		walkIdentifiers(reflect.ValueOf(n), func(name string) {
			switch name {
			case "varargs":
				u.CatchVarargs = true
			case "kwargs":
				u.CatchKwargs = true
			case "caller":
				u.Caller = true
			}
		})
	}
	// A declared parameter with one of these names is an ordinary parameter,
	// as in Jinja2: it switches the special handling off for that name.
	for _, name := range params {
		switch name {
		case "varargs":
			u.CatchVarargs = false
		case "kwargs":
			u.CatchKwargs = false
		case "caller":
			if _, ok := defaults[name]; !ok {
				return u, fmt.Errorf(`when defining macros or call blocks the special "caller" argument must be omitted or be given a default`)
			}
			u.Caller = false
		}
	}
	return u, nil
}

// walkIdentifiers calls fn for every IdentifierNode reachable from v through
// exported fields, slices and maps.
func walkIdentifiers(v reflect.Value, fn func(string)) {
	switch v.Kind() {
	case reflect.Interface, reflect.Ptr:
		if v.IsNil() {
			return
		}
		if id, ok := v.Interface().(*IdentifierNode); ok {
			fn(id.Name)
			return
		}
		walkIdentifiers(v.Elem(), fn)
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if v.Type().Field(i).IsExported() {
				walkIdentifiers(v.Field(i), fn)
			}
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			walkIdentifiers(v.Index(i), fn)
		}
	case reflect.Map:
		iter := v.MapRange()
		for iter.Next() {
			walkIdentifiers(iter.Value(), fn)
		}
	}
}
