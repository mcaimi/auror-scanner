# Assessment Procedure

Follow these nine steps in order. Run the helper scripts (see the Scripts section in SKILL.md) at the indicated steps to augment manual analysis.

## Step 1 — Read the Full File

Read every line of the PKGBUILD verbatim. Do not skip comments; they sometimes contain encoded payloads or reveal intent.

## Step 2 — Parse Structure

Identify and list:

1. All declared variables and their values
2. All function definitions (names and bodies)
3. All external commands invoked (executables called via shell)
4. All network operations (URLs, hostnames, protocol schemes)
5. All file system writes (paths written, created, or modified)

## Step 3 — Checksum Integrity Check

- Count `source` array entries
- Count checksum array entries
- Flag `SKIP` entries — they bypass integrity verification entirely
- Flag mismatched array lengths
- Flag `sha1sums` or `md5sums` without a stronger algorithm alongside them

Run `scripts/checksum_validator.py` to automate this step.

## Step 4 — Entropy Analysis

Scan every string literal, here-doc, and variable assignment for high-entropy content.

### Shannon Entropy Algorithm

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

Run `scripts/entropy_scan.py` to automate this step.

## Step 5 — Obfuscation Detection

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

## Step 6 — Network Interaction Audit

For every URL or hostname found:

1. Record the full URI
2. Classify the protocol (`https`, `http`, `ftp`, `git+https`, `git+ssh`, …)
3. Note whether it appears in `source` (auditable) or is fetched at runtime (unauditable)
4. Flag `http://` without TLS
5. Flag IP-address-only endpoints
6. Flag uncommon TLDs or freshly-registered-looking domains
7. Flag `--no-check-certificate` / `--insecure` / `-k` flags on `curl`/`wget`
8. Flag data sent TO a remote (exfiltration): POST bodies, query strings

## Step 7 — Privilege and System Integrity Checks

Flag:

- `sudo`, `su`, `doas`, `pkexec` — escalation
- Writes to absolute paths outside `$pkgdir` / `$srcdir`
- `chmod +s` / `chown root` — SUID/SGID setting
- `/etc/cron*`, `/etc/systemd/`, `/etc/passwd`, `/etc/shadow` — system config tampering
- `iptables`, `nft`, `ufw` — firewall rule modification
- `systemctl enable`/`start` during build — premature service activation

## Step 8 — Supply Chain Checks

- Does the package declare itself to `provides` a widely-used package? (typosquatting / shadowing)
- Does `replaces` or `conflicts` remove a security tool or system package?
- Do `.install` hook scripts download additional payloads?
- Are `makedepends` pulling in packages not published in official repos?
- Is `pkgname` a near-homoglyph of a popular package?

Run `scripts/red_flags_detector.py` to automate Steps 5–8.

## Step 9 — VCS Package Specifics

If `pkgver()` is present:

- Verify it only reads from the cloned source, not from a remote endpoint
- Flag network calls inside `pkgver()`
- Flag dynamic version strings that could be poisoned by a compromised upstream tag
