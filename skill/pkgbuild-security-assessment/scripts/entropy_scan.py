#!/usr/bin/env python3
"""
Quick entropy scanner for PKGBUILD string literals.
Usage: python3 entropy_scan.py PKGBUILD
"""

import math, re, sys, collections
from pathlib import Path

# Thresholds (bits/char)
MEDIUM_THRESHOLD = 4.0
HIGH_THRESHOLD = 4.5
CRIT_THRESHOLD = 5.0
MIN_LENGTH = 20

# Patterns that produce high entropy but are expected/benign
CHECKSUM_RE = re.compile(r"^[0-9a-fA-F]{32,128}$")
UUID_RE = re.compile(
    r"^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$"
)
# Variables whose values are always benign high-entropy data
BENIGN_VARS = {
    "sha256sums",
    "sha512sums",
    "b2sums",
    "md5sums",
    "sha1sums",
    "validpgpkeys",
}


def shannon(s: str) -> float:
    if not s:
        return 0.0
    freq = collections.Counter(s)
    n = len(s)
    return -sum((c / n) * math.log2(c / n) for c in freq.values())


def severity(h: float) -> str:
    if h >= CRIT_THRESHOLD:
        return "CRITICAL"
    if h >= HIGH_THRESHOLD:
        return "HIGH    "
    if h >= MEDIUM_THRESHOLD:
        return "MEDIUM  "
    return "INFO    "


def is_benign(s: str) -> bool:
    return bool(CHECKSUM_RE.match(s))


def scan(path: str):
    src = Path(path).read_text(errors="replace")
    lines = src.splitlines()

    # Track which array variable a line belongs to
    current_array_var = None
    findings = []

    for lineno, line in enumerate(lines, 1):
        # Detect array variable assignments (may span lines)
        arr_start = re.match(r"^\s*([a-zA-Z_][a-zA-Z0-9_]*)=\(", line)
        if arr_start:
            current_array_var = arr_start.group(1)
        if ")" in line and current_array_var:
            current_array_var = None  # reset after closing paren on same line

        # Extract all quoted string literals (single and double quoted)
        candidates = re.findall(r'"([^"]{20,})"', line)
        candidates += re.findall(r"'([^']{20,})'", line)

        for cand in candidates:
            if is_benign(cand):
                continue
            if current_array_var in BENIGN_VARS:
                continue
            h = shannon(cand)
            if h < MEDIUM_THRESHOLD:
                continue
            sev = severity(h)
            findings.append((lineno, h, sev, cand[:80]))

    if not findings:
        print("No high-entropy strings found.")
        return

    print(f"{'Line':>5}  {'Entropy':>7}  {'Severity':<10}  String")
    print("-" * 80)
    for lineno, h, sev, snip in sorted(findings, key=lambda x: -x[1]):
        print(f"{lineno:>5}  {h:>7.4f}  {sev:<10}  {snip}")


if __name__ == "__main__":
    if len(sys.argv) != 2:
        sys.exit(f"Usage: {sys.argv[0]} PKGBUILD")
    scan(sys.argv[1])
