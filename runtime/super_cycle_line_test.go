package runtime

import (
	"testing"

	"github.com/zipreport/miya/parser"
)

// Issue #37 claim 3: the node standing for a super() cycle reports the line of
// the statement holding that super(), as superMissing does.
func TestIssue37SuperCycleNodeLine(t *testing.T) {
	env := newMockEnvironment()
	env.addTemplate("base", parseForInheritanceTest(t, `{% block a %}A{% block b %}B{% endblock %}{% endblock %}`))
	env.addTemplate("c", parseForInheritanceTest(t, "{% extends 'base' %}{% block b %}{{ super() }}\n\n{% block a %}{{ super() }}{% endblock %}{% endblock %}"))
	c, _ := env.GetTemplate("c")
	ast, err := NewInheritanceProcessorWithCache(env, NewInheritanceCache()).ResolveInheritance(c, &simpleContext{variables: map[string]interface{}{}})
	if err != nil {
		t.Fatal(err)
	}
	var cycle *superCycle
	parser.WalkStatements([]parser.Node{ast}, func(n parser.Node) bool {
		if sc, ok := n.(*superCycle); ok {
			cycle = sc
		}
		return cycle == nil
	})
	if cycle == nil || cycle.Line() != 3 {
		t.Fatalf("got %v; want a super() cycle at line 3", cycle)
	}
}
