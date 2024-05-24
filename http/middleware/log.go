package middleware

import (
	"time"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/lunuan/gopkg/log"
	"go.uber.org/zap"
)

var logger *zap.Logger

func init() {
	logger = log.NewLogger(&log.Config{
		Level:  "debug",
		Format: "common",
	})
}

func Logger() gin.HandlerFunc {
	return ginzap.Ginzap(logger, time.RFC3339, true)
}

func Recovery() gin.HandlerFunc {
	return ginzap.RecoveryWithZap(logger, true)
}
