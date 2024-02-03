package finance

import (
	"errors"
	"log"
	"math"
	"sort"
	"strconv"
	"time"

	"github.com/markcheno/go-quote"
)

// Definition of a equity to hold the transactions for a particular stock/ETF.
type Equity struct {
	id                              uint
	Ticker                          string            `json:"ticker"`
	EquityType                      string            `json:"equityType"`
	MarketPrice                     float64           `json:"marketPrice"`
	MarketPrevClosePrice            float64           `json:"marketPrevClosePrice"`
	MarketValue                     float64           `json:"marketValue"`
	DailyGain                       float64           `json:"dailyGain"`
	DailyGainPercentage             float64           `json:"dailyGainPercentage"`
	UnitCostBasis                   float64           `json:"unitCostBasis"`
	TotalCostBasis                  float64           `json:"totalCostBasis"`
	NumShares                       float64           `json:"numShares"`
	CurrentlyHeld                   bool              `json:"currentlyHeld"`
	RealizedGain                    float64           `json:"realizedGain"`
	UnrealizedGain                  float64           `json:"unrealizedGain"`
	UnrealizedGainPercentage        float64           `json:"unrealizedGainPercentage"`
	TotalGain                       float64           `json:"totalGain"`
	ValueAllTimeHigh                float64           `json:"valueAllTimeHigh"`
	HoldingDays                     uint              `json:"holdingDays"` // TODO: Calculate this and use it...
	Sector                          string            `json:"sector"`
	Industry                        string            `json:"industry"`
	CurrentQuarter                  string            `json:"currentQuarter"`
	MarketCap                       float64           `json:"marketCap"`
	RevenueTtm                      float64           `json:"revenueTtm"`
	RevenueCurrentYearEstimate      float64           `json:"revenueCurrentYearEstimate"`
	RevenueNextYearEstimate         float64           `json:"revenueNextYearEstimate"`
	GrossMargin                     float64           `json:"grossMargin"`
	PriceToSalesTtm                 float64           `json:"priceToSalesTtm"`
	PriceToSalesNtm                 float64           `json:"priceToSalesNtm"`
	PriceToFcfTtm                   float64           `json:"priceToFcfTtm"`
	RevenueGrowthPercentageYoy      float64           `json:"revenueGrowthPercentageYoy"`
	RevenueGrowthPercentageNextYear float64           `json:"revenueGrowthPercentageNextYear"`
	TrailingPE                      float64           `json:"trailingPE"`
	ForwardPE                       float64           `json:"forwardPE"`
	ValueHistory                    map[int64]float64 `json:"valueHistory"`
	// Some arrays/objects to support metric calculation.
	buyQ          []Transaction
	priceHistory  quote.Quote
	sp500History  quote.Quote
	splitMultiple float64
	transactions  []Transaction
	// Financial history data
	revenueUnits                   string
	fcfTtm                         float64
	quarterlyDates                 []string
	quarterlyRevenue               []float64
	quarterlyFCF                   []float64
	quarterlyGrossProfitPercentage []float64
	quarterlyPercentSM             []float64
	quarterlyPercentFCF            []float64
	quarterlyPercentSBC            []float64
}

// Constructor for a new Equity object.
func NewEquity(tkr string, eqType string) (*Equity, error) {
	// Validate the equity type before creating the object.
	if eqType != "Stock" && eqType != "ETF" && eqType != "Mutual Fund" && eqType != "Crypto" && eqType != "Cash" {
		return nil, errors.New("Could not create Equity. Invalid equity type (" + eqType + ") for " + tkr)
	}
	var s Equity
	s.Ticker = tkr
	s.EquityType = eqType
	s.ValueHistory = make(map[int64]float64)
	s.transactions = make([]Transaction, 0)
	// Create a slice to use as a FIFO queue for calculating metrics.
	s.buyQ = make([]Transaction, 0)
	// Create slices for the financial data.
	s.quarterlyDates = make([]string, 0)
	s.quarterlyRevenue = make([]float64, 0)
	s.quarterlyFCF = make([]float64, 0)
	s.quarterlyGrossProfitPercentage = make([]float64, 0)
	s.quarterlyPercentSM = make([]float64, 0)
	s.quarterlyPercentFCF = make([]float64, 0)
	s.quarterlyPercentSBC = make([]float64, 0)
	return &s, nil
}

