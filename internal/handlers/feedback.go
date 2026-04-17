package handlers

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"

	"bereaucat/internal/store"
)

// Anonymous POSTs are abuse-prone, so the public endpoint enforces a simple
// per-IP sliding-window limit. Numbers are deliberately generous: a human
// submitting feedback will never hit them, but a script hammering the endpoint
// will trip the cap quickly.
const (
	feedbackRateLimit  = 5
	feedbackRateWindow = time.Hour
	// ipHashMaxSize caps the rate-limit map so an attacker rotating source IPs
	// can't use this data structure to grow memory unbounded. Once we cross
	// the cap, the oldest entries are evicted on the next successful Allow.
	ipHashMaxSize = 50_000
)

// ipRateLimiter is a pure in-memory sliding-window limiter keyed by IP. For a
// self-hosted app with a single process, this is adequate; a multi-replica
// deployment would swap in Redis or similar.
type ipRateLimiter struct {
	mu       sync.Mutex
	hits     map[string][]time.Time
	limit    int
	window   time.Duration
	maxSize  int
}

func newIPRateLimiter(limit int, window time.Duration, maxSize int) *ipRateLimiter {
	return &ipRateLimiter{
		hits:    make(map[string][]time.Time),
		limit:   limit,
		window:  window,
		maxSize: maxSize,
	}
}

// Allow reports whether an IP may submit another feedback right now. Also
// opportunistically prunes stale entries to keep memory bounded.
func (l *ipRateLimiter) Allow(ip string) bool {
	if ip == "" {
		// Couldn't identify a client — be conservative and still allow, since
		// stdlib misconfig shouldn't break the feature. Abuse on unknown IPs
		// is still bounded by the total map size.
		return true
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-l.window)

	timestamps := l.hits[ip]
	// Drop anything older than the window so the slice doesn't grow forever.
	pruned := timestamps[:0]
	for _, t := range timestamps {
		if t.After(cutoff) {
			pruned = append(pruned, t)
		}
	}
	if len(pruned) >= l.limit {
		l.hits[ip] = pruned
		return false
	}
	l.hits[ip] = append(pruned, now)

	// Global map-size guard. Cheap check on every admitted request; when we
	// exceed the cap, do a one-shot sweep dropping any IP whose window is
	// fully in the past. This keeps the per-request cost O(1) amortised.
	if len(l.hits) > l.maxSize {
		for k, ts := range l.hits {
			fresh := false
			for _, t := range ts {
				if t.After(cutoff) {
					fresh = true
					break
				}
			}
			if !fresh {
				delete(l.hits, k)
			}
		}
	}

	return true
}

// FeedbackHandler handles anonymous feedback submissions and admin moderation.
// Data lives in the `feedback` table; raw pgx is used so we don't need a
// separate sqlc round-trip for this tiny, standalone domain.
type FeedbackHandler struct {
	pool    *pgxpool.Pool
	store   store.Querier
	limiter *ipRateLimiter
}

// NewFeedbackHandler wires up the handler with a pgx pool (for data access) and
// the sqlc Querier (reused for the existing settings table).
func NewFeedbackHandler(pool *pgxpool.Pool, s store.Querier) *FeedbackHandler {
	return &FeedbackHandler{
		pool:    pool,
		store:   s,
		limiter: newIPRateLimiter(feedbackRateLimit, feedbackRateWindow, ipHashMaxSize),
	}
}

