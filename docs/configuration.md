# Configuration

Marvin can load reusable analysis settings from a JSON config file:

```sh
marvin analyze --config marvin.json cost-explorer.csv
```

Generate a starter file:

```sh
marvin config sample --output marvin.json
```

Validate a file before using it:

```sh
marvin config validate marvin.json
marvin config sample | marvin config validate -
```

## Fields

All fields are optional.

| Field | Type | Description |
| --- | --- | --- |
| `$schema` | string | Optional editor hint for JSON Schema validation. |
| `currency` | string | Include only records for this three-letter currency code. |
| `fail_on_warning` | boolean | Exit with code 3 when warnings are present. |
| `total_budget` | number | Warn when total filtered spend exceeds this amount. |
| `format` | string | Report output format: `terminal`, `markdown`, `json`, or `csv`. Aliases: `text`, `md`. |
| `growth_limit_percent` | number | Warn when month-over-month growth exceeds this percentage. |
| `service_budgets` | object | Map of service name to budget amount. |
| `include_services` | array of strings | Services to include in totals, warnings, and report output. |
| `ignore_services` | array of strings | Services to exclude from totals, warnings, and report output. |
| `from_month` | string | Include records from this month onward. Format: `YYYY-MM`. |
| `min_service_spend` | number | Hide service rows below this spend amount. Totals and warnings still use all filtered records. |
| `to_month` | string | Include records through this month. Format: `YYYY-MM`. |
| `top_services` | number | Limit service rows in report output. |

Unknown fields are rejected so typos do not silently change report behavior.
The `$schema` field is the only metadata field Marvin accepts.

## JSON Schema

The config schema is published at
[`docs/marvin.schema.json`](marvin.schema.json). Use this URL in `marvin.json`
for editor validation:

```json
{
  "$schema": "https://raw.githubusercontent.com/kaleab-kali/marvin/main/docs/marvin.schema.json"
}
```

Marvin also validates rules that JSON Schema cannot express, such as requiring
`from_month` to be before or equal to `to_month`.

## Example

```json
{
  "$schema": "https://raw.githubusercontent.com/kaleab-kali/marvin/main/docs/marvin.schema.json",
  "currency": "USD",
  "total_budget": 300,
  "format": "terminal",
  "growth_limit_percent": 10,
  "from_month": "2026-01",
  "min_service_spend": 10,
  "to_month": "2026-02",
  "top_services": 10,
  "include_services": [
    "Amazon Elastic Compute Cloud - Compute",
    "Amazon Simple Storage Service"
  ],
  "ignore_services": [
    "Tax",
    "Credits"
  ],
  "service_budgets": {
    "Amazon Elastic Compute Cloud - Compute": 200
  }
}
```

Command-line flags are applied in the order they appear. If a config file is
loaded before a flag, the later flag can override that setting. If a config file
is loaded after a flag, the config value takes precedence.
