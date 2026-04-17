import type { TaskAssignee } from "./task";

export type CycleStatus = "upcoming" | "active" | "completed";

export interface Cycle {
  id: string;
  project_id: string;
  title: string;
  description?: string;
  start_date: string; // YYYY-MM-DD
  end_date: string;   // YYYY-MM-DD
  status: CycleStatus;
  created_by: string;
  created_at: string;
  updated_at: string;
  total_tasks: number;
  completed_tasks: number;
  project_key?: string;
  project_name?: string;
}

export interface PaginatedCyclesResponse {
  cycles: Cycle[];
  total: number;
  page: number;
  per_page: number;
  total_pages: number;
}

export interface CycleTask {
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

export interface CycleStateBucket {
  state_id: string;
  state_name: string;
  state_color: string;
  state_type: string;
  count: number;
}

export interface CycleAssigneeSummary {
  user_id: string;
  username: string;
  first_name: string;
  last_name: string;
  avatar_url?: string;
  task_count: number;
}

export interface CycleMetrics {
  total: number;
  completed: number;
  in_progress: number;
  todo: number;
  cancelled: number;
  state_breakdown: CycleStateBucket[];
  assignees: CycleAssigneeSummary[];
}

export interface CreateCycleRequest {
  title: string;
  description?: string;
  start_date: string;
  end_date: string;
}

export interface UpdateCycleRequest {
  title?: string;
  description?: string;
  start_date?: string;
  end_date?: string;
}

export interface CycleSibling {
  id: string;
  title: string;
  start_date: string;
  end_date: string;
  status: CycleStatus;
}
