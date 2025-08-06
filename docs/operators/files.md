# File Operators (`files.*`)

File system operations in YISP. All operators in this module require the `files.` prefix.

## `files.glob`

Reads files matching a glob pattern and returns information about them.

**Syntax:**
```yaml
!yisp
- files.glob
- "glob/pattern/*.yaml"
- ...
```

**Example:**
```yaml
files: !yisp
  - files.glob
  - "config/*.yaml"
# Returns an array of maps with path, name, and body keys
```

**Return Format:**
Each file is returned as a map containing:
- `path`: The full path to the file
- `name`: The filename
- `body`: The file contents