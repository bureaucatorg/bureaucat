package planeimport

import (
	"fmt"
	"regexp"
	"strings"
)

// htmlToMarkdown converts Plane's ProseMirror HTML output to Markdown.
// Handles: p, a, strong/b, em/i, u, code, pre>code, ul/ol/li,
// blockquote, h1-h6, br, img, mention-component, tables.
func htmlToMarkdown(html string) string {
	if html == "" {
		return ""
	}

	s := html

	// Normalize escaped newlines from pg_dump (literal \n in COPY data).
	s = strings.ReplaceAll(s, "\\n", "\n")

	// Pre-process: protect code blocks by extracting them first.
	s = convertCodeBlocks(s)

	// Mentions: <mention-component ... label="name" ...>text</mention-component> → @name
	s = regReplace(s, `<mention-component[^>]*\blabel="([^"]*)"[^>]*>.*?</mention-component>`, "@$1")

	// Images: <img ... src="url" ... alt="text" /> → ![text](url)
	s = regReplace(s, `<img[^>]*\bsrc="([^"]*)"[^>]*\balt="([^"]*)"[^>]*\/?>`, "![$2]($1)")
	s = regReplace(s, `<img[^>]*\bsrc="([^"]*)"[^>]*\/?>`, "![]($1)")
	// <image> tags (Plane custom)
	s = regReplace(s, `<image[^>]*\bsrc="([^"]*)"[^>]*\balt="([^"]*)"[^>]*\/?>`, "![$2]($1)")
	s = regReplace(s, `<image[^>]*\bsrc="([^"]*)"[^>]*\/?>`, "![]($1)")
	s = regReplace(s, `</image>`, "")

	// Links: <a href="url">text</a> → [text](url)
	s = regReplace(s, `<a[^>]*\bhref="([^"]*)"[^>]*>(.*?)</a>`, "[$2]($1)")

	// Bold: <strong>text</strong> or <b>text</b>
	s = regReplace(s, `<(?:strong|b)>(.*?)</(?:strong|b)>`, "**$1**")

	// Italic: <em>text</em> or <i>text</i>
	s = regReplace(s, `<(?:em|i)>(.*?)</(?:em|i)>`, "*$1*")

	// Underline: not standard markdown, use emphasis
	s = regReplace(s, `<u>(.*?)</u>`, "_${1}_")

	// Inline code: <code>text</code> (not inside pre)
	s = regReplace(s, `<code>(.*?)</code>`, "`$1`")

	// Headings
	s = regReplace(s, `<h1[^>]*>(.*?)</h1>`, "\n# $1\n")
	s = regReplace(s, `<h2[^>]*>(.*?)</h2>`, "\n## $1\n")
	s = regReplace(s, `<h3[^>]*>(.*?)</h3>`, "\n### $1\n")
	s = regReplace(s, `<h4[^>]*>(.*?)</h4>`, "\n#### $1\n")
	s = regReplace(s, `<h5[^>]*>(.*?)</h5>`, "\n##### $1\n")
	s = regReplace(s, `<h6[^>]*>(.*?)</h6>`, "\n###### $1\n")

	// Blockquotes
	s = convertBlockquotes(s)

	// Tables
	s = convertTables(s)

	// Lists - must be done before removing other tags.
	s = convertLists(s)

	// Task list checkboxes: <input type="checkbox" checked /> → [x], <input type="checkbox" /> → [ ]
	s = regReplace(s, `<input[^>]*\bchecked[^>]*\/?>`, "[x] ")
	s = regReplace(s, `<input[^>]*type="checkbox"[^>]*\/?>`, "[ ] ")

	// Line breaks
	s = regReplace(s, `<br\s*/?>`, "\n")

	// Paragraphs: <p>text</p> → text + double newline
	s = regReplace(s, `<p[^>]*>(.*?)</p>`, "$1\n\n")

	// Divs
	s = regReplace(s, `<div[^>]*>(.*?)</div>`, "$1\n")

	// Strip remaining HTML tags.
	s = regReplace(s, `<[^>]+>`, "")

	// Decode common HTML entities.
	s = strings.ReplaceAll(s, "&amp;", "&")
	s = strings.ReplaceAll(s, "&lt;", "<")
	s = strings.ReplaceAll(s, "&gt;", ">")
	s = strings.ReplaceAll(s, "&quot;", "\"")
	s = strings.ReplaceAll(s, "&#39;", "'")
	s = strings.ReplaceAll(s, "&nbsp;", " ")

	// Clean up excessive blank lines.
	s = regReplace(s, `\n{3,}`, "\n\n")
	s = strings.TrimSpace(s)

	return s
}

