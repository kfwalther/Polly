package finance

import (
	"time"
)

// Definition of a equity to hold the transactions for a particular stock/ETF.
type PortfolioSummary struct {
	// Includes current cash balance as well.
	TotalMarketValue  float64         `json:"totalMarketValue"`
	TotalEquities     float64         `json:"totalEquities"`
	TotalCostBasis    float64         `json:"totalCostBasis"`
	PercentageGain    float64         `json:"percentageGain"`
	DailyGain         float64         `json:"dailyGain"`
	LastUpdated       time.Time       `json:"lastUpdated"`
	MarketValueJan1   float64         `json:"marketValueJan1"`
	AnnualPerformance map[int]float64 `json:"annualPerformance"`
}

// Constructor for a new PortfolioSummary object.
func NewPortfolioSummary() *PortfolioSummary {
	var s PortfolioSummary
	return &s
}

// Calculate annual performance of the portfolio for each year.
func (ps *PortfolioSummary) CalculateHistoricalPerformance(portfolioHistory map[time.Time]float64, cashFlows map[int]float64) {
	ps.AnnualPerformance = make(map[int]float64)
	// Find the first datetime for this portfolio.
	beginDatetime, initialValue := ps.findEarliestTimeValue(portfolioHistory)
	if beginDatetime.IsZero() {
		return
	}
	nextValue := initialValue
	year := beginDatetime.Year()
	// Iterate through each completed year of portfolio history.
	for ; year < time.Now().Year(); year++ {
		// Find the earliest datetime of the following year.
		nextDatetime, nextYearValue := ps.findEarliestTimeValue(portfolioHistory, year+1)
		if nextDatetime.IsZero() {
			break
		}
		nextValue = nextYearValue
		// Save current year's performance.
		if denominator := initialValue + cashFlows[year]; denominator > 0.001 {
			ps.AnnualPerformance[year] = (nextValue/denominator - 1) * 100.0
		}
		initialValue = nextValue
	}
	ps.MarketValueJan1 = nextValue
	if denominator := nextValue + cashFlows[year]; denominator > 0.001 {
		ps.AnnualPerformance[year] = (ps.TotalMarketValue/denominator - 1) * 100.0
	}
}

// Helper function to locate the earliest time value in the time/value map.
func (ps *PortfolioSummary) findEarliestTimeValue(portfolioHistory map[time.Time]float64, targetYear ...int) (time.Time, float64) {
	var earliestTime time.Time
	var earliestValue float64
	firstIteration := true

	// Check if target year was provided.
	useTargetYear := len(targetYear) > 0

	for key, value := range portfolioHistory {
		if !useTargetYear || key.Year() == targetYear[0] {
			if firstIteration || key.Before(earliestTime) {
				earliestTime = key
				earliestValue = value
				firstIteration = false
			}
		}
	}
	return earliestTime, earliestValue
}
