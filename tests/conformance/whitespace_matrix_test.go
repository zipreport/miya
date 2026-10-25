package conformance

import (
	"fmt"
	"testing"

	"github.com/zipreport/miya"
)

type whitespaceCase struct {
	ID       string `json:"id"`
	Template string `json:"template"`
	Options  struct {
		TrimBlocks          bool `json:"trim_blocks"`
		LstripBlocks        bool `json:"lstrip_blocks"`
		KeepTrailingNewline bool `json:"keep_trailing_newline"`
	} `json:"options"`
	Expect struct {
		Output *string `json:"output"`
		Error  string  `json:"error"`
	} `json:"expect"`
}

func renderWhitespaceCase(c whitespaceCase) (out string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	env := miya.NewEnvironment(
		miya.WithLoader(mapLoader{"main": c.Template}),
		miya.WithAutoEscape(false),
		miya.WithTrimBlocks(c.Options.TrimBlocks),
		miya.WithLstripBlocks(c.Options.LstripBlocks),
		miya.WithKeepTrailingNewline(c.Options.KeepTrailingNewline),
	)
	tmpl, err := env.GetTemplate("main")
	if err != nil {
		return "", err
	}
	return tmpl.Render(miya.NewContextFrom(map[string]interface{}{"v": "V"}))
}

// TestWhitespaceMatrix has no allow-list: every case must match Jinja2.
func TestWhitespaceMatrix(t *testing.T) {
	var suite struct {
		Cases []whitespaceCase `json:"cases"`
	}
	loadJSON(t, "whitespace_cases.json", &suite)
	if len(suite.Cases) != 3240 {
		t.Fatalf("whitespace_cases.json has %d cases, want 3240", len(suite.Cases))
	}
	for _, c := range suite.Cases {
		out, err := renderWhitespaceCase(c)
		switch {
		case c.Expect.Output == nil && err == nil:
			t.Errorf("%s: want error %s, got output %q", c.ID, c.Expect.Error, out)
		case c.Expect.Output != nil && err != nil:
			t.Errorf("%s: want %q, got error: %v", c.ID, *c.Expect.Output, err)
		case c.Expect.Output != nil && out != *c.Expect.Output:
			t.Errorf("%s: want %q, got %q", c.ID, *c.Expect.Output, out)
		}
	}
}
