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

## `maps.patch`

Applies a patch to a map, similar to merge but with more advanced merging capabilities.

**Syntax:**
```yaml
!yisp
- maps.patch
- base_map
- patch_map
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