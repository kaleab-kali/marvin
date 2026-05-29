package cli

import (
	"bytes"
	"compress/gzip"
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

	if code != ExitOK {
		t.Fatalf("expected exit code %d, got %d", ExitOK, code)
	}
	if !strings.Contains(stdout.String(), "marvin analyze [flags] <cost-explorer.csv|-> [more.csv ...]") {
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

	if code != ExitOK {
		t.Fatalf("expected exit code %d, got %d", ExitOK, code)
	}
	if got, want := stdout.String(), "marvin "+Version+"\n"; got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
}

func TestSampleWritesCostExplorerCSV(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"sample"}, &stdout, &stderr)

	if code != ExitOK {
		t.Fatalf("expected exit code %d, got %d with stderr %q", ExitOK, code, stderr.String())
	}
	output := stdout.String()
	for _, want := range []string{
		"Start Date,End Date,Service,Unblended Cost,Currency",
		"Amazon Elastic Compute Cloud - Compute",
		"AWS Key Management Service",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected sample output to contain %q, got:\n%s", want, output)
		}
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
}

func TestSampleWritesCostExplorerCSVToOutputFile(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "sample.csv")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"sample", "--output", outputPath}, &stdout, &stderr)

	if code != ExitOK {
		t.Fatalf("expected exit code %d, got %d with stderr %q", ExitOK, code, stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected empty stdout when output file is used, got %q", stdout.String())
	}
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("expected sample file to be written: %v", err)
	}
	if !strings.Contains(string(content), "Start Date,End Date,Service,Unblended Cost,Currency") {
		t.Fatalf("expected sample CSV in output file, got:\n%s", string(content))
	}
}

func TestSampleRejectsUnexpectedArgument(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"sample", "extra.csv"}, &stdout, &stderr)

	if code != ExitUsageError {
		t.Fatalf("expected exit code %d, got %d", ExitUsageError, code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected empty stdout, got %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), `unexpected sample argument "extra.csv"`) {
		t.Fatalf("expected unexpected argument error, got %q", stderr.String())
	}
}

func TestConfigValidateAcceptsValidConfig(t *testing.T) {
	configPath := writeTempFile(t, "marvin.json", `{
  "total_budget": 200,
  "growth_limit_percent": 20,
  "service_budgets": {
    "Amazon EC2": 150
  },
  "ignore_services": ["Tax"]
}`)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"config", "validate", configPath}, &stdout, &stderr)

	if code != ExitOK {
		t.Fatalf("expected exit code %d, got %d with stderr %q", ExitOK, code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "is valid") {
		t.Fatalf("expected valid config message, got %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
}

func TestConfigValidateAcceptsConfigFromStdin(t *testing.T) {
	input := strings.NewReader(`{"total_budget": 200}`)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := RunWithIO([]string{"config", "validate", "-"}, input, &stdout, &stderr)

	if code != ExitOK {
		t.Fatalf("expected exit code %d, got %d with stderr %q", ExitOK, code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "config - is valid") {
		t.Fatalf("expected valid stdin config message, got %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
}

func TestConfigSampleWritesJSON(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"config", "sample"}, &stdout, &stderr)

	if code != ExitOK {
		t.Fatalf("expected exit code %d, got %d with stderr %q", ExitOK, code, stderr.String())
	}
	if !strings.Contains(stdout.String(), `"total_budget": 300`) {
		t.Fatalf("expected sample config JSON, got:\n%s", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
}

func TestConfigSampleWritesJSONToOutputFile(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "marvin.json")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"config", "sample", "--output", outputPath}, &stdout, &stderr)

	if code != ExitOK {
		t.Fatalf("expected exit code %d, got %d with stderr %q", ExitOK, code, stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected empty stdout when output file is used, got %q", stdout.String())
	}
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("expected sample config file to be written: %v", err)
	}
	if !strings.Contains(string(content), `"service_budgets"`) {
		t.Fatalf("expected sample config JSON in output file, got:\n%s", string(content))
	}

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"config", "validate", outputPath}, &stdout, &stderr)
	if code != ExitOK {
		t.Fatalf("expected generated sample config to validate, got %d with stderr %q", code, stderr.String())
	}
}

func TestConfigValidateRejectsMissingPath(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"config", "validate"}, &stdout, &stderr)

	if code != ExitUsageError {
		t.Fatalf("expected exit code %d, got %d", ExitUsageError, code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected empty stdout, got %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "config validate requires a config path") {
		t.Fatalf("expected missing path error, got %q", stderr.String())
	}
}

