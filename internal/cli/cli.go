package cli

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
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
var errValidateHelp = errors.New("validate help requested")

type analyzeOptions struct {
	paths            []string
	currency         string
	format           string
	outputPath       string
	failOnWarning    bool
	fromMonth        time.Time
	ignoredServices  []string
	includedServices []string
	minServiceSpend  float64
	serviceSort      string
	toMonth          time.Time
	topServices      int
	rules            cost.WarningRules
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
		return runConfigCommand(args[1:], stdin, stdout, stderr)
	case "sample":
		return runSample(args[1:], stdout, stderr)
	case "validate":
		return runValidate(args[1:], stdin, stdout, stderr)
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
	records = cost.FilterCurrency(records, options.currency)
	records = cost.FilterIncludedServices(records, options.includedServices)
	records = cost.FilterIgnoredServices(records, options.ignoredServices)
	records = filterRecordsByMonth(records, options.fromMonth, options.toMonth)
	if len(records) == 0 {
		fmt.Fprintln(stderr, "analyze produced no records after applying filters")
		return ExitRuntimeError
	}
	if err := validateSingleCurrency(records); err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return ExitRuntimeError
	}
	summary := report.BuildSummary(records, options.rules)
	summary = filterServiceSpend(summary, options.minServiceSpend)
	summary = limitServiceSpend(summary, options.topServices)
	summary = sortServiceSpend(summary, options.serviceSort)

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

func runValidate(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	paths, err := parseValidateArgs(args)
	if errors.Is(err, errValidateHelp) {
		printValidateUsage(stdout)
		return ExitOK
	}
	if err != nil {
		fmt.Fprintln(stderr, err)
		return ExitUsageError
	}

	records, err := readCostRecords(paths, stdin)
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return ExitRuntimeError
	}

	fmt.Fprintf(stdout, "validated %d cost records from %d input(s)\n", len(records), len(paths))
	return ExitOK
}

