# Documentation Enforcement Guidelines

This document describes the automated enforcement mechanisms that validate documentation structure. When editing documentation files, you must maintain the required sections and formats.

## Overview

Several documentation files are automatically validated by tests and CI scripts. These mechanisms ensure consistency and enable automated updates (e.g., coverage tables).

## Enforced Documentation Files

### README.md

**Purpose:** Main project documentation

**Enforcement mechanisms:**

- `scripts/update_coverage_table.py` via `make update-coverage-table` regenerates the coverage table
- `make verify-docs` checks the Coverage section format with a regex
- `make coverage-enforce` fails when total statement coverage drops below 100.0% (also part of `make check` and CI)
- `readme_ui_screenshots_test.go` requires every `assets/ui-*.png` referenced in the Usage Tutorial to exist
- `readme_quickstart_test.go` forbids references to removed Make targets in the Quick Start section

**Required sections:**

- `### Coverage` - Must exist and contain the coverage table

**Required format:**

```markdown
### Coverage

Package-level statement coverage from `CGO_ENABLED=1 go test ./... -coverprofile coverage.out`:

| Package                                          |  Coverage |
| ------------------------------------------------ | --------: |
| llm-tournament                                   |     XX.X% |
| llm-tournament/handlers                          |     XX.X% |
| llm-tournament/integration                       |         - |
| llm-tournament/middleware                        |     XX.X% |
| llm-tournament/templates                         |     XX.X% |
| llm-tournament/testutil                          |     XX.X% |
| llm-tournament/tools/screenshots/cmd/demo-server |     XX.X% |
| **Total**                                        | **XX.X%** |
```

**How to update:**

- Run `make update-coverage-table` locally to regenerate the table
- Or run `make update-coverage` to update both badge and table

**What breaks it:**

- Removing or renaming the `### Coverage` section
- Changing the table format (columns, headers, or package names)
- Removing the backticks around the command
- Deleting a screenshot file that the Usage Tutorial references

### templates/*.html

**Purpose:** UI templates

**Enforcement mechanism:** `arena_theme_test.go` - `TestArenaTheme_AllTemplatesUseArenaCSS`

**Required elements:**

- Every page template must reference `href="/templates/arena.css"` or `href="/templates/output.css"`
- Every page template must declare `<html lang="en"` and the responsive viewport meta (`arena_theme_test.go`, `TestArenaTheme_AllTemplatesDeclareLanguageAndViewport`)
- No page template may load a script or stylesheet from a CDN; everything is served locally, including `templates/vendor/` (`arena_offline_test.go`, `TestTemplates_NoCDNScripts`)
- Nav anchors must not carry duplicate `class` attributes, which browsers silently drop (`arena_nav_test.go`, `TestArenaNav_NoDuplicateClassAttributes`)
- `design_preview.html` and `nav.html` are exempt from the stylesheet and lang/viewport rules (preview page and partial)

**What breaks it:**

- Adding a new page template without one of the required stylesheet references, the lang attribute, or the viewport meta
- Pointing a template at a CDN-hosted script instead of `templates/vendor/`

### DESIGN_CONCEPT.md / DESIGN_ROLLOUT.md / design_preview.html

**Purpose:** UI design system documentation and static preview (informational)

These files are not enforced by dedicated structure tests. Keep them consistent with the shipped UI when making design changes.

## Automated Enforcement in CI

### GitHub Actions

The CI pipeline (`.github/workflows/ci.yml`) includes:

1. **Tests** - Run all Go tests, including the documentation enforcement tests listed above

2. **Coverage updates** - Automatically updates README.md coverage table on main branch
   - Runs `make update-coverage-table`
   - Commits changes with `[skip ci]` tag

UI screenshots are not generated in CI. Regenerate them offline with `make screenshots` before pushing UI changes.

### Pre-commit Workflow (Recommended)

Before committing documentation changes:

1. Run `make verify-docs`

2. If updating coverage-related sections:

   ```bash
   make update-coverage-table
   ```

3. Run full test suite:
   ```bash
   make test
   ```

**Note:** All scripts in `scripts/` directory work from any directory. They automatically find the repository root and required files, so you don't need to be in the repo root when running them.

## Common Mistakes

### Mistake 1: Renaming Section Headers

**Wrong:**

```markdown
### Coverage Statistics # renamed (breaks regex matching)
```

**Right:**

```markdown
### Coverage # Keep exact header text
```

### Mistake 2: Removing Coverage Table Section

**Wrong:**

```markdown
### Coverage

(removed the table)
```

**Right:**

```markdown
### Coverage

Package-level statement coverage from `CGO_ENABLED=1 go test ./... -coverprofile coverage.out`:

| Package | Coverage |
| ------- | -------: |

| ...
```

## Adding New Enforcement

If you need to add enforcement for a new documentation file:

1. **Add a Go test** in the appropriate `*_test.go` file
   - Use table-driven tests for multiple assertions
   - Use clear error messages (e.g., "FILENAME.md missing required section: '## Section Name'")

2. **Add verification commands** to `Makefile` (if applicable)
   - Example: `make verify-docs` target

3. **Update this document** (`DOCUMENTATION_ENFORCEMENT.md`) with:
   - File name and purpose
   - Enforcement mechanism (test name, script, or CI job)
   - Required sections/format
   - How to update correctly
   - What breaks it

4. **Add pre-commit hook** (optional but recommended)
   - Create `.git/hooks/pre-commit` to run relevant tests

## Troubleshooting

### Test fails: "missing required section"

**Diagnosis:**

- Check if you renamed or removed a section header
- Verify exact string matching (case-sensitive, spacing matters)

**Solution:**

- Restore the original section header
- Run the test again to verify

### Coverage table doesn't update

**Diagnosis:**

- Coverage table format changed (columns, headers, etc.)
- Script regex doesn't match the new format

**Solution:**

- Run `make update-coverage-table` manually to see errors
- Check the script output for parsing errors
- Update the script regex if you intentionally changed the format

### CI fails on documentation changes

**Diagnosis:**

- Enforcement test fails
- Automated script fails to parse documentation

**Solution:**

- Run the failing test locally: `CGO_ENABLED=1 go test -run <test_name> -v`
- Fix the issue as described in this document
- Commit and push again

## Summary

- **Always run relevant tests** after editing documentation
- **Keep exact section headers** for files with regex-based enforcement
- **Run `make update-coverage-table`** if touching coverage sections
- **Check CI output** if automated updates fail
- **Update this document** when adding new enforcement mechanisms

For questions, refer to the enforcement test files:

- `readme_ui_screenshots_test.go` - README.md screenshot references
- `readme_quickstart_test.go` - README.md quick start targets
- `arena_theme_test.go` - template stylesheet references, lang attribute, viewport meta
- `arena_nav_test.go` - nav.html structure and duplicate class attributes
- `arena_offline_test.go` - no CDN script/stylesheet references in templates
- `scripts/update_coverage_table.py` - README.md coverage table
