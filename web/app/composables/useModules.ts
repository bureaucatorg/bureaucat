import type {
  Module,
  ModuleListFilters,
  ModuleMetrics,
  ModuleTask,
  ModuleUserBrief,
  PaginatedModulesResponse,
  CreateModuleRequest,
  UpdateModuleRequest,
  DuplicateModuleRequest,
} from "~/types";

interface ModulesState {
  modules: Module[];
  currentModule: Module | null;
  tasks: ModuleTask[];
  metrics: ModuleMetrics | null;
  members: ModuleUserBrief[];
  activeModules: Module[];
  loading: boolean;
  total: number;
  page: number;
  perPage: number;
  totalPages: number;
}

const state = reactive<ModulesState>({
  modules: [],
  currentModule: null,
  tasks: [],
  metrics: null,
  members: [],
  activeModules: [],
  loading: false,
  total: 0,
  page: 1,
  perPage: 20,
  totalPages: 0,
});

function qs(filters?: ModuleListFilters) {
  const p = new URLSearchParams();
  if (!filters) return p;
  if (filters.status) p.set("status", filters.status);
  if (filters.lead_id) p.set("lead_id", filters.lead_id);
  if (filters.start_after) p.set("start_after", filters.start_after);
  if (filters.end_before) p.set("end_before", filters.end_before);
  if (filters.sort_by) p.set("sort_by", filters.sort_by);
  if (filters.sort_dir) p.set("sort_dir", filters.sort_dir);
  return p;
}

