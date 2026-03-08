package observability

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v5"
	"hackathonTemplate/pkg/metrics"
)

type ReadinessChecker interface {
	Ready(ctx context.Context) error
}

type Handler struct {
	readyChecker ReadinessChecker
}

func NewHandler(readyChecker ReadinessChecker) *Handler {
	return &Handler{readyChecker: readyChecker}
}

// Health возвращает статус живости сервиса (liveness).
//
// @Summary      Liveness
// @Description  Проверка живости сервиса. Всегда возвращает 200 OK, если сервер отвечает.
// @Tags         observability
// @Produce      plain
// @Success      200  {string}  string  "OK"
// @Router       /health [get]
func (h *Handler) Health(c *echo.Context) error {
	return c.String(http.StatusOK, "OK")
}

// Ready возвращает статус готовности сервиса (readiness).
//
// @Summary      Readiness
// @Description  Проверка готовности сервиса к приёму трафика (в т.ч. доступность БД).
// @Tags         observability
// @Produce      plain
// @Success      200   {string}  string  "OK"
// @Failure      503  {string}  string  "not ready"
// @Router       /ready [get]
func (h *Handler) Ready(c *echo.Context) error {
	if err := h.readyChecker.Ready(c.Request().Context()); err != nil {
		return c.String(http.StatusServiceUnavailable, "not ready")
	}

	return c.String(http.StatusOK, "OK")
}

// Metrics экспортирует метрики в формате Prometheus text/plain.
//
// @Summary      Prometheus metrics
// @Description  Экспорт счётчиков платформы в формате Prometheus text/plain (B9-3).
// @Tags         observability
// @Produce      plain
// @Success      200  {string}  string  "Prometheus text/plain metrics"
// @Router       /metrics [get]
func (h *Handler) Metrics(c *echo.Context) error {
	w := c.Response()
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	metrics.Global.WritePrometheusText(w)

	return nil
}
