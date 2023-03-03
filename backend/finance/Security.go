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
	Ticker                   string  `json:"ticker"`
	SecurityType             string  `json:"securityType"`
	MarketPrice              float64 `json:"marketPrice"`
	MarketValue              float64 `json:"marketValue"`
	DailyGain                float64 `json:"dailyGain"`
	DailyGainPercentage      float64 `json:"dailyGainPercentage"`
	UnitCostBasis            float64 `json:"unitCostBasis"`
	TotalCostBasis           float64 `json:"totalCostBasis"`
	NumShares                float64 `json:"numShares"`
	RealizedGain             float64 `json:"realizedGain"`
	UnrealizedGain           float64 `json:"unrealizedGain"`
	UnrealizedGainPercentage float64 `json:"unrealizedGainPercentage"`
	TotalGain                float64 `json:"totalGain"`
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

func (s *Security) GetQuoteOfSP500(quoteDate time.Time, sp500Quotes quote.Quote) float64 {
	// Get index of this date (Yahoo dates are returned in UTC).
	idx := indexOf(time.Date(quoteDate.Year(), quoteDate.Month(), quoteDate.Day(), 0, 0, 0, 0, time.UTC), sp500Quotes.Date)
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

// Calculate various metrics about this security.
func (s *Security) CalculateMetrics(sp500Quotes quote.Quote) {
	hasSplits := false
	// Lookup if this security has any stock splits to account for.
	if val, ok := StockSplits[s.Ticker]; ok {
		s.transactions = append(s.transactions, val...)
		hasSplits = true
	}
	// Order the transactions by date, using anonymous function.
	sort.Slice(s.transactions, func(i, j int) bool {
		return s.transactions[i].dateTime.Before(s.transactions[j].dateTime)
	})
	// Get the current info for the given ticker.
	q, err := equity.Get(s.Ticker)
	if err != nil || q == nil {
		return
	}
	// Get current market price.
	s.DetermineCurrentPrice(q)
	// Create a slice to use as a FIFO queue.
	buyQ := make([]Transaction, 0)
	// Iterate thru each transaction (by index, using range creates a copy).
	for idx := 0; idx < len(s.transactions); idx++ {
		// Get a reference to the current txn.
		t := &s.transactions[idx]
		// Ignore cash withdraws/deposits
		if t.ticker == "CASH" {
			continue
		}
		// For buys, increment number of shares.
		if t.action == "Buy" {
			// Add the txn to the buy queue.
			buyQ = append(buyQ, *t)
			// Calculate the return had we bought S&P500 for this transaction.
			spDateOfTxn := s.GetQuoteOfSP500(t.dateTime, sp500Quotes)
			spNow := s.GetQuoteOfSP500(time.Now(), sp500Quotes)
			t.sp500Return = ((spNow - spDateOfTxn) / spDateOfTxn) * 100.0
		} else if t.action == "Sell" {
			// TODO Calculate SP500 theoretical comparison return as done above for BUYs.
			remainingShares := t.shares
			for remainingShares > 0 {
				// Make sure we have buys to cover remaining shares in the sell.
				if len(buyQ) > 0 {
					if buyQ[0].shares > remainingShares {
						// Remaining sell shares are covered by this buy.
						s.RealizedGain += (t.price - buyQ[0].price) * remainingShares
						buyQ[0].shares -= remainingShares
						remainingShares = 0
					} else if buyQ[0].shares == remainingShares {
						// Shares in this buy equal remaining shares, pop the buy off the queue.
						s.RealizedGain += (t.price - buyQ[0].price) * remainingShares
						remainingShares = 0
						buyQ = buyQ[1:]
					} else if buyQ[0].shares < remainingShares {
						// This buy is completely covered by sell, pop it.
						s.RealizedGain += (t.price - buyQ[0].price) * buyQ[0].shares
						remainingShares -= buyQ[0].shares
						buyQ = buyQ[1:]
					}
				} else {
					// Queue is empty, but apparently have more sold shares to account for.
					// There was either a stock split, or re-invested dividends.
					additionalGains := remainingShares * t.price
					s.RealizedGain += additionalGains
					log.Printf("%s is oversold - Adding remaining shares to realized gain (%f shares, total $%f)\n", t.ticker, remainingShares, additionalGains)
					break
				}
			}
		} else if t.action == "Split" {
			// Apply the split to all txns in the buy queue.
			for i := range buyQ {
				buyQ[i].price /= t.shares
				buyQ[i].shares *= t.shares
			}
		}

		// Calculate the theoretical return on this txn, if we held.
		// TODO: Calculate for SELL.
		if t.action == "Buy" {
			// Keep track of the total multiplier to adjust for any stock splits.
			splitMultiple := 1.0
			// Iterate from the current txn to the end, checking for any stock splits.
			if (idx != len(s.transactions)-1) && hasSplits {
				for n := idx + 1; n < len(s.transactions); n++ {
					if s.transactions[n].action == "Split" {
						splitMultiple *= s.transactions[n].shares
					}
				}
			}
			// Calculate the theoretical total return (%) of each txn (using any split multiple from above).
			t.totalReturn = ((s.MarketPrice*t.shares*splitMultiple - t.value) / t.value) * 100.0
			t.excessReturn = t.totalReturn - t.sp500Return
		}
	}

	// Calculate cost bases and unrealized gains with any remaining buy shares in the buy queue.
	for _, txn := range buyQ {
		s.NumShares += txn.shares
		s.TotalCostBasis += txn.shares * txn.price
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
		log.Printf("   -----TXN: %s -----\n", txn.dateTime.Format("2006-01-02"))
		log.Printf("   Num Shares: $%f\n", txn.shares)
		log.Printf("   Price: $%f\n", txn.price)
		log.Printf("   Total Return: $%f\n", txn.totalReturn)
		log.Printf("   SP500 Return: $%f\n", txn.sp500Return)
		log.Printf("   Excess Return: $%f\n", txn.excessReturn)
	}
}
