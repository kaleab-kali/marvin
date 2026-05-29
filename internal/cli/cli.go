package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/kaleab-kali/marvin/internal/config"
	"github.com/kaleab-kali/marvin/internal/cost"
	"github.com/kaleab-kali/marvin/internal/report"
)

var Version = "0.1.0-dev"

const (
	ExitOK           = 0
	ExitRuntimeError = 1
	ExitUsageError   = 2
	ExitWarning      = 3
)

var errAnalyzeHelp = errors.New("analyze help requested")
var errConfigHelp = errors.New("config help requested")
var errSampleHelp = errors.New("sample help requested")

type analyzeOptions struct {
	paths           []string
	format          string
	outputPath      string
	failOnWarning   bool
	fromMonth       time.Time
	ignoredServices []string
	toMonth         time.Time
	topServices     int
	rules           cost.WarningRules
}

func Run(args []string, stdout, stderr io.Writer) int {
	return RunWithIO(args, os.Stdin, stdout, stderr)
}

func RunWithIO(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printUsage(stdout)
		return ExitOK
	}

	switch args[0] {
	case "-h", "--help", "help":
		printUsage(stdout)
		return ExitOK
	case "-v", "--version", "version":
		fmt.Fprintf(stdout, "marvin %s\n", Version)
		return ExitOK
	case "analyze":
		return runAnalyze(args[1:], stdin, stdout, stderr)
	case "config":
		return runConfigCommand(args[1:], stdout, stderr)
	case "sample":
		return runSample(args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "unknown command %q\n\n", args[0])
		printUsage(stderr)
		return ExitUsageError
	}
}

func runAnalyze(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	options, err := parseAnalyzeArgs(args)
	if errors.Is(err, errAnalyzeHelp) {
		printAnalyzeUsage(stdout)
		return ExitOK
	}
	if err != nil {
		fmt.Fprintln(stderr, err)
		return ExitUsageError
	}

	records, err := readCostRecords(options.paths, stdin)
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return ExitRuntimeError
	}
	records = cost.FilterIgnoredServices(records, options.ignoredServices)
	records = filterRecordsByMonth(records, options.fromMonth, options.toMonth)
	summary := report.BuildSummary(records, options.rules)
	summary = limitServiceSpend(summary, options.topServices)

	output := stdout
	var outputFile *os.File
	if options.outputPath != "" {
		outputFile, err = os.Create(options.outputPath)
		if err != nil {
			fmt.Fprintf(stderr, "create output file: %v\n", err)
			return 1
		}
		defer outputFile.Close()
		output = outputFile
	}

	if err := writeReport(output, summary, options); err != nil {
		fmt.Fprintf(stderr, "write report: %v\n", err)
		return ExitRuntimeError
	}

	if options.failOnWarning && len(summary.Warnings) > 0 {
		return ExitWarning
	}

	return ExitOK
}

func readCostRecords(paths []string, stdin io.Reader) ([]cost.Record, error) {
	var records []cost.Record
	for _, path := range paths {
		input := stdin
		var file *os.File
		if path != "-" {
			var err error
			file, err = os.Open(path)
			if err != nil {
				return nil, fmt.Errorf("open cost CSV %q: %w", path, err)
			}
			input = file
		}

		parsed, err := cost.ParseCostExplorerCSV(input)
		if file != nil {
			if closeErr := file.Close(); err == nil && closeErr != nil {
				err = closeErr
			}
		}
		if err != nil {
			return nil, fmt.Errorf("parse cost CSV %q: %w", path, err)
		}
		records = append(records, parsed...)
	}
	return records, nil
}

func runConfigCommand(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printConfigUsage(stdout)
		return ExitOK
	}

	switch args[0] {
	case "-h", "--help", "help":
		printConfigUsage(stdout)
		return ExitOK
	case "sample":
		return runConfigSample(args[1:], stdout, stderr)
	case "validate":
		return runConfigValidate(args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "unknown config command %q\n\n", args[0])
		printConfigUsage(stderr)
		return ExitUsageError
	}
}

