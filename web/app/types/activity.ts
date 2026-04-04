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
  | "comment_deleted";

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
};

export interface VerifyActivityResponse {
  valid: boolean;
  message: string;
}

export interface UserActivityEntry extends ActivityLogEntry {
  task_number: number;
  project_key: string;
  task_title: string;
}

export interface UserActivityDateCount {
  date: string;
  count: number;
}
