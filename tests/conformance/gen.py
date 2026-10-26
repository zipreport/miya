#!/usr/bin/env python3
"""Generate cases.json for miya's Jinja2 conformance suite.

Run with Jinja2 3.1.6 and MarkupSafe 3.0.3 (see the README in this directory):
    python3 tests/conformance/gen.py > tests/conformance/cases.json
Each case is rendered by Jinja2; the result (output, or error class plus the
phase it was raised in) becomes the expected value that miya must reproduce.
"""
import importlib.metadata
import json
import re
import sys

import jinja2

assert jinja2.__version__ == "3.1.6", f"need jinja2 3.1.6, have {jinja2.__version__}"
MARKUPSAFE = importlib.metadata.version("markupsafe")
assert MARKUPSAFE == "3.0.3", f"need markupsafe 3.0.3, have {MARKUPSAFE}"


# pass_eval_context stops Jinja2 from folding a constant output such as
# {{ 3 in [1, 2, 3] }} at compile time, which would escape it before finalize
# runs and print "True". With it, every top-level output goes through finalize.
@jinja2.pass_eval_context
def finalize(_eval_ctx, v):
    # miya's deliberate deviations, applied to top-level {{ }} output only:
    # booleans print true/false, None prints empty, whole floats print without ".0".
    if v is True:
        return "true"
    if v is False:
        return "false"
    if v is None:
        return ""
    if isinstance(v, float) and v.is_integer() and abs(v) < 1e16:
        return str(int(v))
    return v


CASES = []


def case(cid, template=None, ctx=None, templates=None, entry="main", **opts):
    if template is not None:
        templates = {"main": template}
    CASES.append({
        "id": cid,
        "templates": templates,
        "entry": entry,
        "context": ctx or {},
        "options": {
            "autoescape": opts.get("autoescape", True),
            "trim_blocks": opts.get("trim_blocks", False),
            "lstrip_blocks": opts.get("lstrip_blocks", False),
            "keep_trailing_newline": opts.get("keep_trailing_newline", False),
            "strict": opts.get("strict", False),
        },
    })


# Context values reach Jinja2 exactly as cases.json stores them (a JSON round
# trip with sorted keys), so a dict iterates in sorted key order in both
# engines' expectations.

