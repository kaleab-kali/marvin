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
| `total_budget` | number | Warn when total filtered spend exceeds this amount. |
| `growth_limit_percent` | number | Warn when month-over-month growth exceeds this percentage. |
| `service_budgets` | object | Map of service name to budget amount. |
| `ignore_services` | array of strings | Services to exclude from totals, warnings, and report output. |
| `from_month` | string | Include records from this month onward. Format: `YYYY-MM`. |
| `to_month` | string | Include records through this month. Format: `YYYY-MM`. |
| `top_services` | number | Limit service rows in report output. |

Unknown fields are rejected so typos do not silently change report behavior.

## Example

```json
{
  "total_budget": 300,
  "growth_limit_percent": 10,
  "from_month": "2026-01",
  "to_month": "2026-02",
  "top_services": 10,
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
