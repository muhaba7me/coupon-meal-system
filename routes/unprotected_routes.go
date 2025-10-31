package routes

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
     controller "github.com/muhaba7me/coupon-meal-system/controllers"
)

func SetupUnProtectedRoutes(router *gin.Engine, client *mongo.Client) {
	router.POST("/login", controller.LoginUser(client))
}