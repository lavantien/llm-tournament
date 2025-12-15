#!/bin/bash
# Update coverage badge in README.md

# Extract coverage percentage (with decimals)
COV=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')

# Determine badge color based on coverage (using awk for floating point comparison)
COLOR=$(awk -v cov="$COV" 'BEGIN {
    if (cov >= 80) print "green"
    else if (cov >= 60) print "yellowgreen"
    else print "yellow"
}')

# Update README.md badge (handles both integer and decimal percentages)
sed -i "s/Coverage-[0-9.]*%25-[a-z]*/Coverage-${COV}%25-${COLOR}/" README.md

echo "Coverage badge updated to ${COV}% (${COLOR})"