// clientIP extracts the best-guess client IP from the request. Prefers the
// first entry of X-Forwarded-For (when present and this process trusts its
// reverse proxy), then X-Real-IP, finally RemoteAddr.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// XFF is a comma-separated list; the left-most entry is the original
		// client. We don't drop spoofable entries here because we also don't
		// treat this as identity — it's strictly a rate-limit bucket.
		if comma := strings.IndexByte(xff, ','); comma > 0 {
			xff = xff[:comma]
		}
		return strings.TrimSpace(xff)
	}
	if xr := r.Header.Get("X-Real-IP"); xr != "" {
		return strings.TrimSpace(xr)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// FeedbackSettings is stored under the `feedback` settings key. It controls
// whether this instance accepts incoming feedback and whether its users can
// submit outgoing feedback to the main bureaucat.org instance.
type FeedbackSettings struct {
	// ReceiveEnabled gates POST /api/v1/feedback on this instance.
	// Defaults to false — admins (including on the main instance) must opt in
	// before any feedback is accepted.
	ReceiveEnabled bool `json:"receive_enabled"`
	// SendToMainEnabled toggles the feedback button in the sidebar for users
	// on *this* instance. Defaults to false so self-hosted admins must opt in.
	SendToMainEnabled bool `json:"send_to_main_enabled"`
	// StoreSentLocally mirrors outbound submissions into this instance's own
	// feedback table. On by default so admins can see what their users are
	// saying even when the main instance is the ultimate recipient.
	StoreSentLocally bool `json:"store_sent_locally"`
}

// defaultFeedbackSettings mirrors the shipping defaults when no row exists.
func defaultFeedbackSettings() FeedbackSettings {
	return FeedbackSettings{
		ReceiveEnabled:    false,
		SendToMainEnabled: false,
		StoreSentLocally:  true,
	}
}

// loadFeedbackSettings reads and parses the settings row, falling back to
// defaults on any error so the endpoints never 500 due to missing config.
func (h *FeedbackHandler) loadFeedbackSettings(ctx context.Context) FeedbackSettings {
	setting, err := h.store.GetSetting(ctx, "feedback")
	if err != nil {
		return defaultFeedbackSettings()
	}
	var s FeedbackSettings
	if err := json.Unmarshal(setting.Value, &s); err != nil {
		return defaultFeedbackSettings()
	}
	return s
}

// FeedbackPublicSettings is returned on the unauthenticated endpoint. Exposes
// just enough for the sidebar / dialog to make display decisions — the
// receive-side toggle stays internal.
type FeedbackPublicSettings struct {
	SendToMainEnabled bool `json:"send_to_main_enabled"`
	StoreSentLocally  bool `json:"store_sent_locally"`
}

// GetPublicSettings lets the sidebar decide whether to render the feedback
// button and whether to dual-post a local copy, without leaking receive-side
// configuration.
//
//	@Summary	Get public feedback settings
//	@Tags		Settings
//	@Produce	json
//	@Success	200	{object}	FeedbackPublicSettings
//	@Router		/settings/feedback [get]
func (h *FeedbackHandler) GetPublicSettings(c *echo.Context) error {
	s := h.loadFeedbackSettings(c.Request().Context())
	return c.JSON(http.StatusOK, FeedbackPublicSettings{
		SendToMainEnabled: s.SendToMainEnabled,
		StoreSentLocally:  s.StoreSentLocally,
	})
}

// SubmitFeedbackRequest is the public submission body. Kept deliberately tiny:
// no user identity, no project linkage — this is anonymous feedback.
type SubmitFeedbackRequest struct {
	Message string `json:"message"`
}

// SubmitFeedback receives anonymous feedback. Public, CORS-enabled, and gated
// by the receive_enabled toggle so admins can stop accepting submissions.
//
//	@Summary	Submit anonymous feedback
//	@Tags		Feedback
//	@Accept		json
//	@Produce	json
//	@Param		body	body	SubmitFeedbackRequest	true	"Feedback"
//	@Success	202
//	@Router		/feedback [post]
func (h *FeedbackHandler) SubmitFeedback(c *echo.Context) error {
	ctx := c.Request().Context()
	settings := h.loadFeedbackSettings(ctx)
	if !settings.ReceiveEnabled {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "feedback is not being accepted")
	}

	// Rate-limit BEFORE binding so an attacker can't keep this handler busy
	// parsing bodies once they've tripped the cap.
	ip := clientIP(c.Request())
	if !h.limiter.Allow(ip) {
		// Give browsers a rough hint for when to retry.
		c.Response().Header().Set("Retry-After", strconv.Itoa(int(feedbackRateWindow.Seconds())))
		return echo.NewHTTPError(http.StatusTooManyRequests, "too many submissions — try again later")
	}

	msg, err := parseFeedbackBody(c)
	if err != nil {
		return err
	}

	origin := c.Request().Header.Get("Origin")
	if origin == "" {
		origin = c.Request().Header.Get("Referer")
	}
	if len(origin) > 255 {
		origin = origin[:255]
	}
	ua := c.Request().UserAgent()

	if _, err := h.pool.Exec(ctx,
		`INSERT INTO feedback (message, source_origin, user_agent) VALUES ($1, $2, $3)`,
		msg, origin, ua,
	); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to store feedback")
	}
	return c.NoContent(http.StatusAccepted)
}

// SubmitLocalFeedback mirrors a sidebar submission into *this* instance's
// feedback table, so admins can see what their own users sent to bureaucat.org
// without scraping the main instance. Authenticated (the caller is one of our
// users), and therefore exempt from the anonymous rate limiter — auth itself
// is the abuse brake here. Gated by the store_sent_locally setting.
//
//	@Summary	Store a local copy of outbound feedback
//	@Tags		Feedback
//	@Accept		json
//	@Produce	json
//	@Param		body	body	SubmitFeedbackRequest	true	"Feedback"
//	@Success	202
//	@Security	BearerAuth
//	@Router		/me/feedback [post]
func (h *FeedbackHandler) SubmitLocalFeedback(c *echo.Context) error {
	ctx := c.Request().Context()
	settings := h.loadFeedbackSettings(ctx)
	if !settings.StoreSentLocally {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "local storage of feedback is disabled")
	}

	msg, err := parseFeedbackBody(c)
	if err != nil {
		return err
	}

	// Tag the origin so the admin view can tell local copies apart from
	// external submissions at a glance. We intentionally do *not* record the
	// user ID — the feedback remains anonymous in the table.
	origin := "local:" + c.Request().Host
	if len(origin) > 255 {
		origin = origin[:255]
	}
	ua := c.Request().UserAgent()

	if _, err := h.pool.Exec(ctx,
		`INSERT INTO feedback (message, source_origin, user_agent) VALUES ($1, $2, $3)`,
		msg, origin, ua,
	); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to store feedback")
	}
	return c.NoContent(http.StatusAccepted)
}

