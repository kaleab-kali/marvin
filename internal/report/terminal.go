package report

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/kaleab-kali/marvin/internal/cost"
)

func WriteTerminal(w io.Writer, records []cost.Record, rules cost.WarningRules) error {
	total := cost.TotalSpend(records)
	services := cost.GroupByService(records)
	months := cost.GroupByMonth(records)
	comparisons := cost.CompareMonths(months)
	warnings := cost.EvaluateWarnings(total, services, comparisons, rules)

	tabbed := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)

	fmt.Fprintln(tabbed, "Marvin Cost Report")
	fmt.Fprintf(tabbed, "Total spend: %s\n\n", formatMoney(total))

	fmt.Fprintln(tabbed, "Monthly spend")
	fmt.Fprintln(tabbed, "Month\tCost")
	for _, month := range months {
		fmt.Fprintf(tabbed, "%s\t%s\n", month.Month.Format("2006-01"), formatMoney(month.Cost))
	}

	if len(comparisons) > 0 {
		fmt.Fprintln(tabbed)
		fmt.Fprintln(tabbed, "Month-over-month")
		fmt.Fprintln(tabbed, "Month\tPrevious\tCurrent\tChange\tChange %")
		for _, comparison := range comparisons {
			fmt.Fprintf(
				tabbed,
				"%s\t%s\t%s\t%s\t%s\n",
				comparison.Month.Format("2006-01"),
				formatMoney(comparison.PreviousCost),
				formatMoney(comparison.Cost),
				formatSignedMoney(comparison.Change),
				formatPercent(comparison.ChangePercent),
			)
		}
	}

	fmt.Fprintln(tabbed)
	fmt.Fprintln(tabbed, "Service spend")
	fmt.Fprintln(tabbed, "Service\tCost")
	for _, service := range services {
		fmt.Fprintf(tabbed, "%s\t%s\n", service.Service, formatMoney(service.Cost))
	}

	fmt.Fprintln(tabbed)
	fmt.Fprintln(tabbed, "Warnings")
	if len(warnings) == 0 {
		fmt.Fprintln(tabbed, "None")
	} else {
		for _, warning := range warnings {
			fmt.Fprintf(tabbed, "- %s\n", formatWarning(warning))
		}
	}

	return tabbed.Flush()
}

func formatWarning(warning cost.Warning) string {
	switch warning.Type {
	case cost.WarningTotalBudget:
		return fmt.Sprintf("total spend %s exceeds budget %s", formatMoney(warning.Actual), formatMoney(warning.Limit))
	case cost.WarningServiceBudget:
		return fmt.Sprintf("%s spend %s exceeds budget %s", warning.Service, formatMoney(warning.Actual), formatMoney(warning.Limit))
	case cost.WarningGrowth:
		return fmt.Sprintf("%s spend grew %s from %s to %s", warning.Month.Format("2006-01"), formatPercent(warning.ChangePercent), formatMoney(warning.Previous), formatMoney(warning.Actual))
	default:
		return "unknown warning"
	}
}

func formatMoney(value float64) string {
	return fmt.Sprintf("$%.2f", value)
}

func formatSignedMoney(value float64) string {
	if value > 0 {
		return fmt.Sprintf("+$%.2f", value)
	}
	if value < 0 {
		return fmt.Sprintf("-$%.2f", -value)
	}
	return "$0.00"
}

func formatPercent(value float64) string {
	if value > 0 {
		return fmt.Sprintf("+%.2f%%", value)
	}
	return fmt.Sprintf("%.2f%%", value)
}
