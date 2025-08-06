# Type Operators (`types.*`)

Type checking and conversion operators in YISP. All operators in this module require the `types.` prefix.

## `types.assert`

Asserts that a value matches a specific type, raising an error if it doesn't.

**Syntax:**
```yaml
!yisp
- types.assert
- value
- type_schema
```

## `types.get`

Gets type information about a value.

**Syntax:**
```yaml
!yisp
- types.get
- value
```

## `types.of`

Returns the type of a value as a string.

**Syntax:**
```yaml
!yisp
- types.of
- value
```

**Example:**
```yaml
result: !yisp
  - types.of
  - 42
# Evaluates to: result: "int"
```

## Type Tags

YISP provides special type tags for parameters in lambda functions:

- `!string`: Indicates that the parameter should be a string
- `!int`: Indicates that the parameter should be an integer
- `!float`: Indicates that the parameter should be a floating-point number
- `!bool`: Indicates that the parameter should be a boolean

**Example:**
```yaml
!yisp &greet
- lambda
- [!string name, !int age]
- strings.concat
- "Hello, "
- *name
- "! You are "
- strings.concat
- *age
- " years old."
```