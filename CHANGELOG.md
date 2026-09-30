# Changelog

All notable changes are documented in this file.

## [v4.5] - 2026-10-01

Logic-path and test-suite audit. Every fix landed test-first, statement coverage stays at 100.0%, and no schema changed.

### Fixed

- **Prompt writes destroyed all scores and saved responses:** the prompt suite writer deleted every prompt row and reinserted with fresh ids, so `ON DELETE CASCADE` wiped the suite's scores and model responses on any add, edit, delete, reorder, import, or profile-rename write. It now upserts in place: surviving rows keep their ids, new rows insert, and the cascade only fires for prompts actually removed.
- **Profile writes severed every prompt link:** the profile suite writer used the same delete-plus-reinsert pattern with AUTOINCREMENT ids, so adding or deleting any one profile unlinked all prompts. It now preserves row ids the same way.
- **Model rename deleted saved responses:** renaming deleted the old model row and inserted a fresh one, cascading away its responses. A new single-UPDATE rename keeps the row id.
- **Profile rename severed links:** the prompt rewrite now resolves against the renamed profile by persisting profiles before the rewritten prompts, with the snapshot taken while the old rows still resolve.
- **Suite resolution recursion and races:** `GetCurrentSuiteID` recursed forever on a suites table with no current and no default row; recovery is now a single transaction that clears other current flags, and name resolution routes through the same recovery so both paths agree.
- **Foreign keys only enforced on one connection:** the `PRAGMA foreign_keys` exec applied to a single pooled connection, so cascades were nondeterministic under load. The DSN now carries `_foreign_keys=on` for every connection.
- **Evaluate page input handling:** the GET branch rejected nothing, so a negative index rendered prompt 0's response under a "Prompt -4 of 2" header and non-numeric indexes silently fell back to 0. Invalid indexes now redirect, the model lookup is suite-scoped, and non-GET/POST methods get 405.
- **UpdateResult accepted anything:** GET requests mutated scores, form-parse and index errors silently wrote cell 0. The endpoint is POST-only with strict validation (400s) and no writes on rejection.
- **Stats totals and errors:** arbitrary 0-100 scores were overwritten by the tier bucket sum, hiding real totals; prompt-count query errors fell back to a hardcoded 50, silently skewing tiers. Totals now stay true and errors return 500.
- **Phantom models and silent reorders:** renaming a missing model created an empty result row (now 404); a filtered prompt list posted a partial order that was silently dropped with a success redirect (now 400 with data untouched).
- **Nondeterministic ranking:** equal-total models shuffled between renders; the results page, mock page, and websocket broadcast now tie-break alphabetically.
- **Empty-suite broadcast failure:** zero-prompt suites produced NaN pass percentages that failed the whole JSON broadcast; they now report 0. `WriteResults` logs dropped scores beyond the prompt count.
- **Dead port config:** the listen port was hardcoded to :8080 with `Config.Port` unused; `-port` now configures it with validation that rejects 0, negatives, and out-of-range values before the DB opens.

### Security

- **WebSocket origin enforcement:** upgrades now require a same-origin Origin header, with empty Origin still allowed for CLI clients.

### Changed

- **Test suite effectiveness:** log-only error tests assert exact statuses and table counts, silent-pass guards fail loudly, zero-assertion tests verify routing and data readback, duplicate test twins were dropped after confirming stronger successors, and a Go/JS parity test pins the score constants bidirectionally.
- **Test file splits:** the four oversized test files (results, prompt, state, profiles) are split by feature, every result under the 1000 SLOC limit, verified move-only by function inventory.
- **Dead scaffolding removed:** the unused testutil DB half (1135 lines) is deleted; only the consumed MockRenderer remains.

## [v4.4] - 2026-10-01

### Fixed

- **Model responses in non-default suites:** the Evaluate page resolved prompt IDs against hardcoded `suite_id = 1`, so saving or loading responses in any other suite silently targeted the first suite's prompts. It now resolves the current suite.
- **Stats MaxScore across suites:** the stats page counted prompts from every suite when computing tier thresholds; it now scopes to the current suite.
- **Broken markdown preview:** the pinned CDN URL for marked pointed at `marked.min.js`, which does not exist in the marked 18 package. Marked 18.0.14 and Chart.js 4.5.1 are now vendored under `templates/vendor/`, so the app runs fully offline (enforced by a test that forbids CDN URLs in templates).
- **Nav link styling:** five nav links carried duplicate `class` attributes; browsers drop the second one, so `text-xs` never applied.
- **Duplicate refresh handler:** `/refresh_results` was a byte-identical copy of `/confirm_refresh_results`; the confirm page now posts to the surviving route and the duplicate is gone.
- **Copy-prompt injection:** prompt text was interpolated into an inline `onclick` JS string; copy buttons now read from the page's JSON data via a delegated listener.

