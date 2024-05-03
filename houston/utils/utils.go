package utils

import (
	"log"
	"os"

	"go.uber.org/zap"
)

// Lookup for an env variable and if not found returns a fallback string value.
// Currently works only with values of string type.
func GetEnvWithFallback(envVarKey string, fallbackValue string) string {
	envVarValue, ok := os.LookupEnv(envVarKey)
	if !ok {
		envVarValue = fallbackValue
	}
	return envVarValue
}

// Initializes the zap logger with the given log level
func InitLogger() *zap.Logger {
	logLevel := GetEnvWithFallback("LOG_LEVEL", "DEBUG")
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
