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
