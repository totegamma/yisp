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
