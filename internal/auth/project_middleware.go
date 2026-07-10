package auth

import (
	"net/http"
	"strings"

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
	// HeaderProjectDisabled is "true" when the project is disabled (read-only).
	HeaderProjectDisabled = "X-Project-Disabled"
)

// isProjectWrite reports whether the request mutates project data. Read methods
// (GET/HEAD/OPTIONS) are always allowed on a disabled project; the enable/disable
// toggle (PATCH .../disabled) is also allowed so a project can be re-enabled, and
// deleting the project itself is allowed so a disabled project can still be removed.
func isProjectWrite(method, path, projectKey string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return false
	}
	if method == http.MethodPatch && strings.HasSuffix(path, "/disabled") {
		return false
	}
	// DELETE on the project root (.../projects/{key}) removes the project itself.
	if method == http.MethodDelete && strings.HasSuffix(path, "/projects/"+projectKey) {
		return false
	}
	return true
}

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

			// Admin users can access all projects with admin role
			userType := c.Request().Header.Get(HeaderUserType)
			var role string
			if userType == "admin" {
				role = "admin"
			} else {
				// Check if user is a member of the project
				role, err = queryer.GetProjectMemberRole(ctx, store.GetProjectMemberRoleParams{
					ProjectID: project.ID,
					UserID:    userID,
				})
				if err != nil {
					return echo.NewHTTPError(http.StatusForbidden, "access denied: not a project member")
				}
			}

			// A disabled project is read-only: reject any write except the
			// toggle that re-enables it.
			if project.Disabled && isProjectWrite(c.Request().Method, c.Request().URL.Path, projectKey) {
				return echo.NewHTTPError(http.StatusForbidden, "project is disabled (read-only)")
			}

			// Set project info in headers for downstream handlers
			c.Request().Header.Set(HeaderProjectID, project.ID.String())
			c.Request().Header.Set(HeaderProjectKey, project.ProjectKey)
			c.Request().Header.Set(HeaderProjectRole, role)
			if project.Disabled {
				c.Request().Header.Set(HeaderProjectDisabled, "true")
			}

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
