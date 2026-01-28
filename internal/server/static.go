package server

import (
	"io"
	"io/fs"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v5"
)

// setupStatic configures static file serving from embedded filesystem
func (s *Server) setupStatic(distFS fs.FS) {
	// Handle all non-API routes
	s.echo.Any("/*", func(c *echo.Context) error {
		path := c.Request().URL.Path

		// Skip API routes
		if strings.HasPrefix(path, "/api/") {
			return echo.ErrNotFound
		}

		// Normalize path: remove leading slash and trailing slash
		filePath := strings.TrimPrefix(path, "/")
		filePath = strings.TrimSuffix(filePath, "/")
		if filePath == "" {
			filePath = "index.html"
		}

		// Try to open the file
		f, err := distFS.Open(filePath)
		if err == nil {
			defer f.Close()
			// Check if it's a file (not directory)
			stat, statErr := f.Stat()
			if statErr == nil && !stat.IsDir() {
				// Serve the file directly
				return serveFile(c, f, filePath)
			}
		}

		// File doesn't exist or is a directory - serve index.html for SPA routing
		indexFile, err := distFS.Open("index.html")
		if err != nil {
			return echo.ErrNotFound
		}
		defer indexFile.Close()

		return serveFile(c, indexFile, "index.html")
	})

	// Handle _nuxt assets explicitly
	s.echo.Any("/_nuxt/*", func(c *echo.Context) error {
		path := c.Request().URL.Path
		filePath := strings.TrimPrefix(path, "/")

		f, err := distFS.Open(filePath)
		if err != nil {
			return echo.ErrNotFound
		}
		defer f.Close()

		return serveFile(c, f, filePath)
	})
}

// serveFile serves a file with the appropriate content type
func serveFile(c *echo.Context, f fs.File, filePath string) error {
	// Determine content type from extension
	ext := filepath.Ext(filePath)
	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Read file content
	content, err := io.ReadAll(f)
	if err != nil {
		return echo.ErrInternalServerError
	}

	return c.Blob(http.StatusOK, contentType, content)
}
