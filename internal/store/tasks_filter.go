package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Filter limits.
const (
	MaxFilterPredicates = 20
	MaxFilterJSONBytes  = 8192
)

// FilterTree is the top-level filter document. Children are all ANDed together.
type FilterTree struct {
	Children []FilterNode `json:"children"`
}

// FilterNode holds a single Predicate. The struct wrapper is retained so the
// wire shape can grow without a breaking change.
type FilterNode struct {
	Predicate *Predicate `json:"predicate,omitempty"`
}

// Predicate is a single comparison: field op value.
type Predicate struct {
	Field string          `json:"field"`
	Op    string          `json:"op"`
	Value json.RawMessage `json:"value,omitempty"`
}

// ParseFilterTree decodes a filter tree from JSON bytes and enforces size.
func ParseFilterTree(data []byte) (FilterTree, error) {
	if len(data) > MaxFilterJSONBytes {
		return FilterTree{}, fmt.Errorf("filter JSON exceeds %d bytes", MaxFilterJSONBytes)
	}
	var tree FilterTree
	if len(data) == 0 {
		return tree, nil
	}
	if err := json.Unmarshal(data, &tree); err != nil {
		return FilterTree{}, fmt.Errorf("invalid filter JSON: %w", err)
	}
	return tree, nil
}

// CompiledFilter is the output of Compile — a WHERE fragment plus positional args.
type CompiledFilter struct {
	// WhereSQL is either empty or starts with " AND " so callers can append directly.
	WhereSQL string
	Args     []any
}

// Valid sort keys map the external name to the qualified column. Closed list —
// anything else falls back to the default.
var validSortKeys = map[string]string{
	"created_at": "t.created_at",
	"updated_at": "t.updated_at",
	"priority":   "t.priority",
	"due_date":   "t.due_date",
	"start_date": "t.start_date",
	"title":      "t.title",
}

var validSortDirs = map[string]string{
	"asc":  "ASC",
	"desc": "DESC",
}

// argBuffer accumulates positional args and renders $N placeholders that pick
// up at startAt + 1.
type argBuffer struct {
	args    []any
	startAt int
}

func newArgBuffer(startAt int) *argBuffer {
	return &argBuffer{startAt: startAt}
}

func (a *argBuffer) push(v any) string {
	a.args = append(a.args, v)
	return "$" + strconv.Itoa(a.startAt+len(a.args))
}

// Compile walks the tree and produces a WHERE fragment plus args. Children are
// ANDed together. startAt should be the number of fixed args already bound by
// the caller so the first compiled arg placeholder becomes $(startAt+1).
func Compile(tree FilterTree, callerID uuid.UUID, now time.Time, startAt int) (CompiledFilter, error) {
	if len(tree.Children) == 0 {
		return CompiledFilter{}, nil
	}
	if len(tree.Children) > MaxFilterPredicates {
		return CompiledFilter{}, fmt.Errorf("too many filter predicates (max %d)", MaxFilterPredicates)
	}
	for _, c := range tree.Children {
		if c.Predicate == nil {
			return CompiledFilter{}, errors.New("filter node must carry a predicate")
		}
	}

	buf := newArgBuffer(startAt)
	andParts := make([]string, 0, len(tree.Children))
	for _, child := range tree.Children {
		s, err := compilePredicate(buf, callerID, now, *child.Predicate)
		if err != nil {
			return CompiledFilter{}, err
		}
		andParts = append(andParts, s)
	}
	return CompiledFilter{
		WhereSQL: " AND " + strings.Join(andParts, " AND "),
		Args:     buf.args,
	}, nil
}

func compilePredicate(a *argBuffer, callerID uuid.UUID, now time.Time, p Predicate) (string, error) {
	h, ok := predicateHandlers[predicateKey{Field: p.Field, Op: p.Op}]
	if !ok {
		return "", fmt.Errorf("unsupported predicate %s %s", p.Field, p.Op)
	}
	return h(a, callerID, now, p.Value)
}

// predicateKey indexes handlers by (field, op).
type predicateKey struct {
	Field string
	Op    string
}

type predicateHandler func(a *argBuffer, callerID uuid.UUID, now time.Time, value json.RawMessage) (string, error)

