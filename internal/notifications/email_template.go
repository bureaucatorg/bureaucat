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
	"bereaucat/internal/branding"
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
	"inc":     func(i int) int { return i + 1 },
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
  .card { border-radius:0 !important; border-left:0 !important; border-right:0 !important; }
  .px { padding-left:16px !important; padding-right:16px !important; }
  .title { font-size:20px !important; line-height:28px !important; }
}
</style>
</head>
<body>
<table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0" bgcolor="#f4f4f5">
<tr><td class="outer" align="center" style="padding:32px 12px 20px">
<table class="card" role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0" bgcolor="#ffffff" style="max-width:580px;border:1px solid #e4e4e7;border-radius:12px">
<tr><td height="3" bgcolor="#f59e0b" style="font-size:0;line-height:0;border-radius:12px 12px 0 0">&nbsp;</td></tr>

<tr><td class="px" style="padding:14px 28px">
<table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0"><tr>
{{if .EmbedLogo}}<td width="{{.LogoWidth}}" valign="middle" style="padding-right:8px;font-size:0;line-height:0"><img src="cid:{{logoCID}}" width="{{.LogoWidth}}" height="24" alt="" style="border:0;outline:none;text-decoration:none"></td>{{end}}
<td valign="middle" style="font-family:{{sans}};font-size:16px;line-height:24px;color:#09090b;letter-spacing:-0.01em"><b>{{if .Custom}}{{.AppName}}{{else}}Bureau<span style="color:#f59e0b">Cat</span>{{end}}</b></td>
<td align="right" valign="middle" style="font-family:{{sans}};font-size:12px;line-height:24px;color:#a1a1aa">{{.ProjectName}}</td>
</tr></table>
</td></tr>
<tr><td height="1" bgcolor="#f4f4f5" style="font-size:0;line-height:0">&nbsp;</td></tr>

<tr><td class="px" style="padding:28px 28px 0">
<table role="presentation" cellpadding="0" cellspacing="0" border="0"><tr>
<td valign="middle" style="font-family:{{mono}};font-size:12px;line-height:20px;color:#b45309;letter-spacing:0.02em"><b>{{.TaskKey}}</b></td>
{{with .State}}<td valign="middle" style="padding-left:10px">
<table role="presentation" cellpadding="0" cellspacing="0" border="0"><tr>
<td valign="middle" style="padding:1px 9px 2px;border:1px solid #e4e4e7;border-radius:9999px;font-family:{{sans}};font-size:12px;line-height:18px;color:#3f3f46;white-space:nowrap"><span style="color:{{.Color}}">&#9679;</span>&nbsp;{{.Name}}</td>
</tr></table>
</td>{{end}}
</tr></table>
</td></tr>
<tr><td class="px title" style="padding:8px 28px 0;font-family:{{sans}};font-size:22px;line-height:30px;color:#09090b;letter-spacing:-0.015em"><b>{{.TaskTitle}}</b></td></tr>

<tr><td class="px" style="padding:28px 28px 18px">
<table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0"><tr>
<td valign="bottom" style="font-family:{{sans}};font-size:13px;line-height:18px;color:#09090b"><b>{{.Heading}}</b></td>
{{if .Count}}<td align="right" valign="bottom" style="font-family:{{sans}};font-size:12px;line-height:18px;color:#a1a1aa">{{.Count}}</td>{{end}}
</tr>
<tr><td colspan="2" style="padding-top:10px"><table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0"><tr><td height="1" bgcolor="#f4f4f5" style="font-size:0;line-height:0">&nbsp;</td></tr></table></td></tr>
</table>
</td></tr>

