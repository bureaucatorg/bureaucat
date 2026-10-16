package notifications

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"html"
	"html/template"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"bereaucat/internal/mailer"
	"bereaucat/internal/auth"
	"bereaucat/internal/store"
)

// Mirrors ACTIVITY_TYPE_LABELS in web/app/types/activity.ts.
var activityLabels = map[string]string{
	"task_created":     "created the task",
	"task_updated":     "updated",
	"task_deleted":     "deleted the task",
	"assignee_added":   "added assignee",
	"assignee_removed": "removed assignee",
	"label_added":      "added label",
	"label_removed":    "removed label",
	"state_changed":    "changed state",
	"comment_created":  "added a comment",
	"comment_updated":  "edited a comment",
	"comment_deleted":  "deleted a comment",
	"blocker_added":    "linked a blocker",
	"blocker_removed":  "unlinked a blocker",
}

// Mirrors PRIORITY_LABELS in web/app/types/task.ts.
var priorityLabels = map[int]string{0: "No priority", 1: "Low", 2: "Medium", 3: "High", 4: "Urgent"}

const (
	fontSans     = `-apple-system,BlinkMacSystemFont,'Segoe UI',Helvetica,Arial,sans-serif`
	fontMono     = `ui-monospace,SFMono-Regular,Menlo,Consolas,monospace`
	excerptRunes = 280
	quoteRunes   = 600
	neutralColor = "#a1a1aa"
)

const logoCID = "logo@bureaucat"

//go:embed email-logo.png
var logoPNG []byte

var (
	tagRe      = regexp.MustCompile(`<[^>]*>`)
	hexColorRe = regexp.MustCompile(`^#[0-9a-fA-F]{3}([0-9a-fA-F]{3})?$`)
)

