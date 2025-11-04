package main

import (
	"os"
	"paymentservice/internal/app"
	"strconv"
)

func getEnvAsInt64(key string, defaultVal int64) int64 {
	if value, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intVal
		}
	}
	return defaultVal
}

func main() {
	app.Run(
		os.Getenv("BEANSTALK_ADR"),
		os.Getenv("BEANSTALK_TUBE_NAME"),
		os.Getenv("REDIS_DSN"),
		getEnvAsInt64("RATELIMIT_LINIT", 5),
		getEnvAsInt64("RATELIMIT_WINDOW", 1),
		getEnvAsInt64("APP_PORT", 8080),
	)
}
