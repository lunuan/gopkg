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
	"go.uber.org/zap/zapcore"
)

var ginzapConfig *ginzap.Config
var logger *zap.Logger

// init init logger use default config
func init() {
	cfg := &log.Config{
		Level:  "debug",
		Format: "common",
	}
	InitLogMiddleware(cfg, &ginzap.Config{})
}

func InitLogMiddleware(cfg *log.Config, GinzapConfig *ginzap.Config) {
	ginzapConfig = GinzapConfig
	logger = log.NewLogger(cfg)
}

func Logger() gin.HandlerFunc {
	return GinzapWithConfig(logger, ginzapConfig)
}

// GinzapWithConfig returns a gin.HandlerFunc using configs
func GinzapWithConfig(logger ginzap.ZapLogger, conf *ginzap.Config) gin.HandlerFunc {
	skipPaths := make(map[string]bool, len(conf.SkipPaths))
	for _, path := range conf.SkipPaths {
		skipPaths[path] = true
	}

	return func(c *gin.Context) {
		start := time.Now()
		// some evil middlewares modify this values
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		c.Next()
		track := true

		if _, ok := skipPaths[path]; ok || (conf.Skipper != nil && conf.Skipper(c)) {
			track = false
		}

		if track && len(conf.SkipPathRegexps) > 0 {
			for _, reg := range conf.SkipPathRegexps {
				if !reg.MatchString(path) {
					continue
				}

				track = false
				break
			}
		}

		if track {
			end := time.Now()
			duration := end.Sub(start) * 1000
			// if conf.UTC {
			// 	end = end.UTC()
			// }

			fields := []zapcore.Field{
				zap.Int("status", c.Writer.Status()),
				zap.String("method", c.Request.Method),
				zap.String("path", path),
				zap.Duration("duration", duration),
				zap.String("ip", c.ClientIP()),
				zap.String("user-agent", c.Request.UserAgent()),
			}
			if query != "" {
				fields = append(fields, zap.String("query", query))
			}
			// if conf.TimeFormat != "" {
			// 	fields = append(fields, zap.String("time", end.Format(conf.TimeFormat)))
			// }

			if conf.Context != nil {
				fields = append(fields, conf.Context(c)...)
			}

			if len(c.Errors) > 0 {
				// Append error field if this is an erroneous request.
				for _, e := range c.Errors.Errors() {
					logger.Error(e, fields...)
				}
			} else {
				if zl, ok := logger.(*zap.Logger); ok {
					zl.Log(conf.DefaultLevel, "", fields...)
				} else if conf.DefaultLevel == zapcore.InfoLevel {
					logger.Info(path, fields...)
				} else {
					logger.Error(path, fields...)
				}
			}
		}
	}
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
