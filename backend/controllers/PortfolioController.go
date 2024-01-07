package controllers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/kfwalther/Polly/backend/auth"
	"github.com/kfwalther/Polly/backend/config"
	"github.com/kfwalther/Polly/backend/data"
	"github.com/kfwalther/Polly/backend/finance"
)

type PortfolioController struct {
	stockCatalogue       *finance.EquityCatalogue
	etfCatalogue         *finance.EquityCatalogue
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
	// Create the new equity catalogues to house our portfolio data.
	c.stockCatalogue = finance.NewEquityCatalogue(c.googleSheetMgr, c.dbClient, c.yfinPythonScriptFile)
	c.etfCatalogue = finance.NewEquityCatalogue(c.googleSheetMgr, c.dbClient, c.yfinPythonScriptFile)
	// Read from portfolio transactions sheets.
	stockTxns := c.googleSheetMgr.GetStockTransactionData()
	etfTxns := c.googleSheetMgr.GetEtfTransactionData()
	// Process the imported data to organize it by ticker.
	c.stockCatalogue.ProcessImport(stockTxns.Values)
	c.etfCatalogue.ProcessImport(etfTxns.Values)
	log.Printf("Number of stock transactions processed: %d", len(stockTxns.Values))
	log.Printf("Number of ETF transactions processed: %d", len(etfTxns.Values))
	// Calculate metrics for each stock.
	c.stockCatalogue.Calculate()
	c.etfCatalogue.Calculate()
}

// Define the endpoint for the Google Sheets API OAuth redirect URL.
func (c *PortfolioController) OAuthRedirectCallback(ctx *gin.Context) {
	c.oauthHandler.HandleTokenResponse(ctx.Writer, ctx.Request)
	// If valid OAuth token received, we can initialize using new HttpClient.
	if httpClient := c.oauthHandler.GetHttpClient(); httpClient != nil {
		c.CreatePortfolioCatalogueAndProcess(httpClient)
	}
}

func (c *PortfolioController) GetStockSummary(ctx *gin.Context) {
	summary := c.stockCatalogue.GetPortfolioSummary()
	if summary == nil {
		log.Print("No stock portfolio summary to forward thru API to front-end!")
		ctx.JSON(400, gin.H{
			"error": "No stock portfolio summary found!",
		})
	} else {
		log.Print("Sending stock portfolio summary to front-end...")
		ctx.JSON(200, gin.H{
			"summary": summary,
		})
	}
}

func (c *PortfolioController) GetEtfSummary(ctx *gin.Context) {
	summary := c.etfCatalogue.GetPortfolioSummary()
	if summary == nil {
		log.Print("No ETF portfolio summary to forward thru API to front-end!")
		ctx.JSON(400, gin.H{
			"error": "No ETF portfolio summary found!",
		})
	} else {
		log.Print("Sending ETF portfolio summary to front-end...")
		ctx.JSON(200, gin.H{
			"summary": summary,
		})
	}
}

func (c *PortfolioController) GetFullSummary(ctx *gin.Context) {
	summary := c.fullPortfolioSummary
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
	if c.stockCatalogue.PortfolioHistory == nil || len(c.stockCatalogue.PortfolioHistory) == 0 {
		log.Print("No stock portfolio history to forward thru API to front-end!")
		ctx.JSON(400, gin.H{
			"error": "No stock portfolio history found!",
		})
	} else {
		log.Print("Sending stock portfolio history to front-end...")
		ctx.JSON(200, gin.H{
			"history": c.stockCatalogue.PortfolioHistory,
		})
	}
}

func (c *PortfolioController) GetStocks(ctx *gin.Context) {
	eqs := c.stockCatalogue.GetEquityList()
	if len(eqs) == 0 {
		log.Print("No stocks to forward thru API to front-end!")
		ctx.JSON(400, gin.H{
			"error": "No stocks found in the portfolio!",
		})
	} else {
		log.Printf("Sending %d stocks to front-end...", len(eqs))
		ctx.JSON(200, gin.H{
			"equities": eqs,
		})
	}
}

