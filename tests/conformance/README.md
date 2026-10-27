# Jinja2 conformance suite

`cases.json` holds templates rendered by Jinja2 3.1.6 (MarkupSafe 3.0.3).
`TestConformance` renders each one with miya 200 times and compares the output.
For cases where Jinja2 raises, miya must also fail, in the same phase
(compile = `GetTemplate`, render = `Render`); the error class is not compared.

## Deliberate deviations

miya prints booleans as `true`/`false`, `None` as empty, and whole floats
without `.0`. `gen.py` applies these through a Jinja2 `finalize` hook, so they
affect top-level `{{ }}` output only (`{{ [1, 2] }}` still expects `[1, 2]`).
`finalize` carries `@jinja2.pass_eval_context` because otherwise Jinja2 folds a
context-free output such as `{{ 3 in [1, 2, 3] }}` into a constant and escapes
it before `finalize` runs, printing `True`. `gen.py` asserts no expected output
is `True`, `False`, `None` or `N.0`. Contexts are passed to Jinja2 after a JSON
round trip with sorted keys, so dicts iterate in sorted key order.

## Adding cases

Edit `gen.py`, then regenerate with the pinned versions:

    python3 -m venv /tmp/j2 && /tmp/j2/bin/pip install jinja2==3.1.6 markupsafe==3.0.3
    /tmp/j2/bin/python3 tests/conformance/gen.py > tests/conformance/cases.json

Run `go test ./tests/conformance/ -run '^TestConformance$' -count=1`. A new case
miya gets wrong is either fixed or added by hand to `known_divergences.json`
with `plan`, `kind` and `reason`. `-args -conformance.dump=FILE` writes all
current divergences (the test fails as usual, but still writes the file).
`build_allowlist.py --existing known_divergences.json FILE` refreshes the
allow-list from a dump: entries that now match are dropped, the rest keep their
`plan` and take `kind` and `reason` from the dump. Any divergence not already
in the allow-list makes it exit 1 and list the IDs; `--allow-new` adds them
with plan `unplanned`, and is only for deliberate divergences. Write the output
to a temporary file and move it over `known_divergences.json`.

## Contract for fix plans

A change that makes an allow-listed case match must delete that entry from
`known_divergences.json` in the same change; the test fails until it does.
The `plan` field is a routing hint, not a rule: whichever change makes the case
match deletes it. A change that alters how an allow-listed case diverges (its
`kind`: `output`, `error`, `panic`) must update that entry's `kind` and
`reason` in the same change and list it in its report; the `reason` text alone
is not compared. Entries are never added to hide a regression.

## Why mapLoader

`mapLoader` implements only `GetSource`, the path zipreport-go's loader takes.
`loader.StringLoader` also compiles through `Environment` and honours
`trim_blocks`, `lstrip_blocks` and `{{-`/`-}}` (since plan 044), so either would
exercise whitespace options; the suite keeps `mapLoader` so the plain
`GetSource` path stays covered.

## Whitespace matrix

`gen_whitespace.py` generates `whitespace_cases.json` (every whitespace option
against every tag kind) and `TestWhitespaceMatrix` checks it. It has no
allow-list. Regenerate it with:

    REPO=$(pwd); (cd / && python3 -P "$REPO/tests/conformance/gen_whitespace.py") > tests/conformance/whitespace_cases.json
