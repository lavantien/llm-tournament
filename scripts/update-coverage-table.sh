#!/usr/bin/env bash
# Update coverage table in README.md based on test results

set -euo pipefail

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
NC='\033[0m'

COVERAGE_FILE="${1:-coverage.out}"
README_PATH="${2:-README.md}"

if [[ ! -f "$COVERAGE_FILE" ]]; then
    echo -e "${RED}Coverage file not found: $COVERAGE_FILE${NC}"
    exit 1
fi

if [[ ! -f "$README_PATH" ]]; then
    echo -e "${RED}README.md not found${NC}"
    exit 1
fi

echo -e "${CYAN}Parsing coverage data...${NC}"

# Run Python script to update README
python3 scripts/update_coverage_table.py "$COVERAGE_FILE" "$README_PATH"

echo -e "${GREEN}Coverage table updated successfully${NC}"
