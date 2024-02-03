package finance

import (
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/kfwalther/Polly/backend/data"
	"github.com/markcheno/go-quote"

	"golang.org/x/exp/maps"
)

// Define a global map to store the stock split history relevant to our portfolio
var StockSplits = map[string][]Transaction{
	"TSLA": {
		Transaction{
			Ticker:   "TSLA",
			DateTime: time.Date(2020, 8, 31, 0, 0, 0, 0, time.Local),
			Action:   "Split",
			Shares:   5},
		Transaction{
			Ticker:   "TSLA",
			DateTime: time.Date(2022, 8, 25, 0, 0, 0, 0, time.Local),
			Action:   "Split",
			Shares:   3},
	},
	"NVDA": {
		Transaction{
			Ticker:   "NVDA",
			DateTime: time.Date(2021, 7, 20, 0, 0, 0, 0, time.Local),
			Action:   "Split",
			Shares:   4},
	},
	"SHOP": {
		Transaction{
			Ticker:   "SHOP",
			DateTime: time.Date(2022, 06, 29, 0, 0, 0, 0, time.Local),
			Action:   "Split",
			Shares:   10},
	},
	"AMZN": {
		Transaction{
			Ticker:   "AMZN",
			DateTime: time.Date(2022, 6, 6, 0, 0, 0, 0, time.Local),
			Action:   "Split",
			Shares:   20},
	},
	"GOOG": {
		Transaction{
			Ticker:   "GOOG",
			DateTime: time.Date(2022, 7, 18, 0, 0, 0, 0, time.Local),
			Action:   "Split",
			Shares:   20},
	},
	"GOOGL": {
		Transaction{
			Ticker:   "GOOGL",
			DateTime: time.Date(2022, 7, 18, 0, 0, 0, 0, time.Local),
			Action:   "Split",
			Shares:   20},
	},
	"IAU": {
		Transaction{
			Ticker:   "IAU",
			DateTime: time.Date(2021, 5, 24, 0, 0, 0, 0, time.Local),
			Action:   "Split",
			Shares:   0.5},
	},
	"CELH": {
		Transaction{
			Ticker:   "CELH",
			DateTime: time.Date(2023, 11, 15, 0, 0, 0, 0, time.Local),
			Action:   "Split",
			Shares:   3},
	},
}

// Define a list of delisted stocks no longer on exchanges so we don't query Yahoo for these.
var DelistedTickers = map[string][]string{
	// Went bankrupt in 2022
	"VYGVF": {"Voyager Digital Ltd."},
	// Delisted from Nasdaq in July 2023 to OTC as APPH.Q
	"APPH": {"AppHarvest Inc."},
	// Delisted from Nasdaq in August 2023, filed for bankrupcy
	"PTRA": {"Proterra Inc."},
	// Bought by AMD
	"XLNX": {"Xilinx Inc."},
	// Delisted mutual fund in 2022
	"PONDX": {"PIMCO Income Fund Class D"},
	// This represents cash transactions in our tracker, ignore as a ticker.
	"CASH": {"Placeholder symbol"},
}

// Definition of a equity catalogue to house a portfolio of stock/ETF info in a map.
type EquityCatalogue struct {
	yFinInterface    *YahooFinanceExtension
	sp500quotes      quote.Quote
	sheetMgr         *GoogleSheetManager
	dbClient         *data.MongoDbClient
	portfolioSummary *PortfolioSummary
	equityType       string
	equities         map[string]*Equity
	transactions     []Transaction
	CashFlowByYear   map[int]float64
	PortfolioHistory map[time.Time]float64 `json:"portfolioHistory"`
}

// Constructor for a new EquityCatalogue object, initializing the map.
func NewEquityCatalogue(equityType string, sheetMgr *GoogleSheetManager, dbClient *data.MongoDbClient, pyScript string) *EquityCatalogue {
	var ec EquityCatalogue
	ec.equityType = equityType
	// Initialize the interfaces.
	ec.yFinInterface = NewYahooFinanceExtension(pyScript)
	ec.sheetMgr = sheetMgr
	ec.dbClient = dbClient
	// Initialize the data structures for this class.
	ec.equities = make(map[string]*Equity)
	ec.transactions = make([]Transaction, 0)
	ec.PortfolioHistory = make(map[time.Time]float64)
	ec.portfolioSummary = NewPortfolioSummary()
	return &ec
}

// Return the full portfolio summary followed by the stock-only portfolio summary.
func (ec *EquityCatalogue) GetPortfolioSummary() *PortfolioSummary {
	return ec.portfolioSummary
}

func (ec *EquityCatalogue) GetEquityList() []*Equity {
	return maps.Values(ec.equities)
}

// Get the transaction lists from each equity, which have derived metrics calculated on each.
func (ec *EquityCatalogue) GetTransactionList() []Transaction {
	var txns []Transaction
	// Iterate through each equity, appending the transactions to a slice as we go.
	for _, s := range ec.equities {
		txns = append(txns, s.transactions...)
	}
	return txns
}

