package miya_test

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	miya "github.com/zipreport/miya"
	"github.com/zipreport/miya/loader"
	"github.com/zipreport/miya/runtime"
)

// Issue #3: an import of a template that cannot be loaded names it once.
func TestIssue3ImportErrorNamesTemplateOnce(t *testing.T) {
	dir := t.TempDir()
	write := func(name, src string) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
		return p
	}
	write("bad.html", "{% if %}")
	locked := write("m.html", "{% macro f() %}x{% endmacro %}")
	if err := os.Chmod(locked, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(locked, 0o644) })
	readable := true
	if _, err := os.ReadFile(locked); err != nil {
		readable = false
	}
	env := miya.NewEnvironment(miya.WithLoader(loader.NewFileSystemLoader([]string{dir}, loader.NewDirectTemplateParser())))
	noLoader := miya.NewEnvironment()

	tests := []struct {
		name, src, tmpl string
		unreadable      bool
		wantIs          error
		env             *miya.Environment
	}{
		{"import_unreadable", `{% import "m.html" as m %}`, "m.html", true, fs.ErrPermission, nil},
		{"from_unreadable", `{% from "m.html" import f %}`, "m.html", true, fs.ErrPermission, nil},
		{"import_missing", `{% import "nope.html" as m %}`, "nope.html", false, runtime.ErrTemplateNotFound, nil},
		{"from_missing", `{% from "nope.html" import f %}`, "nope.html", false, runtime.ErrTemplateNotFound, nil},
		// guards: errors that do not already name the template keep the prefix
		{"import_compile_error_names_template", `{% import "bad.html" as m %}`, "bad.html", false, nil, nil},
		{"import_no_loader_names_template", `{% import "x.html" as m %}`, "x.html", false, nil, noLoader},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.unreadable && readable {
				t.Skip("a mode-000 file is still readable here (root, or Windows)")
			}
			e := env
			if tt.env != nil {
				e = tt.env
			}
			_, err := e.RenderString(tt.src, miya.NewContext())
			if err == nil {
				t.Fatal("want an error")
			}
			if tt.wantIs != nil && !errors.Is(err, tt.wantIs) {
				t.Errorf("error does not wrap %v: %v", tt.wantIs, err)
			}
			prefix := `failed to load template "` + tt.tmpl + `"`
			if n := strings.Count(err.Error(), prefix); n != 1 {
				t.Errorf("%q appears %d times, want 1: %v", prefix, n, err)
			}
		})
	}
}
