"""One-off: mechanically split handlers/results.go by line ranges.

Move-only. Usage: python playground/split_results_go.py
"""
from pathlib import Path

src_path = Path("handlers/results.go")
lines = src_path.read_text(encoding="utf-8").splitlines(keepends=True)


def take(a, b):
    """1-based inclusive range."""
    return "".join(lines[a - 1 : b])


HEADER_EVALUATE = """package handlers

import (
\t"llm-tournament/middleware"
\t"llm-tournament/templates"
\t"log"
\t"net/http"
\t"strconv"
)

"""

HEADER_MOCK = """package handlers

import (
\t"database/sql"
\t"encoding/json"
\t"io"
\t"llm-tournament/middleware"
\t"log"
\t"math/rand"
\t"net/http"
\t"time"
)

"""

HEADER_IO = """package handlers

import (
\t"encoding/json"
\t"log"
\t"net/http"
)

"""

evaluate = HEADER_EVALUATE + take(72, 76) + "\n" + take(391, 532)
mock = HEADER_MOCK + take(18, 38) + "\n" + take(82, 86) + "\n" + take(557, 1213)
io_file = HEADER_IO + take(77, 81) + "\n" + take(534, 555)

keep = take(1, 17) + take(39, 70) + take(87, 390)

Path("handlers/evaluate.go").write_text(evaluate, encoding="utf-8", newline="")
Path("handlers/mock_results.go").write_text(mock, encoding="utf-8", newline="")
Path("handlers/results_io.go").write_text(io_file, encoding="utf-8", newline="")
src_path.write_text(keep, encoding="utf-8", newline="")
print("split done")