func TestConfigValidateRejectsInvalidConfig(t *testing.T) {
	configPath := writeTempFile(t, "marvin.json", `{"total_budget": -1}`)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"config", "validate", configPath}, &stdout, &stderr)

	if code != ExitRuntimeError {
		t.Fatalf("expected exit code %d, got %d", ExitRuntimeError, code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected empty stdout, got %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "validate config: total_budget must not be negative") {
		t.Fatalf("expected invalid config error, got %q", stderr.String())
	}
}

func TestRunRejectsUnknownCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"nope"}, &stdout, &stderr)

	if code != ExitUsageError {
		t.Fatalf("expected exit code %d, got %d", ExitUsageError, code)
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

	if code != ExitUsageError {
		t.Fatalf("expected exit code %d, got %d", ExitUsageError, code)
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

	if code != ExitOK {
		t.Fatalf("expected exit code %d, got %d with stderr %q", ExitOK, code, stderr.String())
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

func TestAnalyzeCombinesMultipleCostCSVs(t *testing.T) {
	firstCSV := writeTempFile(t, "cost-1.csv", `Start Date,Service,Unblended Cost,Currency
2026-01-01,Amazon EC2,100,USD
`)
	secondCSV := writeTempFile(t, "cost-2.csv", `Start Date,Service,Unblended Cost,Currency
2026-01-01,Amazon S3,25,USD
`)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"analyze", firstCSV, secondCSV}, &stdout, &stderr)

	if code != ExitOK {
		t.Fatalf("expected exit code %d, got %d with stderr %q", ExitOK, code, stderr.String())
	}
	output := stdout.String()
	for _, want := range []string{
		"Total spend: $125.00",
		"Amazon EC2",
		"Amazon S3",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, output)
		}
	}
}

func TestAnalyzeReadsGzipCostCSV(t *testing.T) {
	csvPath := writeTempGzipFile(t, "cost.csv.gz", `Start Date,Service,Unblended Cost,Currency
2026-01-01,Amazon EC2,100,USD
`)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"analyze", csvPath}, &stdout, &stderr)

	if code != ExitOK {
		t.Fatalf("expected exit code %d, got %d with stderr %q", ExitOK, code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Total spend: $100.00") {
		t.Fatalf("expected gzip CSV to be analyzed, got:\n%s", stdout.String())
	}
}

func TestAnalyzeLimitsServiceRows(t *testing.T) {
	csvPath := writeTempCSV(t, `Start Date,Service,Unblended Cost,Currency
2026-01-01,Amazon EC2,100,USD
2026-01-01,Amazon S3,25,USD
2026-01-01,AWS Lambda,5,USD
`)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"analyze", "--top-services=2", csvPath}, &stdout, &stderr)

	if code != ExitOK {
		t.Fatalf("expected exit code %d, got %d with stderr %q", ExitOK, code, stderr.String())
	}
	output := stdout.String()
	for _, want := range []string{
		"Total spend: $130.00",
		"Amazon EC2",
		"Amazon S3",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, output)
		}
	}
	if strings.Contains(output, "AWS Lambda") {
		t.Fatalf("expected AWS Lambda to be hidden by --top-services, got:\n%s", output)
	}
}

func TestAnalyzeFiltersMonthRange(t *testing.T) {
	csvPath := writeTempCSV(t, `Start Date,Service,Unblended Cost,Currency
2026-01-01,Amazon EC2,100,USD
2026-02-01,Amazon EC2,150,USD
2026-03-01,Amazon EC2,200,USD
`)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"analyze", "--from=2026-02", "--to", "2026-02", csvPath}, &stdout, &stderr)

	if code != ExitOK {
		t.Fatalf("expected exit code %d, got %d with stderr %q", ExitOK, code, stderr.String())
	}
	output := stdout.String()
	if !strings.Contains(output, "Total spend: $150.00") {
		t.Fatalf("expected filtered total, got:\n%s", output)
	}
	if strings.Contains(output, "2026-01") || strings.Contains(output, "2026-03") {
		t.Fatalf("expected report to include only 2026-02, got:\n%s", output)
	}
}

func TestAnalyzeRejectsEmptyFilteredResult(t *testing.T) {
	csvPath := writeTempCSV(t, `Start Date,Service,Unblended Cost,Currency
2026-01-01,Amazon EC2,100,USD
`)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"analyze", "--from=2026-02", csvPath}, &stdout, &stderr)

	if code != ExitRuntimeError {
		t.Fatalf("expected exit code %d, got %d", ExitRuntimeError, code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected empty stdout, got %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "analyze produced no records after applying filters") {
		t.Fatalf("expected empty filtered result error, got %q", stderr.String())
	}
}