// predicateHandlers is the closed set of supported filter operations.
var predicateHandlers = map[predicateKey]predicateHandler{
	// ---- search (free-text across title + description) ----
	// A single predicate whose SQL spans both columns. The frontend's
	// search box emits this; it is deliberately not in the filter catalog
	// so users don't see it as a chip.
	{"search", "contains"}: searchContains,

	// ---- title ----
	{"title", "contains"}:     textLike("t.title", false),
	{"title", "not_contains"}: textLike("t.title", true),
	{"title", "is"}:           textCompare("t.title", "="),
	{"title", "is_not"}:       textCompare("t.title", "<>"),
	{"title", "is_empty"}:     textEmpty("t.title", true),
	{"title", "is_set"}:       textEmpty("t.title", false),

	// ---- description ----
	{"description", "contains"}:     textLike("t.description", false),
	{"description", "not_contains"}: textLike("t.description", true),
	{"description", "is_empty"}:     textEmpty("t.description", true),
	{"description", "is_set"}:       textEmpty("t.description", false),

	// ---- state (by state_id) ----
	{"state", "in"}:     uuidIn("t.state_id", false),
	{"state", "not_in"}: uuidIn("t.state_id", true),

	// ---- state_type (enum) ----
	{"state_type", "in"}:     stringIn("ps.state_type::text", false),
	{"state_type", "not_in"}: stringIn("ps.state_type::text", true),

	// ---- priority (0-4) ----
	{"priority", "in"}:     intIn("t.priority", false),
	{"priority", "not_in"}: intIn("t.priority", true),
	{"priority", "gte"}:    intCompare("t.priority", ">="),
	{"priority", "lte"}:    intCompare("t.priority", "<="),

	// ---- created_by ----
	{"created_by", "in"}:        uuidIn("t.created_by", false),
	{"created_by", "not_in"}:    uuidIn("t.created_by", true),
	{"created_by", "is_me"}:     meCompare("t.created_by", "="),
	{"created_by", "is_not_me"}: meCompare("t.created_by", "<>"),

	// ---- assignees (join) ----
	{"assignees", "has_any"}:  assigneesExists(false, false),
	{"assignees", "has_none"}: assigneesExists(true, false),
	{"assignees", "has_all"}:  assigneesHasAll,
	{"assignees", "is_empty"}: assigneesPresence(false),
	{"assignees", "is_set"}:   assigneesPresence(true),

	// ---- labels (join) ----
	{"labels", "has_any"}:  labelsExists(false, false),
	{"labels", "has_none"}: labelsExists(true, false),
	{"labels", "has_all"}:  labelsHasAll,
	{"labels", "is_empty"}: labelsPresence(false),
	{"labels", "is_set"}:   labelsPresence(true),

	// ---- dates ----
	{"start_date", "before"}:  dateOp("t.start_date", "before"),
	{"start_date", "after"}:   dateOp("t.start_date", "after"),
	{"start_date", "between"}: dateBetween("t.start_date"),
	{"start_date", "is_empty"}: nullCheck("t.start_date", true),
	{"start_date", "is_set"}:   nullCheck("t.start_date", false),

	{"due_date", "before"}:  dateOp("t.due_date", "before"),
	{"due_date", "after"}:   dateOp("t.due_date", "after"),
	{"due_date", "between"}: dateBetween("t.due_date"),
	{"due_date", "is_empty"}: nullCheck("t.due_date", true),
	{"due_date", "is_set"}:   nullCheck("t.due_date", false),
	{"due_date", "overdue"}:  dueDateOverdue,

	{"created_at", "before"}:  dateOp("t.created_at", "before"),
	{"created_at", "after"}:   dateOp("t.created_at", "after"),
	{"created_at", "between"}: dateBetween("t.created_at"),

	{"updated_at", "before"}:  dateOp("t.updated_at", "before"),
	{"updated_at", "after"}:   dateOp("t.updated_at", "after"),
	{"updated_at", "between"}: dateBetween("t.updated_at"),

	// ---- comment_count ----
	{"comment_count", "eq"}:  commentCount("="),
	{"comment_count", "ne"}:  commentCount("<>"),
	{"comment_count", "gt"}:  commentCount(">"),
	{"comment_count", "gte"}: commentCount(">="),
	{"comment_count", "lt"}:  commentCount("<"),
	{"comment_count", "lte"}: commentCount("<="),
}