func readCostRecords(paths []string, stdin io.Reader) ([]cost.Record, error) {
	var records []cost.Record
	for _, path := range paths {
		input := stdin
		var file io.Closer
		if path != "-" {
			var err error
			file, input, err = openCostCSV(path)
			if err != nil {
				return nil, fmt.Errorf("open cost CSV %q: %w", path, err)
			}
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

func openCostCSV(path string) (io.Closer, io.Reader, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}

	if !strings.HasSuffix(strings.ToLower(path), ".gz") {
		return file, file, nil
	}

	gz, err := gzip.NewReader(file)
	if err != nil {
		file.Close()
		return nil, nil, err
	}
	return joinedCloser{closers: []io.Closer{gz, file}}, gz, nil
}

type joinedCloser struct {
	closers []io.Closer
}

func (closer joinedCloser) Close() error {
	var closeErr error
	for _, item := range closer.closers {
		if err := item.Close(); closeErr == nil && err != nil {
			closeErr = err
		}
	}
	return closeErr
}

func runConfigCommand(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
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
		return runConfigValidate(args[1:], stdin, stdout, stderr)
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

func runConfigValidate(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	path, err := parseConfigValidateArgs(args)
	if errors.Is(err, errConfigHelp) {
		printConfigValidateUsage(stdout)
		return ExitOK
	}
	if err != nil {
		fmt.Fprintln(stderr, err)
		return ExitUsageError
	}

	input := stdin
	if path != "-" {
		file, err := os.Open(path)
		if err != nil {
			fmt.Fprintf(stderr, "open config: %v\n", err)
			return ExitRuntimeError
		}
		defer file.Close()
		input = file
	}

	if _, err := config.Load(input); err != nil {
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
		format:      "terminal",
		rules:       cost.WarningRules{ServiceLimits: make(map[string]float64)},
		serviceSort: "cost",
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
				return analyzeOptions{}, errors.New("--format requires terminal, markdown, json, or csv")
			}
			if err := setReportFormat(&options, value); err != nil {
				return analyzeOptions{}, err
			}
		case strings.HasPrefix(arg, "--format="):
			if err := setReportFormat(&options, strings.TrimPrefix(arg, "--format=")); err != nil {
				return analyzeOptions{}, err
			}
		case arg == "--currency":
			value, ok := nextArg(args, &i)
			if !ok {
				return analyzeOptions{}, errors.New("--currency requires a three-letter currency code")
			}
			currency, err := parseCurrencyFlag("--currency", value)
			if err != nil {
				return analyzeOptions{}, err
			}
			options.currency = currency
		case strings.HasPrefix(arg, "--currency="):
			currency, err := parseCurrencyFlag("--currency", strings.TrimPrefix(arg, "--currency="))
			if err != nil {
				return analyzeOptions{}, err
			}
			options.currency = currency
		case arg == "--min-service-spend":
			value, ok := nextArg(args, &i)
			if !ok {
				return analyzeOptions{}, errors.New("--min-service-spend requires an amount")
			}
			amount, err := parsePositiveFloat("--min-service-spend", value)
			if err != nil {
				return analyzeOptions{}, err
			}
			options.minServiceSpend = amount
		case strings.HasPrefix(arg, "--min-service-spend="):
			amount, err := parsePositiveFloat("--min-service-spend", strings.TrimPrefix(arg, "--min-service-spend="))
			if err != nil {
				return analyzeOptions{}, err
			}
			options.minServiceSpend = amount
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
		case arg == "--sort-services":
			value, ok := nextArg(args, &i)
			if !ok {
				return analyzeOptions{}, errors.New("--sort-services requires cost or name")
			}
			if err := setServiceSort(&options, value); err != nil {
				return analyzeOptions{}, err
			}
		case strings.HasPrefix(arg, "--sort-services="):
			if err := setServiceSort(&options, strings.TrimPrefix(arg, "--sort-services=")); err != nil {
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
		case arg == "--only-service":
			value, ok := nextArg(args, &i)
			if !ok {
				return analyzeOptions{}, errors.New("--only-service requires a service name")
			}
			if err := addIncludedService(&options, value); err != nil {
				return analyzeOptions{}, err
			}
		case strings.HasPrefix(arg, "--only-service="):
			if err := addIncludedService(&options, strings.TrimPrefix(arg, "--only-service=")); err != nil {
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
		case arg == "-":
			if path != "" {
				return "", fmt.Errorf("unexpected extra argument %q", arg)
			}
			path = arg
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

func parseValidateArgs(args []string) ([]string, error) {
	var paths []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-h" || arg == "--help":
			return nil, errValidateHelp
		case arg == "-":
			if containsString(paths, "-") {
				return nil, errors.New("validate accepts standard input only once")
			}
			paths = append(paths, arg)
		case strings.HasPrefix(arg, "-"):
			return nil, fmt.Errorf("unknown validate flag %q", arg)
		default:
			paths = append(paths, arg)
		}
	}
	if len(paths) == 0 {
		return nil, errors.New("validate requires a Cost Explorer CSV path")
	}
	return paths, nil
}

func writeReport(stdout io.Writer, summary report.Summary, options analyzeOptions) error {
	switch options.format {
	case "terminal":
		return report.WriteTerminalSummary(stdout, summary)
	case "markdown":
		return report.WriteMarkdownSummary(stdout, summary)
	case "json":
		return report.WriteJSONSummary(stdout, summary)
	case "csv":
		return report.WriteCSVSummary(stdout, summary)
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

func setServiceSort(options *analyzeOptions, value string) error {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "cost", "name":
		options.serviceSort = value
		return nil
	default:
		return fmt.Errorf("unsupported --sort-services %q, expected cost or name", value)
	}
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

func filterServiceSpend(summary report.Summary, minServiceSpend float64) report.Summary {
	if minServiceSpend <= 0 {
		return summary
	}

	services := make([]report.ServiceSpend, 0, len(summary.ServiceSpend))
	for _, service := range summary.ServiceSpend {
		if service.Cost >= minServiceSpend {
			services = append(services, service)
		}
	}
	summary.ServiceSpend = services
	return summary
}

func sortServiceSpend(summary report.Summary, serviceSort string) report.Summary {
	switch serviceSort {
	case "", "cost":
		sort.Slice(summary.ServiceSpend, func(i, j int) bool {
			if summary.ServiceSpend[i].Cost == summary.ServiceSpend[j].Cost {
				return summary.ServiceSpend[i].Service < summary.ServiceSpend[j].Service
			}
			return summary.ServiceSpend[i].Cost > summary.ServiceSpend[j].Cost
		})
	case "name":
		sort.Slice(summary.ServiceSpend, func(i, j int) bool {
			return summary.ServiceSpend[i].Service < summary.ServiceSpend[j].Service
		})
	}
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

func validateSingleCurrency(records []cost.Record) error {
	currencies := make(map[string]bool)
	for _, record := range records {
		currencies[normalizeCurrency(record.Currency)] = true
	}
	if len(currencies) <= 1 {
		return nil
	}

	values := make([]string, 0, len(currencies))
	for currency := range currencies {
		values = append(values, currency)
	}
	sort.Strings(values)
	return fmt.Errorf("multiple currencies found after filters: %s", strings.Join(values, ", "))
}

func normalizeCurrency(value string) string {
	value = strings.ToUpper(strings.TrimSpace(value))
	if value == "" {
		return "USD"
	}
	return value
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
	if settings.Currency != "" {
		options.currency = settings.Currency
	}
	if settings.Format != "" {
		options.format = settings.Format
	}
	if settings.FailOnWarning != nil {
		options.failOnWarning = *settings.FailOnWarning
	}
	options.fromMonth = settings.FromMonth
	options.ignoredServices = settings.IgnoreServices
	options.includedServices = settings.IncludeServices
	options.minServiceSpend = settings.MinServiceSpend
	if settings.OutputPath != nil {
		options.outputPath = *settings.OutputPath
	}
	if settings.SortServices != "" {
		options.serviceSort = settings.SortServices
	}
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

func addIncludedService(options *analyzeOptions, value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return errors.New("--only-service requires a service name")
	}
	options.includedServices = append(options.includedServices, value)
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
	case "csv":
		options.format = "csv"
		return nil
	default:
		return fmt.Errorf("unsupported --format %q, expected terminal, markdown, md, json, csv, or text", value)
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

func parseCurrencyFlag(name, value string) (string, error) {
	value = strings.ToUpper(strings.TrimSpace(value))
	if len(value) != 3 {
		return "", fmt.Errorf("invalid %s value %q, expected a three-letter currency code", name, value)
	}
	for _, char := range value {
		if char < 'A' || char > 'Z' {
			return "", fmt.Errorf("invalid %s value %q, expected a three-letter currency code", name, value)
		}
	}
	return value, nil
}

func printUsage(w io.Writer) {
	fmt.Fprint(w, `Marvin reads exported AWS Cost Explorer CSV files and reports cost changes.

Usage:
  marvin analyze [flags] <cost-explorer.csv|-> [more.csv ...]
  marvin config sample [flags]
  marvin config validate <marvin.json|->
  marvin sample [flags]
  marvin validate <cost-explorer.csv|-> [more.csv ...]
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
  --currency <code>                     Include only records for this currency code.
  --fail-on-warning                     Exit with code 3 when warnings are present.
  --format <terminal|markdown|json|csv> Output format. Defaults to terminal.
  --from <YYYY-MM>                     Include records from this month onward.
  --ignore-service <service>           Exclude a service from totals and warnings. Repeatable.
  --min-service-spend <amount>         Hide service rows below this spend amount.
  --only-service <service>             Include only this service. Repeatable.
  --output <path>                       Write the report to a file instead of stdout.
  --sort-services <cost|name>          Sort service rows. Defaults to cost.
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
  marvin config validate <marvin.json|->
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
  marvin config validate <marvin.json|->
`)
}

func printSampleUsage(w io.Writer) {
	fmt.Fprint(w, `Usage:
  marvin sample [flags]

Flags:
  --output <path>    Write the sample CSV to a file instead of stdout.
`)
}

func printValidateUsage(w io.Writer) {
	fmt.Fprint(w, `Usage:
  marvin validate <cost-explorer.csv|-> [more.csv ...]
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
  "$schema": "https://raw.githubusercontent.com/kaleab-kali/marvin/main/docs/marvin.schema.json",
  "currency": "USD",
  "total_budget": 300,
  "format": "terminal",
  "growth_limit_percent": 10,
  "from_month": "2026-01",
  "min_service_spend": 10,
  "sort_services": "cost",
  "to_month": "2026-02",
  "top_services": 10,
  "service_budgets": {
    "Amazon Elastic Compute Cloud - Compute": 200
  },
  "include_services": [
    "Amazon Elastic Compute Cloud - Compute",
    "Amazon Simple Storage Service"
  ],
  "ignore_services": [
    "Tax"
  ]
}
`
