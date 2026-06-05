package report

import (
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/kaleab-kali/marvin/internal/cost"
)

func BenchmarkBuildSummary(b *testing.B) {
	records := benchmarkRecords()
	rules := cost.WarningRules{
		TotalLimit:         100000,
		GrowthLimitPercent: 25,
		ServiceLimits: map[string]float64{
			"Service 001": 1000,
			"Service 010": 1000,
		},
	}
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		summary := BuildSummary(records, rules)
		if summary.TotalSpend == 0 {
			b.Fatal("expected non-zero total spend")
		}
	}
}

func BenchmarkWriteTerminalSummary(b *testing.B) {
	summary := BuildSummary(benchmarkRecords(), cost.WarningRules{})
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		if err := WriteTerminalSummary(io.Discard, summary); err != nil {
			b.Fatalf("write terminal summary: %v", err)
		}
	}
}

func BenchmarkWriteJSONSummary(b *testing.B) {
	summary := BuildSummary(benchmarkRecords(), cost.WarningRules{})
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		if err := WriteJSONSummary(io.Discard, summary); err != nil {
			b.Fatalf("write JSON summary: %v", err)
		}
	}
}

func BenchmarkWriteCSVSummary(b *testing.B) {
	summary := BuildSummary(benchmarkRecords(), cost.WarningRules{})
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		if err := WriteCSVSummary(io.Discard, summary); err != nil {
			b.Fatalf("write CSV summary: %v", err)
		}
	}
}

func benchmarkRecords() []cost.Record {
	records := make([]cost.Record, 0, 600)
	for month := 1; month <= 12; month++ {
		for service := 1; service <= 50; service++ {
			records = append(records, cost.Record{
				Service:   "Service " + threeDigit(service),
				StartDate: time.Date(2026, time.Month(month), 1, 0, 0, 0, 0, time.UTC),
				EndDate:   time.Date(2026, time.Month(month), 28, 0, 0, 0, 0, time.UTC),
				Cost:      float64(month*service) / 3,
				Currency:  "USD",
			})
		}
	}
	return records
}

func threeDigit(value int) string {
	return fmt.Sprintf("%03d", value)
}
