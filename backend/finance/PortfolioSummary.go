package finance

import "time"

// Definition of a security to hold the transactions for a particular stock/ETF.
type PortfolioSummary struct {
	// Includes current cash balance as well.
	TotalMarketValue         float64   `json:"totalMarketValue"`
	TotalSecurities          float64   `json:"totalSecurities"`
	TotalCostBasis           float64   `json:"totalCostBasis"`
	PercentageGain           float64   `json:"percentageGain"`
	DailyGain                float64   `json:"dailyGain"`
	YearToDatePercentageGain float64   `json:"yearToDatePercentageGain"`
	LastUpdated              time.Time `json:"lastUpdated"`
}

// Constructor for a new PortfolioSummary object.
func NewPortfolioSummary() *PortfolioSummary {
	var s PortfolioSummary
	return &s
}
