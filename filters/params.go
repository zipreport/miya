package filters

import (
	"fmt"
	"sort"

	"github.com/zipreport/miya/runtime"
)

// Param describes one parameter of a filter, in positional order. Keyword
// arguments are bound to positions by Name.
//
// When a keyword fills a later position, every earlier unfilled position
// receives its Param's Default. A Param without a Default therefore passes
// nil: a filter whose code type-asserts args[i] must handle nil, or declare a
// Default.
type Param struct {
	Name     string
	Default  interface{} // passed for an unfilled position before a bound one; nil if unset
	Required bool        // a keyword after it cannot leave it unfilled
	// Unsupported marks a Jinja2 parameter miya does not implement yet:
	// passing it by keyword is an error, never silently ignored.
	Unsupported bool
}

// KwArgs carries keyword arguments, as the last positional argument, to the
// filters listed in kwargsFilters (whose Jinja2 signature is *args/**kwargs).
type KwArgs = runtime.KwArgs

var builtinParams = map[string][]Param{
	"default": {{Name: "default_value", Default: ""}, {Name: "boolean", Default: false}},
	"d":       {{Name: "default_value", Default: ""}, {Name: "boolean", Default: false}},
	"join":    {{Name: "d", Default: ""}, {Name: "attribute"}},
	"sort":    {{Name: "reverse", Default: false}, {Name: "case_sensitive", Default: false}, {Name: "attribute"}},
	"unique":  {{Name: "case_sensitive", Default: false}, {Name: "attribute"}},
	"batch":   {{Name: "linecount", Required: true}, {Name: "fill_with"}},
	"slice":   {{Name: "slices", Required: true}, {Name: "fill_with"}},
	"truncate": {
		{Name: "length", Default: 255}, {Name: "killwords", Default: false},
		{Name: "end", Default: "..."}, {Name: "leeway"},
	},
	"center":         {{Name: "width", Default: 80}},
	"indent":         {{Name: "width", Default: 4}, {Name: "first", Default: false}, {Name: "blank", Default: false}},
	"round":          {{Name: "precision", Default: 0}, {Name: "method", Default: "common"}},
	"int":            {{Name: "default", Default: 0}, {Name: "base", Default: 10}},
	"float":          {{Name: "default", Default: 0.0}},
	"sum":            {{Name: "attribute"}, {Name: "start", Default: 0}},
	"min":            {{Name: "case_sensitive", Default: false}, {Name: "attribute"}},
	"max":            {{Name: "case_sensitive", Default: false}, {Name: "attribute"}},
	"replace":        {{Name: "old", Required: true}, {Name: "new", Required: true}, {Name: "count"}},
	"filesizeformat": {{Name: "binary", Default: false}},
	"tojson":         {{Name: "indent"}},
	"dictsort":       {{Name: "case_sensitive", Default: false}, {Name: "by", Default: "key"}, {Name: "reverse", Default: false}},
	"groupby": {
		{Name: "attribute", Required: true}, {Name: "default"},
		{Name: "case_sensitive", Default: false},
	},
	"trim":  {{Name: "chars"}},
	"strip": {{Name: "chars"}},
}

var kwargsFilters = map[string]bool{"map": true}

// bindKwargs places keyword arguments after the positional ones, by parameter
// name, filling skipped positions with each parameter's Default.
func bindKwargs(name string, params []Param, args []interface{}, kwargs map[string]interface{}) ([]interface{}, error) {
	keys := make([]string, 0, len(kwargs))
	for k := range kwargs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	bound := make(map[int]interface{}, len(keys))
	last := len(args) - 1
	for _, k := range keys {
		idx := -1
		for i, p := range params {
			if p.Name == k {
				idx = i
				break
			}
		}
		if idx < 0 {
			return nil, fmt.Errorf("filter %q got an unexpected keyword argument %q", name, k)
		}
		if params[idx].Unsupported {
			return nil, fmt.Errorf("filter %q: keyword argument %q is not supported", name, k)
		}
		if idx < len(args) {
			return nil, fmt.Errorf("filter %q got multiple values for argument %q", name, k)
		}
		bound[idx] = kwargs[k]
		if idx > last {
			last = idx
		}
	}

	result := make([]interface{}, last+1)
	copy(result, args)
	for i := len(args); i <= last; i++ {
		if v, ok := bound[i]; ok {
			result[i] = v
			continue
		}
		if params[i].Required {
			return nil, fmt.Errorf("filter %q missing required argument %q", name, params[i].Name)
		}
		result[i] = params[i].Default
	}
	return result, nil
}
