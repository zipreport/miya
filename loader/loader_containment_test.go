package loader

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestLoaderContainment(t *testing.T) {
	write := func(t *testing.T, p, content string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	link := func(t *testing.T, target, name string) {
		t.Helper()
		if err := os.Symlink(target, name); err != nil {
			t.Skipf("symlinks unavailable: %v", err)
		}
	}
	// setup returns base, root and a loader (followLinks false) over root.
	setup := func(t *testing.T) (string, string, *FileSystemLoader) {
		t.Helper()
		base := t.TempDir()
		root := filepath.Join(base, "root")
		if err := os.MkdirAll(filepath.Join(base, "outside"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(root, 0o755); err != nil {
			t.Fatal(err)
		}
		return base, root, NewFileSystemLoader([]string{root}, NewDirectTemplateParser())
	}
	wantSource := func(t *testing.T, l *FileSystemLoader, name, want string) {
		t.Helper()
		got, err := l.GetSource(name)
		if err != nil {
			t.Fatalf("GetSource(%q): %v", name, err)
		}
		if got != want {
			t.Fatalf("GetSource(%q) = %q, want %q", name, got, want)
		}
	}
	wantErr := func(t *testing.T, l *FileSystemLoader, name string) {
		t.Helper()
		got, err := l.GetSource(name)
		if err == nil {
			t.Fatalf("GetSource(%q) = %q, want error", name, got)
		}
	}

	t.Run("symlinked_dir_escape_rejected", func(t *testing.T) {
		base, root, l := setup(t)
		write(t, filepath.Join(base, "outside", "secret.html"), "SECRET")
		link(t, filepath.Join(base, "outside"), filepath.Join(root, "linkdir"))
		got, err := l.GetSource("linkdir/secret.html")
		if err == nil || got == "SECRET" {
			t.Fatalf("got %q, %v; want error and no secret", got, err)
		}
	})
	t.Run("symlinked_dir_followed_when_enabled", func(t *testing.T) {
		base, root, l := setup(t)
		write(t, filepath.Join(base, "outside", "secret.html"), "SECRET")
		link(t, filepath.Join(base, "outside"), filepath.Join(root, "linkdir"))
		l.SetFollowLinks(true)
		wantSource(t, l, "linkdir/secret.html", "SECRET")
	})
	t.Run("symlinked_file_escape_rejected", func(t *testing.T) {
		base, root, l := setup(t)
		write(t, filepath.Join(base, "outside", "secret.html"), "SECRET")
		link(t, filepath.Join(base, "outside", "secret.html"), filepath.Join(root, "f.html"))
		wantErr(t, l, "f.html")
	})
	t.Run("relative_in_root_dir_link_ok", func(t *testing.T) {
		_, root, l := setup(t)
		write(t, filepath.Join(root, "v2", "p.html"), "P")
		link(t, "v2", filepath.Join(root, "latest"))
		wantSource(t, l, "latest/p.html", "P")
	})
	t.Run("relative_in_root_file_link_ok", func(t *testing.T) {
		_, root, l := setup(t)
		write(t, filepath.Join(root, "m.html"), "M")
		link(t, "m.html", filepath.Join(root, "alias.html"))
		wantSource(t, l, "alias.html", "M")
	})
	t.Run("absolute_in_root_link_rejected", func(t *testing.T) {
		_, root, l := setup(t)
		write(t, filepath.Join(root, "v2", "p.html"), "P")
		link(t, filepath.Join(root, "v2"), filepath.Join(root, "abs"))
		wantErr(t, l, "abs/p.html")
	})
	t.Run("symlinked_root_ok", func(t *testing.T) {
		base, root, _ := setup(t)
		write(t, filepath.Join(root, "m.html"), "M")
		write(t, filepath.Join(root, "v2", "p.html"), "P")
		link(t, "v2", filepath.Join(root, "latest"))
		link(t, root, filepath.Join(base, "rootlink"))
		l := NewFileSystemLoader([]string{filepath.Join(base, "rootlink")}, NewDirectTemplateParser())
		wantSource(t, l, "m.html", "M")
		wantSource(t, l, "latest/p.html", "P")
	})
	t.Run("plain_nested_file_ok", func(t *testing.T) {
		_, root, l := setup(t)
		write(t, filepath.Join(root, "sub", "p.html"), "P")
		wantSource(t, l, "sub/p.html", "P")
	})
	t.Run("directory_skipped_for_extension", func(t *testing.T) {
		_, root, l := setup(t)
		write(t, filepath.Join(root, "sub", "p.html"), "P")
		write(t, filepath.Join(root, "sub.html"), "S")
		wantSource(t, l, "sub", "S")
	})
	t.Run("dotdot_inside_segment_ok", func(t *testing.T) {
		_, root, l := setup(t)
		write(t, filepath.Join(root, "a..b.html"), "D")
		wantSource(t, l, "a..b.html", "D")
	})
	t.Run("dotdot_segment_rejected", func(t *testing.T) {
		for _, follow := range []bool{false, true} {
			base, root, l := setup(t)
			l.SetFollowLinks(follow)
			// Without the ".." check each name would resolve to one of these.
			write(t, filepath.Join(root, "m.html"), "M")
			write(t, filepath.Join(base, "x.html"), "OUTSIDE")
			if err := os.MkdirAll(filepath.Join(root, "sub"), 0o755); err != nil {
				t.Fatal(err)
			}
			names := []string{"sub/../m.html", "../x.html", "a/../../x.html"}
			if filepath.Separator == '\\' {
				names = append(names, `..\x.html`)
			} else {
				// On POSIX a backslash is an ordinary character: this is a file in root.
				write(t, filepath.Join(root, `..\x.html`), "LITERAL")
				wantSource(t, l, `..\x.html`, "LITERAL")
			}
			for _, name := range names {
				got, err := l.GetSource(name)
				if err == nil || !strings.Contains(err.Error(), "invalid template name") {
					t.Errorf("followLinks=%v: GetSource(%q) = %q, %v; want invalid template name", follow, name, got, err)
				}
			}
		}
	})
	t.Run("backslash_is_literal_on_posix", func(t *testing.T) {
		if filepath.Separator == '\\' {
			t.Skip("backslash is a separator on this OS")
		}
		_, root, l := setup(t)
		write(t, filepath.Join(root, `a\b.html`), "BS")
		write(t, filepath.Join(root, "v2", "p.html"), "P")
		wantSource(t, l, `a\b.html`, "BS")
		wantErr(t, l, `v2\p.html`)
	})
	t.Run("cache_sees_mtime", func(t *testing.T) {
		_, root, l := setup(t)
		p := filepath.Join(root, "m.html")
		write(t, p, "v1")
		first, err := l.LoadTemplate("m.html")
		if err != nil {
			t.Fatal(err)
		}
		write(t, p, "v2")
		later := time.Now().Add(time.Second)
		if err := os.Chtimes(p, later, later); err != nil {
			t.Fatal(err)
		}
		second, err := l.LoadTemplate("m.html")
		if err != nil {
			t.Fatal(err)
		}
		if first == second {
			t.Fatal("LoadTemplate returned the stale cached AST after the file changed")
		}
	})
	t.Run("nil_parser_errors", func(t *testing.T) {
		_, root, _ := setup(t)
		write(t, filepath.Join(root, "m.html"), "M")
		try := func(name string, f func() error) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("%s panicked: %v", name, r)
				}
			}()
			if err := f(); err == nil {
				t.Fatalf("%s: want error", name)
			}
		}
		try("FileSystemLoader", func() error {
			_, err := NewFileSystemLoader([]string{root}, nil).LoadTemplate("m.html")
			return err
		})
		try("StringLoader", func() error {
			s := NewStringLoader(nil)
			s.AddTemplate("m.html", "x")
			_, err := s.LoadTemplate("m.html")
			return err
		})
	})
	t.Run("concurrent_stats", func(t *testing.T) {
		_, root, l := setup(t)
		write(t, filepath.Join(root, "m.html"), "M")
		var wg sync.WaitGroup
		for i := 0; i < 8; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < 50; j++ {
					_, _ = l.LoadTemplate("m.html")
				}
			}()
		}
		wg.Wait()
		s := l.GetCacheStats()
		if s.Hits+s.Misses != 400 {
			t.Fatalf("Hits+Misses = %d, want 400", s.Hits+s.Misses)
		}
	})
}
