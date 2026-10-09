export interface GraphUser {
  id: string;
  username: string;
  email: string;
  first_name: string;
  last_name: string;
  avatar_url?: string;
}

export interface GraphTask {
  id: string;
  project_key: string;
  project_name: string;
  task_number: number;
  title: string;
  is_subtask: boolean;
  parent_id?: string;
  workspace_id: string;
  workspace_key: string;
  workspace_name: string;
  state_name: string;
  state_type: string;
  state_color?: string;
}

export interface GraphEdge {
  user_id: string;
  task_id: string;
}

export interface GraphBlockerEdge {
  blocker_id: string;
  blocked_id: string;
}

export interface TaskGraph {
  users: GraphUser[];
  tasks: GraphTask[];
  edges: GraphEdge[];
  blocker_edges: GraphBlockerEdge[];
}

export interface TaskGraphFilterOptions {
  workspaces: { key: string; name: string }[];
  projects: { key: string; name: string; workspace_key: string }[];
  users: { username: string; email: string; first_name: string; last_name: string }[];
}

export type HoveredGraphNode =
  | { type: "user"; data: GraphUser & { task_count: number } }
  | { type: "task"; data: GraphTask };

export interface TaskGraphFilters {
  workspace?: string;
  projects?: string[];
  stateTypes?: string[];
  users?: string[];
}

export function useTaskGraph(apiBase: string) {
  const { getAuthHeader } = useAuth();

  async function getTaskGraph(filters: TaskGraphFilters): Promise<{ success: boolean; data?: TaskGraph; error?: string; }> {
    try {
      const params = new URLSearchParams();
      if (filters.workspace) params.set("workspace", filters.workspace);
      if (filters.projects?.length) params.set("projects", filters.projects.join(","));
      if (filters.stateTypes?.length) params.set("state_types", filters.stateTypes.join(","));
      if (filters.users?.length) params.set("users", filters.users.join(","));
      const response = await fetch(`${apiBase}?${params}`, { headers: getAuthHeader() });
      if (!response.ok) {
        const error = await response.json();
        return { success: false, error: error.message || "Failed to fetch graph" };
      }
      const data = await response.json();
      return { success: true, data };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function getTaskGraphFilters(): Promise<{ success: boolean; data?: TaskGraphFilterOptions; error?: string; }> {
    try {
      const response = await fetch(`${apiBase}/filters`, { headers: getAuthHeader() });
      if (!response.ok) {
        const error = await response.json();
        return { success: false, error: error.message || "Failed to fetch graph filters" };
      }
      const data = await response.json();
      return { success: true, data };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  return { getTaskGraph, getTaskGraphFilters };
}
