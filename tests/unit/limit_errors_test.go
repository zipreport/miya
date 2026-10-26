package miya_test

import (
	"errors"
	"strings"
	"testing"

	miya "github.com/zipreport/miya"
	"github.com/zipreport/miya/loader"
	"github.com/zipreport/miya/runtime"
)

func TestLimitErrors(t *testing.T) {
	tests := []struct {
		name, tpl, text string
		data            map[string]interface{}
	}{
		{"join", `{% set s = 'x' * (1024*1024) %}{{ ''.join([s]*100) }}`, "join(): result too large", nil},
		{"replace", `{% set s = 'x' * (1024*1024) %}{{ ('a '*100).replace(' ', s) }}`, "replace(): result too large", nil},
		{"repeat", `{{ 'x' * (100*1024*1024) }}`, "repetition result too large", nil},
		{"list_repeat", `{{ [0] * 100000000 }}`, "list repetition result too large", nil},
		{"center", `{{ 'x'|center(100000000) }}`, "filter output too large", nil},
		{"filter_block", `{% filter center(100000000) %}x{% endfilter %}`, "in filter block", nil},
		{"dynamic_extends", `{% extends ('x' * (100*1024*1024)) %}`, "dynamic template name", nil},
		{"range", `{{ range(2000000)|length }}`, "range() too large", nil},
		{"lipsum", `{{ lipsum(2000) }}`, "lipsum() too large", nil},
		{"list_concat", `{{ big + [] }}`, "list concatenation too large", map[string]interface{}{"big": make([]struct{}, 10000001)}},
		{"batch_fill", `{{ [1]|batch(2000000, 0)|list }}`, "batch size too large to fill", nil},
		{"slice_count", `{{ [1]|slice(2000000)|list }}`, "slice count too large", nil},
		{"format_method", `{% set s = 'x' * (1024*1024) %}{{ ('{0}' * 100).format(s) }}`, "format(): result too large", nil},
		{"extend_growth", `{% set l = [] %}{% do l.extend(big) %}`, "list too large", map[string]interface{}{"big": make([]int, 10000001)}},
		{"indent_lines", `{{ ('\n' * 1000)|indent(100000) }}`, "indent output too large", nil},
		{"format_width", `{{ '%2000000d'|format(1) }}`, "format: width too big", nil},
		{"format_star_width", `{{ '%*d'|format(2000000, 1) }}`, "format: width too big", nil},
		{"format_precision", `{{ '%.2000000f'|format(1) }}`, "format: precision too big", nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sl := loader.NewStringLoader(loader.NewDirectTemplateParser())
			sl.AddTemplate("t", tc.tpl)
			env := miya.NewEnvironment(miya.WithLoader(sl))
			tmpl, err := env.GetTemplate("t")
			if err == nil {
				_, err = tmpl.Render(ctxOf(tc.data))
			}
			if err == nil {
				t.Fatal("expected an error")
			}
			if !errors.Is(err, runtime.ErrOutputTooLarge) {
				t.Errorf("errors.Is(ErrOutputTooLarge) false for: %v", err)
			}
			if !strings.Contains(err.Error(), tc.text) {
				t.Errorf("error %q lacks %q", err.Error(), tc.text)
			}
		})
	}
}

func ctxOf(d map[string]interface{}) miya.Context {
	if d == nil {
		return miya.NewContext()
	}
	return miya.NewContextFrom(d)
}
