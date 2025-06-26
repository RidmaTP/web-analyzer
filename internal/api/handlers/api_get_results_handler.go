package handlers

import (
	"fmt"
	"runtime"
	"time"

	"github.com/RidmaTP/web-analyzer/internal/configs"
	"github.com/RidmaTP/web-analyzer/internal/analyzers"
	"github.com/RidmaTP/web-analyzer/internal/fetcher"
	"github.com/RidmaTP/web-analyzer/internal/models"
	"github.com/RidmaTP/web-analyzer/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
)

// Gin Api handler used to get a url
// sends a text/event-stream in http1.1
func GetResultsHandler(c *gin.Context) {

	url := c.Query("url")

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Flush()

	ctx := c.Request.Context()
	a := analyzers.BodyAnalyzer{
		Fetcher: &fetcher.Fetcher{},
		Stream:  make(chan string, 20),
		Output:  models.Output{},
		Workers: runtime.NumCPU(),
	}

	errObj := utils.UrlValidationCheck(url)
	if errObj != nil {
		fmt.Fprintf(c.Writer, "data: %s\n\n", *utils.ErrStreamObj(*errObj))
		c.Writer.Flush()
		return
	}

	//checking cache for results for the given url
	cacheObj := configs.GetCacheConfig()

	if cachedData, found := cacheObj.Get(url); found {
		a := cachedData.(string)
		fmt.Fprintf(c.Writer, "data: %s\n\n", a)
		c.Writer.Flush()
		return
	}

	errChan := make(chan *models.ErrorOut)

	go func(cache *cache.Cache) {
		defer close(a.Stream)
		errObj := a.Analyze(url)
		if errObj != nil {
			errChan <- errObj
		}
		strObj, _ := utils.JsonToText(a.Output)
		cache.Set(url, *strObj, 2*time.Hour)

	}(cacheObj)
	for {
		select {
		//used to push the error
		case errObj := <-errChan:
			if errObj != nil {
				fmt.Fprintf(c.Writer, "data: %s\n\n", *utils.ErrStreamObj(*errObj))
				c.Writer.Flush()

				return
			}
		case <-ctx.Done():
			fmt.Println("client disconnected")
			return
		case msg, ok := <-a.Stream:
			if !ok {
				return
			}
			fmt.Fprintf(c.Writer, "data: %s\n\n", msg)
			c.Writer.Flush()
		}
	}
}
