export interface ActivityLogEntry {
  id: string;
  task_id: string;
  activity_type: ActivityType;
  actor_id: string;
  username: string;
  first_name: string;
  last_name: string;
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
  assignee_added: "added an assignee",
  assignee_removed: "removed an assignee",
  label_added: "added a label",
  label_removed: "removed a label",
  state_changed: "changed the state",
  comment_created: "added a comment",
  comment_updated: "edited a comment",
  comment_deleted: "deleted a comment",
};

export interface VerifyActivityResponse {
  valid: boolean;
  message: string;
}
