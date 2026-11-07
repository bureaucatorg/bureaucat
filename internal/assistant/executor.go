package assistant

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"

	"bereaucat/internal/auth"
	"bereaucat/internal/store"
)

const maxToolOutput = 16 * 1024

// executor runs catalog tools against the app's own HTTP router as the given user,
// so every existing auth, role and activity-log path applies unchanged.
type executor struct {
	router http.Handler
	auth   *auth.Manager
}

// RunTool runs one catalog tool directly, without the agent loop or an LLM.
func (s *Service) RunTool(ctx context.Context, user store.GetUserByIDRow, name string, args json.RawMessage, via string) (string, bool) {
	tool := toolByName(name)
	if tool == nil {
		return fmt.Sprintf("unknown tool %q", name), true
	}
	return s.exec.run(ctx, user, tool, args, via)
}

// BuildRequest resolves args into the tool's /api/v1 target (path + query) and JSON body.
// The body is nil for GET. Shared by the in-process executor and the HTTP CLI.
func (t Tool) BuildRequest(args map[string]any) (string, []byte, error) {
	path := t.Path
	qs := url.Values{}
	payload := map[string]any{}
	for _, p := range t.Params {
		v, ok := args[p.Name]
		if !ok || (v == nil && p.In != InBody) {
			if p.Required {
				return "", nil, fmt.Errorf("missing required argument %q", p.Name)
			}
			continue
		}
		switch {
		case p.In == InPath:
			// Tokens are minted for the real user, so never let an argument escape the catalogued route.
			seg := scalar(v)
			if seg == "" || seg == "." || seg == ".." {
				return "", nil, fmt.Errorf("invalid value for %q", p.Name)
			}
			path = strings.ReplaceAll(path, "{"+p.Name+"}", url.PathEscape(seg))
		case p.In == InQuery && p.Type == "object":
			// The task list API takes its filter tree as base64url JSON in ?f=.
			raw, _ := json.Marshal(v)
			qs.Set("f", base64.RawURLEncoding.EncodeToString(raw))
		case p.In == InQuery:
			qs.Set(p.Name, scalar(v))
		default:
			payload[p.Name] = v
		}
	}

	target := "/api/v1" + path
	if len(qs) > 0 {
		target += "?" + qs.Encode()
	}
	if t.Method == http.MethodGet {
		return target, nil, nil
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", nil, err
	}
	return target, body, nil
}

func (e *executor) run(ctx context.Context, user store.GetUserByIDRow, tool *Tool, rawArgs json.RawMessage, via string) (string, bool) {
	target, body, err := tool.BuildRequest(argsObject(rawArgs))
	if err != nil {
		return err.Error(), true
	}
	var reqBody io.Reader
	if body != nil {
		reqBody = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, tool.Method, target, reqBody)
	if err != nil {
		return err.Error(), true
	}
	token, _, err := e.auth.GenerateToolToken(user.ID, user.Username, user.UserType, via)
	if err != nil {
		return "failed to authenticate tool call", true
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "127.0.0.1:0"

	rec := httptest.NewRecorder()
	e.router.ServeHTTP(rec, req)

	out := truncate(rec.Body.String(), maxToolOutput)
	if rec.Code >= 400 {
		return fmt.Sprintf("HTTP %d: %s", rec.Code, out), true
	}
	if out == "" {
		out = fmt.Sprintf("HTTP %d: ok", rec.Code)
	}
	return out, false
}

func scalar(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	case bool:
		return fmt.Sprintf("%t", t)
	default:
		b, _ := json.Marshal(t)
		return string(b)
	}
}
