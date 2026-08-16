---
name: pkgbuild-security-assessment
description: Conduct structured security assessment of Arch Linux AUR PKGBUILD files. Analyze declared intent vs actual behavior, detect obfuscated code, high-entropy strings, external resource interactions, and produce detailed risk verdicts. Use when reviewing AUR packages for security vulnerabilities or malicious code.
license: MIT
metadata:
  author: auror-team
  version: "1.0"
  domain: security
  language: bash
  target: arch-linux-aur
allowed-tools: Bash(grep,find,cat,python3) Read Glob Grep
---

# PKGBUILD Security Assessment Skill

## Overview

Conduct a structured security assessment of Arch Linux AUR `PKGBUILD` files. This skill analyzes declared intent vs. actual behavior, detects obfuscated code, high-entropy strings, external resource interactions, and produces detailed risk verdicts.

## When to Use

- Reviewing AUR packages for security vulnerabilities
- Assessing PKGBUILD files before installation
- Detecting malicious or tampered packages
- Auditing package metadata and build scripts

## PKGBUILD Primer

A `PKGBUILD` is a `bash` script sourced by `makepkg`. Understanding the expected schema is the baseline for spotting anomalies.

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
| `source` | Array of source URIs or local filenames |
| `sha256sums` | Integrity checksums, one per `source` entry |
| `validpgpkeys` | Fingerprints for GPG signature verification |

### Lifecycle Functions

```
pkgver()   – dynamic version computation (VCS packages)
prepare()  – patch application, pre-build setup
build()    – compilation / transpilation
check()    – upstream test suite
package()  – install files into $pkgdir
```

### What a Legitimate PKGBUILD Should Contain

- Metadata variables matching reality (version, URL, license)
- `source` entries pointing to versioned upstream tarballs, VCS URLs, or local patches
- Checksums for every non-VCS source entry
- Build instructions using standard tools (`make`, `cmake`, `cargo`, `pip`, `meson`, …)
- `install` into `$pkgdir` only — never touching the live system
- No network access beyond what `makepkg` fetches via `source`

### What a Legitimate PKGBUILD Should NOT Contain

- Downloading additional files at build time with `curl`/`wget`/`fetch` outside of `source`
- Dynamic checksum computation that could be influenced by a remote endpoint
- Execution of unverified binaries fetched at build time
- Modification of files outside `$pkgdir` / `$srcdir`
- Obfuscated strings, base64-decoded payloads, or eval chains
- Hard-coded credentials, tokens, or private keys
- `sudo`, `su`, or privilege escalation
- Modifications to `/etc/`, `/usr/`, or any system path directly
- `post_install` hooks that download and execute arbitrary code
- Stripping GPG signature checks (`--no-check-certificate`, `--insecure`)

## Assessment Procedure

### Step 1 — Read the Full File

Read every line of the PKGBUILD verbatim. Do not skip comments; they sometimes contain encoded payloads or reveal intent.

### Step 2 — Parse Structure

Identify and list:

1. All declared variables and their values
2. All function definitions (names and bodies)
3. All external commands invoked (executables called via shell)
4. All network operations (URLs, hostnames, protocol schemes)
5. All file system writes (paths written, created, or modified)

### Step 3 — Checksum Integrity Check

- Count `source` array entries
- Count checksum array entries
- Flag `SKIP` entries — they bypass integrity verification entirely
- Flag mismatched array lengths
- Flag `sha1sums` or `md5sums` without a stronger algorithm alongside them

### Step 4 — Entropy Analysis

Scan every string literal, here-doc, and variable assignment for high-entropy content.

#### Shannon Entropy Algorithm

```
H(s) = -∑ (p_c × log₂(p_c))   for each unique character c in s
where p_c = count(c) / n
```

**Thresholds:**

| Range (bits/char) | Interpretation |
|---|---|
| 0.0 – 3.0 | Normal prose / shell keywords — ignore |
| 3.0 – 4.0 | Paths, version strings, URLs — review context |
| 4.0 – 4.5 | Borderline — flag only if context is suspicious |
| 4.5 – 5.0 | **HIGH** — likely encoded payload or key material; flag |
| > 5.0 | **CRITICAL** — near-random; almost certainly encoded/encrypted data |

Apply threshold only to strings of **≥ 20 characters**.

**Known-Good Exclusions:**

