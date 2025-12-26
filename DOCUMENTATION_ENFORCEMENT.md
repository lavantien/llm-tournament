# Documentation Enforcement Guidelines

This document describes the automated enforcement mechanisms that validate documentation structure. When editing documentation files, you must maintain the required sections and formats.

## Overview

Several documentation files are automatically validated by tests and CI scripts. These mechanisms ensure consistency and enable automated updates (e.g., coverage tables, screenshots).

## Enforced Documentation Files

### README.md

**Purpose:** Main project documentation

**Enforcement mechanism:** `scripts/update_coverage_table.py` and `make update-coverage-table`

**Required sections:**
- `### Coverage` - Must exist and contain the coverage table

**Required format:**
```markdown
### Coverage

Package-level statement coverage from `CGO_ENABLED=1 go test ./... -coverprofile coverage.out`:

| Package | Coverage |
| --- | ---: |
| llm-tournament | XX.X% |
| llm-tournament/evaluator | XX.X% |
| llm-tournament/handlers | XX.X% |
| llm-tournament/integration | - |
| llm-tournament/middleware | XX.X% |
| llm-tournament/templates | XX.X% |
| llm-tournament/testutil | XX.X% |
| llm-tournament/tools/screenshots/cmd/demo-server | XX.X% |
| **Total** | **XX.X%** |
```

**How to update:**
- Run `make update-coverage-table` locally to regenerate the table
- Or run `make update-coverage` to update both badge and table

**What breaks it:**
- Removing or renaming the `### Coverage` section
- Changing the table format (columns, headers, or package names)
- Removing the backticks around the command

### DESIGN_CONCEPT.md

**Purpose:** UI design specifications and migration plan

**Enforcement mechanism:** `design_preview_test.go:10` - `TestDesignConceptAndPreview_ExistAndStructured`

**Required sections:**
- `# LLM Tournament Arena — Design Concept` (header)
- `## Color Palette` (exact string match)
- `## Typography` (exact string match)
- `## UI Elements` (exact string match)
- `## Libraries (CDN Only)` (exact string match)

**How to update:**
- Maintain these exact section headers
- Add content under sections as needed
- Do not rename or remove these sections

**What breaks it:**
- Changing section header text (e.g., "## Color System" instead of "## Color Palette")
- Using different header levels (e.g., "### Color Palette" instead of "## Color Palette")
- Removing any of the required sections

**To verify:**
```bash
CGO_ENABLED=1 go test -run TestDesignConceptAndPreview_ExistAndStructured -v
```

### design_preview.html

**Purpose:** Visual preview of the UI design

**Enforcement mechanism:** `design_preview_test.go:10` - `TestDesignConceptAndPreview_ExistAndStructured`

**Required elements:**
- `<title>LLM Tournament Arena — Design Preview</title>` (exact string)
- `href="/assets/favicon.ico"` (exact string)
- `src="/assets/logo.webp"` (exact string)
- Class names: `arena-shell`, `glass-panel`, `neon-button`
- Library references: `Chart.js`

**Constraints:**
- Must be hardcoded HTML (no Go template actions like `{{` or `}}`)

**How to update:**
- Keep the required elements in the file
- Do not convert to a Go template

**What breaks it:**
- Removing or changing the required strings or class names
- Adding Go template syntax (`{{`, `}}`)

**To verify:**
```bash
CGO_ENABLED=1 go test -run TestDesignConceptAndPreview_ExistAndStructured -v
```

## Automated Enforcement in CI

### GitHub Actions

The CI pipeline (`.github/workflows/ci.yml`) includes:

1. **Tests** - Run all Go tests, including documentation enforcement tests
   - `TestDesignConceptAndPreview_ExistAndStructured` validates DESIGN_CONCEPT.md and design_preview.html

2. **Coverage updates** - Automatically updates README.md coverage table on main branch
   - Runs `make update-coverage-table`
   - Commits changes with `[skip ci]` tag

3. **Screenshots** - Automatically generates UI screenshots on main branch
   - Runs `make screenshots`
   - Commits changes with `[skip ci]` tag

### Pre-commit Workflow (Recommended)

Before committing documentation changes:

1. Run relevant enforcement test:
   ```bash
   CGO_ENABLED=1 go test -run <test_name> -v
   ```

2. If updating coverage-related sections:
   ```bash
   make update-coverage-table
   ```

3. Run full test suite:
   ```bash
   make test
   ```

**Note:** All scripts in `scripts/` directory work from any directory. They automatically find the repository root and required files, so you don't need to be in the repo root when running them.

2. If updating coverage-related sections:
   ```bash
   make update-coverage-table
   ```

3. Run full test suite:
   ```bash
   make test
   ```

## Common Mistakes

### Mistake 1: Renaming Section Headers

**Wrong:**
```markdown
## Color Palette  # Original
## Color System  # Changed (breaks regex matching)
```

**Right:**
```markdown
## Color Palette  # Keep exact header text
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
| --- | ---: |
| ...
```

### Mistake 3: Adding Go Template Syntax to Preview HTML

**Wrong:**
```html
<title>{{.Title}}</title>  <!-- Breaks test -->
```

**Right:**
```html
<title>LLM Tournament Arena — Design Preview</title>  <!-- Hardcoded -->
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
- `design_preview_test.go:10` - DESIGN_CONCEPT.md and design_preview.html
- `scripts/update_coverage_table.py` - README.md coverage table
