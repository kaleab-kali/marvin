# Agent Instructions

These instructions capture project-specific preferences for future agent work in
this repository.

## Repository Direction

- Marvin is a Go CLI for local analysis of exported AWS Cost Explorer CSV files.
- Keep v0.1 focused on CSV input with no AWS credentials or live AWS API access.
- The public README must explain the project, its purpose, status, and usage. Do
  not use the README as an internal checklist or planning document.
- Keep internal planning notes out of public-facing docs unless they are framed
  as a clear roadmap for users and contributors.

## Git Workflow

- Use small, reviewable pull requests.
- Use branch names that describe the actual functionality, not the phase of
  work. Good examples:
  - `feature/parse-cost-explorer-csv`
  - `feature/group-costs-by-service`
  - `docs/document-cost-export-workflow`
- Do not use generic phase branch names such as `phase-1`, `scaffold-phase`, or
  similar.
- Commit progress cleanly with focused commits.
- Use imperative commit messages that describe the change, such as
  `Add Cost Explorer CSV parser`.
- Pull request titles should describe the functionality being added or changed.
- Pull request descriptions should include a short summary and validation steps.
- Do not squash merge unless the user explicitly changes this preference.
- Use normal merge commits for PRs. Do not use squash merge or rebase merge
  unless the user explicitly asks for that specific merge strategy.
- Do not delete local or remote branches after closing or merging PRs unless the
  user explicitly asks for branch deletion.
- Do not rewrite published history unless the user explicitly asks for it.

## Attribution

- Never add Claude, Anthropic, or AI co-author trailers.
- Do not add `Co-authored-by` lines unless the user explicitly provides the
  exact human co-author attribution to use.
- Use the repository's configured Git author only.

## Validation

- For Go changes, run:
  - `gofmt -l cmd internal`
  - `go test ./...`
  - `go vet ./...`
  - `go build ./cmd/marvin`
- If the default Windows Go cache is blocked, use repo-local caches:
  - `GOCACHE=$PWD/.cache/go-build`
  - `GOMODCACHE=$PWD/.cache/gomod`
- Keep `.cache/` ignored.

## Review Standard

- Prefer one behavior change per PR.
- Add tests with behavior changes.
- Update docs when user-facing behavior changes.
- Avoid unrelated formatting churn.
- Keep generated or local-only artifacts out of commits.
