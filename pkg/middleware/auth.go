package middleware

import (
	"strings"

	"github.com/labstack/echo/v5"

	"hackathonTemplate/pkg/apierr"
)

const UserIDKey = "user_id"

type TokenParser interface {
	ParseAccessToken(token string) (string, error)
}

func JWTAuth(parser TokenParser) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			auth := c.Request().Header.Get("Authorization")
			if !strings.HasPrefix(auth, "Bearer ") {
				return apierr.Unauthorized(c, apierr.CodeAuthRequired, "authorization header required")
			}

			token := strings.TrimPrefix(auth, "Bearer ")

			userID, err := parser.ParseAccessToken(token)
			if err != nil {
				return apierr.Unauthorized(c, apierr.CodeAuthTokenInvalid, "invalid or expired token")
			}

			c.Set(UserIDKey, userID)

			return next(c)
		}
	}
}

func UserIDFromContext(c *echo.Context) string {
	id, _ := c.Get(UserIDKey).(string)
	return id
}
