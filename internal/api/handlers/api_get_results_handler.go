package handlers

import (
	"net/http"

	"github.com/RidmaTP/web-analyzer/internal/analyzers"
	"github.com/RidmaTP/web-analyzer/internal/fetcher"
	"github.com/RidmaTP/web-analyzer/internal/models"
	"github.com/RidmaTP/web-analyzer/internal/utils"
	"github.com/gin-gonic/gin"
)

func GetResultsHandler(c *gin.Context) {
	input := models.Input{}
	err := c.ShouldBindJSON(&input)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.SendErrResponse(err))
	}
	f := fetcher.Fetcher{}
	err = analyzers.Analyze(input.Url, &f)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.SendErrResponse(err))
	}

	c.JSON(http.StatusOK, map[string]interface{}{})
}
