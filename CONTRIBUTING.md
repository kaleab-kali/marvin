# Contributing

Thanks for helping improve Marvin.

Marvin is currently in the CSV MVP phase. The first useful version should read
AWS Cost Explorer CSV exports without requiring AWS credentials.

## Local Setup

Requirements:

- Go 1.22 or newer.
- Git.

Run the checks:

```sh
go test ./...
go vet ./...
go build ./cmd/marvin
```

## Pull Requests

Keep pull requests focused and reviewable:

- Prefer one behavior change per PR.
- Add tests with behavior changes.
- Update documentation when user-facing behavior changes.
- Avoid unrelated formatting churn.
- Explain the motivation and validation steps in the PR description.

## Commit Style

Use short imperative commit messages:

```text
Add CSV parser tests
Document Cost Explorer export columns
```

## Project Direction

For v0.1, do not add live AWS API access, credentials handling, scheduled jobs,
Slack, or email integrations. Those belong after the local CSV workflow is
useful and tested.
