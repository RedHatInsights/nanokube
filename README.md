Nanokube
========

Tiny server, based off of the test environment setup from the k8s project. Does not have any
authentication so use with care. Loaded with a couple of CRDs from the HAC project.

```shell
KUBEBUILDER_ASSETS=`setup-envtest use 1.21 -p path` ./nanokube
```