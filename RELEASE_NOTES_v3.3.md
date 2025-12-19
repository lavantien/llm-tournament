# v3.3 (2025-12-20)
This release is focused on documentation correctness and removing stale tooling.

## Highlights
### Remove Dead Make Targets
- Deleted the legacy `make migrate` / `make dedup` targets that referenced removed CLI flags.

### Doc Consistency Pass
- Updated `AUTOMATED_EVALUATION_SETUP.md` to match the README’s commands and formatting (notably `CGO_ENABLED=1 go run .` and PowerShell-friendly `ENCRYPTION_KEY` setup).

## Verification
- `CGO_ENABLED=1 go test ./... -v -race -cover`
