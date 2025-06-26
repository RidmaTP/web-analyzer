package api

import (
	"net/http"

	"github.com/RidmaTP/web-analyzer/internal/configs"
	"github.com/RidmaTP/web-analyzer/internal/api/handlers"
	"github.com/RidmaTP/web-analyzer/internal/api/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

//defines all the routes
func Router(engine *gin.Engine) {
	engine.Use(cors.Default())
	
	group := engine.Group("/api/")
	group.Use(middleware.LogMiddleware)
	routes(group)
}

func routes(rg *gin.RouterGroup) {
	rg.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "service up and running",
			"status":  "success",
			"version": configs.GetAppVersion()})
	})

	rg.GET("/result" , handlers.GetResultsHandler)

}
