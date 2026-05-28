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
