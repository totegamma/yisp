# Exec Operators (`exec.*`)

Command execution operators in YISP. All operators in this module require the `exec.` prefix.

⚠️ **Security Note**: Command execution operators require the `--allow-cmd` flag to be enabled for security reasons.

## `exec.cmd`

Executes a command and returns its output.

**Syntax:**
```yaml
!yisp
- exec.cmd
- cmd: "command"
  args:
    - "arg1"
    - "arg2"
  asString: true/false
```

**Example:**
```yaml
date: !yisp
  - exec.cmd
  - cmd: date
# Evaluates to: date: <current date>

output: !yisp
  - exec.cmd
  - cmd: echo
    args:
      - "Hello, world!"
    asString: true
# Evaluates to: output: "Hello, world!"
```

## `exec.go`

Executes Go code. Requires the `--allow-cmd` flag.

**Syntax:**
```yaml
!yisp
- exec.go
- go_code_configuration
```