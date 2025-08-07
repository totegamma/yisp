# String Operators (`strings.*`)

String manipulation operators in YISP. All operators in this module require the `strings.` prefix.

## `strings.concat`

Concatenates two or more strings together.

**Syntax:**
```yaml
!yisp
- strings.concat
- string1
- string2
- ...
```

**Example:**
```yaml
greeting: !yisp
  - strings.concat
  - "Hello, "
  - "world!"
# Evaluates to: greeting: "Hello, world!"
```

## `strings.format`

Formats a string using printf-style formatting.

**Syntax:**
```yaml
!yisp
- strings.format
- format_string
- arg1
- arg2
- ...
```

**Example:**
```yaml
message: !yisp
  - strings.format
  - "User %s has %d points"
  - "Alice"
  - 150
# Evaluates to: message: "User Alice has 150 points"
```

## `strings.escape`

Escapes a value for safe use in strings, converting it to a string representation.

**Syntax:**
```yaml
!yisp
- strings.escape
- value
```

**Example:**
```yaml
escaped: !yisp
  - strings.escape
  - 42
# Evaluates to: escaped: "42"
```

## `strings.sha256`

Calculates the SHA-256 hash of a string.

**Syntax:**
```yaml
!yisp
- strings.sha256
- string
```

**Example:**
```yaml
hash: !yisp
  - strings.sha256
  - "hello world"
# Evaluates to: hash: "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
`
