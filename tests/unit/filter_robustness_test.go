package miya_test

import (
	"errors"
	"runtime/debug"
	"strings"
	"testing"

	miya "github.com/zipreport/miya"
	"github.com/zipreport/miya/runtime"
)

func renderWith(t *testing.T, autoescape bool, src string, data map[string]interface{}) (string, error) {
	t.Helper()
	env := miya.NewEnvironment(miya.WithAutoEscape(autoescape))
	tmpl, err := env.FromString(src)
	if err != nil {
		t.Fatalf("parse %q: %v", src, err)
	}
	return tmpl.Render(miya.NewContextFrom(data))
}

func expectOutput(t *testing.T, src string, data map[string]interface{}, want string) {
	t.Helper()
	got, err := renderWith(t, false, src, data)
	if err != nil {
		t.Fatalf("%s: unexpected error: %v", src, err)
	}
	if got != want {
		t.Errorf("%s: got %q, want %q", src, got, want)
	}
}

func expectError(t *testing.T, src string, data map[string]interface{}, contains string) {
	t.Helper()
	got, err := renderWith(t, false, src, data)
	if err == nil {
		t.Fatalf("%s: expected an error, got %q", src, got)
	}
	if !strings.Contains(err.Error(), contains) {
		t.Errorf("%s: error %q does not contain %q", src, err, contains)
	}
}

func TestFilterResultSizeCaps(t *testing.T) {
	for _, src := range []string{
		`{{ ("x"*100)|replace("", "y"*1500000)|length }}`,
		`{{ ("ab"*100)|replace("a", "y"*1500000, 60)|length }}`,
		`{{ (["a"]*100)|join("y"*1500000)|length }}`,
		`{{ (["a"|safe]*100)|join("y"*1500000)|length }}`,
		`{{ ("%(a)1000000s"*150)|format(a=1)|length }}`,
		`{{ ("%(a)1000000s"*150)|safe|format(a=1)|length }}`,
	} {
		_, err := renderWith(t, true, src, nil)
		if !errors.Is(err, runtime.ErrOutputTooLarge) {
			t.Errorf("%s: want ErrOutputTooLarge, got %v", src, err)
		}
	}
	// Results under the cap are unchanged.
	expectOutput(t, `{{ ("ab"*3)|replace("a", "xy", 2) }}`, nil, "xybxybab")
	expectOutput(t, `{{ "abc"|replace("", "-") }}`, nil, "-a-b-c-")
	expectOutput(t, `{{ [1, 2, 3]|join(", ") }}`, nil, "1, 2, 3")
	expectOutput(t, `{{ "%s=%d"|format("a", 1) }}`, nil, "a=1")
}

func TestRandomFilterPicksAnyItem(t *testing.T) {
	env := miya.NewEnvironment(miya.WithAutoEscape(false))
	tmpl, err := env.FromString(`{% for i in range(200) %}{{ items|random }},{% endfor %}`)
	if err != nil {
		t.Fatal(err)
	}
	for name, items := range map[string]interface{}{
		"interface_slice": []interface{}{1, 2, 3},
		"int_slice":       []int{1, 2, 3},
		"int_pointer":     &[]int{1, 2, 3},
		"array":           [3]string{"1", "2", "3"},
	} {
		out, err := tmpl.Render(miya.NewContextFrom(map[string]interface{}{"items": items}))
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		seen := map[string]bool{}
		for _, v := range strings.Split(strings.TrimSuffix(out, ","), ",") {
			if v != "1" && v != "2" && v != "3" {
				t.Fatalf("%s: random returned %q, not an item", name, v)
			}
			seen[v] = true
		}
		if len(seen) < 2 {
			t.Errorf("%s: 200 draws gave only %v", name, seen)
		}
	}

	tmpl, err = env.FromString(`{% for i in range(200) %}{{ "çab"|random }}{% endfor %}`)
	if err != nil {
		t.Fatal(err)
	}
	out, err := tmpl.Render(miya.NewContext())
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range out {
		if r != 'ç' && r != 'a' && r != 'b' {
			t.Fatalf("random of a string returned %q, not a whole character", string(r))
		}
	}
	if !strings.ContainsRune(out, 'ç') || strings.Count(out, "ç") == 200 {
		t.Errorf("200 draws did not pick both the non-ASCII and other characters: %q", out)
	}

	// Jinja2: an empty sequence gives undefined.
	expectOutput(t, `{{ []|random }}|{{ ([]|random) is undefined }}|{{ ""|random }}`, nil, "|true|")
	expectError(t, `{{ 5|random }}`, nil, "random")
}

