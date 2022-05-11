# Contributing Guide

Badgr has been developed under the auspices of the Brigade project and as such
follows all of the practices and policies laid out in the main
[Brigade Contributor Guide](https://docs.brigade.sh/topics/contributor-guide/).
Anyone interested in contributing to Badgr should familiarize themselves
with that guide _first_.

The remainder of _this_ document only supplements the above with things specific
to this project.

## Running `make hack-kind-up`

As with the main Brigade repository, running `make hack-kind-up` in this
repository will utilize [ctlptl](https://github.com/tilt-dev/ctlptl) and
[KinD](https://kind.sigs.k8s.io/) to launch a local, development-grade
Kubernetes cluster that is also connected to a local Docker registry.

> ⚠️&nbsp;&nbsp; Despite Badgr existing on the fringes of the Brigade ecosystem,
> Badgr has no dependency on Brigade. Running `make-hack-up` will _not_ launch
> Brigade in the new cluster.

## Running `tilt up`

As with the main Brigade repository, running `tilt up` will build and deploy
project code (Badgr, in this case) from source.
