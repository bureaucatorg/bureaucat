interface User {
  id: string;
  username: string;
  email: string;
  first_name: string;
  last_name: string;
  user_type: string;
  created_at: string;
}

interface PaginatedUsersResponse {
  users: User[];
  total: number;
  page: number;
  per_page: number;
  total_pages: number;
}

export interface MergeUserResult {
  workspaces: number;
  projects: number;
  assignments: number;
  modules: number;
  followed_tasks: number;
  views: number;
}

interface TokenInfo {
  id: string;
  user_id: string;
  username: string;
  email: string;
  created_at: string;
  expires_at: string;
}

interface PaginatedTokensResponse {
  tokens: TokenInfo[];
  total: number;
  page: number;
  per_page: number;
  total_pages: number;
}

interface StatCount {
  label: string;
  count: number;
}

interface ProjectStat {
  project_key: string;
  name: string;
  task_count: number;
}

interface WorkspaceStat {
  workspace_key: string;
  name: string;
  project_count: number;
}

interface DayCount {
  day: string;
  count: number;
}

interface ViewDayCount {
  day: string;
  private: number;
  shared: number;
}

export interface AdminStats {
  totals: {
    workspaces: number;
    projects: number;
    tasks: number;
    subtasks: number;
    pages: number;
    users: number;
    attachments: number;
    attachments_bytes: number;
  };
  tasks_by_state: StatCount[];
  tasks_by_priority: StatCount[];
  top_projects: ProjectStat[];
  projects_per_workspace: WorkspaceStat[];
  series: {
    from: string;
    to: string;
    days: number;
    tasks: DayCount[];
    subtasks: DayCount[];
    pages: DayCount[];
    views: ViewDayCount[];
    comments: DayCount[];
    activity: DayCount[];
    attachments: DayCount[];
    cycles: DayCount[];
    modules: DayCount[];
  };
}

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