func (ec *EquityCatalogue) GetSp500() quote.Quote {
	return ec.sp500quotes
}

// Re-run the Google Sheets retrieval and portfolio calculations again to refresh all data.
func (ec *EquityCatalogue) Refresh() {
	// Re-init the data structures for this class.
	ec.equities = make(map[string]*Equity)
	ec.transactions = make([]Transaction, 0)
	ec.PortfolioHistory = make(map[time.Time]float64)
	ec.portfolioSummary = NewPortfolioSummary()
}

// Method to process the imported data, by creating a new [Transaction] for
// each row of data, and inserting it into the appropriate [Equity] object
// within the catalogue.
func (ec *EquityCatalogue) ProcessImport(txnData [][]interface{}) {
	// Iterate thru each row of data.
	for _, row := range txnData {
		// Make sure we haven't reached the end of the data.
		if row[0] != "" {
			// Create a new transaction with this row of data.
			txn := NewTransaction(row[0].(string), row[1].(string), row[2].(string), row[3].(string), row[4].(string))
			if txn != nil {
				// Add it to the total txns list.
				ec.transactions = append(ec.transactions, *txn)
				// Check if we've seen the current ticker yet.
				if val, ok := ec.equities[txn.Ticker]; ok {
					// Yes, append the next transaction
					val.transactions = append(val.transactions, *txn)
				} else {
					// Create a new Equity to track transactions for it, then append.
					if sec, err := NewEquity(txn.Ticker, row[5].(string)); err == nil {
						sec.transactions = append(sec.transactions, *txn)
						ec.equities[txn.Ticker] = sec
					}
				}
			}
		}
	}
}

// Retrieves data from Yahoo for the given ticker, and stores the data in the DB.
func (ec *EquityCatalogue) RetrieveAndStoreStockData(ticker string, startDate string, endDate string) {
	queryTicker := ticker
	if ec.equityType == "crypto" {
		queryTicker = queryTicker + "-USD"
	}
	log.Printf("Querying %s data from Yahoo: %s ---> %s", queryTicker, startDate, endDate)
	newQuotes, err := quote.NewQuoteFromYahoo(queryTicker, startDate, endDate, quote.Daily, true)
	if err != nil {
		log.Printf("WARNING: Couldn't get ticker (%s) data from Yahoo: %s", queryTicker, err)
		return
	}
	newQuotes.Symbol = strings.TrimSuffix(newQuotes.Symbol, "-USD")
	ec.dbClient.StoreTickerData(newQuotes)
}

// Checks existing ticker history data in our DB, and pulls any missing data from Yahoo to fill in gaps. Careful modifying this method...
func (ec *EquityCatalogue) RefreshStockHistory(txns *[]Transaction, currentlyOwned bool) {
	// Does the ticker exist in the DB?
	ticker := (*txns)[0].Ticker
	if ec.dbClient.TickerExists(ticker) {
		latestDate := ec.dbClient.GetLatestQuote(ticker)
		// Do we currently own this equity?
		if currentlyOwned {
			// Are we up to date on the quotes? More than 3 days have passed?
			if time.Now().Sub(latestDate).Hours() > 72 {
				ec.RetrieveAndStoreStockData(ticker, latestDate.Add(24*time.Hour).UTC().Format("2006-01-02"), time.Now().UTC().Format("2006-01-02"))
			}
		} else {
			// If we don't own it, ensure we have all the data through the last sell date (get one day past sell date to be safe).
			sellDate := (*txns)[len(*txns)-1].DateTime
			if sellDate.Sub(latestDate).Hours() > 24 {
				ec.RetrieveAndStoreStockData(ticker, latestDate.Add(24*time.Hour).UTC().Format("2006-01-02"), sellDate.Add(24*time.Hour).UTC().Format("2006-01-02"))
			}
		}
	} else {
		// Ticker doesn't exist in DB yet, query all its data.
		ec.RetrieveAndStoreStockData(ticker, (*txns)[0].DateTime.Format("2006-01-02"), time.Now().UTC().Format("2006-01-02"))
	}
}

// Add an individual's equity history to the total portfolio value history.
func (ec *EquityCatalogue) AccumulateValueHistory(stockHistory map[int64]float64) {
	// Iterate through each date for this equity and add to the total.
	for date, dailyVal := range stockHistory {
		if dailyVal > 0.0 {
			ec.PortfolioHistory[time.Unix(date, 0).UTC()] += dailyVal
		}
	}
}

