"""One-off: remove whole console.log(...) statements from a JS-in-HTML file.

Balanced-paren scan from each `console.log(` occurrence; deletes the statement
and its trailing newline/indent. Leaves console.error/warn untouched.
Usage: python playground/strip_console_log.py <file>
"""
import sys


def strip_console_logs(src: str) -> str:
    out = []
    i = 0
    needle = "console.log("
    while i < len(src):
        idx = src.find(needle, i)
        if idx == -1:
            out.append(src[i:])
            break
        # keep any leading whitespace on the line out of the output
        line_start = src.rfind("\n", i, idx) + 1
        prefix = src[line_start:idx]
        if prefix.strip():
            out.append(src[i:idx])
            i = idx
        else:
            out.append(src[i:line_start])
        # find balanced closing paren
        depth = 0
        j = idx + len(needle) - 1  # at the '('
        in_str = None
        while j < len(src):
            c = src[j]
            if in_str:
                if c == in_str:
                    if src[j - 1] != "\\":
                        in_str = None
            elif c in ("'", '"', "`"):
                in_str = c
            elif c == "(":
                depth += 1
            elif c == ")":
                depth -= 1
                if depth == 0:
                    break
            j += 1
        # skip trailing semicolon if present
        k = j + 1
        while k < len(src) and src[k] in " \t":
            k += 1
        if k < len(src) and src[k] == ";":
            k += 1
        # skip the newline (and following blank line if the statement was alone)
        if k < len(src) and src[k] == "\r":
            k += 1
        if k < len(src) and src[k] == "\n":
            k += 1
            if k < len(src) and src[k] == "\r" and k + 1 < len(src) and src[k + 1] == "\n":
                k += 2
        i = k
    return "".join(out)


if __name__ == "__main__":
    path = sys.argv[1]
    with open(path, encoding="utf-8") as f:
        content = f.read()
    stripped = strip_console_logs(content)
    with open(path, "w", encoding="utf-8", newline="") as f:
        f.write(stripped)
    print(f"removed {content.count('console.log(')} console.log statements")
