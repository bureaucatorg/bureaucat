package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v5"

	"bereaucat/internal/buildinfo"
)

const (
	releasesURL      = "https://api.github.com/repos/bureaucatorg/bureaucat/releases?per_page=15"
	releasesCacheTTL = time.Hour
)

// Only releases at or above this version are shown.
var minReleaseVersion = [3]int{0, 16, 20}

// Release is a GitHub release. BodyHTML is rendered and sanitized by GitHub.
type Release struct {
	Name        string    `json:"name"`
	TagName     string    `json:"tag_name"`
	HTMLURL     string    `json:"html_url"`
	BodyHTML    string    `json:"body_html"`
	PublishedAt time.Time `json:"published_at"`
	Draft       bool      `json:"draft"`
	Prerelease  bool      `json:"prerelease"`
}

type ReleasesResponse struct {
	CurrentVersion string    `json:"current_version"`
	Releases       []Release `json:"releases"`
}

type ReleasesHandler struct {
	client    *http.Client
	mu        sync.Mutex
	cached    []Release
	fetchedAt time.Time
}

func NewReleasesHandler() *ReleasesHandler {
	return &ReleasesHandler{client: &http.Client{Timeout: 10 * time.Second}}
}

// ListReleases returns recent bureaucat releases from GitHub, cached in memory.
func (h *ReleasesHandler) ListReleases(c *echo.Context) error {
	releases, err := h.getReleases(c)
	if err != nil {
		return c.JSON(http.StatusBadGateway, map[string]string{"error": "Failed to fetch releases"})
	}
	return c.JSON(http.StatusOK, ReleasesResponse{
		CurrentVersion: buildinfo.Version,
		Releases:       releases,
	})
}

func (h *ReleasesHandler) getReleases(c *echo.Context) ([]Release, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.cached != nil && time.Since(h.fetchedAt) < releasesCacheTTL {
		return h.cached, nil
	}

	req, err := http.NewRequestWithContext(c.Request().Context(), http.MethodGet, releasesURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.html+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("User-Agent", "bureaucat/"+buildinfo.Version)

	resp, err := h.client.Do(req)
	if err != nil {
		return h.staleOr(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return h.staleOr(fmt.Errorf("github returned %d", resp.StatusCode))
	}

	var all []Release
	if err := json.NewDecoder(resp.Body).Decode(&all); err != nil {
		return h.staleOr(err)
	}

	releases := make([]Release, 0, len(all))
	for _, r := range all {
		if !r.Draft && isAtLeastMin(r.TagName) {
			releases = append(releases, r)
		}
	}

	h.cached = releases
	h.fetchedAt = time.Now()
	return releases, nil
}

func (h *ReleasesHandler) staleOr(err error) ([]Release, error) {
	if h.cached != nil {
		return h.cached, nil
	}
	return nil, err
}

func isAtLeastMin(tag string) bool {
	core, _, _ := strings.Cut(strings.TrimPrefix(tag, "v"), "-")
	parts := strings.Split(core, ".")
	if len(parts) != 3 {
		return false
	}
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			return false
		}
		if n != minReleaseVersion[i] {
			return n > minReleaseVersion[i]
		}
	}
	return true
}
