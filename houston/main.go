package main

import (
	"time"

	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sw360cab/gno-devops/graphql"
	"github.com/sw360cab/gno-devops/metrics"
	"github.com/sw360cab/gno-devops/router"
	"github.com/sw360cab/gno-devops/utils"
)

func main() {
	godotenv.Load()
	// init logger
	logger := utils.InitLogger()
	defer logger.Sync()
	// rdb := initRedisClient()
	txMetrics := metrics.NewTransactionMetric(logger)

	// HTTP Server config
	ginRouter := gin.Default()
	ginRouteHanler := router.GinRouteHandler{
		Tmc: txMetrics,
	}
	// Dashboard Endpoints
	ginRouter.GET("/count", ginRouteHanler.GetTransactionCount)
	ginRouter.GET("/successRate", ginRouteHanler.GetTransactionSuccessRate)
	ginRouter.GET("/messageTypes", ginRouteHanler.GetMessageTypes)
	ginRouter.GET("/senders", ginRouteHanler.GetTopTransactionSenders)
	// `/realms/(deployed|called)`
	// `/packages/(deployed|called)`
	ginRouter.GET("/:type/:status", ginRouteHanler.GetItemsTransctionWithStatus)

	// GQL Client
	graphqlEndpoint := utils.GetEnvWithFallback(graphql.GRAPHQL_URL_ENV, graphql.DEFAULT_GRAPHQL_URL)
	gqlClient := graphql.GraphQLClient{
		Endpoint:                     graphqlEndpoint,
		SubscriptionResponseHandler:  txMetrics,
		SubscriptionConnRetryTimeout: 2 * time.Minute, // configurable if a future version
	}

	// Launch GQL suscription in another GO routine
	go func() {
		err := gqlClient.CreateGQLSubscription()
		if err != nil {
			logger.Fatal("Unable to subscribe to GraphQL server", zap.Error(err))
		}
	}()

	// Launch GQL query for unprocessed / previous blocks
	go func() {
		bootstrapTime := time.Now()
		defer close(txMetrics.LatestBlockCh)
		leftoverBlock := <-txMetrics.LatestBlockCh
		transactions, err := gqlClient.QueryPreExistingBlocks(leftoverBlock, bootstrapTime)
		if err != nil {
			logger.Error("Unable to fetch blocks before subscription",
				zap.Int64("last block", leftoverBlock.ToBlock),
				zap.Error(err))
		}
		for _, transaction := range transactions {
			txMetrics.HandleTransactionMessage(transaction)
		}
	}()

	// Launch Gin in the main routine
	ginRouter.Run("0.0.0.0:8080")
}
