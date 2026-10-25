package conformance

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"testing"

	"github.com/zipreport/miya"
)

// renders is how many times each case is rendered. A case matches only when
// every render matches, so output that depends on Go map order (dict
// iteration, with-assignment order) cannot pass by chance: a 2-key map whose
// expected order Go happens to produce 89% of the time still fails 200 renders
// with probability 1 - 0.89^200, about 1 - 1e-10.
const renders = 200

var dump = flag.String("conformance.dump", "", "write every current divergence to this JSON file")

// Divergence kinds. An allow-listed case must keep diverging the same way;
// a change of kind (wrong output becoming a panic, say) is reported.
const (
	kindOutput = "output" // wrong output, or output where Jinja2 raises
	kindError  = "error"  // an error where Jinja2 renders, or an error in the wrong phase
	kindPanic  = "panic"  // miya panicked (recovered by this test)
)

type conformanceCase struct {
	ID        string                 `json:"id"`
	Templates map[string]string      `json:"templates"`
	Entry     string                 `json:"entry"`
	Context   map[string]interface{} `json:"context"`
	Options   struct {
		Autoescape          bool `json:"autoescape"`
		TrimBlocks          bool `json:"trim_blocks"`
		LstripBlocks        bool `json:"lstrip_blocks"`
		KeepTrailingNewline bool `json:"keep_trailing_newline"`
		Strict              bool `json:"strict"`
	} `json:"options"`
	Expect struct {
		Output *string `json:"output"`
		Error  string  `json:"error"`
		Phase  string  `json:"phase"`
	} `json:"expect"`
}

type divergence struct {
	Plan   string `json:"plan"`
	Kind   string `json:"kind"`
	Reason string `json:"reason"`
}

// result is one miya render: its output, or an error and the phase it came
// from ("compile" for GetTemplate, "render" for Render).
type result struct {
	out      string
	err      error
	phase    string
	panicked bool
}

func loadJSON(t *testing.T, path string, v interface{}) {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, v); err != nil {
		t.Fatalf("%s: %v", path, err)
	}
}

// mapLoader serves templates from memory through GetSource only, the path
// zipreport-go's ReportFileLoader takes. miya's StringLoader is an
// AdvancedLoader that parses templates itself and so skips the environment's
// trim_blocks, lstrip_blocks and "-" whitespace handling.
type mapLoader map[string]string

func (m mapLoader) GetSource(name string) (string, error) {
	src, ok := m[name]
	if !ok {
		return "", fmt.Errorf("template not found: %s", name)
	}
	return src, nil
}

func (m mapLoader) IsCached(name string) bool { return false }

func (m mapLoader) ListTemplates() ([]string, error) {
	names := make([]string, 0, len(m))
	for name := range m {
		names = append(names, name)
	}
	return names, nil
}

func render(c conformanceCase) (r result) {
	r.phase = "compile"
	defer func() {
		if p := recover(); p != nil {
			r = result{err: fmt.Errorf("panic: %v", p), phase: r.phase, panicked: true}
		}
	}()
	env := miya.NewEnvironment(
		miya.WithLoader(mapLoader(c.Templates)),
		miya.WithAutoEscape(c.Options.Autoescape),
		miya.WithTrimBlocks(c.Options.TrimBlocks),
		miya.WithLstripBlocks(c.Options.LstripBlocks),
		miya.WithKeepTrailingNewline(c.Options.KeepTrailingNewline),
		miya.WithStrictUndefined(c.Options.Strict),
	)
	tmpl, err := env.GetTemplate(c.Entry)
	if err != nil {
		return result{err: err, phase: "compile"}
	}
	r.phase = "render"
	out, err := tmpl.Render(miya.NewContextFrom(c.Context))
	return result{out: out, err: err, phase: "render"}
}

// diverges reports how miya's result differs from Jinja2's as a kind and a
// reason, or ("", "") when it matches. For expected errors the phase is
// compared, not the error class (miya's error types do not map onto Jinja2's).
func diverges(c conformanceCase, r result) (kind, reason string) {
	switch {
	case r.panicked:
		return kindPanic, fmt.Sprintf("%s-time %v", r.phase, r.err)
	case c.Expect.Output == nil && r.err == nil:
		return kindOutput, fmt.Sprintf("want %s-time error %s, got output %q", c.Expect.Phase, c.Expect.Error, r.out)
	case c.Expect.Output == nil && r.phase != c.Expect.Phase:
		return kindError, fmt.Sprintf("want %s-time error %s, got %s-time error: %v", c.Expect.Phase, c.Expect.Error, r.phase, r.err)
	case c.Expect.Output == nil:
		return "", ""
	case r.err != nil:
		return kindError, fmt.Sprintf("want %q, got %s-time error: %v", *c.Expect.Output, r.phase, r.err)
	case r.out != *c.Expect.Output:
		return kindOutput, fmt.Sprintf("want %q, got %q", *c.Expect.Output, r.out)
	}
	return "", ""
}

func TestConformance(t *testing.T) {
	var suite struct {
		Cases []conformanceCase `json:"cases"`
	}
	loadJSON(t, "cases.json", &suite)
	known := map[string]divergence{}
	loadJSON(t, "known_divergences.json", &known)

	seen := map[string]bool{}
	current := map[string]divergence{}
	for _, c := range suite.Cases {
		seen[c.ID] = true
		// Stop at the first divergent render, except when dumping: then render
		// every time and keep the lexicographically smallest reason, so the
		// dump is the same on every run even for map-order cases.
		var kind, reason string
		for i := 0; i < renders; i++ {
			k, r := diverges(c, render(c))
			if k != "" && (kind == "" || r < reason) {
				kind, reason = k, r
			}
			if kind != "" && *dump == "" {
				break
			}
		}
		if kind != "" {
			current[c.ID] = divergence{Kind: kind, Reason: reason}
		}
		entry, allowed := known[c.ID]
		switch {
		case kind != "" && !allowed:
			t.Errorf("%s: diverges from Jinja2: %s", c.ID, reason)
		case kind == "" && allowed:
			t.Errorf("%s: now matches Jinja2; remove it from known_divergences.json (plan %s)", c.ID, entry.Plan)
		case kind != "" && kind != entry.Kind:
			t.Errorf("%s: divergence kind changed from %q to %q: %s; if intended, update its kind and reason in known_divergences.json and say so in your report", c.ID, entry.Kind, kind, reason)
		}
	}
	for id := range known {
		if !seen[id] {
			t.Errorf("%s: in known_divergences.json but not in cases.json", id)
		}
	}
	if *dump != "" {
		ids := make([]string, 0, len(current))
		for id := range current {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		b, _ := json.MarshalIndent(current, "", " ")
		if err := os.WriteFile(*dump, b, 0o644); err != nil {
			t.Fatal(err)
		}
		t.Logf("%d divergences written to %s", len(ids), *dump)
	}
}
