# k8s Operators (`k8s.*`)

k8s manifest manipulation operators in YISP. All operators in this module require the `k8s.` prefix.

## `k8s.patch`

Applies a patch to a map, similar to merge but with more advanced merging capabilities.
Target identified by `apiVersion`, `kind`, `metadata.namespace` and `metadata.name`.

**Syntax:**
```yaml
!yisp
- k8s.patch
- base_map
- patch_map
```


