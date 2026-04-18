import type { TaskAssignee } from "./task";

export type ModuleStatus =
  | "backlog"
  | "planned"
  | "in_progress"
  | "paused"
  | "completed"
  | "cancelled";

export const MODULE_STATUSES: ModuleStatus[] = [
  "backlog",
  "planned",
  "in_progress",
  "paused",
  "completed",
  "cancelled",
];

export interface ModuleUserBrief {
  user_id: string;
  username: string;
  email?: string;
  first_name: string;
  last_name: string;
  avatar_url?: string;
}

export interface Module {
  id: string;
  project_id: string;
  title: string;
  description?: string;
  status: ModuleStatus;
  start_date?: string;
  end_date?: string;
  lead?: ModuleUserBrief;
  members: ModuleUserBrief[];
  created_by: string;
  created_at: string;
  updated_at: string;
  total_tasks: number;
  completed_tasks: number;
  project_key?: string;
  project_name?: string;
}

export interface PaginatedModulesResponse {
  modules: Module[];
  total: number;
  page: number;
  per_page: number;
  total_pages: number;
}

export interface ModuleTask {
  id: string;
  project_key: string;
  task_number: number;
  task_id: string;
  title: string;
  state_id: string;
  state_name: string;
  state_type: string;
  state_color: string;
  priority: number;
  start_date?: string;
  due_date?: string;
  assignees: TaskAssignee[];
}

export interface ModuleStateBucket {
  state_id: string;
  state_name: string;
  state_color: string;
  state_type: string;
  count: number;
}

export interface ModuleMetrics {
  total: number;
  completed: number;
  in_progress: number;
  todo: number;
  cancelled: number;
  state_breakdown: ModuleStateBucket[];
}

export interface CreateModuleRequest {
  title: string;
  description?: string;
  status?: ModuleStatus;
  start_date?: string;
  end_date?: string;
  lead_id?: string;
  member_ids?: string[];
}

export interface UpdateModuleRequest {
  title?: string;
  description?: string;
  status?: ModuleStatus;
  start_date?: string;
  end_date?: string;
  lead_id?: string;
  clear_start_date?: boolean;
  clear_end_date?: boolean;
  clear_lead?: boolean;
}

export interface DuplicateModuleRequest {
  title: string;
  start_date?: string;
  end_date?: string;
  task_ids: string[];
}

export interface ModuleListFilters {
  status?: ModuleStatus;
  lead_id?: string;
  start_after?: string;
  end_before?: string;
  sort_by?: "created_at" | "end_date" | "progress";
  sort_dir?: "asc" | "desc";
}
