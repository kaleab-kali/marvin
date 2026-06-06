# Report Formats

Marvin can write terminal, Markdown, JSON, and CSV reports from the same
analysis summary.

Use terminal or Markdown for human review. Use JSON or CSV when another tool
needs to consume the report.

The `inspect` command also supports terminal and JSON output for checking parsed
CSV contents before running analysis.

## Terminal

Terminal output is the default:

```sh
marvin analyze cost-explorer.csv
marvin analyze --format terminal cost-explorer.csv
marvin analyze --format text cost-explorer.csv
```

The terminal report includes total spend, monthly spend, month-over-month
changes, service spend with share of total spend, and warnings.

## Markdown

Markdown is useful for pull requests, incident notes, and shared reports:

```sh
marvin analyze --format markdown --output cost-report.md cost-explorer.csv
marvin analyze --format md cost-explorer.csv
```

The Markdown report contains the same sections as the terminal report, formatted
as headings, tables, and warning bullets.

## JSON

JSON is best for automation:

```sh
marvin analyze --format json cost-explorer.csv
```

Top-level fields:

- `total_spend`: total spend after record filters.
- `currency`: currency code used in the report.
- `monthly_spend`: sorted monthly totals.
- `month_over_month`: sorted month-over-month comparisons.
- `service_spend`: services sorted by the active service sort option.
- `warnings`: warning objects when thresholds are exceeded.

Service rows include:

- `service`: service name from the CSV.
- `cost`: total service spend.
- `share_percent`: service spend divided by total report spend.

Warning rows use these `type` values:

- `total_budget`: total spend exceeded `--total-budget` or `total_budget`.
- `service_budget`: service spend exceeded `--service-budget` or
  `service_budgets`.
- `growth`: month-over-month growth exceeded `--growth-limit-percent` or
  `growth_limit_percent`.

## CSV

CSV is useful for spreadsheets and simple pipelines:

```sh
marvin analyze --format csv --output cost-report.csv cost-explorer.csv
```

CSV reports use one header row and then mixed section rows. The `section`
column identifies how to interpret the row.

Columns:

- `section`: `total`, `monthly_spend`, `month_over_month`, `service_spend`, or
  `warning`.
- `month`: month for monthly, comparison, and growth warning rows.
- `previous_month`: previous month for month-over-month rows.
- `service`: service name for service spend and service warning rows.
- `currency`: report currency code.
- `cost`: total, monthly, comparison, or service cost.
- `share_percent`: service share of total spend for service rows.
- `previous_cost`: previous month cost for month-over-month rows.
- `change`: cost change for month-over-month and growth warning rows.
- `change_percent`: percentage change for month-over-month and growth warning
  rows.
- `warning_type`: warning type for warning rows.
- `limit`: threshold that triggered a warning.
- `actual`: observed value that triggered a warning.
- `previous`: previous month value for growth warning rows.

Empty cells mean the column is not applicable to that row type.

## Inspect JSON

Use JSON inspection when scripts need to check which months, currencies, and
services are present before choosing analysis filters:

```sh
marvin inspect --format json cost-explorer.csv
```

Fields:

- `input_count`: number of input streams or files read.
- `record_count`: number of parsed cost records.
- `first_month`: first parsed billing month.
- `last_month`: last parsed billing month.
- `currency_spend`: sorted currency totals.
- `services`: sorted service names parsed from the export.