# --- basics: expected to pass today (regression net) -----------------------
case("basic-001", "Hello {{ name }}!", {"name": "World"})
case("basic-002", "{{ 1 + 2 }}")
case("basic-003", "{{ 'a' ~ 'b' }}")
case("basic-004", "{% if x %}yes{% else %}no{% endif %}", {"x": 1})
case("basic-005", "{% if x %}yes{% elif y %}maybe{% else %}no{% endif %}", {"x": 0, "y": 1})
case("basic-006", "{% for i in items %}{{ i }},{% endfor %}", {"items": [1, 2, 3]})
case("basic-007", "{% for i in items %}{{ loop.index }}{% endfor %}", {"items": ["a", "b", "c"]})
case("basic-008", "{% for i in items %}{{ loop.first }}{{ loop.last }}|{% endfor %}", {"items": [1, 2]})
case("basic-009", "{% for i in [] %}x{% else %}empty{% endfor %}")
case("basic-010", "{% set x = 5 %}{{ x * 2 }}")
case("basic-011", "{{ name|upper }}", {"name": "ann"})
case("basic-012", "{{ name|lower }}", {"name": "ANN"})
case("basic-013", "{{ items|length }}", {"items": [1, 2, 3]})
case("basic-014", "{{ items|join(', ') }}", {"items": ["a", "b"]})
case("basic-015", "{{ items|first }}-{{ items|last }}", {"items": [1, 2, 3]})
case("basic-016", "{{ u.name }}", {"u": {"name": "Ann"}})
case("basic-017", "{{ u['name'] }}", {"u": {"name": "Ann"}})
case("basic-018", "{{ items[1] }}", {"items": ["a", "b"]})
case("basic-019", "{{ nope }}|")
case("basic-020", "{{ nope|default('d') }}")
case("basic-021", "{{ 10 / 4 }}")
case("basic-022", "{{ 2 ** 3 }}")
case("basic-023", "{{ 7 % 3 }}")
case("basic-024", "{{ 7 // 2 }}")
case("basic-025", "{{ 'abc'|reverse }}")
case("basic-026", "{{ x is defined }}", {"x": 1})
case("basic-027", "{{ y is defined }}")
case("basic-028", "{{ 3 in [1, 2, 3] }}")
case("basic-029", "{{ 'b' in s }}", {"s": "abc"})
case("basic-030", "{{ not x }}", {"x": 0})
case("basic-031", "{{ x if x else 'none' }}", {"x": ""})
case("basic-032", "{{ s|title }}", {"s": "hello world"})
case("basic-033", "{{ s|capitalize }}", {"s": "hello"})
case("basic-034", "{{ s|trim }}", {"s": "  x  "})
case("basic-035", "{{ s|replace('a', 'o') }}", {"s": "banana"})
case("basic-036", "{{ n|abs }}", {"n": -3})
case("basic-037", "{{ items|sum }}", {"items": [1, 2, 3]})
case("basic-038", "{{ s|wordcount }}", {"s": "a b c"})
case("basic-039", "{{ range(3)|list|length }}")
case("basic-040", "{% for i in range(1, 4) %}{{ i }}{% endfor %}")
case("basic-041", "{{ items|batch(2)|list|length }}", {"items": [1, 2, 3]})
case("basic-042", "{% set ns = namespace(c=0) %}{% for i in [1,2,3] %}{% set ns.c = ns.c + i %}{% endfor %}{{ ns.c }}")
case("basic-043", "{{ 'x'|center(5) }}|")
case("basic-044", "{{ items|reverse|join }}", {"items": [1, 2, 3]})
case("basic-045", "{{ s|safe }}", {"s": "<b>"})
case("basic-046", "{{ s }}", {"s": "<b>&\"'"})
case("basic-047", "{% autoescape false %}{{ s }}{% endautoescape %}", {"s": "<b>"})
case("basic-048", "{{ n|string }}", {"n": "a"})
case("basic-049", "{% macro m(a) %}[{{ a }}]{% endmacro %}{{ m('x')|safe }}")
case("basic-050", "{% filter upper %}abc{% endfilter %}")
case("basic-051", templates={"main": "{% include 'p' %}!", "p": "part"})
case("basic-052", templates={"main": "{% extends 'base' %}{% block b %}child{% endblock %}", "base": "[{% block b %}base{% endblock %}]"})
case("basic-053", templates={"main": "{% extends 'base' %}{% block b %}{{ super() }}+c{% endblock %}", "base": "[{% block b %}base{% endblock %}]"})
case("basic-054", "{% with a = 1 %}{{ a }}{% endwith %}")
case("basic-055", "{{ 0.5 + 0.25 }}")
case("basic-056", "{{ a == 1.0 }}", {"a": 1})
case("basic-057", "{{ total > 0 }}", {"total": 0})
case("basic-058", "{{ [1, 2, 3]|max }}-{{ [1, 2, 3]|min }}")
case("basic-059", "{% for k, v in d.items() %}{{ k }}={{ v }};{% endfor %}", {"d": {"a": 1}})
case("basic-060", "{{ s|striptags }}", {"s": "<b>x</b>"})
case("basic-061", "{{ 15 / 3 }}")

# --- and / or (plan 032) ----------------------------------------------------
case("andor-001", "{{ name or 'default' }}", {"name": "Ann"})
case("andor-002", "{{ name or 'default' }}", {"name": ""})
case("andor-003", "{{ '' or 'x' }}")
case("andor-004", "{{ 0 or 1 and 2 }}")
case("andor-005", "{{ a and 'y' }}", {"a": "x"})
case("andor-006", "{{ l and l[0] }}", {"l": []})
case("andor-007", "{% if x is defined and x > 0 %}t{% else %}f{% endif %}")
case("andor-008", "{{ (n or 5) + 1 }}", {"n": 0})
case("andor-009", "{% set v = s or 'dflt' %}{{ v }}", {"s": "hi"})
case("andor-010", "{{ n == 0 or 10 / n }}", {"n": 0})
case("andor-011", "{% if n != 0 and 10 / n > 1 %}a{% else %}b{% endif %}", {"n": 0})
case("andor-012", "{{ none or 'd' }}")