// Helper function to find the index in an array containing the entry matching the given time.
func indexOf(target time.Time, arr []time.Time) int {
	for idx, t := range arr {
		if t.Equal(target) {
			return idx
		}
	}
	return -1 // not found.
}

func getUtcDate(inDateTime time.Time) time.Time {
	return time.Date(inDateTime.Year(), inDateTime.Month(), inDateTime.Day(), 0, 0, 0, 0, time.UTC)
}

func datesEqual(inDate1 time.Time, inDate2 time.Time) bool {
	return inDate1.Format("2006-01-02") == inDate2.Format("2006-01-02")
}

// A simple helper function to calculate and save the max value in this equity's value history.
func (s *Equity) getMaxValueFromHistory() {
	max := math.Inf(-1)
	for _, v := range s.ValueHistory {
		if v > max {
			max = v
		}
	}
	// If no data, set max to 0.
	if max == math.Inf(-1) {
		max = 0.0
	}
	s.ValueAllTimeHigh = max
}

func (s *Equity) GetQuoteOfSP500(quoteDate time.Time) float64 {
	// Get index of this date (Yahoo dates are returned in UTC).
	idx := indexOf(getUtcDate(quoteDate), s.sp500History.Date)
	if idx != -1 {
		return s.sp500History.Close[idx]
	} else {
		y1, m1, d1 := quoteDate.Date()
		yNow, mNow, dNow := time.Now().Date()
		if y1 == yNow && m1 == mNow && d1 == dNow {
			// Likely don't have today's S&P500 quote queried yet, just return the latest quote.
			return s.sp500History.Close[len(s.sp500History.Close)-1]
		} else {
			// Transaction likely occurred on a weekend. Try +/- a few days.
			for i := 1; i < 4; i++ {
				idxMinusOne := indexOf(getUtcDate(quoteDate.Add(time.Duration(-24*i)*time.Hour)), s.sp500History.Date)
				if idxMinusOne != -1 {
					return s.sp500History.Close[idxMinusOne]
				}
				idxPlusOne := indexOf(getUtcDate(quoteDate.Add(time.Duration(24*i)*time.Hour)), s.sp500History.Date)
				if idxPlusOne != -1 {
					return s.sp500History.Close[idxPlusOne]
				}
			}
			log.Printf("WARNING: Could not retrieve S&P500 quote for date: %v", quoteDate)
			return 0.0
		}
	}
}

