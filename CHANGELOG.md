# Changelog

All notable changes to Marvin will be documented in this file.

The format is based on Keep a Changelog, and this project aims to follow
semantic versioning after the first tagged release.

## Unreleased

### Added

- Cost value parsing for generic three-letter currency code prefixes and suffixes.
- Empty CSV data detection for header-only cost exports.
- Report format aliases for `text` and `md`.
- Config sample generation with `marvin config sample`.
- Config validation with `marvin config validate`.
- Release version injection for tagged CLI binaries.
- Release build workflow for tagged CLI binaries.
- Standard input support with `marvin analyze -`.
- Golden test coverage for terminal and JSON reports.
- Cost Explorer CSV export documentation.
- Sample CSV generation with `marvin sample`.
- Support for additional AWS cost CSV column names and timestamp formats.
- Stable analyze exit codes and `--fail-on-warning`.
- Report file output with `marvin analyze --output`.
- Golden test coverage for Markdown reports.
- Ignore-service rules for excluding services from reports and warnings.
- JSON config loading for reusable warning rules.
- Markdown and JSON report output for `marvin analyze`.
- Terminal report output for `marvin analyze`.
- Budget, service, and growth warning evaluation.
- Month-over-month cost comparison helper.
- Cost aggregation helpers for total, service, and month spend.
- Cost Explorer CSV parser and sample fixture.
- Repository line-ending normalization for cross-platform CI.
- Initial Go CLI scaffold.
- Cross-platform Go CI workflow.
- Open-source project metadata.
