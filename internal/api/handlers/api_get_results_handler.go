package handlers

import (
	"fmt"
	"runtime"

	"github.com/RidmaTP/web-analyzer/internal/analyzers"
	"github.com/RidmaTP/web-analyzer/internal/fetcher"
	"github.com/RidmaTP/web-analyzer/internal/models"
	"github.com/RidmaTP/web-analyzer/internal/utils"
	"github.com/gin-gonic/gin"
)

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
	errChan := make(chan error)
	go func() {
		defer close(a.Stream)
		err := a.Analyze(url)
		errChan <- err
	}()
	for {
		select {
		case err := <-errChan:
			if err != nil {
				fmt.Fprintf(c.Writer, "data: %s\n\n", *utils.ErrStreamObj(err.Error()))
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

