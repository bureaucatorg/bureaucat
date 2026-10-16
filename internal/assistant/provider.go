package assistant

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	ProviderOpenAI     = "openai"
	ProviderAnthropic  = "anthropic"
	ProviderOpenRouter = "openrouter"
	ProviderBedrock    = "bedrock"
)

const (
	RoleUser      = "user"
	RoleAssistant = "assistant"
	RoleTool      = "tool"
)

const maxOutputTokens = 4096

// ToolCall is a tool invocation requested by the model.
type ToolCall struct {
	ID   string          `json:"id"`
	Name string          `json:"name"`
	Args json.RawMessage `json:"args"`
}

// ToolResult is the output of a tool call fed back to the model.
type ToolResult struct {
	CallID  string
	Name    string
	Content string
	IsError bool
}

// Message is a provider-neutral chat message. Role tool carries ToolResults only.
type Message struct {
	Role        string
	Text        string
	ToolCalls   []ToolCall
	ToolResults []ToolResult
}

// ToolSpec describes a tool to the model.
type ToolSpec struct {
	Name        string
	Description string
	Schema      map[string]any
}

type ChatRequest struct {
	Model    string
	System   string
	Messages []Message
	Tools    []ToolSpec
}

type ChatResponse struct {
	Text         string
	ToolCalls    []ToolCall
	InputTokens  int
	OutputTokens int
}

// Provider is a single LLM backend.
type Provider interface {
	Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)
}

// NewProvider builds the provider configured by the admin.
func NewProvider(ctx context.Context, s Settings) (Provider, error) {
	switch s.Provider {
	case ProviderOpenAI:
		return &openAIProvider{baseURL: "https://api.openai.com/v1", apiKey: s.APIKey}, nil
	case ProviderOpenRouter:
		return &openAIProvider{baseURL: "https://openrouter.ai/api/v1", apiKey: s.APIKey}, nil
	case ProviderAnthropic:
		return &anthropicProvider{apiKey: s.APIKey}, nil
	case ProviderBedrock:
		return newBedrockProvider(ctx, s.AWSRegion)
	default:
		return nil, fmt.Errorf("unknown provider %q", s.Provider)
	}
}

var httpClient = &http.Client{Timeout: 3 * time.Minute}

func postJSON(ctx context.Context, url string, headers map[string]string, body, out any) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return err
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("provider returned %d: %s", resp.StatusCode, truncate(string(data), 500))
	}
	return json.Unmarshal(data, out)
}

// argsObject guarantees tool arguments decode to a JSON object.
func argsObject(raw json.RawMessage) map[string]any {
	m := map[string]any{}
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &m)
	}
	return m
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…[truncated]"
}
