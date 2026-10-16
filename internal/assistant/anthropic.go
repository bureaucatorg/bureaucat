package assistant

import (
	"context"
	"encoding/json"
)

type anthropicProvider struct {
	apiKey string
}

type antBlock struct {
	Type      string         `json:"type"`
	Text      string         `json:"text,omitempty"`
	ID        string         `json:"id,omitempty"`
	Name      string         `json:"name,omitempty"`
	Input     any            `json:"input,omitempty"`
	ToolUseID string         `json:"tool_use_id,omitempty"`
	Content   string         `json:"content,omitempty"`
	IsError   bool           `json:"is_error,omitempty"`
}

type antMessage struct {
	Role    string     `json:"role"`
	Content []antBlock `json:"content"`
}

type antResponse struct {
	Content []struct {
		Type  string          `json:"type"`
		Text  string          `json:"text"`
		ID    string          `json:"id"`
		Name  string          `json:"name"`
		Input json.RawMessage `json:"input"`
	} `json:"content"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

func (p *anthropicProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	var msgs []antMessage
	add := func(role string, blocks ...antBlock) {
		if n := len(msgs); n > 0 && msgs[n-1].Role == role {
			msgs[n-1].Content = append(msgs[n-1].Content, blocks...)
			return
		}
		msgs = append(msgs, antMessage{Role: role, Content: blocks})
	}

	for _, m := range req.Messages {
		switch m.Role {
		case RoleUser:
			add("user", antBlock{Type: "text", Text: m.Text})
		case RoleAssistant:
			var blocks []antBlock
			if m.Text != "" {
				blocks = append(blocks, antBlock{Type: "text", Text: m.Text})
			}
			for _, tc := range m.ToolCalls {
				input := argsObject(tc.Args)
				blocks = append(blocks, antBlock{Type: "tool_use", ID: tc.ID, Name: tc.Name, Input: input})
			}
			if len(blocks) > 0 {
				add("assistant", blocks...)
			}
		case RoleTool:
			for _, r := range m.ToolResults {
				add("user", antBlock{Type: "tool_result", ToolUseID: r.CallID, Content: r.Content, IsError: r.IsError})
			}
		}
	}

	tools := make([]map[string]any, 0, len(req.Tools))
	for _, t := range req.Tools {
		tools = append(tools, map[string]any{"name": t.Name, "description": t.Description, "input_schema": t.Schema})
	}

	body := map[string]any{
		"model":      req.Model,
		"max_tokens": maxOutputTokens,
		"system":     req.System,
		"messages":   msgs,
	}
	if len(tools) > 0 {
		body["tools"] = tools
	}

	headers := map[string]string{"x-api-key": p.apiKey, "anthropic-version": "2023-06-01"}
	var out antResponse
	if err := postJSON(ctx, "https://api.anthropic.com/v1/messages", headers, body, &out); err != nil {
		return nil, err
	}

	resp := &ChatResponse{InputTokens: out.Usage.InputTokens, OutputTokens: out.Usage.OutputTokens}
	for _, b := range out.Content {
		switch b.Type {
		case "text":
			resp.Text += b.Text
		case "tool_use":
			args := b.Input
			if len(args) == 0 {
				args = json.RawMessage("{}")
			}
			resp.ToolCalls = append(resp.ToolCalls, ToolCall{ID: b.ID, Name: b.Name, Args: args})
		}
	}
	return resp, nil
}
