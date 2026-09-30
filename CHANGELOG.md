# Changelog

All notable changes are documented in this file.

## [v4.5.1] - 2026-10-01

Docs only. The changelog is compressed to plain prose, 209 lines to 85. No code changes, coverage stays at 100.0%.

## [v4.5] - 2026-10-01

Logic-path and test-suite audit. Every fix landed test-first, statement coverage stays at 100.0%, no schema changed.

### Fixed

Data loss in the suite writers. Prompt writes deleted every prompt row and reinserted with fresh ids, so `ON DELETE CASCADE` wiped the suite's scores and model responses on any add, edit, reorder, import, or profile rename. Profile writes severed every prompt link the same way, and model rename deleted that model's saved responses. All three writers now upsert in place and keep row ids, so the cascade only fires for rows actually removed.

Suite resolution. `GetCurrentSuiteID` recursed forever on a suites table with no current and no default row, and its recovery could race into two current rows. Recovery is now one transaction that leaves a single current row, and name resolution goes through it so both paths agree. Foreign keys ran on one pooled connection only, the DSN now carries `_foreign_keys=on` for every connection.

Endpoint contracts. The evaluate GET branch accepted any prompt index, so a negative index rendered prompt 0's response under a "Prompt -4 of 2" header and non-numeric input fell back to 0. Invalid input now redirects, the model lookup is suite-scoped, and other methods get 405. `UpdateResult` accepted GET and wrote cell 0 on parse errors, it is POST-only with 400s and no writes on rejection.

Stats. The tier bucket sum overwrote real totals, hiding arbitrary 0-100 scores, and prompt-count query errors fell back to a hardcoded 50 that skewed tiers. Totals stay true now and the errors return 500.

Smaller fixes. Renaming a missing model returns 404 instead of creating a phantom row. A filtered prompt list posts a partial order that was silently dropped with a success redirect, it now gets 400 with data untouched. Equal-total models sort alphabetically on the results page, mock page, and websocket broadcast. Zero-prompt suites broadcast 0 instead of a NaN that failed the whole payload. `WriteResults` logs dropped scores. The listen port comes from `-port` with validation instead of a hardcoded :8080.

### Security

WebSocket upgrades require a same-origin Origin header. Empty Origin still works for CLI clients.

### Changed

Test repairs. Log-only error tests assert exact statuses and table counts, silent-pass guards fail loudly, zero-assertion tests verify routing and data readback, and duplicate twins were dropped for stronger successors. A parity test pins the JS score constants to the Go source. The 4 oversized test files are split by feature, every result under the 1000 SLOC limit, verified move-only by function inventory. The dead testutil DB half, 1135 lines, is deleted.

## [v4.4] - 2026-10-01

### Fixed

Evaluate resolved prompt ids against a hardcoded `suite_id` of 1, so responses saved to and loaded from the wrong suite outside the first one. Stats counted prompts from every suite when computing tier thresholds. The pinned marked CDN URL named a file that does not exist in marked 18 and the markdown preview broke, marked 18.0.14 and Chart.js 4.5.1 are vendored under `templates/vendor/` with a test that forbids CDN URLs. Five nav links had duplicate `class` attributes so `text-xs` never applied. `/refresh_results` was a byte copy of `/confirm_refresh_results` and is gone. Copy buttons read prompt text from page data instead of an inline `onclick` string.

### Added

`middleware.CheckOrigin` rejects cross-site POSTs. `/templates` and `/assets` serve css, js, and images only. `make coverage-enforce` fails below 100.0% total coverage. Every page declares `html lang` and the viewport meta, stats has an empty state, the results badge offers a reconnect after WebSocket retries, and Evaluate submit stays disabled until a score is picked. One constants file plus the Go `ScoreColors` map replace 3 conflicting score palettes.

### Changed

Prompt and solution render through the sanitized server markdown pipeline instead of marked into `innerHTML`, 6 duplicated score button blocks collapsed into one loop, and emoji buttons carry aria-labels. 35 `console.log` calls and dead helpers left results.html. Debug scripts and untracked personal tooling are gone or gitignored, the dead tailwind config and deps are deleted, `output.css` is rebuilt as the canonical v4 build, and npm deps are pinned exact. README claims match reality, the design docs are marked historical, and `handlers/results.go` is split under the 1000 SLOC limit with `EvaluateResult` as a `*Handler` method.

## [v4.3] - 2026-09-30

### Removed

Breaking: the entire auto-judge surface. The `evaluator/` package, the `python_service/` FastAPI judge, `handlers/evaluation.go`, the settings page and its store, AES-256-GCM encrypted storage for API keys, 4 tables, and all judge routes and broadcasts. The app is manual-only now: CRUD for prompts, models, profiles, and suites, manual scoring, saved responses, results, and stats. No third-party APIs, fully offline.

### Added

Every package tests at 100.0% statement coverage, error paths included.

### Changed

Toolchain refresh: Go 1.27.1, go-sqlite3 v1.14.52, golangci-lint v2.14.0, Tailwind v4.3.3, DaisyUI v5.7.47, Node.js 24, GitHub Actions pinned to current majors. CI dropped the screenshots job, and the commit-updates job only refreshes the coverage badge and table with scoped write permissions.

### Fixed

`WriteResults` error paths roll back their transaction instead of leaking it, which held the SQLite file on Windows. `capture.mjs` was missing an `await` on parallel page operations.

## [v3.4] - 2025-12-20

README gained a per-package coverage table, entrypoints were refactored for testability with more error-path unit tests, and `SetCurrentSuite` updates `CurrentSuite` when no error hook is set.

## [v3.3] - 2025-12-20

Removed dead Makefile targets for the legacy JSON to SQLite tooling. `AUTOMATED_EVALUATION_SETUP.md` matches the README commands and style.

## [v3.2] - 2025-12-20

The Chart.js container got a stable height so stacked bars render. The left rail got narrower and the score controls are centered.

## [v3.1] - 2025-12-19

Playwright screenshot automation (`npm run screenshots`) and Arena CSS regression tests. Layout compacted: thinner rail, flex rows for sticky headers and toolbars, styled selects and file inputs, scroll buttons moved into the sidebar on shell pages.

## [v3.0] - 2025-12-19

Added optional automated LLM evaluation with a Python FastAPI judge service, multi-judge consensus, job persistence, WebSocket progress, cost tracking, encrypted storage for API keys, the Arena UI overhaul, and coverage badge scripts. Templates standardized on a shared top bar and left rail, with handler dependency injection for broader testing. Removed the v2.0 JSON migration tooling, its CLI flags, and Gemini CLI integration.

## [v2.1] - 2025-03-16

Dynamic profile grouping with color-coded borders, centralized score colors in `score-utils.js`, results table row highlighting, sticky headers, tooltips, and a progress bar, keyboard scoring in the evaluate grid, tiered mock score generation, and WebSocket auto-reconnect. Fixed prompt move limits that kept profile groups contiguous, profile border and cell sizing bugs, model deletion in `WriteResults`, and profile case sensitivity. Grouping logic moved to `middleware/utils.go` and legacy JSON data files were deleted.

Full details live in the git history.
