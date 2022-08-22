package main

// Definition of a security to hold the transactions for a particular stock/ETF.
type Security struct {
	id           uint
	ticker       string
	transactions []Transaction
}

// Constructor for a new SecurityCatalogue object, initializing the map.
func NewSecurity(tkr string) *Security {
	var s Security
	s.ticker = tkr
	s.transactions = make([]Transaction, 0)
	return &s
}
