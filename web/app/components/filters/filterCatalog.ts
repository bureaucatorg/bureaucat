/**
 * Filter catalog — closed list of fields, the operators they accept, and the
 * kind of value each (field, op) expects. Shared by the predicate editor and
 * the chip summariser so both stay in lock-step with the backend predicate
 * handlers in internal/store/tasks_filter.go.
 */

import type { Component } from "vue";
import {
  Type,
  FileText,
  Tag,
  Layers,
  Flame,
  Users,
  User,
  Calendar as CalendarIcon,
  MessageSquare,
  Clock,
} from "lucide-vue-next";
import type { FilterField, FilterOp } from "~/types";

export type ValueKind =
  | "text"
  | "uuid-array"
  | "int-array"
  | "string-array"
  | "date"
  | "date-range"
  | "number"
  | "none";

export type EntityKind = "state" | "state_type" | "priority" | "member" | "label";

export interface OpDef {
  op: FilterOp;
  label: string;
  valueKind: ValueKind;
}

export interface FieldDef {
  field: FilterField;
  label: string;
  icon: Component;
  entityKind?: EntityKind;
  ops: OpDef[];
}

export const FILTER_CATALOG: FieldDef[] = [
  {
    field: "title",
    label: "Title",
    icon: Type,
    ops: [
      { op: "contains", label: "contains", valueKind: "text" },
      { op: "not_contains", label: "does not contain", valueKind: "text" },
      { op: "is", label: "is exactly", valueKind: "text" },
      { op: "is_not", label: "is not", valueKind: "text" },
      { op: "is_empty", label: "is empty", valueKind: "none" },
      { op: "is_set", label: "is set", valueKind: "none" },
    ],
  },
  {
    field: "description",
    label: "Description",
    icon: FileText,
    ops: [
      { op: "contains", label: "contains", valueKind: "text" },
      { op: "not_contains", label: "does not contain", valueKind: "text" },
      { op: "is_empty", label: "is empty", valueKind: "none" },
      { op: "is_set", label: "is set", valueKind: "none" },
    ],
  },
  {
    field: "state",
    label: "State",
    icon: Tag,
    entityKind: "state",
    ops: [
      { op: "in", label: "is any of", valueKind: "uuid-array" },
      { op: "not_in", label: "is none of", valueKind: "uuid-array" },
    ],
  },
  {
    field: "state_type",
    label: "Status category",
    icon: Layers,
    entityKind: "state_type",
    ops: [
      { op: "in", label: "is any of", valueKind: "string-array" },
      { op: "not_in", label: "is none of", valueKind: "string-array" },
    ],
  },
  {
    field: "priority",
    label: "Priority",
    icon: Flame,
    entityKind: "priority",
    ops: [
      { op: "in", label: "is any of", valueKind: "int-array" },
      { op: "not_in", label: "is none of", valueKind: "int-array" },
      { op: "gte", label: "is at least", valueKind: "number" },
      { op: "lte", label: "is at most", valueKind: "number" },
    ],
  },
  {
    field: "assignees",
    label: "Assignees",
    icon: Users,
    entityKind: "member",
    ops: [
      { op: "has_any", label: "include any of", valueKind: "uuid-array" },
      { op: "has_all", label: "include all of", valueKind: "uuid-array" },
      { op: "has_none", label: "exclude all of", valueKind: "uuid-array" },
      { op: "is_empty", label: "is unassigned", valueKind: "none" },
      { op: "is_set", label: "has any assignee", valueKind: "none" },
    ],
  },
  {
    field: "created_by",
    label: "Created by",
    icon: User,
    entityKind: "member",
    ops: [
      { op: "in", label: "is any of", valueKind: "uuid-array" },
      { op: "not_in", label: "is none of", valueKind: "uuid-array" },
      { op: "is_me", label: "is me", valueKind: "none" },
      { op: "is_not_me", label: "is not me", valueKind: "none" },
    ],
  },
  {
    field: "labels",
    label: "Labels",
    icon: Tag,
    entityKind: "label",
    ops: [
      { op: "has_any", label: "include any of", valueKind: "uuid-array" },
      { op: "has_all", label: "include all of", valueKind: "uuid-array" },
      { op: "has_none", label: "exclude all of", valueKind: "uuid-array" },
      { op: "is_empty", label: "has no labels", valueKind: "none" },
      { op: "is_set", label: "has any label", valueKind: "none" },
    ],
  },
  {
    field: "start_date",
    label: "Start date",
    icon: CalendarIcon,
    ops: [
      { op: "before", label: "is before", valueKind: "date" },
      { op: "after", label: "is after", valueKind: "date" },
      { op: "between", label: "is between", valueKind: "date-range" },
      { op: "is_empty", label: "is unset", valueKind: "none" },
      { op: "is_set", label: "is set", valueKind: "none" },
    ],
  },
  {
    field: "due_date",
    label: "Due date",
    icon: CalendarIcon,
    ops: [
      { op: "before", label: "is before", valueKind: "date" },
      { op: "after", label: "is after", valueKind: "date" },
      { op: "between", label: "is between", valueKind: "date-range" },
      { op: "overdue", label: "is overdue", valueKind: "none" },
      { op: "is_empty", label: "is unset", valueKind: "none" },
      { op: "is_set", label: "is set", valueKind: "none" },
    ],
  },
  {
    field: "created_at",
    label: "Created at",
    icon: Clock,
    ops: [
      { op: "before", label: "is before", valueKind: "date" },
      { op: "after", label: "is after", valueKind: "date" },
      { op: "between", label: "is between", valueKind: "date-range" },
    ],
  },
  {
    field: "updated_at",
    label: "Updated at",
    icon: Clock,
    ops: [
      { op: "before", label: "is before", valueKind: "date" },
      { op: "after", label: "is after", valueKind: "date" },
      { op: "between", label: "is between", valueKind: "date-range" },
    ],
  },
  {
    field: "comment_count",
    label: "Comments",
    icon: MessageSquare,
    ops: [
      { op: "eq", label: "=", valueKind: "number" },
      { op: "ne", label: "≠", valueKind: "number" },
      { op: "gt", label: ">", valueKind: "number" },
      { op: "gte", label: "≥", valueKind: "number" },
      { op: "lt", label: "<", valueKind: "number" },
      { op: "lte", label: "≤", valueKind: "number" },
    ],
  },
];

