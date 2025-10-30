package routes
import (
	"github.com/gin-gonic/gin"
	controller "github.com/muhaba7me/coupon-meal-system/controllers"
	"github.com/muhaba7me/coupon-meal-system/middleware"
	"go.mongodb.org/mongo-driver/v2/mongo"
)
func SetupProtectedRoutes(router *gin.Engine, client *mongo.Client) {
	router.Use(middleware.AuthMiddleware())
	router.POST("/employee", controller.CreateEmployee(client))
	router.POST("/qr-codes/generate", controller.GenerateQrCode(client))
}