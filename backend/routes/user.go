package routes

import (
	"backend/middlewares"

	"github.com/gin-gonic/gin"
)

// UserRoutes defines the user management routes
func UserRoutes() func(*gin.RouterGroup) {
	return func(rg *gin.RouterGroup) {
		rg.GET("/", middlewares.AuthMiddleware(), GetAllUsersHandler)
		rg.POST("/", middlewares.AuthMiddleware(), CreateUserHandler)
		rg.PATCH("/:id", middlewares.AuthMiddleware(), UpdateUserPasswordHandler)
		rg.DELETE("/:id", middlewares.AuthMiddleware(), DeleteUserHandler)
	}
	// Note: AuthMiddleware() is applied individually to ensure flexibility.
}