export function useModules() {
  const { getAuthHeader } = useAuth();

  async function listModules(
    projectKey: string,
    page = 1,
    perPage = 20,
    filters: ModuleListFilters = {}
  ): Promise<{ success: boolean; data?: PaginatedModulesResponse; error?: string }> {
    try {
      state.loading = true;
      const params = qs(filters);
      params.set("page", String(page));
      params.set("per_page", String(perPage));
      const response = await fetch(
        `/api/v1/projects/${projectKey}/modules?${params}`,
        { headers: getAuthHeader() }
      );
      if (!response.ok) {
        const error = await response.json().catch(() => ({}));
        return { success: false, error: error.message || "Failed to fetch modules" };
      }
      const data: PaginatedModulesResponse = await response.json();
      state.modules = data.modules || [];
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

  async function createModule(
    projectKey: string,
    data: CreateModuleRequest
  ): Promise<{ success: boolean; data?: Module; error?: string }> {
    try {
      const response = await fetch(`/api/v1/projects/${projectKey}/modules`, {
        method: "POST",
        headers: { "Content-Type": "application/json", ...getAuthHeader() },
        body: JSON.stringify(data),
      });
      if (!response.ok) {
        const error = await response.json().catch(() => ({}));
        return { success: false, error: error.message || "Failed to create module" };
      }
      return { success: true, data: await response.json() };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function getModule(
    projectKey: string,
    moduleId: string
  ): Promise<{ success: boolean; data?: Module; error?: string }> {
    try {
      const response = await fetch(
        `/api/v1/projects/${projectKey}/modules/${moduleId}`,
        { headers: getAuthHeader() }
      );
      if (!response.ok) {
        const error = await response.json().catch(() => ({}));
        return { success: false, error: error.message || "Failed to fetch module" };
      }
      const m: Module = await response.json();
      state.currentModule = m;
      state.members = m.members ?? [];
      return { success: true, data: m };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function updateModule(
    projectKey: string,
    moduleId: string,
    data: UpdateModuleRequest
  ): Promise<{ success: boolean; data?: Module; error?: string }> {
    try {
      const response = await fetch(
        `/api/v1/projects/${projectKey}/modules/${moduleId}`,
        {
          method: "PATCH",
          headers: { "Content-Type": "application/json", ...getAuthHeader() },
          body: JSON.stringify(data),
        }
      );
      if (!response.ok) {
        const error = await response.json().catch(() => ({}));
        return { success: false, error: error.message || "Failed to update module" };
      }
      const m: Module = await response.json();
      state.currentModule = m;
      state.members = m.members ?? [];
      return { success: true, data: m };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function deleteModule(
    projectKey: string,
    moduleId: string
  ): Promise<{ success: boolean; error?: string }> {
    try {
      const response = await fetch(
        `/api/v1/projects/${projectKey}/modules/${moduleId}`,
        { method: "DELETE", headers: getAuthHeader() }
      );
      if (!response.ok) {
        const error = await response.json().catch(() => ({}));
        return { success: false, error: error.message || "Failed to delete module" };
      }
      return { success: true };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function duplicateModule(
    projectKey: string,
    moduleId: string,
    payload: DuplicateModuleRequest
  ): Promise<{ success: boolean; data?: Module; error?: string }> {
    try {
      const response = await fetch(
        `/api/v1/projects/${projectKey}/modules/${moduleId}/duplicate`,
        {
          method: "POST",
          headers: { "Content-Type": "application/json", ...getAuthHeader() },
          body: JSON.stringify(payload),
        }
      );
      if (!response.ok) {
        const error = await response.json().catch(() => ({}));
        return { success: false, error: error.message || "Failed to duplicate module" };
      }
      return { success: true, data: await response.json() };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function listModuleTasks(
    projectKey: string,
    moduleId: string,
    assigneeId?: string | null
  ): Promise<{ success: boolean; data?: ModuleTask[]; error?: string }> {
    try {
      const qs = assigneeId ? `?assignee=${assigneeId}` : "";
      const response = await fetch(
        `/api/v1/projects/${projectKey}/modules/${moduleId}/tasks${qs}`,
        { headers: getAuthHeader() }
      );
      if (!response.ok) {
        const error = await response.json().catch(() => ({}));
        return { success: false, error: error.message || "Failed to fetch module tasks" };
      }
      const tasks: ModuleTask[] = await response.json();
      state.tasks = tasks;
      return { success: true, data: tasks };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function addTasksToModule(
    projectKey: string,
    moduleId: string,
    taskIds: string[]
  ): Promise<{ success: boolean; error?: string }> {
    try {
      const response = await fetch(
        `/api/v1/projects/${projectKey}/modules/${moduleId}/tasks`,
        {
          method: "POST",
          headers: { "Content-Type": "application/json", ...getAuthHeader() },
          body: JSON.stringify({ task_ids: taskIds }),
        }
      );
      if (!response.ok) {
        const error = await response.json().catch(() => ({}));
        return { success: false, error: error.message || "Failed to add tasks" };
      }
      return { success: true };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function removeTaskFromModule(
    projectKey: string,
    moduleId: string,
    taskId: string
  ): Promise<{ success: boolean; error?: string }> {
    try {
      const response = await fetch(
        `/api/v1/projects/${projectKey}/modules/${moduleId}/tasks/${taskId}`,
        { method: "DELETE", headers: getAuthHeader() }
      );
      if (!response.ok) {
        const error = await response.json().catch(() => ({}));
        return { success: false, error: error.message || "Failed to remove task" };
      }
      return { success: true };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function listPickerTasks(
    projectKey: string,
    moduleId: string,
    search = "",
    limit = 50
  ): Promise<{ success: boolean; data?: ModuleTask[]; error?: string }> {
    try {
      const params = new URLSearchParams({
        module_id: moduleId,
        limit: String(limit),
      });
      if (search) params.set("search", search);
      const response = await fetch(
        `/api/v1/projects/${projectKey}/modules/tasks-picker?${params}`,
        { headers: getAuthHeader() }
      );
      if (!response.ok) {
        const error = await response.json().catch(() => ({}));
        return { success: false, error: error.message || "Failed to fetch tasks" };
      }
      return { success: true, data: await response.json() };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function listModuleMembers(
    projectKey: string,
    moduleId: string
  ): Promise<{ success: boolean; data?: ModuleUserBrief[]; error?: string }> {
    try {
      const response = await fetch(
        `/api/v1/projects/${projectKey}/modules/${moduleId}/members`,
        { headers: getAuthHeader() }
      );
      if (!response.ok) {
        const error = await response.json().catch(() => ({}));
        return { success: false, error: error.message || "Failed to fetch members" };
      }
      const members: ModuleUserBrief[] = await response.json();
      state.members = members;
      return { success: true, data: members };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function addModuleMember(
    projectKey: string,
    moduleId: string,
    userId: string
  ): Promise<{ success: boolean; error?: string }> {
    try {
      const response = await fetch(
        `/api/v1/projects/${projectKey}/modules/${moduleId}/members`,
        {
          method: "POST",
          headers: { "Content-Type": "application/json", ...getAuthHeader() },
          body: JSON.stringify({ user_id: userId }),
        }
      );
      if (!response.ok) {
        const error = await response.json().catch(() => ({}));
        return { success: false, error: error.message || "Failed to add member" };
      }
      return { success: true };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function removeModuleMember(
    projectKey: string,
    moduleId: string,
    userId: string
  ): Promise<{ success: boolean; error?: string }> {
    try {
      const response = await fetch(
        `/api/v1/projects/${projectKey}/modules/${moduleId}/members/${userId}`,
        { method: "DELETE", headers: getAuthHeader() }
      );
      if (!response.ok) {
        const error = await response.json().catch(() => ({}));
        return { success: false, error: error.message || "Failed to remove member" };
      }
      return { success: true };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function getModuleMetrics(
    projectKey: string,
    moduleId: string
  ): Promise<{ success: boolean; data?: ModuleMetrics; error?: string }> {
    try {
      const response = await fetch(
        `/api/v1/projects/${projectKey}/modules/${moduleId}/metrics`,
        { headers: getAuthHeader() }
      );
      if (!response.ok) {
        const error = await response.json().catch(() => ({}));
        return { success: false, error: error.message || "Failed to fetch metrics" };
      }
      const metrics: ModuleMetrics = await response.json();
      state.metrics = metrics;
      return { success: true, data: metrics };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function listActiveModules(): Promise<{
    success: boolean;
    data?: Module[];
    error?: string;
  }> {
    try {
      state.loading = true;
      const response = await fetch(`/api/v1/modules/active`, { headers: getAuthHeader() });
      if (!response.ok) {
        const error = await response.json().catch(() => ({}));
        return { success: false, error: error.message || "Failed to fetch active modules" };
      }
      const modules: Module[] = await response.json();
      state.activeModules = modules;
      return { success: true, data: modules };
    } catch {
      return { success: false, error: "Network error" };
    } finally {
      state.loading = false;
    }
  }

  function clearCurrent() {
    state.currentModule = null;
    state.tasks = [];
    state.metrics = null;
    state.members = [];
  }

  return {
    modules: computed(() => state.modules),
    currentModule: computed(() => state.currentModule),
    tasks: computed(() => state.tasks),
    metrics: computed(() => state.metrics),
    members: computed(() => state.members),
    activeModules: computed(() => state.activeModules),
    loading: computed(() => state.loading),
    total: computed(() => state.total),
    page: computed(() => state.page),
    perPage: computed(() => state.perPage),
    totalPages: computed(() => state.totalPages),

    listModules,
    createModule,
    getModule,
    updateModule,
    deleteModule,
    duplicateModule,
    listModuleTasks,
    addTasksToModule,
    removeTaskFromModule,
    listPickerTasks,
    listModuleMembers,
    addModuleMember,
    removeModuleMember,
    getModuleMetrics,
    listActiveModules,
    clearCurrent,
  };
}
