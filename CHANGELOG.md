# Changelog

All notable changes are documented in this file.

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
