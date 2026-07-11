package finance

import (
	"testing"
	"time"
)

func TestProcessImportGroupsTransactionsByTicker(t *testing.T) {
	catalogue := NewEquityCatalogue("stock", nil, nil, "")
	catalogue.ProcessImport([][]interface{}{
		{"1/2/2024", "ACME", "Buy", "10", "12.50", "Stock"},
		{"1/3/2024", "ACME", "Sell", "2", "15", "Stock"},
		{"1/4/2024", "CASH", "Deposit", "100", "", "Cash"},
		{"", "", "", "", "", ""},
	})

	if len(catalogue.transactions) != 3 {
		t.Fatalf("transactions = %d, want 3", len(catalogue.transactions))
	}
	if len(catalogue.equities) != 2 {
		t.Fatalf("equities = %d, want 2", len(catalogue.equities))
	}
	if len(catalogue.equities["ACME"].transactions) != 2 {
		t.Fatalf("ACME transactions = %d, want 2", len(catalogue.equities["ACME"].transactions))
	}
	if len(catalogue.equities["CASH"].transactions) != 1 {
		t.Fatalf("CASH transactions = %d, want 1", len(catalogue.equities["CASH"].transactions))
	}
}

func TestCalculateCashBalanceHistoryOrdersAndClassifiesTransactions(t *testing.T) {
	catalogue := NewEquityCatalogue("stock", nil, nil, "")
	catalogue.ProcessImport([][]interface{}{
		{"1/3/2024", "CASH", "Deposit", "1000", "", "Cash"},
		{"1/1/2024", "ACME", "Buy", "10", "10", "Stock"},
		{"1/2/2024", "ACME", "Sell", "10", "16", "Stock"},
		{"1/4/2024", "CASH", "Withdraw", "200", "", "Cash"},
	})

	catalogue.CalculateCashBalanceHistory()
	cashHistory := catalogue.equities["CASH"].ValueHistory
	utcNoon := func(day int) int64 {
		return time.Date(2024, time.January, day, 12, 0, 0, 0, time.UTC).Unix()
	}

	requireFloat(t, cashHistory[utcNoon(1)], -100)
	requireFloat(t, cashHistory[utcNoon(2)], 60)
	requireFloat(t, cashHistory[utcNoon(3)], 1060)
	requireFloat(t, cashHistory[utcNoon(4)], 860)
	requireFloat(t, catalogue.equities["CASH"].MarketValue, 860)
	requireFloat(t, catalogue.CashFlowByYear[2024], 800)
}

func TestAccumulateValueHistoryCombinesPositiveDailyValues(t *testing.T) {
	catalogue := NewEquityCatalogue("stock", nil, nil, "")
	day := time.Date(2024, time.January, 2, 0, 0, 0, 0, time.UTC)
	catalogue.PortfolioHistory[day] = 10

	catalogue.AccumulateValueHistory(map[int64]float64{
		day.Unix():                  20,
		day.AddDate(0, 0, 1).Unix(): 30,
		day.AddDate(0, 0, 2).Unix(): 0,
	})

	requireFloat(t, catalogue.PortfolioHistory[day], 30)
	requireFloat(t, catalogue.PortfolioHistory[day.AddDate(0, 0, 1)], 30)
	if _, exists := catalogue.PortfolioHistory[day.AddDate(0, 0, 2)]; exists {
		t.Fatal("zero history entry should not be added")
	}
}

func TestCalculatePortfolioSummaryMetricsAggregatesEquitiesAndCash(t *testing.T) {
	catalogue := NewEquityCatalogue("stock", nil, nil, "")
	currentYear := time.Now().Year()
	historyDay := time.Date(currentYear, time.January, 2, 0, 0, 0, 0, time.UTC)
	stock, err := NewEquity("ACME", "Stock")
	if err != nil {
		t.Fatal(err)
	}
	stock.ValueHistory[historyDay.Unix()] = 90
	stock.MarketValue = 100
	stock.TotalCostBasis = 80
	stock.DailyGain = 2
	stock.CurrentlyHeld = true
	cash, err := NewEquity("CASH", "Cash")
	if err != nil {
		t.Fatal(err)
	}
	cash.MarketValue = 20
	catalogue.equities = map[string]*Equity{"ACME": stock, "CASH": cash}
	catalogue.CashFlowByYear = map[int]float64{currentYear: 10}

	catalogue.CalculatePortfolioSummaryMetrics()
	summary := catalogue.GetPortfolioSummary()

	requireFloat(t, summary.TotalMarketValue, 120)
	requireFloat(t, summary.TotalCostBasis, 80)
	requireFloat(t, summary.DailyGain, 2)
	requireFloat(t, summary.TotalEquities, 1)
	requireFloat(t, summary.PercentageGain, 50)
	requireFloat(t, catalogue.PortfolioHistory[historyDay], 90)
	requireFloat(t, summary.MarketValueJan1, 90)
	requireFloat(t, summary.AnnualPerformance[currentYear], 20)
	if summary.LastUpdated.IsZero() {
		t.Fatal("LastUpdated should be set")
	}
}

func TestCalculateCashBalanceHistoryCreatesMissingCashEquity(t *testing.T) {
	catalogue := NewEquityCatalogue("stock", nil, nil, "")
	stock, err := NewEquity("ACME", "Stock")
	if err != nil {
		t.Fatal(err)
	}
	catalogue.equities["ACME"] = stock
	catalogue.transactions = []Transaction{
		testTransaction("Buy", 2, 10, time.Date(2024, time.January, 2, 12, 0, 0, 0, time.UTC)),
	}

	catalogue.CalculateCashBalanceHistory()

	cash, exists := catalogue.equities["CASH"]
	if !exists {
		t.Fatal("CASH equity should be created when absent")
	}
	requireFloat(t, cash.MarketValue, -20)
}
