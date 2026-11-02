package miya_test

import (
	"strings"
	"testing"

	miya "github.com/zipreport/miya"
	"github.com/zipreport/miya/extensions"
	"github.com/zipreport/miya/loader"
	"github.com/zipreport/miya/parser"
)

// Expected values are Jinja2 3.1.6's.

func issue38Env(templates map[string]string) *miya.Environment {
	sl := loader.NewStringLoader(loader.NewDirectTemplateParser())
	for k, v := range templates {
		sl.AddTemplate(k, v)
	}
	return miya.NewEnvironment(miya.WithLoader(sl))
}

// issue38Render renders template c twice, so a cached build is checked too.
func issue38Render(t *testing.T, templates map[string]string) []string {
	t.Helper()
	env := issue38Env(templates)
	var got []string
	for i := 0; i < 2; i++ {
		tmpl, err := env.GetTemplate("c")
		if err != nil {
			t.Fatalf("build: %v", err)
		}
		out, err := tmpl.Render(miya.NewContext())
		if err != nil {
			got = append(got, "error: "+err.Error())
		} else {
			got = append(got, out)
		}
	}
	return got
}

func issue38WantNoParent(t *testing.T, templates map[string]string, block string) {
	t.Helper()
	want := "there is no parent block called '" + block + "'"
	for i, got := range issue38Render(t, templates) {
		if !strings.HasPrefix(got, "error: ") || !strings.Contains(got, want) {
			t.Errorf("render %d: got %q; want an error containing %q", i+1, got, want)
		}
	}
}

func issue38WantOutput(t *testing.T, templates map[string]string, want string) {
	t.Helper()
	for i, got := range issue38Render(t, templates) {
		if got != want {
			t.Errorf("render %d: got %q; want %q", i+1, got, want)
		}
	}
}

// Claim 1: in a single template, {{ super() }} in a block raises an error
// naming the block, from a loader and from FromString.
func TestIssue38SingleTemplateSuperErrors(t *testing.T) {
	const src = "[{% block a %}x{{ super() }}{% endblock %}]"
	t.Run("loader", func(t *testing.T) {
		issue38WantNoParent(t, map[string]string{"c": src}, "a")
	})
	t.Run("from_string", func(t *testing.T) {
		tmpl, err := miya.NewEnvironment().FromString(src)
		if err != nil {
			t.Fatalf("build: %v", err)
		}
		out, err := tmpl.Render(miya.NewContext())
		if err == nil || !strings.Contains(err.Error(), "there is no parent block called 'a'") {
			t.Fatalf("got %q, %v; want the no-parent-block error", out, err)
		}
	})
}

// Claim 2: the same with super() inside an expression.
func TestIssue38SingleTemplateSuperExpressionErrors(t *testing.T) {
	issue38WantNoParent(t, map[string]string{"c": "[{% block a %}x{{ super()|upper }}{% endblock %}]"}, "a")
}

// Claim 3: a super() in the base's block, reached from the child's super(),
// raises the same error, as a whole expression and inside one.
func TestIssue38BaseBlockSuperErrors(t *testing.T) {
	const child = "{% extends 'base' %}{% block a %}a{{ super() }}{% endblock %}"
	t.Run("whole", func(t *testing.T) {
		issue38WantNoParent(t, map[string]string{"base": "[{% block a %}A{{ super() }}{% endblock %}]", "c": child}, "a")
	})
	t.Run("expression", func(t *testing.T) {
		issue38WantNoParent(t, map[string]string{"base": "[{% block a %}A{{ super()|lower }}{% endblock %}]", "c": child}, "a")
	})
	t.Run("expression_child_calls_no_super", func(t *testing.T) {
		issue38WantNoParent(t, map[string]string{"base": "[{% block a %}A{% endblock %}{% block b %}B{{ super()|upper }}{% endblock %}]", "c": "{% extends 'base' %}{% block a %}a{% endblock %}"}, "b")
	})
}

// Claim 5: a block only the child defines has no parent block either.
func TestIssue38ChildOnlyBlockSuperErrors(t *testing.T) {
	const base = "[{% block a %}A{% endblock %}]"
	t.Run("hoisted_from_loop", func(t *testing.T) {
		issue38WantNoParent(t, map[string]string{"base": base, "c": "{% extends 'base' %}{% for i in [1] %}{% block z %}z{{ super() }}{% endblock %}{% endfor %}"}, "z")
	})
	t.Run("nested_in_override", func(t *testing.T) {
		issue38WantNoParent(t, map[string]string{"base": base, "c": "{% extends 'base' %}{% block a %}{% block z %}z{{ super() }}{% endblock %}{% endblock %}"}, "z")
	})
}