func runConfigSample(args []string, stdout, stderr io.Writer) int {
	outputPath, err := parseConfigSampleArgs(args)
	if errors.Is(err, errConfigHelp) {
		printConfigSampleUsage(stdout)
		return ExitOK
	}
	if err != nil {
		fmt.Fprintln(stderr, err)
		return ExitUsageError
	}

	output := stdout
	var outputFile *os.File
	if outputPath != "" {
		outputFile, err = os.Create(outputPath)
		if err != nil {
			fmt.Fprintf(stderr, "create sample config: %v\n", err)
			return ExitRuntimeError
		}
		defer outputFile.Close()
		output = outputFile
	}

	if _, err := io.WriteString(output, sampleConfigJSON); err != nil {
		fmt.Fprintf(stderr, "write sample config: %v\n", err)
		return ExitRuntimeError
	}

	return ExitOK
}

func runConfigValidate(args []string, stdout, stderr io.Writer) int {
	path, err := parseConfigValidateArgs(args)
	if errors.Is(err, errConfigHelp) {
		printConfigValidateUsage(stdout)
		return ExitOK
	}
	if err != nil {
		fmt.Fprintln(stderr, err)
		return ExitUsageError
	}

	file, err := os.Open(path)
	if err != nil {
		fmt.Fprintf(stderr, "open config: %v\n", err)
		return ExitRuntimeError
	}
	defer file.Close()

	if _, err := config.Load(file); err != nil {
		fmt.Fprintf(stderr, "validate config: %v\n", err)
		return ExitRuntimeError
	}

	fmt.Fprintf(stdout, "config %s is valid\n", path)
	return ExitOK
}

func runSample(args []string, stdout, stderr io.Writer) int {
	outputPath, err := parseSampleArgs(args)
	if errors.Is(err, errSampleHelp) {
		printSampleUsage(stdout)
		return ExitOK
	}
	if err != nil {
		fmt.Fprintln(stderr, err)
		return ExitUsageError
	}

	output := stdout
	var outputFile *os.File
	if outputPath != "" {
		outputFile, err = os.Create(outputPath)
		if err != nil {
			fmt.Fprintf(stderr, "create sample CSV: %v\n", err)
			return ExitRuntimeError
		}
		defer outputFile.Close()
		output = outputFile
	}

	if _, err := io.WriteString(output, sampleCostExplorerCSV); err != nil {
		fmt.Fprintf(stderr, "write sample CSV: %v\n", err)
		return ExitRuntimeError
	}

	return ExitOK
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
		case arg == "--from":
			value, ok := nextArg(args, &i)
			if !ok {
				return analyzeOptions{}, errors.New("--from requires a month")
			}
			month, err := parseMonthFlag("--from", value)
			if err != nil {
				return analyzeOptions{}, err
			}
			options.fromMonth = month
		case strings.HasPrefix(arg, "--from="):
			month, err := parseMonthFlag("--from", strings.TrimPrefix(arg, "--from="))
			if err != nil {
				return analyzeOptions{}, err
			}
			options.fromMonth = month
		case arg == "--fail-on-warning":
			options.failOnWarning = true
		case arg == "--output":
			value, ok := nextArg(args, &i)
			if !ok {
				return analyzeOptions{}, errors.New("--output requires a path")
			}
			if err := setOutputPath(&options, value); err != nil {
				return analyzeOptions{}, err
			}
		case strings.HasPrefix(arg, "--output="):
			if err := setOutputPath(&options, strings.TrimPrefix(arg, "--output=")); err != nil {
				return analyzeOptions{}, err
			}
		case arg == "--top-services":
			value, ok := nextArg(args, &i)
			if !ok {
				return analyzeOptions{}, errors.New("--top-services requires a count")
			}
			count, err := parsePositiveInt("--top-services", value)
			if err != nil {
				return analyzeOptions{}, err
			}
			options.topServices = count
		case strings.HasPrefix(arg, "--top-services="):
			count, err := parsePositiveInt("--top-services", strings.TrimPrefix(arg, "--top-services="))
			if err != nil {
				return analyzeOptions{}, err
			}
			options.topServices = count
		case arg == "--to":
			value, ok := nextArg(args, &i)
			if !ok {
				return analyzeOptions{}, errors.New("--to requires a month")
			}
			month, err := parseMonthFlag("--to", value)
			if err != nil {
				return analyzeOptions{}, err
			}
			options.toMonth = month
		case strings.HasPrefix(arg, "--to="):
			month, err := parseMonthFlag("--to", strings.TrimPrefix(arg, "--to="))
			if err != nil {
				return analyzeOptions{}, err
			}
			options.toMonth = month
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
		case arg == "-":
			if containsString(options.paths, "-") {
				return analyzeOptions{}, errors.New("analyze accepts standard input only once")
			}
			options.paths = append(options.paths, arg)
		case strings.HasPrefix(arg, "-"):
			return analyzeOptions{}, fmt.Errorf("unknown analyze flag %q", arg)
		default:
			options.paths = append(options.paths, arg)
		}
	}

	if len(options.paths) == 0 {
		return analyzeOptions{}, errors.New("analyze requires a Cost Explorer CSV path")
	}
	if !options.fromMonth.IsZero() && !options.toMonth.IsZero() && options.fromMonth.After(options.toMonth) {
		return analyzeOptions{}, errors.New("--from must be before or equal to --to")
	}

	return options, nil
}

