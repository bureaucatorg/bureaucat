package mattermostbot

import (
	"slices"
	"testing"
)

func TestUnseen(t *testing.T) {
	const bot, me, other = "bot", "me", "other"
	mention := mentionRe("bureaucat")
	at := func(id, user, root, msg string, ts int64) post {
		return post{ID: id, UserID: user, RootID: root, Message: msg, CreateAt: ts}
	}
	ids := func(ps []post) []string {
		out := []string{}
		for _, p := range ps {
			out = append(out, p.ID)
		}
		return out
	}

	channel := []post{
		at("c1", other, "", "old news", 1),
		at("c2", other, "c1", "a reply", 2),
		{ID: "c3", UserID: other, Type: "system_join_channel", Message: "joined", CreateAt: 3},
		at("c4", other, "", "  ", 4),
		at("c5", other, "", "shown before", 5),
		at("c6", me, "", "context", 6),
	}
	top := at("p", me, "", "@bureaucat help", 7)

	thread := []post{
		at("r", other, "", "root question", 1),
		at("t1", bot, "r", "bot answering someone else", 2),
		at("t2", other, "r", "more detail", 3),
		at("u1", me, "r", "@bureaucat first turn", 4),
		at("t3", other, "r", "posted while bot worked", 5),
		at("b1", bot, "r", "bot reply", 6),
		at("t4", other, "r", "after reply", 7),
		at("t5", me, "r", "no mention, just context", 8),
	}
	turn2 := at("u2", me, "r", "@bureaucat second turn", 9)

	tests := []struct {
		name    string
		posts   []post
		p       post
		shown   []string
		newOnly bool
		want    []string
	}{
		{"channel skips replies, system, empty and shown", channel, top, []string{"c5"}, true, []string{"c1", "c6"}},
		{"first turn in thread keeps earlier bot posts", thread[:3], at("u1", me, "r", "@bureaucat first turn", 4), nil, true, []string{"r", "t1", "t2"}},
		{"second turn only sees posts since previous turn", thread, turn2, nil, true, []string{"t3", "t4", "t5"}},
		{"history skips turns, bot replies and shown posts", thread, turn2, []string{"t4", "t5", "t2"}, false, []string{"r", "t1", "t3"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shown := map[string]bool{}
			for _, id := range tt.shown {
				shown[id] = true
			}
			got := ids(unseen(tt.posts, tt.p, bot, mention, shown, tt.newOnly))
			if !slices.Equal(got, tt.want) {
				t.Errorf("unseen = %v, want %v", got, tt.want)
			}
		})
	}
}
