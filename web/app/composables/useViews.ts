import type { ProjectView, CreateViewInput, UpdateViewInput } from "~/types";

interface ViewsState {
  views: ProjectView[];
  loading: boolean;
}

const state = reactive<ViewsState>({
  views: [],
  loading: false,
});

export function useViews() {
  const { getAuthHeader } = useAuth();

  async function listViews(projectKey: string) {
    state.loading = true;
    try {
      const resp = await fetch(`/api/v1/projects/${projectKey}/views`, {
        headers: getAuthHeader(),
      });
      if (!resp.ok) {
        return { success: false, error: await readError(resp) };
      }
      const data = (await resp.json()) as ProjectView[];
      state.views = data ?? [];
      return { success: true, data: state.views };
    } catch {
      return { success: false, error: "Network error" };
    } finally {
      state.loading = false;
    }
  }

  async function getView(projectKey: string, slug: string) {
    try {
      const resp = await fetch(`/api/v1/projects/${projectKey}/views/${slug}`, {
        headers: getAuthHeader(),
      });
      if (!resp.ok) return { success: false, error: await readError(resp) };
      return { success: true, data: (await resp.json()) as ProjectView };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function createView(projectKey: string, input: CreateViewInput) {
    try {
      const resp = await fetch(`/api/v1/projects/${projectKey}/views`, {
        method: "POST",
        headers: { "Content-Type": "application/json", ...getAuthHeader() },
        body: JSON.stringify(input),
      });
      if (!resp.ok) return { success: false, error: await readError(resp) };
      const view = (await resp.json()) as ProjectView;
      state.views = [...state.views, view];
      return { success: true, data: view };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function updateView(projectKey: string, slug: string, patch: UpdateViewInput) {
    try {
      const resp = await fetch(`/api/v1/projects/${projectKey}/views/${slug}`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json", ...getAuthHeader() },
        body: JSON.stringify(patch),
      });
      if (!resp.ok) return { success: false, error: await readError(resp) };
      const view = (await resp.json()) as ProjectView;
      const idx = state.views.findIndex((v) => v.slug === slug);
      if (idx !== -1) state.views[idx] = view;
      return { success: true, data: view };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function deleteView(projectKey: string, slug: string) {
    try {
      const resp = await fetch(`/api/v1/projects/${projectKey}/views/${slug}`, {
        method: "DELETE",
        headers: getAuthHeader(),
      });
      if (!resp.ok) return { success: false, error: await readError(resp) };
      state.views = state.views.filter((v) => v.slug !== slug);
      return { success: true };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function reorderViews(
    projectKey: string,
    items: { id: string; position: number }[]
  ) {
    try {
      const resp = await fetch(`/api/v1/projects/${projectKey}/views/reorder`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json", ...getAuthHeader() },
        body: JSON.stringify({ items }),
      });
      if (!resp.ok) return { success: false, error: await readError(resp) };
      return { success: true };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  return {
    views: computed(() => state.views),
    loading: computed(() => state.loading),
    listViews,
    getView,
    createView,
    updateView,
    deleteView,
    reorderViews,
  };
}

async function readError(resp: Response): Promise<string> {
  try {
    const body = await resp.json();
    return (body && (body.message || body.error)) || `HTTP ${resp.status}`;
  } catch {
    return `HTTP ${resp.status}`;
  }
}