# --- escaping of plain values and filters (plan 031) ------------------------
case("esc-001", "{{ s|e }}", {"s": "<b>&\"q"})
case("esc-002", "{{ s|e|e }}", {"s": "<b>"})
case("esc-003", "{% set y = s|e %}{{ y }}", {"s": "<b>"})
case("esc-004", "{{ s|urlize }}", {"s": "<script>x</script>"})
case("esc-005", "{{ s|urlize }}", {"s": "see http://a.com/<b> now"})
case("esc-006", "<a title='{{ q|tojson }}'>", {"q": "a'b"})
case("esc-007", "{{ q|tojson }}", {"q": "<&>"})
case("esc-008", "{{ ['<'] }}")
case("esc-009", "{{ items }}", {"items": ["<i>"]})
case("esc-010", "<a{{ d|xmlattr }}>", {"d": {"k": "v"}})
case("esc-011", "{{ s|forceescape }}", {"s": "<b>"})
case("esc-012", "{{ s|escape }}", {"s": "&"})
case("esc-013", "{{ s|striptags|e }}", {"s": "a & b"})
case("esc-014", "{{ s }}", {"s": "<b>&"}, autoescape=False)
case("esc-015", "{{ s|e }}", {"s": "<b>&"}, autoescape=False)
case("esc-016", "{{ s|urlize }}", {"s": "<b> http://a.com"}, autoescape=False)
case("esc-017", "{{ q|tojson }}", {"q": "<a'b>"}, autoescape=False)
case("esc-018", "{{ \"a'b\"|tojson }}")
case("esc-019", "<a{{ d|xmlattr }}>", {"d": {"k": "<v\">"}})
case("esc-020", "{{ 'go http://x.com <b>'|urlize }}")

# --- macro / caller / set-block output escaping (plan 027) ------------------
case("macesc-001", "{% macro b(x) %}<b>{{ x }}</b>{% endmacro %}{{ b('a&') }}")
case("macesc-002", "{% macro w() %}<div>{{ caller() }}</div>{% endmacro %}{% call w() %}<p>c</p>{% endcall %}")
case("macesc-003", "{% set s %}<i>{% endset %}{{ s }}")
case("macesc-004", "{% macro b(x) %}<b>{{ x }}</b>{% endmacro %}{{ b('x')|length }}")
case("macesc-005", "{% macro e() %}{% endmacro %}{% if e() %}yes{% else %}no{% endif %}")
case("macesc-006", "{% macro b(x) %}<b>{{ x }}</b>{% endmacro %}{{ b('<s>') }}")
case("macesc-007", templates={"main": "{% import 'm' as m %}{{ m.b('x') }}", "m": "{% macro b(x) %}<b>{{ x }}</b>{% endmacro %}"})

# --- whitespace, comments, raw (plan 033) -----------------------------------
case("ws-001", "{% raw %}{{ a }} {% for x in y %}{% endraw %}")
case("ws-002", "{% raw %}{{ 'str' }}{% endraw %}")
case("ws-003", "{% raw %}<p>{# keep #}</p>{% endraw %}")
case("ws-004", "x = 1;\n{# note #}\ny = 2;", trim_blocks=True)
case("ws-005", "<script>\nvar a = 1\n{# note #}\nvar b = 2\n</script>", trim_blocks=True, lstrip_blocks=True)
case("ws-006", "Total {# amount #} due", trim_blocks=True)
case("ws-007", "{% if true %}{{ '-%}' }}{% endif %}", trim_blocks=True)
case("ws-008", "{{ 'a {# x #} b' }}", trim_blocks=True)
case("ws-009", "a {#- c #} b")
case("ws-010", "a  {{- 1 -}}  b")
case("ws-011", "{{ a }}\n", {"a": 1})
case("ws-012", "{% if true %}\nx\n{% endif %}\ny", trim_blocks=True)
case("ws-013", "  {% if true %}\nx\n  {% endif %}\n", trim_blocks=True, lstrip_blocks=True)
case("ws-014", "<pre>line1\n{# c #}\nline2</pre>", trim_blocks=True)
case("ws-015", "{{ d }}", {"d": "x %}\n\ny"}, trim_blocks=True)
case("ws-016", "<option>  a\n   b </option>")
case("ws-017", "<pre>{{ o }}</pre>", {"o": "<option>  a\n   b </option>"}, autoescape=False)
case("ws-018", "{%- if true -%}\n  x\n{%- endif %}\n  y")
case("ws-019", "a\n{% for i in [1,2] %}\n{{ i }}\n{% endfor %}\nb", trim_blocks=True, lstrip_blocks=True)
case("ws-020", "{{ a }}\n", {"a": 1}, keep_trailing_newline=True)
case("ws-021", "{% if true %}\nx\n{% endif %}\n", trim_blocks=True, keep_trailing_newline=True)

