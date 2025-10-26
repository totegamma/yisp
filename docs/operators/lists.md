# List Operators (`lists.*`)

List manipulation operators in YISP. All operators in this module require the `lists.` prefix.

## `lists.car`

Returns the first element of a list.

**Syntax:**
```yaml
!yisp
- lists.car
- list
```

**Example:**
```yaml
first: !yisp
  - lists.car
  - !quote
    - 1
    - 2
    - 3
# Evaluates to: first: 1
```

## `lists.cdr`

Returns all but the first element of a list.

**Syntax:**
```yaml
!yisp
- lists.cdr
- list
```

**Example:**
```yaml
rest: !yisp
  - lists.cdr
  - !quote
    - 1
    - 2
    - 3
# Evaluates to: rest: [2, 3]
```

## `lists.cons`

Constructs a new list by adding an element to the front of a list.

**Syntax:**
```yaml
!yisp
- lists.cons
- element
- list
```

**Example:**
```yaml
newList: !yisp
  - lists.cons
  - 0
  - !quote
    - 1
    - 2
    - 3
# Evaluates to: newList: [0, 1, 2, 3]
```

## `lists.flatten`

Flattens multiple lists into a single list.

**Syntax:**
```yaml
!yisp
- lists.flatten
- list1
- list2
- ...
```

**Example:**
```yaml
result: !yisp
  - lists.flatten
  - !quote [a, b, c]
  - !quote [d, e, f]
# Evaluates to: result: [a, b, c, d, e, f]
```

## `lists.map`

Applies a function to each element of one or more lists.

**Syntax:**
```yaml
!yisp
- lists.map
- function
- list1
- list2
- ...
```

**Example:**
```yaml
result: !yisp
  - lists.map
  - +
  - !quote [1, 2, 3]
  - !quote [4, 5, 6]
# Evaluates to: result: [5, 7, 9]
```

## `lists.filter`

Filters a list by keeping only elements that match a predicate function.

**Syntax:**
```yaml
!yisp
- lists.filter
- predicate_function
- list
```

**Example:**
```yaml
evens: !yisp
  - lists.filter
  - eq 0
  - !quote [1, 2, 3, 4, 5, 6]
# Note: This would require a more complex predicate setup
```

## `lists.reduce`

Reduces a list to a single value by applying a function cumulatively.

**Syntax:**
```yaml
!yisp
- lists.reduce
- list
- function
```

**Example:**
```yaml
sum: !yisp
  - lists.reduce
  - !quote [1, 2, 3, 4, 5]
  - +
# Evaluates to: sum: 15
```

## `lists.iota`

Generates a list of integers from start to start+n-1. If only one argument is provided, starts from 0.

**Syntax:**
```yaml
!yisp
- lists.iota
- n
- start  # optional, defaults to 0
```

**Example:**
```yaml
numbers: !yisp
  - lists.iota
  - 5
# Evaluates to: numbers: [0, 1, 2, 3, 4]

numbersFrom10: !yisp
  - lists.iota
  - 5
  - 10
# Evaluates to: numbersFrom10: [10, 11, 12, 13, 14]
```

## `lists.length`

Returns the length of a list.

**Syntax:**
```yaml
!yisp
- lists.length
- list
```

**Example:**
```yaml
len: !yisp
  - lists.length
  - !quote [a, b, c, d]
# Evaluates to: len: 4
```

## `lists.at`

Gets an element from a list at the specified index (0-based).

**Syntax:**
```yaml
!yisp
- lists.at
- list
- index
```

**Example:**
```yaml
element: !yisp
  - lists.at
  - !quote [a, b, c, d]
  - 2
# Evaluates to: element: c
```

## `lists.as-toplevel`

Flattens and marks the result as a top-level document root for YAML output, causing elements to be output as separate YAML documents.

**Syntax:**
```yaml
!yisp
- lists.as-toplevel
- value1
- value2
- ...
```

**Example:**
```yaml
documents: !yisp
  - lists.as-toplevel
  - !quote
    - kind: ConfigMap
      name: config1
    - kind: ConfigMap
      name: config2
# Outputs multiple YAML documents
```