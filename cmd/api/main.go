package main

import (
	"github.com/ebubekir/event-stream/pkg/config"
	"github.com/gin-gonic/gin"
)

func main() {

	_ := config.Read()

	api := gin.Default()
	// TODO: Custom recovery endpoint
	// TODO: Graceful shutdown
	// TODO: Logging
	api.Use(gin.Logger())

	if err := api.Run(":8080"); err != nil {
		panic(err)
	}
}
