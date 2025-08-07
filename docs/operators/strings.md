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
```

## `strings.contains`

Checks if a string contains a substring.

**Syntax:**
```yaml
!yisp
- strings.contains
- haystack
- needle
```

**Example:**
```yaml
found: !yisp
  - strings.contains
  - "hello world"
  - "world"
# Evaluates to: found: true
```

## `strings.join`

Joins a list of strings with a separator.

**Syntax:**
```yaml
!yisp
- strings.join
- separator
- [string1, string2, ...]
```

**Example:**
```yaml
result: !yisp
  - strings.join
  - ", "
  - ["apple", "banana", "cherry"]
# Evaluates to: result: "apple, banana, cherry"
```

## `strings.replace`

Replaces all occurrences of a substring with another string.

**Syntax:**
```yaml
!yisp
- strings.replace
- original_string
- old_substring
- new_substring
```

**Example:**
```yaml
result: !yisp
  - strings.replace
  - "hello world"
  - "world"
  - "YISP"
# Evaluates to: result: "hello YISP"
```

## `strings.split`

Splits a string by a separator into a list of strings.

**Syntax:**
```yaml
!yisp
- strings.split
- string
- separator
```

**Example:**
```yaml
parts: !yisp
  - strings.split
  - "apple,banana,cherry"
  - ","
# Evaluates to: parts: ["apple", "banana", "cherry"]
```

## `strings.toUpper`

Converts a string to uppercase.

**Syntax:**
```yaml
!yisp
- strings.toUpper
- string
```

**Example:**
```yaml
upper: !yisp
  - strings.toUpper
  - "hello world"
# Evaluates to: upper: "HELLO WORLD"
```

## `strings.toLower`

Converts a string to lowercase.

**Syntax:**
```yaml
!yisp
- strings.toLower
- string
```

**Example:**
```yaml
lower: !yisp
  - strings.toLower
  - "HELLO WORLD"
# Evaluates to: lower: "hello world"
```

## `strings.trim`

Removes whitespace from the beginning and end of a string.

**Syntax:**
```yaml
!yisp
- strings.trim
- string
```

**Example:**
```yaml
trimmed: !yisp
  - strings.trim
  - "  hello world  "
# Evaluates to: trimmed: "hello world"
```
