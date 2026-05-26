package cost

import (
	"sort"
	"time"
)

type ServiceTotal struct {
	Service string
	Cost    float64
}

type MonthTotal struct {
	Month time.Time
	Cost  float64
}

func TotalSpend(records []Record) float64 {
	var total float64
	for _, record := range records {
		total += record.Cost
	}
	return total
}

func GroupByService(records []Record) []ServiceTotal {
	totals := make(map[string]float64)
	for _, record := range records {
		totals[record.Service] += record.Cost
	}

	services := make([]ServiceTotal, 0, len(totals))
	for service, total := range totals {
		services = append(services, ServiceTotal{
			Service: service,
			Cost:    total,
		})
	}

	sort.Slice(services, func(i, j int) bool {
		if services[i].Cost == services[j].Cost {
			return services[i].Service < services[j].Service
		}
		return services[i].Cost > services[j].Cost
	})

	return services
}

func GroupByMonth(records []Record) []MonthTotal {
	totals := make(map[time.Time]float64)
	for _, record := range records {
		month := Month(record.StartDate)
		totals[month] += record.Cost
	}

	months := make([]MonthTotal, 0, len(totals))
	for month, total := range totals {
		months = append(months, MonthTotal{
			Month: month,
			Cost:  total,
		})
	}

	sort.Slice(months, func(i, j int) bool {
		return months[i].Month.Before(months[j].Month)
	})

	return months
}

func Month(date time.Time) time.Time {
	if date.IsZero() {
		return time.Time{}
	}
	return time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, time.UTC)
}