func TestSelfReferencingDataDoesNotCrash(t *testing.T) {
	defer debug.SetMaxStack(debug.SetMaxStack(16 << 20))

	cs := []interface{}{1, nil}
	cs[1] = cs
	cyc := map[string]interface{}{"a": 1}
	cyc["self"] = cyc
	a := []interface{}{nil}
	b := []interface{}{a}
	a[0] = b
	data := map[string]interface{}{"cs": cs, "cyc": cyc, "a": a, "b": b,
		"pcyc": &cyc, "mpcs": map[string]interface{}{"a": &cs}, "lpcs": []interface{}{&cs}}

	// Both autoescape settings: with it off (outside HTML escaping)
	// runtime.unmarkItems also walks a filter's list result.
	check := func(src, want string) {
		t.Helper()
		for _, autoescape := range []bool{true, false} {
			got, err := renderWith(t, autoescape, src, data)
			if err != nil || got != want {
				t.Errorf("%s (autoescape %v): got %q, %v; want %q", src, autoescape, got, err, want)
			}
		}
	}
	check(`{{ cs|unique|length }}`, "2")
	check(`{{ [cs, cs]|unique|length }}`, "1")
	check(`{{ [cyc, cyc]|unique|length }}`, "1")
	check(`{{ [cs, cs]|sort|length }}`, "2")
	check(`{{ [cs, cs]|max|length }}`, "2")
	check(`{{ [cs, cs]|min|length }}`, "2")
	check(`{{ {"a": cs, "b": cs}|dictsort(by="value")|length }}`, "2")
	check(`{{ cyc|pprint|safe }}`, "{'a': 1, 'self': {...}}")
	check(`{{ cs|pprint(true)|safe }}`, "[1, [...]]")
	// A pointer to a self-containing list or map inside the printed value.
	check(`{{ {"a": pcyc}|pprint|safe }}`, "{a: &map[a:1 self:map[...]]}")
	check(`{{ [pcyc]|pprint|safe }}`, "[&map[a:1 self:map[...]]]")
	check(`{{ mpcs|pprint|safe }}`, "{a: &[1 [...]]}")
	check(`{{ lpcs|pprint|safe }}`, "[&[1 [...]]]")
	check(`{{ cs is in "abc" }}`, "false")
	// Two lists that contain each other never compare equal by identity.
	if _, err := renderWith(t, true, `{{ [a, b]|sort|length }}`, data); err == nil || !strings.Contains(err.Error(), "recursion") {
		t.Errorf("sort of mutually containing lists: want a recursion error, got %v", err)
	}

	// Values without cycles keep their output.
	expectOutput(t, `{{ {"a": [1, "x"]}|pprint }}`, nil, "{a: [1 x]}")
	expectOutput(t, `{{ [[1, "x"], [1, "X"]]|unique|length }}`, nil, "2")
}

type stringMap map[string]string

func TestPointerAndTypedCollectionsInFiltersAndTests(t *testing.T) {
	pm := map[string]string{"a": "1"}
	data := map[string]interface{}{
		"pxs":   &[]int{3, 1, 2},
		"pim":   &map[string]int{"a": 1},
		"pm":    &pm,
		"sm":    stringMap{"b": "<y>", "a": "x"},
		"yaml":  map[interface{}]interface{}{"a": 1, "b": "<x>"},
		"names": []string{"b", "a"},
	}
	expectOutput(t, `{{ pxs|first }}|{{ pxs|last }}|{{ pxs|list }}|{{ pxs|sort }}|{{ pxs|length }}`, data, "3|2|[3, 1, 2]|[1, 2, 3]|3")
	expectOutput(t, `{{ pxs|unique|list }}|{{ pxs|sum }}|{{ pxs|max }}|{{ pxs|join("-") }}`, data, "[3, 1, 2]|6|3|3-1-2")
	expectOutput(t, `{{ 2 is in pxs }}|{{ 5 is in pxs }}|{{ "a" is in pim }}|{{ pxs is sequence }}|{{ pim is mapping }}|{{ pxs is iterable }}`, data, "true|false|true|true|true|true")
	expectOutput(t, `{{ pim is sequence }}|{{ pxs is mapping }}`, data, "true|false")
	expectOutput(t, `<a{{ sm|xmlattr }}>`, data, `<a a="x" b="&lt;y&gt;">`)
	expectOutput(t, `<a{{ pim|xmlattr }}>`, data, `<a a="1">`)
	expectOutput(t, `<a{{ pm|xmlattr }}>`, data, `<a a="1">`)
	expectOutput(t, `<a{{ yaml|xmlattr }}>`, data, `<a a="1" b="&lt;x&gt;">`)

	// Jinja2 raises TypeError for a non-iterable; undefined stays empty.
	expectError(t, `{{ 5|first }}`, nil, "not iterable")
	expectError(t, `{{ 5|last }}`, nil, "not reversible")
	expectError(t, `{{ 5|list }}`, nil, "not iterable")
	expectOutput(t, `[{{ missing|first }}][{{ missing|last }}]{{ missing|list }}`, nil, "[][][]")
	expectOutput(t, `{{ "ab"|first }}{{ names|last }}{{ names|list }}`, data, "aa['b', 'a']")
}

