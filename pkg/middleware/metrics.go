package middleware

import (
	"fmt"

	"github.com/labstack/echo/v5"

	"hackathonTemplate/pkg/metrics"
)

// HTTPMetrics инкрементирует ab_http_requests_total{method, path, status} для каждого запроса.
func HTTPMetrics() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			err := next(c)

			status := 200
			if resp, unwrapErr := echo.UnwrapResponse(c.Response()); unwrapErr == nil && resp.Status != 0 {
				status = resp.Status
			}

			metrics.HTTPRequestsTotal(
				c.Request().Method,
				c.Request().URL.Path,
				fmt.Sprintf("%d", status),
			).Inc()

			return err
		}
	}
}