// Email-client safe: tables, bgcolor and inline padding (no margin/display); rounded corners degrade to square.
// The <style> block only tightens spacing on phones; clients that drop it get the desktop layout.
var emailHTML = template.Must(template.New("email").Funcs(template.FuncMap{
	"sans":    func() template.CSS { return fontSans },
	"mono":    func() template.CSS { return fontMono },
	"logoCID": func() string { return logoCID },
}).Parse(`<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<meta name="color-scheme" content="light">
<title>{{.TaskKey}} {{.TaskTitle}}</title>
<style>
@media only screen and (max-width:600px) {
  .outer { padding:0 !important; }
  .card { border-radius:0 !important; }
  .px { padding-left:16px !important; padding-right:16px !important; }
  .title { font-size:20px !important; line-height:28px !important; }
}
</style>
</head>
<body>
<table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0" bgcolor="#f4f4f5">
<tr><td class="outer" align="center" style="padding:32px 12px">
<table class="card" role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0" bgcolor="#ffffff" style="max-width:580px;border-radius:12px">

<tr><td class="px" style="padding:16px 28px">
<table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0"><tr>
{{if .EmbedLogo}}<td width="26" valign="middle" style="padding-right:8px;font-size:0;line-height:0"><img src="cid:{{logoCID}}" width="26" height="26" alt="" style="border:0;outline:none;text-decoration:none"></td>{{end}}
<td valign="middle" style="font-family:{{sans}};font-size:17px;line-height:26px;color:#09090b"><b>Bureau<span style="color:#f59e0b">Cat</span></b></td>
<td align="right" valign="middle" style="font-family:{{sans}};font-size:12px;line-height:26px;color:#71717a">{{.ProjectName}}</td>
</tr></table>
</td></tr>
<tr><td height="1" bgcolor="#e4e4e7" style="font-size:0;line-height:0">&nbsp;</td></tr>

<tr><td class="px" style="padding:24px 28px 0;font-family:{{mono}};font-size:12px;line-height:18px;color:#d97706"><b>{{.TaskKey}}</b></td></tr>
<tr><td class="px title" style="padding:4px 28px 0;font-family:{{sans}};font-size:22px;line-height:30px;color:#09090b"><b>{{.TaskTitle}}</b></td></tr>
{{with .State}}<tr><td class="px" style="padding:10px 28px 0">
<table role="presentation" cellpadding="0" cellspacing="0" border="0"><tr>
<td bgcolor="#f4f4f5" valign="middle" style="padding:2px 10px 4px;border-radius:9999px;font-family:{{sans}};font-size:12px;line-height:18px;color:#3f3f46"><span style="color:{{.Color}}">&#9679;</span>&nbsp;{{.Name}}</td>
</tr></table>
</td></tr>{{end}}

<tr><td class="px" style="padding:22px 28px 0"><table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0"><tr><td height="1" bgcolor="#e4e4e7" style="font-size:0;line-height:0">&nbsp;</td></tr></table></td></tr>
<tr><td class="px" style="padding:16px 28px 0;font-family:{{mono}};font-size:11px;line-height:16px;letter-spacing:0.12em;color:#a1a1aa">{{.Heading}}</td></tr>

{{range .Events}}<tr><td class="px" style="padding:12px 28px 0">
<table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0">
<tr>
<td width="26" valign="top">
<table role="presentation" cellpadding="0" cellspacing="0" border="0"><tr>
<td width="26" height="26" align="center" valign="middle" bgcolor="#fef3c7" style="border-radius:13px;font-family:{{mono}};font-size:10px;line-height:26px;color:#92400e"><b>{{.Initials}}</b></td>
</tr></table>
</td>
<td valign="middle" style="padding-left:10px;font-family:{{sans}};font-size:14px;line-height:20px;color:#52525b">
<b style="color:#09090b">{{.Actor}}</b> {{.Verb}}{{if .Target}} <b style="color:#09090b">{{.Target}}</b>{{end}}{{with .Label}} <span style="padding:1px 8px 3px;border-radius:9999px;background-color:#f4f4f5;font-size:12px;color:#3f3f46"><span style="color:{{.Color}}">&#9679;</span>&nbsp;{{.Name}}</span>{{end}}
</td>
</tr>
{{if .To}}<tr><td colspan="2" style="padding-top:8px">
<table role="presentation" cellpadding="0" cellspacing="0" border="0"><tr>
<td bgcolor="#f4f4f5" valign="middle" style="padding:2px 10px 4px;border-radius:9999px;font-family:{{sans}};font-size:12px;line-height:18px;color:#71717a"><span style="color:{{.From.Color}}">&#9679;</span>&nbsp;{{.From.Name}}</td>
<td valign="middle" style="padding:0 8px 6px;font-family:{{sans}};font-size:14px;line-height:12px;color:#a1a1aa">&rarr;</td>
<td bgcolor="#f4f4f5" valign="middle" style="padding:2px 10px 4px;border-radius:9999px;font-family:{{sans}};font-size:12px;line-height:18px;color:#09090b"><span style="color:{{.To.Color}}">&#9679;</span>&nbsp;{{.To.Name}}</td>
</tr></table>
</td></tr>{{end}}
{{if .Quote}}<tr><td colspan="2" style="padding-top:8px">
<table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0" bgcolor="#fafafa"><tr>
<td width="3" bgcolor="#f59e0b" style="font-size:0;line-height:0">&nbsp;</td>
<td style="padding:10px 12px;font-family:{{sans}};font-size:14px;line-height:21px;color:#3f3f46">{{.QuoteHTML}}</td>
</tr></table>
</td></tr>{{end}}
</table>
</td></tr>{{end}}

<tr><td class="px" style="padding:24px 28px 28px">
<table role="presentation" cellpadding="0" cellspacing="0" border="0"><tr>
<td bgcolor="#18181b" style="padding:10px 18px;border-radius:8px;font-family:{{sans}};font-size:14px;line-height:20px"><a href="{{.Link}}" style="color:#fafafa;text-decoration:none"><b>{{.LinkLabel}}&nbsp;&rarr;</b></a></td>
</tr></table>
</td></tr>

<tr><td height="1" bgcolor="#e4e4e7" style="font-size:0;line-height:0">&nbsp;</td></tr>
<tr><td class="px" style="padding:16px 28px 20px;font-family:{{sans}};font-size:12px;line-height:18px;color:#71717a">
{{.Note}}{{if .SettingsLink}} <a href="{{.SettingsLink}}" style="color:#71717a">Manage email notifications</a>{{end}}
</td></tr>

</table>
</td></tr>
</table>
</body>
</html>`))

