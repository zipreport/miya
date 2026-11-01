package miya_test

import (
	"fmt"
	"runtime"
	"sync/atomic"
	"testing"

	miya "github.com/zipreport/miya"
)

// scopeAllocPerRender renders src over 2000 rows with vars extra data keys and
// returns the bytes allocated per render.
func scopeAllocPerRender(t *testing.T, src string, vars int) uint64 {
	t.Helper()
	env := miya.NewEnvironment(miya.WithAutoEscape(false))
	tmpl, err := env.FromString(src)
	if err != nil {
		t.Fatal(err)
	}
	rows := make([]interface{}, 2000)
	for i := range rows {
		rows[i] = float64(i)
	}
	data := map[string]interface{}{"rows": rows}
	for i := 0; i < vars; i++ {
		data[fmt.Sprintf("v%d", i)] = i
	}
	if _, err := tmpl.Render(miya.NewContextFrom(data)); err != nil {
		t.Fatalf("render: %v", err)
	}
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	for i := 0; i < 3; i++ {
		if _, err := tmpl.Render(miya.NewContextFrom(data)); err != nil {
			t.Fatal(err)
		}
	}
	runtime.ReadMemStats(&after)
	return (after.TotalAlloc - before.TotalAlloc) / 3
}

// A macro call, a for-if filter and a comprehension must not copy every
// variable in scope per call or per item.
func TestScopeCostIndependentOfContextWidth(t *testing.T) {
	tests := []struct{ name, src string }{
		{"macro_call", "{% macro cell(r) %}{{ r }}{% endmacro %}{% for r in rows %}{{ cell(r) }}{% endfor %}"},
		{"for_if", "{% for r in rows if r > -1 %}{{ r }}{% endfor %}"},
		{"list_comprehension", "{{ [r for r in rows]|length }}"},
		{"dict_comprehension", "{{ {r: r for r in rows}|length }}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			narrow, wide := scopeAllocPerRender(t, tt.src, 5), scopeAllocPerRender(t, tt.src, 200)
			if narrow == 0 {
				t.Fatal("no allocation measured for the narrow context")
			}
			if ratio := float64(wide) / float64(narrow); ratio > 1.5 {
				t.Fatalf("allocates %.1fx more with 200 variables in scope than with 5 (narrow %d B, wide %d B); the scope must not be copied per call or item", ratio, narrow, wide)
			}
		})
	}
}

type countedStringer struct{ calls *int64 }

func (c countedStringer) String() string { atomic.AddInt64(c.calls, 1); return "counted" }

// A recursive loop() must not format its children: nothing reads the text,
// and formatting calls String() on Go values the template never prints.
func TestRecursiveLoopDoesNotFormatChildren(t *testing.T) {
	var calls int64
	node := func(name string, children ...interface{}) map[string]interface{} {
		return map[string]interface{}{"name": name, "meta": countedStringer{&calls}, "children": children}
	}
	tree := []interface{}{node("a", node("b"), node("c"))}
	env := miya.NewEnvironment(miya.WithAutoEscape(false))
	got, err := env.RenderString("{% for n in tree recursive %}{{ n.name }}({{ loop(n.children) }}){% endfor %}",
		miya.NewContextFrom(map[string]interface{}{"tree": tree}))
	if err != nil || got != "a(b()c())" {
		t.Fatalf("got %q, %v; want %q", got, err, "a(b()c())")
	}
	if n := atomic.LoadInt64(&calls); n != 0 {
		t.Errorf("String() called %d times; the template never prints meta", n)
	}
}
