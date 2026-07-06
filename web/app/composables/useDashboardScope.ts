import { ref, watch } from "vue";

// Shared "All workspaces" preference for the dashboard's "Assigned to You"
// section. Lifted into a composable (singleton reactive, mirroring
// useAuth/useProjects) so the global Create Task dialog opened via Shift+C can
// honor the same toggle while the user is on the dashboard.
const STORAGE_KEY = "bureaucat.dashboardAllWorkspaces";

function readInitial(): boolean {
  if (typeof window === "undefined") return true;
  const stored = localStorage.getItem(STORAGE_KEY);
  // Defaults to true (all workspaces) until the user explicitly toggles it.
  return stored === null ? true : stored === "1";
}

const showAllWorkspaces = ref(readInitial());

// Persist changes once, at module scope, for the singleton ref.
if (typeof window !== "undefined") {
  watch(showAllWorkspaces, (v) => {
    localStorage.setItem(STORAGE_KEY, v ? "1" : "0");
  });
}

export function useDashboardScope() {
  return { showAllWorkspaces };
}
