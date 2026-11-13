interface User {
  id: string;
  username: string;
  email: string;
  first_name: string;
  last_name: string;
  user_type: string;
  avatar_url?: string;
  created_at: string;
}

interface AuthResponse {
  user: User;
  access_token: string;
  expires_at: number;
}

interface AuthState {
  user: User | null;
  accessToken: string | null;
  expiresAt: number | null;
  isLoading: boolean;
}

// Singleton state
const state = reactive<AuthState>({
  user: null,
  accessToken: null,
  expiresAt: null,
  isLoading: true,
});

let refreshTimer: ReturnType<typeof setTimeout> | null = null;
let refreshInFlight: Promise<boolean> | null = null;
let wakeListenersAdded = false;
let retryAttempt = 0;
let router: ReturnType<typeof useRouter> | null = null;
const MAX_RETRY_DELAY_MS = 60000;

export function useAuth() {
  const isAuthenticated = computed(() => !!state.accessToken && !!state.user);

  async function signup(data: {
    username: string;
    email: string;
    password: string;
    first_name: string;
    last_name: string;
  }): Promise<{ success: boolean; error?: string }> {
    try {
      const response = await fetch("/api/v1/signup", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(data),
      });

      if (!response.ok) {
        const error = await response.json();
        return { success: false, error: error.message || "Signup failed" };
      }

      const authResponse: AuthResponse = await response.json();
      setAuthState(authResponse);
      scheduleTokenRefresh();
      await useWorkspaces().listWorkspaces();
      return { success: true };
    } catch (e) {
      return { success: false, error: "Network error" };
    }
  }

  async function signin(data: {
    identifier: string;
    password: string;
  }): Promise<{ success: boolean; error?: string }> {
    try {
      const response = await fetch("/api/v1/signin", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(data),
      });

      if (!response.ok) {
        const error = await response.json();
        return { success: false, error: error.message || "Invalid credentials" };
      }

      const authResponse: AuthResponse = await response.json();
      setAuthState(authResponse);
      scheduleTokenRefresh();
      await useWorkspaces().listWorkspaces();
      return { success: true };
    } catch (e) {
      return { success: false, error: "Network error" };
    }
  }

  async function logout(): Promise<void> {
    try {
      await fetch("/api/v1/logout", {
        method: "POST",
        credentials: "include",
      });
    } catch {
      // Ignore errors, clear state anyway
    }
    clearAuthState();
    useWorkspaces().clearWorkspaces();
  }

  function refreshToken(): Promise<boolean> {
    refreshInFlight ??= doRefresh().finally(() => {
      refreshInFlight = null;
    });
    return refreshInFlight;
  }

  // Tabs share the refresh cookie and each refresh rotates it, so serialise across tabs.
  async function doRefresh(): Promise<boolean> {
    if (!navigator.locks) return refreshOnce();
    return navigator.locks.request("bureaucat-token-refresh", refreshOnce);
  }

  async function refreshOnce(): Promise<boolean> {
    let response: Response;
    try {
      response = await fetch("/api/v1/token_refresh", {
        method: "POST",
        credentials: "include",
      });
    } catch {
      if (state.user) scheduleRetry();
      return false;
    }

    if (response.status === 401) {
      const wasAuthenticated = !!state.user;
      clearAuthState();
      if (wasAuthenticated && router) {
        const redirect = router.currentRoute.value.fullPath;
        router.push({ path: "/signin", query: { redirect } });
      }
      return false;
    }

    if (!response.ok) {
      if (state.user) scheduleRetry();
      return false;
    }

    const authResponse: AuthResponse = await response.json();
    retryAttempt = 0;
    setAuthState(authResponse);
    scheduleTokenRefresh();
    return true;
  }

  function scheduleRetry(): void {
    if (refreshTimer) clearTimeout(refreshTimer);
    const delay = Math.min(1000 * 2 ** retryAttempt, MAX_RETRY_DELAY_MS);
    retryAttempt++;
    refreshTimer = setTimeout(refreshToken, delay);
  }

  function refreshIfStale(): void {
    if (document.visibilityState !== "visible" || !state.expiresAt) return;
    if (state.expiresAt * 1000 - Date.now() < 60000) refreshToken();
  }

  async function initAuth(): Promise<void> {
    state.isLoading = true;
    router = useRouter();

    if (!wakeListenersAdded) {
      wakeListenersAdded = true;
      document.addEventListener("visibilitychange", refreshIfStale);
      window.addEventListener("online", refreshIfStale);
    }

    // Try to refresh token on app load (uses httpOnly cookie)
    const success = await refreshToken();

    if (!success) {
      clearAuthState();
    }

    state.isLoading = false;
  }

  function getAuthHeader(): Record<string, string> {
    if (!state.accessToken) {
      return {};
    }
    return { Authorization: `Bearer ${state.accessToken}` };
  }

  function setAuthState(response: AuthResponse): void {
    state.user = response.user;
    state.accessToken = response.access_token;
    state.expiresAt = response.expires_at;
  }

  function clearAuthState(): void {
    state.user = null;
    state.accessToken = null;
    state.expiresAt = null;
    retryAttempt = 0;
    if (refreshTimer) {
      clearTimeout(refreshTimer);
      refreshTimer = null;
    }
  }

  function scheduleTokenRefresh(): void {
    if (refreshTimer) {
      clearTimeout(refreshTimer);
    }

    if (!state.expiresAt) return;

    // Refresh 1 minute before expiry
    const expiresAt = state.expiresAt * 1000; // Convert to ms
    const now = Date.now();
    const refreshIn = expiresAt - now - 60000; // 1 minute before expiry

    if (refreshIn > 0) {
      refreshTimer = setTimeout(() => {
        refreshToken();
      }, refreshIn);
    } else {
      // Token already expired or about to expire, refresh immediately
      refreshToken();
    }
  }

  return {
    // State (readonly)
    user: computed(() => state.user),
    accessToken: computed(() => state.accessToken),
    expiresAt: computed(() => state.expiresAt),
    isAuthenticated,
    isLoading: computed(() => state.isLoading),

    // Methods
    signup,
    signin,
    logout,
    refreshToken,
    initAuth,
    getAuthHeader,
  };
}
