package main

import (
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/lunuan/gopkg/http"
	"github.com/lunuan/gopkg/http/middleware"
	"github.com/lunuan/gopkg/log"
	"go.uber.org/zap/zapcore"
)

func main() {
	gin.SetMode(gin.ReleaseMode)

	logConfig := &log.Config{
		Level:  "debug",
		Format: "common",
	}
	log.Init(logConfig)
	middleware.InitLogMiddleware(logConfig, &ginzap.Config{
		DefaultLevel: zapcore.DebugLevel,
	})

	server := http.NewGinServer(8080)
	rg := server.GetHandler().Group("/api")
	rg.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	if err := server.Run(); err != nil {
		log.Errorf("failed to run server, %s", err.Error())
	}

}
