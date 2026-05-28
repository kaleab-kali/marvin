package report

import (
	"os"
	"testing"

	"github.com/kaleab-kali/marvin/internal/cost"
)

func sampleReportRecords(t *testing.T) []cost.Record {
	t.Helper()

	return []cost.Record{
		{Service: "Amazon EC2", StartDate: mustDate(t, "2026-01-01"), Cost: 100},
		{Service: "Amazon EC2", StartDate: mustDate(t, "2026-02-01"), Cost: 150},
	}
}

func readGoldenFile(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden file: %v", err)
	}
	return string(content)
}
