import type { ActivityLogEntry, VerifyActivityResponse } from "~/types";

interface ActivityState {
  activities: ActivityLogEntry[];
  loading: boolean;
}

const state = reactive<ActivityState>({
  activities: [],
  loading: false,
});

export function useActivity() {
  const { getAuthHeader } = useAuth();

  async function listActivity(
    projectKey: string,
    taskNum: number
  ): Promise<{ success: boolean; data?: ActivityLogEntry[]; error?: string }> {
    try {
      state.loading = true;
      const response = await fetch(
        `/api/v1/projects/${projectKey}/tasks/${taskNum}/activity`,
        { headers: getAuthHeader() }
      );

      if (!response.ok) {
        const error = await response.json();
        return { success: false, error: error.message || "Failed to fetch activity" };
      }

      const activities: ActivityLogEntry[] = await response.json();
      state.activities = activities;
      return { success: true, data: activities };
    } catch {
      return { success: false, error: "Network error" };
    } finally {
      state.loading = false;
    }
  }

  async function verifyActivity(
    projectKey: string,
    taskNum: number
  ): Promise<{ success: boolean; data?: VerifyActivityResponse; error?: string }> {
    try {
      const response = await fetch(
        `/api/v1/projects/${projectKey}/tasks/${taskNum}/activity/verify`,
        { headers: getAuthHeader() }
      );

      if (!response.ok) {
        const error = await response.json();
        return { success: false, error: error.message || "Failed to verify activity" };
      }

      const result: VerifyActivityResponse = await response.json();
      return { success: true, data: result };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  function clearActivity() {
    state.activities = [];
  }

  return {
    // State (readonly)
    activities: computed(() => state.activities),
    loading: computed(() => state.loading),

    // Methods
    listActivity,
    verifyActivity,
    clearActivity,
  };
}
