#!/usr/bin/env python3
"""
Checksum integrity validator for PKGBUILD files.
Validates that source entries have matching checksums.
Usage: python3 checksum_validator.py PKGBUILD
"""

import re
import sys
from pathlib import Path


def parse_pkgbuild(path: str):
    """Parse PKGBUILD and extract source and checksum arrays."""
    src = Path(path).read_text(errors="replace")
    lines = src.splitlines()

    sources = []
    checksums = []
    current_array = None
    current_var = None

    for line in lines:
        # Detect array start
        arr_match = re.match(r"^\s*([a-zA-Z_][a-zA-Z0-9_]*)=(\()", line)
        if arr_match:
            current_var = arr_match.group(1)
            current_array = []

        # Detect array end
        if current_array is not None and ")" in line:
            # Extract array contents
            array_content = re.search(r"\((.+)\)", line)
            if array_content:
                content = array_content.group(1)
                # Split by comma, handling quotes
                items = re.findall(r'"([^"]*)"', content)
                items += re.findall(r"'([^']*)'", content)
                current_array.extend(items)

            if current_var in [
                "source",
                "sha256sums",
                "sha512sums",
                "b2sums",
                "md5sums",
                "sha1sums",
            ]:
                if current_var == "source":
                    sources = current_array
                else:
                    checksums = current_array

            current_array = None
            current_var = None

        # Handle multi-line arrays
        elif current_array is not None:
            # Check if line contains array items
            items = re.findall(r'"([^"]*)"', line)
            items += re.findall(r"'([^']*)'", line)
            current_array.extend(items)

    return sources, checksums


def validate_checksums(sources: list, checksums: list) -> list:
    """Validate checksum integrity and return findings."""
    findings = []

    # Check for SKIP entries
    if "SKIP" in sources:
        findings.append(
            {
                "type": "SKIP_FOUND",
                "severity": "HIGH",
                "message": "Found SKIP entry in source array - integrity verification bypassed",
            }
        )

    # Check for mismatched counts
    if len(sources) != len(checksums):
        findings.append(
            {
                "type": "MISMATCH",
                "severity": "HIGH",
                "message": f"Source count ({len(sources)}) does not match checksum count ({len(checksums)})",
            }
        )

    # Check for weak checksums
    weak_checksums = ["md5sums", "sha1sums"]
    for weak in weak_checksums:
        if weak in checksums:
            findings.append(
                {
                    "type": "WEAK_CHECKSUM",
                    "severity": "MEDIUM",
                    "message": f"Found {weak} - consider using stronger algorithm (sha256sums or sha512sums)",
                }
            )

    # Check for missing checksums on non-VCS sources
    vcs_patterns = ["git+", "svn+", "hg+", "bzr+"]
    non_vcs_sources = []
    for src in sources:
        is_vcs = any(pattern in src for pattern in vcs_patterns)
        if not is_vcs and "http" not in src and "https" not in src:
            non_vcs_sources.append(src)

    # If we have non-VCS sources, we should have checksums
    if non_vcs_sources and len(checksums) == 0:
        findings.append(
            {
                "type": "MISSING_CHECKSUMS",
                "severity": "HIGH",
                "message": f"Non-VCS sources found without checksums: {non_vcs_sources}",
            }
        )

    return findings


def main():
    if len(sys.argv) != 2:
        print(f"Usage: {sys.argv[0]} PKGBUILD")
        sys.exit(1)

    path = sys.argv[1]
    try:
        sources, checksums = parse_pkgbuild(path)
        findings = validate_checksums(sources, checksums)

        if not findings:
            print("Checksum validation passed.")
            print(f"Sources: {len(sources)}, Checksums: {len(checksums)}")
        else:
            print(f"Found {len(findings)} issue(s):")
            for finding in findings:
                print(f"\n[{finding['severity']}] {finding['type']}")
                print(f"  {finding['message']}")
    except FileNotFoundError:
        print(f"Error: File not found: {path}")
        sys.exit(1)
    except Exception as e:
        print(f"Error: {e}")
        sys.exit(1)


if __name__ == "__main__":
    main()
