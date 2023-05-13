package finance

import (
	"encoding/json"
	"log"
	"os/exec"
)

// A class to support calling yfinance python package to query extended data about individual symbols.
type YahooFinanceExtension struct {
	yfinScript string
}

// Constructor for a new YahooFinanceExtension.
func NewYahooFinanceExtension(pyScript string) *YahooFinanceExtension {
	var ext YahooFinanceExtension
	ext.yfinScript = pyScript
	return &ext
}

func (ext *YahooFinanceExtension) GetTickerData(ticker string) *map[string]interface{} {
	// Scrape the Yahoo finance site to get extra data about some stocks.
	cmd := exec.Command("python", ext.yfinScript, ticker)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Error: could not retrieve extended stock info via yfinance: %v", err)
	}
	// Parse the JSON string into a map
	var myDict map[string]interface{}
	err = json.Unmarshal(output, &myDict)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	return &myDict
}
