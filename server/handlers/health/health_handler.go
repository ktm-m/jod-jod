package health

import (
	"fmt"
	"github.com/Montheankul-K/jod-jod/config"
	"github.com/labstack/echo/v4"
	"net/http"
)

type HealthHandler interface {
	HealthCheck(c echo.Context) error
}

type healthHandler struct {
	cfg *config.Server
}

func NewHealthHandler(cfg *config.Server) HealthHandler {
	return &healthHandler{
		cfg: cfg,
	}
}

func (h *healthHandler) HealthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, echo.Map{
		"message": fmt.Sprintf("%s server version: %s is healthy", h.cfg.Name, h.cfg.Version),
	})
}