func TestAnalyzeRejectsMixedCurrencies(t *testing.T) {
	csvPath := writeTempCSV(t, `Start Date,Service,Unblended Cost,Currency
2026-01-01,Amazon EC2,100,USD
2026-01-01,Amazon S3,100,GBP
`)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"analyze", csvPath}, &stdout, &stderr)

	if code != ExitRuntimeError {
		t.Fatalf("expected exit code %d, got %d", ExitRuntimeError, code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected empty stdout, got %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "multiple currencies found after filters: GBP, USD") {
		t.Fatalf("expected mixed currency error, got %q", stderr.String())
	}
}

func TestAnalyzeRejectsInvalidMonthRange(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"analyze", "--from=2026-03", "--to=2026-02", "cost.csv"}, &stdout, &stderr)

	if code != ExitUsageError {
		t.Fatalf("expected exit code %d, got %d", ExitUsageError, code)
	}
	if !strings.Contains(stderr.String(), "--from must be before or equal to --to") {
		t.Fatalf("expected invalid month range error, got %q", stderr.String())
	}
}

func TestAnalyzeRejectsInvalidTopServices(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"analyze", "--top-services=0", "cost.csv"}, &stdout, &stderr)

	if code != ExitUsageError {
		t.Fatalf("expected exit code %d, got %d", ExitUsageError, code)
	}
	if !strings.Contains(stderr.String(), "--top-services must be greater than zero") {
		t.Fatalf("expected invalid top services error, got %q", stderr.String())
	}
}