type emailPill struct {
	Name, Color string
}

type emailEvent struct {
	Actor     string
	Initials  string
	Verb      string
	Target    string
	Quote     string
	QuoteHTML template.HTML
	Label     *emailPill
	From      *emailPill
	To        *emailPill
}

type emailData struct {
	ProjectName  string
	TaskKey      string
	TaskTitle    string
	State        *emailPill
	Heading      string
	Events       []emailEvent
	Link         string
	LinkLabel    string
	Note         string
	SettingsLink string
	EmbedLogo    bool
}

func activityLabel(activityType string) string {
	if label, ok := activityLabels[activityType]; ok {
		return label
	}
	return strings.ReplaceAll(activityType, "_", " ")
}

// systemActor names changes made by workflow rules "Workflow Engine" rather than the triggering user.
func systemActor(via, firstName, lastName string) (string, string) {
	if via == auth.ViaWorkflow {
		return "Workflow", "Engine"
	}
	return firstName, lastName
}

func newEvent(firstName, lastName, verb string) emailEvent {
	return emailEvent{
		Actor:    strings.TrimSpace(firstName + " " + lastName),
		Initials: initials(firstName, lastName),
		Verb:     verb,
	}
}

func eventFromActivity(a store.ListEmailActivityRow) emailEvent {
	first, last := systemActor(a.Via, a.FirstName, a.LastName)
	e := newEvent(first, last, activityLabel(a.ActivityType))
	oldV, newV := jsonMap(a.OldValue), jsonMap(a.NewValue)

	switch a.ActivityType {
	case "comment_created", "comment_updated":
		e.setQuote(str(newV["content"]))
	case "assignee_added":
		e.Verb, e.Target = "assigned", strings.TrimSpace(str(newV["first_name"])+" "+str(newV["last_name"]))
	case "assignee_removed":
		e.Verb, e.Target = "unassigned", strings.TrimSpace(str(oldV["first_name"])+" "+str(oldV["last_name"]))
	case "label_added", "label_removed":
		v := newV
		if a.ActivityType == "label_removed" {
			v = oldV
		}
		if name := str(v["name"]); name != "" {
			e.Label = &emailPill{Name: name, Color: safeColor(str(v["color"]))}
		}
	case "state_changed":
		if from, to := str(oldV["name"]), str(newV["name"]); from != "" && to != "" {
			e.From = &emailPill{Name: from, Color: neutralColor}
			e.To = &emailPill{Name: to, Color: "#f59e0b"}
		}
	case "task_updated":
		var quote string
		e.Verb, e.Target, quote = describeFieldChange(a.FieldName.String, jsonValue(a.NewValue))
		e.setQuote(quote)
	}
	return e
}

func (e *emailEvent) setQuote(content string) {
	if e.Quote = excerpt(content); e.Quote != "" {
		e.QuoteHTML = richExcerpt(content, quoteRunes)
	}
}

func describeFieldChange(field string, value any) (verb, target, quote string) {
	switch field {
	case "title":
		return "renamed the task to", str(value), ""
	case "description":
		return "updated the description", "", str(value)
	case "priority":
		if n, ok := value.(float64); ok {
			return "set priority to", priorityLabels[int(n)], ""
		}
	case "start_date", "due_date":
		label := strings.ReplaceAll(field, "_", " ")
		t, err := time.Parse(time.RFC3339, str(value))
		if err != nil {
			return "removed the " + label, "", ""
		}
		return "set the " + label + " to", t.Format("Jan 2, 2006"), ""
	}
	return "updated " + strings.ReplaceAll(field, "_", " "), "", ""
}

