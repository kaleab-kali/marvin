package cost

import "strings"

func FilterIgnoredServices(records []Record, ignoredServices []string) []Record {
	ignored := make(map[string]bool)
	for _, service := range ignoredServices {
		service = strings.TrimSpace(service)
		if service != "" {
			ignored[service] = true
		}
	}
	if len(ignored) == 0 {
		return records
	}

	filtered := make([]Record, 0, len(records))
	for _, record := range records {
		if ignored[record.Service] {
			continue
		}
		filtered = append(filtered, record)
	}
	return filtered
}

func FilterIncludedServices(records []Record, includedServices []string) []Record {
	included := make(map[string]bool)
	for _, service := range includedServices {
		service = strings.TrimSpace(service)
		if service != "" {
			included[service] = true
		}
	}
	if len(included) == 0 {
		return records
	}

	filtered := make([]Record, 0, len(records))
	for _, record := range records {
		if !included[record.Service] {
			continue
		}
		filtered = append(filtered, record)
	}
	return filtered
}

func FilterCurrency(records []Record, currency string) []Record {
	currency = strings.ToUpper(strings.TrimSpace(currency))
	if currency == "" {
		return records
	}

	filtered := make([]Record, 0, len(records))
	for _, record := range records {
		if normalizeRecordCurrency(record.Currency) != currency {
			continue
		}
		filtered = append(filtered, record)
	}
	return filtered
}

func normalizeRecordCurrency(value string) string {
	value = strings.ToUpper(strings.TrimSpace(value))
	if value == "" {
		return "USD"
	}
	return value
}
