package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/kaleab-kali/marvin/internal/cost"
	"github.com/kaleab-kali/marvin/internal/report"
)

const Version = "0.1.0-dev"

var errAnalyzeHelp = errors.New("analyze help requested")

type analyzeOptions struct {
	path  string
	rules cost.WarningRules
}

func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printUsage(stdout)
		return 0
	}

	switch args[0] {
	case "-h", "--help", "help":
		printUsage(stdout)
		return 0
	case "-v", "--version", "version":
		fmt.Fprintf(stdout, "marvin %s\n", Version)
		return 0
	case "analyze":
		return runAnalyze(args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "unknown command %q\n\n", args[0])
		printUsage(stderr)
		return 2
	}
}

func runAnalyze(args []string, stdout, stderr io.Writer) int {
	options, err := parseAnalyzeArgs(args)
	if errors.Is(err, errAnalyzeHelp) {
		printAnalyzeUsage(stdout)
		return 0
	}
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}

	file, err := os.Open(options.path)
	if err != nil {
		fmt.Fprintf(stderr, "open cost CSV: %v\n", err)
		return 1
	}
	defer file.Close()

	records, err := cost.ParseCostExplorerCSV(file)
	if err != nil {
		fmt.Fprintf(stderr, "parse cost CSV: %v\n", err)
		return 1
	}

	if err := report.WriteTerminal(stdout, records, options.rules); err != nil {
		fmt.Fprintf(stderr, "write report: %v\n", err)
		return 1
	}

	return 0
}

func parseAnalyzeArgs(args []string) (analyzeOptions, error) {
	options := analyzeOptions{
		rules: cost.WarningRules{ServiceLimits: make(map[string]float64)},
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-h" || arg == "--help":
			return analyzeOptions{}, errAnalyzeHelp
		case arg == "--total-budget":
			value, ok := nextArg(args, &i)
			if !ok {
				return analyzeOptions{}, errors.New("--total-budget requires an amount")
			}
			amount, err := parsePositiveFloat("--total-budget", value)
			if err != nil {
				return analyzeOptions{}, err
			}
			options.rules.TotalLimit = amount
		case strings.HasPrefix(arg, "--total-budget="):
			amount, err := parsePositiveFloat("--total-budget", strings.TrimPrefix(arg, "--total-budget="))
			if err != nil {
				return analyzeOptions{}, err
			}
			options.rules.TotalLimit = amount
		case arg == "--growth-limit-percent":
			value, ok := nextArg(args, &i)
			if !ok {
				return analyzeOptions{}, errors.New("--growth-limit-percent requires a percent")
			}
			amount, err := parsePositiveFloat("--growth-limit-percent", value)
			if err != nil {
				return analyzeOptions{}, err
			}
			options.rules.GrowthLimitPercent = amount
		case strings.HasPrefix(arg, "--growth-limit-percent="):
			amount, err := parsePositiveFloat("--growth-limit-percent", strings.TrimPrefix(arg, "--growth-limit-percent="))
			if err != nil {
				return analyzeOptions{}, err
			}
			options.rules.GrowthLimitPercent = amount
		case arg == "--service-budget":
			value, ok := nextArg(args, &i)
			if !ok {
				return analyzeOptions{}, errors.New("--service-budget requires service=amount")
			}
			if err := parseServiceBudget(options.rules.ServiceLimits, value); err != nil {
				return analyzeOptions{}, err
			}
		case strings.HasPrefix(arg, "--service-budget="):
			if err := parseServiceBudget(options.rules.ServiceLimits, strings.TrimPrefix(arg, "--service-budget=")); err != nil {
				return analyzeOptions{}, err
			}
		case strings.HasPrefix(arg, "-"):
			return analyzeOptions{}, fmt.Errorf("unknown analyze flag %q", arg)
		default:
			if options.path != "" {
				return analyzeOptions{}, fmt.Errorf("unexpected extra argument %q", arg)
			}
			options.path = arg
		}
	}

	if options.path == "" {
		return analyzeOptions{}, errors.New("analyze requires a Cost Explorer CSV path")
	}

	return options, nil
}

func nextArg(args []string, index *int) (string, bool) {
	if *index+1 >= len(args) {
		return "", false
	}
	*index = *index + 1
	return args[*index], true
}

func parseServiceBudget(serviceLimits map[string]float64, value string) error {
	service, amountText, ok := strings.Cut(value, "=")
	if !ok || strings.TrimSpace(service) == "" || strings.TrimSpace(amountText) == "" {
		return fmt.Errorf("invalid --service-budget %q, expected service=amount", value)
	}

	amount, err := parsePositiveFloat("--service-budget", amountText)
	if err != nil {
		return err
	}
	serviceLimits[strings.TrimSpace(service)] = amount
	return nil
}

func parsePositiveFloat(name, value string) (float64, error) {
	amount, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s value %q", name, value)
	}
	if amount <= 0 {
		return 0, fmt.Errorf("%s must be greater than zero", name)
	}
	return amount, nil
}

func printUsage(w io.Writer) {
	fmt.Fprint(w, `Marvin reads exported AWS Cost Explorer CSV files and reports cost changes.

Usage:
  marvin analyze [flags] <cost-explorer.csv>
  marvin version
  marvin help

Status:
  CSV MVP. Cost analysis runs locally from exported Cost Explorer CSV files.
`)
}

func printAnalyzeUsage(w io.Writer) {
	fmt.Fprint(w, `Usage:
  marvin analyze [flags] <cost-explorer.csv>

Flags:
  --total-budget <amount>             Warn when total spend exceeds amount.
  --service-budget <service=amount>   Warn when service spend exceeds amount. Repeatable.
  --growth-limit-percent <percent>    Warn when month-over-month growth exceeds percent.
`)
}
