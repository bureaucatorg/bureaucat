package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v5"

	"bereaucat/internal/auth"
	"bereaucat/internal/store"
)

// PATHandler handles Personal Access Token endpoints.
type PATHandler struct {
	store store.Querier
}

// NewPATHandler creates a new PATHandler.
func NewPATHandler(store store.Querier) *PATHandler {
	return &PATHandler{store: store}
}

type createTokenRequest struct {
	Name      string  `json:"name"`
	ExpiresAt *string `json:"expires_at"`
}

type tokenResponse struct {
	ID         uuid.UUID  `json:"id"`
	Name       string     `json:"name"`
	Token      string     `json:"token,omitempty"`
	ExpiresAt  *time.Time `json:"expires_at"`
	LastUsedAt *time.Time `json:"last_used_at"`
	CreatedAt  time.Time  `json:"created_at"`
}

// CreateToken creates a new Personal Access Token.
func (h *PATHandler) CreateToken(c *echo.Context) error {
	userID, err := uuid.Parse(c.Request().Header.Get("X-User-ID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid user")
	}

	var req createTokenRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" || len(req.Name) > 100 {
		return echo.NewHTTPError(http.StatusBadRequest, "name is required and must be 100 characters or less")
	}

	// Generate token: bcat_ + 64 hex chars (32 random bytes)
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate token")
	}
	plaintext := "bcat_" + hex.EncodeToString(randomBytes)
	tokenHash := auth.HashToken(plaintext)

	// Parse optional expiry
	var expiresAt pgtype.Timestamptz
	if req.ExpiresAt != nil && *req.ExpiresAt != "" {
		t, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid expires_at format, use RFC3339")
		}
		if t.Before(time.Now()) {
			return echo.NewHTTPError(http.StatusBadRequest, "expires_at must be in the future")
		}
		expiresAt = pgtype.Timestamptz{Time: t, Valid: true}
	}

	pat, err := h.store.CreatePersonalAccessToken(c.Request().Context(), store.CreatePersonalAccessTokenParams{
		UserID:    userID,
		Name:      req.Name,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create token")
	}

	resp := tokenResponse{
		ID:        pat.ID,
		Name:      pat.Name,
		Token:     plaintext,
		CreatedAt: pat.CreatedAt.Time,
	}
	if pat.ExpiresAt.Valid {
		resp.ExpiresAt = &pat.ExpiresAt.Time
	}

	return c.JSON(http.StatusCreated, resp)
}

// ListTokens lists all Personal Access Tokens for the current user.
func (h *PATHandler) ListTokens(c *echo.Context) error {
	userID, err := uuid.Parse(c.Request().Header.Get("X-User-ID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid user")
	}

	tokens, err := h.store.ListPersonalAccessTokensByUser(c.Request().Context(), userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list tokens")
	}

	resp := make([]tokenResponse, len(tokens))
	for i, t := range tokens {
		resp[i] = tokenResponse{
			ID:        t.ID,
			Name:      t.Name,
			CreatedAt: t.CreatedAt.Time,
		}
		if t.ExpiresAt.Valid {
			resp[i].ExpiresAt = &t.ExpiresAt.Time
		}
		if t.LastUsedAt.Valid {
			resp[i].LastUsedAt = &t.LastUsedAt.Time
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"tokens": resp,
	})
}

// DeleteToken deletes a Personal Access Token.
func (h *PATHandler) DeleteToken(c *echo.Context) error {
	userID, err := uuid.Parse(c.Request().Header.Get("X-User-ID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid user")
	}

	tokenID, err := uuid.Parse(c.Param("tokenId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid token ID")
	}

	err = h.store.DeletePersonalAccessToken(c.Request().Context(), store.DeletePersonalAccessTokenParams{
		ID:     tokenID,
		UserID: userID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete token")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Token deleted",
	})
}

