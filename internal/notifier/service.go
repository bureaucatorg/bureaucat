package notifier

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"

	"bereaucat/internal/store"
)

// Service dispatches notifications to all enabled providers.
// It reads provider configuration from the database on each call,
// with a short cache to avoid excessive DB reads.
type Service struct {
	store store.Querier

	mu           sync.RWMutex
	cachedAt     time.Time
	cachedProvs  []Notifier
	cacheTTL     time.Duration
}

// NewService creates a notification service that dynamically loads providers from settings.
func NewService(s store.Querier) *Service {
	return &Service{
		store:    s,
		cacheTTL: 30 * time.Second,
	}
}

// Notify sends a notification to a single recipient via all enabled providers.
// Sends are dispatched asynchronously to avoid blocking the caller.
func (s *Service) Notify(ctx context.Context, n Notification) {
	providers := s.getProviders()
	if len(providers) == 0 {
		return
	}

	email, err := s.getRecipientEmail(ctx, n.RecipientID)
	if err != nil {
		log.Printf("notifier: failed to look up email for user %s: %v", n.RecipientID, err)
		return
	}

	for _, provider := range providers {
		go func(p Notifier) {
			if err := p.Send(context.Background(), email, n); err != nil {
				log.Printf("notifier [%s]: failed to send to %s: %v", p.Name(), email, err)
			}
		}(provider)
	}
}

// NotifyAll sends notifications to multiple recipients via all providers.
func (s *Service) NotifyAll(ctx context.Context, notifications []Notification) {
	for _, n := range notifications {
		s.Notify(ctx, n)
	}
}

// NotifyExcluding sends notifications to all recipients except the excluded user.
func (s *Service) NotifyExcluding(ctx context.Context, notifications []Notification, excludeUserID uuid.UUID) {
	for _, n := range notifications {
		if n.RecipientID != excludeUserID {
			s.Notify(ctx, n)
		}
	}
}

// InvalidateCache forces the next call to reload providers from the database.
func (s *Service) InvalidateCache() {
	s.mu.Lock()
	s.cachedAt = time.Time{}
	s.mu.Unlock()
}

func (s *Service) getProviders() []Notifier {
	s.mu.RLock()
	if !s.cachedAt.IsZero() && time.Since(s.cachedAt) < s.cacheTTL {
		provs := s.cachedProvs
		s.mu.RUnlock()
		return provs
	}
	s.mu.RUnlock()

	// Reload from DB
	s.mu.Lock()
	defer s.mu.Unlock()

	// Double-check after acquiring write lock
	if !s.cachedAt.IsZero() && time.Since(s.cachedAt) < s.cacheTTL {
		return s.cachedProvs
	}

	var providers []Notifier

	// Load Mattermost
	setting, err := s.store.GetSetting(context.Background(), "mattermost")
	if err == nil {
		var mm struct {
			Enabled   bool   `json:"enabled"`
			ServerURL string `json:"server_url"`
			BotToken  string `json:"bot_token"`
		}
		if json.Unmarshal(setting.Value, &mm) == nil && mm.Enabled && mm.ServerURL != "" && mm.BotToken != "" {
			providers = append(providers, NewMattermostNotifier(MattermostConfig{
				ServerURL: mm.ServerURL,
				BotToken:  mm.BotToken,
			}))
		}
	}

	s.cachedProvs = providers
	s.cachedAt = time.Now()
	return providers
}

func (s *Service) getRecipientEmail(ctx context.Context, userID uuid.UUID) (string, error) {
	user, err := s.store.GetUserByID(ctx, userID)
	if err != nil {
		return "", err
	}
	return user.Email, nil
}
