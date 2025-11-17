# Getting Started with YISP

This guide will help you get started with YISP, a Lisp-like evaluator for YAML documents.

## Installation

You can install YISP using Go's package manager:

```sh
go install github.com/totegamma/yisp@latest
```

This will download and install the latest version of YISP to your Go bin directory.

## Basic Usage

The primary command for using YISP is `yisp build`, which processes a YAML file with YISP expressions and outputs the evaluated result.

```sh
yisp build path/to/your/file.yaml
```

By default, the output is printed to stdout. You can redirect it to a file if needed:

```sh
yisp build path/to/your/file.yaml > output.yaml
```

### Build Command Options

The `yisp build` command supports several flags to customize the build process:

- `--output`, `-o`: Specify the output format (`yaml` or `json`, default: `yaml`)
- `--disable-type-check`: Disable type checking during output generation
- `--allow-untyped-manifest`: Allow manifests without type information (useful for Kubernetes resources)
- `--show-trace`: Show detailed trace information for debugging
- `--enable-sourcemap`: Include source map comments in the output YAML
- `--render-special-objects`: Display special objects like types and lambdas in the output
- `--allow-cmd`: Allow command execution through `exec.*` operators

**Example:**
```sh
# Build with JSON output
yisp build input.yaml --output json

# Build with type checking disabled
yisp build input.yaml --disable-type-check

# Build with trace information for debugging
yisp build input.yaml --show-trace
```

## Your First YISP File

Let's create a simple YISP file to demonstrate the basics:

1. Create a file named `hello.yaml` with the following content:

```yaml
message: !yisp
  - strings.concat
  - "Hello, "
  - "YISP!"

calculation: !yisp
  - +
  - 10
  - 20
  - 30

conditional: !yisp
  - if
  - - <
    - 5
    - 10
  - "5 is less than 10"
  - "5 is not less than 10"
```

2. Process the file with YISP:

```sh
yisp build hello.yaml
```

3. You should see the following output:

```yaml
message: Hello, YISP!
calculation: 60
conditional: 5 is less than 10
```

## Understanding YISP Syntax

### The `!yisp` Tag

The `!yisp` tag marks a YAML node for evaluation. When YISP encounters this tag, it evaluates the node as a YISP expression.

```yaml
result: !yisp
  - operator
  - arg1
  - arg2
  - ...
```

In YISP expressions, the first element of a list is the operator, and the remaining elements are the arguments.

### The `!quote` Tag

The `!quote` tag suppresses evaluation, allowing you to embed unevaluated YAML structures inside expressions.

```yaml
data: !yisp
  - some_function
  - !quote
    - this
    - will
    - not
    - be
    - evaluated
```

### Anchors and References

YISP supports YAML anchors (`&name`) and references (`*name`) for defining and reusing values:

```yaml
!yisp &add_five
- lambda
- [x]
- - +
  - *x
  - 5

---

result: !yisp
  - *add_five
  - 10
# Evaluates to: result: 15
```

## Creating Functions

You can define functions using the `lambda` operator:

```yaml
!yisp &greet
- lambda
- [name]
- - strings.concat
  - "Hello, "
  - *name
  - "!"

---

greeting: !yisp
  - *greet
  - "World"
# Evaluates to: greeting: "Hello, World!"
```

## Type Checking

YISP includes a type checking system to help catch errors early and ensure data consistency. The type system supports basic types like strings, integers, floats, and booleans, as well as custom type schemas for more complex structures.

### Type Tags

You can specify types for lambda function parameters using type tags:

```yaml
!yisp &calculate
- lambda
- [!int x, !int y]
- - +
  - *x
  - *y

---
result: !yisp
  - *calculate
  - 10
  - 20
# Evaluates to: result: 30
```

**Available type tags:**
- `!string`: String type
- `!int`: Integer type
- `!float`: Floating-point number type
- `!bool`: Boolean type

### Type Assertions

You can use the `types.assert` operator to validate that a value matches a specific type:

```yaml
validated: !yisp
  - types.assert
  - "hello"
  - !string
# Passes validation

invalid: !yisp
  - types.assert
  - 42
  - !string
# Fails with a type error
```

### Type Checking for Kubernetes Manifests

When working with Kubernetes manifests, YISP can validate resources against their schemas. By default, type checking is enabled. If you're working with resources that don't have schema information, you can use the `--allow-untyped-manifest` flag.

If you want to disable type checking entirely (e.g., during development or for debugging), use the `--disable-type-check` flag:

```sh
yisp build manifests.yaml --disable-type-check
```

### Getting Type Information

You can retrieve type information about values using type operators:

```yaml
# Get the type of a value as a string
type: !yisp
  - types.of
  - 42
# Evaluates to: type: "int"

# Get detailed type information
info: !yisp
  - types.get
  - "hello world"
# Returns type metadata
```

For more information about type operators, see the [Types documentation](operators/types.md).

## Working with Multiple Files

YISP allows you to include and import files, making it easy to organize your code:

### Including Files

The `include` operator evaluates files and returns their results as a list:

```yaml
# main.yaml
results: !yisp
  - include
  - "part1.yaml"
  - "part2.yaml"
```

### Importing Modules

The `import` operator imports modules, making their definitions available in the current environment:

```yaml
# main.yaml
!yisp
- import
- ["utils", "./utils.yaml"]
---
result: !yisp
  - *utils.some_function
  - arg1
  - arg2
```

## Next Steps

Now that you understand the basics of YISP, you can:

- Explore the [Operators](operators.md) documentation to learn about all available operators
- Check out the [Examples](examples.md) for real-world use cases

Happy coding with YISP!
