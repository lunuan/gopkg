package middleware

import (
	"bytes"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"
	"time"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/lunuan/gopkg/conv"
	"github.com/lunuan/gopkg/log"
	"go.uber.org/zap"
)

var logger *zap.Logger

// init init logger use default config
func init() {
	def := &log.Config{
		Level:  "debug",
		Format: "json",
	}
	InitLoggerMiddleware(def)
}

func InitLoggerMiddleware(cfg *log.Config) {
	logger = log.NewLogger(cfg)
}

func Logger() gin.HandlerFunc {
	return ginzap.Ginzap(logger, time.RFC3339, true)
}

func Recovery() gin.HandlerFunc {
	return customRecoveryWithZap(logger, true, defaultHandleRecovery)
}

func defaultHandleRecovery(c *gin.Context, err interface{}) {
	c.AbortWithStatus(http.StatusInternalServerError)
}

// customRecoveryWithZap returns a gin.HandlerFunc (middleware) with a custom recovery handler
// that recovers from any panics and logs requests using uber-go/zap.
// All errors are logged using zap.Error().
// stack means whether output the stack info.
// The stack info is easy to find where the error occurs but the stack info is too large.
func customRecoveryWithZap(logger ginzap.ZapLogger, stack bool, recovery gin.RecoveryFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Check for a broken connection, as it is not really a
				// condition that warrants a panic stack trace.
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") ||
							strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				if brokenPipe {
					httpRequest, _ := httputil.DumpRequest(c.Request, false)
					logger.Error(c.Request.URL.Path,
						zap.Any("error", err),
						zap.String("request", conv.BytesToString(httpRequest)),
					)
					// If the connection is dead, we can't write a status to it.
					c.Error(err.(error)) //nolint: errcheck
					c.Abort()
					return
				}

				buf := &bytes.Buffer{}
				buf.WriteString("recovery from panic, ")
				if stack {
					buf.WriteString(conv.BytesToString(debug.Stack()))
				}
				logger.Error(buf.String(),
					zap.String("method", c.Request.Method),
					zap.String("path", c.Request.URL.Path),
					zap.String("ip", c.ClientIP()),
					zap.String("user-agent", c.Request.UserAgent()),
					zap.Any("error", err),
				)
				recovery(c, err)
			}
		}()
		c.Next()
	}
}
