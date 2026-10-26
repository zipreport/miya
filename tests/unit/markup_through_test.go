package miya_test

import (
	"testing"

	miya "github.com/zipreport/miya"
)

type Name string

func TestMarkupThrough(t *testing.T) {
	const x = "<b>x</b>"
	rows := []struct {
		name, tpl string
		off       bool
		v         string
		want      string
	}{
		{"M1", `{{ [s]|map('upper')|join }}`, false, "", `<B>`},
		{"M2", `{{ [s, v]|map('upper')|join }}`, false, "", `<B>&lt;B&gt;`},
		{"M3", `{{ [s]|map('trim')|join }}`, false, "", `<b>`},
		{"M4", `{{ [s]|map('replace', 'b', 'i')|join }}`, false, "", `<i>`},
		{"M5", `{{ [s]|map('string')|join }}`, false, "", `<b>`},
		{"M6", `{{ [s]|map('format')|join }}`, false, "", `<b>`},
		{"M7", `{{ [s]|select|join }}`, false, "", `<b>`},
		{"D1", `{{ nope|default(s) }}`, false, "", `<b>`},
		{"D2", `{{ none|default(s, true) }}`, false, "", `<b>`},
		{"D3", `{{ nope|default(s)|length }}`, false, "", `3`},
		{"C1", `{% for i in [1,2] %}{{ loop.cycle(s, 'y') }}{% endfor %}`, false, "", `<b>y`},
		{"C2", `{% for i in [1,2] %}{{ loop.cycle(v, s) }}{% endfor %}`, false, "", `&lt;b&gt;<b>`},
		{"C3", `{{ cycler(s, 'y').next() }}`, false, "", `<b>`},
		{"J1", `{{ ['a','b']|join(s) }}`, false, "", `a<b>b`},
		{"J2", `{{ [s, s]|join(d=s) }}`, false, "", `<b><b><b>`},
		{"J3", `{% set items = [{'n': s}, {'n': v}] %}{{ items|join(',', attribute='n') }}`, false, "", `<b>,&lt;b&gt;`},
		{"J4", `{% set items = [{'n': s}, {}] %}{{ items|join(',', attribute='n') }}`, false, "", `<b>,`},
		{"J5", `{% set items = [{'n': s}, {'n': 1}] %}{{ items|join(s, attribute='n') }}`, false, "", `<b><b>1`},
		{"A1", `{% set items = [{'n': s}, {}] %}{{ items|map(attribute='n', default=s)|join }}`, false, "", `<b><b>`},
		{"R1", `{{ v|replace('b', s) }}`, false, "", `&lt;<b>&gt;`},
		{"R2", `{{ v|replace(s, 'x') }}`, false, "", `&lt;b&gt;`},
		{"R3", `{{ [s, v]|map('replace', 'b', s)|join }}`, false, "", `<<b>>&lt;<b>&gt;`},
		{"R4", `{{ v|replace('b', s) }}|{{ [s]|map('upper')|join }}|{{ nope|default(s) }}`, true, "", `<<b>>|<B>|<b>`},
		{"R5", `{{ v|replace('b', new=s) }}`, false, "", `&lt;<b>&gt;`},
		{"R6", `{{ v|replace('<', s) }}`, false, "", `&lt;b&gt;`},
		{"N1", `{{ n + s }}|{{ s + n }}`, false, "", `&lt;t&gt;<b>|<b>&lt;t&gt;`},
		{"N2", `{{ n ~ s }}|{{ [n, s]|join }}`, false, "", `&lt;t&gt;<b>|&lt;t&gt;<b>`},
		{"T1", `{{ t + s }}|{{ s + t }}`, false, "", `&lt;t&gt;<b>|<b>&lt;t&gt;`},
		{"P2", `{{ s * 2 }}|{{ 2 * s }}`, false, x, `<b>x</b><b>x</b>|<b>x</b><b>x</b>`},
		{"M8", `{{ s|map('upper')|join }}`, false, "", `&lt;B&gt;`},
		{"P3", `{{ s|capitalize }}|{{ s|upper }}`, false, x, `<b>x</b>|<B>X</B>`},
		{"Q9_replace_number", `{{ 1|replace('1', s) }}`, false, "", `<b>`},
		{"Q9_map_join_lists", `{{ [[s, v]]|map('join')|join }}`, false, "", `<b>&lt;b&gt;`},
		{"Q9_map_replace_numbers", `{{ [1, v]|map('replace','1',s)|join }}`, false, "", `<b>&lt;b&gt;`},
		{"Q9_replace_rounded", `{{ (2.567|round(1))|replace('.', s) }}`, false, "", `2<b>6`},
		{"Q9_map_rounded", `{{ [2.567|round(1)]|map('replace','.',s)|join }}`, false, "", `2<b>6`},
		{"deviation_bool_str", `{{ true|replace('e', s) }}`, false, "", `tru<b>`},
		{"grouptuple_attr_on", `{{ [{'k':1},{'k':2}]|groupby('k')|map('grouper')|join(',') }}`, false, "", `1,2`},
		{"grouptuple_attr_off", `{{ [{'k':1},{'k':2}]|groupby('k')|map('grouper')|join(',') }}`, true, "", `1,2`},
		{"cycle_none", `{{ [1]|map('string')|list }}{% for i in [1] %}{{ loop.cycle() }}{% endfor %}`, false, "", `ERR:cycle() requires`},
		{"deviation_markup_repr", `{{ [s]|map('upper')|list }}`, false, "", `[&#39;&lt;B&gt;&#39;]`},
		{"deviation_default_none", `{{ [none]|map('default', s)|join }}`, false, "", `<b>`},
		{"deviation_title", `{{ s|title }}`, false, x, `<B>x</b>`},
	}
	for _, tc := range rows {
		t.Run(tc.name, func(t *testing.T) {
			env := miya.NewEnvironment(miya.WithAutoEscape(!tc.off))
			v := tc.v
			if v == "" {
				v = "<b>"
			}
			tmpl, err := env.FromString(`{% set s = v|safe %}` + tc.tpl)
			var got string
			if err == nil {
				got, err = tmpl.Render(miya.NewContextFrom(map[string]interface{}{"v": v, "t": "<t>", "n": Name("<t>")}))
			}
			if len(tc.want) > 4 && tc.want[:4] == "ERR:" {
				if err == nil {
					t.Fatalf("want error, got %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}
