# v3.4 (2025-12-20)

This release focuses on testability and coverage, pushing the repo’s Go statement coverage comfortably above 90% and making it easier to keep it there.

## Highlights

### Coverage > 90%
- Expanded unit tests across middleware, CLI entrypoints, and the demo screenshot server.
- Repo-wide Go statement coverage now sits at **~90.3%** (rounded to one decimal).

### Testable Entrypoints
- Main server and demo-server are refactored around dependency-injected `run(...) int` helpers (unit-testable without `os.Exit`).

### Coverage Visibility
- Added a package-level coverage table to the README (and refreshed the local coverage badge/report workflow).

## Verification
- `CGO_ENABLED=1 go test ./... -v -race -cover`
