package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunShowsHelpWithoutArgs(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run(nil, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if !strings.Contains(stdout.String(), "marvin analyze [flags] <cost-explorer.csv>") {
		t.Fatalf("expected usage in stdout, got %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
}

func TestRunShowsVersion(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"version"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if got, want := stdout.String(), "marvin "+Version+"\n"; got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
}

func TestRunRejectsUnknownCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"nope"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("expected exit code 2, got %d", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected empty stdout, got %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), `unknown command "nope"`) {
		t.Fatalf("expected unknown command error, got %q", stderr.String())
	}
}

func TestAnalyzeRequiresPath(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"analyze"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("expected exit code 2, got %d", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected empty stdout, got %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "requires a Cost Explorer CSV path") {
		t.Fatalf("expected path error, got %q", stderr.String())
	}
}

func TestAnalyzeWritesTerminalReport(t *testing.T) {
	csvPath := writeTempCSV(t, `Start Date,End Date,Service,Unblended Cost,Currency
2026-01-01,2026-01-31,Amazon EC2,100,USD
2026-02-01,2026-02-28,Amazon EC2,150,USD
`)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"analyze", "--total-budget=200", "--growth-limit-percent=20", "--service-budget", "Amazon EC2=200", csvPath}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d with stderr %q", code, stderr.String())
	}
	output := stdout.String()
	for _, want := range []string{
		"Marvin Cost Report",
		"Total spend: $250.00",
		"Amazon EC2",
		"total spend $250.00 exceeds budget $200.00",
		"Amazon EC2 spend $250.00 exceeds budget $200.00",
		"2026-02 spend grew +50.00%",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, output)
		}
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
}

func TestAnalyzeRejectsInvalidBudgetFlag(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"analyze", "--total-budget=free", "cost.csv"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("expected exit code 2, got %d", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected empty stdout, got %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), `invalid --total-budget value "free"`) {
		t.Fatalf("expected invalid budget error, got %q", stderr.String())
	}
}

func writeTempCSV(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "cost.csv")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write temp CSV: %v", err)
	}
	return path
}
