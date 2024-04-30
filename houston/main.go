package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sw360cab/gno-devops/metrics"
	"github.com/sw360cab/gno-devops/router"
)

func main() {
	godotenv.Load()
	txMetrics := metrics.NewTransactionMetric()

	ginRouter := gin.Default()
	ginRouteHanler := router.GinRouteHandler{
		Tmc: txMetrics,
	}
	// Dashboard Endpoints
	ginRouter.GET("/count", ginRouteHanler.GetTransactionCount)
	ginRouter.GET("/successRate", ginRouteHanler.GetTransactionSuccessRate)
	ginRouter.GET("/messageTypes", ginRouteHanler.GetMessageTypes)
	ginRouter.GET("/senders", ginRouteHanler.GetTopTransactionSenders)
	// /realms/(deployed|called)
	// /packages/(deployed|called)
	ginRouter.GET("/:type/:status", ginRouteHanler.GetItemsTransctionWithStatus)

	// GQL Client
	graphqlEndpoint, ok := os.LookupEnv(metrics.GRAPHQL_URL_ENV)
	if !ok {
		graphqlEndpoint = metrics.DEFAULT_GRAPHQL_URL
	}
	gqlClient := metrics.GraphQLClient{
		Endpoint:                    graphqlEndpoint,
		SubscriptionResponseHandler: txMetrics,
	}
	// Launch GQL suscription in another GO routine
	go func() {
		err := gqlClient.CreateGQLSubscription()
		if err != nil {
			log.Fatal("Unable to subscribe to GraphQL server", err)
		}
	}()

	ginRouter.Run("0.0.0.0:8080")
}