export interface TaskGraph {
  users: GraphUser[];
  tasks: GraphTask[];
  edges: GraphEdge[];
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

interface CreateUserData {
  username: string;
  email: string;
  password: string;
  first_name: string;
  last_name: string;
  user_type: string;
}

export interface DeletedProject {
  id: string;
  project_key: string;
  name: string;
  description?: string;
  workspace_id: string;
  workspace_name: string;
  created_by: string;
  creator_name: string;
  creator_username: string;
  created_at: string;
  deleted_at: string;
}

interface PaginatedDeletedProjectsResponse {
  projects: DeletedProject[];
  total: number;
  page: number;
  per_page: number;
  total_pages: number;
}

export function useAdmin() {
  const { getAuthHeader } = useAuth();

  async function listUsers(
    page = 1,
    perPage = 20,
    search = ""
  ): Promise<{
    success: boolean;
    data?: PaginatedUsersResponse;
    error?: string;
  }> {
    try {
      const params = new URLSearchParams({ page: String(page), per_page: String(perPage) });
      if (search) params.set("search", search);
      const response = await fetch(
        `/api/v1/admin/users?${params}`,
        { headers: getAuthHeader() }
      );

      if (!response.ok) {
        const error = await response.json();
        return { success: false, error: error.message || "Failed to fetch users" };
      }

      const data = await response.json();
      return { success: true, data };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function createUser(userData: CreateUserData): Promise<{
    success: boolean;
    data?: User;
    error?: string;
  }> {
    try {
      const response = await fetch("/api/v1/admin/users", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          ...getAuthHeader(),
        },
        body: JSON.stringify(userData),
      });

      if (!response.ok) {
        const error = await response.json();
        return { success: false, error: error.message || "Failed to create user" };
      }

      const data = await response.json();
      return { success: true, data };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function deleteUser(userId: string): Promise<{
    success: boolean;
    error?: string;
  }> {
    try {
      const response = await fetch(`/api/v1/admin/users/${userId}`, {
        method: "DELETE",
        headers: getAuthHeader(),
      });

      if (!response.ok) {
        const error = await response.json();
        return { success: false, error: error.message || "Failed to delete user" };
      }

      return { success: true };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function listTokens(
    page = 1,
    perPage = 20
  ): Promise<{
    success: boolean;
    data?: PaginatedTokensResponse;
    error?: string;
  }> {
    try {
      const response = await fetch(
        `/api/v1/admin/tokens?page=${page}&per_page=${perPage}`,
        { headers: getAuthHeader() }
      );

      if (!response.ok) {
        const error = await response.json();
        return { success: false, error: error.message || "Failed to fetch tokens" };
      }

      const data = await response.json();
      return { success: true, data };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function revokeToken(tokenId: string): Promise<{
    success: boolean;
    error?: string;
  }> {
    try {
      const response = await fetch(`/api/v1/admin/tokens/${tokenId}`, {
        method: "DELETE",
        headers: getAuthHeader(),
      });

      if (!response.ok) {
        const error = await response.json();
        return { success: false, error: error.message || "Failed to revoke token" };
      }

      return { success: true };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function cleanupExpiredTokens(): Promise<{
    success: boolean;
    deleted?: number;
    error?: string;
  }> {
    try {
      const response = await fetch("/api/v1/admin/tokens/expired", {
        method: "DELETE",
        headers: getAuthHeader(),
      });

      if (!response.ok) {
        const error = await response.json();
        return { success: false, error: error.message || "Failed to cleanup tokens" };
      }

      const data = await response.json();
      return { success: true, deleted: data.deleted };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function updateUserRole(userId: string, userType: string): Promise<{
    success: boolean;
    data?: User;
    error?: string;
  }> {
    try {
      const response = await fetch(`/api/v1/admin/users/${userId}/role`, {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
          ...getAuthHeader(),
        },
        body: JSON.stringify({ user_type: userType }),
      });

      if (!response.ok) {
        const error = await response.json();
        return { success: false, error: error.message || "Failed to update role" };
      }

      const data = await response.json();
      return { success: true, data };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function resetUserPassword(userId: string, password: string): Promise<{
    success: boolean;
    error?: string;
  }> {
    try {
      const response = await fetch(`/api/v1/admin/users/${userId}/password`, {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
          ...getAuthHeader(),
        },
        body: JSON.stringify({ password }),
      });

      if (!response.ok) {
        const error = await response.json();
        return { success: false, error: error.message || "Failed to reset password" };
      }

      return { success: true };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function mergeUser(sourceId: string, targetId: string): Promise<{
    success: boolean;
    data?: MergeUserResult;
    error?: string;
  }> {
    try {
      const response = await fetch(`/api/v1/admin/users/${sourceId}/merge`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          ...getAuthHeader(),
        },
        body: JSON.stringify({ target_user_id: targetId }),
      });

      if (!response.ok) {
        const error = await response.json();
        return { success: false, error: error.message || "Failed to merge user" };
      }

      const data = await response.json();
      return { success: true, data };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function getStats(from: string, to: string): Promise<{
    success: boolean;
    data?: AdminStats;
    error?: string;
  }> {
    try {
      const params = new URLSearchParams({ from, to });
      const response = await fetch(`/api/v1/admin/stats?${params}`, {
        headers: getAuthHeader(),
      });

      if (!response.ok) {
        const error = await response.json();
        return { success: false, error: error.message || "Failed to fetch stats" };
      }

      const data = await response.json();
      return { success: true, data };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function listDeletedProjects(
    page = 1,
    perPage = 20
  ): Promise<{
    success: boolean;
    data?: PaginatedDeletedProjectsResponse;
    error?: string;
  }> {
    try {
      const response = await fetch(
        `/api/v1/admin/projects/deleted?page=${page}&per_page=${perPage}`,
        { headers: getAuthHeader() }
      );

      if (!response.ok) {
        const error = await response.json();
        return { success: false, error: error.message || "Failed to fetch deleted projects" };
      }

      const data = await response.json();
      return { success: true, data };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function restoreProject(projectId: string): Promise<{
    success: boolean;
    error?: string;
  }> {
    try {
      const response = await fetch(`/api/v1/admin/projects/${projectId}/restore`, {
        method: "POST",
        headers: getAuthHeader(),
      });

      if (!response.ok) {
        const error = await response.json();
        return { success: false, error: error.message || "Failed to restore project" };
      }

      return { success: true };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function getTaskGraph(filters: TaskGraphFilters): Promise<{ success: boolean; data?: TaskGraph; error?: string; }> {
    try {
      const params = new URLSearchParams();
      if (filters.workspace) params.set("workspace", filters.workspace);
      if (filters.projects?.length) params.set("projects", filters.projects.join(","));
      if (filters.stateTypes?.length) params.set("state_types", filters.stateTypes.join(","));
      if (filters.users?.length) params.set("users", filters.users.join(","));
      const response = await fetch(`/api/v1/admin/graph?${params}`, { headers: getAuthHeader() });
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
      const response = await fetch("/api/v1/admin/graph/filters", { headers: getAuthHeader() });
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

  return {
    getStats,
    getTaskGraph,
    getTaskGraphFilters,
    listUsers,
    createUser,
    deleteUser,
    updateUserRole,
    resetUserPassword,
    mergeUser,
    listTokens,
    revokeToken,
    cleanupExpiredTokens,
    listDeletedProjects,
    restoreProject,
  };
}