{{range $g := .Groups}}<tr><td class="px" style="padding:0 28px">
<table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0">
<tr>
<td colspan="3" width="28" valign="top">
<table role="presentation" cellpadding="0" cellspacing="0" border="0"><tr>
<td width="28" height="28" align="center" valign="middle" bgcolor="{{if $g.System}}#f4f4f5{{else}}#fef3c7{{end}}" style="border-radius:14px;font-family:{{mono}};font-size:10px;line-height:28px;color:{{if $g.System}}#52525b{{else}}#92400e{{end}}"><b>{{$g.Initials}}</b></td>
</tr></table>
</td>
<td valign="middle" style="padding-left:12px;font-family:{{sans}};font-size:14px;line-height:28px;color:#09090b"><b>{{$g.Actor}}</b>{{if $g.System}} <span style="font-size:12px;color:#a1a1aa">automation</span>{{end}}</td>
</tr>
<tr>
<td width="13" style="font-size:0;line-height:0">&nbsp;</td>
<td width="2" {{if not $g.Last}}bgcolor="#e4e4e7" {{end}}style="font-size:0;line-height:0">&nbsp;</td>
<td width="13" style="font-size:0;line-height:0">&nbsp;</td>
<td valign="top" style="padding:2px 0 {{if $g.Last}}4{{else}}24{{end}}px 12px">
<table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0">
{{range $j, $e := $g.Items}}<tr><td style="padding-top:{{if $j}}14{{else}}0{{end}}px;font-family:{{sans}};font-size:14px;line-height:21px;color:#52525b">
{{$e.Verb}}{{if $e.Target}} <b style="color:#09090b">{{$e.Target}}</b>{{end}}{{with $e.Label}} <span style="padding:1px 8px 2px;border:1px solid #e4e4e7;border-radius:9999px;font-size:12px;color:#3f3f46;white-space:nowrap"><span style="color:{{.Color}}">&#9679;</span>&nbsp;{{.Name}}</span>{{end}}
</td></tr>
{{if $e.To}}<tr><td style="padding-top:8px">
<table role="presentation" cellpadding="0" cellspacing="0" border="0"><tr>
<td valign="middle" style="padding:1px 9px 2px;border:1px solid #e4e4e7;border-radius:9999px;font-family:{{sans}};font-size:12px;line-height:18px;color:#71717a;white-space:nowrap">{{$e.From.Name}}</td>
<td valign="middle" style="padding:0 8px;font-family:{{sans}};font-size:13px;line-height:18px;color:#d4d4d8">&rarr;</td>
<td bgcolor="#fffbeb" valign="middle" style="padding:1px 9px 2px;border:1px solid #fcd34d;border-radius:9999px;font-family:{{sans}};font-size:12px;line-height:18px;color:#92400e;white-space:nowrap"><span style="color:#f59e0b">&#9679;</span>&nbsp;<b>{{$e.To.Name}}</b></td>
</tr></table>
</td></tr>{{end}}
{{if $e.Workflow}}<tr><td style="padding-top:10px">
<table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0" bgcolor="#fafafa" style="border:1px solid #f0f0f1;border-radius:8px">
<tr><td colspan="2" style="padding:8px 12px 2px;font-family:{{sans}};font-size:11px;line-height:16px;color:#a1a1aa">&#8627;&nbsp;Workflow Engine</td></tr>
{{range $k, $w := $e.Workflow}}<tr>
<td width="76" valign="top" style="padding:2px 0 {{if eq (len $e.Workflow) (inc $k)}}8{{else}}0{{end}}px 12px;font-family:{{sans}};font-size:12px;line-height:18px;color:#a1a1aa">{{$w.Verb}}</td>
<td valign="top" style="padding:2px 12px {{if eq (len $e.Workflow) (inc $k)}}8{{else}}0{{end}}px 0;font-family:{{sans}};font-size:12px;line-height:18px;color:#3f3f46">{{$w.Names}}</td>
</tr>{{end}}
</table>
</td></tr>{{end}}
{{if $e.Quote}}<tr><td style="padding-top:10px">
<table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0" bgcolor="#fafafa" style="border:1px solid #f0f0f1;border-radius:8px"><tr>
<td style="padding:10px 14px;font-family:{{sans}};font-size:14px;line-height:22px;color:#3f3f46;word-break:break-word">{{$e.QuoteHTML}}</td>
</tr></table>
</td></tr>{{end}}
{{end}}</table>
</td>
</tr>
</table>
</td></tr>{{end}}

