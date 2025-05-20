# YISP Examples

This document provides practical examples of YISP in action, demonstrating how it can be used to solve real-world problems. Each example includes detailed explanations to help you understand the concepts and techniques being used.

## Kubernetes Configuration Management

YISP excels at managing complex Kubernetes configurations by allowing you to create reusable templates, apply conditional logic, and generate dynamic content. The following examples demonstrate common patterns for Kubernetes manifest management.

### Templating Kubernetes Manifests

One of the most powerful features of YISP is the ability to create reusable templates for Kubernetes resources. This example demonstrates how to create a template for a Pod and then use it to generate a specific Pod configuration.

**template.yaml**
```yaml
!yisp &mkpod
- lambda
- [name, image]
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

In this template:
- We define a function named `mkpod` using the `lambda` operator
- The function takes two parameters: `name` and `image`
- The `!quote` tag ensures the YAML structure is preserved without evaluation
- The `*name` and `*image` references are replaced with the actual values when the function is called

**use.yaml**
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

Here we:
1. Import the template file, making the `mkpod` function available as `template.mkpod`
2. Call the function with specific values for the pod name and image

**Result:**
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

This approach allows you to maintain consistent resource definitions while customizing specific properties. It's particularly useful when you need to create multiple similar resources with slight variations.

### Patching Kubernetes Manifests

Sometimes you need to modify existing Kubernetes manifests without completely rewriting them. YISP makes this easy with its functional approach to data transformation.

**patch.yaml**
```yaml
!yisp &selective-patch
- lambda
- [props]
- - map
  - - lambda
    - [x]
    - - if
      - - ==
        - *x.kind
        - *props.kind
      - - merge
        - *x
        - *props.patch
      - *x
  - *props.input
```

This function:
- Takes a `props` object containing `input` (the manifests to patch), `kind` (the resource kind to target), and `patch` (the changes to apply)
- Uses `map` to iterate through each manifest in the input
- For each manifest, checks if its `kind` matches the target kind
- If it matches, merges the original manifest with the patch
- If it doesn't match, returns the original manifest unchanged

**Usage:**
```yaml
- *kube.selective-patch
- input:
    - include
    - ./mymanifests.yaml
  kind: Deployment
  patch:
    spec:
      template:
        metadata:
          annotations:
            checksum/config:
              - sha256
              - - to-yaml
                - *config
```

This example:
1. Loads manifests from `./mymanifests.yaml`
2. Finds all Deployment resources
3. Adds a `checksum/config` annotation with the SHA-256 hash of a configuration object

This pattern is particularly useful for operations like:
- Adding annotations or labels to specific resources
- Updating image versions across multiple deployments
- Injecting configuration values into existing manifests

### Generating Kubernetes ConfigMaps

ConfigMaps are commonly used to store configuration files in Kubernetes. YISP can automate the process of generating ConfigMaps from a directory of files, similar to Kustomize's configMapGenerator.

```yaml
!yisp &configmap-generator
- lambda
- [props]
- !quote
  apiVersion: v1
  kind: ConfigMap
  metadata:
    name: *props.name
  data: !yisp
    - from-entries
    - - map
      - - lambda
        - [file]
        - !quote
          - *file.name
          - *file.body
      - - read-files
        - *props.dir
```

This function:
- Takes a `props` object with `name` (the ConfigMap name) and `dir` (the directory containing configuration files)
- Uses `read-files` to read all files in the specified directory
- Maps each file to a key-value pair where the key is the filename and the value is the file content
- Converts these pairs to a map using `from-entries`
- Embeds this map in a ConfigMap resource

**Usage:**
```yaml
!yisp
- *kube.configmap-generator
- name: "my-config"
  dir: "./cm-files/*"