# --- inheritance (plan 034) -------------------------------------------------
case("inh-001", templates={"main": "{% extends 'base' %}{% block x %}X{% endblock %}", "base": "{% if f %}[{% block x %}d{% endblock %}]{% endif %}"}, ctx={"f": True})
case("inh-002", templates={"main": "{% extends 'base' %}{% block x %}X{% endblock %}", "base": "{% for i in [1,2] %}[{% block x %}d{% endblock %}]{% endfor %}"})
case("inh-003", templates={"main": "{% extends 'base' %}{% set title = 'T' %}{% block b %}{{ title }}{% endblock %}", "base": "{{ title }}|{% block b %}{% endblock %}"})
case("inh-004", templates={"main": "{% extends 'base' %}{% macro m() %}M{% endmacro %}{% block b %}{{ m() }}{% endblock %}", "base": "[{% block b %}{% endblock %}]"})
case("inh-005", templates={"main": "{% extends 'base' %}{% block x %}X{% endblock %}", "base": "{% with a = 1 %}[{% block x %}d{% endblock %}]{% endwith %}"})
case("inh-006", templates={"main": "{% extends 'mid' %}{% block b %}C{% endblock %}", "mid": "{% extends 'base' %}{% block b %}M{% endblock %}", "base": "[{% block b %}B{% endblock %}]"})
case("inh-007", templates={"main": "{% extends 'mid' %}{% block b %}{{ super() }}C{% endblock %}", "mid": "{% extends 'base' %}{% block b %}{{ super() }}M{% endblock %}", "base": "[{% block b %}B{% endblock %}]"})
case("inh-008", templates={"main": "{% extends 'base' %}{% block x %}X{% endblock %}", "base": "{% if f %}[{% block x %}d{% endblock %}]{% else %}none{% endif %}"}, ctx={"f": False})
case("inh-009", templates={"main": "{% extends 'base' %}{% block t %}T{% endblock %}", "base": "{% block t %}{% endblock %}|{{ self.t() }}"})
case("inh-101", templates={"main": "{% extends 'base' %}{% block b %}{% for i in [1,2,3] if i > 1 %}{{ i }}{% endfor %}{% endblock %}", "base": "[{% block b %}{% endblock %}]"})
case("inh-102", templates={"main": "{% extends 'base' %}{% block b %}{% for i in [1,2,3] if i > 5 %}a{% else %}e{% endfor %}{% endblock %}", "base": "[{% block b %}{% endblock %}]"})
case("inh-103", templates={"main": "{% extends 'base' %}{% block b %}{% for n in tree recursive %}{{ loop.depth }}{% if n.c %}({{ loop(n.c) }}){% endif %}{% endfor %}{% endblock %}", "base": "[{% block b %}{% endblock %}]"}, ctx={"tree": [{"c": [{"c": [{"c": []}]}]}]})

