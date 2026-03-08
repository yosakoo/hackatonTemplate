package observability

import (
	"github.com/labstack/echo/v5"
)

const (
	PathHealth  = "/health"
	PathReady   = "/ready"
	PathMetrics = "/metrics"
)

func RegisterRoutes(e *echo.Echo, handler *Handler) {
	e.GET(PathHealth, handler.Health)
	e.GET(PathReady, handler.Ready)
	e.GET(PathMetrics, handler.Metrics)
}
