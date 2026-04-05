package server

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
	echoSwagger "github.com/swaggo/echo-swagger"

	"bereaucat/internal/auth"
	"bereaucat/internal/buildinfo"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	All     bool   `json:"all"`
	DB      bool   `json:"db"`
	API     bool   `json:"api"`
	Version string `json:"version"`
}

func (s *Server) registerRoutes() {
	// Swagger documentation
	s.echo.GET("/docs", func(c *echo.Context) error {
		return c.Redirect(http.StatusMovedPermanently, "/docs/index.html")
	})
	s.echo.GET("/docs/*", echoSwagger.WrapHandler)

	// API routes under /api/v1
	api := s.echo.Group("/api/v1")

	api.GET("/health", healthCheck)
	api.GET("/ht/", s.healthCheckDetailed)

	// Public settings routes
	if s.settingsHandler != nil {
		api.GET("/settings/branding", s.settingsHandler.GetBranding)
		api.GET("/settings/sso", s.settingsHandler.GetSSOProviders)
		api.GET("/settings/signup", s.settingsHandler.GetSignupSettings)
	}

	// SSO auth routes (public)
	if s.oauthHandler != nil {
		api.GET("/auth/sso/:provider", s.oauthHandler.StartSSO)
		api.GET("/auth/sso/:provider/callback", s.oauthHandler.CallbackSSO)
	}

	// Dynamic OG image
	if s.ogHandler != nil {
		api.GET("/og-image", s.ogHandler.OGImage)
	}

	// Auth routes (public)
	if s.authHandler != nil {
		api.POST("/signup", s.authHandler.Signup)
		api.POST("/signin", s.authHandler.Signin)
		api.POST("/token_refresh", s.authHandler.TokenRefresh)
		api.POST("/logout", s.authHandler.Logout)

		// Public upload serving (no auth required)
		if s.uploadHandler != nil {
			api.GET("/uploads/:id", s.uploadHandler.Serve)
		}

		// Protected routes
		protected := api.Group("", auth.Middleware(s.authManager, s.store))
		protected.GET("/me", s.authHandler.Me)
		protected.GET("/me/tasks", s.authHandler.MyTasks)
		protected.GET("/me/notifications", s.authHandler.GetMyNotifications)

		// Personal Access Token routes
		if s.patHandler != nil {
			protected.GET("/me/tokens", s.patHandler.ListTokens)
			protected.POST("/me/tokens", s.patHandler.CreateToken)
			protected.DELETE("/me/tokens/:tokenId", s.patHandler.DeleteToken)
		}
		protected.GET("/users/:id", s.authHandler.GetUserProfile)
		protected.GET("/users/:id/activity", s.authHandler.GetUserActivity)
		protected.GET("/users/:id/activity/graph", s.authHandler.GetUserActivityGraph)

		// File uploads (authenticated)
		if s.uploadHandler != nil {
			protected.POST("/uploads", s.uploadHandler.Upload)
		}

		// Project routes (authenticated)
		if s.projectHandler != nil {
			protected.GET("/projects", s.projectHandler.ListProjects)
			protected.POST("/projects", s.projectHandler.CreateProject)

			// Project-specific routes (requires project membership)
			projectGroup := protected.Group("/projects/:projectKey", auth.ProjectMiddleware(s.store))

			// Project CRUD
			projectGroup.GET("", s.projectHandler.GetProject)
			projectGroup.PATCH("", s.projectHandler.UpdateProject, auth.ProjectRoleMiddleware("admin"))
			projectGroup.DELETE("", s.projectHandler.DeleteProject, auth.ProjectRoleMiddleware("admin"))

			// Project members
			projectGroup.GET("/members", s.projectHandler.ListMembers)
			projectGroup.POST("/members", s.projectHandler.AddMember, auth.ProjectRoleMiddleware("admin"))
			projectGroup.PATCH("/members/:userId", s.projectHandler.UpdateMemberRole, auth.ProjectRoleMiddleware("admin"))
			projectGroup.DELETE("/members/:userId", s.projectHandler.RemoveMember, auth.ProjectRoleMiddleware("admin"))

			// Project states
			projectGroup.GET("/states", s.projectHandler.ListStates)
			projectGroup.POST("/states", s.projectHandler.CreateState, auth.ProjectRoleMiddleware("admin"))
			projectGroup.PATCH("/states/:stateId", s.projectHandler.UpdateState, auth.ProjectRoleMiddleware("admin"))
			projectGroup.DELETE("/states/:stateId", s.projectHandler.DeleteState, auth.ProjectRoleMiddleware("admin"))

			// Project labels
			projectGroup.GET("/labels", s.projectHandler.ListLabels)
			projectGroup.POST("/labels", s.projectHandler.CreateLabel, auth.ProjectRoleMiddleware("member"))
			projectGroup.PATCH("/labels/:labelId", s.projectHandler.UpdateLabel, auth.ProjectRoleMiddleware("admin"))
			projectGroup.DELETE("/labels/:labelId", s.projectHandler.DeleteLabel, auth.ProjectRoleMiddleware("admin"))

			// Task templates
			projectGroup.GET("/templates", s.projectHandler.ListTemplates)
			projectGroup.POST("/templates", s.projectHandler.CreateTemplate, auth.ProjectRoleMiddleware("admin"))
			projectGroup.PATCH("/templates/:templateId", s.projectHandler.UpdateTemplate, auth.ProjectRoleMiddleware("admin"))
			projectGroup.DELETE("/templates/:templateId", s.projectHandler.DeleteTemplate, auth.ProjectRoleMiddleware("admin"))

			// Tasks
			if s.taskHandler != nil {
				projectGroup.GET("/tasks", s.taskHandler.ListTasks)
				projectGroup.POST("/tasks", s.taskHandler.CreateTask, auth.ProjectRoleMiddleware("member"))
				projectGroup.GET("/tasks/:taskNum", s.taskHandler.GetTask)
				projectGroup.PATCH("/tasks/:taskNum", s.taskHandler.UpdateTask, auth.ProjectRoleMiddleware("member"))
				projectGroup.DELETE("/tasks/:taskNum", s.taskHandler.DeleteTask, auth.ProjectRoleMiddleware("member"))

				// Task assignees
				projectGroup.POST("/tasks/:taskNum/assignees", s.taskHandler.AddAssignee, auth.ProjectRoleMiddleware("member"))
				projectGroup.DELETE("/tasks/:taskNum/assignees/:userId", s.taskHandler.RemoveAssignee, auth.ProjectRoleMiddleware("member"))

				// Task labels
				projectGroup.POST("/tasks/:taskNum/labels", s.taskHandler.AddLabel, auth.ProjectRoleMiddleware("member"))
				projectGroup.DELETE("/tasks/:taskNum/labels/:labelId", s.taskHandler.RemoveLabel, auth.ProjectRoleMiddleware("member"))
			}

			// Comments and Activity
			if s.commentHandler != nil {
				projectGroup.GET("/tasks/:taskNum/comments", s.commentHandler.ListComments)
				projectGroup.POST("/tasks/:taskNum/comments", s.commentHandler.CreateComment, auth.ProjectRoleMiddleware("member"))
				projectGroup.PATCH("/tasks/:taskNum/comments/:commentId", s.commentHandler.UpdateComment, auth.ProjectRoleMiddleware("member"))
				projectGroup.DELETE("/tasks/:taskNum/comments/:commentId", s.commentHandler.DeleteComment, auth.ProjectRoleMiddleware("member"))

				// Activity log
				projectGroup.GET("/tasks/:taskNum/activity", s.commentHandler.GetActivity)
				projectGroup.GET("/tasks/:taskNum/activity/verify", s.commentHandler.VerifyActivity)
			}
		}

		// Admin routes (requires auth + admin)
		admin := api.Group("/admin", auth.Middleware(s.authManager, s.store), auth.AdminMiddleware())
		admin.GET("/users", s.adminHandler.ListUsers)
		admin.POST("/users", s.adminHandler.CreateUser)
		admin.DELETE("/users/:id", s.adminHandler.DeleteUser)
		admin.PUT("/users/:id/role", s.adminHandler.UpdateUserRole)
		admin.PUT("/users/:id/password", s.adminHandler.ResetUserPassword)
		admin.GET("/tokens", s.adminHandler.ListTokens)
		admin.DELETE("/tokens/:id", s.adminHandler.RevokeToken)
		admin.DELETE("/tokens/expired", s.adminHandler.CleanupExpiredTokens)

		// Admin settings
		if s.settingsHandler != nil {
			admin.PUT("/settings/branding", s.settingsHandler.UpdateBranding)
			admin.PUT("/settings/signup", s.settingsHandler.UpdateSignupSettings)
			admin.GET("/settings/sso", s.settingsHandler.GetSSOSettings)
			admin.PUT("/settings/sso", s.settingsHandler.UpdateSSOSettings)
			admin.GET("/settings/mattermost", s.settingsHandler.GetMattermostSettings)
			admin.PUT("/settings/mattermost", s.settingsHandler.UpdateMattermostSettings)
			admin.POST("/settings/mattermost/test", s.settingsHandler.TestMattermostConnection)
		}

		// Admin data import
		if s.importHandler != nil {
			admin.POST("/import/plane", s.importHandler.ImportPlane)
		}
	}
}

func healthCheck(c *echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status":  "ok",
		"version": buildinfo.Version,
	})
}

func (s *Server) healthCheckDetailed(c *echo.Context) error {
	resp := HealthResponse{
		API:     true,
		DB:      false,
		Version: buildinfo.Version,
	}

	// Check database connection with timeout
	if s.db != nil {
		ctx, cancel := context.WithTimeout(c.Request().Context(), 3*time.Second)
		defer cancel()

		if err := s.db.PingContext(ctx); err == nil {
			resp.DB = true
		}
	}

	resp.All = resp.API && resp.DB

	return c.JSON(http.StatusOK, resp)
}
