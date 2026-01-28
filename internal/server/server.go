package server

import (
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

// Server wraps the Echo server with application configuration
type Server struct {
	echo    *echo.Echo
	devMode bool
}

// New creates a new Server instance
func New(devMode bool) *Server {
	e := echo.New()

	// Middleware
	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	srv := &Server{
		echo:    e,
		devMode: devMode,
	}

	// Register routes
	srv.registerRoutes()

	// Set up reverse proxy in dev mode
	if devMode {
		srv.setupProxy()
	}

	return srv
}

// Start starts the HTTP server
func (s *Server) Start(addr string) error {
	return s.echo.Start(addr)
}

// Echo returns the underlying Echo instance
func (s *Server) Echo() *echo.Echo {
	return s.echo
}
