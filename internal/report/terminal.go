package report

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/kaleab-kali/marvin/internal/cost"
)

func WriteTerminal(w io.Writer, records []cost.Record, rules cost.WarningRules) error {
	return WriteTerminalSummary(w, BuildSummary(records, rules))
}

func WriteTerminalSummary(w io.Writer, summary Summary) error {
	tabbed := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)

	fmt.Fprintln(tabbed, "Marvin Cost Report")
	fmt.Fprintf(tabbed, "Total spend: %s\n\n", formatMoney(summary.TotalSpend, summary.Currency))

	fmt.Fprintln(tabbed, "Monthly spend")
	fmt.Fprintln(tabbed, "Month\tCost")
	for _, month := range summary.MonthlySpend {
		fmt.Fprintf(tabbed, "%s\t%s\n", month.Month, formatMoney(month.Cost, summary.Currency))
	}

	if len(summary.MonthOverMonth) > 0 {
		fmt.Fprintln(tabbed)
		fmt.Fprintln(tabbed, "Month-over-month")
		fmt.Fprintln(tabbed, "Month\tPrevious\tCurrent\tChange\tChange %")
		for _, comparison := range summary.MonthOverMonth {
			fmt.Fprintf(
				tabbed,
				"%s\t%s\t%s\t%s\t%s\n",
				comparison.Month,
				formatMoney(comparison.PreviousCost, summary.Currency),
				formatMoney(comparison.Cost, summary.Currency),
				formatSignedMoney(comparison.Change, summary.Currency),
				formatPercent(comparison.ChangePercent),
			)
		}
	}

	fmt.Fprintln(tabbed)
	fmt.Fprintln(tabbed, "Service spend")
	fmt.Fprintln(tabbed, "Service\tCost\tShare")
	for _, service := range summary.ServiceSpend {
		fmt.Fprintf(tabbed, "%s\t%s\t%s\n", service.Service, formatMoney(service.Cost, summary.Currency), formatSharePercent(service.SharePercent))
	}

	fmt.Fprintln(tabbed)
	fmt.Fprintln(tabbed, "Warnings")
	if len(summary.Warnings) == 0 {
		fmt.Fprintln(tabbed, "None")
	} else {
		for _, warning := range summary.Warnings {
			fmt.Fprintf(tabbed, "- %s\n", formatWarning(warning, summary.Currency))
		}
	}

	return tabbed.Flush()
}

func formatWarning(warning Warning, currency string) string {
	switch warning.Type {
	case string(cost.WarningTotalBudget):
		return fmt.Sprintf("total spend %s exceeds budget %s", formatMoney(warning.Actual, currency), formatMoney(warning.Limit, currency))
	case string(cost.WarningServiceBudget):
		return fmt.Sprintf("%s spend %s exceeds budget %s", warning.Service, formatMoney(warning.Actual, currency), formatMoney(warning.Limit, currency))
	case string(cost.WarningGrowth):
		return fmt.Sprintf("%s spend grew %s from %s to %s", warning.Month, formatPercent(warning.ChangePercent), formatMoney(warning.Previous, currency), formatMoney(warning.Actual, currency))
	default:
		return "unknown warning"
	}
}

func formatMoney(value float64, currency string) string {
	if currency != "" && currency != "USD" {
		return fmt.Sprintf("%s %.2f", currency, value)
	}
	return fmt.Sprintf("$%.2f", value)
}

func formatSignedMoney(value float64, currency string) string {
	if value > 0 {
		if currency != "" && currency != "USD" {
			return fmt.Sprintf("+%s %.2f", currency, value)
		}
		return fmt.Sprintf("+$%.2f", value)
	}
	if value < 0 {
		if currency != "" && currency != "USD" {
			return fmt.Sprintf("-%s %.2f", currency, -value)
		}
		return fmt.Sprintf("-$%.2f", -value)
	}
	return formatMoney(0, currency)
}

func formatPercent(value float64) string {
	if value > 0 {
		return fmt.Sprintf("+%.2f%%", value)
	}
	return fmt.Sprintf("%.2f%%", value)
}

func formatSharePercent(value float64) string {
	return fmt.Sprintf("%.2f%%", value)
}
