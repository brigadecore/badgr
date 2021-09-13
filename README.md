# Badgr

![badgr](https://badgr.brigade2.io/v1/github/checks/brigadecore/badgr/badge.svg?appID=99005)
[![codecov](https://codecov.io/gh/brigadecore/badgr/branch/main/graph/badge.svg?token=N1SQx2TZt0)](https://codecov.io/gh/brigadecore/badgr)
[![Go Report Card](https://goreportcard.com/badge/github.com/brigadecore/badgr)](https://goreportcard.com/report/github.com/brigadecore/badgr)

Badgr's creators love using [shields.io](https://shields.io/), for displaying
various project statuses in our READMEs, but it doesn't support something we
wanted badly-- the ability to generate a badge based on the results of a GitHub
check suite associated with a specific GitHub App.

Badgr is a simple server that, given a repo owner, repo name, and optional
branch name, can query for and consolidate GitHub check suite results and then
delegate to [shields.io](https://shields.io/) to generate the corresponding
badge. If a GitHub App ID is also specified, the badge will reflect _only_ 
the results of check suites associates with that App.

Badgr also uses Redis to cache results to avoid getting rate limited by the
GitHub Checks API. The cache is composed of warm and cold layers. The warm layer
caches results short term to balance the need for up-to-date results against the
desire to not be rate limited. The cold layer caches results longer term to
return a _relatively_ recent result in the event of a communication failure with
GitHub.

## Installation

Prerequisites:

* A Kubernetes cluster for which you have the `admin` cluster role

* `kubectl` and `helm`

### 1. Install Badgr

For now, we're using the [GitHub Container Registry](https://ghcr.io) (which is
an [OCI registry](https://helm.sh/docs/topics/registries/)) to host our Helm
chart. Helm 3.7 has _experimental_ support for OCI registries. In the event that
the Helm 3.7 dependency proves troublesome for users, or in the event that this
experimental feature goes away, or isn't working like we'd hope, we will revisit
this choice before going GA.

First, be sure you are using
[Helm 3.7.0-rc.3](https://github.com/helm/helm/releases/tag/v3.7.0-rc.3) and
enable experimental OCI support:

```console
$ export HELM_EXPERIMENTAL_OCI=1
```

Since this chart requires some slight bit of custom configuration, we'll need to
create a chart values file with said config.

Use the following command to extract the full set of configuration options into
a file you can modify:

```console
$ helm inspect values oci://ghcr.io/brigadecore/badgr \
    --version v0.1.0 > ~/badgr-values.yaml
```

Edit `~/badgr-values.yaml`, making the following changes:

* `host`: Set this to the host name where you'd like Badgr to be accessible.

Save your changes to `~/badgr-values.yaml` and use the following command to
install the gateway using the above customizations:

```console
$ helm install badgr oci://ghcr.io/brigadecore/badgr \
    --version v0.1.0 \
    --create-namespace \
    --namespace badgr \
    --values ~/badgr-values.yaml
```

### 2. (RECOMMENDED) Create a DNS Entry

If you installed the gateway without enabling support for an ingress controller,
this command should help you find the gateway's public IP address:

```console
$ kubectl get svc badgr \
  --namespace badgr \
  --output jsonpath='{.status.loadBalancer.ingress[0].ip}'
```

If you overrode defaults and enabled support for an ingress controller, you
probably know what you're doing well enough to track down the correct IP without
our help. 😉

With this public IP in hand, edit your name servers and add an `A` record
pointing your domain to the public IP.

## Usage

To use, add the following markdown (with appropriate substitutions where you
see angled brackets) to your `README.md` or any other markdown doc needing such
a badge:

```markdown
![badgr](https://<host name>/v1/github/checks/<user or org name>/<repo name>/badge.svg?branch=<optional branch name>&appID=<optional GitHub App ID>)
```

## Contributing

Badgr is part of the Brigade project and accepts contributions via GitHub pull
requests. The [Contributing](CONTRIBUTING.md) document outlines the process to
help get your contribution accepted.

## Support & Feedback

We have a slack channel!
[Kubernetes/#brigade](https://kubernetes.slack.com/messages/C87MF1RFD) Feel free
to join for any support questions or feedback, we are happy to help. To report
an issue or to request a feature open an issue
[here](https://github.com/brigadecore/badgr/issues)
