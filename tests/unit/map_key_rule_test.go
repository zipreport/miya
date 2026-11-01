package miya_test

import (
	"math"
	"strconv"
	"strings"
	"testing"
	"time"

	miya "github.com/zipreport/miya"
	"github.com/zipreport/miya/runtime"
)

type mkrStatus string

type mkrKey struct{ A int }

func TestMapKeyRule(t *testing.T) {
	one := "1"
	mk := runtime.Markup("k")
	sk := mkrStatus("k")
	skv := mkrKey{1}
	data := map[string]interface{}{
		"ms":      map[string]string{"None": "N", "": "E", "1": "one"},
		"ifm":     map[interface{}]interface{}{nil: "nil", 1: "one", "a": "A"},
		"yam":     map[interface{}]interface{}{"None": "N", "1": "one", "true": "T"},
		"d":       map[string]interface{}{"None": "N", "": "E"},
		"p":       &one,
		"bs":      []byte("1"),
		"ifm1":    map[interface{}]interface{}{1: "one"},
		"ift":     map[interface{}]interface{}{true: "T"},
		"hm":      map[interface{}]interface{}{runtime.Markup("k"): "MK"},
		"mk":      mk,
		"pmk":     &mk,
		"hs":      map[interface{}]interface{}{"k": "S"},
		"hn":      map[interface{}]interface{}{mkrStatus("k"): "NK"},
		"sk":      sk,
		"mb":      map[bool]string{true: "T"},
		"ma":      map[[2]int]string{{1, 2}: "A"},
		"mst":     map[mkrKey]string{{1}: "ST"},
		"skv":     skv,
		"im_min":  map[interface{}]interface{}{int(math.MinInt64): "MIN"},
		"im_umax": map[interface{}]interface{}{uint64(math.MaxUint64): "UMAX"},
		"im_neg":  map[interface{}]interface{}{-1: "NEG"},
		"im_big":  map[interface{}]interface{}{uint64(1 << 63): "BIG"},
		"huge":    1e300,
		"negf":    -1.0,
		"umax":    uint64(math.MaxUint64),
		"f63":     float64(1 << 63),
	}
	rows := []struct{ name, tpl, want string }{
		// Jinja2: [|None|False]
		{"deviation_ms_none", `[{{ ms[none] }}|{{ ms.get(none) }}|{{ none in ms }}]`, `[N|N|true]`},
		// Jinja2: [|None|False]
		{"deviation_d_none", `[{{ d[none] }}|{{ d.get(none) }}|{{ none in d }}]`, `[N|N|true]`},
		// Jinja2: [|one|False] (1 is not "1")
		{"deviation_ms_int", `[{{ ms[1] }}|{{ ms.get(1) }}|{{ 1 in ms }}]`, `[one|one|true]`},
		{"ifm_none", `[{{ ifm[none] }}|{{ ifm.get(none) }}|{{ none in ifm }}]`, `[nil|nil|true]`},
		{"ifm_float", `[{{ ifm[1.0] }}|{{ ifm.get(1.0) }}|{{ 1.0 in ifm }}]`, `[one|one|true]`},
		// Jinja2: [|None|False]
		{"deviation_yam_none", `[{{ yam[none] }}|{{ yam.get(none) }}|{{ none in yam }}]`, `[N|N|true]`},
		// Jinja2: [|None|False]
		{"deviation_yam_int", `[{{ yam[1] }}|{{ yam.get(1) }}|{{ 1 in yam }}]`, `[one|one|true]`},
		{"ptr_key", `{{ ms[p] }}|{{ p in ms }}`, `one|true`},
		{"ptr_get", `[{{ ms[p] }}|{{ ms.get(p) }}|{{ p in ms }}]`, `[one|one|true]`},
		// Jinja2:
		{"deviation_bytes_key", `{{ ms[bs] }}`, `one`},
		{"mixed_num", `{{ ifm[1.0] }}|{{ ifm.get(1.0) }}|{{ 1.0 in ifm }}`, `one|one|true`},
		{"mixed_miss", `[{{ ifm[2] }}|{{ ifm.get(2) }}|{{ 2 in ifm }}]`, `[||false]`},
		// Jinja2: error: TypeError (cannot use 'list' as a dict key (unhashable type: 'list'))
		{"deviation_slice_key", `[{{ ifm[[1]] }}|{{ ifm.get([1]) }}|{{ [1] in ifm }}]`, `[||false]`},
		{"ptr_dict_literal", `{{ {p: 1}|list }}`, `[&#39;1&#39;]`},
		{"bool_int_key", `{{ ifm1[true] }}|{{ ift[1] }}`, `one|T`},
		{"markup_key", `[{{ hm[mk] }}|{{ hm.get(mk) }}|{{ mk in hm }}|{{ mk is in hm }}|{{ hm['k'] }}|{{ hm.get('k') }}|{{ 'k' in hm }}|{{ 'k' is in hm }}]`, `[MK|MK|true|true|MK|MK|true|true]`},
		{"markup_key_str", `[{{ hs[mk] }}|{{ hs.get(mk) }}|{{ mk in hs }}]`, `[S|S|true]`},
		{"ptr_markup_key", `[{{ hm[pmk] }}|{{ hm.get(pmk) }}|{{ pmk in hm }}]`, `[MK|MK|true]`},
		// Jinja2: [NK|NK|True|NK] (a str subclass key equals 'k')
		{"deviation_named_key", `[{{ hn[sk] }}|{{ hn.get(sk) }}|{{ sk in hn }}|{{ hn['k'] }}]`, `[NK|NK|true|]`},
		{"bool_key", `[{{ mb[true] }}|{{ mb.get(true) }}|{{ true in mb }}|{{ mb[1] }}|{{ mb.get(1) }}|{{ 1 in mb }}|{{ 1 is in mb }}|{{ mb['true'] }}]`, `[T|T|true|T|T|true|true|]`},
		// Jinja2: error: TypeError (cannot use 'list' as a dict key (unhashable type: 'list'))
		{"deviation_array_key", `[{{ ma[[1,2]] }}|{{ ma.get([1,2]) }}|{{ [1,2] in ma }}|{{ [1,2] is in ma }}|{{ ma[[1]] }}]`, `[A|A|true|true|]`},
		{"struct_key", `[{{ mst[skv] }}|{{ mst.get(skv) }}|{{ skv in mst }}|{{ mst.get(1) }}]`, `[ST|ST|true|]`},
		// Jinja2: [|None|False]
		{"deviation_bytes_get", `[{{ ms[bs] }}|{{ ms.get(bs) }}|{{ bs in ms }}]`, `[one|one|true]`},
		{"range_float_int", `[{{ im_min[huge] }}|{{ im_min.get(huge) }}|{{ huge in im_min }}]`, `[||false]`},
		{"range_neg_uint", `[{{ im_umax[-1] }}|{{ im_umax[negf] }}|{{ im_umax.get(-1) }}|{{ -1 in im_umax }}]`, `[|||false]`},
		{"range_uint_int", `[{{ im_neg[umax] }}|{{ im_neg.get(umax) }}|{{ umax in im_neg }}]`, `[||false]`},
		{"big_float_uint", `[{{ im_big[f63] }}|{{ im_big[huge] }}|{{ im_big.get(f63) }}|{{ f63 in im_big }}]`, `[BIG||BIG|true]`},
	}
	env := miya.NewEnvironment()
	for _, tc := range rows {
		t.Run(tc.name, func(t *testing.T) {
			tmpl, err := env.FromString(tc.tpl)
			if err != nil {
				t.Fatal(err)
			}
			got, err := tmpl.Render(miya.NewContextFrom(data))
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestMapLookupCost(t *testing.T) {
	bigi := map[interface{}]interface{}{"s": "S"}
	bigt := map[int]int{}
	for i := 0; i < 20000; i++ {
		bigi[i] = i
		bigt[i] = i
	}
	bigs := map[string]int{}
	bigss := map[string]string{}
	bigst := map[mkrKey]int{}
	for i := 0; i < 20000; i++ {
		bigs[strconv.Itoa(i)] = i
		bigss[strconv.Itoa(i)] = strconv.Itoa(i)
		bigst[mkrKey{i}] = i
	}
	data := map[string]interface{}{"bigi": bigi, "bigt": bigt, "bigs": bigs, "bigss": bigss, "bigst": bigst}
	ops := []struct{ name, expr string }{
		{"miss", `bigX[-1]`},
		{"in", `-1 in bigX`},
		{"get_miss", `bigX.get(-1)`},
		{"get_hit", `bigX.get(5)`},
	}
	env := miya.NewEnvironment()
	run := func(t *testing.T, expr string) {
		src := "{% for i in range(2000) %}{{ " + expr + " }}{% endfor %}"
		tmpl, err := env.FromString(src)
		if err != nil {
			t.Fatal(err)
		}
		done := make(chan error, 1)
		start := time.Now()
		go func() {
			_, err := tmpl.Render(miya.NewContextFrom(data))
			done <- err
		}()
		select {
		case err := <-done:
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("took %v", time.Since(start))
		case <-time.After(300 * time.Millisecond):
			t.Fatal("2000 lookups took over 300ms: the lookup scans the map")
		}
	}
	for _, m := range []string{"bigi", "bigt"} {
		for _, op := range ops {
			t.Run(m+"_"+op.name, func(t *testing.T) {
				run(t, strings.ReplaceAll(op.expr, "bigX", m))
			})
		}
	}
	for _, op := range []struct{ name, expr string }{
		{"bigt_get_wrongtype", `bigt.get('5')`},
		{"bigt_get_none", `bigt.get(none)`},
		{"bigt_in_wrongtype", `'5' in bigt`},
		{"bigt_isin_wrongtype", `'5' is in bigt`},
		{"bigs_get", `bigs.get('5')`},
		{"bigss_get", `bigss.get('5')`},
		{"bigst_get_wrongtype", `bigst.get(1)`},
	} {
		t.Run(op.name, func(t *testing.T) {
			run(t, op.expr)
		})
	}
}
