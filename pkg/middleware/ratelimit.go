package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"

	ratelimiter "hackathonTemplate/pkg/limiter"
)

func RateLimitByIP(client ratelimiter.Client, keyPrefix string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			key := strings.TrimSuffix(keyPrefix, ":") + ":ip:" + c.RealIP()
			if !client.IsAllowed(c.Request().Context(), key) {
				return c.JSON(http.StatusTooManyRequests, map[string]any{
					"error": map[string]any{
						"code":    "TOO_MANY_REQUESTS",
						"message": "rate limit exceeded",
					},
				})
			}

			return next(c)
		}
	}
}

// RateLimitByPath ограничивает частоту запросов по пути (глобально на путь). keyPrefix — префикс ключа (например "rl:").
func RateLimitByPath(client ratelimiter.Client, keyPrefix string, path string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			key := strings.TrimSuffix(keyPrefix, ":") + ":path:" + path
			if !client.IsAllowed(c.Request().Context(), key) {
				return c.JSON(http.StatusTooManyRequests, map[string]any{
					"error": map[string]any{
						"code":    "TOO_MANY_REQUESTS",
						"message": "rate limit exceeded",
					},
				})
			}

			return next(c)
		}
	}
}

func RateLimitByIPPerPath(clients map[string]ratelimiter.Client, keyPrefix string) echo.MiddlewareFunc {
	prefix := strings.TrimSuffix(keyPrefix, ":")

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			path := c.Request().URL.Path

			client, ok := clients[path]
			if !ok {
				return next(c)
			}

			key := prefix + ":ip:" + c.RealIP() + ":path:" + path
			if !client.IsAllowed(c.Request().Context(), key) {
				return c.JSON(http.StatusTooManyRequests, map[string]any{
					"error": map[string]any{
						"code":    "TOO_MANY_REQUESTS",
						"message": "rate limit exceeded",
					},
				})
			}

			return next(c)
		}
	}
}
