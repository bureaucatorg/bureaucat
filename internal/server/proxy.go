package server

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/labstack/echo/v5"
)

const nuxtDevServer = "http://localhost:3041"

func (s *Server) setupProxy() {
	target, _ := url.Parse(nuxtDevServer)
	proxy := httputil.NewSingleHostReverseProxy(target)

	// Catch-all handler for non-API routes
	s.echo.Any("/*", func(c *echo.Context) error {
		path := c.Request().URL.Path

		// Skip API routes
		if strings.HasPrefix(path, "/api/") {
			return echo.ErrNotFound
		}

		// Proxy to Nuxt dev server
		proxy.ServeHTTP(c.Response(), c.Request())
		return nil
	})

	// Handle Nuxt HMR and dev assets
	s.echo.Any("/_nuxt/*", func(c *echo.Context) error {
		proxy.ServeHTTP(c.Response(), c.Request())
		return nil
	})

	// Handle Nuxt websocket for HMR
	s.echo.Any("/__nuxt_devtools__/*", func(c *echo.Context) error {
		proxy.ServeHTTP(c.Response(), c.Request())
		return nil
	})
}

// proxyHandler creates a handler that proxies requests to the target URL
func proxyHandler(target *url.URL) http.Handler {
	return httputil.NewSingleHostReverseProxy(target)
}
