package cost

import "time"

type Record struct {
	Service   string
	StartDate time.Time
	EndDate   time.Time
	Cost      float64
	Currency  string
}
