package auth

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

// AdminMiddleware returns an Echo middleware that requires admin user type.
// This should be used AFTER the regular Middleware, as it expects
// the X-User-Type header to be set.
func AdminMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			userType := c.Request().Header.Get(HeaderUserType)
			if userType != "admin" {
				return echo.NewHTTPError(http.StatusForbidden, "admin access required")
			}
			return next(c)
		}
	}
}
