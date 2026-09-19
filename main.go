package main

import (
	"log"

	"github.com/davi/omc-be/database"
	"github.com/davi/omc-be/handlers"
	"github.com/davi/omc-be/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	database.Connect()

	r := gin.Default()

	// CORS config
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	r.Use(cors.New(config))

	// Public routes
	r.POST("/register", handlers.Register)
	r.POST("/login", handlers.Login)

	// Protected routes
	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware()) // require valid token
	{
		api.GET("/me", handlers.GetMe)
		api.GET("/products", handlers.GetProducts)
		api.GET("/weapon-attachments", handlers.GetWeaponAttachments)
		api.GET("/sync", handlers.SSEHandler)

		api.GET("/sgts", handlers.GetSgts)

		// Member routes
		api.POST("/orders", handlers.CreateOrder)
		api.GET("/orders/me", handlers.GetMyOrders)

		// Admin & SGT routes
		adminOfficer := api.Group("/")
		adminOfficer.Use(middleware.AuthMiddleware("admin", "sgt"))
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
		adminOnly := api.Group("/")
		adminOnly.Use(middleware.AuthMiddleware("admin"))
		{
			adminOnly.POST("/users", handlers.CreateUser)
			adminOnly.PUT("/users/:id", handlers.UpdateUser)
			adminOnly.DELETE("/users/:id", handlers.DeleteUser)
		}
	}

	log.Println("Server running on port 8081")
	if err := r.Run(":8081"); err != nil {
		log.Fatal(err)
	}
}