func TestToJSONIndentIsBounded(t *testing.T) {
	for _, src := range []string{
		`{{ [1, 2]|tojson(100000000)|length }}`,
		`{{ 1|tojson(9223372036854775807) }}`,
		`{{ [[[[[[[[1]]]]]]]]|tojson(10000000)|length }}`,
	} {
		_, err := renderWith(t, false, src, nil)
		if !errors.Is(err, runtime.ErrOutputTooLarge) {
			t.Errorf("%s: want ErrOutputTooLarge, got %v", src, err)
		}
	}
	expectOutput(t, `{{ [1, 2]|tojson(2) }}`, nil, "[\n  1,\n  2\n]")
	expectOutput(t, `{{ {"a": [1, {}], "b": "x,y"}|tojson(1) }}`, nil, "{\n \"a\": [\n  1,\n  {}\n ],\n \"b\": \"x,y\"\n}")
	// Jinja2 (json.dumps) puts each item on its own line for 0 and negative indents.
	expectOutput(t, `{{ [1, 2]|tojson(-1) }}|{{ [1, 2]|tojson(0) }}`, nil, "[\n1,\n2\n]|[\n1,\n2\n]")
	expectOutput(t, `{{ [1, 2]|tojson }}`, nil, "[1,2]")
}

func TestSumAndReplaceRejectBadOperands(t *testing.T) {
	expectError(t, `{{ [1, "a", 2]|sum }}`, nil, "sum")
	expectError(t, `{{ [1, none, 2]|sum }}`, nil, "sum")
	expectError(t, `{{ [{"v": "a"}]|sum(attribute="v") }}`, nil, "sum")
	expectOutput(t, `{{ [1, true, 2.5]|sum }}|{{ [1, missing, 2]|sum }}|{{ [{"v": 1}, {}]|sum(attribute="v") }}`, nil, "4.5|3|1")

	expectError(t, `{{ "aaa"|replace("a", "b", "x") }}`, nil, "integer")
	expectError(t, `{{ "aaa"|replace("a", "b", 1.5) }}`, nil, "integer")
	expectError(t, `{{ "aaa"|replace("a", "b", "2") }}`, nil, "integer")
	expectOutput(t, `{{ "aaa"|replace("a", "b", none) }}|{{ "aaa"|replace("a", "b", true) }}|{{ "aaa"|replace("a", "b", 2) }}|{{ "aaa"|replace("a", "b", count=1) }}|{{ "aaa"|replace("a", "b", -1) }}`, nil, "bbb|baa|bba|baa|bbb")
}

func TestOutOfRangeIntegerArgumentsError(t *testing.T) {
	data := map[string]interface{}{"big": uint64(1) << 63}
	for _, src := range []string{
		`{{ "x"|pad_left(1e19) }}`,
		`{{ "x"|pad_right(-1e19) }}`,
		`{{ "x"|center(1e19) }}`,
		`{{ "x"|indent(1e19, true) }}`,
		`{{ "x"|pad_left(big) }}`,
	} {
		got, err := renderWith(t, false, src, data)
		if err == nil {
			t.Errorf("%s: expected an error, got %q", src, got)
		}
	}
	expectOutput(t, `{{ "x"|pad_left(3.0) }}|{{ "x"|center(3) }}`, nil, "  x| x ")
}

func TestUrlizeTruncateKeepsSafeInput(t *testing.T) {
	src := `{{ "http://x.com/?a=1&amp;b=2"|safe|urlizetruncate(50) }}`
	got, err := renderWith(t, true, src, nil)
	if err != nil {
		t.Fatal(err)
	}
	want := `<a href="http://x.com/?a=1&amp;b=2">http://x.com/?a=1&amp;b=2</a>`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	got, err = renderWith(t, true, `{{ "http://x.com/?a=1&b=2"|urlizetruncate(50) }}`, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("unsafe input: got %q, want %q", got, want)
	}
}
