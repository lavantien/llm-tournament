# Update coverage table in README.md based on test results

param(
    [string]$CoverageFile = "coverage.out",
    [string]$ReadmePath = "README.md"
)

# Color output functions
function Write-ColorOutput($ForegroundColor) {
    $fc = $host.UI.RawUI.ForegroundColor
    $host.UI.RawUI.ForegroundColor = $ForegroundColor
    if ($args) {
        Write-Output $args
    }
    $host.UI.RawUI.ForegroundColor = $fc
}

function Write-Red { Write-ColorOutput Red $args }
function Write-Green { Write-ColorOutput Green $args }
function Write-Cyan { Write-ColorOutput Cyan $args }

if (-not (Test-Path $CoverageFile)) {
    Write-Red "Coverage file not found: $CoverageFile"
    exit 1
}

if (-not (Test-Path $ReadmePath)) {
    Write-Red "README.md not found"
    exit 1
}

Write-Cyan "Parsing coverage data..."

# Run Python script to update README
python3 scripts/update_coverage_table.py $CoverageFile $ReadmePath

if ($LASTEXITCODE -eq 0) {
    Write-Green "Coverage table updated successfully"
} else {
    Write-Red "Failed to update coverage table"
    exit 1
}
