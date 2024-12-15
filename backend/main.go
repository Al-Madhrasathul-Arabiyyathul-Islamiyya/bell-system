package main

import (
	"github.com/gin-gonic/gin"

	"backend/db"
	"backend/routes"
)

func main() {
	// Initialize Gin router
	r := gin.Default()

	// Initialize database connection
	db.ConnectDB()

	// Create admin user
	db.CreateAdminUser()

	// Set up routes
	authRoutes := routes.AuthRoutes()
	scheduleRoutes := routes.ScheduleRoutes()

	// Group routes under `/api`
	api := r.Group("/api")
	{
		apiGroupAuth := api.Group("/auth")
		{
			authRoutes(apiGroupAuth)
		}

		apiGroupSchedule := api.Group("/schedule")
		{
			scheduleRoutes(apiGroupSchedule)
		}

		apiGroupUsers := api.Group("/users")
		{
			routes.UserRoutes()(apiGroupUsers)
		}
	}

	// Start the server
	r.Run(":8080")
}
