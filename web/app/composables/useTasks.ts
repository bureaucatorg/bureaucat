import type {
  Task,
  PaginatedTasksResponse,
  CreateTaskRequest,
  UpdateTaskRequest,
  TaskFilters,
} from "~/types";

interface TasksState {
  tasks: Task[];
  currentTask: Task | null;
  loading: boolean;
  total: number;
  page: number;
  perPage: number;
  totalPages: number;
  filters: TaskFilters;
}

const state = reactive<TasksState>({
  tasks: [],
  currentTask: null,
  loading: false,
  total: 0,
  page: 1,
  perPage: 20,
  totalPages: 0,
  filters: {},
});

export function useTasks() {
  const { getAuthHeader } = useAuth();

  function buildQueryString(page: number, perPage: number, filters: TaskFilters): string {
    const params = new URLSearchParams();
    params.set("page", page.toString());
    params.set("per_page", perPage.toString());

    if (filters.state_id) params.set("state_id", filters.state_id);
    if (filters.state_type) params.set("state_type", filters.state_type);
    if (filters.created_by) params.set("created_by", filters.created_by);
    if (filters.priority !== undefined) params.set("priority", filters.priority.toString());
    if (filters.q) params.set("q", filters.q);

    return params.toString();
  }

  async function listTasks(
    projectKey: string,
    page = 1,
    perPage = 20,
    filters: TaskFilters = {}
  ): Promise<{ success: boolean; data?: PaginatedTasksResponse; error?: string }> {
    try {
      state.loading = true;
      state.filters = filters;
      const queryString = buildQueryString(page, perPage, filters);
      const response = await fetch(
        `/api/v1/projects/${projectKey}/tasks?${queryString}`,
        { headers: getAuthHeader() }
      );

      if (!response.ok) {
        const error = await response.json();
        return { success: false, error: error.message || "Failed to fetch tasks" };
      }

      const data: PaginatedTasksResponse = await response.json();
      state.tasks = data.tasks || [];
      state.total = data.total;
      state.page = data.page;
      state.perPage = data.per_page;
      state.totalPages = data.total_pages;
      return { success: true, data };
    } catch {
      return { success: false, error: "Network error" };
    } finally {
      state.loading = false;
    }
  }

  async function createTask(
    projectKey: string,
    data: CreateTaskRequest
  ): Promise<{ success: boolean; data?: Task; error?: string }> {
    try {
      const response = await fetch(`/api/v1/projects/${projectKey}/tasks`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          ...getAuthHeader(),
        },
        body: JSON.stringify(data),
      });

      if (!response.ok) {
        const error = await response.json();
        return { success: false, error: error.message || "Failed to create task" };
      }

      const task: Task = await response.json();
      return { success: true, data: task };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function getTask(
    projectKey: string,
    taskNum: number
  ): Promise<{ success: boolean; data?: Task; error?: string }> {
    try {
      const response = await fetch(`/api/v1/projects/${projectKey}/tasks/${taskNum}`, {
        headers: getAuthHeader(),
      });

      if (!response.ok) {
        const error = await response.json();
        return { success: false, error: error.message || "Failed to fetch task" };
      }

      const task: Task = await response.json();
      state.currentTask = task;
      return { success: true, data: task };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function updateTask(
    projectKey: string,
    taskNum: number,
    data: UpdateTaskRequest
  ): Promise<{ success: boolean; data?: Task; error?: string }> {
    try {
      const response = await fetch(`/api/v1/projects/${projectKey}/tasks/${taskNum}`, {
        method: "PATCH",
        headers: {
          "Content-Type": "application/json",
          ...getAuthHeader(),
        },
        body: JSON.stringify(data),
      });

      if (!response.ok) {
        const error = await response.json();
        return { success: false, error: error.message || "Failed to update task" };
      }

      const task: Task = await response.json();
      state.currentTask = task;

      // Update task in list if present
      const idx = state.tasks.findIndex((t) => t.id === task.id);
      if (idx !== -1) {
        state.tasks[idx] = task;
      }

      return { success: true, data: task };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function deleteTask(
    projectKey: string,
    taskNum: number
  ): Promise<{ success: boolean; error?: string }> {
    try {
      const response = await fetch(`/api/v1/projects/${projectKey}/tasks/${taskNum}`, {
        method: "DELETE",
        headers: getAuthHeader(),
      });

      if (!response.ok) {
        const error = await response.json();
        return { success: false, error: error.message || "Failed to delete task" };
      }

      return { success: true };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  // Assignees
  async function addAssignee(
    projectKey: string,
    taskNum: number,
    userId: string
  ): Promise<{ success: boolean; error?: string }> {
    try {
      const response = await fetch(
        `/api/v1/projects/${projectKey}/tasks/${taskNum}/assignees`,
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            ...getAuthHeader(),
          },
          body: JSON.stringify({ user_id: userId }),
        }
      );

      if (!response.ok) {
        const error = await response.json();
        return { success: false, error: error.message || "Failed to add assignee" };
      }

      return { success: true };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function removeAssignee(
    projectKey: string,
    taskNum: number,
    userId: string
  ): Promise<{ success: boolean; error?: string }> {
    try {
      const response = await fetch(
        `/api/v1/projects/${projectKey}/tasks/${taskNum}/assignees/${userId}`,
        {
          method: "DELETE",
          headers: getAuthHeader(),
        }
      );

      if (!response.ok) {
        const error = await response.json();
        return { success: false, error: error.message || "Failed to remove assignee" };
      }

      return { success: true };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  // Labels
  async function addLabel(
    projectKey: string,
    taskNum: number,
    labelId: string
  ): Promise<{ success: boolean; error?: string }> {
    try {
      const response = await fetch(
        `/api/v1/projects/${projectKey}/tasks/${taskNum}/labels`,
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            ...getAuthHeader(),
          },
          body: JSON.stringify({ label_id: labelId }),
        }
      );

      if (!response.ok) {
        const error = await response.json();
        return { success: false, error: error.message || "Failed to add label" };
      }

      return { success: true };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function removeLabel(
    projectKey: string,
    taskNum: number,
    labelId: string
  ): Promise<{ success: boolean; error?: string }> {
    try {
      const response = await fetch(
        `/api/v1/projects/${projectKey}/tasks/${taskNum}/labels/${labelId}`,
        {
          method: "DELETE",
          headers: getAuthHeader(),
        }
      );

      if (!response.ok) {
        const error = await response.json();
        return { success: false, error: error.message || "Failed to remove label" };
      }

      return { success: true };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  function clearCurrentTask() {
    state.currentTask = null;
  }

  function setFilters(filters: TaskFilters) {
    state.filters = filters;
  }

  function clearFilters() {
    state.filters = {};
  }

  return {
    // State (readonly)
    tasks: computed(() => state.tasks),
    currentTask: computed(() => state.currentTask),
    loading: computed(() => state.loading),
    total: computed(() => state.total),
    page: computed(() => state.page),
    perPage: computed(() => state.perPage),
    totalPages: computed(() => state.totalPages),
    filters: computed(() => state.filters),

    // Tasks CRUD
    listTasks,
    createTask,
    getTask,
    updateTask,
    deleteTask,

    // Assignees
    addAssignee,
    removeAssignee,

    // Labels
    addLabel,
    removeLabel,

    // Utils
    clearCurrentTask,
    setFilters,
    clearFilters,
  };
}
