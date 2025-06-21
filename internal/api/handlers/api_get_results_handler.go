package handlers

import (
	"net/http"

	"github.com/RidmaTP/web-analyzer/internal/models"
	"github.com/RidmaTP/web-analyzer/internal/utils"
	"github.com/gin-gonic/gin"
)

func GetResultsHandler(c *gin.Context) {
	input := models.Input{}
	err := c.Bind(&input)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.SendErrResponse(err))
	}

	//fetch & analyze

	c.JSON(http.StatusOK, map[string]interface{}{})
}