# --- imports (plan 026) -----------------------------------------------------
case("imp-001", templates={"main": "{% import 'lib' as l %}{{ l.show() }}", "lib": "{% macro show() %}[{{ v }}]{% endmacro %}"}, ctx={"v": "SECRET"})
case("imp-002", templates={"main": "{% for v in ['A','B'] %}{% import 'lib' as l %}{{ l.m(v) }}|{% endfor %}", "lib": "{% macro m(a) %}{{ a }}:{{ v }}{% endmacro %}"})
case("imp-003", templates={"main": "{% from 'lib' import hi %}{{ hi('x') }}", "lib": "{% macro hi(n) %}hi {{ n }}{% endmacro %}"})
case("imp-004", templates={"main": "{% import 'lib' as l %}{{ l.g }}", "lib": "{% set g = 'G' %}"})
case("imp-005", templates={"main": "{% import 'lib' as l %}{{ l.m() }}", "lib": "{% set g = 'G' %}{% macro m() %}{{ g }}{% endmacro %}"})
case("imp-006", templates={"main": "{% import 'nope' as l %}x"})
case("imp-007", templates={"main": "{% from 'lib' import nothere %}{{ nothere() }}", "lib": "{% macro hi() %}{% endmacro %}"})
case("imp-008", templates={"main": "{% import 'lib' as l with context %}{{ l.show() }}", "lib": "{% macro show() %}[{{ v }}]{% endmacro %}"}, ctx={"v": "V"})

# --- macro calls (plan 028) -------------------------------------------------
case("mac-001", "{% macro m(a, b=2) %}{{ a }}-{{ b }}{% endmacro %}{{ m(a=1) }}")
case("mac-002", "{% macro m(a, b=2) %}{{ a }}-{{ b }}{% endmacro %}{{ m(b=5, a=1) }}")
case("mac-003", "{% macro m(a, b=2) %}{{ a }}-{{ b }}{% endmacro %}{{ m(1, b=9) }}")
case("mac-004", "{% macro m(a, b=a~'x') %}{{ a }}{{ b }}{% endmacro %}{{ m(1) }}")
case("mac-005", "{% macro m(a, b) %}{% if b %}B{% else %}nob{% endif %}{% endmacro %}{{ m(1) }}")
case("mac-006", "{% macro m(a) %}{{ a }}{% endmacro %}{{ m(1, 2) }}")
case("mac-007", templates={"main": "{% import 'lib' as l %}{{ l.r(3) }}", "lib": "{% macro r(n) %}{{ n }}{% if n > 0 %}{{ r(n - 1) }}{% endif %}{% endmacro %}"})
case("mac-008", templates={"main": "{% import 'lib' as l %}{{ l.a() }}", "lib": "{% macro b() %}B{% endmacro %}{% macro a() %}A{{ b() }}{% endmacro %}"})
case("mac-009", templates={"main": "{% import 'lib' as l %}{% call l.wrap() %}in{% endcall %}", "lib": "{% macro wrap() %}[{{ caller() }}]{% endmacro %}"}, autoescape=False)
case("mac-010", templates={"main": "{% import 'lib' as l %}{{ l.hi(g='G', n='N') }}", "lib": "{% macro hi(n, g='Hello') %}{{ g }} {{ n }}{% endmacro %}"})
case("mac-011", "{% macro m() %}{{ varargs|length }}{% endmacro %}{{ m(1, 2) }}")
case("mac-012", "{% macro m(a) %}[{{ a }}]{% endmacro %}{% call(v) m(2) %}{% endcall %}x")

# --- filter keyword arguments (plan 036) ------------------------------------
case("fkw-001", "{{ l|sort(reverse=true)|join(',') }}", {"l": [3, 1, 2]})
case("fkw-002", "{{ ld|map(attribute='n')|join(',') }}", {"ld": [{"n": "a"}, {"n": "b"}]})
case("fkw-003", "{{ ld|sort(attribute='v')|map(attribute='v')|join(',') }}", {"ld": [{"v": 2}, {"v": 1}]})
case("fkw-004", "{{ ld|sum(attribute='v') }}", {"ld": [{"v": 2}, {"v": 3}]})
case("fkw-005", "{{ l|sum(start=10) }}", {"l": [1, 2]})
case("fkw-006", "{% for g in ld|groupby(attribute='g') %}{{ g.grouper }}:{{ g.list|length }};{% endfor %}", {"ld": [{"g": "x"}, {"g": "x"}, {"g": "y"}]})
case("fkw-007", "{{ 3.14159|round(precision=1) }}")
case("fkw-008", "{{ s|indent(2, first=true) }}", {"s": "a\nb"})
case("fkw-009", "{{ d|dictsort(reverse=true)|map('first')|join(',') }}", {"d": {"a": 1, "b": 2}})
case("fkw-010", "{{ l|join(d='-') }}", {"l": [1, 2]})
case("fkw-011", "{{ s|truncate(length=5) }}", {"s": "Hello World Foo"})
case("fkw-012", "{{ s|replace('a', 'o', count=1) }}", {"s": "aaa"})

