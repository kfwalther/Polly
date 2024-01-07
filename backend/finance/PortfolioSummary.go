package finance

import "time"

// Definition of a equity to hold the transactions for a particular stock/ETF.
type PortfolioSummary struct {
	// Includes current cash balance as well.
	TotalMarketValue         float64   `json:"totalMarketValue"`
	TotalEquities            float64   `json:"totalEquities"`
	TotalCostBasis           float64   `json:"totalCostBasis"`
	PercentageGain           float64   `json:"percentageGain"`
	DailyGain                float64   `json:"dailyGain"`
	YearToDatePercentageGain float64   `json:"yearToDatePercentageGain"`
	LastUpdated              time.Time `json:"lastUpdated"`
	MarketValueJan1          float64   `json:"marketValueJan1"`
}

// Constructor for a new PortfolioSummary object.
func NewPortfolioSummary() *PortfolioSummary {
	var s PortfolioSummary
	return &s
}
