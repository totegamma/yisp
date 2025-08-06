# Built-in Operators

Built-in operators are always available in YISP and don't require any module prefix.

## Arithmetic Operators

### `+` (Addition)

Adds two or more numbers together.

**Syntax:**
```yaml
!yisp
- +
- arg1
- arg2
- ...
```

**Example:**
```yaml
result: !yisp
  - +
  - 1
  - 2
  - 3
# Evaluates to: result: 6
```

### `-` (Subtraction)

Subtracts numbers from the first argument.

**Syntax:**
```yaml
!yisp
- -
- base
- arg1
- arg2
- ...
```

**Example:**
```yaml
result: !yisp
  - -
  - 10
  - 3
  - 2
# Evaluates to: result: 5
```

### `*` (Multiplication)

Multiplies two or more numbers together.

**Syntax:**
```yaml
!yisp
- *
- arg1
- arg2
- ...
```

**Example:**
```yaml
result: !yisp
  - *
  - 2
  - 3
  - 4
# Evaluates to: result: 24
```

### `/` (Division)

Divides the first argument by the remaining arguments.

**Syntax:**
```yaml
!yisp
- /
- dividend
- divisor1
- divisor2
- ...
```

**Example:**
```yaml
result: !yisp
  - /
  - 24
  - 2
  - 3
# Evaluates to: result: 4
```

## Comparison Operators

### `==` (Equal)

Checks if two values are equal. Works with numbers, strings, and booleans.

**Syntax:**
```yaml
!yisp
- ==
- value1
- value2
```

**Example:**
```yaml
isEqual: !yisp
  - ==
  - 5
  - 5
# Evaluates to: isEqual: true
```

### `!=` (Not Equal)

Checks if two values are not equal.

**Syntax:**
```yaml
!yisp
- !=
- value1
- value2
```

**Example:**
```yaml
isNotEqual: !yisp
  - !=
  - 5
  - 10
# Evaluates to: isNotEqual: true
```

### `<` (Less Than)

Checks if the first value is less than the second. Works with numbers.

**Syntax:**
```yaml
!yisp
- <
- value1
- value2
```

**Example:**
```yaml
isLess: !yisp
  - <
  - 5
  - 10
# Evaluates to: isLess: true
```

### `<=` (Less Than or Equal)

Checks if the first value is less than or equal to the second.

**Syntax:**
```yaml
!yisp
- <=
- value1
- value2
```

**Example:**
```yaml
isLessOrEqual: !yisp
  - <=
  - 5
  - 5
# Evaluates to: isLessOrEqual: true
```

### `>` (Greater Than)

Checks if the first value is greater than the second.

**Syntax:**
```yaml
!yisp
- >
- value1
- value2
```

**Example:**
```yaml
isGreater: !yisp
  - >
  - 10
  - 5
# Evaluates to: isGreater: true
```

### `>=` (Greater Than or Equal)

Checks if the first value is greater than or equal to the second.

**Syntax:**
```yaml
!yisp
- >=
- value1
- value2
```

**Example:**
```yaml
isGreaterOrEqual: !yisp
  - >=
  - 5
  - 5
# Evaluates to: isGreaterOrEqual: true
```

## Logical Operators

### `and` (Logical AND)

Performs a logical AND operation on all arguments. Returns true if all arguments are truthy.

**Syntax:**
```yaml
!yisp
- and
- arg1
- arg2
- ...
```

**Example:**
```yaml
result: !yisp
  - and
  - true
  - true
  - true
# Evaluates to: result: true
```

### `or` (Logical OR)

Performs a logical OR operation on all arguments. Returns true if any argument is truthy.

**Syntax:**
```yaml
!yisp
- or
- arg1
- arg2
- ...
```

**Example:**
```yaml
result: !yisp
  - or
  - false
  - true
  - false
# Evaluates to: result: true
```

### `not` (Logical NOT)

Performs a logical NOT operation on the argument.

**Syntax:**
```yaml
!yisp
- not
- arg
```

**Example:**
```yaml
result: !yisp
  - not
  - false
# Evaluates to: result: true
```

## Null Coalesce Operators

### `??` / `default`

Returns the first non-null value from the arguments.

**Syntax:**
```yaml
!yisp
- ??
- value1
- value2
- ...
```

**Example:**
```yaml
result: !yisp
  - ??
  - null
  - "default value"
  - "fallback"
# Evaluates to: result: "default value"
```

## Conditional Operators

### `if`

Evaluates a condition and returns one of two values based on the result.

**Syntax:**
```yaml
!yisp
- if
- condition
- true_value
- false_value
```

**Example:**
```yaml
result: !yisp
  - if
  - !yisp [<, 5, 10]
  - "Less"
  - "Greater or Equal"
# Evaluates to: result: "Less"
```

## Special Operators

### `include`

Includes and evaluates files, returning their results as a list.

**Syntax:**
```yaml
!yisp
- include
- "path/to/file1.yaml"
- "path/to/file2.yaml"
- ...
```

**Example:**
```yaml
results: !yisp
  - include
  - "config/database.yaml"
  - "config/server.yaml"
```

### `import`

Imports modules, making their definitions available in the current environment.

**Syntax:**
```yaml
!yisp
- import
- ["module_name", "path/to/module.yaml"]
- ...
```

**Example:**
```yaml
!yisp
- import
- ["utils", "./utils.yaml"]
```

### `progn`

Evaluates all arguments in sequence and returns the value of the last one.

**Syntax:**
```yaml
!yisp
- progn
- expr1
- expr2
- ...
```

**Example:**
```yaml
result: !yisp
  - progn
  - some_operation
  - another_operation
  - final_result
# Evaluates to: result: final_result
```

### `pipeline`

Applies functions in sequence, passing the result of each to the next.

**Syntax:**
```yaml
!yisp
- pipeline
- initial_value
- function1
- function2
- ...
```

### `schema`

Creates a schema for type validation.

**Syntax:**
```yaml
!yisp
- schema
- schema_definition
```

### `as-document-root`

Marks the result as a document root for YAML output.

**Syntax:**
```yaml
!yisp
- as-document-root
- value1
- value2
- ...
```

## Lambda Functions

### `lambda`

Creates a lambda function that can be called later.

**Syntax:**
```yaml
!yisp &function_name
- lambda
- [param1, param2, ...]
- body
```

**Example:**
```yaml
!yisp &add
- lambda
- [a, b]
- - +
  - *a
  - *b

result: !yisp
  - *add
  - 3
  - 4
# Evaluates to: result: 7
```

### Calling Lambda Functions

Lambda functions are called using the `*` prefix followed by the function name:

```yaml
result: !yisp
  - *function_name
  - arg1
  - arg2
```

### Variable References

In lambda functions, parameters are referenced using the `*` prefix:

```yaml
!yisp &greet
- lambda
- [name]
- - strings.concat
  - "Hello, "
  - *name
  - "!"
```