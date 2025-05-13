# YISP Project Overview

## Introduction

YISP (YAML Lisp) is a lightweight evaluation engine for YAML, inspired by Lisp. It allows you to embed logic, expressions, and includes within YAML files, making it a powerful tool for generating structured configuration such as Kubernetes manifests, Ansible playbooks, and more.

## Core Concepts

### YAML as Data

In YISP, YAML documents are treated as plain data by default. This means that any YAML file can be processed by YISP without modification.

### Evaluation Tags

To enable evaluation, YISP uses special tags:

- `!yisp`: Marks a YAML node for evaluation. When a list or object is tagged with `!yisp`, its contents are recursively evaluated as YISP expressions.
- `!quote`: Suppresses evaluation, allowing you to embed unevaluated YAML structures inside expressions.

### Lisp-like Syntax

YISP uses a Lisp-like syntax for expressions, where the first element of a list is the operator and the remaining elements are the arguments:

```yaml
!yisp
- operator
- arg1
- arg2
- ...
```

### Anchors and References

YISP supports YAML anchors (`&name`) and references (`*name`) for defining and reusing values:

```yaml
!yisp &my_function
- lambda
- [arg1, arg2]
- body

# Later in the document
!yisp
- *my_function
- value1
- value2
```

## Architecture

YISP is built around several key components:

### Parser

The parser reads YAML files and converts them into an internal representation that can be evaluated. It handles YAML-specific features like tags, anchors, and references.

### Evaluator

The evaluator processes the parsed YAML and executes any YISP expressions it encounters. It maintains an environment of variables and functions that can be referenced during evaluation.

### Operators

YISP provides a set of built-in operators for common operations like arithmetic, string manipulation, conditionals, and list processing. These operators form the core functionality of the language.

### Environment

The environment keeps track of variables, functions, and imported modules. It provides a scope for variable resolution and allows for the creation of child environments for local scoping.

## Workflow

The typical workflow for using YISP is:

1. Write YAML files with embedded YISP expressions
2. Process these files with the YISP tool
3. Use the resulting pure YAML output for your application

This workflow allows you to maintain the simplicity and readability of YAML while adding the power and flexibility of a programming language.

