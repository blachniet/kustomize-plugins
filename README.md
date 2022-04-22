# Kustomize Plugins

A collection of Kustomize plugins developed according to the [Kustomize plugins guide](https://kubectl.docs.kubernetes.io/guides/extending_kustomize/).

> You must provide the `--enable-alpha-plugins` argument to `kubectl` or `kustomize` to use these plugins because the [Plugin feature status] is currently in **alpha**.

## Plugins

### PullPolicyTransformer

Set the `imagePullPolicy` for all containers whose `image` name matches a given value.

This is particularly useful in combination with the [ImageTagTransformer] when changing the tag to/from a "rolling" tag (e.g. `latest`). When you change the tag to `latest`, you may want to set `imagePullPolicy: Always`. When you change the tag to a specific version (e.g. `v1.2.3`), you may want to set `imagePullPolicy: IfNotPresent`.

#### Example

```yaml
apiVersion: k8s.blachniet.com/v1alpha1
kind: PullPolicyTransformer
metadata:
  name: set-pull-policy
  annotations:
    config.kubernetes.io/function: |
      container:
        image: ghcr.io/blachniet/kustomize-plugins:0.3.0
images:
- name: nginx
  newPullPolicy: Always
- name: busybox
  newPullPolicy: IfNotPresent
```

## Contributing

### Generate Dockerfile

```sh
go run main.go gen .
```

### Run example one-liner

Build the Docker image and run an example.

```sh
docker build . -t ghcr.io/blachniet/kustomize-plugins:dev \
    && kubectl kustomize --enable-alpha-plugins examples/pull-policy-transformer \
    | less
```

### Release checklist

1. List existing releases and choose the next release version.

    ```sh
    gh release list
    ```

1. Update documentation and examples with the new version number. This is a manual step. Search the repository for `ghcr.io/blachniet/kustomize-plugins:`.
1. Commit changes, push and create the release.

    ```sh
    git commit <updated-files>
    git push
    gh release create
    ```

1. Confirm that GitHub workflows publish the new Docker image.

## Resources

- [Extending Kustomize](https://kubectl.docs.kubernetes.io/guides/extending_kustomize/)
- [Kustomize Built-Ins](https://kubectl.docs.kubernetes.io/references/kustomize/builtins/)

[ImageTagTransformer]: https://kubectl.docs.kubernetes.io/references/kustomize/builtins/#_imagetagtransformer_
[Plugin feature status]: https://kubectl.docs.kubernetes.io/guides/extending_kustomize/#plugin-feature-status
