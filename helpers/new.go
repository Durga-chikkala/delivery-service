package helpers

import (
	"github.com/Durga-Chikkala/delivery-service/models"
	"os"
)

func New() *models.Helpers {
	LoadConfigs()
	logger := InitializeLogger()
	db := InitializeMongo(logger)
	redisDB := initializeRedis(logger)
	metrics := NewMetrics()

	appName := os.Getenv("APP_NAME")
	if appName == "" {
		appName = "delivery-service"
	}

	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8000"
	}

	return &models.Helpers{AppName: appName, AppPort: port, DB: db, Redis: redisDB, Metrics: metrics, Logger: logger}
}
