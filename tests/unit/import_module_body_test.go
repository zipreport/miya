package miya_test

import (
	"strings"
	"testing"
)

func TestImportModuleBody(t *testing.T) {
	mods := map[string]string{
		"m1":  `{% set a, b = 1, 2 %}`,
		"m2":  `{% set ns = namespace(x=0) %}{% set ns.x = 3 %}`,
		"m3":  `{% set ns = namespace(c=0) %}{% for i in range(3) %}{% set ns.c = ns.c + 1 %}{% endfor %}`,
		"m4":  `{% set l = [] %}{% do l.append(1) %}`,
		"m5":  `{% if true %}{% set x = 1 %}{% else %}{% set x = 2 %}{% endif %}`,
		"m6":  `{% for i in [1] %}{% set y = 1 %}{% endfor %}`,
		"m7":  `{% set items = [] %}{% macro add(x) %}{% do items.append(x) %}{% endmacro %}`,
		"m8":  `{% set _p = 1 %}{% set q = 2 %}`,
		"m9":  `hello{% set z = 1 %}`,
		"m10": `{% if true %}{% macro mm() %}A{% endmacro %}{% else %}{% macro mm() %}B{% endmacro %}{% endif %}`,
		"m11": `{% set w = v * 2 %}`,
		"m12": `{% macro _h() %}H{% endmacro %}{% macro p() %}{{ _h() }}{% endmacro %}`,
		"m13": `{% set x = 1 %}{% block b %}{% set x = 2 %}{% macro bm() %}M{% endmacro %}{% endblock %}{% macro g() %}{{ x }}{% endmacro %}`,
		"m14": `{% extends 'mb' %}{% block x %}{{ super() }}{% endblock %}{% macro s() %}S{% endmacro %}{% set ev = 2 %}`,
		"mb":  `B{% block x %}bx{% endblock %}{% macro bm() %}BM{% endmacro %}{% set bv = 1 %}`,
		"m15": `{% do acc.append(1) %}`,
		"m16": `{% set h = 1 %}{% import 'm1' as h %}`,
		"m17": `{% for i in [1] %}{% macro fm() %}F{% endmacro %}{% endfor %}`,
		"mi":  `{% if false %}{% set x = 1 %}{% elif true %}{% set x = 3 %}{% else %}{% set x = 4 %}{% endif %}{% if false %}{% set y = 1 %}{% else %}{% set y = 2 %}{% endif %}`,
		"mf":  `{% set x = 0 %}{% from "mi" import x %}{% set z = 1 %}`,
	}
	cases := []struct {
		name, src, want, errHas string
	}{
		{"m1", `{% import 'm1' as m %}{{ m.a }}{{ m.b }}|{% from 'm1' import a, b %}{{ a }}{{ b }}`, `12|12`, ""},
		{"m2", `{% import 'm2' as m %}{{ m.ns.x }}`, `3`, ""},
		{"m3", `{% import 'm3' as m %}{{ m.ns.c }}`, `3`, ""},
		{"m4", `{% import 'm4' as m %}{{ m.l }}`, `[1]`, ""},
		{"m5", `{% import 'm5' as m %}{{ m.x }}`, `1`, ""},
		{"m6", `{% import 'm6' as m %}{{ m.y is defined }}`, `false`, ""},
		{"m7", `{% import 'm7' as m %}{{ m.add(1) }}{{ m.add(2) }}{{ m.items }}`, `[1, 2]`, ""},
		{"m8", `{% import 'm8' as m %}{{ m._p is defined }}{{ m.q }}`, `false2`, ""},
		{"m9", `[{% import 'm9' as m %}]{{ m.z }}`, `[]1`, ""},
		{"m10", `{% import 'm10' as m %}{{ m.mm() }}`, `A`, ""},
		{"m11", `{% import 'm11' as m with context %}{{ m.w }}`, `8`, ""},
		{"from_private", `{% from 'm8' import _p %}`, ``, "underline"},
		{"m12", `{% import 'm12' as m %}{{ m.p() }}|{{ m._h is defined }}`, `H|false`, ""},
		{"m13", `{% import 'm13' as m %}{{ m.x }}|{{ m.g() }}|{{ m.bm is defined }}`, `1|1|false`, ""},
		{"m14", `{% import 'm14' as m %}{{ m.s() }}|{{ m.ev is defined }}`, `S|true`, ""},
		// Jinja2 runs the parent's top level too: True|True
		{"deviation_import_extends_parent", `{% import 'm14' as m %}{{ m.bm is defined }}|{{ m.bv is defined }}`, `false|false`, ""},
		{"m15", `{% set acc = [] %}{% for i in [1,2] %}{% import 'm15' as m with context %}{% endfor %}{{ acc }}`, `[1, 1]`, ""},
		{"m16", `{% import 'm16' as m %}{{ m.h is defined }}`, `false`, ""},
		{"m17", `{% import 'm17' as m %}{{ m.fm is defined }}`, `false`, ""},
		{"m18_elif", `{% import 'mi' as m %}{{ m.x }}|{{ m.y }}`, `3|2`, ""},
		{"m19_from_not_exported", `{% import 'mf' as m %}{{ m.x is defined }}|{{ m.z }}`, `false|1`, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			tm := map[string]string{"i.html": c.src}
			for k, v := range mods {
				tm[k] = v
			}
			env := newImportEnv(tm, false)
			got, err := renderImport(env, map[string]interface{}{"v": 4})
			if c.errHas != "" {
				if err == nil || !strings.Contains(err.Error(), c.errHas) {
					t.Fatalf("want error containing %q, got %q, %v", c.errHas, got, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != c.want {
				t.Errorf("got %q, want %q", got, c.want)
			}
		})
	}
}
