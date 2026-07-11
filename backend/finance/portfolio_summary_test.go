package finance

import (
	"testing"
	"time"
)

func TestFindEarliestTimeValue(t *testing.T) {
	summary := NewPortfolioSummary()
	history := map[time.Time]float64{
		time.Date(2024, time.March, 1, 0, 0, 0, 0, time.UTC):     130,
		time.Date(2023, time.December, 31, 0, 0, 0, 0, time.UTC): 100,
		time.Date(2024, time.January, 2, 0, 0, 0, 0, time.UTC):   120,
	}

	gotTime, gotValue := summary.findEarliestTimeValue(history, 2024)
	wantTime := time.Date(2024, time.January, 2, 0, 0, 0, 0, time.UTC)
	if !gotTime.Equal(wantTime) {
		t.Fatalf("earliest 2024 time = %v, want %v", gotTime, wantTime)
	}
	requireFloat(t, gotValue, 120)
}

func TestCalculateHistoricalPerformanceAdjustsForCashFlows(t *testing.T) {
	currentYear := time.Now().Year()
	previousYear := currentYear - 1
	summary := NewPortfolioSummary()
	summary.TotalMarketValue = 143
	history := map[time.Time]float64{
		time.Date(previousYear, time.January, 2, 0, 0, 0, 0, time.UTC): 100,
		time.Date(currentYear, time.January, 2, 0, 0, 0, 0, time.UTC):  120,
	}

	summary.CalculateHistoricalPerformance(history, map[int]float64{currentYear: 10})

	requireFloat(t, summary.AnnualPerformance[previousYear], 20)
	requireFloat(t, summary.MarketValueJan1, 120)
	requireFloat(t, summary.AnnualPerformance[currentYear], 10)
}

func TestCalculateHistoricalPerformanceHandlesEmptyHistory(t *testing.T) {
	summary := NewPortfolioSummary()
	summary.TotalMarketValue = 100

	summary.CalculateHistoricalPerformance(map[time.Time]float64{}, nil)

	if len(summary.AnnualPerformance) != 0 {
		t.Fatalf("AnnualPerformance = %v, want no entries", summary.AnnualPerformance)
	}
	requireFloat(t, summary.MarketValueJan1, 0)
}
