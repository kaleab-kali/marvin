package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/kaleab-kali/marvin/internal/cost"
)

type warningRulesFile struct {
	TotalBudget        float64            `json:"total_budget"`
	GrowthLimitPercent float64            `json:"growth_limit_percent"`
	ServiceBudgets     map[string]float64 `json:"service_budgets"`
	IgnoreServices     []string           `json:"ignore_services"`
	TopServices        int                `json:"top_services"`
}

type Settings struct {
	Rules          cost.WarningRules
	IgnoreServices []string
	TopServices    int
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
	ignored, err := ignoredServicesFromFile(file.IgnoreServices)
	if err != nil {
		return Settings{}, err
	}

	if err := validateOptionalNonNegativeInt("top_services", file.TopServices); err != nil {
		return Settings{}, err
	}

	return Settings{Rules: rules, IgnoreServices: ignored, TopServices: file.TopServices}, nil
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

func ignoredServicesFromFile(values []string) ([]string, error) {
	ignored := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			return nil, errors.New("ignore_services contains an empty service name")
		}
		ignored = append(ignored, value)
	}
	return ignored, nil
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
