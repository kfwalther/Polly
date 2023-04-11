package finance

import (
	"errors"
	"log"
	"sort"
	"time"

	"github.com/markcheno/go-quote"
	"github.com/piquette/finance-go"
	"github.com/piquette/finance-go/equity"
)

// Definition of a security to hold the transactions for a particular stock/ETF.
type Security struct {
	id                       uint
	Ticker                   string                `json:"ticker"`
	SecurityType             string                `json:"securityType"`
	MarketPrice              float64               `json:"marketPrice"`
	MarketValue              float64               `json:"marketValue"`
	DailyGain                float64               `json:"dailyGain"`
	DailyGainPercentage      float64               `json:"dailyGainPercentage"`
	UnitCostBasis            float64               `json:"unitCostBasis"`
	TotalCostBasis           float64               `json:"totalCostBasis"`
	NumShares                float64               `json:"numShares"`
	RealizedGain             float64               `json:"realizedGain"`
	UnrealizedGain           float64               `json:"unrealizedGain"`
	UnrealizedGainPercentage float64               `json:"unrealizedGainPercentage"`
	TotalGain                float64               `json:"totalGain"`
	ValueHistory             map[time.Time]float64 `json:"valueHistory"`
	priceHistory             quote.Quote
	transactions             []Transaction
}

