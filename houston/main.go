package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sw360cab/gno-devops/metrics"
)

func main() {
	txMetrics := metrics.NewTransactionMetric()
	router := gin.Default()

	godotenv.Load()
	graphqlEndpoint, ok := os.LookupEnv(metrics.GRAPHQL_URL_ENV)
	if !ok {
		graphqlEndpoint = metrics.DEFAULT_GRAPHQL_URL
	}

	router.GET("/count", func(c *gin.Context) {
		count, _ := txMetrics.GetTransactionCount()
		c.JSON(200, count)
	})

	router.GET("/messageTypes", func(c *gin.Context) {
		types, _ := txMetrics.GetMessageTypes()
		c.JSON(200, types)
	})

	router.GET("/senders", func(c *gin.Context) {
		senders, _ := txMetrics.GetTopTransactionSenders()
		c.JSON(200, senders)
	})

	gqlClient := metrics.GraphQLClient{
		Endpoint:                    graphqlEndpoint,
		SubscriptionResponseHandler: txMetrics,
	}
	// err := gqlClient.CreateGQLStaticQuery()
	err := gqlClient.CreateGQLSubscription()
	if err != nil {
		log.Fatal("Unable to subscribe to GraphQL server", err)
	}

	router.Run("0.0.0.0:8080")
}
