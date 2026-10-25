package miya_test

import (
	"testing"

	miya "github.com/zipreport/miya"
	"github.com/zipreport/miya/loader"
)

func TestNoOutputRewrites(t *testing.T) {
	newEnv := func() (*miya.Environment, *loader.StringLoader) {
		sl := loader.NewStringLoader(loader.NewDirectTemplateParser())
		return miya.NewEnvironment(miya.WithLoader(sl), miya.WithAutoEscape(false)), sl
	}

	t.Run("name_blog_post", func(t *testing.T) {
		env, sl := newEnv()
		src := `By <a href="/u">Ann</a>`
		sl.AddTemplate("blog_post.html", src)
		tmpl, err := env.GetTemplate("blog_post.html")
		if err != nil {
			t.Fatal(err)
		}
		got, err := tmpl.Render(nil)
		if err != nil {
			t.Fatal(err)
		}
		if got != src {
			t.Errorf("got %q, want %q", got, src)
		}
	})

	t.Run("output_blog_post", func(t *testing.T) {
		env, _ := newEnv()
		src := `<div class="blog-post">By <a href="/u">Ann</a></div>`
		tmpl, err := env.FromString(src)
		if err != nil {
			t.Fatal(err)
		}
		got, err := tmpl.Render(nil)
		if err != nil {
			t.Fatal(err)
		}
		if got != src {
			t.Errorf("got %q, want %q", got, src)
		}
	})

	t.Run("empty_imported_macro", func(t *testing.T) {
		env, sl := newEnv()
		sl.AddTemplate("m.html", `{% macro e() %}{% endmacro %}`)
		tmpl, err := env.FromString(`{% import "m.html" as m %}[{{ m.e() }}]`)
		if err != nil {
			t.Fatal(err)
		}
		got, err := tmpl.Render(nil)
		if err != nil {
			t.Fatal(err)
		}
		if got != "[]" {
			t.Errorf("got %q, want %q", got, "[]")
		}
	})
}
