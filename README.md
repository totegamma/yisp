<div align="center">
  <img src="./docs/assets/yisp-logo.png" alt="yisp logo" width="200px"/>
  
  # YISP - A Lisp-like evaluator for YAML documents
</div>

**YISP** (suggested pronunciation: `/ˈjɪsp/`) is a lightweight evaluation engine for YAML, inspired by Lisp.  
It allows you to embed logic, expressions, and includes within YAML files.  
This is useful for generating structured configuration such as Kubernetes manifests, Ansible playbooks, and more.

> [!NOTE]
> This project is currently under active development.
> Interfaces and features may change without notice. Use with caution in production environments.

## Installation

```sh
go install github.com/totegamma/yisp@latest
```

## Syntax
In yisp, YAML documents are treated as plain data by default.  
To enable evaluation, you explicitly mark expressions using the `!yisp` tag.

When a list or object is tagged with `!yisp`, its contents are recursively evaluated as yisp expressions.  
To embed unevaluated YAML structures inside expressions, you can use the `!quote` tag to suppress evaluation.

### simple example:

hello_world.yaml
```yaml
mystring: !yisp
  - concat
  - hello
  - ' '
  - world
```

build:
```sh
yisp build hello_world.yaml
```

result:
```yaml
mystring: hello world
```

### Define functions and call it from another file:

template.yaml:
```yaml
!yisp &mkpod
- lambda
- [!string name, !string image]
- !quote
  apiVersion: v1
  kind: Pod
  metadata:
    name: *name
  spec:
    containers:
      - name: *name
        image: *image
```

main.yaml
```yaml
!yisp
- import
- ["template", "./template.yaml"]
---
!yisp
- *template.mkpod
- mypod1
- myimage1
```

result:
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: mypod1
spec:
  containers:
    - name: mypod1
      image: myimage1
```

More examples are available in `/testfiles`.

### TODO
- Add evaluation tracing for debugging
- Add remote file loading
- Add map/filter/reduce functions

