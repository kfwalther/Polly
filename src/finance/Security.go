package finance

import (
	"fmt"

	"github.com/piquette/finance-go/quote"
)

// Definition of a security to hold the transactions for a particular stock/ETF.
type Security struct {
	id              uint
	ticker          string
	unitCostBasis   float64
	totalCostBasis  float64
	numShares       float64
	realizedGains   float64
	unrealizedGains float64
	transactions    []Transaction
}

// Constructor for a new SecurityCatalogue object, initializing the map.
func NewSecurity(tkr string) *Security {
	var s Security
	s.ticker = tkr
	s.transactions = make([]Transaction, 0)
	return &s
}

func (s *Security) CurrentPrice() float64 {
	q, err := quote.Get(s.ticker)
	if err != nil || q == nil {
		return 0.0
	}
	return q.Ask
}

// Calculate various metrics about this security.
func (s *Security) CalculateMetrics() {
	// Create a slice to use as a FIFO queue.
	q := make([]Transaction, 0)
	// Iterate thru each transaction.
	for _, t := range s.transactions {
		// For buys, increment number of shares.
		if t.action == "Buy" {
			// Add the txn to the queue.
			q = append(q, t)
		} else if t.action == "Sell" {
			remainingShares := t.shares
			for remainingShares > 0 {
				// Make sure we have buys to cover remaining shares in the sell.
				if len(q) > 0 {
					if q[0].shares > remainingShares {
						// Remaining sell shares are covered by this buy.
						s.realizedGains += (t.price - q[0].price) * remainingShares
						q[0].shares -= remainingShares
						remainingShares = 0
					} else if q[0].shares == remainingShares {
						// Shares in this buy equal remaining shares, pop the buy off the queue.
						s.realizedGains += (t.price - q[0].price) * remainingShares
						remainingShares = 0
						q = q[1:]
					} else if q[0].shares < remainingShares {
						// This buy is completely covered by sell, pop it.
						s.realizedGains += (t.price - q[0].price) * q[0].shares
						remainingShares -= q[0].shares
						q = q[1:]
					}
				} else {
					// Queue is empty, but apparently have more sold shares to account for. ERROR!
					fmt.Printf("Security %s is oversold! (%s, %f)\n", t.ticker, t.dateTime, t.shares)
					break
				}
			}
		}
	}

	// Calculate cost bases and unrealized gains with any remaining buy shares in the queue.
	for _, txn := range q {
		s.numShares += txn.shares
		s.totalCostBasis += txn.shares * txn.price
	}
	// Current market price
	curPrice := s.CurrentPrice()
	// Unit cost basis
	s.unitCostBasis = s.totalCostBasis / s.numShares
	// Unrealized gain
	s.unrealizedGains = (s.CurrentPrice() * s.numShares) - s.totalCostBasis

	fmt.Printf("%s Market Price: %f\n", s.ticker, curPrice)
	fmt.Printf("%s Number of Shares: %f\n", s.ticker, s.numShares)
	fmt.Printf("%s Unit Cost Basis: %f\n", s.ticker, s.unitCostBasis)
	fmt.Printf("%s Total Cost Basis: %f\n", s.ticker, s.totalCostBasis)
	fmt.Printf("%s Unrealized Gains: %f\n", s.ticker, s.unrealizedGains)
	fmt.Printf("%s Realized Gains: %f\n", s.ticker, s.realizedGains)
	fmt.Println("---------------------------------")
}
