package finance

import (
	"log"
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
}

// Definition of a security catalogue to house a portfolio of stock/ETF info in a map.
type SecurityCatalogue struct {
	yFinInterface    *YahooFinanceExtension
	sp500quotes      quote.Quote
	sheetMgr         *GoogleSheetManager
	dbClient         *data.MongoDbClient
	summary          *PortfolioSummary
	securities       map[string]*Security
	PortfolioHistory map[time.Time]float64 `json:"portfolioHistory"`
}

// Constructor for a new SecurityCatalogue object, initializing the map.
func NewSecurityCatalogue(sheetMgr *GoogleSheetManager, dbClient *data.MongoDbClient, pyScript string) *SecurityCatalogue {
	var sc SecurityCatalogue
	sc.yFinInterface = NewYahooFinanceExtension(pyScript)
	sc.securities = make(map[string]*Security)
	sc.PortfolioHistory = make(map[time.Time]float64)
	sc.summary = NewPortfolioSummary()
	sc.sheetMgr = sheetMgr
	sc.dbClient = dbClient
	return &sc
}

func (sc *SecurityCatalogue) GetPortfolioSummary() *PortfolioSummary {
	return sc.summary
}

func (sc *SecurityCatalogue) GetSecurityList() []*Security {
	return maps.Values(sc.securities)
}

func (sc *SecurityCatalogue) GetTransactionList() []Transaction {
	var txns []Transaction
	// Iterate through each security, appending the transactions to a slice as we go.
	for _, s := range sc.securities {
		txns = append(txns, s.transactions...)
	}
	return txns
}

func (sc *SecurityCatalogue) GetSp500() quote.Quote {
	return sc.sp500quotes
}

// Method to process the imported data, by creating a new [Transaction] for
// each row of data, and inserting it into the appropriate [Security] object
// within the catalogue.
func (sc *SecurityCatalogue) ProcessImport(txnData [][]interface{}) {
	// TODO: Would be good to perform column name verification before processing.
	// Iterate thru each row of data.
	for _, row := range txnData {
		// Make sure we haven't reached the end of the data.
		if row[0] != "" {
			// Create a new transaction with this row of data.
			txn := NewTransaction(row[0].(string), row[1].(string), row[2].(string), row[3].(string), row[4].(string))
			if txn != nil {
				// Check if we've seen the current ticker yet.
				if val, ok := sc.securities[txn.Ticker]; ok {
					// Yes, append the next transaction
					val.transactions = append(val.transactions, *txn)
				} else {
					// Create a new Security to track transactions for it, then append.
					if sec, err := NewSecurity(txn.Ticker, row[6].(string)); err == nil {
						sec.transactions = append(sec.transactions, *txn)
						sc.securities[txn.Ticker] = sec
					}
				}
			}
		}
	}
}

// Retrieves data from Yahoo for the given ticker, and stores the data in the DB.
func (sc *SecurityCatalogue) RetrieveAndStoreStockData(ticker string, startDate string, endDate string) {
	log.Printf("Querying %s data from Yahoo: %s ---> %s", ticker, startDate, endDate)
	newQuotes, err := quote.NewQuoteFromYahoo(ticker, startDate, endDate, quote.Daily, true)
	if err != nil {
		log.Printf("WARNING: Couldn't get ticker (%s) data from Yahoo: %s", ticker, err)
		return
	}
	sc.dbClient.StoreTickerData(newQuotes)
}

// Checks existing ticker history data in our DB, and pulls any missing data from Yahoo to fill in gaps.
func (sc *SecurityCatalogue) RefreshStockHistory(ticker string, initialDate string, currentlyOwned bool) {
	// Does the ticker exist in the DB?
	if sc.dbClient.TickerExists(ticker) {
		// Do we currently own this security?
		if currentlyOwned {
			// Are we up to date on the quotes? More than 3 days have passed?
			latestDate := sc.dbClient.GetLatestQuote(ticker)
			if time.Now().Sub(latestDate).Hours() > 72 {
				sc.RetrieveAndStoreStockData(ticker, latestDate.Add(24*time.Hour).UTC().Format("2006-01-02"), time.Now().UTC().Format("2006-01-02"))
			}
		}
	} else {
		// Ticker doesn't exist in DB yet, query all its data.
		sc.RetrieveAndStoreStockData(ticker, initialDate, time.Now().UTC().Format("2006-01-02"))
	}
}

// Add an individual's security history to the total portfolio value history.
func (sc *SecurityCatalogue) AccumulateValueHistory(stockHistory map[time.Time]float64) {
	// Iterate through each date for this security and add to the total.
	for date, dailyVal := range stockHistory {
		sc.PortfolioHistory[date] += dailyVal
	}
}

// Kicks off async functions in go-routines to calculate metrics for each security
func (sc *SecurityCatalogue) Calculate() {

	// Grab the historical S&P 500 data to compare against (2015 to present).
	sc.RefreshStockHistory("SPY", "2015-01-01", true)
	sc.sp500quotes = sc.dbClient.GetTickerData("SPY")
	// Preprocess all the securities, and add the stocks to a list to grab their historical info.
	for _, s := range sc.securities {
		if s.Ticker != "CASH" {
			s.PreProcess(sc.yFinInterface, sc.sheetMgr)
			// Make sure the stock's history data is up-to-date.
			sc.RefreshStockHistory(s.Ticker, s.transactions[0].DateTime.Format("2006-01-02"), s.CurrentlyHeld)
		}
	}

	// Setup a wait group.
	var waitGroup sync.WaitGroup
	// Specify the number of metrics to wait for.
	waitGroup.Add(len(sc.securities))

	log.Printf("Processing %d securities...\n", len(sc.securities))
	// Iterate thru each security in the map, and calculate its data.
	for _, s := range sc.securities {
		// Launch a new goroutine for this security.
		go func(s *Security) {
			// Pass SP500 quotes to this function to use when calculating transaction level metrics.
			s.CalculateMetrics(sc.dbClient.GetTickerData(s.Ticker), sc.sp500quotes)
			waitGroup.Done()
		}(s)
	}

	// Wait/monitor until all work is complete.
	waitGroup.Wait()

	// Calculate total invested market value across all securities.
	for _, s := range sc.securities {
		if s.Ticker != "CASH" {
			// s.DisplayMetrics()
			// Add this security's value history to the portfolio total.
			sc.AccumulateValueHistory(s.ValueHistory)
			sc.summary.TotalMarketValue += s.MarketValue
			sc.summary.TotalCostBasis += s.TotalCostBasis
			sc.summary.DailyGain += s.DailyGain
			sc.summary.TotalSecurities++
		}
	}
	// Store the last updated time, and percentage gain.
	sc.summary.LastUpdated = time.Now()
	sc.summary.PercentageGain = ((sc.summary.TotalMarketValue - sc.summary.TotalCostBasis) / sc.summary.TotalCostBasis) * 100.0
	log.Println("---------------------------------")
	log.Printf("Total Market Value: $%f", sc.summary.TotalMarketValue)
	log.Printf("Percentage Gain/Loss: %f", sc.summary.PercentageGain)
	log.Println("---------------------------------")
}