// -------- value decoders --------

func decodeString(raw json.RawMessage) (string, error) {
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return "", fmt.Errorf("expected string value: %w", err)
	}
	return s, nil
}

func decodeStringArray(raw json.RawMessage) ([]string, error) {
	var arr []string
	if err := json.Unmarshal(raw, &arr); err != nil {
		return nil, fmt.Errorf("expected string array: %w", err)
	}
	return arr, nil
}

func decodeUUIDArray(raw json.RawMessage, callerID uuid.UUID) ([]uuid.UUID, error) {
	arr, err := decodeStringArray(raw)
	if err != nil {
		return nil, err
	}
	out := make([]uuid.UUID, 0, len(arr))
	for _, s := range arr {
		if s == "@me" {
			out = append(out, callerID)
			continue
		}
		id, err := uuid.Parse(s)
		if err != nil {
			return nil, fmt.Errorf("invalid UUID %q: %w", s, err)
		}
		out = append(out, id)
	}
	return out, nil
}

func decodeInt(raw json.RawMessage) (int, error) {
	var n float64
	if err := json.Unmarshal(raw, &n); err != nil {
		return 0, fmt.Errorf("expected number: %w", err)
	}
	return int(n), nil
}

func decodeIntArray(raw json.RawMessage) ([]int32, error) {
	var arr []float64
	if err := json.Unmarshal(raw, &arr); err != nil {
		return nil, fmt.Errorf("expected number array: %w", err)
	}
	out := make([]int32, len(arr))
	for i, v := range arr {
		out[i] = int32(v)
	}
	return out, nil
}

// -------- text handlers --------

// escapeLike sanitises user-supplied text for use with ILIKE `%` / `_`.
// The surrounding SQL uses `ESCAPE '\'` so `\` is the escape character.
func escapeLike(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}

// searchContains matches the term against title and description in a single
// predicate. The OR is a SQL-level implementation detail of this opcode, not
// a user-visible boolean; the DSL has no OR node.
func searchContains(a *argBuffer, _ uuid.UUID, _ time.Time, v json.RawMessage) (string, error) {
	s, err := decodeString(v)
	if err != nil {
		return "", err
	}
	p := a.push(escapeLike(s))
	return "(t.title ILIKE '%' || " + p + " || '%' ESCAPE '\\' OR t.description ILIKE '%' || " + p + " || '%' ESCAPE '\\')", nil
}

func textLike(col string, negate bool) predicateHandler {
	return func(a *argBuffer, _ uuid.UUID, _ time.Time, v json.RawMessage) (string, error) {
		s, err := decodeString(v)
		if err != nil {
			return "", err
		}
		p := a.push(escapeLike(s))
		op := " ILIKE "
		if negate {
			op = " NOT ILIKE "
		}
		return col + op + `'%' || ` + p + ` || '%'` + ` ESCAPE '\'`, nil
	}
}

func textCompare(col, op string) predicateHandler {
	return func(a *argBuffer, _ uuid.UUID, _ time.Time, v json.RawMessage) (string, error) {
		s, err := decodeString(v)
		if err != nil {
			return "", err
		}
		p := a.push(s)
		return col + " " + op + " " + p, nil
	}
}

func textEmpty(col string, empty bool) predicateHandler {
	return func(*argBuffer, uuid.UUID, time.Time, json.RawMessage) (string, error) {
		if empty {
			return "(" + col + " IS NULL OR " + col + " = '')", nil
		}
		return "(" + col + " IS NOT NULL AND " + col + " <> '')", nil
	}
}

// -------- uuid / int / string handlers --------

func uuidIn(col string, negate bool) predicateHandler {
	return func(a *argBuffer, callerID uuid.UUID, _ time.Time, v json.RawMessage) (string, error) {
		ids, err := decodeUUIDArray(v, callerID)
		if err != nil {
			return "", err
		}
		if len(ids) == 0 {
			if negate {
				return "TRUE", nil
			}
			return "FALSE", nil
		}
		p := a.push(ids)
		op := " = ANY("
		if negate {
			op = " <> ALL("
		}
		return col + op + p + "::uuid[])", nil
	}
}