func (c *PortfolioController) GetEtfs(ctx *gin.Context) {
	eqs := c.etfCatalogue.GetEquityList()
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

func (c *PortfolioController) GetAllEquities(ctx *gin.Context) {
	stocks := c.stockCatalogue.GetEquityList()
	etfs := c.etfCatalogue.GetEquityList()
	if len(stocks) == 0 && len(etfs) == 0 {
		log.Print("No stocks or ETFs to forward thru API to front-end!")
		ctx.JSON(400, gin.H{
			"error": "No stocks or ETFs found in the portfolio!",
		})
	} else {
		mergedList := append(stocks, etfs...)
		log.Printf("Sending %d stocks and ETFs to front-end...", len(mergedList))
		ctx.JSON(200, gin.H{
			"equities": mergedList,
		})
	}
}

func (c *PortfolioController) GetTransactions(ctx *gin.Context) {
	stockTxns := c.stockCatalogue.GetTransactionList()
	etfTxns := c.etfCatalogue.GetTransactionList()
	if len(stockTxns) == 0 && len(etfTxns) == 0 {
		log.Print("No transactions to forward thru API to front-end!")
		ctx.JSON(400, gin.H{
			"error": "No transactions found in the portfolio!",
		})
	} else {
		mergedList := append(stockTxns, etfTxns...)
		log.Printf("Sending %d transactions to front-end...", len(mergedList))
		ctx.JSON(200, gin.H{
			"transactions": mergedList,
		})
	}
}

func (c *PortfolioController) GetSp500History(ctx *gin.Context) {
	sp500 := c.stockCatalogue.GetSp500()
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
	// TODO: Fix the percentages when doing two refreshes.
	c.stockCatalogue.Refresh(progressSocket)
	c.etfCatalogue.Refresh(progressSocket)
	c.fullPortfolioSummary = finance.NewPortfolioSummary()
	c.stockCatalogue.SendProgressUpdate(2.0)
	// Read from portfolio transactions sheets.
	stockTxns := c.googleSheetMgr.GetStockTransactionData()
	c.stockCatalogue.SendProgressUpdate(4.0)
	etfTxns := c.googleSheetMgr.GetEtfTransactionData()
	c.stockCatalogue.SendProgressUpdate(6.0)
	// Process the imported data to organize it by ticker.
	c.stockCatalogue.ProcessImport(stockTxns.Values)
	c.stockCatalogue.SendProgressUpdate(9.0)
	c.etfCatalogue.ProcessImport(etfTxns.Values)
	c.stockCatalogue.SendProgressUpdate(12.0)
	log.Printf("Number of stock transactions processed: %d", len(stockTxns.Values))
	log.Printf("Number of ETF transactions processed: %d", len(etfTxns.Values))
	// Calculate metrics for each stock.
	c.stockCatalogue.Calculate()
	c.stockCatalogue.SendProgressUpdate(56.0)
	c.etfCatalogue.Calculate()
	c.CalculatePortfolioSummaryMetrics()
	c.stockCatalogue.SendProgressUpdate(100.0)
}

func (c *PortfolioController) CalculatePortfolioSummaryMetrics() {
	// Define the first trading day of the year to reference for YTD calculations.
	stockSummary := c.stockCatalogue.GetPortfolioSummary()
	etfSummary := c.etfCatalogue.GetPortfolioSummary()
	c.fullPortfolioSummary.MarketValueJan1 = stockSummary.MarketValueJan1 + etfSummary.MarketValueJan1
	c.fullPortfolioSummary.TotalMarketValue = stockSummary.TotalMarketValue + etfSummary.TotalMarketValue
	c.fullPortfolioSummary.TotalCostBasis = stockSummary.TotalCostBasis + etfSummary.TotalCostBasis
	c.fullPortfolioSummary.DailyGain = stockSummary.DailyGain + etfSummary.DailyGain
	c.fullPortfolioSummary.TotalEquities = stockSummary.TotalEquities + etfSummary.TotalEquities
	c.fullPortfolioSummary.LastUpdated = time.Now()
	if c.fullPortfolioSummary.TotalCostBasis > 0.001 {
		c.fullPortfolioSummary.PercentageGain = ((c.fullPortfolioSummary.TotalMarketValue - c.fullPortfolioSummary.TotalCostBasis) / c.fullPortfolioSummary.TotalCostBasis) * 100.0
	}
	c.fullPortfolioSummary.YearToDatePercentageGain = (c.fullPortfolioSummary.TotalMarketValue/
		(c.fullPortfolioSummary.MarketValueJan1+c.stockCatalogue.CashFlowYtd+c.etfCatalogue.CashFlowYtd) - 1) * 100.0
}
