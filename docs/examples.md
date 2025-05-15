# Examples

### Templating kubernetes manifests

template.yaml
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

use.yaml
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

### Patching kubernetes manifests

patch.yaml
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

usage:
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

### Generating kubernetes configmaps

This functions works like a kustomize's configMapGenerator

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

usage:
```yaml
!yisp
- *kube.configmap-generator
- name: "my-config"
  dir: "./cm-files/*"
```


### Use if-else in mappings

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

