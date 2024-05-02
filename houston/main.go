package main

import (
	"log"
	"os"
	"time"

	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sw360cab/gno-devops/metrics"
	"github.com/sw360cab/gno-devops/router"
)

// func initRedisClient() *redis.Client {
// 	redisHost, ok := os.LookupEnv("REDIS_HOST")
// 	if !ok {
// 		redisHost = "127.0.0.1"
// 	}
// 	redisPort, ok := os.LookupEnv("REDIS_PORT")
// 	if !ok {
// 		redisPort = "6379"
// 	}

// 	return redis.NewClient(&redis.Options{
// 		Addr:     fmt.Sprintf("%s:%s", redisHost, redisPort),
// 		Password: "",
// 		DB:       0,
// 	})
// }

func getEnvVarFallback(envVarKey string, fallbackValue string) string {
	envVarValue, ok := os.LookupEnv(envVarKey)
	if !ok {
		envVarValue = fallbackValue
	}
	return envVarValue
}

func initLogger() *zap.Logger {
	logLevel := getEnvVarFallback("LOG_LEVEL", "DEBUG")
	logLevelCfg, err := zap.ParseAtomicLevel(logLevel)
	if err != nil {
		log.Fatal("unable to parse log level, %w", err)
	}

	cfg := zap.NewDevelopmentConfig()
	cfg.Level = logLevelCfg
	// Create a new logger
	logger, err := cfg.Build()
	if err != nil {
		log.Fatal("unable to create logger, %w", err)
	}
	return logger
}

func main() {
	godotenv.Load()
	// init logger
	logger := initLogger()
	defer logger.Sync()
	// rdb := initRedisClient()
	txMetrics := metrics.NewTransactionMetric(nil, logger)

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
	graphqlEndpoint := getEnvVarFallback(metrics.GRAPHQL_URL_ENV, metrics.DEFAULT_GRAPHQL_URL)
	gqlClient := metrics.GraphQLClient{
		Endpoint:                    graphqlEndpoint,
		SubscriptionResponseHandler: txMetrics,
	}

	// Launch GQL suscription in another GO routine
	go func() {
		err := gqlClient.CreateGQLSubscription()
		if err != nil {
			logger.Fatal("Unable to subscribe to GraphQL server", zap.Error(err))

			// // attempt to fetch last value from redis
			// if err = txMetrics.FetchLatestBlockFromRedis(); err != nil {
			// 	log.Fatal("Unable to get value from redis server... killing myself ", err)
			// }
		}
	}()

	// Launch GQL query for unprocessed / previous blocks
	go func() {
		bootstrapTime := time.Now()
		leftoverBlock := <-txMetrics.LatestBlockCh
		transactions, err := gqlClient.CreateGQLStaticQuery(leftoverBlock, bootstrapTime)
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
