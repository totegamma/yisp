# Crypto Operators (`crypto.*`)

Cryptographic functions in YISP. All operators in this module require the `crypto.` prefix.

## `crypto.sha256`

Calculates the SHA-256 hash of a string.

**Syntax:**
```yaml
!yisp
- crypto.sha256
- string
```

**Example:**
```yaml
hash: !yisp
  - crypto.sha256
  - "hello world"
# Evaluates to: hash: "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
```