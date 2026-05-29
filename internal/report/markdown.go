package report

import (
	"fmt"
	"io"

	"github.com/kaleab-kali/marvin/internal/cost"
)

func WriteMarkdown(w io.Writer, records []cost.Record, rules cost.WarningRules) error {
	return WriteMarkdownSummary(w, BuildSummary(records, rules))
}

func WriteMarkdownSummary(w io.Writer, summary Summary) error {
	if _, err := fmt.Fprintln(w, "# Marvin Cost Report"); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "\nTotal spend: **%s**\n", formatMoney(summary.TotalSpend, summary.Currency)); err != nil {
		return err
	}

	if _, err := fmt.Fprint(w, "\n## Monthly Spend\n\n"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "| Month | Cost |"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "| --- | ---: |"); err != nil {
		return err
	}
	for _, month := range summary.MonthlySpend {
		if _, err := fmt.Fprintf(w, "| %s | %s |\n", month.Month, formatMoney(month.Cost, summary.Currency)); err != nil {
			return err
		}
	}

	if len(summary.MonthOverMonth) > 0 {
		if _, err := fmt.Fprint(w, "\n## Month-over-Month\n\n"); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w, "| Month | Previous | Current | Change | Change % |"); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w, "| --- | ---: | ---: | ---: | ---: |"); err != nil {
			return err
		}
		for _, comparison := range summary.MonthOverMonth {
			if _, err := fmt.Fprintf(w, "| %s | %s | %s | %s | %s |\n", comparison.Month, formatMoney(comparison.PreviousCost, summary.Currency), formatMoney(comparison.Cost, summary.Currency), formatSignedMoney(comparison.Change, summary.Currency), formatPercent(comparison.ChangePercent)); err != nil {
				return err
			}
		}
	}

	if _, err := fmt.Fprint(w, "\n## Service Spend\n\n"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "| Service | Cost |"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "| --- | ---: |"); err != nil {
		return err
	}
	for _, service := range summary.ServiceSpend {
		if _, err := fmt.Fprintf(w, "| %s | %s |\n", service.Service, formatMoney(service.Cost, summary.Currency)); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprint(w, "\n## Warnings\n\n"); err != nil {
		return err
	}
	if len(summary.Warnings) == 0 {
		_, err := fmt.Fprintln(w, "None.")
		return err
	}
	for _, warning := range summary.Warnings {
		if _, err := fmt.Fprintf(w, "- %s\n", formatWarning(warning, summary.Currency)); err != nil {
			return err
		}
	}

	return nil
}
