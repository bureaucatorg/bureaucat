import type {
  Cycle,
  PaginatedCyclesResponse,
  CycleTask,
  CycleMetrics,
  CycleSibling,
  CreateCycleRequest,
  UpdateCycleRequest,
} from "~/types";

interface CyclesState {
  cycles: Cycle[];
  currentCycle: Cycle | null;
  siblings: CycleSibling[];
  tasks: CycleTask[];
  metrics: CycleMetrics | null;
  activeCycles: Cycle[];
  loading: boolean;
  total: number;
  page: number;
  perPage: number;
  totalPages: number;
}

const state = reactive<CyclesState>({
  cycles: [],
  currentCycle: null,
  siblings: [],
  tasks: [],
  metrics: null,
  activeCycles: [],
  loading: false,
  total: 0,
  page: 1,
  perPage: 20,
  totalPages: 0,
});

export function useCycles() {
  const { getAuthHeader } = useAuth();

  async function listCycles(
    projectKey: string,
    page = 1,
    perPage = 20
  ): Promise<{ success: boolean; data?: PaginatedCyclesResponse; error?: string }> {
    try {
      state.loading = true;
      const response = await fetch(
        `/api/v1/projects/${projectKey}/cycles?page=${page}&per_page=${perPage}`,
        { headers: getAuthHeader() }
      );
      if (!response.ok) {
        const error = await response.json().catch(() => ({}));
        return { success: false, error: error.message || "Failed to fetch cycles" };
      }
      const data: PaginatedCyclesResponse = await response.json();
      state.cycles = data.cycles || [];
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

  async function createCycle(
    projectKey: string,
    data: CreateCycleRequest
  ): Promise<{ success: boolean; data?: Cycle; error?: string }> {
    try {
      const response = await fetch(`/api/v1/projects/${projectKey}/cycles`, {
        method: "POST",
        headers: { "Content-Type": "application/json", ...getAuthHeader() },
        body: JSON.stringify(data),
      });
      if (!response.ok) {
        const error = await response.json().catch(() => ({}));
        return { success: false, error: error.message || "Failed to create cycle" };
      }
      return { success: true, data: await response.json() };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function getCycle(
    projectKey: string,
    cycleId: string
  ): Promise<{ success: boolean; data?: Cycle; error?: string }> {
    try {
      const response = await fetch(`/api/v1/projects/${projectKey}/cycles/${cycleId}`, {
        headers: getAuthHeader(),
      });
      if (!response.ok) {
        const error = await response.json().catch(() => ({}));
        return { success: false, error: error.message || "Failed to fetch cycle" };
      }
      const cycle: Cycle = await response.json();
      state.currentCycle = cycle;
      return { success: true, data: cycle };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function updateCycle(
    projectKey: string,
    cycleId: string,
    data: UpdateCycleRequest
  ): Promise<{ success: boolean; data?: Cycle; error?: string }> {
    try {
      const response = await fetch(`/api/v1/projects/${projectKey}/cycles/${cycleId}`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json", ...getAuthHeader() },
        body: JSON.stringify(data),
      });
      if (!response.ok) {
        const error = await response.json().catch(() => ({}));
        return { success: false, error: error.message || "Failed to update cycle" };
      }
      const cycle: Cycle = await response.json();
      state.currentCycle = cycle;
      return { success: true, data: cycle };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function deleteCycle(
    projectKey: string,
    cycleId: string
  ): Promise<{ success: boolean; error?: string }> {
    try {
      const response = await fetch(`/api/v1/projects/${projectKey}/cycles/${cycleId}`, {
        method: "DELETE",
        headers: getAuthHeader(),
      });
      if (!response.ok) {
        const error = await response.json().catch(() => ({}));
        return { success: false, error: error.message || "Failed to delete cycle" };
      }
      return { success: true };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function listCycleTasks(
    projectKey: string,
    cycleId: string,
    assigneeId?: string | null
  ): Promise<{ success: boolean; data?: CycleTask[]; error?: string }> {
    try {
      const qs = assigneeId ? `?assignee=${assigneeId}` : "";
      const response = await fetch(
        `/api/v1/projects/${projectKey}/cycles/${cycleId}/tasks${qs}`,
        { headers: getAuthHeader() }
      );
      if (!response.ok) {
        const error = await response.json().catch(() => ({}));
        return { success: false, error: error.message || "Failed to fetch cycle tasks" };
      }
      const tasks: CycleTask[] = await response.json();
      state.tasks = tasks;
      return { success: true, data: tasks };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function getCycleMetrics(
    projectKey: string,
    cycleId: string
  ): Promise<{ success: boolean; data?: CycleMetrics; error?: string }> {
    try {
      const response = await fetch(
        `/api/v1/projects/${projectKey}/cycles/${cycleId}/metrics`,
        { headers: getAuthHeader() }
      );
      if (!response.ok) {
        const error = await response.json().catch(() => ({}));
        return { success: false, error: error.message || "Failed to fetch metrics" };
      }
      const metrics: CycleMetrics = await response.json();
      state.metrics = metrics;
      return { success: true, data: metrics };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function addTasksToCycle(
    projectKey: string,
    cycleId: string,
    taskIds: string[]
  ): Promise<{ success: boolean; error?: string }> {
    try {
      const response = await fetch(
        `/api/v1/projects/${projectKey}/cycles/${cycleId}/tasks`,
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

  async function removeTaskFromCycle(
    projectKey: string,
    cycleId: string,
    taskId: string
  ): Promise<{ success: boolean; error?: string }> {
    try {
      const response = await fetch(
        `/api/v1/projects/${projectKey}/cycles/${cycleId}/tasks/${taskId}`,
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

  async function listUnassignedTasks(
    projectKey: string,
    search = "",
    limit = 50
  ): Promise<{ success: boolean; data?: CycleTask[]; error?: string }> {
    try {
      const params = new URLSearchParams({ limit: limit.toString() });
      if (search) params.set("search", search);
      const response = await fetch(
        `/api/v1/projects/${projectKey}/cycles/unassigned-tasks?${params}`,
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

  async function listActiveCycles(): Promise<{
    success: boolean;
    data?: Cycle[];
    error?: string;
  }> {
    try {
      state.loading = true;
      const response = await fetch(`/api/v1/cycles/active`, { headers: getAuthHeader() });
      if (!response.ok) {
        const error = await response.json().catch(() => ({}));
        return { success: false, error: error.message || "Failed to fetch active cycles" };
      }
      const cycles: Cycle[] = await response.json();
      state.activeCycles = cycles;
      return { success: true, data: cycles };
    } catch {
      return { success: false, error: "Network error" };
    } finally {
      state.loading = false;
    }
  }

  async function listSiblings(
    projectKey: string
  ): Promise<{ success: boolean; data?: CycleSibling[]; error?: string }> {
    try {
      const response = await fetch(`/api/v1/projects/${projectKey}/cycles/all`, {
        headers: getAuthHeader(),
      });
      if (!response.ok) {
        const error = await response.json().catch(() => ({}));
        return { success: false, error: error.message || "Failed to fetch cycles" };
      }
      const siblings: CycleSibling[] = await response.json();
      state.siblings = siblings;
      return { success: true, data: siblings };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  function clearCurrent() {
    state.currentCycle = null;
    state.tasks = [];
    state.metrics = null;
    state.siblings = [];
  }

  return {
    cycles: computed(() => state.cycles),
    currentCycle: computed(() => state.currentCycle),
    siblings: computed(() => state.siblings),
    tasks: computed(() => state.tasks),
    metrics: computed(() => state.metrics),
    activeCycles: computed(() => state.activeCycles),
    loading: computed(() => state.loading),
    total: computed(() => state.total),
    page: computed(() => state.page),
    perPage: computed(() => state.perPage),
    totalPages: computed(() => state.totalPages),

    listCycles,
    createCycle,
    getCycle,
    updateCycle,
    deleteCycle,
    listCycleTasks,
    getCycleMetrics,
    addTasksToCycle,
    removeTaskFromCycle,
    listUnassignedTasks,
    listActiveCycles,
    listSiblings,
    clearCurrent,
  };
}
