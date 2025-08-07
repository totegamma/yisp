# YISP Operators

⚠️ **Important**: This documentation is now organized into separate files. For detailed operator documentation, please visit the [operators directory](operators/).

## Quick Links

- [**Overview**](operators/README.md) - Complete operator reference with migration guide
- [**Built-in Operators**](operators/builtin.md) - Core operators (no prefix required)
- [**Lambda Functions**](operators/lambda.md) - Creating and using functions

## Library Operators (require module prefix)

- [**Strings**](operators/strings.md) - `strings.*` operators
- [**Lists**](operators/lists.md) - `lists.*` operators  
- [**Maps**](operators/maps.md) - `maps.*` operators
- [**Files**](operators/files.md) - `files.*` operators
- [**Crypto**](operators/crypto.md) - `crypto.*` operators
- [**Exec**](operators/exec.md) - `exec.*` operators
- [**Types**](operators/types.md) - `types.*` operators
- [**YAML**](operators/yaml.md) - `yaml.*` operators
- [**k8s**](operators/k8s.md) - `k8s.*` operators

## Migration Note

**Breaking Change**: Many operators that were previously available without prefixes now require module prefixes:

- `concat` → `strings.concat`
- `car` → `lists.car`
- `mapping-get` → `maps.get`
- `to-yaml` → `yaml.marshal`
- `sha256` → `crypto.sha256`
- And more...

See the [full migration guide](operators/README.md#migration-from-previous-versions) for complete details.