// Process the financial history data for one stock from the growth stock spreadsheet.
func (s *Equity) processFinancialHistoryData(data [][]interface{}) {
	numQs := 0
	if data == nil {
		return
	}
	// Iterate thru each date we have data for (until blank). Format is YYYYMMDD
	for i := 1; i < len(data[0]); i++ {
		if data[0][i] != "" {
			s.quarterlyDates = append(s.quarterlyDates, data[0][i].(string))
			numQs++
		} else {
			break
		}
	}

	// Check if any quarterly history data found.
	if numQs > 1 {
		multiplier := 1.0
		if s.revenueUnits == "M" {
			multiplier = 1000.0
		}
		// Iterate through the indices we have data for, filling in the rest of the financials arrays.
		for _, row := range data {
			// Read revenue history, and account for revenue units in millions where necessary.
			if row[0].(string) == "Total Revenue" {
				for i := 1; i <= numQs; i++ {
					val, _ := strconv.ParseFloat(row[i].(string), 64)
					s.quarterlyRevenue = append(s.quarterlyRevenue, val*multiplier)
				}
				// Calculate last 4 Qs of revenue (TTM), and revenue growth YoY.
				s.RevenueTtm = s.quarterlyRevenue[numQs-1] + s.quarterlyRevenue[numQs-2] + s.quarterlyRevenue[numQs-3] + s.quarterlyRevenue[numQs-4]
				if s.quarterlyRevenue[numQs-5] > 0.001 {
					s.RevenueGrowthPercentageYoy = (s.quarterlyRevenue[numQs-1] - s.quarterlyRevenue[numQs-5]) / s.quarterlyRevenue[numQs-5]
				} else {
					log.Printf("WARNING: Couldn't calculate rev growth YoY for %s", s.Ticker)
				}
			}
			// Read all gross margin history.
			if row[0].(string) == "Gross Margins (%)" {
				for i := 1; i <= numQs; i++ {
					val, _ := strconv.ParseFloat(row[i].(string), 64)
					s.quarterlyGrossProfitPercentage = append(s.quarterlyGrossProfitPercentage, val)
				}
				// Record latest gross margin, we make a percentage on front-end.
				s.GrossMargin = s.quarterlyGrossProfitPercentage[numQs-1] / 100
			}
			// Read Sales & Marketing as a percentage of revenue history.
			if row[0].(string) == "S&M / Revenue (%)" {
				for i := 1; i <= numQs; i++ {
					val, _ := strconv.ParseFloat(row[i].(string), 64)
					s.quarterlyPercentSM = append(s.quarterlyPercentSM, val)
				}
			}
			// Read FCF history, and account for units in millions where necessary.
			if row[0].(string) == "FCF" {
				for i := 1; i <= numQs; i++ {
					val, _ := strconv.ParseFloat(row[i].(string), 64)
					s.quarterlyFCF = append(s.quarterlyFCF, val*multiplier)
				}
				// Calculate last 4 Qs of FCF (TTM).
				s.fcfTtm = s.quarterlyFCF[numQs-1] + s.quarterlyFCF[numQs-2] + s.quarterlyFCF[numQs-3] + s.quarterlyFCF[numQs-4]
			}
			// Read FCF as a percentage of revenue history.
			if row[0].(string) == "FCF / Revenue (%)" {
				for i := 1; i <= numQs; i++ {
					val, _ := strconv.ParseFloat(row[i].(string), 64)
					s.quarterlyPercentFCF = append(s.quarterlyPercentFCF, val)
				}
			}
			// Read SBC as a percentage of revenue history.
			if row[0].(string) == "SBC / Revenue (%)" {
				for i := 1; i <= numQs; i++ {
					val, _ := strconv.ParseFloat(row[i].(string), 64)
					s.quarterlyPercentSBC = append(s.quarterlyPercentSBC, val)
				}
			}
		}
	}
}

