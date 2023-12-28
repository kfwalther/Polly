package main

// List of imported packages
import (
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/kfwalther/Polly/backend/auth"
	"github.com/kfwalther/Polly/backend/config"
	"github.com/kfwalther/Polly/backend/controllers"
	"golang.org/x/oauth2/google"
)

// Program entry point.
func main() {
	// Read the config file for this server.
	config := config.NewConfiguration("../go-server-config.json")
	// Get the GCP credentials file.
	b, err := os.ReadFile(config.GcpCredentialsFile)
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved auth_token.json.
	oauthConfig, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets.readonly")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	// Define the file to store our auth token.
	tokenFile := config.AuthTokenFile
	// Create a new OAuth handler to manage OAuth with Google Sheets API.
	oauthHandler := auth.NewOAuthHandler(tokenFile, oauthConfig)
	// Define the Google Sheets IDs file.
	sheetIdFile := config.GoogleSheetsIdsFile
	// Name the python script to use with yfinance to grab extended stock info.
	pyScript := "yahooFinanceHelper.py"
	// Create a controller to manage front-end interaction.
	ctrlr := controllers.NewPortfolioController(oauthHandler, sheetIdFile, pyScript)
	ctrlr.Init(config)

	// Set gin web server to release mode. Comment out to enable debug logging.
	gin.SetMode(gin.ReleaseMode)
	// Setup the Go web server, with default logger.
	router := gin.Default()
	router.Use(cors.Default())
	// Setup the GET routes for our web server.
	router.GET("/summary", ctrlr.GetSummary)
	router.GET("/equities", ctrlr.GetEquities)
	router.GET("/transactions", ctrlr.GetTransactions)
	router.GET("/sp500", ctrlr.GetSp500History)
	router.GET("/history", ctrlr.GetPortfolioHistory)
	router.GET("/refresh", ctrlr.WebSocketHandler)
	router.GET("/tokenresponse", ctrlr.OAuthRedirectCallback)

	// Disable trusted proxies.
	router.SetTrustedProxies(nil)
	// Run the web server.
	router.Run(":" + config.WebServerPort)
}
