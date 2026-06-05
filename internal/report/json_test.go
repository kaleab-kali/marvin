package report

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/kaleab-kali/marvin/internal/cost"
)

func TestWriteJSON(t *testing.T) {
	records := []cost.Record{
		{Service: "Amazon EC2", StartDate: mustDate(t, "2026-01-01"), Cost: 100},
	}

	var output bytes.Buffer
	if err := WriteJSON(&output, records, cost.WarningRules{}); err != nil {
		t.Fatalf("expected JSON report to write, got %v", err)
	}

	var summary Summary
	if err := json.Unmarshal(output.Bytes(), &summary); err != nil {
		t.Fatalf("expected valid JSON, got %v with output %q", err, output.String())
	}
	if summary.TotalSpend != 100 {
		t.Fatalf("expected total spend 100, got %f", summary.TotalSpend)
	}
	if len(summary.ServiceSpend) != 1 || summary.ServiceSpend[0].Service != "Amazon EC2" {
		t.Fatalf("expected Amazon EC2 service spend, got %+v", summary.ServiceSpend)
	}
	if summary.ServiceSpend[0].SharePercent != 100 {
		t.Fatalf("expected service share percent 100, got %f", summary.ServiceSpend[0].SharePercent)
	}
}

func TestWriteJSONSummary(t *testing.T) {
	summary := Summary{TotalSpend: 100}

	var output bytes.Buffer
	if err := WriteJSONSummary(&output, summary); err != nil {
		t.Fatalf("expected JSON summary to write, got %v", err)
	}

	var decoded Summary
	if err := json.Unmarshal(output.Bytes(), &decoded); err != nil {
		t.Fatalf("expected valid JSON, got %v with output %q", err, output.String())
	}
	if decoded.TotalSpend != 100 {
		t.Fatalf("expected total spend 100, got %f", decoded.TotalSpend)
	}
}

func TestWriteJSONMatchesGoldenFile(t *testing.T) {
	var output bytes.Buffer
	if err := WriteJSON(&output, sampleReportRecords(t), cost.WarningRules{}); err != nil {
		t.Fatalf("expected JSON report to write, got %v", err)
	}

	want := readGoldenFile(t, "testdata/sample.golden.json")
	if got := output.String(); got != want {
		t.Fatalf("json output mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}