<tr><td class="px" style="padding:24px 28px 32px">
<table role="presentation" cellpadding="0" cellspacing="0" border="0"><tr>
<td bgcolor="#18181b" style="padding:11px 20px;border-radius:8px;font-family:{{sans}};font-size:14px;line-height:20px"><a href="{{.Link}}" style="color:#fafafa;text-decoration:none"><b>{{.LinkLabel}}&nbsp;&rarr;</b></a></td>
</tr></table>
</td></tr>

</table>
</td></tr>
<tr><td class="px" align="center" style="padding:0 28px 32px;font-family:{{sans}};font-size:12px;line-height:19px;color:#a1a1aa">
{{.Note}}{{if .SettingsLink}}<br><a href="{{.SettingsLink}}" style="color:#71717a;text-decoration:underline">Manage email notifications</a>{{end}}
{{if .Custom}}<br>Powered by <a href="https://bureaucat.org" style="color:#a1a1aa">Bureaucat</a>{{end}}
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
	Workflow  []workflowChange
	via       string
	isTrigger bool
}

type workflowChange struct {
	Verb, Names string
}

// emailGroup is a run of consecutive events by one actor, rendered under a single avatar.
type emailGroup struct {
	Actor    string
	Initials string
	System   bool
	Last     bool
	Items    []emailEvent
}

