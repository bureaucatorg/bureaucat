package server

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	_ "github.com/lib/pq"

	_ "bereaucat/docs"
	"bereaucat/internal/activity"
	"bereaucat/internal/auth"
	"bereaucat/internal/handlers"
	"bereaucat/internal/notifier"
	"bereaucat/internal/store"
	"bereaucat/internal/uploads"
)

// AuthConfig holds authentication configuration
type AuthConfig struct {
	JWTSecret              string
	AccessTokenExpiryMins  int
	RefreshTokenExpiryDays int
}

// Server wraps the Echo server with application configuration
type Server struct {
	echo            *echo.Echo
	devMode         bool
	db              *sql.DB
	pool            *pgxpool.Pool
	store           store.Querier
	authManager     *auth.Manager
	authHandler     *handlers.AuthHandler
	adminHandler    *handlers.AdminHandler
	uploadHandler   *handlers.UploadHandler
	projectHandler  *handlers.ProjectHandler
	taskHandler     *handlers.TaskHandler
	commentHandler    *handlers.CommentHandler
	attachmentHandler *handlers.AttachmentHandler
	settingsHandler *handlers.SettingsHandler
	ogHandler       *handlers.OGHandler
	importHandler   *handlers.ImportHandler
	oauthHandler    *handlers.OAuthHandler
	patHandler      *handlers.PATHandler
	activityService     *activity.Service
	notificationService *notifier.Service
	uploadService       *uploads.Service
	distFS              fs.FS
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

	// Initialize handlers
	if srv.store != nil {
		srv.authHandler = handlers.NewAuthHandler(srv.store, srv.authManager, devMode)
		srv.adminHandler = handlers.NewAdminHandler(srv.store, srv.authManager, devMode)

		// Initialize upload service (S3-backed)
		maxUploadSize := int64(10 * 1024 * 1024) // 10MB default
		if sizeStr := os.Getenv("MAX_UPLOAD_SIZE"); sizeStr != "" {
			if size, err := strconv.ParseInt(sizeStr, 10, 64); err == nil {
				maxUploadSize = size
			}
		}
		uploadService, err := uploads.NewService(uploads.Config{
			S3Endpoint:  os.Getenv("S3_ENDPOINT"),
			BucketName:  os.Getenv("FILES_BUCKET_NAME"),
			AccessKeyID: os.Getenv("FILES_BUCKET_ACCESS_KEY_ID"),
			SecretKey:   os.Getenv("FILES_BUCKET_SECRET_ACCESS_KEY"),
			MaxFileSize: maxUploadSize,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create upload service: %w", err)
		}
		srv.uploadService = uploadService
		srv.uploadHandler = handlers.NewUploadHandler(srv.store, uploadService)

		// Initialize activity service
		srv.activityService = activity.NewService(srv.store)

		// Initialize notification service (loads providers dynamically from settings)
		srv.notificationService = notifier.NewService(srv.store)

		// Initialize project and task handlers
		srv.projectHandler = handlers.NewProjectHandler(srv.store)
		srv.taskHandler = handlers.NewTaskHandler(srv.store, srv.activityService, srv.notificationService)
		srv.commentHandler = handlers.NewCommentHandler(srv.store, srv.activityService, srv.notificationService)
		srv.attachmentHandler = handlers.NewAttachmentHandler(srv.store, uploadService)
		srv.settingsHandler = handlers.NewSettingsHandler(srv.store)
		srv.oauthHandler = handlers.NewOAuthHandler(srv.store, srv.authManager, srv.authHandler, devMode)
		srv.ogHandler = handlers.NewOGHandler(srv.store)
		srv.importHandler = handlers.NewImportHandler(srv.pool)
		srv.patHandler = handlers.NewPATHandler(srv.store)
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
