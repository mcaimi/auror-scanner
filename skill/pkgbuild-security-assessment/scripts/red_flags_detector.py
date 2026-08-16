#!/usr/bin/env python3
"""
Red flags detector for PKGBUILD files.
Detects common attack patterns and suspicious behaviors.
Usage: python3 red_flags_detector.py PKGBUILD
"""

import re
import sys
from pathlib import Path

# Common attack patterns
ATTACK_PATTERNS = [
    {
        "name": "Remote Code Execution via curl",
        "pattern": r"curl\s+.*\|\s*(bash|sh|zsh|ksh)",
        "severity": "CRITICAL",
        "description": "Downloads and executes remote code",
    },
    {
        "name": "Remote Code Execution via wget",
        "pattern": r"wget\s+.*\|\s*(bash|sh|zsh|ksh)",
        "severity": "CRITICAL",
        "description": "Downloads and executes remote code",
    },
    {
        "name": "Eval with base64",
        "pattern": r"eval\s+.*base64.*-d",
        "severity": "CRITICAL",
        "description": "Obfuscated code execution via base64 decoding",
    },
    {
        "name": "Python inline execution",
        "pattern": r'\$\([^)]*python[^)]*-\s*c\s+[\'"]',
        "severity": "HIGH",
        "description": "Inline Python code execution",
    },
    {
        "name": "Perl inline execution",
        "pattern": r'\$\([^)]*perl[^)]*-\s*e\s+[\'"]',
        "severity": "HIGH",
        "description": "Inline Perl code execution",
    },
    {
        "name": "SUID binary installation",
        "pattern": r"install\s+.*-m4755|install\s+.*-m755.*chmod\s+\+s",
        "severity": "HIGH",
        "description": "Installing SUID binary",
    },
    {
        "name": "chmod SUID/SGID",
        "pattern": r"chmod\s+\+s",
        "severity": "HIGH",
        "description": "Setting SUID/SGID bit",
    },
    {
        "name": "Privilege escalation via sudo",
        "pattern": r"\bsudo\b|\bsu\b|\bdoas\b|\bpkexec\b",
        "severity": "CRITICAL",
        "description": "Privilege escalation attempt",
    },
    {
        "name": "System config modification",
        "pattern": r"/etc/(passwd|shadow|sudoers|cron|systemd)",
        "severity": "CRITICAL",
        "description": "Modifying system configuration files",
    },
    {
        "name": "Insecure curl/wget",
        "pattern": r"curl\s+.*(--insecure|--no-check-certificate|-k)",
        "severity": "MEDIUM",
        "description": "TLS verification disabled",
    },
    {
        "name": "HTTP source (no TLS)",
        "pattern": r"source\s*=\s*\([^)]*http://[^)]*\)",
        "severity": "MEDIUM",
        "description": "Using unencrypted HTTP for sources",
    },
    {
        "name": "Checksum bypass",
        "pattern": r"sha256sums\s*=\s*\(\s*\'SKIP\'\s*",
        "severity": "HIGH",
        "description": "Integrity check bypassed",
    },
    {
        "name": "System info exfiltration",
        "pattern": r"curl\s+-X\s+POST.*\$(uname|hostname|whoami)",
        "severity": "CRITICAL",
        "description": "Sending system information to remote server",
    },
    {
        "name": "Premature service activation",
        "pattern": r"systemctl\s+enable",
        "severity": "HIGH",
        "description": "Enabling system service during build",
    },
    {
        "name": "Package shadowing",
        "pattern": r'provides\s*=\s*\(\s*[\'"]openssl[\'"]',
        "severity": "HIGH",
        "description": "Shadowing critical system package",
    },
    {
        "name": "Credential exposure",
        "pattern": r'(password|secret|token|api_key)\s*=\s*[\'"][^\'"]+[\'"]',
        "severity": "CRITICAL",
        "description": "Potential credential exposure",
    },
    {
        "name": "Write to system paths",
        "pattern": r"install\s+.*\s+(/usr/|/etc/|/bin/|/sbin/)",
        "severity": "HIGH",
        "description": "Writing to system paths outside $pkgdir",
    },
]


def scan_pkgbuild(path: str) -> list:
    """Scan PKGBUILD for attack patterns."""
    src = Path(path).read_text(errors="replace")
    findings = []

    for pattern_info in ATTACK_PATTERNS:
        matches = re.finditer(pattern_info["pattern"], src, re.IGNORECASE)
        for match in matches:
            lineno = src[: match.start()].count("\n") + 1
            findings.append(
                {
                    "pattern_name": pattern_info["name"],
                    "severity": pattern_info["severity"],
                    "description": pattern_info["description"],
                    "line_number": lineno,
                    "matched_text": match.group()[:100],
                }
            )

    return findings


def main():
    if len(sys.argv) != 2:
        print(f"Usage: {sys.argv[0]} PKGBUILD")
        sys.exit(1)

    path = sys.argv[1]
    try:
        findings = scan_pkgbuild(path)

        if not findings:
            print("No red flags detected.")
        else:
            # Group by severity
            severity_order = {"CRITICAL": 0, "HIGH": 1, "MEDIUM": 2, "LOW": 3}
            sorted_findings = sorted(
                findings, key=lambda x: severity_order.get(x["severity"], 4)
            )

            print(f"Found {len(findings)} red flag(s):")
            print()

            for finding in sorted_findings:
                print(f"[{finding['severity']}] {finding['pattern_name']}")
                print(f"  Line {finding['line_number']}: {finding['description']}")
                print(f"  Match: {finding['matched_text']}")
                print()
    except FileNotFoundError:
        print(f"Error: File not found: {path}")
        sys.exit(1)
    except Exception as e:
        print(f"Error: {e}")
        sys.exit(1)


if __name__ == "__main__":
    main()
