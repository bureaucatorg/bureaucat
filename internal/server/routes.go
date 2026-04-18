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

// feedbackCORS opens the feedback endpoints to any origin so users on
// self-hosted instances can POST to the main bureaucat.org from their browser.
// Written inline rather than pulling in echo's CORS middleware so this stays
// compatible with echo v5's moving API surface.
func feedbackCORS(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		h := c.Response().Header()
		h.Set("Access-Control-Allow-Origin", "*")
		h.Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		h.Set("Access-Control-Allow-Headers", "Content-Type")
		h.Set("Access-Control-Max-Age", "86400")
		if c.Request().Method == http.MethodOptions {
			return c.NoContent(http.StatusNoContent)
		}
		return next(c)
	}
}

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

	// Public feedback endpoint — accepts cross-origin submissions from other
	// self-hosted Bureaucat instances, so CORS is explicitly opened.
	if s.feedbackHandler != nil {
		api.POST("/feedback", s.feedbackHandler.SubmitFeedback, feedbackCORS)
		api.OPTIONS("/feedback", func(c *echo.Context) error {
			return c.NoContent(http.StatusNoContent)
		}, feedbackCORS)
		api.GET("/settings/feedback", s.feedbackHandler.GetPublicSettings)
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
		protected := api.Group("", auth.Middleware(s.authManager, s.store), auth.EnforcePATScope())
		protected.GET("/me", s.authHandler.Me)
		protected.GET("/me/tasks", s.authHandler.MyTasks)
		protected.GET("/me/notifications", s.authHandler.GetMyNotifications)

		// Local mirror of outbound feedback. Authenticated — the sidebar
		// dialog calls this alongside the cross-origin POST to bureaucat.org
		// so admins can see what their own users submitted.
		if s.feedbackHandler != nil {
			protected.POST("/me/feedback", s.feedbackHandler.SubmitLocalFeedback)
		}

		// Personal Access Token routes (not accessible via PAT)
		if s.patHandler != nil {
			patGroup := protected.Group("", auth.RejectPAT())
			patGroup.GET("/me/tokens", s.patHandler.ListTokens)
			patGroup.POST("/me/tokens", s.patHandler.CreateToken)
			patGroup.PATCH("/me/tokens/:tokenId", s.patHandler.UpdateTokenScope)
			patGroup.DELETE("/me/tokens/:tokenId", s.patHandler.DeleteToken)
		}
		// Workspace-level active cycles dashboard (across all projects)
		if s.cycleHandler != nil {
			protected.GET("/cycles/active", s.cycleHandler.ListActiveCycles)
		}
		// Workspace-level active modules dashboard (across all projects)
		if s.moduleHandler != nil {
			protected.GET("/modules/active", s.moduleHandler.ListActiveModules)
		}

		// Global search across tasks, cycles, projects the user can see.
		if s.searchHandler != nil {
			protected.GET("/search", s.searchHandler.Search)
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

			// Views (saved filter combinations)
			if s.viewHandler != nil {
				projectGroup.GET("/views", s.viewHandler.ListViews)
				projectGroup.POST("/views", s.viewHandler.CreateView, auth.ProjectRoleMiddleware("member"))
				projectGroup.PATCH("/views/reorder", s.viewHandler.ReorderViews, auth.ProjectRoleMiddleware("member"))
				projectGroup.GET("/views/:slug", s.viewHandler.GetView)
				projectGroup.PATCH("/views/:slug", s.viewHandler.UpdateView, auth.ProjectRoleMiddleware("member"))
				projectGroup.DELETE("/views/:slug", s.viewHandler.DeleteView, auth.ProjectRoleMiddleware("member"))
			}

			// Cycles
			if s.cycleHandler != nil {
				projectGroup.GET("/cycles", s.cycleHandler.ListCycles)
				projectGroup.GET("/cycles/all", s.cycleHandler.ListAllProjectCycles)
				projectGroup.POST("/cycles", s.cycleHandler.CreateCycle, auth.ProjectRoleMiddleware("admin"))
				projectGroup.GET("/cycles/unassigned-tasks", s.cycleHandler.ListUnassignedTasks)
				projectGroup.GET("/cycles/:cycleId", s.cycleHandler.GetCycle)
				projectGroup.PATCH("/cycles/:cycleId", s.cycleHandler.UpdateCycle, auth.ProjectRoleMiddleware("admin"))
				projectGroup.DELETE("/cycles/:cycleId", s.cycleHandler.DeleteCycle, auth.ProjectRoleMiddleware("admin"))
				projectGroup.GET("/cycles/:cycleId/tasks", s.cycleHandler.ListCycleTasks)
				projectGroup.POST("/cycles/:cycleId/tasks", s.cycleHandler.AddCycleTasks, auth.ProjectRoleMiddleware("admin"))
				projectGroup.DELETE("/cycles/:cycleId/tasks/:taskId", s.cycleHandler.RemoveCycleTask, auth.ProjectRoleMiddleware("admin"))
				projectGroup.GET("/cycles/:cycleId/metrics", s.cycleHandler.GetCycleMetrics)
			}

			// Modules
			if s.moduleHandler != nil {
				projectGroup.GET("/modules", s.moduleHandler.ListModules)
				projectGroup.POST("/modules", s.moduleHandler.CreateModule, auth.ProjectRoleMiddleware("admin"))
				projectGroup.GET("/modules/tasks-picker", s.moduleHandler.ListProjectTasksNotInModule)
				projectGroup.GET("/modules/:moduleId", s.moduleHandler.GetModule)
				projectGroup.PATCH("/modules/:moduleId", s.moduleHandler.UpdateModule, auth.ProjectRoleMiddleware("admin"))
				projectGroup.DELETE("/modules/:moduleId", s.moduleHandler.DeleteModule, auth.ProjectRoleMiddleware("admin"))
				projectGroup.POST("/modules/:moduleId/duplicate", s.moduleHandler.DuplicateModule, auth.ProjectRoleMiddleware("admin"))
				projectGroup.GET("/modules/:moduleId/tasks", s.moduleHandler.ListModuleTasks)
				projectGroup.POST("/modules/:moduleId/tasks", s.moduleHandler.AddModuleTasks, auth.ProjectRoleMiddleware("admin"))
				projectGroup.DELETE("/modules/:moduleId/tasks/:taskId", s.moduleHandler.RemoveModuleTask, auth.ProjectRoleMiddleware("admin"))
				projectGroup.GET("/modules/:moduleId/members", s.moduleHandler.ListModuleMembers)
				projectGroup.POST("/modules/:moduleId/members", s.moduleHandler.AddModuleMember, auth.ProjectRoleMiddleware("admin"))
				projectGroup.DELETE("/modules/:moduleId/members/:userId", s.moduleHandler.RemoveModuleMember, auth.ProjectRoleMiddleware("admin"))
				projectGroup.GET("/modules/:moduleId/metrics", s.moduleHandler.GetModuleMetrics)
			}

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

			// Attachments
			if s.attachmentHandler != nil {
				projectGroup.GET("/tasks/:taskNum/attachments", s.attachmentHandler.ListTaskAttachments)
				projectGroup.POST("/tasks/:taskNum/attachments", s.attachmentHandler.AttachToTask, auth.ProjectRoleMiddleware("member"))
				projectGroup.DELETE("/tasks/:taskNum/attachments/:attachmentId", s.attachmentHandler.DeleteTaskAttachment, auth.ProjectRoleMiddleware("member"))

				projectGroup.GET("/tasks/:taskNum/comments/:commentId/attachments", s.attachmentHandler.ListCommentAttachments)
				projectGroup.POST("/tasks/:taskNum/comments/:commentId/attachments", s.attachmentHandler.AttachToComment, auth.ProjectRoleMiddleware("member"))
				projectGroup.DELETE("/tasks/:taskNum/comments/:commentId/attachments/:attachmentId", s.attachmentHandler.DeleteCommentAttachment, auth.ProjectRoleMiddleware("member"))
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
		admin := api.Group("/admin", auth.Middleware(s.authManager, s.store), auth.RejectPAT(), auth.AdminMiddleware())
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

		// Admin feedback (list / delete / settings)
		if s.feedbackHandler != nil {
			admin.GET("/feedback", s.feedbackHandler.ListFeedback)
			admin.DELETE("/feedback/:id", s.feedbackHandler.DeleteFeedback)
			admin.GET("/settings/feedback", s.feedbackHandler.GetAdminSettings)
			admin.PUT("/settings/feedback", s.feedbackHandler.UpdateAdminSettings)
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
