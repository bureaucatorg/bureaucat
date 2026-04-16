export interface Task {
  id: string;
  project_key: string;
  task_number: number;
  task_id: string; // e.g., "DEVOP-123"
  title: string;
  description?: string;
  state_id: string;
  state_name: string;
  state_type: string;
  state_color: string;
  priority: number;
  start_date?: string;
  due_date?: string;
  created_by: string;
  creator_username: string;
  creator_first_name: string;
  creator_last_name: string;
  creator_avatar_url?: string;
  assignees?: TaskAssignee[];
  labels?: TaskLabel[];
  comment_count: number;
  created_at: string;
  updated_at: string;
}

export interface PaginatedTasksResponse {
  tasks: Task[];
  total: number;
  page: number;
  per_page: number;
  total_pages: number;
}

export interface TaskAssignee {
  id: string;
  user_id: string;
  username: string;
  email: string;
  first_name: string;
  last_name: string;
  avatar_url?: string;
}

export interface TaskLabel {
  id: string;
  name: string;
  color: string;
}

export interface CreateTaskRequest {
  title: string;
  description?: string;
  state_id?: string;
  priority?: number;
  start_date?: string;
  due_date?: string;
  assignees?: string[];
  labels?: string[];
}

export interface UpdateTaskRequest {
  title?: string;
  description?: string;
  state_id?: string;
  priority?: number;
  // Use `null` to clear; omit to leave unchanged.
  start_date?: string | null;
  due_date?: string | null;
}

/** @deprecated — retained so legacy URL migration can parse old bookmarks. */
export interface TaskFilters {
  state_id?: string;
  state_type?: string;
  created_by?: string;
  assigned_to?: string;
  priority?: number;
  q?: string;
  from_date?: string;
  to_date?: string;
}

// ================ FilterTree DSL ================
// Every child is a single Predicate. All predicates are joined with AND;
// the DSL has no OR opcode. `search` is a special internal field emitted
// by the search box and matches title+description in one SQL predicate.

export type FilterField =
  | "search"
  | "title"
  | "description"
  | "state"
  | "state_type"
  | "priority"
  | "assignees"
  | "created_by"
  | "labels"
  | "start_date"
  | "due_date"
  | "created_at"
  | "updated_at"
  | "comment_count";

export type FilterOp =
  | "contains"
  | "not_contains"
  | "is"
  | "is_not"
  | "in"
  | "not_in"
  | "has_any"
  | "has_all"
  | "has_none"
  | "is_empty"
  | "is_set"
  | "before"
  | "after"
  | "between"
  | "overdue"
  | "is_me"
  | "is_not_me"
  | "eq"
  | "ne"
  | "gt"
  | "gte"
  | "lt"
  | "lte";

/**
 * FilterValue is a generic JSON value whose shape depends on the (field, op).
 * - Text ops: string
 * - Number ops: number
 * - *in / has_*: string[] (UUIDs) or number[] (ints)
 * - between: { from: string; to: string } where strings are ISO dates or relative keywords
 * - before / after: string (ISO date or relative keyword)
 * - is_me / is_not_me / is_empty / is_set / overdue: no value (omit or undefined)
 *
 * The literal string "@me" in any id-array means "the currently-signed-in user".
 */
export type FilterValue =
  | string
  | number
  | string[]
  | number[]
  | { from: string; to: string };

export interface Predicate {
  field: FilterField;
  op: FilterOp;
  value?: FilterValue;
}

export interface FilterNode {
  predicate?: Predicate;
}

export interface FilterTree {
  children: FilterNode[];
}

// ================ View grouping and sorting ================

export type ViewGroupBy =
  | "state"
  | "state_type"
  | "priority"
  | "assignee"
  | "label"
  | "due_bucket";

export type SortKey =
  | "created_at"
  | "updated_at"
  | "priority"
  | "due_date"
  | "start_date"
  | "title";

export type SortDir = "asc" | "desc";

// Relative date keywords accepted by date predicates. Evaluated server-side.
export const RELATIVE_DATE_KEYWORDS = [
  "today",
  "yesterday",
  "tomorrow",
  "this_week",
  "last_week",
  "next_week",
  "this_month",
  "last_month",
  "next_month",
  "last_7_days",
  "last_30_days",
  "last_90_days",
] as const;

export type RelativeDateKeyword = (typeof RELATIVE_DATE_KEYWORDS)[number];

export const PRIORITY_LABELS: Record<number, { label: string; color: string }> = {
  0: { label: "No priority", color: "#6B7280" },
  1: { label: "Low", color: "#3B82F6" },
  2: { label: "Medium", color: "#EAB308" },
  3: { label: "High", color: "#F97316" },
  4: { label: "Urgent", color: "#EF4444" },
};

export const STATE_TYPE_COLORS: Record<string, string> = {
  backlog: "#6B7280",
  unstarted: "#3B82F6",
  started: "#10B981",
  completed: "#22C55E",
  cancelled: "#9CA3AF",
};
