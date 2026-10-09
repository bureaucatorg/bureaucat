export interface ActivityLogEntry {
  id: string;
  task_id: string;
  activity_type: ActivityType;
  actor_id: string;
  username: string;
  first_name: string;
  last_name: string;
  avatar_url?: string;
  field_name?: string;
  old_value?: unknown;
  new_value?: unknown;
  created_at: string;
}

export type ActivityType =
  | "task_created"
  | "task_updated"
  | "task_deleted"
  | "assignee_added"
  | "assignee_removed"
  | "label_added"
  | "label_removed"
  | "state_changed"
  | "comment_created"
  | "comment_updated"
  | "comment_deleted"
  | "blocker_added"
  | "blocker_removed";

export const ACTIVITY_TYPE_LABELS: Record<ActivityType, string> = {
  task_created: "created the task",
  task_updated: "updated",
  task_deleted: "deleted the task",
  assignee_added: "added assignee",
  assignee_removed: "removed assignee",
  label_added: "added label",
  label_removed: "removed label",
  state_changed: "changed state",
  comment_created: "added a comment",
  comment_updated: "edited a comment",
  comment_deleted: "deleted a comment",
  blocker_added: "linked a blocker",
  blocker_removed: "unlinked a blocker",
};

// field_name holds the relation as seen from the task the entry belongs to.
const BLOCKER_LABELS: Record<string, string> = {
  "blocker_added:blocked_by": "marked as blocked by",
  "blocker_added:blocking": "marked as blocking",
  "blocker_removed:blocked_by": "removed blocker",
  "blocker_removed:blocking": "removed blocked task",
};

export function activityLabel(entry: ActivityLogEntry): string {
  return (
    BLOCKER_LABELS[`${entry.activity_type}:${entry.field_name}`] ||
    ACTIVITY_TYPE_LABELS[entry.activity_type] ||
    entry.activity_type
  );
}

export interface VerifyActivityResponse {
  valid: boolean;
  message: string;
}

export interface UserActivityEntry extends ActivityLogEntry {
  task_number: number;
  project_key: string;
  task_title: string;
}

// A persisted, per-user notification. Extends the activity shape with batching
// (event_count) and read/unread state (is_read).
export interface NotificationEntry extends UserActivityEntry {
  event_count: number;
  is_read: boolean;
  updated_at: string;
  comment_id?: string;
}

export interface NotificationListResponse {
  activities: NotificationEntry[];
  unread_count: number;
  total: number;
  page: number;
  per_page: number;
  total_pages: number;
}

export interface UserActivityDateCount {
  date: string;
  count: number;
}
