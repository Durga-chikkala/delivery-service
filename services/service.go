package services

import (
	"github.com/gin-gonic/gin"
	"strings"

	"github.com/Durga-Chikkala/delivery-service/models"
	"github.com/Durga-Chikkala/delivery-service/stores"
)

type Service struct {
	stores.Delivery
}

func New(store stores.Delivery) Service {
	return Service{Delivery: store}
}

func (s Service) Get(ctx *gin.Context, dimensions *models.Dimension) (*[]models.Response, error) {
	convertDimensionsToLowerCase(dimensions)
	return s.Delivery.Get(ctx, dimensions)
}

func convertDimensionsToLowerCase(dimensions *models.Dimension) {
	dimensions.APPID = strings.ToLower(dimensions.APPID)
	dimensions.Country = strings.ToLower(dimensions.Country)
	dimensions.OS = strings.ToLower(dimensions.OS)
}
