---
name: pkgbuild-security-assessment
description: Conduct structured security assessment of Arch Linux AUR PKGBUILD files. Analyze declared intent vs actual behavior, detect obfuscated code, high-entropy strings, external resource interactions, and produce detailed risk verdicts. Use when reviewing AUR packages for security vulnerabilities or malicious code.
license: MIT
compatibility: Requires Python 3 for helper scripts.
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

Follow the nine-step procedure defined in [Assessment Procedure](references/PROCEDURE.md). The procedure covers file reading, structure parsing, checksum validation, entropy analysis, obfuscation detection, network auditing, privilege checks, supply chain checks, and VCS-specific checks. Each step indicates which helper script to run.

## Output Format

The output report must follow the format defined in [Output Format](references/OUTPUT_FORMAT.md).

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

This skill comes with helper scripts. Script details and definition is found in this [SCRIPTS document](references/SCRIPTS.md)
Incorporate their output into the findings section of the report. Script results alone are not sufficient — they miss context-dependent issues and multi-line patterns that require manual analysis.

## References

Do not load all reference files upfront. Read each file only when the assessment reaches the step or topic it covers.

- [Assessment Procedure](references/PROCEDURE.md) - Nine-step assessment methodology. Load first — it drives the entire assessment workflow.
- [Output Format](references/OUTPUT_FORMAT.md) - Report template for assessment output. Load when writing the final report.
- [Detailed Reference](references/REFERENCE.md) - PKGBUILD schema and best practices. Load when you need to verify whether a variable, function, or convention is standard.
- [Examples](references/EXAMPLES.md) - Positive and negative PKGBUILD examples. Load when comparing a PKGBUILD against known-good or known-bad patterns.
- [Red Flags Quick Reference](references/RED-FLAGS.md) - Common attack patterns and detection commands. Load during Steps 5-8 to cross-check findings.
- [Scripts Description](references/SCRIPTS.md) - Description of available scripts and how to use them. Load this document during any step if you need to run a command or script.
