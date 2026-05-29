# Architecture

Marvin is a small Go CLI that analyzes exported AWS Cost Explorer CSV files
locally. The v0.1 architecture intentionally avoids AWS credentials, live AWS
API calls, background jobs, and hosted services.

## Package Layout

```text
cmd/marvin        CLI entry point.
internal/cli      Command parsing, input/output wiring, and exit codes.
internal/config   JSON config loading and validation.
internal/cost     Cost Explorer CSV parsing, filtering, aggregation, comparisons, and warnings.
internal/report   Summary model and terminal, Markdown, JSON, and CSV writers.
fixtures          Sample Cost Explorer CSV data for docs and tests.
examples          Example Marvin config files.
docs              User and contributor documentation.
```

## Analyze Flow

`cmd/marvin` delegates to `internal/cli.RunWithIO`, which keeps command
execution testable by accepting stdin, stdout, and stderr as interfaces.

The `analyze` command follows this flow:

1. Parse CLI flags and optional config files.
2. Read one or more CSV inputs from files, gzip files, or standard input.
3. Parse records with `internal/cost.ParseCostExplorerCSV`.
4. Apply record filters such as included services, ignored services, and month
   ranges.
5. Validate that at least one record remains and that all remaining records use
   a single currency.
6. Build a `report.Summary` from totals, service spend, monthly spend,
   month-over-month comparisons, and warning rules.
7. Apply report-row presentation options such as minimum service spend and top
   service limits.
8. Write the selected report format.

Record filters affect totals and warnings. Report-row presentation options only
affect rows shown in the service section.

## Config Loading

Config files are JSON and are loaded by `internal/config`. Unknown fields are
rejected so typos do not silently change analysis behavior. Config settings and
CLI flags are applied in the order they appear on the command line, so later
flags can override earlier config values.

The public schema lives in `docs/marvin.schema.json`. Runtime validation still
enforces rules that JSON Schema cannot express, such as month range ordering.

## Reporting

`internal/report` uses a shared `Summary` type for all output formats. Terminal
and Markdown formats are optimized for people. JSON and CSV formats are intended
for automation and spreadsheet workflows.

Report writers should not read files, parse CLI flags, or apply record filters.
Those responsibilities stay in `internal/cli` and `internal/cost`.

## Boundaries

Keep v0.1 focused on local Cost Explorer CSV analysis:

- Do not add AWS SDK clients or credential handling.
- Do not add hosted APIs, scheduled jobs, Slack, or email integrations.
- Do not make reports depend on network access.
- Prefer small package-level functions with tests over framework-style wiring.
