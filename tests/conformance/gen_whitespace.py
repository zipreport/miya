#!/usr/bin/env python3
"""Generate whitespace_cases.json: every whitespace option against every tag kind.

Run with Jinja2 3.1.6 (see the README in this directory):
    REPO=$(pwd); (cd / && python3 -P "$REPO/tests/conformance/gen_whitespace.py") > tests/conformance/whitespace_cases.json
Each template places one construct, indented by two spaces, between optional
text before it and optional text after it. Jinja2's result (output or error
class) is the expected value that miya must reproduce exactly.
"""
import json
import sys

import jinja2

assert jinja2.__version__ == "3.1.6", f"need jinja2 3.1.6, have {jinja2.__version__}"

OPTIONS = {
    "none": {},
    "trim": {"trim_blocks": True},
    "lstrip": {"lstrip_blocks": True},
    "both": {"trim_blocks": True, "lstrip_blocks": True},
    "keep": {"trim_blocks": True, "lstrip_blocks": True, "keep_trailing_newline": True},
}
# L is the marker after the opening delimiter, R the marker before the closing one.
CONSTRUCTS = {
    "block": "{%L if v R%}\n  X\n  {%L endif R%}",
    "comment": "{#L c R#}",
    "var": "{{L v R}}",
    "raw": "{%L raw R%}\n  {{ v }} {# c #}\n  {%L endraw R%}",
}
MARKERS = {"n": "", "d": "-", "p": "+"}
BEFORE = {"first": "", "line": "a\n", "same": "a "}
AFTER = {"last": "", "nl": "\n", "text": "\nb"}
EOLS = {"lf": "\n", "crlf": "\r\n"}


def render(source, opts):
    env = jinja2.Environment(**opts)
    try:
        return {"output": env.from_string(source).render(v="V")}
    except Exception as e:  # noqa: BLE001 - the error class is the expectation
        return {"error": type(e).__name__}


cases = []
for oname, opts in OPTIONS.items():
    for kind, construct in CONSTRUCTS.items():
        for lname, left in MARKERS.items():
            for rname, right in MARKERS.items():
                for bname, before in BEFORE.items():
                    for aname, after in AFTER.items():
                        for ename, eol in EOLS.items():
                            body = construct.replace("L", left).replace("R", right)
                            source = (before + "  " + body + after).replace("\n", eol)
                            cases.append({
                                "id": f"wsm-{oname}-{kind}-{lname}{rname}-{bname}-{aname}-{ename}",
                                "template": source,
                                "options": {
                                    "trim_blocks": opts.get("trim_blocks", False),
                                    "lstrip_blocks": opts.get("lstrip_blocks", False),
                                    "keep_trailing_newline": opts.get("keep_trailing_newline", False),
                                },
                                "expect": render(source, opts),
                            })
json.dump({"jinja2": jinja2.__version__, "cases": cases}, sys.stdout, indent=1, ensure_ascii=False, sort_keys=True)
sys.stdout.write("\n")
