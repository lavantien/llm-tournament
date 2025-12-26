# Update coverage table in README.md based on test results

param(
    [string]$CoverageFile = "coverage.out",
    [string]$ReadmePath = "README.md"
)

# Get script directory
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$RepoRoot = Split-Path -Parent $ScriptDir

# Resolve paths
if (-not [System.IO.Path]::IsPathRooted($CoverageFile)) {
    $CoverageFile = Join-Path $RepoRoot $CoverageFile
}
if (-not [System.IO.Path]::IsPathRooted($ReadmePath)) {
    $ReadmePath = Join-Path $RepoRoot $ReadmePath
}

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

# Change to repo root and run Python script
Push-Location $RepoRoot
try {
    $pythonScript = Join-Path $ScriptDir "update_coverage_table.py"
    python3 $pythonScript $CoverageFile $ReadmePath

    if ($LASTEXITCODE -eq 0) {
        Write-Green "Coverage table updated successfully"
    } else {
        Write-Red "Failed to update coverage table"
        exit 1
    }
} finally {
    Pop-Location
}