// Sorts the transactions for this equity, adds any stock splits, and pre-populates data from Growth Stock Google Sheet.
func (s *Equity) PreProcess(sheetMgr *GoogleSheetManager, stockDataMap *map[string]interface{}) {
	// Cash shouldn't be considered here.
	if s.Ticker == "CASH" {
		return
	}
	// Lookup if this equity has any stock splits to account for.
	if val, ok := StockSplits[s.Ticker]; ok {
		s.transactions = append(s.transactions, val...)
	}

	// Order the transactions by date, using anonymous function.
	sort.Slice(s.transactions, func(i, j int) bool {
		return s.transactions[i].DateTime.Before(s.transactions[j].DateTime)
	})
	// Calculate number of shares currently owned, and split multiple.
	s.splitMultiple = 1.0
	curShares := 0.0
	for _, txn := range s.transactions {
		if txn.Action == "Buy" {
			curShares += txn.Shares
		} else if txn.Action == "Sell" {
			curShares -= txn.Shares
		} else if txn.Action == "Split" {
			curShares *= txn.Shares
			s.splitMultiple *= txn.Shares
		}
	}
	var stockData map[string]interface{} = nil
	// Crypto tickers from Yahoo are searchable only with '-USD' after coin type.
	tickerLookup := s.Ticker
	if s.EquityType == "Crypto" {
		tickerLookup = s.Ticker + "-USD"
	}
	// If Yahoo returned data for this equity, try to extract it from the JSON map.
	if stockMapEntry, ok := (*stockDataMap)[tickerLookup]; ok {
		if stockData, ok = stockMapEntry.(map[string]interface{}); ok {
			curPriceName := "currentPrice"
			if s.EquityType == "ETF" {
				// ETFs don't have currentPrice, use navPrice instead.
				curPriceName = "navPrice"
			} else if s.EquityType == "Mutual Fund" || s.EquityType == "Crypto" {
				// Mutual funds and crypto don't have currentPrice, use previousClose instead.
				curPriceName = "previousClose"
			}
			// Save the current market price.
			if s.MarketPrice, ok = stockData[curPriceName].(float64); !ok {
				s.MarketPrice = 0.0
				log.Printf("WARNING: Couldn't obtain current market price for %s", s.Ticker)
			}
			if s.MarketPrevClosePrice, ok = stockData["previousClose"].(float64); !ok {
				s.MarketPrevClosePrice = 0.0
			}
		} else {
			log.Printf("WARNING: Couldn't convert data map from Yahoo for ticker %s", s.Ticker)
		}
	} else {
		log.Printf("WARNING: No data returned from Yahoo for ticker %s", s.Ticker)
	}
	// Do we currently hold this stock (account for minor accounting differences).
	if curShares > 0.001 {
		s.CurrentlyHeld = true
		// If a stock we currently own, save some addtl data.
		if s.EquityType == "Stock" && stockData != nil {
			var err error
			var ok bool
			// Check if Yahoo returned any data for these fields.
			if s.Sector, ok = stockData["sector"].(string); !ok {
				s.Sector = ""
			}
			if s.Industry, ok = stockData["industry"].(string); !ok {
				s.Industry = ""
			}
			if s.PriceToSalesTtm, ok = stockData["priceToSalesTrailing12Months"].(float64); !ok {
				s.PriceToSalesTtm = 0.0
			}
			if s.MarketCap, ok = stockData["marketCap"].(float64); !ok {
				s.MarketCap = 0.0
			}
			if s.GrossMargin, ok = stockData["grossMargins"].(float64); !ok {
				s.GrossMargin = 0.0
			}
			if s.RevenueGrowthPercentageYoy, ok = stockData["revenueGrowth"].(float64); !ok {
				s.RevenueGrowthPercentageYoy = 0.0
			}
			if s.TrailingPE, ok = stockData["trailingPE"].(float64); !ok {
				s.TrailingPE = 0.0
			}
			if s.ForwardPE, ok = stockData["forwardPE"].(float64); !ok {
				s.ForwardPE = 0.0
			}
			log.Printf("Grabbing %s from Revenue sheet...", s.Ticker)
			sheetData := sheetMgr.GetRevenueData(s.Ticker)
			if sheetData != nil {
				// Save initial revenue info from top of current Google sheet.
				s.revenueUnits = sheetData.Values[0][0].(string)
				s.CurrentQuarter = sheetData.Values[1][0].(string)
				// We save these revenue values in thousands (not $M or $B).
				if s.RevenueCurrentYearEstimate, err = strconv.ParseFloat(sheetData.Values[3][0].(string), 64); err != nil {
					log.Printf("WARNING: Unable to parse current year Revenue estimate from %s sheet: %v", s.Ticker, err)
				}
				if s.RevenueNextYearEstimate, err = strconv.ParseFloat(sheetData.Values[4][0].(string), 64); err != nil {
					log.Printf("WARNING: Unable to parse next year Revenue estimate from %s sheet: %v", s.Ticker, err)
				}
				// Adjust for revenue logged in $M, not thousands.
				if s.revenueUnits == "M" {
					s.RevenueCurrentYearEstimate *= 1000
					s.RevenueNextYearEstimate *= 1000
				}
				// Save info from revenue history within sheet, if populated.
				s.processFinancialHistoryData(sheetMgr.GetAllRevenueData(s.Ticker).Values)
			}
		}
	} else {
		s.CurrentlyHeld = false
	}
}

