#!/usr/bin/env python3
"""
Update coverage table in README.md based on test results
"""
import sys
import re
import subprocess
import os

def list_packages(repo_root):
    """List all Go packages in the module via go list"""
    result = subprocess.run(
        ['go', 'list', './...'],
        capture_output=True,
        text=True,
        timeout=60,
        cwd=repo_root
    )
    return [line.strip() for line in result.stdout.splitlines() if line.strip()]

def get_package_coverage(packages, repo_root):
    """Run go test on each package and extract statement coverage"""
    coverage_map = {}

    # Use NUL on Windows, /dev/null on Unix-like systems
    null_device = 'NUL' if os.name == 'nt' else '/dev/null'

    for pkg in packages:
        try:
            result = subprocess.run(
                ['go', 'test', pkg, f'-coverprofile={null_device}'],
                capture_output=True,
                text=True,
                timeout=60,
                cwd=repo_root
            )

            # Parse coverage: "ok  llm-tournament/handlers	1.572s	coverage: 98.8% of statements"
            match = re.search(r'coverage:\s+([0-9]+\.[0-9]+)%', result.stdout)
            if match:
                coverage_map[pkg] = float(match.group(1))
            else:
                # Check for "coverage: [no statements]"
                if '[no statements]' in result.stdout:
                    coverage_map[pkg] = None
                else:
                    coverage_map[pkg] = 0.0
        except subprocess.TimeoutExpired:
            coverage_map[pkg] = None
        except Exception:
            coverage_map[pkg] = 0.0

    return coverage_map

def get_total_coverage(repo_root, coverage_file):
    """Get actual total coverage from go tool cover -func"""
    try:
        result = subprocess.run(
            ['go', 'tool', 'cover', '-func', coverage_file],
            capture_output=True,
            text=True,
            timeout=60,
            cwd=repo_root
        )
        # Parse "total:statements:	99.3%"
        match = re.search(r'total:.*?\s+([0-9]+\.[0-9]+)%', result.stdout)
        if match:
            return float(match.group(1))
    except Exception:
        pass
    return None

def main():
    if len(sys.argv) < 3:
        print("Usage: update_coverage_table.py <coverage_file> <readme_path>")
        sys.exit(1)

    coverage_file = sys.argv[1]
    readme_path = sys.argv[2]
    
    # Get repository root (script is in scripts/, so parent dir is repo root)
    script_dir = os.path.dirname(os.path.abspath(__file__))
    repo_root = os.path.dirname(script_dir)
    
    # Get absolute path for readme
    if not os.path.isabs(readme_path):
        readme_path = os.path.join(repo_root, readme_path)

    # Packages to check (discovered from the module, so the table always
    # matches the current tree)
    packages = list_packages(repo_root)

    # Get statement-level coverage for each package
    package_coverage = get_package_coverage(packages, repo_root)

    # Get actual total coverage from go tool cover -func (weighted by statement count)
    total_coverage = get_total_coverage(repo_root, coverage_file)
    if total_coverage is None:
        # Fallback to simple average if go tool cover fails
        total_coverage = 0.0
        total_packages = 0
        for pkg_key in packages:
            if pkg_key in package_coverage:
                cov = package_coverage[pkg_key]
                if cov is not None:
                    total_coverage += cov
                    total_packages += 1
        if total_packages > 0:
            total_coverage = total_coverage / total_packages

    # Generate ordered table output (go list order: root package, then
    # subpackages sorted by import path)
    table_lines = []
    table_lines.append("| Package | Coverage |")
    table_lines.append("| --- | ---: |")

    for pkg in packages:
        coverage = package_coverage.get(pkg)
        if coverage is None:
            coverage_str = "-"
        else:
            coverage_str = f"{coverage:.1f}%"
        table_lines.append(f"| {pkg} | {coverage_str} |")

    table_lines.append(f"| **Total** | **{total_coverage:.1f}%** |")

    new_table = '\n'.join(table_lines)

    # Update README
    with open(readme_path, 'r', encoding='utf-8') as f:
        content = f.read()

    # Find and replace coverage section (handles numbered or unnumbered subsections)
    pattern = r'###(?:\s+[\d.]+\s+)?Coverage.*?(?=\n\n##|\n\[)'
    replacement = f"""### 9.2 Coverage

Package-level statement coverage from `CGO_ENABLED=1 go test ./... -coverprofile coverage.out`:

{new_table}
"""

    new_content = re.sub(pattern, replacement, content, flags=re.DOTALL)

    with open(readme_path, 'w', encoding='utf-8') as f:
        f.write(new_content)

    # Count packages with actual coverage (excluding "no statements" ones)
    packages_with_coverage = sum(1 for pkg in packages
                                  if package_coverage.get(pkg) is not None)

    print(f"Updated README.md successfully with {packages_with_coverage} packages and {total_coverage:.1f}% coverage")

if __name__ == '__main__':
    main()
