# PKGBUILD Security Assessment Skill

## Purpose

Conduct a structured security assessment of an Arch Linux AUR `PKGBUILD` file. Produce a line-by-line annotated report covering: declared intent vs. actual behavior, suspicious patterns, obfuscated code, high-entropy strings, external resource interactions, and a final risk verdict.

---

## PKGBUILD Primer

A `PKGBUILD` is a `bash` script sourced by `makepkg`. It defines metadata variables and lifecycle functions. Understanding the expected schema is the baseline for spotting anomalies.

### Required Variables

| Variable | Purpose |
|---|---|
| `pkgname` | Package name (must match directory name) |
| `pkgver` | Upstream version string |
| `pkgrel` | Arch packaging release number (integer ≥ 1) |
| `arch` | Target architectures (`x86_64`, `aarch64`, `any`, …) |

### Common Optional Variables

| Variable | Notes |
|---|---|
| `pkgdesc` | One-line description |
| `url` | Upstream project URL |
| `license` | SPDX identifier(s) |
| `depends` | Runtime dependencies |
| `makedepends` | Build-time-only dependencies |
| `optdepends` | Optional dependencies with descriptions |
| `provides` / `conflicts` / `replaces` | Virtual package relationships |
| `source` | Array of source URIs or local filenames |
| `sha256sums` / `sha512sums` / `b2sums` | Integrity checksums, one per `source` entry |
| `validpgpkeys` | Fingerprints for GPG signature verification |
| `options` | Build flags (`!strip`, `!debug`, …) |
| `install` | Path to an `.install` hook script |
| `backup` | Files to preserve on upgrade |

### Lifecycle Functions (in execution order)

```
pkgver()   – dynamic version computation (VCS packages)
prepare()  – patch application, pre-build setup
build()    – compilation / transpilation
check()    – upstream test suite
package()  – install files into $pkgdir
```

There may also be split-package `package_<name>()` functions for multi-package builds.

### What a Legitimate PKGBUILD Should Contain

- Metadata variables matching reality (version, URL, license).
- `source` entries pointing to versioned upstream tarballs, VCS URLs, or local patches.
- Checksums for every non-VCS source entry.
- Build instructions using standard tools (`make`, `cmake`, `cargo`, `pip`, `meson`, …).
- `install` into `$pkgdir` only — never touching the live system.
- No network access beyond what `makepkg` fetches via `source`.

### What a Legitimate PKGBUILD Should NOT Contain

- Downloading additional files at build time with `curl`/`wget`/`fetch` outside of `source`.
- Dynamic checksum computation that could be influenced by a remote endpoint.
- Execution of unverified binaries fetched at build time.
- Modification of files outside `$pkgdir` / `$srcdir`.
- Obfuscated strings, base64-decoded payloads, or eval chains.
- Hard-coded credentials, tokens, or private keys.
- `sudo`, `su`, or privilege escalation.
- Modifications to `/etc/`, `/usr/`, or any system path directly.
- `post_install` hooks that download and execute arbitrary code.
- Stripping GPG signature checks (`--no-check-certificate`, `--insecure`).

---

## Assessment Procedure

### Step 1 — Read the Full File

Read every line of the PKGBUILD verbatim. Do not skip comments; they sometimes contain encoded payloads or reveal intent.

### Step 2 — Parse Structure

Identify and list:
1. All declared variables and their values.
2. All function definitions (names and bodies).
3. All external commands invoked (executables called via shell).
4. All network operations (URLs, hostnames, protocol schemes).
5. All file system writes (paths written, created, or modified).

### Step 3 — Checksum Integrity Check

- Count `source` array entries.
- Count checksum array entries.
- Flag `SKIP` entries — they bypass integrity verification entirely.
- Flag mismatched array lengths.
- Flag `sha1sums` or `md5sums` without a stronger algorithm alongside them.

### Step 4 — Entropy Analysis

Scan every string literal, here-doc, and variable assignment for high-entropy content. Apply the Shannon entropy algorithm below to every candidate string before flagging it.

#### Shannon Entropy Algorithm