func notificationEmail(cfg mailer.Settings, row store.ClaimEmailNotificationsRow, events []emailEvent) (mailer.Message, error) {
	base := strings.TrimRight(cfg.AppURL, "/")
	taskKey := fmt.Sprintf("%s-%d", row.ProjectKey, row.TaskNumber)
	link := fmt.Sprintf("%s/projects/%s/tasks/%d", base, row.ProjectKey, row.TaskNumber)
	if row.CommentID.Valid {
		link += "#comment-" + uuid.UUID(row.CommentID.Bytes).String()
	}

	heading := "ACTIVITY"
	if n := max(len(events), int(row.EventCount)); n > 1 {
		heading = fmt.Sprintf("ACTIVITY · %d UPDATES", n)
	}

	data := emailData{
		ProjectName:  row.ProjectName,
		TaskKey:      taskKey,
		TaskTitle:    row.TaskTitle,
		State:        &emailPill{Name: row.StateName, Color: safeColor(row.StateColor.String)},
		Heading:      heading,
		Events:       events,
		Link:         link,
		LinkLabel:    "View task",
		Note:         "You are receiving this because you are involved in " + taskKey + ".",
		SettingsLink: base + "/settings",
	}
	return buildEmail(cfg, row.RecipientEmail, fmt.Sprintf("[%s] %s", taskKey, row.TaskTitle), data)
}

// TestEmail builds the SMTP connection test email in the same layout as notifications.
func TestEmail(cfg mailer.Settings, to, senderFirstName, senderLastName string) (mailer.Message, error) {
	data := emailData{
		ProjectName: "Admin",
		TaskKey:     "SMTP TEST",
		TaskTitle:   "Email delivery is working",
		Heading:     "DETAILS",
		Events:      []emailEvent{newEvent(senderFirstName, senderLastName, "sent this test email from the admin email settings.")},
		Link:        strings.TrimRight(cfg.AppURL, "/"),
		LinkLabel:   "Open Bureaucat",
		Note:        "Notification emails will arrive in this format.",
	}
	return buildEmail(cfg, to, "Bureaucat test email", data)
}

// SampleEmail builds the exact email sent for activityType ("batched" for several updates),
// using placeholder task data and the real formatting path.
func SampleEmail(cfg mailer.Settings, to, activityType, actorFirstName, actorLastName string) (mailer.Message, error) {
	types := []string{activityType}
	if activityType == "batched" {
		types = []string{"comment_created", "label_added", "state_changed"}
	} else if _, ok := activityLabels[activityType]; !ok && activityType != "description_updated" {
		return mailer.Message{}, fmt.Errorf("unknown activity type %q", activityType)
	}

	events := make([]emailEvent, 0, len(types))
	for _, t := range types {
		events = append(events, eventFromActivity(sampleActivity(t, actorFirstName, actorLastName)))
	}

	row := store.ClaimEmailNotificationsRow{
		ActivityType:   types[len(types)-1],
		EventCount:     int32(len(types)),
		RecipientEmail: to,
		TaskNumber:     1,
		TaskTitle:      "Sample task",
		ProjectKey:     "DEMO",
		ProjectName:    "Demo project",
		StateName:      "In Progress",
		StateColor:     pgtype.Text{String: "#f59e0b", Valid: true},
	}
	if strings.HasPrefix(row.ActivityType, "comment_") {
		row.CommentID = pgtype.UUID{Bytes: uuid.New(), Valid: true}
	}
	return notificationEmail(cfg, row, events)
}