# --- filter semantics (plan 039) --------------------------------------------
case("fsem-001", "{{ 'Hello World'|truncate(8) }}")
case("fsem-002", "{{ 'Hello World'|truncate(5) }}")
case("fsem-003", "{{ s|truncate(30, true) }}", {"s": "The quick brown fox jumps over the lazy dog and keeps running"})
case("fsem-004", "{{ s|indent(2) }}", {"s": "a\n\nb"})
case("fsem-005", "{{ x|round }}", {"x": 2.5})
case("fsem-006", "{{ 2.5|round(2) }}")
case("fsem-007", "{{ 2.675|round(2) }}")
case("fsem-008", "{{ l|slice(2)|list|length }}", {"l": [3, 1, 2]})
case("fsem-009", "{{ l|map('upper')|join(',') }}", {"l": ["a", "b"]})
case("fsem-010", "{{ l|select('odd')|join(',') }}", {"l": [1, 2, 3]})
case("fsem-011", "{{ l|reject('odd')|join(',') }}", {"l": [1, 2, 3]})
case("fsem-012", "{{ l|select('gt', 1)|join(',') }}", {"l": [1, 2, 3]})
case("fsem-013", "{{ l|unique|join(',') }}", {"l": ["a", "B", "b", "A"]})
case("fsem-014", "{{ '3.7'|int }}")
case("fsem-015", "{{ ' 7 '|int }}")
case("fsem-016", "{{ s|urlencode }}", {"s": "a b/c"})
case("fsem-017", "{{ s|striptags }}", {"s": "<p>a</p>  <p>b</p>"})
case("fsem-018", "{{ l|sort|join(',') }}", {"l": [10, 9, 2, 33]})
case("fsem-019", "{{ l|max }}", {"l": ["3", "10", "2"]})
case("fsem-020", "{{ 1000|filesizeformat }}")
case("fsem-021", "{{ s|wordwrap(3) }}", {"s": "abcdefgh ij"})
case("fsem-022", "{{ s|center }}|", {"s": "x"})
case("fsem-023", "{{ ld|selectattr('a')|list|length }}", {"ld": [{"a": 1}, {"a": 0}]})
case("fsem-024", "{{ ld|selectattr('a', 'equalto', 1)|list|length }}", {"ld": [{"a": 1}, {"a": 2}]})
case("fsem-025", "{{ ld|rejectattr('a')|list|length }}", {"ld": [{"a": 1}, {"a": 0}]})
case("fsem-026", "{{ l|list }}", {"l": [1, 2]})
case("fsem-027", "{{ d|dictsort|first|first }}", {"d": {"b": 1, "a": 2}})
case("fsem-028", "{{ 5|round(2) }}")
case("fsem-029", "{{ ls|max }}", {"ls": ["apple", "cherry", "banana"]})
case("fsem-030", "{{ '%d'|format(n) }}", {"n": 10023})
case("fsem-031", "{{ '%s-%s'|format(1, 2) }}")

# --- dicts and unicode strings (plan 035) -----------------------------------
case("dict-001", "{% for k in d %}{{ k }},{% endfor %}", {"d": {"a": 10}})
case("dict-002", "{% for k in d %}{{ k }}={{ d[k] }};{% endfor %}", {"d": {"a": 1}})
case("dict-003", "{% for k, v in d.items() %}{{ k }}{% endfor %}", {"d": {"b": 1, "a": 2, "c": 3}})
case("dict-004", "{{ d.keys()|list|join }}", {"d": {"b": 1, "a": 2}})
case("uni-001", "{{ s[1] }}", {"s": "açúcar"})
case("uni-002", "{{ s[0:3] }}", {"s": "açúcar"})
case("uni-003", "{{ s[::-1] }}", {"s": "açúcar"})
case("uni-004", "{% for c in s %}{{ c }}.{% endfor %}", {"s": "açú"})
case("uni-005", "{{ s|last }}", {"s": "señor ñ"})
case("uni-006", "{{ s|length }}", {"s": "açúcar"})
case("uni-007", "{{ s|upper }}", {"s": "ação"})

