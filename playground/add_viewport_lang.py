"""One-off: add lang attribute and viewport meta to every page template.

Usage: python playground/add_viewport_lang.py
"""
import glob
import re

VIEWPORT = '<meta name="viewport" content="width=device-width, initial-scale=1" />'

changed = []
for path in glob.glob("templates/*.html"):
    name = path.replace("\\", "/").split("/")[-1]
    if name in ("design_preview.html", "nav.html"):
        continue
    with open(path, encoding="utf-8") as f:
        src = f.read()
    orig = src
    if "<html lang=" not in src:
        src = src.replace("<html ", '<html lang="en" ', 1)
    if 'name="viewport"' not in src:
        src = src.replace("<head>", "<head>\n    " + VIEWPORT, 1)
    if src != orig:
        with open(path, "w", encoding="utf-8", newline="") as f:
            f.write(src)
        changed.append(name)

print(f"updated {len(changed)} templates: {', '.join(sorted(changed))}")
