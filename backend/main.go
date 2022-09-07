package main

// List of imported packages
import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kfwalther/Polly/backend/auth"
	"github.com/kfwalther/Polly/backend/finance"
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

	// Get the spreadsheet ID from the input file.
	portfolioIdFile := "../portfolio-sheet-id.txt"
	spreadsheetId, err := ioutil.ReadFile(portfolioIdFile)
	if err != nil {
		log.Fatalf("Can't read file (%s): %s", portfolioIdFile, err)
	}

	// Specify the sheet and columns.
	readRange := "TransactionList!A2:G"
	resp, err := srv.Spreadsheets.Values.Get(string(spreadsheetId), readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}

	// Setup the Go web server.
	router := gin.Default()
	// router.GET("/api/securities", taskController.GetTasks)
	router.GET("/securities", SecurityListHandler)

	// Check if we parsed any data from the spreadsheet.
	if len(resp.Values) == 0 {
		fmt.Println("No data found.")
	} else {
		catalogue := finance.NewSecurityCatalogue()
		// Process the imported data to organize it by ticker.
		(*catalogue).ProcessImport(resp.Values)
		fmt.Println("Number of transactions processed: " + strconv.Itoa(len(resp.Values)))
		// Calculate metrics for each stock.
		catalogue.Calculate()
	}
	fmt.Println("Successful Completion!")
}

// JokeHandler retrieves a list of available jokes
func SecurityListHandler(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK)
}
