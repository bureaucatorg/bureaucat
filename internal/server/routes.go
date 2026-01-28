package server

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"

	"bereaucat/internal/auth"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	All bool `json:"all"`
	DB  bool `json:"db"`
	API bool `json:"api"`
}

func (s *Server) registerRoutes() {
	// API routes under /api/v1
	api := s.echo.Group("/api/v1")

	api.GET("/health", healthCheck)
	api.GET("/ht/", s.healthCheckDetailed)

	// Auth routes (public)
	if s.authHandler != nil {
		api.POST("/signup", s.authHandler.Signup)
		api.POST("/signin", s.authHandler.Signin)
		api.POST("/token_refresh", s.authHandler.TokenRefresh)
		api.POST("/logout", s.authHandler.Logout)

		// Protected routes
		protected := api.Group("", auth.Middleware(s.authManager))
		protected.GET("/me", s.authHandler.Me)
	}
}

func healthCheck(c *echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status": "ok",
	})
}

func (s *Server) healthCheckDetailed(c *echo.Context) error {
	resp := HealthResponse{
		API: true,
		DB:  false,
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
