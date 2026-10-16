// Package mattermostbot lets users talk to the assistant through the Mattermost notification bot.
package mattermostbot

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"regexp"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"bereaucat/internal/assistant"
	"bereaucat/internal/store"
)

const (
	configPoll = 30 * time.Second
	userTTL    = 5 * time.Minute
	runTimeout = 10 * time.Minute

	contextMessages = 10
)

// config mirrors the "mattermost" setting saved by the admin settings handler.
type config struct {
	Enabled          bool   `json:"enabled"`
	ServerURL        string `json:"server_url"`
	BotToken         string `json:"bot_token"`
	AssistantEnabled bool   `json:"assistant_enabled"`
	AppURL           string `json:"app_url"`
}

func (c config) active() bool {
	return c.Enabled && c.AssistantEnabled && c.ServerURL != "" && c.BotToken != "" && c.AppURL != ""
}

type Bot struct {
	store store.Querier
	agent *assistant.Service

	mu    sync.Mutex
	users map[string]cachedUser
}

type cachedUser struct {
	user mmUser
	at   time.Time
}

func New(s store.Querier, agent *assistant.Service) *Bot {
	return &Bot{store: s, agent: agent, users: map[string]cachedUser{}}
}

// Run keeps a WebSocket connection to Mattermost open while the integration is enabled.
func (b *Bot) Run(ctx context.Context) {
	backoff := time.Second
	for {
		wait := configPoll
		cfg, err := b.loadConfig(ctx)
		if err != nil {
			log.Printf("mattermost assistant: load settings: %v", err)
		} else if cfg.active() {
			started := time.Now()
			err := b.listen(ctx, cfg)
			if ctx.Err() != nil {
				return
			}
			if time.Since(started) > time.Minute {
				backoff = time.Second
			}
			if err == nil {
				wait = 0
			} else {
				log.Printf("mattermost assistant: %v", err)
				wait, backoff = backoff, min(backoff*2, time.Minute)
			}
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(wait):
		}
	}
}

func (b *Bot) loadConfig(ctx context.Context) (config, error) {
	var cfg config
	setting, err := b.store.GetSetting(ctx, "mattermost")
	if errors.Is(err, pgx.ErrNoRows) {
		return cfg, nil
	}
	if err != nil {
		return cfg, err
	}
	return cfg, json.Unmarshal(setting.Value, &cfg)
}

// listen returns nil when the settings change, so Run reconnects with the new ones.
func (b *Bot) listen(parent context.Context, cfg config) error {
	c := newClient(cfg)
	me, err := c.user(parent, "me")
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(parent)
	defer cancel()
	conn, _, err := websocket.Dial(ctx, c.websocketURL(), &websocket.DialOptions{
		HTTPHeader: map[string][]string{"Authorization": {"Bearer " + cfg.BotToken}},
	})
	if err != nil {
		return err
	}
	defer conn.CloseNow()
	conn.SetReadLimit(1 << 20)
	go b.watchConfig(ctx, cfg, cancel)
	log.Printf("mattermost assistant: connected as @%s", me.Username)

	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			if parent.Err() == nil && ctx.Err() != nil {
				return nil
			}
			return err
		}
		b.dispatch(parent, c, cfg, me, data)
	}
}

func (b *Bot) watchConfig(ctx context.Context, cfg config, changed context.CancelFunc) {
	t := time.NewTicker(configPoll)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			if next, err := b.loadConfig(ctx); err == nil && next != cfg {
				changed()
				return
			}
		}
	}
}

type wsEvent struct {
	Event string `json:"event"`
	Data  struct {
		ChannelType string `json:"channel_type"`
		Post        string `json:"post"`
		Mentions    string `json:"mentions"`
	} `json:"data"`
}

type post struct {
	ID        string `json:"id"`
	RootID    string `json:"root_id"`
	ChannelID string `json:"channel_id"`
	UserID    string `json:"user_id"`
	Message   string `json:"message"`
	Type      string `json:"type"`
	CreateAt  int64  `json:"create_at"`
}

