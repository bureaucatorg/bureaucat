package notifications

import (
	"html"
	"html/template"
	"regexp"
	"strings"
	"unicode/utf8"
)

var (
	htmlTokenRe = regexp.MustCompile(`<(/?)([a-zA-Z][a-zA-Z0-9]*)([^>]*)>`)
	hrefRe      = regexp.MustCompile(`(?i)\bhref\s*=\s*(?:"([^"]*)"|'([^']*)')`)
	spaceRe     = regexp.MustCompile(`\s+`)
)

type richOpen struct {
	name, closer string
}

// richWriter rebuilds user HTML from an allowlist into inline-styled, email-safe markup.
// Input is never passed through: text is re-escaped and only tags built here are emitted.
type richWriter struct {
	b         strings.Builder
	stack     []richOpen
	runes     int
	limit     int
	truncated bool
	blocks    int
	inPre     int
	inList    int
}

func (w *richWriter) blockStyle() string {
	if w.blocks == 0 || w.inList > 0 {
		return ""
	}
	return "padding-top:8px;"
}

func (w *richWriter) open(name, tag, closer string) {
	w.b.WriteString(tag)
	w.stack = append(w.stack, richOpen{name, closer})
}

func (w *richWriter) openBlock(name, style string) {
	w.open(name, `<div style="`+w.blockStyle()+style+`">`, "</div>")
	w.blocks++
}

func (w *richWriter) close(name string) {
	for i := len(w.stack) - 1; i >= 0; i-- {
		if w.stack[i].name != name {
			continue
		}
		for j := len(w.stack) - 1; j >= i; j-- {
			w.b.WriteString(w.stack[j].closer)
			switch w.stack[j].name {
			case "pre":
				w.inPre--
			case "ul", "ol":
				w.inList--
			}
		}
		w.stack = w.stack[:i]
		return
	}
}

func (w *richWriter) text(s string) {
	if w.truncated || s == "" {
		return
	}
	if w.inPre == 0 {
		s = spaceRe.ReplaceAllString(s, " ")
	}
	if n := utf8.RuneCountInString(s); w.runes+n > w.limit {
		s = strings.TrimSpace(string([]rune(s)[:w.limit-w.runes])) + "…"
		w.truncated = true
	}
	w.runes += utf8.RuneCountInString(s)
	w.b.WriteString(template.HTMLEscapeString(s))
}

func (w *richWriter) startTag(name, attrs string) {
	const mono = "font-family:" + fontMono + ";"
	switch name {
	case "p", "div":
		if w.inList > 0 {
			w.open(name, "<span>", "</span>")
			return
		}
		w.openBlock(name, "")
	case "h1", "h2", "h3", "h4", "h5", "h6":
		w.openBlock(name, "font-weight:700;color:#09090b;")
	case "blockquote":
		w.openBlock(name, "border-left:3px solid #d4d4d8;padding-left:10px;color:#52525b;")
	case "pre":
		w.openBlock(name, mono+"font-size:12px;line-height:18px;background-color:#f4f4f5;padding:8px 10px;border-radius:6px;white-space:pre-wrap;")
		w.inPre++
	case "ul", "ol":
		w.open(name, `<`+name+` style="`+w.blockStyle()+`margin:0;padding-left:22px">`, "</"+name+">")
		w.blocks++
		w.inList++
	case "li":
		w.open(name, `<li style="padding-top:2px">`, "</li>")
	case "strong", "b":
		w.open(name, `<b style="color:#09090b">`, "</b>")
	case "em", "i":
		w.open(name, "<i>", "</i>")
	case "s", "del", "strike":
		w.open(name, "<s>", "</s>")
	case "u":
		w.open(name, "<u>", "</u>")
	case "code":
		if w.inPre > 0 {
			w.open(name, "", "")
			return
		}
		w.open(name, `<code style="`+mono+`font-size:12px;background-color:#f4f4f5;padding:1px 4px;border-radius:4px">`, "</code>")
	case "a":
		w.openLink(attrs)
	case "br":
		w.b.WriteString("<br>")
	case "hr":
		w.b.WriteString(`<div style="padding-top:8px;border-bottom:1px solid #e4e4e7;font-size:0;line-height:0">&nbsp;</div>`)
	case "img":
		w.text(" [image] ")
	}
}

// Mentions link to /profile/<id>; other relative links are dropped since emails have no base URL.
func (w *richWriter) openLink(attrs string) {
	var href string
	if m := hrefRe.FindStringSubmatch(attrs); m != nil {
		href = html.UnescapeString(m[1] + m[2])
	}
	lower := strings.ToLower(href)
	switch {
	case strings.HasPrefix(href, "/profile/"):
		w.open("a", `<b style="color:#b45309">`, "</b>")
	case strings.HasPrefix(lower, "https://"), strings.HasPrefix(lower, "http://"):
		w.open("a", `<a href="`+template.HTMLEscapeString(href)+`" style="color:#b45309;text-decoration:underline">`, "</a>")
	default:
		w.open("a", "", "")
	}
}

// richExcerpt renders stored rich text (Tiptap HTML or legacy plain text) for an email quote, capped at limit runes.
func richExcerpt(s string, limit int) template.HTML {
	w := &richWriter{limit: limit}
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "<") {
		for i, line := range strings.Split(s, "\n") {
			if w.truncated {
				break
			}
			if i > 0 {
				w.b.WriteString("<br>")
			}
			w.text(line)
		}
		return template.HTML(w.b.String())
	}

	last := 0
	for _, m := range htmlTokenRe.FindAllStringSubmatchIndex(s, -1) {
		w.text(html.UnescapeString(s[last:m[0]]))
		last = m[1]
		if w.truncated {
			break
		}
		name := strings.ToLower(s[m[4]:m[5]])
		if s[m[2]:m[3]] == "/" {
			w.close(name)
		} else {
			w.startTag(name, s[m[6]:m[7]])
		}
	}
	if !w.truncated {
		w.text(html.UnescapeString(s[last:]))
	}
	for len(w.stack) > 0 {
		w.close(w.stack[len(w.stack)-1].name)
	}
	return template.HTML(w.b.String())
}
