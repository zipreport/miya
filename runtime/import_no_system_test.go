package runtime

import (
	"strings"
	"testing"

	"github.com/zipreport/miya/parser"
)

func TestImportWithoutImportSystem(t *testing.T) {
	e := NewEvaluator()
	ctx := &simpleContext{variables: make(map[string]interface{})}
	name := parser.NewLiteralNode("m.html", "", 1, 1)
	if _, err := e.EvalImportNode(parser.NewImportNode(1, 1, name, "m"), ctx); err == nil || !strings.Contains(err.Error(), "no import system") {
		t.Fatalf("import: got %v, want a no-import-system error", err)
	}
	if _, err := e.EvalFromNode(parser.NewFromNode(1, 1, name, []string{"s"}, map[string]string{}), ctx); err == nil || !strings.Contains(err.Error(), "no import system") {
		t.Fatalf("from: got %v, want a no-import-system error", err)
	}
}
