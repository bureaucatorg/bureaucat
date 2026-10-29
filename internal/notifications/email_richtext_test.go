package notifications

import (
	"strings"
	"testing"
)

func TestRichExcerpt(t *testing.T) {
	tests := []struct {
		name, in string
		limit    int
		want     []string
		notWant  []string
	}{
		{
			name: "keeps formatting and word spacing",
			in:   `<p>open port <strong>6379</strong> now</p><p></p><p>Reason: <s>old</s> testing</p>`,
			want: []string{`open port <b style="color:#09090b">6379</b> now`, `<s>old</s>`, `padding-top:8px`},
		},
		{
			name: "lists and mentions",
			in:   `<ul><li><p>ask <a href="/profile/abc">@Sam Sample</a></p></li></ul>`,
			want: []string{"<ul ", "<li ", `<b style="color:#b45309">@Sam Sample</b>`},
		},
		{
			name:    "checklists",
			in:      `<ul data-type="taskList"><li data-checked="true" data-type="taskItem"><label><input type="checkbox" checked="checked"><span></span></label><div><p>done</p></div></li><li data-checked="false" data-type="taskItem"><label><input type="checkbox"><span></span></label><div><p>todo</p></div></li></ul>`,
			want:    []string{"list-style:none", "&#9745; <span>", "&#9744; <span>", "done", "todo"},
			notWant: []string{"<input"},
		},
		{
			name:    "drops unsafe markup",
			in:      `<p onclick="x()"><script>alert(1)</script><a href="javascript:alert(1)">click</a><img src=x onerror=alert(1)></p>`,
			want:    []string{"alert(1)click [image]"},
			notWant: []string{"<script", "javascript:", "onclick", "onerror", "<img"},
		},
		{
			name:    "escapes entities in text",
			in:      `<p>a &lt;b&gt; &amp; c</p>`,
			want:    []string{"a &lt;b&gt; &amp; c"},
			notWant: []string{"<b>"},
		},
		{
			name:    "truncates and closes open tags",
			in:      `<ul><li><p>` + strings.Repeat("x", 50) + `</p></li><li>more</li></ul>`,
			limit:   10,
			want:    []string{"xxxxxxxxxx…</span></li></ul>"},
			notWant: []string{"more"},
		},
		{
			name: "legacy plain text keeps line breaks",
			in:   "line one\nline <two>",
			want: []string{"line one<br>line &lt;two&gt;"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limit := tt.limit
			if limit == 0 {
				limit = quoteRunes
			}
			got := string(richExcerpt(tt.in, limit))
			for _, w := range tt.want {
				if !strings.Contains(got, w) {
					t.Errorf("want %q in %q", w, got)
				}
			}
			for _, w := range tt.notWant {
				if strings.Contains(got, w) {
					t.Errorf("did not want %q in %q", w, got)
				}
			}
		})
	}
}
