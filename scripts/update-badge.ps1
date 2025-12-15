#!/usr/bin/env pwsh
# Update coverage badge based on test results

param(
    [string]$OutputPath = "coverage-badge.svg",
    [switch]$Verbose
)

$ErrorActionPreference = "Stop"

function Get-CoverageColor {
    param([decimal]$coverage)

    if ($coverage -ge 80) { return "brightgreen" }
    elseif ($coverage -ge 60) { return "green" }
    elseif ($coverage -ge 40) { return "yellow" }
    elseif ($coverage -ge 20) { return "orange" }
    else { return "red" }
}

function Get-CoveragePercentage {
    Write-Host "Running tests with coverage..." -ForegroundColor Cyan

    # Set CGO_ENABLED=1 for SQLite support
    $env:CGO_ENABLED = "1"

    # Run tests and capture output
    $coverageFile = "coverage.out"
    $testOutput = & go test ./... -coverprofile=$coverageFile 2>&1

    if ($LASTEXITCODE -ne 0) {
        Write-Host "Tests failed:" -ForegroundColor Red
        Write-Host $testOutput
        throw "Test execution failed with exit code $LASTEXITCODE"
    }

    if ($Verbose) {
        Write-Host $testOutput
    }

    # Parse coverage from output
    if (Test-Path $coverageFile) {
        $coverageOutput = & go tool cover -func=$coverageFile | Select-Object -Last 1

        if ($coverageOutput -match 'total:\s+\(statements\)\s+([\d.]+)%') {
            $coverage = [decimal]$matches[1]
            Write-Host "Total coverage: $coverage%" -ForegroundColor Green

            # Cleanup
            Remove-Item $coverageFile -ErrorAction SilentlyContinue

            return $coverage
        }
    }

    throw "Could not parse coverage percentage from test output"
}

function New-Badge {
    param(
        [decimal]$coverage,
        [string]$output
    )

    $color = Get-CoverageColor -coverage $coverage
    $coverageText = "{0:N1}%" -f $coverage

    # Generate shields.io badge URL
    $badgeUrl = "https://img.shields.io/badge/coverage-$coverageText-$color"

    Write-Host "Downloading badge from: $badgeUrl" -ForegroundColor Cyan

    try {
        Invoke-WebRequest -Uri $badgeUrl -OutFile $output -UseBasicParsing
        Write-Host "Badge saved to: $output" -ForegroundColor Green
        return $true
    }
    catch {
        Write-Host "Failed to download badge: $_" -ForegroundColor Red
        return $false
    }
}

function Update-ReadmeBadge {
    param([decimal]$coverage)

    $readmePath = "README.md"

    if (-not (Test-Path $readmePath)) {
        Write-Host "README.md not found, skipping badge update" -ForegroundColor Yellow
        return
    }

    $badgeMarkdown = "[![Coverage](./coverage-badge.svg)]()"

    $readme = Get-Content $readmePath -Raw

    # Try to replace existing coverage badge
    if ($readme -match '!\[Coverage\]\([^)]+\)') {
        $readme = $readme -replace '!\[Coverage\]\([^)]+\)', $badgeMarkdown
        Set-Content -Path $readmePath -Value $readme -NoNewline
        Write-Host "Updated coverage badge in README.md to reference local SVG" -ForegroundColor Green
    }
    else {
        Write-Host "No existing coverage badge found in README.md" -ForegroundColor Yellow
        Write-Host "Add this line to your README.md:" -ForegroundColor Cyan
        Write-Host $badgeMarkdown -ForegroundColor White
    }
}

# Main execution
try {
    Write-Host "=== Coverage Badge Updater ===" -ForegroundColor Magenta
    Write-Host ""

    $coverage = Get-CoveragePercentage

    Write-Host ""
    New-Badge -coverage $coverage -output $OutputPath

    Write-Host ""
    Update-ReadmeBadge -coverage $coverage

    Write-Host ""
    Write-Host "Done!" -ForegroundColor Green
}
catch {
    Write-Host "Error: $_" -ForegroundColor Red
    exit 1
}
