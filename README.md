# YISP - A Lisp-like evaluator for YAML documents

> 🚧 Note: This project is currently under active development.
> Interfaces and features may change without notice. Use with caution in production environments.

**YISP** (suggested pronunciation: `/ˈjɪsp/`) is a lightweight evaluation engine for YAML, inspired by Lisp.  
It allows you to embed logic, expressions, and includes within YAML files.  
This is useful for generating structured configuration such as Kubernetes manifests, Ansible playbooks, and more.

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

### Handling multiple documents:

```yaml
!yisp
- include
- "./manifest1.yaml"
- "./manifest2.yaml"
```

such as:
manifest1.yaml:
```yaml 
apiVersion: v1
kind: Pod
metadata:
  name: mypod
```
manifest2.yaml:
```yaml
apiVersion: v1
kind: Service
metadata:
  name: myservice
```

results:
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: mypod
---
apiVersion: v1
kind: Service
metadata:
  name: myservice
```

### Define a function:

```yaml
!yisp
- discard
- &mkpod
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
---
!yisp
- *mkpod
- mypod1
- myimage1

---
!yisp
- *mkpod
- mypod2
- myimage2
```

results:
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: mypod1
spec:
  containers:
    - name: mypod1
      image: myimage1
---
apiVersion: v1
kind: Pod
metadata:
  name: mypod2
spec:
  containers:
    - name: mypod2
      image: myimage2
```