// Calculate stock holdings info and stats based on individual transactions.
func (s *Equity) CalculateTransactionData(txnIdx int, curShares float64) float64 {
	// Get a reference to the current txn.
	t := &s.transactions[txnIdx]
	// For buys, increment number of shares.
	if t.Action == "Buy" {
		curShares += t.Shares
		// Add the txn to the buy queue.
		s.buyQ = append(s.buyQ, *t)
		// Calculate the return had we bought S&P500 for this transaction.
		spDateOfTxn := s.GetQuoteOfSP500(t.DateTime)
		spNow := s.GetQuoteOfSP500(time.Now())
		t.Sp500Return = ((spNow - spDateOfTxn) / spDateOfTxn) * 100.0
	} else if t.Action == "Sell" {
		curShares -= t.Shares
		// Calculate the return had we sold S&P500 for this transaction.
		spDateOfTxn := s.GetQuoteOfSP500(t.DateTime)
		spNow := s.GetQuoteOfSP500(time.Now())
		t.Sp500Return = -((spNow - spDateOfTxn) / spDateOfTxn) * 100.0
		// In the loop below, calculate the realized gain from this sale.
		remainingShares := t.Shares
		for remainingShares > 0 {
			// Make sure we have buys to cover remaining shares in the sell.
			if len(s.buyQ) > 0 {
				if s.buyQ[0].Shares > remainingShares {
					// Remaining sell shares are covered by this buy.
					s.RealizedGain += (t.Price - s.buyQ[0].Price) * remainingShares
					s.buyQ[0].Shares -= remainingShares
					remainingShares = 0
				} else if s.buyQ[0].Shares == remainingShares {
					// Shares in this buy equal remaining shares, pop the buy off the queue.
					s.RealizedGain += (t.Price - s.buyQ[0].Price) * remainingShares
					remainingShares = 0
					s.buyQ = s.buyQ[1:]
				} else if s.buyQ[0].Shares < remainingShares {
					// This buy is completely covered by sell, pop it.
					s.RealizedGain += (t.Price - s.buyQ[0].Price) * s.buyQ[0].Shares
					remainingShares -= s.buyQ[0].Shares
					s.buyQ = s.buyQ[1:]
				}
			} else {
				// Queue is empty, but apparently have more sold shares to account for.
				// This is usually due to re-invested dividends.
				additionalGains := remainingShares * t.Price
				s.RealizedGain += additionalGains
				// log.Printf("%s is oversold - Adding remaining shares to realized gain (%f shares, total $%f)\n", t.Ticker, remainingShares, additionalGains)
				break
			}
		}
	} else if t.Action == "Split" {
		curShares *= t.Shares
		// Apply the split to all txns in the buy queue.
		for i := range s.buyQ {
			s.buyQ[i].Price /= t.Shares
			s.buyQ[i].Shares *= t.Shares
		}
		s.splitMultiple /= t.Shares
	}

	if s.MarketPrice > 0.0 {
		// Calculate the theoretical return on this txn, if we held.
		if t.Action == "Buy" || t.Action == "Sell" {
			isNeg := 1.0
			if t.Action == "Sell" {
				isNeg = -1.0
			}
			// Calculate the theoretical total return (%) of each txn (using any split multiple from above).
			t.TotalReturn = isNeg * ((s.MarketPrice*t.Shares*s.splitMultiple - t.Value) / t.Value) * 100.0
			t.ExcessReturn = t.TotalReturn - t.Sp500Return
		}
	}
	// Round down small values to essentially zero.
	if curShares < 0.001 {
		curShares = 0.0
	}
	return curShares
}

