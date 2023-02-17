package finance

import (
	"errors"
	"log"
	"sort"

	"github.com/piquette/finance-go/equity"
)

// Definition of a security to hold the transactions for a particular stock/ETF.
type Security struct {
	id                  uint
	Ticker              string  `json:"ticker"`
	SecurityType        string  `json:"securityType"`
	MarketPrice         float64 `json:"marketPrice"`
	MarketValue         float64 `json:"marketValue"`
	DailyGain           float64 `json:"dailyGain"`
	DailyGainPercentage float64 `json:"dailyGainPercentage"`
	UnitCostBasis       float64 `json:"unitCostBasis"`
	TotalCostBasis      float64 `json:"totalCostBasis"`
	NumShares           float64 `json:"numShares"`
	RealizedGain        float64 `json:"realizedGain"`
	UnrealizedGain      float64 `json:"unrealizedGain"`
	TotalGain           float64 `json:"totalGain"`
	transactions        []Transaction
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

// Grab the current market price of this ticker symbol, from the web.
func (s *Security) DetermineCurrentPrice() {
	// Get the current info for the given ticker.
	q, err := equity.Get(s.Ticker)
	if err != nil || q == nil {
		return
	}
	// Based on the market's current state, grab the proper current quoted price.
	if q.MarketState == "PRE" && q.PreMarketPrice > 0.0 {
		s.MarketPrice = q.PreMarketPrice
	} else if q.MarketState == "POST" && q.PostMarketPrice > 0.0 {
		s.MarketPrice = q.PostMarketPrice
	} else {
		s.MarketPrice = q.RegularMarketPrice
	}
	// Calculate the total 1-day gain/loss for this stock.
	s.DailyGain = q.RegularMarketChange * s.NumShares
	s.DailyGainPercentage = q.RegularMarketChangePercent
}

// Calculate various metrics about this security.
func (s *Security) CalculateMetrics() {
	// Lookup if this security has any stock splits to account for.
	if val, ok := StockSplits[s.Ticker]; ok {
		s.transactions = append(s.transactions, val...)
	}
	// Order the transactions by date, using anonymous function.
	sort.Slice(s.transactions, func(i, j int) bool {
		return s.transactions[i].dateTime.Before(s.transactions[j].dateTime)
	})
	// Create a slice to use as a FIFO queue.
	buyQ := make([]Transaction, 0)
	// Iterate thru each transaction.
	for _, t := range s.transactions {
		// Ignore cash withdraws/deposits
		if t.ticker == "CASH" {
			continue
		}
		// For buys, increment number of shares.
		if t.action == "Buy" {
			// Add the txn to the buy queue.
			buyQ = append(buyQ, t)
		} else if t.action == "Sell" {
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
	}

	// Calculate cost bases and unrealized gains with any remaining buy shares in the buy queue.
	for _, txn := range buyQ {
		s.NumShares += txn.shares
		s.TotalCostBasis += txn.shares * txn.price
	}
	// Request the current market price, and calculate daily gain values.
	s.DetermineCurrentPrice()
	if s.NumShares > 0 {
		// Unit cost basis
		s.UnitCostBasis = s.TotalCostBasis / s.NumShares
		// Total market value
		s.MarketValue = s.MarketPrice * s.NumShares
		// Unrealized gain
		s.UnrealizedGain = s.MarketValue - s.TotalCostBasis
	}
	// Get the total gain.
	s.TotalGain = s.UnrealizedGain + s.RealizedGain
}

func (s *Security) DisplayMetrics() {
	log.Println("---------------------------------")
	log.Printf("%s Market Price: $%f\n", s.Ticker, s.MarketPrice)
	log.Printf("%s Number of Shares: %f\n", s.Ticker, s.NumShares)
	log.Printf("%s Market Value: %f\n", s.Ticker, s.MarketValue)
	log.Printf("%s Daily Gain: $%f\n", s.Ticker, s.DailyGain)
	log.Printf("%s Daily Gain Percent: %f\n", s.Ticker, s.DailyGainPercentage)
	log.Printf("%s Unit Cost Basis: $%f\n", s.Ticker, s.UnitCostBasis)
	log.Printf("%s Total Cost Basis: $%f\n", s.Ticker, s.TotalCostBasis)
	log.Printf("%s Unrealized Gains: $%f\n", s.Ticker, s.UnrealizedGain)
	log.Printf("%s Realized Gains: $%f\n", s.Ticker, s.RealizedGain)
}
