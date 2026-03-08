package middleware

import (
	"bytes"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

const RequestIDHeader = "X-Request-ID"

func Logging(logger *slog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			start := time.Now()

			requestID := c.Request().Header.Get(RequestIDHeader)
			if requestID == "" {
				requestID = uuid.New().String()
			}

			c.Response().Header().Set(RequestIDHeader, requestID)

			var requestBody string

			if c.Request().Body != nil {
				bodyBytes, err := io.ReadAll(c.Request().Body)
				if err == nil {
					requestBody = string(bodyBytes)
					c.Request().Body = io.NopCloser(bytes.NewReader(bodyBytes))
				}
			}

			err := next(c)

			var status int
			if resp, unwrapErr := echo.UnwrapResponse(c.Response()); unwrapErr == nil {
				status = resp.Status
			}

			if status == 0 {
				status = 200
			}

			durationMs := time.Since(start).Milliseconds()

			level := slog.LevelInfo
			if status >= 400 && status < 500 {
				level = slog.LevelWarn
			} else if status >= 500 {
				level = slog.LevelError
			}

			maxBodyLength := 1000
			if len(requestBody) > maxBodyLength {
				requestBody = requestBody[:maxBodyLength] + "..."
			}

			requestBody = strings.ReplaceAll(strings.TrimSpace(requestBody), "\n", " ")

			logger.Log(c.Request().Context(), level, "HTTP Request",
				slog.String("request_id", requestID),
				slog.String("method", c.Request().Method),
				slog.String("path", c.Request().RequestURI),
				slog.Int("status", status),
				slog.Int64("duration_ms", durationMs),
				slog.String("ip", c.RealIP()),
				slog.String("user_agent", c.Request().UserAgent()),
				slog.String("request_body", requestBody),
			)

			return err
		}
	}
}
