package cost

import (
	"math"
	"testing"
)

func TestCompareMonths(t *testing.T) {
	months := []MonthTotal{
		{Month: mustDate(t, "2026-01-01"), Cost: 100},
		{Month: mustDate(t, "2026-02-01"), Cost: 125},
		{Month: mustDate(t, "2026-03-01"), Cost: 75},
	}

	got := CompareMonths(months)

	want := []MonthComparison{
		{
			Month:         mustDate(t, "2026-02-01"),
			PreviousMonth: mustDate(t, "2026-01-01"),
			Cost:          125,
			PreviousCost:  100,
			Change:        25,
			ChangePercent: 25,
		},
		{
			Month:         mustDate(t, "2026-03-01"),
			PreviousMonth: mustDate(t, "2026-02-01"),
			Cost:          75,
			PreviousCost:  125,
			Change:        -50,
			ChangePercent: -40,
		},
	}

	if len(got) != len(want) {
		t.Fatalf("expected %d comparisons, got %d", len(want), len(got))
	}
	for i := range want {
		if !got[i].Month.Equal(want[i].Month) {
			t.Fatalf("index %d: expected month %s, got %s", i, want[i].Month, got[i].Month)
		}
		if !got[i].PreviousMonth.Equal(want[i].PreviousMonth) {
			t.Fatalf("index %d: expected previous month %s, got %s", i, want[i].PreviousMonth, got[i].PreviousMonth)
		}
		assertFloat(t, "cost", got[i].Cost, want[i].Cost)
		assertFloat(t, "previous cost", got[i].PreviousCost, want[i].PreviousCost)
		assertFloat(t, "change", got[i].Change, want[i].Change)
		assertFloat(t, "change percent", got[i].ChangePercent, want[i].ChangePercent)
	}
}

func TestCompareMonthsReturnsNoComparisonsForSingleMonth(t *testing.T) {
	got := CompareMonths([]MonthTotal{
		{Month: mustDate(t, "2026-01-01"), Cost: 100},
	})

	if len(got) != 0 {
		t.Fatalf("expected no comparisons, got %d", len(got))
	}
}

func TestCompareMonthsAvoidsDivideByZero(t *testing.T) {
	got := CompareMonths([]MonthTotal{
		{Month: mustDate(t, "2026-01-01"), Cost: 0},
		{Month: mustDate(t, "2026-02-01"), Cost: 25},
	})

	if len(got) != 1 {
		t.Fatalf("expected 1 comparison, got %d", len(got))
	}
	assertFloat(t, "change", got[0].Change, 25)
	assertFloat(t, "change percent", got[0].ChangePercent, 0)
}

func assertFloat(t *testing.T, label string, got, want float64) {
	t.Helper()

	if math.Abs(got-want) > 0.000001 {
		t.Fatalf("expected %s %f, got %f", label, want, got)
	}
}
