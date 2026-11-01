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

// Issue #11: an include of a template that cannot be loaded names it once.
func TestIssue11IncludeErrorNamesTemplateOnce(t *testing.T) {
	dir := t.TempDir()
	write := func(name, src string) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
		return p
	}
	write("bad.html", "{% if %}")
	locked := write("m.html", "x")
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
		{"claim1_unreadable", `{% include "m.html" %}`, "m.html", true, fs.ErrPermission, nil},
		{"claim2_unreadable_in_list", `{% include ["m.html", "z.html"] %}`, "m.html", true, fs.ErrPermission, nil},
		{"claim3_missing", `{% include "nope.html" %}`, "nope.html", false, runtime.ErrTemplateNotFound, nil},
		// guards: errors that do not already name the template still name it
		{"compile_error_names_template", `{% include "bad.html" %}`, "bad.html", false, nil, nil},
		{"no_loader_names_template", `{% include "x.html" %}`, "x.html", false, nil, noLoader},
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
			if n := strings.Count(err.Error(), `"`+tt.tmpl+`"`); n != 1 {
				t.Errorf("%q appears %d times, want 1: %v", tt.tmpl, n, err)
			}
		})
	}

	t.Run("claim4_ignore_missing", func(t *testing.T) {
		got, err := env.RenderString(`{% include "nope.html" ignore missing %}`, miya.NewContext())
		if err != nil || got != "" {
			t.Errorf("got %q, %v; want \"\", nil", got, err)
		}
	})
}
