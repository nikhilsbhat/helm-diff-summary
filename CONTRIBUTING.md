# Contributing

Thanks for helping improve `helm-diff-summary`.

Good contributions include parser fixes, sample Helm diffs, CI examples, policy ideas, notification improvements, and documentation that makes the tool easier to adopt.

## Development

```bash
go test ./...
go run . < diff.txt
```

When changing CLI flags or command help, regenerate the docs:

```bash
make generate/document
```

Before opening a pull request, run:

```bash
go test ./...
go build ./...
```

## Bug Reports

Parser bugs are easiest to fix when the issue includes:

* the `helm diff` command
* a small diff input sample
* the actual summary
* the expected summary
* the `helm-diff-summary` and `helm diff` versions

Please redact secrets, image pull credentials, and private hostnames before sharing diffs.
