package finance

import (
	"fmt"
	"github.com/kfwalther/Polly/backend/data"
	"strings"
	"testing"
	"time"
)

func TestGetHistoricalDataBatchChunksTickersAndWaitsBetweenChunks(t *testing.T) {
	ext := NewYahooFinanceExtension("yahooFinanceHelper.py", HistoricalDownloadConfig{
		ChunkSize:         2,
		WaitBetweenChunks: 25 * time.Millisecond,
	})
	var calls [][]string
	var waits []time.Duration
	ext.runCommand = func(script string, args ...string) ([]byte, error) {
		calls = append(calls, append([]string(nil), args...))
		return historicalQuotesJSON(strings.Split(args[1], ",")), nil
	}
	ext.sleep = func(wait time.Duration) {
		waits = append(waits, wait)
	}

	quotes, err := ext.GetHistoricalDataBatch([]string{"AAPL", "MSFT", "GOOG", "AMZN", "NVDA"}, "2024-01-01", "2024-02-01")
	if err != nil {
		t.Fatalf("GetHistoricalDataBatch() error = %v", err)
	}

	if len(quotes) != 5 {
		t.Fatalf("quotes = %d, want 5", len(quotes))
	}
	if len(calls) != 3 {
		t.Fatalf("command calls = %d, want 3", len(calls))
	}
	wantTickerArgs := []string{"AAPL,MSFT", "GOOG,AMZN", "NVDA"}
	for i, want := range wantTickerArgs {
		if calls[i][0] != "--historical-batch" || calls[i][1] != want || calls[i][2] != "2024-01-01" || calls[i][3] != "2024-02-01" {
			t.Fatalf("call %d = %#v, want batch arguments for %q", i, calls[i], want)
		}
	}
	if len(waits) != 2 || waits[0] != 25*time.Millisecond || waits[1] != 25*time.Millisecond {
		t.Fatalf("waits = %#v, want two 25ms waits", waits)
	}
	if quotes["NVDA"].Symbol != "NVDA" {
		t.Fatalf("NVDA symbol = %q, want NVDA", quotes["NVDA"].Symbol)
	}
}

func TestGetHistoricalDataBatchUsesDefaultChunkSizeForInvalidConfig(t *testing.T) {
	ext := NewYahooFinanceExtension("yahooFinanceHelper.py", HistoricalDownloadConfig{ChunkSize: 0})
	if ext.historicalConfig.ChunkSize != DefaultHistoricalDownloadConfig.ChunkSize {
		t.Fatalf("chunk size = %d, want default %d", ext.historicalConfig.ChunkSize, DefaultHistoricalDownloadConfig.ChunkSize)
	}
}

func historicalQuotesJSON(symbols []string) []byte {
	parts := make([]string, 0, len(symbols))
	for _, symbol := range symbols {
		parts = append(parts, fmt.Sprintf(`%q:{"symbol":%q,"date":[],"open":[],"high":[],"low":[],"close":[],"volume":[]}`, symbol, symbol))
	}
	return []byte("{" + strings.Join(parts, ",") + "}")
}

func TestTrimHistoricalQuoteToRangeKeepsOnlyRequestedDates(t *testing.T) {
	quote := data.Quote{
		Symbol: "AAPL",
		Date:   []time.Time{time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)},
		Close:  []float64{10, 20, 30},
	}

	trimmed := trimHistoricalQuoteToRange(quote, "2024-01-02", "2024-01-03")
	if len(trimmed.Date) != 1 || !trimmed.Date[0].Equal(quote.Date[1]) {
		t.Fatalf("dates = %#v, want only 2024-01-02", trimmed.Date)
	}
	if len(trimmed.Close) != 1 || trimmed.Close[0] != 20 {
		t.Fatalf("close = %#v, want [20]", trimmed.Close)
	}
}

func TestGetHistoricalDataNormalizesDailyTimestampsToUTCMidnight(t *testing.T) {
	ext := NewYahooFinanceExtension("yahooFinanceHelper.py", HistoricalDownloadConfig{})
	ext.runCommand = func(script string, args ...string) ([]byte, error) {
		return []byte(historicalQuoteJSONWithDates()), nil
	}

	quote, err := ext.GetHistoricalData("AAPL", "2024-01-01", "2024-01-04")
	if err != nil {
		t.Fatalf("GetHistoricalData() error = %v", err)
	}
	assertUTCMidnightDates(t, quote.Date)
}

func TestGetHistoricalDataBatchNormalizesDailyTimestampsToUTCMidnight(t *testing.T) {
	ext := NewYahooFinanceExtension("yahooFinanceHelper.py", HistoricalDownloadConfig{ChunkSize: 2})
	ext.runCommand = func(script string, args ...string) ([]byte, error) {
		return []byte(`{"AAPL":` + historicalQuoteJSONWithDates() + `}`), nil
	}

	quotes, err := ext.GetHistoricalDataBatch([]string{"AAPL"}, "2024-01-01", "2024-01-04")
	if err != nil {
		t.Fatalf("GetHistoricalDataBatch() error = %v", err)
	}
	assertUTCMidnightDates(t, quotes["AAPL"].Date)
}

func historicalQuoteJSONWithDates() string {
	return `{"symbol":"AAPL","date":["2024-01-02T00:00:00-05:00","2024-01-03T00:00:00+09:00"],"open":[1,2],"high":[1,2],"low":[1,2],"close":[1,2],"volume":[1,2]}`
}

func assertUTCMidnightDates(t *testing.T, dates []time.Time) {
	t.Helper()
	want := []time.Time{
		time.Date(2024, time.January, 2, 0, 0, 0, 0, time.UTC),
		time.Date(2024, time.January, 3, 0, 0, 0, 0, time.UTC),
	}
	if len(dates) != len(want) {
		t.Fatalf("dates = %#v, want %#v", dates, want)
	}
	for i := range want {
		if !dates[i].Equal(want[i]) || dates[i].Location() != time.UTC {
			t.Fatalf("date %d = %v (%s), want UTC midnight %v", i, dates[i], dates[i].Location(), want[i])
		}
	}
}
