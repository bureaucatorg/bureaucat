export type PATScope = "read_only" | "read_write";

interface PAToken {
  id: string;
  name: string;
  token?: string;
  scope: PATScope;
  expires_at: string | null;
  last_used_at: string | null;
  created_at: string;
}

interface PATokenListResponse {
  tokens: PAToken[];
}

export function usePAT() {
  const { getAuthHeader } = useAuth();

  async function listTokens(): Promise<{
    success: boolean;
    data?: PATokenListResponse;
    error?: string;
  }> {
    try {
      const res = await fetch("/api/v1/me/tokens", {
        headers: getAuthHeader(),
      });
      if (!res.ok) {
        const data = await res.json().catch(() => null);
        return { success: false, error: data?.message || "Failed to fetch tokens" };
      }
      const data = await res.json();
      return { success: true, data };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function createToken(
    name: string,
    scope: PATScope,
    expiresAt?: string,
  ): Promise<{ success: boolean; data?: PAToken; error?: string }> {
    try {
      const res = await fetch("/api/v1/me/tokens", {
        method: "POST",
        headers: { ...getAuthHeader(), "Content-Type": "application/json" },
        body: JSON.stringify({ name, scope, expires_at: expiresAt || null }),
      });
      if (!res.ok) {
        const data = await res.json().catch(() => null);
        return { success: false, error: data?.message || "Failed to create token" };
      }
      const data = await res.json();
      return { success: true, data };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function updateTokenScope(
    id: string,
    scope: PATScope,
  ): Promise<{ success: boolean; data?: PAToken; error?: string }> {
    try {
      const res = await fetch(`/api/v1/me/tokens/${id}`, {
        method: "PATCH",
        headers: { ...getAuthHeader(), "Content-Type": "application/json" },
        body: JSON.stringify({ scope }),
      });
      if (!res.ok) {
        const data = await res.json().catch(() => null);
        return { success: false, error: data?.message || "Failed to update token" };
      }
      const data = await res.json();
      return { success: true, data };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function deleteToken(id: string): Promise<{ success: boolean; error?: string }> {
    try {
      const res = await fetch(`/api/v1/me/tokens/${id}`, {
        method: "DELETE",
        headers: getAuthHeader(),
      });
      if (!res.ok) {
        const data = await res.json().catch(() => null);
        return { success: false, error: data?.message || "Failed to delete token" };
      }
      return { success: true };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  return { listTokens, createToken, updateTokenScope, deleteToken };
}
