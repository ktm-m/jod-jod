package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/Montheankul-K/jod-jod/config"
	"github.com/Montheankul-K/jod-jod/db"
	"github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Server interface {
	Start() error
}

type server struct {
	app         *echo.Echo
	db          db.DB
	cfg         *config.Config
	redisClient *redis.Client
}

var (
	srv  *server
	once sync.Once
)

func InitServer(cfg *config.Config, db db.DB) Server {
	app := echo.New()
	redisClient := &redis.Client{}
	once.Do(func() {
		srv = &server{
			app:         app,
			db:          db,
			cfg:         cfg,
			redisClient: redisClient,
		}
	})
	return srv
}

func (s *server) Start() error {
	timeoutMiddleware := setTimeoutMiddleware(s.cfg.Server.Timeout)
	s.app.Use(timeoutMiddleware)
	s.app.Use(middleware.Logger())
	s.app.Use(middleware.Recover())

	s.healthCheckRouter()
	s.userRouter()
	s.transactionRouter()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
	go s.gracefullyShutdown(shutdown)
	err := s.listenAndServe()
	if err != nil {
		return err
	}
	return nil
}

func setTimeoutMiddleware(timeout time.Duration) echo.MiddlewareFunc {
	return middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Skipper:      middleware.DefaultSkipper,
		ErrorMessage: "error: request timeout",
		Timeout:      timeout * time.Second,
	})
}

func (s *server) listenAndServe() error {
	port := fmt.Sprintf(":%d", s.cfg.Server.Port)
	if err := s.app.Start(port); err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.app.Logger.Fatal("cannot start server")
		return errors.New("cannot start server")
	}
	s.app.Logger.Infof("server started at %s", port)
	return nil
}

func (s *server) gracefullyShutdown(shutdown <-chan os.Signal) {
	<-shutdown
	s.app.Logger.Info("shutting down the server")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := s.app.Shutdown(ctx); err != nil {
		s.app.Logger.Fatal("cannot gracefully shutdown the server")
	}
}
