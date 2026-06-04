# Marvin

Marvin is a Go CLI for understanding AWS cost exports without connecting to an
AWS account.

It is designed to read exported AWS Cost Explorer CSV files, group spend by
service and month, and produce plain reports that explain where the bill moved.
The goal is a pessimistic but accurate local cost monitor: no credentials, no
cloud access, and no hidden state required for the first release.

## Status

Marvin is in CSV MVP development. It can analyze exported AWS Cost Explorer CSV
files locally and produce terminal, Markdown, or JSON reports.

The current CLI supports:

- Importing AWS Cost Explorer CSV files.
- Grouping spend by service and month.
- Showing month-over-month changes.
- Emitting budget, service, and growth warnings from command-line thresholds.
- Producing terminal, Markdown, and JSON reports.
- Generating a small sample CSV for local testing.

Live AWS API access, scheduled reports, Slack/email notifications, and anomaly
detection are intentionally out of scope until the local CSV workflow is useful
and tested.

## Why Marvin Exists

AWS bills can be hard to explain after the fact. Marvin aims to make exported
billing data easier to inspect by answering practical questions:

- Which services cost the most this month?
- What changed compared with last month?
- Did total spend cross a budget?
- Did one service grow unusually fast?
- Can the report be saved and shared as Markdown or JSON?

Marvin is not intended to replace AWS Cost Explorer. It is a small local tool for
turning exported data into repeatable reports that are easy to review in a
terminal, pull request, or incident notes.

Reports use the currency code from the CSV export. Mixed-currency inputs are
rejected so totals are not accidentally combined across currencies.

## Quickstart

Run the sample report:

```sh
marvin analyze fixtures/sample-cost-explorer.csv
```

Analyze multiple CSV exports together:

```sh
marvin analyze account-a.csv account-b.csv
```

Read a gzip-compressed CSV export:

```sh
marvin analyze cost-explorer.csv.gz
```

Generate a sample CSV:

```sh
marvin sample --output sample-costs.csv
marvin analyze sample-costs.csv
```

Validate a CSV before reporting:

```sh
marvin validate fixtures/sample-cost-explorer.csv
```

Run with warning thresholds:

```sh
marvin analyze --total-budget=300 --growth-limit-percent=10 --service-budget "Amazon Elastic Compute Cloud - Compute=200" fixtures/sample-cost-explorer.csv
```

Limit the service table for large exports:

```sh
marvin analyze --top-services=10 fixtures/sample-cost-explorer.csv
```

Sort service rows alphabetically:

```sh
marvin analyze --sort-services=name fixtures/sample-cost-explorer.csv
```

Focus on selected services:

```sh
marvin analyze --only-service "Amazon Elastic Compute Cloud - Compute" fixtures/sample-cost-explorer.csv
```

Analyze one currency from a multi-currency export:

```sh
marvin analyze --currency=USD fixtures/sample-cost-explorer.csv
```

Hide small service rows while keeping total spend and warnings intact:

```sh
marvin analyze --min-service-spend=10 fixtures/sample-cost-explorer.csv
```

Analyze a month range:

```sh
marvin analyze --from=2026-01 --to=2026-02 fixtures/sample-cost-explorer.csv
```

Or load warning thresholds from JSON:

```sh
marvin analyze --config examples/marvin.json fixtures/sample-cost-explorer.csv
```

Config files can also set reusable report options such as `currency`,
`fail_on_warning`, `format`, `include_services`, `ignore_services`, `from_month`,
`min_service_spend`, `output_path`, `to_month`, and `top_services`.

For the full config reference and JSON Schema, see
[`docs/configuration.md`](docs/configuration.md).

Validate a config file:

```sh
marvin config sample --output marvin.json
marvin config validate marvin.json
marvin config sample | marvin config validate -
```

Choose an output format:

