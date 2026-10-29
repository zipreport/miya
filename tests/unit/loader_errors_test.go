package miya_test

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	miya "github.com/zipreport/miya"
	"github.com/zipreport/miya/loader"
	"github.com/zipreport/miya/runtime"
)

func unreadableDir(t *testing.T) string {
	dir := t.TempDir()
	p := filepath.Join(dir, "t.html")
	if err := os.WriteFile(p, []byte("secret"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(p, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(p, 0o644) })
	if _, err := os.ReadFile(p); err == nil {
		t.Skip("a mode-000 file is still readable here (root, or Windows)")
	}
	return dir
}
func stringLoader(src string) *loader.StringLoader {
	sl := loader.NewStringLoader(loader.NewDirectTemplateParser())
	sl.AddTemplate("t.html", src)
	return sl
}

// A template that exists but cannot be read is an error, as in Jinja2
// (PermissionError); only a template that does not exist is "not found".
func TestLoaderReadErrors(t *testing.T) {
	permDenied := loader.LoaderFunc(func(name string) (string, error) {
		return "", &fs.PathError{Op: "open", Path: name, Err: fs.ErrPermission}
	})
	plainErr := loader.LoaderFunc(func(name string) (string, error) {
		return "", fmt.Errorf("no template %s", name)
	})
	pathErr := func(err error) loader.LoaderFunc {
		return func(name string) (string, error) {
			return "", &fs.PathError{Op: "open", Path: name, Err: err}
		}
	}

	tests := []struct {
		name         string
		loader       func(t *testing.T) miya.Loader
		wantNotFound bool   // errors.Is(GetTemplate err, runtime.ErrTemplateNotFound)
		wantWrapped  error  // the GetTemplate and include errors must wrap this; nil: no check
		wantInclude  string // output of the ignore-missing include, "" means an error
	}{
		{"func_permission_denied", func(*testing.T) miya.Loader { return permDenied }, false, fs.ErrPermission, ""},
		{"func_io_error", func(*testing.T) miya.Loader { return pathErr(syscall.EIO) }, false, syscall.EIO, ""},
		{"func_plain_error_is_missing", func(*testing.T) miya.Loader { return plainErr }, true, nil, "[]"},
		{"func_not_exist_is_missing", func(*testing.T) miya.Loader { return pathErr(fs.ErrNotExist) }, true, nil, "[]"},
		// os.ReadFile / fs.ReadFile errors for names that are not files: Jinja2 says not found.
		{"func_not_dir_is_missing", func(*testing.T) miya.Loader { return pathErr(syscall.ENOTDIR) }, true, nil, "[]"},
		{"func_invalid_path_is_missing", func(*testing.T) miya.Loader { return pathErr(fs.ErrInvalid) }, true, nil, "[]"},
		{"filesystem_unreadable", func(t *testing.T) miya.Loader {
			return loader.NewFileSystemLoader([]string{unreadableDir(t)}, loader.NewDirectTemplateParser())
		}, false, fs.ErrPermission, ""},
		{"filesystem_missing", func(t *testing.T) miya.Loader {
			return loader.NewFileSystemLoader([]string{t.TempDir()}, loader.NewDirectTemplateParser())
		}, true, nil, "[]"},
		{"chain_unreadable_first", func(t *testing.T) miya.Loader {
			fsl := loader.NewFileSystemLoader([]string{unreadableDir(t)}, loader.NewDirectTemplateParser())
			return loader.NewChainLoader(fsl, stringLoader("fallback"))
		}, false, fs.ErrPermission, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := miya.NewEnvironment(miya.WithLoader(tt.loader(t)))
			_, err := env.GetTemplate("t.html")
			if err == nil {
				t.Fatal("GetTemplate: want an error")
			}
			if got := errors.Is(err, runtime.ErrTemplateNotFound); got != tt.wantNotFound {
				t.Errorf("errors.Is(err, ErrTemplateNotFound) = %v, want %v (err: %v)", got, tt.wantNotFound, err)
			}
			if tt.wantWrapped != nil && !errors.Is(err, tt.wantWrapped) {
				t.Errorf("GetTemplate error does not wrap %v: %v", tt.wantWrapped, err)
			}
			out, err := env.RenderString(`[{% include 't.html' ignore missing %}]`, miya.NewContext())
			if tt.wantInclude == "" {
				if err == nil {
					t.Fatalf("include ignore missing: want an error, got %q", out)
				}
				if !errors.Is(err, tt.wantWrapped) {
					t.Errorf("include error does not wrap %v: %v", tt.wantWrapped, err)
				}
				return
			}
			if err != nil || out != tt.wantInclude {
				t.Errorf("include ignore missing: got %q, %v; want %q", out, err, tt.wantInclude)
			}
		})
	}

	// A template that exists in an earlier loader but does not compile is an
	// error; the chain does not fall through to a later loader's copy.
	t.Run("chain_compile_error_first", func(t *testing.T) {
		env := miya.NewEnvironment(miya.WithLoader(loader.NewChainLoader(stringLoader("{% if %}"), stringLoader("fallback"))))
		tmpl, err := env.GetTemplate("t.html")
		if err == nil {
			out, _ := tmpl.Render(miya.NewContext())
			t.Fatalf("want a compile error, got a template rendering %q", out)
		}
		if errors.Is(err, runtime.ErrTemplateNotFound) {
			t.Errorf("compile error reported as not found: %v", err)
		}
	})
}

func TestChainLoaderDirect(t *testing.T) {
	fsl := func(dir string) *loader.FileSystemLoader {
		return loader.NewFileSystemLoader([]string{dir}, loader.NewDirectTemplateParser())
	}

	t.Run("load_unreadable_first", func(t *testing.T) {
		c := loader.NewChainLoader(fsl(unreadableDir(t)), stringLoader("fallback"))
		ast, err := c.LoadTemplate("t.html")
		if err == nil || !errors.Is(err, fs.ErrPermission) {
			t.Fatalf("err = %v, want permission error", err)
		}
		if ast != nil {
			t.Fatal("AST must be nil")
		}
	})
	t.Run("meta_unreadable_first", func(t *testing.T) {
		c := loader.NewChainLoader(fsl(unreadableDir(t)), stringLoader("fallback"))
		if _, err := c.GetSourceWithMetadata("t.html"); !errors.Is(err, fs.ErrPermission) {
			t.Fatalf("err = %v, want permission error", err)
		}
	})
	t.Run("load_compile_error_first", func(t *testing.T) {
		c := loader.NewChainLoader(stringLoader("{% if %}"), stringLoader("fallback"))
		ast, err := c.LoadTemplate("t.html")
		if err == nil {
			t.Fatal("want a parse error")
		}
		if ast != nil {
			t.Fatal("AST must be nil")
		}
	})
	t.Run("load_missing_first_falls_through", func(t *testing.T) {
		c := loader.NewChainLoader(fsl(t.TempDir()), stringLoader("fallback"))
		ast, err := c.LoadTemplate("t.html")
		if err != nil || ast == nil {
			t.Fatalf("ast = %v, err = %v, want fallback template", ast, err)
		}
	})
	t.Run("load_all_missing", func(t *testing.T) {
		c := loader.NewChainLoader(fsl(t.TempDir()))
		if _, err := c.LoadTemplate("t.html"); err == nil {
			t.Fatal("want an error")
		}
	})
}
