#!/usr/bin/env python3
"""Refresh known_divergences.json from a -conformance.dump file.

    python3 tests/conformance/build_allowlist.py --existing tests/conformance/known_divergences.json DUMP.json > /tmp/kd.json \
        && mv /tmp/kd.json tests/conformance/known_divergences.json

Entries whose case no longer diverges are dropped. Entries still diverging keep
their plan and take kind and reason from the dump. A divergence that is not
already allow-listed is a regression: the script lists it and exits 1 unless
--allow-new is given, in which case it is added with plan "unplanned".
"""
import argparse
import json
import sys

ap = argparse.ArgumentParser()
ap.add_argument("--existing", required=True, help="current known_divergences.json")
ap.add_argument("--allow-new", action="store_true", help="add divergences not in --existing")
ap.add_argument("dump", help="file written by -conformance.dump")
args = ap.parse_args()

existing = json.load(open(args.existing))
dump = json.load(open(args.dump))

new = sorted(cid for cid in dump if cid not in existing)
if new and not args.allow_new:
    sys.stderr.write("not in %s (fix them, or pass --allow-new for deliberate divergences):\n" % args.existing)
    for cid in new:
        sys.stderr.write("  %s: %s\n" % (cid, dump[cid]["reason"]))
    sys.exit(1)

for cid in sorted(existing):
    if cid not in dump:
        sys.stderr.write("dropped %s: now matches Jinja2\n" % cid)
    elif dump[cid]["kind"] != existing[cid]["kind"]:
        sys.stderr.write("kind changed %s: %s -> %s\n" % (cid, existing[cid]["kind"], dump[cid]["kind"]))

out = {}
for cid in sorted(dump):
    plan = existing[cid]["plan"] if cid in existing else "unplanned"
    out[cid] = {"plan": plan, "kind": dump[cid]["kind"], "reason": dump[cid]["reason"]}
json.dump(out, sys.stdout, indent=1, ensure_ascii=False, sort_keys=True)
sys.stdout.write("\n")
