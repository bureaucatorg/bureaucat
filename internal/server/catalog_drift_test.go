package server

import (
	"net/http"
	"regexp"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"

	"bereaucat/internal/assistant"
	"bereaucat/internal/handlers"
	"bereaucat/internal/store"
)

// nonCatalogRoutes are /api/v1 routes deliberately absent from the assistant catalog.
// Admin and assistant routes reject PATs and are skipped wholesale.
var nonCatalogRoutes = map[string]bool{
	// Auth, sessions, PAT management and public plumbing.
	"GET /api/v1/auth/sso/{}":          true,
	"GET /api/v1/auth/sso/{}/callback": true,
	"POST /api/v1/signin":              true,
	"POST /api/v1/signup":              true,
	"POST /api/v1/token_refresh":       true,
	"POST /api/v1/logout":              true,
	"GET /api/v1/me/tokens":            true,
	"POST /api/v1/me/tokens":           true,
	"PATCH /api/v1/me/tokens/{}":       true,
	"DELETE /api/v1/me/tokens/{}":      true,
	"GET /api/v1/health":               true,
	"GET /api/v1/ht/":                  true,
	"GET /api/v1/openapi.json":         true,
	"GET /api/v1/og-image":             true,
	"GET /api/v1/releases":             true,
	"GET /api/v1/settings/branding":    true,
	"GET /api/v1/settings/feedback":    true,
	"GET /api/v1/settings/signup":      true,
	"GET /api/v1/settings/sso":         true,
	"POST /api/v1/feedback":            true,
	"POST /api/v1/me/feedback":         true,
	"GET /api/v1/uploads/{}":           true,
	"POST /api/v1/uploads":             true,

	// Not catalogued yet; candidates for new tools.
	"GET /api/v1/graph":                                              true,
	"GET /api/v1/graph/filters":                                      true,
	"GET /api/v1/users/{}/activity/graph":                            true,
	"GET /api/v1/me/notification_settings":                           true,
	"PUT /api/v1/me/notification_settings":                           true,
	"GET /api/v1/me/notifications/unread_count":                      true,
	"POST /api/v1/workspaces":                                        true,
	"PATCH /api/v1/workspaces/{}":                                    true,
	"DELETE /api/v1/workspaces/{}":                                   true,
	"POST /api/v1/workspaces/{}/members":                             true,
	"DELETE /api/v1/workspaces/{}/members/{}":                        true,
	"GET /api/v1/workspaces/{}/members/search":                       true,
	"PATCH /api/v1/projects/{}/disabled":                             true,
	"PATCH /api/v1/projects/{}/workspace":                            true,
	"GET /api/v1/projects/{}/workspace/move-impact":                  true,
	"PATCH /api/v1/projects/{}/views/reorder":                        true,
	"GET /api/v1/projects/{}/tasks/{}/activity/verify":               true,
	"DELETE /api/v1/projects/{}/tasks/{}/attachments/{}":             true,
	"GET /api/v1/projects/{}/tasks/{}/comments/{}/attachments":       true,
	"DELETE /api/v1/projects/{}/tasks/{}/comments/{}/attachments/{}": true,
}

var routeParam = regexp.MustCompile(`:[A-Za-z_]+|\{[a-z_]+\}`)

func routeKey(method, path string) string {
	return method + " " + routeParam.ReplaceAllString(path, "{}")
}

func TestAPIRoutesMatchAssistantCatalog(t *testing.T) {
	s := &Server{
		echo:                 echo.New(),
		store:                store.New(nil),
		authHandler:          &handlers.AuthHandler{},
		adminHandler:         &handlers.AdminHandler{},
		uploadHandler:        &handlers.UploadHandler{},
		workspaceHandler:     &handlers.WorkspaceHandler{},
		projectHandler:       &handlers.ProjectHandler{},
		pageHandler:          &handlers.PageHandler{},
		taskHandler:          &handlers.TaskHandler{},
		viewHandler:          &handlers.ViewHandler{},
		commentHandler:       &handlers.CommentHandler{},
		attachmentHandler:    &handlers.AttachmentHandler{},
		cycleHandler:         &handlers.CycleHandler{},
		moduleHandler:        &handlers.ModuleHandler{},
		settingsHandler:      &handlers.SettingsHandler{},
		ogHandler:            &handlers.OGHandler{},
		importHandler:        &handlers.ImportHandler{},
		oauthHandler:         &handlers.OAuthHandler{},
		patHandler:           &handlers.PATHandler{},
		feedbackHandler:      &handlers.FeedbackHandler{},
		searchHandler:        &handlers.SearchHandler{},
		releasesHandler:      &handlers.ReleasesHandler{},
		assistantHandler:     &handlers.AssistantHandler{},
		notificationsHandler: &handlers.NotificationHandler{},
	}
	s.registerRoutes()

	catalog := map[string]bool{}
	for _, tool := range assistant.Catalog() {
		catalog[routeKey(tool.Method, "/api/v1"+tool.Path)] = true
	}

	methods := map[string]bool{http.MethodGet: true, http.MethodPost: true, http.MethodPut: true, http.MethodPatch: true, http.MethodDelete: true}
	routes := map[string]bool{}
	for _, r := range s.echo.Router().Routes() {
		if !methods[r.Method] || !strings.HasPrefix(r.Path, "/api/v1/") {
			continue
		}
		routes[routeKey(r.Method, r.Path)] = true
	}

	for key := range routes {
		path := strings.SplitN(key, " ", 2)[1]
		if strings.HasPrefix(path, "/api/v1/admin/") || strings.HasPrefix(path, "/api/v1/assistant/") {
			continue
		}
		if !catalog[key] && !nonCatalogRoutes[key] {
			t.Errorf("%s has no assistant catalog tool: add one in internal/assistant/tools.go or list it in nonCatalogRoutes", key)
		}
	}
	for key := range catalog {
		if !routes[key] {
			t.Errorf("catalog tool route %s is not registered", key)
		}
	}
	for key := range nonCatalogRoutes {
		if !routes[key] {
			t.Errorf("stale nonCatalogRoutes entry %s: route no longer exists", key)
		}
		if catalog[key] {
			t.Errorf("%s is now in the catalog: remove it from nonCatalogRoutes", key)
		}
	}
}
