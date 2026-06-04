# Changelog

All notable changes to Marvin will be documented in this file.

The format is based on Keep a Changelog, and this project aims to follow
semantic versioning after the first tagged release.

## Unreleased

### Added

- Include-service filtering with `marvin analyze --only-service`.
- `include_services` config setting for reusable service include filters.
- JSON Schema for `marvin.json` config files.
- CSV report output with `marvin analyze --format csv`.
- Minimum service spend report filtering with `marvin analyze --min-service-spend`.
- `min_service_spend` config setting for reusable service spend report filters.
- Expanded GitHub issue and pull request templates for contributors.
- Architecture documentation for contributors.
- Troubleshooting guide for common CSV, config, and report issues.
- Linux race-detector job in CI.
- `format` config setting for reusable report output format selection.
- `fail_on_warning` config setting for reusable warning exit behavior.
- Currency filtering with `marvin analyze --currency`.
- `currency` config setting for reusable currency filters.
- `output_path` config setting for reusable report file output.
- Cost Explorer CSV validation with `marvin validate`.
- Configuration reference documentation.
- Standard input support for `marvin config validate -`.
- Gzip-compressed CSV input support.
- Currency-aware report formatting and mixed-currency analysis validation.
- Empty filtered-result detection for analysis filters.
- `from_month` and `to_month` config settings for reusable month range filters.
- Month range filtering with `marvin analyze --from` and `--to`.
- `top_services` config setting for reusable service row limits.
- Service row limiting with `marvin analyze --top-services`.
- Multi-file analysis with `marvin analyze <file> [more.csv ...]`.
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