func (b *Bot) dispatch(ctx context.Context, c *client, cfg config, me mmUser, data []byte) {
	var ev wsEvent
	if json.Unmarshal(data, &ev) != nil || ev.Event != "posted" {
		return
	}
	var p post
	if json.Unmarshal([]byte(ev.Data.Post), &p) != nil || p.UserID == me.ID || p.Type != "" {
		return
	}
	var mentions []string
	_ = json.Unmarshal([]byte(ev.Data.Mentions), &mentions)
	if ev.Data.ChannelType != "D" && !slices.Contains(mentions, me.ID) {
		return
	}
	go b.handle(ctx, c, cfg, me, p, ev.Data.ChannelType == "D")
}

func (b *Bot) handle(parent context.Context, c *client, cfg config, me mmUser, p post, direct bool) {
	ctx, cancel := context.WithTimeout(parent, runTimeout)
	defer cancel()
	root := p.RootID
	if root == "" {
		root = p.ID
	}
	r := &replier{ctx: ctx, client: c, channelID: p.ChannelID, rootID: root, appURL: cfg.AppURL}
	if !direct {
		r.answerPrefix = "@" + me.Username + " "
	}

	sender, err := b.mmUser(ctx, c, p.UserID)
	if err != nil {
		log.Printf("mattermost assistant: lookup user %s: %v", p.UserID, err)
		return
	}
	if sender.IsBot {
		return
	}
	userID, problem, err := b.bureaucatUser(ctx, sender.Email)
	if err != nil {
		log.Printf("mattermost assistant: match email: %v", err)
		return
	}
	if problem != "" {
		r.post(problem)
		return
	}

	convID, err := b.store.GetAssistantMattermostThread(ctx, store.GetAssistantMattermostThreadParams{RootPostID: root, UserID: userID})
	found := err == nil
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		log.Printf("mattermost assistant: thread lookup: %v", err)
		return
	}

	text := strings.TrimSpace(mentionRe(me.Username).ReplaceAllString(p.Message, ""))
	if text == "" {
		r.post("Ask me something about your Bureaucat tasks and projects.")
		return
	}
	if !found {
		conv, err := b.agent.CreateConversation(ctx, userID)
		if err == nil {
			convID = conv.ID
			err = b.store.CreateAssistantMattermostThread(ctx, store.CreateAssistantMattermostThreadParams{RootPostID: root, UserID: userID, ConversationID: convID})
		}
		if err != nil {
			log.Printf("mattermost assistant: create conversation: %v", err)
			r.post(assistant.ErrorText(err))
			return
		}
	}

	r.react(me.ID, p.ID, reactionWorking)
	stopTyping := r.typing(me.ID)
	em := &emitter{r: r}
	err = b.run(ctx, c, me, p, direct, userID, convID, text, em)
	stopTyping()
	em.flush()
	r.unreact(me.ID, p.ID, reactionWorking)
	r.react(me.ID, p.ID, reactionDone)
	if errors.Is(err, assistant.ErrBusy) {
		r.post("Still working on your previous message in this thread.")
	} else if err != nil {
		log.Printf("mattermost assistant: run: %v", err)
		r.post(assistant.ErrorText(err))
	}
}

// run answers pending approvals with a yes/no reply; anything else is a new message,
// sent with recent channel or thread messages when outside a DM.
func (b *Bot) run(ctx context.Context, c *client, me mmUser, p post, direct bool, userID, convID uuid.UUID, text string, em *emitter) error {
	pending, err := b.store.ListPendingAssistantActions(ctx, convID)
	if err != nil {
		return err
	}
	approve, isDecision := decision(text)
	if len(pending) == 0 || !isDecision {
		var meta assistant.Meta
		if !direct {
			if meta.Chat, err = b.chatContext(ctx, c, me, p); err != nil {
				log.Printf("mattermost assistant: load chat context: %v", err)
			}
		}
		return b.agent.Send(ctx, userID, convID, text, nil, meta, em)
	}
	for _, a := range pending {
		err := b.agent.Resolve(ctx, userID, convID, a.ID, approve, em)
		if err != nil && !errors.Is(err, assistant.ErrNotPending) {
			return err
		}
	}
	return nil
}

