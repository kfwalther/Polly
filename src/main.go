package main

// List of imported packages
import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"

	"github.com/kfwalther/Polly/src/auth"
	"github.com/kfwalther/Polly/src/finance"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// Program entry point.
func main() {
	ctx := context.Background()
	b, err := ioutil.ReadFile("../credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved auth_token.json.
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets.readonly")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := auth.GetClient(config)

	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}

	// Test spreadsheet: https://docs.google.com/spreadsheets/d/1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms/edit
	// ID of the portfolio spreadsheet
	spreadsheetId := "18H9kc7j-lMBH4TJP4nEiccVU4Dj6Yqnz7RLfOwaXluI"
	// Specify the sheet and columns.
	readRange := "TransactionList!A2:G"
	// valueRendering := "UNFORMATTED_VALUE"
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}

	// Check if we parsed any data from the spreadsheet.
	if len(resp.Values) == 0 {
		fmt.Println("No data found.")
	} else {
		catalogue := finance.NewSecurityCatalogue()
		// Process the imported data to organize it by ticker.
		if (*catalogue).ProcessImport(resp.Values) {
			fmt.Println("Number of transactions processed: " + strconv.Itoa(len(resp.Values)))
		}
	}
	fmt.Println("Successful Completion!")
}
