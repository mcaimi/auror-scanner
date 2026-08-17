This skill includes helper scripts in the `scripts/` directory. Run them during the assessment to augment manual analysis. Each script takes a single argument: the path to the PKGBUILD file.

### `scripts/checksum_validator.py` — Checksum Integrity (Step 3)

Validates that `source` entries have matching checksums. Detects SKIP entries, source/checksum count mismatches, missing checksums on non-VCS sources, and weak algorithms (md5sums, sha1sums).

```bash
python3 scripts/checksum_validator.py /path/to/PKGBUILD
```

Output: per-finding severity and description, or "Checksum validation passed." if clean.

### `scripts/entropy_scan.py` — High-Entropy String Detection (Step 4)

Scans string literals for high Shannon entropy. Automatically excludes known-benign values (checksums in checksum arrays, hex strings matching standard checksum lengths, UUIDs). Flags strings >= 20 characters above 4.0 bits/char as MEDIUM, 4.5 as HIGH, 5.0 as CRITICAL.

```bash
python3 scripts/entropy_scan.py /path/to/PKGBUILD
```

Output: table sorted by descending entropy with line number, entropy value, severity, and string snippet.

### `scripts/red_flags_detector.py` — Attack Pattern Detection (Steps 5-8)

Pattern-matches against 16 known attack signatures covering remote code execution, obfuscated payloads, privilege escalation, system config modification, data exfiltration, credential exposure, SUID installation, checksum bypass, insecure transport, premature service activation, package shadowing, and writes to system paths.

```bash
python3 scripts/red_flags_detector.py /path/to/PKGBUILD
```

Output: findings grouped by severity (CRITICAL, HIGH, MEDIUM) with line numbers and matched text.

### Recommended Execution Order

Scripts are usually requested during the 9-step procedure but a sensible default order is this:

1- Execute checksum_validator

```bash
python3 scripts/checksum_validator.py PKGBUILD
```

2- Execute entropy_scan

```bash
python3 scripts/entropy_scan.py PKGBUILD

```

3- Execute red_flags_detector

```bash
python3 scripts/red_flags_detector.py PKGBUILD
```
