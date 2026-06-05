package report

import (
	"testing"

	"github.com/kaleab-kali/marvin/internal/cost"
)

func TestBuildSummaryCalculatesServiceSharePercent(t *testing.T) {
	summary := BuildSummary([]cost.Record{
		{Service: "Amazon EC2", StartDate: mustDate(t, "2026-01-01"), Cost: 75},
		{Service: "Amazon S3", StartDate: mustDate(t, "2026-01-01"), Cost: 25},
	}, cost.WarningRules{})

	if len(summary.ServiceSpend) != 2 {
		t.Fatalf("expected 2 service rows, got %+v", summary.ServiceSpend)
	}
	if summary.ServiceSpend[0].Service != "Amazon EC2" || summary.ServiceSpend[0].SharePercent != 75 {
		t.Fatalf("expected EC2 share percent 75, got %+v", summary.ServiceSpend[0])
	}
	if summary.ServiceSpend[1].Service != "Amazon S3" || summary.ServiceSpend[1].SharePercent != 25 {
		t.Fatalf("expected S3 share percent 25, got %+v", summary.ServiceSpend[1])
	}
}
