package miya_test

import (
	"strings"
	"testing"
	"time"

	miya "github.com/zipreport/miya"
)

func TestStripTagsLinear(t *testing.T) {
	env := miya.NewEnvironment(miya.WithAutoEscape(false))
	for _, tc := range []struct{ name, in, want string }{
		{"comments", strings.Repeat("<!---->x", 200000), strings.Repeat("x", 200000)},
		{"tags", strings.Repeat("<b>x", 200000), strings.Repeat("x", 200000)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tmpl, err := env.FromString(`{{ v|striptags }}`)
			if err != nil {
				t.Fatal(err)
			}
			start := time.Now()
			got, err := tmpl.Render(miya.NewContextFrom(map[string]interface{}{"v": tc.in}))
			if err != nil {
				t.Fatal(err)
			}
			if d := time.Since(start); d > 2*time.Second {
				t.Errorf("striptags took %v, want < 2s", d)
			}
			if got != tc.want {
				t.Errorf("got %d bytes, want %d", len(got), len(tc.want))
			}
		})
	}
}
