package controllers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/kfwalther/Polly/backend/finance"
)

type SecurityController struct {
	securityCatalogue *finance.SecurityCatalogue
}

func (c *SecurityController) Init(catalogue *finance.SecurityCatalogue) {
	c.securityCatalogue = catalogue
}

func (c *SecurityController) GetSummary(ctx *gin.Context) {
	summary := c.securityCatalogue.GetPortfolioSummary()
	if summary == nil {
		log.Printf("No portfolio summary to forward thru API to front-end!")
		ctx.JSON(400, gin.H{
			"error": "No portfolio summary found!",
		})
	} else {
		log.Printf("Sending portfolio summary to front-end...")
		ctx.JSON(200, gin.H{
			"summary": summary,
		})
	}
}

func (c *SecurityController) GetPortfolioHistory(ctx *gin.Context) {
	if c.securityCatalogue.PortfolioHistory == nil || len(c.securityCatalogue.PortfolioHistory) == 0 {
		log.Printf("No portfolio history to forward thru API to front-end!")
		ctx.JSON(400, gin.H{
			"error": "No portfolio history found!",
		})
	} else {
		log.Printf("Sending portfolio history to front-end...")
		ctx.JSON(200, gin.H{
			"history": c.securityCatalogue.PortfolioHistory,
		})
	}
}

func (c *SecurityController) GetSecurities(ctx *gin.Context) {
	secs := c.securityCatalogue.GetSecurityList()
	if len(secs) == 0 {
		log.Printf("No securities to forward thru API to front-end!")
		ctx.JSON(400, gin.H{
			"error": "No securities found in the portfolio!",
		})
	} else {
		log.Printf("Sending %d securities to front-end...", len(secs))
		ctx.JSON(200, gin.H{
			"securities": secs,
		})
	}
}

func (c *SecurityController) GetTransactions(ctx *gin.Context) {
	txns := c.securityCatalogue.GetTransactionList()
	if len(txns) == 0 {
		log.Printf("No transactions to forward thru API to front-end!")
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

func (c *SecurityController) GetSp500History(ctx *gin.Context) {
	sp500 := c.securityCatalogue.GetSp500()
	if len(sp500.Date) == 0 {
		log.Printf("No historical S&P500 data to forward thru API to front-end!")
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
func (c *SecurityController) WebSocketHandler(ctx *gin.Context) {
	log.Printf("Setting up web socket...")
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
	c.securityCatalogue.Refresh(progressSocket)
}