```

This generates a ConfigMap named "my-config" containing all files in the "./cm-files/" directory. This approach:
- Automatically updates the ConfigMap when files change
- Maintains a clean separation between configuration files and Kubernetes resources
- Reduces manual work when managing multiple configuration files

## Advanced YAML Manipulation

YISP provides powerful tools for manipulating YAML structures beyond simple templating. The following examples demonstrate more advanced techniques.

### Conditional Field Selection in Mappings

YISP allows you to conditionally include or modify fields in YAML mappings, which is particularly useful for configuration that varies based on environment or other factors.

```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: postgres
spec:
  replicas: 1
  serviceName: postgres
  selector:
    matchLabels:
      name: postgres
  template:
    metadata:
      labels:
        name: postgres
    spec:
      containers:
        - name: posgresql
          image: postgres:16-bookworm
          ports:
            - name: postgres
              containerPort: 5432
          <<: !yisp
            - if
            - *values.postgres.useSecret
            - envFrom: !quote
              - secretRef
                name: postgres-secret
            - env: !quote
              - name: POSTGRES_USER
                value: postgres
              - name: POSTGRES_PASSWORD
                value: postgres
              - name: POSTGRES_DB
                value: postgres
          volumeMounts:
            - name: postgres-varlib
              mountPath: "var/lib/postgresql/data"
```

In this example:
- We use the YAML merge key (`<<:`) with a YISP expression
- The expression uses an `if` operator to check if `values.postgres.useSecret` is true
- If true, it includes an `envFrom` field that references a Kubernetes Secret
- If false, it includes an `env` field with hardcoded environment variables

This pattern allows you to:
- Create a single template that works in multiple environments
- Toggle between different configuration approaches based on conditions
- Keep related configuration options together for better readability

### Calling external tools

#### go package

The package to be executed must be listed in Yisp's config file at ~/.config/yisp/config.yaml.
You can edit this file directly or use the yisp allow <pkgname> command to add an entry.
Wildcard patterns (*) are supported in the allowlist.

```yaml
!yisp &helm-chart
- lambda
- [!helm-chart-props props]
- - go-run
  - pkg: github.com/totegamma/yisp-helm-adapter@v0.1.0
    args:
      - *props.repo
      - *props.release
      - *props.version
    stdin:
      - to-yaml
      - *props.values
```

#### command

Requires `--allow-cmd` flag

```yaml
!yisp
- cmd
- cmd: date
```

By default, the output of the command is interpreted as YAML.
To interpret the output as a plain string, set `asString: true`.

## Working with Data Transformations

YISP's functional approach makes it well-suited for data transformation tasks. Here are some examples of common data manipulation patterns.

### List Processing

```yaml
# Map a function over a list
doubled: !yisp
  - map
  - - lambda
    - [x]
    - - *
      - *x
      - 2
  - !quote [1, 2, 3, 4, 5]
# Result: doubled: [2, 4, 6, 8, 10]

# Filter a list
evens: !yisp
  - filter
  - - lambda
    - [x]
    - - ==
      - - %
        - *x
        - 2
      - 0
  - !quote [1, 2, 3, 4, 5, 6]
# Result: evens: [2, 4, 6]

# Reduce a list
sum: !yisp
  - reduce
  - - lambda
    - [acc, x]
    - - +
      - *acc
      - *x
  - 0
  - !quote [1, 2, 3, 4, 5]
# Result: sum: 15
```

### Map Transformations

```yaml
# Convert a map to entries and back
transformed: !yisp
  - from-entries
  - - map
    - - lambda
      - [entry]
      - !quote
        - *entry.0
        - !yisp
          - +
          - *entry.1
          - 10
    - - to-entries
      - !quote
          a: 1
          b: 2
          c: 3
# Result: transformed: {a: 11, b: 12, c: 13}
```

## Recursive Functions

YISP supports recursive functions, allowing you to implement algorithms like Fibonacci:

```yaml
!yisp &fib
- lambda
- [n]
- - if
  - - <=
    - *n
    - 1
  - *n
  - - +
    - - *fib
      - - -
        - *n
        - 1
    - - *fib
      - - -
        - *n
        - 2

result: !yisp
  - *fib
  - 10
# Result: result: 55
```

## Conclusion

These examples demonstrate the power and flexibility of YISP for managing complex YAML configurations. By combining functional programming concepts with YAML's readable syntax, YISP enables you to create maintainable, reusable, and dynamic configuration templates.

For more information on available operators and syntax, refer to the [Operators](operators.md) documentation. To get started with YISP, check out the [Getting Started](getting-started.md) guide.
