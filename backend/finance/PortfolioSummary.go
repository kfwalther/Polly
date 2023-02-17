package finance

// Definition of a security to hold the transactions for a particular stock/ETF.
type PortfolioSummary struct {
	TotalMarketValue float64 `json:"totalMarketValue"`
	TotalSecurities  float64 `json:"totalSecurities"`
	TotalCostBasis   float64 `json:"totalCostBasis"`
	PercentageGain   float64 `json:"percentageGain"`
	DailyGain        float64 `json:"dailyGain"`
}

// Constructor for a new PortfolioSummary object.
func NewPortfolioSummary() *PortfolioSummary {
	var s PortfolioSummary
	return &s
}
