interface User {
  id: string;
  username: string;
  email: string;
  first_name: string;
  last_name: string;
  user_type: string;
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
  }

  async function refreshToken(): Promise<boolean> {
    try {
      const response = await fetch("/api/v1/token_refresh", {
        method: "POST",
        credentials: "include",
      });

      if (!response.ok) {
        clearAuthState();
        return false;
      }

      const authResponse: AuthResponse = await response.json();
      setAuthState(authResponse);
      scheduleTokenRefresh();
      return true;
    } catch {
      clearAuthState();
      return false;
    }
  }

  async function initAuth(): Promise<void> {
    state.isLoading = true;

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
