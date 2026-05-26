package cost

import (
	"math"
	"testing"
	"time"
)

func TestTotalSpend(t *testing.T) {
	records := []Record{
		{Service: "Amazon EC2", Cost: 10.25},
		{Service: "Amazon S3", Cost: 5.75},
	}

	got := TotalSpend(records)
	if math.Abs(got-16.0) > 0.000001 {
		t.Fatalf("expected total 16.0, got %f", got)
	}
}

func TestGroupByServiceSortsByCostDescending(t *testing.T) {
	records := []Record{
		{Service: "Amazon S3", Cost: 10},
		{Service: "Amazon EC2", Cost: 25},
		{Service: "Amazon S3", Cost: 15},
		{Service: "AWS KMS", Cost: 2},
	}

	got := GroupByService(records)

	want := []ServiceTotal{
		{Service: "Amazon EC2", Cost: 25},
		{Service: "Amazon S3", Cost: 25},
		{Service: "AWS KMS", Cost: 2},
	}
	if len(got) != len(want) {
		t.Fatalf("expected %d service totals, got %d", len(want), len(got))
	}
	for i := range want {
		if got[i].Service != want[i].Service {
			t.Fatalf("index %d: expected service %q, got %q", i, want[i].Service, got[i].Service)
		}
		if math.Abs(got[i].Cost-want[i].Cost) > 0.000001 {
			t.Fatalf("index %d: expected cost %f, got %f", i, want[i].Cost, got[i].Cost)
		}
	}
}

func TestGroupByMonthSortsChronologically(t *testing.T) {
	records := []Record{
		{StartDate: mustDate(t, "2026-02-17"), Cost: 8},
		{StartDate: mustDate(t, "2026-01-01"), Cost: 10},
		{StartDate: mustDate(t, "2026-02-01"), Cost: 4},
	}

	got := GroupByMonth(records)

	want := []MonthTotal{
		{Month: mustDate(t, "2026-01-01"), Cost: 10},
		{Month: mustDate(t, "2026-02-01"), Cost: 12},
	}
	if len(got) != len(want) {
		t.Fatalf("expected %d month totals, got %d", len(want), len(got))
	}
	for i := range want {
		if !got[i].Month.Equal(want[i].Month) {
			t.Fatalf("index %d: expected month %s, got %s", i, want[i].Month.Format(time.DateOnly), got[i].Month.Format(time.DateOnly))
		}
		if math.Abs(got[i].Cost-want[i].Cost) > 0.000001 {
			t.Fatalf("index %d: expected cost %f, got %f", i, want[i].Cost, got[i].Cost)
		}
	}
}

func TestMonthNormalizesDateToFirstOfMonthUTC(t *testing.T) {
	date := time.Date(2026, time.May, 26, 14, 30, 0, 0, time.FixedZone("EAT", 3*60*60))

	got := Month(date)
	want := time.Date(2026, time.May, 1, 0, 0, 0, 0, time.UTC)

	if !got.Equal(want) {
		t.Fatalf("expected %s, got %s", want.Format(time.RFC3339), got.Format(time.RFC3339))
	}
}

func mustDate(t *testing.T, value string) time.Time {
	t.Helper()

	parsed, err := time.Parse(time.DateOnly, value)
	if err != nil {
		t.Fatalf("invalid test date %q: %v", value, err)
	}
	return parsed
}
