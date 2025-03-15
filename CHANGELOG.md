# Changelog

All notable changes for version v2.1 are documented in this file.

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
