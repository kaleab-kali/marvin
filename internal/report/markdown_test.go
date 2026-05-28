package report

import (
	"bytes"
	"strings"
	"testing"

	"github.com/kaleab-kali/marvin/internal/cost"
)

func TestWriteMarkdown(t *testing.T) {
	records := []cost.Record{
		{Service: "Amazon EC2", StartDate: mustDate(t, "2026-01-01"), Cost: 100},
	}

	var output bytes.Buffer
	if err := WriteMarkdown(&output, records, cost.WarningRules{}); err != nil {
		t.Fatalf("expected markdown report to write, got %v", err)
	}

	text := output.String()
	for _, want := range []string{
		"# Marvin Cost Report",
		"Total spend: **$100.00**",
		"| Amazon EC2 | $100.00 |",
		"None.",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected markdown to contain %q, got:\n%s", want, text)
		}
	}
}

func TestWriteMarkdownMatchesGoldenFile(t *testing.T) {
	var output bytes.Buffer
	if err := WriteMarkdown(&output, sampleReportRecords(t), cost.WarningRules{}); err != nil {
		t.Fatalf("expected markdown report to write, got %v", err)
	}

	want := readGoldenFile(t, "testdata/sample.golden.md")
	if got := output.String(); got != want {
		t.Fatalf("markdown output mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}
