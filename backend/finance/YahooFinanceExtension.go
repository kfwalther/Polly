package finance

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/kfwalther/Polly/backend/data"
)

type HistoricalDownloadConfig struct {
	ChunkSize         int
	WaitBetweenChunks time.Duration
}

var DefaultHistoricalDownloadConfig = HistoricalDownloadConfig{
	ChunkSize:         25,
	WaitBetweenChunks: 5 * time.Second,
}

type pythonCommandRunner func(script string, args ...string) ([]byte, error)

type YahooFinanceExtension struct {
	yfinScript       string
	historicalConfig HistoricalDownloadConfig
	runCommand       pythonCommandRunner
	sleep            func(time.Duration)
}

func NewYahooFinanceExtension(pyScript string, config HistoricalDownloadConfig) *YahooFinanceExtension {
	if config.ChunkSize <= 0 {
		config.ChunkSize = DefaultHistoricalDownloadConfig.ChunkSize
	}
	if config.WaitBetweenChunks < 0 {
		config.WaitBetweenChunks = DefaultHistoricalDownloadConfig.WaitBetweenChunks
	}
	return &YahooFinanceExtension{
		yfinScript:       pyScript,
		historicalConfig: config,
		runCommand:       runPythonCommand,
		sleep:            time.Sleep,
	}
}

func runPythonCommand(script string, args ...string) ([]byte, error) {
	commandArgs := append([]string{script}, args...)
	return exec.Command("python", commandArgs...).CombinedOutput()
}

// Call the python yFinance script to query data for the given tickers (accepts single ticker or comma-separated list of tickers).
func (ext *YahooFinanceExtension) GetTickerData(tickers string) *map[string]interface{} {
	retrySecs := 1
	output, err := ext.runCommand(ext.yfinScript, tickers)
	for err != nil && retrySecs < 8 {
		log.Printf("Warning: could not retrieve extended stock info (%s) via yfinance - will retry: %v", tickers, err)
		time.Sleep(time.Duration(retrySecs) * time.Second)
		retrySecs += retrySecs
		output, err = ext.runCommand(ext.yfinScript, tickers)
	}
	if err != nil {
		log.Printf("Error: could not retrieve extended stock info (%s) via yfinance: %v", tickers, err)
	}
	var myDict map[string]interface{}
	err = json.Unmarshal(output, &myDict)
	if err != nil {
		log.Printf("Error unmarshaling yfinance data: %v", err)
	}
	log.Printf("Successfully pulled yfinance data for ticker(s): %s", tickers)
	return &myDict
}

func normalizeHistoricalQuoteDates(quote data.Quote) data.Quote {
	for i, date := range quote.Date {
		quote.Date[i] = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	}
	return quote
}

// GetHistoricalDataBatch retrieves daily historical data for multiple symbols.
// Symbols are split into configured chunks and requests are serialized with a
// configurable pause between chunks.
func (ext *YahooFinanceExtension) GetHistoricalDataBatch(tickers []string, startDate string, endDate string) (map[string]data.Quote, error) {
	cleanTickers := make([]string, 0, len(tickers))
	seen := make(map[string]struct{}, len(tickers))
	for _, ticker := range tickers {
		ticker = strings.TrimSpace(ticker)
		if ticker == "" {
			continue
		}
		if _, ok := seen[ticker]; ok {
			continue
		}
		seen[ticker] = struct{}{}
		cleanTickers = append(cleanTickers, ticker)
	}

	quotes := make(map[string]data.Quote)
	for start := 0; start < len(cleanTickers); start += ext.historicalConfig.ChunkSize {
		if start > 0 && ext.historicalConfig.WaitBetweenChunks > 0 {
			ext.sleep(ext.historicalConfig.WaitBetweenChunks)
		}
		end := start + ext.historicalConfig.ChunkSize
		if end > len(cleanTickers) {
			end = len(cleanTickers)
		}
		chunk := cleanTickers[start:end]
		output, err := ext.runCommand(ext.yfinScript, "--historical-batch", strings.Join(chunk, ","), startDate, endDate)
		if err != nil {
			return nil, fmt.Errorf("retrieve historical data for %s: %w; output: %s", strings.Join(chunk, ","), err, strings.TrimSpace(string(output)))
		}
		var chunkQuotes map[string]data.Quote
		if err := json.Unmarshal(output, &chunkQuotes); err != nil {
			return nil, fmt.Errorf("unmarshal historical data for %s: %w", strings.Join(chunk, ","), err)
		}
		for ticker, quote := range chunkQuotes {
			quotes[ticker] = normalizeHistoricalQuoteDates(quote)
		}
	}
	return quotes, nil
}

func (ext *YahooFinanceExtension) GetHistoricalData(ticker string, startDate string, endDate string) (*data.Quote, error) {
	output, err := ext.runCommand(ext.yfinScript, ticker, startDate, endDate)
	if err != nil {
		return nil, err
	}
	var quote data.Quote
	err = json.Unmarshal(output, &quote)
	if err != nil {
		return nil, err
	}
	quote = normalizeHistoricalQuoteDates(quote)
	return &quote, nil
}
