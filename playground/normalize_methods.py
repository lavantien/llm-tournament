"""One-off: replace string method literals with http.Method* constants.

Only rewrites exact comparisons and case labels on r.Method in non-test Go
files under handlers/. Usage: python playground/normalize_methods.py
"""
import glob
import re

MAP = {
    '"GET"': "http.MethodGet",
    '"POST"': "http.MethodPost",
}
changed = 0
for path in glob.glob("handlers/*.go"):
    if path.endswith("_test.go"):
        continue
    with open(path, encoding="utf-8") as f:
        src = f.read()
    orig = src
    for lit, const in MAP.items():
        src = re.sub(rf"r\.Method == {re.escape(lit)}\b", f"r.Method == {const}", src)
        src = re.sub(rf"r\.Method != {re.escape(lit)}\b", f"r.Method != {const}", src)
        src = re.sub(rf"case {re.escape(lit)}:", f"case {const}:", src)
    if src != orig:
        with open(path, "w", encoding="utf-8", newline="") as f:
            f.write(src)
        changed += 1
        print(f"normalized {path}")
print(f"done: {changed} files")
