package routes

import (
	"github.com/gin-gonic/gin"
	controller "github.com/muhaba7me/coupon-meal-system/controllers"
	"github.com/muhaba7me/coupon-meal-system/middleware"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// SetupProtectedRoutes registers all authenticated (protected) routes
func SetupProtectedRoutes(router *gin.Engine, client *mongo.Client) {
	// Apply AuthMiddleware to all routes in this group
	protected := router.Group("/")
	protected.Use(middleware.AuthMiddleware())

	// ========================================
	// EMPLOYEE ROUTES
	// ========================================
	router.POST("/register", controller.RegisterUser(client))
	employee := protected.Group("/employees")
	{
		employee.POST("/", controller.CreateEmployee(client))
		employee.GET("/", controller.GetAllEmployees(client))
		employee.GET("/id/:id", controller.GetEmployeeByID(client))       
		employee.GET("/code/:code", controller.GetEmployeeByCode(client)) 
		employee.PATCH("/:id", controller.UpdateEmployee(client))
		employee.GET("/profile/me", controller.GetMyProfile(client))
		employee.GET("/balance/me", controller.GetMyBalance(client))
		// QR Code Management
		qr := employee.Group("/qr-codes")
		{
			qr.POST("/generate", controller.GenerateQrCode(client))
		}
	}
}
