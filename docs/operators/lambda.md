# Lambda Functions

YISP supports creating and calling lambda functions for reusable code.

## `lambda`

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

## Calling Lambda Functions

Lambda functions are called using the `*` prefix followed by the function name:

```yaml
result: !yisp
  - *function_name
  - arg1
  - arg2
```

## Conditional Logic

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

## Variable References

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