Shannon entropy measures information density. For a string *s* of length *n* over alphabet *Σ*:

```
H(s) = -∑ (p_c × log₂(p_c))   for each unique character c in s
where p_c = count(c) / n
```

Range: 0 bits/char (all chars identical) → log₂(|Σ|) bits/char (perfectly uniform).

**Reference script** — run this on any extracted candidate string:

```python
import math, collections, sys

def shannon_entropy(s: str) -> float:
    if not s:
        return 0.0
    freq = collections.Counter(s)
    n = len(s)
    return -sum((c / n) * math.log2(c / n) for c in freq.values())

for line in sys.stdin:
    s = line.rstrip('\n')
    if s:
        print(f"{shannon_entropy(s):.4f}  {s[:120]}")
```

Usage during assessment:

```bash
# Extract all quoted strings from the PKGBUILD, one per line, then score:
grep -oP '(?<=")[^"]{20,}(?=")' PKGBUILD | python3 entropy.py
grep -oP "(?<=')[^']{20,}(?=')" PKGBUILD | python3 entropy.py
```

#### Entropy Thresholds

| Range (bits/char) | Interpretation |
|---|---|
| 0.0 – 3.0 | Normal prose / shell keywords — ignore |
| 3.0 – 4.0 | Paths, version strings, URLs — review context |
| 4.0 – 4.5 | Borderline — flag only if context is suspicious |
| 4.5 – 5.0 | **HIGH** — likely encoded payload or key material; flag |
| > 5.0 | **CRITICAL** — near-random; almost certainly encoded/encrypted data |

Apply the threshold only to strings of **≥ 20 characters**. Shorter strings produce unreliable entropy estimates.

Alternatively, a more thorough but complex entropy estimator is provided below: Save as `entropy_scan.py` and run it against any PKGBUILD to surface high-entropy candidates automatically. This approach is useful to scan an entire PKGBUILD without grep-ing values.

Use this approach if the provided PKGBUILD is more than 100 lines long.

```python
#!/usr/bin/env python3
"""
Quick entropy scanner for PKGBUILD string literals.
Usage: python3 entropy_scan.py PKGBUILD
"""
import math, re, sys, collections
from pathlib import Path

# Thresholds (bits/char)
MEDIUM_THRESHOLD = 4.0
HIGH_THRESHOLD   = 4.5
CRIT_THRESHOLD   = 5.0
MIN_LENGTH       = 20

# Patterns that produce high entropy but are expected/benign
CHECKSUM_RE = re.compile(r'^[0-9a-fA-F]{32,128}$')
UUID_RE     = re.compile(r'^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$')
# Variables whose values are always benign high-entropy data
BENIGN_VARS = {'sha256sums', 'sha512sums', 'b2sums', 'md5sums', 'sha1sums', 'validpgpkeys'}

def shannon(s: str) -> float:
    if not s:
        return 0.0
    freq = collections.Counter(s)
    n = len(s)
    return -sum((c / n) * math.log2(c / n) for c in freq.values())

def severity(h: float) -> str:
    if h >= CRIT_THRESHOLD:   return 'CRITICAL'
    if h >= HIGH_THRESHOLD:   return 'HIGH    '
    if h >= MEDIUM_THRESHOLD: return 'MEDIUM  '
    return 'INFO    '

def is_benign(s: str) -> bool:
    return bool(CHECKSUM_RE.match(s) or UUID_RE.match(s))

def scan(path: str):
    src = Path(path).read_text(errors='replace')
    lines = src.splitlines()

    # Track which array variable a line belongs to
    current_array_var = None
    findings = []

    for lineno, line in enumerate(lines, 1):
        # Detect array variable assignments (may span lines)
        arr_start = re.match(r'^\s*([a-zA-Z_][a-zA-Z0-9_]*)=\(', line)
        if arr_start:
            current_array_var = arr_start.group(1)
        if ')' in line and current_array_var:
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

if __name__ == '__main__':
    if len(sys.argv) != 2:
        sys.exit(f"Usage: {sys.argv[0]} PKGBUILD")
    scan(sys.argv[1])
```

