package finance

import (
	"encoding/json"
	"log"
	"os/exec"
	"time"

	"github.com/kfwalther/Polly/backend/data"
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

// Call the python yFinance script to query data for the given tickers (accepts single ticker or comma-separated list of tickers).
func (ext *YahooFinanceExtension) GetTickerData(tickers string) *map[string]interface{} {
	// Retry counter (1,2,4,8 seconds).
	retrySecs := 1
	// Scrape the Yahoo finance site to get extra data about some stocks.
	cmd := exec.Command("python", ext.yfinScript, tickers)
	output, err := cmd.CombinedOutput()
	// If failure, retry a few times, waiting up to 8 seconds.
	for err != nil && retrySecs < 8 {
		log.Printf("Warning: could not retrieve extended stock info (%s) via yfinance - will retry: %v", tickers, err)
		time.Sleep(time.Duration(retrySecs) * time.Second)
		retrySecs += retrySecs
		output, err = cmd.CombinedOutput()
	}
	if output == nil {
		log.Fatalf("Error: could not retrieve extended stock info (%s) via yfinance: %v", tickers, err)
	}
	// Parse the JSON string into a map
	var myDict map[string]interface{}
	err = json.Unmarshal(output, &myDict)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	log.Printf("Successfully pulled yfinance data for ticker(s): %s", tickers)
	return &myDict
}

// GetHistoricalData retrieves historical price data for a ticker
func (ext *YahooFinanceExtension) GetHistoricalData(ticker string, startDate string, endDate string) (*data.Quote, error) {
	cmd := exec.Command("python", ext.yfinScript, ticker, startDate, endDate)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	// Parse the JSON output into a Quote structure
	var quote data.Quote
	err = json.Unmarshal(output, &quote)
	if err != nil {
		return nil, err
	}
	return &quote, nil
}
