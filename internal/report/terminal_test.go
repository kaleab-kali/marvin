package report

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/kaleab-kali/marvin/internal/cost"
)

func TestWriteTerminal(t *testing.T) {
	records := []cost.Record{
		{Service: "Amazon EC2", StartDate: mustDate(t, "2026-01-01"), Cost: 100},
		{Service: "Amazon S3", StartDate: mustDate(t, "2026-01-01"), Cost: 25},
		{Service: "Amazon EC2", StartDate: mustDate(t, "2026-02-01"), Cost: 150},
		{Service: "Amazon S3", StartDate: mustDate(t, "2026-02-01"), Cost: 30},
	}
	rules := cost.WarningRules{
		TotalLimit:         250,
		GrowthLimitPercent: 20,
		ServiceLimits: map[string]float64{
			"Amazon EC2": 200,
		},
	}

	var output bytes.Buffer
	if err := WriteTerminal(&output, records, rules); err != nil {
		t.Fatalf("expected terminal report to write, got %v", err)
	}

	text := output.String()
	for _, want := range []string{
		"Marvin Cost Report",
		"Total spend: $305.00",
		"2026-01",
		"2026-02",
		"Amazon EC2",
		"$250.00",
		"total spend $305.00 exceeds budget $250.00",
		"Amazon EC2 spend $250.00 exceeds budget $200.00",
		"2026-02 spend grew +44.00%",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected report to contain %q, got:\n%s", want, text)
		}
	}
}

func TestWriteTerminalShowsNoWarnings(t *testing.T) {
	records := []cost.Record{
		{Service: "Amazon EC2", StartDate: mustDate(t, "2026-01-01"), Cost: 100},
	}

	var output bytes.Buffer
	if err := WriteTerminal(&output, records, cost.WarningRules{}); err != nil {
		t.Fatalf("expected terminal report to write, got %v", err)
	}
	if !strings.Contains(output.String(), "Warnings\nNone") {
		t.Fatalf("expected no warning text, got:\n%s", output.String())
	}
}

func TestWriteTerminalSummary(t *testing.T) {
	summary := Summary{
		TotalSpend:   100,
		MonthlySpend: []MonthSpend{{Month: "2026-01", Cost: 100}},
		ServiceSpend: []ServiceSpend{{Service: "Amazon EC2", Cost: 100}},
	}

	var output bytes.Buffer
	if err := WriteTerminalSummary(&output, summary); err != nil {
		t.Fatalf("expected terminal summary to write, got %v", err)
	}
	if !strings.Contains(output.String(), "Total spend: $100.00") {
		t.Fatalf("expected terminal summary, got:\n%s", output.String())
	}
}

func mustDate(t *testing.T, value string) time.Time {
	t.Helper()

	parsed, err := time.Parse(time.DateOnly, value)
	if err != nil {
		t.Fatalf("invalid test date %q: %v", value, err)
	}
	return parsed
}
