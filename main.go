package main

import (
	"github.com/gin-gonic/gin"
	"github.com/lunuan/gopkg/http/middleware"
	"github.com/lunuan/gopkg/log"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	log.Init(&log.Config{
		Level:  "debug",
		Format: "common",
	})
	log.Infow("test", "test", "test\n222")
	r.Use(middleware.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.Recovery())

	r.GET("/ping", func(c *gin.Context) {
		panic("test")
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.Run(":8080")
}
