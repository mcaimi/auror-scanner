# PKGBUILD Security Assessment Report Template

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

(Table: source entry | checksum present | algorithm | SKIP flag | notes)

---

## Line-by-Line Annotation

For every logical block:

### Line <N>–<M>: <Short Label>

    ```bash
    <verbatim code>
    ```

**What it does:** <Plain-English explanation>  
**Interactions:** <Any external service, file path, or binary involved>  
**Security note:** <NONE / or finding with severity [INFO / LOW / MEDIUM / HIGH / CRITICAL]>

---

## Findings

### Finding F-<NNN>: <Title>

| Field    | Value |
|---|---|
| Severity | INFO / LOW / MEDIUM / HIGH / CRITICAL |
| Lines    | <line numbers> |
| Category | Obfuscation / Network / Privilege / Supply-Chain / Integrity / Entropy |

**Description:** <What the code does and why it is suspicious>  
**Evidence:** <Verbatim snippet>  
**Recommendation:** <What a package maintainer should do instead>

---

## Risk Verdict

**Overall Risk: <LEVEL>**

(Two to four sentences summarising the key risk drivers and whether the package should be trusted as-is, reviewed further, or rejected.)
```
