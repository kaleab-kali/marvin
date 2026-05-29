package report

import (
	"encoding/csv"
	"io"
	"strconv"

	"github.com/kaleab-kali/marvin/internal/cost"
)

func WriteCSV(w io.Writer, records []cost.Record, rules cost.WarningRules) error {
	return WriteCSVSummary(w, BuildSummary(records, rules))
}

func WriteCSVSummary(w io.Writer, summary Summary) error {
	writer := csv.NewWriter(w)

	if err := writer.Write([]string{
		"section",
		"month",
		"previous_month",
		"service",
		"currency",
		"cost",
		"previous_cost",
		"change",
		"change_percent",
		"warning_type",
		"limit",
		"actual",
		"previous",
	}); err != nil {
		return err
	}

	if err := writer.Write([]string{"total", "", "", "", summary.Currency, formatCSVNumber(summary.TotalSpend), "", "", "", "", "", "", ""}); err != nil {
		return err
	}
	for _, month := range summary.MonthlySpend {
		if err := writer.Write([]string{"monthly_spend", month.Month, "", "", summary.Currency, formatCSVNumber(month.Cost), "", "", "", "", "", "", ""}); err != nil {
			return err
		}
	}
	for _, comparison := range summary.MonthOverMonth {
		if err := writer.Write([]string{
			"month_over_month",
			comparison.Month,
			comparison.PreviousMonth,
			"",
			summary.Currency,
			formatCSVNumber(comparison.Cost),
			formatCSVNumber(comparison.PreviousCost),
			formatCSVNumber(comparison.Change),
			formatCSVNumber(comparison.ChangePercent),
			"",
			"",
			"",
			"",
		}); err != nil {
			return err
		}
	}
	for _, service := range summary.ServiceSpend {
		if err := writer.Write([]string{"service_spend", "", "", service.Service, summary.Currency, formatCSVNumber(service.Cost), "", "", "", "", "", "", ""}); err != nil {
			return err
		}
	}
	for _, warning := range summary.Warnings {
		if err := writer.Write(csvWarningRow(summary.Currency, warning)); err != nil {
			return err
		}
	}

	writer.Flush()
	return writer.Error()
}

func formatCSVNumber(value float64) string {
	return strconv.FormatFloat(value, 'f', 2, 64)
}

func csvWarningRow(currency string, warning Warning) []string {
	row := []string{
		"warning",
		warning.Month,
		"",
		warning.Service,
		currency,
		"",
		"",
		"",
		"",
		warning.Type,
		"",
		"",
		"",
	}

	switch warning.Type {
	case string(cost.WarningTotalBudget), string(cost.WarningServiceBudget):
		row[10] = formatCSVNumber(warning.Limit)
		row[11] = formatCSVNumber(warning.Actual)
	case string(cost.WarningGrowth):
		row[7] = formatCSVNumber(warning.Change)
		row[8] = formatCSVNumber(warning.ChangePercent)
		row[10] = formatCSVNumber(warning.Limit)
		row[11] = formatCSVNumber(warning.Actual)
		row[12] = formatCSVNumber(warning.Previous)
	}

	return row
}
