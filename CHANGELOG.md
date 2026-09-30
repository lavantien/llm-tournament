# Changelog

All notable changes are documented in this file.

## [v4.6] - 2026-10-01

Documentation harmonization and repo hygiene, no behavior change. The stale v4.0 design planning docs (`DESIGN_CONCEPT.md`, `DESIGN_ROLLOUT.md`) are deleted with every reference cleaned up, and the legacy aider make targets plus the `dev.sh` and `dev.ps1` launchers are gone. The vestigial `data/current_suite.txt` is removed because the current suite lives in SQLite, the arena theme test only accepts `output.css` now that `arena.css` does not exist, and the stray `nul` and `$coverageFile` artifacts left the repo root along with their gitignore masks.

README fixes: the mermaid diagram and the API list name the real `/update_result` endpoint, the listen port is documented as the `-port` flag wired in `run.go`, the stats anchor points at `main.go:43`, and the vendored Chart.js and Marked versions are pinned to 4.5.1 and 18.0.14. Every doc follows the writing standards now: sentence case headings, no em dashes or semicolons in prose, digits for numbers. The coverage note in README and `update_coverage_table.py` drops the enforced-versus-aspirational contrast, and the readme enforcement tests pin the sentence-cased section headers.

## [v4.5.1] - 2026-10-01

Docs only. The changelog was rewritten in condensed prose, the versions were sorted, and the release descriptions follow the same style. No code changed and coverage stayed at 100.0%.

## [v4.5] - 2026-10-01

A logic-path and test-suite audit. Every fix landed test-first, coverage stayed at 100.0%, and no schema changed.

The suite writers destroyed live data because they deleted every row and reinserted with fresh ids, which tripped `ON DELETE CASCADE`. Any prompt write wiped the suite's scores and model responses, any profile write severed every prompt link, and model rename deleted that model's saved responses. All 3 writers now upsert in place and keep row ids, so the cascade only fires for removed rows.

Suite resolution could hang or land on the wrong suite. `GetCurrentSuiteID` recursed forever on a suites table with no current and no default row, and its recovery could race into two current rows, so recovery is now one transaction that leaves a single current row with name resolution going through it. Foreign keys ran on one pooled connection only, so cascades were nondeterministic under load, and the DSN now carries `_foreign_keys=on` for every connection.

Endpoint contracts were tightened because they accepted anything. The evaluate GET branch took any prompt index, so a negative index rendered prompt 0's response under a "Prompt -4 of 2" header and non-numeric input fell back to 0. Invalid input now redirects, the model lookup is suite-scoped, and other methods get 405. `UpdateResult` accepted GET and wrote cell 0 on parse errors, it is POST-only with 400s and no writes on rejection. Stats overwrote real totals with the tier bucket sum, hiding arbitrary 0-100 scores, and prompt-count query errors fell back to a hardcoded 50 that skewed tiers. Totals stay true now and the errors return 500.

The rest: renaming a missing model returns 404 instead of creating a phantom row, a filtered prompt list posting a partial order gets 400 with data untouched instead of a silent success redirect, equal-total models sort alphabetically on the results page, mock page, and websocket broadcast so renders stop shuffling, zero-prompt suites broadcast 0 instead of a NaN that failed the whole payload, `WriteResults` logs dropped scores, and the listen port comes from `-port` with validation instead of a hardcoded :8080. WebSocket upgrades require a same-origin Origin header because the upgrader accepted any origin, and empty Origin still works for CLI clients.

The test suite was repaired because 100.0% statement coverage hid tests that could not fail. Log-only error tests assert exact statuses and table counts, silent-pass guards fail loudly, zero-assertion tests verify routing and data readback, duplicate twins were dropped for stronger successors, and a parity test pins the JS score constants to the Go source. The 4 oversized test files are split by feature under the 1000 SLOC limit, verified move-only by function inventory, and the dead testutil DB half, 1135 lines, is deleted.

## [v4.4] - 2026-10-01

Evaluate resolved prompt ids against a hardcoded `suite_id` of 1, so responses saved to and loaded from the wrong suite outside the first one, and stats counted prompts from every suite when computing tier thresholds. The pinned marked CDN URL named a file that does not exist in marked 18 and the markdown preview broke, so marked 18.0.14 and Chart.js 4.5.1 are vendored under `templates/vendor/` with a test that forbids CDN URLs. Five nav links had duplicate `class` attributes so `text-xs` never applied, `/refresh_results` was a byte copy of `/confirm_refresh_results` and is gone, and copy buttons read prompt text from page data instead of an inline `onclick` string that invited injection.

`middleware.CheckOrigin` rejects cross-site POSTs, `/templates` and `/assets` serve css, js, and images only so Go sources are no longer downloadable, and `make coverage-enforce` fails below 100.0% total coverage. Every page declares `html lang` and the viewport meta, stats has an empty state, the results badge offers a reconnect after WebSocket retries, Evaluate submit stays disabled until a score is picked, and one constants file plus the Go `ScoreColors` map replace 3 conflicting score palettes.

