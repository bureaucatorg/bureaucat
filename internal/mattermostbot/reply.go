package mattermostbot

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"bereaucat/internal/assistant"
	"bereaucat/internal/store"
)

const maxPostRunes = 16000

type replier struct {
	ctx          context.Context
	client       *client
	channelID    string
	rootID       string
	appURL       string
	answerPrefix string // mention users must include to answer outside DMs
}

func (r *replier) post(msg string) {
	if runes := []rune(msg); len(runes) > maxPostRunes {
		msg = string(runes[:maxPostRunes]) + "\n\n…(truncated)"
	}
	if err := r.client.createPost(r.ctx, r.channelID, r.rootID, msg); err != nil {
		log.Printf("mattermost assistant: post reply: %v", err)
	}
}

const (
	reactionWorking = "large_yellow_circle"
	reactionDone    = "large_green_circle"
)

func (r *replier) react(botID, postID, emoji string) {
	if err := r.client.addReaction(r.ctx, botID, postID, emoji); err != nil {
		log.Printf("mattermost assistant: add reaction: %v", err)
	}
}

func (r *replier) unreact(botID, postID, emoji string) {
	if err := r.client.removeReaction(r.ctx, botID, postID, emoji); err != nil {
		log.Printf("mattermost assistant: remove reaction: %v", err)
	}
}

// typing shows "bot is typing" in the thread until the returned func is called.
func (r *replier) typing(botID string) func() {
	ctx, cancel := context.WithCancel(r.ctx)
	go func() {
		t := time.NewTicker(4 * time.Second)
		defer t.Stop()
		for {
			_ = r.client.typing(ctx, botID, r.channelID, r.rootID)
			select {
			case <-ctx.Done():
				return
			case <-t.C:
			}
		}
	}()
	return cancel
}

// emitter posts assistant replies as they arrive and batches approval requests into one post.
type emitter struct {
	r       *replier
	users   store.Querier
	pending []*assistant.ActionView
}

func (e *emitter) Emit(ev assistant.Event) {
	switch ev.Type {
	case "message":
		if strings.TrimSpace(ev.Text) != "" {
			e.r.post(sanitize(ev.Text, e.r.appURL))
		}
	case "pending":
		e.pending = append(e.pending, ev.Action)
	case "action":
		e.r.post(actionLine(ev.Action))
	}
}

func (e *emitter) flush() {
	if len(e.pending) == 0 {
		return
	}
	var b strings.Builder
	b.WriteString("**Approval needed**\n")
	for _, a := range e.pending {
		fmt.Fprintf(&b, "\n**%s**\n%s", a.Title, argsList(a.Args, e.username))
	}
	if len(e.pending) > 1 {
		fmt.Fprintf(&b, "\nReply `%[1]syes` to approve all of these or `%[1]sno` to reject them, in this thread.", e.r.answerPrefix)
	} else {
		fmt.Fprintf(&b, "\nReply `%[1]syes` to approve or `%[1]sno` to reject, in this thread.", e.r.answerPrefix)
	}
	e.r.post(b.String())
	e.pending = nil
}

func actionLine(a *assistant.ActionView) string {
	switch a.Status {
	case "approved":
		return "✅ Approved: " + a.Title
	case "rejected":
		return "❌ Rejected: " + a.Title
	default:
		return fmt.Sprintf("⚠️ Failed: %s %s", a.Title, code(a.Result, 300))
	}
}

var userArgs = map[string]bool{"user_id": true, "assignees": true, "lead_id": true, "member_ids": true}

// username turns a user UUID into @username, leaving anything it cannot resolve as is.
func (e *emitter) username(id string) string {
	uid, err := uuid.Parse(id)
	if err != nil || e.users == nil {
		return id
	}
	u, err := e.users.GetUserByID(e.r.ctx, uid)
	if err != nil {
		return id
	}
	return "@" + u.Username
}

func argsList(raw json.RawMessage, username func(string) string) string {
	var args map[string]any
	if json.Unmarshal(raw, &args) != nil || len(args) == 0 {
		return ""
	}
	keys := make([]string, 0, len(args))
	for k := range args {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for _, k := range keys {
		if userArgs[k] {
			args[k] = resolveUsers(args[k], username)
		}
		v, _ := json.Marshal(args[k])
		fmt.Fprintf(&b, "- %s: %s\n", k, code(strings.Trim(string(v), `"`), 200))
	}
	return b.String()
}

func resolveUsers(v any, username func(string) string) any {
	switch t := v.(type) {
	case string:
		return username(t)
	case []any:
		for i, x := range t {
			if s, ok := x.(string); ok {
				t[i] = username(s)
			}
		}
	}
	return v
}

// code renders untrusted text as inline code so Mattermost shows no links, mentions or formatting.
func code(s string, limit int) string {
	s = strings.Join(strings.Fields(strings.ReplaceAll(s, "`", "'")), " ")
	if r := []rune(s); len(r) > limit {
		s = string(r[:limit]) + "…"
	}
	if s == "" {
		return ""
	}
	return "`" + s + "`"
}

var (
	linkRe    = regexp.MustCompile(`(!?)\[([^\]]*)\]\(\s*<?([^)\s>]*)>?(?:\s+"[^"]*")?\s*\)`)
	bareURLRe = regexp.MustCompile(`(?i)\b(?:[a-z][a-z0-9+.-]*://|www\.)[^\s<>()\[\]` + "`" + `]+`)
	massRe    = regexp.MustCompile(`(?i)@(?:channel|all|here)\b`)
)

// sanitize keeps only links into Bureaucat, so Mattermost never previews or fetches other URLs
// a prompt injection could plant; images become their alt text and @channel-style pings are defused.
func sanitize(text, appURL string) string {
	base := strings.TrimRight(appURL, "/")
	inApp := func(u string) bool { return u == base || strings.HasPrefix(u, base+"/") }

	var links []string
	text = linkRe.ReplaceAllStringFunc(text, func(m string) string {
		g := linkRe.FindStringSubmatch(m)
		label, target := g[2], g[3]
		if strings.HasPrefix(target, "/") && !strings.HasPrefix(target, "//") {
			target = base + target
		}
		if g[1] != "" || !inApp(target) {
			return label
		}
		links = append(links, fmt.Sprintf("[%s](%s)", label, target))
		return fmt.Sprintf("\x00%d\x00", len(links)-1)
	})
	text = defuseURLs(text, inApp)
	for i := range links {
		links[i] = defuseMentions(links[i])
	}
	text = defuseMentions(text)
	for i, l := range links {
		text = strings.Replace(text, fmt.Sprintf("\x00%d\x00", i), l, 1)
	}
	return separateTables(text)
}

var tableDelimRe = regexp.MustCompile(`^\s*\|?\s*:?-+:?\s*(\|\s*:?-+:?\s*)+\|?\s*$`)

// separateTables adds the blank line Mattermost needs before a table that directly follows text.
func separateTables(text string) string {
	lines := strings.Split(text, "\n")
	out := make([]string, 0, len(lines))
	for i, line := range lines {
		isHeader := i+1 < len(lines) && strings.Contains(line, "|") && tableDelimRe.MatchString(lines[i+1])
		if isHeader && len(out) > 0 && strings.TrimSpace(out[len(out)-1]) != "" {
			out = append(out, "")
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}

func defuseMentions(s string) string {
	return massRe.ReplaceAllStringFunc(s, func(m string) string { return "`" + m + "`" })
}

func defuseURLs(s string, inApp func(string) bool) string {
	return bareURLRe.ReplaceAllStringFunc(s, func(u string) string {
		if inApp(u) {
			return u
		}
		return "`" + u + "`"
	})
}
