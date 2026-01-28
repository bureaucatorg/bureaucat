package auth

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"
)

const (
	// HeaderUserID is the header name for the user ID.
	HeaderUserID = "X-User-ID"
	// HeaderUsername is the header name for the username.
	HeaderUsername = "X-Username"
	// HeaderUserType is the header name for the user type.
	HeaderUserType = "X-User-Type"
)

// Middleware returns an Echo middleware that validates JWT tokens.
func Middleware(manager *Manager) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing authorization header")
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid authorization header format")
			}

			tokenString := parts[1]
			claims, err := manager.ValidateAccessToken(tokenString)
			if err != nil {
				if err == ErrExpiredToken {
					return echo.NewHTTPError(http.StatusUnauthorized, "token has expired")
				}
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
			}

			// Set user info in request headers for downstream handlers
			c.Request().Header.Set(HeaderUserID, claims.UserID.String())
			c.Request().Header.Set(HeaderUsername, claims.Username)
			c.Request().Header.Set(HeaderUserType, claims.UserType)

			return next(c)
		}
	}
}
