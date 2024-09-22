package main

import (
	"github.com/gin-gonic/gin"

	"github.com/Durga-Chikkala/delivery-service/handlers"
	"github.com/Durga-Chikkala/delivery-service/helpers"
	"github.com/Durga-Chikkala/delivery-service/middlewares"
	"github.com/Durga-Chikkala/delivery-service/services"
	"github.com/Durga-Chikkala/delivery-service/stores"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	router := gin.Default()
	helper := helpers.New()

	// middlewares
	middlewareMetrics := middlewares.Metrics{RequestCount: helper.Metrics.RequestCounter, RequestDuration: helper.Metrics.RequestDuration}
	router.Use(middlewares.CORS(), middlewareMetrics.MetricsMiddleware())

	// Injections
	store := stores.New(helper.DB, helper.Redis, helper.Logger, helper.Metrics.CacheHits, helper.Metrics.CacheMisses)
	svc := services.New(&store)
	handler := handlers.New(svc, helper.Metrics.ErrorCounter)

	// Endpoints
	router.GET("/v1/delivery", handler.Get)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	err := router.Run(":" + helper.AppPort)
	if err != nil {
		helper.Logger.Error("Error While Running the Service", "Error", err.Error())
		return
	}

}
