# Exporting Cost Explorer CSV Files

Marvin v0.1 reads local CSV files. It does not connect to AWS, read from S3, or
use AWS credentials.

Use this guide to export a CSV from AWS Cost Explorer that is easy for Marvin to
analyze.

## Recommended Cost Explorer Settings

In AWS Cost Explorer, configure the report before downloading the CSV:

- Granularity: monthly.
- Group by: service.
- Metric: unblended cost, net unblended cost, amortized cost, or blended cost.
- Date range: include at least two complete months if you want
  month-over-month comparisons.
- Filters: choose whatever account, region, tag, or cost category scope you want
  Marvin to analyze.

Then choose **Download CSV** in Cost Explorer. AWS documents the CSV download
workflow in the AWS Cost Management User Guide:
https://docs.aws.amazon.com/cost-management/latest/userguide/ce-download-csv.html

## Supported Columns

Marvin looks for a service column, a start date column, a cost column, and an
optional currency or end date column. Header matching is case-insensitive and
ignores spaces, underscores, hyphens, and slash characters.

Common supported service columns include:

- `Service`
- `Service Name`
- `Service Code`
- `Product`
- `Product Name`
- `lineItem/ProductCode`

Common supported date columns include:

- `Start Date`
- `Usage Start Date`
- `Usage Start Time`
- `Billing Period Start Date`
- `lineItem/UsageStartDate`

Common supported cost columns include:

- `Unblended Cost`
- `Net Unblended Cost`
- `Amortized Cost`
- `Net Amortized Cost`
- `Blended Cost`
- `lineItem/UnblendedCost`

Common supported currency columns include:

- `Currency`
- `Currency Code`
- `lineItem/CurrencyCode`
- `Pricing Currency`

## Validate An Export

Run:

```sh
marvin analyze path/to/cost-explorer.csv
```

To combine multiple exports into one report:

```sh
marvin analyze path/to/account-a.csv path/to/account-b.csv
```

Marvin can also read a CSV from standard input:

```sh
cat path/to/cost-explorer.csv | marvin analyze -
```

To create a shareable Markdown report:

```sh
marvin analyze --format markdown --output cost-report.md path/to/cost-explorer.csv
```

To keep the service section short for large exports:

```sh
marvin analyze --top-services=10 path/to/cost-explorer.csv
```

To fail automation when configured warning rules trigger:

```sh
marvin analyze --config examples/marvin.json --fail-on-warning path/to/cost-explorer.csv
```

To validate a config file before using it:

```sh
marvin config sample --output marvin.json
marvin config validate marvin.json
```

## Handling Private Billing Data

Do not commit real billing exports to the repository. Cost exports can include
account names, linked accounts, product usage, tags, and internal allocation
details depending on the filters you use in AWS.

For issues and pull requests, reproduce parser problems with a minimal sanitized
CSV instead of a real account export.
