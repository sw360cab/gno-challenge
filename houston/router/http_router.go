package router

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/sw360cab/gno-devops/metrics"
)

// Binding type for parametrized endpoints
type ItemTransaction struct {
	Type   string `uri:"type" binding:"required,oneof=realms packages"`
	Status string `uri:"status" binding:"required,oneof=deployed called"`
}

type GinRouteHandler struct {
	Tmc metrics.TransactionMetricsCollector
}

/**********************/
/* Endpoints handlers */
/**********************/

func (r GinRouteHandler) GetTransactionCount(c *gin.Context) {
	c.JSON(200, gin.H{
		"count": r.Tmc.GetTransactionCount(),
	})
}

func (r GinRouteHandler) GetTransactionSuccessRate(c *gin.Context) {
	c.JSON(200, gin.H{
		"successRate": fmt.Sprintf("%.2f", r.Tmc.GetTransactionSuccessRate()),
	})
}

func (r GinRouteHandler) GetMessageTypes(c *gin.Context) {
	c.JSON(200, r.Tmc.GetMessageTypes())
}

func (r GinRouteHandler) GetTopTransactionSenders(c *gin.Context) {
	c.JSON(200, r.Tmc.GetTopTransactionSenders())
}

// Handler for parametrized endpoints
// `/realms/(deployed|called)`
// `/packages/(deployed|called)`
func (r GinRouteHandler) GetItemsTransctionWithStatus(c *gin.Context) {
	var itemTx ItemTransaction
	if err := c.ShouldBindUri(&itemTx); err != nil {
		c.JSON(400, gin.H{"msg": err})
		return
	}

	var handler func() []metrics.SlicedMap
	switch itemTx.Type {
	case "realms":
		if itemTx.Status == "deployed" {
			handler = r.Tmc.GetMostActiveRealmsDeployed
		} else {
			handler = r.Tmc.GetMostActiveRealmsCalled
		}
	case "packages":
		if itemTx.Status == "deployed" {
			handler = r.Tmc.GetMostActivePackagesDeployed
		} else {
			handler = r.Tmc.GetMostActivePackagesCalled
		}
	}
	c.JSON(200, handler())
}