func intIn(col string, negate bool) predicateHandler {
	return func(a *argBuffer, _ uuid.UUID, _ time.Time, v json.RawMessage) (string, error) {
		arr, err := decodeIntArray(v)
		if err != nil {
			return "", err
		}
		if len(arr) == 0 {
			if negate {
				return "TRUE", nil
			}
			return "FALSE", nil
		}
		p := a.push(arr)
		op := " = ANY("
		if negate {
			op = " <> ALL("
		}
		return col + op + p + "::int[])", nil
	}
}

func intCompare(col, op string) predicateHandler {
	return func(a *argBuffer, _ uuid.UUID, _ time.Time, v json.RawMessage) (string, error) {
		n, err := decodeInt(v)
		if err != nil {
			return "", err
		}
		p := a.push(int32(n))
		return col + " " + op + " " + p, nil
	}
}

func stringIn(colExpr string, negate bool) predicateHandler {
	return func(a *argBuffer, _ uuid.UUID, _ time.Time, v json.RawMessage) (string, error) {
		arr, err := decodeStringArray(v)
		if err != nil {
			return "", err
		}
		if len(arr) == 0 {
			if negate {
				return "TRUE", nil
			}
			return "FALSE", nil
		}
		p := a.push(arr)
		op := " = ANY("
		if negate {
			op = " <> ALL("
		}
		return colExpr + op + p + "::text[])", nil
	}
}

func meCompare(col, op string) predicateHandler {
	return func(a *argBuffer, callerID uuid.UUID, _ time.Time, _ json.RawMessage) (string, error) {
		p := a.push(callerID)
		return col + " " + op + " " + p, nil
	}
}

// -------- assignees / labels handlers --------

func assigneesExists(negate, _ bool) predicateHandler {
	return func(a *argBuffer, callerID uuid.UUID, _ time.Time, v json.RawMessage) (string, error) {
		ids, err := decodeUUIDArray(v, callerID)
		if err != nil {
			return "", err
		}
		if len(ids) == 0 {
			if negate {
				return "TRUE", nil
			}
			return "FALSE", nil
		}
		p := a.push(ids)
		prefix := ""
		if negate {
			prefix = "NOT "
		}
		return prefix + "EXISTS (SELECT 1 FROM task_assignees ta WHERE ta.task_id = t.id AND ta.user_id = ANY(" + p + "::uuid[]))", nil
	}
}

func assigneesHasAll(a *argBuffer, callerID uuid.UUID, _ time.Time, v json.RawMessage) (string, error) {
	ids, err := decodeUUIDArray(v, callerID)
	if err != nil {
		return "", err
	}
	if len(ids) == 0 {
		return "TRUE", nil
	}
	p := a.push(ids)
	return "(SELECT COUNT(DISTINCT ta.user_id) FROM task_assignees ta WHERE ta.task_id = t.id AND ta.user_id = ANY(" + p + "::uuid[])) = " + strconv.Itoa(len(ids)), nil
}

func assigneesPresence(set bool) predicateHandler {
	return func(*argBuffer, uuid.UUID, time.Time, json.RawMessage) (string, error) {
		prefix := "NOT "
		if set {
			prefix = ""
		}
		return prefix + "EXISTS (SELECT 1 FROM task_assignees ta WHERE ta.task_id = t.id)", nil
	}
}

func labelsExists(negate, _ bool) predicateHandler {
	return func(a *argBuffer, _ uuid.UUID, _ time.Time, v json.RawMessage) (string, error) {
		ids, err := decodeUUIDArray(v, uuid.Nil)
		if err != nil {
			return "", err
		}
		if len(ids) == 0 {
			if negate {
				return "TRUE", nil
			}
			return "FALSE", nil
		}
		p := a.push(ids)
		prefix := ""
		if negate {
			prefix = "NOT "
		}
		return prefix + "EXISTS (SELECT 1 FROM task_labels tl WHERE tl.task_id = t.id AND tl.label_id = ANY(" + p + "::uuid[]))", nil
	}
}

