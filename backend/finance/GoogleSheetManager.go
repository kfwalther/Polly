package finance

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// A class to support calling yfinance python package to query extended data about individual symbols.
type GoogleSheetManager struct {
	sheetIds                []string
	service                 *sheets.Service
	growthStocksSpreadsheet *sheets.Spreadsheet
}

// Constructor for a new GoogleSheetManager.
func NewGoogleSheetManager(httpClient *http.Client, ctx *context.Context, sheetIdFile string) *GoogleSheetManager {
	var mgr GoogleSheetManager
	var err error
	// Initialize the service to interact with Google sheets.
	if mgr.service, err = sheets.NewService(*ctx, option.WithHTTPClient(httpClient)); err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}
	var spreadsheetIds []byte
	// Read the sheet IDs from the input file.
	if spreadsheetIds, err = os.ReadFile(sheetIdFile); err != nil {
		log.Fatalf("Can't read file (%s): %s", sheetIdFile, err)
	}
	// Get the IDs for both sheets to reference.
	ids := strings.Split(string(spreadsheetIds), "\r\n")
	// Verify we got two IDs.
	if len(ids) < 2 {
		log.Fatalf("Spreadsheet ID file (%s) only has %v IDs. Expecting 2.", sheetIdFile, len(ids))
	}
	mgr.sheetIds = ids
	// Grab the growth stock spreadsheet now.
	mgr.growthStocksSpreadsheet, err = mgr.service.Spreadsheets.Get(ids[1]).Do()
	if err != nil {
		log.Fatalf("Unable to get spreadsheet (ID: %s). If token error, "+
			"delete auth_token file and retry: %v", ids[1], err)
	}
	return &mgr
}

func (mgr *GoogleSheetManager) GetTransactionData(equityType string) *sheets.ValueRange {
	resp := mgr.getSheetData(mgr.sheetIds[0], equityType+"!A2:G")
	// Check if we parsed any data from the spreadsheet.
	if len(resp.Values) == 0 {
		log.Fatalf("No transaction data found in %s spreadsheet... Exiting!", equityType)
	}
	return resp
}

func (mgr *GoogleSheetManager) GetAllRevenueData(ticker string) *sheets.ValueRange {
	// Check if sheet exists for this ticker.
	if mgr.sheetExists(mgr.sheetIds[1], ticker) {
		resp := mgr.getSheetData(mgr.sheetIds[1], ticker+"!A1:28")
		// Check if we parsed any data from the spreadsheet.
		if len(resp.Values) > 0 {
			return resp
		}
		log.Printf("WARNING: No revenue data for %s found in spreadsheet!", ticker)
	}
	return nil
}

func (mgr *GoogleSheetManager) sheetExists(sheetId string, sheetName string) bool {
	for _, sheet := range mgr.growthStocksSpreadsheet.Sheets {
		if sheet.Properties.Title == sheetName {
			return true
		}
	}
	return false
}

func (mgr *GoogleSheetManager) getSheetData(sheetId string, sheetRange string) *sheets.ValueRange {
	resp, err := mgr.service.Spreadsheets.Values.Get(sheetId, sheetRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from spreadsheet (ID: %s). If token error, "+
			"delete auth_token file and retry: %v", sheetId, err)
	}
	return resp
}
