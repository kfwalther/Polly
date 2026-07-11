package finance

import (
	"testing"
	"time"

	"github.com/kfwalther/Polly/backend/data"
)

func testTransaction(action string, shares, price float64, date time.Time) Transaction {
	return Transaction{
		Ticker:   "ACME",
		DateTime: date,
		Action:   action,
		Shares:   shares,
		Price:    price,
		Value:    shares * price,
	}
}

func requireFloat(t *testing.T, got, want float64) {
	t.Helper()
	const tolerance = 0.000001
	if difference := got - want; difference > tolerance || difference < -tolerance {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestCalculateTransactionDataUsesFIFOAcrossLots(t *testing.T) {
	equity, err := NewEquity("ACME", "Stock")
	if err != nil {
		t.Fatal(err)
	}
	date := time.Date(2024, time.January, 2, 12, 0, 0, 0, time.UTC)
	equity.transactions = []Transaction{
		testTransaction("Buy", 10, 10, date),
		testTransaction("Buy", 5, 12, date.AddDate(0, 0, 1)),
		testTransaction("Sell", 12, 15, date.AddDate(0, 0, 2)),
	}

	shares := 0.0
	for idx := range equity.transactions {
		shares = equity.CalculateTransactionData(idx, shares)
	}

	requireFloat(t, shares, 3)
	requireFloat(t, equity.RealizedGain, 56)
	if len(equity.buyQ) != 1 {
		t.Fatalf("remaining buy lots = %d, want 1", len(equity.buyQ))
	}
	requireFloat(t, equity.buyQ[0].Shares, 3)
	requireFloat(t, equity.buyQ[0].Price, 12)
}

func TestCalculateTransactionDataAdjustsLotsForStockSplit(t *testing.T) {
	equity, err := NewEquity("ACME", "Stock")
	if err != nil {
		t.Fatal(err)
	}
	equity.splitMultiple = 1
	date := time.Date(2024, time.January, 2, 12, 0, 0, 0, time.UTC)
	equity.transactions = []Transaction{
		testTransaction("Buy", 10, 100, date),
		testTransaction("Split", 2, 0, date.AddDate(0, 0, 1)),
		testTransaction("Sell", 10, 60, date.AddDate(0, 0, 2)),
	}

	shares := 0.0
	for idx := range equity.transactions {
		shares = equity.CalculateTransactionData(idx, shares)
	}

	requireFloat(t, shares, 10)
	requireFloat(t, equity.RealizedGain, 100)
	requireFloat(t, equity.splitMultiple, 0.5)
	if len(equity.buyQ) != 1 {
		t.Fatalf("remaining buy lots = %d, want 1", len(equity.buyQ))
	}
	requireFloat(t, equity.buyQ[0].Shares, 10)
	requireFloat(t, equity.buyQ[0].Price, 50)
}

func TestCalculateTransactionDataAccountsForOversoldShares(t *testing.T) {
	equity, err := NewEquity("ACME", "Stock")
	if err != nil {
		t.Fatal(err)
	}
	equity.transactions = []Transaction{
		testTransaction("Sell", 3, 20, time.Date(2024, time.January, 2, 12, 0, 0, 0, time.UTC)),
	}

	shares := equity.CalculateTransactionData(0, 0)
	requireFloat(t, shares, 0)
	requireFloat(t, equity.RealizedGain, 60)
}

func TestCalculateMetricsBuildsValuesAndHoldingMetrics(t *testing.T) {
	equity, err := NewEquity("ACME", "Stock")
	if err != nil {
		t.Fatal(err)
	}
	firstDay := time.Date(2024, time.January, 2, 0, 0, 0, 0, time.UTC)
	equity.transactions = []Transaction{testTransaction("Buy", 10, 10, firstDay.Add(12*time.Hour))}
	equity.MarketPrice = 15
	equity.MarketPrevClosePrice = 10
	equity.CurrentlyHeld = true

	quotes := data.Quote{
		Date:  []time.Time{firstDay, firstDay.AddDate(0, 0, 1)},
		Close: []float64{10, 12},
	}
	today := getUtcDate(time.Now())
	sp500Quotes := data.Quote{
		Date:  []time.Time{firstDay, today},
		Close: []float64{100, 110},
	}

	equity.CalculateMetrics(quotes, sp500Quotes)

	requireFloat(t, equity.ValueHistory[firstDay.Unix()], 100)
	requireFloat(t, equity.ValueHistory[firstDay.AddDate(0, 0, 1).Unix()], 120)
	requireFloat(t, equity.ValueAllTimeHigh, 120)
	requireFloat(t, equity.NumShares, 10)
	requireFloat(t, equity.TotalCostBasis, 100)
	requireFloat(t, equity.UnitCostBasis, 10)
	requireFloat(t, equity.MarketValue, 150)
	requireFloat(t, equity.UnrealizedGain, 50)
	requireFloat(t, equity.TotalGain, 50)
	requireFloat(t, equity.DailyGain, 50)
	requireFloat(t, equity.DailyGainPercentage, 50)
}

func TestCalculateMetricsIsIdempotent(t *testing.T) {
	equity, err := NewEquity("ACME", "Stock")
	if err != nil {
		t.Fatal(err)
	}
	day := time.Date(2024, time.January, 2, 0, 0, 0, 0, time.UTC)
	equity.transactions = []Transaction{testTransaction("Buy", 10, 10, day.Add(12*time.Hour))}
	equity.MarketPrice = 15
	equity.CurrentlyHeld = true
	quotes := data.Quote{Date: []time.Time{day}, Close: []float64{10}}
	today := getUtcDate(time.Now())
	sp500Quotes := data.Quote{Date: []time.Time{day, today}, Close: []float64{100, 110}}

	equity.CalculateMetrics(quotes, sp500Quotes)
	equity.CalculateMetrics(quotes, sp500Quotes)

	requireFloat(t, equity.NumShares, 10)
	requireFloat(t, equity.TotalCostBasis, 100)
	requireFloat(t, equity.RealizedGain, 0)
	requireFloat(t, equity.MarketValue, 150)
	requireFloat(t, equity.TotalGain, 50)
}

func TestCalculateMetricsIsIdempotentWithStockSplit(t *testing.T) {
	equity, err := NewEquity("ACME", "Stock")
	if err != nil {
		t.Fatal(err)
	}
	firstDay := time.Date(2024, time.January, 2, 0, 0, 0, 0, time.UTC)
	equity.transactions = []Transaction{
		testTransaction("Buy", 10, 100, firstDay.Add(12*time.Hour)),
		testTransaction("Split", 2, 0, firstDay.AddDate(0, 0, 1).Add(12*time.Hour)),
	}
	equity.MarketPrice = 60
	equity.MarketPrevClosePrice = 50
	equity.CurrentlyHeld = true
	quotes := data.Quote{
		Date:  []time.Time{firstDay, firstDay.AddDate(0, 0, 1)},
		Close: []float64{50, 55},
	}
	today := getUtcDate(time.Now())
	sp500Quotes := data.Quote{Date: []time.Time{firstDay, today}, Close: []float64{100, 110}}

	equity.CalculateMetrics(quotes, sp500Quotes)
	equity.CalculateMetrics(quotes, sp500Quotes)

	requireFloat(t, equity.ValueHistory[firstDay.Unix()], 1000)
	requireFloat(t, equity.ValueHistory[firstDay.AddDate(0, 0, 1).Unix()], 1100)
	requireFloat(t, equity.NumShares, 20)
	requireFloat(t, equity.TotalCostBasis, 1000)
	requireFloat(t, equity.MarketValue, 1200)
	requireFloat(t, equity.UnrealizedGain, 200)
}

func TestProcessFinancialHistoryDataCalculatesTTMAndGrowth(t *testing.T) {
	equity, err := NewEquity("ACME", "Stock")
	if err != nil {
		t.Fatal(err)
	}
	equity.revenueUnits = "M"
	equity.processFinancialHistoryData([][]interface{}{
		{"Units", "M", "", "", "", ""},
		{"Total Revenue", "1", "2", "3", "4", "5"},
		{"Gross Margins (%)", "10", "20", "30", "40", "50"},
		{"S&M / Revenue (%)", "1", "2", "3", "4", "5"},
		{"FCF", "1", "2", "3", "4", "5"},
		{"FCF / Revenue (%)", "1", "2", "3", "4", "5"},
		{"Quarter", "20230101", "20230401", "20230701", "20231001", "20240101"},
		{"SBC / Revenue (%)", "1", "2", "3", "4", "5"},
	})

	if len(equity.quarterlyDates) != 5 {
		t.Fatalf("quarterly dates = %d, want 5", len(equity.quarterlyDates))
	}
	requireFloat(t, equity.quarterlyRevenue[0], 1000)
	requireFloat(t, equity.RevenueTtm, 14000)
	requireFloat(t, equity.RevenueGrowthPercentageYoy, 4)
	requireFloat(t, equity.GrossMargin, 0.5)
	requireFloat(t, equity.fcfTtm, 14000)
}

func TestProcessFinancialHistoryDataHandlesFourQuarters(t *testing.T) {
	equity, err := NewEquity("ACME", "Stock")
	if err != nil {
		t.Fatal(err)
	}
	equity.processFinancialHistoryData([][]interface{}{
		{"Units", "B", "", "", ""},
		{"Total Revenue", "1", "2", "3", "4"},
		{"Gross Margins (%)", "10", "20", "30", "40"},
		{"S&M / Revenue (%)", "1", "2", "3", "4"},
		{"FCF", "1", "2", "3", "4"},
		{"FCF / Revenue (%)", "1", "2", "3", "4"},
		{"Quarter", "20230401", "20230701", "20231001", "20240101"},
		{"SBC / Revenue (%)", "1", "2", "3", "4"},
	})

	if len(equity.quarterlyDates) != 4 {
		t.Fatalf("quarterly dates = %d, want 4", len(equity.quarterlyDates))
	}
	requireFloat(t, equity.RevenueTtm, 10)
	requireFloat(t, equity.fcfTtm, 10)
	requireFloat(t, equity.RevenueGrowthPercentageYoy, 0)
	requireFloat(t, equity.GrossMargin, 0.4)
}
