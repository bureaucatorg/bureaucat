// Package mcpserver exposes the assistant's tool catalog over the Model Context Protocol.
package mcpserver

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"bereaucat/internal/assistant"
	"bereaucat/internal/auth"
	"bereaucat/internal/buildinfo"
	"bereaucat/internal/store"
)

type userKey struct{}

// Handler serves stateless Streamable HTTP. It must run behind auth.Middleware and only
// accepts PATs; a read_only PAT gets only GET tools.
func Handler(s store.Querier, svc *assistant.Service) http.Handler {
	mcpHandler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		user, _ := r.Context().Value(userKey{}).(store.GetUserByIDRow)
		readOnly := r.Header.Get(auth.HeaderPATScope) == auth.PATScopeReadOnly
		return newServer(svc, user, readOnly)
	}, &mcp.StreamableHTTPOptions{Stateless: true})

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(auth.HeaderAuthMethod) != auth.AuthMethodPAT {
			http.Error(w, "a personal access token is required", http.StatusUnauthorized)
			return
		}
		userID, err := uuid.Parse(r.Header.Get(auth.HeaderUserID))
		if err != nil {
			http.Error(w, "invalid user", http.StatusUnauthorized)
			return
		}
		user, err := s.GetUserByID(r.Context(), userID)
		if err != nil {
			http.Error(w, "invalid user", http.StatusUnauthorized)
			return
		}
		mcpHandler.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userKey{}, user)))
	})
}

func newServer(svc *assistant.Service, user store.GetUserByIDRow, readOnly bool) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{Name: "bureaucat", Title: "Bureaucat", Version: buildinfo.Version}, nil)
	for _, t := range allowedTools(readOnly) {
		name := t.Name
		server.AddTool(toMCPTool(t), func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			out, isErr := svc.RunTool(ctx, user, name, req.Params.Arguments, auth.ViaMCP)
			return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: out}}, IsError: isErr}, nil
		})
	}
	return server
}

// allowedTools drops tools that need a chat upload, and writes for read-only tokens.
// Filtering here matters: tool calls run with a full-scope token minted for the user.
func allowedTools(readOnly bool) []assistant.Tool {
	var out []assistant.Tool
	for _, t := range assistant.Catalog() {
		if t.UsesUpload() || (readOnly && t.Method != http.MethodGet) {
			continue
		}
		out = append(out, t)
	}
	return out
}

func toMCPTool(t assistant.Tool) *mcp.Tool {
	readOnly := t.Method == http.MethodGet
	destructive := t.Confirm
	openWorld := false
	return &mcp.Tool{
		Name:        t.Name,
		Title:       t.Title,
		Description: t.Description,
		InputSchema: t.InputSchema(),
		Annotations: &mcp.ToolAnnotations{
			Title:           t.Title,
			ReadOnlyHint:    readOnly,
			DestructiveHint: &destructive,
			IdempotentHint:  readOnly,
			OpenWorldHint:   &openWorld,
		},
	}
}
