# KAB - Kubernetes Application Bundle
A CNAB bundle that allows creation of other declarative CNAB bundles

## Background

[CNAB is a spec](https://github.com/deislabs/cnab-spec) for installing and managing cloud native applications in a
cloud agnostic manner.

The spec also has a section about [Declarative Invocation Images](https://github.com/deislabs/cnab-spec/blob/master/801-declarative-images.md)
which states the following:

> By providing a run tool (/cnab/app/run), a middleware image can remove the necessity to write the imperative portions
> of a CNAB bundle, essentially allowing construction of declarative CNAB bundles. In this model, the middleware image
> provides the tooling necessary for handling CNAB actions. Images layered on top of this middleware merely need to
> describe what entities are being installed, uninstalled, or upgraded.

This project is an implementation of this "Declarative Invocation Image" for applications that want to target kubernetes
as the deployment platform.

## Structure of manifest file

This project expects the following structure of the declarative manifest file:
```yaml
apiVersion: projectriff.io/v1alpha1
kind: Manifest
metadata:
  name: riff-install
  namespace: default
spec:
  resources:
  - name: istio
    namespace: istio-system
    path: https://storage.googleapis.com/knative-releases/serving/previous/v0.3.0/istio.yaml
    checks:
    - jsonpath: .status.phase
      kind: Pod
      pattern: Running
      selector:
        matchLabels:
          istio: sidecar-injector
  - name: riff-build-template
    path: https://storage.googleapis.com/projectriff/riff-buildtemplate/riff-cnb-clusterbuildtemplate-0.2.0-snapshot-ci-63cd05079e1f.yaml
```
The `.spec.resources` section expects a list of resources that make up your product. The resource to be installed could
either be a url or its contents can inlined in the manifest. Please see [types.go](https://github.com/projectriff/cnab-k8s-installer-base/blob/master/pkg/apis/kab/v1alpha1/types.go)
for the complete structure of the manifest.

### Resource Dependencies
Please ensure that a resource's dependencies are defined before the resource itself. To ensure that the resource has
been successfully installed, you can add a `checks` section as shown above. The above example check will ensure that
the `sidecar-injector` Pod is running before the next resource is installed. At the moment only Pod checks are supported.


## Custom Resource Definition
This base bundle defines a CRD named `manifests.projectriff.io`, and it will create objects of this CRD for all bundles
that extend this bundle. This will allow your product's configuration to be stored in the k8s cluster itself. This
project also defines a golang client so that you'r product's config can be looked up programmatically.

## Steps for creating your installer bundle

Install and setup the [duffle cli](https://github.com/deislabs/duffle). To bootstrap your project run:
```bash
$ duffle create foo
```
This should create a directory for your project with the following structure:
```bash
$ tree
.
├── cnab
│   ├── Dockerfile
│   └── app
│       └── run
└── duffle.json
```
Change the Dockerfile to extend from the image for this base bundle and to copy your product's installation template to
`/cnab/app/kab`, something like this:
```dockerfile
FROM sbawaska/kab-cnab:latest

COPY Dockerfile /cnab
COPY app/kab /cnab/app/kab
```

Finally, create your application's manifest as defined above under `app/kab`.