Prompt and solution render through the sanitized server markdown pipeline instead of marked into `innerHTML`, 6 duplicated score button blocks collapsed into one loop, and emoji buttons carry aria-labels. 35 `console.log` calls and dead helpers left results.html, debug scripts and untracked personal tooling are gone or gitignored, the dead tailwind config and deps are deleted, `output.css` is rebuilt as the canonical v4 build, npm deps are pinned exact, README claims match reality, and `handlers/results.go` is split under the 1000 SLOC limit with `EvaluateResult` as a `*Handler` method.

## [v4.3] - 2026-09-30

Breaking: the entire auto-judge surface was removed because the app was going manual-only. The `evaluator/` package, the `python_service/` FastAPI judge, `handlers/evaluation.go`, the settings page and its store, AES-256-GCM encrypted storage for API keys, 4 tables, and all judge routes and broadcasts are gone. What remains is CRUD for prompts, models, profiles, and suites, manual scoring, saved responses, results, and stats, with no third-party APIs and fully offline. Every package tests at 100.0% statement coverage, error paths included.

The toolchain was refreshed to Go 1.27.1, go-sqlite3 v1.14.52, golangci-lint v2.14.0, Tailwind v4.3.3, DaisyUI v5.7.47, Node.js 24, and GitHub Actions pinned to current majors. CI dropped the screenshots job, and the commit-updates job only refreshes the coverage badge and table with scoped write permissions. `WriteResults` error paths roll back their transaction instead of leaking it, which held the SQLite file on Windows, and `capture.mjs` gained a missing `await` on parallel page operations.

## [v4.2] - 2026-01-01

The UI moved to Tailwind with DaisyUI and no custom CSS. Added a Randomize Scores button, profile-based mock prompt generation, confirmation dialogs, and scroll buttons on pages that lacked them. CI began updating the coverage badge, table, and screenshots automatically, parsing statement-level coverage after the badge had been reading function-level numbers. Screenshot capture gained the edit page, and error-case tests grew across results, evaluate, and mock handlers.

## [v4.1] - 2025-12-23

CI was added, all lint findings were fixed, and formatting was standardized. Coverage moved from 100.0% to 99.9%, an accepted balance at the time.

## [v4.0] - 2025-12-20

Total statement coverage reached 100.0%.

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

## [v2.0] - 2025-03-15

Persistence moved from JSON files to SQLite. `middleware/database.go` arrived with go-sqlite3, profiles, prompts, and results state went through database queries, and the flags `--migrate-to-sqlite`, `--remigrate-scores`, and `--cleanup-duplicates` plus Makefile `setenv`, `migrate`, and `dedup` targets handled the transition. Tier definitions gained a transcendental tier with matching CSS, and state and database modules were reorganized with better transaction handling.

## [v1.7] - 2025-03-10

Scoring moved to an 11-tier system with names like Divine, Legendary, and Mythical over a 0-3000+ range, with reversed colors for hierarchy. Mock generation got even distribution, revised weights, and proper RNG seeding. phi-4-mini joined the roster and the random scores button stopped flickering and desyncing.

## [v1.6] - 2025-03-07

README gained screenshots, score buttons became emoji, results gained a Previous button to restore prior state, and evaluate rendered prompt and solution as markdown with a raw markdown copy button and previous and next navigation. Templates gained math functions. A batch of fixes covered results table rendering, websocket stability including first-load rendering, and the JSON parse error in random mock scores.

## [v1.5] - 2025-03-02

Added tools (openwebui, pipes, anthropic-claude-thinking-96k) and a tools section in the README, fixed the chart overflow, polished the UI, and updated screenshots.

## [v1.4] - 2025-03-01

The codebase was modularized. Scoring got granular schemes with matching cell colors, stats got a tiered ranking page, mock scores became persistent random with live updates, import and export switched from CSV to JSON, prompt suite rename was fixed, and websockets stabilized. Prompt and contestant counts settled at 20, with aider configs for o1 high, o3-mini high, v3, r1, 3.7 sonnet, and codestral.

## [v1.3] - 2025-01-19

The prompts page gained a profile filter. Added a full set of XML system prompts, a local TTS tool on Kokoro 82M and ONNX, contestant and default prompt lists at 33, and prompt and README quality work.

## [v1.2] - 2025-01-19

Project structure simplified, profile renames now reflect in prompt rendering and selection, and full text search on the profile page handles XML.

## [v1.1] - 2025-01-18

Copy button on profiles, contestant list at 32, default prompt list at 32.

## [v1.0] - 2025-01-15

First release. The manual tournament workflow with a 30-model contestant list, a 30-prompt suite, and profiles for chain-of-thought plus ReAct and Vietnamese translation.

Full details live in the git history.
