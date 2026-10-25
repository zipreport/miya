#!/usr/bin/env python3
"""Turn a -conformance.dump file into known_divergences.json.

    python3 tests/conformance/build_allowlist.py DUMP.json > tests/conformance/known_divergences.json
The plan number is a routing hint: whichever change makes a case match must
delete its entry, whatever plan it names. Where a case needs two plans, it is
routed to the one that lands last.
"""
import json
import sys

PLAN_BY_PREFIX = {
    "andor": "032", "esc": "031", "macesc": "027", "ws": "033", "inh": "034",
    "imp": "026", "mac": "028", "fkw": "036", "fsem": "050", "dict": "035",
    "uni": "035", "dlit": "046", "undef": "049", "loop": "037", "math": "037",
    "with": "049", "inc": "049", "parse": "038",
}
PLAN_BY_ID = {
    # |e / |escape double-escaping: plan 031 keeps |e's type, 027 makes it Markup.
    "esc-001": "027", "esc-002": "027", "esc-003": "027", "esc-012": "027", "esc-013": "027",
    # urlize output also needs rel="noopener" (054, after 031's escaping).
    "esc-005": "054", "esc-016": "054",
    # Python repr of containers, value methods, string repetition.
    "esc-008": "048", "esc-009": "048", "fsem-026": "048",
    "parse-025": "048", "parse-026": "048", "parse-027": "048",
    # collection filters
    "fsem-008": "039", "fsem-009": "039", "fsem-010": "039", "fsem-011": "039",
    "fsem-012": "039", "fsem-013": "039", "fsem-018": "039", "fsem-019": "039",
    "fsem-024": "039", "fsem-029": "039", "fkw-006": "039", "fkw-009": "039",
    "fkw-011": "050",
    # round keeps miya's fixed decimals and half-away rounding (050's deliberate deviation).
    "fsem-005": "deviation", "fsem-006": "deviation", "fsem-007": "deviation", "fsem-028": "deviation",
    # string filters split out of 050: wordwrap 052, format 053, urlencode/striptags 054.
    "fsem-021": "052", "fsem-030": "053", "fsem-031": "053", "fsem-016": "054", "fsem-017": "054",
    "dlit-003": "035",
    "parse-023": "033",
    "err-006": "034",
    # deferred by the plans that would own them
    "inh-009": "unplanned",    # self.blockname(), deferred by 034
    "parse-015": "unplanned",  # % string formatting, deferred by 038
}

dump = json.load(open(sys.argv[1]))
out = {}
for cid in sorted(dump):
    plan = PLAN_BY_ID.get(cid) or PLAN_BY_PREFIX.get(cid.split("-")[0], "unplanned")
    out[cid] = {"plan": plan, "kind": dump[cid]["kind"], "reason": dump[cid]["reason"]}
json.dump(out, sys.stdout, indent=1, ensure_ascii=False, sort_keys=True)
sys.stdout.write("\n")
