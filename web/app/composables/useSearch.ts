export interface SearchTaskResult {
  id: string;
  task_number: number;
  task_key: string;
  title: string;
  description?: string;
  project_id: string;
  project_key: string;
  project_name: string;
  state_name: string;
  state_type: string;
  state_color: string;
  updated_at: string;
}

export interface SearchCycleResult {
  id: string;
  title: string;
  start_date: string;
  end_date: string;
  project_id: string;
  project_key: string;
  project_name: string;
}

export interface SearchProjectResult {
  id: string;
  project_key: string;
  name: string;
  description?: string;
  icon_url?: string;
}

export interface SearchResponse {
  tasks: SearchTaskResult[];
  cycles: SearchCycleResult[];
  projects: SearchProjectResult[];
}

const emptyResults: SearchResponse = { tasks: [], cycles: [], projects: [] };

export function useSearch() {
  const { getAuthHeader } = useAuth();

  const query = ref("");
  const results = ref<SearchResponse>({ ...emptyResults });
  const loading = ref(false);

  let debounceTimer: ReturnType<typeof setTimeout> | null = null;
  let activeController: AbortController | null = null;

  async function runSearch(q: string) {
    if (activeController) activeController.abort();

    const trimmed = q.trim();
    if (!trimmed) {
      results.value = { ...emptyResults };
      loading.value = false;
      return;
    }

    const controller = new AbortController();
    activeController = controller;
    loading.value = true;

    try {
      const response = await fetch(
        `/api/v1/search?q=${encodeURIComponent(trimmed)}`,
        { headers: getAuthHeader(), signal: controller.signal },
      );
      if (!response.ok) {
        results.value = { ...emptyResults };
        return;
      }
      const data: SearchResponse = await response.json();
      // Only commit if this is still the active request
      if (activeController === controller) {
        results.value = {
          tasks: data.tasks ?? [],
          cycles: data.cycles ?? [],
          projects: data.projects ?? [],
        };
      }
    } catch (err) {
      // Ignore aborts; swallow network errors silently for a palette.
      if ((err as Error)?.name !== "AbortError") {
        results.value = { ...emptyResults };
      }
    } finally {
      if (activeController === controller) {
        loading.value = false;
        activeController = null;
      }
    }
  }

  function search(q: string, debounceMs = 150) {
    query.value = q;
    if (debounceTimer) clearTimeout(debounceTimer);
    if (debounceMs <= 0) {
      runSearch(q);
      return;
    }
    debounceTimer = setTimeout(() => runSearch(q), debounceMs);
  }

  function reset() {
    if (debounceTimer) clearTimeout(debounceTimer);
    if (activeController) activeController.abort();
    query.value = "";
    results.value = { ...emptyResults };
    loading.value = false;
  }

  return {
    query,
    results: computed(() => results.value),
    loading: computed(() => loading.value),
    search,
    reset,
  };
}
