package config

import (
	"strings"
	"testing"
	"time"
)

func TestLoadWarningRules(t *testing.T) {
	input := strings.NewReader(`{
  "total_budget": 300,
  "growth_limit_percent": 10,
  "service_budgets": {
    "Amazon EC2": 200
  }
}`)

	rules, err := LoadWarningRules(input)
	if err != nil {
		t.Fatalf("expected config to load, got %v", err)
	}
	if rules.TotalLimit != 300 {
		t.Fatalf("expected total budget 300, got %f", rules.TotalLimit)
	}
	if rules.GrowthLimitPercent != 10 {
		t.Fatalf("expected growth limit 10, got %f", rules.GrowthLimitPercent)
	}
	if rules.ServiceLimits["Amazon EC2"] != 200 {
		t.Fatalf("expected EC2 service budget 200, got %f", rules.ServiceLimits["Amazon EC2"])
	}
}

func TestLoadIncludesServiceFilters(t *testing.T) {
	input := strings.NewReader(`{
  "ignore_services": ["Tax", "Credits"],
  "include_services": ["Amazon EC2", "Amazon S3"],
  "currency": "usd",
  "fail_on_warning": true,
  "format": "md",
  "from_month": "2026-01",
  "min_service_spend": 10,
  "output_path": "report.md",
  "sort_services": "name",
  "to_month": "2026-02",
  "top_services": 10
}`)

	settings, err := Load(input)
	if err != nil {
		t.Fatalf("expected config to load, got %v", err)
	}
	if len(settings.IgnoreServices) != 2 {
		t.Fatalf("expected 2 ignored services, got %d", len(settings.IgnoreServices))
	}
	if settings.IgnoreServices[0] != "Tax" || settings.IgnoreServices[1] != "Credits" {
		t.Fatalf("unexpected ignored services: %+v", settings.IgnoreServices)
	}
	if len(settings.IncludeServices) != 2 {
		t.Fatalf("expected 2 included services, got %d", len(settings.IncludeServices))
	}
	if settings.IncludeServices[0] != "Amazon EC2" || settings.IncludeServices[1] != "Amazon S3" {
		t.Fatalf("unexpected included services: %+v", settings.IncludeServices)
	}
	if settings.Format != "markdown" {
		t.Fatalf("expected markdown format, got %q", settings.Format)
	}
	if settings.Currency != "USD" {
		t.Fatalf("expected USD currency, got %q", settings.Currency)
	}
	if settings.FailOnWarning == nil || !*settings.FailOnWarning {
		t.Fatalf("expected fail on warning to be true, got %+v", settings.FailOnWarning)
	}
	assertMonth(t, "from month", settings.FromMonth, "2026-01")
	assertMonth(t, "to month", settings.ToMonth, "2026-02")
	if settings.MinServiceSpend != 10 {
		t.Fatalf("expected min service spend 10, got %f", settings.MinServiceSpend)
	}
	if settings.OutputPath == nil || *settings.OutputPath != "report.md" {
		t.Fatalf("expected output path report.md, got %+v", settings.OutputPath)
	}
	if settings.SortServices != "name" {
		t.Fatalf("expected service sort name, got %q", settings.SortServices)
	}
	if settings.TopServices != 10 {
		t.Fatalf("expected top services 10, got %d", settings.TopServices)
	}
}

func TestLoadWarningRulesRejectsUnknownFields(t *testing.T) {
	_, err := LoadWarningRules(strings.NewReader(`{"total_budget": 300, "budget": 100}`))
	if err == nil {
		t.Fatal("expected unknown field error")
	}
	if !strings.Contains(err.Error(), `unknown field "budget"`) {
		t.Fatalf("expected unknown field error, got %v", err)
	}
}

func TestLoadWarningRulesRejectsNegativeValues(t *testing.T) {
	_, err := LoadWarningRules(strings.NewReader(`{"growth_limit_percent": -1}`))
	if err == nil {
		t.Fatal("expected negative value error")
	}
	if !strings.Contains(err.Error(), "growth_limit_percent must not be negative") {
		t.Fatalf("expected negative value error, got %v", err)
	}
}

