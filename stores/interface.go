package stores

import (
	"github.com/Durga-Chikkala/delivery-service/models"
	"github.com/gin-gonic/gin"
)

type Delivery interface {
	Get(ctx *gin.Context, dimensions *models.Dimension) (*[]models.Response, error)
}
