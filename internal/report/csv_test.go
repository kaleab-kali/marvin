package report

import (
	"bytes"
	"encoding/csv"
	"strings"
	"testing"

	"github.com/kaleab-kali/marvin/internal/cost"
)

func TestWriteCSV(t *testing.T) {
	records := []cost.Record{
		{Service: "Amazon EC2", StartDate: mustDate(t, "2026-01-01"), Cost: 100},
		{Service: "Amazon EC2", StartDate: mustDate(t, "2026-02-01"), Cost: 150},
	}
	rules := cost.WarningRules{TotalLimit: 200}

	var output bytes.Buffer
	if err := WriteCSV(&output, records, rules); err != nil {
		t.Fatalf("expected CSV report to write, got %v", err)
	}

	rows, err := csv.NewReader(strings.NewReader(output.String())).ReadAll()
	if err != nil {
		t.Fatalf("expected valid CSV, got %v with output:\n%s", err, output.String())
	}
	if got, want := rows[0][0], "section"; got != want {
		t.Fatalf("expected header %q, got %q", want, got)
	}
	if !hasCSVRow(rows, "total", "", "", "", "USD", "250.00", "", "", "", "", "", "", "") {
		t.Fatalf("expected total row, got %+v", rows)
	}
	if !hasCSVRow(rows, "warning", "", "", "", "USD", "", "", "", "", "total_budget", "200.00", "250.00", "") {
		t.Fatalf("expected total budget warning row, got %+v", rows)
	}
}

func TestWriteCSVMatchesGoldenFile(t *testing.T) {
	var output bytes.Buffer
	if err := WriteCSV(&output, sampleReportRecords(t), cost.WarningRules{}); err != nil {
		t.Fatalf("expected CSV report to write, got %v", err)
	}

	want := readGoldenFile(t, "testdata/sample.golden.csv")
	if got := output.String(); got != want {
		t.Fatalf("csv output mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func hasCSVRow(rows [][]string, want ...string) bool {
	for _, row := range rows {
		if len(row) != len(want) {
			continue
		}
		matched := true
		for i := range row {
			if row[i] != want[i] {
				matched = false
				break
			}
		}
		if matched {
			return true
		}
	}
	return false
}
