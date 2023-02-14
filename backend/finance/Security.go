package finance

import (
	"log"
	"sort"

	"github.com/piquette/finance-go/equity"
)

// Definition of a security to hold the transactions for a particular stock/ETF.
type Security struct {
	id              uint
	Ticker          string  `json:"ticker"`
	MarketPrice     float64 `json:"marketPrice"`
	MarketValue     float64 `json:"marketValue"`
	UnitCostBasis   float64 `json:"unitCostBasis"`
	TotalCostBasis  float64 `json:"totalCostBasis"`
	NumShares       float64 `json:"numShares"`
	RealizedGains   float64 `json:"realizedGains"`
	UnrealizedGains float64 `json:"unrealizedGains"`
	transactions    []Transaction
}

// Constructor for a new SecurityCatalogue object, initializing the map.
func NewSecurity(tkr string) *Security {
	var s Security
	s.Ticker = tkr
	s.transactions = make([]Transaction, 0)
	return &s
}

// Grab the current market price of this ticker symbol, from the web.
func (s *Security) CurrentPrice() float64 {
	q, err := equity.Get(s.Ticker)
	if err != nil || q == nil {
		return 0.0
	}
	// Based on the market's current state, grab the proper price.
	if q.MarketState == "PRE" && q.PreMarketPrice > 0.0 {
		return q.PreMarketPrice
	} else if q.MarketState == "POST" && q.PostMarketPrice > 0.0 {
		return q.PostMarketPrice
	} else {
		return q.RegularMarketPrice
	}
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
						s.RealizedGains += (t.price - buyQ[0].price) * remainingShares
						buyQ[0].shares -= remainingShares
						remainingShares = 0
					} else if buyQ[0].shares == remainingShares {
						// Shares in this buy equal remaining shares, pop the buy off the queue.
						s.RealizedGains += (t.price - buyQ[0].price) * remainingShares
						remainingShares = 0
						buyQ = buyQ[1:]
					} else if buyQ[0].shares < remainingShares {
						// This buy is completely covered by sell, pop it.
						s.RealizedGains += (t.price - buyQ[0].price) * buyQ[0].shares
						remainingShares -= buyQ[0].shares
						buyQ = buyQ[1:]
					}
				} else {
					// Queue is empty, but apparently have more sold shares to account for.
					// There was either a stock split, or re-invested dividends.
					additionalGains := remainingShares * t.price
					s.RealizedGains += additionalGains
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
	// Current market price
	s.MarketPrice = s.CurrentPrice()
	if s.NumShares > 0 {
		// Unit cost basis
		s.UnitCostBasis = s.TotalCostBasis / s.NumShares
		// Total market value
		s.MarketValue = s.MarketPrice * s.NumShares
		// Unrealized gain
		s.UnrealizedGains = s.MarketValue - s.TotalCostBasis
	}
}

func (s *Security) DisplayMetrics() {
	log.Println("---------------------------------")
	log.Printf("%s Market Price: $%f\n", s.Ticker, s.MarketPrice)
	log.Printf("%s Number of Shares: %f\n", s.Ticker, s.NumShares)
	log.Printf("%s Market Value: %f\n", s.Ticker, s.MarketValue)
	log.Printf("%s Unit Cost Basis: $%f\n", s.Ticker, s.UnitCostBasis)
	log.Printf("%s Total Cost Basis: $%f\n", s.Ticker, s.TotalCostBasis)
	log.Printf("%s Unrealized Gains: $%f\n", s.Ticker, s.UnrealizedGains)
	log.Printf("%s Realized Gains: $%f\n", s.Ticker, s.RealizedGains)
}