func parseConfigSampleArgs(args []string) (string, error) {
	var outputPath string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-h" || arg == "--help":
			return "", errConfigHelp
		case arg == "--output":
			value, ok := nextArg(args, &i)
			if !ok {
				return "", errors.New("--output requires a path")
			}
			path, err := sampleOutputPath(value)
			if err != nil {
				return "", err
			}
			outputPath = path
		case strings.HasPrefix(arg, "--output="):
			path, err := sampleOutputPath(strings.TrimPrefix(arg, "--output="))
			if err != nil {
				return "", err
			}
			outputPath = path
		case strings.HasPrefix(arg, "-"):
			return "", fmt.Errorf("unknown config sample flag %q", arg)
		default:
			return "", fmt.Errorf("unexpected config sample argument %q", arg)
		}
	}
	return outputPath, nil
}

func parseConfigValidateArgs(args []string) (string, error) {
	var path string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-h" || arg == "--help":
			return "", errConfigHelp
		case strings.HasPrefix(arg, "-"):
			return "", fmt.Errorf("unknown config validate flag %q", arg)
		default:
			if path != "" {
				return "", fmt.Errorf("unexpected extra argument %q", arg)
			}
			path = arg
		}
	}
	if path == "" {
		return "", errors.New("config validate requires a config path")
	}
	return path, nil
}

func parseSampleArgs(args []string) (string, error) {
	var outputPath string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-h" || arg == "--help":
			return "", errSampleHelp
		case arg == "--output":
			value, ok := nextArg(args, &i)
			if !ok {
				return "", errors.New("--output requires a path")
			}
			path, err := sampleOutputPath(value)
			if err != nil {
				return "", err
			}
			outputPath = path
		case strings.HasPrefix(arg, "--output="):
			path, err := sampleOutputPath(strings.TrimPrefix(arg, "--output="))
			if err != nil {
				return "", err
			}
			outputPath = path
		case strings.HasPrefix(arg, "-"):
			return "", fmt.Errorf("unknown sample flag %q", arg)
		default:
			return "", fmt.Errorf("unexpected sample argument %q", arg)
		}
	}
	return outputPath, nil
}

func writeReport(stdout io.Writer, summary report.Summary, options analyzeOptions) error {
	switch options.format {
	case "terminal":
		return report.WriteTerminalSummary(stdout, summary)
	case "markdown":
		return report.WriteMarkdownSummary(stdout, summary)
	case "json":
		return report.WriteJSONSummary(stdout, summary)
	default:
		return fmt.Errorf("unsupported report format %q", options.format)
	}
}

func setOutputPath(options *analyzeOptions, value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return errors.New("--output requires a path")
	}
	options.outputPath = value
	return nil
}

func sampleOutputPath(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", errors.New("--output requires a path")
	}
	return value, nil
}

func limitServiceSpend(summary report.Summary, topServices int) report.Summary {
	if topServices <= 0 || topServices >= len(summary.ServiceSpend) {
		return summary
	}
	summary.ServiceSpend = summary.ServiceSpend[:topServices]
	return summary
}