// Constructor for a new Security object.
func NewSecurity(tkr string, secType string) (*Security, error) {
	// Validate the security type before creating the object.
	if secType != "Stock" && secType != "ETF" && secType != "Mutual Fund" && secType != "Cash" {
		return nil, errors.New("Could not create Security. Invalid security type (" + secType + ") for " + tkr)
	}
	var s Security
	s.Ticker = tkr
	s.SecurityType = secType
	s.ValueHistory = make(map[time.Time]float64)
	s.transactions = make([]Transaction, 0)
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

func (s *Security) GetQuoteOfSP500(quoteDate time.Time, sp500Quotes quote.Quote) float64 {
	// Get index of this date (Yahoo dates are returned in UTC).
	idx := indexOf(getUtcDate(quoteDate), sp500Quotes.Date)
	if idx != -1 {
		return sp500Quotes.Close[idx]
	} else {
		y1, m1, d1 := quoteDate.Date()
		yNow, mNow, dNow := time.Now().Date()
		if y1 == yNow && m1 == mNow && d1 == dNow {
			// Likely don't have today's S&P500 quote queried yet, just return the latest quote.
			return sp500Quotes.Close[len(sp500Quotes.Close)-1]
		} else {
			// TODO: ERROR HERE
			return 0.0
		}
	}
}

// Grab the current market price of this ticker symbol, from the web.
func (s *Security) DetermineCurrentPrice(q *finance.Equity) {
	// Based on the market's current state, grab the proper current quoted price.
	if q.MarketState == "PRE" && q.PreMarketPrice > 0.0 {
		s.MarketPrice = q.PreMarketPrice
	} else if q.MarketState == "POST" && q.PostMarketPrice > 0.0 {
		s.MarketPrice = q.PostMarketPrice
	} else {
		s.MarketPrice = q.RegularMarketPrice
	}
}

// Sorts the transactions for this security, and adds in any stock splits we need to account for.
func (s *Security) PreProcess() {
	// Ignore ticker CASH for now, may use this later.
	if s.Ticker == "CASH" {
		return
	}
	// Lookup if this security has any stock splits to account for.
	if val, ok := StockSplits[s.Ticker]; ok {
		s.transactions = append(s.transactions, val...)
	}
	// Order the transactions by date, using anonymous function.
	sort.Slice(s.transactions, func(i, j int) bool {
		return s.transactions[i].DateTime.Before(s.transactions[j].DateTime)
	})
}

// Calculate various metrics about this security.
func (s *Security) CalculateMetrics(histQuotes quote.Quote, sp500Quotes quote.Quote) {
	// Ignore ticker CASH for now, may use this later.
	if s.Ticker == "CASH" {
		return
	}
	// hasSplits := false
	splitMultiple := 1.0
	if _, ok := StockSplits[s.Ticker]; ok {
		// hasSplits = true
		// Iterate thru all transactions to calculate the total split multiple for this stock.
		for _, txn := range s.transactions {
			if txn.Action == "Split" {
				splitMultiple *= txn.Shares
			}
		}
	}
	// Get the current info for the given ticker.
	q, err := equity.Get(s.Ticker)
	if err != nil || q == nil {
		return
	}
	// Note, for calculations below, Yahoo stock price history is already split-adjusted.
	s.priceHistory = histQuotes

	// // Upon first buy, grab historical quotes from that date forward to calculate running value total each day.
	// if s.transactions[0].Action == "Buy" {
	// 	// TODO: WTF WHY THIS RETURN NOTHING?!?!
	// 	s.priceHistory, err = quote.NewQuoteFromYahoo(s.Ticker, s.transactions[0].DateTime.Format("2006-01-02"), time.Now().Format("2006-01-02"), quote.Daily, true)
	// 	if err != nil {
	// 		log.Printf("Can't query history for ticker %s: %s", s.Ticker, err)
	// 		return
	// 	} else {
	// 		log.Printf("Successfully queried history for ticker: %s", s.Ticker)
	// 	}
	// } else {
	// 	log.Printf("WARNING: First transaction for %s is not a Buy (%s), what is going on?!\n", s.Ticker, s.transactions[0].Action)
	// 	return
	// }

	// TODO: Figure out what fields we can grab from this quote...
	// if s.Ticker == "S" {
	// 	log.Printf("Book Value: %f", q.BookValue)
	// 	log.Printf("Fwd PE: %f", q.ForwardPE)
	// 	log.Printf("Price / Book: %f", q.PriceToBook)
	// 	log.Printf("Trailing PE: %f", q.TrailingPE)
	// 	log.Printf("Market Cap: %d", q.MarketCap)
	// 	log.Printf("Quote Source: %s", q.QuoteSource)
	// }

	// Get current market price.
	s.DetermineCurrentPrice(q)
	// Create a slice to use as a FIFO queue.
	buyQ := make([]Transaction, 0)

	tIdx := 0
	curShares := 0.0

	// Iterate through each date from first purchase.
	for dIdx := 0; dIdx < len(s.priceHistory.Date); dIdx++ {
		// if s.Ticker == "TSLA" {
		// 	log.Printf("%v", s.priceHistory.Date[dIdx])
		// }
		// Was there a transaction on this date? (Keep iterating if multiple on this date)
		for tIdx < len(s.transactions) &&
			s.priceHistory.Date[dIdx].Equal(getUtcDate(s.transactions[tIdx].DateTime)) {

			// Get a reference to the current txn.
			t := &s.transactions[tIdx]
			tIdx++
			// if s.Ticker == "TSLA" {
			// 	log.Printf("%s - %s", t.Action, t.DateTime.Format("2006-01-02"))
			// }
			// For buys, increment number of shares.
			if t.Action == "Buy" {
				curShares += t.Shares
				// Add the txn to the buy queue.
				buyQ = append(buyQ, *t)
				// Calculate the return had we bought S&P500 for this transaction.
				spDateOfTxn := s.GetQuoteOfSP500(t.DateTime, sp500Quotes)
				spNow := s.GetQuoteOfSP500(time.Now(), sp500Quotes)
				t.Sp500Return = ((spNow - spDateOfTxn) / spDateOfTxn) * 100.0
			} else if t.Action == "Sell" {
				curShares -= t.Shares
				// Calculate the return had we sold S&P500 for this transaction.
				spDateOfTxn := s.GetQuoteOfSP500(t.DateTime, sp500Quotes)
				spNow := s.GetQuoteOfSP500(time.Now(), sp500Quotes)
				t.Sp500Return = -((spNow - spDateOfTxn) / spDateOfTxn) * 100.0
				// In the loop below, calculate the realized gain from this sale.
				remainingShares := t.Shares
				for remainingShares > 0 {
					// Make sure we have buys to cover remaining shares in the sell.
					if len(buyQ) > 0 {
						if buyQ[0].Shares > remainingShares {
							// Remaining sell shares are covered by this buy.
							s.RealizedGain += (t.Price - buyQ[0].Price) * remainingShares
							buyQ[0].Shares -= remainingShares
							remainingShares = 0
						} else if buyQ[0].Shares == remainingShares {
							// Shares in this buy equal remaining shares, pop the buy off the queue.
							s.RealizedGain += (t.Price - buyQ[0].Price) * remainingShares
							remainingShares = 0
							buyQ = buyQ[1:]
						} else if buyQ[0].Shares < remainingShares {
							// This buy is completely covered by sell, pop it.
							s.RealizedGain += (t.Price - buyQ[0].Price) * buyQ[0].Shares
							remainingShares -= buyQ[0].Shares
							buyQ = buyQ[1:]
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
				for i := range buyQ {
					buyQ[i].Price /= t.Shares
					buyQ[i].Shares *= t.Shares
				}
				splitMultiple /= t.Shares
			}

			// Calculate the theoretical return on this txn, if we held.
			if t.Action == "Buy" || t.Action == "Sell" {
				isNeg := 1.0
				if t.Action == "Sell" {
					isNeg = -1.0
				}
				// Calculate the theoretical total return (%) of each txn (using any split multiple from above).
				t.TotalReturn = isNeg * ((s.MarketPrice*t.Shares*splitMultiple - t.Value) / t.Value) * 100.0
				t.ExcessReturn = t.TotalReturn - t.Sp500Return
			}
		}
		// Save the value of this stock in our portfolio on this date (if still owned).
		if curShares > 0 {
			s.ValueHistory[s.priceHistory.Date[dIdx]] = curShares * s.priceHistory.Close[dIdx] * splitMultiple
		} else {
			// If share count is zero, we need not calculate any more dates for this stock.
			break
		}
	}

	// if s.Ticker == "TSLA" {
	// 	log.Printf("%v", s.ValueHistory)
	// }
	// Calculate cost bases and unrealized gains with any remaining buy shares in the buy queue.
	for _, txn := range buyQ {
		s.NumShares += txn.Shares
		s.TotalCostBasis += txn.Shares * txn.Price
	}
	// Calculate the total 1-day gain/loss for this stock.
	s.DailyGain = q.RegularMarketChange * s.NumShares
	s.DailyGainPercentage = q.RegularMarketChangePercent
	if s.NumShares > 0 {
		// Unit cost basis
		s.UnitCostBasis = s.TotalCostBasis / s.NumShares
		// Total market value
		s.MarketValue = s.MarketPrice * s.NumShares
		// Unrealized gain (and percentage)
		s.UnrealizedGain = s.MarketValue - s.TotalCostBasis
		s.UnrealizedGainPercentage = (s.UnrealizedGain / s.TotalCostBasis) * 100.0
	}
	// Get the total gain.
	s.TotalGain = s.UnrealizedGain + s.RealizedGain
}

func (s *Security) DisplayMetrics() {
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
		log.Printf("   Num Shares: $%f\n", txn.Shares)
		log.Printf("   Price: $%f\n", txn.Price)
		log.Printf("   Total Return: $%f\n", txn.TotalReturn)
		log.Printf("   SP500 Return: $%f\n", txn.Sp500Return)
		log.Printf("   Excess Return: $%f\n", txn.ExcessReturn)
	}
}
