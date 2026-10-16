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

func (e *executor) run(ctx context.Context, user store.GetUserByIDRow, tool *Tool, rawArgs json.RawMessage) (string, bool) {
	args := argsObject(rawArgs)

	path := tool.Path
	qs := url.Values{}
	payload := map[string]any{}
	for _, p := range tool.Params {
		v, ok := args[p.Name]
		if !ok || (v == nil && p.In != InBody) {
			if p.Required {
				return fmt.Sprintf("missing required argument %q", p.Name), true
			}
			continue
		}
		switch {
		case p.In == InPath:
			// Tokens are minted for the real user, so never let an argument escape the catalogued route.
			seg := scalar(v)
			if seg == "" || seg == "." || seg == ".." {
				return fmt.Sprintf("invalid value for %q", p.Name), true
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

	var reqBody io.Reader
	if tool.Method != http.MethodGet {
		raw, err := json.Marshal(payload)
		if err != nil {
			return err.Error(), true
		}
		reqBody = bytes.NewReader(raw)
	}

	req, err := http.NewRequestWithContext(ctx, tool.Method, target, reqBody)
	if err != nil {
		return err.Error(), true
	}
	token, _, err := e.auth.GenerateAssistantToken(user.ID, user.Username, user.UserType)
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
