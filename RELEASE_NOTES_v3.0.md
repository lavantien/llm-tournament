# v3.0 (2025-12-19)

This is a major release that bundles **all changes since `v2.0`** (including the `v2.1` UI improvements), with a focus on:

- An optional **Automated LLM Evaluation** system (multi-judge consensus, job queue, cost tracking)
- A full **Arena UI overhaul** (no build step; still SSR Go templates)
- Significant **test suite expansion** + automated coverage reporting

## Highlights

### Automated Evaluation (New)
- Optional **Python FastAPI judge service** (`python_service/`) integrating **LiteLLM** + provider SDKs.
- **Multi-judge consensus scoring** (Claude Opus 4.5 / GPT-5.2 / Gemini 3 Pro) with audit trail (reasoning + confidence).
- **Async job queue** in Go (`evaluator/`) with persistence in SQLite and **WebSocket** progress broadcasts.
- **Cost estimation** and budget alerting, plus cancelable evaluations from the UI.
- **AES-256-GCM encrypted API key storage** (configured via `ENCRYPTION_KEY`; keys are masked in the UI).

### UI Overhaul (Arena)
- New “Arena” layout + design system (“Neon Glass Foundry”) documented in `DESIGN_CONCEPT.md` and `DESIGN_ROLLOUT.md`.
- Shared stylesheet `templates/arena.css` applied across templates (no bundler/build tooling).
- Results-grid UX upgrades (from `v2.1`): sticky headers, row highlighting, tooltips, keyboard navigation, and a unified score color scheme.
- Dynamic profile grouping and separation on the Results page, with consistent borders/colors.

### Quality & Testing
- Large Go test suite across `handlers/`, `middleware/`, `evaluator/`, `templates/`, and `integration/`.
- Coverage badge and auto-update scripts (`scripts/update-badge.sh`, `scripts/update-badge.ps1`) and `make update-coverage`.
- Refactors for testability (e.g., handler dependency injection).

## Verification

- `CGO_ENABLED=1 go test ./... -v -race -cover` (pass)
- Total statement coverage: **79.6%** (`CGO_ENABLED=1 go test ./... -count=1 -coverprofile coverage.out` + `go tool cover -func coverage.out`)

## Breaking / Migration Notes

- **Removed legacy JSON→SQLite migration tooling from `v2.0`:** `--migrate-to-sqlite`, `--remigrate-scores`, `--cleanup-duplicates`, and related code paths are no longer present.
  - If you still need to migrate old JSON state, run the migration on `v2.0` to produce `data/tournament.db`, then upgrade to `v3.0`.
- **Automated evaluation now relies on an encryption key:** if you enable automated evaluation, set `ENCRYPTION_KEY` (32-byte hex) and run the Python judge service (`python_service/main.py`, default `:8001`).
  - Manual evaluation continues to work without the Python service.

## Upgrade Checklist

1. Ensure Go is installed and `CGO_ENABLED=1` is enabled (SQLite).
2. Back up your existing `data/tournament.db` before upgrading.
3. Start the Go server and verify the Arena UI loads.
4. Optional: start the Python judge service and configure provider keys at `/settings` (see `AUTOMATED_EVALUATION_SETUP.md`).

## Full Changelog (v2.0 → v3.0)

