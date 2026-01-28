package server

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	_ "github.com/lib/pq"

	"bereaucat/internal/auth"
	"bereaucat/internal/handlers"
	"bereaucat/internal/store"
)

// AuthConfig holds authentication configuration
type AuthConfig struct {
	JWTSecret              string
	AccessTokenExpiryMins  int
	RefreshTokenExpiryDays int
}

// Server wraps the Echo server with application configuration
type Server struct {
	echo         *echo.Echo
	devMode      bool
	db           *sql.DB
	pool         *pgxpool.Pool
	store        store.Querier
	authManager  *auth.Manager
	authHandler  *handlers.AuthHandler
	adminHandler *handlers.AdminHandler
	distFS       fs.FS
}

// New creates a new Server instance
// distFS should be provided in production mode (non-dev) for serving embedded static files
func New(devMode bool, dbURL string, authConfig AuthConfig, distFS fs.FS) (*Server, error) {
	e := echo.New()

	// Middleware
	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	srv := &Server{
		echo:    e,
		devMode: devMode,
		distFS:  distFS,
	}

	// Open database connection if URL provided
	if dbURL != "" {
		// sql.DB for health checks (existing)
		db, err := sql.Open("postgres", dbURL)
		if err != nil {
			return nil, fmt.Errorf("failed to open sql.DB: %w", err)
		}
		srv.db = db

		// pgxpool for sqlc queries
		pool, err := pgxpool.New(context.Background(), dbURL)
		if err != nil {
			return nil, fmt.Errorf("failed to create pgx pool: %w", err)
		}
		srv.pool = pool
		srv.store = store.New(pool)
	}

	// Initialize auth manager
	srv.authManager = auth.NewManager(auth.Config{
		JWTSecret:              authConfig.JWTSecret,
		AccessTokenExpiryMins:  authConfig.AccessTokenExpiryMins,
		RefreshTokenExpiryDays: authConfig.RefreshTokenExpiryDays,
	})

	// Initialize auth handler
	if srv.store != nil {
		srv.authHandler = handlers.NewAuthHandler(srv.store, srv.authManager, devMode)
		srv.adminHandler = handlers.NewAdminHandler(srv.store, srv.authManager, devMode)
	}

	// Register routes
	srv.registerRoutes()

	// Set up static file serving
	if devMode {
		// Dev mode: proxy to Nuxt dev server
		srv.setupProxy()
	} else if distFS != nil {
		// Production mode: serve embedded static files
		srv.setupStatic(distFS)
	}

	return srv, nil
}

// Close closes any open resources
func (s *Server) Close() error {
	if s.pool != nil {
		s.pool.Close()
	}
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
