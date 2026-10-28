package runtime

import (
	"runtime/debug"
	"testing"
	"time"
)

// A filter's list result that contains itself is un-marked without
// overflowing the stack.
func TestUnmarkItemsSelfContaining(t *testing.T) {
	defer debug.SetMaxStack(debug.SetMaxStack(64 << 20))
	cs := []interface{}{Markup("<b>"), nil}
	cs[1] = cs
	out := unmarkItems([]interface{}{Markup("<i>"), cs}, nil)
	if s, ok := out[0].(string); !ok || s != "<i>" {
		t.Fatalf("out[0] = %#v, want plain string", out[0])
	}
	if s, ok := out[1].([]interface{})[0].(string); !ok || s != "<b>" {
		t.Fatalf("nested item not un-marked: %#v", out[1].([]interface{})[0])
	}

	// The top-level list is tracked too: where it recurs it is left as is,
	// like a nested list.
	top := unmarkItems(cs, nil)
	if s, ok := top[0].(string); !ok || s != "<b>" {
		t.Fatalf("top[0] = %#v, want plain string", top[0])
	}
	if sub := top[1].([]interface{}); &sub[0] != &cs[0] {
		t.Fatalf("top[1] is a copy, want the original list where it recurs")
	}
}

// FmtCycles walks each shared list once: a 40-level DAG whose levels hold
// the level below twice has 2^40 paths but 41 lists.
func TestFmtCyclesSharedListsLinear(t *testing.T) {
	var dag interface{} = []interface{}{1}
	for i := 0; i < 40; i++ {
		dag = []interface{}{dag, dag}
	}
	cs := []interface{}{1, nil}
	cs[1] = cs
	done := make(chan [2]bool, 1)
	go func() { done <- [2]bool{FmtCycles(dag), FmtCycles([]interface{}{dag, cs})} }()
	select {
	case got := <-done:
		if got != [2]bool{false, true} {
			t.Fatalf("FmtCycles(dag), FmtCycles([dag, cycle]) = %v, want [false true]", got)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("FmtCycles did not finish on a DAG of shared lists")
	}
}

// unmarkItems un-marks each shared list once, so a DAG of shared lists takes
// linear time.
func TestUnmarkItemsSharedListsLinear(t *testing.T) {
	var dag interface{} = []interface{}{Markup("<i>")}
	for i := 0; i < 40; i++ {
		dag = []interface{}{dag, dag}
	}
	done := make(chan interface{}, 1)
	go func() { done <- unmarkItems(dag.([]interface{}), nil) }()
	select {
	case got := <-done:
		leaf := got
		for i := 0; i < 40; i++ {
			leaf = leaf.([]interface{})[1]
		}
		if s, ok := leaf.([]interface{})[0].(string); !ok || s != "<i>" {
			t.Fatalf("leaf = %#v, want the plain string \"<i>\"", leaf.([]interface{})[0])
		}
	case <-time.After(10 * time.Second):
		t.Fatal("unmarkItems did not finish on a DAG of shared lists")
	}
}

// The fallback filters, used without an environment, print a self-containing
// list without overflowing the stack.
func TestFallbackFiltersSelfContaining(t *testing.T) {
	defer debug.SetMaxStack(debug.SetMaxStack(64 << 20))
	cs := []interface{}{"a", nil}
	cs[1] = cs
	e := NewEvaluator()
	for name, want := range map[string]string{
		"upper": "[A [...]]", "lower": "[a [...]]", "capitalize": "[a [...]]",
		"trim": "[a [...]]", "escape": "[a [...]]",
	} {
		got, err := e.applyFilter(name, cs, nil)
		if err != nil || ToString(got) != want {
			t.Errorf("%s: got %v, %v; want %q", name, got, err, want)
		}
	}
	if got, _ := e.applyFilter("upper", []interface{}{"a", 1}, nil); got != "[A 1]" {
		t.Errorf("upper of a plain list: got %v", got)
	}
}

// The fallback sequence, mapping and iterable tests follow pointers, like
// the branching package.
func TestApplyTestFollowsPointers(t *testing.T) {
	e := NewEvaluator()
	m := map[string]interface{}{"a": 1}
	type st struct{ A int }
	rows := []struct {
		test  string
		value interface{}
		want  bool
	}{
		{"sequence", &[]int{1}, true},
		{"iterable", &[]int{1}, true},
		{"mapping", &m, true},
		{"iterable", &m, true},
		{"sequence", &st{}, false},
		{"mapping", &st{}, false},
		{"iterable", &st{}, false},
		{"sequence", (*[]int)(nil), false},
	}
	for _, r := range rows {
		got, err := e.applyTest(r.test, r.value, nil)
		if err != nil || got != r.want {
			t.Errorf("%T is %s = %v, %v; want %v", r.value, r.test, got, err, r.want)
		}
	}
}
