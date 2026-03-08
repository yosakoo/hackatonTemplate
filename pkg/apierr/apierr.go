package apierr

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v5"
)

const (
	CodeAuthRequired           = "AUTH_REQUIRED"
	CodeAuthTokenInvalid       = "AUTH_TOKEN_INVALID"
	CodeAuthInvalidCredentials = "AUTH_INVALID_CREDENTIALS"
	CodeUserDeactivated        = "USER_DEACTIVATED"
	CodeAdminRequired          = "ADMIN_REQUIRED"
	CodeForbidden              = "FORBIDDEN"

	CodeUserNotFound       = "USER_NOT_FOUND"
	CodeUserDuplicateEmail = "USER_DUPLICATE_EMAIL"

	CodeValidationError = "VALIDATION_ERROR"
	CodeInternalError   = "INTERNAL_ERROR"
	CodeNotFound        = "NOT_FOUND"
)

type ErrorBody struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details any `json:"details"`
}

// Response — стандартный HTTP-ответ с ошибкой.
//
// swagger:model
type Response struct {
	Success bool       `json:"success" example:"false"`
	Error   *ErrorBody `json:"error"`
}

func New(code, message string) Response {
	return Response{
		Success: false,
		Error:   &ErrorBody{Code: code, Message: message, Details: nil},
	}
}

func Send(c *echo.Context, status int, code, message string) error {
	return c.JSON(status, New(code, message))
}

func BadRequest(c *echo.Context, code, message string) error {
	return Send(c, http.StatusBadRequest, code, message)
}

func Unauthorized(c *echo.Context, code, message string) error {
	return Send(c, http.StatusUnauthorized, code, message)
}

func Forbidden(c *echo.Context, code, message string) error {
	return Send(c, http.StatusForbidden, code, message)
}

func NotFound(c *echo.Context, code, message string) error {
	return Send(c, http.StatusNotFound, code, message)
}

func Conflict(c *echo.Context, code, message string) error {
	return Send(c, http.StatusConflict, code, message)
}

func Internal(c *echo.Context, code, message string) error {
	return Send(c, http.StatusInternalServerError, code, message)
}

// HTTPErrorHandler — глобальный обработчик ошибок для Echo.
// Подключается через e.SetHTTPErrorHandler(apierr.HTTPErrorHandler(logger)).
//
// Маппинг:
//   - *echo.HTTPError          → берёт статус и сообщение из ошибки Echo
//   - любая другая ошибка      → 500 INTERNAL_ERROR
//   - если ответ уже записан   → ничего не делает
func HTTPErrorHandler(logger *slog.Logger) func(c *echo.Context, err error) {
	return func(c *echo.Context, err error) {
		if resp, unwrapErr := echo.UnwrapResponse(c.Response()); unwrapErr == nil && resp.Committed {
			return
		}

		var he *echo.HTTPError
		if errors.As(err, &he) {
			msg := he.Message
			if msg == "" {
				msg = http.StatusText(he.Code)
			}

			_ = c.JSON(he.Code, New(httpStatusCode(he.Code), msg))
			return
		}

		logger.Error("unhandled error", "error", err,
			"method", c.Request().Method,
			"path", c.Request().URL.Path,
		)

		_ = c.JSON(http.StatusInternalServerError, New(CodeInternalError, "internal server error"))
	}
}

// httpStatusCode maps an HTTP status to an apierr code string.
func httpStatusCode(status int) string {
	switch status {
	case http.StatusBadRequest:
		return CodeValidationError
	case http.StatusUnauthorized:
		return CodeAuthRequired
	case http.StatusForbidden:
		return CodeForbidden
	case http.StatusNotFound:
		return CodeNotFound
	case http.StatusConflict:
		return CodeInternalError
	default:
		return CodeInternalError
	}
}
