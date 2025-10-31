package routes

import (
	"github.com/gin-gonic/gin"
	controller "github.com/muhaba7me/coupon-meal-system/controllers"
	"github.com/muhaba7me/coupon-meal-system/middleware"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// SetupProtectedRoutes registers all routes that require authentication
func SetupProtectedRoutes(router *gin.Engine, client *mongo.Client) {
	api := router.Group("/api")
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware())

	// =======================================
	// 👑 ADMIN ROUTES
	// =======================================
	admin := protected.Group("/admin")
	admin.Use(middleware.RoleMiddleware("ADMIN"))
	{
		// ✅ Admin-only: Register new users (employees/suppliers)
		admin.POST("/register", controller.RegisterUser(client))

		// --- Employees Management ---
		employees := admin.Group("/employees")
		{
			employees.POST("", controller.CreateEmployee(client))
			employees.GET("", controller.GetAllEmployees(client))
			employees.GET("/:id", controller.GetEmployeeByID(client))
			employees.GET("/code/:code", controller.GetEmployeeByCode(client))
			employees.PATCH("/:id", controller.UpdateEmployee(client))
		}

		// --- Suppliers Management ---
		suppliers := admin.Group("/suppliers")
		{
			suppliers.POST("", controller.CreateSupplier(client))
			suppliers.GET("", controller.GetAllSuppliers(client))
			suppliers.GET("/:id", controller.GetSupplierByID(client))
			suppliers.PATCH("/:id", controller.UpdateSupplier(client))
			suppliers.PATCH("/:id/activate", controller.ActivateSupplier(client))
			// suppliers.PATCH("/:id/verify", controller.Ve(client))
		}
	}

	// =======================================
	// 👷 EMPLOYEE ROUTES
	// =======================================
	employee := protected.Group("/employee")
	{
		employee.GET("/profile", controller.GetMyProfile(client))
		employee.GET("/balance", controller.GetMyBalance(client))

		// --- QR Codes ---
		qr := employee.Group("/qr-codes")
		{
			qr.POST("/generate", controller.GenerateQrCode(client))
			qr.GET("/history", controller.GetMyQRCodes(client))
		}

		// --- Transactions ---
		employee.GET("/transactions", controller.GetMyTransactions(client))
		employee.GET("/transactions/pending", controller.GetPendingTransactions(client))
		employee.POST("/transactions/approve", controller.ApproveTransaction(client))
	}

	// =======================================
	// 🏭 SUPPLIER ROUTES
	// =======================================
	supplier := protected.Group("/supplier")
	supplier.Use(middleware.RoleMiddleware("SUPPLIER", "ADMIN"))
	{
		supplier.GET("/profile", controller.GetMySupplierProfile(client))
		supplier.GET("/totals", controller.GetMyTotals(client))

		supplier.POST("/validate-qr", controller.ValidateQRcode(client))

		transactions := supplier.Group("/transactions")
		{
			transactions.POST("/initiate", controller.InitiateTransaction(client))
			transactions.GET("/daily", controller.GetSupplierTransactions(client))
			transactions.GET("/monthly", controller.GetSupplierTransactions(client))
		}
	}
}
