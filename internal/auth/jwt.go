package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

// Config holds JWT and token configuration.
type Config struct {
	JWTSecret              string
	AccessTokenExpiryMins  int
	RefreshTokenExpiryDays int
}

// UserClaims represents the claims stored in the JWT.
type UserClaims struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	UserType string    `json:"user_type"`
	jwt.RegisteredClaims
}

// Manager handles JWT token generation and validation.
type Manager struct {
	config Config
}

// NewManager creates a new JWT manager with the given config.
func NewManager(config Config) *Manager {
	return &Manager{config: config}
}

// GenerateAccessToken creates a JWT access token for the user.
func (m *Manager) GenerateAccessToken(userID uuid.UUID, username, userType string) (string, time.Time, error) {
	expiresAt := time.Now().Add(time.Duration(m.config.AccessTokenExpiryMins) * time.Minute)

	claims := UserClaims{
		UserID:   userID,
		Username: username,
		UserType: userType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(m.config.JWTSecret))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

// GenerateRefreshToken creates a random 32-byte base64-encoded refresh token.
func (m *Manager) GenerateRefreshToken() (string, time.Time, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", time.Time{}, err
	}

	token := base64.URLEncoding.EncodeToString(bytes)
	expiresAt := time.Now().Add(time.Duration(m.config.RefreshTokenExpiryDays) * 24 * time.Hour)

	return token, expiresAt, nil
}

// ValidateAccessToken parses and validates a JWT access token.
func (m *Manager) ValidateAccessToken(tokenString string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(m.config.JWTSecret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// GetRefreshTokenExpiryDays returns the configured refresh token expiry in days.
func (m *Manager) GetRefreshTokenExpiryDays() int {
	return m.config.RefreshTokenExpiryDays
}

// GetJWTSecret returns the JWT secret (used for HMAC signing of OAuth state).
func (m *Manager) GetJWTSecret() string {
	return m.config.JWTSecret
}
