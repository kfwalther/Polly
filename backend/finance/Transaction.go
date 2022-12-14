package finance

import (
	"log"
	"strconv"
	"strings"
	"time"
)

// Definition of a transaction containing metadata about the trade.
type Transaction struct {
	id       uint
	ticker   string
	dateTime time.Time
	action   string
	shares   float64
	price    float64
}

func NormalizeAmerican(num string) string {
	return strings.Replace(num, ",", "", -1)
}

// Constructor for a new SecurityCatalogue object, initializing the map.
func NewTransaction(dTime string, tkr string, act string, numShares string, txnPrice string) *Transaction {
	var t Transaction
	var err error
	// Attempt to parse each field into appropriate type in the object.
	if t.dateTime, err = time.Parse("2006-01-02", dTime); err != nil {
		log.Fatalf("Unable to parse date field from transaction: %v", dTime)
	}
	t.ticker = tkr
	t.action = act
	if t.shares, err = strconv.ParseFloat(NormalizeAmerican(numShares), 64); err != nil {
		log.Fatalf("Unable to parse shares field from transaction: %v", numShares)
	}
	if txnPrice != "" {
		if t.price, err = strconv.ParseFloat(NormalizeAmerican(txnPrice), 64); err != nil {
			log.Fatalf("Unable to parse price field from transaction: %v|", txnPrice)
		}
	} else {
		t.price = 1
	}
	return &t
}