type emailData struct {
	ProjectName  string
	TaskKey      string
	TaskTitle    string
	State        *emailPill
	Heading      string
	Count        string
	Events       []emailEvent
	Groups       []emailGroup
	Link         string
	LinkLabel    string
	Note         string
	SettingsLink string
	EmbedLogo    bool
	LogoWidth    int
	AppName      string
	Custom       bool
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

// groupEvents nests workflow assignee changes under the state change/creation by the user who
// triggered them, and merges back-to-back assignee changes with the same actor and verb.
func groupEvents(activity []store.ListEmailActivityRow) []emailEvent {
	events := make([]emailEvent, 0, len(activity))
	triggers := map[string]int{}
	var workflow map[int][]workflowChange

	for _, a := range activity {
		e := eventFromActivity(a)
		isAssignee := a.ActivityType == "assignee_added" || a.ActivityType == "assignee_removed"
		trigger := strings.TrimSpace(a.FirstName + " " + a.LastName)

		if i, ok := triggers[trigger]; ok && a.Via == auth.ViaWorkflow && isAssignee && e.Target != "" {
			if workflow == nil {
				workflow = map[int][]workflowChange{}
			}
			workflow[i] = appendGrouped(workflow[i], e.Verb, e.Target)
			continue
		}
		if n := len(events); n > 0 && isAssignee && e.Target != "" {
			if prev := &events[n-1]; prev.Actor == e.Actor && prev.via == e.via && prev.Verb == e.Verb && prev.Target != "" && prev.Workflow == nil && !prev.isTrigger {
				prev.Target += ", " + e.Target
				continue
			}
		}
		if a.Via != auth.ViaWorkflow && (a.ActivityType == "state_changed" || a.ActivityType == "task_created") {
			e.isTrigger = true
			triggers[trigger] = len(events)
		}
		events = append(events, e)
	}

	for i, changes := range workflow {
		events[i].Workflow = changes
	}
	return events
}

func appendGrouped(changes []workflowChange, verb, name string) []workflowChange {
	for i := range changes {
		if changes[i].Verb == verb {
			changes[i].Names += ", " + name
			return changes
		}
	}
	return append(changes, workflowChange{verb, name})
}

// groupByActor folds consecutive events by the same actor into one group and capitalises each verb.
func groupByActor(events []emailEvent) []emailGroup {
	var groups []emailGroup
	for _, e := range events {
		e.Verb = capitalize(e.Verb)
		if n := len(groups); n > 0 && groups[n-1].Actor == e.Actor && groups[n-1].System == (e.via == auth.ViaWorkflow) {
			groups[n-1].Items = append(groups[n-1].Items, e)
			continue
		}
		groups = append(groups, emailGroup{Actor: e.Actor, Initials: e.Initials, System: e.via == auth.ViaWorkflow, Items: []emailEvent{e}})
	}
	if n := len(groups); n > 0 {
		groups[n-1].Last = true
	}
	return groups
}

func capitalize(s string) string {
	r := []rune(s)
	if len(r) == 0 {
		return s
	}
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

func eventFromActivity(a store.ListEmailActivityRow) emailEvent {
	first, last := systemActor(a.Via, a.FirstName, a.LastName)
	e := newEvent(first, last, activityLabel(a.ActivityType))
	e.via = a.Via
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

func notificationEmail(cfg mailer.Settings, brand branding.Brand, row store.ClaimEmailNotificationsRow, events []emailEvent) (mailer.Message, error) {
	base := strings.TrimRight(cfg.AppURL, "/")
	taskKey := fmt.Sprintf("%s-%d", row.ProjectKey, row.TaskNumber)
	link := fmt.Sprintf("%s/projects/%s/tasks/%d", base, row.ProjectKey, row.TaskNumber)
	if row.CommentID.Valid {
		link += "#comment-" + uuid.UUID(row.CommentID.Bytes).String()
	}

	var count string
	if n := max(len(events), int(row.EventCount)); n > 1 {
		count = fmt.Sprintf("%d updates", n)
	}

	data := emailData{
		ProjectName:  row.ProjectName,
		TaskKey:      taskKey,
		TaskTitle:    row.TaskTitle,
		State:        &emailPill{Name: row.StateName, Color: safeColor(row.StateColor.String)},
		Heading:      "Activity",
		Count:        count,
		Events:       events,
		Link:         link,
		LinkLabel:    "View task",
		Note:         "You are receiving this because you are involved in " + taskKey + ".",
		SettingsLink: base + "/settings",
	}
	return buildEmail(cfg, brand, row.RecipientEmail, fmt.Sprintf("[%s] %s", taskKey, row.TaskTitle), data)
}

// TestEmail builds the SMTP connection test email in the same layout as notifications.
func TestEmail(cfg mailer.Settings, brand branding.Brand, to, senderFirstName, senderLastName string) (mailer.Message, error) {
	data := emailData{
		ProjectName: "Admin",
		TaskKey:     "SMTP TEST",
		TaskTitle:   "Email delivery is working",
		Heading:     "Details",
		Events:      []emailEvent{newEvent(senderFirstName, senderLastName, "sent this test email from the admin email settings.")},
		Link:        strings.TrimRight(cfg.AppURL, "/"),
		LinkLabel:   "Open " + brand.Name,
		Note:        "Notification emails will arrive in this format.",
	}
	return buildEmail(cfg, brand, to, brand.Name+" test email", data)
}

// SampleEmail builds the exact email sent for activityType ("batched" for several updates,
// "workflow" for a state change with workflow assignee changes), using placeholder task data.
func SampleEmail(cfg mailer.Settings, brand branding.Brand, to, activityType, actorFirstName, actorLastName string) (mailer.Message, error) {
	types := []string{activityType}
	switch activityType {
	case "batched":
		types = []string{"comment_created", "label_added", "state_changed"}
	case "workflow":
		types = []string{"state_changed"}
	case "everything":
		return everythingSample(cfg, brand, to, actorFirstName, actorLastName)
	default:
		if _, ok := activityLabels[activityType]; !ok && activityType != "description_updated" {
			return mailer.Message{}, fmt.Errorf("unknown activity type %q", activityType)
		}
	}

	activity := make([]store.ListEmailActivityRow, 0, len(types)+3)
	for _, t := range types {
		activity = append(activity, sampleActivity(t, actorFirstName, actorLastName))
	}
	if activityType == "workflow" {
		for _, c := range []struct{ activityType, assignee string }{
			{"assignee_removed", `{"first_name":"Sam","last_name":"Sample"}`},
			{"assignee_added", `{"first_name":"Alex","last_name":"Example"}`},
			{"assignee_added", `{"first_name":"Jordan","last_name":"Doe"}`},
		} {
			a := store.ListEmailActivityRow{ActivityType: c.activityType, Via: auth.ViaWorkflow, FirstName: actorFirstName, LastName: actorLastName}
			if c.activityType == "assignee_added" {
				a.NewValue = []byte(c.assignee)
			} else {
				a.OldValue = []byte(c.assignee)
			}
			activity = append(activity, a)
		}
	}
	events := groupEvents(activity)

	row := store.ClaimEmailNotificationsRow{
		ActivityType:   types[len(types)-1],
		EventCount:     int32(len(activity)),
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
	return notificationEmail(cfg, brand, row, events)
}

// everythingSample builds a long digest covering every activity type, workflow nesting and rich comments.
func everythingSample(cfg mailer.Settings, brand branding.Brand, to, recipientFirstName, recipientLastName string) (mailer.Message, error) {
	type person struct{ first, last string }
	alex, jordan, casey := person{"ALEX", "EXAMPLE"}, person{"Jordan", "Doe"}, person{"Casey", "Tester"}
	riley, morgan, taylor := person{"Riley", "Placeholder"}, person{"Morgan", "Demo"}, person{"Taylor", "Sample"}
	name := func(p person) string { return fmt.Sprintf(`{"first_name":%q,"last_name":%q}`, p.first, p.last) }

	var activity []store.ListEmailActivityRow
	add := func(by person, activityType, field, oldV, newV string) {
		a := store.ListEmailActivityRow{ActivityType: activityType, FirstName: by.first, LastName: by.last}
		if field != "" {
			a.FieldName = pgtype.Text{String: field, Valid: true}
		}
		if oldV != "" {
			a.OldValue = []byte(oldV)
		}
		if newV != "" {
			a.NewValue = []byte(newV)
		}
		activity = append(activity, a)
	}
	workflow := func(trigger person, activityType string, assignee person) {
		add(trigger, activityType, "", "", "")
		a := &activity[len(activity)-1]
		a.Via = auth.ViaWorkflow
		if activityType == "assignee_added" {
			a.NewValue = []byte(name(assignee))
		} else {
			a.OldValue = []byte(name(assignee))
		}
	}
	comment := func(by person, activityType, html string) {
		b, _ := json.Marshal(map[string]string{"content": html})
		add(by, activityType, "", "", string(b))
	}

	add(alex, "task_created", "", "", "")
	workflow(alex, "assignee_added", jordan)
	workflow(alex, "assignee_added", casey)
	add(alex, "assignee_added", "", "", name(person{recipientFirstName, recipientLastName}))
	comment(jordan, "comment_created", "<p>reasoning please?</p>")
	comment(alex, "comment_created", `<p><a href="/profile/0">@Jordan Doe</a> Need to see these dashboards : <a href="https://dashboards.example.com/dashboard/673-admin-table-diagnosis-and-square-off-stats?segment=Both&amp;tab=68-historical-stats">https://dashboards.example.com/dashboard/673-admin-table-diagnosis-and-square-off-stats?segment=Both&amp;tab=68-historical-stats</a>, <a href="https://dashboards.example.com/dashboard/624-monitoring-dashboard">https://dashboards.example.com/dashboard/624-monitoring-dashboard</a>, <a href="https://dashboards.example.com/dashboard/669-market-adapter">https://dashboards.example.com/dashboard/669-market-adapter</a></p>`)
	comment(alex, "comment_created", "Also, I publish data for multiple tables, including ticks, profiles, etc. I want to check those as well.")
	add(jordan, "state_changed", "", `{"name":"Approval Pending"}`, `{"name":"CSec Approved"}`)
	workflow(jordan, "assignee_added", riley)
	workflow(jordan, "assignee_added", morgan)
	workflow(jordan, "assignee_removed", jordan)
	workflow(jordan, "assignee_removed", casey)
	add(morgan, "state_changed", "", `{"name":"CSec Approved"}`, `{"name":"ISec Approved"}`)
	workflow(morgan, "assignee_removed", riley)
	workflow(morgan, "assignee_removed", morgan)
	add(casey, "label_added", "", "", `{"name":"Access request","color":"#3b82f6"}`)
	add(casey, "label_removed", "", `{"name":"Needs info","color":"#ef4444"}`, "")
	add(casey, "task_updated", "priority", "", `4`)
	add(casey, "task_updated", "title", "", `"Need access for analytics dashboards"`)
	add(riley, "task_updated", "due_date", "", `"2026-11-20T00:00:00Z"`)
	add(riley, "task_updated", "start_date", "", `null`)
	activity = append(activity, sampleActivity("description_updated", riley.first, riley.last))
	comment(jordan, "comment_updated", "<p>reasoning please? Which <strong>tables</strong> exactly?</p><pre><code>ticks\nprofiles</code></pre>")
	add(jordan, "comment_deleted", "", "", "")
	add(morgan, "blocker_added", "", "", "")
	add(morgan, "blocker_removed", "", "", "")
	add(morgan, "assignee_removed", "", name(casey), "")
	add(morgan, "task_updated", "estimate", "", `3`)
	workflow(taylor, "assignee_added", taylor)

	row := store.ClaimEmailNotificationsRow{
		ActivityType:   "state_changed",
		EventCount:     int32(len(activity)),
		RecipientEmail: to,
		TaskNumber:     985,
		TaskTitle:      "Need access for analytics dashboards",
		ProjectKey:     "DEVOP",
		ProjectName:    "Devops",
		StateName:      "ISec Approved",
		StateColor:     pgtype.Text{String: "#22c55e", Valid: true},
	}
	return notificationEmail(cfg, brand, row, groupEvents(activity))
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

func buildEmail(cfg mailer.Settings, brand branding.Brand, to, subject string, data emailData) (mailer.Message, error) {
	logo, logoWidth := logoPNG, 26
	if brand.Logo != nil {
		logo, logoWidth = brand.Logo.Data, min(160, 26*brand.Logo.Width/brand.Logo.Height)
	}
	data.EmbedLogo, data.LogoWidth = cfg.EmbedLogo, max(logoWidth, 1)
	data.AppName, data.Custom = brand.Name, brand.Custom
	data.Groups = groupByActor(data.Events)
	var body bytes.Buffer
	if err := emailHTML.Execute(&body, data); err != nil {
		return mailer.Message{}, err
	}
	msg := mailer.Message{To: to, Subject: subject, Text: plainText(data), HTML: body.String()}
	if cfg.EmbedLogo {
		msg.Inline = []mailer.InlineFile{{ContentID: logoCID, ContentType: "image/png", Filename: "logo.png", Data: logo}}
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
		for _, w := range e.Workflow {
			fmt.Fprintf(&b, "  ↳ Workflow Engine %s %s\n", w.Verb, w.Names)
		}
		if e.Quote != "" {
			fmt.Fprintf(&b, "  > %s\n", e.Quote)
		}
	}
	fmt.Fprintf(&b, "\n%s: %s\n\n--\n%s\n", data.LinkLabel, data.Link, data.Note)
	if data.SettingsLink != "" {
		fmt.Fprintf(&b, "Manage email notifications: %s\n", data.SettingsLink)
	}
	if data.Custom {
		b.WriteString("Powered by Bureaucat\n")
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