func decision(text string) (approve, ok bool) {
	switch strings.ToLower(strings.Trim(text, " .!")) {
	case "yes", "y", "approve":
		return true, true
	case "no", "n", "reject":
		return false, true
	}
	return false, false
}

func mentionRe(username string) *regexp.Regexp {
	return regexp.MustCompile(`(?i)@` + regexp.QuoteMeta(username) + `\b`)
}

func (b *Bot) mmUser(ctx context.Context, c *client, id string) (mmUser, error) {
	b.mu.Lock()
	cached, ok := b.users[id]
	b.mu.Unlock()
	if ok && time.Since(cached.at) < userTTL {
		return cached.user, nil
	}
	u, err := c.user(ctx, id)
	if err != nil {
		return u, err
	}
	b.mu.Lock()
	b.users[id] = cachedUser{user: u, at: time.Now()}
	b.mu.Unlock()
	return u, nil
}

// chatContext returns the last few top-level channel posts, or thread posts, before p.
func (b *Bot) chatContext(ctx context.Context, c *client, me mmUser, p post) (*assistant.ChatContext, error) {
	where := "thread"
	var posts []post
	var err error
	if p.RootID == "" {
		where = "channel"
		posts, err = c.channelPostsBefore(ctx, p.ChannelID, p.ID, 2*contextMessages)
	} else {
		posts, err = c.thread(ctx, p.RootID)
	}
	if err != nil {
		return nil, err
	}

	var keep []post
	for _, q := range posts {
		if q.ID == p.ID || q.Type != "" || q.CreateAt >= p.CreateAt || strings.TrimSpace(q.Message) == "" {
			continue
		}
		if where == "channel" && q.RootID != "" {
			continue
		}
		keep = append(keep, q)
	}
	if len(keep) > contextMessages {
		keep = keep[len(keep)-contextMessages:]
	}
	if len(keep) == 0 {
		return nil, nil
	}

	authors := map[string]assistant.ChatMessage{}
	out := &assistant.ChatContext{Where: where}
	for _, q := range keep {
		m, ok := authors[q.UserID]
		if !ok {
			m = b.author(ctx, c, me, q.UserID)
			authors[q.UserID] = m
		}
		m.Text = q.Message
		out.Messages = append(out.Messages, m)
	}
	return out, nil
}

func (b *Bot) author(ctx context.Context, c *client, me mmUser, id string) assistant.ChatMessage {
	if id == me.ID {
		return assistant.ChatMessage{Author: "@" + me.Username, IsAssistant: true}
	}
	u, err := b.mmUser(ctx, c, id)
	if err != nil {
		return assistant.ChatMessage{Author: "unknown user"}
	}
	m := assistant.ChatMessage{Author: "@" + u.Username}
	if u.Email != "" {
		if rows, err := b.store.ListUsersByEmailInsensitive(ctx, u.Email); err == nil && len(rows) == 1 {
			m.BureaucatUser = rows[0].Username
		}
	}
	return m
}

// bureaucatUser matches a Mattermost email to exactly one Bureaucat user, or explains why not.
func (b *Bot) bureaucatUser(ctx context.Context, email string) (uuid.UUID, string, error) {
	if email == "" {
		return uuid.Nil, "I can't see your Mattermost email address. Ask an admin to let the bot read user emails.", nil
	}
	rows, err := b.store.ListUsersByEmailInsensitive(ctx, email)
	if err != nil {
		return uuid.Nil, "", err
	}
	switch len(rows) {
	case 0:
		return uuid.Nil, "No Bureaucat account matches your Mattermost email.", nil
	case 1:
		return rows[0].ID, "", nil
	default:
		return uuid.Nil, "More than one Bureaucat account matches your Mattermost email. Ask an admin to fix it.", nil
	}
}