// Claim 4: a super() that is never evaluated does not fail.
func TestIssue38UntakenSuperRenders(t *testing.T) {
	tests := []struct {
		name      string
		templates map[string]string
		want      string
	}{
		{"single_if", map[string]string{"c": "[{% block a %}x{% if false %}{{ super() }}{% endif %}{% endblock %}]"}, "[x]"},
		{"single_ternary", map[string]string{"c": "[{% block a %}x{{ super() if false else 'n' }}{% endblock %}]"}, "[xn]"},
		{"base_if", map[string]string{
			"base": "[{% block a %}A{% if false %}{{ super() }}{% endif %}{% endblock %}]",
			"c":    "{% extends 'base' %}{% block a %}a{{ super() }}{% endblock %}",
		}, "[aA]"},
		{"base_ternary", map[string]string{
			"base": "[{% block a %}A{{ super() if false else '' }}{% endblock %}]",
			"c":    "{% extends 'base' %}{% block a %}a{{ super() }}{% endblock %}",
		}, "[aA]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) { issue38WantOutput(t, tt.templates, tt.want) })
	}
}

// Guard: in a single template the error reaches super() wherever it sits in
// the block, and names the innermost block.
func TestIssue38SingleTemplateSuperShapes(t *testing.T) {
	tests := []struct{ name, src, block string }{
		{"nested_block", "[{% block a %}{% block b %}b{{ super() }}{% endblock %}{% endblock %}]", "b"},
		{"with_body", "[{% block a %}{% with q = 1 %}{{ super() }}{% endwith %}{% endblock %}]", "a"},
		{"call_body", "{% macro m() %}{{ caller() }}{% endmacro %}[{% block a %}{% call m() %}{{ super() }}{% endcall %}{% endblock %}]", "a"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) { issue38WantNoParent(t, map[string]string{"c": tt.src}, tt.block) })
	}
}

// Guard: a super() with a parent block still renders it.
func TestIssue38SuperWithParentRenders(t *testing.T) {
	issue38WantOutput(t, map[string]string{
		"base": "[{% block a %}A{% endblock %}]",
		"c":    "{% extends 'base' %}{% block a %}a{{ super() }}{{ super()|lower }}{% endblock %}",
	}, "[aAa]")
}

// Claim 7: a block of an imported module calls super() with no parent block.
func TestIssue38ImportedModuleSuperErrors(t *testing.T) {
	const mod = "{% macro f() %}F{% endmacro %}[{% block a %}i{{ super() }}{% endblock %}]"
	for _, tt := range []struct{ name, src string }{
		{"import", "{% import 'mod' as m %}x{{ m.f() }}"},
		{"from_import", "{% from 'mod' import f %}x{{ f() }}"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			issue38WantNoParent(t, map[string]string{"mod": mod, "c": tt.src}, "a")
		})
	}
	t.Run("untaken_super", func(t *testing.T) {
		issue38WantOutput(t, map[string]string{
			"mod": "{% macro f() %}F{% endmacro %}[{% block a %}i{% if false %}{{ super() }}{% endif %}{% endblock %}]",
			"c":   "{% import 'mod' as m %}x{{ m.f() }}",
		}, "xF")
	})
}

