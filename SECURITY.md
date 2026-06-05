# Security Policy

Marvin does not require AWS credentials in the CSV MVP phase. Do not include
real account IDs, invoices, credentials, or private billing exports in issues,
pull requests, fixtures, or screenshots.

## Reporting Security Issues

Please do not open a public issue for a security concern. Contact the maintainer
privately through the GitHub profile for `kaleab-kali`.

Include:

- A clear description of the issue.
- Reproduction steps, if available.
- The Marvin version or commit.
- Any relevant logs with secrets removed.

## Supported Versions

Marvin has not reached a stable release yet. Security fixes will target the
current `main` branch until versioned releases begin.

## Automated Checks

Pull requests and `main` branch updates run Go CI, race-detector tests, and
CodeQL analysis. CodeQL also runs weekly on the default branch. Dependabot
monitors GitHub Actions and Go module updates weekly.
