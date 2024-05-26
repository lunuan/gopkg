package http

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lunuan/gopkg/http/middleware"
	"github.com/lunuan/gopkg/log"
)

type GinServer struct {
	*http.Server
}

func NewGinServer(port int) *GinServer {
	handler := gin.New()
	handler.Use(middleware.Logger())
	handler.Use(middleware.Recovery())
	s := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: handler,
	}
	return &GinServer{s}
}

func (s *GinServer) Run() error {
	log.Infof("server listen on %s", s.Addr)
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)
	go s.ListenAndServe()
	sig := <-ch
	log.Infof("receive signal %s, shutdown server", sig.String())
	// 设置为5秒超时
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// 最后关闭
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown server, %s", err.Error())
	}
	log.Infof("server shutdown successfully")
	return nil
}

func (s *GinServer) SetMode(mode string) {
	gin.SetMode(mode)
}

func (s *GinServer) GetHandler() *gin.Engine {
	return s.Handler.(*gin.Engine)
}