func TestLoadRejectsEmptyIncludeService(t *testing.T) {
	_, err := Load(strings.NewReader(`{"include_services": ["Amazon EC2", " "]}`))
	if err == nil {
		t.Fatal("expected empty include service error")
	}
	if !strings.Contains(err.Error(), "include_services contains an empty service name") {
		t.Fatalf("expected empty include service error, got %v", err)
	}
}

func TestLoadRejectsInvalidServiceSort(t *testing.T) {
	_, err := Load(strings.NewReader(`{"sort_services": "month"}`))
	if err == nil {
		t.Fatal("expected invalid service sort error")
	}
	if !strings.Contains(err.Error(), `unsupported sort_services value "month"`) {
		t.Fatalf("expected invalid service sort error, got %v", err)
	}
}

func TestLoadRejectsEmptyOutputPath(t *testing.T) {
	_, err := Load(strings.NewReader(`{"output_path": " "}`))
	if err == nil {
		t.Fatal("expected empty output path error")
	}
	if !strings.Contains(err.Error(), "output_path must not be empty") {
		t.Fatalf("expected empty output path error, got %v", err)
	}
}

func TestLoadRejectsInvalidCurrency(t *testing.T) {
	_, err := Load(strings.NewReader(`{"currency": "US"}`))
	if err == nil {
		t.Fatal("expected invalid currency error")
	}
	if !strings.Contains(err.Error(), `invalid currency value "US"`) {
		t.Fatalf("expected invalid currency error, got %v", err)
	}
}

func TestLoadRejectsUnsupportedFormat(t *testing.T) {
	_, err := Load(strings.NewReader(`{"format": "xml"}`))
	if err == nil {
		t.Fatal("expected unsupported format error")
	}
	if !strings.Contains(err.Error(), `unsupported format value "xml"`) {
		t.Fatalf("expected unsupported format error, got %v", err)
	}
}

func TestLoadRejectsNegativeTopServices(t *testing.T) {
	_, err := Load(strings.NewReader(`{"top_services": -1}`))
	if err == nil {
		t.Fatal("expected negative top services error")
	}
	if !strings.Contains(err.Error(), "top_services must not be negative") {
		t.Fatalf("expected negative top services error, got %v", err)
	}
}

func TestLoadRejectsNegativeMinServiceSpend(t *testing.T) {
	_, err := Load(strings.NewReader(`{"min_service_spend": -1}`))
	if err == nil {
		t.Fatal("expected negative min service spend error")
	}
	if !strings.Contains(err.Error(), "min_service_spend must not be negative") {
		t.Fatalf("expected negative min service spend error, got %v", err)
	}
}

func TestLoadRejectsInvalidMonth(t *testing.T) {
	_, err := Load(strings.NewReader(`{"from_month": "2026-13"}`))
	if err == nil {
		t.Fatal("expected invalid month error")
	}
	if !strings.Contains(err.Error(), `invalid from_month value "2026-13", expected YYYY-MM`) {
		t.Fatalf("expected invalid month error, got %v", err)
	}
}

func TestLoadRejectsInvalidMonthRange(t *testing.T) {
	_, err := Load(strings.NewReader(`{"from_month": "2026-03", "to_month": "2026-02"}`))
	if err == nil {
		t.Fatal("expected invalid month range error")
	}
	if !strings.Contains(err.Error(), "from_month must be before or equal to to_month") {
		t.Fatalf("expected invalid month range error, got %v", err)
	}
}

func assertMonth(t *testing.T, label string, got time.Time, want string) {
	t.Helper()

	parsed, err := time.Parse("2006-01", want)
	if err != nil {
		t.Fatalf("invalid expected month %q: %v", want, err)
	}
	if got.Format("2006-01") != parsed.Format("2006-01") {
		t.Fatalf("expected %s %s, got %s", label, want, got.Format("2006-01"))
	}
}
