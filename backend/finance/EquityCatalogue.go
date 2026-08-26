package finance

import (
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/kfwalther/Polly/backend/data"

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
		Transaction{
			Ticker:   "NVDA",
			DateTime: time.Date(2024, 6, 7, 0, 0, 0, 0, time.Local),
			Action:   "Split",
			Shares:   10},
	},
	"SHOP": {
		Transaction{
			Ticker:   "SHOP",
			DateTime: time.Date(2022, 6, 29, 0, 0, 0, 0, time.Local),
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
	// Delisted from Nasdaq in May 2024, failed to comply with listing rules
	"FFIE": {"Faraday Future Intelligent Electric Inc."},
	// Delisted from Nasdaq in June 2024, filed for bankrupcy
	"DMTK": {"Dermtech Inc."},
	// Bought by AMD
	"XLNX": {"Xilinx Inc."},
	// Merged with Gen Digital, delisted in Apr 2025
	"ML": {"MoneyLion"},
	// Acquired by CoStar in Feb 2025
	"MTTR": {"Matterport"},
	// Delisted mutual fund in 2022
	"PONDX": {"PIMCO Income Fund Class D"},
	// This represents cash transactions in our tracker, ignore as a ticker.
	"CASH": {"Placeholder symbol"},
}

// Definition of a equity catalogue to house a portfolio of stock/ETF info in a map.
type EquityCatalogue struct {
	yFinInterface    *YahooFinanceExtension
	sp500quotes      data.Quote
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
func NewEquityCatalogue(equityType string, sheetMgr *GoogleSheetManager, dbClient *data.MongoDbClient, pyScript string, historicalConfig HistoricalDownloadConfig) *EquityCatalogue {
	var ec EquityCatalogue
	ec.equityType = equityType
	// Initialize the interfaces.
	ec.yFinInterface = NewYahooFinanceExtension(pyScript, historicalConfig)
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

func (ec *EquityCatalogue) GetSp500() data.Quote {
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

// Retrieves and stores historical data for the given tickers.
func (ec *EquityCatalogue) RetrieveAndStoreStockData(tickers []string, startDate string, endDate string) {
	quotes, queryToTicker, err := ec.queryHistoricalData(tickers, startDate, endDate)
	if err != nil {
		log.Printf("WARNING: Couldn't get historical ticker data from Yahoo: %s", err)
		return
	}
	for queryTicker, ticker := range queryToTicker {
		quote, ok := quotes[queryTicker]
		if !ok {
			log.Printf("WARNING: No historical data returned from Yahoo for ticker %s", queryTicker)
			continue
		}
		quote.Symbol = strings.TrimSuffix(quote.Symbol, "-USD")
		if quote.Symbol == "" {
			quote.Symbol = ticker
		}
		ec.dbClient.StoreTickerData(quote)
	}
}

// Adds Yahoo-specific ticker suffixes and retrieves historical data in chunks.
func (ec *EquityCatalogue) queryHistoricalData(tickers []string, startDate string, endDate string) (map[string]data.Quote, map[string]string, error) {
	queryTickers := make([]string, 0, len(tickers))
	queryToTicker := make(map[string]string, len(tickers))
	for _, ticker := range tickers {
		queryTicker := ticker
		if ec.equityType == "crypto" {
			queryTicker += "-USD"
		}
		queryTickers = append(queryTickers, queryTicker)
		queryToTicker[queryTicker] = ticker
	}
	if len(queryTickers) == 0 {
		return map[string]data.Quote{}, queryToTicker, nil
	}

	log.Printf("Querying %d tickers from Yahoo: %s ---> %s", len(queryTickers), startDate, endDate)
	quotes, err := ec.yFinInterface.GetHistoricalDataBatch(queryTickers, startDate, endDate)
	return quotes, queryToTicker, err
}

type historicalDataRequest struct {
	ticker    string
	startDate string
	endDate   string
}

func (ec *EquityCatalogue) getHistoricalDataRequest(txns *[]Transaction, currentlyOwned bool) (*historicalDataRequest, bool) {
	if txns == nil || len(*txns) == 0 {
		return nil, false
	}

	ticker := (*txns)[0].Ticker
	if ec.dbClient.TickerExists(ticker) {
		latestDate := ec.dbClient.GetLatestQuote(ticker)
		if currentlyOwned {
			if time.Since(latestDate).Hours() <= 72 {
				return nil, false
			}
			return &historicalDataRequest{
				ticker:    ticker,
				startDate: latestDate.Add(24 * time.Hour).UTC().Format("2006-01-02"),
				endDate:   time.Now().UTC().Format("2006-01-02"),
			}, true
		}

		sellDate := (*txns)[len(*txns)-1].DateTime
		if sellDate.Sub(latestDate).Hours() <= 24 {
			return nil, false
		}
		return &historicalDataRequest{
			ticker:    ticker,
			startDate: latestDate.Add(24 * time.Hour).UTC().Format("2006-01-02"),
			endDate:   sellDate.Add(24 * time.Hour).UTC().Format("2006-01-02"),
		}, true
	}

	return &historicalDataRequest{
		ticker:    ticker,
		startDate: (*txns)[0].DateTime.Format("2006-01-02"),
		endDate:   time.Now().UTC().Format("2006-01-02"),
	}, true
}

func trimHistoricalQuoteToRange(quote data.Quote, startDate string, endDate string) data.Quote {
	start, startErr := time.Parse("2006-01-02", startDate)
	end, endErr := time.Parse("2006-01-02", endDate)
	if startErr != nil || endErr != nil {
		return quote
	}

	trimmed := data.Quote{Symbol: quote.Symbol}
	for i, date := range quote.Date {
		if date.Before(start) || !date.Before(end) {
			continue
		}
		trimmed.Date = append(trimmed.Date, date)
		if i < len(quote.Open) {
			trimmed.Open = append(trimmed.Open, quote.Open[i])
		}
		if i < len(quote.High) {
			trimmed.High = append(trimmed.High, quote.High[i])
		}
		if i < len(quote.Low) {
			trimmed.Low = append(trimmed.Low, quote.Low[i])
		}
		if i < len(quote.Close) {
			trimmed.Close = append(trimmed.Close, quote.Close[i])
		}
		if i < len(quote.Volume) {
			trimmed.Volume = append(trimmed.Volume, quote.Volume[i])
		}
	}
	return trimmed
}

// Retrieves all pending ticker histories in one common date range. Each
// returned quote is trimmed back to its ticker-specific range before storage.
func (ec *EquityCatalogue) retrievePendingHistoricalData(requests []historicalDataRequest) {
	if len(requests) == 0 {
		return
	}
	startDate := requests[0].startDate
	endDate := requests[0].endDate
	tickers := make([]string, 0, len(requests))
	for _, request := range requests {
		if request.startDate < startDate {
			startDate = request.startDate
		}
		if request.endDate > endDate {
			endDate = request.endDate
		}
		tickers = append(tickers, request.ticker)
	}
	quotes, queryToTicker, err := ec.queryHistoricalData(tickers, startDate, endDate)
	if err != nil {
		log.Printf("WARNING: Couldn't get historical ticker data from Yahoo: %s", err)
		return
	}
	for _, request := range requests {
		queryTicker := request.ticker
		if ec.equityType == "crypto" {
			queryTicker += "-USD"
		}
		quote, ok := quotes[queryTicker]
		if !ok {
			log.Printf("WARNING: No historical data returned from Yahoo for ticker %s", queryTicker)
			continue
		}
		quote = trimHistoricalQuoteToRange(quote, request.startDate, request.endDate)
		quote.Symbol = strings.TrimSuffix(quote.Symbol, "-USD")
		if quote.Symbol == "" {
			quote.Symbol = queryToTicker[queryTicker]
		}
		ec.dbClient.StoreTickerData(quote)
	}
}

// Checks existing ticker history data in the DB and pulls any missing data.
// Date-range selection is centralized in getHistoricalDataRequest.
func (ec *EquityCatalogue) RefreshStockHistory(txns *[]Transaction, currentlyOwned bool) {
	request, ok := ec.getHistoricalDataRequest(txns, currentlyOwned)
	if !ok {
		return
	}
	ec.RetrieveAndStoreStockData([]string{request.ticker}, request.startDate, request.endDate)
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
	cash, ok := ec.equities["CASH"]
	if !ok {
		cash, _ = NewEquity("CASH", "Cash")
		ec.equities["CASH"] = cash
	}
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
		cash.ValueHistory[txn.DateTime.Unix()] = curCashAmount
		// Keep track of portfolio cash flow by year.
		if txn.Action == "Deposit" {
			ec.CashFlowByYear[txn.DateTime.Year()] += txn.Value
		} else if txn.Action == "Withdraw" {
			ec.CashFlowByYear[txn.DateTime.Year()] -= txn.Value
		}
	}
	cash.MarketValue = curCashAmount
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
	ec.RefreshStockHistory(&[]Transaction{*NewTransaction("1/1/2015", "SPY", "Buy", "1", "100.0")}, true)
	ec.sp500quotes = ec.dbClient.GetTickerData("SPY")

	// Get the tickers for all equities we've ever owned and request their
	// extended info in one Python invocation. GetTickerData falls back to
	// smaller chunks only for failed or partially returned symbols.
	tickers := make([]string, 0, len(ec.equities))
	for t := range ec.equities {
		if _, ok := DelistedTickers[t]; ok {
			continue
		}
		ticker := t
		if ec.equityType == "crypto" {
			ticker += "-USD"
		}
		tickers = append(tickers, ticker)
	}
	allStocksData := ec.yFinInterface.GetTickerData(tickers)
	log.Printf("Queried Yahoo finance for %d equities...\n", len(tickers))
	// Preprocess equities before planning history requests because ownership is
	// determined during preprocessing.
	pendingRequests := make([]historicalDataRequest, 0, len(ec.equities))
	for _, s := range ec.equities {
		if _, ok := DelistedTickers[s.Ticker]; ok {
			continue
		}
		s.PreProcess(ec.sheetMgr, &allStocksData)
		if request, ok := ec.getHistoricalDataRequest(&s.transactions, s.CurrentlyHeld); ok {
			pendingRequests = append(pendingRequests, *request)
		}
	}
	ec.retrievePendingHistoricalData(pendingRequests)

	// Setup a wait group.
	var waitGroup sync.WaitGroup
	// Specify the number of metrics to wait for.
	waitGroup.Add(len(ec.equities))

	log.Printf("Processing %d equities...\n", len(ec.equities))
	// Iterate thru each equity in the map, and calculate its data.
	for _, s := range ec.equities {
		// Launch a new goroutine for this equity.
		go func(s *Equity) {
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
	log.Printf("Market Value Jan 1: $%f", ec.portfolioSummary.MarketValueJan1)
	log.Printf("Percentage Gain/Loss: %f%%", ec.portfolioSummary.PercentageGain)
	log.Printf("Cash Flow YTD: %v", ec.CashFlowByYear)
	log.Println("---------------------------------")
}
