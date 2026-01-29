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
  created_by: string;
  creator_username: string;
  assignees?: TaskAssignee[];
  labels?: TaskLabel[];
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
  assignees?: string[];
  labels?: string[];
}

export interface UpdateTaskRequest {
  title?: string;
  description?: string;
  state_id?: string;
  priority?: number;
}

export interface TaskFilters {
  state_id?: string;
  state_type?: string;
  created_by?: string;
  priority?: number;
  q?: string;
}

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
