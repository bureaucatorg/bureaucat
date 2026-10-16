package mattermostbot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

type client struct {
	serverURL string
	token     string
	http      *http.Client
}

type mmUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	IsBot    bool   `json:"is_bot"`
}

func newClient(cfg config) *client {
	return &client{
		serverURL: strings.TrimRight(cfg.ServerURL, "/"),
		token:     cfg.BotToken,
		http:      &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *client) do(ctx context.Context, method, path string, body, out any) error {
	var r io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		r = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.serverURL+"/api/v4"+path, r)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("mattermost %s %s returned %d: %s", method, path, resp.StatusCode, msg)
	}
	if out == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func (c *client) user(ctx context.Context, id string) (mmUser, error) {
	var u mmUser
	err := c.do(ctx, http.MethodGet, "/users/"+url.PathEscape(id), nil, &u)
	return u, err
}

func (c *client) createPost(ctx context.Context, channelID, rootID, message string) error {
	return c.do(ctx, http.MethodPost, "/posts", map[string]string{
		"channel_id": channelID,
		"root_id":    rootID,
		"message":    message,
	}, nil)
}

func (c *client) typing(ctx context.Context, botID, channelID, parentID string) error {
	return c.do(ctx, http.MethodPost, "/users/"+url.PathEscape(botID)+"/typing", map[string]string{
		"channel_id": channelID,
		"parent_id":  parentID,
	}, nil)
}

func (c *client) addReaction(ctx context.Context, botID, postID, emoji string) error {
	return c.do(ctx, http.MethodPost, "/reactions", map[string]string{
		"user_id":    botID,
		"post_id":    postID,
		"emoji_name": emoji,
	}, nil)
}

func (c *client) removeReaction(ctx context.Context, botID, postID, emoji string) error {
	return c.do(ctx, http.MethodDelete, "/users/"+url.PathEscape(botID)+"/posts/"+url.PathEscape(postID)+"/reactions/"+url.PathEscape(emoji), nil, nil)
}

type postList struct {
	Order []string        `json:"order"`
	Posts map[string]post `json:"posts"`
}

func (l postList) oldestFirst() []post {
	out := make([]post, 0, len(l.Order))
	for _, id := range l.Order {
		if p, ok := l.Posts[id]; ok {
			out = append(out, p)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreateAt < out[j].CreateAt })
	return out
}

// channelPostsBefore lists top-level posts only; collapsedThreads leaves out thread replies.
func (c *client) channelPostsBefore(ctx context.Context, channelID, beforeID string, n int) ([]post, error) {
	q := url.Values{"before": {beforeID}, "per_page": {strconv.Itoa(n)}, "collapsedThreads": {"true"}}
	var l postList
	err := c.do(ctx, http.MethodGet, "/channels/"+url.PathEscape(channelID)+"/posts?"+q.Encode(), nil, &l)
	return l.oldestFirst(), err
}

func (c *client) thread(ctx context.Context, rootID string) ([]post, error) {
	var l postList
	err := c.do(ctx, http.MethodGet, "/posts/"+url.PathEscape(rootID)+"/thread", nil, &l)
	return l.oldestFirst(), err
}

func (c *client) websocketURL() string {
	return strings.Replace(c.serverURL, "http", "ws", 1) + "/api/v4/websocket"
}
