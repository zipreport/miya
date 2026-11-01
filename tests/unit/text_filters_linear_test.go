package miya_test

import (
	"testing"
	"time"

	miya "github.com/zipreport/miya"
)

func TestTextFiltersLinear(t *testing.T) {
	for _, tc := range []struct{ name, tpl, want string }{
		{"ww_long_word", `{{ ('a' * 800000)|wordwrap|length }}`, "810126"},
		{"ww_dashes", `{{ ('a-' * 100000)|wordwrap(3)|length }}`, "299998"},
		{"ww_hyphen_splits", `{{ ('ab-' * 66666)|wordwrap(300)|length }}`, "200664"},
		{"ww_hyphen_splits_narrow", `{{ ('ab-' * 66666)|wordwrap(3)|length }}`, "266663"},
		{"ww_blank_run", `{{ (' ' * 100000 ~ 'a')|wordwrap(3)|length }}`, "2"},
		{"uz_open_run", `{{ ('(' * 400000 ~ 'http://x.com')|urlize|length }}`, "400054"},
		{"uz_close_run", `{{ ('http://x.com' ~ ')' * 400000)|urlize|length }}`, "400054"},
		{"uz_balanced_run", `{{ ('http://x.com/' ~ '(' * 200000 ~ ')' * 200000)|urlize|length }}`, "800056"},
		{"uz_balance_small", `{{ ('(' * 3 ~ 'http://x.com/' ~ '(' * 2 ~ ')' * 5 ~ '.')|urlize }}`, `(((<a href="http://x.com/(())" rel="noopener">http://x.com/(())</a>))).`},
		{"ww_blank_lead", `{{ (' ' * 100000 ~ 'a')|wordwrap(1)|length }}`, "1"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			env := miya.NewEnvironment()
			tmpl, err := env.FromString(tc.tpl)
			if err != nil {
				t.Fatal(err)
			}
			start := time.Now()
			got, err := tmpl.Render(miya.NewContext())
			d := time.Since(start)
			if err != nil {
				t.Fatal(err)
			}
			if d > 2*time.Second {
				t.Errorf("took %v, want < 2s", d)
			}
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}
