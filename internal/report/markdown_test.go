package report

import (
	"bytes"
	"os"
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
	records := []cost.Record{
		{Service: "Amazon EC2", StartDate: mustDate(t, "2026-01-01"), Cost: 100},
		{Service: "Amazon EC2", StartDate: mustDate(t, "2026-02-01"), Cost: 150},
	}

	var output bytes.Buffer
	if err := WriteMarkdown(&output, records, cost.WarningRules{}); err != nil {
		t.Fatalf("expected markdown report to write, got %v", err)
	}

	want, err := os.ReadFile("testdata/sample.golden.md")
	if err != nil {
		t.Fatalf("read golden file: %v", err)
	}
	if got := output.String(); got != string(want) {
		t.Fatalf("markdown output mismatch\nwant:\n%s\ngot:\n%s", string(want), got)
	}
}
