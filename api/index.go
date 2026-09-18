package api

import (
	"net/http"

	"github.com/davi/omc-be/database"
	"github.com/davi/omc-be/handlers"
	"github.com/davi/omc-be/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var app *gin.Engine

func init() {
	database.Connect()

	r := gin.Default()

	// CORS config
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	r.Use(cors.New(config))

	// Public routes
	r.POST("/register", handlers.Register)
	r.POST("/login", handlers.Login)

	// Protected routes
	apiGroup := r.Group("/api")
	apiGroup.Use(middleware.AuthMiddleware()) // require valid token
	{
		apiGroup.GET("/me", handlers.GetMe)
		apiGroup.GET("/products", handlers.GetProducts)
		apiGroup.GET("/weapon-attachments", handlers.GetWeaponAttachments)
		apiGroup.GET("/sync", handlers.SSEHandler)
		apiGroup.GET("/sgts", handlers.GetSgts)

		// Member routes
		apiGroup.POST("/orders", handlers.CreateOrder)
		apiGroup.GET("/orders/me", handlers.GetMyOrders)

		// Admin & SGT routes
		adminOfficer := apiGroup.Group("/")
		adminOfficer.Use(middleware.AuthMiddleware("admin", "sgt", "officer"))
		{
			adminOfficer.POST("/products", handlers.CreateProduct)
			adminOfficer.PUT("/products/:id", handlers.UpdateProduct)
			adminOfficer.DELETE("/products/:id", handlers.DeleteProduct)

			adminOfficer.GET("/orders", handlers.GetAllOrders)
			adminOfficer.PUT("/orders/:id/status", handlers.UpdateOrderStatus)
			adminOfficer.DELETE("/orders/:id", handlers.DeleteOrder)
			adminOfficer.GET("/analytics", handlers.GetAnalytics)
			adminOfficer.GET("/users", handlers.GetAllUsers)
		}

		// Admin only routes
		adminOnly := apiGroup.Group("/")
		adminOnly.Use(middleware.AuthMiddleware("admin"))
		{
			adminOnly.POST("/users", handlers.CreateUser)
			adminOnly.PUT("/users/:id", handlers.UpdateUserRole)
			adminOnly.DELETE("/users/:id", handlers.DeleteUser)
		}
	}

	app = r
}

func Handler(w http.ResponseWriter, r *http.Request) {
	app.ServeHTTP(w, r)
}
