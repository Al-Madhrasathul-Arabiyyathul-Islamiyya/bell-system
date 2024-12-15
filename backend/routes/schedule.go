package routes

import (
	"github.com/gin-gonic/gin"
)

// ScheduleRoutes defines the schedule routes
func ScheduleRoutes() func(*gin.RouterGroup) {
	return func(rg *gin.RouterGroup) {
		rg.GET("/", GetScheduleHandler)
	}
}

// GetScheduleHandler is a placeholder response
func GetScheduleHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"schedule": "Sample Schedule Data",
	})
}
