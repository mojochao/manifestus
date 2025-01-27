# manifestus

`manifestus` is a CLI to render Kubernetes resource manifests for application
stack workloads from various sources of Kubernetes resource manifests defined
in a single, loaded *Renderfile* configuration.

While the rendered manifests may be imperatively applied to a Kubernetes cluster,
it is intended *primarily for management of Kubernetes cluster state in pull-based
[GitOps](https://opengitops.dev/) workflows implemented in terms of
[Rendered Manifests](https://medium.com/@PlanB./rendered-manifests-pattern-the-new-standard-for-gitops-c0b9b020f3b6).

The Latin word "manifestus" means "clear," "public,", "obvious", or "transparent".
It is hoped that this tool will make Kubernetes resource manifests, and changes
to them, the same.
The word "manifestus" may also have connotations of "notorious" or "infamous".
Time will tell.

`manifestus` is primarily intended for use by Platform Engineers in the
configuration and change management of their platform stacks across multiple
environments, although it may be also useful to application developers and/or
operators working with application stacks of their own. 

- [Overview](#overview)
- [A simple change management workflow](#a-simple-change-management-workflow)
- [Installation](#installation)
  - [Local binaries](#local-binaries)
  - [Docker image](#docker-image)
- [Configuration](#configuration)
  - [Apps configuration](#apps-configuration)
  - [Releases configuration](#releases-configuration)
  - [Kustomizations configuration](#kustomizations-configuration)
  - [Bundles configuration](#bundles-configuration)
- [Usage](#usage)
  - [Getting help](#getting-help)
  - [General conventions](#general-conventions)
  - [Listing apps](#listing-apps)
  - [Targeting specific apps](#targeting-specific-apps)
  - [Listing outputs of the rendered manifests](#listing-outputs-of-the-rendered-manifests)
  - [Previewing rendered manifests](#previewing-rendered-manifests)
  - [Updating rendered manifests](#updating-rendered-manifests)
  - [Checking rendered manifests](#checking-rendered-manifests)
- [Prior art](#prior-art)
- [References](#references)

## Overview

`manifestus` was designed with a few primary goals in mind:

- to provide a simple, consistent, and repeatable way to render Kubernetes
  resource manifests for workload stacks of of multiple applications defined in
  a single, simple configuration file
- to reuse common, existing, well-known tools like [Helm](https://helm.sh/),
  [Helmfile](https://helmfile.readthedocs.io/), and [Kustomize](https://kustomize.io/)
  to render the manifests with a consistent interface, using existing
  source artifacts
- to provide a way to manage static manifests, whether local or remote, that
  are not "rendered" but should still be included in the rendered manifests
- to be easily extensible to other tools and sources of Kubernetes resource
  manifests in the future

It is somewhat opinionated, but also quite flexible:

- instead of storing the rendered manifests in environment-specific branches,
  it stores them in environment-specific output directories in the `main` branch
- it organizes the rendered manifests in a directory structure that mirrors the
  loaded configuration structure
- it assumes full ownership of the contents of the output directories, and
  deletes any files that are not present in the rendered manifests
- it currently supports rendering manifests from [Helm chart](https://helm.sh/),
  [Helmfile](https://helmfile.readthedocs.io/), and [Kustomization](https://kustomize.io/)
  sources, but may be extended to support other sources such as [CUE](https://cuelang.org/)
  or [KCL](https://www.kcl-lang.io/) in the future
- it supports managing static manifests in bundles, which are useful for storing
  CRDs, ExternalSecrets, or other static manifests not rendered, but instead
  copied from local filesystem or remote HTTP manifest sources

## A simple change management workflow

The purpose of the `manifestus` CLI is to render Kubernetes resource manifests
for application stacks of Kubernetes clusters stored in configuration
Git repositories for application by GitOps controllers watching them for
commits containing changes to reconcile.

In general, it is expected that rendered manifests are made in response to some
configuration *Change Request* (CR) being created in some issue tracking,
service desk, or configuration management system for accountability, compliance,
and tracking purposes.

Although it is expected that you will have your own workflows tailored to your
needs, a simple change management workflow using `manifestus` is provided here
as an example.

In this example, the rendered manifests and their source files are stored in a
Git repository `example-system-k8s` in the `main` branch.
Source files are stored in `sources/<cluster_name>/` directories, and rendered
manifests are written to `manifests/<cluster_name>/` directories.

Changes to the rendered manifests are made in response to some configuration
*Change Request* event in some issue tracking or configuration management system,
persisted for auditability, and given a unique identifier for tracking purposes.

Our *Change Request* involves three different roles:

- *Stakeholders* request a change and provide the necessary details
- *Editors* fulfill change requests by making the desired changes in a CR branch
- *Reviewers* review the changes made by the Editor in the CR branch

The lifecycle of a configuration Change Request is defined here as follows:

1. *Stakeholders* create a CR for modifications to cluster configuration in some
   issue tracking, service desk, or change management system.
2. An *Editor* picks up the change request and creates a new CR branch off the
   `main` branch to make the modifications to the cluster configuration
3. The *Editor* makes the changes to the configuration sources and renders the
   manifests with the `manifestus` CLI for local inspection
4. Once satisfied with the changes, the *Editor* adds the changes to a single
   commit and pushes the changes to the repository remote and opens a pull
   request (PR) for review by *Reviewers* with the necessary expertise
5. Once approval is given by the desired number of *Reviewers*, the *Editor* merges
   the changes to the `main` branch in a fast-forward commit to keep the change
   history linear and destroys the remote CR branch
6. Once merged and applied, the *Editor* closes the issue and the CR is marked as
   complete
7. Once CR is marked as complete, the *Stakeholders* are notified that the CR is
   complete and the changes have been applied to the cluster

## Installation

### Local binaries

`manifestus` is a Golang app that may be installed with `go install`:

```shell
go install github.com/mojochao/manifestus@latest
```

`manifestus` uses the following additional tools to render manifests:

- the [`helm`](https://helm.sh/docs/helm/) CLI to render manifests from local or
  remote [Helm charts](https://helm.sh/docs/topics/charts/)
- the [`helmfile`](https://helmfile.readthedocs.io/) and [`helm`](https://helm.sh/docs/helm/)
  CLIs to render manifests from local or remote [Helm charts](https://helm.sh/docs/topics/charts/)
- the [`kustomize`](https://kustomize.io/) CLI to render manifests from local or remote
  [kustomizations](https://kubectl.docs.kubernetes.io/references/kustomize/glossary/#kustomization)

These tools must be installed and available in the `PATH` for the `manifestus`
CLI to work correctly.

These tools are easily installable with the [`brew`](https://brew.sh/) package
manager:

```shell
brew install helm helmfile kustomize
```

Alternatively, they may be installed with the [`asdf`](https://asdf-vm.com/)
package manager:

```bash
plugins=("helm" "helmfile" "kustomize")
for plugin in "${plugins[@]}"; do
  asdf plugin add "$plugin" && asdf install "$plugin"
done
asdf install
```

If you wish to render a chart release directly from its chart source in a Git
repository, you will also need to install the [`helm-git`](https://github.com/aslafy-z/helm-git)
plugin.

You do *not* need to install the [`helm-diff`](https://github.com/databus23/helm-diff)
plugin, as `helm` is only used to template chart manifests, and not to manage
chart releases in the cluster.

As the rendered manifests are modified in your locally cloned Git repository,
you can examine changes to them, and the source changes responsible for the
changed renders,  with the `git` CLI, or your favorite Git GUI.

An additional benefit of pushing these changes to a change request branch off
the `main`branch is that both the source and output changes will be displayed
in the diff view of the pull request (PR) when reviewed, before acceptance and
merge back to the `main`branch for application by the cluster GitOps controller.

> Note: The `diff` CLI is also required when running the `manifestus check` command.
> It is probably already installed on your system, but if it is not, you will
> need to install it as well.

### Docker image

The `manifestus` app is also available as a Docker image with all the necessary
tools pre-installed. The following example demonstrates use of the `manifestus`
Docker image to list apps defined in the local `testdata/renderfile.yaml` file.

```shell
docker run --rm -it -v $(pwd)/testdata:/testdata ghcr.io/mojochao/manifestus:latest -- apps -f /testdata/renderfile.yaml
```

## Configuration

Configuration for `manifestus` is defined in a `renderfile.yaml` Renderfile.
By default, `manifestus` looks for this file in the current working directory.
The name and location of the configuration can be overridden with the
`--config-file` flag.

Its schema as follows, from the top:

```yaml
# Root renderfile object fields
renderfile:
  schema: str  # Required but '1' is the only version at this point
  apps: []App  # Required list of apps to render
```

### Apps configuration

Each `App` object is defined as follows:

```yaml
# App object fields
name: str                        # Required name of the app
disabled: bool                   # Optional flag to disable the app
releases: []Release              # Optional Helm chart releases
kustomizations: []Kustomization  # Optional kustomizations
bundles: []Bundle                # Optional static manifest bundles
```

### Releases configuration

Each `Release` object in `.renderfile.apps.*.releases` contains:

```yaml
# Release object fields
name: str       # Required name of the Helm chart release
helmfile: str   # Optional path to a Helmfile, defaults to "helmfile.yaml" in the same directory as the config file in use
namespace: str  # Optional namespace for the release, if not using a Helmfile to specify it
chart: str      # Optional Helm chart name for the release, if not using a Helmfile to specify it
version: str    # Optional Helm chart version for the release, if not using a Helmfile to specify it
values: str     # Optional path to a Helm chart values file for the release, if not using a Helmfile to specify it
```

### Kustomizations configuration

Each `Kustomization` object in `.renderfile.apps.*.kustomizations` contains:

```yaml
# Kustomization object fields
name: str    # Required name of the kustomization
source: str  # Required local path or remote URL to a kustomization.yaml file
```

### Bundles configuration

Each `Bundle` object in `.renderfile.apps.*.bundles` contains:

```yaml
# Bundle object fields
name: str          # Required name of the bundle
data: map[str]str  # Optional arbitrary string data to pass to the bundle renderer for expansion in 'sources' items
sources: []str     # Required list of local paths or remote URLs to static manifests
```


See the [test configuration](../testdata/renderfile.yaml) for a full example.

## Usage

> Pro tip: When using interactively, go "Samuel L. Jackson" on your manifests
> with an alias:
> 
> ```shell
> alias mf=manifestus
> ```
> 
> Now render like a badass:
> 
> ```shell
> mf render
> ```
>
> Much better, right?

### Getting help

Run `manifestus --help` to see the available commands.

Run `manifestus <command> --help` to see the available options for a specific
command.

### General conventions

The `manifestus` command uses the `stdout` stream for the actual output of
commands, such as the list of apps available or renders of the manifest for one
or more apps.

Informational messages are sent to the `stderr` stream, but may be silenced
with the `--quiet` option.

An exit code of zero always indicates success, while negative ones always
indicate failure. Some commands, such as the `diff` command, use positive codes
to indicate a non-empty diff, in addition to success.

### Listing apps

To list the available apps in the configuration, run:

```shell
manifestus apps
```

### Targeting specific apps

Most `manifestus` commands can be targeted to specific apps by using the`--app`
flag.

```shell
manifestus render --app <app_name>
```

The `--app` flag may be provided multiple times to render multiple apps in a
single invocation.

If the `--app` flag is not provided, all enabled apps in the configuration will
be rendered.

### Targeting specific sources

It is sometimes useful to target specific source types for rendering. This
can be done with the `--type` flag.

```shell
manifestus render --app cert-manager --type bundle
```

The `--type` flag may be provided multiple times to render multiple source types
in a single invocation.

If the `--type` flag is not provided, all source types in the configuration will
be rendered.

Sources may be further targeted by name with the `--name` flag.

```shell
manifestus render --app cert-manager --type bundle --name crds
```

The `--name` flag may be provided multiple times to render multiple sources in
a single invocation.

If the `--name` flag is not provided, all sources of the specified type will be
rendered.

### Listing outputs of the rendered manifests

To list the filenames of the available output files in the configuration, run:

```shell
manifestus outputs
```

### Previewing rendered manifests

To preview the rendered manifests that would be updated in the cluster, just
render them with:

```shell
manifestus render
```

This is useful for examining the complete manifests for the changes that would
be applied to the cluster before pushing the changes in your change request
branch to their origin.

If you want to preview the commands that would be run to render the manifests,
you can use the `--dry-run` flag:

```shell
manifestus render --dry-run
```

### Updating rendered manifests

To update the rendered manifests for the cluster, run:

```shell
manifestus update
```

After this command is run, the rendered manifests will be updated in the output
directory, and the changes can be examined locally with the `git` CLI, or your
Git GUI of choice, before adding them to a commit in your change request branch.

The rendered manifests are output to a directory structure that mirrors the
configuration structure. The output directory is created in the current working
directory by default, but can be overridden with the `--output-dir` flag.

The output directory structure is as follows:

```text
$OUTPUT_DIR/
$OUTPUT_DIR/<app_name>/
$OUTPUT_DIR/<app_name>/<release_name_a>.release.manifest.yaml
$OUTPUT_DIR/<app_name>/<release_name_b>.release.manifest.yaml
$OUTPUT_DIR/<app_name>/<kustomization_name_a>.kustomization.manifest.yaml
$OUTPUT_DIR/<app_name>/<kustomization_name_b>.kustomization.manifest.yaml
$OUTPUT_DIR/<app_name>/<bundle_name_a>.bundle.manifest.yaml
$OUTPUT_DIR/<app_name>/<bundle_name_b>.bundle.manifest.yaml
```

One common category name for bundles is `crds`.

```text
apps/cert-manager/crds.bundle.manifest.yaml
```

Another useful category name might be `externalsecrets`.

```text
apps/grafana/externalsecrets.bundle.manifest.yaml
```

Note that the `<bundle_name>` is sourced by the category name of the bundle.

### Checking rendered manifests

When rendering manifests it is useful to know if the rendered manifests in an
output directory are consistent with its sources. This can be checked with the
`manifestus check` command

```shell
manifestus check
```

If no differences exist, the command will return an exit code of `0`.
If differences do exist, they will be printed as a diff to standard output
and the command will return an exit code of `1`.

## Prior art

The `manifestus` app is inspired by the [Rendered Manifests](https://medium.com/@PlanB./rendered-manifests-pattern-the-new-standard-for-gitops-c0b9b020f3b6)
pattern, which is a new-ish practice for GitOps workflows that uses rendered
manifests as the source of truth for the state of a Kubernetes cluster.

Other tools existing in this space include:

- [holos](https://holos.run/) renders Kubernetes resource manifests from [CUE](https://cuelang.org/) definitions
- [kpt](https://kpt.dev/) renders Kubernetes resource manifests from [Kustomize](https://kustomize.io/) definitions

## References

- the [Kubernetes workloads management docs](https://kubernetes.io/docs/concepts/workloads/management/)
  provide detailed information on configuring Kubernetes workloads with manifests
- the [Rendered Manifests Pattern blog post](https://medium.com/@PlanB./rendered-manifests-pattern-the-new-standard-for-gitops-c0b9b020f3b6)
  introduces the concept of rendered manifests and how they can be used to manage
  Kubernetes resource state in GitOps workflows
- the [OpenGitOps website](https://opengitops.dev/) provides a comprehensive
  overview of GitOps workflows and practices
- the [Helm docs](https://helm.sh/docs/) provide detailed info on how to use
  `helm` to build manifests for Helm chart releases
- the [Helmfile docs](https://helmfile.readthedocs.io/) provide detailed info
  on how to use `helmfile` to build manifests for multiple Helm chart releases
- the [Kustomize docs](https://kustomize.io/) provide detailed info on how to
  use `kustomize` to build Kubernetes resources
