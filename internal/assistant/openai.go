package assistant

import (
	"context"
	"encoding/json"
	"errors"
)

// openAIProvider speaks the Chat Completions API, which OpenRouter also implements.
type openAIProvider struct {
	baseURL string
	apiKey  string
}

type oaiMessage struct {
	Role       string        `json:"role"`
	Content    *string       `json:"content"`
	ToolCalls  []oaiToolCall `json:"tool_calls,omitempty"`
	ToolCallID string        `json:"tool_call_id,omitempty"`
}

type oaiToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type oaiTool struct {
	Type     string `json:"type"`
	Function struct {
		Name        string         `json:"name"`
		Description string         `json:"description"`
		Parameters  map[string]any `json:"parameters"`
	} `json:"function"`
}

type oaiResponse struct {
	Choices []struct {
		Message oaiMessage `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
}

func (p *openAIProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	msgs := []oaiMessage{{Role: "system", Content: &req.System}}
	for _, m := range req.Messages {
		switch m.Role {
		case RoleUser:
			msgs = append(msgs, oaiMessage{Role: "user", Content: strPtr(m.Text)})
		case RoleAssistant:
			om := oaiMessage{Role: "assistant"}
			if m.Text != "" {
				om.Content = strPtr(m.Text)
			}
			for _, tc := range m.ToolCalls {
				var c oaiToolCall
				c.ID, c.Type = tc.ID, "function"
				c.Function.Name = tc.Name
				c.Function.Arguments = string(tc.Args)
				if c.Function.Arguments == "" {
					c.Function.Arguments = "{}"
				}
				om.ToolCalls = append(om.ToolCalls, c)
			}
			msgs = append(msgs, om)
		case RoleTool:
			for _, r := range m.ToolResults {
				msgs = append(msgs, oaiMessage{Role: "tool", ToolCallID: r.CallID, Content: strPtr(r.Content)})
			}
		}
	}

	tools := make([]oaiTool, 0, len(req.Tools))
	for _, t := range req.Tools {
		var ot oaiTool
		ot.Type = "function"
		ot.Function.Name, ot.Function.Description, ot.Function.Parameters = t.Name, t.Description, t.Schema
		tools = append(tools, ot)
	}

	body := map[string]any{"model": req.Model, "messages": msgs}
	if len(tools) > 0 {
		body["tools"] = tools
	}

	var out oaiResponse
	headers := map[string]string{"Authorization": "Bearer " + p.apiKey}
	if err := postJSON(ctx, p.baseURL+"/chat/completions", headers, body, &out); err != nil {
		return nil, err
	}
	if len(out.Choices) == 0 {
		return nil, errors.New("provider returned no choices")
	}

	msg := out.Choices[0].Message
	resp := &ChatResponse{InputTokens: out.Usage.PromptTokens, OutputTokens: out.Usage.CompletionTokens}
	if msg.Content != nil {
		resp.Text = *msg.Content
	}
	for _, tc := range msg.ToolCalls {
		args := json.RawMessage(tc.Function.Arguments)
		if !json.Valid(args) {
			args = json.RawMessage("{}")
		}
		resp.ToolCalls = append(resp.ToolCalls, ToolCall{ID: tc.ID, Name: tc.Function.Name, Args: args})
	}
	return resp, nil
}

func strPtr(s string) *string { return &s }