// parseFeedbackBody extracts and normalises the message field, centralising
// the same validation the public and local endpoints need.
func parseFeedbackBody(c *echo.Context) (string, error) {
	var req SubmitFeedbackRequest
	if err := c.Bind(&req); err != nil {
		return "", echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	msg := strings.TrimSpace(req.Message)
	if msg == "" {
		return "", echo.NewHTTPError(http.StatusBadRequest, "message is required")
	}
	if len(msg) > 5000 {
		msg = msg[:5000]
	}
	return msg, nil
}


// FeedbackItem is the admin listing shape.
type FeedbackItem struct {
	ID           uuid.UUID `json:"id"`
	Message      string    `json:"message"`
	SourceOrigin string    `json:"source_origin,omitempty"`
	UserAgent    string    `json:"user_agent,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// PaginatedFeedback is returned from the admin list endpoint.
type PaginatedFeedback struct {
	Items      []FeedbackItem `json:"items"`
	Total      int            `json:"total"`
	Page       int            `json:"page"`
	PerPage    int            `json:"per_page"`
	TotalPages int            `json:"total_pages"`
}

// ListFeedback paginates all submitted feedback, newest first.
//
//	@Summary	List received feedback
//	@Tags		Admin - Feedback
//	@Produce	json
//	@Param		page		query	int	false	"Page (default 1)"
//	@Param		per_page	query	int	false	"Page size (default 50, max 200)"
//	@Success	200	{object}	PaginatedFeedback
//	@Security	BearerAuth
//	@Router		/admin/feedback [get]
func (h *FeedbackHandler) ListFeedback(c *echo.Context) error {
	ctx := c.Request().Context()

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))
	if perPage < 1 {
		perPage = 50
	}
	if perPage > 200 {
		perPage = 200
	}
	offset := (page - 1) * perPage

	var total int
	if err := h.pool.QueryRow(ctx, `SELECT COUNT(*) FROM feedback`).Scan(&total); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to count feedback")
	}

	rows, err := h.pool.Query(ctx,
		`SELECT id, message, COALESCE(source_origin, ''), COALESCE(user_agent, ''), created_at
		 FROM feedback
		 ORDER BY created_at DESC
		 LIMIT $1 OFFSET $2`, perPage, offset,
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list feedback")
	}
	defer rows.Close()

	items := make([]FeedbackItem, 0, perPage)
	for rows.Next() {
		var it FeedbackItem
		if err := rows.Scan(&it.ID, &it.Message, &it.SourceOrigin, &it.UserAgent, &it.CreatedAt); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to scan feedback")
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to read feedback")
	}

	totalPages := (total + perPage - 1) / perPage
	if totalPages < 1 {
		totalPages = 1
	}

	return c.JSON(http.StatusOK, PaginatedFeedback{
		Items:      items,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	})
}

// DeleteFeedback removes a single feedback row.
//
//	@Summary	Delete a feedback entry
//	@Tags		Admin - Feedback
//	@Produce	json
//	@Param		id	path	string	true	"Feedback ID"
//	@Success	200	{object}	MessageResponse
//	@Security	BearerAuth
//	@Router		/admin/feedback/{id} [delete]
func (h *FeedbackHandler) DeleteFeedback(c *echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	tag, err := h.pool.Exec(c.Request().Context(), `DELETE FROM feedback WHERE id = $1`, id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete feedback")
	}
	if tag.RowsAffected() == 0 {
		return echo.NewHTTPError(http.StatusNotFound, "feedback not found")
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "deleted"})
}

// GetAdminSettings returns the full feedback settings for admins.
//
//	@Summary	Get feedback settings
//	@Tags		Admin - Feedback
//	@Produce	json
//	@Success	200	{object}	FeedbackSettings
//	@Security	BearerAuth
//	@Router		/admin/settings/feedback [get]
func (h *FeedbackHandler) GetAdminSettings(c *echo.Context) error {
	return c.JSON(http.StatusOK, h.loadFeedbackSettings(c.Request().Context()))
}

// UpdateAdminSettings persists the feedback settings.
//
//	@Summary	Update feedback settings
//	@Tags		Admin - Feedback
//	@Accept		json
//	@Produce	json
//	@Param		body	body		FeedbackSettings	true	"Settings"
//	@Success	200		{object}	FeedbackSettings
//	@Security	BearerAuth
//	@Router		/admin/settings/feedback [put]
func (h *FeedbackHandler) UpdateAdminSettings(c *echo.Context) error {
	var req FeedbackSettings
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	ctx := c.Request().Context()

	value, err := json.Marshal(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to marshal settings")
	}
	if _, err := h.store.UpsertSetting(ctx, store.UpsertSettingParams{
		Key:   "feedback",
		Value: value,
	}); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update settings")
	}
	return c.JSON(http.StatusOK, req)
}
