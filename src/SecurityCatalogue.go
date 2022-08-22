package main

import "fmt"

// Definition of a security catalogue to house a portfolio of stock/ETF info in a map.
type SecurityCatalogue struct {
	id         uint
	securities map[string]Security
}

// Constructor for a new SecurityCatalogue object, initializing the map.
func NewSecurityCatalogue() *SecurityCatalogue {
	var sc SecurityCatalogue
	sc.securities = make(map[string]Security)
	return &sc
}

// Method to process the imported data, by creating a new [Transaction] for
// each row of data, and inserting it into the appropriate [Security] object
// within the catalogue.
func (sc *SecurityCatalogue) processImport(txnData [][]interface{}) bool {
	// TODO: Would be good to perform column name verification before processing.
	// Iterate thru each row of data.
	for _, row := range txnData {
		// Make sure we haven't reached the end of the data.
		if row[0] != "" {
			fmt.Printf("%s, %s, %s, %s, %s\n", row[0], row[1], row[2], row[3], row[4])
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
					sc.securities[txn.ticker] = *sec
				}
			}
		}
	}
	return true
}
