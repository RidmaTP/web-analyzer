package api

import (
	"github.com/RidmaTP/web-analyzer/configs"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"net/http"
)

func Router(engine *gin.Engine) {
	engine.Use(cors.Default())

	group := engine.Group("/api/")
	routes(group)
}

func routes(rg *gin.RouterGroup) {
	rg.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "service up and running",
			"status":  "success",
			"version": configs.APP_VERSION})
	})

}
