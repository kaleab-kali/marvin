package cost

import (
	"math"
	"strings"
	"testing"
	"time"
)

func TestParseCostExplorerCSV(t *testing.T) {
	input := strings.NewReader(`Start Date,End Date,Service,Unblended Cost,Currency
2026-01-01,2026-01-31,Amazon EC2,$123.45,USD
2026-02-01,2026-02-28,Amazon S3,"USD 1,234.560000",USD
`)

	records, err := ParseCostExplorerCSV(input)
	if err != nil {
		t.Fatalf("expected parser to succeed, got %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}

	assertRecord(t, records[0], "Amazon EC2", "2026-01-01", "2026-01-31", 123.45, "USD")
	assertRecord(t, records[1], "Amazon S3", "2026-02-01", "2026-02-28", 1234.56, "USD")
}

func TestParseCostExplorerCSVSupportsCommonHeaderAliases(t *testing.T) {
	input := strings.NewReader(`Date,Product,Cost
2026-03,Amazon CloudWatch,42.10
`)

	records, err := ParseCostExplorerCSV(input)
	if err != nil {
		t.Fatalf("expected parser to succeed, got %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}

	assertRecord(t, records[0], "Amazon CloudWatch", "2026-03-01", "", 42.10, "USD")
}

func TestParseCostExplorerCSVSupportsCURColumnNames(t *testing.T) {
	input := strings.NewReader(`lineItem/UsageStartDate,lineItem/UsageEndDate,lineItem/ProductCode,lineItem/UnblendedCost,lineItem/CurrencyCode
2026-03-01T00:00:00Z,2026-03-01T01:00:00Z,AmazonEC2,12.3400000000,USD
`)

	records, err := ParseCostExplorerCSV(input)
	if err != nil {
		t.Fatalf("expected parser to succeed, got %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}

	assertRecord(t, records[0], "AmazonEC2", "2026-03-01", "2026-03-01", 12.34, "USD")
}

func TestParseCostExplorerCSVSupportsSpacedTimestampColumns(t *testing.T) {
	input := strings.NewReader(`Usage Start Time,Usage End Time,Service Code,Net Amortized Cost,Pricing Currency
2026-03-01 12:30:00,2026-03-01 13:30:00,AmazonS3,£42.10,GBP
`)

	records, err := ParseCostExplorerCSV(input)
	if err != nil {
		t.Fatalf("expected parser to succeed, got %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}

	assertRecord(t, records[0], "AmazonS3", "2026-03-01", "2026-03-01", 42.10, "GBP")
}

func TestParseCostExplorerCSVRejectsMissingRequiredColumns(t *testing.T) {
	input := strings.NewReader(`Start Date,Unblended Cost,Currency
2026-01-01,10.00,USD
`)

	_, err := ParseCostExplorerCSV(input)
	if err == nil {
		t.Fatal("expected missing required column error")
	}
	if !strings.Contains(err.Error(), "missing required column(s): service") {
		t.Fatalf("expected service column error, got %v", err)
	}
}

func TestParseCostExplorerCSVIncludesLineNumberOnBadCost(t *testing.T) {
	input := strings.NewReader(`Start Date,Service,Unblended Cost,Currency
2026-01-01,Amazon EC2,nope,USD
`)

	_, err := ParseCostExplorerCSV(input)
	if err == nil {
		t.Fatal("expected invalid cost error")
	}
	if !strings.Contains(err.Error(), "parse CSV line 2") {
		t.Fatalf("expected line number in error, got %v", err)
	}
	if !strings.Contains(err.Error(), "invalid cost") {
		t.Fatalf("expected invalid cost in error, got %v", err)
	}
}

func TestParseCostValue(t *testing.T) {
	tests := map[string]float64{
		"12.34":         12.34,
		"$12.34":        12.34,
		"€12.34":        12.34,
		"£12.34":        12.34,
		"USD 12.34":     12.34,
		"12.34 USD":     12.34,
		"1,234.560000":  1234.56,
		"($1,234.5600)": -1234.56,
	}

	for input, want := range tests {
		t.Run(input, func(t *testing.T) {
			got, err := parseCostValue(input)
			if err != nil {
				t.Fatalf("expected parse to succeed, got %v", err)
			}
			if math.Abs(got-want) > 0.000001 {
				t.Fatalf("expected %f, got %f", want, got)
			}
		})
	}
}

func assertRecord(t *testing.T, record Record, service, start, end string, cost float64, currency string) {
	t.Helper()

	if record.Service != service {
		t.Fatalf("expected service %q, got %q", service, record.Service)
	}
	assertDate(t, "start date", record.StartDate, start)
	if end == "" {
		if !record.EndDate.IsZero() {
			t.Fatalf("expected zero end date, got %s", record.EndDate.Format(time.DateOnly))
		}
	} else {
		assertDate(t, "end date", record.EndDate, end)
	}
	if math.Abs(record.Cost-cost) > 0.000001 {
		t.Fatalf("expected cost %f, got %f", cost, record.Cost)
	}
	if record.Currency != currency {
		t.Fatalf("expected currency %q, got %q", currency, record.Currency)
	}
}

func assertDate(t *testing.T, label string, got time.Time, want string) {
	t.Helper()

	parsed, err := time.Parse(time.DateOnly, want)
	if err != nil {
		t.Fatalf("invalid expected %s %q: %v", label, want, err)
	}
	if got.Format(time.DateOnly) != parsed.Format(time.DateOnly) {
		t.Fatalf("expected %s %s, got %s", label, want, got.Format(time.DateOnly))
	}
}
