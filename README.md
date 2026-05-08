# helm-diff-summary

A Terraform-style summary tool for `helm diff`.

`helm-diff-summary` converts noisy Helm diff output into a concise, human-readable table showing:

* What Kubernetes resources are changing
* Whether resources are being created, updated, or deleted
* Approximate logical change counts
* A summarized deployment plan

Inspired by:

* [helm-diff](https://github.com/databus23/helm-diff)
* [tf-summarize](https://github.com/dineshba/tf-summarize)

---

# Why?

Raw `helm diff` output becomes difficult to review for:

* large charts
* GitOps workflows
* CI/CD pipelines
* Pull requests
* platform engineering teams

Example:

```diff
- replicas: 2
+ replicas: 3
- image: app:v1
+ image: app:v2
```

After a few hundred lines, identifying the actual impact becomes painful.

`helm-diff-summary` provides a higher-level overview similar to Terraform plans.

---

# Features

* Summarizes Helm diff output in a table
* Detects:
    * CREATE
    * UPDATE
    * DELETE
* Counts logical resource changes
* Works with:
    * `helm diff upgrade`
    * `--allow-unreleased`
* CI/CD friendly
* Lightweight single binary
* No Kubernetes cluster access required

---

# Example

## Input

```bash
helm diff upgrade sample ./chart \
  --allow-unreleased \
  -n crossplane-system \
  --output diff | ./helm-diff-summary
```

## Output

```text
+------------+-------------------+-------------------+---------+---------+
| KIND       | NAME              | NAMESPACE         | ACTION  | CHANGES |
+------------+-------------------+-------------------+---------+---------+
| Deployment | sample            | crossplane-system | CREATE  | 12      |
| Service    | sample            | crossplane-system | CREATE  | 4       |
| ConfigMap  | sample-config     | crossplane-system | CREATE  | 7       |
+------------+-------------------+-------------------+---------+---------+

Plan: 3 to create, 0 to update, 0 to delete.
```

---

# Installation

* Recommend installing released versions. Release binaries are available on the [releases](https://github.com/nikhilsbhat/helm-diff-summary/releases) page.

#### Prerequisites

Install:

* Helm
* helm-diff plugin
* Go 1.22+

Install helm-diff:

```bash
helm plugin install https://github.com/databus23/helm-diff
```

#### Homebrew

Install `helm-diff-summary` on `macOS`

```shell
brew tap nikshilsbhat/stable git@github.com:nikhilsbhat/homebrew-stable.git
# for latest version
brew install nikshilsbhat/stable/helm-diff-summary
# for specific version
brew install nikshilsbhat/stable/helm-diff-summary@0.2.5
```

Check [repo](https://github.com/nikhilsbhat/homebrew-stable) for all available versions of the formula.

#### Docker

Latest version of docker images are published to [ghcr.io](https://github.com/nikhilsbhat/helm-diff-summary/pkgs/container/helm-diff-summary), all available images can be found there. </br>

```bash
docker pull ghcr.io/nikhilsbhat/helm-diff-summary:latest
docker pull ghcr.io/nikhilsbhat/helm-diff-summary:<github-release-tag>
```

#### Build from Source

1. Clone the repository:
    ```sh
    git clone https://github.com/nikhilsbhat/helm-diff-summary.git
    cd helm-diff-summary
    ```
2. Build the project:
    ```sh
    make local.build
    ```

---

# Usage

## Basic

```bash
helm diff upgrade my-release ./chart \
  --output diff | ./helm-diff-summary
```

---

## Fresh Install

```bash
helm diff upgrade my-release ./chart \
  --allow-unreleased \
  --output diff | ./helm-diff-summary
```

---

## Namespace

```bash
helm diff upgrade my-release ./chart \
  -n production \
  --output diff | ./helm-diff-summary
```

---

# How It Works

`helm-diff-summary` parses the structured resource headers emitted by `helm diff`.

Example:

```text
crossplane-system, sample, Deployment (apps) has been added:
```

From this it extracts:

* namespace
* resource name
* resource kind
* action type

It then counts meaningful diff lines inside the resource block.

---

# Logical Change Counting

Raw unified diffs tend to overcount changes.

Example:

```diff
- image: app:v1
+ image: app:v2
```

Technically this is:

* 1 deletion
* 1 addition

But semantically it is a single field update.

To make output more human-friendly:

* CREATE → counts additions
* DELETE → counts deletions
* UPDATE → counts additions only

This produces Terraform-style summaries instead of raw diff math.

```shell
# Sample output on fresh installation 
+----------------+------------------------------+-----------+--------+---------+
| KIND           | NAME                         | NAMESPACE | ACTION | CHANGES |
+----------------+------------------------------+-----------+--------+---------+
| DaemonSet      | fluentd-elasticsearch        | default   | CREATE |      53 |
| ReplicaSet     | frontend                     | default   | CREATE |      20 |
| Function       | function-patch-and-transform | default   | CREATE |       7 |
| CronJob        | hello                        | default   | CREATE |      19 |
| Configuration  | my-configuration             | default   | CREATE |       7 |
| Pod            | nginx                        | default   | CREATE |      14 |
| Pod            | nginx-2                      | default   | CREATE |      10 |
| Job            | pi                           | default   | CREATE |      13 |
| Provider       | provider-aws                 | default   | CREATE |       7 |
| Deployment     | sample                       | default   | CREATE |      45 |
| Service        | sample                       | default   | CREATE |      20 |
| ServiceAccount | sample                       | default   | CREATE |      10 |
| ConfigMap      | sample-config-map            | default   | CREATE |      10 |
| ConfigMap      | sample-config-map-json       | default   | CREATE |      16 |
| ConfigMap      | sample-config-map-test       | default   | CREATE |      17 |
| ConfigMap      | sample-config-map-test-2     | default   | CREATE |      13 |
| ConfigMap      | sample-config-map-yaml       | default   | CREATE |      15 |
| ConfigMap      | test-cm                      | default   | CREATE |       7 |
| StatefulSet    | web                          | default   | CREATE |      35 |
+----------------+------------------------------+-----------+--------+---------+
```

```shell
# Sample output for following diff

# default, sample-config-map-test-2, ConfigMap (v1) has changed:
#  # Source: sample/templates/configmap.yaml
#  apiVersion: v1
#  kind: ConfigMap
#  metadata:
#    name: sample-config-map-test-2
#    namespace: default
#  data:
#    config: |
#      - name: test
#        image: ghcr.io/virtu/test:v2.2.0
#      - name: virtu
#-       type: foo
#+       type: foor
#      - name: foolist
#        type: bar

+-----------+--------------------------+-----------+--------+---------+
| KIND      | NAME                     | NAMESPACE | ACTION | CHANGES |
+-----------+--------------------------+-----------+--------+---------+
| ConfigMap | sample-config-map-test-2 | default   | UPDATE |       1 |
+-----------+--------------------------+-----------+--------+---------+
```
---

# Current Limitations

This tool currently uses unified diff parsing.

It does not yet perform:

* semantic YAML diffing
* Kubernetes-aware field comparison
* exact field-level mutation analysis

Therefore:

* counts are approximate
* formatting changes may still appear as updates

---

# CI/CD Usage

Useful for:

* GitHub Actions
* GitLab CI
* ArgoCD
* FluxCD
* Atlantis-style workflows
* PR review automation

Example:

```bash
helm diff upgrade app ./chart \
  --output diff | ./helm-diff-summary
```

---

# Development

## Run

```bash
go run . < diff.txt
```

---

## Test Against Helm Diff

```bash
helm diff upgrade sample ./chart \
  --allow-unreleased \
  --output diff > diff.txt

cat diff.txt | go run .
```

# License

MIT