func labelsHasAll(a *argBuffer, _ uuid.UUID, _ time.Time, v json.RawMessage) (string, error) {
	ids, err := decodeUUIDArray(v, uuid.Nil)
	if err != nil {
		return "", err
	}
	if len(ids) == 0 {
		return "TRUE", nil
	}
	p := a.push(ids)
	return "(SELECT COUNT(DISTINCT tl.label_id) FROM task_labels tl WHERE tl.task_id = t.id AND tl.label_id = ANY(" + p + "::uuid[])) = " + strconv.Itoa(len(ids)), nil
}

func labelsPresence(set bool) predicateHandler {
	return func(*argBuffer, uuid.UUID, time.Time, json.RawMessage) (string, error) {
		prefix := "NOT "
		if set {
			prefix = ""
		}
		return prefix + "EXISTS (SELECT 1 FROM task_labels tl WHERE tl.task_id = t.id)", nil
	}
}

// -------- date handlers --------

type dateAnchor struct {
	lo time.Time // inclusive
	hi time.Time // exclusive
}

// resolveDateAnchor parses either a YYYY-MM-DD date or a relative keyword into
// a half-open [lo, hi) range. All times are returned in UTC.
func resolveDateAnchor(raw string, now time.Time) (dateAnchor, error) {
	raw = strings.TrimSpace(raw)
	now = now.UTC()
	switch raw {
	case "today":
		d := startOfDay(now)
		return dateAnchor{d, d.Add(24 * time.Hour)}, nil
	case "yesterday":
		d := startOfDay(now).Add(-24 * time.Hour)
		return dateAnchor{d, d.Add(24 * time.Hour)}, nil
	case "tomorrow":
		d := startOfDay(now).Add(24 * time.Hour)
		return dateAnchor{d, d.Add(24 * time.Hour)}, nil
	case "this_week":
		lo, hi := isoWeekRange(now, 0)
		return dateAnchor{lo, hi}, nil
	case "last_week":
		lo, hi := isoWeekRange(now, -1)
		return dateAnchor{lo, hi}, nil
	case "next_week":
		lo, hi := isoWeekRange(now, 1)
		return dateAnchor{lo, hi}, nil
	case "this_month":
		lo, hi := monthRange(now, 0)
		return dateAnchor{lo, hi}, nil
	case "last_month":
		lo, hi := monthRange(now, -1)
		return dateAnchor{lo, hi}, nil
	case "next_month":
		lo, hi := monthRange(now, 1)
		return dateAnchor{lo, hi}, nil
	case "last_7_days":
		return dateAnchor{startOfDay(now).Add(-7 * 24 * time.Hour), startOfDay(now).Add(24 * time.Hour)}, nil
	case "last_30_days":
		return dateAnchor{startOfDay(now).Add(-30 * 24 * time.Hour), startOfDay(now).Add(24 * time.Hour)}, nil
	case "last_90_days":
		return dateAnchor{startOfDay(now).Add(-90 * 24 * time.Hour), startOfDay(now).Add(24 * time.Hour)}, nil
	}
	t, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return dateAnchor{}, fmt.Errorf("invalid date value %q", raw)
	}
	return dateAnchor{t.UTC(), t.UTC().Add(24 * time.Hour)}, nil
}

