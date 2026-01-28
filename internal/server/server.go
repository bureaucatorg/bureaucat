package server

import (
	"database/sql"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	_ "github.com/lib/pq"
)

// Server wraps the Echo server with application configuration
type Server struct {
	echo    *echo.Echo
	devMode bool
	db      *sql.DB
}

// New creates a new Server instance
func New(devMode bool, dbURL string) (*Server, error) {
	e := echo.New()

	// Middleware
	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	srv := &Server{
		echo:    e,
		devMode: devMode,
	}

	// Open database connection if URL provided
	if dbURL != "" {
		db, err := sql.Open("postgres", dbURL)
		if err != nil {
			return nil, err
		}
		srv.db = db
	}

	// Register routes
	srv.registerRoutes()

	// Set up reverse proxy in dev mode
	if devMode {
		srv.setupProxy()
	}

	return srv, nil
}

// Close closes any open resources
func (s *Server) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// Start starts the HTTP server
func (s *Server) Start(addr string) error {
	return s.echo.Start(addr)
}

// Echo returns the underlying Echo instance
func (s *Server) Echo() *echo.Echo {
	return s.echo
}
