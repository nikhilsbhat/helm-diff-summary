# helm-diff-summary

[![CI](https://github.com/nikhilsbhat/helm-diff-summary/actions/workflows/ci.yml/badge.svg)](https://github.com/nikhilsbhat/helm-diff-summary/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/nikhilsbhat/helm-diff-summary)](https://goreportcard.com/report/github.com/nikhilsbhat/helm-diff-summary)
[![Go Reference](https://pkg.go.dev/badge/github.com/nikhilsbhat/helm-diff-summary.svg)](https://pkg.go.dev/github.com/nikhilsbhat/helm-diff-summary)
[![License](https://img.shields.io/badge/license-MIT-blue)](https://github.com/nikhilsbhat/helm-diff-summary/blob/main/LICENSE)
[![Release](https://img.shields.io/github/v/release/nikhilsbhat/helm-diff-summary)](https://github.com/nikhilsbhat/helm-diff-summary/releases)
[![Downloads](https://img.shields.io/github/downloads/nikhilsbhat/helm-diff-summary/total.svg)](https://github.com/nikhilsbhat/helm-diff-summary/releases)

Review Helm changes like a Terraform plan.

`helm-diff-summary` turns noisy [`helm diff`](https://github.com/databus23/helm-diff) output into a compact deployment summary that is easy to read in terminals, CI logs, pull requests, and chat notifications.

```bash
helm diff upgrade app ./chart --allow-unreleased --output diff | helm-diff-summary
```

```text
+------------+-------------------+-------------+--------+----------+-----------+---------+
| KIND       | NAME              | NAMESPACE   | ACTION | SEVERITY | CATEGORY  | CHANGES |
+------------+-------------------+-------------+--------+----------+-----------+---------+
| Deployment | sample-api        | production  | UPDATE | HIGH     | WORKLOAD  |      12 |
| ConfigMap  | sample-config     | production  | UPDATE | MEDIUM   | CONFIG    |       3 |
| Service    | sample            | production  | CREATE | LOW      | NETWORK   |       4 |
+------------+-------------------+-------------+--------+----------+-----------+---------+

Plan: 1 to create, 2 to update, 0 to delete.
```

## Why Use It?

Raw Helm diffs are useful, but they become hard to review when charts grow, generated manifests are large, or deployment changes need approval in a pull request.

`helm-diff-summary` helps platform and application teams answer the questions that matter before a release:

* What Kubernetes resources will change?
* Are resources being created, updated, or deleted?
* Are risky resources or namespaces involved?
* Should CI fail because a delete or high-severity policy violation was detected?
* Can the same summary be sent to Slack, Microsoft Teams, Google Chat, or a webhook?

## Features

* Summarizes `helm diff upgrade --output diff` into a Terraform-style plan
* Detects `CREATE`, `UPDATE`, and `DELETE`
* Adds approximate logical change counts per resource
* Classifies resources by severity and category
* Supports custom policy checks from `helm-diff-summary.yaml`
* Can fail CI on deletes or policy severity thresholds
* Renders terminal tables, JSON, and YAML
* Sends summaries to Slack, Microsoft Teams, Google Chat, and generic webhooks
* Works without Kubernetes cluster access because it reads `helm diff` output from stdin
* Ships as a single binary, Docker image, and Homebrew formula

## Installation

### Go

```bash
go install github.com/nikhilsbhat/helm-diff-summary@latest
```

### Release Binary

Download signed release artifacts from the [releases page](https://github.com/nikhilsbhat/helm-diff-summary/releases).

### Homebrew

```bash
brew tap nikhilsbhat/stable https://github.com/nikhilsbhat/homebrew-stable.git
brew install nikhilsbhat/stable/helm-diff-summary
```

Install a specific version:

```bash
brew install nikhilsbhat/stable/helm-diff-summary@0.2.5
```

### Docker

```bash
docker pull ghcr.io/nikhilsbhat/helm-diff-summary:latest
docker pull ghcr.io/nikhilsbhat/helm-diff-summary:<github-release-tag>
```

### Build From Source

```bash
git clone https://github.com/nikhilsbhat/helm-diff-summary.git
cd helm-diff-summary
make local/build
```

## Prerequisites

Install Helm and the `helm diff` plugin:

```bash
helm plugin install https://github.com/databus23/helm-diff
```

## Usage

### Basic

```bash
helm diff upgrade my-release ./chart \
  --output diff | helm-diff-summary
```

### Fresh Install Preview

```bash
helm diff upgrade my-release ./chart \
  --allow-unreleased \
  --output diff | helm-diff-summary
```

### JSON or YAML Output

```bash
helm diff upgrade my-release ./chart \
  --output diff | helm-diff-summary -o json

helm diff upgrade my-release ./chart \
  --output diff | helm-diff-summary -o yaml
```

### Fail CI on Deletes

```bash
helm diff upgrade my-release ./chart \
  --output diff | helm-diff-summary --fail-on-delete
```

### Fail CI on Policy Severity

```bash
helm diff upgrade my-release ./chart \
  --output diff | helm-diff-summary --fail-on high
```

Supported thresholds are `low`, `medium`, `high`, and `critical`.

## CI Examples

### GitHub Actions

```yaml
name: Helm diff summary

on:
  pull_request:

jobs:
  helm-diff-summary:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: azure/setup-helm@v4

      - uses: actions/setup-go@v5
        with:
          go-version: stable

      - name: Install helm-diff
        run: helm plugin install https://github.com/databus23/helm-diff

      - name: Install helm-diff-summary
        run: go install github.com/nikhilsbhat/helm-diff-summary@latest

      - name: Summarize Helm diff
        run: |
          helm diff upgrade my-release ./chart \
            --allow-unreleased \
            --namespace production \
            --output diff | helm-diff-summary --fail-on high
```

### GitLab CI

```yaml
helm-diff-summary:
  image: golang:1
  script:
    - curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
    - helm plugin install https://github.com/databus23/helm-diff
    - go install github.com/nikhilsbhat/helm-diff-summary@latest
    - helm diff upgrade my-release ./chart --allow-unreleased --output diff | helm-diff-summary --fail-on high
```

## Policies

Create `helm-diff-summary.yaml` in the working directory to flag risky changes.

```yaml
policies:
  - name: resource-deletion
    action: DELETE
    severity: CRITICAL
    message: resource deletion detected

  - name: production-namespace
    namespace: production
    severity: HIGH
    message: change detected in production namespace

  - name: crd-update
    kind: CustomResourceDefinition
    action: UPDATE
    severity: CRITICAL
    message: CRD modification detected

  - name: large-change
    min_changes: 100
    severity: MEDIUM
    message: large resource change detected
```

See [`helm-diff-summary.sample.yaml`](helm-diff-summary.sample.yaml) for a fuller example.

Supported policy fields:

| Field | Description |
| --- | --- |
| `name` | Unique policy name |
| `kind` | Kubernetes resource kind |
| `category` | Resource category |
| `action` | Change action: `CREATE`, `UPDATE`, or `DELETE` |
| `namespace` | Kubernetes namespace |
| `severity` | `LOW`, `MEDIUM`, `HIGH`, or `CRITICAL` |
| `message` | Violation message |
| `min_changes` | Minimum changed lines threshold |

## Notifications

Send deployment summaries and policy violations to external systems.

### Slack

```bash
export SLACK_WEBHOOK_URL=https://hooks.slack.com/services/xxx

helm diff upgrade app ./chart \
  --output diff | helm-diff-summary --notify slack
```

### Microsoft Teams

```bash
export TEAMS_WEBHOOK_URL=https://example.webhook.office.com/xxx

helm diff upgrade app ./chart \
  --output diff | helm-diff-summary --notify teams
```

### Google Chat

```bash
export GCHAT_WEBHOOK_URL=https://chat.googleapis.com/xxx

helm diff upgrade app ./chart \
  --output diff | helm-diff-summary --notify gchat
```

### Generic Webhook

```bash
export WEBHOOK_URL=https://example.com/webhook

helm diff upgrade app ./chart \
  --output diff | helm-diff-summary --notify webhook
```

Multiple targets can be used together:

```bash
helm diff upgrade app ./chart \
  --output diff | helm-diff-summary --notify slack,teams
```

Webhook credentials are read from environment variables so they do not appear in shell history, CI logs, or process lists.

## How It Works

`helm-diff-summary` parses resource headers emitted by `helm diff`.

```text
production, sample-api, Deployment (apps) has changed:
```

From each resource block it extracts the namespace, name, kind, action, severity, and logical change count.

Logical change counts are intentionally approximate. For an update like this:

```diff
- image: app:v1
+ image: app:v2
```

the tool reports one logical update instead of two raw diff lines.

## Current Limitations

This tool currently parses unified diff output. It does not yet perform semantic YAML diffing, Kubernetes-aware field comparison, or exact field-level mutation analysis. Formatting-only changes may still appear as updates.

## Documentation

Generated command documentation is available in [`docs/doc/helm-diff-summary.md`](docs/doc/helm-diff-summary.md).

## Development

```bash
go test ./...
go run . < diff.txt
```

Generate command documentation:

```bash
make generate/document
```

Build a local binary:

```bash
make local/build
```

## Community And Discovery

If this project helps your Helm or GitOps workflow, a star, issue, or example workflow goes a long way.

Recommended GitHub topics for the repository:

```text
helm, helm-diff, kubernetes, gitops, ci-cd, devops, platform-engineering, sre, terraform-plan, pull-request, argocd, fluxcd
```

## Related Projects

* [`helm-diff`](https://github.com/databus23/helm-diff)
* [`tf-summarize`](https://github.com/dineshba/tf-summarize)

## License

MIT
