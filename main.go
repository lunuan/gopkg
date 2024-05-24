package main

import (
	"github.com/gin-gonic/gin"
	"github.com/lunuan/gopkg/http/middleware"
	"github.com/lunuan/gopkg/log"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	logConfig := &log.Config{
		Level:  "debug",
		Format: "common",
	}
	log.Init(logConfig)
	middleware.InitLoggerMiddleware(logConfig)
	r.Use(middleware.Logger())
	// r.Use(gin.Recovery())
	r.Use(middleware.Recovery())

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	if err := r.Run(":8080"); err != nil {
		panic(err)
	}
}
