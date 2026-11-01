package miya_test

import (
	"errors"
	"strings"
	"testing"

	miya "github.com/zipreport/miya"
	"github.com/zipreport/miya/loader"
	"github.com/zipreport/miya/runtime"
)

// Issue #12: importing a template whose {% extends %} parent is missing or
// broken fails, as in Jinja2.
func TestIssue12ImportMissingParent(t *testing.T) {
	sl := loader.NewStringLoader(loader.NewDirectTemplateParser())
	for n, s := range map[string]string{
		"ext.html":    `{% extends "nope.html" %}{% macro f() %}x{% endmacro %}`,
		"bad.html":    `{% extends "broken.html" %}{% macro f() %}x{% endmacro %}`,
		"broken.html": `{% if %}`,
		"ok.html":     `{% extends "okbase.html" %}{% macro f() %}x{% endmacro %}`,
		"okbase.html": `{% macro g() %}g{% endmacro %}`,
		"deep.html":   `{% extends "mid.html" %}{% macro f() %}x{% endmacro %}`,
		"mid.html":    `{% extends "nope.html" %}`,
		"self.html":   `{% extends "self.html" %}{% macro f() %}x{% endmacro %}`,
		"dyn.html":    `{% extends p %}{% macro f() %}x{% endmacro %}`,
	} {
		sl.AddTemplate(n, s)
	}
	env := miya.NewEnvironment(miya.WithLoader(sl))

	// Jinja2 3.1.6 raises TemplateNotFound, TemplateSyntaxError or RecursionError
	errs := []struct {
		name, src, names, contains string
		wantIs                     error
	}{
		{"claim1_import_missing_parent", `{% import "ext.html" as m %}[{{ m.f() }}]`, "nope.html", "", runtime.ErrTemplateNotFound},
		{"claim2_from_missing_parent", `{% from "ext.html" import f %}[{{ f() }}]`, "nope.html", "", runtime.ErrTemplateNotFound},
		{"claim3_import_broken_parent", `{% import "bad.html" as m %}[{{ m.f() }}]`, "broken.html", "", nil},
		{"missing_grandparent", `{% import "deep.html" as m %}[{{ m.f() }}]`, "nope.html", "", runtime.ErrTemplateNotFound},
		{"extends_itself", `{% import "self.html" as m %}[{{ m.f() }}]`, "self.html", "circular inheritance detected", nil},
	}
	for _, tt := range errs {
		t.Run(tt.name, func(t *testing.T) {
			got, err := env.RenderString(tt.src, miya.NewContext())
			if err == nil {
				t.Fatalf("want an error, got %q", got)
			}
			if tt.wantIs != nil && !errors.Is(err, tt.wantIs) {
				t.Errorf("error does not wrap %v: %v", tt.wantIs, err)
			}
			if tt.wantIs == nil && errors.Is(err, runtime.ErrTemplateNotFound) {
				t.Errorf("error is reported as not found: %v", err)
			}
			var nf *runtime.TemplateNotFoundError
			if errors.As(err, &nf) && nf.Name != tt.names {
				t.Errorf("not-found error names %q, want %q", nf.Name, tt.names)
			}
			if !strings.Contains(err.Error(), tt.contains) {
				t.Errorf("error does not say %q: %v", tt.contains, err)
			}
			if n := strings.Count(err.Error(), `"`+tt.names+`"`); n != 1 {
				t.Errorf("%q appears %d times, want 1: %v", tt.names, n, err)
			}
		})
	}

	renders := []struct{ name, src, want string }{
		{"claim4_existing_parent_exports_child_macros", `{% import "ok.html" as m %}[{{ m.f() }}]`, "[x]"},
		{"from_existing_parent", `{% from "ok.html" import f %}[{{ f() }}]`, "[x]"},
		// a parent named by an expression is not followed
		{"dynamic_parent_not_followed", `{% import "dyn.html" as m %}[{{ m.f() }}]`, "[x]"},
	}
	for _, tt := range renders {
		t.Run(tt.name, func(t *testing.T) {
			got, err := env.RenderString(tt.src, miya.NewContext())
			if err != nil {
				t.Fatalf("render: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
