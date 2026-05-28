package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/kaleab-kali/marvin/internal/config"
	"github.com/kaleab-kali/marvin/internal/cost"
	"github.com/kaleab-kali/marvin/internal/report"
)

const Version = "0.1.0-dev"

var errAnalyzeHelp = errors.New("analyze help requested")

type analyzeOptions struct {
	path            string
	format          string
	ignoredServices []string
	rules           cost.WarningRules
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
	records = cost.FilterIgnoredServices(records, options.ignoredServices)

	if err := writeReport(stdout, records, options); err != nil {
		fmt.Fprintf(stderr, "write report: %v\n", err)
		return 1
	}

	return 0
}

func parseAnalyzeArgs(args []string) (analyzeOptions, error) {
	options := analyzeOptions{
		format: "terminal",
		rules:  cost.WarningRules{ServiceLimits: make(map[string]float64)},
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
		case arg == "--format":
			value, ok := nextArg(args, &i)
			if !ok {
				return analyzeOptions{}, errors.New("--format requires terminal, markdown, or json")
			}
			if err := setReportFormat(&options, value); err != nil {
				return analyzeOptions{}, err
			}
		case strings.HasPrefix(arg, "--format="):
			if err := setReportFormat(&options, strings.TrimPrefix(arg, "--format=")); err != nil {
				return analyzeOptions{}, err
			}
		case arg == "--config":
			value, ok := nextArg(args, &i)
			if !ok {
				return analyzeOptions{}, errors.New("--config requires a path")
			}
			if err := loadConfig(&options, value); err != nil {
				return analyzeOptions{}, err
			}
		case strings.HasPrefix(arg, "--config="):
			if err := loadConfig(&options, strings.TrimPrefix(arg, "--config=")); err != nil {
				return analyzeOptions{}, err
			}
		case arg == "--ignore-service":
			value, ok := nextArg(args, &i)
			if !ok {
				return analyzeOptions{}, errors.New("--ignore-service requires a service name")
			}
			if err := addIgnoredService(&options, value); err != nil {
				return analyzeOptions{}, err
			}
		case strings.HasPrefix(arg, "--ignore-service="):
			if err := addIgnoredService(&options, strings.TrimPrefix(arg, "--ignore-service=")); err != nil {
				return analyzeOptions{}, err
			}
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

func writeReport(stdout io.Writer, records []cost.Record, options analyzeOptions) error {
	switch options.format {
	case "terminal":
		return report.WriteTerminal(stdout, records, options.rules)
	case "markdown":
		return report.WriteMarkdown(stdout, records, options.rules)
	case "json":
		return report.WriteJSON(stdout, records, options.rules)
	default:
		return fmt.Errorf("unsupported report format %q", options.format)
	}
}

func loadConfig(options *analyzeOptions, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	defer file.Close()

	settings, err := config.Load(file)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	options.rules = settings.Rules
	options.ignoredServices = settings.IgnoreServices
	if options.rules.ServiceLimits == nil {
		options.rules.ServiceLimits = make(map[string]float64)
	}
	return nil
}

func addIgnoredService(options *analyzeOptions, value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return errors.New("--ignore-service requires a service name")
	}
	options.ignoredServices = append(options.ignoredServices, value)
	return nil
}

func setReportFormat(options *analyzeOptions, value string) error {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "terminal", "markdown", "json":
		options.format = value
		return nil
	default:
		return fmt.Errorf("unsupported --format %q, expected terminal, markdown, or json", value)
	}
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
  --config <path>                       Load warning rules from a JSON config file.
  --format <terminal|markdown|json>    Output format. Defaults to terminal.
  --ignore-service <service>           Exclude a service from totals and warnings. Repeatable.
  --total-budget <amount>             Warn when total spend exceeds amount.
  --service-budget <service=amount>   Warn when service spend exceeds amount. Repeatable.
  --growth-limit-percent <percent>    Warn when month-over-month growth exceeds percent.
`)
}
