package main

import (
	"log"

	"github.com/RidmaTP/web-analyzer/internal/api"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	api.Router(r)
	err := r.Run(":8000")
	if err != nil {
		log.Fatal("unable to initialize server")
	}
}