func filterRecordsByMonth(records []cost.Record, fromMonth, toMonth time.Time) []cost.Record {
	if fromMonth.IsZero() && toMonth.IsZero() {
		return records
	}

	filtered := make([]cost.Record, 0, len(records))
	for _, record := range records {
		month := cost.Month(record.StartDate)
		if !fromMonth.IsZero() && month.Before(fromMonth) {
			continue
		}
		if !toMonth.IsZero() && month.After(toMonth) {
			continue
		}
		filtered = append(filtered, record)
	}
	return filtered
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
	options.fromMonth = settings.FromMonth
	options.ignoredServices = settings.IgnoreServices
	options.toMonth = settings.ToMonth
	options.topServices = settings.TopServices
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
	case "terminal", "text":
		options.format = "terminal"
		return nil
	case "markdown", "md":
		options.format = "markdown"
		return nil
	case "json":
		options.format = "json"
		return nil
	default:
		return fmt.Errorf("unsupported --format %q, expected terminal, markdown, md, json, or text", value)
	}
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
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

func parseMonthFlag(name, value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	parsed, err := time.Parse("2006-01", value)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid %s value %q, expected YYYY-MM", name, value)
	}
	return time.Date(parsed.Year(), parsed.Month(), 1, 0, 0, 0, 0, time.UTC), nil
}

func parsePositiveInt(name, value string) (int, error) {
	amount, err := strconv.Atoi(strings.TrimSpace(value))
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
  marvin analyze [flags] <cost-explorer.csv|-> [more.csv ...]
  marvin config sample [flags]
  marvin config validate <marvin.json>
  marvin sample [flags]
  marvin version
  marvin help

Status:
  CSV MVP. Cost analysis runs locally from exported Cost Explorer CSV files.
`)
}

func printAnalyzeUsage(w io.Writer) {
	fmt.Fprint(w, `Usage:
  marvin analyze [flags] <cost-explorer.csv|-> [more.csv ...]

Flags:
  --config <path>                       Load warning rules from a JSON config file.
  --fail-on-warning                     Exit with code 3 when warnings are present.
  --format <terminal|markdown|json>    Output format. Defaults to terminal.
  --from <YYYY-MM>                     Include records from this month onward.
  --ignore-service <service>           Exclude a service from totals and warnings. Repeatable.
  --output <path>                       Write the report to a file instead of stdout.
  --to <YYYY-MM>                       Include records through this month.
  --total-budget <amount>             Warn when total spend exceeds amount.
  --top-services <count>              Limit service rows in the report.
  --service-budget <service=amount>   Warn when service spend exceeds amount. Repeatable.
  --growth-limit-percent <percent>    Warn when month-over-month growth exceeds percent.
`)
}

func printConfigUsage(w io.Writer) {
	fmt.Fprint(w, `Usage:
  marvin config sample [flags]
  marvin config validate <marvin.json>
  marvin config help
`)
}

func printConfigSampleUsage(w io.Writer) {
	fmt.Fprint(w, `Usage:
  marvin config sample [flags]

Flags:
  --output <path>    Write the sample config to a file instead of stdout.
`)
}

func printConfigValidateUsage(w io.Writer) {
	fmt.Fprint(w, `Usage:
  marvin config validate <marvin.json>
`)
}

func printSampleUsage(w io.Writer) {
	fmt.Fprint(w, `Usage:
  marvin sample [flags]

Flags:
  --output <path>    Write the sample CSV to a file instead of stdout.
`)
}

const sampleCostExplorerCSV = `Start Date,End Date,Service,Unblended Cost,Currency
2026-01-01,2026-01-31,Amazon Elastic Compute Cloud - Compute,$124.32,USD
2026-01-01,2026-01-31,Amazon Simple Storage Service,$18.90,USD
2026-02-01,2026-02-28,Amazon Elastic Compute Cloud - Compute,$143.81,USD
2026-02-01,2026-02-28,Amazon Simple Storage Service,$21.44,USD
2026-02-01,2026-02-28,AWS Key Management Service,$3.12,USD
`

const sampleConfigJSON = `{
  "total_budget": 300,
  "growth_limit_percent": 10,
  "from_month": "2026-01",
  "to_month": "2026-02",
  "top_services": 10,
  "service_budgets": {
    "Amazon Elastic Compute Cloud - Compute": 200
  },
  "ignore_services": [
    "Tax"
  ]
}
`
