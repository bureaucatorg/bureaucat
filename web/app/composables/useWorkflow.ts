import type { SaveWorkflowRequest, Workflow } from "~/types";

type Result<T> = { success: boolean; data?: T; error?: string };

const workflows = reactive<Record<string, Workflow>>({});
const inflight = new Map<string, Promise<Result<Workflow>>>();

export function useWorkflow() {
  const { getAuthHeader } = useAuth();

  async function request(
    url: string,
    init: RequestInit,
    projectKey: string,
    fallback: string
  ): Promise<Result<Workflow>> {
    try {
      const response = await fetch(url, {
        ...init,
        headers: { "Content-Type": "application/json", ...getAuthHeader() },
      });
      if (!response.ok) {
        const error = await response.json().catch(() => ({}));
        return { success: false, error: error.message || fallback };
      }
      const data: Workflow = await response.json();
      workflows[projectKey] = data;
      return { success: true, data };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  function getWorkflow(projectKey: string, force = false): Promise<Result<Workflow>> {
    const cached = workflows[projectKey];
    if (cached && !force) return Promise.resolve({ success: true, data: cached });
    const pending = inflight.get(projectKey);
    if (pending && !force) return pending;
    const p = request(
      `/api/v1/projects/${projectKey}/workflow`,
      { method: "GET" },
      projectKey,
      "Failed to load workflow"
    ).finally(() => inflight.delete(projectKey));
    inflight.set(projectKey, p);
    return p;
  }

  function saveWorkflow(projectKey: string, data: SaveWorkflowRequest) {
    return request(
      `/api/v1/projects/${projectKey}/workflow`,
      { method: "PUT", body: JSON.stringify(data) },
      projectKey,
      "Failed to save workflow"
    );
  }

  function setWorkflowEnabled(projectKey: string, enabled: boolean) {
    return request(
      `/api/v1/projects/${projectKey}/workflow/enabled`,
      { method: "PUT", body: JSON.stringify({ enabled }) },
      projectKey,
      "Failed to update workflow"
    );
  }

  return { workflows, getWorkflow, saveWorkflow, setWorkflowEnabled };
}
