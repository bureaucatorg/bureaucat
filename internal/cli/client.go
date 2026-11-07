package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

var ErrUnauthorized = errors.New("unauthorized")

// Client is a thin HTTP client for the Bureaucat API.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// APIError represents a non-2xx API response.
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	if e.Message == "" {
		return fmt.Sprintf("api error (%d)", e.StatusCode)
	}
	return fmt.Sprintf("api error (%d): %s", e.StatusCode, e.Message)
}

// NewClient creates a client with explicit credentials.
func NewClient(baseURL, token string) *Client {
	baseURL = strings.TrimSpace(baseURL)
	baseURL = strings.TrimRight(baseURL, "/")
	baseURL = strings.TrimSuffix(baseURL, "/api/v1")

	return &Client{
		baseURL:    baseURL,
		token:      strings.TrimSpace(token),
		httpClient: &http.Client{},
	}
}

// NewClientFromEnv creates a client using env/config credentials.
func NewClientFromEnv() (*Client, error) {
	baseURL, token, err := GetCredentials()
	if err != nil {
		return nil, err
	}
	return NewClient(baseURL, token), nil
}

func (c *Client) Get(path string, out interface{}) error {
	respBody, err := c.Send(http.MethodGet, "/api/v1"+path, nil)
	if err != nil || out == nil || len(bytes.TrimSpace(respBody)) == 0 {
		return err
	}
	if err := json.Unmarshal(respBody, out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

// Send performs a bearer-authenticated request; target is a server-relative path with query.
func (c *Client) Send(method, target string, body []byte) ([]byte, error) {
	var payload io.Reader
	if body != nil {
		payload = bytes.NewReader(body)
	}

	req, err := http.NewRequest(method, c.baseURL+target, payload)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("perform request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrUnauthorized
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errResp struct {
			Message string `json:"message"`
		}
		if err := json.Unmarshal(respBody, &errResp); err == nil && errResp.Message != "" {
			return nil, &APIError{StatusCode: resp.StatusCode, Message: errResp.Message}
		}
		return nil, &APIError{StatusCode: resp.StatusCode, Message: strings.TrimSpace(string(respBody))}
	}

	return respBody, nil
}
