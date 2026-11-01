package parser

import (
	"testing"

	"github.com/zipreport/miya/lexer"
)

// The parser marks a for loop whose body cannot read loop, so the evaluator
// skips building it.
func TestForLoopUnused(t *testing.T) {
	tests := []struct {
		name, src string
		unused    bool
	}{
		{"plain", "{% for i in x %}{{ i|upper }}{% endfor %}", true},
		{"else_only", "{% for i in x %}{{ i }}{% else %}{{ loop }}{% endfor %}", true},
		{"attribute_named_loop", "{% for i in x %}{{ i.loop }}{% endfor %}", true},
		{"names_loop", "{% for i in x %}{{ loop.index }}{% endfor %}", false},
		{"names_loop_nested", "{% for i in x %}{% if i %}{% set l = [loop] %}{% endif %}{% endfor %}", false},
		{"macro_names_loop", "{% for i in x %}{% macro m() %}{{ loop.index }}{% endmacro %}{{ m() }}{% endfor %}", false},
		{"inner_loop_plain", "{% for i in x %}{% for j in i %}{{ j }}{% endfor %}{% endfor %}", true},
		{"inner_loop_names_loop", "{% for i in x %}{% for j in i %}{{ loop.index }}{% endfor %}{% endfor %}", false},
		{"inner_loop_else_names_loop", "{% for i in x %}{% for j in i %}{{ j }}{% else %}{{ loop.index }}{% endfor %}{% endfor %}", false},
		{"inner_loop_iterable_names_loop", "{% for i in x %}{% for j in loop.cycle(i, i) %}{{ j }}{% endfor %}{% endfor %}", false},
		{"inner_loop_holds_block", "{% for i in x %}{% for j in i %}{% block b %}{% endblock %}{% endfor %}{% endfor %}", false},
		{"recursive", "{% for i in x recursive %}{{ i }}{% endfor %}", false},
		{"block", "{% for i in x %}{% block b %}{% endblock %}{% endfor %}", false},
		{"include", "{% for i in x %}{% include 'y' %}{% endfor %}", false},
		{"import", "{% for i in x %}{% import 'y' as m with context %}{% endfor %}", false},
		{"from", "{% for i in x %}{% from 'y' import f with context %}{% endfor %}", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens, err := lexer.NewLexer(tt.src, nil).Tokenize()
			if err != nil {
				t.Fatal(err)
			}
			tmpl, err := NewParser(tokens).Parse()
			if err != nil {
				t.Fatal(err)
			}
			f, ok := tmpl.Children[0].(*ForNode)
			if !ok {
				t.Fatalf("got %T", tmpl.Children[0])
			}
			if f.LoopUnused != tt.unused {
				t.Errorf("LoopUnused = %v, want %v", f.LoopUnused, tt.unused)
			}
		})
	}
}
