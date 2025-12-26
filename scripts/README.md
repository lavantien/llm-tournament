# Scripts Directory

This directory contains utility scripts for development, testing, and automation.

## Development Scripts

### update-badge.sh / update-badge.ps1
**Purpose:** Update the coverage badge SVG in the repository root

**Usage:**
- Linux/Mac: `./scripts/update-badge.sh`
- Windows: `powershell -ExecutionPolicy Bypass -File scripts\update-badge.ps1`

**What it does:**
- Runs tests with coverage
- Extracts coverage percentage
- Downloads and saves SVG badge from shields.io
- Updates coverage badge reference in README.md

**Automated:** Runs on CI via `make update-coverage`

### update-coverage-table.sh / update-coverage-table.ps1
**Purpose:** Update the coverage table in README.md

**Usage:**
- Linux/Mac: `bash scripts/update-coverage-table.sh`
- Windows: `powershell -ExecutionPolicy Bypass -File scripts\update-coverage-table.ps1`
- Or via Makefile: `make update-coverage-table`

**What it does:**
- Runs tests with coverage
- Parses `go tool cover -func` output
- Calculates per-package coverage
- Updates README.md coverage table with current values

**Automated:** Runs on CI via `make update-coverage-table` and commits changes

### update_coverage_table.py
**Purpose:** Python script that does the actual coverage table update

**Usage:** Called by the shell/PowerShell wrapper scripts

**What it does:**
- Parses coverage data from `go tool cover -func`
- Calculates average coverage per package
- Updates README.md coverage section via regex replacement

**Enforced:** See [DOCUMENTATION_ENFORCEMENT.md](../DOCUMENTATION_ENFORCEMENT.md) for required format

### pre-commit
**Purpose:** Git pre-commit hook for documentation and code quality checks

**Usage:**
```bash
# Install as git hook
cp scripts/pre-commit .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit

# Then it runs automatically on each commit
```

**What it does:**
- Checks if documentation files (README.md, DESIGN_CONCEPT.md, design_preview.html) are staged
- If yes, runs `make verify-docs`
- Checks if Go files are staged
- If yes, runs `make lint`

**Prevents:** Committing documentation that fails enforcement tests or code that fails linting

## Makefile Targets Related to Scripts

### make update-coverage
- Runs tests with coverage
- Generates coverage.html
- Updates coverage badge SVG

### make update-coverage-table
- Runs tests with coverage
- Generates coverage.html
- Updates coverage table in README.md

### make verify-docs
- Runs documentation enforcement tests
- Validates README.md coverage table format
- Checks DESIGN_CONCEPT.md and design_preview.html structure

### make screenshots
- Builds CSS
- Runs Playwright screenshot generation
- Generates UI screenshots in assets/

## Documentation Enforcement

Several scripts and tests enforce specific documentation formats. See [DOCUMENTATION_ENFORCEMENT.md](../DOCUMENTATION_ENFORCEMENT.md) for:

- List of enforced documentation files
- Required sections and formats
- How to update documentation without breaking automation
- Troubleshooting common mistakes

## Adding New Scripts

When adding a new script to this directory:

1. **Make it executable** (bash/sh scripts): `chmod +x scripts/your-script.sh`
2. **Add to Makefile** if it's a common task: Add new target to `.PHONY` and implement
3. **Document here**: Add section explaining purpose, usage, and what it does
4. **Update CI** if it should run automatically: Edit `.github/workflows/ci.yml`
5. **Test locally**: Verify script works before committing

## Platform Support

- **Linux/Mac:** Bash scripts (`.sh`) with Python 3
- **Windows:** PowerShell scripts (`.ps1`) with Python 3
- **Cross-platform:** Python scripts (`.py`) work on all platforms

Makefile automatically detects OS and calls appropriate scripts.

**Important:** All scripts are designed to work from any directory. They automatically locate the repository root and required files using relative paths to their own location. This ensures they work correctly in CI environments, even when the working directory varies.
