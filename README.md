# knamespacer ![Main Workflow](https://github.com/ejether/knamespacer/actions/workflows/main/badge.svg)

Kubernetes Namespace Controller

![The k is silent](/images/knamespacer.png)

**Note** This project is alpha and undergoing development. Expect errors and issue.

When you want to pre-define a set of Namespaces with Annotations, Labels all in one place at the same time.
This also serves as a policy enforcement mechanism so that and changes in the Namespace's Annotations and Labels
not allowed by the configuration are immediately reverted/corrected.

## Usage

### Local

```shell
make tidy
make build
knamespacer -debug -config examples/namespaces.yaml
```

### Helm

`helm upgrade --install --namespace knamespacer --create-namespace knamespacer oci://ghcr.io/ejether/knamespacer/charts/knamespacer`
