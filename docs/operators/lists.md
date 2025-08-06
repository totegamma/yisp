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