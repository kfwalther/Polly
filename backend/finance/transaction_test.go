package finance

import (
	"testing"
	"time"
)

func TestNormalizeAmerican(t *testing.T) {
	if got := NormalizeAmerican("1,234,567.89"); got != "1234567.89" {
		t.Fatalf("NormalizeAmerican() = %q, want %q", got, "1234567.89")
	}
}

func TestNewTransactionParsesSheetValues(t *testing.T) {
	txn := NewTransaction("2/29/2024", "NVDA", "Buy", "1,234.5", "875.25")

	if txn.Ticker != "NVDA" || txn.Action != "Buy" {
		t.Fatalf("transaction identity = (%q, %q), want (NVDA, Buy)", txn.Ticker, txn.Action)
	}
	if txn.Shares != 1234.5 || txn.Price != 875.25 {
		t.Fatalf("transaction amounts = (%v, %v), want (1234.5, 875.25)", txn.Shares, txn.Price)
	}
	if txn.Value != 1080496.125 {
		t.Fatalf("Value = %v, want 1080496.125", txn.Value)
	}

	wantDate := time.Date(2024, time.February, 29, 12, 0, 0, 0, time.UTC)
	if !txn.DateTime.Equal(wantDate) {
		t.Fatalf("DateTime = %v, want %v", txn.DateTime, wantDate)
	}
}

func TestNewTransactionDefaultsBlankPriceToOne(t *testing.T) {
	txn := NewTransaction("1/1/2024", "CASH", "Deposit", "2,500", "")

	if txn.Price != 1 || txn.Value != 2500 {
		t.Fatalf("blank-price transaction = price %v, value %v; want price 1, value 2500", txn.Price, txn.Value)
	}
}
