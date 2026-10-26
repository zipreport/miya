package runtime

import (
	"errors"
	"testing"

	"github.com/zipreport/miya/parser"
)

// ownErrLoader is a TemplateLoader that returns its own errors, which do not
// wrap ErrTemplateNotFound.
type ownErrLoader struct {
	special map[string]error
}

func (l *ownErrLoader) LoadTemplate(name string) (*parser.TemplateNode, error) {
	if err, ok := l.special[name]; ok {
		return nil, err
	}
	if name == "p" {
		return &parser.TemplateNode{Name: "p", Children: []parser.Node{parser.NewLiteralNode("P", "", 1, 1)}}, nil
	}
	return nil, errors.New("my loader: no such template")
}

func (l *ownErrLoader) TemplateExists(name string) bool { return name == "p" }

func TestIncludeCustomLoaderNotFound(t *testing.T) {
	newEval := func(l TemplateLoader) *DefaultEvaluator {
		evaluator := NewEvaluator()
		evaluator.SetImportSystem(NewImportSystem(l, evaluator))
		return evaluator
	}
	include := func(name string, ignore bool) *parser.IncludeNode {
		return &parser.IncludeNode{Template: parser.NewLiteralNode(name, "", 1, 1), IgnoreMissing: ignore}
	}

	t.Run("ignored", func(t *testing.T) {
		got, err := newEval(&ownErrLoader{}).EvalIncludeNode(include("nope", true), &simpleContext{variables: map[string]interface{}{}})
		if err != nil || got != "" {
			t.Fatalf("got %v, %v; want empty, nil", got, err)
		}
	})
	t.Run("error", func(t *testing.T) {
		if _, err := newEval(&ownErrLoader{}).EvalIncludeNode(include("nope", false), &simpleContext{variables: map[string]interface{}{}}); err == nil {
			t.Fatal("expected an error")
		}
	})
	t.Run("found", func(t *testing.T) {
		got, err := newEval(&ownErrLoader{}).EvalIncludeNode(include("p", false), &simpleContext{variables: map[string]interface{}{}})
		if err != nil || got != "P" {
			t.Fatalf("got %v, %v; want P, nil", got, err)
		}
	})
	t.Run("other_name", func(t *testing.T) {
		l := &ownErrLoader{special: map[string]error{"c": &TemplateNotFoundError{Name: "parent", Err: errors.New("x")}}}
		if _, err := newEval(l).EvalIncludeNode(include("c", true), &simpleContext{variables: map[string]interface{}{}}); err == nil {
			t.Fatal("a missing other template must not be ignored")
		}
	})
	t.Run("same_name", func(t *testing.T) {
		l := &ownErrLoader{special: map[string]error{"c": &TemplateNotFoundError{Name: "c", Err: errors.New("x")}}}
		got, err := newEval(l).EvalIncludeNode(include("c", true), &simpleContext{variables: map[string]interface{}{}})
		if err != nil || got != "" {
			t.Fatalf("got %v, %v; want empty, nil", got, err)
		}
	})
}

func TestEvalWithNodeWithoutNames(t *testing.T) {
	// An AST built without the parser has no Names/Exprs: the assignments are
	// evaluated in sorted key order.
	node := parser.NewWithNode(map[string]parser.ExpressionNode{
		"b": parser.NewLiteralNode("B", "", 1, 1),
		"a": parser.NewLiteralNode("A", "", 1, 1),
	}, []parser.Node{parser.NewIdentifierNode("a", 1, 1)}, 1, 1)
	got, err := NewEvaluator().EvalWithNode(node, &simpleContext{variables: map[string]interface{}{}})
	if err != nil || got != "A" {
		t.Fatalf("got %v, %v; want A, nil", got, err)
	}
}