// convertCodeBlocks handles <pre><code>...</code></pre> → fenced code blocks.
func convertCodeBlocks(s string) string {
	re := regexp.MustCompile(`(?s)<pre[^>]*>\s*<code[^>]*(?:data-code-block-language="([^"]*)")?[^>]*>(.*?)</code>\s*</pre>`)
	return re.ReplaceAllStringFunc(s, func(match string) string {
		parts := re.FindStringSubmatch(match)
		lang := ""
		code := parts[2]
		if parts[1] != "" {
			lang = parts[1]
			// "markup" is not a real language
			if lang == "markup" {
				lang = ""
			}
		}

		// Decode HTML entities inside code.
		code = strings.ReplaceAll(code, "&amp;", "&")
		code = strings.ReplaceAll(code, "&lt;", "<")
		code = strings.ReplaceAll(code, "&gt;", ">")
		code = strings.ReplaceAll(code, "&quot;", "\"")
		code = strings.ReplaceAll(code, "&#39;", "'")
		code = strings.ReplaceAll(code, "&nbsp;", " ")

		// Handle literal \n from pg_dump COPY data.
		code = strings.ReplaceAll(code, "\\n", "\n")

		return "\n```" + lang + "\n" + code + "\n```\n"
	})
}

// convertBlockquotes handles <blockquote>...</blockquote>.
func convertBlockquotes(s string) string {
	re := regexp.MustCompile(`(?s)<blockquote[^>]*>(.*?)</blockquote>`)
	return re.ReplaceAllStringFunc(s, func(match string) string {
		parts := re.FindStringSubmatch(match)
		inner := parts[1]
		// Strip inner <p> tags.
		inner = regReplace(inner, `<p[^>]*>(.*?)</p>`, "$1")
		inner = strings.TrimSpace(inner)
		lines := strings.Split(inner, "\n")
		var quoted []string
		for _, line := range lines {
			quoted = append(quoted, "> "+strings.TrimSpace(line))
		}
		return "\n" + strings.Join(quoted, "\n") + "\n"
	})
}

// convertLists handles <ul>/<ol> with <li> items.
func convertLists(s string) string {
	// Process nested lists by repeatedly converting innermost lists first.
	for i := 0; i < 5; i++ { // max 5 levels of nesting
		prev := s
		s = convertListPass(s, "ul", "- ")
		s = convertOlPass(s)
		if s == prev {
			break
		}
	}
	return s
}

func convertListPass(s, tag, prefix string) string {
	reList := regexp.MustCompile(`(?s)<` + tag + `[^>]*>(.*?)</` + tag + `>`)
	return reList.ReplaceAllStringFunc(s, func(match string) string {
		parts := reList.FindStringSubmatch(match)
		inner := parts[1]
		reLi := regexp.MustCompile(`(?s)<li[^>]*>(.*?)</li>`)
		items := reLi.FindAllStringSubmatch(inner, -1)
		var lines []string
		for _, item := range items {
			content := strings.TrimSpace(item[1])
			// Strip <p> tags inside <li>.
			content = regReplace(content, `<p[^>]*>(.*?)</p>`, "$1")
			content = strings.TrimSpace(content)
			lines = append(lines, prefix+content)
		}
		return "\n" + strings.Join(lines, "\n") + "\n"
	})
}

func convertOlPass(s string) string {
	reList := regexp.MustCompile(`(?s)<ol[^>]*>(.*?)</ol>`)
	return reList.ReplaceAllStringFunc(s, func(match string) string {
		parts := reList.FindStringSubmatch(match)
		inner := parts[1]
		reLi := regexp.MustCompile(`(?s)<li[^>]*>(.*?)</li>`)
		items := reLi.FindAllStringSubmatch(inner, -1)
		var lines []string
		for i, item := range items {
			content := strings.TrimSpace(item[1])
			content = regReplace(content, `<p[^>]*>(.*?)</p>`, "$1")
			content = strings.TrimSpace(content)
			lines = append(lines, fmt.Sprintf("%d. %s", i+1, content))
		}
		return "\n" + strings.Join(lines, "\n") + "\n"
	})
}

// convertTables handles <table> with <tr>/<td>/<th>.
func convertTables(s string) string {
	reTable := regexp.MustCompile(`(?s)<table[^>]*>(.*?)</table>`)
	return reTable.ReplaceAllStringFunc(s, func(match string) string {
		parts := reTable.FindStringSubmatch(match)
		inner := parts[1]

		// Strip thead/tbody wrappers.
		inner = regReplace(inner, `</?(?:thead|tbody)[^>]*>`, "")

		reTr := regexp.MustCompile(`(?s)<tr[^>]*>(.*?)</tr>`)
		rows := reTr.FindAllStringSubmatch(inner, -1)
		if len(rows) == 0 {
			return match
		}

		reTd := regexp.MustCompile(`(?s)<(?:td|th)[^>]*>(.*?)</(?:td|th)>`)
		var mdRows []string
		for i, row := range rows {
			cells := reTd.FindAllStringSubmatch(row[1], -1)
			var cellTexts []string
			for _, cell := range cells {
				text := strings.TrimSpace(regReplace(cell[1], `<[^>]+>`, ""))
				cellTexts = append(cellTexts, text)
			}
			mdRows = append(mdRows, "| "+strings.Join(cellTexts, " | ")+" |")
			// Add separator after first row (header).
			if i == 0 {
				var seps []string
				for range cellTexts {
					seps = append(seps, "---")
				}
				mdRows = append(mdRows, "| "+strings.Join(seps, " | ")+" |")
			}
		}

		return "\n" + strings.Join(mdRows, "\n") + "\n"
	})
}

// regReplace is a helper for regex replacement.
func regReplace(s, pattern, repl string) string {
	re := regexp.MustCompile(pattern)
	return re.ReplaceAllString(s, repl)
}
