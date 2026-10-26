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
	}{
		{"join", `{% set s = 'x' * (1024*1024) %}{{ ''.join([s]*100) }}`, "join(): result too large"},
		{"replace", `{% set s = 'x' * (1024*1024) %}{{ ('a '*100).replace(' ', s) }}`, "replace(): result too large"},
		{"repeat", `{{ 'x' * (100*1024*1024) }}`, "repetition result too large"},
		{"list_repeat", `{{ [0] * 100000000 }}`, "list repetition result too large"},
		{"center", `{{ 'x'|center(100000000) }}`, "filter output too large"},
		{"filter_block", `{% filter center(100000000) %}x{% endfilter %}`, "in filter block"},
		{"dynamic_extends", `{% extends ('x' * (100*1024*1024)) %}`, "dynamic template name"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sl := loader.NewStringLoader(loader.NewDirectTemplateParser())
			sl.AddTemplate("t", tc.tpl)
			env := miya.NewEnvironment(miya.WithLoader(sl))
			tmpl, err := env.GetTemplate("t")
			if err == nil {
				_, err = tmpl.Render(miya.NewContext())
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
