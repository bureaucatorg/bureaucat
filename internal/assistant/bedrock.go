package assistant

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/document"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
)

// bedrockProvider uses the Converse API with the host's default AWS credential chain.
type bedrockProvider struct {
	client *bedrockruntime.Client
}

func newBedrockProvider(ctx context.Context, region string) (*bedrockProvider, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, err
	}
	return &bedrockProvider{client: bedrockruntime.NewFromConfig(cfg)}, nil
}

func (p *bedrockProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	var msgs []types.Message
	add := func(role types.ConversationRole, blocks ...types.ContentBlock) {
		if n := len(msgs); n > 0 && msgs[n-1].Role == role {
			msgs[n-1].Content = append(msgs[n-1].Content, blocks...)
			return
		}
		msgs = append(msgs, types.Message{Role: role, Content: blocks})
	}

	for _, m := range req.Messages {
		switch m.Role {
		case RoleUser:
			add(types.ConversationRoleUser, &types.ContentBlockMemberText{Value: m.Text})
		case RoleAssistant:
			var blocks []types.ContentBlock
			if m.Text != "" {
				blocks = append(blocks, &types.ContentBlockMemberText{Value: m.Text})
			}
			for _, tc := range m.ToolCalls {
				blocks = append(blocks, &types.ContentBlockMemberToolUse{Value: types.ToolUseBlock{
					ToolUseId: aws.String(tc.ID),
					Name:      aws.String(tc.Name),
					Input:     document.NewLazyDocument(argsObject(tc.Args)),
				}})
			}
			if len(blocks) > 0 {
				add(types.ConversationRoleAssistant, blocks...)
			}
		case RoleTool:
			for _, r := range m.ToolResults {
				block := types.ToolResultBlock{
					ToolUseId: aws.String(r.CallID),
					Content:   []types.ToolResultContentBlock{&types.ToolResultContentBlockMemberText{Value: r.Content}},
				}
				if r.IsError {
					block.Status = types.ToolResultStatusError
				}
				add(types.ConversationRoleUser, &types.ContentBlockMemberToolResult{Value: block})
			}
		}
	}

	input := &bedrockruntime.ConverseInput{
		ModelId:         aws.String(req.Model),
		Messages:        msgs,
		System:          []types.SystemContentBlock{&types.SystemContentBlockMemberText{Value: req.System}},
		InferenceConfig: &types.InferenceConfiguration{MaxTokens: aws.Int32(maxOutputTokens)},
	}
	if len(req.Tools) > 0 {
		tools := make([]types.Tool, 0, len(req.Tools))
		for _, t := range req.Tools {
			tools = append(tools, &types.ToolMemberToolSpec{Value: types.ToolSpecification{
				Name:        aws.String(t.Name),
				Description: aws.String(t.Description),
				InputSchema: &types.ToolInputSchemaMemberJson{Value: document.NewLazyDocument(t.Schema)},
			}})
		}
		input.ToolConfig = &types.ToolConfiguration{Tools: tools}
	}

	out, err := p.client.Converse(ctx, input)
	if err != nil {
		return nil, err
	}
	msg, ok := out.Output.(*types.ConverseOutputMemberMessage)
	if !ok {
		return nil, errors.New("bedrock returned no message")
	}

	resp := &ChatResponse{}
	if out.Usage != nil {
		resp.InputTokens = int(aws.ToInt32(out.Usage.InputTokens))
		resp.OutputTokens = int(aws.ToInt32(out.Usage.OutputTokens))
	}
	for _, b := range msg.Value.Content {
		switch v := b.(type) {
		case *types.ContentBlockMemberText:
			resp.Text += v.Value
		case *types.ContentBlockMemberToolUse:
			args := json.RawMessage("{}")
			if v.Value.Input != nil {
				if raw, err := v.Value.Input.MarshalSmithyDocument(); err == nil {
					args = raw
				}
			}
			resp.ToolCalls = append(resp.ToolCalls, ToolCall{
				ID:   aws.ToString(v.Value.ToolUseId),
				Name: aws.ToString(v.Value.Name),
				Args: args,
			})
		}
	}
	return resp, nil
}
