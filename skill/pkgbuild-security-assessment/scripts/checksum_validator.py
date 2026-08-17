#!/usr/bin/env python3
"""
Checksum integrity validator for PKGBUILD files.
Validates that source entries have matching checksums.
Usage: python3 checksum_validator.py PKGBUILD
"""

import re
import sys
from pathlib import Path

CHECKSUM_VARS = {"sha256sums", "sha512sums", "b2sums", "md5sums", "sha1sums"}
WEAK_ALGORITHMS = {"md5sums", "sha1sums"}


def parse_pkgbuild(path: str):
    """Parse PKGBUILD and extract source and checksum arrays."""
    src = Path(path).read_text(errors="replace")
    lines = src.splitlines()

    arrays = {}
    current_array = None
    current_var = None

    for line in lines:
        arr_match = re.match(r"^\s*([a-zA-Z_][a-zA-Z0-9_]*)=(\()", line)
        if arr_match:
            current_var = arr_match.group(1)
            current_array = []

        if current_array is not None:
            items = re.findall(r'"([^"]*)"', line)
            items += re.findall(r"'([^']*)'", line)
            current_array.extend(items)

            if ")" in line:
                arrays[current_var] = current_array
                current_array = None
                current_var = None

    sources = arrays.get("source", [])
    checksum_var = None
    checksums = []
    for var in CHECKSUM_VARS:
        if var in arrays:
            checksum_var = var
            checksums = arrays[var]
            break

    return sources, checksums, checksum_var


def validate_checksums(sources: list, checksums: list, checksum_var: str | None) -> list:
    """Validate checksum integrity and return findings."""
    findings = []

    for i, cs in enumerate(checksums):
        if cs == "SKIP":
            findings.append(
                {
                    "type": "SKIP_FOUND",
                    "severity": "HIGH",
                    "message": f"SKIP entry at index {i} in {checksum_var or 'checksums'} — integrity verification bypassed",
                }
            )

    vcs_patterns = ["git+", "svn+", "hg+", "bzr+"]
    non_vcs_sources = [s for s in sources if not any(p in s for p in vcs_patterns)]

    if non_vcs_sources and len(sources) != len(checksums):
        findings.append(
            {
                "type": "MISMATCH",
                "severity": "HIGH",
                "message": f"Source count ({len(sources)}) does not match checksum count ({len(checksums)})",
            }
        )

    if checksum_var in WEAK_ALGORITHMS:
        findings.append(
            {
                "type": "WEAK_CHECKSUM",
                "severity": "MEDIUM",
                "message": f"Using {checksum_var} — consider a stronger algorithm (sha256sums or sha512sums)",
            }
        )

    if non_vcs_sources and checksum_var is None:
        findings.append(
            {
                "type": "MISSING_CHECKSUMS",
                "severity": "HIGH",
                "message": f"Non-VCS sources found without any checksums: {non_vcs_sources}",
            }
        )

    return findings


def main():
    if len(sys.argv) != 2:
        print(f"Usage: {sys.argv[0]} PKGBUILD")
        sys.exit(1)

    path = sys.argv[1]
    try:
        sources, checksums, checksum_var = parse_pkgbuild(path)
        findings = validate_checksums(sources, checksums, checksum_var)

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
