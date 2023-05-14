package main

// List of imported packages
import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/kfwalther/Polly/backend/auth"
	"github.com/kfwalther/Polly/backend/controllers"
	"github.com/kfwalther/Polly/backend/data"
	"github.com/kfwalther/Polly/backend/finance"
	"golang.org/x/oauth2/google"
)

// Program entry point.
func main() {
	ctx := context.Background()
	b, err := os.ReadFile("../credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved auth_token.json.
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets.readonly")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := auth.GetClient(config)

	// Get the spreadsheet IDs from the input file.
	sheetIdFile := "../portfolio-sheet-id.txt"
	// Initialize the Google sheet interface.
	googleSheetMgr := finance.NewGoogleSheetManager(client, &ctx, sheetIdFile)

	// Specify the sheet and columns, and read from transactions sheet.
	txns := googleSheetMgr.GetTransactionData()
	// Set gin web server to release mode. Comment out to enable debug logging.
	gin.SetMode(gin.ReleaseMode)
	// Setup the Go web server, with default logger.
	router := gin.Default()
	router.Use(cors.Default())

	// Connect to our MongoDB instance.
	dbClient := data.NewMongoDbClient()
	dbClient.ConnectMongoDb("polly-data-prod")
	// Name the python script to use with yfinance to grab extended stock info.
	pyScript := "yahooFinanceHelper.py"
	catalogue := finance.NewSecurityCatalogue(googleSheetMgr, dbClient, pyScript)
	// Process the imported data to organize it by ticker.
	(*catalogue).ProcessImport(txns.Values)
	fmt.Println("Number of transactions processed: " + strconv.Itoa(len(txns.Values)))
	// Calculate metrics for each stock.
	catalogue.Calculate()
	// Create a controller to manage front-end interaction.
	ctrlr := controllers.SecurityController{}
	ctrlr.Init(catalogue)
	// Setup the GET routes for presenting data to the web server.
	router.GET("/summary", ctrlr.GetSummary)
	router.GET("/securities", ctrlr.GetSecurities)
	router.GET("/transactions", ctrlr.GetTransactions)
	router.GET("/sp500", ctrlr.GetSp500History)
	router.GET("/history", ctrlr.GetPortfolioHistory)

	// Disable trusted proxies.
	router.SetTrustedProxies(nil)
	// Run the server.
	router.Run(":5000")
	fmt.Println("Successful Completion!")
}
