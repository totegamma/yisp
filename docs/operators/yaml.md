# YAML Operators (`yaml.*`)

YAML serialization operators in YISP. All operators in this module require the `yaml.` prefix.

## `yaml.marshal`

Converts a YISP value to a YAML string.

**Syntax:**
```yaml
!yisp
- yaml.marshal
- value
```

**Example:**
```yaml
yaml: !yisp
  - yaml.marshal
  - hoge: piyo
    a: 1
# Evaluates to: yaml: "hoge: piyo\na: 1\n"
```