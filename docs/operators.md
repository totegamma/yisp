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

Checks if two values are equal. Works with numbers, strings, and booleans. Different types are never considered equal.

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

Checks if two values are not equal. Works with numbers, strings, and booleans. Different types are always considered not equal.

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

Checks if the first value is less than or equal to the second. Works with numbers.

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

Checks if the first value is greater than the second. Works with numbers.

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

Checks if the first value is greater than or equal to the second. Works with numbers.

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

### `flatten`

Flattens multiple lists into a single list.

**Syntax:**
```yaml
!yisp
- flatten
- list1
- list2
- ...
```

**Example:**
```yaml
result: !yisp
  - flatten
  - !quote [a, b, c]
  - !quote [d, e, f]
# Evaluates to: result: [a, b, c, d, e, f]
```

### `map`

Applies a function to each element of one or more lists.

**Syntax:**
```yaml
!yisp
- map
- function
- list1
- list2
- ...
```

**Example:**
```yaml
result: !yisp
  - map
  - +
  - !quote [1, 2, 3]
  - !quote [4, 5, 6]
# Evaluates to: result: [5, 7, 9]
```

## Map Operators

### `mapping-get`

Gets a value from a map by key.

**Syntax:**
```yaml
!yisp
- mapping-get
- map
- key
```

**Example:**
```yaml
result: !yisp
  - mapping-get
  - hoge: piyo
    fuga: miyo
  - hoge
# Evaluates to: result: piyo
```

### `merge`

Merges multiple maps together. When keys conflict, later maps override earlier ones.

**Syntax:**
```yaml
!yisp
- merge
- map1
- map2
- ...
```

**Example:**
```yaml
result: !yisp
  - merge
  - !quote
      app:
        name: myapp
        version: 1.0
  - !quote
      app:
        version: 1.1
        description: "Updated app"
# Evaluates to: result: {app: {name: myapp, version: 1.1, description: "Updated app"}}
```

### `to-entries`

Converts a map to an array of key-value pairs.

**Syntax:**
```yaml
!yisp
- to-entries
- map
```

**Example:**
```yaml
entries: !yisp
  - to-entries
  - a: 1
    b: 2
# Evaluates to: entries: [["a", 1], ["b", 2]]
```

### `from-entries`

Converts an array of key-value pairs to a map.

**Syntax:**
```yaml
!yisp
- from-entries
- array_of_pairs
```

**Example:**
```yaml
map: !yisp
  - from-entries
  - !quote
    - ["a", 1]
    - ["b", 2]
# Evaluates to: map: {a: 1, b: 2}
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

### `to-yaml`

Converts a YISP value to a YAML string.

**Syntax:**
```yaml
!yisp
- to-yaml
- value
```

**Example:**
```yaml
yaml: !yisp
  - to-yaml
  - hoge: piyo
    a: 1
# Evaluates to: yaml: "hoge: piyo\na: 1\n"
```

### `sha256`

Calculates the SHA-256 hash of a string.

**Syntax:**
```yaml
!yisp
- sha256
- string
```

**Example:**
```yaml
hash: !yisp
  - sha256
  - "hello world"
# Evaluates to: hash: "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
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

### `read-files`

Reads files matching a glob pattern and returns information about them.

**Syntax:**
```yaml
!yisp
- read-files
- "glob/pattern/*.yaml"
- ...
```

**Example:**
```yaml
files: !yisp
  - read-files
  - "config/*.yaml"
# Returns an array of maps with path, name, and body keys
```

## Command Execution

### `cmd`

Executes a command and returns its output. Requires the `--allow-cmd` flag to be enabled.

**Syntax:**
```yaml
!yisp
- cmd
- cmd: "command"
  args:
    - "arg1"
    - "arg2"
  asString: true/false
```

**Example:**
```yaml
date: !yisp
  - cmd
  - cmd: date
# Evaluates to: date: <current date>

output: !yisp
  - cmd
  - cmd: echo
    args:
      - "Hello, world!"
    asString: true
# Evaluates to: output: "Hello, world!"
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
