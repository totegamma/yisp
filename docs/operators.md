# YISP Operators

YISP provides a variety of built-in operators that form the core functionality of the language. This document provides detailed information about each operator, including its purpose, syntax, and examples.

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

## String Operators

### `concat`

Concatenates two or more strings together.

**Syntax:**
```yaml
!yisp
- concat
- string1
- string2
- ...
```

**Example:**
```yaml
greeting: !yisp
  - concat
  - "Hello, "
  - "world!"
# Evaluates to: greeting: "Hello, world!"
```

## Comparison Operators

### `==` (Equal)

Checks if two values are equal.

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

Checks if the first value is less than the second.

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
  - <
  - 5
  - 10
  - "Less"
  - "Greater or Equal"
# Evaluates to: result: "Less"
```

## List Operators

### `car`

Returns the first element of a list.

**Syntax:**
```yaml
!yisp
- car
- list
```

**Example:**
```yaml
first: !yisp
  - car
  - !quote
    - 1
    - 2
    - 3
# Evaluates to: first: 1
```

### `cdr`

Returns all but the first element of a list.

**Syntax:**
```yaml
!yisp
- cdr
- list
```

**Example:**
```yaml
rest: !yisp
  - cdr
  - !quote
    - 1
    - 2
    - 3
# Evaluates to: rest: [2, 3]
```

### `cons`

Constructs a new list by adding an element to the front of a list.

**Syntax:**
```yaml
!yisp
- cons
- element
- list
```

**Example:**
```yaml
newList: !yisp
  - cons
  - 0
  - !quote
    - 1
    - 2
    - 3
# Evaluates to: newList: [0, 1, 2, 3]
```

## Miscellaneous Operators

### `discard`

Evaluates all arguments and returns nil. Useful for side effects.

**Syntax:**
```yaml
!yisp
- discard
- arg1
- arg2
- ...
```

**Example:**
```yaml
!yisp
- discard
- some_operation
- with_side_effects
```

## File and Module Operators

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

## Function Operators

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
- concat
- "Hello, "
- *name
- "! You are "
- concat
- *age
- " years old."
