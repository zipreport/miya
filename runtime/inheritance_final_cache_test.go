package runtime

import (
	"testing"

	"github.com/zipreport/miya/lexer"
	"github.com/zipreport/miya/parser"
)

func parseForInheritanceTest(t *testing.T, src string) *parser.TemplateNode {
	t.Helper()
	tokens, err := lexer.NewLexer(src, nil).Tokenize()
	if err != nil {
		t.Fatal(err)
	}
	tmpl, err := parser.NewParser(tokens).Parse()
	if err != nil {
		t.Fatal(err)
	}
	return tmpl
}

// The merged template depends on the render's autoescape setting (a child's
// {% autoescape %} around super() keeps the parent block's own escaping), so
// one cache serves both settings correctly.
func TestMergedTemplatePerAutoescape(t *testing.T) {
	env := newMockEnvironment()
	env.addTemplate("base", parseForInheritanceTest(t, `{% block b %}<i>{{ v }}</i>{% endblock %}`))
	env.addTemplate("child", parseForInheritanceTest(t, `{% extends 'base' %}{% block b %}{% autoescape false %}{{ super() }}{% endautoescape %}{% endblock %}`))
	p := NewInheritanceProcessorWithCache(env, NewInheritanceCache())
	child, _ := env.GetTemplate("child")
	want := map[bool]string{true: "<i>&lt;x&gt;</i>", false: "<i><x></i>"}
	for _, on := range []bool{true, false, true, false} {
		ctx := &autoescapeContext{Context: &simpleContext{variables: map[string]interface{}{"v": "<x>"}}, autoescape: on}
		ast, err := p.ResolveInheritance(child, ctx)
		if err != nil {
			t.Fatal(err)
		}
		out, err := NewEvaluator().EvalNode(ast, ctx)
		if err != nil || out != want[on] {
			t.Fatalf("autoescape=%v: got %v, %v; want %q", on, out, err, want[on])
		}
	}
}
