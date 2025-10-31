package routes

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
	controller "github.com/muhaba7me/coupon-meal-system/controllers"
)

func SetupUnProtectedRoutes(router *gin.Engine, client *mongo.Client) {
	public := router.Group("/api/auth")
	{
		public.POST("/login", controller.LoginUser(client))
		public.POST("/refresh", controller.RefreshTokenHandler(client))
		public.POST("/logout", controller.LogoutHandler(client))
	}
}
