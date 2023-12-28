package controllers

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/kfwalther/Polly/backend/auth"
	"github.com/kfwalther/Polly/backend/config"
	"github.com/kfwalther/Polly/backend/data"
	"github.com/kfwalther/Polly/backend/finance"
)

type PortfolioController struct {
	equityCatalogue      *finance.EquityCatalogue
	oauthHandler         *auth.OAuthHandler
	dbClient             *data.MongoDbClient
	googleSheetMgr       *finance.GoogleSheetManager
	googleSheetIdsFile   string
	yfinPythonScriptFile string
}

// Constructor for the controller for interfacing with the front-end.
func NewPortfolioController(oauthHandler *auth.OAuthHandler, googleSheetIdsFile string, pyScript string) *PortfolioController {
	var ctrlr PortfolioController
	ctrlr.oauthHandler = oauthHandler
	ctrlr.googleSheetIdsFile = googleSheetIdsFile
	ctrlr.yfinPythonScriptFile = pyScript
	return &ctrlr
}

// Initialize the MongoDB client, and attempt to get our Sheet API auth token.
func (c *PortfolioController) Init(config *config.Configuration) {
	// Connect to our MongoDB instance.
	c.dbClient = data.NewMongoDbClient()
	c.dbClient.ConnectMongoDb(config.MongoDbConnectionUri, config.MongoDbName)
	// If valid OAuth token received, we can initialize here. Otherwise, wait for redirect callback.
	if httpClient := c.oauthHandler.GetHttpClient(); httpClient != nil {
		c.CreatePortfolioCatalogueAndProcess(httpClient)
	}
}

// Initialize the Sheets API and portfolio catalogue, then calculate metrics.
func (c *PortfolioController) CreatePortfolioCatalogueAndProcess(httpClient *http.Client) {
	ctx := context.Background()
	// Initialize the Google sheet interface.
	c.googleSheetMgr = finance.NewGoogleSheetManager(httpClient, &ctx, c.googleSheetIdsFile)
	// Create the new equity catalogue to house our portfolio data.
	c.equityCatalogue = finance.NewEquityCatalogue(c.googleSheetMgr, c.dbClient, c.yfinPythonScriptFile)
	// Read from portfolio transactions sheet.
	txns := c.googleSheetMgr.GetTransactionData()
	// Process the imported data to organize it by ticker.
	c.equityCatalogue.ProcessImport(txns.Values)
	log.Printf("Number of transactions processed: %d", len(txns.Values))
	// Calculate metrics for each stock.
	c.equityCatalogue.Calculate()
}

// Define the endpoint for the Google Sheets API OAuth redirect URL.
func (c *PortfolioController) OAuthRedirectCallback(ctx *gin.Context) {
	c.oauthHandler.HandleTokenResponse(ctx.Writer, ctx.Request)
	// If valid OAuth token received, we can initialize using new HttpClient.
	if httpClient := c.oauthHandler.GetHttpClient(); httpClient != nil {
		c.CreatePortfolioCatalogueAndProcess(httpClient)
	}
}

func (c *PortfolioController) GetSummary(ctx *gin.Context) {
	summary := c.equityCatalogue.GetPortfolioSummary()
	if summary == nil {
		log.Print("No portfolio summary to forward thru API to front-end!")
		ctx.JSON(400, gin.H{
			"error": "No portfolio summary found!",
		})
	} else {
		log.Print("Sending portfolio summary to front-end...")
		ctx.JSON(200, gin.H{
			"summary": summary,
		})
	}
}

func (c *PortfolioController) GetPortfolioHistory(ctx *gin.Context) {
	if c.equityCatalogue.PortfolioHistory == nil || len(c.equityCatalogue.PortfolioHistory) == 0 {
		log.Print("No portfolio history to forward thru API to front-end!")
		ctx.JSON(400, gin.H{
			"error": "No portfolio history found!",
		})
	} else {
		log.Print("Sending portfolio history to front-end...")
		ctx.JSON(200, gin.H{
			"history": c.equityCatalogue.PortfolioHistory,
		})
	}
}

func (c *PortfolioController) GetEquities(ctx *gin.Context) {
	eqs := c.equityCatalogue.GetEquityList()
	if len(eqs) == 0 {
		log.Print("No equities to forward thru API to front-end!")
		ctx.JSON(400, gin.H{
			"error": "No equities found in the portfolio!",
		})
	} else {
		log.Printf("Sending %d equities to front-end...", len(eqs))
		ctx.JSON(200, gin.H{
			"equities": eqs,
		})
	}
}

func (c *PortfolioController) GetTransactions(ctx *gin.Context) {
	txns := c.equityCatalogue.GetTransactionList()
	if len(txns) == 0 {
		log.Print("No transactions to forward thru API to front-end!")
		ctx.JSON(400, gin.H{
			"error": "No transactions found in the portfolio!",
		})
	} else {
		log.Printf("Sending %d transactions to front-end...", len(txns))
		ctx.JSON(200, gin.H{
			"transactions": txns,
		})
	}
}

func (c *PortfolioController) GetSp500History(ctx *gin.Context) {
	sp500 := c.equityCatalogue.GetSp500()
	if len(sp500.Date) == 0 {
		log.Print("No historical S&P500 data to forward thru API to front-end!")
		ctx.JSON(400, gin.H{
			"error": "No historical S&P500 data found! Try restarting server to re-query.",
		})
	} else {
		log.Printf("Sending %d historical S&P500 quotes to front-end...", len(sp500.Date))
		ctx.JSON(200, gin.H{
			"sp500": sp500,
		})
	}
}

// Connect to a web socket on this endpoint to send periodic progress updates.
func (c *PortfolioController) WebSocketHandler(ctx *gin.Context) {
	log.Print("Setting up web socket...")
	// Define the headers/settings for upgrading the http:// request to ws://.
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		// Account for cross-domain issues
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	// Received the socket request, now upgrade it to ws://
	progressSocket, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		log.Println("WARNING: Web socket protocol upgrade error: ", err)
		return
	}
	// Close the socket when we're done.
	defer progressSocket.Close()
	// Write the web socket message to the client (1% progress).
	err = progressSocket.WriteMessage(websocket.TextMessage, []byte("1"))
	if err != nil {
		log.Println("WARNING: Web socket write error: ", err)
	}
	// Refresh the portfolio data, providing the web socket.
	c.equityCatalogue.Refresh(progressSocket)
}
