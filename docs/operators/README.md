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
- [Exec](exec.md) - Command execution (`exec.*`)
- [Types](types.md) - Type checking and conversion (`types.*`)
- [YAML](yaml.md) - YAML serialization (`yaml.*`)
- [k8s](k8s.md) - k8s manifest manipulation operators (`k8s.*`)