- Strings assigned to `sha256sums`, `sha512sums`, `b2sums`, `md5sums`, `sha1sums`
- Exactly 32, 40, 56, 64, or 128 hex characters (checksums)
- GPG fingerprints in `validpgpkeys`
- Well-formed UUIDs

**Flag These:**

- Strings passed to `eval`, `sh -c`, `bash -c`
- Strings decoded with `base64 -d`, `openssl enc -d`, `python -c`, `perl -e`
- Strings assigned to generic variable names (`_x`, `DATA`, `PAYLOAD`, `CMD`, `_0`)
- Hex or base64 literals inside `echo`/`printf` pipelines

### Step 5 — Obfuscation Detection

Flag these patterns:

| Pattern | Why Suspicious |
|---|---|
| `eval $(…)` / `eval "$(…)"` | Executes dynamically-constructed code |
| `base64 -d \| bash` | Classic one-liner dropper |
| `$(python -c '…')` / `$(perl -e '…')` | Inline interpreter abuse |
| Here-docs piped to `sh` or `bash` | Hidden script execution |
| Variable names like `_0x…`, `__`, single chars | Obfuscated variable naming |
| `tr`, `sed`, `rev`, `xxd` used on strings then executed | Runtime string construction |
| `IFS=` manipulation before command execution | Argument splitting tricks |

### Step 6 — Network Interaction Audit

For every URL or hostname found:

1. Record the full URI
2. Classify the protocol (`https`, `http`, `ftp`, `git+https`, `git+ssh`, …)
3. Note whether it appears in `source` (auditable) or is fetched at runtime (unauditable)
4. Flag `http://` without TLS
5. Flag IP-address-only endpoints
6. Flag uncommon TLDs or freshly-registered-looking domains
7. Flag `--no-check-certificate` / `--insecure` / `-k` flags on `curl`/`wget`
8. Flag data sent TO a remote (exfiltration): POST bodies, query strings

### Step 7 — Privilege and System Integrity Checks

Flag:

- `sudo`, `su`, `doas`, `pkexec` — escalation
- Writes to absolute paths outside `$pkgdir` / `$srcdir`
- `chmod +s` / `chown root` — SUID/SGID setting
- `/etc/cron*`, `/etc/systemd/`, `/etc/passwd`, `/etc/shadow` — system config tampering
- `iptables`, `nft`, `ufw` — firewall rule modification
- `systemctl enable`/`start` during build — premature service activation

### Step 8 — Supply Chain Checks

- Does the package declare itself to `provides` a widely-used package? (typosquatting / shadowing)
- Does `replaces` or `conflicts` remove a security tool or system package?
- Do `.install` hook scripts download additional payloads?
- Are `makedepends` pulling in packages not published in official repos?
- Is `pkgname` a near-homoglyph of a popular package?

### Step 9 — VCS Package Specifics

If `pkgver()` is present:

- Verify it only reads from the cloned source, not from a remote endpoint
- Flag network calls inside `pkgver()`
- Flag dynamic version strings that could be poisoned by a compromised upstream tag

## Output Format

The output report must follow the format defined here `assets/templates/report_template.md`

## Severity Definitions

| Level | Meaning |
| --- | --- |
| INFO | Noteworthy but not inherently dangerous; may be explained by legitimate use |
| LOW | Minor deviation from best practice; low exploitation potential |
| MEDIUM | Pattern that warrants manual review; could enable harm if combined with other issues |
| HIGH | Strong indicator of malicious or negligent behavior; do not install without resolution |
| CRITICAL | Confirmed or near-certain malicious code; do not install |

## Overall Risk Roll-Up

- Any CRITICAL finding → overall CRITICAL
- Any HIGH finding → overall at least HIGH
- Three or more MEDIUM findings → overall at least HIGH
- Only LOW / INFO → overall LOW
- Nothing flagged → overall LOW

## Scripts

This skill includes helper scripts in the `scripts/` directory:

- `scripts/entropy_scan.py` - High-entropy string detection
- `scripts/checksum_validator.py` - Checksum integrity verification
- `scripts/red_flags_detector.py` - Pattern matching for common attacks

## References

- [Detailed Reference](references/REFERENCE.md) - PKGBUILD schema and best practices
- [Examples](references/EXAMPLES.md) - Positive and negative PKGBUILD examples
- [Red Flags Quick Reference](references/RED-FLAGS.md) - Common attack patterns

## Common Red Flags at a Glance

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
