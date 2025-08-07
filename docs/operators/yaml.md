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

## `yaml.unmarshal`

Parses a YAML string into a structured data object.

**Syntax:**
```yaml
!yisp
- yaml.unmarshal
- yaml_string
```

**Example:**
```yaml
data: !yisp
  - yaml.unmarshal
  - "name: Alice\nage: 30"
# Evaluates to: data: {name: "Alice", age: 30}
```