// Claim 8: a bare super is undefined where there is no parent block, so the
// Jinja2 guard {% if super is defined %} works.
func TestIssue38BareSuperIsUndefined(t *testing.T) {
	const child = "{% extends 'base' %}{% block a %}a{{ super() }}{% endblock %}"
	tests := []struct {
		name      string
		templates map[string]string
		want      string
	}{
		{"guard_in_base", map[string]string{"base": "[{% block a %}A{% if super is defined %}{{ super() }}{% else %}-{% endif %}{% endblock %}]", "c": child}, "[aA-]"},
		{"guard_in_child_only_block", map[string]string{"base": "[{% block a %}A{% endblock %}]", "c": "{% extends 'base' %}{% block a %}{% block z %}z{% if super is defined %}{{ super() }}{% endif %}{% endblock %}{% endblock %}"}, "[z]"},
		{"default_filter", map[string]string{"base": "[{% block a %}A{{ super|default('') }}{% endblock %}]", "c": child}, "[aA]"},
		{"whole_bare", map[string]string{"c": "[{% block a %}x{{ super }}{% endblock %}]"}, "[x]"},
		{"guard_child_calls_no_super", map[string]string{"base": "[{% block a %}A{% endblock %}{% block b %}B{% if super is defined %}S{% endif %}{% endblock %}]", "c": "{% extends 'base' %}{% block a %}a{% endblock %}"}, "[aB]"},
		{"default_child_calls_no_super", map[string]string{"base": "[{% block a %}A{% endblock %}{% block b %}B{{ super|default('d') }}{% endblock %}]", "c": "{% extends 'base' %}{% block a %}a{% endblock %}"}, "[aBd]"},
		{"single_guard", map[string]string{"c": "[{% block a %}x{% if super is defined %}{{ super() }}{% endif %}{% endblock %}]"}, "[x]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) { issue38WantOutput(t, tt.templates, tt.want) })
	}
	t.Run("defined_with_parent", func(t *testing.T) {
		issue38WantOutput(t, map[string]string{
			"base": "[{% block a %}A{% endblock %}]",
			"c":    "{% extends 'base' %}{% block a %}{% if super is defined %}{{ super() }}{% endif %}a{% endblock %}",
		}, "[Aa]")
	})
}

type issue38Tree struct {
	Parent *issue38Tree
	Kids   []*issue38Tree
}

type issue38CyclicExtension struct {
	*extensions.BaseExtension
}

func (x *issue38CyclicExtension) ParseTag(tagName string, p extensions.ExtensionParser) (parser.Node, error) {
	tok := p.Current()
	node := p.NewExtensionNode("cyclic", tagName, tok.Line, tok.Column)
	root := &issue38Tree{}
	root.Kids = []*issue38Tree{{Parent: root}}
	node.SetProperty("tree", root)
	node.SetEvaluateFunc(func(*parser.ExtensionNode, interface{}) (interface{}, error) { return "M", nil })
	return node, p.ExpectBlockEnd()
}

// Guard: deciding whether a template uses super does not walk the values an
// extension node holds, which may refer to themselves.
func TestIssue38ExtensionPropertiesNotWalked(t *testing.T) {
	env := miya.NewEnvironment()
	if err := env.AddExtension(&issue38CyclicExtension{extensions.NewBaseExtension("cyclic", []string{"meta"})}); err != nil {
		t.Fatal(err)
	}
	for src, want := range map[string]string{
		"{% meta %}x": "Mx",
		"{% meta %}[{% block a %}{% if false %}{{ super() }}{% endif %}{% endblock %}]": "M[]",
	} {
		tmpl, err := env.FromString(src)
		if err != nil {
			t.Fatalf("build: %v", err)
		}
		if out, err := tmpl.Render(miya.NewContext()); err != nil || out != want {
			t.Fatalf("%s: got %q, %v; want %q", src, out, err, want)
		}
	}
	// a template parsed with extensions records its super() too
	tmpl, err := env.FromString("{% meta %}[{% block a %}{{ super()|upper }}{% endblock %}]")
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if out, err := tmpl.Render(miya.NewContext()); err == nil || !strings.Contains(err.Error(), "there is no parent block called 'a'") {
		t.Fatalf("got %q, %v; want the no-parent-block error", out, err)
	}
}

// Claim 1: the error gives the line of the super() call.
func TestIssue38ErrorGivesLine(t *testing.T) {
	for _, call := range []string{"{{ super() }}", "{{ super()|upper }}"} {
		for i, got := range issue38Render(t, map[string]string{"c": "[{% block a %}\n\nx" + call + "{% endblock %}]"}) {
			if !strings.Contains(got, "there is no parent block called 'a' at line 3") {
				t.Errorf("%s render %d: got %q; want the error at line 3", call, i+1, got)
			}
		}
	}
}

// Claim 8 under a child {% autoescape %}: the guard still sees super undefined.
func TestIssue38BareSuperUndefinedUnderAutoescape(t *testing.T) {
	sl := loader.NewStringLoader(loader.NewDirectTemplateParser())
	sl.AddTemplate("c", "[{% block a %}{% autoescape false %}{% if super is defined %}{{ super() }}{% endif %}{% endautoescape %}x{% endblock %}]")
	env := miya.NewEnvironment(miya.WithLoader(sl), miya.WithAutoEscape(true))
	tmpl, err := env.GetTemplate("c")
	if err != nil {
		t.Fatal(err)
	}
	if out, err := tmpl.Render(miya.NewContext()); err != nil || out != "[x]" {
		t.Fatalf("got %q, %v; want %q", out, err, "[x]")
	}
}

// Claim 8 with strict undefined behaviour: using a bare super with no parent
// block fails, as in Jinja2, but "is defined" still sees it undefined.
func TestIssue38BareSuperStrict(t *testing.T) {
	render := func(src string) (string, error) {
		tmpl, err := miya.NewEnvironment(miya.WithStrictUndefined(true)).FromString(src)
		if err != nil {
			t.Fatalf("build: %v", err)
		}
		return tmpl.Render(miya.NewContext())
	}
	for _, src := range []string{
		"[{% block a %}x{{ super }}{% endblock %}]",
		"[{% block a %}x{{ super ~ 'q' }}{% endblock %}]",
		"[{% block a %}x{% if super %}T{% endif %}{% endblock %}]",
	} {
		if out, err := render(src); err == nil || !strings.Contains(err.Error(), "there is no parent block called 'a'") {
			t.Errorf("%s: got %q, %v; want the no-parent-block error", src, out, err)
		}
	}
	if out, err := render("[{% block a %}x{% if super is defined %}{{ super() }}{% endif %}{% endblock %}]"); err != nil || out != "[x]" {
		t.Errorf("guard: got %q, %v; want %q", out, err, "[x]")
	}
}
