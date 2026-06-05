# Troubleshooting

This guide covers common Marvin errors when analyzing local AWS Cost Explorer
CSV exports.

## `analyze requires a Cost Explorer CSV path`

Pass at least one CSV path or `-` for standard input:

```sh
marvin analyze cost-explorer.csv
marvin validate cost-explorer.csv
cat cost-explorer.csv | marvin analyze -
```

## `missing required column(s)`

Marvin needs a date column, a service column, and a cost column. Export monthly
Cost Explorer data grouped by service, then confirm the CSV contains compatible
headers. See [`cost-explorer-export.md`](cost-explorer-export.md) for supported
column names.

## `cost CSV is empty` or `cost CSV contains no data rows`

The CSV has headers but no data rows. Re-export the Cost Explorer report with a
date range that contains spend, or remove filters that exclude every row in the
AWS export before downloading it.

## `invalid cost`

The cost column must contain finite numeric values. Marvin accepts common
currency markers such as `$12.34`, `USD 12.34`, and `12.34 USD`, but rejects
empty, non-numeric, `NaN`, and infinite values.

## `analyze produced no records after applying filters`

Marvin parsed the CSV, but local filters removed every record. Check:

- `--only-service`
- `--ignore-service`
- `--from` and `--to`
- `include_services`, `ignore_services`, `from_month`, and `to_month` in config

Service names must match the CSV values exactly after trimming surrounding
spaces.

## `multiple currencies found after filters`

Marvin reports one currency at a time. If a CSV contains multiple currencies,
use `--currency` to isolate one currency before generating a report, or export
separate Cost Explorer files per currency.

```sh
marvin analyze --currency=USD cost-explorer.csv
```

## `invalid --from value` or `invalid --to value`

Month filters must use `YYYY-MM`:

```sh
marvin analyze --from=2026-01 --to=2026-03 cost-explorer.csv
```

The start month must be before or equal to the end month.

## `validate config: decode config`

The config file is not valid JSON or contains an unknown field. Generate a fresh
starter file and compare it with your config:

```sh
marvin validate cost-explorer.csv
marvin config sample --output marvin.json
marvin config validate marvin.json
```

The JSON Schema is documented in [`configuration.md`](configuration.md).

## Reports Are Too Noisy

Use service row controls without changing total spend or warning evaluation:

```sh
marvin analyze --top-services=10 cost-explorer.csv
marvin analyze --min-service-spend=10 cost-explorer.csv
```

Use record filters when you want totals and warnings to focus on a subset of
the export:

```sh
marvin analyze --only-service "Amazon Elastic Compute Cloud - Compute" cost-explorer.csv
marvin analyze --ignore-service Tax cost-explorer.csv
```