export function findFieldDef(field: FilterField): FieldDef | undefined {
  return FILTER_CATALOG.find((f) => f.field === field);
}

export function findOpDef(field: FilterField, op: FilterOp): OpDef | undefined {
  return findFieldDef(field)?.ops.find((o) => o.op === op);
}

/** Return a friendly label for a state_type enum value. */
export function stateTypeLabel(kind: string): string {
  switch (kind) {
    case "backlog":
      return "Backlog";
    case "unstarted":
      return "To Do";
    case "started":
      return "In Progress";
    case "completed":
      return "Done";
    case "cancelled":
      return "Cancelled";
    default:
      return kind;
  }
}

export const STATE_TYPE_OPTIONS = [
  { id: "backlog", label: "Backlog" },
  { id: "unstarted", label: "To Do" },
  { id: "started", label: "In Progress" },
  { id: "completed", label: "Done" },
  { id: "cancelled", label: "Cancelled" },
];

export const PRIORITY_OPTIONS = [
  { id: "4", label: "Urgent", color: "#EF4444" },
  { id: "3", label: "High", color: "#F97316" },
  { id: "2", label: "Medium", color: "#EAB308" },
  { id: "1", label: "Low", color: "#3B82F6" },
  { id: "0", label: "No priority", color: "#6B7280" },
];

export const RELATIVE_DATE_OPTIONS = [
  { id: "today", label: "Today" },
  { id: "yesterday", label: "Yesterday" },
  { id: "tomorrow", label: "Tomorrow" },
  { id: "this_week", label: "This week" },
  { id: "last_week", label: "Last week" },
  { id: "next_week", label: "Next week" },
  { id: "this_month", label: "This month" },
  { id: "last_month", label: "Last month" },
  { id: "next_month", label: "Next month" },
  { id: "last_7_days", label: "Last 7 days" },
  { id: "last_30_days", label: "Last 30 days" },
  { id: "last_90_days", label: "Last 90 days" },
];