### Added

- **Cross-origin guard:** `middleware.CheckOrigin` rejects browser-driven cross-site POSTs against the local server.
- **Static file allowlist:** `/templates` and `/assets` only serve css/js/image extensions; Go sources and templates are no longer downloadable.
- **Coverage gate:** `make coverage-enforce` (part of `make check` and CI) fails when total statement coverage drops below 100.0%.
- **UI essentials:** every page declares `html lang` and the responsive viewport meta (test-enforced), the stats page has an empty state, the results connection badge starts neutral with a Reconnect button after max WebSocket retries, the Evaluate submit stays disabled until a score is picked, and the save-response button shows pending state.
- **Score palette single-sourced:** one constants file plus the Go `ScoreColors` map and `scoreColor` funcmap replace three conflicting copies with divergent labels.

### Changed

- **Evaluate page hardening:** prompt and solution render through the sanitized server markdown pipeline instead of re-parsing escaped text through marked into `innerHTML`; six duplicated score button blocks collapsed into one loop; emoji buttons carry aria-labels.
- **JS cleanup:** 35 `console.log` calls removed from results.html, dead helpers and duplicated scroll functions deleted, score-utils.js and utils.js no longer define overlapping globals.
- **Repo hygiene:** removed one-off debug scripts (`check.js`, `check_spacers.js`) and untracked personal tooling (`tools/tts`, `tools/bg_batch_eraser`, `tools/openwebui`, `tools/ragweb_agent`, root system prompt XMLs; files remain on disk, now gitignored). Deleted dead `tailwind.config.js` and the unused `@tailwindcss/forms` and `autoprefixer` deps (Tailwind v4 needs neither; `templates/output.css` was regenerated as the canonical v4 build, dropping obsolete vendor prefixes a previous hook's prettier pass had also mangled), pinned all npm deps to exact versions, added `.prettierignore` for vendored and generated files.
- **Docs honesty:** README no longer claims zero custom CSS, one theme, or CDN-backed operation; troubleshooting reflects the CSS-first Tailwind v4 setup; DESIGN_CONCEPT/DESIGN_ROLLOUT marked historical; handlers split (see below).
- **handlers/results.go split** into results, evaluate, mock, and import/export files to stay under the 1000 SLOC limit, and `EvaluateResult` is now a `*Handler` method like every other handler.

## [v4.3] - 2026-09-30

### Removed

- **Automated evaluation and third-party APIs (BREAKING):** Removed the entire auto-judge surface: the `evaluator/` Go package, the `python_service/` FastAPI judge service, `handlers/evaluation.go`, the settings page and its store, AES-256-GCM API-key encryption, the `evaluation_jobs`/`evaluation_history`/`cost_tracking`/`settings` tables, and all judge/evaluation routes and WebSocket broadcasts. The app is now manual-only: CRUD for prompts, models, profiles, and suites plus manual scoring, saved model responses, results, and stats. No third-party APIs, no API keys, fully offline.

### Added

- **100% statement coverage:** Every package now tests at 100.0%, including error paths for suite seeding, score writes, and prompt lookups.

### Changed

- **Sept 2026 toolchain refresh:** Go 1.27.1, go-sqlite3 v1.14.52, golangci-lint v2.14.0, Tailwind CSS v4.3.3, DaisyUI v5.7.47, PostCSS CLI 12, Node.js 24, and all GitHub Actions pinned to current majors (checkout v7, setup-go v7, cache v6, upload-artifact v7, codecov v7). CDN scripts (Marked, Chart.js) pinned to exact versions.
- **CI slimming:** Dropped the screenshots job (regenerate offline with `npm run screenshots` before pushing UI changes). The `commit-updates` job now only refreshes the coverage badge and table, and write permissions are scoped to that job alone.

### Fixed

- **`WriteResults` transaction leak:** Error paths now roll back the open transaction instead of leaking it (which held the SQLite file on Windows and broke temp-dir cleanup).
- **Screenshot capture:** `tools/screenshots/capture.mjs` was missing an `await` on parallel page operations.

## [v3.4] - 2025-12-20

### Added

- **README coverage breakdown:** Added a package-level coverage table for quick visibility into what’s tested.

### Changed

- **Testability + coverage:** Refactored entrypoints (including the demo screenshot server) to be more test-friendly and expanded unit tests across error paths.

### Fixed

- **MockDataStore suite state:** `SetCurrentSuite` now updates `CurrentSuite` when no error hook is configured.

## [v3.3] - 2025-12-20

### Fixed

- **Stale tooling references:** Removed dead Makefile targets that referenced legacy JSON→SQLite tooling.

### Changed

- **Documentation consistency:** Updated `AUTOMATED_EVALUATION_SETUP.md` to match the README’s current commands and style.

## [v3.2] - 2025-12-20

### Fixed

- **Stats chart rendering:** Ensured the Chart.js container has a stable height so stacked bars render correctly.

### Changed

- **Navigation rail width:** Further reduced the left sidebar width to free up content space.
- **Manual evaluation layout:** Centered score selection and action controls for a more balanced layout.

## [v3.1] - 2025-12-19

### Added

- **Arena UI CSS regression tests:** Basic tests to prevent layout regressions (sidebar width variable, toolbar flex layouts, dropdown + file input styling, and scroll button anchoring).
- **UI Tour screenshot automation:** A Playwright-based capture script (`npm run screenshots`) to regenerate the README UI Tour screenshots deterministically.

### Changed

- **More compact Arena layout:** Thinner left navigation rail and tighter top bar spacing to prioritize content.
- **Toolbar layout fixes:** Sticky headers/footers and title/tool rows now compact into flex rows and only wrap when needed.
- **Styled dropdowns and file inputs:** `<select>` and `input[type="file"]` now match the Arena theme instead of rendering unstyled defaults.
- **Scroll buttons repositioned:** “↑/↓” buttons use the left sidebar space on shell pages (and stay bottom-right on solo pages).

## [v3.0] - 2025-12-19

### Added

- **Automated LLM evaluation (optional):** Multi-judge consensus scoring (via a Python FastAPI judge service) with job persistence, WebSocket progress, cost tracking, and an audit trail. See `AUTOMATED_EVALUATION_SETUP.md` and `python_service/`.
- **Encrypted API key storage:** AES-256-GCM encrypted key storage in SQLite with UI masking (configured via `ENCRYPTION_KEY`).
- **Arena UI overhaul:** A shared “Neon Glass Foundry” visual system + layout shell (`templates/arena.css`) applied across templates. See `DESIGN_CONCEPT.md` and `DESIGN_ROLLOUT.md`.
- **Coverage automation:** Coverage badge generation plus update scripts (`scripts/update-badge.sh`, `scripts/update-badge.ps1`) and `make update-coverage`.

### Changed

- **UI layout:** Templates standardized around a shared top bar + left rail (Arena shell), while preserving SSR Go templates and zero build tooling.
- **Evaluation workflow:** Added automated evaluation endpoints and async processing while keeping manual scoring fully supported.
- **Testability:** Refactors (e.g., handler dependency injection) to enable broader unit and integration testing.

### Removed

- **Legacy JSON migration tooling:** Removed the `v2.0` JSON→SQLite migration/remigration/dedup code paths and CLI flags.
- **Gemini CLI support:** Removed Gemini CLI integration.

## [v2.1] - 2025-03-16

### Added

- **Dynamic Profile Grouping:**  
  Implemented dynamic grouping of prompts by profile with color-coded borders and enhanced visual separation.  
  (_See middleware/utils.go and templates/results.html for implementation details._)

- **Enhanced Score Color Management:**  
  Centralized score colors and added utility functions in `templates/score-utils.js` so that pages and charts now share a unified color scheme.

- **Results Table Enhancements:**  
  Introduced row highlighting, sticky headers, tooltips, and a progress bar in the results table to boost user experience.

- **Keyboard Navigation:**  
  Enabled keyboard navigation in the evaluation grid for rapid score selection.

- **Smart Mock Score Generation:**  
  Improved random mock score generation using tiered, weighted distributions for realistic prototype testing.

- **WebSocket Recovery:**  
  Added auto-reconnection and connection status monitoring on the results page for reliable real-time updates.

### Fixed

- **Prompt Move Restrictions:**  
  Limited prompt moves to preserve profile group contiguity (_feat: restrict prompt moves_).

- **Profile Border Application:**  
  Corrected the application of border classes in both header and data cells, ensuring proper vertical borders, handling text overflow, and maintaining a consistent 50px cell size.

- **Model Deletion Logic:**  
  Fixed deletion in the WriteResults function to correctly remove models from the database.

- **UI Styling and Consistency:**  
  Addressed issues such as cell size uniformity, column separator accuracy, and constrained header widths in `templates/style.css`.

- **Case Sensitivity in Profiles:**  
  Resolved problems with profile ID mismatches and case sensitivity in grouping.

### Refactored

- **Unified Code Cleanup:**  
  Removed duplicate code, extracted hardcoded values, and consolidated inline CSS into centralized styles in `templates/style.css`.

- **Profile Group Utility:**  
  Moved profile grouping logic to `middleware/utils.go` to enhance maintainability.

- **Score Visualization Synchronization:**  
  Standardized score color schemes across results pages and charts through centralized utilities in `templates/score-utils.js`.

### Removed

- **Legacy Data Files:**  
  Deleted outdated JSON files (e.g., `data/current_suite.txt`, `data/profiles-default.json`, etc.) and obsolete SQLite WAL/shm files to streamline data management.

---

For full details, please refer to the git commit history.
