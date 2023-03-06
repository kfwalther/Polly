package controllers

import (
	"log"

	"github.com/gin-gonic/gin"

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
