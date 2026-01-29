package auth

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"

	"bereaucat/internal/store"
)

const (
	// HeaderProjectID is the header name for the project ID.
	HeaderProjectID = "X-Project-ID"
	// HeaderProjectKey is the header name for the project key.
	HeaderProjectKey = "X-Project-Key"
	// HeaderProjectRole is the header name for the user's role in the project.
	HeaderProjectRole = "X-Project-Role"
)

// ProjectMiddleware returns an Echo middleware that validates project access.
// It expects a :projectKey path parameter and the auth headers to be set.
// Sets X-Project-ID, X-Project-Key, and X-Project-Role headers for downstream handlers.
func ProjectMiddleware(queryer store.Querier) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			projectKey := c.Param("projectKey")
			if projectKey == "" {
				return echo.NewHTTPError(http.StatusBadRequest, "project key is required")
			}

			userIDStr := c.Request().Header.Get(HeaderUserID)
			userID, err := uuid.Parse(userIDStr)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid user ID")
			}

			ctx := c.Request().Context()

			// Get project by key
			project, err := queryer.GetProjectByKey(ctx, projectKey)
			if err != nil {
				return echo.NewHTTPError(http.StatusNotFound, "project not found")
			}

			// Check if user is a member of the project
			role, err := queryer.GetProjectMemberRole(ctx, store.GetProjectMemberRoleParams{
				ProjectID: project.ID,
				UserID:    userID,
			})
			if err != nil {
				return echo.NewHTTPError(http.StatusForbidden, "access denied: not a project member")
			}

			// Set project info in headers for downstream handlers
			c.Request().Header.Set(HeaderProjectID, project.ID.String())
			c.Request().Header.Set(HeaderProjectKey, project.ProjectKey)
			c.Request().Header.Set(HeaderProjectRole, role)

			return next(c)
		}
	}
}

// ProjectRoleMiddleware returns middleware that requires a minimum role level.
// This should be used AFTER ProjectMiddleware, as it expects the X-Project-Role header.
// Role hierarchy: guest < member < admin
func ProjectRoleMiddleware(minRole string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			role := c.Request().Header.Get(HeaderProjectRole)

			if !hasMinimumRole(role, minRole) {
				return echo.NewHTTPError(http.StatusForbidden, "insufficient project permissions")
			}

			return next(c)
		}
	}
}

// hasMinimumRole checks if the given role meets the minimum required role.
// Role hierarchy: guest < member < admin
func hasMinimumRole(role, minRole string) bool {
	roleLevel := map[string]int{
		"guest":  1,
		"member": 2,
		"admin":  3,
	}

	currentLevel, ok := roleLevel[role]
	if !ok {
		return false
	}

	requiredLevel, ok := roleLevel[minRole]
	if !ok {
		return false
	}

	return currentLevel >= requiredLevel
}
