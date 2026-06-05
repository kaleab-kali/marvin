package cost

import (
	"fmt"
	"strings"
	"testing"
)

func BenchmarkParseCostExplorerCSV(b *testing.B) {
	input := benchmarkCostExplorerCSV(12, 50)
	b.ReportAllocs()
	b.SetBytes(int64(len(input)))

	for i := 0; i < b.N; i++ {
		records, err := ParseCostExplorerCSV(strings.NewReader(input))
		if err != nil {
			b.Fatalf("parse benchmark CSV: %v", err)
		}
		if len(records) != 600 {
			b.Fatalf("expected 600 records, got %d", len(records))
		}
	}
}

func benchmarkCostExplorerCSV(months, services int) string {
	var builder strings.Builder
	builder.WriteString("Start Date,End Date,Service,Unblended Cost,Currency\n")
	for month := 1; month <= months; month++ {
		for service := 1; service <= services; service++ {
			fmt.Fprintf(
				&builder,
				"2026-%02d-01,2026-%02d-28,Service %03d,%.2f,USD\n",
				month,
				month,
				service,
				float64(month*service)/3,
			)
		}
	}
	return builder.String()
}
