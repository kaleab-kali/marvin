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
}

func LoadWarningRules(r io.Reader) (cost.WarningRules, error) {
	decoder := json.NewDecoder(r)
	decoder.DisallowUnknownFields()

	var file warningRulesFile
	if err := decoder.Decode(&file); err != nil {
		return cost.WarningRules{}, fmt.Errorf("decode config: %w", err)
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		return cost.WarningRules{}, errors.New("decode config: multiple JSON documents")
	}

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

func validatePositive(name string, value float64) error {
	if value < 0 {
		return fmt.Errorf("%s must not be negative", name)
	}
	return nil
}
