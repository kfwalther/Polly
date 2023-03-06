package finance

import (
	"log"
	"sync"
	"time"

	"github.com/markcheno/go-quote"

	"golang.org/x/exp/maps"
)

// Define a global map to store the stock split history relevant to our portfolio
var StockSplits = map[string][]Transaction{
	"NFLX": {
		Transaction{
			Ticker:   "NFLX",
			DateTime: time.Date(2015, 7, 15, 0, 0, 0, 0, time.Local),
			Action:   "Split",
			Shares:   7},
	},
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
	id          uint
	sp500quotes quote.Quote
	summary     *PortfolioSummary
	securities  map[string]*Security
}

// Constructor for a new SecurityCatalogue object, initializing the map.
func NewSecurityCatalogue() *SecurityCatalogue {
	var sc SecurityCatalogue
	sc.securities = make(map[string]*Security)
	sc.summary = NewPortfolioSummary()
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

// Kicks off async functions in go-routines to calculate metrics for each security
func (sc *SecurityCatalogue) Calculate() {
	// Grab the historical S&P 500 data to compare against (2015 to present).
	sc.sp500quotes, _ = quote.NewQuoteFromYahoo("spy", "2015-01-01", time.Now().Format("2006-01-02"), quote.Daily, true)

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
			s.CalculateMetrics(sc.sp500quotes)
			waitGroup.Done()
		}(s)
	}

	// Wait/monitor until all work is complete.
	waitGroup.Wait()

	// Calculate total invested market value across all securities.
	for _, s := range sc.securities {
		// s.DisplayMetrics()
		sc.summary.TotalMarketValue += s.MarketValue
		sc.summary.TotalCostBasis += s.TotalCostBasis
		sc.summary.DailyGain += s.DailyGain
		sc.summary.TotalSecurities++
	}
	sc.summary.PercentageGain = ((sc.summary.TotalMarketValue - sc.summary.TotalCostBasis) / sc.summary.TotalCostBasis) * 100.0
	log.Println("---------------------------------")
	log.Printf("Total Market Value: $%f", sc.summary.TotalMarketValue)
	log.Printf("Percentage Gain/Loss: %f", sc.summary.PercentageGain)
	log.Println("---------------------------------")
}
