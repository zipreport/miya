package miya_test

import (
	"strings"
	"testing"

	miya "github.com/zipreport/miya"
)

func TestEnumerateArguments(t *testing.T) {
	env := miya.NewEnvironment()
	loop := func(args string) string {
		return "{% for i, v in enumerate(['a','b']" + args + ") %}{{ i }}{{ v }} {% endfor %}"
	}

	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected string
		wantErr  bool
		wantMsg  string
	}{
		{name: "start_kw", template: loop(", start=5"), expected: "5a 6b "},
		{name: "start_pos", template: loop(", 5"), expected: "5a 6b "},
		{name: "start_negative", template: loop(", -1"), expected: "-1a 0b "},
		{name: "start_string_errors", template: loop(", '5'"), wantErr: true},
		{name: "start_float_errors", template: loop(", 2.5"), wantErr: true},
		{name: "unknown_kw_errors", template: loop(", stat=5"), wantErr: true},
		{name: "too_many_errors", template: loop(", 1, 2"), wantErr: true},
		{name: "start_and_positional", template: "{% for i, v in enumerate(['a'], 1, start=2) %}{{ i }}{{ v }}{% endfor %}", wantErr: true, wantMsg: "unexpected keyword argument"},
		{
			name:     "lone_map_is_not_keywords",
			template: "{% for i, k in enumerate(d) %}{{ i }}{{ k }}{% endfor %}",
			data:     map[string]interface{}{"d": map[string]interface{}{"a": 1}},
			wantErr:  true,
			wantMsg:  "not iterable",
		},
		{
			name:     "int64_start",
			template: "{% for i, v in enumerate(['a'], n) %}{{ i }}{{ v }}{% endfor %}",
			data:     map[string]interface{}{"n": int64(7)},
			expected: "7a",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := miya.NewContext()
			for k, v := range tt.data {
				ctx.Set(k, v)
			}
			got, err := env.RenderString(tt.template, ctx)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %q", got)
				}
				want := tt.wantMsg
				if want == "" {
					want = "enumerate()"
				}
				if !strings.Contains(err.Error(), want) {
					t.Fatalf("error does not contain %q: %v", want, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.expected {
				t.Fatalf("got %q, want %q", got, tt.expected)
			}
		})
	}
}