### Commits included
- fix: Correctly delete models from the database in WriteResults function (3206778)
- update: models (237e5c1)
- clean: old json (eafb665)
- feat: Implement dynamic profile grouping with color-coded borders (cffd009)
- refactor: Group prompts by profile for results page (a0fa661)
- fix: Correct profile order, uncategorized display, and separator lines (c27e5ea)
- fix: Remove hardcoded profile order and use database order instead (14c14bf)
- fix: Order profiles by prompt appearance and handle missing profiles (d090047)
- Refactor: Extract profile group logic to middleware/utils.go (45455b4)
- refactor: Remove unused imports from results.go (73a9ee1)
- fix: Resolve type mismatch in profile group handling (c25aa0b)
- style: Fix UI issues: cell size, column separators, column 42 size (48091bb)
- style: Fix header overflow and standardize cell size to 50px square (97c78cb)
- refactor: Remove ad-hoc fix for last column in results table (aa5d706)
- style: Constrain profile headers to prevent column stretching (6f0821c)
- fix: imports (0c591ad)
- fix: handle text overflow on header (7bbcec2)
- fix: Reduce max-width of profile headers to 50px for better display (326c8cd)
- refactor: Move CSS from results.html to style.css for better styling (4f5f864)
- fix: center the score table (6be14dc)
- refactor: Move CSS from results.html to style.css and use CSS classes (049e167)
- refactor: Improve profile styling with CSS classes and variables (9c9d693)
- refactor: Extract hardcoded header styles to style.css (7ff392c)
- style: Extract inline CSS to style.css (325c2ee)
- style: Extract hardcoded styles to CSS (a2c7063)
- fix: Remove duplicate className assignments in total score cell creation (e11a436)
- fix: ensure progress-bar-standard-width takes precedence (a15a810)
- style: synchronize score color scheme across pages and chart (a4df97e)
- style: Sync score color scheme across pages and chart (71ddd81)
- refactor: standardize score color management (daa274b)
- feat: Update totalScoresChart to use score-buttons color scheme (e1c60cd)
- feat: add score color utility functions and debugging (d904542)
- feat: centralize score colors in score-utils.js (9a96c92)
- fix: remove undefined scoreColorDefault function from evaluate.html (48b60e3)
- style: Add separator lines between profile headers (b6b898d)
- feat: enhance profile group separation with 5px borders (9fa53d7)
- refactor: Simplify profile group border styling (f9f14a5)
- fix: ensure profile-start class is applied to first cell of each profile group row (3f5e6de)
- fix: handle case sensitivity in profile ID properties (1c975cd)
- fix: handle case mismatch in profile group properties (e7833dc)
- fix: strengthen border styling for profile groups (6901431)
- refactor: Simplify border application and remove unnecessary CSS variables (622b7b31)
- style: Extract inline styles to CSS and add profile-specific classes (afe562b)
- refactor: remove hardcoded profile classes and use dynamic inline styles (02f00b24)
- style: Extract cell style manipulations to CSS (acbadd5)
- feat: apply borders directly with inline styles (eb099d3)
- fix: Ensure borders are correctly applied to header and content rows (725a4ec)
- refactor: Simplify profile border application and improve styling (b085d39)
- fix: ensure vertical profile borders appear in table data cells (2f0a4bd)
- feat: add row highlighting, sticky header, tooltips, and keyboard navigation (8c05adf)
- feat: add styles for results table (0db55c3)
- fix: profile group separator (4bf19cb)
- refactor: Clean up code, remove duplicates, extract hardcoded values, and improve maintainability (e42c25c)
- fix: Correct SEARCH block to match exact lines in templates/style.css (e227efcc)
- style: Use #333 for profile color (cb914fa)
- feat: restrict prompt moves to maintain profile group contiguity (7c3f34a)
- docs: Update CHANGELOG.md for v2.1 release (8bd272f)
- docs: update images (f507dc0)
- update: latest models (279db80)
- ai: add rule files for claude code and gemini cli (f44408a)
- feat: Add automated LLM evaluation system with multi-judge consensus (73d39b2)
- chore: ignore built artifracts (17b627d)
- chore: add Automated LLM Evaluation System - Implementation Plan (8e4463c)
- docs: Update README.md to reflect automated evaluation system (4daa24f)
- ai: enhance system prompt (72d2ae8)
- claudecode: full tdd-guard integration (c8b66ac)
- feat: full claude code intergration (e71ed3f)
- feat: enhance readme (7d06935)
- ai: custom tdd-guard (580d207)
- test: add comprehensive test suite (42% coverage) (3d9f8ae)
- test: expand test coverage to 51% (164381d)
- refactor: delete legacy migration code, add template tests (68.6% coverage) (cf31d76)
- chore: settings updated (55db0c2)
- test: expand test coverage to 71.8% (dc90a66)
- test: expand coverage to 73% with evaluator and handler tests (3fa10da)
- feat: add make update-coverage to auto-update badge (be8a785)
- test: expand coverage to 74% with handler edge case tests (fceb42e)
- qol: improve code coverage (d617beb)
- qol: improve coverage (7578b82)
- chore: ignore correctly (0190261)
- chore: drop gemini cli support (5203d1b)
- qol: add plans (1945a0a)
- update claude.md (72f3e56)
- refactor: add dependency injection to handlers for testability (5a88660)
- test: increase coverage from 74.6% to 79.1% (517b700)
- chore: ignore trash (fdb34b4)
- fix: current suite in db instead of file (8e7c8fd)
- chore: ignore (a4b2bcb)
- chore: update docs (4786f62)
- enh: improve stability (71cf0aa)
- feat: auto coverage badge (8a9e4e3)
- feat: auto update badge (8db545b)
- feat: enhance greatme (2495e51)
- claudecode: optimize setup (83ffe99)
- claudecode: update rules (5f18368)
- ai: update agents rules (075b1fc)
- ai: update rules (e7c2ff4)
- ai: update rules (7c2a058)
- ai: update rules (5dd89f1)
- docs: update architecture & development guide (545bf64)
- ui: overhaul, check DESIGN_*.md (4fb1624)
