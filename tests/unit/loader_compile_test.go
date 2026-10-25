package miya_test

import (
	"os"
	"path/filepath"
	"testing"

	miya "github.com/zipreport/miya"
	"github.com/zipreport/miya/lexer"
	"github.com/zipreport/miya/loader"
)

func TestLoaderTemplatesCompileWithEnvironment(t *testing.T) {
	tests := []struct {
		name  string
		files map[string]string
		parse func() loader.TemplateParser
		opts  []miya.EnvironmentOption
		setup func(*miya.Environment)
		main  string
		want  string
	}{
		{
			name:  "inline_ws_control",
			files: map[string]string{"t.html": "a  {{- 1 -}}  b"},
			main:  "t.html",
			want:  "a1b",
		},
		{
			name:  "trim_blocks",
			files: map[string]string{"t.html": "{% if true %}\nx\n{% endif %}\ny"},
			opts:  []miya.EnvironmentOption{miya.WithTrimBlocks(true)},
			main:  "t.html",
			want:  "x\ny",
		},
		{
			name:  "custom_delimiters",
			files: map[string]string{"t.html": "<<= 1 >>"},
			setup: func(e *miya.Environment) { e.SetDelimiters("<<=", ">>", "<%", "%>") },
			main:  "t.html",
			want:  "1",
		},
		{
			name: "include_inherits",
			files: map[string]string{
				"main.html": `main: {% include "ws.html" %}`,
				"ws.html":   "a  {{- 1 -}}  b",
			},
			main: "main.html",
			want: "main: a1b",
		},
		{
			name:  "parser_config_ignored",
			files: map[string]string{"t.html": "<<= 1 >>"},
			parse: func() loader.TemplateParser {
				return loader.NewDirectTemplateParserWithConfig(&lexer.LexerConfig{
					VarStartString:     "<<=",
					VarEndString:       ">>",
					BlockStartString:   "{%",
					BlockEndString:     "%}",
					CommentStartString: "{#",
					CommentEndString:   "#}",
				})
			},
			main: "t.html",
			want: "<<= 1 >>",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			for n, c := range tt.files {
				if err := os.WriteFile(filepath.Join(dir, n), []byte(c), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			var p loader.TemplateParser = loader.NewDirectTemplateParser()
			if tt.parse != nil {
				p = tt.parse()
			}
			fl := loader.NewFileSystemLoader([]string{dir}, p)
			env := miya.NewEnvironment(append([]miya.EnvironmentOption{miya.WithLoader(fl), miya.WithAutoEscape(false)}, tt.opts...)...)
			if tt.setup != nil {
				tt.setup(env)
			}
			got, err := env.RenderTemplate(tt.main, miya.NewContext())
			if err != nil {
				t.Fatalf("render: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}

	t.Run("string_loader", func(t *testing.T) {
		sl := loader.NewStringLoader(loader.NewDirectTemplateParser())
		sl.AddTemplate("ws.html", "a  {{- 1 -}}  b")
		env := miya.NewEnvironment(miya.WithLoader(sl), miya.WithAutoEscape(false))
		got, err := env.RenderTemplate("ws.html", miya.NewContext())
		if err != nil {
			t.Fatalf("render: %v", err)
		}
		if got != "a1b" {
			t.Errorf("got %q, want %q", got, "a1b")
		}
	})
}