**Interpreting output of the script `entropy_scan.py`:**

- `CRITICAL` (> 5.0) — treat as confirmed obfuscation until proven otherwise; raise Finding F-xxx with severity CRITICAL.
- `HIGH` (4.5 – 5.0) — file as HIGH finding; cross-check with Step 5 (Obfuscation) and Step 6 (Network).
- `MEDIUM` (4.0 – 4.5) — note in the line annotation; escalate to HIGH if the string is passed to a shell interpreter.
- `INFO` (< 4.0) — record context only if other indicators are present.

#### Known-Good Exclusions

Do **not** flag strings that match all of the following:
- Assigned directly to `sha256sums`, `sha512sums`, `b2sums`, `md5sums`, or `sha1sums`.
- Are exactly 32, 40, 56, 64, or 128 hex characters (MD5/SHA1/SHA224/SHA256/SHA512 digests).
- Appear in `validpgpkeys` (40 or 160 hex chars, GPG fingerprints).
- Are well-formed UUIDs (`[0-9a-f]{8}-[0-9a-f]{4}-…`).
- Are standard base-64-encoded values assigned to a documented environment variable with a known non-secret purpose.

Flag all other high-entropy strings, especially:

- Strings passed to `eval`, `sh -c`, `bash -c`, or piped into a shell.
- Strings decoded with `base64 -d`, `openssl enc -d`, `python -c`, or `perl -e`.
- Strings assigned to generic variable names (`_x`, `DATA`, `PAYLOAD`, `CMD`, `_0`, …).
- Strings embedded in here-docs that are then executed.
- Hex or base64 literals inside `echo`/`printf` pipelines.

### Step 5 — Obfuscation Detection

Flag any of the following patterns:

| Pattern | Why Suspicious |
|---|---|
| `eval $(…)` / `eval "$(…)"` | Executes dynamically-constructed code |
| `base64 -d \| bash` or equivalent | Classic one-liner dropper |
| `$(python -c '…')` / `$(perl -e '…')` | Inline interpreter abuse |
| Here-docs piped to `sh` or `bash` | Hidden script execution |
| Variable names resembling `_0x…`, `__`, single chars | Obfuscated variable naming |
| `tr`, `sed`, `rev`, `xxd` used on strings then executed | Runtime string construction |
| `IFS=` manipulation before command execution | Argument splitting tricks |

### Step 6 — Network Interaction Audit

For every URL or hostname found:

1. Record the full URI.
2. Classify the protocol (`https`, `http`, `ftp`, `git+https`, `git+ssh`, …).
3. Note whether it appears in `source` (auditable) or is fetched at runtime inside a function (unauditable by `makepkg`).
4. Flag `http://` without TLS.
5. Flag IP-address-only endpoints (bypasses DNS-based blocking).
6. Flag uncommon TLDs or freshly-registered-looking domains.
7. Flag `--no-check-certificate` / `--insecure` / `-k` flags on `curl`/`wget`.
8. Flag data sent TO a remote (exfiltration): POST bodies, query strings containing system info.

### Step 7 — Privilege and System Integrity Checks

Flag:

- `sudo`, `su`, `doas`, `pkexec` — escalation.
- Writes to absolute paths outside `$pkgdir` / `$srcdir`.
- `chmod +s` / `chown root` — SUID/SGID setting.
- `/etc/cron*`, `/etc/systemd/`, `/etc/passwd`, `/etc/shadow` — system config tampering.
- `iptables`, `nft`, `ufw` — firewall rule modification.
- `systemctl enable`/`start` during build — premature service activation.

### Step 8 — Supply Chain Checks

- Does the package declare itself to `provides` a widely-used package? (typosquatting / shadowing)
- Does `replaces` or `conflicts` remove a security tool or system package?
- Do `.install` hook scripts download additional payloads?
- Are `makedepends` pulling in packages not published in official repos?
- Is `pkgname` a near-homoglyph of a popular package?

### Step 9 — VCS Package Specifics

If `pkgver()` is present:

