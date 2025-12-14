#!/bin/bash
# Update coverage badge in README.md

# Extract coverage percentage (integer only)
COV=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/\..*//' | sed 's/%//')

# Determine badge color based on coverage
if [ "$COV" -ge 80 ]; then
    COLOR="green"
elif [ "$COV" -ge 60 ]; then
    COLOR="yellowgreen"
else
    COLOR="yellow"
fi

# Update README.md badge
sed -i "s/Coverage-[0-9]*%25-[a-z]*/Coverage-${COV}%25-${COLOR}/" README.md

echo "Coverage badge updated to ${COV}% (${COLOR})"
