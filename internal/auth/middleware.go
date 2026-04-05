package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"

	"bereaucat/internal/store"
)

const (
	// HeaderUserID is the header name for the user ID.
	HeaderUserID = "X-User-ID"
	// HeaderUsername is the header name for the username.
	HeaderUsername = "X-Username"
	// HeaderUserType is the header name for the user type.
	HeaderUserType = "X-User-Type"
	// HeaderAuthMethod is the header name for the authentication method.
	HeaderAuthMethod = "X-Auth-Method"

	// AuthMethodPAT is the value for PAT authentication.
	AuthMethodPAT = "pat"

	// patPrefix is the prefix for Personal Access Tokens.
	patPrefix = "bcat_"
)

// Middleware returns an Echo middleware that validates JWT tokens and Personal Access Tokens.
func Middleware(manager *Manager, queries store.Querier) echo.MiddlewareFunc {
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

			// Check if this is a Personal Access Token
			if strings.HasPrefix(tokenString, patPrefix) {
				return authenticateWithPAT(c, queries, tokenString, next)
			}

			// Otherwise, validate as JWT
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

// RejectPAT returns a middleware that rejects requests authenticated via Personal Access Tokens.
func RejectPAT() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if c.Request().Header.Get(HeaderAuthMethod) == AuthMethodPAT {
				return echo.NewHTTPError(http.StatusForbidden, "personal access tokens cannot access this endpoint")
			}
			return next(c)
		}
	}
}

// authenticateWithPAT validates a Personal Access Token and sets user headers.
func authenticateWithPAT(c *echo.Context, queries store.Querier, token string, next echo.HandlerFunc) error {
	tokenHash := HashToken(token)

	pat, err := queries.GetPersonalAccessTokenByHash(c.Request().Context(), tokenHash)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
	}

	// Set user info from the joined user data
	c.Request().Header.Set(HeaderUserID, pat.UserID.String())
	c.Request().Header.Set(HeaderUsername, pat.Username)
	c.Request().Header.Set(HeaderUserType, pat.UserType)
	c.Request().Header.Set(HeaderAuthMethod, AuthMethodPAT)

	// Update last_used_at in the background
	go func() {
		_ = queries.UpdatePersonalAccessTokenLastUsed(context.Background(), pat.ID)
	}()

	return next(c)
}
