# YISP Operators

YISP provides two categories of operators: built-in operators and library operators.

## Built-in Operators

Built-in operators are always available and don't require any module prefix:

- [Built-in Operators](builtin.md) - Core operators implemented in the engine

## Library Operators  

Library operators are organized into modules and require a module prefix (e.g., `strings.concat`):

- [Strings](strings.md) - String manipulation operators (`strings.*`)
- [Lists](lists.md) - List manipulation operators (`lists.*`)
- [Maps](maps.md) - Map/object manipulation operators (`maps.*`)
- [Files](files.md) - File system operations (`files.*`)
- [Crypto](crypto.md) - Cryptographic functions (`crypto.*`)
- [Exec](exec.md) - Command execution (`exec.*`)
- [Types](types.md) - Type checking and conversion (`types.*`)
- [YAML](yaml.md) - YAML serialization (`yaml.*`)

## Migration from Previous Versions

If you're updating from a previous version where operators didn't require module prefixes, you'll need to update your code:

- `concat` → `strings.concat`
- `car` → `lists.car`
- `cdr` → `lists.cdr`
- `cons` → `lists.cons`
- `flatten` → `lists.flatten`
- `map` → `lists.map`
- `mapping-get` → `maps.get`
- `merge` → `maps.merge`
- `to-entries` → `maps.to-entries`
- `from-entries` → `maps.from-entries`
- `to-yaml` → `yaml.marshal`
- `sha256` → `crypto.sha256`
- `cmd` → `exec.cmd`
- `read-files` → `files.glob`