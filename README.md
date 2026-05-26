# Marvin

Marvin is a Go CLI for understanding AWS cost exports without connecting to an
AWS account.

It is designed to read exported AWS Cost Explorer CSV files, group spend by
service and month, and produce plain reports that explain where the bill moved.
The goal is a pessimistic but accurate local cost monitor: no credentials, no
cloud access, and no hidden state required for the first release.

## Status

Marvin is in early development. The first milestone is a CSV-based MVP that can
analyze sample Cost Explorer exports locally.

The initial release will focus on:

- Importing AWS Cost Explorer CSV files.
- Grouping spend by service and month.
- Showing month-over-month changes.
- Emitting budget and growth warnings.
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

## Planned Workflow

The target v0.1 workflow is:

```sh
marvin analyze fixtures/sample-cost-explorer.csv
```

The command will read a Cost Explorer CSV export and print a report with totals,
service-level spend, month-over-month movement, and budget warnings.

## Project Principles

- **Local first:** v0.1 should work with files on disk.
- **No credentials required:** CSV analysis should not need AWS SDK setup.
- **Explain the bill:** reports should favor clarity over dashboards.
- **Useful fixtures:** sample data should let contributors test the tool quickly.
- **Small releases:** each feature should be reviewable and covered by tests.

## Contributing

Marvin is not ready for broad contribution yet, but the project is being
structured as an open-source Go CLI from the start. Contributions should stay
focused on the CSV MVP until the first release is complete.

See `CONTRIBUTING.md` once the development scaffold lands.

## License

Marvin is planned to be released under the MIT License.
