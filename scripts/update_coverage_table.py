#!/usr/bin/env python3
"""
Update coverage table in README.md based on test results
"""
import sys
import re
import subprocess
import os

def get_package_coverage(packages, repo_root):
    """Run go test on each package and extract statement coverage"""
    coverage_map = {}
    
    # Use NUL on Windows, /dev/null on Unix-like systems
    null_device = 'NUL' if os.name == 'nt' else '/dev/null'
    
    for pkg in packages:
        try:
            # Use relative path from repo root
            if pkg == '.':
                pkg_path = '.'
            else:
                pkg_path = f'./{pkg}'
            
            result = subprocess.run(
                ['go', 'test', pkg_path, f'-coverprofile={null_device}'],
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

    # Packages to check
    packages = ['.', 'evaluator', 'handlers', 'integration', 'middleware', 'templates', 'testutil', 'tools/screenshots/cmd/demo-server']
    
    # Get statement-level coverage for each package
    package_coverage = get_package_coverage(packages, repo_root)

    # Generate ordered table output
    ordered_packages = [
        (".", "llm-tournament"),
        ("evaluator", "llm-tournament/evaluator"),
        ("handlers", "llm-tournament/handlers"),
        ("integration", "llm-tournament/integration"),
        ("middleware", "llm-tournament/middleware"),
        ("templates", "llm-tournament/templates"),
        ("testutil", "llm-tournament/testutil"),
        ("tools/screenshots/cmd/demo-server", "llm-tournament/tools/screenshots/cmd/demo-server")
    ]

    total_coverage = 0.0
    total_packages = 0

    table_lines = []
    table_lines.append("| Package | Coverage |")
    table_lines.append("| --- | ---: |")

    for pkg_key, pkg_name in ordered_packages:
        if pkg_key in package_coverage:
            coverage = package_coverage[pkg_key]
            if coverage is None:
                coverage_str = "-"
            else:
                coverage_str = f"{coverage:.1f}%"
                total_coverage += coverage
                total_packages += 1
            table_lines.append(f"| {pkg_name} | {coverage_str} |")

    if total_packages > 0:
        avg_coverage = total_coverage / total_packages
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
