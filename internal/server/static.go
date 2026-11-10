package server

import (
	"bytes"
	"html"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v5"

	"bereaucat/internal/branding"
	"bereaucat/internal/store"
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
				return serveFile(c, f, filePath, s.store)
			}
		}

		// File doesn't exist or is a directory - serve index.html for SPA routing
		indexFile, err := distFS.Open("index.html")
		if err != nil {
			return echo.ErrNotFound
		}
		defer indexFile.Close()

		return serveFile(c, indexFile, "index.html", s.store)
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

		return serveFile(c, f, filePath, s.store)
	})
}

// serveFile serves a file with the appropriate content type
func serveFile(c *echo.Context, f fs.File, filePath string, q store.Querier) error {
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

	if filePath == "index.html" {
		content = rewriteOGURLs(content, c.Request())
		if q != nil {
			content = rewriteBranding(content, branding.Load(c.Request().Context(), q))
		}
	}

	return c.Blob(http.StatusOK, contentType, content)
}

// rewriteOGURLs replaces relative og:image / twitter:image URLs with absolute
// URLs derived from the incoming request, so social crawlers that require
// absolute URLs (Facebook, LinkedIn, Slack) can fetch the preview image.
func rewriteOGURLs(content []byte, r *http.Request) []byte {
	scheme := "https"
	if r.TLS == nil && r.Header.Get("X-Forwarded-Proto") != "https" {
		scheme = "http"
	}
	if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		scheme = proto
	}
	host := r.Host
	if fwd := r.Header.Get("X-Forwarded-Host"); fwd != "" {
		host = fwd
	}
	if host == "" {
		return content
	}
	base := scheme + "://" + host

	for _, rel := range []string{"/api/v1/og-image"} {
		abs := base + rel
		content = bytes.ReplaceAll(content, []byte(`content="`+rel+`"`), []byte(`content="`+abs+`"`))
	}
	return content
}

const (
	defaultIconSVG = `<link rel="icon" type="image/svg+xml" href="/logo.svg">`
	defaultIconICO = `<link rel="icon" type="image/x-icon" href="/favicon.ico">`
)

// rewriteBranding swaps the static name and favicon in index.html, since crawlers don't run JS.
func rewriteBranding(content []byte, b branding.Brand) []byte {
	if !b.Custom {
		return content
	}
	name := html.EscapeString(b.Name)
	for _, attr := range []string{`property="og:site_name"`, `property="og:title"`, `name="twitter:title"`} {
		content = bytes.ReplaceAll(content,
			[]byte(attr+` content="`+branding.DefaultName+`"`),
			[]byte(attr+` content="`+name+`"`))
	}
	content = bytes.ReplaceAll(content,
		[]byte(`property="og:image:alt" content="`+branding.DefaultName+` `),
		[]byte(`property="og:image:alt" content="`+name+` `))

	if href, ctype := b.FaviconURL(); href != "" {
		icon := `<link rel="icon" type="` + ctype + `" href="` + href + `">`
		content = bytes.Replace(content, []byte(defaultIconSVG), []byte(icon), 1)
		content = bytes.Replace(content, []byte(defaultIconICO), nil, 1)
	}
	return content
}
