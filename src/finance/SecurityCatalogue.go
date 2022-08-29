package finance

import (
	"log"
	"sync"
	"time"
)

// Define a global map to store the stock split history relevant to our portfolio
var StockSplits = map[string][]Transaction{
	"NFLX": {
		Transaction{
			ticker:   "NFLX",
			dateTime: time.Date(2015, 7, 15, 0, 0, 0, 0, time.Local),
			action:   "Split",
			shares:   7},
	},
	"TSLA": {
		Transaction{
			ticker:   "TSLA",
			dateTime: time.Date(2020, 8, 31, 0, 0, 0, 0, time.Local),
			action:   "Split",
			shares:   5},
		Transaction{
			ticker:   "TSLA",
			dateTime: time.Date(2022, 8, 25, 0, 0, 0, 0, time.Local),
			action:   "Split",
			shares:   3},
	},
	"AMZN": {
		Transaction{
			ticker:   "AMZN",
			dateTime: time.Date(2022, 6, 6, 0, 0, 0, 0, time.Local),
			action:   "Split",
			shares:   20},
	},
	"GOOG": {
		Transaction{
			ticker:   "GOOG",
			dateTime: time.Date(2022, 7, 18, 0, 0, 0, 0, time.Local),
			action:   "Split",
			shares:   20},
	},
	"GOOGL": {
		Transaction{
			ticker:   "GOOGL",
			dateTime: time.Date(2022, 7, 18, 0, 0, 0, 0, time.Local),
			action:   "Split",
			shares:   20},
	},
}

// Definition of a security catalogue to house a portfolio of stock/ETF info in a map.
type SecurityCatalogue struct {
	id         uint
	securities map[string]*Security
}

// Constructor for a new SecurityCatalogue object, initializing the map.
func NewSecurityCatalogue() *SecurityCatalogue {
	var sc SecurityCatalogue
	sc.securities = make(map[string]*Security)
	return &sc
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
				if val, ok := sc.securities[txn.ticker]; ok {
					// Yes, append the next transaction
					val.transactions = append(val.transactions, *txn)
				} else {
					// Create a new Security to track transactions for it, then append.
					sec := NewSecurity(txn.ticker)
					sec.transactions = append(sec.transactions, *txn)
					sc.securities[txn.ticker] = sec
				}
			}
		}
	}
}

// Kicks off async functions in go-routines to calculate metrics for each security
func (sc *SecurityCatalogue) Calculate() {
	// Setup a wait group.
	var waitGroup sync.WaitGroup
	// Specify the number of metrics to wait for.
	waitGroup.Add(len(sc.securities))

	log.Printf("Processing %d securities...\n", len(sc.securities))
	// Iterate thru each security in the map, and calculate its data.
	for _, s := range sc.securities {

		// Launch a new goroutine for this security.
		go func(s *Security) {
			s.CalculateMetrics()
			waitGroup.Done()
		}(s)
	}

	// Wait/monitor until all work is complete.
	waitGroup.Wait()
}
