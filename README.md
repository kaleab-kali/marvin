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

## Quickstart

Run the sample report:

```sh
marvin analyze fixtures/sample-cost-explorer.csv
```

Run with warning thresholds:

```sh
marvin analyze --total-budget=300 --growth-limit-percent=10 --service-budget "Amazon Elastic Compute Cloud - Compute=200" fixtures/sample-cost-explorer.csv
```

Or load warning thresholds from JSON:

```sh
marvin analyze --config examples/marvin.json fixtures/sample-cost-explorer.csv
```

Choose an output format:

```sh
marvin analyze --format terminal fixtures/sample-cost-explorer.csv
marvin analyze --format markdown fixtures/sample-cost-explorer.csv
marvin analyze --format json fixtures/sample-cost-explorer.csv
```

`terminal` is the default format.

Write a report to a file:

```sh
marvin analyze --format markdown --output report.md fixtures/sample-cost-explorer.csv
```

## CLI Usage

```text
marvin analyze [flags] <cost-explorer.csv>
marvin version
marvin help
```

Analyze flags:

```text
--config <path>                         Load warning rules from a JSON config file.
--format <terminal|markdown|json>       Output format. Defaults to terminal.
--ignore-service <service>              Exclude a service from totals and warnings. Repeatable.
--output <path>                         Write the report to a file instead of stdout.
--total-budget <amount>                 Warn when total spend exceeds amount.
--service-budget <service=amount>       Warn when service spend exceeds amount. Repeatable.
--growth-limit-percent <percent>        Warn when month-over-month growth exceeds percent.
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

## License

Marvin is released under the MIT License.