```sh
marvin analyze --format terminal fixtures/sample-cost-explorer.csv
marvin analyze --format text fixtures/sample-cost-explorer.csv
marvin analyze --format markdown fixtures/sample-cost-explorer.csv
marvin analyze --format md fixtures/sample-cost-explorer.csv
marvin analyze --format json fixtures/sample-cost-explorer.csv
marvin analyze --format csv fixtures/sample-cost-explorer.csv
```

`terminal` is the default format.

Write a report to a file:

```sh
marvin analyze --format markdown --output report.md fixtures/sample-cost-explorer.csv
```

Read a CSV from standard input:

```sh
cat fixtures/sample-cost-explorer.csv | marvin analyze -
```

For guidance on exporting compatible AWS data, see
[`docs/cost-explorer-export.md`](docs/cost-explorer-export.md).
For common errors and fixes, see
[`docs/troubleshooting.md`](docs/troubleshooting.md).

## CLI Usage

```text
marvin analyze [flags] <cost-explorer.csv|-> [more.csv ...]
marvin config sample [flags]
marvin config validate <marvin.json|->
marvin sample [flags]
marvin validate <cost-explorer.csv|-> [more.csv ...]
marvin version
marvin help
```

Analyze flags:

```text
--config <path>                         Load warning rules from a JSON config file.
--currency <code>                       Include only records for this currency code.
--fail-on-warning                       Exit with code 3 when warnings are present.
--format <terminal|markdown|json|csv>   Output format. Defaults to terminal. Aliases: text, md.
--from <YYYY-MM>                        Include records from this month onward.
--ignore-service <service>              Exclude a service from totals and warnings. Repeatable.
--min-service-spend <amount>            Hide service rows below this spend amount.
--only-service <service>                Include only this service. Repeatable.
--output <path>                         Write the report to a file instead of stdout.
--sort-services <cost|name>             Sort service rows. Defaults to cost.
--to <YYYY-MM>                          Include records through this month.
--total-budget <amount>                 Warn when total spend exceeds amount.
--top-services <count>                  Limit service rows in the report.
--service-budget <service=amount>       Warn when service spend exceeds amount. Repeatable.
--growth-limit-percent <percent>        Warn when month-over-month growth exceeds percent.
```

Sample flags:

```text
--output <path>                         Write the sample CSV to a file instead of stdout.
```

Config sample flags:

```text
--output <path>                         Write the sample config to a file instead of stdout.
```

Exit codes:

```text
0  Success.
1  Runtime error, such as an unreadable CSV or unwritable output file.
2  Usage error, such as an invalid flag.
3  Warnings were present and --fail-on-warning was set.
```

## Example Output

```text
Marvin Cost Report
Total spend: $311.59

Monthly spend
Month    Cost
2026-01  $143.22
2026-02  $168.37

Month-over-month
Month    Previous  Current  Change   Change %
2026-02  $143.22   $168.37  +$25.15  +17.56%

Service spend
Service                                 Cost
Amazon Elastic Compute Cloud - Compute  $268.13
Amazon Simple Storage Service           $40.34
AWS Key Management Service              $3.12
```

## Project Principles

- **Local first:** v0.1 should work with files on disk.
- **No credentials required:** CSV analysis should not need AWS SDK setup.
- **Explain the bill:** reports should favor clarity over dashboards.
- **Useful fixtures:** sample data should let contributors test the tool quickly.
- **Small releases:** each feature should be reviewable and covered by tests.

## Contributing

Marvin is still early, but the project is structured as an open-source Go CLI
from the start. Contributions should stay focused on the local CSV workflow
until the first release is complete.

See `CONTRIBUTING.md` for local setup and pull request expectations.
For package boundaries and data flow, see
[`docs/architecture.md`](docs/architecture.md).

## Releases

Versioned release builds are produced by GitHub Actions when a tag matching
`v*.*.*` is pushed. The release workflow builds Linux, macOS, and Windows
binaries, injects the tag into `marvin version`, and uploads SHA-256 checksums
with each artifact.

## License

Marvin is released under the MIT License.
