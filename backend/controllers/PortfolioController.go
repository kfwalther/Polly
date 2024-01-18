package controllers

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/kfwalther/Polly/backend/auth"
	"github.com/kfwalther/Polly/backend/config"
	"github.com/kfwalther/Polly/backend/data"
	"github.com/kfwalther/Polly/backend/finance"
)

type PortfolioController struct {
	equityTypes          []string
	equityCatalogues     map[string]*finance.EquityCatalogue
	fullPortfolioSummary *finance.PortfolioSummary
	oauthHandler         *auth.OAuthHandler
	dbClient             *data.MongoDbClient
	googleSheetMgr       *finance.GoogleSheetManager
	googleSheetIdsFile   string
	yfinPythonScriptFile string
}

// Constructor for the controller for interfacing with the front-end.
func NewPortfolioController(oauthHandler *auth.OAuthHandler, googleSheetIdsFile string, pyScript string) *PortfolioController {
	var ctrlr PortfolioController
	ctrlr.equityCatalogues = make(map[string]*finance.EquityCatalogue)
	ctrlr.oauthHandler = oauthHandler
	ctrlr.googleSheetIdsFile = googleSheetIdsFile
	ctrlr.yfinPythonScriptFile = pyScript
	ctrlr.fullPortfolioSummary = finance.NewPortfolioSummary()
	return &ctrlr
}

// Initialize the MongoDB client, and attempt to get our Sheet API auth token.
func (c *PortfolioController) Init(config *config.Configuration) {
	// Connect to our MongoDB instance.
	c.dbClient = data.NewMongoDbClient()
	c.dbClient.ConnectMongoDb(config.MongoDbConnectionUri, config.MongoDbName)
	c.equityTypes = config.EquityTypes
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
	for _, equityType := range c.equityTypes {
		// Create the new equity catalogues to house our portfolio data.
		catalogue := finance.NewEquityCatalogue(equityType, c.googleSheetMgr, c.dbClient, c.yfinPythonScriptFile)
		// Read from portfolio transactions sheets.
		txns := c.googleSheetMgr.GetTransactionData(equityType)
		// Process the imported data to organize it by ticker.
		catalogue.ProcessImport(txns.Values)
		log.Printf("Number of %s transactions processed: %d", equityType, len(txns.Values))
		// Calculate metrics for each catalogue's holdings.
		catalogue.Calculate()
		c.equityCatalogues[equityType] = catalogue
	}
	c.CalculatePortfolioSummaryMetrics()
}

// Define the endpoint for the Google Sheets API OAuth redirect URL.
func (c *PortfolioController) OAuthRedirectCallback(ctx *gin.Context) {
	c.oauthHandler.HandleTokenResponse(ctx.Writer, ctx.Request)
	// If valid OAuth token received, we can initialize using new HttpClient.
	if httpClient := c.oauthHandler.GetHttpClient(); httpClient != nil {
		c.CreatePortfolioCatalogueAndProcess(httpClient)
	}
}

func (c *PortfolioController) GetSummary(ctx *gin.Context, equityType string) {
	var summary *finance.PortfolioSummary
	if equityType == "full" {
		summary = c.fullPortfolioSummary
	} else {
		summary = c.equityCatalogues[equityType].GetPortfolioSummary()
	}
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

// Only sending the stock portfolio history for now.
func (c *PortfolioController) GetPortfolioHistory(ctx *gin.Context) {
	if c.equityCatalogues["stock"].PortfolioHistory == nil || len(c.equityCatalogues["stock"].PortfolioHistory) == 0 {
		log.Print("No stock portfolio history to forward thru API to front-end!")
		ctx.JSON(400, gin.H{
			"error": "No stock portfolio history found!",
		})
	} else {
		log.Print("Sending stock portfolio history to front-end...")
		ctx.JSON(200, gin.H{
			"history": c.equityCatalogues["stock"].PortfolioHistory,
		})
	}
}

func (c *PortfolioController) GetEquities(ctx *gin.Context, equityType string) {
	var eqs []*finance.Equity
	if equityType == "full" {
		eqs = c.equityCatalogues["stock"].GetEquityList()
		eqs = append(eqs, c.equityCatalogues["etf"].GetEquityList()...)
		eqs = append(eqs, c.equityCatalogues["crypto"].GetEquityList()...)
	} else {
		eqs = c.equityCatalogues[equityType].GetEquityList()
	}
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
	var txns []finance.Transaction
	txns = c.equityCatalogues["stock"].GetTransactionList()
	txns = append(txns, c.equityCatalogues["etf"].GetTransactionList()...)
	txns = append(txns, c.equityCatalogues["crypto"].GetTransactionList()...)
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
	sp500 := c.equityCatalogues["stock"].GetSp500()
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
	c.fullPortfolioSummary = finance.NewPortfolioSummary()
	prog := 0.0
	for _, equityType := range c.equityTypes {
		c.equityCatalogues[equityType].Refresh()
		// Read from portfolio transactions sheets.
		txns := c.googleSheetMgr.GetTransactionData(equityType)
		prog += 3.0
		c.SendProgressUpdate(progressSocket, prog)
		// Process the imported data to organize it by ticker.
		c.equityCatalogues[equityType].ProcessImport(txns.Values)
		prog += 3.0
		c.SendProgressUpdate(progressSocket, prog)
		log.Printf("Number of %s transactions processed: %d", equityType, len(txns.Values))
		// Calculate metrics for each stock.
		c.equityCatalogues[equityType].Calculate()
		prog += 27.0
		c.SendProgressUpdate(progressSocket, prog)
	}
	c.CalculatePortfolioSummaryMetrics()
	c.SendProgressUpdate(progressSocket, 100.0)
}

// Send progress updates to the client via the web socket, if it's initialized.
func (c *PortfolioController) SendProgressUpdate(progressSocket *websocket.Conn, progressPercent float64) {
	percentStr := strconv.FormatFloat(progressPercent, 'E', -1, 64)
	// Write the web socket message to the client (1% progress).
	err := progressSocket.WriteMessage(websocket.TextMessage, []byte(percentStr))
	if err != nil {
		log.Println("WARNING: Web socket write error: ", err)
	}
}

func (c *PortfolioController) CalculatePortfolioSummaryMetrics() {
	// Define the first trading day of the year to reference for YTD calculations.
	c.fullPortfolioSummary.LastUpdated = time.Now()
	totalCashFlowYtd := 0.0
	for _, equityType := range c.equityTypes {
		summary := c.equityCatalogues[equityType].GetPortfolioSummary()
		c.fullPortfolioSummary.MarketValueJan1 += summary.MarketValueJan1
		c.fullPortfolioSummary.TotalMarketValue += summary.TotalMarketValue
		c.fullPortfolioSummary.TotalCostBasis += summary.TotalCostBasis
		c.fullPortfolioSummary.DailyGain += summary.DailyGain
		c.fullPortfolioSummary.TotalEquities += summary.TotalEquities
		totalCashFlowYtd += c.equityCatalogues[equityType].CashFlowYtd
	}
	if c.fullPortfolioSummary.TotalCostBasis > 0.001 {
		c.fullPortfolioSummary.PercentageGain = ((c.fullPortfolioSummary.TotalMarketValue - c.fullPortfolioSummary.TotalCostBasis) / c.fullPortfolioSummary.TotalCostBasis) * 100.0
	}
	c.fullPortfolioSummary.YearToDatePercentageGain = (c.fullPortfolioSummary.TotalMarketValue/(c.fullPortfolioSummary.MarketValueJan1+totalCashFlowYtd) - 1) * 100.0
}
