package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
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

func (c *Client) Get(path string, query map[string]string, out interface{}) error {
	return c.Do(http.MethodGet, path, query, nil, out)
}

func (c *Client) Post(path string, body interface{}, out interface{}) error {
	return c.Do(http.MethodPost, path, nil, body, out)
}

func (c *Client) Patch(path string, body interface{}, out interface{}) error {
	return c.Do(http.MethodPatch, path, nil, body, out)
}

func (c *Client) Delete(path string, out interface{}) error {
	return c.Do(http.MethodDelete, path, nil, nil, out)
}

// Do performs a JSON request with bearer auth.
func (c *Client) Do(method, path string, query map[string]string, body interface{}, out interface{}) error {
	endpoint, err := url.Parse(c.baseURL + "/api/v1" + path)
	if err != nil {
		return fmt.Errorf("build request url: %w", err)
	}

	q := endpoint.Query()
	for key, value := range query {
		if strings.TrimSpace(value) != "" {
			q.Set(key, value)
		}
	}
	endpoint.RawQuery = q.Encode()

	var payload io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		payload = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, endpoint.String(), payload)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
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
		return fmt.Errorf("perform request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return ErrUnauthorized
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errResp struct {
			Message string `json:"message"`
		}
		if err := json.Unmarshal(respBody, &errResp); err == nil && errResp.Message != "" {
			return &APIError{StatusCode: resp.StatusCode, Message: errResp.Message}
		}
		return &APIError{StatusCode: resp.StatusCode, Message: strings.TrimSpace(string(respBody))}
	}

	if out == nil || len(bytes.TrimSpace(respBody)) == 0 {
		return nil
	}

	if raw, ok := out.(*json.RawMessage); ok {
		*raw = append((*raw)[:0], respBody...)
		return nil
	}

	if err := json.Unmarshal(respBody, out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	return nil
}
