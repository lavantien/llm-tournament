#!/usr/bin/env bash
# Update coverage badge based on test results

set -euo pipefail

OUTPUT_PATH="${1:-coverage-badge.svg}"
VERBOSE="${VERBOSE:-false}"

# Color codes for terminal output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
WHITE='\033[1;37m'
NC='\033[0m' # No Color

get_coverage_color() {
    local coverage=$1
    echo "$coverage" | awk '
    {
        if ($1 >= 90) { print "blueviolet"; exit }
        else if ($1 >= 80) { print "brightgreen"; exit }
        else if ($1 >= 60) { print "green"; exit }
        else if ($1 >= 40) { print "yellow"; exit }
        else if ($1 >= 20) { print "orange"; exit }
        else { print "red"; exit }
    }'
}

get_coverage_percentage() {
    echo -e "${CYAN}Running tests with coverage...${NC}"

    # Set CGO_ENABLED=1 for SQLite support
    export CGO_ENABLED=1

    # Run tests and capture output
    local coverage_file="coverage.out"
    local test_output

    if ! test_output=$(go test ./... -coverprofile="$coverage_file" 2>&1); then
        echo -e "${RED}Tests failed:${NC}"
        echo "$test_output"
        return 1
    fi

    if [[ "$VERBOSE" == "true" ]]; then
        echo "$test_output"
    fi

    # Parse coverage from output
    if [[ -f "$coverage_file" ]]; then
        local coverage_output
        coverage_output=$(go tool cover -func="$coverage_file" | tail -n 1)

        if [[ "$coverage_output" =~ total:[[:space:]]+\(statements\)[[:space:]]+([0-9.]+)% ]]; then
            local coverage="${BASH_REMATCH[1]}"
            echo -e "${GREEN}Total coverage: ${coverage}%${NC}"

            # Cleanup
            rm -f "$coverage_file"

            echo "$coverage"
            return 0
        fi
    fi

    echo -e "${RED}Could not parse coverage percentage from test output${NC}"
    return 1
}

create_badge() {
    local coverage=$1
    local output=$2

    local color
    color=$(get_coverage_color "$coverage")

    # Format coverage with one decimal place
    local coverage_text
    coverage_text=$(printf "%.1f%%" "$coverage")

    # URL encode the percentage sign
    coverage_text="${coverage_text//%/%25}"

    # Generate shields.io badge URL
    local badge_url="https://img.shields.io/badge/coverage-${coverage_text}-${color}"

    echo -e "${CYAN}Downloading badge from: ${badge_url}${NC}"

    if curl -fsSL "$badge_url" -o "$output"; then
        echo -e "${GREEN}Badge saved to: ${output}${NC}"
        return 0
    else
        echo -e "${RED}Failed to download badge${NC}"
        return 1
    fi
}

update_readme_badge() {
    local coverage=$1
    local readme_path="README.md"

    if [[ ! -f "$readme_path" ]]; then
        echo -e "${YELLOW}README.md not found, skipping badge update${NC}"
        return 0
    fi

    local badge_markdown="![Coverage](./coverage-badge.svg)"

    # Try to replace existing coverage badge
    if grep -q '!\[Coverage\]' "$readme_path"; then
        # Use sed for in-place replacement (compatible with both GNU and BSD sed)
        if sed --version >/dev/null 2>&1; then
            # GNU sed
            sed -i "s|!\[Coverage\]([^)]*)|${badge_markdown}|g" "$readme_path"
        else
            # BSD sed (macOS)
            sed -i '' "s|!\[Coverage\]([^)]*)|${badge_markdown}|g" "$readme_path"
        fi
        echo -e "${GREEN}Updated coverage badge in README.md to reference local SVG${NC}"
    else
        echo -e "${YELLOW}No existing coverage badge found in README.md${NC}"
        echo -e "${CYAN}Add this line to your README.md:${NC}"
        echo -e "${WHITE}[${badge_markdown}](./coverage.html)${NC}"
    fi
}

# Main execution
main() {
    echo -e "${MAGENTA}=== Coverage Badge Updater ===${NC}"
    echo ""

    local coverage
    if ! coverage=$(get_coverage_percentage); then
        echo -e "${RED}Error: Failed to get coverage percentage${NC}"
        exit 1
    fi

    echo ""
    if ! create_badge "$coverage" "$OUTPUT_PATH"; then
        echo -e "${RED}Error: Failed to create badge${NC}"
        exit 1
    fi

    echo ""
    update_readme_badge "$coverage"

    echo ""
    echo -e "${GREEN}Done!${NC}"
}

main "$@"