func startOfDay(t time.Time) time.Time {
	t = t.UTC()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

func isoWeekRange(now time.Time, offsetWeeks int) (time.Time, time.Time) {
	d := startOfDay(now)
	weekday := int(d.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	monday := d.AddDate(0, 0, -(weekday - 1))
	monday = monday.AddDate(0, 0, offsetWeeks*7)
	return monday, monday.AddDate(0, 0, 7)
}

func monthRange(now time.Time, offsetMonths int) (time.Time, time.Time) {
	d := startOfDay(now)
	first := time.Date(d.Year(), d.Month(), 1, 0, 0, 0, 0, time.UTC)
	first = first.AddDate(0, offsetMonths, 0)
	return first, first.AddDate(0, 1, 0)
}

// dateOp builds `before` or `after` predicates against a column.
//   - before anchor: col < anchor.lo  (the anchor's earliest moment)
//   - after anchor:  col >= anchor.hi (strictly past the anchor's last moment)
func dateOp(col, kind string) predicateHandler {
	return func(a *argBuffer, _ uuid.UUID, now time.Time, v json.RawMessage) (string, error) {
		s, err := decodeString(v)
		if err != nil {
			return "", err
		}
		anchor, err := resolveDateAnchor(s, now)
		if err != nil {
			return "", err
		}
		switch kind {
		case "before":
			p := a.push(anchor.lo)
			return col + " < " + p, nil
		case "after":
			p := a.push(anchor.hi)
			return col + " >= " + p, nil
		default:
			return "", fmt.Errorf("internal: unknown date op %q", kind)
		}
	}
}

func dateBetween(col string) predicateHandler {
	return func(a *argBuffer, _ uuid.UUID, now time.Time, v json.RawMessage) (string, error) {
		var payload struct {
			From string `json:"from"`
			To   string `json:"to"`
		}
		if err := json.Unmarshal(v, &payload); err != nil {
			return "", fmt.Errorf("between expects {from,to}: %w", err)
		}
		from, err := resolveDateAnchor(payload.From, now)
		if err != nil {
			return "", err
		}
		to, err := resolveDateAnchor(payload.To, now)
		if err != nil {
			return "", err
		}
		pLo := a.push(from.lo)
		pHi := a.push(to.hi)
		return col + " >= " + pLo + " AND " + col + " < " + pHi, nil
	}
}

func dueDateOverdue(a *argBuffer, _ uuid.UUID, now time.Time, _ json.RawMessage) (string, error) {
	p := a.push(now.UTC())
	return "t.due_date < " + p + " AND ps.state_type NOT IN ('completed', 'cancelled')", nil
}

func nullCheck(col string, isNull bool) predicateHandler {
	return func(*argBuffer, uuid.UUID, time.Time, json.RawMessage) (string, error) {
		if isNull {
			return col + " IS NULL", nil
		}
		return col + " IS NOT NULL", nil
	}
}

// -------- comment_count handler --------

func commentCount(op string) predicateHandler {
	return func(a *argBuffer, _ uuid.UUID, _ time.Time, v json.RawMessage) (string, error) {
		n, err := decodeInt(v)
		if err != nil {
			return "", err
		}
		p := a.push(int64(n))
		return "(SELECT COUNT(*) FROM comments c WHERE c.task_id = t.id AND c.deleted_at IS NULL) " + op + " " + p, nil
	}
}

// =========================================================================
// FilterRunner — runs compiled filter trees against the database via pgx.
// =========================================================================

// FilteredTaskRow matches the columns previously returned by
// ListProjectTasksFiltered. The handler maps this into TaskResponse.
type FilteredTaskRow struct {
	ID               uuid.UUID
	ProjectID        uuid.UUID
	TaskNumber       int32
	Title            string
	Description      pgtype.Text
	StateID          uuid.UUID
	Priority         int32
	CreatedBy        uuid.UUID
	StartDate        pgtype.Timestamptz
	DueDate          pgtype.Timestamptz
	CreatedAt        pgtype.Timestamptz
	UpdatedAt        pgtype.Timestamptz
	DeletedAt        pgtype.Timestamptz
	ProjectKey       string
	StateName        string
	StateType        string
	StateColor       pgtype.Text
	CreatorUsername  string
	CreatorFirstName string
	CreatorLastName  string
	CreatorAvatarUrl pgtype.Text
	CommentCount     int64
}

// FilterListParams describes a paginated task list request.
type FilterListParams struct {
	ProjectID uuid.UUID
	CallerID  uuid.UUID
	Tree      FilterTree
	SortBy    string
	SortDir   string
	Limit     int32
	Offset    int32
	Now       time.Time // override for tests; zero means time.Now().UTC()
}

// FilterRunner executes filter trees against a live database.
type FilterRunner struct {
	pool *pgxpool.Pool
}

// NewFilterRunner constructs a FilterRunner against the given pool.
func NewFilterRunner(pool *pgxpool.Pool) *FilterRunner {
	return &FilterRunner{pool: pool}
}

const filterSelectBase = `SELECT t.id, t.project_id, t.task_number, t.title, t.description, t.state_id, t.priority, t.created_by, t.start_date, t.due_date, t.created_at, t.updated_at, t.deleted_at,
       p.project_key,
       ps.name as state_name, ps.state_type, ps.color as state_color,
       u.username as creator_username, u.first_name as creator_first_name, u.last_name as creator_last_name, u.avatar_url as creator_avatar_url,
       (SELECT COUNT(*) FROM comments c WHERE c.task_id = t.id AND c.deleted_at IS NULL)::bigint as comment_count
FROM tasks t
JOIN projects p ON t.project_id = p.id
JOIN project_states ps ON t.state_id = ps.id
JOIN users u ON t.created_by = u.id
WHERE t.project_id = $1
  AND t.deleted_at IS NULL`

const filterCountBase = `SELECT COUNT(*)
FROM tasks t
JOIN projects p ON t.project_id = p.id
JOIN project_states ps ON t.state_id = ps.id
JOIN users u ON t.created_by = u.id
WHERE t.project_id = $1
  AND t.deleted_at IS NULL`

// ListTasks compiles the filter tree, runs the dynamic query, and returns rows.
func (r *FilterRunner) ListTasks(ctx context.Context, p FilterListParams) ([]FilteredTaskRow, error) {
	now := p.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}
	// Reserve $1 for project_id; compiled filter args start at $2.
	compiled, err := Compile(p.Tree, p.CallerID, now, 1)
	if err != nil {
		return nil, err
	}

	sortCol, ok := validSortKeys[p.SortBy]
	if !ok {
		sortCol = "t.created_at"
	}
	sortDir, ok := validSortDirs[p.SortDir]
	if !ok {
		sortDir = "DESC"
	}
	nullsClause := ""
	if sortCol == "t.due_date" || sortCol == "t.start_date" {
		nullsClause = " NULLS LAST"
	}

	nextArg := 1 + len(compiled.Args)
	limitPH := "$" + strconv.Itoa(nextArg+1)
	offsetPH := "$" + strconv.Itoa(nextArg+2)

	// Stable tiebreaker by task_id to keep pagination deterministic.
	sqlStr := filterSelectBase + compiled.WhereSQL +
		" ORDER BY " + sortCol + " " + sortDir + nullsClause + ", t.id ASC" +
		" LIMIT " + limitPH + " OFFSET " + offsetPH

	args := make([]any, 0, 1+len(compiled.Args)+2)
	args = append(args, p.ProjectID)
	args = append(args, compiled.Args...)
	args = append(args, p.Limit, p.Offset)

	rows, err := r.pool.Query(ctx, sqlStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]FilteredTaskRow, 0)
	for rows.Next() {
		var i FilteredTaskRow
		if err := rows.Scan(
			&i.ID, &i.ProjectID, &i.TaskNumber, &i.Title, &i.Description,
			&i.StateID, &i.Priority, &i.CreatedBy, &i.StartDate, &i.DueDate,
			&i.CreatedAt, &i.UpdatedAt, &i.DeletedAt, &i.ProjectKey,
			&i.StateName, &i.StateType, &i.StateColor,
			&i.CreatorUsername, &i.CreatorFirstName, &i.CreatorLastName, &i.CreatorAvatarUrl,
			&i.CommentCount,
		); err != nil {
			return nil, err
		}
		out = append(out, i)
	}
	return out, rows.Err()
}

// CountTasks runs the filter tree as a COUNT(*) query.
func (r *FilterRunner) CountTasks(ctx context.Context, p FilterListParams) (int64, error) {
	now := p.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}
	compiled, err := Compile(p.Tree, p.CallerID, now, 1)
	if err != nil {
		return 0, err
	}
	sqlStr := filterCountBase + compiled.WhereSQL
	args := make([]any, 0, 1+len(compiled.Args))
	args = append(args, p.ProjectID)
	args = append(args, compiled.Args...)
	var total int64
	if err := r.pool.QueryRow(ctx, sqlStr, args...).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}