func TestAnalyzeReadsCostCSVFromStdin(t *testing.T) {
	input := strings.NewReader(`Start Date,Service,Unblended Cost,Currency
2026-01-01,Amazon EC2,100,USD
`)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := RunWithIO([]string{"analyze", "-"}, input, &stdout, &stderr)

	if code != ExitOK {
		t.Fatalf("expected exit code %d, got %d with stderr %q", ExitOK, code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Total spend: $100.00") {
		t.Fatalf("expected stdin CSV to be analyzed, got:\n%s", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
}

func TestAnalyzeRejectsRepeatedStdinPath(t *testing.T) {
	input := strings.NewReader(`Start Date,Service,Unblended Cost,Currency
2026-01-01,Amazon EC2,100,USD
`)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := RunWithIO([]string{"analyze", "-", "-"}, input, &stdout, &stderr)

	if code != ExitUsageError {
		t.Fatalf("expected exit code %d, got %d", ExitUsageError, code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected empty stdout, got %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "analyze accepts standard input only once") {
		t.Fatalf("expected repeated stdin error, got %q", stderr.String())
	}
}

func TestAnalyzeFailsOnWarningWhenRequested(t *testing.T) {
	csvPath := writeTempCSV(t, `Start Date,Service,Unblended Cost,Currency
2026-01-01,Amazon EC2,100,USD
`)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"analyze", "--fail-on-warning", "--total-budget=50", csvPath}, &stdout, &stderr)

	if code != ExitWarning {
		t.Fatalf("expected exit code %d, got %d", ExitWarning, code)
	}
	if !strings.Contains(stdout.String(), "total spend $100.00 exceeds budget $50.00") {
		t.Fatalf("expected warning report, got:\n%s", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
}

func TestAnalyzeDoesNotFailOnWarningByDefault(t *testing.T) {
	csvPath := writeTempCSV(t, `Start Date,Service,Unblended Cost,Currency
2026-01-01,Amazon EC2,100,USD
`)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"analyze", "--total-budget=50", csvPath}, &stdout, &stderr)

	if code != ExitOK {
		t.Fatalf("expected exit code %d, got %d", ExitOK, code)
	}
	if !strings.Contains(stdout.String(), "total spend $100.00 exceeds budget $50.00") {
		t.Fatalf("expected warning report, got:\n%s", stdout.String())
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

	if code != ExitOK {
		t.Fatalf("expected exit code %d, got %d with stderr %q", ExitOK, code, stderr.String())
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

func TestAnalyzeUsesTopServicesFromConfig(t *testing.T) {
	csvPath := writeTempCSV(t, `Start Date,Service,Unblended Cost,Currency
2026-01-01,Amazon EC2,100,USD
2026-01-01,Amazon S3,25,USD
`)
	configPath := writeTempFile(t, "marvin.json", `{"top_services": 1}`)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"analyze", "--config", configPath, csvPath}, &stdout, &stderr)

	if code != ExitOK {
		t.Fatalf("expected exit code %d, got %d with stderr %q", ExitOK, code, stderr.String())
	}
	output := stdout.String()
	if !strings.Contains(output, "Amazon EC2") {
		t.Fatalf("expected top service in report, got:\n%s", output)
	}
	if strings.Contains(output, "Amazon S3") {
		t.Fatalf("expected second service to be hidden by config top_services, got:\n%s", output)
	}
}

func TestAnalyzeUsesMonthRangeFromConfig(t *testing.T) {
	csvPath := writeTempCSV(t, `Start Date,Service,Unblended Cost,Currency
2026-01-01,Amazon EC2,100,USD
2026-02-01,Amazon EC2,150,USD
`)
	configPath := writeTempFile(t, "marvin.json", `{"from_month": "2026-02", "to_month": "2026-02"}`)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"analyze", "--config", configPath, csvPath}, &stdout, &stderr)

	if code != ExitOK {
		t.Fatalf("expected exit code %d, got %d with stderr %q", ExitOK, code, stderr.String())
	}
	output := stdout.String()
	if !strings.Contains(output, "Total spend: $150.00") {
		t.Fatalf("expected filtered total from config month range, got:\n%s", output)
	}
	if strings.Contains(output, "2026-01") {
		t.Fatalf("expected January to be hidden by config month range, got:\n%s", output)
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

	if code != ExitOK {
		t.Fatalf("expected exit code %d, got %d with stderr %q", ExitOK, code, stderr.String())
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

	if code != ExitOK {
		t.Fatalf("expected exit code %d, got %d with stderr %q", ExitOK, code, stderr.String())
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

	if code != ExitOK {
		t.Fatalf("expected exit code %d, got %d with stderr %q", ExitOK, code, stderr.String())
	}
	if strings.Contains(stdout.String(), "exceeds budget") {
		t.Fatalf("expected later flag to override config budget, got:\n%s", stdout.String())
	}
}

func TestAnalyzeRejectsInvalidBudgetFlag(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"analyze", "--total-budget=free", "cost.csv"}, &stdout, &stderr)

	if code != ExitUsageError {
		t.Fatalf("expected exit code %d, got %d", ExitUsageError, code)
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

	if code != ExitOK {
		t.Fatalf("expected exit code %d, got %d with stderr %q", ExitOK, code, stderr.String())
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

func TestAnalyzeAcceptsMarkdownFormatAlias(t *testing.T) {
	csvPath := writeTempCSV(t, `Start Date,Service,Unblended Cost,Currency
2026-01-01,Amazon EC2,100,USD
`)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"analyze", "--format=md", csvPath}, &stdout, &stderr)

	if code != ExitOK {
		t.Fatalf("expected exit code %d, got %d with stderr %q", ExitOK, code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "# Marvin Cost Report") {
		t.Fatalf("expected markdown report, got:\n%s", stdout.String())
	}
}

func TestAnalyzeAcceptsTerminalFormatAlias(t *testing.T) {
	csvPath := writeTempCSV(t, `Start Date,Service,Unblended Cost,Currency
2026-01-01,Amazon EC2,100,USD
`)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"analyze", "--format=text", csvPath}, &stdout, &stderr)

	if code != ExitOK {
		t.Fatalf("expected exit code %d, got %d with stderr %q", ExitOK, code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Marvin Cost Report") {
		t.Fatalf("expected terminal report, got:\n%s", stdout.String())
	}
}

func TestAnalyzeWritesJSONReport(t *testing.T) {
	csvPath := writeTempCSV(t, `Start Date,Service,Unblended Cost,Currency
2026-01-01,Amazon EC2,100,USD
`)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"analyze", "--format", "json", csvPath}, &stdout, &stderr)

	if code != ExitOK {
		t.Fatalf("expected exit code %d, got %d with stderr %q", ExitOK, code, stderr.String())
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

func TestAnalyzeWritesReportToOutputFile(t *testing.T) {
	csvPath := writeTempCSV(t, `Start Date,Service,Unblended Cost,Currency
2026-01-01,Amazon EC2,100,USD
`)
	outputPath := filepath.Join(t.TempDir(), "report.md")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"analyze", "--format=markdown", "--output", outputPath, csvPath}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d with stderr %q", code, stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected empty stdout when output file is used, got %q", stdout.String())
	}
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("expected output file to be written: %v", err)
	}
	if !strings.Contains(string(content), "# Marvin Cost Report") {
		t.Fatalf("expected markdown report in output file, got:\n%s", string(content))
	}
}

func TestAnalyzeRejectsEmptyOutputPath(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"analyze", "--output=", "cost.csv"}, &stdout, &stderr)

	if code != ExitUsageError {
		t.Fatalf("expected exit code %d, got %d", ExitUsageError, code)
	}
	if !strings.Contains(stderr.String(), "--output requires a path") {
		t.Fatalf("expected output path error, got %q", stderr.String())
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

func writeTempGzipFile(t *testing.T, name string, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), name)
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("create temp gzip file: %v", err)
	}
	gz := gzip.NewWriter(file)
	if _, err := gz.Write([]byte(content)); err != nil {
		t.Fatalf("write gzip content: %v", err)
	}
	if err := gz.Close(); err != nil {
		t.Fatalf("close gzip writer: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("close temp gzip file: %v", err)
	}
	return path
}
