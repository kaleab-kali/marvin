package cli

import (
	"bytes"
	"encoding/json"
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

func TestAnalyzeUsesConfigFile(t *testing.T) {
	csvPath := writeTempCSV(t, `Start Date,Service,Unblended Cost,Currency
2026-01-01,Amazon EC2,100,USD
2026-02-01,Amazon EC2,150,USD
`)
	configPath := writeTempFile(t, "marvin.json", `{
  "total_budget": 200,
  "growth_limit_percent": 20,
  "service_budgets": {
    "Amazon EC2": 200
  }
}`)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"analyze", "--config", configPath, csvPath}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d with stderr %q", code, stderr.String())
	}
	output := stdout.String()
	for _, want := range []string{
		"total spend $250.00 exceeds budget $200.00",
		"Amazon EC2 spend $250.00 exceeds budget $200.00",
		"2026-02 spend grew +50.00%",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, output)
		}
	}
}

func TestAnalyzeIgnoresServicesFromConfig(t *testing.T) {
	csvPath := writeTempCSV(t, `Start Date,Service,Unblended Cost,Currency
2026-01-01,Amazon EC2,100,USD
2026-01-01,Tax,20,USD
`)
	configPath := writeTempFile(t, "marvin.json", `{"ignore_services": ["Tax"]}`)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"analyze", "--config", configPath, csvPath}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d with stderr %q", code, stderr.String())
	}
	output := stdout.String()
	if !strings.Contains(output, "Total spend: $100.00") {
		t.Fatalf("expected ignored tax to be excluded, got:\n%s", output)
	}
	if strings.Contains(output, "Tax") {
		t.Fatalf("expected Tax to be absent from report, got:\n%s", output)
	}
}

func TestAnalyzeIgnoresServicesFromFlag(t *testing.T) {
	csvPath := writeTempCSV(t, `Start Date,Service,Unblended Cost,Currency
2026-01-01,Amazon EC2,100,USD
2026-01-01,Tax,20,USD
`)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"analyze", "--ignore-service=Tax", csvPath}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d with stderr %q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Total spend: $100.00") {
		t.Fatalf("expected ignored tax to be excluded, got:\n%s", stdout.String())
	}
}

func TestAnalyzeLetsFlagsOverrideEarlierConfig(t *testing.T) {
	csvPath := writeTempCSV(t, `Start Date,Service,Unblended Cost,Currency
2026-01-01,Amazon EC2,100,USD
`)
	configPath := writeTempFile(t, "marvin.json", `{"total_budget": 50}`)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"analyze", "--config", configPath, "--total-budget=200", csvPath}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d with stderr %q", code, stderr.String())
	}
	if strings.Contains(stdout.String(), "exceeds budget") {
		t.Fatalf("expected later flag to override config budget, got:\n%s", stdout.String())
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

func TestAnalyzeWritesMarkdownReport(t *testing.T) {
	csvPath := writeTempCSV(t, `Start Date,Service,Unblended Cost,Currency
2026-01-01,Amazon EC2,100,USD
`)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"analyze", "--format=markdown", csvPath}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d with stderr %q", code, stderr.String())
	}
	output := stdout.String()
	for _, want := range []string{
		"# Marvin Cost Report",
		"Total spend: **$100.00**",
		"| Amazon EC2 | $100.00 |",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, output)
		}
	}
}

func TestAnalyzeWritesJSONReport(t *testing.T) {
	csvPath := writeTempCSV(t, `Start Date,Service,Unblended Cost,Currency
2026-01-01,Amazon EC2,100,USD
`)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"analyze", "--format", "json", csvPath}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d with stderr %q", code, stderr.String())
	}

	var payload struct {
		TotalSpend   float64 `json:"total_spend"`
		ServiceSpend []struct {
			Service string  `json:"service"`
			Cost    float64 `json:"cost"`
		} `json:"service_spend"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("expected valid JSON, got %v with output %q", err, stdout.String())
	}
	if payload.TotalSpend != 100 {
		t.Fatalf("expected total spend 100, got %f", payload.TotalSpend)
	}
	if len(payload.ServiceSpend) != 1 || payload.ServiceSpend[0].Service != "Amazon EC2" {
		t.Fatalf("expected Amazon EC2 service spend, got %+v", payload.ServiceSpend)
	}
}

func TestAnalyzeRejectsUnsupportedFormat(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"analyze", "--format=xml", "cost.csv"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("expected exit code 2, got %d", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected empty stdout, got %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), `unsupported --format "xml"`) {
		t.Fatalf("expected unsupported format error, got %q", stderr.String())
	}
}

func writeTempCSV(t *testing.T, content string) string {
	t.Helper()

	return writeTempFile(t, "cost.csv", content)
}

func writeTempFile(t *testing.T, name string, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	return path
}