- Verify it only reads from the cloned source, not from a remote endpoint.
- Flag network calls inside `pkgver()`.
- Flag dynamic version strings that could be poisoned by a compromised upstream tag.

---

## Output Format

Produce the report in the following structure:

```markdown
# PKGBUILD Security Assessment: <pkgname>

## Summary

| Field         | Value                        |
|---|---|
| File          | <path or "provided inline">  |
| pkgname       | …                            |
| pkgver        | …                            |
| pkgrel        | …                            |
| Assessed      | <ISO-8601 date>              |
| Overall Risk  | LOW / MEDIUM / HIGH / CRITICAL |

---

## Metadata Analysis

<List each declared variable, its value, and whether it looks correct.>

---

## Source & Checksum Integrity

<Table: source entry | checksum present | algorithm | SKIP flag | notes>

---

## Line-by-Line Annotation

For every logical block (variable assignment, function definition, individual statement):

### Line <N>–<M>: <Short Label>

```bash
<verbatim code>
```

**What it does:** <Plain-English explanation of the operation — inputs, outputs, side effects.>  
**Interactions:** <Any external service, file path, or binary involved.>  
**Security note:** <NONE / or finding with severity [INFO / LOW / MEDIUM / HIGH / CRITICAL]>

---

## Findings

### Finding F-<NNN>: <Title>

| Field    | Value |
|---|---|
| Severity | INFO / LOW / MEDIUM / HIGH / CRITICAL |
| Lines    | <line numbers> |
| Category | Obfuscation / Network / Privilege / Supply-Chain / Integrity / Entropy |

**Description:** <What the code does and why it is suspicious.>  
**Evidence:** <Verbatim snippet.>  
**Recommendation:** <What a package maintainer should do instead.>

---

## Risk Verdict

**Overall Risk: <LEVEL>**

<Two to four sentences summarising the key risk drivers and whether the package should be trusted as-is, reviewed further, or rejected.>
```

### Severity Definitions

| Level | Meaning |
|---|---|
| INFO | Noteworthy but not inherently dangerous; may be explained by legitimate use. |
| LOW | Minor deviation from best practice; low exploitation potential. |
| MEDIUM | Pattern that warrants manual review; could enable harm if combined with other issues. |
| HIGH | Strong indicator of malicious or negligent behavior; do not install without resolution. |
| CRITICAL | Confirmed or near-certain malicious code; do not install. |

### Overall Risk Roll-Up

- Any CRITICAL finding → overall CRITICAL.
- Any HIGH finding → overall at least HIGH.
- Three or more MEDIUM findings → overall at least HIGH.
- Only LOW / INFO → overall LOW.
- Nothing flagged → overall LOW.

---

## Reference: Common Red Flags at a Glance

```
curl … | bash                          — remote code execution
wget -qO- … | sh                       — remote code execution  
eval "$(base64 -d <<<'…')"             — obfuscated dropper
install -m4755 …                       — SUID binary installation
chmod +s …                             — SUID/SGID bit
source=(… http:// …)                   — insecure transport
sha256sums=('SKIP' …)                  — bypassed integrity check
curl … --insecure / -k                 — TLS verification disabled
curl -X POST … -d "$(uname -a)"        — system info exfiltration
systemctl enable … (in build/package)  — premature service activation
echo "…" >> /etc/sudoers               — privilege escalation
$(python3 -c "exec(__import__('base64').b64decode(…))")  — obfuscation
provides=('openssl')                   — shadowing critical system package
```

---

## Prompt Template

When invoking this skill, supply the PKGBUILD content in the user message and use the following system prompt structure:

```
You are a security analyst specializing in Arch Linux packaging.

You have been given a PKGBUILD file to assess. Follow the PKGBUILD Security Assessment skill exactly:

1. Parse and annotate every line.
2. Apply all detection steps (checksum integrity, entropy, obfuscation, network, privilege, supply chain, VCS).
3. Produce findings with severity ratings.
4. Output the complete markdown report using the prescribed format.

Do not summarize or skip sections. If a section has nothing to report, write "Nothing to report."

PKGBUILD content:
---
<paste PKGBUILD here>
---
```
