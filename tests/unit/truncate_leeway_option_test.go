package miya_test

import (
	"testing"

	miya "github.com/zipreport/miya"
)

func TestTruncateLeewayOption(t *testing.T) {
	zero := miya.NewEnvironment(miya.WithAutoEscape(false), miya.WithTruncateLeeway(0))
	def := miya.NewEnvironment(miya.WithAutoEscape(false))

	tests := []struct {
		name     string
		env      *miya.Environment
		template string
		want     string
	}{
		{"env_leeway0", zero, `{{ 'Hello World'|truncate(8) }}`, "Hello..."},
		{"env_leeway_kw_wins", zero, `{{ 'Hello World'|truncate(8, leeway=5) }}`, "Hello World"},
		{"env_leeway_positional_wins", zero, `{{ 'Hello World'|truncate(8, false, '...', 5) }}`, "Hello World"},
		{"env_default_untouched", def, `{{ 'Hello World'|truncate(8) }}`, "Hello World"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.env.RenderString(tc.template, miya.NewContextFrom(nil))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}