func sampleActivity(activityType, firstName, lastName string) store.ListEmailActivityRow {
	a := store.ListEmailActivityRow{ActivityType: activityType, FirstName: firstName, LastName: lastName}
	assignee := `{"first_name":"Sam","last_name":"Sample"}`
	label := `{"name":"Bug","color":"#ef4444"}`
	switch activityType {
	case "comment_created", "comment_updated":
		a.NewValue = []byte(`{"content":"<p>Looks <strong>good</strong> to me. Two things before the Friday review:</p><ul><li><p>bump the <code>timeout</code> to 30s</p></li><li><p>get sign-off from <a href=\"/profile/0\">@Sam Sample</a></p></li></ul>"}`)
	case "assignee_added":
		a.NewValue = []byte(assignee)
	case "assignee_removed":
		a.OldValue = []byte(assignee)
	case "label_added":
		a.NewValue = []byte(label)
	case "label_removed":
		a.OldValue = []byte(label)
	case "state_changed":
		a.OldValue, a.NewValue = []byte(`{"name":"Todo"}`), []byte(`{"name":"In Progress"}`)
	case "task_updated":
		a.FieldName = pgtype.Text{String: "priority", Valid: true}
		a.NewValue = []byte(`3`)
	case "description_updated":
		a.ActivityType = "task_updated"
		a.FieldName = pgtype.Text{String: "description", Valid: true}
		a.NewValue = []byte(`"<p>Requesting access to port <strong>6379</strong> on the staging cache from the UAT cluster.</p><p></p><p>Reason: testing the new transfer flow internally.</p><ul><li><p>source: <code>10.0.0.10</code></p></li><li><p>target: <code>10.0.0.20</code></p></li></ul>"`)
	}
	return a
}

func buildEmail(cfg mailer.Settings, to, subject string, data emailData) (mailer.Message, error) {
	data.EmbedLogo = cfg.EmbedLogo
	var body bytes.Buffer
	if err := emailHTML.Execute(&body, data); err != nil {
		return mailer.Message{}, err
	}
	msg := mailer.Message{To: to, Subject: subject, Text: plainText(data), HTML: body.String()}
	if cfg.EmbedLogo {
		msg.Inline = []mailer.InlineFile{{ContentID: logoCID, ContentType: "image/png", Filename: "logo.png", Data: logoPNG}}
	}
	return msg, nil
}

func plainText(data emailData) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s · %s\n", data.TaskKey, data.TaskTitle)
	if data.State != nil {
		fmt.Fprintf(&b, "State: %s\n", data.State.Name)
	}
	b.WriteString("\n")
	for _, e := range data.Events {
		fmt.Fprintf(&b, "- %s %s", e.Actor, e.Verb)
		if e.Target != "" {
			b.WriteString(" " + e.Target)
		}
		if e.Label != nil {
			b.WriteString(" " + e.Label.Name)
		}
		if e.To != nil {
			fmt.Fprintf(&b, ": %s → %s", e.From.Name, e.To.Name)
		}
		b.WriteString("\n")
		if e.Quote != "" {
			fmt.Fprintf(&b, "  > %s\n", e.Quote)
		}
	}
	fmt.Fprintf(&b, "\n%s: %s\n\n--\n%s\n", data.LinkLabel, data.Link, data.Note)
	if data.SettingsLink != "" {
		fmt.Fprintf(&b, "Manage email notifications: %s\n", data.SettingsLink)
	}
	return b.String()
}

func excerpt(s string) string {
	s = html.UnescapeString(tagRe.ReplaceAllString(s, " "))
	s = strings.Join(strings.Fields(s), " ")
	if r := []rune(s); len(r) > excerptRunes {
		s = strings.TrimSpace(string(r[:excerptRunes])) + "…"
	}
	return s
}

func initials(firstName, lastName string) string {
	var out []rune
	for _, name := range []string{firstName, lastName} {
		if r := []rune(strings.TrimSpace(name)); len(r) > 0 {
			out = append(out, unicode.ToUpper(r[0]))
		}
	}
	if len(out) == 0 {
		return "?"
	}
	return string(out)
}

func safeColor(c string) string {
	if hexColorRe.MatchString(c) {
		return c
	}
	return neutralColor
}

func jsonMap(b []byte) map[string]any {
	var m map[string]any
	_ = json.Unmarshal(b, &m)
	return m
}

func jsonValue(b []byte) any {
	var v any
	_ = json.Unmarshal(b, &v)
	return v
}

func str(v any) string {
	s, _ := v.(string)
	return s
}
