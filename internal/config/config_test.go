package config

import (
	"strings"
	"testing"
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
