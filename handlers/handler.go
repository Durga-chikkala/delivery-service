package handlers

import (
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/Durga-Chikkala/delivery-service/constants"
	"github.com/Durga-Chikkala/delivery-service/helpers"
	"github.com/Durga-Chikkala/delivery-service/models"
	"github.com/Durga-Chikkala/delivery-service/services"
)

type Handler struct {
	services.Delivery
	ErrorMetrics *prometheus.CounterVec
}

func New(svc services.Delivery, errorMetrics *prometheus.CounterVec) Handler {
	return Handler{Delivery: svc, ErrorMetrics: errorMetrics}
}

func (h *Handler) Get(ctx *gin.Context) {
	appID := ctx.Query(constants.App)
	if strings.TrimSpace(appID) == "" {
		h.ErrorMetrics.WithLabelValues(ctx.Request.Method, ctx.Request.URL.Path, strconv.Itoa(http.StatusBadRequest)).Inc()
		ctx.JSON(helpers.ParseError(&helpers.Error{StatusCode: http.StatusBadRequest,
			Code: "Invalid Param", Reason: "Parameter app is required"}))
		return
	}

	country := ctx.Query(constants.Country)
	if strings.TrimSpace(country) == "" {
		h.ErrorMetrics.WithLabelValues(ctx.Request.Method, ctx.Request.URL.Path, strconv.Itoa(http.StatusBadRequest)).Inc()
		ctx.JSON(helpers.ParseError(&helpers.Error{StatusCode: http.StatusBadRequest,
			Code: "Invalid Param", Reason: "Parameter country is required"}))
		return
	}

	os := ctx.Query(constants.Os)
	if strings.TrimSpace(os) == "" {
		h.ErrorMetrics.WithLabelValues(ctx.Request.Method, ctx.Request.URL.Path, strconv.Itoa(http.StatusBadRequest)).Inc()
		ctx.JSON(helpers.ParseError(&helpers.Error{StatusCode: http.StatusBadRequest,
			Code: "Invalid Param", Reason: "Parameter os is required"}))
		return
	}

	d := &models.Dimension{APPID: appID, Country: country, OS: os}
	campaigns, err := h.Delivery.Get(ctx, d)
	if err != nil {
		statusCode, err := helpers.ParseError(err)
		h.ErrorMetrics.WithLabelValues(ctx.Request.Method, ctx.Request.URL.Path, strconv.Itoa(statusCode)).Inc()
		ctx.JSON(statusCode, err)
		return
	}

	if campaigns == nil || len(*campaigns) == 0 {
		ctx.JSON(http.StatusNoContent, nil)
		return
	}

	ctx.JSON(http.StatusOK, helpers.FormResponse(campaigns))
}
