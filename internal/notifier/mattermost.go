package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// MattermostConfig holds the configuration for the Mattermost notifier.
type MattermostConfig struct {
	ServerURL string
	BotToken  string
}

// MattermostNotifier sends DM notifications via a Mattermost bot.
type MattermostNotifier struct {
	serverURL string
	botToken  string
	botUserID string
	client    *http.Client

	// Cache email → Mattermost user ID to avoid repeated API calls.
	mu         sync.RWMutex
	emailCache map[string]string
}

// NewMattermostNotifier creates a new Mattermost notifier.
func NewMattermostNotifier(cfg MattermostConfig) *MattermostNotifier {
	serverURL := strings.TrimRight(cfg.ServerURL, "/")
	return &MattermostNotifier{
		serverURL:  serverURL,
		botToken:   cfg.BotToken,
		client:     &http.Client{Timeout: 10 * time.Second},
		emailCache: make(map[string]string),
	}
}

func (m *MattermostNotifier) Name() string {
	return "mattermost"
}

// Send delivers a DM to the Mattermost user matching the given email.
func (m *MattermostNotifier) Send(ctx context.Context, recipientEmail string, n Notification) error {
	// Look up Mattermost user ID by email
	mmUserID, err := m.getUserIDByEmail(ctx, recipientEmail)
	if err != nil {
		return fmt.Errorf("lookup user by email %s: %w", recipientEmail, err)
	}

	// Get bot's own user ID (needed to create DM channel)
	botID, err := m.getBotUserID(ctx)
	if err != nil {
		return fmt.Errorf("get bot user ID: %w", err)
	}

	// Create or get existing DM channel
	channelID, err := m.getDirectChannel(ctx, botID, mmUserID)
	if err != nil {
		return fmt.Errorf("create DM channel: %w", err)
	}

	// Build message
	message := formatMessage(n)

	// Post message
	return m.postMessage(ctx, channelID, message)
}

// TestConnection verifies that the bot token is valid and can reach the server.
func (m *MattermostNotifier) TestConnection(ctx context.Context) error {
	_, err := m.getBotUserID(ctx)
	return err
}

func (m *MattermostNotifier) getUserIDByEmail(ctx context.Context, email string) (string, error) {
	// Check cache
	m.mu.RLock()
	if id, ok := m.emailCache[email]; ok {
		m.mu.RUnlock()
		return id, nil
	}
	m.mu.RUnlock()

	// Call Mattermost API
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		m.serverURL+"/api/v4/users/email/"+email, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+m.botToken)

	resp, err := m.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("mattermost API returned %d: %s", resp.StatusCode, string(body))
	}

	var user struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return "", err
	}

	// Cache result
	m.mu.Lock()
	m.emailCache[email] = user.ID
	m.mu.Unlock()

	return user.ID, nil
}

func (m *MattermostNotifier) getBotUserID(ctx context.Context) (string, error) {
	if m.botUserID != "" {
		return m.botUserID, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		m.serverURL+"/api/v4/users/me", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+m.botToken)

	resp, err := m.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("mattermost API returned %d: %s", resp.StatusCode, string(body))
	}

	var user struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return "", err
	}

	m.botUserID = user.ID
	return m.botUserID, nil
}

func (m *MattermostNotifier) getDirectChannel(ctx context.Context, botID, userID string) (string, error) {
	body, err := json.Marshal([]string{botID, userID})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		m.serverURL+"/api/v4/channels/direct", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+m.botToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := m.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("mattermost API returned %d: %s", resp.StatusCode, string(respBody))
	}

	var channel struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&channel); err != nil {
		return "", err
	}
	return channel.ID, nil
}

func (m *MattermostNotifier) postMessage(ctx context.Context, channelID, message string) error {
	payload, err := json.Marshal(map[string]string{
		"channel_id": channelID,
		"message":    message,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		m.serverURL+"/api/v4/posts", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+m.botToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := m.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("mattermost API returned %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func formatMessage(n Notification) string {
	taskRef := fmt.Sprintf("%s-%d: %s", n.ProjectKey, n.TaskNumber, n.TaskTitle)

	switch n.Event {
	case EventTaskAssigned:
		return fmt.Sprintf("**%s** assigned you to **%s**", n.ActorName, taskRef)
	case EventMentioned:
		return fmt.Sprintf("**%s** mentioned you in **%s**", n.ActorName, taskRef)
	default:
		return fmt.Sprintf("**%s** — %s", taskRef, n.ActorName)
	}
}
