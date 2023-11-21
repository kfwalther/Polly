package finance

import (
	"log"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/kfwalther/Polly/backend/data"
	"github.com/markcheno/go-quote"

	"github.com/gorilla/websocket"
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

// Definition of a security catalogue to house a portfolio of stock/ETF info in a map.
type SecurityCatalogue struct {
	yFinInterface     *YahooFinanceExtension
	sp500quotes       quote.Quote
	sheetMgr          *GoogleSheetManager
	dbClient          *data.MongoDbClient
	fullSummary       *PortfolioSummary
	stockSummary      *PortfolioSummary
	progressWebSocket *websocket.Conn
	securities        map[string]*Security
	transactions      []Transaction
	cashFlowYtd       float64
	PortfolioHistory  map[time.Time]float64 `json:"portfolioHistory"`
}

// Constructor for a new SecurityCatalogue object, initializing the map.
func NewSecurityCatalogue(sheetMgr *GoogleSheetManager, dbClient *data.MongoDbClient, pyScript string) *SecurityCatalogue {
	var sc SecurityCatalogue
	// Initialize the interfaces.
	sc.yFinInterface = NewYahooFinanceExtension(pyScript)
	sc.sheetMgr = sheetMgr
	sc.dbClient = dbClient
	// Initialize the data structures for this class.
	sc.securities = make(map[string]*Security)
	sc.transactions = make([]Transaction, 0)
	sc.PortfolioHistory = make(map[time.Time]float64)
	sc.fullSummary = NewPortfolioSummary()
	sc.stockSummary = NewPortfolioSummary()
	sc.progressWebSocket = nil
	return &sc
}

// Return the full portfolio summary followed by the stock-only portfolio summary.
func (sc *SecurityCatalogue) GetPortfolioSummary() []PortfolioSummary {
	summaries := []PortfolioSummary{*sc.fullSummary, *sc.stockSummary}
	return summaries
}

func (sc *SecurityCatalogue) GetSecurityList() []*Security {
	return maps.Values(sc.securities)
}

// Get the transaction lists from each security, which have derived metrics calculated on each.
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

// Send progress updates to the client via the web socket, if it's initialized.
func (sc *SecurityCatalogue) SendProgressUpdate(progressPercent float64) {
	if sc.progressWebSocket != nil {
		percentStr := strconv.FormatFloat(progressPercent, 'E', -1, 64)
		// Write the web socket message to the client (1% progress).
		err := sc.progressWebSocket.WriteMessage(websocket.TextMessage, []byte(percentStr))
		if err != nil {
			log.Println("WARNING: Web socket write error: ", err)
		}
	}
}

// Re-run the Google Sheets retrieval and portfolio calculations again to refresh all data.
func (sc *SecurityCatalogue) Refresh(progressSocket *websocket.Conn) int {
	// Re-init the data structures for this class.
	sc.securities = make(map[string]*Security)
	sc.transactions = make([]Transaction, 0)
	sc.PortfolioHistory = make(map[time.Time]float64)
	sc.fullSummary = NewPortfolioSummary()
	sc.stockSummary = NewPortfolioSummary()
	sc.progressWebSocket = progressSocket
	// Re-read from transactions sheet.
	txns := sc.sheetMgr.GetTransactionData()
	sc.SendProgressUpdate(3.0)
	// Re-process the imported data to organize it by ticker.
	sc.ProcessImport(txns.Values)
	log.Printf("Number of transactions processed: %d", len(txns.Values))
	// Calculate metrics for each stock.
	sc.Calculate()
	sc.SendProgressUpdate(100.0)
	return len(txns.Values)
}

// Method to process the imported data, by creating a new [Transaction] for
// each row of data, and inserting it into the appropriate [Security] object
// within the catalogue.
func (sc *SecurityCatalogue) ProcessImport(txnData [][]interface{}) {
	// Iterate thru each row of data.
	for _, row := range txnData {
		// Make sure we haven't reached the end of the data.
		if row[0] != "" {
			// Create a new transaction with this row of data.
			txn := NewTransaction(row[0].(string), row[1].(string), row[2].(string), row[3].(string), row[4].(string))
			if txn != nil {
				// Add it to the total txns list.
				sc.transactions = append(sc.transactions, *txn)
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
	sc.SendProgressUpdate(5.0)
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

// Checks existing ticker history data in our DB, and pulls any missing data from Yahoo to fill in gaps. Careful modifying this method...
func (sc *SecurityCatalogue) RefreshStockHistory(txns *[]Transaction, currentlyOwned bool) {
	// Does the ticker exist in the DB?
	ticker := (*txns)[0].Ticker
	if sc.dbClient.TickerExists(ticker) {
		latestDate := sc.dbClient.GetLatestQuote(ticker)
		// Do we currently own this security?
		if currentlyOwned {
			// Are we up to date on the quotes? More than 3 days have passed?
			if time.Now().Sub(latestDate).Hours() > 72 {
				sc.RetrieveAndStoreStockData(ticker, latestDate.Add(24*time.Hour).UTC().Format("2006-01-02"), time.Now().UTC().Format("2006-01-02"))
			}
		} else {
			// If we don't own it, ensure we have all the data through the last sell date (get one day past sell date to be safe).
			sellDate := (*txns)[len(*txns)-1].DateTime
			if sellDate.Sub(latestDate).Hours() > 24 {
				sc.RetrieveAndStoreStockData(ticker, latestDate.Add(24*time.Hour).UTC().Format("2006-01-02"), sellDate.Add(24*time.Hour).UTC().Format("2006-01-02"))
			}
		}
	} else {
		// Ticker doesn't exist in DB yet, query all its data.
		sc.RetrieveAndStoreStockData(ticker, (*txns)[0].DateTime.Format("2006-01-02"), time.Now().UTC().Format("2006-01-02"))
	}
}

// Add an individual's security history to the total portfolio value history.
func (sc *SecurityCatalogue) AccumulateValueHistory(stockHistory map[int64]float64) {
	// Iterate through each date for this security and add to the total.
	for date, dailyVal := range stockHistory {
		if dailyVal > 0.0 {
			sc.PortfolioHistory[time.Unix(date, 0).UTC()] += dailyVal
		}
	}
}

// Iterate through every txn to calculate the cash balance history in the portfolio.
func (sc *SecurityCatalogue) CalculateCashBalanceHistory(firstTradeDay time.Time) {
	curCashAmount := 0.0
	sc.cashFlowYtd = 0.0
	// Order the transactions by date, using anonymous function.
	sort.Slice(sc.transactions, func(i, j int) bool {
		return sc.transactions[i].DateTime.Before(sc.transactions[j].DateTime)
	})
	for _, txn := range sc.transactions {
		if txn.Action == "Deposit" || txn.Action == "Sell" {
			curCashAmount += txn.Value
		} else if txn.Action == "Withdraw" || txn.Action == "Buy" {
			curCashAmount -= txn.Value
		}
		sc.securities["CASH"].ValueHistory[txn.DateTime.Unix()] = curCashAmount
		// Keep track of portfolio cash flow YTD.
		if txn.DateTime.After(firstTradeDay) {
			if txn.Action == "Deposit" {
				sc.cashFlowYtd += txn.Value
			} else if txn.Action == "Withdraw" {
				sc.cashFlowYtd -= txn.Value
			}
		}
	}
	sc.securities["CASH"].MarketValue = curCashAmount
}

// Iterate through each equity to calculate summary metrics across the entire portfolio.
func (sc *SecurityCatalogue) CalculatePortfolioSummaryMetrics(firstTradeDay time.Time) {
	fullPortfolioMarketValueJan1 := 0.0
	stockPortfolioMarketValueJan1 := 0.0
	for _, s := range sc.securities {
		if s.Ticker != "CASH" {
			// s.DisplayMetrics()
			// Add this security's value history to the portfolio total.
			sc.AccumulateValueHistory(s.ValueHistory)
			sc.fullSummary.TotalMarketValue += s.MarketValue
			sc.fullSummary.TotalCostBasis += s.TotalCostBasis
			sc.fullSummary.DailyGain += s.DailyGain
			if s.CurrentlyHeld {
				sc.fullSummary.TotalSecurities++
			}
			if s.SecurityType == "Stock" {
				sc.stockSummary.TotalMarketValue += s.MarketValue
				sc.stockSummary.TotalCostBasis += s.TotalCostBasis
				sc.stockSummary.DailyGain += s.DailyGain
				if s.CurrentlyHeld {
					sc.stockSummary.TotalSecurities++
				}
			}
		} else {
			// For cash, just track the current balance in the summary.
			sc.fullSummary.TotalMarketValue += s.MarketValue
			sc.stockSummary.TotalMarketValue += s.MarketValue
		}
		fullPortfolioMarketValueJan1 += s.GetMarketValueOnDate(firstTradeDay)
		if s.SecurityType == "Cash" || s.SecurityType == "Stock" {
			stockPortfolioMarketValueJan1 += s.GetMarketValueOnDate(firstTradeDay)
		}
	}
	// Store the last updated time, and percentage gain.
	sc.fullSummary.LastUpdated = time.Now()
	sc.stockSummary.LastUpdated = time.Now()
	sc.fullSummary.PercentageGain = ((sc.fullSummary.TotalMarketValue - sc.fullSummary.TotalCostBasis) / sc.fullSummary.TotalCostBasis) * 100.0
	sc.stockSummary.PercentageGain = ((sc.stockSummary.TotalMarketValue - sc.stockSummary.TotalCostBasis) / sc.stockSummary.TotalCostBasis) * 100.0
	sc.fullSummary.YearToDatePercentageGain = (sc.fullSummary.TotalMarketValue/(fullPortfolioMarketValueJan1+sc.cashFlowYtd) - 1) * 100.0
	sc.stockSummary.YearToDatePercentageGain = (sc.stockSummary.TotalMarketValue/(stockPortfolioMarketValueJan1+sc.cashFlowYtd) - 1) * 100.0
}

// Kicks off async functions in go-routines to calculate metrics for each security
func (sc *SecurityCatalogue) Calculate() {

	// Grab the historical S&P 500 data to compare against (2015 to present).
	// Define a dummy SPY transaction to pass in, the date is what the function requires.
	sc.RefreshStockHistory(&[]Transaction{*NewTransaction("2015-01-01", "SPY", "Buy", "1", "100.0")}, true)
	sc.sp500quotes = sc.dbClient.GetTickerData("SPY")
	sc.SendProgressUpdate(8.0)

	// Get the tickers for all securities we've ever owned in comma-separated list.
	tickers := make([]string, 0, len(sc.securities))
	for t := range sc.securities {
		// Don't include any delisted securities.
		if _, ok := DelistedTickers[t]; !ok {
			tickers = append(tickers, t)
		}
	}
	tickerList := strings.Join(tickers, ",")
	log.Printf("Querying Yahoo finance for %d equities...\n", len(tickers))
	// Use Python yFinance module to query data for all tickers at once.
	allStocksData := sc.yFinInterface.GetTickerData(tickerList)
	sc.SendProgressUpdate(35.0)

	// Setup a wait group.
	var waitGroup sync.WaitGroup
	// Specify the number of metrics to wait for.
	waitGroup.Add(len(sc.securities))

	log.Printf("Processing %d securities...\n", len(sc.securities))
	// Iterate thru each security in the map, and calculate its data.
	for _, s := range sc.securities {
		// Launch a new goroutine for this security.
		go func(s *Security) {
			if _, ok := DelistedTickers[s.Ticker]; !ok {
				s.PreProcess(sc.sheetMgr, allStocksData)
				// Make sure the stock's history data is up-to-date.
				sc.RefreshStockHistory(&s.transactions, s.CurrentlyHeld)
			}
			// Pass SP500 quotes to this function to use when calculating transaction level metrics.
			s.CalculateMetrics(sc.dbClient.GetTickerData(s.Ticker), sc.sp500quotes)
			waitGroup.Done()
		}(s)
	}

	// Wait/monitor until all work is complete.
	waitGroup.Wait()
	sc.SendProgressUpdate(95.0)

	// Define the first trading day of the year to reference for YTD calculations.
	firstTradeDay := time.Date(time.Now().Year(), time.January, 3, 0, 0, 0, 0, time.UTC)

	// Calculate the cash balance history in our portfolio.
	sc.CalculateCashBalanceHistory(firstTradeDay)

	// Calculate total invested market value and other summary metrics across all equities.
	sc.CalculatePortfolioSummaryMetrics(firstTradeDay)

	log.Println("---------------------------------")
	log.Printf("Total Market Value: $%f", sc.fullSummary.TotalMarketValue)
	log.Printf("Percentage Gain/Loss: %f%%", sc.fullSummary.PercentageGain)
	log.Printf("Cash Flow YTD: $%f", sc.cashFlowYtd)
	log.Println("---------------------------------")
}
