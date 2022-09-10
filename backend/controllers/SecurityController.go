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