// Calculate various metrics about this equity.
func (s *Equity) CalculateMetrics(histQuotes quote.Quote, sp500Quotes quote.Quote) {
	// Ignore ticker CASH for now, may use this later.
	if s.Ticker == "CASH" {
		return
	}
	// Reset the slice to use as a FIFO queue for calculating metrics.
	s.buyQ = make([]Transaction, 0)

	// Note, for calculations below, Yahoo stock price history is already split-adjusted.
	s.priceHistory = histQuotes
	s.sp500History = sp500Quotes

	tIdx := 0
	curShares := 0.0

	// Iterate through each date in the price history since first purchase.
	if len(s.priceHistory.Date) > 0 && s.MarketPrice != 0.0 {
		for dIdx := 0; dIdx < len(s.priceHistory.Date); dIdx++ {
			// Was there a transaction on this date? (Keep iterating if multiple on this date)
			for tIdx < len(s.transactions) && datesEqual(s.priceHistory.Date[dIdx], s.transactions[tIdx].DateTime) {
				// Calculate metrics for this individual txn, and any realized gain.
				curShares = s.CalculateTransactionData(tIdx, curShares)
				tIdx++
			}
			// Save the value of this stock in our portfolio on this date (if still owned).
			if curShares > 0 || tIdx < len(s.transactions) {
				s.ValueHistory[s.priceHistory.Date[dIdx].Unix()] = curShares * s.priceHistory.Close[dIdx] * s.splitMultiple
			} else {
				// If share count is zero, and no more transactions, we need not calculate any more dates for this stock.
				break
			}
		}
		// Now that we have full value history for this equity, calculate the all-time high.
		s.getMaxValueFromHistory()
		// Calculate cost bases and unrealized gains with any remaining buy shares in the buy queue.
		for _, txn := range s.buyQ {
			s.NumShares += txn.Shares
			s.TotalCostBasis += txn.Shares * txn.Price
		}
		if s.CurrentlyHeld {
			// Calculate holding days.
			s.HoldingDays = uint(math.Ceil(time.Now().Sub(s.transactions[0].DateTime).Hours() / 24))
			// Calculate the total 1-day gain/loss for this stock.
			s.DailyGain = (s.MarketPrice - s.MarketPrevClosePrice) * s.NumShares
			if s.MarketPrevClosePrice > 0.001 {
				s.DailyGainPercentage = (s.MarketPrice - s.MarketPrevClosePrice) * 100.0 / s.MarketPrevClosePrice
			}
		}
	} else {
		log.Printf("Calculating reduced metrics for %s", s.Ticker)
		// These stocks had errors, and/or no longer exist. Skip value history calculations.
		for tIdx < len(s.transactions) {
			// Calculate metrics for this individual txn, and any realized gain.
			curShares = s.CalculateTransactionData(tIdx, curShares)
			tIdx++
		}
	}

	// Calculate additional metrics for currently-held equities.
	if s.NumShares > 0.001 {
		// Unit cost basis
		s.UnitCostBasis = s.TotalCostBasis / s.NumShares
		// Total market value
		s.MarketValue = s.MarketPrice * s.NumShares
		// Unrealized gain (and percentage)
		s.UnrealizedGain = s.MarketValue - s.TotalCostBasis
		s.UnrealizedGainPercentage = (s.UnrealizedGain / s.TotalCostBasis) * 100.0
		// Financials (P/S ratios, revenue % increase estimates)
		if s.RevenueTtm > 0.0001 {
			s.PriceToSalesTtm = s.MarketCap / (s.RevenueTtm * 1000)
		}
		if s.fcfTtm > 0.0001 {
			s.PriceToFcfTtm = s.MarketCap / (s.fcfTtm * 1000)
		}
		if s.RevenueCurrentYearEstimate > 0.0001 {
			s.PriceToSalesNtm = s.MarketCap / (s.RevenueNextYearEstimate * 1000)
			s.RevenueGrowthPercentageNextYear = (s.RevenueNextYearEstimate - s.RevenueCurrentYearEstimate) / s.RevenueCurrentYearEstimate
		}
	}
	// Get the total gain.
	s.TotalGain = s.UnrealizedGain + s.RealizedGain
}

func (s *Equity) DisplayMetrics() {
	log.Printf("---------------%s----------------", s.Ticker)
	log.Printf("Market Price: $%f\n", s.MarketPrice)
	log.Printf("Number of Shares: %f\n", s.NumShares)
	log.Printf("Market Value: %f\n", s.MarketValue)
	log.Printf("Daily Gain: $%f\n", s.DailyGain)
	log.Printf("Daily Gain Percent: %f\n", s.DailyGainPercentage)
	log.Printf("Unit Cost Basis: $%f\n", s.UnitCostBasis)
	log.Printf("Total Cost Basis: $%f\n", s.TotalCostBasis)
	log.Printf("Unrealized Gain: $%f\n", s.UnrealizedGain)
	log.Printf("Unrealized Gain Percent: %f\n", s.UnrealizedGainPercentage)
	log.Printf("Realized Gain: $%f\n", s.RealizedGain)
	for _, txn := range s.transactions {
		log.Printf("   -----TXN: %s -----\n", txn.DateTime.Format("2006-01-02"))
		log.Printf("   Num Shares: %f\n", txn.Shares)
		log.Printf("   Price: $%f\n", txn.Price)
		log.Printf("   Total Return: %f\n", txn.TotalReturn)
		log.Printf("   SP500 Return: %f\n", txn.Sp500Return)
		log.Printf("   Excess Return: %f\n", txn.ExcessReturn)
	}
}
