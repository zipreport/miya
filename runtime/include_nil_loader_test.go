package runtime

import (
	"testing"

	"github.com/zipreport/miya/parser"
)

type nilTemplateLoader struct{}

func (nilTemplateLoader) LoadTemplate(string) (*parser.TemplateNode, error) { return nil, nil }
func (nilTemplateLoader) TemplateExists(string) bool                        { return false }

// Issue #11 guard: a loader that returns no template and no error makes the
// include fail instead of panicking.
func TestIncludeLoaderReturnsNothing(t *testing.T) {
	e := NewEvaluator()
	e.SetImportSystem(NewImportSystem(nilTemplateLoader{}, e))
	node := parser.NewIncludeNode(parser.NewLiteralNode("x", `"x"`, 1, 1), 1, 1)
	if _, err := e.EvalIncludeNode(node, &simpleContext{variables: map[string]interface{}{}}); err == nil {
		t.Error("want an error")
	}
}
