package miya_test

import (
	"testing"

	miya "github.com/zipreport/miya"
	"github.com/zipreport/miya/loader"
)

// Issue #13: super() renders the parent block with the most-derived
// definition of each block nested inside it.
func TestIssue13SuperNestedOverrides(t *testing.T) {
	const base = "[{% block o %}<{% block i %}i{% endblock %}>{% endblock %}]"
	const ext = "{% extends 'base' %}"
	two := func(b, child string) map[string]string {
		return map[string]string{"base": b, "child": ext + child}
	}
	// want values are Jinja2 3.1.6's output
	tests := []struct {
		name       string
		autoescape bool
		templates  map[string]string
		want       string
	}{
		{"claim1_override", false, two(base, "{% block o %}{{ super() }}{% endblock %}{% block i %}I{% endblock %}"), "[<I>]"},
		{"claim2_inner_super_expr", false, two(base, "{% block o %}{{ super() }}{% endblock %}{% block i %}{{ super() ~ '!' }}{% endblock %}"), "[<i!>]"},
		{"claim3_both_expr", false, two(base, "{% block o %}{{ super()|upper }}{% endblock %}{% block i %}{{ super() ~ '!' }}{% endblock %}"), "[<I!>]"},
		{"claim4_head_title", false, two("<head>{% block head %}<title>{% block title %}Site{% endblock %}</title>{% endblock %}</head>",
			"{% block head %}{{ super() }}<link>{% endblock %}{% block title %}Page{% endblock %}"), "<head><title>Page</title><link></head>"},
		{"claim5_three_levels", false, map[string]string{
			"base":  "{% block o %}{% block i %}bi{% endblock %}{% endblock %}",
			"mid":   "{% extends 'base' %}{% block o %}({% block i %}mi[{{ super() }}]{% endblock %}){% endblock %}",
			"child": "{% extends 'mid' %}{% block o %}{{ super() }}{% endblock %}",
		}, "(mi[bi])"},
		// further shapes
		{"inner_whole_super", false, two(base, "{% block o %}{{ super() }}{% endblock %}{% block i %}{{ super() }}I{% endblock %}"), "[<iI>]"},
		{"deeply_nested", false, two("[{% block o %}<{% block m %}({% block i %}i{% endblock %}){% endblock %}>{% endblock %}]",
			"{% block o %}{{ super() }}{% endblock %}{% block i %}I{% endblock %}"), "[<(I)>]"},
		{"inner_defined_in_child_outer", false, two(base, "{% block o %}{{ super() }}|{% block i %}I{{ super() }}{% endblock %}{% endblock %}"), "[<Ii>|Ii]"},
		{"super_twice", false, two(base, "{% block o %}{{ super() }}{{ super() }}{% endblock %}{% block i %}I{% endblock %}"), "[<I><I>]"},
		{"middle_overrides_inner", false, map[string]string{
			"base":  base,
			"mid":   "{% extends 'base' %}{% block i %}M{% endblock %}",
			"child": "{% extends 'mid' %}{% block o %}{{ super() }}!{% endblock %}",
		}, "[<M>!]"},
		{"middle_and_child_inner", false, map[string]string{
			"base":  base,
			"mid":   "{% extends 'base' %}{% block i %}M{{ super() }}{% endblock %}",
			"child": "{% extends 'mid' %}{% block o %}{{ super() }}{% endblock %}{% block i %}C{{ super() }}{% endblock %}",
		}, "[<CMi>]"},
		{"four_levels", false, map[string]string{
			"base":  base,
			"m1":    "{% extends 'base' %}{% block o %}1{{ super() }}{% endblock %}",
			"m2":    "{% extends 'm1' %}{% block i %}2{{ super() }}{% endblock %}",
			"child": "{% extends 'm2' %}{% block o %}c{{ super() }}{% endblock %}{% block i %}C{{ super() }}{% endblock %}",
		}, "[c1<C2i>]"},
		// a super() cycle that is never rendered does not fail
		{"cycle_in_untaken_parent_branch", false, two("{% block a %}A{% if false %}{% block b %}B{% endblock %}{% endif %}{% endblock %}",
			"{% block b %}b{% block a %}a{{ super() }}{% endblock %}{% endblock %}"), "aA"},
		{"cycle_in_untaken_child_branch", false, two("{% block a %}A{% if false %}{% block b %}B{% endblock %}{% endif %}{% endblock %}",
			"{% block b %}b{% block a %}a{% if false %}{{ super() }}{% endif %}{% endblock %}{% endblock %}"), "a"},
		{"cycle_in_untaken_ternary", false, two("{% block a %}A{% if false %}{% block b %}B{% endblock %}{% endif %}{% endblock %}",
			"{% block b %}b{% block a %}a{{ super() if false else '' }}{% endblock %}{% endblock %}"), "a"},
		{"autoescape", true, two("[{% block o %}<{% block i %}{{ '<i>' }}{% endblock %}>{% endblock %}]",
			"{% block o %}{{ super() }}{% endblock %}{% block i %}{{ '<I>' }}{{ super() }}{% endblock %}"), "[<&lt;I&gt;&lt;i&gt;>]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := issue13Env(tt.templates, tt.autoescape).GetTemplate("child")
			if err != nil {
				t.Fatal(err)
			}
			got, err := tmpl.Render(miya.NewContext())
			if err != nil {
				t.Fatalf("render: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}

	// a parent block whose nested override calls super() back into it renders
	// itself forever (Jinja2 3.1.6: RecursionError)
	t.Run("cycle_errors", func(t *testing.T) {
		tmpl, err := issue13Env(two("{% block a %}A{% block b %}B{% endblock %}{% endblock %}",
			"{% block b %}{{ super() }}{% block a %}{{ super() }}{% endblock %}{% endblock %}"), false).GetTemplate("child")
		if err != nil {
			t.Fatal(err)
		}
		if got, err := tmpl.Render(miya.NewContext()); err == nil {
			t.Errorf("want an error, got %q", got)
		}
	})
}

func issue13Env(templates map[string]string, autoescape bool) *miya.Environment {
	sl := loader.NewStringLoader(loader.NewDirectTemplateParser())
	for n, s := range templates {
		sl.AddTemplate(n, s)
	}
	return miya.NewEnvironment(miya.WithLoader(sl), miya.WithAutoEscape(autoescape))
}
