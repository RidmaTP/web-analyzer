package main

import (
	"log"

	"github.com/RidmaTP/web-analyzer/internal/configs"
	"github.com/RidmaTP/web-analyzer/internal/api"
	"github.com/gin-gonic/gin"
)

// initiating gin server on default port 8000
// port can be changed from configs/.env
func main() {
	r := gin.Default()

	err := configs.LoadEnv()
	if err !=nil{
		log.Fatal(err)
	}
	configs.LoadLogger()
	configs.LoadCacheConfig()
	api.Router(r)
	err = r.Run(":" + configs.GetPort())
	if err != nil {
		log.Fatal("unable to initialize server")
	}
}
