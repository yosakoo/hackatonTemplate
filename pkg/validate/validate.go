package validate

import (
	"errors"
	"net/http"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"

	"hackathonTemplate/pkg/apierr"
)

var (
	once     sync.Once
	validate *validator.Validate
)

func v() *validator.Validate {
	once.Do(func() {
		validate = validator.New()
	})

	return validate
}

func Bind(c *echo.Context, dst any) error {
	if err := c.Bind(dst); err != nil {
		return apierr.BadRequest(c, apierr.CodeValidationError, err.Error())
	}

	if err := v().Struct(dst); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			fields := make([]map[string]string, 0, len(ve))
			for _, fe := range ve {
				fields = append(fields, map[string]string{
					"field":   fe.Field(),
					"tag":     fe.Tag(),
					"value":   fe.Param(),
				})
			}

			return c.JSON(http.StatusUnprocessableEntity, map[string]any{
				"success": false,
				"error": map[string]any{
					"code":    apierr.CodeValidationError,
					"message": "validation failed",
					"details": fields,
				},
			})
		}

		return apierr.BadRequest(c, apierr.CodeValidationError, err.Error())
	}

	return nil
}
