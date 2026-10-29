package miya_test

import (
	"testing"

	miya "github.com/zipreport/miya"
)

// A macro call, a for-if filter and a comprehension each get their own scope:
// names they bind do not leak out, and a macro defined inside a macro reads
// its enclosing scope when it runs, as in Jinja2.
func TestScopeSemantics(t *testing.T) {
	tests := []struct{ name, src, want string }{
		{"macro_locals_do_not_leak", "{% macro m(a) %}{% set b = 1 %}{% endmacro %}{{ m(1) }}{{ a is defined }}{{ b is defined }}", "falsefalse"},
		{"for_if_target_does_not_leak", "{% for r in [1, 2] if r > 1 %}{{ r }}{% endfor %}{{ r is defined }}", "2false"},
		{"list_comprehension_target_does_not_leak", "{{ [r for r in [1]]|length }}{{ r is defined }}", "1false"},
		{"dict_comprehension_target_does_not_leak", "{{ {r: r for r in [1]}|length }}{{ r is defined }}", "1false"},
		{"caller_locals_do_not_leak", "{% macro m() %}{{ caller(1) }}{% endmacro %}{% call(c) m() %}{% set d = c %}{{ d }}{% endcall %}{{ c is defined }}{{ d is defined }}", "1falsefalse"},
		// Jinja2 3.1.6 renders "2": the nested macro sees the later {% set x = 2 %}.
		{"escaped_nested_macro_reads_late", "{% set ns = namespace() %}{% set x = 1 %}{% macro outer() %}{% macro inner() %}{{ x }}{% endmacro %}{% set ns.f = inner %}{% endmacro %}{{ outer() }}{% set x = 2 %}{{ ns.f() }}", "2"},
	}
	env := miya.NewEnvironment()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := env.RenderString(tt.src, miya.NewContext())
			if err != nil || got != tt.want {
				t.Fatalf("got %q, %v; want %q", got, err, tt.want)
			}
		})
	}
}
