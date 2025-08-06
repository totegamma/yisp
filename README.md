<div align="center">
  <img src="./docs/assets/yisp-wordmark.png" alt="yisp logo" width="400px"/>
  
  # YISP - A Lisp-inspired Functional Engine for YAML
  [![Test](https://github.com/totegamma/yisp/actions/workflows/test.yaml/badge.svg)](https://github.com/totegamma/yisp/actions/workflows/test.yaml)
  
  [[getting-started]](https://github.com/totegamma/yisp/blob/main/docs/getting-started.md) | [[examples]](https://github.com/totegamma/yisp/blob/main/docs/examples.md) | [[other docs]](https://github.com/totegamma/yisp/tree/main/docs)
</div>

**YISP** (suggested pronunciation: `/ˈjɪsp/`) is a lightweight evaluation engine for YAML, inspired by Lisp.  
It allows you to embed logic, expressions, and includes within YAML files.  
This is useful for generating structured configuration such as Kubernetes manifests, Ansible playbooks, and more.

## Installation
Download latest version from [release page](https://github.com/totegamma/yisp/releases).

or use go install:
```sh
go install github.com/totegamma/yisp@latest
```

### Create cache for k8s definitions

If you use kubernetes manifests, you have to create a cache for kubernetes definitions.
Run the following command to download the latest Kubernetes API definitions:

```sh
yisp cache-kube-schemas
```

This command will run `kubectl get --raw /openapi/v2` to fetch the OpenAPI schema and store it in the cache directory.
You have to set up kubectl before running this command, so that it can access your Kubernetes cluster.

## Syntax
In yisp, YAML documents are treated as plain data by default.  
To enable evaluation, you explicitly mark expressions using the `!yisp` tag.

When a list or object is tagged with `!yisp`, its contents are recursively evaluated as yisp expressions.  
To embed unevaluated YAML structures inside expressions, you can use the `!quote` tag to suppress evaluation.

### simple example:

hello_world.yaml
```yaml
mystring: !yisp
  - strings.concat
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

## Use yisp from Go code
```go
package main

import (
	"fmt"
	"github.com/totegamma/yisp/pkg"
)

func main() {
	evaluated, err := yisp.EvaluateFileToYaml("test.yaml")
	if err != nil {
		panic(err)
	}

	fmt.Println("Evaluated YAML:")
	fmt.Println(evaluated)
}
```

also you can use `yisp.EvaluateFileToAny` to get the result as go `any` type.

