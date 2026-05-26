package cost

import "testing"

func TestEvaluateWarningsReturnsTotalBudgetWarning(t *testing.T) {
	warnings := EvaluateWarnings(125, nil, nil, WarningRules{TotalLimit: 100})

	if len(warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d", len(warnings))
	}
	got := warnings[0]
	if got.Type != WarningTotalBudget {
		t.Fatalf("expected total budget warning, got %q", got.Type)
	}
	assertFloat(t, "limit", got.Limit, 100)
	assertFloat(t, "actual", got.Actual, 125)
}

func TestEvaluateWarningsReturnsServiceBudgetWarnings(t *testing.T) {
	services := []ServiceTotal{
		{Service: "Amazon EC2", Cost: 250},
		{Service: "Amazon S3", Cost: 75},
	}
	rules := WarningRules{
		ServiceLimits: map[string]float64{
			"Amazon EC2": 200,
			"Amazon S3":  100,
		},
	}

	warnings := EvaluateWarnings(0, services, nil, rules)

	if len(warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d", len(warnings))
	}
	got := warnings[0]
	if got.Type != WarningServiceBudget {
		t.Fatalf("expected service budget warning, got %q", got.Type)
	}
	if got.Service != "Amazon EC2" {
		t.Fatalf("expected Amazon EC2 warning, got %q", got.Service)
	}
	assertFloat(t, "limit", got.Limit, 200)
	assertFloat(t, "actual", got.Actual, 250)
}

func TestEvaluateWarningsReturnsGrowthWarnings(t *testing.T) {
	comparisons := []MonthComparison{
		{
			Month:         mustDate(t, "2026-02-01"),
			Cost:          125,
			PreviousCost:  100,
			Change:        25,
			ChangePercent: 25,
		},
		{
			Month:         mustDate(t, "2026-03-01"),
			Cost:          130,
			PreviousCost:  125,
			Change:        5,
			ChangePercent: 4,
		},
	}

	warnings := EvaluateWarnings(0, nil, comparisons, WarningRules{GrowthLimitPercent: 20})

	if len(warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d", len(warnings))
	}
	got := warnings[0]
	if got.Type != WarningGrowth {
		t.Fatalf("expected growth warning, got %q", got.Type)
	}
	if !got.Month.Equal(mustDate(t, "2026-02-01")) {
		t.Fatalf("expected February warning, got %s", got.Month)
	}
	assertFloat(t, "limit", got.Limit, 20)
	assertFloat(t, "actual", got.Actual, 125)
	assertFloat(t, "previous", got.Previous, 100)
	assertFloat(t, "change", got.Change, 25)
	assertFloat(t, "change percent", got.ChangePercent, 25)
}

func TestEvaluateWarningsIgnoresDisabledRules(t *testing.T) {
	warnings := EvaluateWarnings(
		125,
		[]ServiceTotal{{Service: "Amazon EC2", Cost: 250}},
		[]MonthComparison{{Cost: 125, PreviousCost: 100, Change: 25, ChangePercent: 25}},
		WarningRules{},
	)

	if len(warnings) != 0 {
		t.Fatalf("expected no warnings, got %d", len(warnings))
	}
}
