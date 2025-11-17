# Map Operators (`maps.*`)

Map/object manipulation operators in YISP. All operators in this module require the `maps.` prefix.

## `maps.get`

Gets a value from a map by key.

**Syntax:**
```yaml
!yisp
- maps.get
- map
- key
```

**Example:**
```yaml
result: !yisp
  - maps.get
  - hoge: piyo
    fuga: miyo
  - hoge
# Evaluates to: result: piyo
```

## `maps.merge`

Merges multiple maps together. When keys conflict, later maps override earlier ones.

**Syntax:**
```yaml
!yisp
- maps.merge
- map1
- map2
- ...
```

**Example:**
```yaml
result: !yisp
  - maps.merge
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

## `maps.to-entries`

Converts a map to an array of key-value pairs.

**Syntax:**
```yaml
!yisp
- maps.to-entries
- map
```

**Example:**
```yaml
entries: !yisp
  - maps.to-entries
  - a: 1
    b: 2
# Evaluates to: entries: [["a", 1], ["b", 2]]
```

## `maps.from-entries`

Converts an array of key-value pairs to a map.

**Syntax:**
```yaml
!yisp
- maps.from-entries
- array_of_pairs
```

**Example:**
```yaml
map: !yisp
  - maps.from-entries
  - !quote
    - ["a", 1]
    - ["b", 2]
# Evaluates to: map: {a: 1, b: 2}
```

## `maps.keys`

Returns a list of all keys from a map.

**Syntax:**
```yaml
!yisp
- maps.keys
- map
```

**Example:**
```yaml
keys: !yisp
  - maps.keys
  - !quote {a: 1, b: 2, c: 3}
# Evaluates to: keys: ["a", "b", "c"]
```

## `maps.values`

Returns a list of all values from a map.

**Syntax:**
```yaml
!yisp
- maps.values
- map
```

**Example:**
```yaml
values: !yisp
  - maps.values
  - !quote {a: 1, b: 2, c: 3}
# Evaluates to: values: [1, 2, 3]
```

## `maps.make`

Creates a new map from a list of key-value pairs. Keys and values are provided as alternating arguments.

**Syntax:**
```yaml
!yisp
- maps.make
- key1
- value1
- key2
- value2
- ...
```

**Example:**
```yaml
config: !yisp
  - maps.make
  - name
  - myapp
  - version
  - 1.0
  - debug
  - true
# Evaluates to: config: {name: myapp, version: 1.0, debug: true}
```

**Notes:**
- Keys must be strings
- The number of arguments must be even (each key must have a corresponding value)
- This operator is useful for dynamically constructing maps when the keys or values are computed

## `maps.patch`

Applies JSON Patch (RFC 6902) operations to a map or array. This operator supports the standard JSON Patch operations: `add`, `remove`, `replace`, `move`, `copy`, and `test`. Additionally, YISP extends RFC 6902 with wildcard support using `*` in paths.

**Syntax:**
```yaml
!yisp
- maps.patch
- target
- patch_operations
```

Where `patch_operations` is either:
- A single patch operation (map with `op`, `path`, and optionally `value` or `from` fields)
- An array of patch operations

**Supported Operations:**

### `add`
Adds a value to a specified path. For maps, this creates or replaces a key. For arrays, this inserts at the specified index or appends to the end.

**Example:**
```yaml
result: !yisp
  - maps.patch
  - !quote
      app:
        name: myapp
  - !quote
    - op: add
      path: /app/version
      value: 1.0
# Evaluates to: result: {app: {name: myapp, version: 1.0}}
```

### `remove`
Removes a value at the specified path.

**Example:**
```yaml
result: !yisp
  - maps.patch
  - !quote {name: myapp, debug: true}
  - !quote
    - op: remove
      path: /debug
# Evaluates to: result: {name: myapp}
```

### `replace`
Replaces an existing value at the specified path.

**Example:**
```yaml
result: !yisp
  - maps.patch
  - !quote {version: 1.0}
  - !quote
    - op: replace
      path: /version
      value: 2.0
# Evaluates to: result: {version: 2.0}
```

### `move`
Moves a value from one path to another.

**Example:**
```yaml
result: !yisp
  - maps.patch
  - !quote {oldKey: value}
  - !quote
    - op: move
      from: /oldKey
      path: /newKey
# Evaluates to: result: {newKey: value}
```

### `copy`
Copies a value from one path to another.

**Example:**
```yaml
result: !yisp
  - maps.patch
  - !quote {original: value}
  - !quote
    - op: copy
      from: /original
      path: /duplicate
# Evaluates to: result: {original: value, duplicate: value}
```

### `test`
Tests that a value at the specified path matches an expected value.

**Example:**
```yaml
result: !yisp
  - maps.patch
  - !quote {version: 1.0}
  - !quote
    - op: test
      path: /version
      value: 1.0
# Succeeds if version is 1.0, otherwise fails
```

**Wildcard Support:**

YISP extends RFC 6902 with wildcard support. You can use `*` in a path to match all keys in a map or all elements in an array.

**Example (wildcard with maps):**
```yaml
result: !yisp
  - maps.patch
  - !quote
      services:
        web:
          replicas: 1
        api:
          replicas: 1
  - !quote
    - op: add
      path: /services/*/replicas
      value: 3
# Evaluates to: result: {services: {web: {replicas: 3}, api: {replicas: 3}}}
```

**Example (wildcard with arrays):**
```yaml
result: !yisp
  - maps.patch
  - !quote
      items:
        - name: item1
        - name: item2
  - !quote
    - op: add
      path: /items/*
      value:
        enabled: true
# Adds 'enabled: true' to all items in the array
```

**Array Operations:**

For arrays, you can:
- Use numeric indices to specify positions (e.g., `/items/0`)
- Use `-` to append to the end (e.g., `/items/-`)
- Insert at a specific position using `add` operation

**Example (array operations):**
```yaml
result: !yisp
  - maps.patch
  - !quote
      items: [a, b, d]
  - !quote
    - op: add
      path: /items/2
      value: c
# Evaluates to: result: {items: [a, b, c, d]}
```

**Notes:**
- JSON Pointer paths use `/` as a separator and start with `/`
- Special characters in keys must be escaped: `~` becomes `~0` and `/` becomes `~1`
- Multiple patch operations are applied in order
- The patch operations modify the target in place
