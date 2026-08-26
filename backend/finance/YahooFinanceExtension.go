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

func cleanTickerList(tickers []string) []string {
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
	return cleanTickers
}

// GetTickerData retrieves extended information for all requested symbols. It
// first tries one Python invocation, then retries only missing symbols in
// configured-sized chunks if Yahoo rejects or partially fulfills the request.
func (ext *YahooFinanceExtension) GetTickerData(tickers []string) map[string]interface{} {
	cleanTickers := cleanTickerList(tickers)
	if len(cleanTickers) == 0 {
		return map[string]interface{}{}
	}

	fetch := func(symbols []string) (map[string]interface{}, error) {
		query := strings.Join(symbols, ",")
		retrySecs := 1
		var lastErr error
		for attempt := 0; attempt < 4; attempt++ {
			output, err := ext.runCommand(ext.yfinScript, query)
			if err == nil {
				var tickerData map[string]interface{}
				if unmarshalErr := json.Unmarshal(output, &tickerData); unmarshalErr == nil && tickerData != nil {
					log.Printf("Successfully pulled yfinance data for ticker(s): %s", query)
					return tickerData, nil
				} else if unmarshalErr != nil {
					err = fmt.Errorf("unmarshal yfinance data: %w", unmarshalErr)
				} else {
					err = fmt.Errorf("yfinance returned no ticker data")
				}
			}
			lastErr = err
			if attempt < 3 {
				log.Printf("Warning: could not retrieve extended stock info (%s) via yfinance - will retry: %v", query, err)
				ext.sleep(time.Duration(retrySecs) * time.Second)
				retrySecs *= 2
			}
		}
		return nil, lastErr
	}

	allData, err := fetch(cleanTickers)
	if err != nil {
		log.Printf("WARNING: all-ticker yfinance info request failed; retrying in chunks: %v", err)
		allData = make(map[string]interface{})
	}

	missing := make([]string, 0)
	for _, ticker := range cleanTickers {
		if _, ok := allData[ticker]; !ok {
			missing = append(missing, ticker)
		}
	}
	if len(missing) == 0 {
		return allData
	}
	if len(missing) < len(cleanTickers) {
		log.Printf("WARNING: yfinance returned data for %d/%d tickers; retrying %d missing symbols in chunks", len(cleanTickers)-len(missing), len(cleanTickers), len(missing))
	}

	chunkSize := ext.historicalConfig.ChunkSize
	if chunkSize <= 0 {
		chunkSize = DefaultHistoricalDownloadConfig.ChunkSize
	}
	for start := 0; start < len(missing); start += chunkSize {
		if start > 0 && ext.historicalConfig.WaitBetweenChunks > 0 {
			ext.sleep(ext.historicalConfig.WaitBetweenChunks)
		}
		end := start + chunkSize
		if end > len(missing) {
			end = len(missing)
		}
		chunk := missing[start:end]
		chunkData, err := fetch(chunk)
		if err != nil {
			log.Printf("WARNING: could not retrieve yfinance info for ticker chunk %s: %v", strings.Join(chunk, ","), err)
			continue
		}
		for ticker, tickerData := range chunkData {
			allData[ticker] = tickerData
		}
	}
	return allData
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
	cleanTickers := cleanTickerList(tickers)
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
