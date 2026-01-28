package server

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

func (s *Server) registerRoutes() {
	// API routes under /api/v1
	api := s.echo.Group("/api/v1")

	api.GET("/health", healthCheck)
}

func healthCheck(c *echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status": "ok",
	})
}
