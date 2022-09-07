package controllers

import (
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
	secs := []finance.Security{}

	if len(c.securityCatalogue.GetSecurityList()) == 0 {
		ctx.JSON(400, gin.H{
			"error": "No securities found in the portfolio!",
		})
	} else {
		secs = c.securityCatalogue.GetSecurityList()
		ctx.JSON(200, gin.H{
			"securities": secs,
		})
	}
}