// Iterate through every txn to calculate the cash balance history in the portfolio.
func (ec *EquityCatalogue) CalculateCashBalanceHistory() {
	curCashAmount := 0.0
	ec.CashFlowByYear = make(map[int]float64)
	// Order the transactions by date, using anonymous function.
	sort.Slice(ec.transactions, func(i, j int) bool {
		return ec.transactions[i].DateTime.Before(ec.transactions[j].DateTime)
	})
	for _, txn := range ec.transactions {
		if txn.Action == "Deposit" || txn.Action == "Sell" {
			curCashAmount += txn.Value
		} else if txn.Action == "Withdraw" || txn.Action == "Buy" {
			curCashAmount -= txn.Value
		}
		ec.equities["CASH"].ValueHistory[txn.DateTime.Unix()] = curCashAmount
		// Keep track of portfolio cash flow by year.
		if txn.Action == "Deposit" {
			ec.CashFlowByYear[txn.DateTime.Year()] += txn.Value
		} else if txn.Action == "Withdraw" {
			ec.CashFlowByYear[txn.DateTime.Year()] -= txn.Value
		}
	}
	ec.equities["CASH"].MarketValue = curCashAmount
}

// Iterate through each equity to calculate summary metrics across the entire portfolio.
func (ec *EquityCatalogue) CalculatePortfolioSummaryMetrics() {
	for _, s := range ec.equities {
		if s.Ticker != "CASH" {
			// s.DisplayMetrics()
			// Add this equity's value history to the portfolio total.
			ec.AccumulateValueHistory(s.ValueHistory)
			ec.portfolioSummary.TotalMarketValue += s.MarketValue
			ec.portfolioSummary.TotalCostBasis += s.TotalCostBasis
			ec.portfolioSummary.DailyGain += s.DailyGain
			if s.CurrentlyHeld {
				ec.portfolioSummary.TotalEquities++
			}
		} else {
			// For cash, just track the current balance in the summary.
			ec.portfolioSummary.TotalMarketValue += s.MarketValue
		}
	}

	ec.portfolioSummary.CalculateHistoricalPerformance(ec.PortfolioHistory, ec.CashFlowByYear)
	// Store the last updated time, and percentage gain.
	ec.portfolioSummary.LastUpdated = time.Now()
	if ec.portfolioSummary.TotalCostBasis > 0.001 {
		ec.portfolioSummary.PercentageGain = ((ec.portfolioSummary.TotalMarketValue - ec.portfolioSummary.TotalCostBasis) / ec.portfolioSummary.TotalCostBasis) * 100.0
	}
}

// Kicks off async functions in go-routines to calculate metrics for each equity
func (ec *EquityCatalogue) Calculate() {

	// Grab the historical S&P 500 data to compare against (2015 to present).
	// Define a dummy SPY transaction to pass in, the date is what the function requires.
	ec.RefreshStockHistory(&[]Transaction{*NewTransaction("2015-01-01", "SPY", "Buy", "1", "100.0")}, true)
	ec.sp500quotes = ec.dbClient.GetTickerData("SPY")

	// Get the tickers for all equities we've ever owned in comma-separated list.
	tickers := make([]string, 0, len(ec.equities))
	for t := range ec.equities {
		// Don't include any delisted equities.
		if _, ok := DelistedTickers[t]; !ok {
			if ec.equityType == "crypto" {
				tickers = append(tickers, t+"-USD")
			} else {
				tickers = append(tickers, t)
			}
		}
	}
	tickerList := strings.Join(tickers, ",")
	log.Printf("Querying Yahoo finance for %d equities...\n", len(tickers))
	// Use Python yFinance module to query data for all tickers at once.
	allStocksData := ec.yFinInterface.GetTickerData(tickerList)

	// Setup a wait group.
	var waitGroup sync.WaitGroup
	// Specify the number of metrics to wait for.
	waitGroup.Add(len(ec.equities))

	log.Printf("Processing %d equities...\n", len(ec.equities))
	// Iterate thru each equity in the map, and calculate its data.
	for _, s := range ec.equities {
		// Launch a new goroutine for this equity.
		go func(s *Equity) {
			if _, ok := DelistedTickers[s.Ticker]; !ok {
				s.PreProcess(ec.sheetMgr, allStocksData)
				// Make sure the stock's history data is up-to-date.
				ec.RefreshStockHistory(&s.transactions, s.CurrentlyHeld)
			}
			// Pass SP500 quotes to this function to use when calculating transaction level metrics.
			s.CalculateMetrics(ec.dbClient.GetTickerData(s.Ticker), ec.sp500quotes)
			waitGroup.Done()
		}(s)
	}

	// Wait/monitor until all work is complete.
	waitGroup.Wait()

	// Calculate the cash balance history in our portfolio.
	ec.CalculateCashBalanceHistory()

	// Calculate total invested market value and other summary metrics across all equities.
	ec.CalculatePortfolioSummaryMetrics()

	log.Println("---------------------------------")
	log.Printf("Total Market Value: $%f", ec.portfolioSummary.TotalMarketValue)
	log.Printf("Percentage Gain/Loss: %f%%", ec.portfolioSummary.PercentageGain)
	log.Printf("Cash Flow YTD: %v", ec.CashFlowByYear)
	log.Println("---------------------------------")
}
