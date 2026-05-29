package report

import (
	"math"
	"strings"

	"github.com/kaleab-kali/marvin/internal/cost"
)

type Summary struct {
	TotalSpend     float64           `json:"total_spend"`
	Currency       string            `json:"currency"`
	MonthlySpend   []MonthSpend      `json:"monthly_spend"`
	MonthOverMonth []MonthComparison `json:"month_over_month,omitempty"`
	ServiceSpend   []ServiceSpend    `json:"service_spend"`
	Warnings       []Warning         `json:"warnings,omitempty"`
}

type MonthSpend struct {
	Month string  `json:"month"`
	Cost  float64 `json:"cost"`
}

type MonthComparison struct {
	Month         string  `json:"month"`
	PreviousMonth string  `json:"previous_month"`
	Cost          float64 `json:"cost"`
	PreviousCost  float64 `json:"previous_cost"`
	Change        float64 `json:"change"`
	ChangePercent float64 `json:"change_percent"`
}

type ServiceSpend struct {
	Service string  `json:"service"`
	Cost    float64 `json:"cost"`
}

type Warning struct {
	Type          string  `json:"type"`
	Service       string  `json:"service,omitempty"`
	Month         string  `json:"month,omitempty"`
	Limit         float64 `json:"limit,omitempty"`
	Actual        float64 `json:"actual,omitempty"`
	Previous      float64 `json:"previous,omitempty"`
	Change        float64 `json:"change,omitempty"`
	ChangePercent float64 `json:"change_percent,omitempty"`
}

func BuildSummary(records []cost.Record, rules cost.WarningRules) Summary {
	total := cost.TotalSpend(records)
	services := cost.GroupByService(records)
	months := cost.GroupByMonth(records)
	comparisons := cost.CompareMonths(months)
	warnings := cost.EvaluateWarnings(total, services, comparisons, rules)

	return Summary{
		TotalSpend:     round(total, 2),
		Currency:       summarizeCurrency(records),
		MonthlySpend:   summarizeMonths(months),
		MonthOverMonth: summarizeComparisons(comparisons),
		ServiceSpend:   summarizeServices(services),
		Warnings:       summarizeWarnings(warnings),
	}
}

func summarizeMonths(months []cost.MonthTotal) []MonthSpend {
	summary := make([]MonthSpend, 0, len(months))
	for _, month := range months {
		summary = append(summary, MonthSpend{
			Month: month.Month.Format("2006-01"),
			Cost:  round(month.Cost, 2),
		})
	}
	return summary
}

func summarizeComparisons(comparisons []cost.MonthComparison) []MonthComparison {
	summary := make([]MonthComparison, 0, len(comparisons))
	for _, comparison := range comparisons {
		summary = append(summary, MonthComparison{
			Month:         comparison.Month.Format("2006-01"),
			PreviousMonth: comparison.PreviousMonth.Format("2006-01"),
			Cost:          round(comparison.Cost, 2),
			PreviousCost:  round(comparison.PreviousCost, 2),
			Change:        round(comparison.Change, 2),
			ChangePercent: round(comparison.ChangePercent, 2),
		})
	}
	return summary
}

func summarizeServices(services []cost.ServiceTotal) []ServiceSpend {
	summary := make([]ServiceSpend, 0, len(services))
	for _, service := range services {
		summary = append(summary, ServiceSpend{
			Service: service.Service,
			Cost:    round(service.Cost, 2),
		})
	}
	return summary
}

func summarizeCurrency(records []cost.Record) string {
	currency := "USD"
	for _, record := range records {
		value := strings.ToUpper(strings.TrimSpace(record.Currency))
		if value != "" {
			currency = value
			break
		}
	}
	return currency
}

func summarizeWarnings(warnings []cost.Warning) []Warning {
	summary := make([]Warning, 0, len(warnings))
	for _, warning := range warnings {
		item := Warning{
			Type:          string(warning.Type),
			Service:       warning.Service,
			Limit:         round(warning.Limit, 2),
			Actual:        round(warning.Actual, 2),
			Previous:      round(warning.Previous, 2),
			Change:        round(warning.Change, 2),
			ChangePercent: round(warning.ChangePercent, 2),
		}
		if !warning.Month.IsZero() {
			item.Month = warning.Month.Format("2006-01")
		}
		summary = append(summary, item)
	}
	return summary
}

func round(value float64, places int) float64 {
	scale := math.Pow(10, float64(places))
	return math.Round(value*scale) / scale
}
