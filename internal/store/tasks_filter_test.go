package store

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func mustRaw(t *testing.T, v any) json.RawMessage {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return b
}

func compileOrFail(t *testing.T, tree FilterTree, caller uuid.UUID) CompiledFilter {
	t.Helper()
	c, err := Compile(tree, caller, time.Date(2026, 4, 16, 12, 0, 0, 0, time.UTC), 1)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	return c
}

func TestCompileEmptyTree(t *testing.T) {
	c, err := Compile(FilterTree{}, uuid.Nil, time.Now(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if c.WhereSQL != "" || len(c.Args) != 0 {
		t.Fatalf("empty tree should yield empty output, got %+v", c)
	}
}

func TestCompileSinglePredicate(t *testing.T) {
	tree := FilterTree{Children: []FilterNode{
		{Predicate: &Predicate{Field: "title", Op: "contains", Value: mustRaw(t, "bug")}},
	}}
	c := compileOrFail(t, tree, uuid.Nil)
	if !strings.HasPrefix(c.WhereSQL, " AND t.title ILIKE ") {
		t.Fatalf("unexpected sql: %q", c.WhereSQL)
	}
	if len(c.Args) != 1 {
		t.Fatalf("want 1 arg, got %d", len(c.Args))
	}
	if c.Args[0] != "bug" {
		t.Fatalf("want arg %q, got %v", "bug", c.Args[0])
	}
}

func TestCompileMultiplePredicatesAreAnded(t *testing.T) {
	tree := FilterTree{Children: []FilterNode{
		{Predicate: &Predicate{Field: "state_type", Op: "in", Value: mustRaw(t, []string{"started"})}},
		{Predicate: &Predicate{Field: "priority", Op: "gte", Value: mustRaw(t, 3)}},
	}}
	c := compileOrFail(t, tree, uuid.Nil)
	if strings.Contains(c.WhereSQL, " OR ") {
		t.Fatalf("filters must never contain top-level OR, got %q", c.WhereSQL)
	}
	if !strings.Contains(c.WhereSQL, " AND ") {
		t.Fatalf("expected AND between predicates, got %q", c.WhereSQL)
	}
}

func TestCompileSearchPredicate(t *testing.T) {
	tree := FilterTree{Children: []FilterNode{
		{Predicate: &Predicate{Field: "search", Op: "contains", Value: mustRaw(t, "bug")}},
	}}
	c := compileOrFail(t, tree, uuid.Nil)
	// The opcode internally matches both columns via a single parenthesised
	// ILIKE … OR ILIKE … expression — the DSL itself has no OR node.
	if !strings.Contains(c.WhereSQL, "t.title ILIKE") || !strings.Contains(c.WhereSQL, "t.description ILIKE") {
		t.Fatalf("search must hit both title and description: %q", c.WhereSQL)
	}
}

func TestCompileAtMeBinding(t *testing.T) {
	caller := uuid.New()
	tree := FilterTree{Children: []FilterNode{
		{Predicate: &Predicate{Field: "assignees", Op: "has_any", Value: mustRaw(t, []string{"@me"})}},
	}}
	c := compileOrFail(t, tree, caller)
	// The @me alias should be substituted into the arg slice.
	found := false
	for _, a := range c.Args {
		if ids, ok := a.([]uuid.UUID); ok {
			for _, id := range ids {
				if id == caller {
					found = true
				}
			}
		}
	}
	if !found {
		t.Fatalf("expected caller UUID in args: %+v", c.Args)
	}
}

func TestCompileRejectsUnknownField(t *testing.T) {
	tree := FilterTree{Children: []FilterNode{
		{Predicate: &Predicate{Field: "nope", Op: "is", Value: mustRaw(t, "x")}},
	}}
	_, err := Compile(tree, uuid.Nil, time.Now(), 1)
	if err == nil || !strings.Contains(err.Error(), "unsupported") {
		t.Fatalf("want unsupported error, got %v", err)
	}
}

func TestCompileRejectsUnknownOp(t *testing.T) {
	tree := FilterTree{Children: []FilterNode{
		{Predicate: &Predicate{Field: "title", Op: "matches_regex", Value: mustRaw(t, "x")}},
	}}
	_, err := Compile(tree, uuid.Nil, time.Now(), 1)
	if err == nil {
		t.Fatal("expected error for unknown op")
	}
}

func TestCompilePredicateLimit(t *testing.T) {
	children := make([]FilterNode, 0, MaxFilterPredicates+1)
	for i := 0; i < MaxFilterPredicates+1; i++ {
		children = append(children, FilterNode{
			Predicate: &Predicate{Field: "title", Op: "contains", Value: mustRaw(t, "x")},
		})
	}
	_, err := Compile(FilterTree{Children: children}, uuid.Nil, time.Now(), 1)
	if err == nil || !strings.Contains(err.Error(), "too many") {
		t.Fatalf("want too many predicates error, got %v", err)
	}
}

func TestDateAnchorKeywords(t *testing.T) {
	// Wed, Apr 15 2026 in UTC
	now := time.Date(2026, 4, 15, 10, 0, 0, 0, time.UTC)

	// today
	a, err := resolveDateAnchor("today", now)
	if err != nil {
		t.Fatal(err)
	}
	if !a.lo.Equal(time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("today lo: %v", a.lo)
	}
	if !a.hi.Equal(time.Date(2026, 4, 16, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("today hi: %v", a.hi)
	}

	// this_week: Apr 13 is a Monday
	a, err = resolveDateAnchor("this_week", now)
	if err != nil {
		t.Fatal(err)
	}
	if !a.lo.Equal(time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("this_week lo: %v", a.lo)
	}
	if !a.hi.Equal(time.Date(2026, 4, 20, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("this_week hi: %v", a.hi)
	}

	// absolute ISO date
	a, err = resolveDateAnchor("2026-06-01", now)
	if err != nil {
		t.Fatal(err)
	}
	if !a.lo.Equal(time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("iso lo: %v", a.lo)
	}
}

func TestCompileOverdue(t *testing.T) {
	tree := FilterTree{Children: []FilterNode{
		{Predicate: &Predicate{Field: "due_date", Op: "overdue"}},
	}}
	c := compileOrFail(t, tree, uuid.Nil)
	if !strings.Contains(c.WhereSQL, "ps.state_type NOT IN ('completed', 'cancelled')") {
		t.Fatalf("overdue must exclude completed/cancelled: %q", c.WhereSQL)
	}
}

func TestCompileEmptyInSetReturnsFalse(t *testing.T) {
	// Empty IN list → FALSE (match nothing). Users sending empty arrays get no rows.
	tree := FilterTree{Children: []FilterNode{
		{Predicate: &Predicate{Field: "priority", Op: "in", Value: mustRaw(t, []int{})}},
	}}
	c := compileOrFail(t, tree, uuid.Nil)
	if !strings.Contains(c.WhereSQL, "FALSE") {
		t.Fatalf("empty IN should compile to FALSE, got %q", c.WhereSQL)
	}
}

func TestParseFilterTreeRespectsSizeLimit(t *testing.T) {
	big := make([]byte, MaxFilterJSONBytes+10)
	for i := range big {
		big[i] = ' '
	}
	_, err := ParseFilterTree(big)
	if err == nil || !strings.Contains(err.Error(), "exceeds") {
		t.Fatalf("want size error, got %v", err)
	}
}

func TestCompileRejectsEmptyFilterNode(t *testing.T) {
	tree := FilterTree{Children: []FilterNode{{}}}
	_, err := Compile(tree, uuid.Nil, time.Now(), 1)
	if err == nil || !strings.Contains(err.Error(), "predicate") {
		t.Fatalf("want predicate-required error, got %v", err)
	}
}
