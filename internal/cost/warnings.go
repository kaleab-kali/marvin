package cost

import "time"

type WarningType string

const (
	WarningTotalBudget   WarningType = "total_budget"
	WarningServiceBudget WarningType = "service_budget"
	WarningGrowth        WarningType = "growth"
)

type WarningRules struct {
	TotalLimit         float64
	ServiceLimits      map[string]float64
	GrowthLimitPercent float64
}

type Warning struct {
	Type          WarningType
	Service       string
	Month         time.Time
	Limit         float64
	Actual        float64
	Previous      float64
	Change        float64
	ChangePercent float64
}

func EvaluateWarnings(total float64, services []ServiceTotal, comparisons []MonthComparison, rules WarningRules) []Warning {
	var warnings []Warning

	if rules.TotalLimit > 0 && total > rules.TotalLimit {
		warnings = append(warnings, Warning{
			Type:   WarningTotalBudget,
			Limit:  rules.TotalLimit,
			Actual: total,
		})
	}

	for _, service := range services {
		limit := rules.ServiceLimits[service.Service]
		if limit <= 0 || service.Cost <= limit {
			continue
		}

		warnings = append(warnings, Warning{
			Type:    WarningServiceBudget,
			Service: service.Service,
			Limit:   limit,
			Actual:  service.Cost,
		})
	}

	for _, comparison := range comparisons {
		if rules.GrowthLimitPercent <= 0 || comparison.ChangePercent <= rules.GrowthLimitPercent {
			continue
		}

		warnings = append(warnings, Warning{
			Type:          WarningGrowth,
			Month:         comparison.Month,
			Limit:         rules.GrowthLimitPercent,
			Actual:        comparison.Cost,
			Previous:      comparison.PreviousCost,
			Change:        comparison.Change,
			ChangePercent: comparison.ChangePercent,
		})
	}

	return warnings
}
