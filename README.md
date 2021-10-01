# Badgr

![badgr](https://badgr.brigade2.io/v1/github/checks/brigadecore/badgr/badge.svg?appID=99005)
[![codecov](https://codecov.io/gh/brigadecore/badgr/branch/main/graph/badge.svg?token=N1SQx2TZt0)](https://codecov.io/gh/brigadecore/badgr)
[![Go Report Card](https://goreportcard.com/badge/github.com/brigadecore/badgr)](https://goreportcard.com/report/github.com/brigadecore/badgr)

Badgr's creators love using [shields.io](https://shields.io/), for displaying
various project statuses in our READMEs, but it doesn't support something we
wanted badly-- the ability to generate a badge based on the results of a GitHub
check suite associated with a specific GitHub App.

Given a repo owner, repo name, and optional branch name, Badgr queries for
GitHub check suite results, consolidates them into a single status (by selecting
the "most severe"<sup>*</sup> among the results), then delegates to
[shields.io](https://shields.io/) to serve the corresponding badge. If a GitHub
App ID is also specified, the badge will reflect _only_ the results of check
suites associates with that App.

<sup>*</sup>Here is how Badgr evaluates check suite severity, from least severe
to most:

* __Passed:__ Check suite has run to completion and succeeded.
* __In Progress:__ One or more checks in the check suite have progressed past
  the queued state, but not all checks are complete.
* __Queued:__ No check in the check suite is either complete or in progress.
* __Neutral:__ Check suite has run to completion and neither failed nor
  succeeded.
* __Canceled:__ Check suite has been voluntarily terminated by a user or some
  other process.
* __Action Required:__ Check suite has run to completion but some action is
  required from a user.
* __Timed Out:__ Check suite has timed out.
* __Failed:__ Check suite has run to completion and failed.
* __Unknown:__ Badgr has been unable to determine the check suite's status.

Badgr also uses Redis to cache results to avoid getting rate limited by the
GitHub Checks API. The cache is composed of warm and cold layers. The warm layer
caches results short term to balance the need for up-to-date results against the
desire to not be rate limited. The cold layer caches results longer term to
return a _relatively_ recent result in the event of a communication failure with
GitHub.

## Installation

Prerequisites:

* A Kubernetes cluster for which you have the `admin` cluster role

* `kubectl` and `helm` (commands below require Helm 3.7.0+)

### 1. Install Badgr

For now, we're using the [GitHub Container Registry](https://ghcr.io) (which is
an [OCI registry](https://helm.sh/docs/topics/registries/)) to host our Helm
chart. Helm 3.7 has _experimental_ support for OCI registries. In the event that
the Helm 3.7 dependency proves troublesome for users, or in the event that this
experimental feature goes away, or isn't working like we'd hope, we will revisit
this choice before going GA.

First, be sure you are using
[Helm 3.7.0](https://github.com/helm/helm/releases/tag/v3.7.0) or greater and
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
    --version v1.0.0 > ~/badgr-values.yaml
```

Edit `~/badgr-values.yaml`, making the following changes:

* `host`: Set this to the host name where you'd like Badgr to be accessible.

Save your changes to `~/badgr-values.yaml` and use the following command to
install the gateway using the above customizations:

```console
$ helm install badgr oci://ghcr.io/brigadecore/badgr \
    --version v1.0.0 \
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

## Code of Conduct

Participation in the Brigade project is governed by the
[CNCF Code of Conduct](https://github.com/cncf/foundation/blob/master/code-of-conduct.md).
