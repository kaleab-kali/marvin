package cost

import "time"

type MonthComparison struct {
	Month         time.Time
	PreviousMonth time.Time
	Cost          float64
	PreviousCost  float64
	Change        float64
	ChangePercent float64
}

func CompareMonths(months []MonthTotal) []MonthComparison {
	if len(months) < 2 {
		return nil
	}

	comparisons := make([]MonthComparison, 0, len(months)-1)
	for i := 1; i < len(months); i++ {
		current := months[i]
		previous := months[i-1]
		change := current.Cost - previous.Cost

		var changePercent float64
		if previous.Cost != 0 {
			changePercent = (change / previous.Cost) * 100
		}

		comparisons = append(comparisons, MonthComparison{
			Month:         current.Month,
			PreviousMonth: previous.Month,
			Cost:          current.Cost,
			PreviousCost:  previous.Cost,
			Change:        change,
			ChangePercent: changePercent,
		})
	}

	return comparisons
}
