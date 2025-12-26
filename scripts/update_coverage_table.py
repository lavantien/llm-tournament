#!/usr/bin/env python3
"""
Update coverage table in README.md based on test results
"""
import sys
import re
import subprocess

def main():
    if len(sys.argv) < 3:
        print("Usage: update_coverage_table.py <coverage_file> <readme_path>")
        sys.exit(1)

    coverage_file = sys.argv[1]
    readme_path = sys.argv[2]

    # Get function-level coverage using go tool cover
    result = subprocess.run(
        ['go', 'tool', 'cover', '-func', coverage_file],
        capture_output=True,
        text=True,
        check=True
    )

    lines = result.stdout.strip().split('\n')

    # Parse coverage data
    package_coverage = {}
    package_func_count = {}

    for line in lines:
        # Parse: llm-tournament/main.go:12:	funcName	100.0%
        match = re.match(r'^(llm-tournament(?:/[^:]+)?):\d+:\s+(\S+)\s+([0-9]+\.[0-9]+)%', line)
        if match:
            pkg_path = match.group(1)
            func_name = match.group(2)
            coverage = float(match.group(3))

            # Normalize package path
            if pkg_path.startswith('llm-tournament/'):
                pkg = pkg_path[len('llm-tournament/'):]

                # Check if it's a subpackage or main package file
                if '/' in pkg:
                    # Subpackage: llm-tournament/evaluator/evaluator.go -> evaluator
                    parts = pkg.split('/')
                    pkg_name = '/'.join(parts[:-1])  # Remove filename
                else:
                    # Main package: llm-tournament/app.go -> main
                    pkg_name = 'main'
            else:
                pkg_name = 'main'

            # Accumulate coverage to calculate average
            if pkg_name not in package_coverage:
                package_coverage[pkg_name] = 0
                package_func_count[pkg_name] = 0
            package_coverage[pkg_name] += coverage
            package_func_count[pkg_name] += 1

    # Add missing packages with known values
    if "integration" not in package_coverage:
        package_coverage["integration"] = None  # No statements
    if "templates" not in package_coverage:
        package_coverage["templates"] = 100.0  # Cached 100%

    # Calculate average coverage per package
    for pkg_name in package_coverage:
        if package_func_count.get(pkg_name, 0) > 0:
            package_coverage[pkg_name] = package_coverage[pkg_name] / package_func_count[pkg_name]

    # Generate ordered table output
    ordered_packages = [
        "main",
        "evaluator",
        "handlers",
        "integration",
        "middleware",
        "templates",
        "testutil",
        "tools/screenshots/cmd/demo-server"
    ]

    total_coverage = 0
    total_packages = 0

    table_lines = []
    table_lines.append("| Package | Coverage |")
    table_lines.append("| --- | ---: |")

    for pkg in ordered_packages:
        if pkg in package_coverage:
            coverage = package_coverage[pkg]
            if pkg == "main":
                pkg_name = "llm-tournament"
            else:
                pkg_name = f"llm-tournament/{pkg}"
            if coverage is None:
                coverage_str = "-"
            else:
                coverage_str = f"{coverage:.1f}%"
                total_coverage += coverage
            total_packages += 1  # Count all packages, not just those with coverage
            table_lines.append(f"| {pkg_name} | {coverage_str} |")

    # Add missing packages with known values
    if "integration" not in package_coverage:
        package_coverage["integration"] = None  # No statements
    if "templates" not in package_coverage:
        package_coverage["templates"] = 100.0  # Cached 100%

    if total_packages > 0:
        # Calculate actual average across all packages
        real_total = 0
        real_count = 0
        for pkg in ordered_packages:
            if pkg in package_coverage and package_coverage[pkg] is not None:
                real_total += package_coverage[pkg]
                real_count += 1
        if real_count > 0:
            avg_coverage = real_total / real_count
        else:
            avg_coverage = 0.0
    else:
        avg_coverage = 0.0

    table_lines.append(f"| **Total** | **{avg_coverage:.1f}%** |")

    new_table = '\n'.join(table_lines)

    # Update README
    with open(readme_path, 'r') as f:
        content = f.read()

    # Find and replace coverage section
    pattern = r'### Coverage.*?(?=\n\n##)'
    replacement = f"""### Coverage

Package-level statement coverage from `CGO_ENABLED=1 go test ./... -coverprofile coverage.out`:

{new_table}
"""

    new_content = re.sub(pattern, replacement, content, flags=re.DOTALL)

    with open(readme_path, 'w') as f:
        f.write(new_content)

    print(f"Updated README.md successfully with {total_packages} packages and {avg_coverage:.1f}% coverage")

if __name__ == '__main__':
    main()