# --- dict literals (plan 046) ------------------------------------------------
case("dlit-001", "{{ {'a': 1}|length }}")
case("dlit-002", "{% set d = {'a': 1, 'b': 2} %}{{ d.a }}")
case("dlit-003", "{% for k in {'x': 1} %}{{ k }}{% endfor %}")
case("dlit-004", "{{ {'a': {'b': 2}}.a.b }}")
case("dlit-005", "{{ {'a': 1, 'b': 2}.b }}")
case("dlit-006", "{{ {1: 'one'}[1] }}")
case("dlit-007", "{{ {'a': 1, 'a': 2}.a }}")

# --- bracket balancing in tags (plan 051) -----------------------------------
case("lexb-001", "{{ [{'a': {'b': 1}}][0].a.b }}")
case("lexb-002", "{{ (1 }}")
case("lexb-003", "{{ x }<div>{{ y }}")
case("lexb-004", "{% set x = 1) %}x")

# --- undefined / None (plan 049) --------------------------------------------
case("undef-001", "{{ a|default('x') }}", strict=True)
case("undef-002", "{{ a is undefined }}", strict=True)
case("undef-003", "{{ a }}", strict=True)
case("undef-004", "{{ a|length }}")
case("undef-005", "{{ a|join(',') }}|")
case("undef-006", "{{ n is defined }}", {"n": None})
case("undef-007", "{{ n is none }}", {"n": None})
case("undef-008", "{{ o.zz|default('x') }}", {"o": {}}, strict=True)
case("undef-009", "{% if a is defined %}y{% else %}n{% endif %}", strict=True)

# --- loops and arithmetic (plan 037); with and include (plan 049) ---------
case("loop-001", "{% for i in [1,2,3] if i > 5 %}a{% else %}e{% endfor %}")
case("loop-002", "{% for i in [1,2,3] if i > 1 %}{{ loop.index }}{% endfor %}")
case("loop-003", "{% for a in [[1],[2]] %}{% for b in a %}{{ loop.depth }}{% endfor %}{% endfor %}")
case("loop-004", "{% for n in tree recursive %}{{ loop.depth }}{% if n.c %}({{ loop(n.c) }}){% endif %}{% endfor %}", {"tree": [{"c": [{"c": []}]}]})
case("loop-005", "{% set x = 1 %}{% for i in [1,2] %}{% set x = x + 1 %}{{ x }}{% endfor %}{{ x }}")
case("loop-006", "{% for i in [1,2,3] %}{% if i == 1 %}{% set t = 'a' %}{% endif %}[{{ t }}]{% endfor %}")
case("loop-007", "{% for i in [1,2,3] %}{{ loop.revindex }}{{ loop.length }}{% endfor %}")
case("loop-008", "{% for i in [1,2,3] %}{{ loop.cycle('a','b') }}{% endfor %}")
case("with-001", "{% set a = 10 %}{% with a = 1, b = a %}{{ b }}{% endwith %}")
case("math-001", "{{ -7 // 2 }}")
case("math-002", "{{ -7 % 3 }}")
case("math-003", "{{ 7.5 % 2 }}")
case("math-004", "{{ x // 2 }}", {"x": 5.5})
case("math-005", "{{ 7 % -3 }}")
case("math-006", "{{ x // 0.3 }}", {"x": 10})
case("math-007", "{{ 1 % 0 }}")
case("inc-001", templates={"main": "a{% include 'bad' ignore missing %}b", "bad": "{{ 1 // 0 }}"})
case("inc-002", templates={"main": "{% include 'p' %}[{{ z }}]", "p": "{% set z = 1 %}"})
case("inc-003", templates={"main": "a{% include 'nope' ignore missing %}b"})

