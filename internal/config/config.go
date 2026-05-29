package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/kaleab-kali/marvin/internal/cost"
)

type warningRulesFile struct {
	Schema             string             `json:"$schema"`
	FailOnWarning      *bool              `json:"fail_on_warning"`
	TotalBudget        float64            `json:"total_budget"`
	Format             string             `json:"format"`
	GrowthLimitPercent float64            `json:"growth_limit_percent"`
	ServiceBudgets     map[string]float64 `json:"service_budgets"`
	IgnoreServices     []string           `json:"ignore_services"`
	IncludeServices    []string           `json:"include_services"`
	FromMonth          string             `json:"from_month"`
	MinServiceSpend    float64            `json:"min_service_spend"`
	ToMonth            string             `json:"to_month"`
	TopServices        int                `json:"top_services"`
}

type Settings struct {
	FailOnWarning   *bool
	Rules           cost.WarningRules
	Format          string
	FromMonth       time.Time
	IgnoreServices  []string
	IncludeServices []string
	MinServiceSpend float64
	ToMonth         time.Time
	TopServices     int
}

func Load(r io.Reader) (Settings, error) {
	file, err := loadFile(r)
	if err != nil {
		return Settings{}, err
	}

	rules, err := rulesFromFile(file)
	if err != nil {
		return Settings{}, err
	}
	ignored, err := serviceNamesFromFile("ignore_services", file.IgnoreServices)
	if err != nil {
		return Settings{}, err
	}
	included, err := serviceNamesFromFile("include_services", file.IncludeServices)
	if err != nil {
		return Settings{}, err
	}

	if err := validateOptionalNonNegativeInt("top_services", file.TopServices); err != nil {
		return Settings{}, err
	}
	if err := validatePositive("min_service_spend", file.MinServiceSpend); err != nil {
		return Settings{}, err
	}
	format, err := reportFormatFromFile(file.Format)
	if err != nil {
		return Settings{}, err
	}
	failOnWarning := optionalBool(file.FailOnWarning)
	fromMonth, err := parseOptionalMonth("from_month", file.FromMonth)
	if err != nil {
		return Settings{}, err
	}
	toMonth, err := parseOptionalMonth("to_month", file.ToMonth)
	if err != nil {
		return Settings{}, err
	}
	if !fromMonth.IsZero() && !toMonth.IsZero() && fromMonth.After(toMonth) {
		return Settings{}, errors.New("from_month must be before or equal to to_month")
	}

	return Settings{
		FailOnWarning:   failOnWarning,
		Rules:           rules,
		Format:          format,
		FromMonth:       fromMonth,
		IgnoreServices:  ignored,
		IncludeServices: included,
		MinServiceSpend: file.MinServiceSpend,
		ToMonth:         toMonth,
		TopServices:     file.TopServices,
	}, nil
}

func optionalBool(value *bool) *bool {
	if value == nil {
		return nil
	}
	copied := *value
	return &copied
}

func LoadWarningRules(r io.Reader) (cost.WarningRules, error) {
	settings, err := Load(r)
	if err != nil {
		return cost.WarningRules{}, err
	}
	return settings.Rules, nil
}

func loadFile(r io.Reader) (warningRulesFile, error) {
	decoder := json.NewDecoder(r)
	decoder.DisallowUnknownFields()

	var file warningRulesFile
	if err := decoder.Decode(&file); err != nil {
		return warningRulesFile{}, fmt.Errorf("decode config: %w", err)
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		return warningRulesFile{}, errors.New("decode config: multiple JSON documents")
	}
	return file, nil
}

func rulesFromFile(file warningRulesFile) (cost.WarningRules, error) {
	if err := validatePositive("total_budget", file.TotalBudget); err != nil {
		return cost.WarningRules{}, err
	}
	if err := validatePositive("growth_limit_percent", file.GrowthLimitPercent); err != nil {
		return cost.WarningRules{}, err
	}

	rules := cost.WarningRules{
		TotalLimit:         file.TotalBudget,
		GrowthLimitPercent: file.GrowthLimitPercent,
		ServiceLimits:      make(map[string]float64),
	}
	for service, limit := range file.ServiceBudgets {
		service = strings.TrimSpace(service)
		if service == "" {
			return cost.WarningRules{}, errors.New("service_budgets contains an empty service name")
		}
		if err := validatePositive("service_budgets."+service, limit); err != nil {
			return cost.WarningRules{}, err
		}
		if limit > 0 {
			rules.ServiceLimits[service] = limit
		}
	}

	return rules, nil
}

func serviceNamesFromFile(field string, values []string) ([]string, error) {
	services := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			return nil, fmt.Errorf("%s contains an empty service name", field)
		}
		services = append(services, value)
	}
	return services, nil
}

func reportFormatFromFile(value string) (string, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "":
		return "", nil
	case "terminal", "text":
		return "terminal", nil
	case "markdown", "md":
		return "markdown", nil
	case "json", "csv":
		return value, nil
	default:
		return "", fmt.Errorf("unsupported format value %q, expected terminal, markdown, md, json, csv, or text", value)
	}
}

func validatePositive(name string, value float64) error {
	if value < 0 {
		return fmt.Errorf("%s must not be negative", name)
	}
	return nil
}

func validateOptionalNonNegativeInt(name string, value int) error {
	if value < 0 {
		return fmt.Errorf("%s must not be negative", name)
	}
	return nil
}

func parseOptionalMonth(name, value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, nil
	}
	parsed, err := time.Parse("2006-01", value)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid %s value %q, expected YYYY-MM", name, value)
	}
	return time.Date(parsed.Year(), parsed.Month(), 1, 0, 0, 0, 0, time.UTC), nil
}
