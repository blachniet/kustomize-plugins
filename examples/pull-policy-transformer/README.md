# PullPolicyTransformer Example

In this example, we use the transformer to:

- Set `imagePullPolicy: Always` for any containers whose `image` name is `nginx`.
- Set `imagePullPolicy: IfNotPresent` for any containers whose `image` name is `busybox`.

Run this example with `kubectl` or `kustomize`:

```sh
kubectl kustomize --enable-alpha-plugins 
kustomize build
```