# --- parser (plan 038) ------------------------------------------------------
case("parse-001", "{{ n is divisibleby 3 }}", {"n": 6})
case("parse-002", "{{ a is eq 1 }}", {"a": 1})
case("parse-003", "{{ n is gt 2 }}", {"n": 3})
case("parse-004", "{{ n is in [1, 2] }}", {"n": 1})
case("parse-005", "{{ x is true }}", {"x": True})
case("parse-006", "{{ x is false }}", {"x": False})
case("parse-007", "{{ msg.from }}", {"msg": {"from": "me"}})
case("parse-008", "{{ d.if }}", {"d": {"if": 1}})
case("parse-009", "{{ 1 < x < 3 }}", {"x": 2})
case("parse-010", "{{ 1 == a == 1 }}", {"a": 1})
case("parse-011", "{{ -2 ** 2 }}")
case("parse-012", "{{ 2 ** 3 ** 2 }}")
case("parse-013", "{{ -i|abs }}", {"i": 5})
case("parse-014", "{{ (1, 2)|length }}")
case("parse-015", "{{ '%s-%s' % (1, 2) }}")
case("parse-016", "{% set a, b = 1, 2 %}{{ a }}{{ b }}")
case("parse-017", "{{ 0x10 }}")
case("parse-018", "{{ 1_000 }}")
case("parse-019", "{{ 1e3 + x }}", {"x": 0})
case("parse-020", "{{ 'a' 'b' }}")
case("parse-021", "{{ b.0 }}", {"b": [1, 2]})
case("parse-022", "{{ f is filter }}", {"f": "upper"})
case("parse-023", "a {%+ if true %}b{% endif %}", lstrip_blocks=True)
case("parse-024", "{{ x if x }}|", {"x": 0})
case("parse-025", "{{ 'abc'.upper() }}")
case("parse-026", "{{ d.get('a') }}", {"d": {"a": 1}})
case("parse-027", "{{ 'a' * 3 }}")
case("parse-028", "{{ --i }}", {"i": 5})
case("parse-029", "{{ missing }}|", {"missing": "m"})

# --- error classes: miya must fail where Jinja2 fails -----------------------
case("err-001", "{% if %}")
case("err-002", "{{ }}")
case("err-003", "{% for %}")
case("err-004", "{% endif %}")
case("err-005", "{{ 1 / 0 }}")
case("err-006", "{% block a %}{% block a %}{% endblock %}{% endblock %}")
case("err-007", templates={"main": "{% extends 'nope' %}"})
case("err-008", templates={"main": "{% include 'nope' %}"})

# --- bad data and malformed tags: errors, not panics (plan 030) -------------
case("nopanic-001", "{{ n + 1 }}", {"n": None})
case("nopanic-002", "{{ l|batch(2)|list }}", {"l": [1, 2, 3]})
case("nopanic-003", "{{ s|indent(-1, true) }}", {"s": "hello world"})
case("nopanic-004", "{% extends 1 %}")


# --- templates cannot modify data (plan 040) --------------------------------
case("mut-001", "{% set d = dict() %}{% set d.k = 1 %}{{ d.k }}")
case("mut-002", "{% set n = namespace(a=1) %}{% set n.a = 2 %}{{ n.a }}")

def render(c):
    env = jinja2.Environment(
        loader=jinja2.DictLoader(c["templates"]),
        autoescape=c["options"]["autoescape"],
        trim_blocks=c["options"]["trim_blocks"],
        lstrip_blocks=c["options"]["lstrip_blocks"],
        keep_trailing_newline=c["options"]["keep_trailing_newline"],
        undefined=jinja2.StrictUndefined if c["options"]["strict"] else jinja2.Undefined,
        finalize=finalize,
    )
    ctx = json.loads(json.dumps(c["context"], sort_keys=True))
    try:
        tmpl = env.get_template(c["entry"])
    except Exception as e:  # noqa: BLE001 - the error class is the expectation
        return {"error": type(e).__name__, "phase": "compile"}
    try:
        return {"output": tmpl.render(**ctx)}
    except Exception as e:  # noqa: BLE001
        return {"error": type(e).__name__, "phase": "render"}


ids = [c["id"] for c in CASES]
assert len(ids) == len(set(ids)), "duplicate case id"
for c in CASES:
    c["expect"] = render(c)
    out = c["expect"].get("output", "")
    assert not re.fullmatch(r"True|False|None|-?\d+\.0", out), f"{c['id']}: constant-folded output {out!r}"
json.dump({"jinja2": jinja2.__version__, "markupsafe": MARKUPSAFE, "cases": CASES}, sys.stdout, indent=1, ensure_ascii=False, sort_keys=True)
sys.stdout.write("\n")
