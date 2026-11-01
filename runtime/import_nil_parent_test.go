package runtime

import (
	"testing"

	"github.com/zipreport/miya/parser"
)

// childOnlyLoader returns a template that extends "p" and no template for "p".
type childOnlyLoader struct{}

func (childOnlyLoader) LoadTemplate(name string) (*parser.TemplateNode, error) {
	if name != "c" {
		return nil, nil
	}
	t := parser.NewTemplateNode("c", 1, 1)
	t.Children = []parser.Node{parser.NewExtendsNode(parser.NewLiteralNode("p", `"p"`, 1, 1), 1, 1)}
	return t, nil
}
func (childOnlyLoader) TemplateExists(name string) bool { return name == "c" }

// Issue #12 guard: a parent the loader returns no template for fails the
// import, as the imported template itself does.
func TestImportParentLoaderReturnsNothing(t *testing.T) {
	e := NewEvaluator()
	is := NewImportSystem(childOnlyLoader{}, e)
	ctx := &simpleContext{variables: map[string]interface{}{}}
	if _, err := is.LoadTemplateNamespace("c", ctx, e); err == nil {
		t.Error("want an error")
	}
}
