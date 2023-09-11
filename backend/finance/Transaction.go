package finance

import (
	"log"
	"strconv"
	"strings"
	"time"
)

// Definition of a transaction containing metadata and calculated metrics about the trade.
type Transaction struct {
	id           uint
	Ticker       string    `json:"ticker"`
	DateTime     time.Time `json:"dateTime"`
	Action       string    `json:"action"`
	Shares       float64   `json:"shares"`
	Price        float64   `json:"price"`
	Value        float64   `json:"value"`
	TotalReturn  float64   `json:"totalReturn"`
	Sp500Return  float64   `json:"sp500Return"`
	ExcessReturn float64   `json:"excessReturn"`
}

func NormalizeAmerican(num string) string {
	return strings.Replace(num, ",", "", -1)
}

// Constructor for a new SecurityCatalogue object, initializing the map.
func NewTransaction(dTime string, tkr string, act string, numShares string, txnPrice string) *Transaction {
	var t Transaction
	var err error
	// Attempt to parse each field into appropriate type in the object.
	if t.DateTime, err = time.Parse("2006-01-02", dTime); err != nil {
		log.Fatalf("Unable to parse date field from transaction: %v", dTime)
	}
	// Place the transaction times at midday, so we can order stock splits at market open first.
	t.DateTime = t.DateTime.Add(time.Hour * 12)
	t.Ticker = tkr
	// Verify the action falls into accepted values.
	if act != "Buy" && act != "Sell" && act != "Deposit" && act != "Withdraw" {
		log.Fatalf("Unable to parse action field from transaction: %v", act)
	}
	t.Action = act
	if t.Shares, err = strconv.ParseFloat(NormalizeAmerican(numShares), 64); err != nil {
		log.Fatalf("Unable to parse shares field from transaction: %v", numShares)
	}
	if txnPrice != "" {
		if t.Price, err = strconv.ParseFloat(NormalizeAmerican(txnPrice), 64); err != nil {
			log.Fatalf("Unable to parse price field from transaction: %v|", txnPrice)
		}
	} else {
		t.Price = 1
	}
	// Calculate the total amount involved with this transaction.
	t.Value = t.Shares * t.Price
	return &t
}
